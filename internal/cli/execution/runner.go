package execution

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/newbpydev/go-sentinel/internal/cli/core"
)

// SmartTestRunner is an intelligent test runner that leverages caching and optimization
type SmartTestRunner struct {
	cache        core.CacheManager
	strategy     core.ExecutionStrategy
	capabilities core.RunnerCapabilities
	processor    *TestProcessor // Integrated processor for JSON output parsing
	verbose      bool
	mu           sync.RWMutex
}

// NewSmartTestRunner creates a new intelligent test runner
func NewSmartTestRunner(cache core.CacheManager, strategy core.ExecutionStrategy) *SmartTestRunner {
	return &SmartTestRunner{
		cache:     cache,
		strategy:  strategy,
		processor: NewTestProcessor(),
		verbose:   false,
		capabilities: core.RunnerCapabilities{
			SupportsCaching:    true,
			SupportsParallel:   true,
			SupportsWatchMode:  true,
			SupportsFiltering:  true,
			MaxConcurrency:     4,
			SupportedFileTypes: []string{".go"},
		},
		mu: sync.RWMutex{},
	}
}

// SetVerbose sets the verbose mode for the runner
func (r *SmartTestRunner) SetVerbose(verbose bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.verbose = verbose
}

// RunTests executes tests based on the provided changes and strategy
func (r *SmartTestRunner) RunTests(ctx context.Context, changes []core.FileChange, strategy core.ExecutionStrategy) (*core.TestResult, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	startTime := time.Now()

	// Step 1: Determine test targets
	targets := r.determineTestTargets(changes)

	if len(targets) == 0 {
		return &core.TestResult{
			Status:    core.StatusPassed,
			Output:    "No test targets identified",
			Duration:  time.Since(startTime),
			StartTime: startTime,
			EndTime:   time.Now(),
			CacheHit:  true,
		}, nil
	}

	// Step 2: Filter targets based on cache and strategy
	targetsToRun := r.filterTargetsForExecution(targets, strategy)

	if len(targetsToRun) == 0 {
		return &core.TestResult{
			Status:    core.StatusPassed,
			Output:    "All targets satisfied by cache",
			Duration:  time.Since(startTime),
			StartTime: startTime,
			EndTime:   time.Now(),
			CacheHit:  true,
		}, nil
	}

	// Step 3: Execute tests with JSON processing
	result, err := r.executeTestsWithProcessing(ctx, targetsToRun)
	if err != nil {
		return nil, err
	}

	result.StartTime = startTime
	result.Duration = time.Since(startTime)

	// Step 4: Update cache
	r.updateCacheAfterExecution(targetsToRun, result)

	return result, nil
}

// GetCapabilities returns the capabilities of this test runner
func (r *SmartTestRunner) GetCapabilities() core.RunnerCapabilities {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.capabilities
}

// GetProcessedResults returns the processed test suites and stats from the last run
func (r *SmartTestRunner) GetProcessedResults() (map[string]*core.TestSuite, *core.TestRunStats) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.processor.GetSuites(), r.processor.GetStats()
}

// determineTestTargets figures out what tests need to be run based on file changes
func (r *SmartTestRunner) determineTestTargets(changes []core.FileChange) []core.TestTarget {
	var targets []core.TestTarget
	targetMap := make(map[string]bool) // Deduplicate targets

	for _, change := range changes {
		switch change.Type {
		case core.ChangeTypeTest:
			// For test file changes, run at package level for compilation safety
			packagePath := filepath.Dir(change.Path)
			if !targetMap[packagePath] {
				targets = append(targets, core.TestTarget{
					Path:              packagePath,
					Type:              "package",
					Priority:          1, // High priority for test changes
					EstimatedDuration: 30 * time.Second,
				})
				targetMap[packagePath] = true
			}

		case core.ChangeTypeSource:
			// For source changes, find related tests
			relatedTargets := r.findRelatedTestTargets(change.Path)
			for _, target := range relatedTargets {
				if !targetMap[target.Path] {
					targets = append(targets, target)
					targetMap[target.Path] = true
				}
			}

			// If no specific tests found, test the package
			if len(relatedTargets) == 0 {
				packagePath := filepath.Dir(change.Path)
				if !targetMap[packagePath] {
					targets = append(targets, core.TestTarget{
						Path:              packagePath,
						Type:              "package",
						Priority:          2, // Medium priority for source changes
						EstimatedDuration: 60 * time.Second,
					})
					targetMap[packagePath] = true
				}
			}

		case core.ChangeTypeConfig, core.ChangeTypeDependency:
			// These affect everything, but let's be smart about it
			if !targetMap["./..."] {
				targets = append(targets, core.TestTarget{
					Path:              "./...",
					Type:              "recursive",
					Priority:          3, // Lower priority for config changes
					EstimatedDuration: 300 * time.Second,
				})
				targetMap["./..."] = true
			}
		}
	}

	return targets
}

