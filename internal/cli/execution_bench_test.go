package cli

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

// BenchmarkTestRunner benchmarks basic test execution
func BenchmarkTestRunner(b *testing.B) {
	runner := &TestRunner{
		Verbose:    false,
		JSONOutput: true,
	}

	ctx := context.Background()
	testPaths := []string{"./internal/cli"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := runner.Run(ctx, testPaths)
		if err != nil {
			b.Logf("Test run failed (expected in benchmark): %v", err)
		}
	}
}

// BenchmarkOptimizedTestRunner benchmarks the optimized test runner
func BenchmarkOptimizedTestRunner(b *testing.B) {
	runner := NewOptimizedTestRunner()

	// Create test changes
	changes := []*FileChange{
		{
			Path: "internal/cli/test_runner.go",
			Type: ChangeTypeSource,
		},
		{
			Path: "internal/cli/processor.go",
			Type: ChangeTypeSource,
		},
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		result, err := runner.RunOptimized(ctx, changes)
		if err != nil {
			b.Logf("Optimized test run failed (expected in benchmark): %v", err)
		}
		_ = result
	}
}

// BenchmarkParallelTestRunner benchmarks parallel test execution
func BenchmarkParallelTestRunner(b *testing.B) {
	testRunner := &TestRunner{JSONOutput: true}
	cache := NewTestResultCache()
	parallelRunner := NewParallelTestRunner(4, testRunner, cache)

	testPaths := []string{
		"./internal/cli",
		"./internal/watch/core",
		"./internal/watch/debouncer",
		"./internal/watch/watcher",
	}

	ctx := context.Background()
	config := &Config{
		Verbosity: 0,
		Colors:    false,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		results, err := parallelRunner.RunParallel(ctx, testPaths, config)
		if err != nil {
			b.Logf("Parallel test run failed (expected in benchmark): %v", err)
		}
		_ = results
	}
}

// BenchmarkTestProcessor benchmarks test result processing
func BenchmarkTestProcessor(b *testing.B) {
	processor := NewTestProcessor(
		io.Discard,
		NewColorFormatter(false),
		NewIconProvider(false),
		80,
	)

	// Create large test suite
	suite := &TestSuite{
		FilePath:     "benchmark_test.go",
		TestCount:    500,
		PassedCount:  450,
		FailedCount:  50,
		SkippedCount: 0,
		Duration:     5 * time.Second,
	}

	// Add many test results
	for i := 0; i < 500; i++ {
		status := StatusPassed
		if i%10 == 0 {
			status = StatusFailed
		}

		test := &TestResult{
			Name:     fmt.Sprintf("TestBenchmark%d", i),
			Status:   status,
			Duration: 10 * time.Millisecond,
			Package:  "github.com/test/benchmark",
		}

		if status == StatusFailed {
			test.Error = &TestError{
				Message: "Benchmark test failed",
				Type:    "AssertionError",
			}
		}

		suite.Tests = append(suite.Tests, test)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		processor.Reset()
		processor.AddTestSuite(suite)
		_ = processor.RenderResults(false)
	}
}

// BenchmarkOptimizedTestProcessor benchmarks the optimized processor
func BenchmarkOptimizedTestProcessor(b *testing.B) {
	processor := NewOptimizedTestProcessor(
		io.Discard,
		NewColorFormatter(false),
		NewIconProvider(false),
		80,
	)

	// Create test suites
	suites := make([]*TestSuite, 10)
	for i := 0; i < 10; i++ {
		suite := &TestSuite{
			FilePath:    fmt.Sprintf("bench_%d_test.go", i),
			TestCount:   100,
			PassedCount: 90,
			FailedCount: 10,
		}

		for j := 0; j < 100; j++ {
			status := StatusPassed
			if j%10 == 0 {
				status = StatusFailed
			}

			test := &TestResult{
				Name:     fmt.Sprintf("Test%d_%d", i, j),
				Status:   status,
				Duration: time.Millisecond,
				Package:  fmt.Sprintf("github.com/test/bench%d", i),
			}
			suite.Tests = append(suite.Tests, test)
		}
		suites[i] = suite
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		processor.Clear()
		for _, suite := range suites {
			processor.AddTestSuite(suite)
		}
		_ = processor.RenderResultsOptimized(false)
	}
}

