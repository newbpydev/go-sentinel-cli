package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	api "github.com/newbpydev/go-sentinel/internal/api"
)

func getTestConfig() api.Config {
	return api.Config{
		Port:         "0",
		ReadTimeout:  10,
		WriteTimeout: 10,
		Env:          "test",
	}
}

func setupTestServer() http.Handler {
	cfg := getTestConfig()
	srv := NewAPIServer(cfg)
	return srv.Router
}

func BenchmarkAPIHealthEndpoint(b *testing.B) {
	h := setupTestServer()
	req := httptest.NewRequest("GET", "/health", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
	}
}

func BenchmarkAPIConfigPost(b *testing.B) {
	h := setupTestServer()
	body := `{"key":"value"}`
	req := httptest.NewRequest("POST", "/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
	}
}

func BenchmarkAPIHealthEndpointMemory(b *testing.B) {
	b.ReportAllocs()
	h := setupTestServer()
	req := httptest.NewRequest("GET", "/health", nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)
	}
}
