package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
)

func TestLoad_DefaultValues(t *testing.T) {
	setupTestEnv(t, "", "")

	cfg := loadConfigOrFail(t)
	assertConfig(t, cfg, "localhost:8080", "")
	if cfg.GRPCAddress != "localhost:9090" {
		t.Errorf("Expected default GRPCAddress localhost:9090, got %q", cfg.GRPCAddress)
	}
}

func TestLoad_GRPCAddressFromEnv(t *testing.T) {
	setupTestEnv(t, "", "")
	t.Setenv("GRPC_ADDRESS", "127.0.0.1:50051")

	cfg := loadConfigOrFail(t)
	if cfg.GRPCAddress != "127.0.0.1:50051" {
		t.Errorf("Expected GRPCAddress from env, got %q", cfg.GRPCAddress)
	}
}

func TestLoad_GRPCAddressFromFlag(t *testing.T) {
	setupTestEnvWithFlags(t, "", "", []string{"test", "-g", "0.0.0.0:7777"})

	cfg := loadConfigOrFail(t)
	if cfg.GRPCAddress != "0.0.0.0:7777" {
		t.Errorf("Expected GRPCAddress from flag, got %q", cfg.GRPCAddress)
	}
}

func TestLoad_ServerAddressFromEnv(t *testing.T) {
	setupTestEnv(t, "0.0.0.0:9090", "")

	cfg := loadConfigOrFail(t)
	assertConfig(t, cfg, "0.0.0.0:9090", "")
}

func TestLoad_BaseURLFromEnv(t *testing.T) {
	setupTestEnv(t, "", "http://example.com")

	cfg := loadConfigOrFail(t)
	assertConfig(t, cfg, "localhost:8080", "http://example.com")
}

func TestLoad_BothFromEnv(t *testing.T) {
	setupTestEnv(t, "127.0.0.1:3000", "https://myserver.com")

	cfg := loadConfigOrFail(t)
	assertConfig(t, cfg, "127.0.0.1:3000", "https://myserver.com")
}

func TestLoad_EnvOverridesFlag(t *testing.T) {
	setupTestEnvWithFlags(t, "env-server:8080", "http://env-base.com", []string{"test", "-a", "flag-server:9090", "-b", "http://flag-base.com"})

	cfg := loadConfigOrFail(t)
	assertConfig(t, cfg, "env-server:8080", "http://env-base.com")
}

func TestLoad_FlagOverridesDefault(t *testing.T) {
	setupTestEnvWithFlags(t, "", "", []string{"test", "-a", "flag-server:7777", "-b", "http://flag-base.com"})

	cfg := loadConfigOrFail(t)
	assertConfig(t, cfg, "flag-server:7777", "http://flag-base.com")
}

func TestLoad_BaseURLValidation_ValidURL(t *testing.T) {
	testCases := []struct {
		name    string
		baseURL string
	}{
		{"HTTP URL", "http://example.com"},
		{"HTTPS URL", "https://example.com"},
		{"URL with port", "http://example.com:8080"},
		{"URL with path", "https://example.com/api"},
		{"URL with trailing slash", "http://example.com/"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupTestEnv(t, "", tc.baseURL)

			cfg := loadConfigOrFail(t)

			expected := tc.baseURL
			if expected != "" && expected[len(expected)-1] == '/' {
				expected = expected[:len(expected)-1]
			}

			if cfg.BaseShortURL != expected {
				t.Errorf("Expected BaseShortURL '%s', got '%s'", expected, cfg.BaseShortURL)
			}
		})
	}
}

