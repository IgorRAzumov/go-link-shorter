package config

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	ServerAddress   string
	BaseShortURL    string
	FileStoragePath string
	DatabaseAddress string
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
	var fileStoragePathFlag string
	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "server address")
	flag.StringVar(&cfg.BaseShortURL, "b", "", "base shorter URL")
	flag.StringVar(&fileStoragePathFlag, "f", "", "file storage path")
	flag.StringVar(&cfg.DatabaseAddress, "d", "", "database DSN")
	flag.Parse()

	envServerAddress := os.Getenv("SERVER_ADDRESS")
	if envServerAddress != "" {
		cfg.ServerAddress = envServerAddress
	}

	envBaseShortURL := os.Getenv("BASE_URL")
	if envBaseShortURL != "" {
		cfg.BaseShortURL = envBaseShortURL
	}

	envFileStoragePath := os.Getenv("FILE_STORAGE_PATH")
	if envFileStoragePath != "" {
		cfg.FileStoragePath = envFileStoragePath
	} else if fileStoragePathFlag != "" {
		cfg.FileStoragePath = fileStoragePathFlag
	} else {
		cfg.FileStoragePath = "/tmp/shortener-db.json"
	}

	envDatabaseAddress := os.Getenv("DATABASE_DSN")
	if envDatabaseAddress != "" {
		cfg.DatabaseAddress = envDatabaseAddress
	}
}
