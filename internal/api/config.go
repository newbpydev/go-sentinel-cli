// Package api provides the core API functionality for the Go Sentinel service.
package api

import (
	"os"
	"strconv"
)

// Config holds API settings for the server, middleware, etc.
type Config struct {
	Port         string
	ReadTimeout  int
	WriteTimeout int
	Env          string
}

// getEnvWithDefault returns the value of an environment variable or a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsIntWithDefault returns the value of an environment variable as an integer or a default value
func getEnvAsIntWithDefault(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// NewConfig returns a Config struct initialized from environment variables or defaults.
// It uses the following environment variables:
//   - API_PORT: The port to run the API server on (default: "8080")
//   - API_READ_TIMEOUT: Read timeout in seconds (default: 10)
//   - API_WRITE_TIMEOUT: Write timeout in seconds (default: 10)
//   - API_ENV: Environment name (default: "development")
func NewConfig() Config {
	return Config{
		Port:         getEnvWithDefault("API_PORT", "8080"),
		ReadTimeout:  getEnvAsIntWithDefault("API_READ_TIMEOUT", 10),
		WriteTimeout: getEnvAsIntWithDefault("API_WRITE_TIMEOUT", 10),
		Env:          getEnvWithDefault("API_ENV", "development"),
	}
}
