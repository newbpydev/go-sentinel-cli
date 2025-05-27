// Package watcher tests for injectable file system watcher with 100% coverage
package watcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/newbpydev/go-sentinel/internal/watch/core"
)

// Mock implementations for dependency injection testing

// mockFsnotifyWatcher provides controllable mock of fsnotify.Watcher
type mockFsnotifyWatcher struct {
	mu           sync.Mutex
	addFunc      func(name string) error
	removeFunc   func(name string) error
	closeFunc    func() error
	events       chan fsnotify.Event
	errors       chan error
	closed       bool
	addedPaths   []string
	removedPaths []string
}

func newMockFsnotifyWatcher() *mockFsnotifyWatcher {
	return &mockFsnotifyWatcher{
		events: make(chan fsnotify.Event, 10),
		errors: make(chan error, 10),
	}
}

func (m *mockFsnotifyWatcher) Add(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.addFunc != nil {
		return m.addFunc(name)
	}

	m.addedPaths = append(m.addedPaths, name)
	return nil
}

func (m *mockFsnotifyWatcher) Remove(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.removeFunc != nil {
		return m.removeFunc(name)
	}

	m.removedPaths = append(m.removedPaths, name)
	return nil
}

func (m *mockFsnotifyWatcher) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closeFunc != nil {
		return m.closeFunc()
	}

	m.closed = true
	close(m.events)
	close(m.errors)
	return nil
}

func (m *mockFsnotifyWatcher) Events() <-chan fsnotify.Event {
	return m.events
}

func (m *mockFsnotifyWatcher) Errors() <-chan error {
	return m.errors
}

func (m *mockFsnotifyWatcher) sendEvent(event fsnotify.Event) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.closed {
		select {
		case m.events <- event:
		default:
		}
	}
}

func (m *mockFsnotifyWatcher) sendError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.closed {
		select {
		case m.errors <- err:
		default:
		}
	}
}

// mockFileSystem provides controllable mock of filesystem operations
type mockFileSystem struct {
	statFunc func(name string) (os.FileInfo, error)
	walkFunc func(root string, walkFn filepath.WalkFunc) error
	absFunc  func(path string) (string, error)
	files    map[string]*mockFileInfo
}

type mockFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return m.mode }
func (m *mockFileInfo) ModTime() time.Time { return m.modTime }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

func newMockFileSystem() *mockFileSystem {
	return &mockFileSystem{
		files: make(map[string]*mockFileInfo),
	}
}

func (m *mockFileSystem) addFile(path string, isDir bool) {
	m.files[path] = &mockFileInfo{
		name:    filepath.Base(path),
		isDir:   isDir,
		modTime: time.Now(),
	}
}

func (m *mockFileSystem) Stat(name string) (os.FileInfo, error) {
	if m.statFunc != nil {
		return m.statFunc(name)
	}

	if info, exists := m.files[name]; exists {
		return info, nil
	}

	return nil, os.ErrNotExist
}

func (m *mockFileSystem) Walk(root string, walkFn filepath.WalkFunc) error {
	if m.walkFunc != nil {
		return m.walkFunc(root, walkFn)
	}

	// Default behavior - walk through mock files
	for path, info := range m.files {
		if err := walkFn(path, info, nil); err != nil {
			if err == filepath.SkipDir {
				continue
			}
			return err
		}
	}
	return nil
}

func (m *mockFileSystem) Abs(path string) (string, error) {
	if m.absFunc != nil {
		return m.absFunc(path)
	}

	if path == "" {
		return "", errors.New("empty path")
	}

	// Use filepath.Join to get the correct path separator for the platform
	return filepath.Join("\\abs", path), nil
}

// mockTimeProvider provides controllable mock of time operations
type mockTimeProvider struct {
	nowFunc   func() time.Time
	fixedTime time.Time
}

func newMockTimeProvider() *mockTimeProvider {
	return &mockTimeProvider{
		fixedTime: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}
}

func (m *mockTimeProvider) Now() time.Time {
	if m.nowFunc != nil {
		return m.nowFunc()
	}
	return m.fixedTime
}

// mockWatcherFactory provides controllable mock of watcher creation
type mockWatcherFactory struct {
	newWatcherFunc func() (FsnotifyWatcher, error)
	watcher        *mockFsnotifyWatcher
}

