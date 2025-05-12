package websocket

import (
	"sync"
	"time"
)

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

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		conns: make(map[string]broadcastConn),
		throttle: 0,
	}
}

func (b *Broadcaster) Add(conn broadcastConn) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.idSeq++
	id := generateID(b.idSeq)
	b.conns[id] = conn
	return id
}

func (b *Broadcaster) Remove(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if c, ok := b.conns[id]; ok {
		c.Close()
		delete(b.conns, id)
	}
}

func (b *Broadcaster) Broadcast(msg []byte) {
	b.mu.RLock()
	conns := make([]broadcastConn, 0, len(b.conns))
	for _, c := range b.conns {
		conns = append(conns, c)
	}
	throttle := b.throttle
	b.mu.RUnlock()

	for _, c := range conns {
		c.Send(msg)
	}
	if throttle > 0 {
		time.Sleep(throttle)
	}
}

func (b *Broadcaster) SetThrottle(d time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.throttle = d
}
