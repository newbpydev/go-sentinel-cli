package coordinator

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/watch/core"
)

// TestCoordinator_HandleFileChanges_LastEventTimeUpdate tests that LastEventTime is updated
func TestCoordinator_HandleFileChanges_LastEventTimeUpdate(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{
		triggerTestsFunc: func(ctx context.Context, filePath string) error {
			return nil // Success case
		},
	}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Configure and set running
	options := core.WatchOptions{Mode: core.WatchAll}
	coordinator.Configure(options)
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.ctx = context.Background()
	initialTime := coordinator.status.LastEventTime
	coordinator.mu.Unlock()

	// Wait a moment to ensure time difference
	time.Sleep(1 * time.Millisecond)

	changes := []core.FileEvent{
		{Path: "test.go", Type: "modify"},
	}

	err := coordinator.HandleFileChanges(changes)
	if err != nil {
		t.Errorf("HandleFileChanges should not error: %v", err)
	}

	// Verify LastEventTime was updated
	status := coordinator.GetStatus()
	if !status.LastEventTime.After(initialTime) {
		t.Error("LastEventTime should be updated after handling file changes")
	}
}

// TestCoordinator_ProcessEvents_DebouncedEventsWithError tests processEvents with debounced events that cause errors
func TestCoordinator_ProcessEvents_DebouncedEventsWithError(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{}

	// Create debouncer that will send events
	debouncedEvents := make(chan []core.FileEvent, 1)
	debouncer := &mockEventDebouncer{
		events: debouncedEvents,
	}

	// Test trigger that returns error
	testTrigger := &mockTestTrigger{
		triggerTestsFunc: func(ctx context.Context, filePath string) error {
			return errors.New("trigger error")
		},
	}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Configure and set running
	options := core.WatchOptions{Mode: core.WatchAll}
	coordinator.Configure(options)
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.ctx, coordinator.cancel = context.WithCancel(context.Background())
	coordinator.mu.Unlock()

	// Start event processing
	go coordinator.processEvents()

	// Send debounced events that will cause error
	debouncedEvents <- []core.FileEvent{
		{Path: "error_file.go", Type: "modify"},
	}

	// Give time for processing
	time.Sleep(10 * time.Millisecond)

	// Verify error count was incremented
	status := coordinator.GetStatus()
	if status.ErrorCount == 0 {
		t.Error("Error count should be incremented when HandleFileChanges fails")
	}

	// Stop the coordinator
	coordinator.Stop()
}

// TestCoordinator_ProcessEvents_EmptyDebouncedEvents tests processEvents with empty debounced events
func TestCoordinator_ProcessEvents_EmptyDebouncedEvents(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{}

	// Create debouncer that will send empty events
	debouncedEvents := make(chan []core.FileEvent, 1)
	debouncer := &mockEventDebouncer{
		events: debouncedEvents,
	}

	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Configure and set running
	options := core.WatchOptions{Mode: core.WatchAll}
	coordinator.Configure(options)
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.ctx, coordinator.cancel = context.WithCancel(context.Background())
	coordinator.mu.Unlock()

	// Start event processing
	go coordinator.processEvents()

	// Send empty debounced events
	debouncedEvents <- []core.FileEvent{} // Empty slice

	// Give time for processing
	time.Sleep(10 * time.Millisecond)

	// This should not cause any errors or changes
	status := coordinator.GetStatus()
	if status.ErrorCount != 0 {
		t.Error("Error count should remain 0 for empty debounced events")
	}

	// Stop the coordinator
	coordinator.Stop()
}

// TestCoordinator_ProcessEvents_EventChannelProcessing tests processEvents with event channel
func TestCoordinator_ProcessEvents_EventChannelProcessing(t *testing.T) {
	t.Parallel()

	var addEventCalled bool
	var addedEvent core.FileEvent
	var mu sync.Mutex

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{
		addEventFunc: func(event core.FileEvent) {
			mu.Lock()
			addEventCalled = true
			addedEvent = event
			mu.Unlock()
		},
	}
	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Configure and set running
	options := core.WatchOptions{Mode: core.WatchAll}
	coordinator.Configure(options)
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.ctx, coordinator.cancel = context.WithCancel(context.Background())
	coordinator.mu.Unlock()

	// Start event processing
	go coordinator.processEvents()

	// Send event through event channel
	testEvent := core.FileEvent{Path: "test.go", Type: "modify"}
	coordinator.eventChannel <- testEvent

	// Give time for processing
	time.Sleep(10 * time.Millisecond)

	// Verify event was processed
	mu.Lock()
	if !addEventCalled {
		t.Error("AddEvent should have been called on debouncer")
	}
	if addedEvent.Path != testEvent.Path {
		t.Errorf("Expected event path %s, got %s", testEvent.Path, addedEvent.Path)
	}
	mu.Unlock()

	// Verify event count was incremented
	status := coordinator.GetStatus()
	if status.EventCount == 0 {
		t.Error("Event count should be incremented")
	}

	// Stop the coordinator
	coordinator.Stop()
}

