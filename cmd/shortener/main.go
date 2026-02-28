package main

import (
	"fmt"
	"os"

	"github.com/IgorRAzumov/go-link-shorter/internal/app"
	"github.com/IgorRAzumov/go-link-shorter/internal/config"
	"github.com/rs/zerolog/log"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	printBuildInfo()

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading config")
	}

	app.Run(cfg)
}

func printBuildInfo() {
	version := buildVersion
	if version == "" {
		version = "N/A"
	}
	date := buildDate
	if date == "" {
		date = "N/A"
	}
	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}
	_, _ = fmt.Fprintf(os.Stdout, "Build version: %s\nBuild date: %s\nBuild commit: %s\n", version, date, commit)
}
