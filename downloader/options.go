package downloader

import (
	"github.com/iikira/BaiduPCS-Go/requester"
)

// Options 下载可选项
type Options struct {
	Client    *requester.HTTPClient
	Testing   bool // 测试下载
	Parallel  int
	CacheSize int
}

// NewOptions 返回预设配置
func NewOptions() *Options {
	return &Options{
		Client:    requester.NewHTTPClient(),
		Testing:   false,
		Parallel:  5,
		CacheSize: 2048,
	}
}

// SetMaxParallel 设置最大下载并发量
func (o *Options) SetMaxParallel(t int) {
	if t <= 0 {
		panic("SetMaxParallel: zero or negative parallel")
	}
	o.Parallel = t
}

// SetCacheSize 设置缓冲大小
func (o *Options) SetCacheSize(size int) {
	if size < 1024 {
		o.CacheSize = 1024
		return
	}
	o.CacheSize = size
}
