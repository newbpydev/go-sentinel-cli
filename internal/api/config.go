package api

import (
	"os"
)

// Config holds API settings for the server, middleware, etc.
type Config struct {
	Port         string
	ReadTimeout  int
	WriteTimeout int
	Env          string
}

// NewConfig returns a Config struct initialized from environment variables or defaults.
func NewConfig() Config {
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}
	return Config{
		Port:         port,
		ReadTimeout:  10, // seconds
		WriteTimeout: 10, // seconds
		Env:          os.Getenv("API_ENV"),
	}
}
