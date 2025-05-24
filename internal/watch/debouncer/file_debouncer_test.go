package debouncer

import (
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/watch/core"
)

// TestNewFileEventDebouncer_Creation verifies debouncer initialization
func TestNewFileEventDebouncer_Creation(t *testing.T) {
	// Test cases for different intervals
	testCases := []struct {
		name             string
		interval         time.Duration
		expectedInterval time.Duration
	}{
		{
			name:             "Valid interval",
			interval:         500 * time.Millisecond,
			expectedInterval: 500 * time.Millisecond,
		},
		{
			name:             "Zero interval uses default",
			interval:         0,
			expectedInterval: 250 * time.Millisecond,
		},
		{
			name:             "Negative interval uses default",
			interval:         -100 * time.Millisecond,
			expectedInterval: 250 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			debouncer := NewFileEventDebouncer(tc.interval)
			defer debouncer.Stop()

			// Assert
			if debouncer == nil {
				t.Fatal("Expected debouncer to be created, got nil")
			}
			if debouncer.interval != tc.expectedInterval {
				t.Errorf("Expected interval %v, got %v", tc.expectedInterval, debouncer.interval)
			}
			if debouncer.events == nil {
				t.Error("Expected events channel to be initialized")
			}
			if debouncer.input == nil {
				t.Error("Expected input channel to be initialized")
			}
			if debouncer.pending == nil {
				t.Error("Expected pending map to be initialized")
			}
		})
	}
}

