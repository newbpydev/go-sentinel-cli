package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	custommiddleware "github.com/newbpydev/go-sentinel/internal/api/middleware"

	"github.com/newbpydev/go-sentinel/internal/api"
	"github.com/newbpydev/go-sentinel/internal/api/metrics"
	"github.com/newbpydev/go-sentinel/internal/api/websocket"
)

// Response cache for frequently requested endpoints
var healthCache = api.NewResultCache(1)
var healthCacheExpiry time.Time

type APIServer struct {
	Config api.Config
	Router *chi.Mux
	HTTP   *http.Server
}

func NewAPIServer(cfg api.Config) *APIServer {
	r := chi.NewRouter()

	// Standard middleware chain
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	// Security middleware
	r.Use(custommiddleware.RateLimit(60, time.Minute)) // 60 requests per minute per IP
	r.Use(custommiddleware.ValidateJSON)

	// Metrics endpoint
	r.Get("/metrics", metrics.Handler)

	// Docs endpoint (stub for now)
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
		if _, err := w.Write([]byte("OpenAPI documentation coming soon")); err != nil {
			log.Printf("failed to write docs stub: %v", err)
		}
	})

	// Health endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		cacheKey := "health_ok"
		var resp []byte
		var ok bool
		if time.Now().Before(healthCacheExpiry) {
			if val, found := healthCache.Get(cacheKey); found {
				resp, ok = val.([]byte)
			}
		}
		if !ok || resp == nil {
			resp = []byte("ok")
			healthCache.Set(cacheKey, resp)
			healthCacheExpiry = time.Now().Add(10 * time.Second)
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(resp); err != nil {
			log.Printf("failed to write health response: %v", err)
		}
	})

	// WebSocket endpoint
	// Create WebSocket connection manager
	connManager := websocket.NewConnectionManager()

	// Register WebSocket endpoint
	r.Get("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Use the connection handler from the websocket package
		HandleWebSocketConnection(w, r, connManager)
	})

	log.Println("WebSocket handler registered at /ws endpoint")

	return &APIServer{
		Config: cfg,
		Router: r,
		HTTP: &http.Server{
			Addr:              cfg.Port,
			Handler:           r,
			ReadTimeout:       15 * time.Second,
			ReadHeaderTimeout: 15 * time.Second,
			WriteTimeout:      15 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
	}
}

// Start begins listening on the configured port and handles graceful shutdown
func (s *APIServer) Start() error {
	// Create a channel to listen for errors coming from the listener.
	serverErrors := make(chan error, 1)

	// Create a channel to listen for an interrupt or terminate signal from the OS.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start the service listening for requests.
	go func() {
		log.Printf("API listening on %s", s.HTTP.Addr)
		serverErrors <- s.HTTP.ListenAndServe()
	}()

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Printf("main: %v: Start shutdown", sig)

		// Give outstanding requests a deadline for completion.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Asking listener to shut down and shed load.
		if err := s.HTTP.Shutdown(ctx); err != nil {
			// Error from closing listeners, or context timeout:
			s.HTTP.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
