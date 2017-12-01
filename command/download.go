package baidupcscmd

import (
	"fmt"
	"github.com/iikira/baidupcs_go/downloader"
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
	fmt.Printf("即将开始下载文件\n\n")

	if downloadInfo.Isdir {
		fmt.Printf("错误: 暂时不支持下载目录 (文件夹)\n\n")
		return
	}

	err = info.FileDownload(path, downloadInfo.Size)
	if err != nil {
		fmt.Println(err)
		return
	}
}
