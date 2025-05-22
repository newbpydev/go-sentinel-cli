package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestRecoverFromTestRunnerCrash tests recovery from test runner crashes
func TestRecoverFromTestRunnerCrash(t *testing.T) {
	// Create a mock test runner that simulates crashes
	processor := NewTestProcessor(io.Discard, NewColorFormatter(false), NewIconProvider(false), 80)

	// Simulate malformed JSON that could cause a crash
	malformedJSON := `{"Time":"2024-01-01T10:00:00.000Z","Action":"run","Package":"github.com/test/example"`

	parser := NewStreamParser()
	reader := strings.NewReader(malformedJSON)
	results := make(chan *TestResult, 10)

	// This should not crash the application
	err := parser.Parse(reader, results)
	close(results)

	// Should handle the error gracefully
	if err == nil {
		t.Error("Expected error from malformed JSON, got nil")
	}

	// Processor should still be functional after error
	suite := &TestSuite{
		FilePath:     "test/recovery_test.go",
		TestCount:    1,
		PassedCount:  1,
		FailedCount:  0,
		SkippedCount: 0,
	}

	processor.AddTestSuite(suite)
	err = processor.RenderResults(false)
	if err != nil {
		t.Errorf("Processor should recover after parser error, got: %v", err)
	}
}

// TestHandleFilesystemPermissionErrors tests handling of filesystem permission errors
func TestHandleFilesystemPermissionErrors(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "permission_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a file with restricted permissions
	restrictedFile := filepath.Join(tempDir, "restricted.go")
	err = os.WriteFile(restrictedFile, []byte("package main"), 0000) // No permissions
	if err != nil {
		t.Fatalf("Failed to create restricted file: %v", err)
	}

	// Test source code extractor with permission error
	extractor := NewSourceExtractor()

	// Try to extract context - behavior varies by platform
	context, err := extractor.ExtractContext(restrictedFile, 1, 5)

	// On Windows, permission restrictions work differently than Unix systems
	// The test should verify graceful handling regardless of platform behavior
	if err != nil {
		// If we get an error, context should be empty
		if len(context) != 0 {
			t.Errorf("Expected empty context on error, got %d lines", len(context))
		}
		t.Logf("Permission error handled gracefully: %v", err)
	} else {
		// On some platforms (like Windows), we might still be able to read the file
		// In this case, verify we got some content
		t.Logf("File readable despite permission restrictions, got %d lines", len(context))
	}

	// The key requirement is that the extractor doesn't crash
	// and handles the situation gracefully, regardless of platform-specific behavior
}

// TestRecoverFromSyntaxErrors tests recovery from syntax errors in test files
func TestRecoverFromSyntaxErrors(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "syntax_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a file with syntax errors
	syntaxErrorFile := filepath.Join(tempDir, "syntax_error.go")
	syntaxErrorContent := `package main
	
	func TestBrokenSyntax(t *testing.T) {
		// Missing closing brace
		if true {
			t.Log("This test has syntax errors"
		// Missing closing brace for function
	`

	err = os.WriteFile(syntaxErrorFile, []byte(syntaxErrorContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create syntax error file: %v", err)
	}

	// Test source code extractor with syntax error file
	extractor := NewSourceExtractor()

	// Should extract context even with syntax errors
	context, err := extractor.ExtractContext(syntaxErrorFile, 3, 5)

	// Should not error when extracting context from syntactically invalid files
	if err != nil {
		t.Errorf("Source extractor should handle syntax errors gracefully, got: %v", err)
	}

	// Should still return context lines
	if len(context) == 0 {
		t.Error("Expected context lines even from syntactically invalid file")
	}

	// Test with invalid line numbers
	context, err = extractor.ExtractContext(syntaxErrorFile, 999, 5)

	// Should handle invalid line numbers gracefully
	if err != nil {
		t.Errorf("Should handle invalid line numbers gracefully, got: %v", err)
	}

	if len(context) != 0 {
		t.Errorf("Expected empty context for invalid line number, got %d lines", len(context))
	}
}

// TestStableBehaviorWithCorruptedFiles tests stable behavior with corrupted/inconsistent Go files
func TestStableBehaviorWithCorruptedFiles(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "corrupted_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testCases := []struct {
		name     string
		content  string
		filename string
	}{
		{
			name:     "binary_content",
			content:  "\x00\x01\x02\x03\xFF\xFE\xFD", // Binary content
			filename: "binary.go",
		},
		{
			name:     "empty_file",
			content:  "",
			filename: "empty.go",
		},
		{
			name:     "mixed_encoding",
			content:  "package main\n\xFF\xFE// This has mixed encoding\nfunc Test() {}",
			filename: "mixed.go",
		},
		{
			name:     "very_long_lines",
			content:  "package main\n// " + strings.Repeat("A", 10000) + "\nfunc Test() {}",
			filename: "longlines.go",
		},
	}

	extractor := NewSourceExtractor()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := filepath.Join(tempDir, tc.filename)
			err := os.WriteFile(filePath, []byte(tc.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Should handle corrupted files without crashing
			context, err := extractor.ExtractContext(filePath, 1, 5)

			// Should not crash, but may return error for binary files
			if err != nil {
				t.Logf("Extractor returned error for %s: %v", tc.name, err)
			}

			// Should return reasonable context (or empty for binary files)
			t.Logf("Context lines for %s: %d", tc.name, len(context))
		})
	}
}

