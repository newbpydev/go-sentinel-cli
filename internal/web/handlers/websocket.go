// Package handlers provides HTTP and WebSocket handlers for the Go Sentinel web server.
// It includes implementations for test results, WebSocket communication, and other web endpoints.
package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketMessage represents a message sent over WebSocket connection.
// It follows a standard format with a type field and an optional payload.
// This structure is used for all communication between the server and clients.
type WebSocketMessage struct {
	// Type of the message (e.g., "test-update", "metrics-update", "notification")
	Type string `json:"type"`

	// Payload of the message as raw JSON, can be unmarshaled into specific types
	Payload json.RawMessage `json:"payload,omitempty"`
}

// WSTestResult represents the result of a single test execution sent over WebSocket.
// It contains all relevant information about a test run including status, duration, and output.
type WSTestResult struct {
	// Unique identifier of the test result
	ID string `json:"id,omitempty"`

	// Name of the test
	Name string `json:"name"`

	// Status of the test (e.g. "passed", "failed", etc.)
	Status string `json:"status"`

	// Duration of the test run
	Duration string `json:"duration"`

	// Timestamp of the last test run
	LastRun time.Time `json:"lastRun"`

	// Timestamp of the test result, if available
	Timestamp string `json:"timestamp,omitempty"`

	// Output of the test run, if any
	Output string `json:"output,omitempty"`
}

// StatusUpdate represents a status update message sent over WebSocket.
// It contains the current status and a timestamp for tracking when the status changed.
type StatusUpdate struct {
	// Status of the operation (e.g., "running", "completed", "failed")
	Status string `json:"status"`

	// Timestamp when the status was updated
	Time time.Time `json:"time"`
}

// WebSocketHandler handles WebSocket connections for real-time communication.
// It manages client connections, message broadcasting, and subscription-based messaging.
type WebSocketHandler struct {
	// WebSocket connection upgrader
	upgrader websocket.Upgrader
	// Map of active client connections
	clients map[*websocket.Conn]bool
	// Count of active connections
	connectionCount int32
	// Mutex for clients map access
	clientsMu sync.RWMutex
	// Channel for broadcasting messages to all clients
	broadcast chan WebSocketMessage
	// Context for cancellation and shutdown
	ctx context.Context
	// Cancel function for the context
	cancelFunc context.CancelFunc
	// Wait group for goroutines
	wg sync.WaitGroup
	// Flag to enable demo mode with periodic updates
	demoMode bool
	// Map of client subscriptions to specific topics
	subscriptions map[*websocket.Conn]map[string]bool
	// Mutex for subscriptions map access
	subscriptionsMu sync.RWMutex
}

// NewWebSocketHandler creates a new WebSocket handler with default configuration.
// It initializes all required maps, channels, and sets up the context for proper lifecycle management.
func NewWebSocketHandler() *WebSocketHandler {
	ctx, cancel := context.WithCancel(context.Background())
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Allow all origins for development
			CheckOrigin: func(_ *http.Request) bool {
				// In production, you should validate the origin
				return true
			},
		},
		clients:       make(map[*websocket.Conn]bool),
		broadcast:     make(chan WebSocketMessage, 256),
		ctx:           ctx,
		cancelFunc:    cancel,
		subscriptions: make(map[*websocket.Conn]map[string]bool),
	}
}

// StartBroadcaster starts the WebSocket broadcaster goroutine that handles message distribution.
// It processes messages from the broadcast channel and sends them to all connected clients.
// This method should be called after creating a new WebSocketHandler instance.
func (h *WebSocketHandler) StartBroadcaster() {
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		if h.demoMode {
			h.wg.Add(1)
			go func() {
				defer h.wg.Done()
				h.sendDemoUpdates()
			}()
		}

		for {
			select {
			case msg, ok := <-h.broadcast:
				if !ok {
					// Channel closed
					return
				}
				h.broadcastMessage(msg)
			case <-h.ctx.Done():
				// Close all connections when context is canceled
				h.closeAllConnections()
				return
			}
		}
	}()
}

// broadcastMessage sends a message to all connected clients
func (h *WebSocketHandler) broadcastMessage(msg WebSocketMessage) {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	for client := range h.clients {
		err := client.WriteJSON(msg)
		if err != nil {
			log.Printf("Error sending WebSocket message: %v", err)
			// Don't remove the client here, let readPump handle it
		}
	}
}

// closeAllConnections closes all active WebSocket connections