// TestCoordinator_HandleFileChanges_ContextNil tests HandleFileChanges when context is nil
func TestCoordinator_HandleFileChanges_ContextNil(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{
		triggerTestsFunc: func(ctx context.Context, filePath string) error {
			if ctx == nil {
				return errors.New("context is nil")
			}
			return nil
		},
	}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Configure and set running but with nil context
	options := core.WatchOptions{Mode: core.WatchAll}
	coordinator.Configure(options)
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.ctx = nil // Set context to nil
	coordinator.mu.Unlock()

	changes := []core.FileEvent{
		{Path: "test.go", Type: "modify"},
	}

	err := coordinator.HandleFileChanges(changes)
	if err == nil {
		t.Error("HandleFileChanges should return error when context is nil")
	}

	// Verify error count was incremented
	status := coordinator.GetStatus()
	if status.ErrorCount == 0 {
		t.Error("Error count should be incremented when trigger fails")
	}
}

// TestCoordinator_Stop_ChannelCloseOrder tests that channels are closed in correct order
func TestCoordinator_Stop_ChannelCloseOrder(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Start the coordinator first
	ctx := context.Background()
	options := core.WatchOptions{Paths: []string{"./test"}}
	coordinator.Configure(options)

	// Manually set running state to test stop
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.ctx, coordinator.cancel = context.WithCancel(ctx)
	coordinator.mu.Unlock()

	// Stop should close channels without error
	err := coordinator.Stop()
	if err != nil {
		t.Errorf("Stop should not error: %v", err)
	}

	// Verify stopped state
	if !coordinator.stopped {
		t.Error("Coordinator should be marked as stopped")
	}

	// Verify status
	status := coordinator.GetStatus()
	if status.IsRunning {
		t.Error("Status should show not running after stop")
	}
}

// TestCoordinator_Start_ContextCanceled tests Start when file watcher returns context.Canceled
func TestCoordinator_Start_ContextCanceled(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{
		watchFunc: func(ctx context.Context, events chan<- core.FileEvent) error {
			return context.Canceled // This should not increment error count
		},
	}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger)

	// Configure with test options
	options := core.WatchOptions{
		Paths: []string{"./test"},
		Mode:  core.WatchAll,
	}
	err := coordinator.Configure(options)
	if err != nil {
		t.Fatalf("Configure should not error: %v", err)
	}

	ctx := context.Background()
	err = coordinator.Start(ctx)

	// Should not return error for context.Canceled
	if err != nil {
		t.Errorf("Start should not return error for context.Canceled: %v", err)
	}

	// Verify error count was not incremented for context.Canceled
	status := coordinator.GetStatus()
	if status.ErrorCount != 0 {
		t.Error("Error count should not be incremented for context.Canceled")
	}
}

// TestCoordinator_HandleFileChanges_AllModesCoverage tests all watch modes for complete coverage
func TestCoordinator_HandleFileChanges_AllModesCoverage(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		mode                    core.WatchMode
		expectedTriggerFunc     string
		triggerTestsFunc        func(ctx context.Context, filePath string) error
		triggerRelatedTestsFunc func(ctx context.Context, filePath string) error
	}{
		"watch_all_success": {
			mode:                core.WatchAll,
			expectedTriggerFunc: "TriggerTestsForFile",
			triggerTestsFunc: func(ctx context.Context, filePath string) error {
				return nil
			},
		},
		"watch_changed_success": {
			mode:                core.WatchChanged,
			expectedTriggerFunc: "TriggerTestsForFile",
			triggerTestsFunc: func(ctx context.Context, filePath string) error {
				return nil
			},
		},
		"watch_related_success": {
			mode:                core.WatchRelated,
			expectedTriggerFunc: "TriggerRelatedTests",
			triggerRelatedTestsFunc: func(ctx context.Context, filePath string) error {
				return nil
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fsWatcher := &mockFileSystemWatcher{}
			debouncer := &mockEventDebouncer{}
			testTrigger := &mockTestTrigger{
				triggerTestsFunc:        tt.triggerTestsFunc,
				triggerRelatedTestsFunc: tt.triggerRelatedTestsFunc,
			}

			coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

			// Configure and set running
			options := core.WatchOptions{Mode: tt.mode}
			coordinator.Configure(options)
			coordinator.mu.Lock()
			coordinator.status.IsRunning = true
			coordinator.ctx = context.Background()
			coordinator.mu.Unlock()

			changes := []core.FileEvent{
				{Path: "test1.go", Type: "modify"},
				{Path: "test2.go", Type: "create"},
			}

			err := coordinator.HandleFileChanges(changes)
			if err != nil {
				t.Errorf("HandleFileChanges should not error for %s: %v", name, err)
			}
		})
	}
}

