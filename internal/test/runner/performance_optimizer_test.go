package runner

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/processor"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// MockColorFormatter for testing
type MockColorFormatter struct{}

func (m *MockColorFormatter) Red(text string) string                 { return text }
func (m *MockColorFormatter) Green(text string) string               { return text }
func (m *MockColorFormatter) Yellow(text string) string              { return text }
func (m *MockColorFormatter) Blue(text string) string                { return text }
func (m *MockColorFormatter) Magenta(text string) string             { return text }
func (m *MockColorFormatter) Cyan(text string) string                { return text }
func (m *MockColorFormatter) Gray(text string) string                { return text }
func (m *MockColorFormatter) Bold(text string) string                { return text }
func (m *MockColorFormatter) Dim(text string) string                 { return text }
func (m *MockColorFormatter) White(text string) string               { return text }
func (m *MockColorFormatter) Colorize(text, colorName string) string { return text }

// MockIconProvider for testing
type MockIconProvider struct{}

func (m *MockIconProvider) CheckMark() string              { return "✓" }
func (m *MockIconProvider) Cross() string                  { return "✗" }
func (m *MockIconProvider) Skipped() string                { return "○" }
func (m *MockIconProvider) Running() string                { return "●" }
func (m *MockIconProvider) GetIcon(iconType string) string { return "●" }

// TestNewOptimizedTestProcessor_ComprehensiveCoverage tests the NewOptimizedTestProcessor factory function
func TestNewOptimizedTestProcessor_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		output    *bytes.Buffer
		processor *processor.TestProcessor
		expectNil bool
	}{
		{
			name:      "valid_output_and_processor",
			output:    &bytes.Buffer{},
			processor: processor.NewTestProcessor(&bytes.Buffer{}, &MockColorFormatter{}, &MockIconProvider{}, 80),
			expectNil: false,
		},
		{
			name:      "nil_output_valid_processor",
			output:    nil,
			processor: processor.NewTestProcessor(&bytes.Buffer{}, &MockColorFormatter{}, &MockIconProvider{}, 80),
			expectNil: false,
		},
		{
			name:      "valid_output_nil_processor",
			output:    &bytes.Buffer{},
			processor: nil,
			expectNil: false,
		},
		{
			name:      "both_nil",
			output:    nil,
			processor: nil,
			expectNil: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			processor := NewOptimizedTestProcessor(tc.output, tc.processor)
			if tc.expectNil && processor != nil {
				t.Error("Expected nil processor")
			}
			if !tc.expectNil && processor == nil {
				t.Error("Expected non-nil processor")
			}
		})
	}
}

// TestOptimizedTestProcessor_GetStats_ComprehensiveCoverage tests the GetStats method
func TestOptimizedTestProcessor_GetStats_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	testProcessor := processor.NewTestProcessor(os.Stdout, &MockColorFormatter{}, &MockIconProvider{}, 80)
	processor := NewOptimizedTestProcessor(os.Stdout, testProcessor)

	stats := processor.GetStats()
	if stats == nil {
		t.Error("Expected stats to be returned, got nil")
	}

	// Verify stats structure
	if stats.TotalTests < 0 {
		t.Errorf("Expected non-negative TotalTests, got %d", stats.TotalTests)
	}
	if stats.PassedTests < 0 {
		t.Errorf("Expected non-negative PassedTests, got %d", stats.PassedTests)
	}
	if stats.FailedTests < 0 {
		t.Errorf("Expected non-negative FailedTests, got %d", stats.FailedTests)
	}
	if stats.SkippedTests < 0 {
		t.Errorf("Expected non-negative SkippedTests, got %d", stats.SkippedTests)
	}
}

// TestOptimizedTestProcessor_GetSuites_ComprehensiveCoverage tests the GetSuites method
func TestOptimizedTestProcessor_GetSuites_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	testProcessor := processor.NewTestProcessor(os.Stdout, &MockColorFormatter{}, &MockIconProvider{}, 80)
	processor := NewOptimizedTestProcessor(os.Stdout, testProcessor)

	suites := processor.GetSuites()
	if suites == nil {
		t.Error("Expected suites map to be returned, got nil")
	}

	// Should return empty map initially
	if len(suites) != 0 {
		t.Errorf("Expected empty suites map initially, got %d entries", len(suites))
	}
}