func (h *WebSocketHandler) closeAllConnections() {
	h.clientsMu.Lock()
	defer h.clientsMu.Unlock()

	for client := range h.clients {
		// Try to send a close message first
		err := client.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Server shutting down"),
		)
		if err != nil && !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
			log.Printf("Failed to write close message to WebSocket client: %v", err)
		}

		// Close the connection
		if err := client.Close(); err != nil {
			log.Printf("Error closing WebSocket connection: %v", err)
		}

		// Remove from clients map
		delete(h.clients, client)
	}

	// Reset connection counter
	atomic.StoreInt32(&h.connectionCount, 0)
}

// ConnectionCount returns the number of currently connected WebSocket clients.
// This method is thread-safe and can be called from any goroutine.
func (h *WebSocketHandler) ConnectionCount() int {
	return int(atomic.LoadInt32(&h.connectionCount))
}

// Close gracefully shuts down the WebSocket handler and all associated resources.
// It cancels the context, closes all client connections, and waits for all goroutines to complete.
// This method should be called when the application is shutting down.
func (h *WebSocketHandler) Close() {
	h.cancelFunc()
	h.wg.Wait()
	close(h.broadcast)
}

// HandleWebSocket upgrades HTTP connections to WebSocket protocol and manages the connection lifecycle.
// It handles the initial connection setup, sends a connected message, and starts the read pump.
// This method should be registered as an HTTP handler for the WebSocket endpoint.
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}

	// Register the client
	h.clientsMu.Lock()
	h.clients[conn] = true
	h.clientsMu.Unlock()

	// Initialize the subscriptions map for this client
	h.subscriptionsMu.Lock()
	h.subscriptions[conn] = make(map[string]bool)
	h.subscriptionsMu.Unlock()

	// Increment connection counter
	atomic.AddInt32(&h.connectionCount, 1)

	// Start reader goroutine
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		h.readPump(conn)
	}()

	// Send welcome message
	h.sendConnectedMessage(conn)
}

func (h *WebSocketHandler) sendConnectedMessage(conn *websocket.Conn) {
	// Create the initial payload
	initialData := map[string]interface{}{
		"connectedAt": time.Now(),
		"clientCount": h.ConnectionCount(),
	}

	// Marshal the payload with error handling
	initialPayload, err := json.Marshal(initialData)
	if err != nil {
		log.Printf("Error marshaling connected message: %v", err)
		// Try to send a basic error message
		h.sendError(conn, "json_error", "Failed to encode connection data")
		return
	}

	// Create and send the WebSocket message
	initialMsg := WebSocketMessage{
		Type:    "connected",
		Payload: initialPayload,
	}

	// Set a write deadline to prevent hanging
	if err := conn.SetWriteDeadline(time.Now().Add(10 * time.Second)); err != nil {
		log.Printf("Error setting write deadline: %v", err)
		return
	}

	// Send the message with error handling
	if err := conn.WriteJSON(initialMsg); err != nil {
		log.Printf("Error sending connected message: %v", err)
	}
}

// sendPong sends a pong message to the client
func (h *WebSocketHandler) sendPong(conn *websocket.Conn) {
	// Create pong message
	pongMsg := WebSocketMessage{
		Type: "pong",
		Payload: json.RawMessage(`{"timestamp":"` + time.Now().Format(time.RFC3339) + `"}`),
	}

	// Send pong with error handling
	if err := conn.WriteJSON(pongMsg); err != nil {
		log.Printf("Error sending pong: %v", err)
		return
	}
}

