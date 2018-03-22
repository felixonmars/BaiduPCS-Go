package baidupcs

// Isdir 检查路径在网盘是否为目录
func (pcs *BaiduPCS) Isdir(pcspath string) bool {
	f, err := pcs.FilesDirectoriesMeta(pcspath)
	if err != nil {
		return false
	}

	return f.Isdir
}
