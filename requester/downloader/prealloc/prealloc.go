//+build !windows,!plan9

// Package prealloc 初始化分配文件包
package prealloc

import (
	"syscall"
)

func InitPrivilege() (err error) {
	return nil
}

func PreAlloc(fd uintptr, length int64) error {
	err := syscall.Ftruncate(int(fd), length)
	if err != nil {
		return &PreAllocError{
			ProcName: "Ftruncate",
			Err:      err,
		}
	}
	return nil
}
