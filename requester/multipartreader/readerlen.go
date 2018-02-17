package multipartreader

import (
	"io"
	"os"
	"sync/atomic"
)

// ReaderLen 实现io.Reader和32-bit长度接口
type ReaderLen interface {
	io.Reader
	Len() int
}

// ReaderLen64 实现io.Reader和64-bit长度接口
type ReaderLen64 interface {
	io.Reader
	Len() int64
}

// ReadedLen64 实现io.Reader, 64-bit长度接口和已读取数据量接口
type ReadedLen64 interface {
	ReaderLen64
	Readed() int64
}

// NewFileReadedLen64 *os.File 实现 ReadedLen64 接口
func NewFileReadedLen64(f *os.File) ReadedLen64 {
	if f == nil {
		return nil
	}

	return &fileReadedlen64{
		f:      f,
		readed: 0,
	}
}

type fileReadedlen64 struct {
	f      *os.File
	readed int64
}

// Read 读文件, 并记录已读取数据量
func (fr *fileReadedlen64) Read(b []byte) (n int, err error) {
	n, err = fr.f.Read(b)
	atomic.AddInt64(&fr.readed, int64(n))
	return n, err
}

// Len 返回文件的大小
func (fr *fileReadedlen64) Len() int64 {
	info, err := fr.f.Stat()
	if err != nil {
		return 0
	}
	return info.Size()
}

func (fr *fileReadedlen64) Readed() int64 {
	return atomic.LoadInt64(&fr.readed)
}