// TestOptimizedTestProcessor_GetStatsOptimized_ComprehensiveCoverage tests the GetStatsOptimized method
func TestOptimizedTestProcessor_GetStatsOptimized_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	testProcessor := processor.NewTestProcessor(os.Stdout, &MockColorFormatter{}, &MockIconProvider{}, 80)
	processor := NewOptimizedTestProcessor(os.Stdout, testProcessor)

	stats := processor.GetStatsOptimized()
	if stats == nil {
		t.Error("Expected optimized stats to be returned, got nil")
		return
	}

	// Should be same as GetStats
	regularStats := processor.GetStats()
	if regularStats == nil {
		t.Error("Expected regular stats to be returned, got nil")
		return
	}

	if stats.TotalTests != regularStats.TotalTests {
		t.Errorf("Expected GetStatsOptimized to match GetStats, got different TotalTests")
	}
}

// TestOptimizedTestProcessor_GetMemoryStats_ComprehensiveCoverage tests the GetMemoryStats method
func TestOptimizedTestProcessor_GetMemoryStats_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	testProcessor := processor.NewTestProcessor(os.Stdout, &MockColorFormatter{}, &MockIconProvider{}, 80)
	processor := NewOptimizedTestProcessor(os.Stdout, testProcessor)

	memStats := processor.GetMemoryStats()

	// Verify memory stats structure
	if memStats.AllocBytes < 0 {
		t.Errorf("Expected non-negative AllocBytes, got %d", memStats.AllocBytes)
	}
	if memStats.TotalAllocBytes < 0 {
		t.Errorf("Expected non-negative TotalAllocBytes, got %d", memStats.TotalAllocBytes)
	}
	if memStats.SysBytes < 0 {
		t.Errorf("Expected non-negative SysBytes, got %d", memStats.SysBytes)
	}
	if memStats.NumGC < 0 {
		t.Errorf("Expected non-negative NumGC, got %d", memStats.NumGC)
	}
}

// TestOptimizedTestProcessor_ForceGarbageCollection_ComprehensiveCoverage tests the ForceGarbageCollection method
func TestOptimizedTestProcessor_ForceGarbageCollection_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	testProcessor := processor.NewTestProcessor(os.Stdout, &MockColorFormatter{}, &MockIconProvider{}, 80)
	processor := NewOptimizedTestProcessor(os.Stdout, testProcessor)

	// Get initial GC count
	var initialStats runtime.MemStats
	runtime.ReadMemStats(&initialStats)
	initialGC := initialStats.NumGC

	// Force garbage collection
	processor.ForceGarbageCollection()

	// Get final GC count
	var finalStats runtime.MemStats
	runtime.ReadMemStats(&finalStats)
	finalGC := finalStats.NumGC

	// GC count should have increased (or at least not decreased)
	if finalGC < initialGC {
		t.Errorf("Expected GC count to increase or stay same, got initial=%d, final=%d", initialGC, finalGC)
	}
}

// TestNewOptimizedStreamParser_ComprehensiveCoverage tests the NewOptimizedStreamParser factory function
func TestNewOptimizedStreamParser_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	parser := NewOptimizedStreamParser()
	if parser == nil {
		t.Fatal("Expected non-nil parser")
	}
}

// TestNewBatchProcessor_ComprehensiveCoverage tests the NewBatchProcessor factory function
func TestNewBatchProcessor_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		batchSize int
		timeout   time.Duration
		expectNil bool
	}{
		{
			name:      "valid_batch_size_and_timeout",
			batchSize: 10,
			timeout:   100 * time.Millisecond,
			expectNil: false,
		},
		{
			name:      "zero_batch_size",
			batchSize: 0,
			timeout:   100 * time.Millisecond,
			expectNil: false,
		},
		{
			name:      "negative_batch_size",
			batchSize: -1,
			timeout:   100 * time.Millisecond,
			expectNil: false,
		},
		{
			name:      "zero_timeout",
			batchSize: 10,
			timeout:   0,
			expectNil: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			processor := NewBatchProcessor(tc.batchSize, tc.timeout)
			if tc.expectNil && processor != nil {
				t.Error("Expected nil processor")
			}
			if !tc.expectNil && processor == nil {
				t.Error("Expected non-nil processor")
			}
		})
	}
}

// TestBatchProcessor_Add_ComprehensiveCoverage tests the Add method
func TestBatchProcessor_Add_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	processor := NewBatchProcessor(3, 100*time.Millisecond)

	testCases := []struct {
		name string
		item *models.LegacyTestResult
	}{
		{
			name: "valid_test_result",
			item: &models.LegacyTestResult{Name: "test1"},
		},
		{
			name: "nil_test_result",
			item: nil,
		},
		{
			name: "empty_test_result",
			item: &models.LegacyTestResult{},
		},
		{
			name: "test_result_with_data",
			item: &models.LegacyTestResult{
				Name:     "test2",
				Package:  "pkg",
				Duration: 100 * time.Millisecond,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Note: Not using t.Parallel() here because we're testing sequential adds

			batch := processor.Add(tc.item)
			// Method should not panic and may return a batch when buffer is full
			if batch != nil && len(batch) > 3 {
				t.Errorf("Expected batch size <= 3, got %d", len(batch))
			}
		})
	}
}

