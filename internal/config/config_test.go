package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	t.Run("loads defaults when no env vars set", func(t *testing.T) {
		// Clear relevant env vars
		os.Unsetenv("HTTP_PORT")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("ENVIRONMENT")

		cfg, err := Load()

		require.NoError(t, err)
		assert.Equal(t, "8080", cfg.HTTPPort)
		assert.Equal(t, "8081", cfg.WSPort)
		assert.Equal(t, "development", cfg.Environment)
		assert.Equal(t, "info", cfg.LogLevel)
		assert.Equal(t, 15*time.Second, cfg.ShutdownTimeout)
	})

	t.Run("loads from environment variables", func(t *testing.T) {
		os.Setenv("HTTP_PORT", "3000")
		os.Setenv("DATABASE_URL", "postgres://localhost/test")
		os.Setenv("ENVIRONMENT", "production")
		defer func() {
			os.Unsetenv("HTTP_PORT")
			os.Unsetenv("DATABASE_URL")
			os.Unsetenv("ENVIRONMENT")
		}()

		cfg, err := Load()

		require.NoError(t, err)
		assert.Equal(t, "3000", cfg.HTTPPort)
		assert.Equal(t, "postgres://localhost/test", cfg.DatabaseURL)
		assert.Equal(t, "production", cfg.Environment)
	})

	t.Run("returns error for invalid HTTP_PORT", func(t *testing.T) {
		os.Setenv("HTTP_PORT", "invalid")
		defer os.Unsetenv("HTTP_PORT")

		_, err := Load()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "HTTP_PORT must be a valid port number")
	})
}

func TestConfig_IsDevelopment(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		want        bool
	}{
		{"development env", "development", true},
		{"production env", "production", false},
		{"staging env", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.environment}
			assert.Equal(t, tt.want, cfg.IsDevelopment())
		})
	}
}

func TestConfig_IsProduction(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		want        bool
	}{
		{"production env", "production", true},
		{"development env", "development", false},
		{"staging env", "staging", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Environment: tt.environment}
			assert.Equal(t, tt.want, cfg.IsProduction())
		})
	}
}

func TestConfig_HTTPAddr(t *testing.T) {
	cfg := &Config{HTTPPort: "8080"}
	assert.Equal(t, ":8080", cfg.HTTPAddr())
}

func TestGetBoolEnv(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue bool
		want         bool
	}{
		{"true string", "true", false, true},
		{"false string", "false", true, false},
		{"1 string", "1", false, true},
		{"0 string", "0", true, false},
		{"empty uses default true", "", true, true},
		{"empty uses default false", "", false, false},
		{"invalid uses default", "invalid", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_BOOL_VAR"
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
				defer os.Unsetenv(key)
			} else {
				os.Unsetenv(key)
			}

			got := getBoolEnv(key, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetDurationEnv(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue time.Duration
		want         time.Duration
	}{
		{"valid duration", "30s", 10 * time.Second, 30 * time.Second},
		{"valid minutes", "5m", 10 * time.Second, 5 * time.Minute},
		{"empty uses default", "", 10 * time.Second, 10 * time.Second},
		{"invalid uses default", "invalid", 10 * time.Second, 10 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_DURATION_VAR"
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
				defer os.Unsetenv(key)
			} else {
				os.Unsetenv(key)
			}

			got := getDurationEnv(key, tt.defaultValue)
			assert.Equal(t, tt.want, got)
		})
	}
}
