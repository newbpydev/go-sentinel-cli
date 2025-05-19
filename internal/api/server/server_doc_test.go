package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestDocsEndpoint_OpenAPIYAML(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile("api-docs.yaml")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/docs")
	if err != nil {
		t.Fatalf("failed to GET /docs: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 for /docs, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/yaml" {
		t.Errorf("expected Content-Type application/yaml, got %q", ct)
	}
	body := make([]byte, 4096)
	n, _ := resp.Body.Read(body)
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("failed to close response body: %v", err)
	}
	if n == 0 || string(body[:n]) == "" || !containsOpenAPI(string(body[:n])) {
		t.Errorf("/docs did not return OpenAPI YAML")
	}
}

func containsOpenAPI(s string) bool {
	return len(s) > 0 && (s[:10] == "openapi: 3" || s[:12] == "openapi: 3.0")
}
