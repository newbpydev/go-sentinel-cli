package websocket

import (
	"sync"
	"testing"
)

func BenchmarkConnectionManager_AddRemove(b *testing.B) {
	cm := NewConnectionManager()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn := &mockConn{}
		id := cm.Add(conn)
		cm.Remove(id)
	}
}

func BenchmarkConnectionManager_ConcurrentAddRemove(b *testing.B) {
	cm := NewConnectionManager()
	b.ResetTimer()
	var wg sync.WaitGroup
	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn := &mockConn{}
			id := cm.Add(conn)
			cm.Remove(id)
		}()
	}
	wg.Wait()
}
