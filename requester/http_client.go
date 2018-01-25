package requester

import (
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"
	"time"
)

// HTTPClient http client
type HTTPClient struct {
	http.Client
}

// NewHTTPClient 返回 HTTPClient 的指针,
// 预设了一些配置
func NewHTTPClient() *HTTPClient {
	jar, _ := cookiejar.New(nil)
	return &HTTPClient{
		Client: http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
				TLSHandshakeTimeout:   10 * time.Second,
				DisableKeepAlives:     false,
				DisableCompression:    false,
				ResponseHeaderTimeout: 10 * time.Second,
				ExpectContinueTimeout: 10 * time.Second,
			},
			Jar: jar,
		},
	}
}

// SetCookiejar 设置 cookie
func (h *HTTPClient) SetCookiejar(c *cookiejar.Jar) {
	h.Jar = c
}

// ClearCookiejar 清空 cookie
func (h *HTTPClient) ClearCookiejar() {
	h.Jar, _ = cookiejar.New(nil)
}

// SetHTTPSecure 是否启用 https 安全检查
func (h *HTTPClient) SetHTTPSecure(b bool) {
	h.Transport.(*http.Transport).TLSClientConfig.InsecureSkipVerify = !b
}

// SetKeepAlive 设置 Keep-Alive
func (h *HTTPClient) SetKeepAlive(b bool) {
	h.Transport.(*http.Transport).DisableKeepAlives = !b
}

//SetGzip 是否启用Gzip
func (h *HTTPClient) SetGzip(b bool) {
	h.Transport.(*http.Transport).DisableCompression = !b
}

//SetResponseHeaderTimeout 设置目标服务器响应超时时间
func (h *HTTPClient) SetResponseHeaderTimeout(t time.Duration) {
	h.Transport.(*http.Transport).ResponseHeaderTimeout = t
}

// SetTimeout 设置 http 请求超时时间 默认30s
func (h *HTTPClient) SetTimeout(t time.Duration) {
	h.Timeout = t
}