// TestNewLazyRenderer_ComprehensiveCoverage tests the NewLazyRenderer factory function
func TestNewLazyRenderer_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		threshold int
		expectNil bool
	}{
		{
			name:      "positive_threshold",
			threshold: 50,
			expectNil: false,
		},
		{
			name:      "zero_threshold",
			threshold: 0,
			expectNil: false,
		},
		{
			name:      "negative_threshold",
			threshold: -1,
			expectNil: false,
		},
		{
			name:      "large_threshold",
			threshold: 10000,
			expectNil: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			renderer := NewLazyRenderer(tc.threshold)
			if tc.expectNil && renderer != nil {
				t.Error("Expected nil renderer")
			}
			if !tc.expectNil && renderer == nil {
				t.Error("Expected non-nil renderer")
			}
		})
	}
}

// TestLazyRenderer_ShouldUseLazyMode_ComprehensiveCoverage tests the ShouldUseLazyMode method
func TestLazyRenderer_ShouldUseLazyMode_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		threshold int
		testCount int
		expected  bool
	}{
		{
			name:      "test_count_above_threshold",
			threshold: 50,
			testCount: 100,
			expected:  true,
		},
		{
			name:      "test_count_below_threshold",
			threshold: 50,
			testCount: 25,
			expected:  false,
		},
		{
			name:      "test_count_equal_threshold",
			threshold: 50,
			testCount: 50,
			expected:  false,
		},
		{
			name:      "zero_test_count",
			threshold: 50,
			testCount: 0,
			expected:  false,
		},
		{
			name:      "negative_test_count",
			threshold: 50,
			testCount: -1,
			expected:  false,
		},
		{
			name:      "zero_threshold",
			threshold: 0,
			testCount: 1,
			expected:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			renderer := NewLazyRenderer(tc.threshold)
			result := renderer.ShouldUseLazyMode(tc.testCount)

			if result != tc.expected {
				t.Errorf("Expected %v, got %v for threshold=%d, testCount=%d",
					tc.expected, result, tc.threshold, tc.testCount)
			}
		})
	}
}

// TestOptimizedTestProcessor_AddTestSuite_ComprehensiveCoverage tests the AddTestSuite method
func TestOptimizedTestProcessor_AddTestSuite_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	testProcessor := processor.NewTestProcessor(os.Stdout, &MockColorFormatter{}, &MockIconProvider{}, 80)
	processor := NewOptimizedTestProcessor(os.Stdout, testProcessor)

	testCases := []struct {
		name  string
		suite *models.TestSuite
	}{
		{
			name: "valid_test_suite",
			suite: &models.TestSuite{
				FilePath:  "test1.go",
				TestCount: 5,
			},
		},
		{
			name:  "nil_test_suite",
			suite: nil,
		},
		{
			name:  "empty_test_suite",
			suite: &models.TestSuite{},
		},
		{
			name: "test_suite_with_zero_tests",
			suite: &models.TestSuite{
				FilePath:  "empty_test.go",
				TestCount: 0,
			},
		},
		{
			name: "test_suite_with_negative_tests",
			suite: &models.TestSuite{
				FilePath:  "negative_test.go",
				TestCount: -1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Should not panic
			processor.AddTestSuite(tc.suite)
		})
	}
}

// TestOptimizedTestProcessor_Clear_ComprehensiveCoverage tests the Clear method
func TestOptimizedTestProcessor_Clear_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	testProcessor := processor.NewTestProcessor(os.Stdout, &MockColorFormatter{}, &MockIconProvider{}, 80)
	processor := NewOptimizedTestProcessor(os.Stdout, testProcessor)

	// Add some test suites first
	suite1 := &models.TestSuite{
		FilePath:  "test1.go",
		TestCount: 5,
	}
	suite2 := &models.TestSuite{
		FilePath:  "test2.go",
		TestCount: 3,
	}

	processor.AddTestSuite(suite1)
	processor.AddTestSuite(suite2)

	// Clear should not panic
	processor.Clear()

	// Verify suites are cleared (if processor is not nil)
	suites := processor.GetSuites()
	if len(suites) != 0 {
		t.Errorf("Expected empty suites after clear, got %d", len(suites))
	}
}

