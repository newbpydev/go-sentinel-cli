package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTestResultCache(t *testing.T) {
	t.Run("Cache basic operations", func(t *testing.T) {
		cache := NewTestResultCache()

		// Test empty cache
		_, exists := cache.GetCachedResult("test/path")
		if exists {
			t.Error("Expected cache miss for non-existent path")
		}

		// Add a result
		suite := &TestSuite{
			FilePath:     "test/path",
			TestCount:    1,
			PassedCount:  1,
			FailedCount:  0,
			SkippedCount: 0,
		}

		cache.CacheResult("test/path", suite)

		// Retrieve result
		cached, exists := cache.GetCachedResult("test/path")
		if !exists {
			t.Error("Expected cache hit after storing result")
		}

		if cached.Suite.TestCount != 1 {
			t.Errorf("Expected TestCount 1, got %d", cached.Suite.TestCount)
		}
	})

	t.Run("Change analysis", func(t *testing.T) {
		// Create temporary test file
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "example_test.go")

		content := `package main
import "testing"
func TestExample(t *testing.T) {
	if 1+1 != 2 {
		t.Error("Math doesn't work")
	}
}`

		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cache := NewTestResultCache()

		// Analyze the change
		change, err := cache.AnalyzeChange(testFile)
		if err != nil {
			t.Fatalf("Failed to analyze change: %v", err)
		}

		if change.Type != ChangeTypeTest {
			t.Errorf("Expected ChangeTypeTest, got %d", change.Type)
		}

		if !change.IsNew {
			t.Error("Expected IsNew to be true for first analysis")
		}

		if len(change.AffectedTests) == 0 {
			t.Error("Expected affected tests to be populated")
		}
	})
}

func TestIncrementalRenderer(t *testing.T) {
	t.Run("Incremental rendering", func(t *testing.T) {
		var buffer bytes.Buffer
		cache := NewTestResultCache()
		renderer := NewIncrementalRenderer(
			&buffer,
			NewColorFormatter(false),
			NewIconProvider(false),
			80,
			cache,
		)

		// Create test suites
		suite1 := &TestSuite{
			FilePath:     "pkg/test1",
			TestCount:    2,
			PassedCount:  2,
			FailedCount:  0,
			SkippedCount: 0,
			Tests: []*TestResult{
				{Name: "TestOne", Status: StatusPassed, Duration: time.Millisecond * 100},
				{Name: "TestTwo", Status: StatusPassed, Duration: time.Millisecond * 150},
			},
		}

		suites := map[string]*TestSuite{
			"pkg/test1": suite1,
		}

		stats := &TestRunStats{
			TotalTests:  2,
			PassedTests: 2,
			FailedTests: 0,
			Duration:    time.Millisecond * 250,
		}

		changes := []*FileChange{
			{
				Path: "pkg/test1/main_test.go",
				Type: ChangeTypeTest,
			},
		}

		// Render incremental results
		err := renderer.RenderIncrementalResults(suites, stats, changes)
		if err != nil {
			t.Fatalf("Failed to render incremental results: %v", err)
		}

		output := buffer.String()
		if output == "" {
			t.Error("Expected output from incremental renderer")
		}

		// Should contain file change information
		if !strings.Contains(output, "File changes detected") {
			t.Error("Expected file changes summary in output")
		}
	})
}

func TestParallelTestRunner(t *testing.T) {
	t.Run("Concurrency control", func(t *testing.T) {
		cache := NewTestResultCache()
		testRunner := &TestRunner{JSONOutput: true}

		// Test concurrency limits
		runner := NewParallelTestRunner(0, testRunner, cache) // Should default to 4
		if runner.maxConcurrency != 4 {
			t.Errorf("Expected default concurrency 4, got %d", runner.maxConcurrency)
		}

		runner = NewParallelTestRunner(8, testRunner, cache)
		if runner.maxConcurrency != 8 {
			t.Errorf("Expected concurrency 8, got %d", runner.maxConcurrency)
		}
	})

	t.Run("Result merging", func(t *testing.T) {
		processor := NewTestProcessor(
			io.Discard,
			NewColorFormatter(false),
			NewIconProvider(false),
			80,
		)

		// Create test results
		results := []*ParallelTestResult{
			{
				TestPath: "pkg/test1",
				Suite: &TestSuite{
					FilePath:    "pkg/test1",
					TestCount:   1,
					PassedCount: 1,
				},
				Error:    nil,
				Duration: time.Millisecond * 100,
			},
			{
				TestPath: "pkg/test2",
				Suite: &TestSuite{
					FilePath:    "pkg/test2",
					TestCount:   1,
					FailedCount: 1,
				},
				Error:    nil,
				Duration: time.Millisecond * 200,
			},
		}

		// Merge results
		MergeResults(processor, results)

		// Verify merged results
		stats := processor.GetStats()
		if stats.TotalTests != 2 {
			t.Errorf("Expected 2 total tests, got %d", stats.TotalTests)
		}

		if stats.PassedTests != 1 {
			t.Errorf("Expected 1 passed test, got %d", stats.PassedTests)
		}

		if stats.FailedTests != 1 {
			t.Errorf("Expected 1 failed test, got %d", stats.FailedTests)
		}
	})
}
