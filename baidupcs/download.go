package baidupcs

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs/pcserror"
	"net/http"
	"net/url"
)

type (
	// DownloadFunc 下载文件处理函数
	DownloadFunc func(downloadURL string, jar http.CookieJar) error

	// URLInfo 下载链接详情
	URLInfo struct {
		URLs []struct {
			URL string `json:"url"`
		} `json:"urls"`
	}

	locateDownloadJSON struct {
		*pcserror.PCSErrInfo
		URLInfo
	}
)

// URLStrings 返回下载链接数组
func (ui *URLInfo) URLStrings(https bool) (urls []*url.URL) {
	urls = make([]*url.URL, 0, len(ui.URLs))
	for k := range ui.URLs {
		thisURL, err := url.Parse(ui.URLs[k].URL)
		if err != nil {
			continue
		}
		thisURL.Scheme = GetHTTPScheme(https)
		urls = append(urls, thisURL)
	}
	return urls
}

// SingleURL 返回单条下载链接
func (ui *URLInfo) SingleURL(https bool) *url.URL {
	urls := ui.URLStrings(https)
	if len(urls) < 1 {
		return nil
	}

	return urls[0]
}

// DownloadFile 下载单个文件
func (pcs *BaiduPCS) DownloadFile(path string, downloadFunc DownloadFunc) (err error) {
	pcs.lazyInit()
	pcsURL := pcs.generatePCSURL("file", "download", map[string]string{
		"path": path,
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationDownloadFile, pcsURL)

	return downloadFunc(pcsURL.String(), pcs.client.Jar)
}

// DownloadStreamFile 下载流式文件
func (pcs *BaiduPCS) DownloadStreamFile(path string, downloadFunc DownloadFunc) (err error) {
	pcs.lazyInit()
	pcsURL := pcs.generatePCSURL("stream", "download", map[string]string{
		"path": path,
	})
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationDownloadStreamFile, pcsURL)

	return downloadFunc(pcsURL.String(), pcs.client.Jar)
}

// LocateDownload 提取下载链接
func (pcs *BaiduPCS) LocateDownload(pcspath string) (info *URLInfo, pcsError pcserror.Error) {
	dataReadCloser, pcsError := pcs.PrepareLocateDownload(pcspath)
	if dataReadCloser != nil {
		defer dataReadCloser.Close()
	}
	if pcsError != nil {
		return nil, pcsError
	}

	errInfo := pcserror.NewPCSErrorInfo(OperationLocateDownload)
	jsonData := locateDownloadJSON{
		PCSErrInfo: errInfo,
	}

	pcsError = handleJSONParse(OperationLocateDownload, dataReadCloser, &jsonData)
	if pcsError != nil {
		return
	}

	return &jsonData.URLInfo, nil
}
