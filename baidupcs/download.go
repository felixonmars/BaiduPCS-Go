package baidupcs

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/config"
	"github.com/iikira/BaiduPCS-Go/downloader"
	"github.com/iikira/BaiduPCS-Go/util"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var saveDir = "download"

// FileDownload 下载网盘内文件
func (p PCSApi) FileDownload(path string, size int64) (err error) {
	// addItem 放在最后
	p.addItem("file", "download", map[string]string{
		"path": path,
	})

	h := downloader.NewHTTPClient()
	jar, _ := cookiejar.New(nil)
	jar.SetCookies(&p.url, []*http.Cookie{
		&http.Cookie{
			Name:  "BDUSS",
			Value: p.bduss,
		},
	})
	h.SetCookiejar(jar)
	h.SetKeepAlive(true)
	h.SetTimeout(2 * time.Minute)

	savePath := saveDir + filepath.Dir("/"+pcsconfig.ActiveBaiduUser.Name+path+"/..")
	fileDl, err := downloader.NewFileDl(h, p.url.String(), savePath, size)
	if err != nil {
		return err
	}

	pa := make(chan struct{})

	var exit = make(chan bool)

	fileDl.OnStart(func() {
		t1 := time.Now()
	for_1:
		for {
			status := fileDl.GetStatus()

			select {
			case <-exit:
				break for_1
			default:
				time.Sleep(time.Second * 1)
				fmt.Printf("\r%v/%v %v/s time: %s %v",
					pcsutil.ConvertFileSize(status.Downloaded, 2),
					pcsutil.ConvertFileSize(fileDl.Size, 2),
					pcsutil.ConvertFileSize(status.Speeds, 2),
					time.Since(t1)/1000000*1000000,
					"[DOWNLOADING]"+strings.Repeat(" ", 10),
				)
				os.Stdout.Sync()
			}
		}
	})

	fileDl.OnFinish(func() {
		exit <- true
		pa <- struct{}{}
	})

	fileDl.Start()
	<-pa
	fmt.Printf("\n\n下载完成, 保存位置: %s\n\n", savePath)
	return nil
}
