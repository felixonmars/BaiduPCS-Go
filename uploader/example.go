package uploader

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"net/http"
	"strings"
)

// DoUpload 执行上传
func DoUpload(uploadURL string, uploadReaderLen ReaderLen, o *Options, checkFunc func(resp *http.Response, err error)) {
	u := NewUploader(uploadURL, uploadReaderLen, o)

	exit := make(chan struct{})
	exit2 := make(chan struct{})

	u.OnExecute(func() {
		for {
			select {
			case <-exit:
				return
			case v, ok := <-u.UploadStatus:
				if !ok {
					return
				}

				fmt.Printf("\r%v/%v %v/s time: %s %v",
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
		exit2 <- struct{}{}
	})

	u.Execute(checkFunc)

	<-exit2
	return
}
