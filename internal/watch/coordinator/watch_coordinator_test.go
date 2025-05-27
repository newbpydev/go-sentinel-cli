package coordinator

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/watch/core"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// Mock implementations for testing

// mockTestRunner is a safe mock that doesn't execute real tests
type mockTestRunner struct {
	runFunc       func(ctx context.Context, targets []string) (string, error)
	runStreamFunc func(ctx context.Context, targets []string) (io.ReadCloser, error)
}

func (m *mockTestRunner) Run(ctx context.Context, targets []string) (string, error) {
	if m.runFunc != nil {
		return m.runFunc(ctx, targets)
	}
	// Return mock JSON output instead of executing real tests
	return `{"Time":"2024-01-01T00:00:00Z","Action":"pass","Package":"mock","Test":"MockTest","Elapsed":0.001}`, nil
}

func (m *mockTestRunner) RunStream(ctx context.Context, targets []string) (io.ReadCloser, error) {
	if m.runStreamFunc != nil {
		return m.runStreamFunc(ctx, targets)
	}
	// Return mock stream instead of executing real tests
	mockOutput := `{"Time":"2024-01-01T00:00:00Z","Action":"pass","Package":"mock","Test":"MockTest","Elapsed":0.001}`
	return io.NopCloser(strings.NewReader(mockOutput)), nil
}

type mockFileSystemWatcher struct {
	watchFunc      func(ctx context.Context, events chan<- core.FileEvent) error
	closeFunc      func() error
	addPathFunc    func(path string) error
	removePathFunc func(path string) error
}

func (m *mockFileSystemWatcher) Watch(ctx context.Context, events chan<- core.FileEvent) error {
	if m.watchFunc != nil {
		return m.watchFunc(ctx, events)
	}
	return nil
}

func (m *mockFileSystemWatcher) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *mockFileSystemWatcher) AddPath(path string) error {
	if m.addPathFunc != nil {
		return m.addPathFunc(path)
	}
	return nil
}

func (m *mockFileSystemWatcher) RemovePath(path string) error {
	if m.removePathFunc != nil {
		return m.removePathFunc(path)
	}
	return nil
}

type mockEventDebouncer struct {
	events       chan []core.FileEvent
	stopFunc     func() error
	setInterval  func(interval time.Duration)
	addEventFunc func(event core.FileEvent)
}

func (m *mockEventDebouncer) Events() <-chan []core.FileEvent {
	if m.events == nil {
		m.events = make(chan []core.FileEvent, 10)
	}
	return m.events
}

func (m *mockEventDebouncer) Stop() error {
	if m.stopFunc != nil {
		return m.stopFunc()
	}
	return nil
}

func (m *mockEventDebouncer) SetInterval(interval time.Duration) {
	if m.setInterval != nil {
		m.setInterval(interval)
	}
}

func (m *mockEventDebouncer) AddEvent(event core.FileEvent) {
	if m.addEventFunc != nil {
		m.addEventFunc(event)
	}
}

type mockTestTrigger struct {
	triggerTestsFunc        func(ctx context.Context, filePath string) error
	triggerRelatedTestsFunc func(ctx context.Context, filePath string) error
	triggerAllTestsFunc     func(ctx context.Context) error
	getTestTargetsFunc      func(changes []core.FileEvent) ([]string, error)
}

func (m *mockTestTrigger) TriggerTestsForFile(ctx context.Context, filePath string) error {
	if m.triggerTestsFunc != nil {
		return m.triggerTestsFunc(ctx, filePath)
	}
	return nil
}

func (m *mockTestTrigger) TriggerRelatedTests(ctx context.Context, filePath string) error {
	if m.triggerRelatedTestsFunc != nil {
		return m.triggerRelatedTestsFunc(ctx, filePath)
	}
	return nil
}

func (m *mockTestTrigger) TriggerAllTests(ctx context.Context) error {
	if m.triggerAllTestsFunc != nil {
		return m.triggerAllTestsFunc(ctx)
	}
	return nil
}

func (m *mockTestTrigger) GetTestTargets(changes []core.FileEvent) ([]string, error) {
	if m.getTestTargetsFunc != nil {
		return m.getTestTargetsFunc(changes)
	}
	return []string{}, nil
}

// createSafeTestWatchCoordinator creates a TestWatchCoordinator with mocked dependencies to prevent real test execution
func createSafeTestWatchCoordinator(options core.WatchOptions) (*TestWatchCoordinator, error) {
	coordinator, err := NewTestWatchCoordinator(options)
	if err != nil {
		return nil, err
	}

	// Replace the real test runner with a safe mock
	coordinator.testRunner = &mockTestRunner{}

	return coordinator, nil
}

