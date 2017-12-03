package baidupcscmd

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/config"
	"github.com/iikira/BaiduPCS-Go/downloader"
	"github.com/iikira/BaiduPCS-Go/util"
	"os"
	"path/filepath"
)

// RunDownload 执行下载网盘内文件
func RunDownload(path string) {
	downloader.SetCacheSize(2048)
	downloader.SetMaxParallel(thread)

	path, err := toAbsPath(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	downloadInfo, err := info.FilesDirectoriesMeta(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(downloadInfo)

	// 如果是一个目录, 递归下载该目录下的所有文件
	if downloadInfo.Isdir {
		fmt.Printf("即将下载目录: %s\n\n", path)

		fileN, directoryN, size := recurseFDCountTotalSize(path)
		statText := fmt.Sprintf("统计: 目录总数: %d, 文件总数: %d, 文件总大小: %s\n\n",
			directoryN,
			fileN,
			pcsutil.ConvertFileSize(size),
		)

		fmt.Printf(statText)
		downloadDirectory(path)
		fmt.Printf("目录 %s 下载完成, %s", path, statText)
		return
	}

	fmt.Printf("即将开始下载文件\n\n")

	err = info.FileDownload(path, downloadInfo.Size)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func downloadDirectory(path string) {
	di, err := info.FileList(path)
	if err != nil {
		fmt.Println("发生错误,", err)
	}

	// 遇到空目录, 则创建目录
	if len(di) == 0 {
		os.MkdirAll("download/"+pcsconfig.ActiveBaiduUser.Name+path, 0777)
		return
	}

	for k := range di {
		if di[k].Isdir {
			downloadDirectory(di[k].Path)
			continue
		}

		// 如果文件存在, 跳过
		savePath := "download" + filepath.Dir("/"+pcsconfig.ActiveBaiduUser.Name+di[k].Path+"/..")
		if _, err = os.Stat(savePath); err == nil {
			if _, err = os.Stat(savePath + downloader.DownloadingFileSuffix); err != nil {
				fmt.Printf("文件已存在 (自动跳过): %s\n\n", savePath)
				continue
			}
		}

		fmt.Println(di[k])
		fmt.Printf("即将开始下载文件: %s\n\n", di[k].Filename)

		err = info.FileDownload(di[k].Path, di[k].Size)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("------------------------------------------------------------")
	}
}
