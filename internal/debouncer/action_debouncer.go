// Package debouncer provides functionality for debouncing actions and events
package debouncer

import (
	"sync"
	"time"
)

// ActionDebouncer debounces actions by a key
type ActionDebouncer struct {
	duration time.Duration
	mutex    sync.Mutex
	timers   map[string]*time.Timer
}

// NewActionDebouncer creates a new action debouncer with the specified timeout
func NewActionDebouncer(duration time.Duration) *ActionDebouncer {
	return &ActionDebouncer{
		duration: duration,
		timers:   make(map[string]*time.Timer),
	}
}

// Debounce executes the callback after the debounce interval if no other
// call with the same key happens in the meantime
func (d *ActionDebouncer) Debounce(key string, callback func()) {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Cancel existing timer if there is one
	if timer, exists := d.timers[key]; exists {
		timer.Stop()
		delete(d.timers, key)
	}

	// Create new timer
	timer := time.AfterFunc(d.duration, func() {
		// Remove the timer from the map when it fires
		d.mutex.Lock()
		delete(d.timers, key)
		d.mutex.Unlock()

		// Execute the callback
		callback()
	})

	// Store the timer
	d.timers[key] = timer
}

// Clear cancels all pending debounced actions
func (d *ActionDebouncer) Clear() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	// Stop all timers
	for key, timer := range d.timers {
		timer.Stop()
		delete(d.timers, key)
	}
}
