package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/newbpydev/go-sentinel/internal/tui"
)

func main() {
	// Parse command line flags
	rootPath := flag.String("path", ".", "Root path to watch for changes")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Get absolute path for the root directory
	absPath, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}
	
	// If root path was specified, use it instead of current directory
	if *rootPath != "." {
		absPath = *rootPath
	}

	// Configure logger
	if *verbose {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	} else {
		// Minimal logging in normal mode
		log.SetFlags(log.Ltime)
	}

	log.Printf("Starting Go-Sentinel in directory: %s", absPath)

	// Create and start the TUI application
	app, err := tui.NewApp(absPath)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the application in a goroutine
	go func() {
		if err := app.Start(); err != nil {
			log.Fatalf("Error running application: %v", err)
		}
	}()

	// Wait for termination signal
	sig := <-sigChan
	log.Printf("Received signal %v, shutting down...", sig)

	// Perform graceful shutdown
	app.Stop()
	log.Println("Shutdown complete")
}
