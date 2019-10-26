package downloader

import (
	"github.com/iikira/BaiduPCS-Go/requester/downloader/prealloc"
	"io"
	"os"
)

type (
	Fder interface {
		Fd() uintptr
	}

	Writer interface {
		io.WriterAt
	}
)

func NewDownloaderWriterByFilename(name string, flag int, perm os.FileMode) (writer Writer, file *os.File, warn error, err error) {
	warn = prealloc.InitPrivilege()
	file, err = os.OpenFile(name, flag, perm)
	if err != nil {
		return
	}

	writer = file
	return
}
