package watch

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/cache"
	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/internal/ui/renderer"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// TestIntelligentWatchEfficiency tests that the watch mode is truly efficient
func TestIntelligentWatchEfficiency(t *testing.T) {
	t.Run("Should only run affected tests", func(t *testing.T) {
		testCache := cache.NewTestResultCache()

		// Create a test file change
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "specific_test.go")

		content := `package main
import "testing"
func TestSpecific(t *testing.T) {
	if 1+1 != 2 {
		t.Error("Math doesn't work")
	}
}`

		err := os.WriteFile(testFile, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		change, err := testCache.AnalyzeChange(testFile)
		if err != nil {
			t.Fatalf("Failed to analyze change: %v", err)
		}

		staleTests := testCache.GetStaleTests([]*cache.FileChange{change})

		// Should only return the directory of the changed test
		expectedDir := filepath.Dir(testFile)
		if len(staleTests) != 1 || staleTests[0] != expectedDir {
			t.Errorf("Expected stale tests [%s], got %v", expectedDir, staleTests)
		}
	})

	t.Run("Cache invalidation works correctly", func(t *testing.T) {
		testCache := cache.NewTestResultCache()

		// Add initial cache result
		suite := &models.TestSuite{
			FilePath:     "pkg/test",
			TestCount:    3,
			PassedCount:  3,
			FailedCount:  0,
			SkippedCount: 0,
		}
		testCache.CacheResult("pkg/test", suite)

		// Verify cache hit
		cached, exists := testCache.GetCachedResult("pkg/test")
		if !exists {
			t.Error("Expected cache hit")
		}
		if cached.Suite.TestCount != 3 {
			t.Errorf("Expected 3 tests in cache, got %d", cached.Suite.TestCount)
		}
	})
}

// TestWatchModePerformance tests that watch mode performs well under load
func TestWatchModePerformance(t *testing.T) {
	t.Run("Handles rapid file changes efficiently", func(t *testing.T) {
		testCache := cache.NewTestResultCache()
		colorFormatter := colors.NewColorFormatter(false)
		iconProvider := colors.NewIconProvider(false)

		incrementalRenderer := renderer.NewIncrementalRenderer(
			io.Discard,
			colorFormatter,
			iconProvider,
			80,
			testCache,
		)

		// Simulate many rapid changes
		tempDir := t.TempDir()
		changes := make([]*cache.FileChange, 50)

		start := time.Now()

		for i := 0; i < 50; i++ {
			testFile := filepath.Join(tempDir, fmt.Sprintf("test_%d.go", i))
			content := fmt.Sprintf(`package main
import "testing"
func Test%d(t *testing.T) {
	if 1+1 != 2 {
		t.Error("Math doesn't work")
	}
}`, i)

			err := os.WriteFile(testFile, []byte(content), 0644)
			if err != nil {
				t.Fatalf("Failed to create test file %d: %v", i, err)
			}

			change, err := testCache.AnalyzeChange(testFile)
			if err != nil {
				t.Fatalf("Failed to analyze change %d: %v", i, err)
			}
			changes[i] = change
		}

		// Process all changes
		staleTests := testCache.GetStaleTests(changes)

		// Render results
		err := incrementalRenderer.RenderIncrementalResults(
			make(map[string]*models.TestSuite),
			&models.TestRunStats{},
			changes,
		)
		if err != nil {
			t.Fatalf("Failed to render results: %v", err)
		}

		duration := time.Since(start)

		// Should handle 50 changes quickly (under 1 second)
		if duration > time.Second {
			t.Errorf("Processing 50 changes took too long: %v", duration)
		}

		// Should return reasonable number of stale tests (not all 50)
		if len(staleTests) > 10 {
			t.Errorf("Too many stale tests returned: %d (should be optimized)", len(staleTests))
		}
	})

	t.Run("Memory usage remains stable", func(t *testing.T) {
		testCache := cache.NewTestResultCache()

		// Add many cache entries
		for i := 0; i < 1000; i++ {
			suite := &models.TestSuite{
				FilePath:     fmt.Sprintf("pkg/test%d", i),
				TestCount:    3,
				PassedCount:  3,
				FailedCount:  0,
				SkippedCount: 0,
			}
			testCache.CacheResult(fmt.Sprintf("pkg/test%d", i), suite)
		}

		stats := testCache.GetStats()
		if stats["cached_results"].(int) != 1000 {
			t.Errorf("Expected 1000 cached results, got %d", stats["cached_results"])
		}

		// Clear cache and verify cleanup
		testCache.Clear()

		stats = testCache.GetStats()
		if stats["cached_results"].(int) != 0 {
			t.Errorf("Expected 0 cached results after clear, got %d", stats["cached_results"])
		}
	})
}

