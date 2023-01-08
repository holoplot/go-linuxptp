package ptp

import (
	"fmt"
	"unsafe"
)

type Pin struct {
	clock *Clock
	index int
	desc  pinDescIoctlStruct
}

func (p *Pin) read() error {
	ptr := unsafe.Pointer(&p.desc)
	code := ptpIoctlMakeCode(ioctlDirReadWrite, magicPinGetFunc2, uintptr(unsafe.Sizeof(p.desc)))

	err := doIoctl(p.clock.file.Fd(), code, ptr)
	if err != nil {
		return fmt.Errorf("failed to send ioctl: %w", err)
	}

	return nil
}

func (p *Pin) GetName() string {
	return string(p.desc.name[:])
}

func (p *Pin) GetFunction() PinFunction {
	return PinFunction(p.desc.function)
}

func (p *Pin) GetChannel() uint32 {
	return p.desc.channel
}

func (p *Pin) SetFunction(function PinFunction, channel uint32) error {
	desc := p.desc

	desc.function = uint32(function)
	desc.channel = channel

	ptr := unsafe.Pointer(&desc)
	code := ptpIoctlMakeCode(ioctlDirWrite, magicPinSetFunc2, uintptr(unsafe.Sizeof(desc)))

	err := doIoctl(p.clock.file.Fd(), code, ptr)
	if err != nil {
		return fmt.Errorf("failed to send ioctl: %w", err)
	}

	return p.read()
}