func TestLoad_BaseURLValidation_InvalidURL(t *testing.T) {
	testCases := []struct {
		name    string
		baseURL string
	}{
		{"Missing scheme", "example.com"},
		{"Missing host", "http://"},
		{"Invalid format", "not-a-url"},
		{"Empty string", ""},
		{"Only scheme", "http://"},
		{"Only path", "/path"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			setupTestEnv(t, "", tc.baseURL)

			cfg, err := Load()

			if tc.baseURL == "" {
				if err != nil {
					t.Fatalf("Load() should not return error for empty BASE_URL, got: %v", err)
				}
				if cfg.BaseShortURL != "" {
					t.Errorf("Expected empty BaseShortURL, got '%s'", cfg.BaseShortURL)
				}
			} else {
				if err == nil {
					t.Errorf("Load() should return error for invalid URL '%s', got nil", tc.baseURL)
				}
			}
		})
	}
}

func TestLoad_BaseURLTrailingSlashRemoved(t *testing.T) {
	setupTestEnv(t, "", "http://example.com/")

	cfg := loadConfigOrFail(t)
	assertConfig(t, cfg, "localhost:8080", "http://example.com")
}

func TestLoad_MixedEnvAndFlag(t *testing.T) {
	setupTestEnvWithFlags(t, "env-server:8080", "", []string{"test", "-b", "http://flag-base.com"})

	cfg := loadConfigOrFail(t)
	assertConfig(t, cfg, "env-server:8080", "http://flag-base.com")
}

func setupTestEnv(t *testing.T, serverAddress, baseURL string) {
	t.Helper()
	t.Setenv("SERVER_ADDRESS", serverAddress)
	t.Setenv("BASE_URL", baseURL)
	t.Setenv("AUDIT_FILE", "")
	t.Setenv("AUDIT_URL", "")
	t.Setenv("GRPC_ADDRESS", "")
	t.Setenv("CONFIG", "")
	pflag.CommandLine = pflag.NewFlagSet("test", pflag.ContinueOnError)
	oldArgs := os.Args
	os.Args = []string{"test"}
	t.Cleanup(func() { os.Args = oldArgs })
}

func setupTestEnvWithFlags(t *testing.T, serverAddress, baseURL string, flagArgs []string) {
	t.Helper()
	setupTestEnv(t, serverAddress, baseURL)
	oldArgs := os.Args
	os.Args = flagArgs
	t.Cleanup(func() { os.Args = oldArgs })
}

func loadConfigOrFail(t *testing.T) *Config {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}
	return cfg
}

func assertConfig(t *testing.T, cfg *Config, expectedServerAddress, expectedBaseURL string) {
	if cfg.ServerAddress != expectedServerAddress {
		t.Errorf("Expected ServerAddress '%s', got '%s'", expectedServerAddress, cfg.ServerAddress)
	}
	if cfg.BaseShortURL != expectedBaseURL {
		t.Errorf("Expected BaseShortURL '%s', got '%s'", expectedBaseURL, cfg.BaseShortURL)
	}
}

func TestLoad_EnableHTTPS_FromEnv(t *testing.T) {
	setupTestEnv(t, "", "")
	t.Setenv("ENABLE_HTTPS", "true")

	cfg := loadConfigOrFail(t)

	if !cfg.EnableHTTPS {
		t.Error("Expected EnableHTTPS true from ENABLE_HTTPS env, got false")
	}
	if cfg.TLSCertFile != "cert.pem" {
		t.Errorf("Expected default TLSCertFile cert.pem when HTTPS enabled, got %q", cfg.TLSCertFile)
	}
	if cfg.TLSKeyFile != "key.pem" {
		t.Errorf("Expected default TLSKeyFile key.pem when HTTPS enabled, got %q", cfg.TLSKeyFile)
	}
}

func TestLoad_EnableHTTPS_FromFlag(t *testing.T) {
	setupTestEnvWithFlags(t, "", "", []string{"test", "-s"})

	cfg := loadConfigOrFail(t)

	if !cfg.EnableHTTPS {
		t.Error("Expected EnableHTTPS true from -s flag, got false")
	}
	if cfg.TLSCertFile != "cert.pem" {
		t.Errorf("Expected default TLSCertFile cert.pem when HTTPS enabled, got %q", cfg.TLSCertFile)
	}
	if cfg.TLSKeyFile != "key.pem" {
		t.Errorf("Expected default TLSKeyFile key.pem when HTTPS enabled, got %q", cfg.TLSKeyFile)
	}
}

