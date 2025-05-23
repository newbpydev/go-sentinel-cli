package cli

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestOptimizedTestRunner(t *testing.T) {
	t.Run("Basic functionality", func(t *testing.T) {
		runner := NewOptimizedTestRunner()

		if runner == nil {
			t.Fatal("NewOptimizedTestRunner returned nil")
		}

		if runner.cache == nil {
			t.Error("Cache should be initialized")
		}

		if !runner.enableGoCache {
			t.Error("Go cache should be enabled by default")
		}

		if !runner.onlyRunChangedTests {
			t.Error("Only run changed tests should be enabled by default")
		}
	})

	t.Run("Determines correct test targets", func(t *testing.T) {
		runner := NewOptimizedTestRunner()

		// Test with different change types
		testCases := []struct {
			name     string
			changes  []*FileChange
			expected int // minimum expected targets
		}{
			{
				name: "Single test file change",
				changes: []*FileChange{
					{Path: "internal/cli/example_test.go", Type: ChangeTypeTest},
				},
				expected: 1,
			},
			{
				name: "Source file change",
				changes: []*FileChange{
					{Path: "internal/cli/example.go", Type: ChangeTypeSource},
				},
				expected: 1,
			},
			{
				name: "Config change affects all",
				changes: []*FileChange{
					{Path: "go.mod", Type: ChangeTypeDependency},
				},
				expected: 1,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				targets := runner.determineTestTargets(tc.changes)

				if len(targets) < tc.expected {
					t.Errorf("Expected at least %d targets, got %d for %s",
						tc.expected, len(targets), tc.name)
				}

				t.Logf("Test targets for %s: %v", tc.name, targets)
			})
		}
	})

	t.Run("Builds optimized commands", func(t *testing.T) {
		runner := NewOptimizedTestRunner()

		testCases := []struct {
			name     string
			targets  []string
			expected []string // expected command parts
		}{
			{
				name:     "Single package",
				targets:  []string{"./internal/cli"},
				expected: []string{"test", "-failfast", "./internal/cli"},
			},
			{
				name:     "Multiple packages",
				targets:  []string{"./internal/cli", "./cmd/go-sentinel-cli"},
				expected: []string{"test", "-failfast"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				cmd := runner.buildOptimizedCommand(tc.targets)

				// Check that expected parts are present
				for _, expectedPart := range tc.expected {
					found := false
					for _, cmdPart := range cmd {
						if cmdPart == expectedPart {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected command part '%s' not found in: %v", expectedPart, cmd)
					}
				}

				t.Logf("Command for %s: %v", tc.name, cmd)
			})
		}
	})

	t.Run("Cache management works", func(t *testing.T) {
		runner := NewOptimizedTestRunner()

		// Test cache operations
		runner.updateCache([]string{"./internal/cli"})

		// Test that we can clear the cache without errors
		runner.ClearCache()

		// Verify cache clearing worked by checking internal state
		r := runner.cache
		r.mu.RLock()
		defer r.mu.RUnlock()

		if len(r.packageCache) > 0 {
			t.Error("Package cache should be empty after clearing")
		}

		if len(r.testDependencies) > 0 {
			t.Error("Test dependencies should be empty after clearing")
		}

		if len(r.fileModTimes) > 0 {
			t.Error("File mod times should be empty after clearing")
		}
	})

	t.Run("Efficiency stats calculation", func(t *testing.T) {
		result := &OptimizedTestResult{
			TestsRun:  2,
			CacheHits: 8,
			Duration:  100 * time.Millisecond,
		}

		stats := result.GetEfficiencyStats()

		expectedCacheHitRate := 80.0 // 8/(2+8) * 100
		if rate := stats["cache_hit_rate"].(float64); rate != expectedCacheHitRate {
			t.Errorf("Expected cache hit rate %.1f%%, got %.1f%%", expectedCacheHitRate, rate)
		}

		if total := stats["total_targets"].(int); total != 10 {
			t.Errorf("Expected total targets 10, got %d", total)
		}

		if duration := stats["duration_ms"].(int64); duration != 100 {
			t.Errorf("Expected duration 100ms, got %dms", duration)
		}

		t.Logf("Efficiency stats: %+v", stats)
	})

	t.Run("Optimization modes work", func(t *testing.T) {
		runner := NewOptimizedTestRunner()

		// Test different optimization modes
		modes := []string{"aggressive", "conservative", "disabled"}

		for _, mode := range modes {
			t.Run(mode, func(t *testing.T) {
				runner.SetOptimizationMode(mode)

				switch mode {
				case "aggressive":
					if !runner.enableGoCache || !runner.onlyRunChangedTests {
						t.Error("Aggressive mode should enable both cache and only changed tests")
					}
				case "conservative":
					if !runner.enableGoCache || runner.onlyRunChangedTests {
						t.Error("Conservative mode should enable cache but not only changed tests")
					}
				case "disabled":
					if runner.enableGoCache || runner.onlyRunChangedTests {
						t.Error("Disabled mode should disable both cache and only changed tests")
					}
				}

				t.Logf("Mode %s: enableGoCache=%t, onlyRunChangedTests=%t",
					mode, runner.enableGoCache, runner.onlyRunChangedTests)
			})
		}
	})
}

