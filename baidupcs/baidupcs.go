// Package baidupcs BaiduPCS RESTful API 工具包
package baidupcs

import (
	"errors"
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"github.com/iikira/BaiduPCS-Go/requester"
	"github.com/iikira/baidu-tools/pan"
	"github.com/json-iterator/go"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
)

const (
	// OperationGetUK 获取UK
	OperationGetUK = "获取UK"
	// OperationQuotaInfo 获取当前用户空间配额信息
	OperationQuotaInfo = "获取当前用户空间配额信息"
	// OperationFilesDirectoriesMeta 获取文件/目录的元信息
	OperationFilesDirectoriesMeta = "获取文件/目录的元信息"
	// OperationFilesDirectoriesList 获取目录下的文件列表
	OperationFilesDirectoriesList = "获取目录下的文件列表"
	// OperationRemove 删除文件/目录
	OperationRemove = "删除文件/目录"
	// OperationMkdir 创建目录
	OperationMkdir = "创建目录"
	// OperationRename 重命名文件/目录
	OperationRename = "重命名文件/目录"
	// OperationCopy 拷贝文件/目录
	OperationCopy = "拷贝文件/目录"
	// OperationMove 移动文件/目录
	OperationMove = "移动文件/目录"
	// OperationRapidUpload 秒传文件
	OperationRapidUpload = "秒传文件"
	// OperationUpload 上传单个文件
	OperationUpload = "上传单个文件"
	// OperationUploadTmpFile 分片上传—文件分片及上传
	OperationUploadTmpFile = "分片上传—文件分片及上传"
	// OperationUploadCreateSuperFile 分片上传—合并分片文件
	OperationUploadCreateSuperFile = "分片上传—合并分片文件"
	// OperationDownloadFile 下载单个文件
	OperationDownloadFile = "下载单个文件"
	// OperationDownloadStreamFile 下载流式文件
	OperationDownloadStreamFile = "下载流式文件"
	// OperationLocateDownload 提取下载链接
	OperationLocateDownload = "提取下载链接"
	// OperationCloudDlAddTask 添加离线下载任务
	OperationCloudDlAddTask = "添加离线下载任务"
	// OperationCloudDlQueryTask 精确查询离线下载任务
	OperationCloudDlQueryTask = "精确查询离线下载任务"
	// OperationCloudDlListTask 查询离线下载任务列表
	OperationCloudDlListTask = "查询离线下载任务列表"
	// OperationCloudDlCancelTask 取消离线下载任务
	OperationCloudDlCancelTask = "取消离线下载任务"
	// OperationCloudDlDeleteTask 删除离线下载任务
	OperationCloudDlDeleteTask = "删除离线下载任务"
	// OperationShareSet 创建分享链接
	OperationShareSet = "创建分享链接"
	// OperationShareCancel 取消分享
	OperationShareCancel = "取消分享"
	// OperationShareList 列出分享列表
	OperationShareList = "列出分享列表"
)

var (
	baiduPCSVerbose = pcsverbose.New("BAIDUPCS")
)

// BaiduPCS 百度 PCS API 详情
type BaiduPCS struct {
	appID   int                   // app_id
	isHTTPS bool                  // 是否启用https
	client  *requester.HTTPClient // http 客户端
}

// NewPCS 提供app_id, 百度BDUSS, 返回 BaiduPCS 对象
func NewPCS(appID int, bduss string) *BaiduPCS {
	client := requester.NewHTTPClient()

	pcsURL := &url.URL{
		Scheme: "http",
		Host:   "pcs.baidu.com",
	}

	cookies := []*http.Cookie{
		&http.Cookie{
			Name:  "BDUSS",
			Value: bduss,
		},
	}

	jar, _ := cookiejar.New(nil)
	jar.SetCookies(pcsURL, cookies)
	jar.SetCookies((&url.URL{
		Scheme: "http",
		Host:   "pan.baidu.com",
	}), cookies)
	client.SetCookiejar(jar)

	return &BaiduPCS{
		appID:  appID,
		client: client,
	}
}

// NewPCSWithClient 提供app_id, 自定义客户端, 返回 BaiduPCS 对象
func NewPCSWithClient(appID int, client *requester.HTTPClient) *BaiduPCS {
	pcs := &BaiduPCS{
		appID:  appID,
		client: client,
	}
	return pcs
}

