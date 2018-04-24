package pcscommand

import (
	"container/list"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/internal/pcsconfig"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/BaiduPCS-Go/requester/downloader"
	"github.com/iikira/BaiduPCS-Go/requester/rio"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	//DownloadSuffix 文件下载后缀
	DownloadSuffix = ".BaiduPCS-Go-downloading"
	//StrDownloadInitError 初始化下载发生错误
	StrDownloadInitError = "初始化下载发生错误"
)

// dtask 下载任务
type dtask struct {
	ListTask
	path         string                  // 下载的路径
	downloadInfo *baidupcs.FileDirectory // 文件或目录详情
}

func getDownloadFunc(id int, savePath string, cfg *downloader.Config, isPrintStatus bool) baidupcs.DownloadFunc {
	if cfg == nil {
		cfg = downloader.NewConfig()
	}

	return func(downloadURL string, jar *cookiejar.Jar) error {
		h := requester.NewHTTPClient()
		h.UserAgent = pcsconfig.Config.UserAgent

		h.SetCookiejar(jar)
		h.SetKeepAlive(true)
		h.SetTimeout(10 * time.Minute)
		h.SetHTTPSecure(pcsconfig.Config.EnableHTTPS)

		var (
			file     rio.WriteCloserAt
			err      error
			exitChan chan struct{}
		)

		if !cfg.IsTest {
			cfg.InstanceStatePath = savePath + DownloadSuffix

			// 创建下载的目录
			dir := filepath.Dir(savePath)
			fileInfo, err := os.Stat(dir)
			if err != nil {
				err = os.MkdirAll(dir, 0777)
				if err != nil {
					return err
				}
			} else if !fileInfo.IsDir() {
				return fmt.Errorf("%s, path %s: not a directory", StrDownloadInitError, dir)
			}

			file, err = os.OpenFile(savePath, os.O_CREATE|os.O_WRONLY, 0777)
			if file != nil {
				defer file.Close()
			}
			if err != nil {
				return fmt.Errorf("%s, %s", StrDownloadInitError, err)
			}
		}

		download := downloader.NewDownloader(downloadURL, file, cfg)
		download.SetClient(h)

		exitChan = make(chan struct{})

		download.OnExecute(func() {
			if cfg.IsTest {
				fmt.Printf("[%d] 测试下载开始\n\n", id)
			}

			var (
				ds                            = download.GetDownloadStatusChan()
				downloaded, totalSize, speeds int64
				leftStr                       string
			)
			for {
				select {
				case <-exitChan:
					return
				case v, ok := <-ds:
					if !ok { // channel 已经关闭
						return
					}

					downloaded, totalSize, speeds = v.Downloaded(), v.TotalSize(), v.SpeedsPerSecond()
					if speeds <= 0 {
						leftStr = "-"
					} else {
						leftStr = (time.Duration((totalSize-downloaded)/(speeds)) * time.Second).String()
					}

					fmt.Printf("\r[%d] ↓ %s/%s %s/s in %s, left %s ............", id,
						pcsutil.ConvertFileSize(v.Downloaded(), 2),
						pcsutil.ConvertFileSize(v.TotalSize(), 2),
						pcsutil.ConvertFileSize(v.SpeedsPerSecond(), 2),
						v.TimeElapsed()/1e7*1e7, leftStr,
					)
				}
			}
		})

		if isPrintStatus {
			go func() {
				for {
					time.Sleep(1 * time.Second)
					select {
					case <-exitChan:
						return
					default:
						download.PrintAllWorkers()
					}
				}
			}()
		}
		err = download.Execute()
		close(exitChan)
		if err != nil {
			return err
		}

		if !cfg.IsTest {
			fmt.Printf("\n\n[%d] 下载完成, 保存位置: %s\n\n", id, savePath)
		} else {
			fmt.Printf("\n\n[%d] 测试下载结束\n\n", id)
		}

		return nil
	}
}

