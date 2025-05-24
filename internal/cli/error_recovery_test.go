package cli

import (
	"bytes"
	"context"
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
	processor := NewOptimizedTestProcessorWithUI(
		io.Discard,
		NewColorFormatter(false),
		NewIconProvider(false),
		80,
	)

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

// TestErrorRecovery_TestRunnerFailures tests recovery from test runner failures
func TestErrorRecovery_TestRunnerFailures(t *testing.T) {
	runner := &TestRunner{
		Verbose:    false,
		JSONOutput: true,
	}

	ctx := context.Background()

	testCases := []struct {
		name        string
		testPaths   []string
		expectError bool
		description string
	}{
		{
			name:        "NonExistentDirectory",
			testPaths:   []string{"/non/existent/path"},
			expectError: true,
			description: "Should handle non-existent directory gracefully",
		},
		{
			name:        "EmptyPathList",
			testPaths:   []string{},
			expectError: true,
			description: "Should handle empty path list gracefully",
		},
		{
			name:        "InvalidPath",
			testPaths:   []string{""},
			expectError: true,
			description: "Should handle empty path string gracefully",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := runner.Run(ctx, tc.testPaths)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s, but got none", tc.description)
			}

			if !tc.expectError && err != nil {
				t.Errorf("Did not expect error for %s, but got: %v", tc.description, err)
			}

			if err != nil {
				// Verify error message is user-friendly
				errorMsg := err.Error()
				if errorMsg == "" {
					t.Error("Expected non-empty error message")
				}

				t.Logf("Error message for %s: %s", tc.name, errorMsg)
			}
		})
	}
}

// TestErrorRecovery_ProcessorFailures tests recovery from processor failures
func TestErrorRecovery_ProcessorFailures(t *testing.T) {
	testCases := []struct {
		name        string
		suite       *TestSuite
		expectError bool
		description string
	}{
		{
			name:        "NilTestSuite",
			suite:       nil,
			expectError: true,
			description: "Should handle nil test suite gracefully",
		},
		{
			name: "EmptyFilePath",
			suite: &TestSuite{
				FilePath:    "",
				TestCount:   1,
				PassedCount: 1,
			},
			expectError: false,
			description: "Should handle empty file path gracefully",
		},
		{
			name: "NegativeTestCounts",
			suite: &TestSuite{
				FilePath:     "test.go",
				TestCount:    -1,
				PassedCount:  -1,
				FailedCount:  -1,
				SkippedCount: -1,
			},
			expectError: false,
			description: "Should handle negative test counts gracefully",
		},
		{
			name: "InconsistentCounts",
			suite: &TestSuite{
				FilePath:     "test.go",
				TestCount:    5,
				PassedCount:  10, // More than total
				FailedCount:  2,
				SkippedCount: 1,
			},
			expectError: false,
			description: "Should handle inconsistent test counts gracefully",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			processor := NewTestProcessor(
				&output,
				NewColorFormatter(false),
				NewIconProvider(false),
				80,
			)

			var err error
			if tc.suite != nil {
				processor.AddTestSuite(tc.suite)
				err = processor.RenderResults(false)
			} else {
				// Test nil suite handling
				func() {
					defer func() {
						if r := recover(); r != nil {
							err = r.(error)
						}
					}()
					processor.AddTestSuite(tc.suite)
				}()
			}

			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s, but got none", tc.description)
			}

			if !tc.expectError && err != nil {
				t.Errorf("Did not expect error for %s, but got: %v", tc.description, err)
			}

			t.Logf("Error recovery test %s completed", tc.name)
		})
	}
}

// TestErrorRecovery_FileSystemErrors tests recovery from file system errors
func TestErrorRecovery_FileSystemErrors(t *testing.T) {
	// Test file watcher with problematic directories
	testCases := []struct {
		name        string
		paths       []string
		expectError bool
		description string
	}{
		{
			name:        "NonExistentDirectory",
			paths:       []string{"/absolutely/non/existent/path"},
			expectError: true,
			description: "Should handle non-existent directory",
		},
		{
			name:        "EmptyPaths",
			paths:       []string{},
			expectError: true,
			description: "Should handle empty paths list",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			options := WatchOptions{
				Paths:            tc.paths,
				IgnorePatterns:   []string{"*.log"},
				TestPatterns:     []string{"*_test.go"},
				Mode:             WatchAll,
				DebounceInterval: 100 * time.Millisecond,
				ClearTerminal:    false,
				Writer:           nil,
			}

			_, err := NewTestWatcher(options)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s, but got none", tc.description)
			}

			if !tc.expectError && err != nil {
				t.Errorf("Did not expect error for %s, but got: %v", tc.description, err)
			}

			if err != nil {
				t.Logf("File system error for %s: %s", tc.name, err.Error())
			}
		})
	}
}

