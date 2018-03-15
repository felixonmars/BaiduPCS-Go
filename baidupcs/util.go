package baidupcs

// Isdir 检查路径在网盘是否为目录
func (p *PCSApi) Isdir(pcspath string) bool {
	f, err := p.FilesDirectoriesMeta(pcspath)
	if err != nil {
		return false
	}

	return f.Isdir
}
