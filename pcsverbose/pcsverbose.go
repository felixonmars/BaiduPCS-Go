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

	// Prefix 调试信息前缀
	Prefix = func() string {
		return "verbose: "
	}
)

// Verbosef 调试格式输出
func Verbosef(format string, a ...interface{}) (n int, err error) {
	if IsVerbose {
		return fmt.Fprintf(Output, Prefix()+format, a...)
	}
	return
}

// Verboseln 调试输出一行
func Verboseln(a ...interface{}) (n int, err error) {
	if IsVerbose {
		n1, err := fmt.Fprint(Output, Prefix())
		n2, err := fmt.Fprintln(Output, a...)
		return n1 + n2, err
	}
	return
}
