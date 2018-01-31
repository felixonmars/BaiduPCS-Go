package uploader

import (
	"os"
)

// NewFileReaderLen 将 *os.File 文件对象实现 ReaderLen 接口
func NewFileReaderLen(f *os.File) ReaderLen {
	return fileReaderlen{f}
}

type fileReaderlen struct {
	*os.File
}

// Len 返回文件的大小
func (fr fileReaderlen) Len() int64 {
	info, err := fr.Stat()
	if err != nil {
		return 0
	}
	return info.Size()
}
