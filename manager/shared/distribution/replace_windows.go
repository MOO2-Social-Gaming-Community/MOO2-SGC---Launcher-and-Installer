//go:build windows

package distribution

import (
	"syscall"
	"unsafe"
)

var moveFileEx = syscall.NewLazyDLL("kernel32.dll").NewProc("MoveFileExW")

func replace(src, dst string) error {
	s, e := syscall.UTF16PtrFromString(src)
	if e != nil {
		return e
	}
	d, e := syscall.UTF16PtrFromString(dst)
	if e != nil {
		return e
	}
	r, _, e := moveFileEx.Call(uintptr(unsafe.Pointer(s)), uintptr(unsafe.Pointer(d)), uintptr(0x1|0x8))
	if r == 0 {
		return e
	}
	return nil
}
