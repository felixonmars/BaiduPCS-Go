package baidupcscmd

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/config"
	"github.com/iikira/BaiduPCS-Go/downloader"
	"github.com/iikira/BaiduPCS-Go/util"
	"os"
)

// RunDownload 执行下载网盘内文件
func RunDownload(paths ...string) {
	downloader.SetCacheSize(2048)
	downloader.SetMaxParallel(pcsconfig.Config.MaxParallel)

	var _paths []string
	for k := range paths {
		_paths = append(_paths, parsePath(paths[k])...)
	}

	fmt.Println()
	for k := range _paths {
		fmt.Printf("添加下载任务: %s\n", _paths[k])
	}
	fmt.Println()

	for k, path := range _paths {
		downloadInfo, err := info.FilesDirectoriesMeta(path)
		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Printf("[ %d / %d ] %s\n", k+1, len(_paths), downloadInfo.String())

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
			continue
		}

		fmt.Printf("即将开始下载文件\n\n")

		err = info.FileDownload(path, downloadInfo.Size)
		if err != nil {
			fmt.Printf("下载文件时发生错误: %s (跳过...)\n\n", err)
		}
	}
}

func downloadDirectory(path string) {
	di, err := info.FileList(path)
	if err != nil {
		fmt.Println("发生错误,", err)
	}

	// 遇到空目录, 则创建目录
	if len(di) == 0 {
		os.MkdirAll(pcsconfig.GetSavePath(path), 0777)
		return
	}

	for k := range di {
		if di[k].Isdir {
			downloadDirectory(di[k].Path)
			continue
		}

		// 如果文件存在, 跳过
		if pcsconfig.CheckFileExist(di[k].Path) {
			fmt.Printf("文件已存在 (自动跳过): %s\n\n", pcsconfig.GetSavePath(di[k].Path))
			continue
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