// TestConcurrentProcessing tests stability under concurrent processing
func TestConcurrentProcessing(t *testing.T) {
	processor := NewOptimizedTestProcessor(io.Discard, NewColorFormatter(false), NewIconProvider(false), 80)

	// Number of concurrent goroutines
	const numGoroutines = 10
	const suitsPerGoroutine = 5

	// Channel to collect errors
	errorChan := make(chan error, numGoroutines*suitsPerGoroutine)
	done := make(chan bool, numGoroutines)

	// Start multiple goroutines adding test suites
	for i := 0; i < numGoroutines; i++ {
		go func(routineID int) {
			defer func() { done <- true }()

			for j := 0; j < suitsPerGoroutine; j++ {
				suite := &TestSuite{
					FilePath:     fmt.Sprintf("test/concurrent_test_%d_%d.go", routineID, j),
					TestCount:    10,
					PassedCount:  8,
					FailedCount:  2,
					SkippedCount: 0,
					Duration:     time.Millisecond * time.Duration(j+1),
				}

				// Add tests to suite
				for k := 0; k < 10; k++ {
					status := StatusPassed
					if k%5 == 0 {
						status = StatusFailed
					}
					test := &TestResult{
						Name:     fmt.Sprintf("TestConcurrent_%d_%d_%d", routineID, j, k),
						Status:   status,
						Duration: time.Millisecond,
						Package:  "github.com/test/concurrent",
					}
					if status == StatusFailed {
						test.Error = &TestError{
							Message: "Concurrent test failed",
							Type:    "AssertionError",
						}
					}
					suite.Tests = append(suite.Tests, test)
				}

				processor.AddTestSuite(suite)

				// Try to render results with timeout to prevent deadlocks
				renderDone := make(chan error, 1)
				go func() {
					renderDone <- processor.RenderResultsOptimized(false)
				}()

				select {
				case err := <-renderDone:
					if err != nil {
						errorChan <- err
						return
					}
				case <-time.After(5 * time.Second):
					errorChan <- fmt.Errorf("render operation timed out - possible deadlock")
					return
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete with overall timeout
	completedGoroutines := 0
	overallTimeout := time.After(30 * time.Second)

	for completedGoroutines < numGoroutines {
		select {
		case <-done:
			completedGoroutines++
		case err := <-errorChan:
			t.Errorf("Concurrent processing error: %v", err)
		case <-overallTimeout:
			t.Fatal("Overall concurrent processing timed out - possible deadlock")
		}
	}

	// Check for any remaining errors
	select {
	case err := <-errorChan:
		t.Errorf("Additional concurrent processing error: %v", err)
	default:
		// No errors
	}
}

// TestWatchModeStability tests stability of watch mode under various conditions
func TestWatchModeStability(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "watch_stability_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create initial test file
	testFile := filepath.Join(tempDir, "watch_test.go")
	initialContent := `package main

import "testing"

func TestWatch(t *testing.T) {
	t.Log("Initial test")
}`
	err = os.WriteFile(testFile, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create initial test file: %v", err)
	}

	// Create file watcher
	watcher, err := NewFileWatcher([]string{tempDir}, []string{})
	if err != nil {
		t.Fatalf("Failed to create file watcher: %v", err)
	}
	defer watcher.Close()

	// Start watching in background
	changes := make(chan FileEvent, 10)
	go func() {
		_ = watcher.Watch(changes)
	}()

	// Test rapid file changes
	for i := 0; i < 5; i++ {
		modifiedContent := initialContent + fmt.Sprintf("// Modification %d\n", i)
		err = os.WriteFile(testFile, []byte(modifiedContent), 0644)
		if err != nil {
			t.Errorf("Failed to modify test file: %v", err)
		}
		// Short delay between modifications
		time.Sleep(10 * time.Millisecond)
	}

	// Collect changes with timeout
	var collectedChanges []FileEvent
	timeout := time.After(2 * time.Second)
collectLoop:
	for {
		select {
		case change := <-changes:
			collectedChanges = append(collectedChanges, change)
			if len(collectedChanges) >= 3 { // Expect at least some changes
				break collectLoop
			}
		case <-timeout:
			break collectLoop
		}
	}

	if len(collectedChanges) == 0 {
		t.Error("Expected file change events, got none")
	}

	// Test file deletion
	err = os.Remove(testFile)
	if err != nil {
		t.Errorf("Failed to remove test file: %v", err)
	}

	// Should handle file deletion gracefully
	select {
	case change := <-changes:
		t.Logf("Received change event after deletion: %+v", change)
	case <-time.After(1 * time.Second):
		// May or may not receive deletion event depending on system
	}
}

// TestLargeOutputHandling tests handling of very large test outputs
func TestLargeOutputHandling(t *testing.T) {
	processor := NewTestProcessor(io.Discard, NewColorFormatter(false), NewIconProvider(false), 80)

	// Create a test with very large output
	largeOutput := strings.Repeat("This is a very long test output line that simulates verbose test output.\n", 10000)

	suite := &TestSuite{
		FilePath:     "test/large_output_test.go",
		TestCount:    1,
		PassedCount:  0,
		FailedCount:  1,
		SkippedCount: 0,
	}

	test := &TestResult{
		Name:     "TestLargeOutput",
		Status:   StatusFailed,
		Duration: time.Second,
		Package:  "github.com/test/large",
		Output:   largeOutput, // Very large output
		Error: &TestError{
			Message: "Test failed with large output",
			Type:    "AssertionError",
			Location: &SourceLocation{
				File: "test/large_output_test.go",
				Line: 42,
			},
		},
	}

	suite.Tests = append(suite.Tests, test)

	// Should handle large output without crashing
	start := time.Now()
	processor.AddTestSuite(suite)
	err := processor.RenderResults(false)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("Failed to process large output: %v", err)
	}

	// Should complete in reasonable time (under 5 seconds)
	if elapsed > 5*time.Second {
		t.Errorf("Large output processing too slow: %v", elapsed)
	}

	t.Logf("Large output processing completed in %v", elapsed)
}
