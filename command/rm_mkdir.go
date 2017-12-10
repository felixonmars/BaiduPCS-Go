package baidupcscmd

import (
	"fmt"
)

// RunRemove 执行 批量删除文件/目录
func RunRemove(paths ...string) {
	paths = getAllPaths(paths...)

	pnt := func() {
		for k := range paths {
			fmt.Printf("%d: %s\n", k+1, paths[k])
		}
	}

	err := info.Remove(paths...)
	if err != nil {
		fmt.Println(err)
		fmt.Println("操作失败, 以下文件/目录删除失败: ")
		pnt()
		return
	}

	fmt.Println("操作成功, 以下文件/目录已删除: ")
	pnt()
}

// RunMkdir 执行 创建目录
func RunMkdir(path string) {
	path = getAbsPath(path)

	err := info.Mkdir(path)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("创建目录成功:", path)
}
