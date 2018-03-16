package downloader

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
)

// DoDownload 执行下载
func DoDownload(url string, cfg Config) {
	download, err := NewDownloader(url, cfg)
	if err != nil {
		fmt.Println(err)
		return
	}

	exitDownloadFunc := make(chan struct{})

	download.OnExecute = func() {
		dc := download.GetStatusChan()
		var ts string

		for {
			select {
			case v, ok := <-dc:
				if !ok { // channel 已经关闭
					return
				}

				if v.TotalSize <= 0 {
					ts = pcsutil.ConvertFileSize(v.Downloaded, 2)
				} else {
					ts = pcsutil.ConvertFileSize(v.TotalSize, 2)
				}

				fmt.Printf("\r↓ %s/%s %s/s in %s ............",
					pcsutil.ConvertFileSize(v.Downloaded, 2),
					ts,
					pcsutil.ConvertFileSize(v.Speeds, 2),
					v.TimeElapsed,
				)
			}
		}
	}

	download.OnFinish = func() {
		exitDownloadFunc <- struct{}{}
	}

	download.Execute()
	<-exitDownloadFunc

	close(exitDownloadFunc)
}
