package debouncer

import (
	"time"
)

// Debouncer provides a mechanism to debounce string events by a specified interval.
type Debouncer struct {
	interval time.Duration
	in       chan string
	out      chan string
	quit     chan struct{}
}

// NewDebouncer creates a new Debouncer with the given interval.
func NewDebouncer(d time.Duration) *Debouncer {
	deb := &Debouncer{
		interval: d,
		in:       make(chan string, 32),
		out:      make(chan string, 32),
		quit:     make(chan struct{}),
	}
	go deb.loop()
	return deb
}

// Emit sends a string event to the debouncer.
func (d *Debouncer) Emit(pkg string) {
	d.in <- pkg
}

// Events returns a channel that receives debounced string events.
func (d *Debouncer) Events() <-chan string {
	return d.out
}

func (d *Debouncer) loop() {
	timers := make(map[string]*time.Timer)
	pending := make(map[string]struct{})
	for {
		select {
		case pkg := <-d.in:
			if t, ok := timers[pkg]; ok {
				t.Stop()
			}
			pending[pkg] = struct{}{}
			timers[pkg] = time.AfterFunc(d.interval, func(p string) func() {
				return func() {
					d.out <- p
					delete(pending, p)
				}
			}(pkg))
		case <-d.quit:
			for _, t := range timers {
				t.Stop()
			}
			return
		}
	}
}

// TODO: Implement event debouncing logic
