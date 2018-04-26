package requester

import (
	"crypto/tls"
	"net/http"
	"net/http/cookiejar"
	"time"
)

var (
	// TLSConfig tls连接配置
	TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
	}
)

// HTTPClient http client
type HTTPClient struct {
	http.Client
	jar       *cookiejar.Jar
	transport *http.Transport
	UserAgent string
}

// NewHTTPClient 返回 HTTPClient 的指针,
// 预设了一些配置
func NewHTTPClient() *HTTPClient {
	h := &HTTPClient{
		Client: http.Client{
			Timeout: 30 * time.Second,
		},
		UserAgent: UserAgent,
	}
	return h
}

func (h *HTTPClient) lazyInit() {
	if h.transport == nil {
		h.transport = &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			DialContext:           dialContext,
			Dial:                  dial,
			DialTLS:               dialTLS,
			TLSClientConfig:       TLSConfig,
			TLSHandshakeTimeout:   10 * time.Second,
			DisableKeepAlives:     false,
			DisableCompression:    false, // gzip
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 10 * time.Second,
		}
		h.Client.Transport = h.transport
	}
	if h.jar == nil {
		h.jar, _ = cookiejar.New(nil)
		h.Client.Jar = h.jar
	}
}

// SetUserAgent 设置 UserAgent 浏览器标识
func (h *HTTPClient) SetUserAgent(ua string) {
	h.UserAgent = ua
}

// SetCookiejar 设置 cookie
func (h *HTTPClient) SetCookiejar(c *cookiejar.Jar) {
	if c == nil {
		h.ResetCookiejar()
		return
	}

	h.jar = c
	h.Client.Jar = c
}

// ResetCookiejar 清空 cookie
func (h *HTTPClient) ResetCookiejar() {
	h.jar, _ = cookiejar.New(nil)
	h.Jar = h.jar
}

// SetHTTPSecure 是否启用 https 安全检查, 默认不检查
func (h *HTTPClient) SetHTTPSecure(b bool) {
	h.lazyInit()
	if b {
		TLSConfig.InsecureSkipVerify = b
		h.transport.TLSClientConfig = nil
	} else {
		h.transport.TLSClientConfig = TLSConfig
	}
}

// SetKeepAlive 设置 Keep-Alive
func (h *HTTPClient) SetKeepAlive(b bool) {
	h.lazyInit()
	h.transport.DisableKeepAlives = !b
}

// SetGzip 是否启用Gzip
func (h *HTTPClient) SetGzip(b bool) {
	h.lazyInit()
	h.transport.DisableCompression = !b
}

// SetResponseHeaderTimeout 设置目标服务器响应超时时间
func (h *HTTPClient) SetResponseHeaderTimeout(t time.Duration) {
	h.lazyInit()
	h.transport.ResponseHeaderTimeout = t
}

// SetTimeout 设置 http 请求超时时间, 默认30s
func (h *HTTPClient) SetTimeout(t time.Duration) {
	h.Client.Timeout = t
}
