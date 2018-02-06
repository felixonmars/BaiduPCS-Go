package downloader

import (
	"regexp"
)

var (
	maxParallel = 5
	cacheSize   = 1024

	// FileNameRE 正则表达式: 匹配文件名
	FileNameRE = regexp.MustCompile("filename=\"(.*?)\"")
)

// SetMaxParallel 设置最大下载并发量
func SetMaxParallel(t int) {
	if t <= 0 {
		panic("downloader.SetMaxParallel: zero or negative parallel")
	}
	maxParallel = t
}

// SetCacheSize 设置缓冲大小
func SetCacheSize(size int) {
	if size < 1024 {
		cacheSize = 1024
		return
	}
	cacheSize = size
}
