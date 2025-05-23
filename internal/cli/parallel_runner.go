package cli

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ParallelTestRunner executes multiple test packages in parallel
type ParallelTestRunner struct {
	maxConcurrency int
	testRunner     *TestRunner
	cache          *TestResultCache
}

// ParallelTestResult represents the result of a parallel test execution
type ParallelTestResult struct {
	TestPath  string
	Suite     *TestSuite
	Error     error
	Duration  time.Duration
	FromCache bool
}

// NewParallelTestRunner creates a new parallel test runner
func NewParallelTestRunner(maxConcurrency int, testRunner *TestRunner, cache *TestResultCache) *ParallelTestRunner {
	if maxConcurrency <= 0 {
		maxConcurrency = 4 // Default concurrency
	}

	return &ParallelTestRunner{
		maxConcurrency: maxConcurrency,
		testRunner:     testRunner,
		cache:          cache,
	}
}

// RunParallel executes multiple test packages in parallel
func (r *ParallelTestRunner) RunParallel(ctx context.Context, testPaths []string, config *Config) ([]*ParallelTestResult, error) {
	if len(testPaths) == 0 {
		return nil, nil
	}

	// Channel for results
	results := make(chan *ParallelTestResult, len(testPaths))

	// Worker pool with semaphore for concurrency control
	semaphore := make(chan struct{}, r.maxConcurrency)

	var wg sync.WaitGroup

	// Start workers for each test path
	for _, testPath := range testPaths {
		wg.Add(1)
		go func(path string) {
			defer wg.Done()

			// Acquire semaphore
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := r.executeTestPath(ctx, path, config)
			results <- result
		}(testPath)
	}

	// Close results channel when all workers complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var allResults []*ParallelTestResult
	for result := range results {
		allResults = append(allResults, result)
	}

	return allResults, nil
}

// executeTestPath executes a single test path with caching
func (r *ParallelTestRunner) executeTestPath(ctx context.Context, testPath string, config *Config) *ParallelTestResult {
	startTime := time.Now()

	// Check cache first
	if cached, exists := r.cache.GetCachedResult(testPath); exists {
		return &ParallelTestResult{
			TestPath:  testPath,
			Suite:     cached.Suite,
			Error:     nil,
			Duration:  time.Since(startTime),
			FromCache: true,
		}
	}

	// Execute the test
	suite, err := r.runSingleTestPath(ctx, testPath, config)
	duration := time.Since(startTime)

	// Cache the result if successful
	if err == nil && suite != nil {
		r.cache.CacheResult(testPath, suite)
	}

	return &ParallelTestResult{
		TestPath:  testPath,
		Suite:     suite,
		Error:     err,
		Duration:  duration,
		FromCache: false,
	}
}

// runSingleTestPath executes a single test path and returns the suite
func (r *ParallelTestRunner) runSingleTestPath(ctx context.Context, testPath string, config *Config) (*TestSuite, error) {
	// Configure the test runner for this execution
	r.testRunner.Verbose = config.Verbosity > 0
	r.testRunner.JSONOutput = true

	// Apply timeout if configured
	testCtx := ctx
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		testCtx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// Execute test command using streaming approach
	stream, err := r.testRunner.RunStream(testCtx, []string{testPath})
	if err != nil {
		return nil, fmt.Errorf("failed to start test stream for %s: %w", testPath, err)
	}
	defer stream.Close()

	// Create a processor for this test execution
	processor := NewTestProcessor(
		&discardWriter{},         // Use discard writer for parallel execution
		NewColorFormatter(false), // No colors for parallel processing
		NewIconProvider(false),   // No icons for parallel processing
		80,
	)

	// Process the stream
	progress := make(chan TestProgress, 10)
	defer close(progress)

	// Start progress monitoring in background (optional)
	go func() {
		for range progress {
			// Consume progress updates without action for parallel execution
		}
	}()

	// Process the stream
	if err := processor.ProcessStream(stream, progress); err != nil {
		return nil, fmt.Errorf("failed to process test stream for %s: %w", testPath, err)
	}

	// Extract the test suite for this path
	if suite, exists := processor.suites[testPath]; exists {
		return suite, nil
	}

	// If no suite found for exact path, look for first suite (fallback)
	for _, suite := range processor.suites {
		return suite, nil
	}

	return nil, fmt.Errorf("no test suite found for path %s", testPath)
}

// discardWriter is a writer that discards all writes (for parallel execution)
type discardWriter struct{}

func (d *discardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// MergeResults merges multiple parallel test results into processor suites
func MergeResults(processor *TestProcessor, results []*ParallelTestResult) {
	for _, result := range results {
		if result.Error != nil {
			// Log error but continue with other results
			continue
		}

		if result.Suite != nil {
			processor.AddTestSuite(result.Suite)
		}
	}
}
