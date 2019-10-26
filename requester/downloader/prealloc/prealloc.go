//+build !windows

// Package prealloc 初始化分配文件包
package prealloc

import (
	"os"
)

func InitPrivilege() (err error) {
	return nil
}

func PreAlloc(fd uintptr, length int64) error {
	file := os.NewFile(fd, "truncfile")
	err := file.Truncate(length)
	if err != nil {
		perr, ok := err.(*os.PathError)
		if !ok {
			return err
		}

		return &PreAllocError{
			ProcName: perr.Op,
			Err:      perr.Err,
		}
	}
	return nil
}
