package pcscommand

import (
	"container/list"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/internal/pcsconfig"
	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/BaiduPCS-Go/requester/downloader"
	"io"
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
	savePath     string                  // 保存的路径
	downloadInfo *baidupcs.FileDirectory // 文件或目录详情
}

//DownloadOption 下载可选参数
type DownloadOption struct {
	IsTest               bool
	IsPrintStatus        bool
	IsExecutedPermission bool
	IsOverwrite          bool
	IsShareDownload      bool
	SaveTo               string
	Parallel             int
}

func download(id int, downloadURL, savePath string, client *requester.HTTPClient, cfg *downloader.Config, isPrintStatus, isExecutedPermission bool) error {
	var (
		file     *os.File
		writerAt io.WriterAt
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

		file, err = os.OpenFile(savePath, os.O_CREATE|os.O_WRONLY, 0666)
		if file != nil {
			defer file.Close()
		}
		if err != nil {
			return fmt.Errorf("%s, %s", StrDownloadInitError, err)
		}

		// 空指针和空接口不等价
		if file != nil {
			writerAt = file
		}
	}

	download := downloader.NewDownloader(downloadURL, writerAt, cfg)
	download.SetClient(client)

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
					converter.ConvertFileSize(v.Downloaded(), 2),
					converter.ConvertFileSize(v.TotalSize(), 2),
					converter.ConvertFileSize(v.SpeedsPerSecond(), 2),
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

	if isExecutedPermission {
		err = file.Chmod(0766)
		if err != nil {
			fmt.Printf("\n\n[%d] 警告, 加执行权限错误: %s\n\n", id, err)
		}
	}

	return nil
}

func getDownloadFunc(id int, savePath string, cfg *downloader.Config, isPrintStatus, isExecutedPermission bool) baidupcs.DownloadFunc {
	if cfg == nil {
		cfg = downloader.NewConfig()
	}

	return func(downloadURL string, jar *cookiejar.Jar) error {
		h := requester.NewHTTPClient()
		h.SetCookiejar(jar)
		h.SetKeepAlive(true)
		h.SetTimeout(10 * time.Minute)
		setupHTTPClient(h)

		err := download(id, downloadURL, savePath, h, cfg, isPrintStatus, isExecutedPermission)
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
func RunDownload(paths []string, option DownloadOption) {
	// 设置下载配置
	cfg := &downloader.Config{
		IsTest:    option.IsTest,
		CacheSize: pcsconfig.Config.CacheSize(),
	}

	// 设置下载最大并发量
	if option.Parallel == 0 {
		option.Parallel = pcsconfig.Config.MaxParallel()
	}
	cfg.MaxParallel = option.Parallel

	paths, err := getAllAbsPaths(paths...)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("\n")
	fmt.Printf("[0] 提示: 当前下载最大并发量为: %d, 下载缓存为: %d\n", cfg.MaxParallel, cfg.CacheSize)

	var (
		pcs    = GetBaiduPCS()
		dlist  = list.New()
		lastID = 0
	)

	for k := range paths {
		lastID++
		ptask := &dtask{
			ListTask: ListTask{
				ID:       lastID,
				MaxRetry: 3,
			},
			path: paths[k],
		}
		if option.SaveTo != "" {
			ptask.savePath = filepath.Join(option.SaveTo, filepath.Base(paths[k]))
		} else {
			ptask.savePath = GetActiveUser().GetSavePath(paths[k])
		}
		dlist.PushBack(ptask)
		fmt.Printf("[%d] 加入下载队列: %s\n", lastID, paths[k])
	}

	var (
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
		e := dlist.Front()
		if e == nil { // 结束
			break
		}

		dlist.Remove(e) // 载入任务后, 移除队列

		task := e.Value.(*dtask)
		if task == nil {
			continue
		}

		if task.downloadInfo == nil {
			task.downloadInfo, err = pcs.FilesDirectoriesMeta(task.path)
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
			if !option.IsTest { // 测试下载, 不建立空目录
				os.MkdirAll(task.savePath, 0777) // 首先在本地创建目录, 保证空目录也能被保存
			}

			fileList, err := pcs.FilesDirectoriesList(task.path, baidupcs.DefaultOrderOptions)
			if err != nil {
				// 不重试
				fmt.Printf("[%d] 获取目录信息错误, %s\n", task.ID, err)
				continue
			}

			for k := range fileList {
				lastID++
				subTask := &dtask{
					ListTask: ListTask{
						ID:       lastID,
						MaxRetry: 3,
					},
					path:         fileList[k].Path,
					downloadInfo: fileList[k],
				}

				if option.SaveTo != "" {
					subTask.savePath = filepath.Join(task.savePath, fileList[k].Filename)
				} else {
					subTask.savePath = GetActiveUser().GetSavePath(subTask.path)
				}

				dlist.PushBack(subTask)
				fmt.Printf("[%d] 加入下载队列: %s\n", lastID, fileList[k].Path)
			}
			continue
		}

		fmt.Printf("[%d] 准备下载: %s\n", task.ID, task.path)

		if !option.IsTest && !option.IsOverwrite && fileExist(task.savePath) {
			fmt.Printf("[%d] 文件已经存在: %s, 跳过...\n", task.ID, task.savePath)
			continue
		}

		if !option.IsTest {
			fmt.Printf("[%d] 将会下载到路径: %s\n\n", task.ID, task.savePath)
		}

		// 以分享文件的方式获取下载链接来下载
		var dlink string
		if option.IsShareDownload {
			dlink = getShareDLink(task.path)
		}
		if dlink != "" {
			fmt.Printf("[%d] 获取到下载链接: %s\n", task.ID, dlink)
			err = download(task.ID, dlink, task.savePath, nil, cfg, option.IsPrintStatus, option.IsExecutedPermission)
		} else {
			err = pcs.DownloadFile(task.path, getDownloadFunc(task.ID, task.savePath, cfg, option.IsPrintStatus, option.IsExecutedPermission))
		}

		if err != nil {
			handleTaskErr(task, "下载文件错误", err)
			continue
		}

		totalSize += task.downloadInfo.Size
	}

	fmt.Printf("任务结束, 数据总量: %s\n", converter.ConvertFileSize(totalSize))
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
