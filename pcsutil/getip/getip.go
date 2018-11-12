// Package getip 获取 ip 信息包
package getip

import (
	"github.com/iikira/BaiduPCS-Go/requester"
	"unsafe"
)

// IPInfoByClient 给定client获取ip地址
func IPInfoByClient(c *requester.HTTPClient) (ipAddr string, err error) {
	if c == nil {
		c = requester.NewHTTPClient()
	}

	body, err := c.Fetch("GET", "https://api.ipify.org", nil, nil)
	if err != nil {
		return "", err
	}

	return *(*string)(unsafe.Pointer(&body)), nil
}

//IPInfo 获取IP地址和IP位置
func IPInfo(https bool) (ipAddr string, err error) {
	c := requester.NewHTTPClient()
	c.SetHTTPSecure(https)
	return IPInfoByClient(c)
}
