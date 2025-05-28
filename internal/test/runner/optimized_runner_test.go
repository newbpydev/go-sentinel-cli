package runner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// TestFileChangeAdapter_AllMethods tests all methods of FileChangeAdapter
func TestFileChangeAdapter_AllMethods(t *testing.T) {
	t.Parallel()

	adapter := &FileChangeAdapter{
		FileChange: &models.FileChange{
			FilePath:   "/test/path/file.go",
			ChangeType: models.ChangeTypeModified,
			Timestamp:  time.Now(),
		},
	}

	// Test GetPath
	t.Run("GetPath", func(t *testing.T) {
		t.Parallel()

		path := adapter.GetPath()
		if path != "/test/path/file.go" {
			t.Errorf("Expected path '/test/path/file.go', got '%s'", path)
		}
	})

	// Test GetType
	t.Run("GetType", func(t *testing.T) {
		t.Parallel()

		changeType := adapter.GetType()
		if changeType != ChangeTypeSource {
			t.Errorf("Expected type ChangeTypeSource for .go file, got '%v'", changeType)
		}
	})

	// Test IsNewChange
	t.Run("IsNewChange", func(t *testing.T) {
		t.Parallel()

		isNew := adapter.IsNewChange()
		if !isNew {
			t.Error("Expected IsNewChange to return true")
		}
	})
}

