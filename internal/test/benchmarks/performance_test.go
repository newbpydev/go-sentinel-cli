package benchmarks

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/processor"
	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// BenchmarkJSONParser tests the performance of parsing Go test JSON output
func BenchmarkJSONParser(b *testing.B) {
	// Create sample JSON test output for benchmarking
	sampleJSON := `{"Time":"2024-01-01T10:00:00.000Z","Action":"run","Package":"github.com/test/example","Test":"TestExample"}
{"Time":"2024-01-01T10:00:00.100Z","Action":"output","Package":"github.com/test/example","Test":"TestExample","Output":"=== RUN   TestExample\n"}
{"Time":"2024-01-01T10:00:00.200Z","Action":"output","Package":"github.com/test/example","Test":"TestExample","Output":"--- PASS: TestExample (0.10s)\n"}
{"Time":"2024-01-01T10:00:00.300Z","Action":"pass","Package":"github.com/test/example","Test":"TestExample","Elapsed":0.1}
`

	parser := processor.NewStreamParser()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(sampleJSON)
		results := make(chan *models.LegacyTestResult, 10)

		go func() {
			defer close(results)
			_ = parser.Parse(reader, results)
		}()

		// Consume results
		for range results {
		}
	}
}

// BenchmarkLargeTestSuiteParser tests parsing performance with large test suites
func BenchmarkLargeTestSuiteParser(b *testing.B) {
	// Generate large JSON test output (1000 tests)
	var jsonBuilder strings.Builder

	for i := 0; i < 1000; i++ {
		jsonBuilder.WriteString(`{"Time":"2024-01-01T10:00:00.000Z","Action":"run","Package":"github.com/test/example","Test":"TestExample`)
		jsonBuilder.WriteString(string(rune('0' + i%10)))
		jsonBuilder.WriteString(`"}` + "\n")

		jsonBuilder.WriteString(`{"Time":"2024-01-01T10:00:00.100Z","Action":"pass","Package":"github.com/test/example","Test":"TestExample`)
		jsonBuilder.WriteString(string(rune('0' + i%10)))
		jsonBuilder.WriteString(`","Elapsed":0.1}` + "\n")
	}

	largeJSON := jsonBuilder.String()
	parser := processor.NewStreamParser()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(largeJSON)
		results := make(chan *models.LegacyTestResult, 1000)

		go func() {
			defer close(results)
			_ = parser.Parse(reader, results)
		}()

		// Consume results
		for range results {
		}
	}
}

// BenchmarkTestSuiteCreation tests the performance of creating test suites
func BenchmarkTestSuiteCreation(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create a test suite with many tests
		suite := &models.TestSuite{
			FilePath:     "test/example_test.go",
			TestCount:    100,
			PassedCount:  90,
			FailedCount:  10,
			SkippedCount: 0,
			Duration:     time.Second,
			MemoryUsage:  50 * 1024 * 1024, // 50MB
		}

		// Add tests to the suite
		for j := 0; j < 100; j++ {
			status := models.TestStatusPassed
			if j%10 == 0 {
				status = models.TestStatusFailed
			}

			test := &models.LegacyTestResult{
				Name:     "TestExample" + string(rune('0'+j%10)),
				Status:   status,
				Duration: 10 * time.Millisecond,
				Package:  "github.com/test/example",
			}

			if status == models.TestStatusFailed {
				test.Error = &models.LegacyTestError{
					Message: "Test failed",
					Type:    "AssertionError",
				}
			}

			suite.Tests = append(suite.Tests, test)
		}
	}
}

// BenchmarkColorFormatterPerformance tests color formatter performance
func BenchmarkColorFormatterPerformance(b *testing.B) {
	formatter := colors.NewColorFormatter(false) // Disable colors for consistent benchmarking
	testMessage := "This is a test message that needs formatting"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer

		// Format different types of messages
		formatter.Green(testMessage)  // Success-like
		formatter.Red(testMessage)    // Error-like
		formatter.Yellow(testMessage) // Warning-like
		formatter.Blue(testMessage)   // Info-like

		// Use the buffer to avoid compiler optimization
		_ = buf.String()
	}
}

