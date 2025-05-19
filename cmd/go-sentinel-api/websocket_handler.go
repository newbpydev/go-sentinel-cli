package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	gorillaws "github.com/gorilla/websocket"
	"github.com/newbpydev/go-sentinel/internal/api/websocket"
)

// SetupWebSocketRoutes adds WebSocket routes to the provided router
func SetupWebSocketRoutes(router *mux.Router) {
	// Create required WebSocket components
	connManager := websocket.NewConnectionManager()
	router.HandleFunc("/ws", handleWebSocket(connManager))

	log.Println("WebSocket routes registered at /ws endpoint")
}

// WebSocket connection upgrader
var upgrader = gorillaws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow connections from any origin for development
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// handleWebSocket handles WebSocket connections
func handleWebSocket(connManager *websocket.ConnectionManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Upgrade the HTTP connection to a WebSocket connection
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade to WebSocket: %v", err)
			return
		}

		// Register the connection with the connection manager
		connID := connManager.Add(NewWebSocketConnection(conn))
		log.Printf("New WebSocket connection established: %s", connID)

		// Handle incoming messages in a goroutine
		go handleMessages(conn, connID, connManager)
	}
}

// WebSocketConnection wraps a gorilla WebSocket connection
type WebSocketConnection struct {
	conn *gorillaws.Conn
}

// NewWebSocketConnection creates a new WebSocketConnection
func NewWebSocketConnection(conn *gorillaws.Conn) *WebSocketConnection {
	return &WebSocketConnection{conn: conn}
}

// Close implements the Connection interface
func (c *WebSocketConnection) Close() error {
	return c.conn.Close()
}

// handleMessages processes incoming WebSocket messages
func handleMessages(conn *gorillaws.Conn, connID string, connManager *websocket.ConnectionManager) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("Error closing WebSocket connection: %v", err)
		}
		connManager.Remove(connID)
		log.Printf("WebSocket connection closed: %s", connID)
	}()

	for {
		// Read message
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if gorillaws.IsUnexpectedCloseError(err,
				gorillaws.CloseGoingAway,
				gorillaws.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		// Log incoming message
		log.Printf("Received message from %s: %s", connID, message)

		// Process message based on type
		// For now, just echo the message back for testing
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}
