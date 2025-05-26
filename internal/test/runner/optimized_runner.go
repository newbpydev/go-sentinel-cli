package runner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/models"
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

// OptimizedTestResult represents the result of an optimized test run
type OptimizedTestResult struct {
	TestsRun  int
	CacheHits int
	Output    string
	ExitCode  int
	Duration  time.Duration
	StartTime time.Time
	Message   string
}

// TestExecutionResult represents the result of test execution
type TestExecutionResult struct {
	Output   string
	ExitCode int
}

// FileChangeInterface defines the interface for file changes
type FileChangeInterface interface {
	GetPath() string
	GetType() ChangeType
	IsNewChange() bool
}

// ChangeType represents different types of file changes
type ChangeType int

const (
	ChangeTypeTest ChangeType = iota
	ChangeTypeSource
	ChangeTypeConfig
	ChangeTypeDependency
)

// FileChangeAdapter adapts models.FileChange to our interface
type FileChangeAdapter struct {
	*models.FileChange
}

func (f *FileChangeAdapter) GetPath() string {
	return f.FilePath
}

func (f *FileChangeAdapter) GetType() ChangeType {
	switch f.ChangeType {
	case models.ChangeTypeCreated, models.ChangeTypeModified:
		// Determine type based on file path
		if strings.HasSuffix(f.FilePath, "_test.go") {
			return ChangeTypeTest
		} else if strings.HasSuffix(f.FilePath, ".go") {
			return ChangeTypeSource
		} else if strings.Contains(f.FilePath, "go.mod") || strings.Contains(f.FilePath, "go.sum") {
			return ChangeTypeDependency
		}
		return ChangeTypeConfig
	default:
		return ChangeTypeConfig
	}
}

func (f *FileChangeAdapter) IsNewChange() bool {
	return f.ChangeType == models.ChangeTypeCreated || f.ChangeType == models.ChangeTypeModified
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
func (r *OptimizedTestRunner) RunOptimized(ctx context.Context, changes []FileChangeInterface) (*OptimizedTestResult, error) {
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
func (r *OptimizedTestRunner) determineTestTargets(changes []FileChangeInterface) []string {
	targets := make(map[string]bool)

	for _, change := range changes {
		switch change.GetType() {
		case ChangeTypeTest:
			// For test file changes, always run at package level to avoid compilation issues
			// Individual test files can't be compiled independently due to package dependencies
			targets[filepath.Dir(change.GetPath())] = true

		case ChangeTypeSource:
			// For source changes, find related test files
			relatedTests := r.findRelatedTestFiles(change.GetPath())
			for _, test := range relatedTests {
				targets[test] = true
			}

			// If no specific tests found, test the package
			if len(relatedTests) == 0 {
				targets[filepath.Dir(change.GetPath())] = true
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
func (r *OptimizedTestRunner) determineNeedsExecution(targets []string, changes []FileChangeInterface) []string {
	needsExecution := make([]string, 0, len(targets))

	// If we have file changes, we should run tests - be less aggressive about caching during development
	hasActualChanges := false
	for _, change := range changes {
		if change.IsNewChange() {
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

	// Build the command with optimizations
	args := r.buildOptimizedCommand(targets)

	// Execute the command
	cmd := exec.CommandContext(ctx, "go", args...)

	// CRITICAL FIX: Set process group to ensure child processes are cleaned up
	setProcessGroup(cmd)

	output, err := cmd.CombinedOutput()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute test command: %w", err)
		}
	}

	return &TestExecutionResult{
		Output:   string(output),
		ExitCode: exitCode,
	}, nil
}

// buildOptimizedCommand builds the optimized test command
func (r *OptimizedTestRunner) buildOptimizedCommand(targets []string) []string {
	args := []string{"test"}

	if r.enableGoCache {
		// Go's built-in caching is enabled by default, no flag needed
	}

	// Add JSON output for easier parsing
	args = append(args, "-json")

	// Add timeout to prevent hanging
	args = append(args, "-timeout=30s")

	// Add targets
	args = append(args, targets...)

	return args
}

// findRelatedTestFiles finds test files related to a source file
func (r *OptimizedTestRunner) findRelatedTestFiles(sourceFile string) []string {
	var related []string

	dir := filepath.Dir(sourceFile)
	baseName := strings.TrimSuffix(filepath.Base(sourceFile), ".go")

	// Look for corresponding test file
	testFile := filepath.Join(dir, baseName+"_test.go")
	if _, err := os.Stat(testFile); err == nil {
		related = append(related, testFile)
	}

	// Look for package test files in the same directory
	entries, err := os.ReadDir(dir)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), "_test.go") {
				testPath := filepath.Join(dir, entry.Name())
				if testPath != testFile { // Don't duplicate
					related = append(related, testPath)
				}
			}
		}
	}

	return related
}

// updateCache updates the cache after successful test execution
func (r *OptimizedTestRunner) updateCache(executedTargets []string) {
	r.cache.mu.Lock()
	defer r.cache.mu.Unlock()

	now := time.Now()

	for _, target := range executedTargets {
		r.cache.packageCache[target] = now

		// Scan dependencies for this target
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

// scanDependencies scans for files that a test target depends on
func (r *OptimizedTestRunner) scanDependencies(target string) []string {
	var deps []string

	// If target is a package path (directory), scan all Go files in it
	if info, err := os.Stat(target); err == nil && info.IsDir() {
		entries, err := os.ReadDir(target)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") {
					deps = append(deps, filepath.Join(target, entry.Name()))
				}
			}
		}
	}

	// Add go.mod and go.sum as dependencies
	if _, err := os.Stat("go.mod"); err == nil {
		deps = append(deps, "go.mod")
	}
	if _, err := os.Stat("go.sum"); err == nil {
		deps = append(deps, "go.sum")
	}

	return deps
}

// GetEfficiencyStats returns efficiency statistics
func (r *OptimizedTestResult) GetEfficiencyStats() map[string]interface{} {
	return map[string]interface{}{
		"tests_run":             r.TestsRun,
		"cache_hits":            r.CacheHits,
		"duration_seconds":      r.Duration.Seconds(),
		"efficiency_percentage": float64(r.CacheHits) / float64(r.TestsRun+r.CacheHits) * 100,
	}
}

// ClearCache clears the test cache
func (r *OptimizedTestRunner) ClearCache() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cache.mu.Lock()
	defer r.cache.mu.Unlock()

	r.cache.packageCache = make(map[string]time.Time)
	r.cache.fileModTimes = make(map[string]time.Time)
	r.cache.testDependencies = make(map[string][]string)
}

// SetCacheEnabled enables or disables Go's built-in test caching
func (r *OptimizedTestRunner) SetCacheEnabled(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.enableGoCache = enabled
}

// SetOnlyRunChangedTests sets whether to only run tests for changed files
func (r *OptimizedTestRunner) SetOnlyRunChangedTests(enabled bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onlyRunChangedTests = enabled
}

// SetOptimizationMode sets the optimization mode
func (r *OptimizedTestRunner) SetOptimizationMode(mode string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	switch mode {
	case "aggressive":
		r.enableGoCache = true
		r.onlyRunChangedTests = true
	case "conservative":
		r.enableGoCache = true
		r.onlyRunChangedTests = false
	case "none":
		r.enableGoCache = false
		r.onlyRunChangedTests = false
	default:
		// Default mode
		r.enableGoCache = true
		r.onlyRunChangedTests = true
	}
}
