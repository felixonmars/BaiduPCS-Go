package downloader

import (
	"fmt"
	"strings"
	"time"
)

// DoDownload 简单网络下载器, 使用默认下载线程,
// 通过调用 SetMaxThread 来修改默认下载线程
func DoDownload(url string, savePath string) {
	downloader, err := NewDownloader(url, savePath, nil)
	if err != nil {
		return
	}

	done := make(chan struct{})

	downloader.OnStart(func() {
		fmt.Println("download started")
		format := "\r%v/%v [%s] %v byte/s %v"

		c := downloader.GetStatusChan()
		for {
			select {
			case v, ok := <-c:
				if !ok {
					return
				}

				var i = float64(v.Downloaded) / float64(downloader.size) * 50
				h := strings.Repeat("=", int(i)) + strings.Repeat(" ", 50-int(i))
				time.Sleep(time.Second * 1)
				fmt.Printf(format, v.Downloaded, downloader.size, h, v.Speeds, "[DOWNLOADING]")
			}
		}
	})

	downloader.OnFinish(func() {
		done <- struct{}{}
	})

	downloader.OnError(func(errCode int, e error) {
		err = e
	})

	downloader.StartDownload()
	<-done
}
