package ptp

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

type Clock struct {
	file *os.File

	caps clockCapsIoctlStruct
	name string

	callbackLock  sync.Mutex
	eventCallback ExternalTimestampEventCallback

	pinsLock sync.Mutex
	pins     map[int]*Pin
}

func Open(index int) (*Clock, error) {
	fname := fmt.Sprintf("/dev/ptp%d", index)

	c := &Clock{
		pins: make(map[int]*Pin),
	}

	var err error

	c.file, err = os.OpenFile(fname, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", fname, err)
	}

	p := unsafe.Pointer(&c.caps)
	code := ptpIoctlMakeCode(ioctlDirRead, magicGetCaps2, uintptr(unsafe.Sizeof(c.caps)))

	err = doIoctl(c.file.Fd(), code, p)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain device capabilities: %w", err)
	}

	fname = fmt.Sprintf("/sys/class/ptp/ptp%d/clock_name", index)
	b, err := os.ReadFile(fname)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain clock name through %s: %w", fname, err)
	}

	c.name = strings.Trim(string(b), "\n\000")

	go c.readEvents()

	return c, nil
}

func (c *Clock) Close() {
	c.file.Close()
}

func (c *Clock) readEvents() {
	for {
		e := externalTimestampEventIoctlStruct{}

		err := binary.Read(c.file, binary.LittleEndian, &e)
		if err != nil {
			return
		}

		i := int(e.Index)
		ts := e.T.time()

		c.callbackLock.Lock()
		if c.eventCallback != nil {
			c.eventCallback(i, ts)
		}
		c.callbackLock.Unlock()
	}
}

// GetName returns the name of this clock.
func (c *Clock) GetName() string {
	return c.name
}

// GetMaxAdj returns the maximum frequency adjustment in parts per billon.
func (c *Clock) GetMaxFrequencyAdjustment() int {
	return int(c.caps.maxAdj)
}

// GetAlarms returns the number of programmable alarms.
func (c *Clock) GetAlarms() int {
	return int(c.caps.nAlarm)
}

// GetExternalTimestampChannels return the number of external time stamp channels.
func (c *Clock) GetExternalTimestampChannels() int {
	return int(c.caps.nExtTs)
}

// GetProgrammablePeriodicSignals return the number of programmable periodic signals.
func (c *Clock) GetProgrammablePeriodicSignals() int {
	return int(c.caps.nPerOut)
}

// GetPpsCallbackSupport returns whether the clock supports a PPS callback.
func (c *Clock) GetPpsCallbackSupport() bool {
	return c.caps.pps != 0
}

func (c *Clock) GetCrossTimestampingSupport() bool {
	return c.caps.crossTimestamping != 0
}

// GetPins returns the number of available input/output pins.
func (c *Clock) GetPins() int {
	return int(c.caps.nPins)
}

func (c *Clock) SetPPSEnabled(enabled bool) error {
	var v uint32

	if enabled {
		v = 1
	}

	p := unsafe.Pointer(&v)
	code := ptpIoctlMakeCode(ioctlDirWrite, magicEnablePps2, uintptr(unsafe.Sizeof(v)))

	err := doIoctl(c.file.Fd(), code, p)
	if err != nil {
		return fmt.Errorf("failed to send ioctl: %w", err)
	}

	return nil
}

func (c *Clock) RequestExternalTimestamp(channel int, flags ExternalTimestampFlag) error {
	o := externalTimestampRequestIoctlStruct{
		index: uint32(channel),
		flags: uint32(flags),
	}
	p := unsafe.Pointer(&o)
	code := ptpIoctlMakeCode(ioctlDirWrite, magicExternalTimestampRequest2, uintptr(unsafe.Sizeof(o)))

	err := doIoctl(c.file.Fd(), code, p)
	if err != nil {
		return fmt.Errorf("failed to send ioctl: %w", err)
	}

	return nil
}

