package websocket

import (
	"fmt"
	"sync"
)

type Connection interface {
	Close() error
}

type ConnectionManager struct {
	mu    sync.RWMutex
	conns map[string]Connection
	idSeq int
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		conns: make(map[string]Connection),
	}
}

func (cm *ConnectionManager) Add(conn Connection) string {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.idSeq++
	id := generateID(cm.idSeq)
	cm.conns[id] = conn
	return id
}

func (cm *ConnectionManager) Get(id string) Connection {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.conns[id]
}

func (cm *ConnectionManager) Remove(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if conn, ok := cm.conns[id]; ok {
		if err := conn.Close(); err != nil {
			// Log but ignore error on close
		}
		delete(cm.conns, id)
	}
}

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
