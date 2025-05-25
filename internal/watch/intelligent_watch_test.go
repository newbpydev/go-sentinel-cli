package watch

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/cache"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

func TestTestResultCache(t *testing.T) {
	t.Run("Cache basic operations", func(t *testing.T) {
		testCache := cache.NewTestResultCache()

		// Test empty cache
		_, exists := testCache.GetCachedResult("test/path")
		if exists {
			t.Error("Expected cache miss for non-existent path")
		}

		// Add a result
		suite := &models.TestSuite{
			FilePath:     "test/path",
			TestCount:    1,
			PassedCount:  1,
			FailedCount:  0,
			SkippedCount: 0,
		}

		testCache.CacheResult("test/path", suite)

		// Retrieve result
		cached, exists := testCache.GetCachedResult("test/path")
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

		testCache := cache.NewTestResultCache()

		// Analyze the change
		change, err := testCache.AnalyzeChange(testFile)
		if err != nil {
			t.Fatalf("Failed to analyze change: %v", err)
		}

		if change.Type != cache.ChangeTypeTest {
			t.Errorf("Expected ChangeTypeTest, got %d", change.Type)
		}

		if !change.IsNew {
			t.Error("Expected IsNew to be true for first analysis")
		}

		if len(change.AffectedTests) == 0 {
			t.Error("Expected affected tests to be populated")
		}
	})

	t.Run("Cache invalidation", func(t *testing.T) {
		testCache := cache.NewTestResultCache()

		// Create a test suite
		suite := &models.TestSuite{
			FilePath:     "test/path",
			TestCount:    1,
			PassedCount:  1,
			FailedCount:  0,
			SkippedCount: 0,
			Duration:     time.Millisecond * 100,
		}

		// Cache the result
		testCache.CacheResult("test/path", suite)

		// Verify it's cached
		cached, exists := testCache.GetCachedResult("test/path")
		if !exists {
			t.Error("Expected cache hit after storing result")
		}

		if cached.Suite.TestCount != 1 {
			t.Errorf("Expected TestCount 1, got %d", cached.Suite.TestCount)
		}

		// Clear the cache
		testCache.Clear()

		// Verify it's cleared
		_, exists = testCache.GetCachedResult("test/path")
		if exists {
			t.Error("Expected cache miss after clearing")
		}
	})

	t.Run("Cache statistics", func(t *testing.T) {
		testCache := cache.NewTestResultCache()

		// Get initial stats
		stats := testCache.GetStats()
		if stats == nil {
			t.Error("Expected stats to be non-nil")
		}

		// Add some results
		suite1 := &models.TestSuite{
			FilePath:     "test/path1",
			TestCount:    2,
			PassedCount:  2,
			FailedCount:  0,
			SkippedCount: 0,
		}

		suite2 := &models.TestSuite{
			FilePath:     "test/path2",
			TestCount:    1,
			PassedCount:  0,
			FailedCount:  1,
			SkippedCount: 0,
		}

		testCache.CacheResult("test/path1", suite1)
		testCache.CacheResult("test/path2", suite2)

		// Get updated stats
		stats = testCache.GetStats()
		if stats == nil {
			t.Error("Expected stats to be non-nil after adding results")
		}

		// Stats should contain information about cached results
		if len(stats) == 0 {
			t.Error("Expected stats to contain information about cached results")
		}
	})
}
