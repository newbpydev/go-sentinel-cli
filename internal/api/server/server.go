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
)

// Response cache for frequently requested endpoints
var healthCache = api.NewResultCache(1)
var healthCacheExpiry time.Time

// APIServer wraps the HTTP server and configuration for the Go-Sentinel API.
type APIServer struct {
	// Config holds the API server configuration.
	Config api.Config
	// Router is the Chi router instance.
	Router *chi.Mux
	// HTTP is the HTTP server instance.
	HTTP *http.Server
	status int
}

// statusRecorder wraps http.ResponseWriter to record the status code for metrics tracking.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader records the status code and delegates to the underlying ResponseWriter.
// WriteHeader records the status code and delegates to the underlying ResponseWriter.
func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// NewAPIServer creates and configures a new Go-Sentinel API server.
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
	//
	// Exposes Prometheus-style metrics about the process and API usage.
	r.Get("/metrics", metrics.Handler)

	// Docs endpoint: serve static OpenAPI YAML
	//
	// Serves the OpenAPI 3.0 YAML for the API at /docs.
	r.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile("internal/api/server/api-docs.yaml")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to load OpenAPI documentation"))
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	// Docs UI endpoint: serve Swagger UI static files at /docs/ui
	fileServer := http.StripPrefix("/docs/ui/", http.FileServer(http.Dir("internal/api/server/swagger-ui")))
	r.Get("/docs/ui/*", func(w http.ResponseWriter, r *http.Request) {
		fileServer.ServeHTTP(w, r)
	})

	// Health endpoint
	//
	// Returns a simple 'ok' if the service is healthy, with response caching.
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

	// Middleware to track request status
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &statusRecorder{ResponseWriter: w, status: 200}
			next.ServeHTTP(rw, r)
			metrics.Track(rw.status)
		})
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


// Start launches the API server and handles graceful shutdown.
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
