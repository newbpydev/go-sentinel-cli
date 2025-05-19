package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

//nolint:unused // Test helper function for future test cases
func assertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

//nolint:unused // Test helper function for future test cases
func assertEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if got != want {
		t.Fatalf("Expected %v, got %v", want, got)
	}
}

//nolint:unused // Test helper function for future test cases
func assertNotNil(t *testing.T, v interface{}) {
	t.Helper()
	if v == nil {
		t.Fatal("Expected non-nil value")
	}
}

func TestWebSocketMessageSerialization(t *testing.T) {
	tests := []struct {
		name    string
		message WebSocketMessage
	}{
		{
			name: "test result message",
			message: WebSocketMessage{
				Type:    "test_result",
				Payload: json.RawMessage(`{"name":"TestExample","passed":true,"message":"Test passed successfully","duration":125000000}`),
			},
		},
		{
			name: "metrics update message",
			message: WebSocketMessage{
				Type:    "metrics_update",
				Payload: json.RawMessage(`{"testsRun":5,"testsPassed":5,"coverage":85.5}`),
			},
		},
		{
			name: "status update message",
			message: WebSocketMessage{
				Type:    "status_update",
				Payload: json.RawMessage(`{"status":"running","time":"` + time.Now().Format(time.RFC3339) + `"}`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to JSON
			data, err := json.Marshal(tt.message)
			if err != nil {
				t.Fatalf("marshaling failed: %v", err)
			}

			// Parse back from JSON
			var parsed WebSocketMessage
			err = json.Unmarshal(data, &parsed)
			if err != nil {
				t.Fatalf("unmarshaling failed: %v", err)
			}

			// Type should match exactly
			if tt.message.Type != parsed.Type {
				t.Fatalf("expected type %q, got %q", tt.message.Type, parsed.Type)
			}

			// For complex payloads, we'll just check they can be marshaled back to JSON
			// since direct comparison might be tricky with time.Time and numeric types
			_, err = json.Marshal(parsed.Payload)
			if err != nil {
				t.Fatalf("payload is not valid JSON: %v", err)
			}
		})
	}
}

func TestWebSocketMessageHandling(t *testing.T) {
	h := NewWebSocketHandler()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.HandleWebSocket(w, r)
	}))
	defer server.Close()

	// Create WebSocket URL
	url := "ws" + server.URL[4:] + "/ws"

	// Connect to the server
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Errorf("Failed to close connection: %v", err)
		}
	}()

	// Test sending a message to the server
	payload, err := json.Marshal(map[string]string{"message": "test"})
	if err != nil {
		t.Fatalf("Failed to marshal test message: %v", err)
	}
	testMessage := WebSocketMessage{
		Type:    "ping",
		Payload: payload,
	}
	err = conn.WriteJSON(testMessage)
	if err != nil {
		t.Fatalf("Failed to write JSON message: %v", err)
	}

	// Read the initial connected message
	var connectedMsg WebSocketMessage
	err = conn.ReadJSON(&connectedMsg)
	if err != nil {
		t.Fatalf("Failed to read connected message: %v", err)
	}
	if connectedMsg.Type != "connected" {
		t.Fatalf("Expected initial message type 'connected', got %q", connectedMsg.Type)
	}

	// Test receiving a message from the server
	// The server should respond with a "pong" message
	var response WebSocketMessage
	err = conn.ReadJSON(&response)
	if err != nil {
		t.Fatalf("Failed to read JSON response: %v", err)
	}
	if response.Type != "pong" {
		t.Fatalf("Expected message type 'pong', got %q", response.Type)
	}
}