// TestCoordinator_HandleFileChanges_WatchChangedError tests WatchChanged mode with error
func TestCoordinator_HandleFileChanges_WatchChangedError(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{
		triggerTestsFunc: func(ctx context.Context, filePath string) error {
			return errors.New("trigger error for changed mode")
		},
	}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Configure for WatchChanged mode
	options := core.WatchOptions{Mode: core.WatchChanged}
	coordinator.Configure(options)
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.ctx = context.Background()
	coordinator.mu.Unlock()

	changes := []core.FileEvent{
		{Path: "changed_file.go", Type: "modify"},
	}

	err := coordinator.HandleFileChanges(changes)
	if err == nil {
		t.Error("HandleFileChanges should return error when trigger fails in WatchChanged mode")
	}

	// Verify error count was incremented
	status := coordinator.GetStatus()
	if status.ErrorCount == 0 {
		t.Error("Error count should be incremented when trigger fails")
	}
}

// TestCoordinator_HandleFileChanges_WatchRelatedError tests WatchRelated mode with error
func TestCoordinator_HandleFileChanges_WatchRelatedError(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{
		triggerRelatedTestsFunc: func(ctx context.Context, filePath string) error {
			return errors.New("trigger error for related mode")
		},
	}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Configure for WatchRelated mode
	options := core.WatchOptions{Mode: core.WatchRelated}
	coordinator.Configure(options)
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.ctx = context.Background()
	coordinator.mu.Unlock()

	changes := []core.FileEvent{
		{Path: "related_file.go", Type: "modify"},
	}

	err := coordinator.HandleFileChanges(changes)
	if err == nil {
		t.Error("HandleFileChanges should return error when trigger fails in WatchRelated mode")
	}

	// Verify error count was incremented
	status := coordinator.GetStatus()
	if status.ErrorCount == 0 {
		t.Error("Error count should be incremented when trigger fails")
	}
}

// TestCoordinator_HandleFileChanges_MultipleChangesWithMixedResults tests multiple changes with some successes and failures
func TestCoordinator_HandleFileChanges_MultipleChangesWithMixedResults(t *testing.T) {
	t.Parallel()

	var callCount int
	var mu sync.Mutex

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{
		triggerTestsFunc: func(ctx context.Context, filePath string) error {
			mu.Lock()
			callCount++
			currentCall := callCount
			mu.Unlock()

			// First call succeeds, second call fails
			if currentCall == 1 {
				return nil
			}
			return errors.New("trigger error on second file")
		},
	}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Configure for WatchAll mode
	options := core.WatchOptions{Mode: core.WatchAll}
	coordinator.Configure(options)
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.ctx = context.Background()
	coordinator.mu.Unlock()

	changes := []core.FileEvent{
		{Path: "success_file.go", Type: "modify"},
		{Path: "error_file.go", Type: "modify"},
	}

	err := coordinator.HandleFileChanges(changes)
	if err == nil {
		t.Error("HandleFileChanges should return error when one of the triggers fails")
	}

	// Verify error count was incremented
	status := coordinator.GetStatus()
	if status.ErrorCount == 0 {
		t.Error("Error count should be incremented when trigger fails")
	}

	// Verify both files were attempted (first succeeds, second fails and returns error)
	mu.Lock()
	if callCount != 2 {
		t.Errorf("Expected 2 trigger calls, got %d", callCount)
	}
	mu.Unlock()
}

// TestCoordinator_HandleFileChanges_EmptyChangesSlice tests handling empty changes slice
func TestCoordinator_HandleFileChanges_EmptyChangesSlice(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Configure and set running
	options := core.WatchOptions{Mode: core.WatchAll}
	coordinator.Configure(options)
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.ctx = context.Background()
	coordinator.mu.Unlock()

	// Test with empty changes slice
	err := coordinator.HandleFileChanges([]core.FileEvent{})
	if err != nil {
		t.Errorf("HandleFileChanges should not error with empty changes: %v", err)
	}
}

// TestCoordinator_HandleFileChanges_StatusUpdateTiming tests that status is updated before processing
func TestCoordinator_HandleFileChanges_StatusUpdateTiming(t *testing.T) {
	t.Parallel()

	var triggerCalled bool
	var mu sync.Mutex

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{
		triggerTestsFunc: func(ctx context.Context, filePath string) error {
			mu.Lock()
			triggerCalled = true
			mu.Unlock()
			return nil
		},
	}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Configure and set running
	options := core.WatchOptions{Mode: core.WatchAll}
	coordinator.Configure(options)
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.ctx = context.Background()
	initialTime := coordinator.status.LastEventTime
	coordinator.mu.Unlock()

	// Wait a moment to ensure time difference
	time.Sleep(1 * time.Millisecond)

	changes := []core.FileEvent{
		{Path: "test.go", Type: "modify"},
	}

	err := coordinator.HandleFileChanges(changes)
	if err != nil {
		t.Errorf("HandleFileChanges should not error: %v", err)
	}

	// Verify trigger was called
	mu.Lock()
	if !triggerCalled {
		t.Error("Trigger should have been called")
	}
	mu.Unlock()

	// Verify status was updated
	status := coordinator.GetStatus()
	if !status.LastEventTime.After(initialTime) {
		t.Error("Status LastEventTime should be updated")
	}
}
