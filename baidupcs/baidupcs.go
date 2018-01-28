package baidupcs

import (
	"fmt"
	"github.com/iikira/BaiduPCS-Go/pcsconfig"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

// PCSApi 百度 PCS API 详情
type PCSApi struct {
	url   url.URL
	bduss string

	writed bool
}

// NewPCS 提供 百度BDUSS, 返回 PCSApi 指针对象
func NewPCS(bduss string) *PCSApi {
	return &PCSApi{
		url: url.URL{
			Scheme:   "http",
			Host:     "pcs.baidu.com",
			Path:     "/rest/2.0/pcs/",
			RawQuery: fmt.Sprintf("app_id=%d", pcsconfig.Config.AppID),
		},
		bduss:  bduss,
		writed: false,
	}
}

func (p *PCSApi) addItem(subPath, method string, param ...map[string]string) {
	if p.writed {
		panic("addItem: Already writed")
	}
	p.url.Path += subPath
	uv := p.url.Query()
	uv.Set("method", method)
	for k := range param {
		for k2 := range param[k] {
			uv.Set(k2, param[k][k2])
		}
	}
	p.url.RawQuery = uv.Encode()
	p.writed = true
}

func (p *PCSApi) getJar() *cookiejar.Jar {
	jar, _ := cookiejar.New(nil)
	jar.SetCookies(&p.url, []*http.Cookie{
		&http.Cookie{
			Name:  "BDUSS",
			Value: p.bduss,
		},
	})
	return jar
}