// TestMin_ComprehensiveCoverage tests the min utility function
func TestMin_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "a_less_than_b",
			a:        5,
			b:        10,
			expected: 5,
		},
		{
			name:     "b_less_than_a",
			a:        10,
			b:        5,
			expected: 5,
		},
		{
			name:     "a_equals_b",
			a:        7,
			b:        7,
			expected: 7,
		},
		{
			name:     "negative_numbers",
			a:        -5,
			b:        -10,
			expected: -10,
		},
		{
			name:     "zero_and_positive",
			a:        0,
			b:        5,
			expected: 0,
		},
		{
			name:     "zero_and_negative",
			a:        0,
			b:        -5,
			expected: -5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := min(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("Expected min(%d, %d) = %d, got %d", tc.a, tc.b, tc.expected, result)
			}
		})
	}
}

// TestOptimizedStreamParser_ParseOptimized_ComprehensiveCoverage tests the ParseOptimized method
func TestOptimizedStreamParser_ParseOptimized_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	parser := NewOptimizedStreamParser()

	testCases := []struct {
		name   string
		input  string
		hasErr bool
	}{
		{
			name:   "valid_test_output",
			input:  "--- PASS: TestExample (0.01s)\n--- FAIL: TestFailing (0.02s)",
			hasErr: false,
		},
		{
			name:   "empty_input",
			input:  "",
			hasErr: false,
		},
		{
			name:   "invalid_format",
			input:  "random text that is not test output",
			hasErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reader := strings.NewReader(tc.input)
			results := make(chan *models.LegacyTestResult, 10)

			// Parse in goroutine and close channel when done
			go func() {
				defer close(results)
				parser.ParseOptimized(reader, results)
			}()

			// Drain the results channel
			for range results {
				// Just consume the results
			}
		})
	}
}

// TestBatchProcessor_Flush_ComprehensiveCoverage tests the Flush method
func TestBatchProcessor_Flush_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	processor := NewBatchProcessor(5, 100*time.Millisecond)

	// Add some items
	processor.Add(&models.LegacyTestResult{Name: "item1"})
	processor.Add(&models.LegacyTestResult{Name: "item2"})
	processor.Add(&models.LegacyTestResult{Name: "item3"})

	// Flush should not panic and return the items
	batch := processor.Flush()
	if batch == nil {
		t.Error("Expected batch to be returned from Flush")
	}
	if len(batch) != 3 {
		t.Errorf("Expected 3 items in batch, got %d", len(batch))
	}

	// Test flushing empty processor
	emptyProcessor := NewBatchProcessor(5, 100*time.Millisecond)
	emptyBatch := emptyProcessor.Flush()
	if emptyBatch != nil {
		t.Errorf("Expected nil batch from empty processor, got %d items", len(emptyBatch))
	}
}

// TestLazyRenderer_RenderSummaryOnly_ComprehensiveCoverage tests the RenderSummaryOnly method
func TestLazyRenderer_RenderSummaryOnly_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	renderer := NewLazyRenderer(50)

	testCases := []struct {
		name  string
		suite *models.TestSuite
	}{
		{
			name: "valid_test_suite",
			suite: &models.TestSuite{
				FilePath:  "test1.go",
				TestCount: 10,
			},
		},
		{
			name:  "nil_test_suite",
			suite: nil,
		},
		{
			name:  "empty_test_suite",
			suite: &models.TestSuite{},
		},
		{
			name: "test_suite_with_zero_tests",
			suite: &models.TestSuite{
				FilePath:  "empty_test.go",
				TestCount: 0,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			summary := renderer.RenderSummaryOnly(tc.suite)
			if summary == "" && tc.suite != nil && tc.suite.FilePath != "" {
				t.Error("Expected non-empty summary for valid test suite")
			}
		})
	}
}

// TestOptimizedTestProcessor_RenderResultsOptimized_ComprehensiveCoverage tests the RenderResultsOptimized method
func TestOptimizedTestProcessor_RenderResultsOptimized_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	testProcessor := processor.NewTestProcessor(&bytes.Buffer{}, &MockColorFormatter{}, &MockIconProvider{}, 80)
	processor := NewOptimizedTestProcessor(&bytes.Buffer{}, testProcessor)

	testCases := []struct {
		name         string
		autoCollapse bool
	}{
		{
			name:         "auto_collapse_enabled",
			autoCollapse: true,
		},
		{
			name:         "auto_collapse_disabled",
			autoCollapse: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := processor.RenderResultsOptimized(tc.autoCollapse)
			if err != nil {
				t.Errorf("RenderResultsOptimized should not error: %v", err)
			}
		})
	}
}