// BenchmarkMemoryUsage tests memory allocation patterns during processing
func BenchmarkMemoryUsage(b *testing.B) {
	formatter := colors.NewColorFormatter(false)
	iconProvider := colors.NewIconProvider(false)
	testProcessor := processor.NewTestProcessor(&bytes.Buffer{}, formatter, iconProvider, 80)

	// Create test suites
	suite := &models.TestSuite{
		FilePath:     "test/memory_test.go",
		TestCount:    100,
		PassedCount:  95,
		FailedCount:  5,
		SkippedCount: 0,
	}

	// Add many tests
	for i := 0; i < 100; i++ {
		test := &models.LegacyTestResult{
			Name:     "TestMemory" + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10)),
			Status:   models.TestStatusPassed,
			Duration: time.Millisecond,
			Package:  "github.com/test/memory",
		}
		suite.Tests = append(suite.Tests, test)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Force garbage collection before each iteration
		runtime.GC()

		testProcessor.AddTestSuite(suite)
	}
}

// TestParsingPerformanceThreshold ensures parsing performance meets requirements
func TestParsingPerformanceThreshold(t *testing.T) {
	// Test that we can parse 1000 test results in under 100ms
	const testCount = 1000
	const maxDuration = 100 * time.Millisecond

	// Generate JSON for many tests
	var jsonBuilder strings.Builder
	for i := 0; i < testCount; i++ {
		jsonBuilder.WriteString(`{"Time":"2024-01-01T10:00:00.000Z","Action":"run","Package":"github.com/test/example","Test":"TestExample`)
		jsonBuilder.WriteString(string(rune('0' + i%10)))
		jsonBuilder.WriteString(`"}` + "\n")

		jsonBuilder.WriteString(`{"Time":"2024-01-01T10:00:00.100Z","Action":"pass","Package":"github.com/test/example","Test":"TestExample`)
		jsonBuilder.WriteString(string(rune('0' + i%10)))
		jsonBuilder.WriteString(`","Elapsed":0.1}` + "\n")
	}

	parser := processor.NewStreamParser()
	reader := strings.NewReader(jsonBuilder.String())
	results := make(chan *models.LegacyTestResult, testCount)

	start := time.Now()

	go func() {
		defer close(results)
		_ = parser.Parse(reader, results)
	}()

	// Consume all results
	count := 0
	for range results {
		count++
	}

	duration := time.Since(start)

	if count != testCount {
		t.Errorf("Expected %d results, got %d", testCount, count)
	}

	if duration > maxDuration {
		t.Errorf("Parsing took %v, expected under %v", duration, maxDuration)
	}

	t.Logf("Parsed %d test results in %v", count, duration)
}

// TestMemoryLeakPrevention ensures no memory leaks during processing
func TestMemoryLeakPrevention(t *testing.T) {
	// Force garbage collection before starting
	runtime.GC()
	runtime.GC()

	var m1, m2 runtime.MemStats
	runtime.ReadMemStats(&m1)

	// Process many test suites
	formatter := colors.NewColorFormatter(false)
	iconProvider := colors.NewIconProvider(false)
	testProcessor := processor.NewTestProcessor(&bytes.Buffer{}, formatter, iconProvider, 80)

	for i := 0; i < 100; i++ {
		suite := &models.TestSuite{
			FilePath:     "test/memory_test.go",
			TestCount:    50,
			PassedCount:  45,
			FailedCount:  5,
			SkippedCount: 0,
		}

		// Add tests
		for j := 0; j < 50; j++ {
			test := &models.LegacyTestResult{
				Name:     "TestMemory" + string(rune('0'+j%10)),
				Status:   models.TestStatusPassed,
				Duration: time.Millisecond,
				Package:  "github.com/test/memory",
			}
			suite.Tests = append(suite.Tests, test)
		}

		testProcessor.AddTestSuite(suite)

		// Force garbage collection periodically
		if i%10 == 0 {
			runtime.GC()
		}
	}

	// Force garbage collection and measure memory
	runtime.GC()
	runtime.GC()
	runtime.ReadMemStats(&m2)

	// Check that memory usage didn't grow significantly
	memGrowth := int64(m2.Alloc) - int64(m1.Alloc)
	maxAcceptableGrowth := int64(10 * 1024 * 1024) // 10MB

	if memGrowth > maxAcceptableGrowth {
		t.Errorf("Memory grew by %d bytes, expected less than %d bytes", memGrowth, maxAcceptableGrowth)
	}

	t.Logf("Memory growth: %d bytes", memGrowth)
}
