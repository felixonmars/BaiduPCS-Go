package downloader

import (
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/iikira/BaiduPCS-Go/requester"
)

var (
	// DownloadingFileSuffix 断点续传临时文件后缀
	DownloadingFileSuffix = ".downloader_downloading"

	// MinParallelSize 单个线程最小的数据量
	MinParallelSize = 128 * pcsutil.KB

	// Verbose 调试
	Verbose = false
)

// Config 下载配置
type Config struct {
	Client    *requester.HTTPClient // http 客户端
	SavePath  string                // relative or absulute path
	Parallel  int                   // 最大下载并发量
	CacheSize int                   // 下载缓冲
	Testing   bool                  // 是否测试下载
}

// NewConfig 返回预设配置
func NewConfig() *Config {
	cfg := &Config{
		Client:    requester.NewHTTPClient(),
		Parallel:  5,
		CacheSize: 2048,
	}
	return cfg
}

// Fix 修正配置信息
func (cfg *Config) Fix() {
	if cfg.CacheSize < 1024 {
		cfg.CacheSize = 1024
	}
	if cfg.Client == nil {
		cfg.Client = requester.NewHTTPClient()
	}
	if cfg.Parallel < 1 {
		cfg.Parallel = 1
	}
}
