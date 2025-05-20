// Package main provides the entry point for the Go Sentinel web server.
// It initializes and starts the web server with all necessary configurations.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/newbpydev/go-sentinel/internal/web/server"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using default values")
	}

	// Define command-line flags with environment variable fallbacks
	webPort := flag.String("web-port", getEnvWithDefault("WEB_PORT", "3000"), "Web server port")
	templatesDir := flag.String("templates", "./web/templates", "Path to templates directory")
	staticDir := flag.String("static", "./web/static", "Path to static files directory")
	flag.Parse()

	// In development, use paths relative to the current working directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// Resolve paths relative to current directory
	resolvedTemplatesDir := filepath.Join(currentDir, *templatesDir)
	resolvedStaticDir := filepath.Join(currentDir, *staticDir)

	log.Printf("Templates directory: %s", resolvedTemplatesDir)
	log.Printf("Static files directory: %s", resolvedStaticDir)

	// Initialize and start the web server
	srv, err := server.NewServer(resolvedTemplatesDir, resolvedStaticDir)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Format the web port with a colon prefix if not provided
	formattedWebPort := *webPort
	if formattedWebPort[0] != ':' {
		formattedWebPort = ":" + formattedWebPort
	}

	fmt.Printf("Go Sentinel Web Server starting on %s\n", formattedWebPort)
	log.Fatal(srv.Start(formattedWebPort))
}

// getEnvWithDefault returns the value of an environment variable or a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
