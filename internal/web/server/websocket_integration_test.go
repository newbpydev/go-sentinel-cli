//go:build integration
// +build integration

package server

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"golang.org/x/net/websocket"
)

type TestEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type TestResult struct {
	Name   string `json:"name"`
	Passed bool   `json:"passed"`
}

type StatusUpdate struct {
	Status string `json:"status"`
	Time   int64  `json:"time"`
}

// TestWebSocketConnection verifies handshake and echo roundtrip
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

// TestWebSocketMessageEncoding verifies JSON encoding/decoding of test events
func TestWebSocketMessageEncoding(t *testing.T) {
	event := TestEvent{
		Type:    "test_result",
		Payload: TestResult{Name: "TestFoo", Passed: true},
	}
	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}
	var decoded TestEvent
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}
	if decoded.Type != event.Type {
		t.Fatalf("Expected type '%s', got '%s'", event.Type, decoded.Type)
	}
}

// TestWebSocketDifferentMessageTypes verifies handling of multiple event types
func TestWebSocketDifferentMessageTypes(t *testing.T) {
	events := []TestEvent{
		{Type: "test_result", Payload: TestResult{Name: "TestFoo", Passed: true}},
		{Type: "status_update", Payload: StatusUpdate{Status: "running", Time: time.Now().Unix()}},
	}
	for _, event := range events {
		data, err := json.Marshal(event)
		if err != nil {
			t.Errorf("Failed to marshal event type '%s': %v", event.Type, err)
		}
		// Optionally decode and check type
		var decoded map[string]interface{}
		err = json.Unmarshal(data, &decoded)
		if err != nil {
			t.Errorf("Failed to unmarshal event type '%s': %v", event.Type, err)
		}
		if decoded["type"] != event.Type {
			t.Errorf("Expected type '%s', got '%v'", event.Type, decoded["type"])
		}
	}
}

// TestWebSocketMalformedMessage verifies error handling for malformed JSON
func TestWebSocketMalformedMessage(t *testing.T) {
	echoHandler := func(ws *websocket.Conn) {
		// Try to read a message (should be malformed JSON)
		var msg = make([]byte, 512)
		_, err := ws.Read(msg) // We expect an error, so we don't need the bytes read
		if err != nil {
			t.Logf("Expected error reading malformed message: %v", err)
			return
		}

		// Send back an error response
		errMsg := map[string]string{
			"error": "invalid JSON",
		}
		err = json.NewEncoder(ws).Encode(errMsg)
		if err != nil {
			t.Errorf("Failed to send error response: %v", err)
		}
	}

	ts := httptest.NewServer(websocket.Handler(echoHandler))
	defer ts.Close()

	u, _ := url.Parse(ts.URL)
	u.Scheme = "ws"

	// Connect to the WebSocket server
	ws, err := websocket.Dial(u.String(), "", "http://localhost")
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer ws.Close()

	// Send malformed JSON
	badJSON := []byte(`{"type": "test_result", "payload": `) // incomplete JSON
	_, err = ws.Write(badJSON)
	if err != nil {
		t.Fatalf("Failed to write to WebSocket: %v", err)
	}

	// Read the error response
	var response map[string]string
	err = json.NewDecoder(ws).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify we got an error response
	if response["error"] != "invalid JSON" {
		t.Errorf("Expected error response, got: %v", response)
	}
}
