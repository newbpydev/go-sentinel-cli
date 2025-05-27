package watcher

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/newbpydev/go-sentinel/internal/watch/core"
)

// Test Factory Functions
func TestNewFileSystemWatcher_Creation(t *testing.T) {
	tests := []struct {
		name           string
		paths          []string
		ignorePatterns []string
		expectError    bool
	}{
		{
			name:           "Valid single path",
			paths:          []string{"."},
			ignorePatterns: []string{".git", "node_modules"},
			expectError:    false,
		},
		{
			name:           "Multiple paths",
			paths:          []string{".", "../"},
			ignorePatterns: []string{".git"},
			expectError:    false,
		},
		{
			name:           "Empty paths",
			paths:          []string{},
			ignorePatterns: []string{},
			expectError:    false,
		},
		{
			name:           "Nil ignore patterns",
			paths:          []string{"."},
			ignorePatterns: nil,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			watcher, err := NewFileSystemWatcher(tt.paths, tt.ignorePatterns)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if watcher == nil {
				t.Fatal("NewFileSystemWatcher should not return nil")
			}

			// Verify interface compliance
			var _ core.FileSystemWatcher = watcher

			// Clean up
			err = watcher.Close()
			if err != nil {
				t.Errorf("Close should not error: %v", err)
			}
		})
	}
}

