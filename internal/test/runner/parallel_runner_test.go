package runner

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/config"
	"github.com/newbpydev/go-sentinel/internal/test/processor"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// MockCache implements CacheInterface for testing
type MockCache struct {
	cache map[string]*CachedResult
}

func NewMockCache() *MockCache {
	return &MockCache{
		cache: make(map[string]*CachedResult),
	}
}

func (m *MockCache) GetCachedResult(testPath string) (*CachedResult, bool) {
	result, exists := m.cache[testPath]
	return result, exists
}

func (m *MockCache) CacheResult(testPath string, suite *models.TestSuite) {
	m.cache[testPath] = &CachedResult{Suite: suite}
}

// MockCacheWithHit implements CacheInterface for testing cache hits
type MockCacheWithHit struct {
	suite *models.TestSuite
}

func (m *MockCacheWithHit) GetCachedResult(testPath string) (*CachedResult, bool) {
	return &CachedResult{Suite: m.suite}, true
}

func (m *MockCacheWithHit) CacheResult(testPath string, suite *models.TestSuite) {
	// Mock implementation - do nothing
}

// TestNewParallelTestRunner_Creation verifies parallel runner initialization
func TestNewParallelTestRunner_Creation(t *testing.T) {
	t.Parallel()

	// Arrange
	testRunner := &TestRunner{}
	testCache := NewMockCache()

	testCases := []struct {
		name                string
		maxConcurrency      int
		expectedConcurrency int
	}{
		{
			name:                "positive_concurrency",
			maxConcurrency:      4,
			expectedConcurrency: 4,
		},
		{
			name:                "zero_concurrency",
			maxConcurrency:      0,
			expectedConcurrency: 4, // Default
		},
		{
			name:                "negative_concurrency",
			maxConcurrency:      -1,
			expectedConcurrency: 4, // Default
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runner := NewParallelTestRunner(tc.maxConcurrency, testRunner, testCache)
			if runner == nil {
				t.Fatal("NewParallelTestRunner should not return nil")
			}

			if runner.maxConcurrency != tc.expectedConcurrency {
				t.Errorf("Expected concurrency %d, got %d", tc.expectedConcurrency, runner.maxConcurrency)
			}
		})
	}
}

// TestRunParallel_EmptyTestPaths tests running with no test paths
func TestRunParallel_EmptyTestPaths(t *testing.T) {
	t.Parallel()

	// Arrange
	testRunner := &TestRunner{}
	testCache := NewMockCache()
	runner := NewParallelTestRunner(2, testRunner, testCache)
	cfg := &config.Config{}

	// Act
	results, err := runner.RunParallel(context.Background(), []string{}, cfg)

	// Assert
	if err != nil {
		t.Errorf("Expected no error for empty test paths, got: %v", err)
	}
	if results != nil {
		t.Errorf("Expected nil results for empty test paths, got: %v", results)
	}
}

// TestRunParallel_NilTestPaths tests running with nil test paths
func TestRunParallel_NilTestPaths(t *testing.T) {
	t.Parallel()

	// Arrange
	testRunner := &TestRunner{}
	testCache := NewMockCache()
	runner := NewParallelTestRunner(2, testRunner, testCache)
	cfg := &config.Config{}

	// Act
	results, err := runner.RunParallel(context.Background(), nil, cfg)

	// Assert
	if err != nil {
		t.Errorf("Expected no error for nil test paths, got: %v", err)
	}
	if results != nil {
		t.Errorf("Expected nil results for nil test paths, got: %v", results)
	}
}

