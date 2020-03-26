package pcsupload

import "sync/atomic"

type (
	UploadStatistic struct {
		totalSize int64
	}
)

func (us *UploadStatistic) AddTotalSize(size int64) int64 {
	return atomic.AddInt64(&us.totalSize, size)
}

func (us *UploadStatistic) TotalSize() int64 {
	return us.totalSize
}
