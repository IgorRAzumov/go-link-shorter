package config

import (
	"flag"
	"os"
	"testing"
)

func TestLoad_DefaultValues(t *testing.T) {
	setupTestEnv(t, "", "")

	cfg := loadConfigOrFail(t)
	assertConfig(t, cfg, "localhost:8080", "")
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
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
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
