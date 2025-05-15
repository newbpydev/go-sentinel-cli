package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketMessage represents a message sent over WebSocket
type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// WSTestResult represents the result of a single test over WebSocket
type WSTestResult struct {
	Name     string    `json:"name"`
	Status   string    `json:"status"`
	Duration string    `json:"duration"`
	LastRun  time.Time `json:"lastRun"`
	Output   string    `json:"output,omitempty"`
}

// StatusUpdate represents a status update message
type StatusUpdate struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	// Upgrader for HTTP connections
	upgrader websocket.Upgrader
	
	// Connected clients
	clients map[*websocket.Conn]bool
	
	// Mutex for thread safety
	clientsMu sync.Mutex
	
	// Broadcast channel for sending messages to all clients
	broadcast chan WebSocketMessage
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Allow all origins for development
			CheckOrigin: func(r *http.Request) bool {
				// In production, you should validate the origin
				return true
			},
		},
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan WebSocketMessage, 256), // Buffered channel to prevent blocking
	}
}

// StartBroadcaster starts the WebSocket broadcaster goroutine
func (h *WebSocketHandler) StartBroadcaster() {
	go func() {
		for {
			// Get the next message from the broadcast channel
			msg := <-h.broadcast
			
			// Send to all clients
			h.clientsMu.Lock()
			for client := range h.clients {
				err := client.WriteJSON(msg)
				if err != nil {
					log.Printf("Error sending WebSocket message: %v", err)
					client.Close()
					delete(h.clients, client)
				}
			}
			h.clientsMu.Unlock()
		}
	}()
	
	// Start a demo goroutine that sends updates periodically
	go h.sendDemoUpdates()
}

// HandleWebSocket upgrades HTTP connections to WebSocket
func (h *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Upgrade the HTTP connection to WebSocket
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}
	
	// Register this client
	h.clientsMu.Lock()
	h.clients[conn] = true
	h.clientsMu.Unlock()
	
	// Send initial data
	initialPayload, _ := json.Marshal(map[string]interface{}{
		"connectedAt": time.Now(),
		"clientCount": len(h.clients),
	})
	initialMsg := WebSocketMessage{
		Type:    "connected",
		Payload: initialPayload,
	}
	conn.WriteJSON(initialMsg)
	
	// Start the reader for this connection
	go h.readPump(conn)
}

// readPump handles incoming WebSocket messages
func (h *WebSocketHandler) readPump(conn *websocket.Conn) {
	defer func() {
		h.clientsMu.Lock()
		delete(h.clients, conn)
		h.clientsMu.Unlock()
		conn.Close()
	}()
	
	// Configure the connection
	conn.SetReadLimit(1024 * 1024) // 1MB max message size
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error { 
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil 
	})

	for {
		var msg WebSocketMessage
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle different message types
		switch msg.Type {
		case "ping":
			// Respond to ping with pong
			pongMsg := WebSocketMessage{
				Type: "pong",
			}
			if err := conn.WriteJSON(pongMsg); err != nil {
				log.Printf("Error sending pong: %v", err)
				return
			}
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

// sendError sends an error message to the client
func (h *WebSocketHandler) sendError(conn *websocket.Conn, code, message string) {
	errMsg := map[string]interface{}{
		"error":   code,
		"message": message,
	}
	payload, _ := json.Marshal(errMsg)
	err := conn.WriteJSON(WebSocketMessage{
		Type:    "error",
		Payload: payload,
	})
	if err != nil {
		log.Printf("Error sending error message: %v", err)
	}
}

// BroadcastTestUpdate broadcasts a test result update to all clients
func (h *WebSocketHandler) BroadcastTestUpdate(test WSTestResult) {
	payload, err := json.Marshal(test)
	if err != nil {
		log.Printf("Error marshaling test update: %v", err)
		return
	}
	h.broadcast <- WebSocketMessage{
		Type:    "test-update",
		Payload: payload,
	}
}

// BroadcastMetricsUpdate broadcasts metrics updates to all clients
func (h *WebSocketHandler) BroadcastMetricsUpdate(metrics map[string]interface{}) {
	payload, err := json.Marshal(metrics)
	if err != nil {
		log.Printf("Error marshaling metrics: %v", err)
		return
	}
	h.broadcast <- WebSocketMessage{
		Type:    "metrics-update",
		Payload: payload,
	}
}

// BroadcastTestResults sends test results to all connected clients
func (h *WebSocketHandler) BroadcastTestResults(testResults []WSTestResult) {
	// Convert test results to JSON
	payload, err := json.Marshal(testResults)
	if err != nil {
		log.Printf("Error marshaling test results: %v", err)
		return
	}

	msg := WebSocketMessage{
		Type:    "test_results",
		Payload: payload,
	}
	
	h.broadcast <- msg
}

// SendNotification broadcasts a notification to all connected clients
func (h *WebSocketHandler) SendNotification(notificationType, title, message string, duration int) {
	notification := map[string]interface{}{
		"type":    notificationType,
		"title":   title,
		"message": message,
		"duration": duration,
	}
	payload, err := json.Marshal(notification)
	if err != nil {
		log.Printf("Error marshaling notification: %v", err)
		return
	}
	
	h.broadcast <- WebSocketMessage{
		Type:    "notification",
		Payload: payload,
	}
}

// sendDemoUpdates sends periodic demo updates for testing
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
		
		// Send a test update
		h.BroadcastTestUpdate(WSTestResult{
			Name:     "TestRandomFunction",
			Status:   randomStatus(),
			Duration: randomDuration(),
			LastRun:  time.Now(),
		})
		
		// Send a metrics update
		metricsPayload, _ := json.Marshal(map[string]interface{}{
			"TotalTests": 128,
			"Passing":    119,
			"Failing":    9,
			"Duration":   "1.3s",
		})
		h.broadcast <- WebSocketMessage{
			Type:    "metrics-update",
			Payload: metricsPayload,
		}
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