func newMockWatcherFactory() *mockWatcherFactory {
	return &mockWatcherFactory{
		watcher: newMockFsnotifyWatcher(),
	}
}

func (m *mockWatcherFactory) NewWatcher() (FsnotifyWatcher, error) {
	if m.newWatcherFunc != nil {
		return m.newWatcherFunc()
	}
	return m.watcher, nil
}

// Test constructor with nil dependencies (JIT injection)
func TestNewInjectableFileSystemWatcher_NilDependencies(t *testing.T) {
	t.Parallel()

	watcher, err := NewInjectableFileSystemWatcher([]string{"/test"}, []string{"*.log"}, nil)

	if err != nil {
		t.Fatalf("Expected no error with nil dependencies, got: %v", err)
	}

	if watcher == nil {
		t.Fatal("Expected watcher to be created with default dependencies")
	}

	if watcher.fs == nil {
		t.Error("Expected filesystem to be injected with default")
	}

	if watcher.timeProvider == nil {
		t.Error("Expected time provider to be injected with default")
	}

	if watcher.factory == nil {
		t.Error("Expected factory to be injected with default")
	}
}

// Test constructor with partial dependencies (JIT injection)
func TestNewInjectableFileSystemWatcher_PartialDependencies(t *testing.T) {
	t.Parallel()

	mockFS := newMockFileSystem()
	deps := &Dependencies{
		FileSystem: mockFS,
		// TimeProvider and Factory left nil to test JIT injection
	}

	watcher, err := NewInjectableFileSystemWatcher([]string{"/test"}, []string{"*.log"}, deps)

	if err != nil {
		t.Fatalf("Expected no error with partial dependencies, got: %v", err)
	}

	if watcher.fs != mockFS {
		t.Error("Expected custom filesystem to be preserved")
	}

	if watcher.timeProvider == nil {
		t.Error("Expected time provider to be injected with default")
	}

	if watcher.factory == nil {
		t.Error("Expected factory to be injected with default")
	}
}

// Test factory error handling (achieving 100% coverage)
func TestNewInjectableFileSystemWatcher_FactoryError(t *testing.T) {
	t.Parallel()

	factoryError := errors.New("factory creation failed")
	factory := newMockWatcherFactory()
	factory.newWatcherFunc = func() (FsnotifyWatcher, error) {
		return nil, factoryError
	}

	deps := &Dependencies{
		FileSystem:   newMockFileSystem(),
		TimeProvider: newMockTimeProvider(),
		Factory:      factory,
	}

	watcher, err := NewInjectableFileSystemWatcher([]string{"/test"}, []string{"*.log"}, deps)

	if err == nil {
		t.Fatal("Expected error when factory fails")
	}

	if watcher != nil {
		t.Error("Expected no watcher when factory fails")
	}

	expectedMsg := "failed to create file watcher"
	if !containsError(err, expectedMsg) {
		t.Errorf("Expected error containing %q, got: %v", expectedMsg, err)
	}
}

// Test Watch with context cancellation - avoid AddPath during initialization
func TestInjectableFileSystemWatcher_Watch_ContextCancellation(t *testing.T) {
	t.Parallel()

	mockWatcher := newMockFsnotifyWatcher()
	factory := newMockWatcherFactory()
	factory.watcher = mockWatcher

	mockFS := newMockFileSystem()

	deps := &Dependencies{
		FileSystem:   mockFS,
		TimeProvider: newMockTimeProvider(),
		Factory:      factory,
	}

	// Create watcher with empty paths to avoid AddPath during initialization
	watcher, err := NewInjectableFileSystemWatcher([]string{}, []string{}, deps)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan core.FileEvent, 10)

	// Cancel immediately to test context cancellation path
	cancel()

	err = watcher.Watch(ctx, events)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
}

// Test Watch with watcher events channel closed - avoid AddPath during initialization
func TestInjectableFileSystemWatcher_Watch_EventsChannelClosed(t *testing.T) {
	t.Parallel()

	mockWatcher := newMockFsnotifyWatcher()
	factory := newMockWatcherFactory()
	factory.watcher = mockWatcher

	mockFS := newMockFileSystem()

	deps := &Dependencies{
		FileSystem:   mockFS,
		TimeProvider: newMockTimeProvider(),
		Factory:      factory,
	}

	// Create watcher with empty paths to avoid AddPath during initialization
	watcher, err := NewInjectableFileSystemWatcher([]string{}, []string{}, deps)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	ctx := context.Background()
	events := make(chan core.FileEvent, 10)

	// Close events channel to simulate channel closure
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(mockWatcher.events)
	}()

	err = watcher.Watch(ctx, events)
	expectedMsg := "watcher channel closed"
	if err == nil || err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got: %v", expectedMsg, err)
	}
}

