package pcscommand

import (
	"container/list"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/downloader"
	"github.com/iikira/BaiduPCS-Go/internal/pcsconfig"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/iikira/BaiduPCS-Go/requester"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"
)

// dtask 下载任务
type dtask struct {
	ListTask
	path         string                  // 下载的路径
	downloadInfo *baidupcs.FileDirectory // 文件或目录详情
}

func getDownloadFunc(id int, cfg *downloader.Config) baidupcs.DownloadFunc {
	if cfg == nil {
		cfg = downloader.NewConfig()
	}

	return func(downloadURL string, jar *cookiejar.Jar) error {
		h := requester.NewHTTPClient()
		h.UserAgent = pcsconfig.Config.UserAgent

		h.SetCookiejar(jar)
		h.SetKeepAlive(true)
		h.SetTimeout(10 * time.Minute)

		cfg.Client = h

		download, err := downloader.NewDownloader(downloadURL, *cfg)
		if err != nil {
			return err
		}

		exitDownloadFunc := make(chan struct{})

		download.OnExecute = func() {
			if cfg.Testing {
				fmt.Printf("[%d] 测试下载开始\n\n", id)
			}

			ds := download.GetStatusChan()
			for {
				select {
				case <-exitDownloadFunc:
					return
				case v, ok := <-ds:
					if !ok { // channel 已经关闭
						return
					}

					fmt.Printf("\r[%d] ↓ %s/%s %s/s in %s ............", id,
						pcsutil.ConvertFileSize(v.Downloaded, 2),
						pcsutil.ConvertFileSize(v.TotalSize, 2),
						pcsutil.ConvertFileSize(v.Speeds, 2),
						v.TimeElapsed,
					)
				}
			}
		}

		download.OnFinish = func() {
			exitDownloadFunc <- struct{}{}
		}

		done, err := download.Execute()
		if err != nil {
			return fmt.Errorf("[%d] 下载发生错误, %s", id, err)
		}
		<-done

		if !cfg.Testing {
			fmt.Printf("\n\n[%d] 下载完成, 保存位置: %s\n\n", id, cfg.SavePath)
		} else {
			fmt.Printf("\n\n[%d] 测试下载结束\n\n", id)
		}

		return nil
	}
}

// RunDownload 执行下载网盘内文件
func RunDownload(testing bool, parallel int, paths []string) {
	// 设置下载配置
	cfg := &downloader.Config{
		Testing:   testing,
		CacheSize: pcsconfig.Config.CacheSize,
	}

	// 设置下载最大并发量
	if parallel == 0 {
		parallel = pcsconfig.Config.MaxParallel
	}
	cfg.Parallel = parallel

	paths, err := getAllAbsPaths(paths...)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("\n")
	fmt.Printf("[0] 提示: 当前下载最大并发量为: %d, 下载缓存为: %d\n", cfg.Parallel, cfg.CacheSize)

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
			case strings.Compare(errManifest, "下载文件错误") == 0 && strings.Contains(err.Error(), "文件已存在"):
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
			if !testing { // 测试下载, 不建立空目录
				os.MkdirAll(pcsconfig.GetSavePath(task.path), 0777) // 首先在本地创建目录
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

		cfg.SavePath = pcsconfig.GetSavePath(task.path)
		err = info.DownloadFile(task.path, getDownloadFunc(task.ID, cfg))
		if err != nil {
			handleTaskErr(task, "下载文件错误", err)
			continue
		}

		totalSize += task.downloadInfo.Size
	}

	fmt.Printf("任务结束, 数据总量: %s\n", pcsutil.ConvertFileSize(totalSize))
}