// TestNewCoordinator_FactoryFunction tests the factory function following TDD patterns
func TestNewCoordinator_FactoryFunction(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger)
	if coordinator == nil {
		t.Fatal("NewCoordinator should not return nil")
	}

	// Verify interface compliance
	_, ok := coordinator.(core.WatchCoordinator)
	if !ok {
		t.Fatal("NewCoordinator should return core.WatchCoordinator interface")
	}

	// Verify initial status
	status := coordinator.GetStatus()
	if status.IsRunning {
		t.Error("Initial status should have IsRunning=false")
	}
	if status.EventCount != 0 {
		t.Error("Initial status should have EventCount=0")
	}
	if status.ErrorCount != 0 {
		t.Error("Initial status should have ErrorCount=0")
	}
	if status.Mode != core.WatchAll {
		t.Error("Initial status should have Mode=WatchAll")
	}
}

// TestCoordinator_Start_Success tests successful start operation
func TestCoordinator_Start_Success(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{
		watchFunc: func(ctx context.Context, events chan<- core.FileEvent) error {
			// Simulate successful watch start
			return nil
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

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = coordinator.Start(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("Start should not error: %v", err)
	}

	// Verify status after start
	status := coordinator.GetStatus()
	if !status.IsRunning {
		t.Error("Status should have IsRunning=true after start")
	}
	if len(status.WatchedPaths) != 1 || status.WatchedPaths[0] != "./test" {
		t.Errorf("Status should have correct WatchedPaths, got %v", status.WatchedPaths)
	}
}

// TestCoordinator_Start_AlreadyRunning tests start when already running
func TestCoordinator_Start_AlreadyRunning(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Manually set running state
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.mu.Unlock()

	ctx := context.Background()
	err := coordinator.Start(ctx)

	if err == nil {
		t.Error("Start should return error when already running")
	}

	var sentinelError *models.SentinelError
	if !errors.As(err, &sentinelError) {
		t.Errorf("Error should be SentinelError, got %T", err)
	}
}

// TestCoordinator_Stop_Success tests successful stop operation
func TestCoordinator_Stop_Success(t *testing.T) {
	t.Parallel()

	var stopCalled, closeCalled bool

	fsWatcher := &mockFileSystemWatcher{
		closeFunc: func() error {
			closeCalled = true
			return nil
		},
	}
	debouncer := &mockEventDebouncer{
		stopFunc: func() error {
			stopCalled = true
			return nil
		},
	}
	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger)

	err := coordinator.Stop()
	if err != nil {
		t.Errorf("Stop should not error: %v", err)
	}

	if !stopCalled {
		t.Error("Debouncer.Stop should have been called")
	}
	if !closeCalled {
		t.Error("FileSystemWatcher.Close should have been called")
	}

	// Verify status after stop
	status := coordinator.GetStatus()
	if status.IsRunning {
		t.Error("Status should have IsRunning=false after stop")
	}
}

// TestCoordinator_Stop_DebouncerError tests stop with debouncer error
func TestCoordinator_Stop_DebouncerError(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("debouncer stop error")

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{
		stopFunc: func() error {
			return expectedError
		},
	}
	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger)

	err := coordinator.Stop()
	if err == nil {
		t.Error("Stop should return error when debouncer fails")
	}

	var sentinelError *models.SentinelError
	if !errors.As(err, &sentinelError) {
		t.Errorf("Error should be SentinelError, got %T", err)
	}
}

// TestCoordinator_Stop_FileWatcherError tests stop with file watcher error
func TestCoordinator_Stop_FileWatcherError(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("file watcher close error")

	fsWatcher := &mockFileSystemWatcher{
		closeFunc: func() error {
			return expectedError
		},
	}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger)

	err := coordinator.Stop()
	if err == nil {
		t.Error("Stop should return error when file watcher fails")
	}

	var sentinelError *models.SentinelError
	if !errors.As(err, &sentinelError) {
		t.Errorf("Error should be SentinelError, got %T", err)
	}
}

// TestCoordinator_Stop_AlreadyStopped tests stop when already stopped
func TestCoordinator_Stop_AlreadyStopped(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Stop once
	err := coordinator.Stop()
	if err != nil {
		t.Errorf("First stop should not error: %v", err)
	}

	// Stop again
	err = coordinator.Stop()
	if err != nil {
		t.Errorf("Second stop should not error: %v", err)
	}
}

