package downloader

import (
	"time"
)

// DownloadStatus 状态
type DownloadStatus struct {
	Downloaded int64 `json:"downloaded"`
	Speeds     int64
	MaxSpeeds  int64

	done bool // 是否已经结束
}

// GetStatusChan 返回 DownloadStatus 对象的 channel
func (der *Downloader) GetStatusChan() <-chan DownloadStatus {
	c := make(chan DownloadStatus)

	go func() {
		var old = der.status.Downloaded
		for {
			if der.status.done {
				return
			}

			time.Sleep(time.Second * 1)
			der.status.Speeds = der.status.Downloaded - old
			old = der.status.Downloaded

			if der.status.Speeds > der.status.MaxSpeeds {
				der.status.MaxSpeeds = der.status.Speeds
			}

			c <- der.status
		}
	}()

	return c
}
