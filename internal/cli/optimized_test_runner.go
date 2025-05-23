package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// OptimizedTestRunner leverages Go's built-in test caching for maximum efficiency
type OptimizedTestRunner struct {
	cache               *SmartTestCache
	mu                  sync.RWMutex
	enableGoCache       bool
	onlyRunChangedTests bool
}

// SmartTestCache tracks file dependencies and test relationships
type SmartTestCache struct {
	testDependencies map[string][]string  // test file -> source files it depends on
	packageCache     map[string]time.Time // package -> last successful run time
	fileModTimes     map[string]time.Time // file -> last modification time
	mu               sync.RWMutex
}

// NewOptimizedTestRunner creates a new optimized test runner
func NewOptimizedTestRunner() *OptimizedTestRunner {
	return &OptimizedTestRunner{
		cache:               NewSmartTestCache(),
		enableGoCache:       true,
		onlyRunChangedTests: true,
	}
}

// NewSmartTestCache creates a new smart test cache
func NewSmartTestCache() *SmartTestCache {
	return &SmartTestCache{
		testDependencies: make(map[string][]string),
		packageCache:     make(map[string]time.Time),
		fileModTimes:     make(map[string]time.Time),
	}
}

// RunOptimized runs tests with maximum efficiency using Go's caching
func (r *OptimizedTestRunner) RunOptimized(ctx context.Context, changes []*FileChange) (*OptimizedTestResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	result := &OptimizedTestResult{
		StartTime: time.Now(),
	}

	// Step 1: Determine what actually needs to be tested
	testTargets := r.determineTestTargets(changes)

	if len(testTargets) == 0 {
		result.CacheHits = len(changes)
		result.TestsRun = 0
		result.Message = "No test targets identified for the detected changes"
		result.Duration = time.Since(result.StartTime)
		return result, nil
	}

	// Step 2: For any file changes detected, we should run tests (be less aggressive about caching)
	needsExecution := r.determineNeedsExecution(testTargets, changes)

	result.CacheHits = len(testTargets) - len(needsExecution)
	result.TestsRun = len(needsExecution)

	if len(needsExecution) == 0 {
		result.Message = "All test targets already satisfied by cache"
		result.Duration = time.Since(result.StartTime)
		return result, nil
	}

	// Step 3: Execute only what's needed
	executionResult, err := r.executeMinimalTests(ctx, needsExecution)
	if err != nil {
		return result, fmt.Errorf("test execution failed: %w", err)
	}

	result.Output = executionResult.Output
	result.ExitCode = executionResult.ExitCode
	result.Duration = time.Since(result.StartTime)

	// Step 4: Update our cache
	r.updateCache(needsExecution)

	return result, nil
}

// determineTestTargets figures out exactly what needs to be tested
func (r *OptimizedTestRunner) determineTestTargets(changes []*FileChange) []string {
	targets := make(map[string]bool)

	for _, change := range changes {
		switch change.Type {
		case ChangeTypeTest:
			// For test file changes, always run at package level to avoid compilation issues
			// Individual test files can't be compiled independently due to package dependencies
			targets[filepath.Dir(change.Path)] = true

		case ChangeTypeSource:
			// For source changes, find related test files
			relatedTests := r.findRelatedTestFiles(change.Path)
			for _, test := range relatedTests {
				targets[test] = true
			}

			// If no specific tests found, test the package
			if len(relatedTests) == 0 {
				targets[filepath.Dir(change.Path)] = true
			}

		case ChangeTypeConfig, ChangeTypeDependency:
			// These affect everything - but let Go's cache handle most of it
			targets["./..."] = true
		}
	}

	result := make([]string, 0, len(targets))
	for target := range targets {
		result = append(result, target)
	}

	return result
}

// determineNeedsExecution determines what actually needs execution based on changes
func (r *OptimizedTestRunner) determineNeedsExecution(targets []string, changes []*FileChange) []string {
	var needsExecution []string

	// If we have file changes, we should run tests - be less aggressive about caching during development
	hasActualChanges := false
	for _, change := range changes {
		if change.IsNew {
			hasActualChanges = true
			break
		}
	}

	// If we detected actual file changes, run the tests
	if hasActualChanges {
		return targets
	}

	// Otherwise, check our cache but be conservative
	r.cache.mu.RLock()
	defer r.cache.mu.RUnlock()

	for _, target := range targets {
		// Only trust cache if it's very recent (less than 1 minute)
		if lastRun, exists := r.cache.packageCache[target]; exists {
			if time.Since(lastRun) < 1*time.Minute && !r.haveDependenciesChanged(target, lastRun) {
				// This target can be cached
				continue
			}
		}
		// Otherwise, we need to execute it
		needsExecution = append(needsExecution, target)
	}

	return needsExecution
}

// haveDependenciesChanged checks if dependencies for a target have changed
func (r *OptimizedTestRunner) haveDependenciesChanged(target string, since time.Time) bool {
	deps, exists := r.cache.testDependencies[target]
	if !exists {
		return true // Unknown dependencies = assume changed
	}

	for _, dep := range deps {
		if modTime, exists := r.cache.fileModTimes[dep]; exists {
			if modTime.After(since) {
				return true
			}
		} else {
			// Check actual file system
			if info, err := os.Stat(dep); err == nil {
				if info.ModTime().After(since) {
					return true
				}
			}
		}
	}

	return false
}

