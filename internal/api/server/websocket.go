package server

import (
	"log"
	"net/http"

	gorillaws "github.com/gorilla/websocket"
	gosentinel "github.com/newbpydev/go-sentinel/internal/api/websocket"
)

// WebSocketConnection implements the gosentinel.Connection interface
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

// WebSocket connection upgrader with default settings
var upgrader = gorillaws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow connections from any origin for development
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// SetupWebSocketHandlers registers WebSocket handlers in the router
func SetupWebSocketHandlers(r http.Handler) {
	// This function is called from NewAPIServer
	log.Println("Setting up WebSocket handlers")
}

// HandleWebSocketConnection handles a new WebSocket connection request
func HandleWebSocketConnection(w http.ResponseWriter, r *http.Request, connManager *gosentinel.ConnectionManager) {
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to WebSocket: %v", err)
		return
	}
	
	// Create a connection wrapper for the manager
	wsConn := NewWebSocketConnection(conn)
	
	// Register with connection manager
	connID := connManager.Add(wsConn)
	log.Printf("New WebSocket connection established: %s", connID)
	
	// Handle incoming messages in a goroutine
	go handleWSMessages(conn, connID, connManager)
}

// handleWSMessages processes incoming WebSocket messages
func handleWSMessages(conn *gorillaws.Conn, connID string, connManager *gosentinel.ConnectionManager) {
	defer func() {
		if err := conn.Close(); err != nil {
		log.Printf("websocket close error: %v", err)
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

		// Process the message
		log.Printf("Received WebSocket message from %s: %s", connID, message)

		// Echo the message back for now (testing)
		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Printf("WebSocket write error: %v", err)
			break
		}
	}
}
