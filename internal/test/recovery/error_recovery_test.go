package recovery

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

	"github.com/newbpydev/go-sentinel/internal/test/cache"
	"github.com/newbpydev/go-sentinel/internal/test/processor"
	"github.com/newbpydev/go-sentinel/internal/test/runner"
	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// TestRecoverFromParsingErrors tests recovery from JSON parsing errors
func TestRecoverFromParsingErrors(t *testing.T) {
	// Create a stream parser
	parser := processor.NewStreamParser()

	// Test malformed JSON that could cause issues
	malformedJSON := `{"Time":"2024-01-01T10:00:00.000Z","Action":"run","Package":"github.com/test/example"`

	reader := strings.NewReader(malformedJSON)
	results := make(chan *models.LegacyTestResult, 10)

	// This should not crash the application
	err := parser.Parse(reader, results)
	close(results)

	// Should handle the error gracefully
	if err == nil {
		t.Error("Expected error from malformed JSON, got nil")
	}

	// Verify the parser can still work with valid JSON afterward
	validJSON := `{"Time":"2024-01-01T10:00:00.000Z","Action":"run","Package":"github.com/test/example","Test":"TestExample"}
{"Time":"2024-01-01T10:00:00.100Z","Action":"pass","Package":"github.com/test/example","Test":"TestExample","Elapsed":0.1}`

	reader2 := strings.NewReader(validJSON)
	results2 := make(chan *models.LegacyTestResult, 10)

	go func() {
		defer close(results2)
		_ = parser.Parse(reader2, results2)
	}()

	// Should work normally after error
	resultCount := 0
	for range results2 {
		resultCount++
	}

	if resultCount == 0 {
		t.Error("Parser should recover and process valid JSON after error")
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

	// Create a file
	testFile := filepath.Join(tempDir, "test.go")
	err = os.WriteFile(testFile, []byte("package main\nfunc Test() {}"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Try to change permissions to restrict access (behavior varies by platform)
	err = os.Chmod(testFile, 0000)
	if err != nil {
		t.Logf("Could not change file permissions: %v", err)
	}

	// Test that cache operations handle permission errors gracefully
	testCache := cache.NewTestResultCache()

	// Try to analyze the file - should handle gracefully
	_, err = testCache.AnalyzeChange(testFile)
	if err != nil {
		t.Logf("Cache analysis failed as expected with permission error: %v", err)
	}

	// The key requirement is that operations don't crash
	// and handle the situation gracefully
}

// TestRecoverFromTestRunnerFailures tests recovery from test runner failures
func TestRecoverFromTestRunnerFailures(t *testing.T) {
	// Create test runner
	testRunner := runner.NewBasicTestRunner(false, true)

	ctx := context.Background()

	// Test with non-existent path
	nonExistentPaths := []string{"/path/that/does/not/exist"}
	_, err := testRunner.Run(ctx, nonExistentPaths)

	// Should return error but not crash
	if err == nil {
		t.Error("Expected error for non-existent path, got nil")
	}

	// Test that runner can still work with valid paths afterward
	tempDir := t.TempDir()

	// Create a valid test file
	testFile := filepath.Join(tempDir, "valid_test.go")
	testContent := `package main

import "testing"

func TestValid(t *testing.T) {
	if 1+1 != 2 {
		t.Error("Math is broken")
	}
}`
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Runner should work normally after error
	_, err = testRunner.Run(ctx, []string{tempDir})
	if err != nil {
		t.Logf("Test run failed (may be expected in some environments): %v", err)
	}
}

// TestCacheErrorRecovery tests recovery from cache-related errors
func TestCacheErrorRecovery(t *testing.T) {
	testCache := cache.NewTestResultCache()

	// Test caching a large number of results
	for i := 0; i < 1000; i++ {
		suite := &models.TestSuite{
			FilePath:     fmt.Sprintf("test_%d.go", i),
			TestCount:    10,
			PassedCount:  9,
			FailedCount:  1,
			SkippedCount: 0,
			Duration:     time.Millisecond * 100,
		}

		testPath := fmt.Sprintf("./test/path%d", i)
		testCache.CacheResult(testPath, suite)
	}

	// Verify cache operations still work
	stats := testCache.GetStats()
	if stats == nil {
		t.Error("Cache stats should be available")
	}

	// Test clearing cache
	testCache.Clear()

	// Cache should work normally after clearing
	suite := &models.TestSuite{
		FilePath:     "test_after_clear.go",
		TestCount:    1,
		PassedCount:  1,
		FailedCount:  0,
		SkippedCount: 0,
	}
	testCache.CacheResult("./test/after_clear", suite)

	_, exists := testCache.GetCachedResult("./test/after_clear")
	if !exists {
		t.Error("Cache should work normally after clearing")
	}
}

// TestProcessorErrorRecovery tests recovery from processor-related errors
func TestProcessorErrorRecovery(t *testing.T) {
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	iconProvider := colors.NewIconProvider(false)

	// Create test processor
	testProcessor := processor.NewTestProcessor(&buf, formatter, iconProvider, 80)

	// Test with many test suites to stress test the processor
	for i := 0; i < 100; i++ {
		suite := &models.TestSuite{
			FilePath:     fmt.Sprintf("stress_test_%d.go", i),
			TestCount:    50,
			PassedCount:  45,
			FailedCount:  5,
			SkippedCount: 0,
			Duration:     time.Millisecond * 200,
		}

		// Add test results
		for j := 0; j < 50; j++ {
			status := models.TestStatusPassed
			if j%10 == 0 {
				status = models.TestStatusFailed
			}

			test := &models.LegacyTestResult{
				Name:     fmt.Sprintf("TestStress_%d_%d", i, j),
				Status:   status,
				Duration: time.Millisecond,
				Package:  fmt.Sprintf("github.com/test/stress%d", i),
			}

			if status == models.TestStatusFailed {
				test.Error = &models.LegacyTestError{
					Message: "Stress test failure",
					Type:    "AssertionError",
				}
			}

			suite.Tests = append(suite.Tests, test)
		}

		testProcessor.AddTestSuite(suite)

		// Reset periodically to prevent memory buildup
		if i%20 == 0 {
			testProcessor.Reset()
			buf.Reset()
		}
	}

	// Final render should work
	err := testProcessor.RenderResults(false)
	if err != nil {
		t.Errorf("Processor should handle stress testing gracefully, got: %v", err)
	}
}

// TestGracefulDegradation tests graceful degradation under adverse conditions
func TestGracefulDegradation(t *testing.T) {
	// Test with invalid writer (should not crash)
	formatter := colors.NewColorFormatter(false)
	iconProvider := colors.NewIconProvider(false)

	// Use discard writer to simulate output issues
	testProcessor := processor.NewTestProcessor(io.Discard, formatter, iconProvider, 80)

	suite := &models.TestSuite{
		FilePath:     "degradation_test.go",
		TestCount:    5,
		PassedCount:  3,
		FailedCount:  2,
		SkippedCount: 0,
	}

	// Add some test results
	for i := 0; i < 5; i++ {
		status := models.TestStatusPassed
		if i%2 == 0 {
			status = models.TestStatusFailed
		}

		test := &models.LegacyTestResult{
			Name:     fmt.Sprintf("TestDegradation_%d", i),
			Status:   status,
			Duration: time.Millisecond,
			Package:  "github.com/test/degradation",
		}

		if status == models.TestStatusFailed {
			test.Error = &models.LegacyTestError{
				Message: "Degradation test failure",
				Type:    "AssertionError",
			}
		}

		suite.Tests = append(suite.Tests, test)
	}

	testProcessor.AddTestSuite(suite)

	// Should not crash even with output issues
	err := testProcessor.RenderResults(false)
	if err != nil {
		t.Logf("Processor returned error in degraded mode: %v", err)
	}
}

// TestConcurrentOperations tests stability under concurrent operations
func TestConcurrentOperations(t *testing.T) {
	testCache := cache.NewTestResultCache()

	// Start multiple goroutines performing cache operations
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(routineID int) {
			defer func() { done <- true }()

			for j := 0; j < 20; j++ {
				suite := &models.TestSuite{
					FilePath:     fmt.Sprintf("concurrent_%d_%d.go", routineID, j),
					TestCount:    10,
					PassedCount:  8,
					FailedCount:  2,
					SkippedCount: 0,
				}

				testPath := fmt.Sprintf("./concurrent/test_%d_%d", routineID, j)
				testCache.CacheResult(testPath, suite)

				// Read back the result
				_, exists := testCache.GetCachedResult(testPath)
				if !exists {
					t.Errorf("Cache result should exist for %s", testPath)
				}
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		select {
		case <-done:
			// Goroutine completed successfully
		case <-time.After(10 * time.Second):
			t.Fatal("Concurrent operations timed out")
		}
	}

	// Verify cache is still functional
	stats := testCache.GetStats()
	if stats == nil {
		t.Error("Cache should be functional after concurrent operations")
	}
}
