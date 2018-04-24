// Package baidupcs BaiduPCS RESTful API 工具包
package baidupcs

import (
	"github.com/iikira/BaiduPCS-Go/pcsverbose"
	"github.com/iikira/BaiduPCS-Go/requester"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
)

const (
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

func (pcs *BaiduPCS) generatePCSURL(subPath, method string, param ...map[string]string) *url.URL {
	pcsURL := &url.URL{
		Scheme: "http",
		Host:   "pcs.baidu.com",
		Path:   "/rest/2.0/pcs/" + subPath,
	}

	if pcs.isHTTPS {
		pcsURL.Scheme = "https"
	}

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
		Scheme: "http",
		Host:   "pan.baidu.com",
		Path:   "/rest/2.0/" + subPath,
	}

	if pcs.isHTTPS {
		pcsURL2.Scheme = "https"
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
