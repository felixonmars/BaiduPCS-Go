package baidupcs

import (
	"github.com/iikira/BaiduPCS-Go/pcsconfig"
	"net/http/cookiejar"
)

// DownloadFunc 下载文件处理函数
type DownloadFunc func(downloadURL string, jar *cookiejar.Jar, savePath string) error

// DownloadFile 下载单个文件
func (pcs *BaiduPCS) DownloadFile(path string, downloadFunc DownloadFunc) (err error) {
	pcs.setPCSURL("file", "download", map[string]string{
		"path": path,
	})

	return downloadFunc(pcs.url.String(), pcs.client.Jar.(*cookiejar.Jar), pcsconfig.GetSavePath(path))
}

// DownloadStreamFile 下载流式文件
func (pcs *BaiduPCS) DownloadStreamFile(path string, downloadFunc DownloadFunc) (err error) {
	pcs.setPCSURL("stream", "download", map[string]string{
		"path": path,
	})

	return downloadFunc(pcs.url.String(), pcs.client.Jar.(*cookiejar.Jar), pcsconfig.GetSavePath(path))
}