// findRelatedTestTargets finds test targets related to a source file
func (r *SmartTestRunner) findRelatedTestTargets(sourceFile string) []core.TestTarget {
	var targets []core.TestTarget
	dir := filepath.Dir(sourceFile)
	base := filepath.Base(sourceFile)

	// Look for direct test file
	nameWithoutExt := strings.TrimSuffix(base, ".go")
	testFile := filepath.Join(dir, nameWithoutExt+"_test.go")

	if _, err := os.Stat(testFile); err == nil {
		targets = append(targets, core.TestTarget{
			Path:              dir, // Run at package level
			Type:              "package",
			Priority:          1,
			EstimatedDuration: 30 * time.Second,
		})
		return targets
	}

	// Look for any test files in the same package
	pattern := filepath.Join(dir, "*_test.go")
	matches, err := filepath.Glob(pattern)
	if err == nil && len(matches) > 0 {
		targets = append(targets, core.TestTarget{
			Path:              dir,
			Type:              "package",
			Priority:          2,
			EstimatedDuration: 60 * time.Second,
		})
	}

	return targets
}

// filterTargetsForExecution determines which targets actually need to be executed
func (r *SmartTestRunner) filterTargetsForExecution(targets []core.TestTarget, strategy core.ExecutionStrategy) []core.TestTarget {
	var targetsToRun []core.TestTarget

	// Handle nil strategy gracefully
	if strategy == nil {
		// If no strategy provided, run all targets
		return targets
	}

	for _, target := range targets {
		if strategy.ShouldRunTest(target, r.cache) {
			targetsToRun = append(targetsToRun, target)
		}
	}

	// Apply execution order optimization
	return strategy.GetExecutionOrder(targetsToRun)
}

// executeTestsWithProcessing runs tests and processes the JSON output
func (r *SmartTestRunner) executeTestsWithProcessing(ctx context.Context, targets []core.TestTarget) (*core.TestResult, error) {
	if len(targets) == 0 {
		return &core.TestResult{
			Status:   core.StatusPassed,
			Output:   "No tests to execute",
			CacheHit: true,
		}, nil
	}

	// For now, execute all targets together for efficiency
	paths := make([]string, len(targets))
	for i, target := range targets {
		paths[i] = target.Path
	}

	// Build command with JSON output for processing
	cmd := r.buildTestCommandWithJSON(paths)
	cmdExec := exec.CommandContext(ctx, "go", cmd...)

	// Execute and capture output
	output, err := cmdExec.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, core.NewTestExecutionError(
				core.TestTarget{Path: strings.Join(paths, ", ")},
				strings.Join(cmd, " "),
				0,
				string(output),
				0,
				fmt.Errorf("failed to execute test command: %w", err),
			)
		}
	}

	// Process the JSON output
	outputStr := string(output)
	if err := r.processor.ProcessJSONOutput(outputStr); err != nil {
		// If JSON processing fails, fall back to basic result
		status := core.StatusPassed
		if exitCode != 0 {
			status = core.StatusFailed
		}

		return &core.TestResult{
			Target: core.TestTarget{
				Path: strings.Join(paths, ", "),
				Type: "multiple",
			},
			Status:   status,
			Output:   outputStr,
			EndTime:  time.Now(),
			CacheHit: false,
		}, nil
	}

	// Get processed statistics
	stats := r.processor.GetStats()

	// Determine overall status from processed results
	status := core.StatusPassed
	if stats.FailedTests > 0 {
		status = core.StatusFailed
	} else if stats.TotalTests == 0 {
		status = core.StatusSkipped
	}

	return &core.TestResult{
		Target: core.TestTarget{
			Path: strings.Join(paths, ", "),
			Type: "multiple",
		},
		Status:       status,
		Output:       outputStr,
		EndTime:      time.Now(),
		TestCount:    stats.TotalTests,
		PassedCount:  stats.PassedTests,
		FailedCount:  stats.FailedTests,
		SkippedCount: stats.SkippedTests,
		CacheHit:     false,
	}, nil
}

// buildTestCommandWithJSON builds a test command with JSON output
func (r *SmartTestRunner) buildTestCommandWithJSON(paths []string) []string {
	args := []string{"test"}

	// Add JSON output for processing
	args = append(args, "-json")

	// Add verbose if enabled
	if r.verbose {
		args = append(args, "-v")
	}

	// Add performance optimizations
	args = append(args, "-failfast") // Stop on first failure for faster feedback

	// Add the paths
	args = append(args, paths...)

	return args
}

// updateCacheAfterExecution updates the cache with execution results
func (r *SmartTestRunner) updateCacheAfterExecution(targets []core.TestTarget, result *core.TestResult) {
	// For each target, store the result in cache
	for _, target := range targets {
		r.cache.StoreResult(target, result)
	}
}