// TestCoordinator_HandleFileChanges_WatchAll tests file change handling in WatchAll mode
func TestCoordinator_HandleFileChanges_WatchAll(t *testing.T) {
	t.Parallel()

	var triggeredFiles []string
	var mu sync.Mutex

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{
		triggerTestsFunc: func(ctx context.Context, filePath string) error {
			mu.Lock()
			triggeredFiles = append(triggeredFiles, filePath)
			mu.Unlock()
			return nil
		},
	}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Configure for WatchAll mode
	options := core.WatchOptions{
		Mode: core.WatchAll,
	}
	coordinator.Configure(options)

	// Set running state
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.mu.Unlock()

	// Test file changes
	changes := []core.FileEvent{
		{Path: "file1.go", Type: "modify"},
		{Path: "file2.go", Type: "create"},
	}

	err := coordinator.HandleFileChanges(changes)
	if err != nil {
		t.Errorf("HandleFileChanges should not error: %v", err)
	}

	mu.Lock()
	if len(triggeredFiles) != 2 {
		t.Errorf("Expected 2 triggered files, got %d", len(triggeredFiles))
	}
	if triggeredFiles[0] != "file1.go" || triggeredFiles[1] != "file2.go" {
		t.Errorf("Expected files [file1.go, file2.go], got %v", triggeredFiles)
	}
	mu.Unlock()
}

// TestCoordinator_HandleFileChanges_WatchChanged tests file change handling in WatchChanged mode
func TestCoordinator_HandleFileChanges_WatchChanged(t *testing.T) {
	t.Parallel()

	var triggeredFiles []string
	var mu sync.Mutex

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{
		triggerTestsFunc: func(ctx context.Context, filePath string) error {
			mu.Lock()
			triggeredFiles = append(triggeredFiles, filePath)
			mu.Unlock()
			return nil
		},
	}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Configure for WatchChanged mode
	options := core.WatchOptions{
		Mode: core.WatchChanged,
	}
	coordinator.Configure(options)

	// Set running state
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.mu.Unlock()

	// Test file changes
	changes := []core.FileEvent{
		{Path: "changed_file.go", Type: "modify"},
	}

	err := coordinator.HandleFileChanges(changes)
	if err != nil {
		t.Errorf("HandleFileChanges should not error: %v", err)
	}

	mu.Lock()
	if len(triggeredFiles) != 1 {
		t.Errorf("Expected 1 triggered file, got %d", len(triggeredFiles))
	}
	if triggeredFiles[0] != "changed_file.go" {
		t.Errorf("Expected file changed_file.go, got %v", triggeredFiles)
	}
	mu.Unlock()
}

// TestCoordinator_HandleFileChanges_WatchRelated tests file change handling in WatchRelated mode
func TestCoordinator_HandleFileChanges_WatchRelated(t *testing.T) {
	t.Parallel()

	var triggeredFiles []string
	var mu sync.Mutex

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{
		triggerRelatedTestsFunc: func(ctx context.Context, filePath string) error {
			mu.Lock()
			triggeredFiles = append(triggeredFiles, filePath)
			mu.Unlock()
			return nil
		},
	}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Configure for WatchRelated mode
	options := core.WatchOptions{
		Mode: core.WatchRelated,
	}
	coordinator.Configure(options)

	// Set running state
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.mu.Unlock()

	// Test file changes
	changes := []core.FileEvent{
		{Path: "related_file.go", Type: "modify"},
	}

	err := coordinator.HandleFileChanges(changes)
	if err != nil {
		t.Errorf("HandleFileChanges should not error: %v", err)
	}

	mu.Lock()
	if len(triggeredFiles) != 1 {
		t.Errorf("Expected 1 triggered file, got %d", len(triggeredFiles))
	}
	if triggeredFiles[0] != "related_file.go" {
		t.Errorf("Expected file related_file.go, got %v", triggeredFiles)
	}
	mu.Unlock()
}

// TestCoordinator_HandleFileChanges_NotRunning tests file change handling when not running
func TestCoordinator_HandleFileChanges_NotRunning(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger)

	changes := []core.FileEvent{
		{Path: "file.go", Type: "modify"},
	}

	err := coordinator.HandleFileChanges(changes)
	if err == nil {
		t.Error("HandleFileChanges should return error when not running")
	}

	var sentinelError *models.SentinelError
	if !errors.As(err, &sentinelError) {
		t.Errorf("Error should be SentinelError, got %T", err)
	}
}

// TestCoordinator_HandleFileChanges_TriggerError tests file change handling with trigger error
func TestCoordinator_HandleFileChanges_TriggerError(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("trigger error")

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{
		triggerTestsFunc: func(ctx context.Context, filePath string) error {
			return expectedError
		},
	}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Configure and set running
	options := core.WatchOptions{Mode: core.WatchAll}
	coordinator.Configure(options)
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.mu.Unlock()

	changes := []core.FileEvent{
		{Path: "file.go", Type: "modify"},
	}

	err := coordinator.HandleFileChanges(changes)
	if err == nil {
		t.Error("HandleFileChanges should return error when trigger fails")
	}

	var sentinelError *models.SentinelError
	if !errors.As(err, &sentinelError) {
		t.Errorf("Error should be SentinelError, got %T", err)
	}

	// Verify error count was incremented
	status := coordinator.GetStatus()
	if status.ErrorCount == 0 {
		t.Error("Error count should be incremented")
	}
}

