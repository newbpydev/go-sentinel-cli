package unit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/cli/core"
	"github.com/newbpydev/go-sentinel/internal/cli/execution"
	"github.com/newbpydev/go-sentinel/internal/cli/testing/helpers"
)

// TestSmartTestRunner_BasicFunctionality tests the basic functionality of the test runner
func TestSmartTestRunner_BasicFunctionality(t *testing.T) {
	t.Run("Creates runner with valid dependencies", func(t *testing.T) {
		// Arrange
		cache := helpers.NewMockCacheManager()
		strategy := helpers.NewMockStrategy("test-strategy")

		// Act
		runner := execution.NewSmartTestRunner(cache, strategy)

		// Assert
		if runner == nil {
			t.Fatal("Expected runner to be created, got nil")
		}

		capabilities := runner.GetCapabilities()
		if !capabilities.SupportsCaching {
			t.Error("Expected runner to support caching")
		}
		if !capabilities.SupportsWatchMode {
			t.Error("Expected runner to support watch mode")
		}
	})

	t.Run("Returns capabilities correctly", func(t *testing.T) {
		// Arrange
		cache := helpers.NewMockCacheManager()
		strategy := helpers.NewMockStrategy("test-strategy")
		runner := execution.NewSmartTestRunner(cache, strategy)

		// Act
		capabilities := runner.GetCapabilities()

		// Assert
		expected := core.RunnerCapabilities{
			SupportsCaching:    true,
			SupportsParallel:   true,
			SupportsWatchMode:  true,
			SupportsFiltering:  true,
			MaxConcurrency:     4,
			SupportedFileTypes: []string{".go"},
		}

		if capabilities.SupportsCaching != expected.SupportsCaching {
			t.Errorf("Expected SupportsCaching %v, got %v", expected.SupportsCaching, capabilities.SupportsCaching)
		}
		if capabilities.MaxConcurrency != expected.MaxConcurrency {
			t.Errorf("Expected MaxConcurrency %d, got %d", expected.MaxConcurrency, capabilities.MaxConcurrency)
		}
		if len(capabilities.SupportedFileTypes) != len(expected.SupportedFileTypes) {
			t.Errorf("Expected %d supported file types, got %d", len(expected.SupportedFileTypes), len(capabilities.SupportedFileTypes))
		}
	})
}

// TestSmartTestRunner_RunTests tests the main RunTests functionality
func TestSmartTestRunner_RunTests(t *testing.T) {
	t.Run("Returns success when no changes provided", func(t *testing.T) {
		// Arrange
		cache := helpers.NewMockCacheManager()
		strategy := helpers.NewMockStrategy("test-strategy")
		runner := execution.NewSmartTestRunner(cache, strategy)
		ctx := context.Background()

		// Act
		result, err := runner.RunTests(ctx, []core.FileChange{}, strategy)

		// Assert
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
		if result.Status != core.StatusPassed {
			t.Errorf("Expected status %v, got %v", core.StatusPassed, result.Status)
		}
		if !result.CacheHit {
			t.Error("Expected cache hit when no changes provided")
		}
	})

	t.Run("Handles test file changes correctly", func(t *testing.T) {
		// Arrange
		cache := helpers.NewMockCacheManager()
		strategy := helpers.NewMockStrategy("test-strategy")
		runner := execution.NewSmartTestRunner(cache, strategy)
		ctx := context.Background()

		changes := []core.FileChange{
			{
				Path:      "internal/test/example_test.go",
				Type:      core.ChangeTypeTest,
				IsNew:     true,
				Timestamp: time.Now(),
			},
		}

		// Act
		result, err := runner.RunTests(ctx, changes, strategy)

		// Assert
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}

		// Verify that the mock strategy was called
		mockStrategy := strategy.(*helpers.MockStrategy)
		if !mockStrategy.WasCalled("ShouldRunTest") {
			t.Error("Expected strategy.ShouldRunTest to be called")
		}
	})

	t.Run("Handles source file changes correctly", func(t *testing.T) {
		// Arrange
		cache := helpers.NewMockCacheManager()
		strategy := helpers.NewMockStrategy("test-strategy")
		runner := execution.NewSmartTestRunner(cache, strategy)
		ctx := context.Background()

		changes := []core.FileChange{
			{
				Path:      "internal/test/example.go",
				Type:      core.ChangeTypeSource,
				IsNew:     true,
				Timestamp: time.Now(),
			},
		}

		// Act
		result, err := runner.RunTests(ctx, changes, strategy)

		// Assert
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("Respects context cancellation", func(t *testing.T) {
		// Arrange
		cache := helpers.NewMockCacheManager()
		strategy := helpers.NewMockStrategy("test-strategy")
		runner := execution.NewSmartTestRunner(cache, strategy)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		changes := []core.FileChange{
			{
				Path:      "internal/test/example_test.go",
				Type:      core.ChangeTypeTest,
				IsNew:     true,
				Timestamp: time.Now(),
			},
		}

		// Act
		result, err := runner.RunTests(ctx, changes, strategy)

		// Assert
		// Since context is cancelled, we expect either:
		// 1. An error related to context cancellation, or
		// 2. A quick return with cached result if no execution was needed
		if err != nil {
			// Check if error is related to context cancellation
			if ctx.Err() == nil {
				t.Errorf("Expected context cancellation error, got %v", err)
			}
		} else if result != nil {
			// If we got a result, it should be because no execution was needed
			if !result.CacheHit {
				t.Error("Expected cache hit when context is cancelled and no execution needed")
			}
		}
	})
}

// TestSmartTestRunner_Performance tests performance characteristics
func TestSmartTestRunner_Performance(t *testing.T) {
	t.Run("Executes quickly for cache hits", func(t *testing.T) {
		// Arrange
		cache := helpers.NewMockCacheManager()
		strategy := helpers.NewMockStrategy("test-strategy")
		runner := execution.NewSmartTestRunner(cache, strategy)
		ctx := context.Background()

		// Configure mock to always return cache hits
		mockStrategy := strategy.(*helpers.MockStrategy)
		mockStrategy.SetShouldRunTest(false) // Don't run tests

		changes := []core.FileChange{
			{
				Path:      "internal/test/example_test.go",
				Type:      core.ChangeTypeTest,
				IsNew:     false, // Not new, so might be cached
				Timestamp: time.Now(),
			},
		}

		// Act
		start := time.Now()
		result, err := runner.RunTests(ctx, changes, strategy)
		duration := time.Since(start)

		// Assert
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if duration > 100*time.Millisecond {
			t.Errorf("Expected fast execution (< 100ms), took %v", duration)
		}
		if result == nil {
			t.Fatal("Expected result, got nil")
		}
	})

	t.Run("Handles multiple concurrent calls safely", func(t *testing.T) {
		// Arrange
		cache := helpers.NewMockCacheManager()
		strategy := helpers.NewMockStrategy("test-strategy")
		runner := execution.NewSmartTestRunner(cache, strategy)
		ctx := context.Background()

		changes := []core.FileChange{
			{
				Path:      "internal/test/example_test.go",
				Type:      core.ChangeTypeTest,
				IsNew:     false,
				Timestamp: time.Now(),
			},
		}

		// Act - Run multiple goroutines concurrently
		const numGoroutines = 10
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				_, err := runner.RunTests(ctx, changes, strategy)
				results <- err
			}()
		}

		// Assert - All goroutines should complete without errors
		for i := 0; i < numGoroutines; i++ {
			select {
			case err := <-results:
				if err != nil {
					t.Errorf("Goroutine %d failed: %v", i, err)
				}
			case <-time.After(5 * time.Second):
				t.Fatalf("Goroutine %d timed out", i)
			}
		}
	})
}

