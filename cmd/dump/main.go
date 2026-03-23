package main

import (
	"flag"
	"log/slog"

	"github.com/holoplot/go-linuxptp/pkg/logger"
	"github.com/holoplot/go-linuxptp/pkg/ptp"
	"github.com/rs/zerolog/log"
)

func main() {
	indexFlag := flag.Int("index", 0, "Index of PTP device to use")
	flag.Parse()

	logger.Setup()

	c, err := ptp.Open(*indexFlag)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open PTP device")
	}

	slog.Info("Read device information",
		"name", c.GetName(),
		"maxFrequencyAdjustment", c.GetMaxFrequencyAdjustment(),
		"alarms", c.GetAlarms(),
		"pins", c.GetPins(),
		"externalTimestampChannels", c.GetExternalTimestampChannels(),
		"programmablePeriodicSignals", c.GetProgrammablePeriodicSignals(),
		"ppsCallbackSupport", c.GetPpsCallbackSupport(),
		"crossTimestamping", c.GetCrossTimestampingSupport(),
	)

	if t, err := c.GetTime(); err != nil {
		slog.Error("Failed to get time", "error", err)
	} else {
		slog.Info("Got current time", "currentTime", t.String())
	}

	if m, err := c.GetSystemOffset(1); err != nil {
		slog.Warn("Failed to get system offset", "error", err)
	} else {
		slog.Info("Got system offset",
			"system", m[0].System.String(),
			"phc", m[0].Phc.String(),
		)
	}

	if m, err := c.GetSystemOffsetExtended(1); err != nil {
		slog.Warn("Failed to get extended system offset", "error", err)
	} else {
		slog.Info("Got extended system offset",
			"system1", m[0].System1.String(),
			"phc", m[0].Phc.String(),
			"system2", m[0].System2.String(),
			"diff(system1,phc)", m[0].System1.Sub(m[0].Phc).String(),
		)
	}

	if m, err := c.GetSystemOffsetPrecise(1); err != nil {
		slog.Warn("Failed to get precise system offset", "error", err)
	} else {
		slog.Info("Got precise system offset",
			"device", m.Device.String(),
			"systemMonotonicRaw", m.SystemMonotonicRaw.String(),
			"systemRealTime", m.SystemRealTime.String(),
		)
	}
}