// readPump handles incoming WebSocket messages
func (h *WebSocketHandler) readPump(conn *websocket.Conn) {
	// Set initial read deadline
	if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		log.Printf("Error setting initial read deadline: %v", err)
		return
	}

	// Set pong handler to update read deadline
	conn.SetPongHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			log.Printf("Error setting read deadline in pong handler: %v", err)
		}
		return nil
	})

	defer func() {
		// Unregister client and decrement counter
		h.clientsMu.Lock()
		if _, ok := h.clients[conn]; ok {
			delete(h.clients, conn)
			atomic.AddInt32(&h.connectionCount, -1)
		}
		h.clientsMu.Unlock()

		// Try to send close message
		err := conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
			time.Now().Add(time.Second*5),
		)
		if err != nil && !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
			log.Printf("Error sending close message: %v", err)
		}

		// Close the connection
		if err := conn.Close(); err != nil {
			log.Printf("Error closing WebSocket connection: %v", err)
		}
	}()

	// Configure connection

	conn.SetReadLimit(512) // 512 bytes max message size

	if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
		log.Printf("Error setting read deadline: %v", err)
		return
	}

	conn.SetPongHandler(func(string) error {
		if err := conn.SetReadDeadline(time.Now().Add(60 * time.Second)); err != nil {
			log.Printf("Error setting read deadline in pong handler: %v", err)
		}
		return nil
	})

	// Start ping-pong keepalive

	h.wg.Add(1)

	go func() {
		defer h.wg.Done()
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if err := conn.WriteControl(
					websocket.PingMessage,
					[]byte{},
					time.Now().Add(10*time.Second),
				); err != nil {
					return
				}
			case <-h.ctx.Done():
				return
			}
		}
	}()

	// Message handling loop

	for {

		var msg WebSocketMessage

		err := conn.ReadJSON(&msg)

		if err != nil {

			if websocket.IsUnexpectedCloseError(

				err,

				websocket.CloseGoingAway,

				websocket.CloseAbnormalClosure,

				websocket.CloseNormalClosure,
			) {

				log.Printf("WebSocket error: %v", err)

			}

			break

		}

		// Handle ping message

		if msg.Type == "ping" {

			h.sendPong(conn)

			continue

		}

		// Handle other message types if needed

		switch msg.Type {

		case "subscribe":

			// Handle subscription requests

		case "unsubscribe":

			// Handle unsubscription requests

		case "test_result":

			// Handle test result message

			var result WSTestResult

			if err := json.Unmarshal(msg.Payload, &result); err != nil {

				h.sendError(conn, "invalid_test_result", "Invalid test result format: "+err.Error())

				continue

			}

			log.Printf("Received test result: %+v", result)

			// Echo back with ack

			ackMsg := WebSocketMessage{

				Type: "test_result_ack",
			}

			if err := conn.WriteJSON(ackMsg); err != nil {

				log.Printf("Error sending ack: %v", err)

			}

		case "status_update":

			// Handle status update

			var status StatusUpdate

			if err := json.Unmarshal(msg.Payload, &status); err != nil {

				h.sendError(conn, "invalid_status_update", "Invalid status update format")

				continue

			}

			log.Printf("Status update: %+v", status)

		default:

			h.sendError(conn, "unknown_message_type", "Unhandled message type: "+msg.Type)

		}

	}

}

// sendError sends an error message to the client with proper error handling
func (h *WebSocketHandler) sendError(conn *websocket.Conn, code, message string) {
	// Create error message with timestamp
	errMsg := map[string]interface{}{
		"error":   code,
		"message": message,
		"time":    time.Now().Format(time.RFC3339),
	}

	// Try to marshal the error message
	payload, err := json.Marshal(errMsg)
	if err != nil {
		log.Printf("Error marshaling error message: %v", err)
		// Fallback to a basic error message
		payload = []byte(`{"error":"json_error","message":"Failed to encode error message"}`)
	}

	// Set a write deadline to prevent hanging
	if deadlineErr := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); deadlineErr != nil {
		log.Printf("Error setting write deadline: %v", deadlineErr)
		return
	}

	// Send the error message
	err = conn.WriteJSON(WebSocketMessage{
		Type:    "error",
		Payload: payload,
	})

	if err != nil && !websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
		log.Printf("Error sending error message (type=%s): %v", code, err)
	}
}

// BroadcastTestUpdate broadcasts a test result update to all connected clients.
// It marshals the test result into JSON and sends it with proper error handling and timeout.
// This method is used to notify clients about individual test execution results in real-time.
func (h *WebSocketHandler) BroadcastTestUpdate(test WSTestResult) {
	// Marshal the test result with error handling
	payload, err := json.Marshal(test)
	if err != nil {
		log.Printf("Error marshaling test update: %v", err)
		return
	}

	// Create the message
	msg := WebSocketMessage{
		Type:    "test-update",
		Payload: payload,
	}

	// Use a select to avoid blocking if the broadcast channel is full
	select {
	case h.broadcast <- msg:
		// Message sent successfully
	case <-time.After(5 * time.Second):
		log.Printf("Timeout sending metrics update to broadcast channel")
	}
}

// BroadcastMetricsUpdate broadcasts metrics updates to all connected clients.
// It marshals the metrics data into JSON and sends it with proper error handling and timeout.
// This method is used to notify clients about test coverage, performance metrics, and other statistical data.
func (h *WebSocketHandler) BroadcastMetricsUpdate(metrics map[string]interface{}) {
	// Marshal the metrics with error handling
	payload, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Error marshaling metrics update: %v", err)
		return
	}

	// Create the message
	msg := WebSocketMessage{
		Type:    "metrics-update",
		Payload: payload,
	}

	// Use a select to avoid blocking if the broadcast channel is full
	select {
	case h.broadcast <- msg:
		// Message sent successfully
	case <-time.After(5 * time.Second):
		log.Printf("Timeout sending metrics update to broadcast channel")
	}
}

