// Package websocket provides WebSocket communication functionality for the Go Sentinel application.
// It includes broadcasting capabilities, connection management, and message type handling.
package websocket

import (
	"log"
	"sync"
	"time"
)

// Broadcaster manages a collection of WebSocket connections and provides methods
// for broadcasting messages to all connected clients with optional throttling.
type Broadcaster struct {
	mu       sync.RWMutex
	conns    map[string]broadcastConn
	idSeq    int
	throttle time.Duration
}

type broadcastConn interface {
	Send([]byte) error
	Close() error
}

// NewBroadcaster creates a new Broadcaster instance with an empty connection pool.
// The broadcaster is used to manage WebSocket connections and broadcast messages to all clients.
func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		conns:    make(map[string]broadcastConn),
		throttle: 0,
	}
}

// Add registers a new WebSocket connection with the broadcaster and returns
// a unique identifier for the connection that can be used later for removal.
func (b *Broadcaster) Add(conn broadcastConn) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.idSeq++
	id := generateID(b.idSeq)
	b.conns[id] = conn
	return id
}

// Remove unregisters and closes a WebSocket connection identified by the given ID.
// If the connection is not found, this operation is a no-op.
func (b *Broadcaster) Remove(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if c, ok := b.conns[id]; ok {
		if err := c.Close(); err != nil {
			// Log error but continue with removal
			log.Printf("Error closing connection %s: %v", id, err)
		}
		delete(b.conns, id)
	}
}

// Broadcast sends a message to all registered WebSocket connections.
// If throttling is enabled, it will wait for the throttle duration between broadcasts.
func (b *Broadcaster) Broadcast(msg []byte) {
	b.mu.RLock()
	conns := make([]broadcastConn, 0, len(b.conns))
	for _, c := range b.conns {
		conns = append(conns, c)
	}
	throttle := b.throttle
	b.mu.RUnlock()

	sem := make(chan struct{}, 16) // Limit concurrency to 16
	var wg sync.WaitGroup
	for _, c := range conns {
		sem <- struct{}{}
		wg.Add(1)
		go func(conn broadcastConn) {
			defer wg.Done()
			// Allocate a new buffer for each message to avoid data races
			buf := make([]byte, len(msg))
			copy(buf, msg)
			if err := conn.Send(buf); err != nil {
				// Log error but continue broadcasting to other connections
				log.Printf("Error sending message to client: %v", err)
			}
			<-sem
		}(c)
	}
	wg.Wait()
	if throttle > 0 {
		time.Sleep(throttle)
	}
}

// SetThrottle configures the delay between broadcast operations to prevent flooding clients.
// A duration of 0 means no throttling is applied.
func (b *Broadcaster) SetThrottle(d time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.throttle = d
}