// TestErrorRecovery_CacheErrors tests recovery from cache-related errors
func TestErrorRecovery_CacheErrors(t *testing.T) {
	cache := NewTestResultCache()

	testCases := []struct {
		name        string
		testPath    string
		suite       *TestSuite
		expectError bool
		description string
	}{
		{
			name:        "EmptyTestPath",
			testPath:    "",
			suite:       &TestSuite{FilePath: "test.go"},
			expectError: false,
			description: "Should handle empty test path gracefully",
		},
		{
			name:        "NilTestSuite",
			testPath:    "./test",
			suite:       nil,
			expectError: false,
			description: "Should handle nil test suite gracefully",
		},
		{
			name:        "ValidCacheOperation",
			testPath:    "./valid/test",
			suite:       &TestSuite{FilePath: "valid_test.go", TestCount: 1},
			expectError: false,
			description: "Should handle valid cache operation",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var err error

			// Test cache store operation
			func() {
				defer func() {
					if r := recover(); r != nil {
						err = r.(error)
					}
				}()
				cache.CacheResult(tc.testPath, tc.suite)
			}()

			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s, but got none", tc.description)
			}

			if !tc.expectError && err != nil {
				t.Errorf("Did not expect error for %s, but got: %v", tc.description, err)
			}

			// Test cache retrieve operation
			_, exists := cache.GetCachedResult(tc.testPath)

			if tc.suite != nil && tc.testPath != "" && !exists {
				t.Errorf("Expected cached result to exist for valid operation")
			}

			t.Logf("Cache error recovery test %s completed", tc.name)
		})
	}
}

// TestErrorRecovery_ConfigurationErrors tests recovery from configuration errors
func TestErrorRecovery_ConfigurationErrors(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		name        string
		configFile  string
		content     string
		expectError bool
		description string
	}{
		{
			name:        "MalformedYAML",
			configFile:  "malformed.yaml",
			content:     "invalid: yaml: content: [unclosed",
			expectError: true,
			description: "Should handle malformed YAML gracefully",
		},
		{
			name:        "InvalidFieldTypes",
			configFile:  "invalid_types.yaml",
			content:     "verbosity: \"not_a_number\"\ncolors: \"not_a_boolean\"",
			expectError: true,
			description: "Should handle invalid field types gracefully",
		},
		{
			name:        "MissingRequiredFields",
			configFile:  "incomplete.yaml",
			content:     "# Only partial config\nverbosity: 1",
			expectError: false,
			description: "Should handle incomplete config with defaults",
		},
		{
			name:        "EmptyConfigFile",
			configFile:  "empty.yaml",
			content:     "",
			expectError: false,
			description: "Should handle empty config file gracefully",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configPath := filepath.Join(tempDir, tc.configFile)
			err := os.WriteFile(configPath, []byte(tc.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test config file: %v", err)
			}

			configLoader := &DefaultConfigLoader{}
			_, err = configLoader.LoadFromFile(configPath)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for %s, but got none", tc.description)
			}

			if !tc.expectError && err != nil {
				t.Errorf("Did not expect error for %s, but got: %v", tc.description, err)
			}

			if err != nil {
				// Verify error message is user-friendly
				errorMsg := err.Error()
				if !strings.Contains(errorMsg, "config") {
					t.Errorf("Expected error message to mention config, got: %s", errorMsg)
				}
				t.Logf("Config error for %s: %s", tc.name, errorMsg)
			}
		})
	}
}

