package downloader

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsutil/converter"
	"io"
	"os"
)

// DoDownload 执行下载
func DoDownload(durl string, savePath string, cfg *Config) {
	var (
		file io.WriterAt
		err  error
	)

	if savePath != "" {
		file, err = os.Create(savePath)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		file = nil
	}

	download := NewDownloader(durl, file, cfg)

	exitDownloadFunc := make(chan struct{})

	download.OnExecute(func() {
		dc := download.GetDownloadStatusChan()
		var ts string

		for {
			select {
			case v, ok := <-dc:
				if !ok { // channel 已经关闭
					return
				}

				if v.TotalSize() <= 0 {
					ts = converter.ConvertFileSize(v.Downloaded(), 2)
				} else {
					ts = converter.ConvertFileSize(v.TotalSize(), 2)
				}

				fmt.Printf("\r ↓ %s/%s %s/s in %s ............",
					converter.ConvertFileSize(v.Downloaded(), 2),
					ts,
					converter.ConvertFileSize(v.SpeedsPerSecond(), 2),
					v.TimeElapsed(),
				)
			}
		}
	})

	err = download.Execute()
	close(exitDownloadFunc)
	if err != nil {
		fmt.Printf("err: %s\n", err)
	}
}
