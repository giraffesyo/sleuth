package main

import (
	"os"

	"github.com/giraffesyo/sleuth/internal/sleuth"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	sleuth := sleuth.NewSleuth(
		sleuth.WithProvider(sleuth.ProviderCNN),
		sleuth.WithProvider(sleuth.ProviderFoxNews),
	)
	sleuth.Run()
}
