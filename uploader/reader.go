package uploader

import (
	"bytes"
	"io"
	"strings"
	"sync/atomic"
)

// ReaderLen 实现 读 和 长度 接口
type ReaderLen interface {
	io.Reader
	Len() int64
}

type reader struct {
	uploadReaderLen ReaderLen
	multipart       *bytes.Buffer
	multipartEnd    *strings.Reader

	readed int64 // 已读取的数据量
}

func (r *reader) Read(p []byte) (n int, err error) {
	var n1, n2, n3 int
	if r.multipart != nil {
		n1, err = r.multipart.Read(p[:])
		r.addReaded(int64(n))
	}
	if n1 < len(p) {
		n2, err = r.uploadReaderLen.Read(p[n1:])
		r.addReaded(int64(n2))
		if n1+n2 < len(p) {
			if r.multipartEnd != nil {
				n3, err = r.multipartEnd.Read(p[n1+n2:])
				r.addReaded(int64(n3))
			}
		}
	}
	n = n1 + n2 + n3
	return
}

func (r *reader) totalLen() int64 {
	if r.uploadReaderLen == nil {
		return 0
	}

	if r.multipart == nil || r.multipartEnd == nil {
		return r.uploadReaderLen.Len()
	}

	laceLen := int64(r.multipart.Len()) + int64(r.multipartEnd.Len())

	return r.uploadReaderLen.Len() + laceLen
}

func (r *reader) getReaded() (readed int64) {
	return atomic.LoadInt64(&r.readed) // 原子操作
}

func (r *reader) addReaded(delta int64) (readed int64) {
	return atomic.AddInt64(&r.readed, delta) // 原子操作
}
