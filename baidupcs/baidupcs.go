package baidupcs

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsconfig"
	"github.com/iikira/BaiduPCS-Go/requester"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

const (
	operationFilesDirectoriesBatchMeta = "获取文件/目录的元信息"
	operationFilesDirectoriesList      = "获取目录下的文件列表"
)

var (
	// AppID 百度 PCS 应用 ID
	AppID int
)

// PCSApi 百度 PCS API 详情
type PCSApi struct {
	url    *url.URL
	writed bool                  // 是否已写入
	client *requester.HTTPClient // http 客户端
}

// NewPCS 提供 百度BDUSS, 返回 PCSApi 指针对象
func NewPCS(bduss string) *PCSApi {
	client := requester.NewHTTPClient()
	client.UserAgent = pcsconfig.Config.UserAgent

	pcsURL := &url.URL{
		Scheme: "http",
		Host:   "pcs.baidu.com",
	}

	jar, _ := cookiejar.New(nil)
	jar.SetCookies(pcsURL, []*http.Cookie{
		&http.Cookie{
			Name:  "BDUSS",
			Value: bduss,
		},
	})
	client.SetCookiejar(jar)

	return &PCSApi{
		url:    pcsURL,
		client: client,
	}
}

func (p *PCSApi) setAPI(subPath, method string, param ...map[string]string) {
	p.url = &url.URL{
		Scheme: "http",
		Host:   "pcs.baidu.com",
		Path:   "/rest/2.0/pcs/" + subPath,
	}

	uv := p.url.Query()
	uv.Set("app_id", fmt.Sprint(pcsconfig.Config.AppID))
	uv.Set("method", method)
	for k := range param {
		for k2 := range param[k] {
			uv.Set(k2, param[k][k2])
		}
	}

	p.url.RawQuery = uv.Encode()
	p.writed = true
}
