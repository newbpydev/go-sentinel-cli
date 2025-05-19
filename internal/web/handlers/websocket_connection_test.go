package handlers

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestConnectionTracking(t *testing.T) {
	h := NewWebSocketHandler()
	h.StartBroadcaster()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.HandleWebSocket(w, r)
	}))
	defer server.Close()

	// Test initial connection count
	if h.ConnectionCount() != 0 {
		t.Errorf("Expected 0 connections initially, got %d", h.ConnectionCount())
	}

	// Connect first client
	url := "ws" + server.URL[4:] + "/ws"
	conn1, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect first WebSocket: %v", err)
	}

	// Verify connection count
	time.Sleep(100 * time.Millisecond) // Give time for connection to be registered
	if count := h.ConnectionCount(); count != 1 {
		t.Errorf("Expected 1 connection after first connect, got %d", count)
	}

	// Connect second client
	conn2, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		conn1.Close()
		t.Fatalf("Failed to connect second WebSocket: %v", err)
	}

	// Verify connection count
	time.Sleep(100 * time.Millisecond)
	if count := h.ConnectionCount(); count != 2 {
		t.Errorf("Expected 2 connections after second connect, got %d", count)
	}

	// Disconnect first client
	if err := conn1.Close(); err != nil {
		t.Errorf("Failed to close connection: %v", err)
	}
	time.Sleep(100 * time.Millisecond) // Give time for cleanup

	// Verify connection count after disconnect
	if count := h.ConnectionCount(); count != 1 {
		t.Errorf("Expected 1 connection after first disconnect, got %d", count)
	}

	// Disconnect second client
	if err := conn2.Close(); err != nil {
		t.Errorf("Failed to close connection: %v", err)
	}
	time.Sleep(100 * time.Millisecond) // Give time for cleanup

	// Verify connection count after all disconnects
	if count := h.ConnectionCount(); count != 0 {
		t.Errorf("Expected 0 connections after all disconnects, got %d", count)
	}
}

func TestBroadcastToDisconnectedClient(t *testing.T) {
	h := NewWebSocketHandler()
	h.StartBroadcaster()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.HandleWebSocket(w, r)
	}))
	defer server.Close()

	// Connect client
	url := "ws" + server.URL[4:] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect WebSocket: %v", err)
	}

	// Close the connection immediately
	defer func() {
		if err := conn.Close(); err != nil {
			t.Errorf("Failed to close connection: %v", err)
		}
	}()

	// Try to broadcast to disconnected client
	h.BroadcastTestResults([]WSTestResult{{
		Name:     "TestExample",
		Status:   "passed",
		Duration: "100ms",
		LastRun:  time.Now(),
	}})

	// This test passes if it doesn't panic or deadlock
}

func TestConcurrentConnections(t *testing.T) {
	h := NewWebSocketHandler()
	h.StartBroadcaster()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.HandleWebSocket(w, r)
	}))
	defer server.Close()

	url := "ws" + server.URL[4:] + "/ws"
	const numConnections = 10
	var connected int32

	// Connect multiple clients concurrently
	for i := 0; i < numConnections; i++ {
		go func() {
			conn, _, err := websocket.DefaultDialer.Dial(url, nil)
			if err != nil {
				t.Logf("Failed to connect: %v", err)
				return
			}
			atomic.AddInt32(&connected, 1)
			defer conn.Close()

			// Keep connection alive for a bit
			time.Sleep(500 * time.Millisecond)
		}()
	}

	// Wait for connections to be established
	time.Sleep(100 * time.Millisecond)

	// Verify all connections are tracked
	if count := h.ConnectionCount(); count != numConnections {
		t.Errorf("Expected %d connections, got %d", numConnections, count)
	}

	// Wait for connections to close
	time.Sleep(600 * time.Millisecond)

	// Verify all connections are cleaned up
	if count := h.ConnectionCount(); count != 0 {
		t.Errorf("Expected 0 connections after cleanup, got %d", count)
	}
}