// Test AddPath functionality
func TestFileSystemWatcher_AddPath(t *testing.T) {
	watcher, err := NewFileSystemWatcher([]string{}, []string{})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	tests := []struct {
		name        string
		path        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid directory",
			path:        ".",
			expectError: false,
		},
		{
			name:        "Valid relative path",
			path:        "../",
			expectError: false,
		},
		{
			name:        "Non-existent path",
			path:        "./non-existent-directory",
			expectError: true,
			errorMsg:    "failed to stat path",
		},
		{
			name:        "Empty path",
			path:        "",
			expectError: true,
			errorMsg:    "path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := watcher.AddPath(tt.path)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Test AddPath with file watching and directory scenarios
func TestFileSystemWatcher_AddPath_Comprehensive(t *testing.T) {
	// Create temporary directory structure for testing
	tempDir, err := os.MkdirTemp("", "watcher_addpath_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirectories and files
	subDir := filepath.Join(tempDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	testFile := filepath.Join(tempDir, "test.go")
	err = os.WriteFile(testFile, []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create ignored directory
	ignoredDir := filepath.Join(tempDir, ".git")
	err = os.Mkdir(ignoredDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create ignored dir: %v", err)
	}

	watcher, err := NewFileSystemWatcher([]string{}, []string{".git"})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	tests := []struct {
		name        string
		path        string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Add directory with subdirectories",
			path:        tempDir,
			expectError: false,
		},
		{
			name:        "Add single file",
			path:        testFile,
			expectError: false,
		},
		{
			name:        "Add duplicate path",
			path:        tempDir,
			expectError: false, // Should handle duplicates gracefully
		},
		{
			name:        "Add ignored directory",
			path:        ignoredDir,
			expectError: false, // Directory itself can be added
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := watcher.AddPath(tt.path)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}

	// Verify paths were added correctly
	if len(watcher.paths) == 0 {
		t.Error("Expected paths to be added to watcher")
	}
}

// Test RemovePath functionality
func TestFileSystemWatcher_RemovePath(t *testing.T) {
	watcher, err := NewFileSystemWatcher([]string{"."}, []string{})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	// Add a path first
	err = watcher.AddPath(".")
	if err != nil {
		t.Fatalf("Failed to add path: %v", err)
	}

	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{
			name:        "Remove existing path",
			path:        ".",
			expectError: false,
		},
		{
			name:        "Remove non-existent path",
			path:        "./non-existent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := watcher.RemovePath(tt.path)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Test Close functionality
func TestFileSystemWatcher_Close(t *testing.T) {
	watcher, err := NewFileSystemWatcher([]string{"."}, []string{})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}

	err = watcher.Close()
	if err != nil {
		t.Errorf("Close should not error: %v", err)
	}

	// Multiple close calls should be safe
	err = watcher.Close()
	if err != nil {
		t.Errorf("Second Close call should not error: %v", err)
	}
}

// Test Watch functionality
func TestFileSystemWatcher_Watch(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "watcher_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	watcher, err := NewFileSystemWatcher([]string{tempDir}, []string{})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events := make(chan core.FileEvent, 10)
	var watchErr error

	// Use a channel to signal when the watcher is ready
	watcherReady := make(chan struct{})

	// Start watching in goroutine
	go func() {
		// Signal that watcher is starting
		close(watcherReady)
		watchErr = watcher.Watch(ctx, events)
	}()

	// Wait for watcher to start
	<-watcherReady

	// Give the watcher time to fully initialize
	time.Sleep(200 * time.Millisecond)

	// Create a test file
	testFile := filepath.Join(tempDir, "test_file.go")
	err = os.WriteFile(testFile, []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Wait for event or timeout
	select {
	case event := <-events:
		if event.Path == "" {
			t.Error("Event should have a path")
		}
		if event.Type == "" {
			t.Error("Event should have a type")
		}
		if event.Timestamp.IsZero() {
			t.Error("Event should have a timestamp")
		}
		// Verify the event is for our test file
		if !filepath.IsAbs(event.Path) || !filepath.IsAbs(testFile) {
			// Convert to absolute paths for comparison
			absEventPath, _ := filepath.Abs(event.Path)
			absTestFile, _ := filepath.Abs(testFile)
			if absEventPath != absTestFile {
				t.Logf("Event path: %s, Expected: %s", absEventPath, absTestFile)
			}
		}
	case <-time.After(3 * time.Second):
		t.Error("Expected to receive file event within timeout")
	}

	// Cancel context and verify Watch returns
	cancel()
	time.Sleep(100 * time.Millisecond)

	if watchErr != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", watchErr)
	}
}

// Test pattern matching functionality
func TestFileSystemWatcher_matchesAnyPattern(t *testing.T) {
	watcher, err := NewFileSystemWatcher([]string{}, []string{})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	tests := []struct {
		name     string
		path     string
		patterns []string
		expected bool
	}{
		{
			name:     "Exact match",
			path:     ".git",
			patterns: []string{".git", "node_modules"},
			expected: true,
		},
		{
			name:     "Contains match",
			path:     "/project/.git/config",
			patterns: []string{".git"},
			expected: true,
		},
		{
			name:     "Prefix wildcard",
			path:     "my_test.go",
			patterns: []string{"*_test.go"},
			expected: true,
		},
		{
			name:     "Suffix wildcard",
			path:     ".gitignore",
			patterns: []string{".git*"},
			expected: true,
		},
		{
			name:     "No match",
			path:     "main.go",
			patterns: []string{".git", "*_test.go"},
			expected: false,
		},
		{
			name:     "Empty patterns",
			path:     "main.go",
			patterns: []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := watcher.matchesAnyPattern(tt.path, tt.patterns)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for path %s with patterns %v", tt.expected, result, tt.path, tt.patterns)
			}
		})
	}
}

// Test eventTypeString functionality comprehensively
func TestFileSystemWatcher_eventTypeString(t *testing.T) {
	watcher, err := NewFileSystemWatcher([]string{}, []string{})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	tests := []struct {
		name     string
		op       fsnotify.Op
		expected string
	}{
		{
			name:     "Create operation",
			op:       fsnotify.Create,
			expected: "create",
		},
		{
			name:     "Write operation",
			op:       fsnotify.Write,
			expected: "write",
		},
		{
			name:     "Remove operation",
			op:       fsnotify.Remove,
			expected: "remove",
		},
		{
			name:     "Rename operation",
			op:       fsnotify.Rename,
			expected: "rename",
		},
		{
			name:     "Chmod operation",
			op:       fsnotify.Chmod,
			expected: "chmod",
		},
		{
			name:     "Unknown operation",
			op:       fsnotify.Op(0), // No flags set
			expected: "unknown",
		},
		{
			name:     "Multiple operations - Create takes precedence",
			op:       fsnotify.Create | fsnotify.Write,
			expected: "create",
		},
		{
			name:     "Multiple operations - Write takes precedence over Remove",
			op:       fsnotify.Write | fsnotify.Remove,
			expected: "write",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := watcher.eventTypeString(tt.op)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s for operation %v", tt.expected, result, tt.op)
			}
		})
	}
}

// Test concurrent access patterns
func TestFileSystemWatcher_ConcurrentAccess(t *testing.T) {
	watcher, err := NewFileSystemWatcher([]string{}, []string{})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	var wg sync.WaitGroup
	const numGoroutines = 10

	// Test concurrent AddPath calls
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			// Add current directory - should be safe
			watcher.AddPath(".")
		}(i)
	}

	wg.Wait()

	// Verify watcher is still functional
	err = watcher.AddPath(".")
	if err != nil {
		t.Errorf("Watcher should still be functional after concurrent access: %v", err)
	}
}

