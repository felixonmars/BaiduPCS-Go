package baidupcs

import (
	"fmt"
)

// ErrInfo 远端服务器返回的错误信息
type ErrInfo struct {
	Operation string `json:"-"`          // 正在进行的操作
	ErrCode   int    `json:"error_code"` // 错误代码
	ErrMsg    string `json:"error_msg"`  // 错误消息
}

func NewErrorInfo(operation string) *ErrInfo {
	return &ErrInfo{
		Operation: operation,
	}
}

// FindErr 查找已知错误
func (e *ErrInfo) FindErr() (errCode int, errMsg string) {
	return findErr(e.ErrCode, e.ErrMsg)
}

func (e *ErrInfo) Error() string {
	if e.ErrCode == 0 {
		return e.Operation + " 操作成功"
	}

	code, msg := e.FindErr()
	return fmt.Sprintf("%s 遇到错误, 远端服务器返回错误代码: %d, 消息: %s", e.Operation, code, msg)
}

// findErr 检查 PCS 错误, 查找已知错误
func findErr(errCode int, errMsg string) (int, string) {
	switch errCode {
	case 0:
		return errCode, ""
	case 31045: // user not exists
		return errCode, "操作失败, 可能百度帐号登录状态过期, 请尝试重新登录, 消息: " + errMsg
	}
	return errCode, errMsg
}
