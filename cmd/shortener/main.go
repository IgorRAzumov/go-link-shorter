package main

import (
	"github.com/IgorRAzumov/go-link-shorter/internal/app"
	"github.com/IgorRAzumov/go-link-shorter/internal/config"
	"github.com/rs/zerolog/log"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config")
	}

	app.Run(cfg.ServerAddress, cfg.BaseShortURL, cfg.FileStoragePath)
}
