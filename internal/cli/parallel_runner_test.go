package cli

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"
)

// TestNewParallelTestRunner_Creation verifies parallel runner initialization
func TestNewParallelTestRunner_Creation(t *testing.T) {
	// Arrange
	testRunner := &TestRunner{}
	cache := NewTestResultCache()

	testCases := []struct {
		name                string
		maxConcurrency      int
		expectedConcurrency int
	}{
		{
			name:                "Valid concurrency",
			maxConcurrency:      8,
			expectedConcurrency: 8,
		},
		{
			name:                "Zero concurrency uses default",
			maxConcurrency:      0,
			expectedConcurrency: 4,
		},
		{
			name:                "Negative concurrency uses default",
			maxConcurrency:      -5,
			expectedConcurrency: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			runner := NewParallelTestRunner(tc.maxConcurrency, testRunner, cache)

			// Assert
			if runner == nil {
				t.Fatal("Expected runner to be created, got nil")
			}
			if runner.maxConcurrency != tc.expectedConcurrency {
				t.Errorf("Expected maxConcurrency %d, got %d", tc.expectedConcurrency, runner.maxConcurrency)
			}
			if runner.testRunner != testRunner {
				t.Error("Expected testRunner to be set correctly")
			}
			if runner.cache != cache {
				t.Error("Expected cache to be set correctly")
			}
		})
	}
}

// TestRunParallel_EmptyTestPaths tests running with no test paths
func TestRunParallel_EmptyTestPaths(t *testing.T) {
	// Arrange
	testRunner := &TestRunner{}
	cache := NewTestResultCache()
	runner := NewParallelTestRunner(2, testRunner, cache)
	config := &Config{}

	// Act
	results, err := runner.RunParallel(context.Background(), []string{}, config)

	// Assert
	if err != nil {
		t.Errorf("Expected no error for empty paths, got: %v", err)
	}
	if results != nil {
		t.Error("Expected results to be nil for empty paths")
	}
}

// TestRunParallel_NilTestPaths tests running with nil test paths
func TestRunParallel_NilTestPaths(t *testing.T) {
	// Arrange
	testRunner := &TestRunner{}
	cache := NewTestResultCache()
	runner := NewParallelTestRunner(2, testRunner, cache)
	config := &Config{}

	// Act
	results, err := runner.RunParallel(context.Background(), nil, config)

	// Assert
	if err != nil {
		t.Errorf("Expected no error for nil paths, got: %v", err)
	}
	if results != nil {
		t.Error("Expected results to be nil for nil paths")
	}
}

// TestRunParallel_ConcurrencyControl tests that concurrency is properly controlled
func TestRunParallel_ConcurrencyControl(t *testing.T) {
	// This test is more complex since we can't easily test actual concurrency
	// without a mock TestRunner, so we test the structure and basic functionality

	// Arrange
	testRunner := &TestRunner{}
	cache := NewTestResultCache()
	runner := NewParallelTestRunner(2, testRunner, cache)
	config := &Config{Timeout: 100 * time.Millisecond} // Short timeout to avoid long test execution

	// Mock test paths (will likely fail but tests structure)
	testPaths := []string{"./testdata/pkg1", "./testdata/pkg2", "./testdata/pkg3"}

	// Act
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	results, err := runner.RunParallel(ctx, testPaths, config)

	// Assert
	// We expect this to likely fail due to nonexistent paths, but test structure
	if err != nil {
		t.Logf("Expected error due to mock paths: %v", err)
	}
	if len(results) != len(testPaths) {
		t.Errorf("Expected %d results, got %d", len(testPaths), len(results))
	}

	// Verify that all test paths are represented in results
	pathsFound := make(map[string]bool)
	for _, result := range results {
		pathsFound[result.TestPath] = true
	}

	for _, path := range testPaths {
		if !pathsFound[path] {
			t.Errorf("Expected result for path '%s', but not found", path)
		}
	}
}

// TestParallelTestResult_StructFields tests ParallelTestResult struct field access
func TestParallelTestResult_StructFields(t *testing.T) {
	// Arrange
	suite := &TestSuite{FilePath: "test.go"}
	duration := 100 * time.Millisecond

	result := &ParallelTestResult{
		TestPath:  "pkg/test",
		Suite:     suite,
		Error:     nil,
		Duration:  duration,
		FromCache: true,
	}

	// Assert
	if result.TestPath != "pkg/test" {
		t.Errorf("Expected TestPath 'pkg/test', got '%s'", result.TestPath)
	}
	if result.Suite != suite {
		t.Error("Expected Suite to be set correctly")
	}
	if result.Error != nil {
		t.Errorf("Expected Error to be nil, got %v", result.Error)
	}
	if result.Duration != duration {
		t.Errorf("Expected Duration %v, got %v", duration, result.Duration)
	}
	if !result.FromCache {
		t.Error("Expected FromCache to be true")
	}
}

