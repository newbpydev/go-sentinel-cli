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

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
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
		w.Write([]byte("OpenAPI documentation coming soon"))
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
		w.Write(resp)
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