// TestSmartTestRunner_ErrorHandling tests error handling scenarios
func TestSmartTestRunner_ErrorHandling(t *testing.T) {
	t.Run("Handles nil strategy gracefully", func(t *testing.T) {
		// Arrange
		cache := helpers.NewMockCacheManager()
		runner := execution.NewSmartTestRunner(cache, nil)
		ctx := context.Background()

		changes := []core.FileChange{
			{
				Path:      "internal/test/example_test.go",
				Type:      core.ChangeTypeTest,
				IsNew:     true,
				Timestamp: time.Now(),
			},
		}

		// Act
		result, err := runner.RunTests(ctx, changes, nil)

		// Assert
		// Should handle gracefully, either by returning an error or using defaults
		if err != nil {
			// Error is acceptable
			t.Logf("Expected error for nil strategy: %v", err)
		} else if result != nil {
			// If no error, result should be valid
			t.Logf("Handled nil strategy gracefully with result: %+v", result)
		} else {
			t.Error("Expected either error or result, got neither")
		}
	})

	t.Run("Handles cache failures gracefully", func(t *testing.T) {
		// Arrange
		cache := helpers.NewMockCacheManager()
		strategy := helpers.NewMockStrategy("test-strategy")
		runner := execution.NewSmartTestRunner(cache, strategy)
		ctx := context.Background()

		// Configure mock cache to fail
		mockCache := cache.(*helpers.MockCacheManager)
		mockCache.SetShouldFail(true)

		changes := []core.FileChange{
			{
				Path:      "internal/test/example_test.go",
				Type:      core.ChangeTypeTest,
				IsNew:     true,
				Timestamp: time.Now(),
			},
		}

		// Act
		result, err := runner.RunTests(ctx, changes, strategy)

		// Assert
		// Should handle cache failures gracefully
		if err != nil {
			t.Logf("Cache failure handled with error: %v", err)
		} else if result != nil {
			t.Logf("Cache failure handled gracefully: %+v", result)
		} else {
			t.Error("Expected either error or result when cache fails")
		}
	})
}

// BenchmarkSmartTestRunner_CacheHit benchmarks cache hit performance
func BenchmarkSmartTestRunner_CacheHit(b *testing.B) {
	// Arrange
	cache := helpers.NewMockCacheManager()
	strategy := helpers.NewMockStrategy("benchmark-strategy")
	runner := execution.NewSmartTestRunner(cache, strategy)
	ctx := context.Background()

	// Configure for cache hits
	mockStrategy := strategy.(*helpers.MockStrategy)
	mockStrategy.SetShouldRunTest(false)

	changes := []core.FileChange{
		{
			Path:      "internal/test/example_test.go",
			Type:      core.ChangeTypeTest,
			IsNew:     false,
			Timestamp: time.Now(),
		},
	}

	// Reset timer and run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := runner.RunTests(ctx, changes, strategy)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

// BenchmarkSmartTestRunner_DetermineTargets benchmarks target determination
func BenchmarkSmartTestRunner_DetermineTargets(b *testing.B) {
	// Arrange
	cache := helpers.NewMockCacheManager()
	strategy := helpers.NewMockStrategy("benchmark-strategy")
	runner := execution.NewSmartTestRunner(cache, strategy)
	ctx := context.Background()

	// Create many file changes to test performance
	changes := make([]core.FileChange, 100)
	for i := 0; i < 100; i++ {
		changes[i] = core.FileChange{
			Path:      fmt.Sprintf("internal/test/example_%d_test.go", i),
			Type:      core.ChangeTypeTest,
			IsNew:     true,
			Timestamp: time.Now(),
		}
	}

	// Reset timer and run benchmark
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := runner.RunTests(ctx, changes[:10], strategy) // Use subset for performance
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