func TestOptimizedTestRunnerIntegration(t *testing.T) {
	t.Run("Real file system integration", func(t *testing.T) {
		runner := NewOptimizedTestRunner()

		// Use actual project files for realistic testing
		changes := []*FileChange{
			{
				Path:  "internal/cli/optimized_test_runner_test.go", // This file!
				Type:  ChangeTypeTest,
				IsNew: true, // Simulate a file change
			},
		}

		result, err := runner.RunOptimized(context.Background(), changes)
		if err != nil {
			t.Fatalf("Integration test failed: %v", err)
		}

		stats := result.GetEfficiencyStats()

		t.Logf("Integration test results:")
		t.Logf("  Tests run: %d", result.TestsRun)
		t.Logf("  Cache hits: %d", result.CacheHits)
		t.Logf("  Duration: %v", result.Duration)
		t.Logf("  Cache hit rate: %.1f%%", stats["cache_hit_rate"].(float64))
		t.Logf("  Exit code: %d", result.ExitCode)

		if result.Duration > 10*time.Second {
			t.Logf("Warning: Test took longer than expected (%v), but this might be normal for integration tests", result.Duration)
		}
	})

	t.Run("Test execution determination", func(t *testing.T) {
		runner := NewOptimizedTestRunner()

		// Test changes that should trigger execution
		changes := []*FileChange{
			{
				Path:  "internal/cli/test_cache.go",
				Type:  ChangeTypeSource,
				IsNew: true, // File is new/changed
			},
		}

		targets := []string{"./internal/cli"}
		needsExecution := runner.determineNeedsExecution(targets, changes)

		// Should need execution since file is new
		if len(needsExecution) == 0 {
			t.Error("Should need execution when files are new/changed")
		}

		t.Logf("Needs execution: %v", needsExecution)

		// Test with no changes
		noChanges := []*FileChange{
			{
				Path:  "internal/cli/test_cache.go",
				Type:  ChangeTypeSource,
				IsNew: false, // File is not new
			},
		}

		needsExecution2 := runner.determineNeedsExecution(targets, noChanges)
		t.Logf("Needs execution (no changes): %v", needsExecution2)
	})

	t.Run("Related test files discovery", func(t *testing.T) {
		runner := NewOptimizedTestRunner()

		// Test finding related test files
		sourceFile := "internal/cli/test_cache.go"
		relatedTests := runner.findRelatedTestFiles(sourceFile)

		t.Logf("Related tests for %s: %v", sourceFile, relatedTests)

		// Should find at least the direct test file if it exists
		expectedTestFile := "internal/cli/test_cache_test.go"
		if _, err := os.Stat(expectedTestFile); err == nil {
			found := false
			for _, test := range relatedTests {
				if test == expectedTestFile {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected to find %s in related tests", expectedTestFile)
			}
		}
	})
}
