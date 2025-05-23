package execution

import (
	"sort"
	"time"

	"github.com/newbpydev/go-sentinel/internal/cli/core"
)

// AggressiveStrategy maximizes cache usage for fastest feedback
type AggressiveStrategy struct {
	name string
}

// NewAggressiveStrategy creates a new aggressive caching strategy
func NewAggressiveStrategy() *AggressiveStrategy {
	return &AggressiveStrategy{
		name: "aggressive",
	}
}

// ShouldRunTest determines if a test should be executed (aggressive caching)
func (s *AggressiveStrategy) ShouldRunTest(target core.TestTarget, cache core.CacheManager) bool {
	// Check if we have a recent cached result
	if cached, exists := cache.GetCachedResult(target); exists {
		// Use cache if it's less than 5 minutes old and was successful
		if time.Since(cached.CacheTime) < 5*time.Minute &&
			cached.Result.Status == core.StatusPassed {
			return false
		}
	}
	return true
}

// GetExecutionOrder determines the order of test execution (fastest first)
func (s *AggressiveStrategy) GetExecutionOrder(targets []core.TestTarget) []core.TestTarget {
	// Sort by priority (higher first) then by estimated duration (shorter first)
	sorted := make([]core.TestTarget, len(targets))
	copy(sorted, targets)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Priority != sorted[j].Priority {
			return sorted[i].Priority > sorted[j].Priority // Higher priority first
		}
		return sorted[i].EstimatedDuration < sorted[j].EstimatedDuration // Faster first
	})

	return sorted
}

// GetName returns the strategy name
func (s *AggressiveStrategy) GetName() string {
	return s.name
}

// ConservativeStrategy balances cache usage with accuracy
type ConservativeStrategy struct {
	name string
}

// NewConservativeStrategy creates a new conservative strategy
func NewConservativeStrategy() *ConservativeStrategy {
	return &ConservativeStrategy{
		name: "conservative",
	}
}

// ShouldRunTest determines if a test should be executed (conservative approach)
func (s *ConservativeStrategy) ShouldRunTest(target core.TestTarget, cache core.CacheManager) bool {
	// Check if we have a very recent cached result
	if cached, exists := cache.GetCachedResult(target); exists {
		// Only use cache if it's less than 1 minute old
		if time.Since(cached.CacheTime) < 1*time.Minute &&
			cached.Result.Status == core.StatusPassed {
			return false
		}
	}
	return true
}

// GetExecutionOrder determines the order of test execution (balanced approach)
func (s *ConservativeStrategy) GetExecutionOrder(targets []core.TestTarget) []core.TestTarget {
	// Sort by priority first, then by type (packages before recursive)
	sorted := make([]core.TestTarget, len(targets))
	copy(sorted, targets)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Priority != sorted[j].Priority {
			return sorted[i].Priority > sorted[j].Priority
		}
		// Prefer package-level tests over recursive tests
		if sorted[i].Type == "package" && sorted[j].Type == "recursive" {
			return true
		}
		if sorted[i].Type == "recursive" && sorted[j].Type == "package" {
			return false
		}
		return sorted[i].EstimatedDuration < sorted[j].EstimatedDuration
	})

	return sorted
}

// GetName returns the strategy name
func (s *ConservativeStrategy) GetName() string {
	return s.name
}

// NoCache strategy always runs tests without using cache
type NoCacheStrategy struct {
	name string
}

// NewNoCacheStrategy creates a new no-cache strategy
func NewNoCacheStrategy() *NoCacheStrategy {
	return &NoCacheStrategy{
		name: "no-cache",
	}
}

// ShouldRunTest always returns true (no caching)
func (s *NoCacheStrategy) ShouldRunTest(target core.TestTarget, cache core.CacheManager) bool {
	return true // Always run tests
}

// GetExecutionOrder determines the order of test execution (priority-based)
func (s *NoCacheStrategy) GetExecutionOrder(targets []core.TestTarget) []core.TestTarget {
	// Sort by priority only
	sorted := make([]core.TestTarget, len(targets))
	copy(sorted, targets)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Priority > sorted[j].Priority
	})

	return sorted
}

// GetName returns the strategy name
func (s *NoCacheStrategy) GetName() string {
	return s.name
}

// WatchModeStrategy is optimized for continuous testing in watch mode
type WatchModeStrategy struct {
	name              string
	lastExecutionTime time.Time
}

// NewWatchModeStrategy creates a new watch mode strategy
func NewWatchModeStrategy() *WatchModeStrategy {
	return &WatchModeStrategy{
		name:              "watch-mode",
		lastExecutionTime: time.Now(),
	}
}

// ShouldRunTest determines if a test should be executed (watch mode optimized)
func (s *WatchModeStrategy) ShouldRunTest(target core.TestTarget, cache core.CacheManager) bool {
	// In watch mode, be more aggressive about caching to provide fast feedback
	if cached, exists := cache.GetCachedResult(target); exists {
		// Use cache if it's less than 2 minutes old and was successful
		if time.Since(cached.CacheTime) < 2*time.Minute &&
			cached.Result.Status == core.StatusPassed {
			return false
		}

		// Also check if nothing has changed since last execution
		if cached.CacheTime.After(s.lastExecutionTime) {
			return false
		}
	}

	s.lastExecutionTime = time.Now()
	return true
}

// GetExecutionOrder determines the order of test execution (watch mode optimized)
func (s *WatchModeStrategy) GetExecutionOrder(targets []core.TestTarget) []core.TestTarget {
	// In watch mode, prioritize fast tests for quick feedback
	sorted := make([]core.TestTarget, len(targets))
	copy(sorted, targets)

	sort.Slice(sorted, func(i, j int) bool {
		// Always prioritize test file changes in watch mode
		if sorted[i].Priority == 1 && sorted[j].Priority != 1 {
			return true
		}
		if sorted[i].Priority != 1 && sorted[j].Priority == 1 {
			return false
		}
		// Then sort by estimated duration (faster first)
		return sorted[i].EstimatedDuration < sorted[j].EstimatedDuration
	})

	return sorted
}

// GetName returns the strategy name
func (s *WatchModeStrategy) GetName() string {
	return s.name
}

// StrategyFactory creates execution strategies based on configuration
type StrategyFactory struct{}

// NewStrategyFactory creates a new strategy factory
func NewStrategyFactory() *StrategyFactory {
	return &StrategyFactory{}
}

// CreateStrategy creates an execution strategy based on name
func (f *StrategyFactory) CreateStrategy(name string) core.ExecutionStrategy {
	switch name {
	case "aggressive":
		return NewAggressiveStrategy()
	case "conservative":
		return NewConservativeStrategy()
	case "no-cache", "disabled":
		return NewNoCacheStrategy()
	case "watch-mode", "watch":
		return NewWatchModeStrategy()
	default:
		// Default to aggressive strategy
		return NewAggressiveStrategy()
	}
}

// GetAvailableStrategies returns a list of available strategy names
func (f *StrategyFactory) GetAvailableStrategies() []string {
	return []string{
		"aggressive",
		"conservative",
		"no-cache",
		"watch-mode",
	}
}

// GetStrategyDescription returns a description of the strategy
func (f *StrategyFactory) GetStrategyDescription(name string) string {
	switch name {
	case "aggressive":
		return "Maximizes cache usage for fastest feedback (5-minute cache window)"
	case "conservative":
		return "Balances cache usage with accuracy (1-minute cache window)"
	case "no-cache":
		return "Always runs tests without using cache"
	case "watch-mode":
		return "Optimized for continuous testing with smart caching (2-minute window)"
	default:
		return "Unknown strategy"
	}
}
