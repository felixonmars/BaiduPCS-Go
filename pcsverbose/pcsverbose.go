// Package pcsverbose 调试包
package pcsverbose

import (
	"fmt"
	"io"
	"os"
)

var (
	// IsVerbose 是否调试
	IsVerbose = false

	// Output 输出
	Output io.Writer = os.Stderr
)

// PCSVerbose 调试
type PCSVerbose struct {
	Module string
}

// New 根据module, 初始化PCSVerbose
func New(module string) *PCSVerbose {
	return &PCSVerbose{
		Module: module,
	}
}

// Info 提示
func (pv *PCSVerbose) Info(l string) {
	Verbosef("DEBUG: %s INFO: %s\n", pv.Module, l)
}

// Infof 提示, 格式输出
func (pv *PCSVerbose) Infof(format string, a ...interface{}) {
	Verbosef("DEBUG: %s INFO: %s", pv.Module, fmt.Sprintf(format, a...))
}

// Warn 警告
func (pv *PCSVerbose) Warn(l string) {
	Verbosef("DEBUG: %s WARN: %s\n", pv.Module, l)
}

// Verbosef 调试格式输出
func Verbosef(format string, a ...interface{}) (n int, err error) {
	if IsVerbose {
		n, err = fmt.Fprintf(Output, format, a...)
	}
	return
}

// Verboseln 调试输出一行
func Verboseln(a ...interface{}) (n int, err error) {
	if IsVerbose {
		n, err = fmt.Fprintln(Output, a...)
	}
	return
}
