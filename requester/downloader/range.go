package downloader

import (
	"errors"
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
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
		total     int64
		begin     int64
		blockSize int64
		parallel  int
		count     int // 已生成次数
		mode      RangeGenMode
		mu        sync.Mutex
	}

	// RangeGenMode 线程分配方式
	RangeGenMode int
)

const (
	// RangeGenModeDefault 根据parallel平均生成
	RangeGenModeDefault RangeGenMode = iota
	// RangeGenModeBlockSize 根据blockSize生成
	RangeGenModeBlockSize
)

const (
	// DefaultBlockSize 默认的BlockSize
	DefaultBlockSize = 256 * converter.KB
)

var (
	// ErrUnknownRangeGenMode RangeGenMode 非法
	ErrUnknownRangeGenMode = errors.New("Unknown RangeGenMode")
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

// NewRangeListGenDefault 初始化默认Range生成器, 根据parallel平均生成
func NewRangeListGenDefault(totalSize, begin int64, count, parallel int) *RangeListGen {
	return &RangeListGen{
		total:    totalSize,
		begin:    begin,
		parallel: parallel,
		count:    count,
		mode:     RangeGenModeDefault,
	}
}

// NewRangeListGenBlockSize 初始化Range生成器, 根据blockSize生成
func NewRangeListGenBlockSize(totalSize, begin, blockSize int64) *RangeListGen {
	return &RangeListGen{
		total:     totalSize,
		begin:     begin,
		blockSize: blockSize,
		mode:      RangeGenModeBlockSize,
	}
}

// Mode 返回Range生成方式
func (gen *RangeListGen) Mode() RangeGenMode {
	return gen.mode
}

// LoadBegin 返回begin
func (gen *RangeListGen) LoadBegin() (begin int64) {
	gen.mu.Lock()
	begin = gen.begin
	gen.mu.Unlock()
	return
}

// LoadBlockSize 返回blockSize
func (gen *RangeListGen) LoadBlockSize() (blockSize int64) {
	switch gen.mode {
	case RangeGenModeDefault:
		if gen.blockSize <= 0 {
			gen.blockSize = gen.total / int64(gen.parallel)
		}
		blockSize = gen.blockSize
	case RangeGenModeBlockSize:
		blockSize = gen.blockSize
	}
	return
}

// IsDone 是否已分配完成
func (gen *RangeListGen) IsDone() bool {
	return gen.begin >= gen.total
}

// GenRange 生成 Range
func (gen *RangeListGen) GenRange() (index int, r *Range) {
	var (
		end int64
	)
	if gen.parallel < 1 {
		gen.parallel = 1
	}
	switch gen.mode {
	case RangeGenModeDefault:
		if gen.blockSize <= 0 {
			gen.blockSize = gen.total / int64(gen.parallel)
		}
		gen.mu.Lock()
		defer gen.mu.Unlock()

		if gen.IsDone() {
			return gen.count, nil
		}

		gen.count++
		if gen.count >= gen.parallel {
			end = gen.total - 1
		} else {
			end = int64(gen.count) * gen.blockSize
		}
		r = &Range{
			Begin: gen.begin,
			End:   end,
		}

		gen.begin = end + 1
		index = gen.count - 1
		return
	case RangeGenModeBlockSize:
		if gen.blockSize <= 0 {
			gen.blockSize = DefaultBlockSize
		}
		gen.mu.Lock()
		defer gen.mu.Unlock()

		if gen.IsDone() {
			return gen.count, nil
		}

		gen.count++
		end = gen.begin + gen.blockSize
		if end >= gen.total {
			end = gen.total - 1
		}
		r = &Range{
			Begin: gen.begin,
			End:   end,
		}
		gen.begin = end + 1
		index = gen.count - 1
		return
	}

	return 0, nil
}
