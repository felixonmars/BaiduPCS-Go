package uploader

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"strings"
	"time"
)

// DoUpload 执行上传
func DoUpload(uploadURL string, uploadReaderLen ReaderLen) {
	u := NewUploader(uploadURL, uploadReaderLen, nil)

	exit := make(chan struct{})
	exit2 := make(chan struct{})

	u.OnExecute(func() {
		t := time.Now()
		c := u.GetStatusChan()
		for {
			select {
			case <-exit:
				return
			case v := <-c:
				fmt.Printf("\r%v/%v %v/s time: %s %v",
					pcsutil.ConvertFileSize(v.Uploaded, 2),
					pcsutil.ConvertFileSize(v.Length, 2),
					pcsutil.ConvertFileSize(v.Speed, 2),
					time.Since(t)/1000000*1000000,
					"[UPLOADING]"+strings.Repeat(" ", 10),
				)
			}
		}
	})

	u.OnFinish(func() {
		exit <- struct{}{}
		exit2 <- struct{}{}
	})

	u.Execute(nil)

	<-exit2
	return
}
