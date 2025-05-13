package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/newbpydev/go-sentinel/internal/web/server"
)

func main() {
	// Define command-line flags
	addr := flag.String("addr", ":8080", "HTTP server address")
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

	fmt.Printf("Go Sentinel Web Server starting on %s\n", *addr)
	log.Fatal(srv.Start(*addr))
}