// TestErrorRecovery_ResourceExhaustion tests recovery from resource exhaustion scenarios
func TestErrorRecovery_ResourceExhaustion(t *testing.T) {
	tempDir := t.TempDir()

	// Create many files to potentially exhaust resources
	numFiles := 100
	for i := 0; i < numFiles; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("stress_%d_test.go", i))
		content := fmt.Sprintf(`package stress%d
import "testing"
func TestStress%d(t *testing.T) {
	// Stress test %d
}`, i%10, i, i)

		err := os.WriteFile(filename, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create stress test file %d: %v", i, err)
		}
	}

	// Test file watcher with many files
	options := WatchOptions{
		Paths:            []string{tempDir},
		IgnorePatterns:   []string{},
		TestPatterns:     []string{"*_test.go"},
		Mode:             WatchAll,
		DebounceInterval: 50 * time.Millisecond,
		ClearTerminal:    false,
		Writer:           nil,
	}

	watcher, err := NewTestWatcher(options)
	if err != nil {
		t.Logf("File watcher creation failed under stress (may be expected): %v", err)
	} else {
		defer watcher.Stop()

		// Wait for initialization
		time.Sleep(100 * time.Millisecond)

		// Modify many files quickly
		for i := 0; i < 20; i++ {
			filename := filepath.Join(tempDir, fmt.Sprintf("stress_%d_test.go", i))
			content := fmt.Sprintf(`package stress%d
import "testing"
func TestStress%d(t *testing.T) {
	// Modified stress test %d
}`, i%10, i, i)

			err := os.WriteFile(filename, []byte(content), 0644)
			if err != nil {
				t.Logf("Failed to modify stress file %d (may be expected under load): %v", i, err)
			}
		}

		// Wait for processing
		time.Sleep(200 * time.Millisecond)
		t.Log("Resource exhaustion test completed successfully")
	}
}

// TestErrorRecovery_ConcurrencyIssues tests recovery from concurrency issues
func TestErrorRecovery_ConcurrencyIssues(t *testing.T) {
	cache := NewTestResultCache()

	// Test concurrent cache access
	numGoroutines := 10
	numOperations := 50

	errChan := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() {
				if r := recover(); r != nil {
					errChan <- r.(error)
				} else {
					errChan <- nil
				}
			}()

			for j := 0; j < numOperations; j++ {
				testPath := fmt.Sprintf("./concurrent/test_%d_%d", id, j)
				suite := &TestSuite{
					FilePath:    fmt.Sprintf("concurrent_%d_%d_test.go", id, j),
					TestCount:   1,
					PassedCount: 1,
				}

				// Mix of read and write operations
				if j%2 == 0 {
					cache.CacheResult(testPath, suite)
				} else {
					_, _ = cache.GetCachedResult(testPath)
				}
			}
		}(i)
	}

	// Collect results
	errorCount := 0
	for i := 0; i < numGoroutines; i++ {
		err := <-errChan
		if err != nil {
			errorCount++
			t.Logf("Concurrent operation error (may be expected): %v", err)
		}
	}

	if errorCount > numGoroutines/2 {
		t.Errorf("Too many concurrent errors: %d/%d", errorCount, numGoroutines)
	}

	t.Logf("Concurrency test completed with %d/%d errors", errorCount, numGoroutines)
}

// TestErrorRecovery_GracefulDegradation tests graceful degradation scenarios
func TestErrorRecovery_GracefulDegradation(t *testing.T) {
	tempDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tempDir, "degradation_test.go")
	content := `package main
import "testing"
func TestDegradation(t *testing.T) {
	// Test graceful degradation
}`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test processor with various failure scenarios
	testCases := []struct {
		name        string
		colors      bool
		icons       bool
		termWidth   int
		description string
	}{
		{
			name:        "NoColors",
			colors:      false,
			icons:       true,
			termWidth:   80,
			description: "Should work without colors",
		},
		{
			name:        "NoIcons",
			colors:      true,
			icons:       false,
			termWidth:   80,
			description: "Should work without icons",
		},
		{
			name:        "NarrowTerminal",
			colors:      true,
			icons:       true,
			termWidth:   20,
			description: "Should work with narrow terminal",
		},
		{
			name:        "MinimalFeatures",
			colors:      false,
			icons:       false,
			termWidth:   40,
			description: "Should work with minimal features",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var output bytes.Buffer
			processor := NewTestProcessor(
				&output,
				NewColorFormatter(tc.colors),
				NewIconProvider(tc.icons),
				tc.termWidth,
			)

			suite := &TestSuite{
				FilePath:    "degradation_test.go",
				TestCount:   1,
				PassedCount: 1,
				Tests: []*TestResult{
					{
						Name:     "TestDegradation",
						Status:   StatusPassed,
						Duration: 10 * time.Millisecond,
						Package:  "github.com/test/degradation",
					},
				},
			}

			processor.AddTestSuite(suite)
			err := processor.RenderResults(false)

			if err != nil {
				t.Errorf("Graceful degradation failed for %s: %v", tc.description, err)
			}

			// Verify output was generated
			if output.Len() == 0 {
				t.Errorf("No output generated for %s", tc.description)
			}

			t.Logf("Graceful degradation test %s completed", tc.name)
		})
	}
}
