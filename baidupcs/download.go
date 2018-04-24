package baidupcs

import (
	"net/http/cookiejar"
)

// DownloadFunc 下载文件处理函数
type DownloadFunc func(downloadURL string, jar *cookiejar.Jar) error

// DownloadFile 下载单个文件
func (pcs *BaiduPCS) DownloadFile(path string, downloadFunc DownloadFunc) (err error) {
	pcs.lazyInit()
	pcsURL := pcs.generatePCSURL("file", "download", map[string]string{
		"path": path,
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationDownloadFile, pcsURL)

	return downloadFunc(pcsURL.String(), pcs.client.Jar.(*cookiejar.Jar))
}

// DownloadStreamFile 下载流式文件
func (pcs *BaiduPCS) DownloadStreamFile(path string, downloadFunc DownloadFunc) (err error) {
	pcs.lazyInit()
	pcsURL := pcs.generatePCSURL("stream", "download", map[string]string{
		"path": path,
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationDownloadStreamFile, pcsURL)

	return downloadFunc(pcsURL.String(), pcs.client.Jar.(*cookiejar.Jar))
}
