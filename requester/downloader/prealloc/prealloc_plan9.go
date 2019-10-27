//+build plan9

package prealloc

func InitPrivilege() (err error) {
	return nil
}

func PreAlloc(fd uintptr, length int64) error {
	return nil
}
