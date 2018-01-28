package pcscommand

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsconfig"
	fpath "path"
	"regexp"
	"strings"
)

var (
	patternRE = regexp.MustCompile(`[\[\]\*\?]`)
)

// getAbsPathNoMatch 获取绝对路径, 不检测通配符
func getAbsPathNoMatch(path string) string {
	if !fpath.IsAbs(path) {
		path = fpath.Dir(pcsconfig.ActiveBaiduUser.Workdir + "/" + path + "/")
	}
	return path
}

// getAllAbsPaths 获取所有绝对路径
func getAllAbsPaths(paths ...string) (_paths []string, err error) {
	for k := range paths {
		p, err := parsePath(paths[k])
		if err != nil {
			return nil, err
		}
		_paths = append(_paths, p...)
	}
	return
}

// getAbsPath 获取绝对路径, 获取错误将会返回 原路径 和 错误信息
func getAbsPath(path string) (string, error) {
	p, err := parsePath(path)
	if err != nil {
		return path, err
	}

	if len(p) != 0 {
		return p[0], nil
	}
	return "", fmt.Errorf("未找到路径")
}

// parsePath 递归解析通配符
func parsePath(path string) (paths []string, err error) {
	path = getAbsPathNoMatch(path)

	if patternRE.MatchString(path) {
		paths = recurseParsePath(path)
		if len(paths) == 0 {
			return nil, fmt.Errorf("文件路径匹配失败, 请检查通配符")
		}
		return paths, nil
	}

	_, err = info.FilesDirectoriesMeta(path)
	if err != nil {
		return nil, err
	}
	paths = []string{path}
	return
}

func recurseParsePath(path string) (paths []string) {
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
					paths = append(paths, recurseParsePath(pfiles[k2].Path+"/"+strings.Join(names[k+1:], "/"))...)
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