// RunDownload 执行下载网盘内文件
func RunDownload(isTest, isPrintStatus bool, parallel int, paths []string) {
	// 设置下载配置
	cfg := &downloader.Config{
		IsTest:    isTest,
		CacheSize: pcsconfig.Config.CacheSize,
	}

	// 设置下载最大并发量
	if parallel == 0 {
		parallel = pcsconfig.Config.MaxParallel
	}
	cfg.MaxParallel = parallel

	paths, err := getAllAbsPaths(paths...)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("\n")
	fmt.Printf("[0] 提示: 当前下载最大并发量为: %d, 下载缓存为: %d\n", cfg.MaxParallel, cfg.CacheSize)

	dlist := list.New()
	lastID := 0

	for k := range paths {
		lastID++
		dlist.PushBack(&dtask{
			ListTask: ListTask{
				ID:       lastID,
				MaxRetry: 3,
			},
			path: paths[k],
		})
		fmt.Printf("[%d] 加入下载队列: %s\n", lastID, paths[k])
	}

	var (
		e             *list.Element
		task          *dtask
		handleTaskErr = func(task *dtask, errManifest string, err error) {
			if task == nil {
				panic("task is nil")
			}

			if err == nil {
				return
			}

			// 不重试的情况
			switch {
			case strings.Compare(errManifest, "下载文件错误") == 0 && strings.Contains(err.Error(), StrDownloadInitError):
				fmt.Printf("[%d] %s, %s\n", task.ID, errManifest, err)
				return
			}

			fmt.Printf("[%d] %s, %s, 重试 %d/%d\n", task.ID, errManifest, err, task.retry, task.MaxRetry)

			// 未达到失败重试最大次数, 将任务推送到队列末尾
			if task.retry < task.MaxRetry {
				task.retry++
				dlist.PushBack(task)
			}
			time.Sleep(3 * time.Duration(task.retry) * time.Second)
		}
		totalSize int64
	)

	for {
		e = dlist.Front()
		if e == nil { // 结束
			break
		}

		dlist.Remove(e) // 载入任务后, 移除队列

		task = e.Value.(*dtask)
		if task == nil {
			continue
		}

		if task.downloadInfo == nil {
			task.downloadInfo, err = info.FilesDirectoriesMeta(task.path)
			if err != nil {
				// 不重试
				fmt.Printf("[%d] 获取路径信息错误, %s\n", task.ID, err)
				continue
			}
		}

		fmt.Printf("\n")
		fmt.Printf("[%d] ----\n%s\n", task.ID, task.downloadInfo.String())

		// 如果是一个目录, 将子文件和子目录加入队列
		if task.downloadInfo.Isdir {
			if !isTest { // 测试下载, 不建立空目录
				os.MkdirAll(pcsconfig.GetSavePath(task.path), 0777) // 首先在本地创建目录, 保证空目录也能被保存
			}

			fileList, err := info.FilesDirectoriesList(task.path)
			if err != nil {
				// 不重试
				fmt.Printf("[%d] 获取目录信息错误, %s\n", task.ID, err)
				continue
			}

			for k := range fileList {
				lastID++
				dlist.PushBack(&dtask{
					ListTask: ListTask{
						ID:       lastID,
						MaxRetry: 3,
					},
					path:         fileList[k].Path,
					downloadInfo: fileList[k],
				})
				fmt.Printf("[%d] 加入下载队列: %s\n", lastID, fileList[k].Path)
			}
			continue
		}

		fmt.Printf("[%d] 准备下载: %s\n\n", task.ID, task.path)
		savePath := pcsconfig.GetSavePath(task.path)
		if !isTest && fileExist(savePath) {
			fmt.Printf("[%d] 文件已经存在: %s, 跳过...\n", task.ID, savePath)
			continue
		}

		err = info.DownloadFile(task.path, getDownloadFunc(task.ID, savePath, cfg, isPrintStatus))
		if err != nil {
			handleTaskErr(task, "下载文件错误", err)
			continue
		}

		totalSize += task.downloadInfo.Size
	}

	fmt.Printf("任务结束, 数据总量: %s\n", pcsutil.ConvertFileSize(totalSize))
}

// fileExist 检查文件是否存在,
// 只有当文件存在, 断点续传文件不存在时, 才判断为存在
func fileExist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		if _, err = os.Stat(path + DownloadSuffix); err != nil {
			return true
		}
	}

	return false
}
