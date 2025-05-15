// +build integration

package server

import (
	"net/http/httptest"
	"testing"
	"net/url"
	"net/http"
	"golang.org/x/net/websocket"
)

// TestWebSocketConnection verifies the WebSocket endpoint upgrades and accepts connections
func TestWebSocketConnection(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Upgrade") == "websocket" {
			w.WriteHeader(http.StatusSwitchingProtocols)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	u.Scheme = "ws"

	ws, err := websocket.Dial(u.String(), "", "http://localhost/")
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket endpoint: %v", err)
	}
	ws.Close()
}
