package dlinkclient

import (
	"github.com/iikira/BaiduPCS-Go/baidupcs/pcserror"
	"github.com/json-iterator/go"
	"io"
)

func handleJSONParse(op string, data io.Reader, info interface{}) (dlinkError pcserror.Error) {
	var (
		d       = jsoniter.NewDecoder(data)
		err     = d.Decode(info)
		errInfo = info.(pcserror.Error)
	)

	if errInfo == nil {
		errInfo = pcserror.NewDlinkErrInfo(op)
	}

	if err != nil {
		errInfo.SetJSONError(err)
		return errInfo
	}

	// 设置出错类型为远程错误
	if errInfo.GetRemoteErrCode() != 0 {
		errInfo.SetRemoteError()
		return errInfo
	}

	return nil
}
