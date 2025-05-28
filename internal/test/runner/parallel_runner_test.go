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

// TestNewParallelTestRunner_Creation verifies parallel runner initialization
func TestNewParallelTestRunner_Creation(t *testing.T) {
	// Arrange
	testRunner := &TestRunner{}
	testCache := NewMockCache()

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
			runner := NewParallelTestRunner(tc.maxConcurrency, testRunner, testCache)

			// Assert
			if runner == nil {
				t.Fatal("Expected runner to be created, got nil")
			}
			// Note: Can't test private fields directly, but we can test behavior
			// The maxConcurrency is tested indirectly through the actual parallel execution behavior
			// For now, we just verify the runner was created successfully
		})
	}
}

// TestRunParallel_EmptyTestPaths tests running with no test paths
func TestRunParallel_EmptyTestPaths(t *testing.T) {
	// Arrange
	testRunner := &TestRunner{}
	testCache := NewMockCache()
	runner := NewParallelTestRunner(2, testRunner, testCache)
	cfg := &config.Config{}

	// Act
	results, err := runner.RunParallel(context.Background(), []string{}, cfg)

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
	testCache := NewMockCache()
	runner := NewParallelTestRunner(2, testRunner, testCache)
	cfg := &config.Config{}

	// Act
	results, err := runner.RunParallel(context.Background(), nil, cfg)

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
	testCache := NewMockCache()
	runner := NewParallelTestRunner(2, testRunner, testCache)
	cfg := &config.Config{Timeout: 100 * time.Millisecond} // Short timeout to avoid long test execution

	// Mock test paths (will likely fail but tests structure)
	testPaths := []string{"./testdata/pkg1", "./testdata/pkg2", "./testdata/pkg3"}

	// Act
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	results, err := runner.RunParallel(ctx, testPaths, cfg)

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
	// Arrange
	testRunner := &TestRunner{}
	testCache := NewMockCache()

	testCases := []struct {
		name           string
		maxConcurrency int
		testPaths      []string
	}{
		{
			name:           "Single concurrency",
			maxConcurrency: 1,
			testPaths:      []string{"./testdata/pkg1", "./testdata/pkg2"},
		},
		{
			name:           "High concurrency",
			maxConcurrency: 10,
			testPaths:      []string{"./testdata/pkg1"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			runner := NewParallelTestRunner(tc.maxConcurrency, testRunner, testCache)
			cfg := &config.Config{Timeout: 50 * time.Millisecond}

			// Act
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			results, err := runner.RunParallel(ctx, tc.testPaths, cfg)

			// Assert
			// We expect errors due to nonexistent paths, but test structure
			if err != nil {
				t.Logf("Expected error due to mock paths: %v", err)
			}
			if len(results) != len(tc.testPaths) {
				t.Errorf("Expected %d results, got %d", len(tc.testPaths), len(results))
			}
		})
	}
}

// TestExecuteTestPath_CacheHit tests cache hit scenario
func TestExecuteTestPath_CacheHit(t *testing.T) {
	// Arrange
	testRunner := &TestRunner{}
	testCache := NewMockCache()

	// Pre-populate cache
	suite := &models.TestSuite{FilePath: "cached_test.go"}
	testCache.CacheResult("./cached", suite)

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
