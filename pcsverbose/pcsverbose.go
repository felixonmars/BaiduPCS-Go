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