type PeriodicOutputConfig struct {
	// The period in which the output is supposed to fire.
	Period time.Duration

	// Configures the periodic output channel with a fixed absolute start time.
	// AbsoluteStartTime and PhaseOffset are mutually exclusive.
	AbsoluteStartTime *time.Time

	// Configures the periodic output channel with a phase offset.
	// The signal should start toggling at an unspecified integer multiple of the given period, plus the value given in the phase parameter.
	// The start time should be "as soon as possible".
	// AbsoluteStartTime and PhaseOffset are mutually exclusive.
	PhaseOffset *time.Duration

	// Optional, determines the 'on' time of the signal. Must be lower than the period.
	DutyCycle *time.Duration

	// Only run once
	OneShot bool
}

func (c *Clock) ConfigurePeriodicOutput(channel uint32, config PeriodicOutputConfig) error {
	flags := periodicOutputFlags(0)
	buf := new(bytes.Buffer)

	if config.AbsoluteStartTime != nil {
		binary.Write(buf, binary.LittleEndian, timeToIoctlStruct(*config.AbsoluteStartTime))
	}

	if config.PhaseOffset != nil {
		binary.Write(buf, binary.LittleEndian, durationToIoctlStruct(*config.PhaseOffset))
		flags |= ptpPerOutPhase
	}

	binary.Write(buf, binary.LittleEndian, durationToIoctlStruct(config.Period))
	binary.Write(buf, binary.LittleEndian, channel)

	if config.OneShot {
		flags |= ptpPerOutOneShot
	}

	if config.DutyCycle != nil {
		binary.Write(buf, binary.LittleEndian, durationToIoctlStruct(*config.DutyCycle))
		flags |= ptpPerOutDutyCycle
	} else {
		var rsv [4]uint32
		binary.Write(buf, binary.LittleEndian, rsv)
	}

	b := buf.Bytes()
	p := unsafe.Pointer(&b[0])
	code := ptpIoctlMakeCode(ioctlDirWrite, magicPerOutRequest2, uintptr(len(b)))

	return doIoctl(c.file.Fd(), code, p)
}

func (c *Clock) GetPin(index int) (*Pin, error) {
	c.pinsLock.Lock()
	defer c.pinsLock.Unlock()

	if pin, ok := c.pins[index]; ok {
		return pin, nil
	}

	pin := &Pin{
		clock: c,
		index: index,
	}

	if err := pin.read(); err != nil {
		return nil, err
	}

	c.pins[index] = pin

	return pin, nil
}

func (c *Clock) GetSystemOffset(samples int) ([]SystemOffsetMeasurement, error) {
	o := sysOffsetIoctlStruct{
		nSamples: uint32(samples),
	}
	p := unsafe.Pointer(&o)
	code := ptpIoctlMakeCode(ioctlDirWrite, magicSysOffset2, uintptr(unsafe.Sizeof(o)))

	err := doIoctl(c.file.Fd(), code, p)
	if err != nil {
		return nil, fmt.Errorf("failed to send ioctl: %w", err)
	}

	m := make([]SystemOffsetMeasurement, 0, samples)

	for i := 0; i < samples; i++ {
		m = append(m, SystemOffsetMeasurement{
			System: o.ts[i*2].time(),
			Phc:    o.ts[i*2+1].time(),
		})
	}

	return m, nil
}

func (c *Clock) GetSystemOffsetExtended(samples int) ([]SystemOffsetMeasurementExtended, error) {
	o := sysOffsetExtendedIoctlStructIoctlStruct{
		nSamples: uint32(samples),
	}
	p := unsafe.Pointer(&o)
	code := ptpIoctlMakeCode(ioctlDirReadWrite, magicSysOffsetExtended2, uintptr(unsafe.Sizeof(o)))

	err := doIoctl(c.file.Fd(), code, p)
	if err != nil {
		return nil, fmt.Errorf("failed to send ioctl: %w", err)
	}

	m := make([]SystemOffsetMeasurementExtended, 0, samples)

	for i := 0; i < samples; i++ {
		m = append(m, SystemOffsetMeasurementExtended{
			System1: o.ts[i][0].time(),
			Phc:     o.ts[i][1].time(),
			System2: o.ts[i][2].time(),
		})
	}

	return m, nil
}

