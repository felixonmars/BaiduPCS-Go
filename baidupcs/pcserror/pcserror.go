// Package pcserror PCS错误包
package pcserror

import (
	"github.com/json-iterator/go"
	"io"
)

type (
	// ErrType 错误类型
	ErrType int

	// Error 错误信息接口
	Error interface {
		error
		SetJSONError(err error)
		SetNetError(err error)
		SetRemoteError()
		GetOperation() string
		GetErrType() ErrType
		GetRemoteErrCode() int
		GetRemoteErrMsg() string
		GetError() error
	}
)

const (
	// ErrorTypeNoError 无错误
	ErrorTypeNoError ErrType = iota
	// ErrTypeInternalError 内部错误
	ErrTypeInternalError
	// ErrTypeRemoteError 远端服务器返回错误
	ErrTypeRemoteError
	// ErrTypeNetError 网络错误
	ErrTypeNetError
	// ErrTypeJSONParseError json 数据解析失败
	ErrTypeJSONParseError
	// ErrTypeOthers 其他错误
	ErrTypeOthers
)

const (
	// StrSuccess 操作成功
	StrSuccess = "操作成功"
	// StrInternalError 内部错误
	StrInternalError = "内部错误"
	// StrRemoteError 远端服务器返回错误
	StrRemoteError = "远端服务器返回错误"
	// StrNetError 网络错误
	StrNetError = "网络错误"
	// StrJSONParseError json 数据解析失败
	StrJSONParseError = "json 数据解析失败"
)

// DecodePCSJSONError 解析PCS JSON的错误
func DecodePCSJSONError(opreation string, data io.Reader) Error {
	errInfo := NewPCSErrorInfo(opreation)
	return decodeJSONError(data, errInfo)
}

// DecodePanJSONError 解析Pan JSON的错误
func DecodePanJSONError(opreation string, data io.Reader) Error {
	errInfo := NewPanErrorInfo(opreation)
	return decodeJSONError(data, errInfo)
}

func decodeJSONError(data io.Reader, errInfo Error) Error {
	var (
		d   = jsoniter.NewDecoder(data)
		err error
	)

	switch value := errInfo.(type) {
	case *PCSErrInfo:
		err = d.Decode(value)
	case *PanErrorInfo:
		err = d.Decode(value)
	}

	if err != nil {
		errInfo.SetJSONError(err)
		return errInfo
	}

	if errInfo.GetRemoteErrCode() != 0 {
		errInfo.SetRemoteError()
		return errInfo
	}

	return nil
}
