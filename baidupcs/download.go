package baidupcs

import (
	"net/http/cookiejar"
)

// DownloadFunc 下载文件处理函数
type DownloadFunc func(downloadURL string, jar *cookiejar.Jar) error

// DownloadFile 下载单个文件
func (pcs *BaiduPCS) DownloadFile(path string, downloadFunc DownloadFunc) (err error) {
	pcsURL := pcs.generatePCSURL("file", "download", map[string]string{
		"path": path,
	})

	return downloadFunc(pcsURL.String(), pcs.client.Jar.(*cookiejar.Jar))
}

// DownloadStreamFile 下载流式文件
func (pcs *BaiduPCS) DownloadStreamFile(path string, downloadFunc DownloadFunc) (err error) {
	pcsURL := pcs.generatePCSURL("stream", "download", map[string]string{
		"path": path,
	})

	return downloadFunc(pcsURL.String(), pcs.client.Jar.(*cookiejar.Jar))
}
