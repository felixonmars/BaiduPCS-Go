package pcscommand

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"github.com/iikira/BaiduPCS-Go/downloader"
	"github.com/iikira/BaiduPCS-Go/pcsconfig"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/iikira/BaiduPCS-Go/requester"
	"net/http/cookiejar"
	"os"
	"strings"
	"time"
)

// downloadFunc 用于下载文件的函数
type downloadFunc func(downloadURL string, jar *cookiejar.Jar, savePath string) error

func getDownloadFunc(o *downloader.Options) downloadFunc {
	return func(downloadURL string, jar *cookiejar.Jar, savePath string) error {
		if o == nil {
			o = downloader.NewOptions()
		}

		h := requester.NewHTTPClient()
		h.UserAgent = pcsconfig.Config.UserAgent

		h.SetCookiejar(jar)
		h.SetKeepAlive(true)
		h.SetTimeout(2 * time.Minute)

		o.Client = h

		downloader, err := downloader.NewDownloader(downloadURL, savePath, o)
		if err != nil {
			return err
		}

		exitDownloadFunc := make(chan struct{})

		downloader.OnStart(func() {
			if o.Testing {
				fmt.Printf("测试下载开始\n\n")
			}

			ds := downloader.GetStatusChan()
			for {
				select {
				case v, ok := <-ds:
					if !ok { // channel 已经关闭
						return
					}

					fmt.Printf("\r↓ %s/%s %s/s %s ............",
						pcsutil.ConvertFileSize(v.Downloaded, 2),
						pcsutil.ConvertFileSize(v.Total, 2),
						pcsutil.ConvertFileSize(v.Speeds, 2),
						v.TimeElapsed,
					)
				}
			}
		})

		downloader.OnFinish(func() {
			exitDownloadFunc <- struct{}{}
		})

		downloader.StartDownload()
		<-exitDownloadFunc

		if !o.Testing {
			fmt.Printf("\n\n下载完成, 保存位置: %s\n\n", savePath)
		} else {
			fmt.Printf("\n\n测试下载结束\n\n")
		}

		close(exitDownloadFunc)
		return nil
	}
}

// RunDownload 执行下载网盘内文件
func RunDownload(testing bool, parallel int, paths []string) {
	// 设置下载选项
	o := &downloader.Options{
		Testing:   testing,
		CacheSize: pcsconfig.Config.CacheSize,
	}

	// 设置下载最大并发量
	if parallel == 0 {
		parallel = pcsconfig.Config.MaxParallel
	}
	o.SetMaxParallel(parallel)

	paths, err := getAllAbsPaths(paths...)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println()
	for k := range paths {
		fmt.Printf("添加下载任务: %s\n", paths[k])
	}
	fmt.Println()

	for k, path := range paths {
		downloadInfo, err := info.FilesDirectoriesMeta(path)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Printf("[ %d / %d ] %s\n", k+1, len(paths), downloadInfo.String())

		// 如果是一个目录, 递归下载该目录下的所有文件
		if downloadInfo.Isdir {
			fmt.Printf("即将下载目录: %s, 获取目录信息中...\n\n", path)

			dirInfo, err := info.FilesDirectoriesList(path, true)
			if err != nil {
				fmt.Printf("发生错误, %s\n", err)
				continue
			}

			fN, dN := dirInfo.Count()
			statText := fmt.Sprintf("统计: 文件总数: %d, 目录总数: %d, 文件总大小: %s\n\n",
				fN, dN,
				pcsutil.ConvertFileSize(dirInfo.TotalSize()),
			)

			fmt.Printf(statText) // 输出统计信息

			downloadDirectory(path, dirInfo, o) // 开始下载目录

			fmt.Printf("目录 %s 下载完成, %s", path, statText) // 再次输出统计信息

			continue
		}

		fmt.Printf("即将开始下载文件\n\n")

		err = info.FileDownload(path, getDownloadFunc(o))
		if err != nil {
			fmt.Printf("下载文件时发生错误: %s (跳过...)\n\n", err)
		}
	}
}

// downloadDirectory 下载目录
func downloadDirectory(pcspath string, dirInfo baidupcs.FileDirectoryList, o *downloader.Options) {
	// 遇到空目录, 则创建目录
	if len(dirInfo) == 0 {
		fmt.Printf("创建目录: %s\n\n", pcspath)
		os.MkdirAll(pcsconfig.GetSavePath(pcspath), 0777)
		return
	}

	for k := range dirInfo {
		if dirInfo[k] == nil {
			continue
		}

		if dirInfo[k].Children != nil {
			downloadDirectory(dirInfo[k].Path, dirInfo[k].Children, o)
		}

		// 如果文件或目录存在, 跳过
		if pcsconfig.CheckFileExist(dirInfo[k].Path) {
			// 如果是目录, 不输出消息
			if !dirInfo[k].Isdir {
				fmt.Printf("文件已存在 (自动跳过): %s\n\n", pcsconfig.GetSavePath(dirInfo[k].Path))
			}

			continue
		}

		fmt.Println(dirInfo[k]) // 输出文件或目录的详情

		if dirInfo[k].Isdir {
			downloadDirectory(dirInfo[k].Path, nil, o)
			continue
		}

		fmt.Printf("即将开始下载文件: %s\n\n", dirInfo[k].Filename)

		err := info.FileDownload(dirInfo[k].Path, getDownloadFunc(o))
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(strings.Repeat("-", 60))
	}
}
