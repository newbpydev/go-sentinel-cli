package benchmarks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/cache"
	"github.com/newbpydev/go-sentinel/internal/test/runner"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// BenchmarkEndToEndWorkflow benchmarks complete end-to-end test execution
func BenchmarkEndToEndWorkflow(b *testing.B) {
	tempDir := b.TempDir()

	// Create realistic project structure
	createRealisticProject(b, tempDir)

	// Configure test runner
	testRunner := runner.NewBasicTestRunner(false, true) // verbose=false, jsonOutput=true

	ctx := context.Background()
	testPaths := []string{tempDir + "/..."}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Run tests
		_, err := testRunner.Run(ctx, testPaths)
		if err != nil {
			b.Logf("Test execution failed (expected in benchmark): %v", err)
		}
	}
}

// BenchmarkCacheOperations benchmarks cache operations
func BenchmarkCacheOperations(b *testing.B) {
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
		suites[i] = suite
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Cache results
		for j, suite := range suites {
			testPath := fmt.Sprintf("./cache/test%d", j)
			testCache.CacheResult(testPath, suite)
		}

		// Retrieve results
		for j := range suites {
			testPath := fmt.Sprintf("./cache/test%d", j)
			_, _ = testCache.GetCachedResult(testPath)
		}

		// Clear cache for next iteration
		testCache.Clear()
	}
}

// BenchmarkFileSystemOperations benchmarks file system operations
func BenchmarkFileSystemOperations(b *testing.B) {
	tempDir := b.TempDir()

	// Create test files
	for i := 0; i < 50; i++ {
		filename := filepath.Join(tempDir, fmt.Sprintf("file_%d.go", i))
		content := fmt.Sprintf("package main\nfunc Function%d() {}\n", i)
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			b.Fatalf("Failed to create test file: %v", err)
		}
	}

	testCache := cache.NewTestResultCache()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Analyze file changes
		for j := 0; j < 25; j++ {
			filename := filepath.Join(tempDir, fmt.Sprintf("file_%d.go", j))
			_, err := testCache.AnalyzeChange(filename)
			if err != nil {
				b.Logf("Failed to analyze change (expected in benchmark): %v", err)
			}
		}
	}
}

// BenchmarkConcurrentTestExecution benchmarks basic concurrent-like test execution
func BenchmarkConcurrentTestExecution(b *testing.B) {
	tempDir := b.TempDir()
	createLargeProject(b, tempDir, 5) // 5 packages

	testRunner := runner.NewBasicTestRunner(false, true)

	// Create test paths for all packages
	testPaths := make([]string, 5)
	for i := 0; i < 5; i++ {
		testPaths[i] = filepath.Join(tempDir, fmt.Sprintf("pkg%d", i))
	}

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Execute tests for each package sequentially (simulating concurrent load)
		for _, testPath := range testPaths {
			_, err := testRunner.Run(ctx, []string{testPath})
			if err != nil {
				b.Logf("Test execution failed (expected in benchmark): %v", err)
			}
		}
	}
}

// BenchmarkMemoryIntensiveWorkload benchmarks memory allocation patterns
func BenchmarkMemoryIntensiveWorkload(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Create many test suites to simulate memory usage
		suites := make([]*models.TestSuite, 50)
		for j := 0; j < 50; j++ {
			suite := &models.TestSuite{
				FilePath:     fmt.Sprintf("large_suite_%d_%d.go", i, j),
				TestCount:    100,
				PassedCount:  95,
				FailedCount:  5,
				SkippedCount: 0,
				Duration:     time.Second,
			}

			// Create test results
			for k := 0; k < 100; k++ {
				test := &models.LegacyTestResult{
					Name:     fmt.Sprintf("TestLarge_%d_%d_%d", i, j, k),
					Status:   models.TestStatusPassed,
					Duration: time.Millisecond,
					Package:  fmt.Sprintf("github.com/test/large%d", j),
				}
				suite.Tests = append(suite.Tests, test)
			}
			suites[j] = suite
		}

		// Process suites (simulate work)
		for _, suite := range suites {
			_ = fmt.Sprintf("Processing %s with %d tests", suite.FilePath, suite.TestCount)
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
