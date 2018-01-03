package uploader

import (
	"io"
	"sync/atomic"
)

type reader struct {
	reader io.Reader
	length int64
	readed int64 // 已读取的数据量
}

func (r *reader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	r.addReaded(int64(n))
	return n, err
}

func (r *reader) getReaded() (readed int64) {
	return atomic.LoadInt64(&r.readed) // 原子操作
}

func (r *reader) addReaded(delta int64) (readed int64) {
	return atomic.AddInt64(&r.readed, delta) // 原子操作
}
