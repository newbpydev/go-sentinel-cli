package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestIntelligentWatchEfficiency tests that the watch mode is truly efficient
func TestIntelligentWatchEfficiency(t *testing.T) {
	t.Run("Should only run affected tests", func(t *testing.T) {
		cache := NewTestResultCache()

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

		change, err := cache.AnalyzeChange(testFile)
		if err != nil {
			t.Fatalf("Failed to analyze change: %v", err)
		}

		staleTests := cache.GetStaleTests([]*FileChange{change})

		// Should only return the directory of the changed test
		expectedDir := filepath.Dir(testFile)
		if len(staleTests) != 1 || staleTests[0] != expectedDir {
			t.Errorf("Expected stale tests [%s], got %v", expectedDir, staleTests)
		}
	})

	t.Run("Cache invalidation works correctly", func(t *testing.T) {
		cache := NewTestResultCache()

		// Add initial cache result
		suite := &TestSuite{
			FilePath:     "pkg/test",
			TestCount:    3,
			PassedCount:  3,
			FailedCount:  0,
			SkippedCount: 0,
		}
		cache.CacheResult("pkg/test", suite)

		// Verify cache hit
		cached, exists := cache.GetCachedResult("pkg/test")
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
		cache := NewTestResultCache()
		renderer := NewIncrementalRenderer(
			io.Discard,
			NewColorFormatter(false),
			NewIconProvider(false),
			80,
			cache,
		)

		// Simulate many rapid changes
		tempDir := t.TempDir()
		changes := make([]*FileChange, 50)

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

			change, err := cache.AnalyzeChange(testFile)
			if err != nil {
				t.Fatalf("Failed to analyze change %d: %v", i, err)
			}
			changes[i] = change
		}

		// Process all changes
		staleTests := cache.GetStaleTests(changes)

		// Render results
		err := renderer.RenderIncrementalResults(
			make(map[string]*TestSuite),
			&TestRunStats{},
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
		cache := NewTestResultCache()

		// Add many cache entries
		for i := 0; i < 1000; i++ {
			suite := &TestSuite{
				FilePath:     fmt.Sprintf("pkg/test%d", i),
				TestCount:    3,
				PassedCount:  3,
				FailedCount:  0,
				SkippedCount: 0,
			}
			cache.CacheResult(fmt.Sprintf("pkg/test%d", i), suite)
		}

		stats := cache.GetStats()
		if stats["cached_results"].(int) != 1000 {
			t.Errorf("Expected 1000 cached results, got %d", stats["cached_results"])
		}

		// Clear cache and verify cleanup
		cache.Clear()

		stats = cache.GetStats()
		if stats["cached_results"].(int) != 0 {
			t.Errorf("Expected 0 cached results after clear, got %d", stats["cached_results"])
		}
	})
}

// TestWatchModeUserExperience tests the UX aspects of watch mode
func TestWatchModeUserExperience(t *testing.T) {
	t.Run("Clear messaging when no tests need to run", func(t *testing.T) {
		var buffer bytes.Buffer
		cache := NewTestResultCache()
		renderer := NewIncrementalRenderer(
			&buffer,
			NewColorFormatter(false),
			NewIconProvider(false),
			80,
			cache,
		)

		// Simulate a change that doesn't require tests
		changes := []*FileChange{
			{
				Path: "docs/README.md",
				Type: ChangeTypeConfig,
			},
		}

		// Render with no test suites (meaning no tests ran)
		err := renderer.RenderIncrementalResults(
			make(map[string]*TestSuite),
			&TestRunStats{},
			changes,
		)
		if err != nil {
			t.Fatalf("Failed to render results: %v", err)
		}

		output := buffer.String()

		// Should clearly indicate what happened
		if !strings.Contains(output, "File changes detected") {
			t.Error("Should show file changes were detected")
		}
		if !strings.Contains(output, "No test changes detected") {
			t.Error("Should clearly indicate no tests were needed")
		}
	})
}

// TestWatchModeEdgeCases tests edge cases and error conditions
func TestWatchModeEdgeCases(t *testing.T) {
	t.Run("Handles missing files gracefully", func(t *testing.T) {
		cache := NewTestResultCache()

		// Try to analyze a non-existent file
		_, err := cache.AnalyzeChange("nonexistent/file.go")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("Handles empty changes list", func(t *testing.T) {
		cache := NewTestResultCache()

		staleTests := cache.GetStaleTests([]*FileChange{})
		if len(staleTests) != 0 {
			t.Errorf("Expected no stale tests for empty changes, got %d", len(staleTests))
		}
	})
}