func TestLoad_EnableHTTPS_DefaultsToFalse(t *testing.T) {
	setupTestEnv(t, "", "")

	cfg := loadConfigOrFail(t)

	if cfg.EnableHTTPS {
		t.Error("Expected EnableHTTPS false by default, got true")
	}
}

func TestLoad_TLSCertKey_FromFlags(t *testing.T) {
	setupTestEnvWithFlags(t, "", "", []string{"test", "-s", "--tls-cert", "/path/to/cert.pem", "--tls-key", "/path/to/key.pem"})

	cfg := loadConfigOrFail(t)

	if !cfg.EnableHTTPS {
		t.Error("Expected EnableHTTPS true, got false")
	}
	if cfg.TLSCertFile != "/path/to/cert.pem" {
		t.Errorf("Expected TLSCertFile /path/to/cert.pem, got %q", cfg.TLSCertFile)
	}
	if cfg.TLSKeyFile != "/path/to/key.pem" {
		t.Errorf("Expected TLSKeyFile /path/to/key.pem, got %q", cfg.TLSKeyFile)
	}
}

func TestLoad_TLSCertKey_FromEnv(t *testing.T) {
	setupTestEnv(t, "", "")
	t.Setenv("ENABLE_HTTPS", "true")
	t.Setenv("TLS_CERT_FILE", "/env/cert.pem")
	t.Setenv("TLS_KEY_FILE", "/env/key.pem")

	cfg := loadConfigOrFail(t)

	if cfg.TLSCertFile != "/env/cert.pem" {
		t.Errorf("Expected TLSCertFile from env /env/cert.pem, got %q", cfg.TLSCertFile)
	}
	if cfg.TLSKeyFile != "/env/key.pem" {
		t.Errorf("Expected TLSKeyFile from env /env/key.pem, got %q", cfg.TLSKeyFile)
	}
}

func TestLoad_FromConfigFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	configContent := `{
		"server_address": "config-server:9000",
		"base_url": "http://config-base.com",
		"file_storage_path": "/config/file.db",
		"database_dsn": "postgres://config/db",
		"enable_https": true
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	setupTestEnv(t, "", "")
	t.Setenv("CONFIG", configPath)

	cfg := loadConfigOrFail(t)

	if cfg.ServerAddress != "config-server:9000" {
		t.Errorf("Expected ServerAddress from config, got %q", cfg.ServerAddress)
	}
	if cfg.BaseShortURL != "http://config-base.com" {
		t.Errorf("Expected BaseShortURL from config, got %q", cfg.BaseShortURL)
	}
	if cfg.FileStoragePath != "/config/file.db" {
		t.Errorf("Expected FileStoragePath from config, got %q", cfg.FileStoragePath)
	}
	if cfg.DatabaseAddress != "postgres://config/db" {
		t.Errorf("Expected DatabaseAddress from config, got %q", cfg.DatabaseAddress)
	}
	if !cfg.EnableHTTPS {
		t.Error("Expected EnableHTTPS true from config")
	}
}

func TestLoad_ConfigFileOverriddenByEnv(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	configContent := `{"server_address": "config-server:9000", "base_url": "http://config-base.com"}`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	setupTestEnv(t, "env-server:8080", "http://env-base.com")
	t.Setenv("CONFIG", configPath)
	os.Args = []string{"test"}

	cfg := loadConfigOrFail(t)

	if cfg.ServerAddress != "env-server:8080" {
		t.Errorf("Expected env to override config ServerAddress, got %q", cfg.ServerAddress)
	}
	if cfg.BaseShortURL != "http://env-base.com" {
		t.Errorf("Expected env to override config BaseShortURL, got %q", cfg.BaseShortURL)
	}
}

func TestLoad_ConfigFileOverriddenByFlag(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	configContent := `{"server_address": "config-server:9000", "base_url": "http://config-base.com"}`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	setupTestEnv(t, "", "")
	t.Setenv("CONFIG", configPath)
	os.Args = []string{"test", "-a", "flag-server:7777", "-b", "http://flag-base.com"}

	cfg := loadConfigOrFail(t)

	if cfg.ServerAddress != "flag-server:7777" {
		t.Errorf("Expected flag to override config ServerAddress, got %q", cfg.ServerAddress)
	}
	if cfg.BaseShortURL != "http://flag-base.com" {
		t.Errorf("Expected flag to override config BaseShortURL, got %q", cfg.BaseShortURL)
	}
}

func TestLoad_ConfigFileOverriddenByEnv_StorageAndDB(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	configContent := `{
		"file_storage_path": "/config/file.db",
		"database_dsn": "postgres://config/db"
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	t.Setenv("CONFIG", configPath)
	t.Setenv("FILE_STORAGE_PATH", "/env/file.db")
	t.Setenv("DATABASE_DSN", "postgres://env/db")
	t.Setenv("AUDIT_FILE", "")
	t.Setenv("AUDIT_URL", "")

	pflag.CommandLine = pflag.NewFlagSet("test", pflag.ContinueOnError)
	os.Args = []string{"test"}

	cfg := loadConfigOrFail(t)

	if cfg.FileStoragePath != "/env/file.db" {
		t.Errorf("Expected env to override config FileStoragePath, got %q", cfg.FileStoragePath)
	}
	if cfg.DatabaseAddress != "postgres://env/db" {
		t.Errorf("Expected env to override config DatabaseAddress, got %q", cfg.DatabaseAddress)
	}
}

