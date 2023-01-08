package main

import (
	"flag"

	"github.com/holoplot/go-linuxptp/pkg/ptp"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	indexFlag := flag.Int("index", 0, "Index of PTP device to use")
	flag.Parse()

	consoleWriter := zerolog.ConsoleWriter{
		Out: colorable.NewColorableStdout(),
	}

	log.Logger = log.Output(consoleWriter)

	c, err := ptp.Open(*indexFlag)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to open PTP device")
	}

	log.Info().
		Str("name", c.GetName()).
		Int("maxFrequencyAdjustment", c.GetMaxFrequencyAdjustment()).
		Int("alarms", c.GetAlarms()).
		Int("pins", c.GetPins()).
		Int("externalTimestampChannels", c.GetExternalTimestampChannels()).
		Int("programmablePeriodicSignals", c.GetProgrammablePeriodicSignals()).
		Bool("ppsCallbackSupport", c.GetPpsCallbackSupport()).
		Bool("crossTimestamping", c.GetCrossTimestampingSupport()).
		Msg("Read device information")

	if t, err := c.GetTime(); err != nil {
		log.Error().
			Err(err).
			Msg("Failed to get current time")
	} else {
		log.Info().
			Str("currentTime", t.String()).
			Msg("current time")
	}

	if m, err := c.GetSystemOffset(1); err != nil {
		log.Warn().
			Err(err).
			Msg("Failed to get system offset")
	} else {
		log.Info().
			Str("system", m[0].System.String()).
			Str("phc", m[0].Phc.String()).
			Msg("Got system offset")
	}

	if m, err := c.GetSystemOffsetExtended(1); err != nil {
		log.Warn().
			Err(err).
			Msg("Failed to get extended system offset")
	} else {
		log.Info().
			Str("system1", m[0].System1.String()).
			Str("phc", m[0].Phc.String()).
			Str("system2", m[0].System2.String()).
			Msg("Got extended system offset")
	}

	if m, err := c.GetSystemOffsetPrecise(1); err != nil {
		log.Warn().
			Err(err).
			Msg("Failed to get precise system offset")
	} else {
		log.Info().
			Str("device", m.Device.String()).
			Str("systemMonotonicRaw", m.SystemMonotonicRaw.String()).
			Str("systemRealTime", m.SystemRealTime.String()).
			Msg("Got precise system offset")
	}
}