// Test Watch with watcher errors channel closed - avoid AddPath during initialization
func TestInjectableFileSystemWatcher_Watch_ErrorsChannelClosed(t *testing.T) {
	t.Parallel()

	mockWatcher := newMockFsnotifyWatcher()
	factory := newMockWatcherFactory()
	factory.watcher = mockWatcher

	mockFS := newMockFileSystem()

	deps := &Dependencies{
		FileSystem:   mockFS,
		TimeProvider: newMockTimeProvider(),
		Factory:      factory,
	}

	// Create watcher with empty paths to avoid AddPath during initialization
	watcher, err := NewInjectableFileSystemWatcher([]string{}, []string{}, deps)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	ctx := context.Background()
	events := make(chan core.FileEvent, 10)

	// Close errors channel to simulate channel closure
	go func() {
		time.Sleep(10 * time.Millisecond)
		close(mockWatcher.errors)
	}()

	err = watcher.Watch(ctx, events)
	expectedMsg := "watcher error channel closed"
	if err == nil || err.Error() != expectedMsg {
		t.Errorf("Expected error %q, got: %v", expectedMsg, err)
	}
}

// Test Watch with fsnotify error - avoid AddPath during initialization
func TestInjectableFileSystemWatcher_Watch_FsnotifyError(t *testing.T) {
	t.Parallel()

	mockWatcher := newMockFsnotifyWatcher()
	factory := newMockWatcherFactory()
	factory.watcher = mockWatcher

	mockFS := newMockFileSystem()

	deps := &Dependencies{
		FileSystem:   mockFS,
		TimeProvider: newMockTimeProvider(),
		Factory:      factory,
	}

	// Create watcher with empty paths to avoid AddPath during initialization
	watcher, err := NewInjectableFileSystemWatcher([]string{}, []string{}, deps)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	ctx := context.Background()
	events := make(chan core.FileEvent, 10)

	// Send an error through the errors channel
	fsnotifyError := errors.New("fsnotify internal error")
	go func() {
		time.Sleep(10 * time.Millisecond)
		mockWatcher.sendError(fsnotifyError)
	}()

	err = watcher.Watch(ctx, events)
	expectedMsg := "watcher error"
	if err == nil || !containsError(err, expectedMsg) {
		t.Errorf("Expected error containing %q, got: %v", expectedMsg, err)
	}
}

// Test Watch with directory creation requiring watcher addition - avoid AddPath during initialization
func TestInjectableFileSystemWatcher_Watch_DirectoryCreation(t *testing.T) {
	t.Parallel()

	mockWatcher := newMockFsnotifyWatcher()
	factory := newMockWatcherFactory()
	factory.watcher = mockWatcher

	mockFS := newMockFileSystem()
	mockFS.addFile("\\abs\\test\\newdir", true) // Use Windows-style path

	deps := &Dependencies{
		FileSystem:   mockFS,
		TimeProvider: newMockTimeProvider(),
		Factory:      factory,
	}

	// Create watcher with empty paths to avoid AddPath during initialization
	watcher, err := NewInjectableFileSystemWatcher([]string{}, []string{}, deps)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	events := make(chan core.FileEvent, 10)

	// Send a directory creation event
	go func() {
		time.Sleep(10 * time.Millisecond)
		mockWatcher.sendEvent(fsnotify.Event{
			Name: "\\abs\\test\\newdir", // Use Windows-style path
			Op:   fsnotify.Create,
		})
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err = watcher.Watch(ctx, events)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}

	// Verify the directory was added to the watcher
	found := false
	for _, path := range mockWatcher.addedPaths {
		if path == "\\abs\\test\\newdir" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected new directory to be added to watcher")
	}
}

