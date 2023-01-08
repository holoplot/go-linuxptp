package ptp

import (
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

const (
	ioctlDirNone      = 0x0
	ioctlDirWrite     = 0x1
	ioctlDirRead      = 0x2
	ioctlDirReadWrite = ioctlDirWrite | ioctlDirRead
)

func doIoctl(fd uintptr, code uint32, ptr unsafe.Pointer) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, uintptr(code), uintptr(ptr))
	if errno != 0 {
		return errors.New(errno.Error())
	}

	return nil
}

func ioctlMakeCode(dir, typ, nr int, size uintptr) uint32 {
	var code uint32
	if dir > ioctlDirWrite|ioctlDirRead {
		panic(fmt.Errorf("invalid ioctl dir value: %d", dir))
	}

	if size > 1<<14 {
		panic(fmt.Errorf("invalid ioctl size value: %d", size))
	}

	code |= uint32(dir) << 30
	code |= uint32(size) << 16
	code |= uint32(typ) << 8
	code |= uint32(nr)

	return code
}

func ptpIoctlMakeCode(dir int, nr magic, size uintptr) uint32 {
	return ioctlMakeCode(dir, '=', int(nr), size)
}

type magic int

const (
	// deprecated
	// magicGetCaps           = magic(1)
	// magicExternalTimestampRequest     = magic(2)
	// magicPerOutRequest     = magic(3)
	// magicEnablePps         = magic(4)
	// magicSysOffset         = magic(5)
	// magicPinGetFunc        = magic(6)
	// magicPinSetFunc        = magic(7)
	// magicSysOffsetPrecise  = magic(8)
	// magicSysOffsetExtended = magic(9)

	magicGetCaps2                  = magic(10)
	magicExternalTimestampRequest2 = magic(11)
	magicPerOutRequest2            = magic(12)
	magicEnablePps2                = magic(13)
	magicSysOffset2                = magic(14)
	magicPinGetFunc2               = magic(15)
	magicPinSetFunc2               = magic(16)
	magicSysOffsetPrecise2         = magic(17)
	magicSysOffsetExtended2        = magic(18)
)
