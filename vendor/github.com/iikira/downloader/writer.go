package downloader

import (
	"io"
)

// Writer 接口
type Writer interface {
	io.WriteCloser
	WriteAt(b []byte, off int64) (n int, err error)
}
