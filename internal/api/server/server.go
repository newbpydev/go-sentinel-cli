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
)

// Server wraps the HTTP server and its dependencies
// Config is imported from the parent api package
import api "github.com/newbpydev/go-sentinel/internal/api"

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
	// TODO: Add CORS and security middleware here if needed

	// Health endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	httpSrv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
	}

	return &APIServer{
		Config: cfg,
		Router: r,
		HTTP:   httpSrv,
	}
}

func (s *APIServer) Start() error {
	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.HTTP.Shutdown(ctx); err != nil {
			log.Printf("API server forced to shutdown: %v", err)
		}
	}()

	fmt.Printf("[api] Listening on %s\n", s.HTTP.Addr)
	return s.HTTP.ListenAndServe()
}
