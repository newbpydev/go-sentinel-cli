package websocket

import (
	"sync"
	"testing"
)

type mockConn struct {
	closed bool
	mu     sync.Mutex
}

func (m *mockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}
func (m *mockConn) IsClosed() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.closed
}

func TestConnectionManager_TrackAndCreate(t *testing.T) {
	cm := NewConnectionManager()
	conn := &mockConn{}
	id := cm.Add(conn)
	if cm.Count() != 1 {
		t.Errorf("expected 1 connection, got %d", cm.Count())
	}
	if cm.Get(id) == nil {
		t.Errorf("expected to retrieve connection by id")
	}
}

func TestConnectionManager_CleanupOnDisconnect(t *testing.T) {
	cm := NewConnectionManager()
	conn := &mockConn{}
	id := cm.Add(conn)
	cm.Remove(id)
	if cm.Count() != 0 {
		t.Errorf("expected 0 connections after removal, got %d", cm.Count())
	}
	// Remove should call Close
	if !conn.IsClosed() {
		t.Errorf("expected connection to be closed on removal")
	}
}

func TestConnectionManager_GoroutineSafety(t *testing.T) {
	cm := NewConnectionManager()
	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn := &mockConn{}
			id := cm.Add(conn)
			cm.Remove(id)
		}()
	}
	wg.Wait()
	if cm.Count() != 0 {
		t.Errorf("expected 0 connections after concurrent add/remove, got %d", cm.Count())
	}
}
