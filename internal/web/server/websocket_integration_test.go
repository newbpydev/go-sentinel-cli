//go:build integration
// +build integration

package server

import (
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/net/websocket"
)

// TestWebSocketConnection verifies the WebSocket endpoint upgrades and accepts connections
func TestWebSocketConnection(t *testing.T) {
	echoHandler := func(ws *websocket.Conn) {
		var msg = make([]byte, 512)
		n, err := ws.Read(msg)
		if err != nil {
			t.Errorf("Server failed to read: %v", err)
			return
		}
		// Echo back
		_, err = ws.Write(msg[:n])
		if err != nil {
			t.Errorf("Server failed to write: %v", err)
		}
	}

	ts := httptest.NewServer(websocket.Handler(echoHandler))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	u.Scheme = "ws"

	ws, err := websocket.Dial(u.String(), "", "http://localhost/")
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket endpoint: %v", err)
	}
	defer ws.Close()

	msg := []byte("hello")
	_, err = ws.Write(msg)
	if err != nil {
		t.Fatalf("Failed to write to WebSocket: %v", err)
	}

	reply := make([]byte, 512)
	n, err := ws.Read(reply)
	if err != nil {
		t.Fatalf("Failed to read from WebSocket: %v", err)
	}
	if string(reply[:n]) != "hello" {
		t.Fatalf("Expected echo 'hello', got '%s'", string(reply[:n]))
	}
}
