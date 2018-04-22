package downloader

import (
	"github.com/iikira/BaiduPCS-Go/requester/rio/speeds"
	"io"
)

//trigger 用于触发事件
func trigger(f func()) {
	if f == nil {
		return
	}
	go f()
}

func fixCacheSize(size *int) {
	if *size < 1024 {
		*size = 1024
	}
}

func readFullFrom(r io.Reader, buf []byte, sps ...speeds.Adder) (n int, err error) {
	var nn int
	for n < len(buf) && err == nil {
		nn, err = r.Read(buf[n:])

		// 更新速度统计
		for _, sp := range sps {
			if sp == nil {
				continue
			}
			sp.Add(int64(nn))
		}
		n += nn
	}
	if n >= len(buf) {
		err = nil
	} else if n > 0 && err == io.EOF {
		err = io.ErrUnexpectedEOF
	}
	return
}
