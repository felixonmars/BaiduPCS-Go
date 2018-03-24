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
func (pcs *BaiduPCS) QuotaInfo() (quota, used int64, err error) {
	dataReadCloser, err := pcs.PrepareQuotaInfo()
	if err != nil {
		return
	}

	defer dataReadCloser.Close()

	quotaInfo := &quotaInfo{
		ErrInfo: NewErrorInfo(OperationQuotaInfo),
	}

	d := jsoniter.NewDecoder(dataReadCloser)
	err = d.Decode(quotaInfo)
	if err != nil {
		return 0, 0, fmt.Errorf("%s, json 数据解析失败, %s", OperationQuotaInfo, err)
	}

	if quotaInfo.ErrCode != 0 {
		return 0, 0, quotaInfo.ErrInfo
	}

	quota = quotaInfo.Quota
	used = quotaInfo.Used
	return
}
