package cli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// BenchmarkEndToEndWorkflow benchmarks complete end-to-end test execution
func BenchmarkEndToEndWorkflow(b *testing.B) {
	tempDir := b.TempDir()

	// Create realistic project structure
	createRealisticProject(b, tempDir)

	// Configure test runner
	runner := &TestRunner{
		Verbose:    false,
		JSONOutput: true,
	}

	var buf bytes.Buffer
	processor := NewTestProcessor(
		&buf, // Use buffer instead of nil
		NewColorFormatter(false),
		NewIconProvider(false),
		80,
	)

	ctx := context.Background()
	testPaths := []string{tempDir + "/..."}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		processor.Reset()

		// Run tests
		_, err := runner.Run(ctx, testPaths)
		if err != nil {
			b.Logf("Test execution failed (expected in benchmark): %v", err)
		}
	}
}

// BenchmarkWatchModeIntegration benchmarks complete watch mode operation
func BenchmarkWatchModeIntegration(b *testing.B) {
	tempDir := b.TempDir()

	// Create test files
	createWatchTestFiles(b, tempDir)

	options := WatchOptions{
		Paths:            []string{tempDir},
		IgnorePatterns:   []string{"*.log", "*.tmp"},
		TestPatterns:     []string{"*_test.go"},
		Mode:             WatchChanged,
		DebounceInterval: 50 * time.Millisecond,
		ClearTerminal:    false,
		Writer:           nil,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		watcher, err := NewTestWatcher(options)
		if err != nil {
			b.Fatalf("Failed to create test watcher: %v", err)
		}

		// Simulate file change
		testFile := filepath.Join(tempDir, "main_test.go")
		content := fmt.Sprintf("package main\n// Modified at %d\nfunc TestMain%d(t *testing.T) {}\n", i, i)
		err = os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			b.Fatalf("Failed to modify test file: %v", err)
		}

		_ = watcher.Stop()
	}
}

// BenchmarkOptimizedPipeline benchmarks the complete optimized test pipeline
func BenchmarkOptimizedPipeline(b *testing.B) {
	tempDir := b.TempDir()
	createRealisticProject(b, tempDir)

	// Set up optimized components
	cache := NewTestResultCache()
	optimizedRunner := NewOptimizedTestRunner()

	changes := []*FileChange{
		{Path: filepath.Join(tempDir, "main.go"), Type: ChangeTypeSource},
		{Path: filepath.Join(tempDir, "main_test.go"), Type: ChangeTypeTest},
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Run optimized pipeline
		result, err := optimizedRunner.RunOptimized(ctx, changes)
		if err != nil {
			b.Logf("Optimized pipeline failed (expected in benchmark): %v", err)
		}
		_ = result

		// Simulate cache operations
		for j, change := range changes {
			suite := &TestSuite{
				FilePath:    change.Path,
				TestCount:   10,
				PassedCount: 9,
				FailedCount: 1,
			}
			testPath := fmt.Sprintf("test_%d_%d", i, j)
			cache.CacheResult(testPath, suite)
		}
	}
}

// BenchmarkConcurrentTestExecution benchmarks parallel test execution under load
func BenchmarkConcurrentTestExecution(b *testing.B) {
	tempDir := b.TempDir()
	createLargeProject(b, tempDir, 20) // 20 packages

	testRunner := &TestRunner{JSONOutput: true}
	cache := NewTestResultCache()
	parallelRunner := NewParallelTestRunner(8, testRunner, cache) // High concurrency

	// Create test paths for all packages
	testPaths := make([]string, 20)
	for i := 0; i < 20; i++ {
		testPaths[i] = filepath.Join(tempDir, fmt.Sprintf("pkg%d", i))
	}

	config := &Config{
		Verbosity: 0,
		Colors:    false,
		Parallel:  8,
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		results, err := parallelRunner.RunParallel(ctx, testPaths, config)
		if err != nil {
			b.Logf("Parallel execution failed (expected in benchmark): %v", err)
		}
		_ = results
	}
}

// BenchmarkMemoryIntensiveWorkload benchmarks performance under memory pressure
func BenchmarkMemoryIntensiveWorkload(b *testing.B) {
	// Create large test suites that consume significant memory
	var buf bytes.Buffer
	processor := NewOptimizedTestProcessor(
		&buf, // Use buffer instead of nil
		NewColorFormatter(false),
		NewIconProvider(false),
		80,
	)

	// Create many large test suites
	largeSuites := make([]*TestSuite, 50)
	for i := 0; i < 50; i++ {
		suite := &TestSuite{
			FilePath:    fmt.Sprintf("large_suite_%d_test.go", i),
			TestCount:   1000,
			PassedCount: 950,
			FailedCount: 50,
			Tests:       make([]*TestResult, 1000),
		}

		// Create many test results
		for j := 0; j < 1000; j++ {
			suite.Tests[j] = &TestResult{
				Name:     fmt.Sprintf("TestLarge_%d_%d", i, j),
				Status:   StatusPassed,
				Duration: time.Millisecond,
				Package:  fmt.Sprintf("github.com/test/large%d", i),
			}
		}
		largeSuites[i] = suite
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		processor.Clear()

		// Process all large suites
		for _, suite := range largeSuites {
			processor.AddTestSuite(suite)
		}

		_ = processor.RenderResultsOptimized(false)
	}
}

