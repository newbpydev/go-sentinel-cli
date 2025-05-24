// Package debouncer provides event temporal processing capabilities
package debouncer

import (
	"sync"
	"time"

	"github.com/newbpydev/go-sentinel/internal/watch/core"
)

// Debouncer implements the EventDebouncer interface
type Debouncer struct {
	mu            sync.RWMutex
	interval      time.Duration
	events        chan []core.FileEvent
	pendingEvents map[string]core.FileEvent
	timer         *time.Timer
	stopCh        chan struct{}
	stopped       bool
}

// NewDebouncer creates a new event debouncer
func NewDebouncer(interval time.Duration) core.EventDebouncer {
	return &Debouncer{
		interval:      interval,
		events:        make(chan []core.FileEvent, 10),
		pendingEvents: make(map[string]core.FileEvent),
		stopCh:        make(chan struct{}),
		stopped:       false,
	}
}

// AddEvent implements the EventDebouncer interface
func (d *Debouncer) AddEvent(event core.FileEvent) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Don't add events if stopped
	if d.stopped {
		return
	}

	// Store the latest event for this path
	d.pendingEvents[event.Path] = event

	// Reset or start the timer
	if d.timer != nil {
		d.timer.Stop()
	}

	d.timer = time.AfterFunc(d.interval, func() {
		d.flushPendingEvents()
	})
}

// Events implements the EventDebouncer interface
func (d *Debouncer) Events() <-chan []core.FileEvent {
	return d.events
}

// SetInterval implements the EventDebouncer interface
func (d *Debouncer) SetInterval(interval time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.interval = interval
}

// Stop implements the EventDebouncer interface
func (d *Debouncer) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.stopped {
		return nil
	}

	d.stopped = true

	// Stop the timer if it exists
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}

	// Flush any remaining events
	if len(d.pendingEvents) > 0 {
		events := make([]core.FileEvent, 0, len(d.pendingEvents))
		for _, event := range d.pendingEvents {
			events = append(events, event)
		}

		// Send final events safely
		select {
		case d.events <- events:
		default:
			// Channel might be full, that's ok
		}

		// Clear pending events
		d.pendingEvents = make(map[string]core.FileEvent)
	}

	// Close the stop channel
	close(d.stopCh)

	// Close the events channel in a separate goroutine to avoid blocking
	go func() {
		time.Sleep(10 * time.Millisecond) // Small delay to ensure final events are processed
		close(d.events)
	}()

	return nil
}

// flushPendingEvents sends all pending events and clears the buffer
func (d *Debouncer) flushPendingEvents() {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Don't flush if stopped
	if d.stopped {
		return
	}

	if len(d.pendingEvents) == 0 {
		return
	}

	// Convert map to slice
	events := make([]core.FileEvent, 0, len(d.pendingEvents))
	for _, event := range d.pendingEvents {
		events = append(events, event)
	}

	// Clear pending events
	d.pendingEvents = make(map[string]core.FileEvent)

	// Send events safely
	select {
	case d.events <- events:
	case <-d.stopCh:
		// Debouncer was stopped, don't try to send
		return
	default:
		// Channel might be full, that's ok for now
	}
}
