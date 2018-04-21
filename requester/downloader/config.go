package downloader

const (
	//CacheSize 默认的下载缓存
	CacheSize = 8192
)

var (
	// MinParallelSize 单个线程最小的数据量
	MinParallelSize int64 = 128 * 1024 // 128kb
)

//Config 下载配置
type Config struct {
	MaxParallel       int    // 最大下载并发量
	CacheSize         int    // 下载缓冲
	InstanceStatePath string // 断点续传信息路径
	IsTest            bool   // 是否测试下载
	cacheSize         int    // 实际下载缓存
	parallel          int    // 实际的下载并行量
}

//NewConfig 返回默认配置
func NewConfig() *Config {
	return &Config{
		MaxParallel: 5,
		CacheSize:   CacheSize,
		IsTest:      false,
	}
}

//Fix 修复配置信息, 使其合法
func (cfg *Config) Fix() {
	fixCacheSize(&cfg.CacheSize)
	if cfg.MaxParallel < 1 {
		cfg.MaxParallel = 1
	}
}
