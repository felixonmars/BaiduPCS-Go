// Package getip 获取 ip 信息包
package getip

import (
	"github.com/iikira/BaiduPCS-Go/requester"
	"unsafe"
)

//IPInfo 获取IP地址和IP位置
func IPInfo(https bool) (ipAddr string, err error) {
	c := requester.NewHTTPClient()
	c.SetHTTPSecure(https)

	var scheme string
	if https {
		scheme = "https"
	} else {
		scheme = "http"
	}

	body, err := c.Fetch("GET", scheme+"://api.ipify.org", nil, nil)
	if err != nil {
		return "", err
	}

	return *(*string)(unsafe.Pointer(&body)), nil
}
