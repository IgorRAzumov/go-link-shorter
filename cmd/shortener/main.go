package main

import (
	"log"

	"github.com/IgorRAzumov/go-link-shorter/internal/app"
	"github.com/IgorRAzumov/go-link-shorter/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	app.Run(cfg.ServerAddress, cfg.BaseShortURL)
}
