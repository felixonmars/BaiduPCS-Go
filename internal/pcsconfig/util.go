package pcsconfig

// AverageParallel 返回平均的下载最大并发量
func AverageParallel(parallel, downloadLoad int) int {
	if downloadLoad < 1 {
		return 1
	}

	p := parallel / downloadLoad
	if p < 1 {
		return 1
	}
	return p
}
