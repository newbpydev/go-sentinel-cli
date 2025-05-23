package cli

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchModeIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "watch_integration_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple test file
	testFile := filepath.Join(tempDir, "example_test.go")
	initialContent := `package main

import "testing"

func TestExample(t *testing.T) {
	if 1+1 != 2 {
		t.Error("Math doesn't work")
	}
}`

	err = os.WriteFile(testFile, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create go.mod file
	goModFile := filepath.Join(tempDir, "go.mod")
	goModContent := "module watch_test\n\ngo 1.23\n"
	err = os.WriteFile(goModFile, []byte(goModContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	t.Run("Config paths conversion", func(t *testing.T) {
		// Test that CLI packages are converted to watch paths correctly
		args := &Args{
			Packages: []string{"./...", "internal/cli"},
			Watch:    true,
		}

		config := GetDefaultConfig()
		merged := config.MergeWithCLIArgs(args)

		if !merged.Watch.Enabled {
			t.Error("Watch mode should be enabled")
		}

		if len(merged.Paths.IncludePatterns) == 0 {
			t.Error("Include patterns should not be empty")
		}

		// Should contain current directory for "./..."
		found := false
		for _, path := range merged.Paths.IncludePatterns {
			if path == "." || path == "internal/cli" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected paths not found in: %v", merged.Paths.IncludePatterns)
		}
	})

	t.Run("File watcher creation", func(t *testing.T) {
		// Test that file watcher can be created with proper paths
		watcher, err := NewFileWatcher(
			[]string{tempDir},
			[]string{"*.log", ".git/*"},
		)
		if err != nil {
			t.Fatalf("Failed to create file watcher: %v", err)
		}
		defer watcher.Close()

		// Verify watcher properties
		if len(watcher.paths) != 1 || watcher.paths[0] != tempDir {
			t.Errorf("Expected paths [%s], got %v", tempDir, watcher.paths)
		}

		if len(watcher.ignorePatterns) < 2 {
			t.Errorf("Expected at least 2 ignore patterns, got %d", len(watcher.ignorePatterns))
		}
	})

	t.Run("File event detection", func(t *testing.T) {
		// Test that file changes are detected
		watcher, err := NewFileWatcher(
			[]string{tempDir},
			[]string{},
		)
		if err != nil {
			t.Fatalf("Failed to create file watcher: %v", err)
		}
		defer watcher.Close()

		events := make(chan FileEvent, 10)
		go func() {
			_ = watcher.Watch(events)
		}()

		// Wait for watcher to initialize
		time.Sleep(100 * time.Millisecond)

		// Modify the test file
		modifiedContent := initialContent + "\n// Modified\n"
		err = os.WriteFile(testFile, []byte(modifiedContent), 0644)
		if err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}

		// Wait for event
		select {
		case event := <-events:
			if filepath.Base(event.Path) != "example_test.go" {
				t.Errorf("Expected event for example_test.go, got %s", event.Path)
			}
			if event.Type != "write" {
				t.Errorf("Expected write event, got %s", event.Type)
			}
			if !event.IsTest {
				t.Error("Expected IsTest to be true for test file")
			}
		case <-time.After(2 * time.Second):
			t.Error("No file change event received within timeout")
		}
	})

	t.Run("Debouncer functionality", func(t *testing.T) {
		// Test that the debouncer works correctly
		debouncer := NewFileEventDebouncer(100 * time.Millisecond)
		defer debouncer.Stop()

		// Add multiple events rapidly
		for i := 0; i < 5; i++ {
			debouncer.AddEvent(FileEvent{
				Path:      testFile,
				Type:      "write",
				Timestamp: time.Now(),
				IsTest:    true,
			})
			time.Sleep(10 * time.Millisecond) // Rapid succession
		}

		// Should get one debounced event
		select {
		case events := <-debouncer.Events():
			if len(events) != 1 {
				t.Errorf("Expected 1 debounced event, got %d", len(events))
			}
			if events[0].Path != testFile {
				t.Errorf("Expected event for %s, got %s", testFile, events[0].Path)
			}
		case <-time.After(500 * time.Millisecond):
			t.Error("No debounced event received within timeout")
		}
	})

	t.Run("Pattern matching", func(t *testing.T) {
		// Test ignore pattern matching
		testCases := []struct {
			path     string
			patterns []string
			expected bool
		}{
			{"test.log", []string{"*.log"}, true},
			{"test.go", []string{"*.log"}, false},
			{".git/config", []string{".git/*"}, true},
			{"src/main.go", []string{".git/*"}, false},
			{"vendor/pkg/file.go", []string{"vendor/**"}, true},
			{"src/vendor.go", []string{"vendor/**"}, false},
		}

		for _, tc := range testCases {
			result := matchesAnyPattern(tc.path, tc.patterns)
			if result != tc.expected {
				t.Errorf("matchesAnyPattern(%s, %v) = %v, expected %v",
					tc.path, tc.patterns, result, tc.expected)
			}
		}
	})

	t.Run("Test determination logic", func(t *testing.T) {
		// Test the logic for determining which tests to run
		controller := NewAppController()

		testCases := []struct {
			changedFile string
			expected    int // number of test packages expected
		}{
			{filepath.Join(tempDir, "example_test.go"), 1}, // Test file -> run its package
			{filepath.Join(tempDir, "main.go"), 1},         // Source file -> run package tests
			{"/some/other/file.txt", 1},                    // Other file -> run directory tests
		}

		for _, tc := range testCases {
			tests := controller.determineTestsToRun(tc.changedFile)
			if len(tests) != tc.expected {
				t.Errorf("determineTestsToRun(%s) returned %d tests, expected %d",
					tc.changedFile, len(tests), tc.expected)
			}
		}
	})
}

func TestWatchModeConfiguration(t *testing.T) {
	t.Run("Default ignore patterns", func(t *testing.T) {
		config := GetDefaultConfig()

		// Should have comprehensive ignore patterns
		if len(config.Watch.IgnorePatterns) < 10 {
			t.Errorf("Expected at least 10 ignore patterns, got %d", len(config.Watch.IgnorePatterns))
		}

		// Check for essential patterns
		essentialPatterns := []string{"*.log", ".git/*", "vendor/*", "node_modules/*"}
		for _, pattern := range essentialPatterns {
			found := false
			for _, ignore := range config.Watch.IgnorePatterns {
				if ignore == pattern {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Essential ignore pattern %s not found", pattern)
			}
		}
	})

	t.Run("Package to path conversion", func(t *testing.T) {
		testCases := []struct {
			packages []string
			expected []string
		}{
			{[]string{"./..."}, []string{"."}},
			{[]string{"."}, []string{"."}},
			{[]string{"internal/cli"}, []string{"internal/cli"}},
			{[]string{"pkg/...", "cmd/..."}, []string{"pkg", "cmd"}},
			{[]string{".", ".", "internal"}, []string{".", "internal"}}, // Deduplication
		}

		for _, tc := range testCases {
			result := convertPackagesToWatchPaths(tc.packages)
			if len(result) != len(tc.expected) {
				t.Errorf("convertPackagesToWatchPaths(%v) = %v, expected %v",
					tc.packages, result, tc.expected)
				continue
			}

			for i, expected := range tc.expected {
				if result[i] != expected {
					t.Errorf("convertPackagesToWatchPaths(%v)[%d] = %s, expected %s",
						tc.packages, i, result[i], expected)
				}
			}
		}
	})
}
