package baidupcs

import (
	"path"
)

// IsFile 检查路径在网盘中是否为文件
func (pcs *BaiduPCS) IsFile(pcspath string) (isfile bool) {
	if path.Clean(pcspath) == "/" {
		return false
	}

	f, err := pcs.FilesDirectoriesMeta(pcspath)
	if err != nil {
		return false
	}

	return !f.Isdir
}
