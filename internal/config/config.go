package config

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	ServerAddress string
	BaseShortURL  string
}

func Load() (*Config, error) {
	config := &Config{}
	extractStartConfig(config)

	if config.BaseShortURL != "" {
		parsedURL, err := url.Parse(config.BaseShortURL)
		if err != nil {
			return nil, fmt.Errorf("invalid base shorter URL: %w", err)
		}

		config.BaseShortURL = strings.TrimSuffix(config.BaseShortURL, "/")
		if parsedURL.Scheme == "" || parsedURL.Host == "" {
			return nil, fmt.Errorf("invalid base shorter URL")
		}
	}

	return config, nil
}

func extractStartConfig(cfg *Config) {
	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "server address")
	flag.StringVar(&cfg.BaseShortURL, "b", "", "base shorter URL")
	flag.Parse()

	envServerAddress := os.Getenv("SERVER_ADDRESS")
	if envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}

	envBaseShortURL := os.Getenv("BASE_URL")
	if envBaseShortURL != "" {
		cfg.BaseShortURL = envBaseShortURL
	}
}
