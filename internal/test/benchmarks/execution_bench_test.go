package benchmarks

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/cache"
	"github.com/newbpydev/go-sentinel/internal/test/processor"
	"github.com/newbpydev/go-sentinel/internal/test/runner"
	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// BenchmarkTestRunner benchmarks basic test execution
func BenchmarkTestRunner(b *testing.B) {
	testRunner := runner.NewBasicTestRunner(false, true) // verbose=false, jsonOutput=true

	ctx := context.Background()
	testPaths := []string{"./internal/test/benchmarks"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := testRunner.Run(ctx, testPaths)
		if err != nil {
			b.Logf("Test run failed (expected in benchmark): %v", err)
		}
	}
}

// BenchmarkTestProcessor benchmarks test result processing
func BenchmarkTestProcessor(b *testing.B) {
	formatter := colors.NewColorFormatter(false)
	iconProvider := colors.NewIconProvider(false)
	testProcessor := processor.NewTestProcessor(io.Discard, formatter, iconProvider, 80)

	// Create large test suite
	suite := &models.TestSuite{
		FilePath:     "benchmark_test.go",
		TestCount:    500,
		PassedCount:  450,
		FailedCount:  50,
		SkippedCount: 0,
		Duration:     5 * time.Second,
	}

	// Add many test results
	for i := 0; i < 500; i++ {
		status := models.TestStatusPassed
		if i%10 == 0 {
			status = models.TestStatusFailed
		}

		test := &models.LegacyTestResult{
			Name:     fmt.Sprintf("TestBenchmark%d", i),
			Status:   status,
			Duration: 10 * time.Millisecond,
			Package:  "github.com/test/benchmark",
		}

		if status == models.TestStatusFailed {
			test.Error = &models.LegacyTestError{
				Message: "Benchmark test failed",
				Type:    "AssertionError",
			}
		}

		suite.Tests = append(suite.Tests, test)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testProcessor.Reset()
		testProcessor.AddTestSuite(suite)
	}
}

// BenchmarkStreamParser benchmarks JSON test output parsing
func BenchmarkStreamParser(b *testing.B) {
	// Generate realistic JSON test output
	var jsonBuilder strings.Builder

	for i := 0; i < 100; i++ {
		jsonBuilder.WriteString(fmt.Sprintf(`{"Time":"2024-01-01T10:00:00.000Z","Action":"run","Package":"github.com/test/bench","Test":"TestBench%d"}`, i))
		jsonBuilder.WriteString("\n")

		jsonBuilder.WriteString(fmt.Sprintf(`{"Time":"2024-01-01T10:00:00.100Z","Action":"output","Package":"github.com/test/bench","Test":"TestBench%d","Output":"=== RUN   TestBench%d\n"}`, i, i))
		jsonBuilder.WriteString("\n")

		if i%10 == 0 {
			// Failed test
			jsonBuilder.WriteString(fmt.Sprintf(`{"Time":"2024-01-01T10:00:00.200Z","Action":"output","Package":"github.com/test/bench","Test":"TestBench%d","Output":"--- FAIL: TestBench%d (0.01s)\n"}`, i, i))
			jsonBuilder.WriteString("\n")
			jsonBuilder.WriteString(fmt.Sprintf(`{"Time":"2024-01-01T10:00:00.300Z","Action":"fail","Package":"github.com/test/bench","Test":"TestBench%d","Elapsed":0.01}`, i))
		} else {
			// Passed test
			jsonBuilder.WriteString(fmt.Sprintf(`{"Time":"2024-01-01T10:00:00.200Z","Action":"output","Package":"github.com/test/bench","Test":"TestBench%d","Output":"--- PASS: TestBench%d (0.01s)\n"}`, i, i))
			jsonBuilder.WriteString("\n")
			jsonBuilder.WriteString(fmt.Sprintf(`{"Time":"2024-01-01T10:00:00.300Z","Action":"pass","Package":"github.com/test/bench","Test":"TestBench%d","Elapsed":0.01}`, i))
		}
		jsonBuilder.WriteString("\n")
	}

	jsonOutput := jsonBuilder.String()
	parser := processor.NewStreamParser()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(jsonOutput)
		results := make(chan *models.LegacyTestResult, 100)

		go func() {
			defer close(results)
			_ = parser.Parse(reader, results)
		}()

		// Consume all results
		for range results {
		}
	}
}

// BenchmarkBatchProcessor benchmarks batch processing of test results
func BenchmarkBatchProcessor(b *testing.B) {
	formatter := colors.NewColorFormatter(false)
	iconProvider := colors.NewIconProvider(false)
	testProcessor := processor.NewTestProcessor(io.Discard, formatter, iconProvider, 80)

	// Create multiple test suites
	suites := make([]*models.TestSuite, 20)
	for i := 0; i < 20; i++ {
		suite := &models.TestSuite{
			FilePath:    fmt.Sprintf("batch_%d_test.go", i),
			TestCount:   50,
			PassedCount: 45,
			FailedCount: 5,
		}

		for j := 0; j < 50; j++ {
			status := models.TestStatusPassed
			if j%10 == 0 {
				status = models.TestStatusFailed
			}

			test := &models.LegacyTestResult{
				Name:     fmt.Sprintf("TestBatch%d_%d", i, j),
				Status:   status,
				Duration: time.Millisecond,
				Package:  fmt.Sprintf("github.com/test/batch%d", i),
			}
			suite.Tests = append(suite.Tests, test)
		}
		suites[i] = suite
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testProcessor.Reset()
		for _, suite := range suites {
			testProcessor.AddTestSuite(suite)
		}
	}
}

// BenchmarkTestResultCache benchmarks cache operations
func BenchmarkTestResultCache(b *testing.B) {
	testCache := cache.NewTestResultCache()

	// Create test suites for caching
	suites := make([]*models.TestSuite, 100)
	for i := 0; i < 100; i++ {
		suite := &models.TestSuite{
			FilePath:     fmt.Sprintf("test_cache_%d.go", i),
			TestCount:    10,
			PassedCount:  9,
			FailedCount:  1,
			SkippedCount: 0,
			Duration:     time.Millisecond * 100,
		}

		// Add some test results
		for j := 0; j < 10; j++ {
			test := &models.LegacyTestResult{
				Name:     fmt.Sprintf("TestCache%d_%d", i, j),
				Status:   models.TestStatusPassed,
				Duration: time.Millisecond,
				Package:  "github.com/test/cache",
			}
			suite.Tests = append(suite.Tests, test)
		}
		suites[i] = suite
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Cache results
		for _, suite := range suites {
			testCache.CacheResult(suite.FilePath, suite)
		}

		// Retrieve results
		for _, suite := range suites {
			_, _ = testCache.GetCachedResult(suite.FilePath)
		}

		// Clear cache for next iteration
		testCache.Clear()
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate creating many test results
		results := make([]*models.LegacyTestResult, 1000)
		for j := 0; j < 1000; j++ {
			results[j] = &models.LegacyTestResult{
				Name:     fmt.Sprintf("TestAlloc%d", j),
				Status:   models.TestStatusPassed,
				Duration: time.Microsecond,
				Package:  "github.com/test/alloc",
			}
		}

		// Simulate processing
		for _, result := range results {
			_ = result.Name + result.Package
		}
	}
}
