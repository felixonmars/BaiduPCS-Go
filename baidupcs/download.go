package baidupcs

import (
	"github.com/iikira/BaiduPCS-Go/pcsconfig"
	"net/http/cookiejar"
)

// FileDownload 下载网盘内文件
func (p PCSApi) FileDownload(path string, downloadFunc func(downloadURL string, jar *cookiejar.Jar, savePath string) error) (err error) {
	// addItem 放在最后
	p.addItem("file", "download", map[string]string{
		"path": path,
	})

	return downloadFunc(p.url.String(), p.getJar(), pcsconfig.GetSavePath(path))
}
