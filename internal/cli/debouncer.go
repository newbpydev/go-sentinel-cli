package cli

import (
	"sync"
	"time"
)

// FileEventDebouncer batches file events to avoid running tests too frequently
type FileEventDebouncer struct {
	interval time.Duration
	events   chan []FileEvent
	input    chan FileEvent
	pending  map[string]FileEvent // Use map to deduplicate events by path
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
		events:   make(chan []FileEvent, 10),
		input:    make(chan FileEvent, 100),
		pending:  make(map[string]FileEvent),
		stopCh:   make(chan struct{}),
	}

	go d.run()
	return d
}

// AddEvent adds a file event to be debounced
func (d *FileEventDebouncer) AddEvent(event FileEvent) {
	select {
	case d.input <- event:
	case <-d.stopCh:
		// Debouncer is stopped, ignore new events
	}
}

// Events returns the channel for debounced events
func (d *FileEventDebouncer) Events() <-chan []FileEvent {
	return d.events
}

// Stop stops the debouncer and closes channels
func (d *FileEventDebouncer) Stop() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.stopped {
		return
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
	events := make([]FileEvent, 0, len(d.pending))
	for _, event := range d.pending {
		events = append(events, event)
	}

	// Clear pending events
	d.pending = make(map[string]FileEvent)

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
