package downloader

import (
	"io"
)

// Writer 接口
type Writer interface {
	io.WriteCloser
	io.WriterAt
}
