package baidupcs

import (
	"fmt"
	"github.com/json-iterator/go"
)

type quotaInfo struct {
	*ErrInfo

	Quota int64 `json:"quota"`
	Used  int64 `json:"used"`
}

// QuotaInfo 获取当前用户空间配额信息
func (p *PCSApi) QuotaInfo() (quota, used int64, err error) {
	operation := "获取当前用户空间配额信息"

	p.setAPI("quota", "info")

	resp, err := p.client.Req("GET", p.url.String(), nil, nil)
	if err != nil {
		return
	}

	defer resp.Body.Close()

	quotaInfo := &quotaInfo{
		ErrInfo: NewErrorInfo(operation),
	}

	d := jsoniter.NewDecoder(resp.Body)
	err = d.Decode(quotaInfo)
	if err != nil {
		return 0, 0, fmt.Errorf("%s, json 数据解析失败, %s", operation, err)
	}

	if quotaInfo.ErrCode != 0 {
		return 0, 0, quotaInfo.ErrInfo
	}

	quota = quotaInfo.Quota
	used = quotaInfo.Used
	return
}
