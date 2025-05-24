package cli

import (
	"testing"
	"time"
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

	event := FileEvent{
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
	debouncer.AddEvent(FileEvent{Path: "test.go", Type: "write"})
	debouncer.AddEvent(FileEvent{Path: "test.go", Type: "modify"})
	debouncer.AddEvent(FileEvent{Path: "test.go", Type: "write"}) // Should override previous

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
	debouncer.AddEvent(FileEvent{Path: "file1.go", Type: "write"})
	debouncer.AddEvent(FileEvent{Path: "file2.go", Type: "create"})
	debouncer.AddEvent(FileEvent{Path: "file3.go", Type: "remove"})

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
		debouncer.AddEvent(FileEvent{Path: "rapid.go", Type: "write"})
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
	debouncer.AddEvent(FileEvent{Path: "timer.go", Type: "write"})
	time.Sleep(50 * time.Millisecond) // Half the debounce interval
	debouncer.AddEvent(FileEvent{Path: "timer.go", Type: "modify"})

	// Assert - Should not receive event until full interval after last event
	select {
	case <-debouncer.Events():
		t.Error("Expected timer to be reset, but received event too early")
	case <-time.After(80 * time.Millisecond): // Should still be waiting
		// This is expected
	}

	// Now wait for the full interval and should receive event
	select {
	case events := <-debouncer.Events():
		if len(events) != 1 {
			t.Errorf("Expected 1 event after timer reset, got %d", len(events))
		}
		if events[0].Type != "modify" {
			t.Errorf("Expected type 'modify' (latest), got '%s'", events[0].Type)
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("Expected to receive event after timer reset")
	}
}

// TestFileEventDebouncer_Stop tests proper shutdown behavior
func TestFileEventDebouncer_Stop(t *testing.T) {
	// Arrange
	debouncer := NewFileEventDebouncer(50 * time.Millisecond)

	// Add an event
	debouncer.AddEvent(FileEvent{Path: "stop.go", Type: "write"})

	// Wait a bit to let any pending timer operations start
	time.Sleep(10 * time.Millisecond)

	// Act
	debouncer.Stop()

	// Assert - Events channel should be closed eventually
	select {
	case events, ok := <-debouncer.Events():
		if !ok {
			// Channel closed - this is expected after Stop()
			return
		}
		// If we get events, that's ok too (final flush)
		if len(events) > 1 {
			t.Errorf("Expected at most 1 event after stop, got %d", len(events))
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Expected events channel to be closed or receive final events")
	}

	// Adding events after stop should not panic and should be ignored
	debouncer.AddEvent(FileEvent{Path: "ignored.go", Type: "write"})

	// Multiple stops should not panic
	debouncer.Stop()
	debouncer.Stop()
}

// TestFileEventDebouncer_ConcurrentAccess tests concurrent access patterns
func TestFileEventDebouncer_ConcurrentAccess(t *testing.T) {
	// Arrange
	debouncer := NewFileEventDebouncer(100 * time.Millisecond) // Longer interval to reduce race conditions
	defer func() {
		// Drain events channel before stopping to prevent race conditions
		go func() {
			for range debouncer.Events() {
				// Drain events
			}
		}()
		time.Sleep(10 * time.Millisecond)
		debouncer.Stop()
	}()

	// Act - Add events concurrently from multiple goroutines
	done := make(chan bool, 3)

	// Goroutine 1: Add events for file1
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 5; i++ { // Reduced iterations
			debouncer.AddEvent(FileEvent{Path: "concurrent1.go", Type: "write"})
			time.Sleep(10 * time.Millisecond) // Slightly longer sleep
		}
	}()

	// Goroutine 2: Add events for file2
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 5; i++ { // Reduced iterations
			debouncer.AddEvent(FileEvent{Path: "concurrent2.go", Type: "modify"})
			time.Sleep(10 * time.Millisecond) // Slightly longer sleep
		}
	}()

	// Goroutine 3: Read events
	go func() {
		defer func() { done <- true }()
		select {
		case events := <-debouncer.Events():
			// Should receive some events (1 or 2 files depending on timing)
			if len(events) == 0 {
				t.Error("Expected at least one event from concurrent access")
			}
			if len(events) > 2 {
				t.Errorf("Expected at most 2 events (one per file), got %d", len(events))
			}
		case <-time.After(300 * time.Millisecond): // Longer timeout
			t.Error("Expected to receive events from concurrent access")
		}
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-time.After(1 * time.Second):
			t.Error("Timeout waiting for concurrent operations to complete")
		}
	}
}

// TestFileEventDebouncer_ChannelBlocking tests non-blocking behavior when channel is full
func TestFileEventDebouncer_ChannelBlocking(t *testing.T) {
	// Arrange
	debouncer := NewFileEventDebouncer(30 * time.Millisecond) // Longer interval
	defer func() {
		// Drain events before stopping
		go func() {
			for range debouncer.Events() {
				// Drain
			}
		}()
		time.Sleep(20 * time.Millisecond)
		debouncer.Stop()
	}()

	// Fill up the events channel by not reading from it and generating many batches
	for i := 0; i < 8; i++ { // Reduced count to be safer
		debouncer.AddEvent(FileEvent{Path: "blocking.go", Type: "write"})
		time.Sleep(35 * time.Millisecond) // Wait for debounce to trigger
	}

	// Act - Add more events (should not block even if channel is full)
	eventSent := false
	done := make(chan bool, 1)

	go func() {
		debouncer.AddEvent(FileEvent{Path: "nonblocking.go", Type: "write"})
		eventSent = true
		done <- true
	}()

	// Assert - Should complete quickly without blocking
	select {
	case <-done:
		if !eventSent {
			t.Error("Expected event to be sent without blocking")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("AddEvent appears to be blocking when channel is full")
	}
}

// TestFileEventDebouncer_EmptyFlush tests that empty flushes are handled correctly
func TestFileEventDebouncer_EmptyFlush(t *testing.T) {
	// Arrange
	debouncer := NewFileEventDebouncer(30 * time.Millisecond)
	defer debouncer.Stop()

	// Act - Wait for potential flush without adding events
	select {
	case events := <-debouncer.Events():
		t.Errorf("Expected no events when nothing was added, got %d events", len(events))
	case <-time.After(100 * time.Millisecond):
		// This is expected - no events should be sent
	}
}

// TestFileEventDebouncer_EventOverride tests that newer events override older ones for same path
func TestFileEventDebouncer_EventOverride(t *testing.T) {
	// Arrange
	debouncer := NewFileEventDebouncer(50 * time.Millisecond)
	defer debouncer.Stop()

	// Act - Add events with different types for the same path
	debouncer.AddEvent(FileEvent{Path: "override.go", Type: "create"})
	debouncer.AddEvent(FileEvent{Path: "override.go", Type: "write"})
	debouncer.AddEvent(FileEvent{Path: "override.go", Type: "remove"})

	// Assert
	select {
	case events := <-debouncer.Events():
		if len(events) != 1 {
			t.Errorf("Expected 1 event after override, got %d", len(events))
		}
		if events[0].Type != "remove" {
			t.Errorf("Expected type 'remove' (last override), got '%s'", events[0].Type)
		}
		if events[0].Path != "override.go" {
			t.Errorf("Expected path 'override.go', got '%s'", events[0].Path)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Expected to receive overridden event within timeout")
	}
}