// TestCoordinator_HandleFileChanges_UnknownMode tests file change handling with unknown mode
func TestCoordinator_HandleFileChanges_UnknownMode(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Configure with invalid mode
	options := core.WatchOptions{Mode: core.WatchMode("invalid")}
	coordinator.Configure(options)
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.mu.Unlock()

	changes := []core.FileEvent{
		{Path: "file.go", Type: "modify"},
	}

	err := coordinator.HandleFileChanges(changes)
	if err == nil {
		t.Error("HandleFileChanges should return error for unknown mode")
	}

	var sentinelError *models.SentinelError
	if !errors.As(err, &sentinelError) {
		t.Errorf("Error should be SentinelError, got %T", err)
	}
}

// TestCoordinator_Configure tests configuration updates
func TestCoordinator_Configure(t *testing.T) {
	t.Parallel()

	var setIntervalCalled bool
	var intervalValue time.Duration

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{
		setInterval: func(interval time.Duration) {
			setIntervalCalled = true
			intervalValue = interval
		},
	}
	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger)

	options := core.WatchOptions{
		Mode:             core.WatchRelated,
		DebounceInterval: 300 * time.Millisecond,
		Paths:            []string{"./src", "./test"},
	}

	err := coordinator.Configure(options)
	if err != nil {
		t.Errorf("Configure should not error: %v", err)
	}

	if !setIntervalCalled {
		t.Error("SetInterval should have been called on debouncer")
	}
	if intervalValue != 300*time.Millisecond {
		t.Errorf("Expected interval 300ms, got %v", intervalValue)
	}

	// Verify status was updated
	status := coordinator.GetStatus()
	if status.Mode != core.WatchRelated {
		t.Errorf("Expected mode WatchRelated, got %v", status.Mode)
	}
}

// TestCoordinator_GetStatus tests status retrieval
func TestCoordinator_GetStatus(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Modify status directly for testing
	coordinator.mu.Lock()
	coordinator.status.IsRunning = true
	coordinator.status.EventCount = 42
	coordinator.status.ErrorCount = 3
	coordinator.status.Mode = core.WatchChanged
	coordinator.status.WatchedPaths = []string{"./test1", "./test2"}
	coordinator.mu.Unlock()

	status := coordinator.GetStatus()

	if !status.IsRunning {
		t.Error("Status should reflect IsRunning=true")
	}
	if status.EventCount != 42 {
		t.Errorf("Expected EventCount=42, got %d", status.EventCount)
	}
	if status.ErrorCount != 3 {
		t.Errorf("Expected ErrorCount=3, got %d", status.ErrorCount)
	}
	if status.Mode != core.WatchChanged {
		t.Errorf("Expected Mode=WatchChanged, got %v", status.Mode)
	}
	if len(status.WatchedPaths) != 2 {
		t.Errorf("Expected 2 watched paths, got %d", len(status.WatchedPaths))
	}
}

// TestCoordinator_ProcessEvents tests the event processing loop
func TestCoordinator_ProcessEvents(t *testing.T) {
	t.Parallel()

	eventReceived := make(chan bool, 1)

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{
		addEventFunc: func(event core.FileEvent) {
			eventReceived <- true
		},
	}
	// Use a safe mock that doesn't trigger real test execution
	testTrigger := &mockTestTrigger{
		triggerTestsFunc: func(ctx context.Context, filePath string) error {
			// Safe mock - no real test execution
			return nil
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

	// Send an event
	coordinator.eventChannel <- core.FileEvent{Path: "test.go", Type: "modify"}

	// Wait for event to be processed
	select {
	case <-eventReceived:
		// Event was processed
	case <-time.After(100 * time.Millisecond):
		t.Error("Event should have been processed")
	}

	// Test debounced events processing (but don't send actual events to avoid HandleFileChanges)
	// Just verify the channel exists and is accessible
	if debouncer.Events() == nil {
		t.Error("Debouncer events channel should be accessible")
	}

	// Stop the coordinator
	coordinator.Stop()
}

// TestCoordinator_IncrementEventCount tests thread-safe event count increment
func TestCoordinator_IncrementEventCount(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Test concurrent increments
	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			coordinator.incrementEventCount()
		}()
	}

	wg.Wait()

	status := coordinator.GetStatus()
	if status.EventCount != numGoroutines {
		t.Errorf("Expected EventCount=%d, got %d", numGoroutines, status.EventCount)
	}
}