// Test Watch with directory creation error - avoid AddPath during initialization
func TestInjectableFileSystemWatcher_Watch_DirectoryCreationError(t *testing.T) {
	t.Parallel()

	mockWatcher := newMockFsnotifyWatcher()
	addError := errors.New("failed to add directory")
	mockWatcher.addFunc = func(name string) error {
		if name == "\\abs\\test\\newdir" {
			return addError
		}
		return nil
	}

	factory := newMockWatcherFactory()
	factory.watcher = mockWatcher

	mockFS := newMockFileSystem()
	mockFS.addFile("\\abs\\test\\newdir", true) // Use Windows-style path

	deps := &Dependencies{
		FileSystem:   mockFS,
		TimeProvider: newMockTimeProvider(),
		Factory:      factory,
	}

	// Create watcher with empty paths to avoid AddPath during initialization
	watcher, err := NewInjectableFileSystemWatcher([]string{}, []string{}, deps)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	ctx := context.Background()
	events := make(chan core.FileEvent, 10)

	// Send a directory creation event
	go func() {
		time.Sleep(10 * time.Millisecond)
		mockWatcher.sendEvent(fsnotify.Event{
			Name: "\\abs\\test\\newdir", // Use Windows-style path
			Op:   fsnotify.Create,
		})
	}()

	err = watcher.Watch(ctx, events)
	expectedMsg := "failed to add new directory"
	if err == nil || !containsError(err, expectedMsg) {
		t.Errorf("Expected error containing %q, got: %v", expectedMsg, err)
	}
}

// Test successful event processing with custom time - avoid AddPath during initialization
func TestInjectableFileSystemWatcher_Watch_SuccessfulEventProcessing(t *testing.T) {
	t.Parallel()

	mockWatcher := newMockFsnotifyWatcher()
	factory := newMockWatcherFactory()
	factory.watcher = mockWatcher

	mockFS := newMockFileSystem()
	mockFS.addFile("\\abs\\test\\file.go", false) // Use Windows-style path

	customTime := time.Date(2024, 12, 25, 10, 30, 0, 0, time.UTC)
	timeProvider := newMockTimeProvider()
	timeProvider.fixedTime = customTime

	deps := &Dependencies{
		FileSystem:   mockFS,
		TimeProvider: timeProvider,
		Factory:      factory,
	}

	// Create watcher with empty paths to avoid AddPath during initialization
	watcher, err := NewInjectableFileSystemWatcher([]string{}, []string{}, deps)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	events := make(chan core.FileEvent, 10)

	// Start the watcher in a goroutine
	go func() {
		_ = watcher.Watch(ctx, events) // Ignore error for this test
	}()

	// Send a file write event
	go func() {
		time.Sleep(10 * time.Millisecond)
		mockWatcher.sendEvent(fsnotify.Event{
			Name: "\\abs\\test\\file.go", // Use Windows-style path
			Op:   fsnotify.Write,
		})
	}()

	// Collect events with timeout
	var receivedEvents []core.FileEvent
	timeout := time.After(50 * time.Millisecond)
eventLoop:
	for {
		select {
		case event := <-events:
			receivedEvents = append(receivedEvents, event)
			if len(receivedEvents) >= 1 {
				break eventLoop
			}
		case <-timeout:
			break eventLoop
		}
	}

	if len(receivedEvents) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(receivedEvents))
	}

	event := receivedEvents[0]
	if event.Path != "\\abs\\test\\file.go" {
		t.Errorf("Expected path %q, got %q", "\\abs\\test\\file.go", event.Path)
	}

	if event.Type != "write" {
		t.Errorf("Expected type %q, got %q", "write", event.Type)
	}

	if !event.Timestamp.Equal(customTime) {
		t.Errorf("Expected timestamp %v, got %v", customTime, event.Timestamp)
	}

	if event.IsTest {
		t.Error("Expected IsTest to be false for .go file")
	}
}

// Test pattern matching edge cases
func TestInjectableFileSystemWatcher_MatchesAnyPattern_EdgeCases(t *testing.T) {
	t.Parallel()

	watcher := &InjectableFileSystemWatcher{}

	tests := []struct {
		name     string
		path     string
		patterns []string
		expected bool
	}{
		{
			name:     "Empty patterns",
			path:     "/test/file.go",
			patterns: []string{},
			expected: false,
		},
		{
			name:     "Simple contains match",
			path:     "/test/node_modules/file.js",
			patterns: []string{"node_modules"},
			expected: true,
		},
		{
			name:     "Prefix wildcard match",
			path:     "/test/file_test.go",
			patterns: []string{"*_test.go"},
			expected: true,
		},
		{
			name:     "Suffix wildcard match",
			path:     "/tmp/tempfile.txt",
			patterns: []string{"/tmp/*"},
			expected: true,
		},
		{
			name:     "No match",
			path:     "/test/file.go",
			patterns: []string{"*.py", "node_modules"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := watcher.matchesAnyPattern(tt.path, tt.patterns)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for path %q with patterns %v",
					tt.expected, result, tt.path, tt.patterns)
			}
		})
	}
}