func (c *Clock) GetSystemOffsetPrecise(samples int) (SystemOffsetMeasurementPrecise, error) {
	o := sysOffsetPreciseIoctlStruct{}
	p := unsafe.Pointer(&o)
	code := ptpIoctlMakeCode(ioctlDirReadWrite, magicSysOffsetPrecise2, uintptr(unsafe.Sizeof(o)))

	err := doIoctl(c.file.Fd(), code, p)
	if err != nil {
		return SystemOffsetMeasurementPrecise{}, fmt.Errorf("failed to send ioctl: %w", err)
	}

	m := SystemOffsetMeasurementPrecise{
		Device:             o.device.time(),
		SystemRealTime:     o.sysRealTime.time(),
		SystemMonotonicRaw: o.sysMonoRaw.time(),
	}

	return m, nil
}

// OnExternalTimestampEvent sets the callback function for external time stamp events that
// are configure through RequestExternalTimestamp().
func (c *Clock) OnExternalTimestampEvent(cb ExternalTimestampEventCallback) {
	c.callbackLock.Lock()
	defer c.callbackLock.Unlock()

	c.eventCallback = cb
}

func (c *Clock) posixId() int {
	// From clock_settime(3):
	// #define CLOCKFD 3
	// #define FD_TO_CLOCKID(fd)   ((~(clockid_t) (fd) << 3) | CLOCKFD)
	return ((^int(c.file.Fd())) << 3) | 3
}

// GetTime returns the current time of the clock.
func (c *Clock) GetTime() (time.Time, error) {
	clockId := c.posixId()
	ts := unix.Timespec{}

	_, _, errno := unix.Syscall(unix.SYS_CLOCK_GETTIME, uintptr(clockId), uintptr(unsafe.Pointer(&ts)), 0)
	if errno != 0 {
		return time.Time{}, fmt.Errorf("ioctl SYS_CLOCK_GETTIME failed: %s", errno.Error())
	}

	t := unixEpoch()
	t = t.Add(time.Duration(ts.Sec) * time.Duration(time.Second))
	t = t.Add(time.Duration(ts.Nsec) * time.Duration(time.Nanosecond))

	return t, nil
}

// SetTime sets the current time of the clock.
func (c *Clock) SetTime(t time.Time) error {
	clockId := c.posixId()
	d := t.Sub(unixEpoch())
	ts := unix.NsecToTimespec(d.Nanoseconds())

	_, _, errno := unix.Syscall(unix.SYS_CLOCK_SETTIME, uintptr(clockId), uintptr(unsafe.Pointer(&ts)), 0)
	if errno != 0 {
		return fmt.Errorf("ioctl SYS_CLOCK_SETTIME failed: %s", errno.Error())
	}

	return nil
}

// AdjustTime gradually adjusts the system clock.
// The amount of time by which the clock is to be adjusted is specified in the structure passes as argument.
// If the adjustment parameter is positive, then the system clock is sped up by some small percentage
// (i.e., by adding a small amount of time to the clock value in each second) until the adjustment has been completed.
// If the adjustment parameter is negative, then the clock is slowed down in a similar fashion.
// Internally, this function calls into the clock_adjtime(3) syscall. Refer the manpage for more information.
func (c *Clock) AdjustTime(t unix.Timex) error {
	clockId := c.posixId()

	_, _, errno := unix.Syscall(unix.SYS_CLOCK_ADJTIME, uintptr(clockId), uintptr(unsafe.Pointer(&t)), 0)
	if errno != 0 {
		return fmt.Errorf("ioctl SYS_CLOCK_ADJTIME failed: %s", errno.Error())
	}

	return nil
}
