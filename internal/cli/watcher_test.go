package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileWatcher(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "watcher-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("failed to remove temp dir: %v", err)
		}
	}()

	// Create some test files
	testFiles := []string{
		filepath.Join(tempDir, "test_file.go"),
		filepath.Join(tempDir, "test_file_test.go"),
		filepath.Join(tempDir, "implementation.go"),
	}

	for _, file := range testFiles {
		// #nosec G306 - Test file, permissions not important
		if err := os.WriteFile(file, []byte("package test"), 0600); err != nil {
			t.Fatalf("failed to create test file %s: %v", file, err)
		}
	}

	tests := []struct {
		name           string
		paths          []string
		ignorePatterns []string
		testPatterns   []string
		writeFile      string
		expectedEvent  string
	}{
		{
			name:          "detects changes to test files",
			paths:         []string{tempDir},
			testPatterns:  []string{"*_test.go"},
			writeFile:     filepath.Join(tempDir, "test_file_test.go"),
			expectedEvent: "test_file_test.go",
		},
		{
			name:          "detects changes to implementation files",
			paths:         []string{tempDir},
			testPatterns:  []string{"*_test.go"},
			writeFile:     filepath.Join(tempDir, "implementation.go"),
			expectedEvent: "implementation.go",
		},
		{
			name:           "respects ignore patterns",
			paths:          []string{tempDir},
			ignorePatterns: []string{"implementation.go"},
			writeFile:      filepath.Join(tempDir, "implementation.go"),
			expectedEvent:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create watcher
			watcher, err := NewFileWatcher(tt.paths, tt.ignorePatterns)
			if err != nil {
				t.Fatalf("failed to create watcher: %v", err)
			}
			defer func() {
				if err := watcher.Close(); err != nil {
					t.Logf("failed to close watcher: %v", err)
				}
			}()

			// Start watching for changes
			eventCh := make(chan FileEvent, 10)
			errCh := make(chan error, 1)
			go func() {
				err := watcher.Watch(eventCh)
				if err != nil {
					errCh <- err
				}
			}()

			// Wait a bit for the watcher to initialize
			time.Sleep(100 * time.Millisecond)

			// Modify the file to trigger an event
			// #nosec G306 - Test file, permissions not important
			if err := os.WriteFile(tt.writeFile, []byte("package test // modified"), 0600); err != nil {
				t.Fatalf("failed to modify file: %v", err)
			}

			// Check for errors from the watcher
			select {
			case err := <-errCh:
				t.Fatalf("watcher error: %v", err)
			default:
			}

			// Wait for events or timeout
			var receivedEvent string
			timeout := time.After(2 * time.Second)
			select {
			case event := <-eventCh:
				receivedEvent = event.Path
			case <-timeout:
				// No event received
			}

			// Verify expectations
			if tt.expectedEvent == "" {
				if receivedEvent != "" {
					t.Errorf("expected no event, got %s", receivedEvent)
				}
			} else {
				if filepath.Base(receivedEvent) != filepath.Base(tt.expectedEvent) {
					t.Errorf("expected event for %s, got %s", tt.expectedEvent, receivedEvent)
				}
			}
		})
	}
}

func TestTestFileFinder(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "finder-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("failed to remove temp dir: %v", err)
		}
	}()

	// Create test directory structure
	pkgDir := filepath.Join(tempDir, "pkg")
	// #nosec G301 - Test directory, permissions not important
	if err := os.Mkdir(pkgDir, 0700); err != nil {
		t.Fatalf("failed to create package dir: %v", err)
	}

	testFiles := map[string]string{
		filepath.Join(pkgDir, "foo.go"):      "package pkg",
		filepath.Join(pkgDir, "foo_test.go"): "package pkg_test",
		filepath.Join(pkgDir, "bar.go"):      "package pkg",
		filepath.Join(pkgDir, "bar_test.go"): "package pkg_test",
		filepath.Join(pkgDir, "baz.go"):      "package pkg",
	}

	for file, content := range testFiles {
		// #nosec G306 - Test file, permissions not important
		if err := os.WriteFile(file, []byte(content), 0600); err != nil {
			t.Fatalf("failed to create file %s: %v", file, err)
		}
	}

	tests := []struct {
		name         string
		file         string
		expectedTest string
	}{
		{
			name:         "finds direct test file",
			file:         filepath.Join(pkgDir, "foo.go"),
			expectedTest: filepath.Join(pkgDir, "foo_test.go"),
		},
		{
			name:         "handles test file input",
			file:         filepath.Join(pkgDir, "foo_test.go"),
			expectedTest: filepath.Join(pkgDir, "foo_test.go"),
		},
		{
			name:         "handles missing test file",
			file:         filepath.Join(pkgDir, "baz.go"),
			expectedTest: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := NewTestFileFinder(tempDir)
			testFile, err := finder.FindTestFile(tt.file)

			if tt.expectedTest == "" {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if testFile != tt.expectedTest {
					t.Errorf("expected %s, got %s", tt.expectedTest, testFile)
				}
			}
		})
	}
}

func TestWatcherUtils(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		patterns       []string
		expectedResult bool
	}{
		{
			name:           "matches test file",
			path:           "foo_test.go",
			patterns:       []string{"*_test.go"},
			expectedResult: true,
		},
		{
			name:           "doesn't match test file",
			path:           "foo.go",
			patterns:       []string{"*_test.go"},
			expectedResult: false,
		},
		{
			name:           "matches multiple patterns",
			path:           "foo.go",
			patterns:       []string{"*.go", "*.ts"},
			expectedResult: true,
		},
		{
			name:           "respects ignore patterns",
			path:           "node_modules/foo.js",
			patterns:       []string{"node_modules/**"},
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesAnyPattern(tt.path, tt.patterns)
			if result != tt.expectedResult {
				t.Errorf("expected %v, got %v", tt.expectedResult, result)
			}
		})
	}
}
