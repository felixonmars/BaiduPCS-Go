package baidupcs

import (
	"fmt"
	"github.com/bitly/go-simplejson"
	"github.com/iikira/BaiduPCS-Go/requester"
)

// QuotaInfo 获取当前用户空间配额信息
func (p PCSApi) QuotaInfo() (quota, used int64, err error) {
	p.addItem("quota", "info")

	h := requester.NewHTTPClient()
	body, err := h.Fetch("GET", p.url.String(), nil, map[string]string{
		"Cookie": "BDUSS=" + p.bduss,
	})
	if err != nil {
		return
	}

	json, err := simplejson.NewJson(body)
	if err != nil {
		return
	}

	code, err := CheckErr(json)
	if err != nil {
		return 0, 0, fmt.Errorf("获取当前用户空间配额信息, 错误代码: %d, 消息: %s", code, err)
	}

	quota = json.Get("quota").MustInt64()
	used = json.Get("used").MustInt64()

	return
}
