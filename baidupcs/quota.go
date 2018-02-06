package baidupcs

import (
	"fmt"
	"github.com/bitly/go-simplejson"
)

// QuotaInfo 获取当前用户空间配额信息
func (p *PCSApi) QuotaInfo() (quota, used int64, err error) {
	p.setApi("quota", "info")

	resp, err := p.client.Req("GET", p.url.String(), nil, nil)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	json, err := simplejson.NewFromReader(resp.Body)
	if err != nil {
		return
	}

	code, msg := CheckErr(json)
	if msg != "" {
		return 0, 0, fmt.Errorf("获取当前用户空间配额信息, 错误代码: %d, 消息: %s", code, msg)
	}

	quota = json.Get("quota").MustInt64()
	used = json.Get("used").MustInt64()
	return
}
