// +build !windows,!plan9

package pcsupdate

import (
	"github.com/iikira/BaiduPCS-Go/pcsutil"
	"syscall"
)

func checkWritable() bool {
	return syscall.Access(pcsutil.ExecutablePath(), syscall.O_RDWR) == nil
}
