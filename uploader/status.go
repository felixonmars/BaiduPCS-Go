package uploader

import (
	"time"
)

// UploadStatus 上传状态
type UploadStatus struct {
	Length   int64 // 总大小
	Uploaded int64 // 已上传数据
	Speed    int64 // 上传速度, 每秒
}

// GetStatusChan 返回 UploadStatus 对象的 channel
func (u *Uploader) GetStatusChan() <-chan UploadStatus {
	c := make(chan UploadStatus)

	go func() {
		for {
			old := u.Reader.getReaded()
			if u.Reader.Len() != 0 && old == u.Reader.Len() {
				// 上传完毕, 结束
				return
			}

			time.Sleep(1 * time.Second) // 每秒统计
			c <- UploadStatus{
				Length:   u.Reader.Len(),
				Uploaded: u.Reader.getReaded(),
				Speed:    u.Reader.getReaded() - old,
			}
		}
	}()

	return c
}
