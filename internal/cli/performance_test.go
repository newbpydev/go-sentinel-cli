package cli

import (
	"bytes"
	"io"
	"runtime"
	"strings"
	"testing"
	"time"
)

// BenchmarkJSONParser tests the performance of parsing Go test JSON output
func BenchmarkJSONParser(b *testing.B) {
	// Create sample JSON test output for benchmarking
	sampleJSON := `{"Time":"2024-01-01T10:00:00.000Z","Action":"run","Package":"github.com/test/example","Test":"TestExample"}
{"Time":"2024-01-01T10:00:00.100Z","Action":"output","Package":"github.com/test/example","Test":"TestExample","Output":"=== RUN   TestExample\n"}
{"Time":"2024-01-01T10:00:00.200Z","Action":"output","Package":"github.com/test/example","Test":"TestExample","Output":"--- PASS: TestExample (0.10s)\n"}
{"Time":"2024-01-01T10:00:00.300Z","Action":"pass","Package":"github.com/test/example","Test":"TestExample","Elapsed":0.1}
`

	parser := NewStreamParser()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(sampleJSON)
		results := make(chan *TestResult, 10)

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
	parser := NewStreamParser()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(largeJSON)
		results := make(chan *TestResult, 1000)

		go func() {
			defer close(results)
			_ = parser.Parse(reader, results)
		}()

		// Consume results
		for range results {
		}
	}
}

// BenchmarkSuiteRenderer tests the rendering performance for test suites
func BenchmarkSuiteRenderer(b *testing.B) {
	// Create a test suite with many tests
	suite := &TestSuite{
		FilePath:     "test/example_test.go",
		TestCount:    100,
		PassedCount:  90,
		FailedCount:  10,
		SkippedCount: 0,
		Duration:     time.Second,
		MemoryUsage:  50 * 1024 * 1024, // 50MB
	}

	// Add tests to the suite
	for i := 0; i < 100; i++ {
		status := StatusPassed
		if i%10 == 0 {
			status = StatusFailed
		}

		test := &TestResult{
			Name:     "TestExample" + string(rune('0'+i%10)),
			Status:   status,
			Duration: 10 * time.Millisecond,
			Package:  "github.com/test/example",
		}

		if status == StatusFailed {
			test.Error = &TestError{
				Message: "Test failed",
				Type:    "AssertionError",
			}
		}

		suite.Tests = append(suite.Tests, test)
	}

	formatter := NewColorFormatter(false) // Disable colors for consistent benchmarking
	icons := NewIconProvider(false)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		renderer := NewSuiteRenderer(&buf, formatter, icons, 80)
		_ = renderer.RenderSuite(suite, false)
	}
}

// BenchmarkFailedTestRenderer tests the rendering performance for failed tests
func BenchmarkFailedTestRenderer(b *testing.B) {
	// Create multiple failed tests with source context
	failedTests := make([]*TestResult, 50)

	for i := 0; i < 50; i++ {
		failedTests[i] = &TestResult{
			Name:   "TestFailedExample" + string(rune('0'+i%10)),
			Status: StatusFailed,
			Error: &TestError{
				Type:    "AssertionError",
				Message: "Expected value, got different value",
				Location: &SourceLocation{
					File:   "test/example_test.go",
					Line:   42 + i,
					Column: 16,
				},
				SourceContext: []string{
					"func TestFailedExample() {",
					"    result := doSomething()",
					"    if result != expected {",
					"        t.Errorf(\"Expected %v, got %v\", expected, result)",
					"    }",
				},
				HighlightedLine: 2,
			},
		}
	}

	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		renderer := NewFailedTestRenderer(&buf, formatter, icons, 80)
		_ = renderer.RenderFailedTests(failedTests)
	}
}

// BenchmarkMemoryUsage tests memory allocation patterns during processing
func BenchmarkMemoryUsage(b *testing.B) {
	processor := NewTestProcessor(io.Discard, NewColorFormatter(false), NewIconProvider(false), 80)

	// Create test suites
	suite := &TestSuite{
		FilePath:     "test/memory_test.go",
		TestCount:    1000,
		PassedCount:  1000,
		FailedCount:  0,
		SkippedCount: 0,
	}

	// Add many tests
	for i := 0; i < 1000; i++ {
		test := &TestResult{
			Name:     "TestMemory" + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10)),
			Status:   StatusPassed,
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

		processor.AddTestSuite(suite)
		_ = processor.RenderResults(false)

		// Clear processor state
		processor = NewTestProcessor(io.Discard, NewColorFormatter(false), NewIconProvider(false), 80)
	}
}

