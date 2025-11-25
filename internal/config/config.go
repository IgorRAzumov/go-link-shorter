package config

import (
	"flag"
	"fmt"
	"net/url"
	"strings"
)

type Config struct {
	ServerAddress string
	BaseShortURL  string
}

func Load() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "server address")
	flag.StringVar(&cfg.BaseShortURL, "b", "", "base shorter URL")

	flag.Parse()

	if cfg.BaseShortURL != "" {
		parsedURL, err := url.Parse(cfg.BaseShortURL)
		if err != nil {
			return nil, fmt.Errorf("invalid base shorter URL: %w", err)
		}
		cfg.BaseShortURL = strings.TrimSuffix(cfg.BaseShortURL, "/")
		if parsedURL.Scheme == "" || parsedURL.Host == "" {
			return nil, fmt.Errorf("invalid base shorter URL")
		}
	}

	return cfg, nil
}