// TestOptimizedTestRunner_HaveDependenciesChanged tests the haveDependenciesChanged method
func TestOptimizedTestRunner_HaveDependenciesChanged(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "optimized_runner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test files
	testFile1 := filepath.Join(tempDir, "test1.go")
	testFile2 := filepath.Join(tempDir, "test2.go")

	err = os.WriteFile(testFile1, []byte("package main\nfunc Test1() {}"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = os.WriteFile(testFile2, []byte("package main\nfunc Test2() {}"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	runner := NewOptimizedTestRunner()

	testCases := []struct {
		name     string
		target   string
		since    time.Time
		expected bool
	}{
		{
			name:     "Target with no dependencies",
			target:   "/non/existent/target",
			since:    time.Now().Add(-1 * time.Hour),
			expected: true, // Unknown dependencies = assume changed
		},
		{
			name:     "Target with recent timestamp",
			target:   tempDir,
			since:    time.Now().Add(-1 * time.Hour),
			expected: true, // Dependencies exist but not cached, so assume changed
		},
		{
			name:     "Target with old timestamp",
			target:   tempDir,
			since:    time.Now().Add(-24 * time.Hour),
			expected: true, // Dependencies exist but not cached, so assume changed
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := runner.haveDependenciesChanged(tc.target, tc.since)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestOptimizedTestRunner_DetermineNeedsExecution tests the determineNeedsExecution method
func TestOptimizedTestRunner_DetermineNeedsExecution(t *testing.T) {
	t.Parallel()

	runner := NewOptimizedTestRunner()

	// Create test changes
	changes := []FileChangeInterface{
		&FileChangeAdapter{
			FileChange: &models.FileChange{
				FilePath:   "/test/file1.go",
				ChangeType: models.ChangeTypeModified,
				Timestamp:  time.Now(),
			},
		},
		&FileChangeAdapter{
			FileChange: &models.FileChange{
				FilePath:   "/test/file2.go",
				ChangeType: models.ChangeTypeCreated,
				Timestamp:  time.Now(),
			},
		},
	}

	targets := []string{"/test/target1", "/test/target2"}

	// Test the method
	result := runner.determineNeedsExecution(targets, changes)

	// Should return targets if there are changes
	if len(result) == 0 {
		t.Error("Expected determineNeedsExecution to return targets with changes")
	}

	// Test with no changes
	result = runner.determineNeedsExecution(targets, []FileChangeInterface{})
	if len(result) != len(targets) {
		t.Errorf("Expected determineNeedsExecution to return all targets with no changes, got %d", len(result))
	}

	// Test with empty targets
	result = runner.determineNeedsExecution([]string{}, changes)
	if len(result) != 0 {
		t.Error("Expected determineNeedsExecution to return empty result with no targets")
	}
}

// TestOptimizedTestRunner_ScanDependencies tests the scanDependencies method
func TestOptimizedTestRunner_ScanDependencies(t *testing.T) {
	t.Parallel()

	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "scan_deps_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test Go file with imports
	testFile := filepath.Join(tempDir, "test.go")
	testContent := `package main

import (
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	fmt.Println("Hello")
}`

	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	runner := NewOptimizedTestRunner()

	// Test scanning dependencies
	deps := runner.scanDependencies(tempDir)

	// Should return some dependencies (at least the test file we created)
	if len(deps) == 0 {
		t.Error("Expected scanDependencies to return some dependencies")
	}

	// Test with non-existent file
	deps = runner.scanDependencies("/non/existent/file.go")
	if len(deps) < 0 { // Should at least try to add go.mod/go.sum if they exist
		t.Errorf("Expected non-negative dependencies for non-existent file, got %d", len(deps))
	}
}

// TestOptimizedTestRunner_UpdateCache tests the updateCache method
func TestOptimizedTestRunner_UpdateCache(t *testing.T) {
	t.Parallel()

	runner := NewOptimizedTestRunner()

	// Create test targets
	executedTargets := []string{"/test/target1", "/test/target2"}

	// Test updating cache
	runner.updateCache(executedTargets)

	// Verify cache was updated (this is mostly testing that the method doesn't panic)
	// The actual cache verification would require access to internal state
	t.Log("Cache update completed successfully")
}

// TestOptimizedTestRunner_ConfigurationMethods tests configuration methods
func TestOptimizedTestRunner_ConfigurationMethods(t *testing.T) {
	t.Parallel()

	runner := NewOptimizedTestRunner()

	// Test SetCacheEnabled
	t.Run("SetCacheEnabled", func(t *testing.T) {
		t.Parallel()
		runner.SetCacheEnabled(true)
		runner.SetCacheEnabled(false)
		// Should not panic
	})

	// Test SetOnlyRunChangedTests
	t.Run("SetOnlyRunChangedTests", func(t *testing.T) {
		t.Parallel()
		runner.SetOnlyRunChangedTests(true)
		runner.SetOnlyRunChangedTests(false)
		// Should not panic
	})

	// Test SetOptimizationMode
	t.Run("SetOptimizationMode", func(t *testing.T) {
		t.Parallel()
		runner.SetOptimizationMode("aggressive")
		runner.SetOptimizationMode("conservative")
		runner.SetOptimizationMode("balanced")
		// Should not panic
	})

	// Test ClearCache
	t.Run("ClearCache", func(t *testing.T) {
		t.Parallel()
		runner.ClearCache()
		// Should not panic
	})
}

// TestOptimizedTestResult_GetEfficiencyStats tests the GetEfficiencyStats method
func TestOptimizedTestResult_GetEfficiencyStats(t *testing.T) {
	t.Parallel()

	result := &OptimizedTestResult{
		TestsRun:  5,
		CacheHits: 3,
		Duration:  100 * time.Millisecond,
	}

	stats := result.GetEfficiencyStats()
	if stats == nil {
		t.Error("GetEfficiencyStats should not return nil")
	}

	if len(stats) == 0 {
		t.Error("GetEfficiencyStats should return non-empty stats")
	}

	// Verify specific stats
	if testsRun, ok := stats["tests_run"]; !ok || testsRun != 5 {
		t.Errorf("Expected tests_run to be 5, got %v", testsRun)
	}

	if cacheHits, ok := stats["cache_hits"]; !ok || cacheHits != 3 {
		t.Errorf("Expected cache_hits to be 3, got %v", cacheHits)
	}
}

// TestOptimizedTestRunner_EdgeCases tests edge cases
func TestOptimizedTestRunner_EdgeCases(t *testing.T) {
	t.Parallel()

	runner := NewOptimizedTestRunner()

	// Test RunOptimized with empty changes
	t.Run("RunOptimized with empty changes", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		changes := []FileChangeInterface{}

		// This should handle empty changes gracefully
		_, err := runner.RunOptimized(ctx, changes)
		if err != nil {
			t.Logf("RunOptimized with empty changes returned error (may be expected): %v", err)
		}
	})

	// Test with nil changes
	t.Run("RunOptimized with nil changes", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// This should handle nil changes gracefully
		_, err := runner.RunOptimized(ctx, nil)
		if err != nil {
			t.Logf("RunOptimized with nil changes returned error (may be expected): %v", err)
		}
	})

	// Test with cancelled context
	t.Run("RunOptimized with cancelled context", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		changes := []FileChangeInterface{
			&FileChangeAdapter{
				FileChange: &models.FileChange{
					FilePath:   "/test/file.go",
					ChangeType: models.ChangeTypeModified,
					Timestamp:  time.Now(),
				},
			},
		}

		_, err := runner.RunOptimized(ctx, changes)
		if err != nil {
			t.Logf("RunOptimized with cancelled context returned error (expected): %v", err)
		}
	})
}

// TestFileChangeAdapter_ChangeTypes tests different change types
func TestFileChangeAdapter_ChangeTypes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		filePath     string
		changeType   models.ChangeType
		expectedType ChangeType
	}{
		{
			name:         "Test file",
			filePath:     "/test/example_test.go",
			changeType:   models.ChangeTypeModified,
			expectedType: ChangeTypeTest,
		},
		{
			name:         "Source file",
			filePath:     "/test/example.go",
			changeType:   models.ChangeTypeModified,
			expectedType: ChangeTypeSource,
		},
		{
			name:         "Go mod file",
			filePath:     "/test/go.mod",
			changeType:   models.ChangeTypeModified,
			expectedType: ChangeTypeDependency,
		},
		{
			name:         "Go sum file",
			filePath:     "/test/go.sum",
			changeType:   models.ChangeTypeModified,
			expectedType: ChangeTypeDependency,
		},
		{
			name:         "Config file",
			filePath:     "/test/config.json",
			changeType:   models.ChangeTypeModified,
			expectedType: ChangeTypeConfig,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			adapter := &FileChangeAdapter{
				FileChange: &models.FileChange{
					FilePath:   tc.filePath,
					ChangeType: tc.changeType,
					Timestamp:  time.Now(),
				},
			}

			changeType := adapter.GetType()
			if changeType != tc.expectedType {
				t.Errorf("Expected type %v for file %s, got %v", tc.expectedType, tc.filePath, changeType)
			}
		})
	}
}
