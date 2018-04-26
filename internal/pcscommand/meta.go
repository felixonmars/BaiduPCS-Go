package pcscommand

import (
	"fmt"
)

// RunGetMeta 执行 获取单个文件/目录的元信息
func RunGetMeta(path string) {
	p, err := getAbsPath(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	data, err := GetBaiduPCS().FilesDirectoriesMeta(p)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println()
	fmt.Println(data)
}
