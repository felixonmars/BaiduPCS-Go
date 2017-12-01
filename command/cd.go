package baidupcscmd

import (
	"fmt"
	"github.com/iikira/baidupcs_go/config"
)

// RunChangeDirectory 执行更改工作目录
func RunChangeDirectory(path string) {
	info.Workdir = pcsconfig.Config.Workdir
	path, err := toAbsPath(path)
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

	info.Workdir = path
	pcsconfig.Config.Workdir = path
	pcsconfig.Config.Save()

	fmt.Printf("改变工作目录: %s\n", path)
}
