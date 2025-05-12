package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimit_AllowsRequestsBelowLimit(t *testing.T) {
	h := RateLimit(3, time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	for i := 0; i < 3; i++ {
		h.ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	}
}

func TestRateLimit_BlocksRequestsAboveLimit(t *testing.T) {
	h := RateLimit(2, time.Second)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	h.ServeHTTP(w, r)
	h.ServeHTTP(w, r)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, r)
	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429, got %d", w2.Code)
	}
}

func TestRateLimit_ResetsAfterWindow(t *testing.T) {
	h := RateLimit(1, 50*time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	h.ServeHTTP(w, r)
	time.Sleep(60 * time.Millisecond)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, r)
	if w2.Code != http.StatusOK {
		t.Errorf("expected 200 after reset, got %d", w2.Code)
	}
}
