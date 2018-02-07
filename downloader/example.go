package downloader

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// DoDownload 简单网络下载器, 使用默认下载线程,
// 通过调用 SetMaxThread 来修改默认下载线程
func DoDownload(url string, fileName string) {
	downloader, err := NewDownloader(url, fileName, nil)
	if err != nil {
		return
	}

	done := make(chan struct{})

	var exit = make(chan bool)
	downloader.OnStart(func() {
		fmt.Println("download started")
		format := "\r%v/%v [%s] %v byte/s %v"

		for {
			c := downloader.GetStatusChan()

			select {
			case <-exit:
				return
			case v := <-c:
				var i = float64(v.Downloaded) / float64(downloader.Size) * 50
				h := strings.Repeat("=", int(i)) + strings.Repeat(" ", 50-int(i))
				time.Sleep(time.Second * 1)
				fmt.Printf(format, v.Downloaded, downloader.Size, h, v.Speeds, "[DOWNLOADING]")
				os.Stdout.Sync()
			}
		}
	})

	downloader.OnFinish(func() {
		exit <- true
		done <- struct{}{}
	})

	downloader.OnError(func(errCode int, e error) {
		err = e
	})

	downloader.StartDownload()
	<-done
}
