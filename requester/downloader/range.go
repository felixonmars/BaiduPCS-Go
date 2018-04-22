package downloader

import (
	"fmt"
	"sync/atomic"
)

//Range 请求范围
type Range struct {
	Begin int64
	End   int64
}

//RangeList 请求范围列表
type RangeList []*Range

//Len 长度
func (r *Range) Len() int64 {
	return r.LoadEnd() - r.LoadBegin() + 1
}

//LoadBegin 读取Begin, 原子操作
func (r *Range) LoadBegin() int64 {
	return atomic.LoadInt64(&r.Begin)
}

//AddBegin 增加Begin, 原子操作
func (r *Range) AddBegin(i int64) {
	atomic.AddInt64(&r.Begin, i)
}

//StoreBegin 储存End, 原子操作
func (r *Range) StoreBegin(end int64) {
	atomic.StoreInt64(&r.Begin, end)
}

//LoadEnd 读取End, 原子操作
func (r *Range) LoadEnd() int64 {
	return atomic.LoadInt64(&r.End)
}

//StoreEnd 储存End, 原子操作
func (r *Range) StoreEnd(end int64) {
	atomic.StoreInt64(&r.End, end)
}

func (r *Range) String() string {
	return fmt.Sprintf("{%d-%d}", r.Begin, r.End)
}

//Len 获取所有的Range长度
func (rl *RangeList) Len() int64 {
	var l int64
	for _, wrange := range *rl {
		if wrange == nil {
			continue
		}
		l += wrange.Len()
	}
	return l
}