// TestRunParallel_ConcurrencyControl tests that concurrency is properly controlled
func TestRunParallel_ConcurrencyControl(t *testing.T) {
	t.Parallel()

	// Arrange
	testRunner := &TestRunner{}
	testCache := NewMockCache()
	runner := NewParallelTestRunner(2, testRunner, testCache)
	cfg := &config.Config{Timeout: 100 * time.Millisecond} // Short timeout to avoid long test execution

	// Create test paths that will likely fail quickly
	testPaths := []string{"./non-existent-1", "./non-existent-2", "./non-existent-3", "./non-existent-4"}

	// Act
	start := time.Now()
	results, err := runner.RunParallel(context.Background(), testPaths, cfg)
	elapsed := time.Since(start)

	// Assert
	// Should complete relatively quickly due to concurrency
	if elapsed > 5*time.Second {
		t.Errorf("Expected parallel execution to complete quickly, took %v", elapsed)
	}

	// Should return results for all paths (even if they failed)
	if len(results) != len(testPaths) {
		t.Errorf("Expected %d results, got %d", len(testPaths), len(results))
	}

	// Error is acceptable since paths don't exist
	_ = err
}

// TestParallelTestResult_StructFields tests ParallelTestResult struct field access
func TestParallelTestResult_StructFields(t *testing.T) {
	// Arrange
	suite := &models.TestSuite{FilePath: "test.go"}
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
		t.Errorf("Expected no error, got: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected to write %d bytes, got %d", len(testData), n)
	}
}

// TestDiscardWriter_MultipleWrites tests multiple writes to discard writer
func TestDiscardWriter_MultipleWrites(t *testing.T) {
	// Arrange
	writer := &discardWriter{}
	writes := [][]byte{
		[]byte("first write"),
		[]byte("second write"),
		[]byte("third write"),
	}

	// Act & Assert
	for i, data := range writes {
		n, err := writer.Write(data)
		if err != nil {
			t.Errorf("Write %d: expected no error, got: %v", i, err)
		}
		if n != len(data) {
			t.Errorf("Write %d: expected to write %d bytes, got %d", i, len(data), n)
		}
	}
}

// TestMergeResults_EmptyResults tests merging empty results
func TestMergeResults_EmptyResults(t *testing.T) {
	// Arrange
	var results []*ParallelTestResult
	testProcessor := processor.NewTestProcessor(
		&discardWriter{},
		&nullColorFormatter{},
		&nullIconProvider{},
		80,
	)

	// Act
	MergeResults(testProcessor, results)

	// Assert
	if len(testProcessor.GetSuites()) != 0 {
		t.Errorf("Expected 0 suites, got %d", len(testProcessor.GetSuites()))
	}
	if testProcessor.GetStats().TotalTests != 0 {
		t.Errorf("Expected 0 total tests, got %d", testProcessor.GetStats().TotalTests)
	}
}

// TestMergeResults_ValidResults tests merging valid results
func TestMergeResults_ValidResults(t *testing.T) {
	// Arrange
	suite1 := &models.TestSuite{
		FilePath:    "pkg1/test.go",
		TestCount:   3,
		PassedCount: 2,
		FailedCount: 1,
		Duration:    100 * time.Millisecond,
	}

	suite2 := &models.TestSuite{
		FilePath:    "pkg2/test.go",
		TestCount:   2,
		PassedCount: 2,
		FailedCount: 0,
		Duration:    50 * time.Millisecond,
	}

	results := []*ParallelTestResult{
		{
			TestPath:  "pkg1",
			Suite:     suite1,
			Duration:  100 * time.Millisecond,
			FromCache: false,
		},
		{
			TestPath:  "pkg2",
			Suite:     suite2,
			Duration:  50 * time.Millisecond,
			FromCache: true,
		},
	}

	testProcessor := processor.NewTestProcessor(
		&discardWriter{},
		&nullColorFormatter{},
		&nullIconProvider{},
		80,
	)

	// Act
	MergeResults(testProcessor, results)

	// Assert
	suites := testProcessor.GetSuites()
	if len(suites) != 2 {
		t.Errorf("Expected 2 suites, got %d", len(suites))
	}

	stats := testProcessor.GetStats()
	if stats.TotalTests != 5 {
		t.Errorf("Expected 5 total tests, got %d", stats.TotalTests)
	}
	if stats.PassedTests != 4 {
		t.Errorf("Expected 4 passed tests, got %d", stats.PassedTests)
	}
	if stats.FailedTests != 1 {
		t.Errorf("Expected 1 failed test, got %d", stats.FailedTests)
	}

	// Check that both suites are present
	if _, exists := suites["pkg1/test.go"]; !exists {
		t.Error("Expected suite 'pkg1/test.go' to be present")
	}
	if _, exists := suites["pkg2/test.go"]; !exists {
		t.Error("Expected suite 'pkg2/test.go' to be present")
	}
}

