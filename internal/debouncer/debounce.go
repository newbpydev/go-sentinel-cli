package debouncer

import (
	"time"
)

type Debouncer struct {
	interval time.Duration
	in       chan string
	out      chan string
	quit     chan struct{}
}

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

func (d *Debouncer) Emit(pkg string) {
	d.in <- pkg
}

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