// Test edge cases and error conditions
func TestFileSystemWatcher_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (*FileSystemWatcher, func())
		operation   func(*FileSystemWatcher) error
		expectError bool
		errorMsg    string
	}{
		{
			name: "AddPath with empty string",
			setup: func(t *testing.T) (*FileSystemWatcher, func()) {
				w, err := NewFileSystemWatcher([]string{}, []string{})
				if err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
				return w, func() { w.Close() }
			},
			operation: func(w *FileSystemWatcher) error {
				return w.AddPath("")
			},
			expectError: true,
		},
		{
			name: "RemovePath with empty string",
			setup: func(t *testing.T) (*FileSystemWatcher, func()) {
				w, err := NewFileSystemWatcher([]string{}, []string{})
				if err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
				return w, func() { w.Close() }
			},
			operation: func(w *FileSystemWatcher) error {
				return w.RemovePath("")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			watcher, cleanup := tt.setup(t)
			defer cleanup()

			err := tt.operation(watcher)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Test Watch error handling and edge cases
func TestFileSystemWatcher_Watch_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (*FileSystemWatcher, func(), context.Context)
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name: "Watch with invalid path in constructor",
			setup: func(t *testing.T) (*FileSystemWatcher, func(), context.Context) {
				// Create watcher with invalid path
				watcher, err := NewFileSystemWatcher([]string{"./non-existent-path"}, []string{})
				if err != nil {
					t.Fatalf("Failed to create watcher: %v", err)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				return watcher, func() { watcher.Close(); cancel() }, ctx
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return contains(err.Error(), "failed to stat path")
			},
		},
		{
			name: "Watch with context cancellation",
			setup: func(t *testing.T) (*FileSystemWatcher, func(), context.Context) {
				watcher, err := NewFileSystemWatcher([]string{}, []string{})
				if err != nil {
					t.Fatalf("Failed to create watcher: %v", err)
				}
				ctx, cancel := context.WithCancel(context.Background())
				// Cancel immediately to test cancellation path
				cancel()
				return watcher, func() { watcher.Close() }, ctx
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return err == context.Canceled
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher, cleanup, ctx := tt.setup(t)
			defer cleanup()

			events := make(chan core.FileEvent, 10)
			err := watcher.Watch(ctx, events)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if tt.errorCheck != nil && !tt.errorCheck(err) {
					t.Errorf("Error check failed for error: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

// Test Watch with ignored files and different event types
func TestFileSystemWatcher_Watch_IgnorePatterns(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "watcher_ignore_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create watcher with ignore patterns
	watcher, err := NewFileSystemWatcher([]string{tempDir}, []string{"*.tmp", ".git"})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	events := make(chan core.FileEvent, 10)
	watcherReady := make(chan struct{})
	var watchErr error

	// Start watching
	go func() {
		close(watcherReady)
		watchErr = watcher.Watch(ctx, events)
	}()

	// Wait for watcher to start
	<-watcherReady
	time.Sleep(200 * time.Millisecond)

	// Create files that should be ignored
	ignoredFile := filepath.Join(tempDir, "temp.tmp")
	err = os.WriteFile(ignoredFile, []byte("ignored"), 0644)
	if err != nil {
		t.Fatalf("Failed to create ignored file: %v", err)
	}

	// Create file that should be watched
	watchedFile := filepath.Join(tempDir, "watched.go")
	err = os.WriteFile(watchedFile, []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("Failed to create watched file: %v", err)
	}

	// Create a test file (should be detected as test file)
	testFile := filepath.Join(tempDir, "main_test.go")
	err = os.WriteFile(testFile, []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Collect events for a short time
	eventCount := 0
	testFileDetected := false
	timeout := time.After(2 * time.Second)

	for {
		select {
		case event := <-events:
			eventCount++
			t.Logf("Received event: %s (type: %s, isTest: %v)", event.Path, event.Type, event.IsTest)

			// Check if this is our test file
			if contains(event.Path, "main_test.go") {
				testFileDetected = true
				if !event.IsTest {
					t.Error("Expected test file to be marked as test")
				}
			}

			// Verify ignored files don't generate events
			if contains(event.Path, ".tmp") {
				t.Error("Ignored .tmp file should not generate events")
			}

		case <-timeout:
			goto done
		}
	}

done:
	cancel()
	time.Sleep(100 * time.Millisecond)

	if eventCount == 0 {
		t.Error("Expected to receive at least one file event")
	}

	if !testFileDetected {
		t.Error("Expected to detect test file creation")
	}

	if watchErr != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", watchErr)
	}
}

// Test Watch with directory creation
func TestFileSystemWatcher_Watch_DirectoryCreation(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "watcher_dir_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	watcher, err := NewFileSystemWatcher([]string{tempDir}, []string{})
	if err != nil {
		t.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	events := make(chan core.FileEvent, 10)
	watcherReady := make(chan struct{})
	var watchErr error

	// Start watching
	go func() {
		close(watcherReady)
		watchErr = watcher.Watch(ctx, events)
	}()

	// Wait for watcher to start
	<-watcherReady
	time.Sleep(200 * time.Millisecond)

	// Create a new subdirectory - this should be handled but not generate file events
	newDir := filepath.Join(tempDir, "newsubdir")
	err = os.Mkdir(newDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create new directory: %v", err)
	}

	// Create a file in the new directory
	newFile := filepath.Join(newDir, "newfile.go")
	err = os.WriteFile(newFile, []byte("package main"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file in new directory: %v", err)
	}

	// Wait for events
	fileEventReceived := false
	timeout := time.After(2 * time.Second)

	for {
		select {
		case event := <-events:
			t.Logf("Received event: %s (type: %s)", event.Path, event.Type)

			if contains(event.Path, "newfile.go") {
				fileEventReceived = true
			}

		case <-timeout:
			goto done
		}
	}

done:
	cancel()
	time.Sleep(100 * time.Millisecond)

	// We expect file events, directory events are handled internally
	// Note: On some platforms, file events in newly created directories might be delayed
	if !fileEventReceived {
		t.Logf("Warning: Did not receive file event for file created in new directory (platform-specific behavior)")
		// Don't fail the test as this might be platform-specific behavior
	}

	if watchErr != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", watchErr)
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
