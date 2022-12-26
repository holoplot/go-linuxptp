package ptp

type ptpClockCaps struct {
	// Maximum frequency adjustment in parts per billon.
	maxAdj uint32

	// Number of programmable alarms.
	nAlarm uint32

	// Number of external time stamp channels.
	nExtTs uint32

	// Number of programmable periodic signals.
	nPerOut uint32

	// Whether the clock supports a PPS callback.
	pps uint32

	// Number of input/output pins.
	nPins uint32

	// Whether the clock supports precise system-device cross timestamps
	crossTimestamping uint32

	// Whether the clock supports adjust phase
	adjustPhase uint32

	// Reserved for future use.
	rsv [12]uint32
}

type Clock struct {
	caps ptpClockCaps
}

// GetMaxAdj returns the maximum frequency adjustment in parts per billon.
func (c *Clock) GetMaxFrequencyAdjustment() uint32 {
	return c.caps.maxAdj
}

// GetAlarms returns the number of programmable alarms.
func (c *Clock) GetAlarms() uint32 {
	return c.caps.nAlarm
}

// GetExternalTimestampChannels return the number of external time stamp channels.
func (c *Clock) GetExternalTimestampChannels() uint32 {
	return c.caps.nExtTs
}

// GetProgrammablePeriodicSignals return the number of programmable periodic signals.
func (c *Clock) GetProgrammablePeriodicSignals() uint32 {
	return c.caps.nPerOut
}

// GetPpsCallbackSupport returns whether the clock supports a PPS callback.
func (c *Clock) GetPpsCallbackSupport() bool {
	return c.caps.pps != 0
}
