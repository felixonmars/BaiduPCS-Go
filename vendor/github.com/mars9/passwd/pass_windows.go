package passwd

import (
	"bufio"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

func init() {
	passFunc = getPasswd
}

const enableEchoInput = 0x0004

var (
	kernel32 syscall.Handle
	getMode  uintptr
	setMode  uintptr
)

func initHandle() (err error) {
	kernel32, err = syscall.LoadLibrary("kernel32.dll")
	if err != nil {
		return
	}
	getMode, err = syscall.GetProcAddress(kernel32, "GetConsoleMode")
	if err != nil {
		return
	}
	setMode, err = syscall.GetProcAddress(kernel32, "SetConsoleMode")
	return
}

func getConsoleMode(handle syscall.Handle, mode *uint32) error {
	if _, _, err := syscall.Syscall(
		getMode,
		2,
		uintptr(handle),
		uintptr(unsafe.Pointer(mode)),
		0,
	); err != 0 {
		return fmt.Errorf("GetConsoleMode: %s", err.Error())
	}
	return nil
}

func setConsoleMode(handle syscall.Handle, mode uint32) error {
	if _, _, err := syscall.Syscall(
		setMode,
		2,
		uintptr(handle),
		uintptr(mode),
		0,
	); err != 0 {
		return fmt.Errorf("SetConsoleMode: %s", err.Error())
	}
	return nil
}

func getPasswd(prompt string) ([]byte, error) {
	if err := initHandle(); err != nil {
		return nil, err
	}
	defer syscall.CloseHandle(kernel32)

	fd, err := syscall.GetStdHandle(syscall.STD_INPUT_HANDLE)
	if err != nil {
		return nil, err
	}

	var omode uint32
	if err := getConsoleMode(fd, &omode); err != nil {
		return nil, err
	}

	nmode := omode
	nmode &^= enableEchoInput
	if err := setConsoleMode(fd, nmode); err != nil {
		return nil, err
	}
	defer setConsoleMode(fd, omode)

	fmt.Fprint(os.Stdout, prompt)
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	fmt.Fprint(os.Stdout, "\n")
	return line[:len(line)-1], nil
}
