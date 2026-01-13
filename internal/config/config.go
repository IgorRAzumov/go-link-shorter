package config

import (
	"flag"
	"fmt"
	"net/url"
	"strings"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" env-default:"localhost:8080"`
	BaseShortURL    string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseAddress string `env:"DATABASE_DSN"`
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
	_ = cleanenv.ReadEnv(cfg)

	envServerAddress := cfg.ServerAddress
	envBaseShortURL := cfg.BaseShortURL
	envFileStoragePath := cfg.FileStoragePath
	envDatabaseAddress := cfg.DatabaseAddress

	var serverAddressFlag string
	var baseShortURLFlag string
	var fileStoragePathFlag string
	var databaseAddressFlag string

	flag.StringVar(&serverAddressFlag, "a", "localhost:8080", "server address")
	flag.StringVar(&baseShortURLFlag, "b", "", "base shorter URL")
	flag.StringVar(&fileStoragePathFlag, "f", "", "file storage path")
	flag.StringVar(&databaseAddressFlag, "d", "", "database DSN")
	flag.Parse()

	if envServerAddress == "" || envServerAddress == "localhost:8080" {
		if serverAddressFlag != "" {
			cfg.ServerAddress = serverAddressFlag
		}
	}
	if envBaseShortURL == "" {
		cfg.BaseShortURL = baseShortURLFlag
	}
	if envFileStoragePath == "" {
		cfg.FileStoragePath = fileStoragePathFlag
	}
	if envDatabaseAddress == "" {
		cfg.DatabaseAddress = databaseAddressFlag
	}
}
