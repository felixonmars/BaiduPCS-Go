package downloader

//trigger 用于触发事件
func trigger(f func()) {
	if f == nil {
		return
	}
	go f()
}

func fixCacheSize(size *int) {
	if *size < 1024 {
		*size = 1024
	}
}
