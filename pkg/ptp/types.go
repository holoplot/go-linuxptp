package ptp

import (
	"time"
)

type clockCapsIoctlStruct struct {
	maxAdj            uint32
	nAlarm            uint32
	nExtTs            uint32
	nPerOut           uint32
	pps               uint32
	nPins             uint32
	crossTimestamping uint32
	adjustPhase       uint32
	rsv               [12]uint32
}

type ExternalTimestampFlag int

const (
	ExternalTimestampEnable      = ExternalTimestampFlag(1 << 0)
	ExternalTimestampRisingEdge  = ExternalTimestampFlag(1 << 1)
	ExternalTimestampFallingEdge = ExternalTimestampFlag(1 << 2)
	ExternalTimestampStrict      = ExternalTimestampFlag(1 << 3)
	ExternalTimestampBothEdges   = ExternalTimestampRisingEdge | ExternalTimestampFallingEdge
)

type externalTimestampRequestIoctlStruct struct {
	index uint32
	flags uint32
	rsv   [2]uint32
}

type periodicOutputFlags uint32

const (
	ptpPerOutOneShot   = periodicOutputFlags(1 << 0)
	ptpPerOutDutyCycle = periodicOutputFlags(1 << 1)
	ptpPerOutPhase     = periodicOutputFlags(1 << 2)
)

type clockTimeIoctlStruct struct {
	Sec      int64
	Nsec     uint32
	Reserved uint32
}

func unixEpoch() time.Time {
	return time.Unix(0, 0).UTC()
}

func (cis clockTimeIoctlStruct) duration() time.Duration {
	d := time.Duration(0)
	d += time.Duration(cis.Sec) * time.Second
	d += time.Duration(cis.Nsec) * time.Nanosecond

	return d
}

func (cis clockTimeIoctlStruct) time() time.Time {
	t := unixEpoch()
	t = t.Add(time.Duration(cis.Sec) * time.Second)
	t = t.Add(time.Duration(cis.Nsec) * time.Nanosecond)

	return t
}

func durationToIoctlStruct(d time.Duration) clockTimeIoctlStruct {
	return clockTimeIoctlStruct{
		Sec:  int64(d.Seconds()),
		Nsec: uint32(d.Nanoseconds()),
	}
}

func timeToIoctlStruct(t time.Time) clockTimeIoctlStruct {
	return durationToIoctlStruct(t.Sub(unixEpoch()))
}

const MaxSamples = 25

type sysOffsetIoctlStruct struct {
	nSamples uint32
	rsv      [3]uint32

	// Array of interleaved system/phc time stamps. The kernel
	// will provide 2*nSamples + 1 time stamps, with the last
	// one as a system time stamp.
	ts [2*MaxSamples + 1]clockTimeIoctlStruct
}

type SystemOffsetMeasurement struct {
	System time.Time
	Phc    time.Time
}

type sysOffsetExtendedIoctlStructIoctlStruct struct {
	nSamples uint32
	rsv      [3]uint32

	// Array of [system, phc, system] time stamps. The kernel will provide
	// 3*nSamples time stamps.
	ts [MaxSamples][3]clockTimeIoctlStruct
}

type SystemOffsetMeasurementExtended struct {
	System1 time.Time
	Phc     time.Time
	System2 time.Time
}

type sysOffsetPreciseIoctlStruct struct {
	device      clockTimeIoctlStruct
	sysRealTime clockTimeIoctlStruct
	sysMonoRaw  clockTimeIoctlStruct
	rsv         [4]uint32
}

type SystemOffsetMeasurementPrecise struct {
	Device             time.Time
	SystemRealTime     time.Time
	SystemMonotonicRaw time.Time
}

type pinDescIoctlStruct struct {
	name     [64]byte
	index    uint32
	function uint32
	channel  uint32
	rsv      [5]uint32
}

type PinFunction uint32

const (
	PinFunctionNone              = PinFunction(0)
	PinFunctionExternalTimestamp = PinFunction(1)
	PinFunctionPerOut            = PinFunction(2)
	PinFunctionPhySync           = PinFunction(3)
)

type externalTimestampEventIoctlStruct struct {
	T     clockTimeIoctlStruct
	Index uint32
	Flags uint32
	Rsv   [2]uint32
}

type ExternalTimestampEventCallback func(channel int, timestamp time.Time)
