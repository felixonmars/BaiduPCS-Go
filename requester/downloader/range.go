package downloader

import (
	"fmt"
	"sync"
	"sync/atomic"
)

type (
	//Range 请求范围
	Range struct {
		Begin int64 `json:"begin"`
		End   int64 `json:"end"`
	}

	//RangeList 请求范围列表
	RangeList []*Range

	//RangeListGen Range 生成器
	RangeListGen struct {
		begin     int64
		end       int64
		parallel  int
		count     int // 已生成次数
		blockSize int64
		total     int64
		left      int64
		mode      RangeGenMode
		mu        sync.Mutex
	}

	//RangeListGenFunc Range 生成函数
	RangeListGenFunc func() *Range

	// RangeGenMode 线程分配方式
	RangeGenMode int
)

const (
	// RangeGenMode1 根据parallel平均生成
	RangeGenMode1 RangeGenMode = iota
	// RangeGenMode2 根据range size生成
	RangeGenMode2
)

//Len 长度
func (r *Range) Len() int64 {
	return r.LoadEnd() - r.LoadBegin() + 1
}

//LoadBegin 读取Begin, 原子操作
func (r *Range) LoadBegin() int64 {
	return atomic.LoadInt64(&r.Begin)
}

//AddBegin 增加Begin, 原子操作
func (r *Range) AddBegin(i int64) (newi int64) {
	return atomic.AddInt64(&r.Begin, i)
}

//LoadEnd 读取End, 原子操作
func (r *Range) LoadEnd() int64 {
	return atomic.LoadInt64(&r.End)
}

//StoreBegin 储存End, 原子操作
func (r *Range) StoreBegin(end int64) {
	atomic.StoreInt64(&r.Begin, end)
}

//StoreEnd 储存End, 原子操作
func (r *Range) StoreEnd(end int64) {
	atomic.StoreInt64(&r.End, end)
}

func (r *Range) String() string {
	return fmt.Sprintf("{%d-%d}", r.LoadBegin(), r.LoadEnd())
}

//Len 获取所有的Range的剩余长度
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

func NewRangeListGen1(totalSize int64, parallel int) *RangeListGen {
	return &RangeListGen{
		begin:    0,
		end:      0,
		parallel: parallel,
		count:    0,
		total:    totalSize,
		left:     0,
		mode:     RangeGenMode1,
	}
}

func (gen *RangeListGen) GenFunc() (blockSize int64, rangeGenF RangeListGenFunc) {
	switch gen.mode {
	case RangeGenMode1:
		blockSize = gen.total / int64(gen.parallel)
		rangeGenF = func() *Range {
			gen.count++
			if gen.count > gen.parallel {
				return nil
			}

			if gen.count == gen.parallel {
				gen.end = gen.total - 1
			} else {
				gen.end = int64(gen.count) * blockSize
			}
			r := &Range{
				Begin: gen.begin,
				End:   gen.end,
			}
			gen.begin = gen.end + 1
			gen.left = gen.total - gen.end - 1
			return r
		}
	case RangeGenMode2:
	}

	return
}
