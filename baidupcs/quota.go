package baidupcs

import (
	"github.com/json-iterator/go"
)

type quotaInfo struct {
	*ErrInfo

	Quota int64 `json:"quota"`
	Used  int64 `json:"used"`
}

// QuotaInfo 获取当前用户空间配额信息
func (pcs *BaiduPCS) QuotaInfo() (quota, used int64, pcsError Error) {
	dataReadCloser, pcsError := pcs.PrepareQuotaInfo()
	if pcsError != nil {
		return
	}

	defer dataReadCloser.Close()

	quotaInfo := &quotaInfo{
		ErrInfo: NewErrorInfo(OperationQuotaInfo),
	}

	d := jsoniter.NewDecoder(dataReadCloser)
	err := d.Decode(quotaInfo)
	if err != nil {
		quotaInfo.ErrInfo.jsonError(err)
		return 0, 0, quotaInfo.ErrInfo
	}

	if quotaInfo.ErrCode != 0 {
		return 0, 0, quotaInfo.ErrInfo
	}

	quota = quotaInfo.Quota
	used = quotaInfo.Used
	return
}
