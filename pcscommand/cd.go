package pcscommand

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsconfig"
)

// RunChangeDirectory 执行更改工作目录
func RunChangeDirectory(path string, isList bool) {
	path, err := getAbsPath(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	data, err := info.FilesDirectoriesMeta(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !data.Isdir {
		fmt.Printf("错误: %s 不是一个目录 (文件夹)\n", path)
		return
	}

	pcsconfig.ActiveBaiduUser.Workdir = path
	pcsconfig.Config.Save()

	fmt.Printf("改变工作目录: %s\n", path)

	if isList {
		RunLs(".")
	}
}
