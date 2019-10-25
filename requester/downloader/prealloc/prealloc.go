//+build !windows

// Package prealloc 初始化分配文件包
package prealloc

func InitPrivilege() (err error) {
	return nil
}

func PreAlloc(fd uintptr, length int64) error {
	return nil
}