// TestCoordinator_IncrementErrorCount tests thread-safe error count increment
func TestCoordinator_IncrementErrorCount(t *testing.T) {
	t.Parallel()

	fsWatcher := &mockFileSystemWatcher{}
	debouncer := &mockEventDebouncer{}
	testTrigger := &mockTestTrigger{}

	coordinator := NewCoordinator(fsWatcher, debouncer, testTrigger).(*Coordinator)

	// Test concurrent increments
	var wg sync.WaitGroup
	numGoroutines := 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			coordinator.incrementErrorCount()
		}()
	}

	wg.Wait()

	status := coordinator.GetStatus()
	if status.ErrorCount != numGoroutines {
		t.Errorf("Expected ErrorCount=%d, got %d", numGoroutines, status.ErrorCount)
	}
}

// Existing tests continue below...

func TestWatchCoordinatorDefaults(t *testing.T) {
	buffer := &bytes.Buffer{}

	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Check defaults were applied
	if coordinator.options.DebounceInterval != 500*time.Millisecond {
		t.Errorf("expected default debounce interval of 500ms, got %v", coordinator.options.DebounceInterval)
	}

	if len(coordinator.options.TestPatterns) != 1 || coordinator.options.TestPatterns[0] != "*_test.go" {
		t.Errorf("expected default test pattern of *_test.go, got %v", coordinator.options.TestPatterns)
	}

	if len(coordinator.options.IgnorePatterns) < 3 {
		t.Errorf("expected at least 3 default ignore patterns, got %v", coordinator.options.IgnorePatterns)
	}
}