// BroadcastTestResults sends a batch of test results to all connected clients.
// It marshals the test results array into JSON and sends it with proper error handling and timeout.
// This method is used to update clients with a complete set of test results, typically after a test run completes.
func (h *WebSocketHandler) BroadcastTestResults(testResults []WSTestResult) {
	// Marshal test results with error handling
	payload, err := json.Marshal(testResults)
	if err != nil {
		log.Printf("Error marshaling test results: %v", err)
		return
	}

	// Create the message
	msg := WebSocketMessage{
		Type:    "test-results",
		Payload: payload,
	}

	// Use a select to avoid blocking if the broadcast channel is full
	select {
	case h.broadcast <- msg:
		// Message sent successfully
	case <-time.After(5 * time.Second):
		log.Printf("Timeout sending test results to broadcast channel")
	}
}

// SendNotification broadcasts a notification message to all connected clients.
// It creates a notification with the specified type, title, message, and duration,
// then marshals it to JSON and sends it with proper error handling and timeout.
// This method is used to display notifications to users in the UI, such as alerts or success messages.
func (h *WebSocketHandler) SendNotification(notificationType, title, message string, duration int) {
	// Create the notification payload
	notification := map[string]interface{}{
		"type":     notificationType,
		"title":    title,
		"message":  message,
		"duration": duration,
	}

	// Marshal the notification with error handling
	payload, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Error marshaling notification: %v", err)
		return
	}

	// Create the message
	msg := WebSocketMessage{
		Type:    "notification",
		Payload: payload,
	}

	// Use a select to avoid blocking if the broadcast channel is full
	select {
	case h.broadcast <- msg:
		// Message sent successfully
	case <-time.After(5 * time.Second):
		log.Printf("Timeout sending notification to broadcast channel")
	}
}

// sendDemoUpdates sends periodic demo updates for testing with proper error handling
func (h *WebSocketHandler) sendDemoUpdates() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		h.clientsMu.Lock()
		if len(h.clients) == 0 {
			h.clientsMu.Unlock()
			continue
		}
		h.clientsMu.Unlock()

		// 1. Send test result update
		h.sendTestUpdate()

		// 2. Send metrics update
		h.sendDemoMetrics()
	}
}

// sendTestUpdate sends a single test update with random data
func (h *WebSocketHandler) sendTestUpdate() {
	// Create test result with random status and duration
	testResult := WSTestResult{
		Name:     "TestRandomFunction",
		Status:   randomStatus(),
		Duration: randomDuration(),
		LastRun:  time.Now(),
	}

	// Marshal the test result with error handling
	payload, err := json.Marshal(testResult)
	if err != nil {
		log.Printf("Error marshaling demo test result: %v", err)
		return
	}

	// Create the message
	msg := WebSocketMessage{
		Type:    "test-update",
		Payload: payload,
	}

	// Use a select to avoid blocking if the broadcast channel is full
	select {
	case h.broadcast <- msg:
		// Message sent successfully
	case <-time.After(5 * time.Second):
		log.Printf("Timeout sending demo update to broadcast channel")
	}
}

// sendDemoMetrics sends demo metrics update
func (h *WebSocketHandler) sendDemoMetrics() {
	metricsPayload, err := json.Marshal(map[string]interface{}{
		"TotalTests": 128,
		"Passing":    119,
		"Failing":    9,
		"Duration":   "1.3s",
	})

	if err != nil {
		log.Printf("Error marshaling demo metrics: %v", err)
		return
	}

	msg := WebSocketMessage{
		Type:    "metrics-update",
		Payload: metricsPayload,
	}

	// Use a select to avoid blocking if the broadcast channel is full
	select {
	case h.broadcast <- msg:
		// Message sent successfully
	case <-time.After(5 * time.Second):
		log.Printf("Timeout sending metrics update to broadcast channel")
	}
}

// Helper functions

// randomStatus returns a random test status for demo

func randomStatus() string {

	if time.Now().UnixNano()%5 == 0 {

		return "failed"

	}

	return "passed"

}

// randomDuration returns a random test duration for demo

func randomDuration() string {

	durations := []string{"0.5s", "0.8s", "1.1s", "1.3s", "0.9s"}

	return durations[time.Now().UnixNano()%int64(len(durations))]

}
