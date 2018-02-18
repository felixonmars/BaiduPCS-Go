package uploader

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"github.com/iikira/BaiduPCS-Go/requester/multipartreader"
	"net/http"
	"strings"
)

// DoUpload 执行上传
func DoUpload(uploadURL string, readedlen64 multipartreader.ReadedLen64, o *Options, checkFunc func(resp *http.Response, err error)) {
	u := NewUploader(uploadURL, readedlen64, o)

	exit := make(chan struct{})

	u.OnExecute(func() {
		for {
			select {
			case v, ok := <-u.UploadStatus:
				if !ok {
					return
				}

				fmt.Printf("\r%s/%s %s/s in %s %v",
					pcsutil.ConvertFileSize(v.Uploaded, 2),
					pcsutil.ConvertFileSize(v.Length, 2),
					pcsutil.ConvertFileSize(v.Speed, 2),
					v.TimeElapsed,
					"[UPLOADING]"+strings.Repeat(" ", 10),
				)
			}
		}
	})

	u.OnFinish(func() {
		exit <- struct{}{}
	})

	u.Execute(checkFunc)

	<-exit
	return
}
