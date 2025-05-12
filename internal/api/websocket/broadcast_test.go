package websocket

import (
	"sync"
	"testing"
	"time"
)

type fakeConn struct {
	mu sync.Mutex
	msgs [][]byte
	closed bool
}
func (f *fakeConn) Close() error {
	f.mu.Lock(); defer f.mu.Unlock(); f.closed = true; return nil }
func (f *fakeConn) Send(msg []byte) error { f.mu.Lock(); defer f.mu.Unlock(); f.msgs = append(f.msgs, msg); return nil }
func (f *fakeConn) Last() []byte { f.mu.Lock(); defer f.mu.Unlock(); if len(f.msgs)==0 {return nil}; return f.msgs[len(f.msgs)-1] }
func (f *fakeConn) Count() int { f.mu.Lock(); defer f.mu.Unlock(); return len(f.msgs) }
func (f *fakeConn) IsClosed() bool { f.mu.Lock(); defer f.mu.Unlock(); return f.closed }

func TestBroadcastToAllConnections(t *testing.T) {
	b := NewBroadcaster()
	conns := []*fakeConn{{}, {}, {}}
	ids := []string{}
	for _, c := range conns {
		ids = append(ids, b.Add(c))
	}
	msg := []byte("test-result")
	b.Broadcast(msg)
	for i, c := range conns {
		if c.Count() != 1 || string(c.Last()) != "test-result" {
			t.Errorf("conn %d did not receive broadcast", i)
		}
	}
}

func TestBroadcastThrottling(t *testing.T) {
	b := NewBroadcaster()
	c := &fakeConn{}
	b.Add(c)
	b.SetThrottle(10*time.Millisecond)
	for i := 0; i < 5; i++ {
		b.Broadcast([]byte("msg"))
	}
	time.Sleep(50*time.Millisecond)
	count := c.Count()
	if count < 2 || count > 5 {
		t.Errorf("unexpected throttle count: %d", count)
	}
}

func TestMessageOrderingAndDelivery(t *testing.T) {
	b := NewBroadcaster()
	c := &fakeConn{}
	b.Add(c)
	b.Broadcast([]byte("first"))
	b.Broadcast([]byte("second"))
	b.Broadcast([]byte("third"))
	time.Sleep(10*time.Millisecond)
	if c.Count() != 3 {
		t.Errorf("expected 3 messages, got %d", c.Count())
	}
	if string(c.msgs[0]) != "first" || string(c.msgs[2]) != "third" {
		t.Errorf("messages out of order: %+v", c.msgs)
	}
}