// BenchmarkFileSystemStress benchmarks file system operations under stress
func BenchmarkFileSystemStress(b *testing.B) {
	tempDir := b.TempDir()

	// Create many files to watch
	for i := 0; i < 500; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("stress_%d.go", i))
		content := fmt.Sprintf("package stress%d\nfunc Test%d() {}\n", i%10, i)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create stress test file: %v", err)
		}
	}

	cache := NewTestResultCache()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Analyze many file changes
		for j := 0; j < 100; j++ {
			filename := filepath.Join(tempDir, fmt.Sprintf("stress_%d.go", j%500))
			change, err := cache.AnalyzeChange(filename)
			if err != nil {
				b.Logf("Failed to analyze change (expected in benchmark): %v", err)
			}
			_ = change
		}
	}
}

// BenchmarkRealWorldScenario benchmarks realistic development workflow
func BenchmarkRealWorldScenario(b *testing.B) {
	tempDir := b.TempDir()
	createRealisticProject(b, tempDir)

	// Set up complete pipeline
	controller := NewAppController()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Simulate complete development workflow
		err := controller.Run([]string{"run"})
		if err != nil {
			b.Logf("Real world scenario failed (expected in benchmark): %v", err)
		}
	}
}

// BenchmarkCacheEfficiency benchmarks cache hit rates and performance
func BenchmarkCacheEfficiency(b *testing.B) {
	cache := NewTestResultCache()

	// Pre-populate cache with many results
	for i := 0; i < 1000; i++ {
		suite := &TestSuite{
			FilePath:    fmt.Sprintf("cache_test_%d.go", i),
			TestCount:   10,
			PassedCount: 9,
			FailedCount: 1,
		}
		testPath := fmt.Sprintf("./cache/test%d", i)
		cache.CacheResult(testPath, suite)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Mix of cache hits and misses
		for j := 0; j < 100; j++ {
			testPath := fmt.Sprintf("./cache/test%d", j%1200) // Some hits, some misses

			if j%3 == 0 {
				// Cache lookup
				_, exists := cache.GetCachedResult(testPath)
				_ = exists
			} else {
				// Cache write
				suite := &TestSuite{
					FilePath:    fmt.Sprintf("new_cache_%d_%d.go", i, j),
					TestCount:   5,
					PassedCount: 5,
				}
				cache.CacheResult(testPath, suite)
			}
		}
	}
}

// Helper functions for benchmark setup

func createRealisticProject(b *testing.B, rootDir string) {
	// Create main source files
	mainContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}

func add(a, b int) int {
	return a + b
}

func multiply(a, b int) int {
	return a * b
}`

	testContent := `package main

import "testing"

func TestAdd(t *testing.T) {
	result := add(2, 3)
	if result != 5 {
		t.Errorf("Expected 5, got %d", result)
	}
}

func TestMultiply(t *testing.T) {
	result := multiply(4, 5)
	if result != 20 {
		t.Errorf("Expected 20, got %d", result)
	}
}

func TestAddZero(t *testing.T) {
	result := add(5, 0)
	if result != 5 {
		t.Errorf("Expected 5, got %d", result)
	}
}`

	mainFile := filepath.Join(rootDir, "main.go")
	testFile := filepath.Join(rootDir, "main_test.go")

	if err := os.WriteFile(mainFile, []byte(mainContent), 0644); err != nil {
		b.Fatalf("Failed to create main.go: %v", err)
	}
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		b.Fatalf("Failed to create main_test.go: %v", err)
	}
}

func createWatchTestFiles(b *testing.B, rootDir string) {
	// Create multiple test files for watch mode
	for i := 0; i < 5; i++ {
		filename := filepath.Join(rootDir, fmt.Sprintf("watch_%d_test.go", i))
		content := fmt.Sprintf(`package main

import "testing"

func TestWatch%d(t *testing.T) {
	if 1+1 != 2 {
		t.Error("Math doesn't work")
	}
}`, i)

		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create watch test file: %v", err)
		}
	}
}

func createLargeProject(b *testing.B, rootDir string, packageCount int) {
	for i := 0; i < packageCount; i++ {
		pkgDir := filepath.Join(rootDir, fmt.Sprintf("pkg%d", i))
		if err := os.MkdirAll(pkgDir, 0755); err != nil {
			b.Fatalf("Failed to create package directory: %v", err)
		}

		// Create source file
		sourceFile := filepath.Join(pkgDir, "source.go")
		sourceContent := fmt.Sprintf(`package pkg%d

func Function%d() int {
	return %d
}`, i, i, i*10)

		// Create test file
		testFile := filepath.Join(pkgDir, "source_test.go")
		testContent := fmt.Sprintf(`package pkg%d

import "testing"

func TestFunction%d(t *testing.T) {
	result := Function%d()
	if result != %d {
		t.Errorf("Expected %d, got %%d", result)
	}
}`, i, i, i, i*10, i*10)

		if err := os.WriteFile(sourceFile, []byte(sourceContent), 0644); err != nil {
			b.Fatalf("Failed to create source file: %v", err)
		}
		if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
	}
}