// BenchmarkStreamParser benchmarks JSON test output parsing
func BenchmarkStreamParser(b *testing.B) {
	// Generate realistic JSON test output
	var jsonBuilder strings.Builder
	for i := 0; i < 1000; i++ {
		testName := fmt.Sprintf("TestBenchmark%d", i)
		pkg := "github.com/test/benchmark"

		// Run event
		jsonBuilder.WriteString(fmt.Sprintf(
			`{"Time":"2024-01-01T10:00:%02d.000Z","Action":"run","Package":"%s","Test":"%s"}%s`,
			i%60, pkg, testName, "\n"))

		// Output event
		jsonBuilder.WriteString(fmt.Sprintf(
			`{"Time":"2024-01-01T10:00:%02d.100Z","Action":"output","Package":"%s","Test":"%s","Output":"=== RUN   %s\n"}%s`,
			i%60, pkg, testName, testName, "\n"))

		// Pass/Fail event
		action := "pass"
		if i%10 == 0 {
			action = "fail"
		}
		jsonBuilder.WriteString(fmt.Sprintf(
			`{"Time":"2024-01-01T10:00:%02d.200Z","Action":"%s","Package":"%s","Test":"%s","Elapsed":0.1}%s`,
			i%60, action, pkg, testName, "\n"))
	}

	jsonOutput := jsonBuilder.String()
	parser := NewStreamParser()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(jsonOutput)
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

// BenchmarkBatchProcessor benchmarks batch processing of test results
func BenchmarkBatchProcessor(b *testing.B) {
	processor := NewBatchProcessor(100, 100*time.Millisecond)

	// Create test results
	results := make([]*TestResult, 1000)
	for i := 0; i < 1000; i++ {
		status := StatusPassed
		if i%10 == 0 {
			status = StatusFailed
		}

		results[i] = &TestResult{
			Name:     fmt.Sprintf("TestBatch%d", i),
			Status:   status,
			Duration: time.Millisecond,
			Package:  "github.com/test/batch",
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, result := range results {
			batch := processor.Add(result)
			_ = batch // Consume batch
		}
		// Flush remaining
		_ = processor.Flush()
	}
}

// BenchmarkConcurrentTestProcessing benchmarks concurrent test processing
func BenchmarkConcurrentTestProcessing(b *testing.B) {
	processor := NewOptimizedTestProcessor(
		io.Discard,
		NewColorFormatter(false),
		NewIconProvider(false),
		80,
	)

	// Create test suites
	createSuite := func(id int) *TestSuite {
		suite := &TestSuite{
			FilePath:    fmt.Sprintf("concurrent_%d_test.go", id),
			TestCount:   50,
			PassedCount: 45,
			FailedCount: 5,
		}

		for j := 0; j < 50; j++ {
			status := StatusPassed
			if j%10 == 0 {
				status = StatusFailed
			}

			test := &TestResult{
				Name:     fmt.Sprintf("TestConcurrent%d_%d", id, j),
				Status:   status,
				Duration: time.Millisecond,
				Package:  fmt.Sprintf("github.com/test/concurrent%d", id),
			}
			suite.Tests = append(suite.Tests, test)
		}
		return suite
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		id := 0
		for pb.Next() {
			suite := createSuite(id)
			processor.AddTestSuite(suite)
			id++
		}
	})
}

// BenchmarkTestResultCache benchmarks the test result caching system
func BenchmarkTestResultCache(b *testing.B) {
	cache := NewTestResultCache()

	// Create test suites to cache
	suites := make([]*TestSuite, 100)
	for i := 0; i < 100; i++ {
		suite := &TestSuite{
			FilePath:    fmt.Sprintf("cache_test_%d.go", i),
			TestCount:   10,
			PassedCount: 10,
			Duration:    100 * time.Millisecond,
		}

		for j := 0; j < 10; j++ {
			test := &TestResult{
				Name:     fmt.Sprintf("TestCache%d_%d", i, j),
				Status:   StatusPassed,
				Duration: 10 * time.Millisecond,
				Package:  fmt.Sprintf("github.com/test/cache%d", i),
			}
			suite.Tests = append(suite.Tests, test)
		}
		suites[i] = suite
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Cache operations
		for j, suite := range suites {
			testPath := fmt.Sprintf("./test/cache%d", j)
			cache.CacheResult(testPath, suite)
		}

		// Retrieve operations
		for j := range suites {
			testPath := fmt.Sprintf("./test/cache%d", j)
			_, exists := cache.GetCachedResult(testPath)
			_ = exists
		}

		// Clear cache for next iteration
		cache.Clear()
	}
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate memory allocation patterns during test processing
		suite := &TestSuite{
			FilePath:    "memory_test.go",
			TestCount:   100,
			PassedCount: 90,
			FailedCount: 10,
			Tests:       make([]*TestResult, 0, 100),
		}

		for j := 0; j < 100; j++ {
			test := &TestResult{
				Name:     fmt.Sprintf("TestMemory%d", j),
				Status:   StatusPassed,
				Duration: time.Millisecond,
				Package:  "github.com/test/memory",
			}
			suite.Tests = append(suite.Tests, test)
		}

		// Process the suite
		processor := NewTestProcessor(
			io.Discard,
			NewColorFormatter(false),
			NewIconProvider(false),
			80,
		)
		processor.AddTestSuite(suite)
		_ = processor.RenderResults(false)
	}
}

// BenchmarkConcurrentCacheAccess benchmarks concurrent cache access
func BenchmarkConcurrentCacheAccess(b *testing.B) {
	cache := NewTestResultCache()

	// Pre-populate cache
	for i := 0; i < 50; i++ {
		suite := &TestSuite{
			FilePath:    fmt.Sprintf("concurrent_cache_%d_test.go", i),
			TestCount:   5,
			PassedCount: 5,
		}
		testPath := fmt.Sprintf("./test/concurrent%d", i)
		cache.CacheResult(testPath, suite)
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			testPath := fmt.Sprintf("./test/concurrent%d", i%50)

			// Mix of reads and writes
			if i%4 == 0 {
				// Write operation
				suite := &TestSuite{
					FilePath:    fmt.Sprintf("new_cache_%d_test.go", i),
					TestCount:   3,
					PassedCount: 3,
				}
				cache.CacheResult(testPath, suite)
			} else {
				// Read operation
				_, exists := cache.GetCachedResult(testPath)
				_ = exists
			}
			i++
		}
	})
}
