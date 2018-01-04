package uploader

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync/atomic"
)

type reader struct {
	uploadReader io.Reader
	multipart    *bytes.Buffer
	multipartEnd *strings.Reader
	size         int64
	readed       int64 // 已读取的数据量
}

func (r *reader) Read(p []byte) (n int, err error) {
	var n1, n2, n3 int
	n1, err = r.multipart.Read(p[:])
	r.addReaded(int64(n))
	if n1 < len(p) {
		n2, err = r.uploadReader.Read(p[n1:])
		r.addReaded(int64(n2))
		if n1+n2 < len(p) {
			n3, err = r.multipartEnd.Read(p[n1+n2:])
			r.addReaded(int64(n3))
		}
	}
	n = n1 + n2 + n3
	return
}

func (r *reader) Len() int64 {
	if r.uploadReader == nil || r.multipart == nil || r.multipartEnd == nil {
		return 0
	}

	laceLen := int64(r.multipart.Len()) + int64(r.multipartEnd.Len())

	// 尝试获取大小
	switch v := r.uploadReader.(type) {
	case *bytes.Buffer:
		return int64(v.Len()) + laceLen
	case *bytes.Reader:
		return int64(v.Len()) + laceLen
	case *strings.Reader:
		return int64(v.Len()) + laceLen
	case *os.File:
		info, err := v.Stat()
		if err != nil {
			return r.size + laceLen
		}
		return info.Size() + laceLen
	}
	return r.size + laceLen
}

func (r *reader) getReaded() (readed int64) {
	return atomic.LoadInt64(&r.readed) // 原子操作
}

func (r *reader) addReaded(delta int64) (readed int64) {
	return atomic.AddInt64(&r.readed, delta) // 原子操作
}
