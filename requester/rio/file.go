package rio

import (
	"os"
	"sync/atomic"
)

// NewFileReaderLen64 *os.File 实现 ReadedLen64 接口
func NewFileReaderLen64(f *os.File) ReaderLen64 {
	if f == nil {
		return nil
	}

	return &fileReadedlen64{
		f: f,
	}
}

type fileReadedlen64 struct {
	readed int64
	f      *os.File
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
	return info.Size() - fr.readed
}