// TestMergeResults_WithErrors tests merging results that include errors
func TestMergeResults_WithErrors(t *testing.T) {
	// Arrange
	suite1 := &models.TestSuite{
		FilePath:    "pkg1/test.go",
		TestCount:   2,
		PassedCount: 2,
		FailedCount: 0,
	}

	results := []*ParallelTestResult{
		{
			TestPath:  "pkg1",
			Suite:     suite1,
			Duration:  100 * time.Millisecond,
			FromCache: false,
		},
		{
			TestPath:  "pkg2",
			Suite:     nil,
			Error:     fmt.Errorf("compilation failed"),
			Duration:  10 * time.Millisecond,
			FromCache: false,
		},
		{
			TestPath:  "pkg3",
			Suite:     nil,
			Error:     fmt.Errorf("timeout"),
			Duration:  5 * time.Second,
			FromCache: false,
		},
	}

	testProcessor := processor.NewTestProcessor(
		&discardWriter{},
		&nullColorFormatter{},
		&nullIconProvider{},
		80,
	)

	// Act
	MergeResults(testProcessor, results)

	// Assert
	suites := testProcessor.GetSuites()
	if len(suites) != 1 {
		t.Errorf("Expected 1 suite (only successful ones), got %d", len(suites))
	}

	stats := testProcessor.GetStats()
	if stats.TotalTests != 2 {
		t.Errorf("Expected 2 total tests (from successful suite), got %d", stats.TotalTests)
	}
	if stats.PassedTests != 2 {
		t.Errorf("Expected 2 passed tests, got %d", stats.PassedTests)
	}
	if stats.FailedTests != 0 {
		t.Errorf("Expected 0 failed tests (errors don't count as failed tests), got %d", stats.FailedTests)
	}

	// Check that only the successful suite is present
	if _, exists := suites["pkg1/test.go"]; !exists {
		t.Error("Expected suite 'pkg1/test.go' to be present")
	}
}

// TestMergeResults_NilSuites tests merging results with nil suites
func TestMergeResults_NilSuites(t *testing.T) {
	// Arrange
	results := []*ParallelTestResult{
		{
			TestPath:  "pkg1",
			Suite:     nil,
			Error:     fmt.Errorf("failed to run"),
			Duration:  10 * time.Millisecond,
			FromCache: false,
		},
		{
			TestPath:  "pkg2",
			Suite:     nil,
			Error:     fmt.Errorf("another error"),
			Duration:  20 * time.Millisecond,
			FromCache: false,
		},
	}

	testProcessor := processor.NewTestProcessor(
		&discardWriter{},
		&nullColorFormatter{},
		&nullIconProvider{},
		80,
	)

	// Act
	MergeResults(testProcessor, results)

	// Assert
	suites := testProcessor.GetSuites()
	if len(suites) != 0 {
		t.Errorf("Expected 0 suites (all failed), got %d", len(suites))
	}

	stats := testProcessor.GetStats()
	if stats.TotalTests != 0 {
		t.Errorf("Expected 0 total tests, got %d", stats.TotalTests)
	}
}

