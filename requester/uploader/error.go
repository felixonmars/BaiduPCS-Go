package uploader

import (
	"errors"
)

type (
	// ErrorControl 多线程上传的出错控制
	ErrorControl interface {
		error
		// IsRetry 是否重试,
		// 当不重试时返回ErrTerminatd
		IsRetry() bool
	}
)

var (
	ErrTerminatd = errors.New("task terminated")
)