func TestCoordinatorStatusPrinting(t *testing.T) {
	buffer := &bytes.Buffer{}

	options := core.WatchOptions{
		Paths:  []string{"."},
		Mode:   core.WatchAll,
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Test status printing
	coordinator.printStatus("Test message")
	output := buffer.String()

	if !strings.Contains(output, "Test message") {
		t.Errorf("expected output to contain 'Test message', got %s", output)
	}

	// Clear buffer
	buffer.Reset()

	// Test watch info printing
	coordinator.printWatchInfo()
	output = buffer.String()

	if !strings.Contains(output, "mode: all") {
		t.Errorf("expected output to contain watch mode, got %s", output)
	}

	if !strings.Contains(output, "Press Ctrl+C to exit") {
		t.Errorf("expected output to contain exit instructions, got %s", output)
	}
}

func TestCoordinatorClearTerminal(t *testing.T) {
	buffer := &bytes.Buffer{}

	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Test terminal clearing
	coordinator.clearTerminal()
	output := buffer.String()

	// Check for ANSI escape sequence
	if output != "\033[2J\033[H" {
		t.Errorf("expected ANSI escape sequence for clearing terminal, got %q", output)
	}
}

func TestCoordinatorFileChange(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "coordinator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("failed to remove temp dir: %v", err)
		}
	}()

	// Create a test file
	testFile := filepath.Join(tempDir, "example_test.go")
	// #nosec G306 - Test file, permissions not important
	if err := os.WriteFile(testFile, []byte("package example_test\n\nfunc TestExample(t *testing.T) {}\n"), 0600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	buffer := &bytes.Buffer{}

	options := core.WatchOptions{
		Paths:         []string{tempDir},
		Mode:          core.WatchChanged,
		Writer:        buffer,
		ClearTerminal: false,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Test runTestsForFile
	if err := coordinator.runTestsForFile(testFile); err != nil {
		t.Errorf("runTestsForFile failed: %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "Running tests for: example_test.go") {
		t.Errorf("expected output to contain running message, got %s", output)
	}
}

func TestCoordinatorHandleFileChanges(t *testing.T) {
	buffer := &bytes.Buffer{}

	options := core.WatchOptions{
		Paths:  []string{"."},
		Mode:   core.WatchChanged,
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Test handling empty file changes
	if err := coordinator.HandleFileChanges([]core.FileEvent{}); err != nil {
		t.Errorf("HandleFileChanges with empty slice failed: %v", err)
	}

	// Check that status was not updated for empty changes
	status := coordinator.GetStatus()
	if status.EventCount != 0 {
		t.Errorf("expected event count to be 0 for empty changes, got %d", status.EventCount)
	}

	// Test that the coordinator was created with correct initial status
	if status.IsRunning {
		t.Error("expected IsRunning to be false initially")
	}

	if len(status.WatchedPaths) != 1 || status.WatchedPaths[0] != "." {
		t.Errorf("expected WatchedPaths to be ['.'], got %v", status.WatchedPaths)
	}
}

func TestCoordinatorConfigure(t *testing.T) {
	buffer := &bytes.Buffer{}

	options := core.WatchOptions{
		Paths:            []string{"."},
		DebounceInterval: 100 * time.Millisecond,
		Writer:           buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Test configuration update
	newOptions := core.WatchOptions{
		Paths:            []string{"./src"},
		DebounceInterval: 200 * time.Millisecond,
		Writer:           buffer,
	}

	if err := coordinator.Configure(newOptions); err != nil {
		t.Errorf("Configure failed: %v", err)
	}

	// Check that options were updated
	if coordinator.options.DebounceInterval != 200*time.Millisecond {
		t.Errorf("expected debounce interval to be 200ms, got %v", coordinator.options.DebounceInterval)
	}
}

func TestCoordinatorGetStatus(t *testing.T) {
	buffer := &bytes.Buffer{}

	options := core.WatchOptions{
		Paths:  []string{"."},
		Mode:   core.WatchAll,
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	status := coordinator.GetStatus()

	// Check initial status
	if status.IsRunning {
		t.Error("expected IsRunning to be false initially")
	}

	if len(status.WatchedPaths) != 1 || status.WatchedPaths[0] != "." {
		t.Errorf("expected WatchedPaths to be ['.'], got %v", status.WatchedPaths)
	}

	if status.Mode != core.WatchAll {
		t.Errorf("expected Mode to be WatchAll, got %v", status.Mode)
	}
}

// TestTestWatchCoordinator_Start tests the Start method of TestWatchCoordinator
func TestTestWatchCoordinator_Start(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{"."},
		Mode:   core.WatchAll,
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Test start with context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = coordinator.Start(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("Start should not error or should timeout: %v", err)
	}

	// Verify status was updated
	status := coordinator.GetStatus()
	if status.IsRunning {
		t.Error("Status should have IsRunning=false after context cancellation")
	}
}

// TestTestWatchCoordinator_Stop tests the Stop method of TestWatchCoordinator
func TestTestWatchCoordinator_Stop(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	err = coordinator.Stop()
	if err != nil {
		t.Errorf("Stop should not error: %v", err)
	}
}

// TestTestWatchCoordinator_ProcessFileEvents tests the processFileEvents method
func TestTestWatchCoordinator_ProcessFileEvents(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Create a channel for file events
	fileEvents := make(chan core.FileEvent, 1)

	// Start processing events in a goroutine
	go coordinator.processFileEvents(fileEvents)

	// Send an event
	fileEvents <- core.FileEvent{Path: "test.go", Type: "modify"}

	// Close the channel to stop processing
	close(fileEvents)

	// Give it time to process
	time.Sleep(10 * time.Millisecond)
}

// TestTestWatchCoordinator_RunAllTests tests the runAllTests method
func TestTestWatchCoordinator_RunAllTests(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	coordinator.runAllTests()

	output := buffer.String()
	if !strings.Contains(output, "Running all tests") {
		t.Errorf("expected output to contain 'Running all tests', got %s", output)
	}
}

// TestTestWatchCoordinator_RunRelatedTests tests the runRelatedTests method
func TestTestWatchCoordinator_RunRelatedTests(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "coordinator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("failed to remove temp dir: %v", err)
		}
	}()

	// Create a test file
	testFile := filepath.Join(tempDir, "example.go")
	// #nosec G306 - Test file, permissions not important
	if err := os.WriteFile(testFile, []byte("package example\n\nfunc Example() {}\n"), 0600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{tempDir},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	err = coordinator.runRelatedTests(testFile)
	if err != nil {
		t.Errorf("runRelatedTests failed: %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "Running related tests for: example.go") {
		t.Errorf("expected output to contain running message, got %s", output)
	}
}

// TestTestWatchCoordinator_ExecuteTests tests the executeTests method
func TestTestWatchCoordinator_ExecuteTests(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Test with empty targets - this should not execute any real tests
	err = coordinator.executeTests([]string{})
	if err != nil {
		t.Errorf("executeTests with empty targets should not error: %v", err)
	}

	// Test with mock targets - safe because we're using mock test runner
	err = coordinator.executeTests([]string{"./mock"})
	if err != nil {
		t.Errorf("executeTests with mock targets should not error: %v", err)
	}
}

// TestTestWatchCoordinator_HandleFileChanges_Comprehensive tests comprehensive file change handling
func TestTestWatchCoordinator_HandleFileChanges_Comprehensive(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "coordinator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("failed to remove temp dir: %v", err)
		}
	}()

	// Create test files
	testFile := filepath.Join(tempDir, "example_test.go")
	// #nosec G306 - Test file, permissions not important
	if err := os.WriteFile(testFile, []byte("package example_test\n\nfunc TestExample(t *testing.T) {}\n"), 0600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	sourceFile := filepath.Join(tempDir, "example.go")
	// #nosec G306 - Test file, permissions not important
	if err := os.WriteFile(sourceFile, []byte("package example\n\nfunc Example() {}\n"), 0600); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	buffer := &bytes.Buffer{}

	// Test WatchAll mode
	options := core.WatchOptions{
		Paths:         []string{tempDir},
		Mode:          core.WatchAll,
		Writer:        buffer,
		ClearTerminal: true,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	changes := []core.FileEvent{
		{Path: testFile, Type: "modify"},
	}

	err = coordinator.HandleFileChanges(changes)
	if err != nil {
		t.Errorf("HandleFileChanges should not error: %v", err)
	}

	// Test WatchChanged mode
	buffer.Reset()
	options.Mode = core.WatchChanged
	options.ClearTerminal = false

	coordinator2, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	err = coordinator2.HandleFileChanges(changes)
	if err != nil {
		t.Errorf("HandleFileChanges should not error: %v", err)
	}

	// Test WatchRelated mode
	buffer.Reset()
	options.Mode = core.WatchRelated

	coordinator3, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	err = coordinator3.HandleFileChanges(changes)
	if err != nil {
		t.Errorf("HandleFileChanges should not error: %v", err)
	}
}

// TestTestWatchCoordinator_RunTestsForFile_Comprehensive tests comprehensive runTestsForFile scenarios
func TestTestWatchCoordinator_RunTestsForFile_Comprehensive(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "coordinator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("failed to remove temp dir: %v", err)
		}
	}()

	// Create a source file (non-test)
	sourceFile := filepath.Join(tempDir, "example.go")
	// #nosec G306 - Test file, permissions not important
	if err := os.WriteFile(sourceFile, []byte("package example\n\nfunc Example() {}\n"), 0600); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{tempDir},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Test with source file (non-test file)
	err = coordinator.runTestsForFile(sourceFile)
	if err != nil {
		t.Errorf("runTestsForFile with source file should not error: %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "Running tests for: example.go") {
		t.Errorf("expected output to contain running message, got %s", output)
	}
}

// TestCoordinator_Start_FileWatcherError tests Start method with file watcher error
func TestCoordinator_Start_FileWatcherError(t *testing.T) {
	t.Parallel()

	expectedError := errors.New("file watcher error")

	fsWatcher := &mockFileSystemWatcher{
		watchFunc: func(ctx context.Context, events chan<- core.FileEvent) error {
			return expectedError
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

	// The error should be returned when file watcher fails
	if err == nil {
		t.Error("Start should return error when file watcher fails")
	}

	var sentinelError *models.SentinelError
	if !errors.As(err, &sentinelError) {
		t.Errorf("Error should be SentinelError, got %T", err)
	}
}

// TestCoordinator_HandleFileChanges_EmptyChanges tests handling empty changes
func TestCoordinator_HandleFileChanges_EmptyChanges(t *testing.T) {
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
	coordinator.mu.Unlock()

	// Test with empty changes
	err := coordinator.HandleFileChanges([]core.FileEvent{})
	if err != nil {
		t.Errorf("HandleFileChanges with empty changes should not error: %v", err)
	}
}

// TestNewTestWatchCoordinator_ErrorCases tests error cases in NewTestWatchCoordinator
func TestNewTestWatchCoordinator_ErrorCases(t *testing.T) {
	t.Parallel()

	// Test with empty paths - this should still work as it defaults to current directory
	// Use a buffer instead of os.Stdout to avoid output during tests
	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{},
		Writer: buffer,
	}

	coordinator, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Errorf("NewTestWatchCoordinator should not error with empty paths: %v", err)
	}
	if coordinator == nil {
		t.Error("NewTestWatchCoordinator should return valid coordinator")
	}

	// Verify defaults were applied
	if coordinator.options.DebounceInterval != 500*time.Millisecond {
		t.Errorf("Expected default debounce interval 500ms, got %v", coordinator.options.DebounceInterval)
	}
}

// TestCoordinator_ProcessEvents_StopChannel tests the processEvents method with stop channel
func TestCoordinator_ProcessEvents_StopChannel(t *testing.T) {
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
	coordinator.ctx, coordinator.cancel = context.WithCancel(context.Background())
	coordinator.mu.Unlock()

	// Start event processing
	go coordinator.processEvents()

	// Send stop signal
	coordinator.stopCh <- struct{}{}

	// Give it time to process
	time.Sleep(10 * time.Millisecond)
}

// TestCoordinator_ProcessEvents_ContextDone tests the processEvents method with context cancellation
func TestCoordinator_ProcessEvents_ContextDone(t *testing.T) {
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
	coordinator.ctx, coordinator.cancel = context.WithCancel(context.Background())
	coordinator.mu.Unlock()

	// Start event processing
	go coordinator.processEvents()

	// Cancel context
	coordinator.cancel()

	// Give it time to process
	time.Sleep(10 * time.Millisecond)
}

// TestTestWatchCoordinator_ExecuteTests_Error tests executeTests error cases
func TestTestWatchCoordinator_ExecuteTests_Error(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Replace with error-returning mock
	coordinator.testRunner = &mockTestRunner{
		runFunc: func(ctx context.Context, targets []string) (string, error) {
			return "", errors.New("test execution error")
		},
	}

	err = coordinator.executeTests([]string{"./test"})
	if err == nil {
		t.Error("executeTests should return error when test runner fails")
	}

	if !strings.Contains(err.Error(), "test execution failed") {
		t.Errorf("Expected error to contain 'test execution failed', got: %v", err)
	}
}

// TestTestWatchCoordinator_RunAllTests_WithProcessor tests runAllTests with processor
func TestTestWatchCoordinator_RunAllTests_WithProcessor(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Replace with mock that returns JSON output
	coordinator.testRunner = &mockTestRunner{
		runFunc: func(ctx context.Context, targets []string) (string, error) {
			return `{"Time":"2024-01-01T00:00:00Z","Action":"pass","Package":"test","Test":"TestExample","Elapsed":0.001}`, nil
		},
	}

	coordinator.runAllTests()

	output := buffer.String()
	if !strings.Contains(output, "Running all tests") {
		t.Errorf("expected output to contain 'Running all tests', got %s", output)
	}
}

// TestTestWatchCoordinator_RunTestsForFile_TestFile tests runTestsForFile with test file
func TestTestWatchCoordinator_RunTestsForFile_TestFile(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Test with a test file
	err = coordinator.runTestsForFile("example_test.go")
	if err != nil {
		t.Errorf("runTestsForFile should not error: %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "Running tests for: example_test.go") {
		t.Errorf("expected output to contain running message, got %s", output)
	}
}

// TestTestWatchCoordinator_RunRelatedTests_TestFile tests runRelatedTests with test file
func TestTestWatchCoordinator_RunRelatedTests_TestFile(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Test with a test file
	err = coordinator.runRelatedTests("example_test.go")
	if err != nil {
		t.Errorf("runRelatedTests should not error: %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "Running related tests for: example_test.go") {
		t.Errorf("expected output to contain running message, got %s", output)
	}
}

// TestTestWatchCoordinator_ExecuteTests_WithProcessor tests executeTests with processor
func TestTestWatchCoordinator_ExecuteTests_WithProcessor(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Replace with mock that returns JSON output
	coordinator.testRunner = &mockTestRunner{
		runFunc: func(ctx context.Context, targets []string) (string, error) {
			return `{"Time":"2024-01-01T00:00:00Z","Action":"pass","Package":"test","Test":"TestExample","Elapsed":0.001}`, nil
		},
	}

	err = coordinator.executeTests([]string{"./test"})
	if err != nil {
		t.Errorf("executeTests should not error: %v", err)
	}
}

// TestTestWatchCoordinator_ExecuteTests_EmptyTargets tests executeTests with empty targets
func TestTestWatchCoordinator_ExecuteTests_EmptyTargets(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	err = coordinator.executeTests([]string{})
	if err != nil {
		t.Errorf("executeTests with empty targets should not error: %v", err)
	}
}

// TestTestWatchCoordinator_ExecuteTests_DuplicateTargets tests executeTests with duplicate targets
func TestTestWatchCoordinator_ExecuteTests_DuplicateTargets(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Test with duplicate targets
	err = coordinator.executeTests([]string{"./test", "./test", "./other"})
	if err != nil {
		t.Errorf("executeTests with duplicate targets should not error: %v", err)
	}
}

// TestTestWatchCoordinator_HandleFileChanges_EmptyChanges tests HandleFileChanges with empty changes
func TestTestWatchCoordinator_HandleFileChanges_EmptyChanges(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	err = coordinator.HandleFileChanges([]core.FileEvent{})
	if err != nil {
		t.Errorf("HandleFileChanges with empty changes should not error: %v", err)
	}
}

// TestTestWatchCoordinator_HandleFileChanges_WithClearTerminal tests HandleFileChanges with clear terminal
func TestTestWatchCoordinator_HandleFileChanges_WithClearTerminal(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:         []string{"."},
		Writer:        buffer,
		ClearTerminal: true,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	changes := []core.FileEvent{
		{Path: "test.go", Type: "modify"},
	}

	err = coordinator.HandleFileChanges(changes)
	if err != nil {
		t.Errorf("HandleFileChanges should not error: %v", err)
	}

	output := buffer.String()
	// Should contain ANSI escape codes for clearing terminal
	if !strings.Contains(output, "\033[2J\033[H") {
		t.Errorf("expected output to contain terminal clear codes, got %s", output)
	}
}

// TestTestWatchCoordinator_Stop_NotStarted tests Stop when not started
func TestTestWatchCoordinator_Stop_NotStarted(t *testing.T) {
	t.Parallel()

	buffer := &bytes.Buffer{}
	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := createSafeTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Stop without starting
	err = coordinator.Stop()
	if err != nil {
		t.Errorf("Stop should not error when not started: %v", err)
	}
}
