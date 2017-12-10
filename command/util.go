package baidupcscmd

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/config"
	fpath "path"
	"regexp"
	"strings"
)

var (
	patternRE = regexp.MustCompile(`[\[\]\*\?]`)
)

// getAbsPath 获取绝对路径, 忽略通配符
func getAbsPath(path string) string {
	if !fpath.IsAbs(path) {
		path = fpath.Dir(pcsconfig.ActiveBaiduUser.Workdir + "/" + path + "/")
	}
	return path
}

func getAllPaths(paths ...string) (_paths []string) {
	for k := range paths {
		_paths = append(_paths, parsePath(paths[k])...)
	}
	return
}

func toAbsPath(path string) (string, error) {
	p := parsePath(path)
	if len(p) == 0 {
		return "", fmt.Errorf("文件路径匹配失败, 请检查通配符")
	}
	return p[0], nil
}

// parsePath 递归解析通配符
func parsePath(path string) (paths []string) {
	path = getAbsPath(path)

	if !patternRE.MatchString(path) {
		paths = []string{path}
		return
	}

	if _, err := fpath.Match(path, ""); err != nil {
		return nil
	}

	names := strings.Split(path, "/")

	for k := range names {
		if names[k] == "" || !patternRE.MatchString(names[k]) {
			continue
		}

		pfiles, err := info.FileList(strings.Join(names[:k], "/"))
		if err != nil {
			fmt.Println(err)
			return nil
		}

		for k2 := range pfiles {
			ok, _ := fpath.Match(names[k], pfiles[k2].Filename)
			if ok {
				if k >= len(names)-1 {
					paths = append(paths, strings.Join(names[:k], "/")+"/"+pfiles[k2].Filename)
				} else if pfiles[k2].Isdir {
					paths = append(paths, parsePath(pfiles[k2].Path+"/"+strings.Join(names[k+1:], "/"))...)
				}
			}
		}
		break
	}

	return
}

func recurseFDCountTotalSize(path string) (fileN, directoryN, size int64) {
	di, err := info.FileList(path)
	if err != nil {
		fmt.Println(err)
	}

	for k := range di {
		if di[k].Isdir {
			f, d, s := recurseFDCountTotalSize(di[k].Path)
			fileN += f
			directoryN += d
			size += s
		}
	}
	f, d := di.Count()
	s := di.TotalSize()
	fileN += f
	directoryN += d
	size += s
	return
}
