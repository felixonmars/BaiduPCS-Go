package baidupcs

import (
	"errors"
	"github.com/iikira/baidu-tools/pan"
	"github.com/json-iterator/go"
	"net/url"
	"strconv"
	"strings"
)

// ShareOption 分享可选项
type ShareOption struct {
	Password string // 密码
	Period   int    // 有效期
}

// Shared 分享信息
type Shared struct {
	Link    string `json:"link"`
	ShareID int64  `json:"shareid"`
}

// ShareSet 分享文件
func (pcs *BaiduPCS) ShareSet(paths []string, option *ShareOption) (s *Shared, pcsError Error) {
	pcs.lazyInit()

	if option == nil {
		option = &ShareOption{}
	}

	pcsURL := &url.URL{
		Scheme: GetHTTPScheme(pcs.isHTTPS),
		Host:   "pan.baidu.com",
		Path:   "share/pset",
	}

	builder := &strings.Builder{}
	builder.WriteRune('[')
	for k := range paths {
		builder.WriteString("\"" + paths[k] + "\"")
		if k < len(paths)-1 {
			builder.WriteRune(',')
		}
	}
	builder.WriteRune(']')

	errInfo := NewErrorInfo(OperationShareSet)
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationShareSet, pcsURL)

	resp, err := pcs.client.Req("POST", pcsURL.String(), map[string]string{
		"path_list":    builder.String(),
		"schannel":     "0",
		"channel_list": "[]",
		"period":       strconv.Itoa(option.Period),
	}, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"User-Agent":   "netdisk;8.3.1",
	})
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		errInfo.errType = ErrTypeNetError
		errInfo.err = err
		return nil, errInfo
	}

	s = &Shared{}
	jsonData := struct {
		*Shared
		*pan.RemoteErrInfo
	}{
		Shared:        s,
		RemoteErrInfo: &pan.RemoteErrInfo{},
	}

	d := jsoniter.NewDecoder(resp.Body)
	err = d.Decode(&jsonData)
	if err != nil {
		errInfo.jsonError(err)
		return nil, errInfo
	}

	if jsonData.RemoteErrInfo.ErrNo != 0 {
		jsonData.RemoteErrInfo.ParseErrMsg()
		errInfo.ErrCode = jsonData.RemoteErrInfo.ErrNo
		errInfo.ErrMsg = jsonData.RemoteErrInfo.ErrMsg
		return nil, errInfo
	}

	if jsonData.Link == "" {
		errInfo.errType = ErrTypeOthers
		errInfo.err = errors.New("未找到分享链接")
		return nil, errInfo
	}

	return jsonData.Shared, nil
}

// ShareCancel 取消分享
func (pcs *BaiduPCS) ShareCancel(shareIDs []int64) (pcsError Error) {
	pcs.lazyInit()

	pcsURL := &url.URL{
		Scheme: GetHTTPScheme(pcs.isHTTPS),
		Host:   "pan.baidu.com",
		Path:   "share/cancel",
	}

	builder := &strings.Builder{}
	builder.WriteRune('[')
	for k := range shareIDs {
		builder.WriteString(strconv.FormatInt(shareIDs[k], 10))
		if k < len(shareIDs)-1 {
			builder.WriteRune(',')
		}
	}
	builder.WriteRune(']')

	errInfo := NewErrorInfo(OperationShareCancel)
	baiduPCSVerbose.Infof("%s URL: %s\n", OperationShareCancel, pcsURL)

	resp, err := pcs.client.Req("POST", pcsURL.String(), map[string]string{
		"shareid_list": builder.String(),
	}, map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
		"User-Agent":   "netdisk;8.3.1",
	})
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		errInfo.errType = ErrTypeNetError
		errInfo.err = err
		return errInfo
	}

	jsonData := struct {
		*pan.RemoteErrInfo
	}{
		RemoteErrInfo: &pan.RemoteErrInfo{},
	}

	d := jsoniter.NewDecoder(resp.Body)
	err = d.Decode(&jsonData)
	if err != nil {
		errInfo.jsonError(err)
		return errInfo
	}

	if jsonData.RemoteErrInfo.ErrNo != 0 {
		jsonData.RemoteErrInfo.ParseErrMsg()
		errInfo.ErrCode = jsonData.RemoteErrInfo.ErrNo
		errInfo.ErrMsg = jsonData.RemoteErrInfo.ErrMsg
		return errInfo
	}

	return nil
}