func TestLoad_ConfigFileOverriddenByFlag_StorageAndDB(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	configContent := `{
		"file_storage_path": "/config/file.db",
		"database_dsn": "postgres://config/db"
	}`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	t.Setenv("CONFIG", configPath)
	t.Setenv("AUDIT_FILE", "")
	t.Setenv("AUDIT_URL", "")
	_ = os.Unsetenv("FILE_STORAGE_PATH")
	_ = os.Unsetenv("DATABASE_DSN")

	pflag.CommandLine = pflag.NewFlagSet("test", pflag.ContinueOnError)
	os.Args = []string{"test", "-f", "/flag/file.db", "-d", "postgres://flag/db"}

	cfg := loadConfigOrFail(t)

	if cfg.FileStoragePath != "/flag/file.db" {
		t.Errorf("Expected flag to override config FileStoragePath, got %q", cfg.FileStoragePath)
	}
	if cfg.DatabaseAddress != "postgres://flag/db" {
		t.Errorf("Expected flag to override config DatabaseAddress, got %q", cfg.DatabaseAddress)
	}
}

func TestLoad_ConfigFileNotFound(t *testing.T) {
	setupTestEnv(t, "", "")
	t.Setenv("CONFIG", "/nonexistent/config.json")
	os.Args = []string{"test"}

	_, err := Load()
	if err == nil {
		t.Fatal("Load() should return error for nonexistent config file")
	}
}

func TestLoad_ConfigFileViaFlag(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	configContent := `{"server_address": "flag-config-server:8000", "base_url": "http://flag-config.com"}`
	if err := os.WriteFile(configPath, []byte(configContent), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	setupTestEnvWithFlags(t, "", "", []string{"test", "-c", configPath})

	cfg := loadConfigOrFail(t)

	if cfg.ServerAddress != "flag-config-server:8000" {
		t.Errorf("Expected ServerAddress from -c config file, got %q", cfg.ServerAddress)
	}
	if cfg.BaseShortURL != "http://flag-config.com" {
		t.Errorf("Expected BaseShortURL from -c config file, got %q", cfg.BaseShortURL)
	}
}
