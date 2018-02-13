package downloader

import (
	"time"
)

// DownloadStatus 状态
type DownloadStatus struct {
	_ bool // alignment, for 32-bit device

	Total       int64         // 总大小
	Downloaded  int64         // 已下载的数据量
	Speeds      int64         // 下载速度, 每秒
	MaxSpeeds   int64         // 最大下载速度
	TimeElapsed time.Duration // 下载的时间

	done bool // 是否已经结束
}

// GetStatusChan 返回 DownloadStatus 对象的 channel
func (der *Downloader) GetStatusChan() <-chan DownloadStatus {
	c := make(chan DownloadStatus)

	go func() {
		var old = der.status.Downloaded
		for {
			time.Sleep(1 * time.Second) // 每秒统计

			der.status.Speeds = der.status.Downloaded - old
			old = der.status.Downloaded

			if der.status.Speeds > der.status.MaxSpeeds {
				der.status.MaxSpeeds = der.status.Speeds
			}

			der.status.TimeElapsed = time.Since(der.sinceTime) / 1e6 * 1e6
			if der.status.done {
				close(c)
				return
			}

			c <- der.status
		}
	}()

	return c
}