// Test event type string conversion
func TestInjectableFileSystemWatcher_EventTypeString_AllTypes(t *testing.T) {
	t.Parallel()

	watcher := &InjectableFileSystemWatcher{}

	tests := []struct {
		op       fsnotify.Op
		expected string
	}{
		{fsnotify.Create, "create"},
		{fsnotify.Write, "write"},
		{fsnotify.Remove, "remove"},
		{fsnotify.Rename, "rename"},
		{fsnotify.Chmod, "chmod"},
		{fsnotify.Op(0), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()

			result := watcher.eventTypeString(tt.op)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q for op %v", tt.expected, result, tt.op)
			}
		})
	}
}

// Helper function to check if error contains expected message
func containsError(err error, expected string) bool {
	if err == nil {
		return false
	}
	return fmt.Sprintf("%v", err) != "" &&
		(err.Error() == expected ||
			fmt.Sprintf("%v", err) == expected ||
			strings.Contains(err.Error(), expected))
}

// Test AddPath with filesystem error scenarios
func TestInjectableFileSystemWatcher_AddPath_FilesystemErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		path          string
		setupMockFS   func(*mockFileSystem)
		expectedError string
	}{
		{
			name:          "Empty path",
			path:          "",
			setupMockFS:   func(fs *mockFileSystem) {},
			expectedError: "path cannot be empty",
		},
		{
			name: "Abs path error",
			path: "/test",
			setupMockFS: func(fs *mockFileSystem) {
				fs.absFunc = func(path string) (string, error) {
					return "", errors.New("abs path failed")
				}
			},
			expectedError: "failed to get absolute path",
		},
		{
			name:          "Stat error",
			path:          "/test",
			setupMockFS:   func(fs *mockFileSystem) {},
			expectedError: "failed to stat path",
		},
		{
			name: "Walk error",
			path: "/test",
			setupMockFS: func(fs *mockFileSystem) {
				fs.addFile("\\abs\\test", true) // Use Windows-style path
				fs.walkFunc = func(root string, walkFn filepath.WalkFunc) error {
					return errors.New("walk failed")
				}
			},
			expectedError: "failed to walk directory",
		},
		{
			name: "Single file success",
			path: "/test/file.go",
			setupMockFS: func(fs *mockFileSystem) {
				fs.addFile("\\abs\\test\\file.go", false) // Use Windows-style path
			},
			expectedError: "", // No error expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockFS := newMockFileSystem()
			tt.setupMockFS(mockFS)

			deps := &Dependencies{
				FileSystem:   mockFS,
				TimeProvider: newMockTimeProvider(),
				Factory:      newMockWatcherFactory(),
			}

			watcher, err := NewInjectableFileSystemWatcher([]string{}, []string{}, deps)
			if err != nil {
				t.Fatalf("Failed to create watcher: %v", err)
			}

			err = watcher.AddPath(tt.path)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("Expected error but got none")
				}

				if !containsError(err, tt.expectedError) {
					t.Errorf("Expected error containing %q, got: %v", tt.expectedError, err)
				}
			}
		})
	}
}

// Test AddPath with watcher add error
func TestInjectableFileSystemWatcher_AddPath_WatcherAddError(t *testing.T) {
	t.Parallel()

	mockWatcher := newMockFsnotifyWatcher()
	addError := errors.New("watcher add failed")
	mockWatcher.addFunc = func(name string) error {
		return addError
	}

	factory := newMockWatcherFactory()
	factory.watcher = mockWatcher

	mockFS := newMockFileSystem()
	mockFS.addFile("\\abs\\test", true)         // Use Windows-style path
	mockFS.addFile("\\abs\\test\\subdir", true) // Use Windows-style path

	deps := &Dependencies{
		FileSystem:   mockFS,
		TimeProvider: newMockTimeProvider(),
		Factory:      factory,
	}

	watcher, err := NewInjectableFileSystemWatcher([]string{}, []string{}, deps)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	err = watcher.AddPath("/test")
	expectedMsg := "failed to add directory"
	if err == nil || !containsError(err, expectedMsg) {
		t.Errorf("Expected error containing %q, got: %v", expectedMsg, err)
	}
}

