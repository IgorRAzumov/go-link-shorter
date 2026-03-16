package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// Config — конфигурация приложения (сервер, хранилище, аудит).
type Config struct {
	ServerAddress   string `mapstructure:"server_address"`
	BaseShortURL    string `mapstructure:"base_url"`
	FileStoragePath string `mapstructure:"file_storage_path"`
	DatabaseAddress string `mapstructure:"database_dsn"`
	SecretKey       string `mapstructure:"secret_key"`
	AuditFile       string `mapstructure:"audit_file"`
	AuditURL        string `mapstructure:"audit_url"`
	EnablePprof     bool   `mapstructure:"enable_pprof"`
	EnableHTTPS     bool   `mapstructure:"enable_https"`
	TLSCertFile     string `mapstructure:"tls_cert_file"`
	TLSKeyFile      string `mapstructure:"tls_key_file"`
}

// Load загружает конфигурацию из файла (если указан), переменных окружения и флагов.
// Приоритет: флаги > переменные окружения > файл конфигурации.
func Load() (*Config, error) {
	flags := parseFlags()
	v, err := buildViper(flags)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	if err := validateURLs(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

type flagValues struct {
	configPath      string
	serverAddress   string
	baseShortURL    string
	fileStoragePath string
	databaseDSN     string
	secretKey       string
	auditFile       string
	auditURL        string
	enablePprof     bool
	enableHTTPS     bool
	tlsCertFile     string
	tlsKeyFile      string
}

func parseFlags() flagValues {
	configPath := pflag.StringP("config", "c", "", "path to JSON config file")
	serverAddress := pflag.StringP("server-address", "a", "", "server address")
	baseShortURL := pflag.StringP("base-url", "b", "", "base shorter URL")
	fileStoragePath := pflag.StringP("file-storage", "f", "", "file storage path")
	databaseDSN := pflag.StringP("database", "d", "", "database DSN")
	secretKey := pflag.StringP("secret-key", "k", "", "secret key for cookie signing")
	auditFile := pflag.String("audit-file", "", "audit file path (append json line events)")
	auditURL := pflag.String("audit-url", "", "audit receiver URL (POST json event)")
	enablePprof := pflag.Bool("pprof", false, "enable pprof debug endpoints at /debug/pprof/")
	enableHTTPS := pflag.BoolP("https", "s", false, "enable HTTPS (TLS) server")
	tlsCertFile := pflag.String("tls-cert", "", "path to TLS certificate file (default: cert.pem when HTTPS enabled)")
	tlsKeyFile := pflag.String("tls-key", "", "path to TLS private key file (default: key.pem when HTTPS enabled)")
	pflag.Parse()

	configFilePath := *configPath
	if configFilePath == "" {
		configFilePath = os.Getenv("CONFIG")
	}

	return flagValues{
		configPath:      configFilePath,
		serverAddress:   *serverAddress,
		baseShortURL:    *baseShortURL,
		fileStoragePath: *fileStoragePath,
		databaseDSN:     *databaseDSN,
		secretKey:       *secretKey,
		auditFile:       *auditFile,
		auditURL:        *auditURL,
		enablePprof:     *enablePprof,
		enableHTTPS:     *enableHTTPS,
		tlsCertFile:     *tlsCertFile,
		tlsKeyFile:      *tlsKeyFile,
	}
}

func buildViper(flags flagValues) (*viper.Viper, error) {
	v := viper.New()
	setDefaults(v)
	bindEnv(v)

	if err := loadConfigFile(v, flags.configPath); err != nil {
		return nil, err
	}
	applyFlags(v, flags)
	applyTLSDefaults(v)
	return v, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("server_address", "localhost:8080")
	v.SetDefault("secret_key", "default-secret-key-change-in-production")
	v.SetDefault("enable_pprof", false)
	v.SetDefault("enable_https", false)
}

func bindEnv(v *viper.Viper) {
	envBindings := map[string]string{
		"server_address":   "SERVER_ADDRESS",
		"base_url":         "BASE_URL",
		"file_storage_path": "FILE_STORAGE_PATH",
		"database_dsn":     "DATABASE_DSN",
		"secret_key":       "SECRET_KEY",
		"audit_file":       "AUDIT_FILE",
		"audit_url":        "AUDIT_URL",
		"enable_pprof":     "ENABLE_PPROF",
		"enable_https":     "ENABLE_HTTPS",
		"tls_cert_file":    "TLS_CERT_FILE",
		"tls_key_file":     "TLS_KEY_FILE",
	}
	for key, env := range envBindings {
		_ = v.BindEnv(key, env)
	}
}

func loadConfigFile(v *viper.Viper, path string) error {
	if path == "" {
		return nil
	}
	v.SetConfigFile(path)
	if err := v.ReadInConfig(); err != nil {
		var configErr viper.ConfigFileNotFoundError
		if errors.As(err, &configErr) {
			return fmt.Errorf("config file not found: %s", path)
		}
		return fmt.Errorf("reading config file: %w", err)
	}
	return nil
}

func applyFlags(v *viper.Viper, flags flagValues) {
	overrideIfEnvEmpty := func(key, envVar, flagVal string) {
		if flagVal != "" && os.Getenv(envVar) == "" {
			v.Set(key, flagVal)
		}
	}

	overrideIfEnvEmpty("server_address", "SERVER_ADDRESS", flags.serverAddress)
	overrideIfEnvEmpty("base_url", "BASE_URL", flags.baseShortURL)
	overrideIfEnvEmpty("file_storage_path", "FILE_STORAGE_PATH", flags.fileStoragePath)
	overrideIfEnvEmpty("database_dsn", "DATABASE_DSN", flags.databaseDSN)
	overrideIfEnvEmpty("secret_key", "SECRET_KEY", flags.secretKey)
	overrideIfEnvEmpty("audit_file", "AUDIT_FILE", flags.auditFile)
	overrideIfEnvEmpty("audit_url", "AUDIT_URL", flags.auditURL)
	overrideIfEnvEmpty("tls_cert_file", "TLS_CERT_FILE", flags.tlsCertFile)
	overrideIfEnvEmpty("tls_key_file", "TLS_KEY_FILE", flags.tlsKeyFile)

	if os.Getenv("ENABLE_PPROF") == "" && flagChanged("pprof") {
		v.Set("enable_pprof", flags.enablePprof)
	}
	if os.Getenv("ENABLE_HTTPS") == "" && flagChanged("https") {
		v.Set("enable_https", flags.enableHTTPS)
	}
}

func flagChanged(name string) bool {
	f := pflag.Lookup(name)
	return f != nil && f.Changed
}

func applyTLSDefaults(v *viper.Viper) {
	if !v.GetBool("enable_https") {
		return
	}
	if v.GetString("tls_cert_file") == "" {
		v.Set("tls_cert_file", "cert.pem")
	}
	if v.GetString("tls_key_file") == "" {
		v.Set("tls_key_file", "key.pem")
	}
}

func validateURLs(cfg *Config) error {
	if cfg.BaseShortURL != "" {
		parsed, err := url.Parse(cfg.BaseShortURL)
		if err != nil {
			return fmt.Errorf("invalid base shorter URL: %w", err)
		}
		cfg.BaseShortURL = strings.TrimSuffix(cfg.BaseShortURL, "/")
		if parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("invalid base shorter URL")
		}
	}
	if cfg.AuditURL != "" {
		parsed, err := url.Parse(cfg.AuditURL)
		if err != nil {
			return fmt.Errorf("invalid audit URL: %w", err)
		}
		if parsed.Scheme == "" || parsed.Host == "" {
			return fmt.Errorf("invalid audit URL")
		}
	}
	return nil
}
