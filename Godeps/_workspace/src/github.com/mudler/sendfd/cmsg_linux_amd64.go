package sendfd

import (
	"syscall"
	"unsafe"
)

func CmsgAlign(length uintptr) uintptr {
	return (length + unsafe.Sizeof(int(0)) - 1) & ^(unsafe.Sizeof(int(0)) - 1)
}

func CmsgSpace(length uintptr) uintptr {
	return CmsgAlign(unsafe.Sizeof(syscall.Cmsghdr{})) + CmsgAlign(length)
}

func CmsgLen(length uintptr) uint64 {
	return uint64(CmsgAlign(unsafe.Sizeof(syscall.Cmsghdr{})) + length)
}