// executeMinimalTests runs only the tests that actually need execution
func (r *OptimizedTestRunner) executeMinimalTests(ctx context.Context, targets []string) (*TestExecutionResult, error) {
	if len(targets) == 0 {
		return &TestExecutionResult{
			Output:   "No tests needed execution\n",
			ExitCode: 0,
		}, nil
	}

	// Build optimized go test command
	args := r.buildOptimizedCommand(targets)

	cmd := exec.CommandContext(ctx, "go", args...)

	// Capture output
	output, err := cmd.CombinedOutput()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute tests: %w", err)
		}
	}

	return &TestExecutionResult{
		Output:   string(output),
		ExitCode: exitCode,
	}, nil
}

// buildOptimizedCommand builds the most efficient go test command
func (r *OptimizedTestRunner) buildOptimizedCommand(targets []string) []string {
	args := []string{"test"}

	// Let Go use its built-in caching unless we specifically need to bypass it
	if !r.enableGoCache {
		args = append(args, "-count=1") // Bypass Go's cache
	}

	// Add specific optimizations
	args = append(args, "-failfast") // Stop on first failure for faster feedback

	// Add the targets
	args = append(args, targets...)

	return args
}

// findRelatedTestFiles finds test files related to a source file
func (r *OptimizedTestRunner) findRelatedTestFiles(sourceFile string) []string {
	dir := filepath.Dir(sourceFile)
	base := filepath.Base(sourceFile)

	// Remove .go extension and add _test.go
	nameWithoutExt := strings.TrimSuffix(base, ".go")
	testFile := filepath.Join(dir, nameWithoutExt+"_test.go")

	var relatedTests []string

	// Check if direct test file exists
	if _, err := os.Stat(testFile); err == nil {
		relatedTests = append(relatedTests, testFile)
	}

	// Look for other test files in the same package that might import this
	pattern := filepath.Join(dir, "*_test.go")
	matches, err := filepath.Glob(pattern)
	if err == nil {
		for _, match := range matches {
			if match != testFile { // Don't duplicate
				relatedTests = append(relatedTests, match)
			}
		}
	}

	return relatedTests
}

// updateCache updates our internal cache after test execution
func (r *OptimizedTestRunner) updateCache(executedTargets []string) {
	r.cache.mu.Lock()
	defer r.cache.mu.Unlock()

	now := time.Now()
	for _, target := range executedTargets {
		r.cache.packageCache[target] = now

		// Update dependencies
		deps := r.scanDependencies(target)
		r.cache.testDependencies[target] = deps

		// Update file modification times
		for _, dep := range deps {
			if info, err := os.Stat(dep); err == nil {
				r.cache.fileModTimes[dep] = info.ModTime()
			}
		}
	}
}

// scanDependencies scans for dependencies of a test target
func (r *OptimizedTestRunner) scanDependencies(target string) []string {
	var deps []string

	// For file targets
	if strings.HasSuffix(target, ".go") {
		dir := filepath.Dir(target)

		// Add all .go files in the same package
		pattern := filepath.Join(dir, "*.go")
		matches, err := filepath.Glob(pattern)
		if err == nil {
			deps = append(deps, matches...)
		}

		// Add go.mod and go.sum
		if _, err := os.Stat("go.mod"); err == nil {
			deps = append(deps, "go.mod")
		}
		if _, err := os.Stat("go.sum"); err == nil {
			deps = append(deps, "go.sum")
		}
	}

	return deps
}

// OptimizedTestResult represents the result of optimized test execution
type OptimizedTestResult struct {
	TestsRun  int
	CacheHits int
	Output    string
	ExitCode  int
	Duration  time.Duration
	StartTime time.Time
	Message   string
}

// TestExecutionResult represents raw test execution result
type TestExecutionResult struct {
	Output   string
	ExitCode int
}

// GetEfficiencyStats returns efficiency statistics
func (r *OptimizedTestResult) GetEfficiencyStats() map[string]interface{} {
	total := r.TestsRun + r.CacheHits
	cacheHitRate := 0.0
	if total > 0 {
		cacheHitRate = float64(r.CacheHits) / float64(total) * 100
	}

	return map[string]interface{}{
		"total_targets":    total,
		"cache_hits":       r.CacheHits,
		"tests_run":        r.TestsRun,
		"cache_hit_rate":   cacheHitRate,
		"duration_ms":      r.Duration.Milliseconds(),
		"efficiency_score": cacheHitRate,
	}
}

// ClearCache clears the internal cache (for testing or reset)
func (r *OptimizedTestRunner) ClearCache() {
	r.cache.mu.Lock()
	defer r.cache.mu.Unlock()

	r.cache.testDependencies = make(map[string][]string)
	r.cache.packageCache = make(map[string]time.Time)
	r.cache.fileModTimes = make(map[string]time.Time)
}

// SetCacheEnabled enables or disables Go's built-in caching
func (r *OptimizedTestRunner) SetCacheEnabled(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enableGoCache = enabled
}

// SetOnlyRunChangedTests enables running only specific changed tests
func (r *OptimizedTestRunner) SetOnlyRunChangedTests(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onlyRunChangedTests = enabled
}

// SetOptimizationMode configures the optimization strategy
func (r *OptimizedTestRunner) SetOptimizationMode(mode string) {
	switch mode {
	case "aggressive":
		r.SetCacheEnabled(true)
		r.SetOnlyRunChangedTests(true)
	case "conservative":
		r.SetCacheEnabled(true)
		r.SetOnlyRunChangedTests(false)
	case "disabled":
		r.SetCacheEnabled(false)
		r.SetOnlyRunChangedTests(false)
	default:
		// Default to aggressive optimization
		r.SetCacheEnabled(true)
		r.SetOnlyRunChangedTests(true)
	}
}
