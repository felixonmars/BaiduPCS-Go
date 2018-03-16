package pcsutil

import (
	"github.com/iikira/osext"
	"os"
	"path/filepath"
	"strings"
)

// ExecutablePath 获取程序所在目录
func ExecutablePath() string {
	folderPath, err := osext.ExecutableFolder()
	if err != nil {
		folderPath, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			folderPath = filepath.Dir(os.Args[0])
		}
	}
	return folderPath
}

// ExecutablePathJoin 返回程序所在目录的子目录
func ExecutablePathJoin(subPath string) string {
	return filepath.Join(ExecutablePath(), subPath)
}

// WalkDir 获取指定目录及所有子目录下的所有文件，可以匹配后缀过滤。
func WalkDir(dirPth, suffix string) (files []string, err error) {
	files = make([]string, 0, 30)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写

	err = filepath.Walk(dirPth, func(filename string, fi os.FileInfo, err error) error { //遍历目录
		if err != nil {
			return err
		}
		if fi.IsDir() { // 忽略目录
			return nil
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, filename)
		}
		return nil
	})
	return files, err
}

// ConvertToUnixPathSeparator 将 windows 目录分隔符转换为 Unix 的
func ConvertToUnixPathSeparator(p string) string {
	return strings.Replace(p, "\\", "/", -1)
}
