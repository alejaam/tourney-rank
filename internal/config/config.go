// Package config provides configuration management for the application.package config

package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration.
type Config struct {
	// Server configuration
	HTTPPort string
	WSPort   string

	// Database configuration
	DatabaseURL string

	// Redis configuration
	RedisURL string

	// Application settings
	Environment     string
	LogLevel        string
	ShutdownTimeout time.Duration

	// Feature flags
	EnableMetrics bool
	EnableTracing bool
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		// Server defaults
		HTTPPort: getEnv("HTTP_PORT", "8080"),
		WSPort:   getEnv("WS_PORT", "8081"),

		// Database defaults (empty means not configured)
		DatabaseURL: getEnv("DATABASE_URL", ""),
		RedisURL:    getEnv("REDIS_URL", ""),

		// Application defaults
		Environment:     getEnv("ENVIRONMENT", "development"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 15*time.Second),

		// Feature flags
		EnableMetrics: getBoolEnv("ENABLE_METRICS", false),
		EnableTracing: getBoolEnv("ENABLE_TRACING", false),
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.HTTPPort == "" {
		return fmt.Errorf("HTTP_PORT is required")
	}

	// Validate port is numeric
	if _, err := strconv.Atoi(c.HTTPPort); err != nil {
		return fmt.Errorf("HTTP_PORT must be a valid port number: %w", err)
	}

	return nil
}

// IsDevelopment returns true if running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode.
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// HTTPAddr returns the HTTP server address.
func (c *Config) HTTPAddr() string {
	return ":" + c.HTTPPort
}

// WSAddr returns the WebSocket server address.
func (c *Config) WSAddr() string {
	return ":" + c.WSPort
}

// getEnv retrieves an environment variable with a fallback default value.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getBoolEnv retrieves a boolean environment variable.
func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}

// getDurationEnv retrieves a duration environment variable.
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}

// MustGetEnv retrieves an environment variable or panics if not set.
func MustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("environment variable %s is required", key))
	}
	return value
}
