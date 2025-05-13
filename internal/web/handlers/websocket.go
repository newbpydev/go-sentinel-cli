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
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
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
				return true
			},
		},
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan WebSocketMessage),
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
	initialMsg := WebSocketMessage{
		Type: "connected",
		Payload: map[string]interface{}{
			"connectedAt": time.Now(),
			"clientCount": len(h.clients),
		},
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
	conn.SetReadLimit(512) // Max message size
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	// Read messages in a loop
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, 
				websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
		
		// Parse the message
		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Failed to parse WebSocket message: %v", err)
			continue
		}
		
		// Handle the message based on type
		switch msg.Type {
		case "ping":
			// Respond to ping with pong
			conn.WriteJSON(WebSocketMessage{
				Type:    "pong",
				Payload: time.Now(),
			})
		}
	}
}

// BroadcastTestUpdate broadcasts a test result update to all clients
func (h *WebSocketHandler) BroadcastTestUpdate(test TestResult) {
	h.broadcast <- WebSocketMessage{
		Type:    "test-update",
		Payload: test,
	}
}

// BroadcastMetricsUpdate broadcasts metrics updates to all clients
func (h *WebSocketHandler) BroadcastMetricsUpdate(metrics map[string]interface{}) {
	h.broadcast <- WebSocketMessage{
		Type:    "metrics-update",
		Payload: metrics,
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
		h.BroadcastTestUpdate(TestResult{
			Name:     "TestRandomFunction",
			Status:   randomStatus(),
			Duration: randomDuration(),
			LastRun:  time.Now(),
		})
		
		// Send a metrics update
		h.BroadcastMetricsUpdate(map[string]interface{}{
			"TotalTests": 128,
			"Passing":    119,
			"Failing":    9,
			"Duration":   "1.3s",
		})
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
