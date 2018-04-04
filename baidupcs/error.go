package baidupcs

import (
	"fmt"
)

// ErrType 错误类型
type ErrType int

const (
	// StrSuccess 操作成功
	StrSuccess = "操作成功"
	// StrRemoteError 远端服务器返回错误
	StrRemoteError = "远端服务器返回错误"
	// StrJSONEncodeError json 数据生成失败
	StrJSONEncodeError = "json 数据生成失败"
	// StrJSONParseError json 数据解析失败
	StrJSONParseError = "json 数据解析失败"

	// ErrTypeRemoteError 远端服务器返回错误
	ErrTypeRemoteError ErrType = iota
	// ErrTypeNetError 网络错误
	ErrTypeNetError
	// ErrTypeJSONEncodeError json 数据生成失败
	ErrTypeJSONEncodeError
	// ErrTypeJSONParseError json 数据解析失败
	ErrTypeJSONParseError
	// ErrTypeOthers 其他错误
	ErrTypeOthers
)

// ErrInfo 错误信息
type ErrInfo struct {
	Operation string  `json:"-"` // 正在进行的操作
	ErrType   ErrType `json:"-"`
	Err       error   `json:"-"`
	ErrCode   int     `json:"error_code"` // 错误代码
	ErrMsg    string  `json:"error_msg"`  // 错误消息
}

// NewErrorInfo 提供operation操作名称, 返回 *ErrInfo
func NewErrorInfo(operation string) *ErrInfo {
	return &ErrInfo{
		Operation: operation,
		ErrType:   ErrTypeRemoteError,
	}
}

func (e *ErrInfo) jsonError(err error) {
	e.ErrType = ErrTypeJSONParseError
	e.Err = err
}

// FindErr 查找已知错误
func (e *ErrInfo) FindErr() (errCode int, errMsg string) {
	return findErr(e.ErrCode, e.ErrMsg)
}

func (e *ErrInfo) Error() string {
	switch e.ErrType {
	case ErrTypeJSONEncodeError:
		return fmt.Sprintf("%s, %s, %s", e.Operation, StrJSONEncodeError, e.Err)
	case ErrTypeJSONParseError:
		return fmt.Sprintf("%s, %s, %s", e.Operation, StrJSONParseError, e.Err)
	case ErrTypeNetError:
		return fmt.Sprintf("%s, %s, %s", e.Operation, "网络错误", e.Err)
	case ErrTypeRemoteError:
		if e.ErrCode == 0 {
			if e.Operation != "" {
				return e.Operation + ", " + StrSuccess
			}
			return StrSuccess
		}

		code, msg := e.FindErr()
		return fmt.Sprintf("%s, 遇到错误, %s, 代码: %d, 消息: %s", e.Operation, StrRemoteError, code, msg)
	case ErrTypeOthers:
		if e.Err == nil {
			if e.Operation != "" {
				return e.Operation + ", " + StrSuccess
			}
			return StrSuccess
		}
		return fmt.Sprintf("%s, 遇到错误, %s", e.Operation, e.Err)
	default:
		panic("unknown ErrType")
	}
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
