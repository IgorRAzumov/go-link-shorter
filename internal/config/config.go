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
	SecretKey       string `env:"SECRET_KEY" env-default:"default-secret-key-change-in-production"`
	AuditFile       string `env:"AUDIT_FILE"`
	AuditURL        string `env:"AUDIT_URL"`
	EnablePprof     bool   `env:"ENABLE_PPROF" env-default:"false"`
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

	if config.AuditURL != "" {
		parsedURL, err := url.Parse(config.AuditURL)
		if err != nil {
			return nil, fmt.Errorf("invalid audit URL: %w", err)
		}
		if parsedURL.Scheme == "" || parsedURL.Host == "" {
			return nil, fmt.Errorf("invalid audit URL")
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
	envSecretKey := cfg.SecretKey
	envAuditFile := cfg.AuditFile
	envAuditURL := cfg.AuditURL

	var serverAddressFlag string
	var baseShortURLFlag string
	var fileStoragePathFlag string
	var databaseAddressFlag string
	var secretKeyFlag string
	var auditFileFlag string
	var auditURLFlag string
	var enablePprofFlag bool

	flag.StringVar(&serverAddressFlag, "a", "localhost:8080", "server address")
	flag.StringVar(&baseShortURLFlag, "b", "", "base shorter URL")
	flag.StringVar(&fileStoragePathFlag, "f", "", "file storage path")
	flag.StringVar(&databaseAddressFlag, "d", "", "database DSN")
	flag.StringVar(&secretKeyFlag, "k", "", "secret key for cookie signing")
	flag.StringVar(&auditFileFlag, "audit-file", "", "audit file path (append json line events)")
	flag.StringVar(&auditURLFlag, "audit-url", "", "audit receiver URL (POST json event)")
	flag.BoolVar(&enablePprofFlag, "pprof", false, "enable pprof debug endpoints at /debug/pprof/")
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
	if envSecretKey == "" || envSecretKey == "default-secret-key-change-in-production" {
		if secretKeyFlag != "" {
			cfg.SecretKey = secretKeyFlag
		}
	}
	if envAuditFile == "" {
		cfg.AuditFile = auditFileFlag
	}
	if envAuditURL == "" {
		cfg.AuditURL = auditURLFlag
	}
	cfg.EnablePprof = cfg.EnablePprof || enablePprofFlag
}
