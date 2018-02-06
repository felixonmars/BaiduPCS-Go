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

func downloadFunc(downloadURL string, jar *cookiejar.Jar, savePath string) error {
	h := requester.NewHTTPClient()

	h.SetCookiejar(jar)
	h.SetKeepAlive(true)
	h.SetTimeout(2 * time.Minute)

	fileDl, err := downloader.NewFileDl(h, downloadURL, savePath)
	if err != nil {
		return err
	}

	pa := make(chan struct{})
	exit := make(chan bool)

	fileDl.OnStart(func() {
		t1 := time.Now()
		for {
			status := fileDl.GetStatus()

			select {
			case <-exit:
				return
			default:
				time.Sleep(time.Second * 1)
				fmt.Printf("\r%v/%v %v/s time: %s %v",
					pcsutil.ConvertFileSize(status.Downloaded, 2),
					pcsutil.ConvertFileSize(fileDl.Size, 2),
					pcsutil.ConvertFileSize(status.Speeds, 2),
					time.Since(t1)/1000000*1000000,
					"[DOWNLOADING]"+strings.Repeat(" ", 10),
				)
				os.Stdout.Sync()
			}
		}
	})

	fileDl.OnFinish(func() {
		exit <- true
		pa <- struct{}{}
	})

	fileDl.Start()
	<-pa
	fmt.Printf("\n\n下载完成, 保存位置: %s\n\n", savePath)
	return nil
}

// RunDownload 执行下载网盘内文件
func RunDownload(paths ...string) {
	downloader.SetCacheSize(pcsconfig.Config.CacheSize)
	downloader.SetMaxParallel(pcsconfig.Config.MaxParallel)

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
			statText := fmt.Sprintf("统计: 目录总数: %d, 文件总数: %d, 文件总大小: %s\n\n",
				fN, dN,
				pcsutil.ConvertFileSize(dirInfo.TotalSize()),
			)

			fmt.Printf(statText) // 输出统计信息

			downloadDirectory(path, dirInfo) // 开始下载目录

			fmt.Printf("目录 %s 下载完成, %s", path, statText) // 再次输出统计信息

			continue
		}

		fmt.Printf("即将开始下载文件\n\n")

		err = info.FileDownload(path, downloadFunc)
		if err != nil {
			fmt.Printf("下载文件时发生错误: %s (跳过...)\n\n", err)
		}
	}
}

func downloadDirectory(pcspath string, dirInfo baidupcs.FileDirectoryList) {
	// 遇到空目录, 则创建目录
	if len(dirInfo) == 0 {
		os.MkdirAll(pcsconfig.GetSavePath(pcspath), 0777)
		return
	}

	for k := range dirInfo {
		if dirInfo[k] == nil {
			continue
		}

		if dirInfo[k].Children != nil {
			downloadDirectory(dirInfo[k].Path, dirInfo[k].Children)
		}

		// 如果文件存在, 跳过
		if pcsconfig.CheckFileExist(dirInfo[k].Path) {
			fmt.Printf("文件已存在 (自动跳过): %s\n\n", pcsconfig.GetSavePath(dirInfo[k].Path))
			continue
		}

		fmt.Println(dirInfo[k])
		fmt.Printf("即将开始下载文件: %s\n\n", dirInfo[k].Filename)

		err := info.FileDownload(dirInfo[k].Path, downloadFunc)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(strings.Repeat("-", 60))
	}
}
