package main

import (
	"log"

	"github.com/newbpydev/go-sentinel/internal/api"
	"github.com/newbpydev/go-sentinel/internal/api/server"
)

func main() {
	// Load config from environment or defaults
	cfg := api.NewConfig()

	// Create API server instance
	srv := server.NewAPIServer(cfg)

	log.Printf("Starting Go-Sentinel API server on port %s", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatalf("API server exited with error: %v", err)
	}
}