// TestParallelTestResult_WithError tests ParallelTestResult with error
func TestParallelTestResult_WithError(t *testing.T) {
	// Arrange
	testErr := fmt.Errorf("test execution failed")

	result := &ParallelTestResult{
		TestPath:  "pkg/test",
		Suite:     nil,
		Error:     testErr,
		Duration:  50 * time.Millisecond,
		FromCache: false,
	}

	// Assert
	if result.Error == nil {
		t.Error("Expected Error to be set")
	}
	if result.Error.Error() != "test execution failed" {
		t.Errorf("Expected error message 'test execution failed', got '%s'", result.Error.Error())
	}
	if result.Suite != nil {
		t.Error("Expected Suite to be nil when there's an error")
	}
	if result.FromCache {
		t.Error("Expected FromCache to be false for failed execution")
	}
}

// TestDiscardWriter_WritesAndDiscards tests the discard writer utility
func TestDiscardWriter_WritesAndDiscards(t *testing.T) {
	// Arrange
	writer := &discardWriter{}
	testData := []byte("test data to discard")

	// Act
	n, err := writer.Write(testData)

	// Assert
	if err != nil {
		t.Errorf("Expected no error from discardWriter, got: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected bytes written %d, got %d", len(testData), n)
	}
}

// TestDiscardWriter_MultipleWrites tests multiple writes to discard writer
func TestDiscardWriter_MultipleWrites(t *testing.T) {
	// Arrange
	writer := &discardWriter{}
	writes := [][]byte{
		[]byte("first write"),
		[]byte("second write"),
		[]byte(""),
		[]byte("final write"),
	}

	// Act & Assert
	for i, data := range writes {
		n, err := writer.Write(data)
		if err != nil {
			t.Errorf("Write %d: expected no error, got: %v", i, err)
		}
		if n != len(data) {
			t.Errorf("Write %d: expected bytes written %d, got %d", i, len(data), n)
		}
	}
}

// TestMergeResults_EmptyResults tests merging empty results
func TestMergeResults_EmptyResults(t *testing.T) {
	// Arrange
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	processor := NewTestProcessor(nil, formatter, icons, 80)
	var results []*ParallelTestResult

	// Act
	MergeResults(processor, results)

	// Assert
	// Should not panic and processor should remain unchanged
	stats := processor.GetStats()
	if stats.TotalTests != 0 {
		t.Error("Expected no tests after merging empty results")
	}
}

// TestMergeResults_ValidResults tests merging valid parallel test results
func TestMergeResults_ValidResults(t *testing.T) {
	// Arrange
	processor := NewTestProcessor(
		&bytes.Buffer{},
		NewColorFormatter(false),
		NewIconProvider(false),
		80,
	)

	results := []*ParallelTestResult{
		{
			TestPath: "pkg1",
			Suite: &TestSuite{
				FilePath:     "pkg1",
				TestCount:    3,
				PassedCount:  2,
				FailedCount:  1,
				SkippedCount: 0,
				Tests: []*TestResult{
					{Name: "Test1", Status: StatusPassed},
					{Name: "Test2", Status: StatusPassed},
					{Name: "Test3", Status: StatusFailed},
				},
			},
			Error: nil,
		},
		{
			TestPath: "pkg2",
			Suite: &TestSuite{
				FilePath:     "pkg2",
				TestCount:    6,
				PassedCount:  5,
				FailedCount:  0,
				SkippedCount: 1,
				Tests: []*TestResult{
					{Name: "TestA", Status: StatusPassed},
					{Name: "TestB", Status: StatusPassed},
					{Name: "TestC", Status: StatusPassed},
					{Name: "TestD", Status: StatusPassed},
					{Name: "TestE", Status: StatusPassed},
					{Name: "TestF", Status: StatusSkipped},
				},
			},
			Error: nil,
		},
	}

	// Act
	MergeResults(processor, results)

	// Assert
	stats := processor.GetStats()
	if stats.TotalTests != 9 {
		t.Errorf("Expected 9 total tests after merge, got %d", stats.TotalTests)
	}
	if stats.PassedTests != 7 {
		t.Errorf("Expected 7 passed tests, got %d", stats.PassedTests)
	}
	if stats.FailedTests != 1 {
		t.Errorf("Expected 1 failed test, got %d", stats.FailedTests)
	}
	if stats.SkippedTests != 1 {
		t.Errorf("Expected 1 skipped test, got %d", stats.SkippedTests)
	}
}