// TestParallelTestRunner_ConcurrencyLimits tests concurrency limits
func TestParallelTestRunner_ConcurrencyLimits(t *testing.T) {
	t.Parallel()

	// Arrange
	testRunner := &TestRunner{}
	testCache := NewMockCache()

	testCases := []struct {
		name           string
		maxConcurrency int
		expectedMax    int
	}{
		{
			name:           "normal_concurrency",
			maxConcurrency: 4,
			expectedMax:    4,
		},
		{
			name:           "zero_concurrency",
			maxConcurrency: 0,
			expectedMax:    4, // Should default to 4
		},
		{
			name:           "negative_concurrency",
			maxConcurrency: -1,
			expectedMax:    4, // Should default to 4
		},
		{
			name:           "high_concurrency",
			maxConcurrency: 100,
			expectedMax:    100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runner := NewParallelTestRunner(tc.maxConcurrency, testRunner, testCache)
			if runner.maxConcurrency != tc.expectedMax {
				t.Errorf("Expected max concurrency %d, got %d", tc.expectedMax, runner.maxConcurrency)
			}
		})
	}
}

// TestExecuteTestPath_CacheHit tests cache hit scenario
func TestExecuteTestPath_CacheHit(t *testing.T) {
	t.Parallel()

	// Arrange
	testRunner := &TestRunner{}
	suite := &models.TestSuite{
		FilePath:     "./cached",
		Tests:        []*models.LegacyTestResult{},
		Duration:     time.Second,
		TestCount:    1,
		PassedCount:  1,
		FailedCount:  0,
		SkippedCount: 0,
	}
	testCache := &MockCacheWithHit{suite: suite}
	runner := NewParallelTestRunner(2, testRunner, testCache)
	cfg := &config.Config{}

	// Act
	result := runner.executeTestPath(context.Background(), "./cached", cfg)

	// Assert
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if !result.FromCache {
		t.Error("Expected result to be from cache")
	}
	if result.Suite != suite {
		t.Error("Expected cached suite to be returned")
	}
	if result.Error != nil {
		t.Errorf("Expected no error for cache hit, got: %v", result.Error)
	}
}

// TestParallelRunner_ProgressDrainingSafety tests that progress draining doesn't cause infinite loops
func TestParallelRunner_ProgressDrainingSafety(t *testing.T) {
	t.Parallel()

	// This test verifies that the progress draining logic has proper safety limits
	// and doesn't cause infinite loops when context is cancelled

	testRunner := &TestRunner{}
	testCache := NewMockCache()
	runner := NewParallelTestRunner(1, testRunner, testCache)

	// Create a context that will be cancelled quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Use non-existent paths to trigger quick failures
	testPaths := []string{"./non-existent-path-1", "./non-existent-path-2"}
	cfg := &config.Config{Timeout: 10 * time.Millisecond}

	start := time.Now()

	// This should complete quickly without hanging
	results, err := runner.RunParallel(ctx, testPaths, cfg)

	elapsed := time.Since(start)

	// Should complete within reasonable time (not hang indefinitely)
	if elapsed > 5*time.Second {
		t.Errorf("RunParallel took too long (%v), possible infinite loop", elapsed)
	}

	// Should handle cancellation gracefully
	if err != nil && !strings.Contains(err.Error(), "cancel") {
		t.Logf("Expected cancellation error, got: %v", err)
	}

	// Results should be returned even on cancellation
	if len(results) > len(testPaths) {
		t.Errorf("Got more results (%d) than test paths (%d)", len(results), len(testPaths))
	}

	t.Logf("Test completed in %v with %d results", elapsed, len(results))
}

// TestNullColorFormatter_AllMethods tests all methods of nullColorFormatter
func TestNullColorFormatter_AllMethods(t *testing.T) {
	t.Parallel()

	formatter := &nullColorFormatter{}
	testText := "test text"

	testCases := []struct {
		name     string
		method   func(string) string
		expected string
	}{
		{"Red", formatter.Red, testText},
		{"Green", formatter.Green, testText},
		{"Yellow", formatter.Yellow, testText},
		{"Blue", formatter.Blue, testText},
		{"Magenta", formatter.Magenta, testText},
		{"Cyan", formatter.Cyan, testText},
		{"Gray", formatter.Gray, testText},
		{"Bold", formatter.Bold, testText},
		{"Dim", formatter.Dim, testText},
		{"White", formatter.White, testText},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.method(testText)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}

	// Test Colorize method separately
	t.Run("Colorize", func(t *testing.T) {
		t.Parallel()

		result := formatter.Colorize(testText, "red")
		if result != testText {
			t.Errorf("Expected %s, got %s", testText, result)
		}
	})
}

