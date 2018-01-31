package pcsverbose

import (
	"fmt"
	"io"
	"os"
)

var (
	// IsVerbose 是否调试
	IsVerbose bool = false

	// Output 输出
	Output io.Writer = os.Stderr
)

// Verbosef 调试格式输出
func Verbosef(format string, a ...interface{}) {
	if IsVerbose {
		fmt.Fprintf(Output, format, a...)
	}
}

// Verboseln 调试输出一行
func Verboseln(a ...interface{}) {
	if IsVerbose {
		fmt.Fprintln(Output, a...)
	}
}
