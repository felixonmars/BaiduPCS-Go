package baidupcs

import (
	"fmt"
)

// Error 错误详情接口
type Error interface {
	error
	Operation() string    // 操作
	ErrorType() ErrType   // 错误类型
	ErrorCode() int       // 远端服务器错误代码
	ErrorMsg() string     // 远端服务器错误信息
	OriginalError() error // 原始错误
}

// ErrType 错误类型
type ErrType int

const (
	// StrSuccess 操作成功
	StrSuccess = "操作成功"
	// StrRemoteError 远端服务器返回错误
	StrRemoteError = "远端服务器返回错误"
	// StrJSONParseError json 数据解析失败
	StrJSONParseError = "json 数据解析失败"

	// ErrTypeInternalError 内部错误
	ErrTypeInternalError ErrType = iota
	// ErrTypeRemoteError 远端服务器返回错误
	ErrTypeRemoteError
	// ErrTypeNetError 网络错误
	ErrTypeNetError
	// ErrTypeJSONParseError json 数据解析失败
	ErrTypeJSONParseError
	// ErrTypeOthers 其他错误
	ErrTypeOthers
)

// ErrInfo 错误信息
type ErrInfo struct {
	operation string // 正在进行的操作
	errType   ErrType
	err       error
	ErrCode   int    `json:"error_code"` // 错误代码
	ErrMsg    string `json:"error_msg"`  // 错误消息
}

// NewErrorInfo 提供operation操作名称, 返回 *ErrInfo
func NewErrorInfo(operation string) *ErrInfo {
	return &ErrInfo{
		operation: operation,
		errType:   ErrTypeRemoteError,
	}
}

func (e *ErrInfo) jsonError(err error) {
	e.errType = ErrTypeJSONParseError
	e.err = err
}

// FindErr 查找已知错误
func (e *ErrInfo) FindErr() (errCode int, errMsg string) {
	return findErr(e.ErrCode, e.ErrMsg)
}

func (e *ErrInfo) Error() string {
	if e.operation == "" {
		if e.err != nil {
			return e.err.Error()
		}
		return StrSuccess
	}

	switch e.errType {
	case ErrTypeInternalError:
		return fmt.Sprintf("%s, %s, %s", e.operation, "内部错误", e.err)
	case ErrTypeJSONParseError:
		return fmt.Sprintf("%s, %s, %s", e.operation, StrJSONParseError, e.err)
	case ErrTypeNetError:
		return fmt.Sprintf("%s, %s, %s", e.operation, "网络错误", e.err)
	case ErrTypeRemoteError:
		if e.ErrCode == 0 {
			if e.operation != "" {
				return e.operation + ", " + StrSuccess
			}
			return StrSuccess
		}

		code, msg := e.FindErr()
		return fmt.Sprintf("%s, 遇到错误, %s, 代码: %d, 消息: %s", e.operation, StrRemoteError, code, msg)
	case ErrTypeOthers:
		if e.err == nil {
			if e.operation != "" {
				return e.operation + ", " + StrSuccess
			}
			return StrSuccess
		}

		return fmt.Sprintf("%s, 遇到错误, %s", e.operation, e.err)
	default:
		panic("unknown ErrType")
	}
}

// Operation return operation
func (e *ErrInfo) Operation() string {
	return e.operation
}

// ErrorType return error type "ErrType"
func (e *ErrInfo) ErrorType() ErrType {
	return e.errType
}

// ErrorCode 返回远端服务器错误代码
func (e *ErrInfo) ErrorCode() int {
	return e.ErrCode
}

// ErrorMsg 返回远端服务器错误信息
func (e *ErrInfo) ErrorMsg() string {
	return e.ErrMsg
}

// OriginalError 返回原始错误
func (e *ErrInfo) OriginalError() error {
	return e.err
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
