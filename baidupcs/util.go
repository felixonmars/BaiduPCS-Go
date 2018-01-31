package baidupcs

import (
	"errors"
	"github.com/bitly/go-simplejson"
)

func checkErr(json *simplejson.Json) (code int, msg error) {
	codeJSON, ok1 := json.CheckGet("error_code")
	msgJSON, ok2 := json.CheckGet("error_msg")
	if !ok1 && !ok2 { // 未发现错误信息
		return 0, nil
	}

	errCode := codeJSON.MustInt()
	errMsg := msgJSON.MustString()
	switch errCode {
	case 31045: // user not exists
		errMsg = "操作失败, 可能百度帐号登录状态过期, 请尝试重新登录, 消息: " + errMsg
	}
	return errCode, errors.New(errMsg)
}