// TestWatchModeUserExperience tests the UX aspects of watch mode
func TestWatchModeUserExperience(t *testing.T) {
	t.Run("Clear messaging when no tests need to run", func(t *testing.T) {
		var buffer bytes.Buffer
		testCache := cache.NewTestResultCache()
		colorFormatter := colors.NewColorFormatter(false)
		iconProvider := colors.NewIconProvider(false)

		incrementalRenderer := renderer.NewIncrementalRenderer(
			&buffer,
			colorFormatter,
			iconProvider,
			80,
			testCache,
		)

		// Simulate a change that doesn't require tests
		changes := []*cache.FileChange{
			{
				Path: "docs/README.md",
				Type: cache.ChangeTypeConfig,
			},
		}

		// Render with no test suites (meaning no tests ran)
		err := incrementalRenderer.RenderIncrementalResults(
			make(map[string]*models.TestSuite),
			&models.TestRunStats{},
			changes,
		)
		if err != nil {
			t.Fatalf("Failed to render results: %v", err)
		}

		output := buffer.String()
		if !strings.Contains(output, "No test changes detected") {
			t.Error("Expected clear messaging when no tests need to run")
		}
	})

	t.Run("Informative output for file changes", func(t *testing.T) {
		var buffer bytes.Buffer
		testCache := cache.NewTestResultCache()
		colorFormatter := colors.NewColorFormatter(false)
		iconProvider := colors.NewIconProvider(false)

		incrementalRenderer := renderer.NewIncrementalRenderer(
			&buffer,
			colorFormatter,
			iconProvider,
			80,
			testCache,
		)

		// Simulate various types of changes
		changes := []*cache.FileChange{
			{
				Path: "src/main.go",
				Type: cache.ChangeTypeSource,
			},
			{
				Path: "test/main_test.go",
				Type: cache.ChangeTypeTest,
			},
		}

		// Render with changes but no test results
		err := incrementalRenderer.RenderIncrementalResults(
			make(map[string]*models.TestSuite),
			&models.TestRunStats{},
			changes,
		)
		if err != nil {
			t.Fatalf("Failed to render results: %v", err)
		}

		output := buffer.String()
		if !strings.Contains(output, "File changes detected") {
			t.Error("Expected file changes summary in output")
		}
		if !strings.Contains(output, "src/main.go") {
			t.Error("Expected source file change in output")
		}
		if !strings.Contains(output, "test/main_test.go") {
			t.Error("Expected test file change in output")
		}
	})
}

// TestWatchModeEdgeCases tests edge cases in watch mode
func TestWatchModeEdgeCases(t *testing.T) {
	t.Run("Handles empty changes gracefully", func(t *testing.T) {
		testCache := cache.NewTestResultCache()

		staleTests := testCache.GetStaleTests([]*cache.FileChange{})
		if len(staleTests) != 0 {
			t.Errorf("Expected no stale tests for empty changes, got %v", staleTests)
		}

		shouldRun, tests := testCache.ShouldRunTests([]*cache.FileChange{})
		if shouldRun {
			t.Error("Expected ShouldRunTests to return false for empty changes")
		}
		if len(tests) != 0 {
			t.Errorf("Expected no tests for empty changes, got %v", tests)
		}
	})

	t.Run("Handles non-existent files gracefully", func(t *testing.T) {
		testCache := cache.NewTestResultCache()

		// Try to analyze a non-existent file
		_, err := testCache.AnalyzeChange("/non/existent/file.go")
		if err == nil {
			t.Error("Expected error when analyzing non-existent file")
		}
	})
}
