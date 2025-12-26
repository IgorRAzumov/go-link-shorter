package config

import (
	"flag"
	"os"
	"testing"
)

func TestLoad_DefaultValues(t *testing.T) {
	testEnv := setupTestEnv("", "")
	defer testEnv.restore()

	cfg := loadConfigOrFail(t)
	assertConfig(t, cfg, "localhost:8080", "")
}

func TestLoad_ServerAddressFromEnv(t *testing.T) {
	testEnv := setupTestEnv("0.0.0.0:9090", "")
	defer testEnv.restore()

	cfg := loadConfigOrFail(t)
	assertConfig(t, cfg, "0.0.0.0:9090", "")
}

func TestLoad_BaseURLFromEnv(t *testing.T) {
	testEnv := setupTestEnv("", "http://example.com")
	defer testEnv.restore()

	cfg := loadConfigOrFail(t)
	assertConfig(t, cfg, "localhost:8080", "http://example.com")
}

func TestLoad_BothFromEnv(t *testing.T) {
	testEnv := setupTestEnv("127.0.0.1:3000", "https://myserver.com")
	defer testEnv.restore()

	cfg := loadConfigOrFail(t)
	assertConfig(t, cfg, "127.0.0.1:3000", "https://myserver.com")
}

func TestLoad_EnvOverridesFlag(t *testing.T) {
	testEnv, restoreFlags := setupTestEnvWithFlags("env-server:8080", "http://env-base.com", []string{"test", "-a", "flag-server:9090", "-b", "http://flag-base.com"})
	defer testEnv.restore()
	defer restoreFlags()

	cfg := loadConfigOrFail(t)
	assertConfig(t, cfg, "env-server:8080", "http://env-base.com")
}

func TestLoad_FlagOverridesDefault(t *testing.T) {
	testEnv, restoreFlags := setupTestEnvWithFlags("", "", []string{"test", "-a", "flag-server:7777", "-b", "http://flag-base.com"})
	defer testEnv.restore()
	defer restoreFlags()

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
			testEnv := setupTestEnv("", tc.baseURL)
			defer testEnv.restore()

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
			testEnv := setupTestEnv("", tc.baseURL)
			defer testEnv.restore()

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
	testEnv := setupTestEnv("", "http://example.com/")
	defer testEnv.restore()

	cfg := loadConfigOrFail(t)
	assertConfig(t, cfg, "localhost:8080", "http://example.com")
}

func TestLoad_MixedEnvAndFlag(t *testing.T) {
	testEnv, restoreFlags := setupTestEnvWithFlags("env-server:8080", "", []string{"test", "-b", "http://flag-base.com"})
	defer testEnv.restore()
	defer restoreFlags()

	cfg := loadConfigOrFail(t)
	assertConfig(t, cfg, "env-server:8080", "http://flag-base.com")
}

type testEnv struct {
	oldServerAddress string
	oldBaseURL       string
	oldArgs          []string
}

func (testEnv *testEnv) restore() {
	if testEnv.oldServerAddress != "" {
		_ = os.Setenv("SERVER_ADDRESS", testEnv.oldServerAddress)
	} else {
		_ = os.Unsetenv("SERVER_ADDRESS")
	}

	if testEnv.oldBaseURL != "" {
		_ = os.Setenv("BASE_URL", testEnv.oldBaseURL)
	} else {
		_ = os.Unsetenv("BASE_URL")
	}

	os.Args = testEnv.oldArgs
}

func setupTestEnv(serverAddress, baseURL string) *testEnv {
	testEnv := &testEnv{
		oldServerAddress: os.Getenv("SERVER_ADDRESS"),
		oldBaseURL:       os.Getenv("BASE_URL"),
		oldArgs:          os.Args,
	}

	if serverAddress != "" {
		_ = os.Setenv("SERVER_ADDRESS", serverAddress)
	} else {
		_ = os.Unsetenv("SERVER_ADDRESS")
	}

	if baseURL != "" {
		_ = os.Setenv("BASE_URL", baseURL)
	} else {
		_ = os.Unsetenv("BASE_URL")
	}

	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	return testEnv
}

func setupTestFlags(args []string) func() {
	oldArgs := os.Args
	os.Args = args
	return func() {
		os.Args = oldArgs
	}
}

func setupTestEnvWithFlags(serverAddress, baseURL string, flagArgs []string) (*testEnv, func()) {
	testEnv := setupTestEnv(serverAddress, baseURL)
	restoreFlags := setupTestFlags(flagArgs)
	return testEnv, restoreFlags
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
