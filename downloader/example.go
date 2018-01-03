package downloader

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/requester"
	"os"
	"strings"
	"time"
)

// DoDownload 简单网络下载器, 使用默认下载线程,
// 通过调用 SetMaxThread 来修改默认下载线程
func DoDownload(url string, fileName string) {
	h := requester.NewHTTPClient()
	fileDl, err := NewFileDl(h, url, fileName)
	if err != nil {
		return
	}

	done := make(chan struct{})

	var exit = make(chan bool)
	fileDl.OnStart(func() {
		fmt.Println("download started")
		format := "\r%v/%v [%s] %v byte/s %v"

	for_1:
		for {
			status := fileDl.GetStatus()
			var i = float64(status.Downloaded) / float64(fileDl.Size) * 50
			h := strings.Repeat("=", int(i)) + strings.Repeat(" ", 50-int(i))

			select {
			case <-exit:
				fmt.Printf(format, status.Downloaded, fileDl.Size, h, 0, "[FINISH]")
				fmt.Println("\ndownload finished")
				break for_1
			default:
				time.Sleep(time.Second * 1)
				fmt.Printf(format, status.Downloaded, fileDl.Size, h, status.Speeds, "[DOWNLOADING]")
				os.Stdout.Sync()
			}
		}
	})

	fileDl.OnFinish(func() {
		exit <- true
		done <- struct{}{}
	})

	fileDl.OnError(func(errCode int, e error) {
		err = e
	})

	fileDl.Start()
	<-done
}
