// +build !windows,!plan9

package checkaccess

import (
	"syscall"
)

func AccessRDWR(path string) bool {
	return syscall.Access(path, syscall.O_RDWR) == nil
}
