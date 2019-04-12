package baidupcs

import (
	"errors"
	"github.com/iikira/BaiduPCS-Go/baidupcs/pcserror"
)

type (
	// ShareOption 分享可选项
	ShareOption struct {
		Password string // 密码
		Period   int    // 有效期
	}

	// Shared 分享信息
	Shared struct {
		Link    string `json:"link"`
		ShareID int64  `json:"shareid"`
	}

	// ShareRecordInfo 分享信息
	ShareRecordInfo struct {
		ShareID         int64   `json:"shareId"`
		FsIds           []int64 `json:"fsIds"`
		Passwd          string  `json:"passwd"`
		Shortlink       string  `json:"shortlink"`
		Status          int     `json:"status"`          // 状态
		TypicalCategory int     `json:"typicalCategory"` // 文件类型
		TypicalPath     string  `json:"typicalPath"`
	}

	sharePSetJSON struct {
		*Shared
		*pcserror.PanErrorInfo
	}

	shareListJSON struct {
		List ShareRecordInfoList `json:"list"`
		*pcserror.PanErrorInfo
	}
)

var (
	// ErrShareLinkNotFound 未找到分享链接
	ErrShareLinkNotFound = errors.New("未找到分享链接")
)

// Clean 清理
func (sri *ShareRecordInfo) Clean() {
	if sri.Passwd == "0" {
		sri.Passwd = ""
	}
}

// HasPasswd 是否需要提取码
func (sri *ShareRecordInfo) HasPasswd() bool {
	return sri.Passwd != "" && sri.Passwd != "0"
}

// ShareRecordInfoList 分享信息列表
type ShareRecordInfoList []*ShareRecordInfo

// Clean 清理
func (sril *ShareRecordInfoList) Clean() {
	for _, sri := range *sril {
		if sri == nil {
			continue
		}

		sri.Clean()
	}
}

// ShareSet 分享文件
func (pcs *BaiduPCS) ShareSet(paths []string, option *ShareOption) (s *Shared, pcsError pcserror.Error) {
	if option == nil {
		option = &ShareOption{}
	}

	dataReadCloser, pcsError := pcs.PrepareSharePSet(paths, option.Period)
	if pcsError != nil {
		return
	}

	defer dataReadCloser.Close()

	errInfo := pcserror.NewPanErrorInfo(OperationShareSet)
	jsonData := sharePSetJSON{
		Shared:       &Shared{},
		PanErrorInfo: errInfo,
	}

	pcsError = pcserror.HandleJSONParse(OperationShareSet, dataReadCloser, &jsonData)
	if pcsError != nil {
		return
	}

	if jsonData.Link == "" {
		errInfo.ErrType = pcserror.ErrTypeOthers
		errInfo.Err = ErrShareLinkNotFound
		return nil, errInfo
	}

	return jsonData.Shared, nil
}

// ShareCancel 取消分享
func (pcs *BaiduPCS) ShareCancel(shareIDs []int64) (pcsError pcserror.Error) {
	dataReadCloser, pcsError := pcs.PrepareShareCancel(shareIDs)
	if pcsError != nil {
		return
	}

	defer dataReadCloser.Close()

	pcsError = pcserror.DecodePanJSONError(OperationShareCancel, dataReadCloser)
	return
}

// ShareList 列出分享列表
func (pcs *BaiduPCS) ShareList(page int) (records ShareRecordInfoList, pcsError pcserror.Error) {
	dataReadCloser, pcsError := pcs.PrepareShareList(page)
	if pcsError != nil {
		return
	}

	defer dataReadCloser.Close()

	errInfo := pcserror.NewPanErrorInfo(OperationShareList)
	jsonData := shareListJSON{
		List:         records,
		PanErrorInfo: errInfo,
	}

	pcsError = pcserror.HandleJSONParse(OperationShareList, dataReadCloser, &jsonData)
	if pcsError != nil {
		return
	}

	if jsonData.List == nil {
		errInfo.ErrType = pcserror.ErrTypeOthers
		errInfo.Err = errors.New("shared list is nil")
		return nil, errInfo
	}

	jsonData.List.Clean()
	return jsonData.List, nil
}