// TestParsingPerformanceThreshold ensures parsing performance meets minimum requirements
func TestParsingPerformanceThreshold(t *testing.T) {
	// Generate test data (100 tests)
	var jsonBuilder strings.Builder
	for i := 0; i < 100; i++ {
		jsonBuilder.WriteString(`{"Time":"2024-01-01T10:00:00.000Z","Action":"run","Package":"github.com/test/example","Test":"TestExample`)
		jsonBuilder.WriteString(string(rune('0' + i%10)))
		jsonBuilder.WriteString(`"}` + "\n")

		jsonBuilder.WriteString(`{"Time":"2024-01-01T10:00:00.100Z","Action":"pass","Package":"github.com/test/example","Test":"TestExample`)
		jsonBuilder.WriteString(string(rune('0' + i%10)))
		jsonBuilder.WriteString(`","Elapsed":0.1}` + "\n")
	}

	testData := jsonBuilder.String()
	parser := NewStreamParser()

	start := time.Now()
	reader := strings.NewReader(testData)
	results := make(chan *TestResult, 100)

	go func() {
		defer close(results)
		err := parser.Parse(reader, results)
		if err != nil {
			t.Errorf("Parser error: %v", err)
		}
	}()

	// Consume results
	count := 0
	for range results {
		count++
	}

	elapsed := time.Since(start)

	// Should parse 100 tests in less than 100ms (1ms per test)
	threshold := 100 * time.Millisecond
	if elapsed > threshold {
		t.Errorf("Parsing too slow: %v > %v (processed %d tests)", elapsed, threshold, count)
	}

	if count != 100 {
		t.Errorf("Expected 100 parsed tests, got %d", count)
	}
}

// TestRenderingPerformanceThreshold ensures rendering performance meets minimum requirements
func TestRenderingPerformanceThreshold(t *testing.T) {
	// Create a test suite with 50 tests
	suite := &TestSuite{
		FilePath:     "test/performance_test.go",
		TestCount:    50,
		PassedCount:  45,
		FailedCount:  5,
		SkippedCount: 0,
		Duration:     time.Second,
		MemoryUsage:  25 * 1024 * 1024,
	}

	// Add tests
	for i := 0; i < 50; i++ {
		status := StatusPassed
		if i%10 == 0 {
			status = StatusFailed
		}

		test := &TestResult{
			Name:     "TestPerformance" + string(rune('0'+i%10)),
			Status:   status,
			Duration: 20 * time.Millisecond,
			Package:  "github.com/test/performance",
		}

		if status == StatusFailed {
			test.Error = &TestError{
				Message: "Performance test failed",
				Type:    "AssertionError",
			}
		}

		suite.Tests = append(suite.Tests, test)
	}

	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)

	start := time.Now()

	var buf bytes.Buffer
	renderer := NewSuiteRenderer(&buf, formatter, icons, 80)
	err := renderer.RenderSuite(suite, false)
	if err != nil {
		t.Errorf("Renderer error: %v", err)
	}

	elapsed := time.Since(start)

	// Should render 50 tests in less than 50ms (1ms per test)
	threshold := 50 * time.Millisecond
	if elapsed > threshold {
		t.Errorf("Rendering too slow: %v > %v", elapsed, threshold)
	}

	// Verify output was generated
	if buf.Len() == 0 {
		t.Error("No output generated by renderer")
	}
}

// TestMemoryLeakPrevention ensures no memory leaks during long-running operations
func TestMemoryLeakPrevention(t *testing.T) {
	// Get initial memory stats
	var initialStats, finalStats runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&initialStats)

	processor := NewTestProcessor(io.Discard, NewColorFormatter(false), NewIconProvider(false), 80)

	// Simulate processing many test suites
	for iteration := 0; iteration < 100; iteration++ {
		suite := &TestSuite{
			FilePath:     "test/leak_test.go",
			TestCount:    10,
			PassedCount:  10,
			FailedCount:  0,
			SkippedCount: 0,
		}

		// Add tests
		for i := 0; i < 10; i++ {
			test := &TestResult{
				Name:     "TestLeak" + string(rune('0'+i)),
				Status:   StatusPassed,
				Duration: time.Millisecond,
				Package:  "github.com/test/leak",
				Output:   "Test output that should be garbage collected",
			}
			suite.Tests = append(suite.Tests, test)
		}

		processor.AddTestSuite(suite)
		_ = processor.RenderResults(false)

		// Clear processor state - simulate new test run
		processor = NewTestProcessor(io.Discard, NewColorFormatter(false), NewIconProvider(false), 80)

		// Force garbage collection every 10 iterations
		if iteration%10 == 0 {
			runtime.GC()
		}
	}

	// Final garbage collection and memory check
	runtime.GC()
	runtime.GC() // Run twice to ensure cleanup
	runtime.ReadMemStats(&finalStats)

	// Check memory growth - should not grow significantly	// Handle potential underflow by checking if final memory is actually less than initial	var memoryGrowth uint64	if finalStats.Alloc > initialStats.Alloc {		memoryGrowth = finalStats.Alloc - initialStats.Alloc	} else {		memoryGrowth = 0 // Memory was garbage collected, which is good	}		maxAllowedGrowth := uint64(10 * 1024 * 1024) // 10MB		if memoryGrowth > maxAllowedGrowth {		t.Errorf("Potential memory leak detected: memory grew by %d bytes (max allowed: %d)", 			memoryGrowth, maxAllowedGrowth)	}		t.Logf("Memory growth: %d bytes (initial: %d, final: %d)", 		memoryGrowth, initialStats.Alloc, finalStats.Alloc)
}