// TestMergeResults_WithErrors tests merging results that include errors
func TestMergeResults_WithErrors(t *testing.T) {
	// Arrange
	processor := NewTestProcessor(
		&bytes.Buffer{},
		NewColorFormatter(false),
		NewIconProvider(false),
		80,
	)

	results := []*ParallelTestResult{
		{
			TestPath: "pkg1",
			Suite: &TestSuite{
				FilePath:     "pkg1",
				TestCount:    2,
				PassedCount:  2,
				FailedCount:  0,
				SkippedCount: 0,
				Tests: []*TestResult{
					{Name: "Test1", Status: StatusPassed},
					{Name: "Test2", Status: StatusPassed},
				},
			},
			Error: nil,
		},
		{
			TestPath: "pkg2",
			Suite:    nil,
			Error:    fmt.Errorf("build failed"),
		},
		{
			TestPath: "pkg3",
			Suite: &TestSuite{
				FilePath:     "pkg3",
				TestCount:    3,
				PassedCount:  3,
				FailedCount:  0,
				SkippedCount: 0,
				Tests: []*TestResult{
					{Name: "TestA", Status: StatusPassed},
					{Name: "TestB", Status: StatusPassed},
					{Name: "TestC", Status: StatusPassed},
				},
			},
			Error: nil,
		},
	}

	// Act
	MergeResults(processor, results)

	// Assert - only successful results should be merged
	stats := processor.GetStats()
	if stats.TotalTests != 5 {
		t.Errorf("Expected 5 total tests (only successful), got %d", stats.TotalTests)
	}
	if stats.PassedTests != 5 {
		t.Errorf("Expected 5 passed tests, got %d", stats.PassedTests)
	}
	if stats.FailedTests != 0 {
		t.Errorf("Expected 0 failed tests, got %d", stats.FailedTests)
	}
}

// TestMergeResults_NilSuites tests merging results with nil suites
func TestMergeResults_NilSuites(t *testing.T) {
	// Arrange
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	processor := NewTestProcessor(nil, formatter, icons, 80)

	results := []*ParallelTestResult{
		{
			TestPath: "pkg1/test",
			Suite:    nil, // Nil suite but no error
			Error:    nil,
		},
		{
			TestPath: "pkg2/test",
			Suite:    nil,
			Error:    nil,
		},
	}

	// Act
	MergeResults(processor, results)

	// Assert
	stats := processor.GetStats()
	// No tests should be merged since all suites are nil
	if stats.TotalTests != 0 {
		t.Errorf("Expected 0 total tests with nil suites, got %d", stats.TotalTests)
	}
}

// TestParallelTestRunner_ConcurrencyLimits tests concurrency boundary conditions
func TestParallelTestRunner_ConcurrencyLimits(t *testing.T) {
	testCases := []struct {
		name           string
		maxConcurrency int
		testPathCount  int
		expectedMax    int
	}{
		{
			name:           "More paths than concurrency",
			maxConcurrency: 2,
			testPathCount:  5,
			expectedMax:    2,
		},
		{
			name:           "Fewer paths than concurrency",
			maxConcurrency: 10,
			testPathCount:  3,
			expectedMax:    10,
		},
		{
			name:           "Equal paths and concurrency",
			maxConcurrency: 4,
			testPathCount:  4,
			expectedMax:    4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			testRunner := &TestRunner{}
			cache := NewTestResultCache()
			runner := NewParallelTestRunner(tc.maxConcurrency, testRunner, cache)

			// Assert
			if runner.maxConcurrency != tc.expectedMax {
				t.Errorf("Expected maxConcurrency %d, got %d", tc.expectedMax, runner.maxConcurrency)
			}
		})
	}
}

// TestExecuteTestPath_CacheHit tests cache hit behavior
func TestExecuteTestPath_CacheHit(t *testing.T) {
	// Arrange
	testRunner := &TestRunner{}
	cache := NewTestResultCache()
	runner := NewParallelTestRunner(2, testRunner, cache)
	config := &Config{}

	// Pre-populate cache
	testPath := "pkg/test"
	cachedSuite := &TestSuite{
		FilePath:    testPath,
		PassedCount: 10,
	}
	cache.CacheResult(testPath, cachedSuite)

	// Act
	result := runner.executeTestPath(context.Background(), testPath, config)

	// Assert
	if result == nil {
		t.Fatal("Expected result to be returned")
	}
	if result.TestPath != testPath {
		t.Errorf("Expected TestPath '%s', got '%s'", testPath, result.TestPath)
	}
	if !result.FromCache {
		t.Error("Expected result to be from cache")
	}
	if result.Suite != cachedSuite {
		t.Error("Expected cached suite to be returned")
	}
	if result.Error != nil {
		t.Errorf("Expected no error for cache hit, got: %v", result.Error)
	}
}
