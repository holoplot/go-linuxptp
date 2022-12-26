package ptp

import (
	"bytes"
	"encoding/binary"
	"time"
)

type PtpExtTtsFlag int

const (
	PtpExtTtsEnable      = PtpExtTtsFlag(1 << 0)
	PtpExtTtsRisingEdge  = PtpExtTtsFlag(1 << 1)
	PtpExtTtsFallingEdge = PtpExtTtsFlag(1 << 2)
	PtpExtTtsStrict      = PtpExtTtsFlag(1 << 3)
	PtpExtTtsBothEdges   = PtpExtTtsRisingEdge | PtpExtTtsFallingEdge
	//	flag fields valid for the new PTP_EXTTS_REQUEST2 ioctl.
	PtpExtTtsValidFlags = PtpExtTtsEnable | PtpExtTtsBothEdges | PtpExtTtsFallingEdge | PtpExtTtsStrict
	//	flag fields valid for the new PTP_EXTTS_REQUEST ioctl.
	PtpExtTtsValidV1Flags = PtpExtTtsEnable | PtpExtTtsBothEdges
)

type PtpPerOutFlags int

const (
	PtpPerOutOneShot   = PtpPerOutFlags(1 << 0)
	PtpPerOutDutyCycle = PtpPerOutFlags(1 << 1)
	PtpPerOutDutyPhase = PtpPerOutFlags(1 << 2)
)

type PtpClockTime time.Duration

func (p PtpClockTime) toIoctlStruct() []byte {
	buf := new(bytes.Buffer)

	var seconds int64 = int64(time.Duration(p) / time.Second)
	binary.Write(buf, binary.LittleEndian, seconds)

	var nanoSeconds uint32 = uint32(time.Duration(p) % time.Second)
	binary.Write(buf, binary.LittleEndian, nanoSeconds)

	// reservered
	binary.Write(buf, binary.LittleEndian, uint32(0))

	return buf.Bytes()
}
