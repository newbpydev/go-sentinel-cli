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

	// Get the executable directory to resolve relative paths
	execDir, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}
	baseDir := filepath.Dir(execDir)

	// Resolve paths relative to the executable
	resolvedTemplatesDir := filepath.Join(baseDir, *templatesDir)
	resolvedStaticDir := filepath.Join(baseDir, *staticDir)

	// Initialize and start the web server
	srv, err := server.NewServer(resolvedTemplatesDir, resolvedStaticDir)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("Go Sentinel Web Server starting on %s\n", *addr)
	log.Fatal(srv.Start(*addr))
}
