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
		file      *os.File
		writer    io.WriterAt
		warn, err error
	)

	if savePath != "" {
		writer, file, warn, err = NewDownloaderWriterByFilename(savePath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		if warn != nil {
			fmt.Printf("warn: %s\n", warn)
		}
		if err != nil {
			fmt.Println(err)
			return
		}
		defer file.Close()
	}

	download := NewDownloader(durl, writer, cfg)

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