// TestNullIconProvider_AllMethods tests all methods of nullIconProvider
func TestNullIconProvider_AllMethods(t *testing.T) {
	t.Parallel()

	provider := &nullIconProvider{}

	testCases := []struct {
		name     string
		method   func() string
		expected string
	}{
		{"CheckMark", provider.CheckMark, "✓"},
		{"Cross", provider.Cross, "✗"},
		{"Skipped", provider.Skipped, "-"},
		{"Running", provider.Running, "..."},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := tc.method()
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}

	// Test GetIcon method separately
	t.Run("GetIcon", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			iconType string
			expected string
		}{
			{"check", "•"},
			{"cross", "•"},
			{"unknown", "•"},
			{"", "•"},
		}

		for _, tc := range testCases {
			result := provider.GetIcon(tc.iconType)
			if result != tc.expected {
				t.Errorf("GetIcon(%s): expected %s, got %s", tc.iconType, tc.expected, result)
			}
		}
	})
}

// TestDiscardWriter_Interface tests that discardWriter implements io.Writer
func TestDiscardWriter_Interface(t *testing.T) {
	t.Parallel()

	writer := &discardWriter{}

	// Verify interface compliance
	var _ io.Writer = writer

	// Test with various data sizes
	testCases := [][]byte{
		[]byte(""),
		[]byte("small"),
		[]byte("medium length test data"),
		make([]byte, 1024), // Large buffer
		make([]byte, 0),    // Empty slice
	}

	for _, data := range testCases {
		t.Run(fmt.Sprintf("Write_%d_bytes", len(data)), func(t *testing.T) {
			t.Parallel()

			n, err := writer.Write(data)
			if err != nil {
				t.Errorf("Write failed: %v", err)
			}
			if n != len(data) {
				t.Errorf("Expected to write %d bytes, got %d", len(data), n)
			}
		})
	}
}