// NewPCSWithCookieStr 提供app_id, cookie 字符串, 返回 BaiduPCS 对象
func NewPCSWithCookieStr(appID int, cookieStr string) *BaiduPCS {
	pcs := &BaiduPCS{
		appID:  appID,
		client: requester.NewHTTPClient(),
	}

	cookies := requester.ParseCookieStr(cookieStr)
	jar, _ := cookiejar.New(nil)
	jar.SetCookies(pcs.URL(), cookies)
	pcs.client.SetCookiejar(jar)

	return pcs
}

func (pcs *BaiduPCS) lazyInit() {
	if pcs.client == nil {
		pcs.client = requester.NewHTTPClient()
	}
}

// SetAPPID 设置app_id
func (pcs *BaiduPCS) SetAPPID(appID int) {
	pcs.appID = appID
}

// SetUserAgent 设置 User-Agent
func (pcs *BaiduPCS) SetUserAgent(ua string) {
	pcs.client.SetUserAgent(ua)
}

// SetHTTPS 是否启用https连接
func (pcs *BaiduPCS) SetHTTPS(https bool) {
	pcs.isHTTPS = https
}

// URL 返回 url
func (pcs *BaiduPCS) URL() *url.URL {
	return &url.URL{
		Scheme: GetHTTPScheme(pcs.isHTTPS),
		Host:   "pcs.baidu.com",
	}
}

func (pcs *BaiduPCS) generatePCSURL(subPath, method string, param ...map[string]string) *url.URL {
	pcsURL := pcs.URL()
	pcsURL.Path = "/rest/2.0/pcs/" + subPath

	uv := pcsURL.Query()
	uv.Set("app_id", strconv.Itoa(pcs.appID))
	uv.Set("method", method)
	for k := range param {
		for k2 := range param[k] {
			uv.Set(k2, param[k][k2])
		}
	}

	pcsURL.RawQuery = uv.Encode()
	return pcsURL
}

func (pcs *BaiduPCS) generatePCSURL2(subPath, method string, param ...map[string]string) *url.URL {
	pcsURL2 := &url.URL{
		Scheme: GetHTTPScheme(pcs.isHTTPS),
		Host:   "pan.baidu.com",
		Path:   "/rest/2.0/" + subPath,
	}

	uv := pcsURL2.Query()
	uv.Set("app_id", "250528")
	uv.Set("method", method)
	for k := range param {
		for k2 := range param[k] {
			uv.Set(k2, param[k][k2])
		}
	}

	pcsURL2.RawQuery = uv.Encode()
	return pcsURL2
}

// UK 获取用户 UK
func (pcs *BaiduPCS) UK() (uk int64, pcsError Error) {
	pcs.lazyInit()

	pcsURL := GetHTTPScheme(pcs.isHTTPS) + "://pan.baidu.com/api/user/getinfo?need_selfinfo=1"

	errInfo := NewErrorInfo(OperationGetUK)
	body, err := pcs.client.Fetch("GET", pcsURL, nil, map[string]string{
		"User-Agent": "netdisk;8.3.1",
	})
	if err != nil {
		errInfo.errType = ErrTypeNetError
		errInfo.err = err
		return 0, errInfo
	}

	jsonData := struct {
		pan.RemoteErrInfo
		Records []struct {
			Uk int64 `json:"uk"`
		} `json:"records"`
	}{}

	err = jsoniter.Unmarshal(body, &jsonData)
	if err != nil {
		errInfo.jsonError(err)
		return 0, errInfo
	}

	if jsonData.ErrNo != 0 {
		jsonData.RemoteErrInfo.ParseErrMsg()
		errInfo.ErrCode = jsonData.RemoteErrInfo.ErrNo
		errInfo.ErrMsg = jsonData.RemoteErrInfo.ErrMsg
		return 0, errInfo
	}

	if len(jsonData.Records) != 1 {
		errInfo.errType = ErrTypeOthers
		errInfo.err = errors.New("Unknown remote data")
		return 0, errInfo
	}

	return jsonData.Records[0].Uk, nil
}
