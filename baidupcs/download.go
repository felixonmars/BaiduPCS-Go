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

	// LocateDownloadInfoV1 locatedownload api v1
	LocateDownloadInfoV1 struct {
		Server []string `json:"server"`
		PathJSON
	}

	locateDownloadJSON struct {
		*pcserror.PCSErrInfo
		URLInfo
	}

	// APIDownloadDlinkInfo 下载信息
	APIDownloadDlinkInfo struct {
		Dlink string `json:"dlink"`
		FsID  string `json:"fs_id"`
	}

	// APIDownloadDlinkInfoList 下载信息列表
	APIDownloadDlinkInfoList []*APIDownloadDlinkInfo

	panAPIDownloadJSON struct {
		*pcserror.PanErrorInfo
		DlinkList APIDownloadDlinkInfoList `json:"dlink"`
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
	if len(ui.URLs) < 1 {
		return nil
	}

	u, err := url.Parse(ui.URLs[0].URL)
	if err != nil {
		return nil
	}
	u.Scheme = GetHTTPScheme(https)
	return u
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

// LocateDownloadWithUserAgent 获取下载链接, 可指定User-Agent
func (pcs *BaiduPCS) LocateDownloadWithUserAgent(pcspath, ua string) (info *URLInfo, pcsError pcserror.Error) {
	dataReadCloser, pcsError := pcs.prepareLocateDownload(pcspath, ua)
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

// LocateDownload 获取下载链接
func (pcs *BaiduPCS) LocateDownload(pcspath string) (info *URLInfo, pcsError pcserror.Error) {
	return pcs.LocateDownloadWithUserAgent(pcspath, "")
}

// LocatePanAPIDownload 从百度网盘首页获取下载链接
func (pcs *BaiduPCS) LocatePanAPIDownload(fidList ...int64) (dlinkInfoList APIDownloadDlinkInfoList, pcsError pcserror.Error) {
	dataReadCloser, pcsError := pcs.PrepareLocatePanAPIDownload(fidList...)
	if dataReadCloser != nil {
		defer dataReadCloser.Close()
	}
	if pcsError != nil {
		return nil, pcsError
	}

	jsonData := panAPIDownloadJSON{
		PanErrorInfo: pcserror.NewPanErrorInfo(OperationLocatePanAPIDownload),
	}
	pcsError = handleJSONParse(OperationLocatePanAPIDownload, dataReadCloser, &jsonData)
	if pcsError != nil {
		if pcsError.GetErrType() == pcserror.ErrTypeRemoteError {
			switch pcsError.GetRemoteErrCode() {
			case 112: // 页面已过期
				pcs.ph.SetSignExpires() // 重置
			}
		}
		return
	}

	return jsonData.DlinkList, nil
}
