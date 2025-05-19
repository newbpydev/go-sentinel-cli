package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

func TestServer_InitializesWithProperRoutes(t *testing.T) {
	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("failed to send GET /health: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}
}

func TestServer_AppliesMiddlewareChain(t *testing.T) {
	called := false
	mw := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			next.ServeHTTP(w, r)
		})
	}

	r := chi.NewRouter()
	r.Use(mw)
	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ts := httptest.NewServer(r)
	defer ts.Close()

	_, err := http.Get(ts.URL + "/test")
	if err != nil {
		t.Fatalf("failed to send GET /test: %v", err)
	}
	if !called {
		t.Error("middleware was not called")
	}
}

func TestServer_GracefulShutdown(t *testing.T) {
	r := chi.NewRouter()
	srv := &http.Server{
		Addr:    "127.0.0.1:0",
		Handler: r,
	}

go func() {
		time.Sleep(100 * time.Millisecond)
		if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("server shutdown failed: %v", err)
	}
	}()

	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		t.Fatalf("unexpected server error: %v", err)
	}
}