// TestNullColorFormatter_ComprehensiveCoverage tests all color formatter methods
func TestNullColorFormatter_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	formatter := &nullColorFormatter{}

	tests := map[string]struct {
		method   func(string) string
		input    string
		expected string
	}{
		"Red": {
			method:   formatter.Red,
			input:    "test",
			expected: "test",
		},
		"Green": {
			method:   formatter.Green,
			input:    "success",
			expected: "success",
		},
		"Yellow": {
			method:   formatter.Yellow,
			input:    "warning",
			expected: "warning",
		},
		"Blue": {
			method:   formatter.Blue,
			input:    "info",
			expected: "info",
		},
		"Magenta": {
			method:   formatter.Magenta,
			input:    "debug",
			expected: "debug",
		},
		"Cyan": {
			method:   formatter.Cyan,
			input:    "cyan",
			expected: "cyan",
		},
		"Gray": {
			method:   formatter.Gray,
			input:    "gray",
			expected: "gray",
		},
		"Bold": {
			method:   formatter.Bold,
			input:    "bold",
			expected: "bold",
		},
		"Dim": {
			method:   formatter.Dim,
			input:    "dim",
			expected: "dim",
		},
		"White": {
			method:   formatter.White,
			input:    "white",
			expected: "white",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := tt.method(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}

	// Test Colorize method
	t.Run("Colorize", func(t *testing.T) {
		t.Parallel()

		result := formatter.Colorize("test", "red")
		if result != "test" {
			t.Errorf("Expected 'test', got %q", result)
		}
	})
}

// TestNullIconProvider_ComprehensiveCoverage tests all icon provider methods
func TestNullIconProvider_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	provider := &nullIconProvider{}

	tests := map[string]struct {
		method   func() string
		expected string
	}{
		"CheckMark": {
			method:   provider.CheckMark,
			expected: "✓",
		},
		"Cross": {
			method:   provider.Cross,
			expected: "✗",
		},
		"Skipped": {
			method:   provider.Skipped,
			expected: "-",
		},
		"Running": {
			method:   provider.Running,
			expected: "...",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result := tt.method()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}

	// Test GetIcon method
	t.Run("GetIcon", func(t *testing.T) {
		t.Parallel()

		result := provider.GetIcon("any")
		if result != "•" {
			t.Errorf("Expected '•', got %q", result)
		}
	})
}

// TestMergeResults_ComprehensiveCoverage tests the MergeResults function thoroughly
func TestMergeResults_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	t.Run("nil_results", func(t *testing.T) {
		t.Parallel()

		testProcessor := processor.NewTestProcessor(&discardWriter{}, &nullColorFormatter{}, &nullIconProvider{}, 80)
		MergeResults(testProcessor, nil)

		// Verify processor state
		suites := testProcessor.GetSuites()
		if len(suites) != 0 {
			t.Errorf("Expected 0 suites, got %d", len(suites))
		}
	})

	t.Run("empty_results", func(t *testing.T) {
		t.Parallel()

		testProcessor := processor.NewTestProcessor(&discardWriter{}, &nullColorFormatter{}, &nullIconProvider{}, 80)
		MergeResults(testProcessor, []*ParallelTestResult{})

		// Verify processor state
		suites := testProcessor.GetSuites()
		if len(suites) != 0 {
			t.Errorf("Expected 0 suites, got %d", len(suites))
		}
	})

	t.Run("single_result", func(t *testing.T) {
		t.Parallel()

		suite := &models.TestSuite{
			FilePath:     "test.go",
			TestCount:    5,
			PassedCount:  3,
			FailedCount:  1,
			SkippedCount: 1,
		}

		results := []*ParallelTestResult{
			{
				TestPath: "pkg1",
				Suite:    suite,
				Duration: 100 * time.Millisecond,
			},
		}

		testProcessor := processor.NewTestProcessor(&discardWriter{}, &nullColorFormatter{}, &nullIconProvider{}, 80)
		MergeResults(testProcessor, results)

		// Verify processor state
		suites := testProcessor.GetSuites()
		if len(suites) != 1 {
			t.Errorf("Expected 1 suite, got %d", len(suites))
		}
	})

	t.Run("multiple_results", func(t *testing.T) {
		t.Parallel()

		suite1 := &models.TestSuite{
			FilePath:     "test1.go",
			TestCount:    3,
			PassedCount:  2,
			FailedCount:  1,
			SkippedCount: 0,
		}

		suite2 := &models.TestSuite{
			FilePath:     "test2.go",
			TestCount:    4,
			PassedCount:  3,
			FailedCount:  0,
			SkippedCount: 1,
		}

		results := []*ParallelTestResult{
			{
				TestPath: "pkg1",
				Suite:    suite1,
				Duration: 100 * time.Millisecond,
			},
			{
				TestPath: "pkg2",
				Suite:    suite2,
				Duration: 150 * time.Millisecond,
			},
		}

		testProcessor := processor.NewTestProcessor(&discardWriter{}, &nullColorFormatter{}, &nullIconProvider{}, 80)
		MergeResults(testProcessor, results)

		// Verify processor state
		suites := testProcessor.GetSuites()
		if len(suites) != 2 {
			t.Errorf("Expected 2 suites, got %d", len(suites))
		}
	})

	t.Run("results_with_errors", func(t *testing.T) {
		t.Parallel()

		suite := &models.TestSuite{
			FilePath:    "test.go",
			TestCount:   2,
			PassedCount: 2,
		}

		results := []*ParallelTestResult{
			{
				TestPath: "pkg1",
				Suite:    suite,
				Duration: 100 * time.Millisecond,
			},
			{
				TestPath: "pkg2",
				Error:    fmt.Errorf("test failed"),
				Duration: 50 * time.Millisecond,
			},
		}

		testProcessor := processor.NewTestProcessor(&discardWriter{}, &nullColorFormatter{}, &nullIconProvider{}, 80)
		MergeResults(testProcessor, results)

		// Verify processor state - only successful results should be added
		suites := testProcessor.GetSuites()
		if len(suites) != 1 {
			t.Errorf("Expected 1 suite (error result should be skipped), got %d", len(suites))
		}
	})

	t.Run("results_with_nil_suites", func(t *testing.T) {
		t.Parallel()

		results := []*ParallelTestResult{
			{
				TestPath: "pkg1",
				Suite:    nil,
				Duration: 100 * time.Millisecond,
			},
			{
				TestPath: "pkg2",
				Suite:    nil,
				Error:    fmt.Errorf("failed"),
				Duration: 50 * time.Millisecond,
			},
		}

		testProcessor := processor.NewTestProcessor(&discardWriter{}, &nullColorFormatter{}, &nullIconProvider{}, 80)
		MergeResults(testProcessor, results)

		// Verify processor state - nil suites should not be added
		suites := testProcessor.GetSuites()
		if len(suites) != 0 {
			t.Errorf("Expected 0 suites for nil suites, got %d", len(suites))
		}
	})
}

