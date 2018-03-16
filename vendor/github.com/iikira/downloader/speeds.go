package downloader

import (
	"sync/atomic"
	"time"
	"unsafe"
)

// SpeedsStat 统计下载速度
type SpeedsStat struct {
	Readed      int64
	TimeElapsed time.Duration
	nowTime     time.Time
}

// Start 开始统计速度
func (sps *SpeedsStat) Start() {
	atomic.StoreInt64(&sps.Readed, 0)
	sps.nowTime = time.Now()
}

// AddReaded 原子操作, 增加数据量
func (sps *SpeedsStat) AddReaded(readed int64) {
	atomic.AddInt64(&sps.Readed, readed)
}

// EndAndGetSpeedsPerSecond 结束统计速度, 并返回每秒的速度
func (sps *SpeedsStat) EndAndGetSpeedsPerSecond() (speeds int64) {
	atomic.StoreInt64((*int64)(unsafe.Pointer(&sps.TimeElapsed)), int64(time.Since(sps.nowTime)))
	if sps.TimeElapsed == 0 {
		return 0
	}

	speeds = int64(float64(atomic.LoadInt64(&sps.Readed)) / sps.TimeElapsed.Seconds())
	return
}
