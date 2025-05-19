package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestValidateJSON_RejectsMalformed(t *testing.T) {
	h := ValidateJSON(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader("{bad json}"))
	h.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestValidateJSON_AllowsWellFormed(t *testing.T) {
	h := ValidateJSON(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", strings.NewReader("{\"foo\":123}"))
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
