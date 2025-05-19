package websocket

import (
	"fmt"
	"log"
	"sync"
)

// Connection represents a WebSocket connection that can be closed.
// This interface allows the connection manager to work with different WebSocket implementations.
type Connection interface {
	Close() error
}

// ConnectionManager provides thread-safe management of WebSocket connections.
// It allows adding, retrieving, and removing connections with unique identifiers.
type ConnectionManager struct {
	mu    sync.RWMutex
	conns map[string]Connection
	idSeq int
}

// NewConnectionManager creates a new ConnectionManager instance with an empty connection pool.
// The connection manager is used to track and manage active WebSocket connections.
func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		conns: make(map[string]Connection),
	}
}

// Add registers a new WebSocket connection with the connection manager and returns
// a unique identifier for the connection that can be used later for retrieval or removal.
func (cm *ConnectionManager) Add(conn Connection) string {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.idSeq++
	id := generateID(cm.idSeq)
	cm.conns[id] = conn
	return id
}

// Get retrieves a WebSocket connection by its unique identifier.
// Returns nil if the connection is not found.
func (cm *ConnectionManager) Get(id string) Connection {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.conns[id]
}

// Remove unregisters and closes a WebSocket connection identified by the given ID.
// If the connection is not found, this operation is a no-op.
func (cm *ConnectionManager) Remove(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if conn, ok := cm.conns[id]; ok {
		if err := conn.Close(); err != nil {
			// Log error but continue with removal
			log.Printf("Error closing connection %s: %v", id, err)
		}
		delete(cm.conns, id)
	}
}

// Count returns the number of active WebSocket connections currently managed.
func (cm *ConnectionManager) Count() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.conns)
}

func generateID(seq int) string {
	return "conn-" + itoa(seq)
}

func itoa(i int) string {
	// Simple integer to string conversion
	return fmt.Sprintf("%d", i)
}