// TestNewParallelTestRunner_ComprehensiveCoverage tests the factory function thoroughly
func TestNewParallelTestRunner_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name                string
		maxConcurrency      int
		expectedConcurrency int
	}{
		{
			name:                "positive_concurrency",
			maxConcurrency:      4,
			expectedConcurrency: 4,
		},
		{
			name:                "zero_concurrency",
			maxConcurrency:      0,
			expectedConcurrency: 4, // Default
		},
		{
			name:                "negative_concurrency",
			maxConcurrency:      -1,
			expectedConcurrency: 4, // Default
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockTestRunner := NewBasicTestRunner(false, true)
			mockCache := NewMockCache()

			runner := NewParallelTestRunner(tc.maxConcurrency, mockTestRunner, mockCache)
			if runner == nil {
				t.Fatal("NewParallelTestRunner should not return nil")
			}

			if runner.maxConcurrency != tc.expectedConcurrency {
				t.Errorf("Expected concurrency %d, got %d", tc.expectedConcurrency, runner.maxConcurrency)
			}
		})
	}
}

// TestDiscardWriter_ComprehensiveCoverage tests the discardWriter.Write method
func TestDiscardWriter_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	writer := &discardWriter{}

	t.Run("write_empty_data", func(t *testing.T) {
		t.Parallel()

		n, err := writer.Write([]byte{})
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if n != 0 {
			t.Errorf("Expected 0 bytes written, got %d", n)
		}
	})

	t.Run("write_small_data", func(t *testing.T) {
		t.Parallel()

		data := []byte("test")
		n, err := writer.Write(data)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if n != len(data) {
			t.Errorf("Expected %d bytes written, got %d", len(data), n)
		}
	})

	t.Run("write_large_data", func(t *testing.T) {
		t.Parallel()

		data := make([]byte, 1024)
		for i := range data {
			data[i] = byte(i % 256)
		}
		n, err := writer.Write(data)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if n != len(data) {
			t.Errorf("Expected %d bytes written, got %d", len(data), n)
		}
	})

	t.Run("write_nil_data", func(t *testing.T) {
		t.Parallel()

		n, err := writer.Write(nil)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if n != 0 {
			t.Errorf("Expected 0 bytes written, got %d", n)
		}
	})
}

