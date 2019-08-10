// Package speeds 速度计算工具包
package speeds

import (
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

//Adder 增加
type Adder interface {
	Add(int64)
}

// Speeds 统计下载速度
type Speeds struct {
	readed      int64
	timeElapsed time.Duration
	inited      bool
	nowTime     time.Time
	once        sync.Once
}

// Init 初始化
func (sps *Speeds) Init() {
	sps.once.Do(func() {
		sps.nowTime = time.Now()
		sps.inited = true
	})
}

// Add 原子操作, 增加数据量
func (sps *Speeds) Add(readed int64) {
	// 初始化
	if !sps.inited {
		sps.Init()
	}

	atomic.AddInt64(&sps.readed, readed)
}

// GetSpeedsPerSecond 结束统计速度, 并返回每秒的速度
func (sps *Speeds) GetSpeedsPerSecond() (speeds int64) {
	if !sps.inited {
		sps.Init()
	}

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