// TestFileEventDebouncer_SingleEvent tests basic event processing
func TestFileEventDebouncer_SingleEvent(t *testing.T) {
	// Arrange
	debouncer := NewFileEventDebouncer(50 * time.Millisecond)
	defer debouncer.Stop()

	event := core.FileEvent{
		Path: "test.go",
		Type: "write",
	}

	// Act
	debouncer.AddEvent(event)

	// Assert
	select {
	case events := <-debouncer.Events():
		if len(events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(events))
		}
		if events[0].Path != "test.go" {
			t.Errorf("Expected path 'test.go', got '%s'", events[0].Path)
		}
		if events[0].Type != "write" {
			t.Errorf("Expected type 'write', got '%s'", events[0].Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive debounced event within timeout")
	}
}

// TestFileEventDebouncer_EventDeduplication tests that duplicate events are deduplicated
func TestFileEventDebouncer_EventDeduplication(t *testing.T) {
	// Arrange
	debouncer := NewFileEventDebouncer(50 * time.Millisecond)
	defer debouncer.Stop()

	// Act - Add multiple events for the same file
	debouncer.AddEvent(core.FileEvent{Path: "test.go", Type: "write"})
	debouncer.AddEvent(core.FileEvent{Path: "test.go", Type: "modify"})
	debouncer.AddEvent(core.FileEvent{Path: "test.go", Type: "write"}) // Should override previous

	// Assert
	select {
	case events := <-debouncer.Events():
		if len(events) != 1 {
			t.Errorf("Expected 1 deduplicated event, got %d", len(events))
		}
		// Should have the last event for the path
		if events[0].Type != "write" {
			t.Errorf("Expected type 'write' (last event), got '%s'", events[0].Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive debounced event within timeout")
	}
}

// TestFileEventDebouncer_MultipleFiles tests handling multiple different files
func TestFileEventDebouncer_MultipleFiles(t *testing.T) {
	// Arrange
	debouncer := NewFileEventDebouncer(50 * time.Millisecond)
	defer debouncer.Stop()

	// Act - Add events for different files
	debouncer.AddEvent(core.FileEvent{Path: "file1.go", Type: "write"})
	debouncer.AddEvent(core.FileEvent{Path: "file2.go", Type: "create"})
	debouncer.AddEvent(core.FileEvent{Path: "file3.go", Type: "remove"})

	// Assert
	select {
	case events := <-debouncer.Events():
		if len(events) != 3 {
			t.Errorf("Expected 3 events for different files, got %d", len(events))
		}

		// Verify all files are present (order may vary due to map iteration)
		filePaths := make(map[string]bool)
		for _, event := range events {
			filePaths[event.Path] = true
		}

		expectedFiles := []string{"file1.go", "file2.go", "file3.go"}
		for _, expectedFile := range expectedFiles {
			if !filePaths[expectedFile] {
				t.Errorf("Expected file '%s' in events, but not found", expectedFile)
			}
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive debounced events within timeout")
	}
}

// TestFileEventDebouncer_RapidEvents tests that rapid events are properly debounced
func TestFileEventDebouncer_RapidEvents(t *testing.T) {
	// Arrange
	debouncer := NewFileEventDebouncer(100 * time.Millisecond)
	defer debouncer.Stop()

	// Act - Add events rapidly
	for i := 0; i < 10; i++ {
		debouncer.AddEvent(core.FileEvent{Path: "rapid.go", Type: "write"})
		time.Sleep(10 * time.Millisecond) // Faster than debounce interval
	}

	// Assert - Should receive only one batch of events
	eventBatches := 0
	timeout := time.After(200 * time.Millisecond)

	for {
		select {
		case events := <-debouncer.Events():
			eventBatches++
			if len(events) != 1 {
				t.Errorf("Expected 1 event in batch, got %d", len(events))
			}
			if events[0].Path != "rapid.go" {
				t.Errorf("Expected path 'rapid.go', got '%s'", events[0].Path)
			}
		case <-timeout:
			// Check that we received exactly one batch
			if eventBatches != 1 {
				t.Errorf("Expected exactly 1 event batch, got %d", eventBatches)
			}
			return
		}
	}
}

// TestFileEventDebouncer_TimerReset tests that timer is properly reset with new events
func TestFileEventDebouncer_TimerReset(t *testing.T) {
	// Arrange
	debouncer := NewFileEventDebouncer(100 * time.Millisecond)
	defer debouncer.Stop()

	// Act - Add event, wait halfway, add another event
	debouncer.AddEvent(core.FileEvent{Path: "timer.go", Type: "write"})
	time.Sleep(50 * time.Millisecond) // Half the debounce interval
	debouncer.AddEvent(core.FileEvent{Path: "timer.go", Type: "modify"})

	// Assert - Should not receive event until full interval after last event
	select {
	case <-debouncer.Events():
		t.Error("Should not receive event before full debounce interval")
	case <-time.After(75 * time.Millisecond): // 75ms total, should not have triggered yet
		// Good, no event received yet
	}

	// Now wait for the full interval and should receive event
	select {
	case events := <-debouncer.Events():
		if len(events) != 1 {
			t.Errorf("Expected 1 event, got %d", len(events))
		}
		if events[0].Type != "modify" {
			t.Errorf("Expected type 'modify' (last event), got '%s'", events[0].Type)
		}
	case <-time.After(50 * time.Millisecond): // Additional 50ms should be enough
		t.Error("Expected to receive event after full debounce interval")
	}
}

// TestFileEventDebouncer_Stop tests proper cleanup when stopping
func TestFileEventDebouncer_Stop(t *testing.T) {
	// Arrange
	debouncer := NewFileEventDebouncer(50 * time.Millisecond)

	// Act
	err := debouncer.Stop()

	// Assert
	if err != nil {
		t.Errorf("Expected no error when stopping, got: %v", err)
	}

	// Verify that adding events after stop doesn't block or panic
	debouncer.AddEvent(core.FileEvent{Path: "test.go", Type: "write"})

	// Verify that stopping again doesn't cause issues
	err = debouncer.Stop()
	if err != nil {
		t.Errorf("Expected no error when stopping again, got: %v", err)
	}

	// Verify events channel is eventually closed
	select {
	case _, ok := <-debouncer.Events():
		if ok {
			t.Error("Expected events channel to be closed after stop")
		}
	case <-time.After(50 * time.Millisecond):
		// Channel might not be closed immediately, that's okay
	}
}

// TestFileEventDebouncer_ConcurrentAccess tests thread safety
func TestFileEventDebouncer_ConcurrentAccess(t *testing.T) {
	// Arrange
	debouncer := NewFileEventDebouncer(100 * time.Millisecond)
	defer debouncer.Stop()

	// Act - Add events concurrently
	done := make(chan bool, 3)

	// Goroutine 1: Add events
	go func() {
		for i := 0; i < 50; i++ {
			debouncer.AddEvent(core.FileEvent{Path: "concurrent1.go", Type: "write"})
		}
		done <- true
	}()

	// Goroutine 2: Add different events
	go func() {
		for i := 0; i < 50; i++ {
			debouncer.AddEvent(core.FileEvent{Path: "concurrent2.go", Type: "create"})
		}
		done <- true
	}()

	// Goroutine 3: Read events
	go func() {
		eventCount := 0
		timeout := time.After(300 * time.Millisecond)
		for {
			select {
			case events := <-debouncer.Events():
				eventCount += len(events)
			case <-timeout:
				// Should have received some events
				if eventCount == 0 {
					t.Error("Expected to receive some events during concurrent access")
				}
				done <- true
				return
			}
		}
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}

// TestFileEventDebouncer_ChannelBlocking tests behavior when events channel is full
func TestFileEventDebouncer_ChannelBlocking(t *testing.T) {
	// Arrange
	debouncer := NewFileEventDebouncer(50 * time.Millisecond)
	defer debouncer.Stop()

	// Fill up the events channel by not reading from it
	for i := 0; i < 15; i++ { // More than channel buffer size (10)
		debouncer.AddEvent(core.FileEvent{Path: "blocking.go", Type: "write"})
		time.Sleep(60 * time.Millisecond) // Wait for debounce to trigger
	}

	// Act - Add one more event (should not block due to non-blocking send)
	debouncer.AddEvent(core.FileEvent{Path: "final.go", Type: "write"})

	// Assert - Should not hang, and we should be able to read events
	eventBatches := 0
	timeout := time.After(200 * time.Millisecond)

	for {
		select {
		case <-debouncer.Events():
			eventBatches++
		case <-timeout:
			// Should have received some events (channel was full)
			if eventBatches == 0 {
				t.Error("Expected to receive some events even when channel was full")
			}
			return
		}
	}
}

// TestFileEventDebouncer_EmptyFlush tests that empty flushes don't send events
func TestFileEventDebouncer_EmptyFlush(t *testing.T) {
	// Arrange
	debouncer := NewFileEventDebouncer(50 * time.Millisecond)
	defer debouncer.Stop()

	// Act - Don't add any events, just wait
	select {
	case <-debouncer.Events():
		t.Error("Should not receive any events when none were added")
	case <-time.After(100 * time.Millisecond):
		// Good, no events received
	}
}

// TestFileEventDebouncer_EventOverride tests that newer events override older ones for same path
func TestFileEventDebouncer_EventOverride(t *testing.T) {
	// Arrange
	debouncer := NewFileEventDebouncer(50 * time.Millisecond)
	defer debouncer.Stop()

	// Act - Add events for same path with different types
	debouncer.AddEvent(core.FileEvent{Path: "override.go", Type: "create", Timestamp: time.Now()})
	time.Sleep(10 * time.Millisecond)
	debouncer.AddEvent(core.FileEvent{Path: "override.go", Type: "write", Timestamp: time.Now()})
	time.Sleep(10 * time.Millisecond)
	debouncer.AddEvent(core.FileEvent{Path: "override.go", Type: "remove", Timestamp: time.Now()})

	// Assert
	select {
	case events := <-debouncer.Events():
		if len(events) != 1 {
			t.Errorf("Expected 1 event (overridden), got %d", len(events))
		}
		if events[0].Type != "remove" {
			t.Errorf("Expected type 'remove' (last event), got '%s'", events[0].Type)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive debounced event within timeout")
	}
}

// TestFileEventDebouncer_SetInterval tests the SetInterval method
func TestFileEventDebouncer_SetInterval(t *testing.T) {
	// Arrange
	debouncer := NewFileEventDebouncer(100 * time.Millisecond)
	defer debouncer.Stop()

	// Act
	debouncer.SetInterval(200 * time.Millisecond)

	// Assert - Check that the interval was updated
	if debouncer.interval != 200*time.Millisecond {
		t.Errorf("Expected interval to be updated to 200ms, got %v", debouncer.interval)
	}

	// Test with invalid interval
	debouncer.SetInterval(-50 * time.Millisecond)
	if debouncer.interval != 250*time.Millisecond {
		t.Errorf("Expected invalid interval to default to 250ms, got %v", debouncer.interval)
	}
}
