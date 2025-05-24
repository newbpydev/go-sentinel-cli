package debouncer

import (
	"sync"
	"time"

	"github.com/newbpydev/go-sentinel/internal/watch/core"
)

// FileEventDebouncer batches file events to avoid running tests too frequently
// Implements the core.EventDebouncer interface
type FileEventDebouncer struct {
	interval time.Duration
	events   chan []core.FileEvent
	input    chan core.FileEvent
	pending  map[string]core.FileEvent // Use map to deduplicate events by path
	timer    *time.Timer
	mutex    sync.Mutex
	stopCh   chan struct{}
	stopped  bool
}

// NewFileEventDebouncer creates a new file event debouncer
func NewFileEventDebouncer(interval time.Duration) *FileEventDebouncer {
	if interval <= 0 {
		interval = 250 * time.Millisecond // Default debounce interval
	}

	d := &FileEventDebouncer{
		interval: interval,
		events:   make(chan []core.FileEvent, 10),
		input:    make(chan core.FileEvent, 100),
		pending:  make(map[string]core.FileEvent),
		stopCh:   make(chan struct{}),
	}

	go d.run()
	return d
}

// AddEvent adds a file event to be debounced
// Implements core.EventDebouncer.AddEvent
func (d *FileEventDebouncer) AddEvent(event core.FileEvent) {
	select {
	case d.input <- event:
	case <-d.stopCh:
		// Debouncer is stopped, ignore new events
	}
}

// Events returns the channel for debounced events
// Implements core.EventDebouncer.Events
func (d *FileEventDebouncer) Events() <-chan []core.FileEvent {
	return d.events
}

// SetInterval configures the debounce interval
// Implements core.EventDebouncer.SetInterval
func (d *FileEventDebouncer) SetInterval(interval time.Duration) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if interval <= 0 {
		interval = 250 * time.Millisecond
	}

	d.interval = interval
}

// Stop stops the debouncer and closes channels
// Implements core.EventDebouncer.Stop
func (d *FileEventDebouncer) Stop() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.stopped {
		return nil
	}

	d.stopped = true
	close(d.stopCh)
	close(d.input)

	// Stop the timer if it's running
	if d.timer != nil {
		d.timer.Stop()
	}

	// Close the events channel after a brief delay to allow final events to be sent
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(d.events)
	}()

	return nil
}

// run is the main debouncer loop
func (d *FileEventDebouncer) run() {
	for {
		select {
		case event, ok := <-d.input:
			if !ok {
				// Input channel closed, flush any pending events and exit
				d.flushPendingEvents()
				return
			}

			d.mutex.Lock()
			// Add or update the event (newer events for the same path override older ones)
			d.pending[event.Path] = event

			// Reset the timer
			if d.timer != nil {
				d.timer.Stop()
			}
			d.timer = time.AfterFunc(d.interval, func() {
				d.flushPendingEvents()
			})
			d.mutex.Unlock()

		case <-d.stopCh:
			return
		}
	}
}

// flushPendingEvents sends all pending events and clears the pending map
func (d *FileEventDebouncer) flushPendingEvents() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if len(d.pending) == 0 || d.stopped {
		return
	}

	// Convert map to slice
	events := make([]core.FileEvent, 0, len(d.pending))
	for _, event := range d.pending {
		events = append(events, event)
	}

	// Clear pending events
	d.pending = make(map[string]core.FileEvent)

	// Send events (non-blocking) only if not stopped
	if !d.stopped {
		select {
		case d.events <- events:
		case <-d.stopCh:
			// Debouncer is stopped, don't send events
		default:
			// Channel is full, skip this batch
			// This prevents blocking if the consumer is slow
		}
	}
}

// Ensure FileEventDebouncer implements the EventDebouncer interface
var _ core.EventDebouncer = (*FileEventDebouncer)(nil)
