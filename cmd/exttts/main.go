package main

import (
	"context"
	"flag"
	"time"

	"github.com/holoplot/go-linuxptp/pkg/ptp"
	"github.com/mattn/go-colorable"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	indexFlag := flag.Int("index", 0, "Index of PTP device to use")
	channelFlag := flag.Int("channel", 0, "Channel to use")

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

	c.OnExternalTimestampEvent(func(channel int, timestamp time.Time) {
		log.Info().
			Int("channel", channel).
			Str("system2", timestamp.String()).
			Msg("external time stamp event")
	})

	if err := c.RequestExternalTimestamp(*channelFlag, ptp.ExternalTimestampEnable); err != nil {
		log.Fatal().Err(err).Msg("Failed to enable external timestamping")
	}

	ctx := context.Background()
	<-ctx.Done()
}