// TestParallelTestRunner_RunParallel_ErrorCases tests RunParallel with various error conditions
func TestParallelTestRunner_RunParallel_ErrorCases(t *testing.T) {
	t.Parallel()

	// Create mock dependencies
	mockTestRunner := NewBasicTestRunner(false, true)
	mockCache := NewMockCache()

	testCases := []struct {
		name        string
		concurrency int
		paths       []string
		timeout     time.Duration
	}{
		{
			name:        "invalid_paths",
			concurrency: 2,
			paths:       []string{"non/existent/path"},
			timeout:     2 * time.Second,
		},
		{
			name:        "empty_paths",
			concurrency: 2,
			paths:       []string{},
			timeout:     2 * time.Second,
		},
		{
			name:        "single_invalid_path",
			concurrency: 1,
			paths:       []string{"./non_existent_dir"},
			timeout:     2 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			runner := NewParallelTestRunner(tc.concurrency, mockTestRunner, mockCache)
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			// Create a mock config
			cfg := &config.Config{
				Timeout: tc.timeout,
			}

			result, err := runner.RunParallel(ctx, tc.paths, cfg)
			// We expect errors for these cases, but the function should not panic
			_ = result
			_ = err
		})
	}
}

// TestParallelTestRunner_ExecuteTestPath_ComprehensiveCoverage tests executeTestPath method
func TestParallelTestRunner_ExecuteTestPath_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	// Create mock dependencies
	mockTestRunner := NewBasicTestRunner(false, true)
	mockCache := NewMockCache()
	runner := NewParallelTestRunner(2, mockTestRunner, mockCache)

	testCases := []struct {
		name    string
		path    string
		timeout time.Duration
	}{
		{
			name:    "current_directory",
			path:    ".",
			timeout: 5 * time.Second,
		},
		{
			name:    "non_existent_path",
			path:    "./non_existent",
			timeout: 2 * time.Second,
		},
		{
			name:    "empty_path",
			path:    "",
			timeout: 2 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			// Create a mock config
			cfg := &config.Config{
				Timeout: tc.timeout,
			}

			result := runner.executeTestPath(ctx, tc.path, cfg)
			if result == nil {
				t.Error("Expected result to be returned, got nil")
			}
		})
	}
}

// TestParallelTestRunner_RunSingleTestPath_ComprehensiveCoverage tests runSingleTestPath method
func TestParallelTestRunner_RunSingleTestPath_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	// Create mock dependencies
	mockTestRunner := NewBasicTestRunner(false, true)
	mockCache := NewMockCache()
	runner := NewParallelTestRunner(1, mockTestRunner, mockCache)

	testCases := []struct {
		name    string
		path    string
		timeout time.Duration
	}{
		{
			name:    "current_directory",
			path:    ".",
			timeout: 5 * time.Second,
		},
		{
			name:    "non_existent_path",
			path:    "./non_existent",
			timeout: 2 * time.Second,
		},
		{
			name:    "empty_path",
			path:    "",
			timeout: 2 * time.Second,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			defer cancel()

			// Create a mock config
			cfg := &config.Config{
				Timeout: tc.timeout,
			}

			suite, err := runner.runSingleTestPath(ctx, tc.path, cfg)
			// We expect errors for some cases, but the function should not panic
			_ = suite
			_ = err
		})
	}
}

// TestDiscardWriter_Write_ComprehensiveCoverage tests discardWriter.Write method
func TestDiscardWriter_Write_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	writer := &discardWriter{}

	testCases := []struct {
		name string
		data []byte
	}{
		{
			name: "normal_data",
			data: []byte("test data"),
		},
		{
			name: "empty_data",
			data: []byte{},
		},
		{
			name: "nil_data",
			data: nil,
		},
		{
			name: "large_data",
			data: make([]byte, 10000),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			n, err := writer.Write(tc.data)
			if err != nil {
				t.Errorf("Write should not error: %v", err)
			}
			if n != len(tc.data) {
				t.Errorf("Expected to write %d bytes, got %d", len(tc.data), n)
			}
		})
	}
}
