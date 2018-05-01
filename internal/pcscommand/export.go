package pcscommand

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/baidupcs"
	"path"
	"strings"
)

// RunExport 执行导出文件和目录
func RunExport(pcspaths []string, rootPath string) {
	pcspaths, err := getAllAbsPaths(pcspaths...)
	if err != nil {
		fmt.Println(err)
		return
	}

	pcs := GetBaiduPCS()

	for _, pcspath := range pcspaths {
		getPath := func(p string) string {
			if rootPath == "" {
				return p
			}

			// 是一个单独文件
			if pcspath == p {
				return path.Join(rootPath, strings.TrimPrefix(p, path.Dir(pcspath)))
			}

			return path.Join(rootPath, strings.TrimPrefix(p, pcspath))
		}

		var d int
		pcs.FilesDirectoriesRecurseList(pcspath, func(depth int, fd *baidupcs.FileDirectory) {
			if fd.Isdir {
				if depth > d {
					d = depth
				} else {
					fmt.Printf("BaiduPCS-Go mkdir \"%s\"\n", getPath(fd.Path))
					d = 0
				}
				return
			}

			fmt.Printf("BaiduPCS-Go rapidupload -length=%d -md5=%s \"%s\"\n", fd.Size, fd.MD5, getPath(fd.Path))
		})
	}
}
