package pcsdownload

// IsSkipMd5Checksum 是否忽略某些校验
func IsSkipMd5Checksum(size int64, md5Str string) bool {
	switch {
	case size == 1749504 && md5Str == "48bb9b0361dc9c672f3dc7b3ffcfde97": //8秒温馨提示
		fallthrough
	case size == 120 && md5Str == "6c1b84914588d09a6e5ec43605557457": //温馨提示文字版
		return true
	}
	return false
}
