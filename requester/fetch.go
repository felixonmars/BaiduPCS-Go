package requester

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// HTTPGet 简单实现 http 访问 GET 请求
func HTTPGet(urlStr string) (body []byte, err error) {
	resp, err := http.Get(urlStr)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

// Req 实现 http／https 访问，
// 根据给定的 method (GET, POST, HEAD, PUT 等等), urlStr (网址),
// post (post 数据), header (header 请求头数据), 进行网站访问。
// 返回值分别为 *http.Response, 错误信息
func (h *HTTPClient) Req(method string, urlStr string, post interface{}, header map[string]string) (resp *http.Response, err error) {
	var (
		req   *http.Request
		obody io.Reader
	)

	if post != nil {
		switch value := post.(type) {
		case map[string]string:
			query := url.Values{}
			for k := range value {
				query.Set(k, value[k])
			}
			obody = strings.NewReader(query.Encode())
		case string:
			obody = strings.NewReader(value)
		case []byte:
			obody = bytes.NewReader(value[:])
		default:
			return nil, fmt.Errorf("requester.Req: unknown post type: %s", value)
		}
	}
	req, err = http.NewRequest(method, urlStr, obody)
	if err != nil {
		return nil, err
	}

	// 设置浏览器标识
	req.Header.Set("User-Agent", UserAgent)

	if header != nil {
		for key := range header {
			req.Header.Add(key, header[key])
		}
	}

	return h.Client.Do(req)
}

// Fetch 实现 http／https 访问，
// 根据给定的 method (GET, POST, HEAD, PUT 等等), urlStr (网址),
// post (post 数据), header (header 请求头数据), 进行网站访问。
// 返回值分别为 网站主体, 错误信息
func (h *HTTPClient) Fetch(method string, urlStr string, post interface{}, header map[string]string) (body []byte, err error) {
	resp, err := h.Req(method, urlStr, post, header)
	if err != nil {
		return nil, err
	}

	body, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return
}
