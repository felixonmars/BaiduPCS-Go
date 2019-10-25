package prealloc

import (
	"golang.org/x/sys/windows"
	"sync"
	"syscall"
)

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	procSetFileValidData = kernel32.NewProc("SetFileValidData")
	onceInit             = sync.Once{}
)

func initPrivilege() error {
	current, err := windows.GetCurrentProcess()
	if err != nil {
		return &PreAllocError{
			ProcName: "GetCurrentProcess",
			Err:      err,
		}
	}

	var hToken windows.Token
	err = windows.OpenProcessToken(current, windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, &hToken)
	if err != nil {
		return &PreAllocError{
			ProcName: "OpenProcessToken",
			Err:      err,
		}
	}

	var (
		seManageVolumeName, _ = windows.UTF16PtrFromString("SeManageVolumePrivilege")
		tp                    = windows.Tokenprivileges{
			PrivilegeCount: 1,
			Privileges: [1]windows.LUIDAndAttributes{
				windows.LUIDAndAttributes{
					Luid:       windows.LUID{},
					Attributes: windows.SE_PRIVILEGE_ENABLED,
				},
			},
		}
	)
	err = windows.LookupPrivilegeValue(nil, seManageVolumeName, &tp.Privileges[0].Luid)
	if err != nil {
		return &PreAllocError{
			ProcName: "LookupPrivilegeValue",
			Err:      err,
		}
	}

	err = windows.AdjustTokenPrivileges(hToken, false, &tp, 0, nil, nil)
	if err != nil {
		return &PreAllocError{
			ProcName: "AdjustTokenPrivileges",
			Err:      err,
		}
	}

	return nil
}

// InitPrivilege 初始化权限
func InitPrivilege() (err error) {
	onceInit.Do(func() {
		err = initPrivilege()
	})
	return err
}

// PreAlloc 预分配文件空间
func PreAlloc(fd uintptr, length int64) error {
	err := syscall.Ftruncate(syscall.Handle(fd), length)
	if err != nil {
		return &PreAllocError{
			ProcName: "Ftruncate",
			Err:      err,
		}
	}

	r1, _, err := procSetFileValidData.Call(fd, uintptr(length))
	if r1 == 0 {
		return &PreAllocError{
			ProcName: "SetFileValidData",
			Err:      err,
		}
	}

	return nil
}
