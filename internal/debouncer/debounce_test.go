package debouncer

import (
	"testing"
	"time"
)

func TestBufferRapidEventsPerPackage(t *testing.T) {
	debounce := NewDebouncer(50 * time.Millisecond)
	ch := debounce.Events()
	
	go func() {
		debounce.Emit("pkgA")
		debounce.Emit("pkgA")
		debounce.Emit("pkgA")
	}()
	
	var count int
	for {
		select {
		case pkg := <-ch:
			if pkg != "pkgA" {
				t.Errorf("expected pkgA, got %v", pkg)
			}
			count++
		case <-time.After(200 * time.Millisecond):
			if count != 1 {
				t.Errorf("expected 1 debounced event, got %d", count)
			}
			return
		}
	}
}

func TestTriggerAfterQuietPeriod(t *testing.T) {
	debounce := NewDebouncer(30 * time.Millisecond)
	ch := debounce.Events()
	
	start := time.Now()
	debounce.Emit("pkgB")
	
	select {
	case <-ch:
		elapsed := time.Since(start)
		if elapsed < 25*time.Millisecond {
			t.Errorf("debounced too soon: %v", elapsed)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("timeout waiting for debounced event")
	}
}
