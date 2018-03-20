package downloader

import (
	"io"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// SpeedsStat 统计下载速度
type SpeedsStat struct {
	readed      int64
	timeElapsed time.Duration
	nowTime     time.Time
	once        sync.Once
}

// AddReaded 原子操作, 增加数据量
func (sps *SpeedsStat) AddReaded(readed int64) {
	// 初始化
	sps.once.Do(func() {
		if sps.nowTime.Unix() == 0 {
			sps.nowTime = time.Now()
		}
	})

	atomic.AddInt64(&sps.readed, readed)
}

// GetSpeedsPerSecond 结束统计速度, 并返回每秒的速度
func (sps *SpeedsStat) GetSpeedsPerSecond() (speeds int64) {
	int64Ptr := (*int64)(unsafe.Pointer(&sps.timeElapsed))
	atomic.StoreInt64(int64Ptr, (int64)(time.Since(sps.nowTime)))
	if atomic.LoadInt64(int64Ptr) == 0 {
		return 0
	}

	speeds = int64(float64(atomic.LoadInt64(&sps.readed)) / sps.timeElapsed.Seconds())

	atomic.StoreInt64(&sps.readed, 0)
	sps.nowTime = time.Now()
	return
}

func readFullFrom(r io.Reader, buf []byte, stats ...*SpeedsStat) (n int, err error) {
	for n < len(buf) && err == nil {
		var nn int
		nn, err = r.Read(buf[n:])

		// 更新速度统计
		for _, stat := range stats {
			if stat == nil {
				continue
			}
			stat.AddReaded(int64(nn))
		}
		n += nn
	}
	if n >= len(buf) {
		err = nil
	} else if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return
}
