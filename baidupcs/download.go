package baidupcs

import (
	"github.com/iikira/BaiduPCS-Go/config"
	"net/http"
	"net/http/cookiejar"
)

// FileDownload 下载网盘内文件
func (p PCSApi) FileDownload(path string, downloadFunc func(downloadURL string, jar *cookiejar.Jar, savePath string) error) (err error) {
	// addItem 放在最后
	p.addItem("file", "download", map[string]string{
		"path": path,
	})

	jar, _ := cookiejar.New(nil)
	jar.SetCookies(&p.url, []*http.Cookie{
		&http.Cookie{
			Name:  "BDUSS",
			Value: p.bduss,
		},
	})

	return downloadFunc(p.url.String(), jar, pcsconfig.GetSavePath(path))
}
