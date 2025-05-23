package helpers

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/newbpydev/go-sentinel/internal/cli/core"
)

// MockCacheManager provides a mock implementation of core.CacheManager for testing
type MockCacheManager struct {
	results    map[string]*core.CachedResult
	shouldFail bool
	callLog    map[string]int
	mu         sync.RWMutex
}

// NewMockCacheManager creates a new mock cache manager
func NewMockCacheManager() core.CacheManager {
	return &MockCacheManager{
		results: make(map[string]*core.CachedResult),
		callLog: make(map[string]int),
		mu:      sync.RWMutex{},
	}
}

// GetCachedResult implements core.CacheManager
func (m *MockCacheManager) GetCachedResult(target core.TestTarget) (*core.CachedResult, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callLog["GetCachedResult"]++

	if m.shouldFail {
		return nil, false
	}

	key := target.Path
	result, exists := m.results[key]
	return result, exists
}

// StoreResult implements core.CacheManager
func (m *MockCacheManager) StoreResult(target core.TestTarget, result *core.TestResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callLog["StoreResult"]++

	if m.shouldFail {
		return
	}

	key := target.Path
	m.results[key] = &core.CachedResult{
		Result:    result,
		CacheTime: time.Now(),
		IsValid:   true,
	}
}

// InvalidateCache implements core.CacheManager
func (m *MockCacheManager) InvalidateCache(changes []core.FileChange) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callLog["InvalidateCache"]++

	if m.shouldFail {
		return
	}

	// Simple invalidation - mark all as invalid
	for _, cached := range m.results {
		cached.IsValid = false
	}
}

// Clear implements core.CacheManager
func (m *MockCacheManager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callLog["Clear"]++
	m.results = make(map[string]*core.CachedResult)
}

// GetStats implements core.CacheManager
func (m *MockCacheManager) GetStats() core.CacheStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.callLog["GetStats"]++

	return core.CacheStats{
		TotalEntries:   len(m.results),
		ValidEntries:   len(m.results),
		InvalidEntries: 0,
		HitRate:        100.0,
		MemoryUsage:    int64(len(m.results) * 1024),
		OldestEntry:    time.Now(),
		NewestEntry:    time.Now(),
	}
}

// SetShouldFail configures the mock to fail operations
func (m *MockCacheManager) SetShouldFail(shouldFail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
}

// GetCallCount returns the number of times a method was called
func (m *MockCacheManager) GetCallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callLog[method]
}

// MockStrategy provides a mock implementation of core.ExecutionStrategy for testing
type MockStrategy struct {
	name          string
	shouldRunTest bool
	callLog       map[string]int
	mu            sync.RWMutex
}

// NewMockStrategy creates a new mock execution strategy
func NewMockStrategy(name string) core.ExecutionStrategy {
	return &MockStrategy{
		name:          name,
		shouldRunTest: true, // Default to running tests
		callLog:       make(map[string]int),
		mu:            sync.RWMutex{},
	}
}

// ShouldRunTest implements core.ExecutionStrategy
func (m *MockStrategy) ShouldRunTest(target core.TestTarget, cache core.CacheManager) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callLog["ShouldRunTest"]++
	return m.shouldRunTest
}

// GetExecutionOrder implements core.ExecutionStrategy
func (m *MockStrategy) GetExecutionOrder(targets []core.TestTarget) []core.TestTarget {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callLog["GetExecutionOrder"]++

	// Return targets as-is for simplicity
	return targets
}

// GetName implements core.ExecutionStrategy
func (m *MockStrategy) GetName() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	m.callLog["GetName"]++
	return m.name
}

// SetShouldRunTest configures whether tests should run
func (m *MockStrategy) SetShouldRunTest(shouldRun bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldRunTest = shouldRun
}

// WasCalled returns true if the specified method was called
func (m *MockStrategy) WasCalled(method string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callLog[method] > 0
}

// GetCallCount returns the number of times a method was called
func (m *MockStrategy) GetCallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callLog[method]
}

// MockTestRunner provides a mock implementation of core.TestRunner for testing
type MockTestRunner struct {
	shouldFail bool
	result     *core.TestResult
	callLog    map[string]int
	mu         sync.RWMutex
}

// NewMockTestRunner creates a new mock test runner
func NewMockTestRunner() core.TestRunner {
	return &MockTestRunner{
		result: &core.TestResult{
			Status:   core.StatusPassed,
			Output:   "Mock test output",
			Duration: 100 * time.Millisecond,
			CacheHit: false,
		},
		callLog: make(map[string]int),
		mu:      sync.RWMutex{},
	}
}

// RunTests implements core.TestRunner
func (m *MockTestRunner) RunTests(ctx context.Context, changes []core.FileChange, strategy core.ExecutionStrategy) (*core.TestResult, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callLog["RunTests"]++

	if m.shouldFail {
		return nil, errors.New("mock test runner failure")
	}

	return m.result, nil
}

// GetCapabilities implements core.TestRunner
func (m *MockTestRunner) GetCapabilities() core.RunnerCapabilities {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callLog["GetCapabilities"]++

	return core.RunnerCapabilities{
		SupportsCaching:    true,
		SupportsParallel:   true,
		SupportsWatchMode:  true,
		SupportsFiltering:  true,
		MaxConcurrency:     4,
		SupportedFileTypes: []string{".go"},
	}
}

// SetShouldFail configures the mock to fail operations
func (m *MockTestRunner) SetShouldFail(shouldFail bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldFail = shouldFail
}

// SetResult configures the result to return
func (m *MockTestRunner) SetResult(result *core.TestResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.result = result
}

// WasCalled returns true if the specified method was called
func (m *MockTestRunner) WasCalled(method string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callLog[method] > 0
}

// GetCallCount returns the number of times a method was called
func (m *MockTestRunner) GetCallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callLog[method]
}
