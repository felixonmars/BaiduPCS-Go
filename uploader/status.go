package uploader

import (
	"time"
)

// UploadStatus 上传状态
type UploadStatus struct {
	Length      int64         // 总大小
	Uploaded    int64         // 已上传数据
	Speed       int64         // 上传速度, 每秒
	TimeElapsed time.Duration // 上传时间
}

// startStatus 开始获取上传统计
func (u *Uploader) startStatus() {
	c := make(chan UploadStatus)

	go func() {
		t := time.Now()
		for {
			old := u.Body.Readed()

			time.Sleep(1 * time.Second) // 每秒统计

			if u.finished {
				// 上传完毕, 结束
				close(c)
				return
			}

			c <- UploadStatus{
				Length:      u.Body.Len(),
				Uploaded:    u.Body.Readed(),
				Speed:       u.Body.Readed() - old,
				TimeElapsed: time.Since(t) / 1000000 * 1000000,
			}
		}
	}()

	u.UploadStatus = c
}