func TestBroadcastTestResults(t *testing.T) {
	h := NewWebSocketHandler()
	h.StartBroadcaster()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.HandleWebSocket(w, r)
	}))
	defer server.Close()

	// Connect first client
	url := "ws" + server.URL[4:] + "/ws"
	conn1, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect first client: %v", err)
	}
	defer func() {
		if err := conn1.Close(); err != nil {
			t.Errorf("Failed to close connection: %v", err)
		}
	}()

	// Connect second client
	conn2, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect second client: %v", err)
	}
	defer func() {
		if err := conn2.Close(); err != nil {
			t.Errorf("Failed to close connection: %v", err)
		}
	}()

	// Send test results
	testResults := []WSTestResult{{
		Name:     "TestExample",
		Status:   "passed",
		Duration: "100ms",
		LastRun:  time.Now(),
	}}
	h.BroadcastTestResults(testResults)

	// Both clients should receive the initial "connected" message first
	var msg1, msg2 WebSocketMessage
	err = conn1.ReadJSON(&msg1)
	if err != nil {
		t.Fatalf("Failed to read from first client: %v", err)
	}
	if msg1.Type != "connected" {
		t.Fatalf("Expected initial message type 'connected' for first client, got %q", msg1.Type)
	}

	err = conn2.ReadJSON(&msg2)
	if err != nil {
		t.Fatalf("Failed to read from second client: %v", err)
	}
	if msg2.Type != "connected" {
		t.Fatalf("Expected initial message type 'connected' for second client, got %q", msg2.Type)
	}

	// Now read the test results
	err = conn1.ReadJSON(&msg1)
	if err != nil {
		t.Fatalf("Failed to read test results from first client: %v", err)
	}
	if msg1.Type != "test_results" {
		t.Errorf("Expected message type 'test_results' for first client, got %q", msg1.Type)
	}

	err = conn2.ReadJSON(&msg2)
	if err != nil {
		t.Fatalf("Failed to read test results from second client: %v", err)
	}
	if msg2.Type != "test_results" {
		t.Errorf("Expected message type 'test_results' for second client, got %q", msg2.Type)
	}
}

func TestMessageRoutingByType(t *testing.T) {
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
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Errorf("Failed to close connection: %v", err)
		}
	}()

	// Read the initial connected message
	var msg WebSocketMessage
	err = conn.ReadJSON(&msg)
	if err != nil {
		t.Fatalf("Failed to read connected message: %v", err)
	}
	if msg.Type != "connected" {
		t.Fatalf("Expected initial message type 'connected', got %q", msg.Type)
	}

	// Test different message types and their routing
	tests := []struct {
		name        string
		messageType string
		sendFunc    func()
		expectData  bool
	}{
		{
			name:        "test results message",
			messageType: "test_results",
			sendFunc: func() {
				h.BroadcastTestResults([]WSTestResult{{
					Name:     "TestExample",
					Status:   "passed",
					Duration: "100ms",
					LastRun:  time.Now(),
				}})
			},
			expectData: true,
		},
		{
			name:        "metrics update message",
			messageType: "metrics-update",
			sendFunc: func() {
				h.BroadcastMetricsUpdate(map[string]interface{}{
					"testsRun":    5,
					"testsPassed": 5,
					"coverage":    85.5,
				})
			},
			expectData: true,
		},
		{
			name:        "notification message",
			messageType: "notification",
			sendFunc: func() {
				h.SendNotification("info", "Test", "This is a test notification", 3000)
			},
			expectData: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Send the test message
			tt.sendFunc()

			// Read the message from the WebSocket
			err = conn.ReadJSON(&msg)
			if err != nil {
				t.Fatalf("Failed to read %s: %v", tt.messageType, err)
			}

			// Verify the message type
			if msg.Type != tt.messageType {
				t.Errorf("Expected message type %q, got %q", tt.messageType, msg.Type)
			}

			// Verify the payload is not empty if expected
			if tt.expectData && len(msg.Payload) == 0 {
				t.Error("Expected non-empty payload")
			}

			// For test results, verify the structure
			if tt.messageType == "test_results" {
				var results []WSTestResult
				if err := json.Unmarshal(msg.Payload, &results); err != nil {
					t.Fatalf("Failed to unmarshal test results: %v", err)
				}
				if len(results) == 0 {
					t.Error("Expected at least one test result")
				}
			}
		})
	}
}

func TestHandleMalformedMessages(t *testing.T) {
	h := NewWebSocketHandler()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.HandleWebSocket(w, r)
	}))
	defer server.Close()

	url := "ws" + server.URL[4:] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Errorf("Failed to close connection: %v", err)
		}
	}()

	// Send malformed JSON
	err = conn.WriteMessage(websocket.TextMessage, []byte("{invalid json"))
	if err != nil {
		t.Fatalf("Failed to write malformed JSON: %v", err)
	}

	// Send message with invalid type
	err = conn.WriteJSON(map[string]interface{}{
		"type": 123, // invalid type, should be string
	})
	if err != nil {
		t.Fatalf("Failed to write message with invalid type: %v", err)
	}
}