// Test RemovePath with error scenarios
func TestInjectableFileSystemWatcher_RemovePath_ErrorScenarios(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		path          string
		setupMockFS   func(*mockFileSystem)
		setupWatcher  func(*mockFsnotifyWatcher)
		expectedError string
	}{
		{
			name:          "Empty path",
			path:          "",
			setupMockFS:   func(fs *mockFileSystem) {},
			setupWatcher:  func(w *mockFsnotifyWatcher) {},
			expectedError: "path cannot be empty",
		},
		{
			name: "Abs path error",
			path: "/test",
			setupMockFS: func(fs *mockFileSystem) {
				fs.absFunc = func(path string) (string, error) {
					return "", errors.New("abs path failed")
				}
			},
			setupWatcher:  func(w *mockFsnotifyWatcher) {},
			expectedError: "failed to get absolute path",
		},
		{
			name:        "Watcher remove error",
			path:        "/test",
			setupMockFS: func(fs *mockFileSystem) {},
			setupWatcher: func(w *mockFsnotifyWatcher) {
				w.removeFunc = func(name string) error {
					return errors.New("remove failed")
				}
			},
			expectedError: "failed to remove path",
		},
		{
			name:          "Successful removal",
			path:          "/test",
			setupMockFS:   func(fs *mockFileSystem) {},
			setupWatcher:  func(w *mockFsnotifyWatcher) {},
			expectedError: "", // No error expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockFS := newMockFileSystem()
			tt.setupMockFS(mockFS)

			mockWatcher := newMockFsnotifyWatcher()
			tt.setupWatcher(mockWatcher)

			factory := newMockWatcherFactory()
			factory.watcher = mockWatcher

			deps := &Dependencies{
				FileSystem:   mockFS,
				TimeProvider: newMockTimeProvider(),
				Factory:      factory,
			}

			watcher, err := NewInjectableFileSystemWatcher([]string{}, []string{}, deps)
			if err != nil {
				t.Fatalf("Failed to create watcher: %v", err)
			}

			// Add the path to the internal list first for successful removal test
			if tt.expectedError == "" {
				watcher.paths = append(watcher.paths, tt.path)
			}

			err = watcher.RemovePath(tt.path)

			if tt.expectedError == "" {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			} else {
				if err == nil {
					t.Fatal("Expected error but got none")
				}

				if !containsError(err, tt.expectedError) {
					t.Errorf("Expected error containing %q, got: %v", tt.expectedError, err)
				}
			}
		})
	}
}

// Test Close with error
func TestInjectableFileSystemWatcher_Close_Error(t *testing.T) {
	t.Parallel()

	mockWatcher := newMockFsnotifyWatcher()
	closeError := errors.New("close failed")
	mockWatcher.closeFunc = func() error {
		return closeError
	}

	factory := newMockWatcherFactory()
	factory.watcher = mockWatcher

	deps := &Dependencies{
		FileSystem:   newMockFileSystem(),
		TimeProvider: newMockTimeProvider(),
		Factory:      factory,
	}

	watcher, err := NewInjectableFileSystemWatcher([]string{}, []string{}, deps)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	err = watcher.Close()
	if err != closeError {
		t.Errorf("Expected close error %v, got: %v", closeError, err)
	}
}

// Test Close with success
func TestInjectableFileSystemWatcher_Close_Success(t *testing.T) {
	t.Parallel()

	mockWatcher := newMockFsnotifyWatcher()

	factory := newMockWatcherFactory()
	factory.watcher = mockWatcher

	deps := &Dependencies{
		FileSystem:   newMockFileSystem(),
		TimeProvider: newMockTimeProvider(),
		Factory:      factory,
	}

	watcher, err := NewInjectableFileSystemWatcher([]string{}, []string{}, deps)
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	err = watcher.Close()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !mockWatcher.closed {
		t.Error("Expected watcher to be closed")
	}
}
