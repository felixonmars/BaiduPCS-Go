package baidupcscmd

import (
	"bytes"
	"fmt"
	fpath "path"
	"regexp"
	"strings"
)

func toAbsPath(path string) (string, error) {
	var _p string
	if !fpath.IsAbs(path) {
		_p = fpath.Dir(info.Workdir + "/" + path + "/")
	} else {
		_p = fpath.Dir(path + "/")
	}
	p, err := parsePath(_p)
	if err != nil {
		return "", err
	}
	return fpath.Dir(p + "/.."), nil
}

func parsePath(path string) (string, error) {
	re := regexp.MustCompile(`[\[\]\*\?]`)
	names := strings.Split(path, "/")

	var ret bytes.Buffer
	ret.WriteRune('/')
	for k := range names {
		if names[k] == "" {
			continue
		}
		if !re.MatchString(names[k]) {
			ret.WriteString("/" + names[k])
			continue
		}
		pfiles, err := info.FileList(ret.String())
		if err != nil {
			return "", err
		}

		errTime := 0
		for k2 := range pfiles {
			ok, err := fpath.Match(names[k], pfiles[k2].Filename)
			if err != nil {
				return "", err
			}
			if ok {
				ret.WriteString("/" + pfiles[k2].Filename)
				break
			} else {
				errTime++
			}
		}
		if len(pfiles) == errTime {
			return "", fmt.Errorf("文件路径匹配失败, 请检查通配符, 停留在 %s", ret.String())
		}
	}

	return ret.String(), nil
}
