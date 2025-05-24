package runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/newbpydev/go-sentinel/internal/config"
	"github.com/newbpydev/go-sentinel/internal/test/processor"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// ParallelTestRunner executes multiple test packages in parallel
type ParallelTestRunner struct {
	maxConcurrency int
	testRunner     TestRunnerInterface
	cache          CacheInterface
}

// ParallelTestResult represents the result of a parallel test execution
type ParallelTestResult struct {
	TestPath  string
	Suite     *models.TestSuite
	Error     error
	Duration  time.Duration
	FromCache bool
}

// CacheInterface defines the interface for test result caching
type CacheInterface interface {
	GetCachedResult(testPath string) (*CachedResult, bool)
	CacheResult(testPath string, suite *models.TestSuite)
}

// CachedResult represents a cached test result
type CachedResult struct {
	Suite *models.TestSuite
}

// ColorFormatterInterface defines the interface for color formatting
type ColorFormatterInterface interface {
	// Add methods as needed for color formatting
}

// IconProviderInterface defines the interface for icon provision
type IconProviderInterface interface {
	// Add methods as needed for icon provision
}

// NewParallelTestRunner creates a new parallel test runner
func NewParallelTestRunner(maxConcurrency int, testRunner TestRunnerInterface, cache CacheInterface) *ParallelTestRunner {
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
func (r *ParallelTestRunner) RunParallel(ctx context.Context, testPaths []string, cfg *config.Config) ([]*ParallelTestResult, error) {
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

			result := r.executeTestPath(ctx, path, cfg)
			results <- result
		}(testPath)
	}

	// Close results channel when all workers complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	allResults := make([]*ParallelTestResult, 0, len(testPaths))
	for result := range results {
		allResults = append(allResults, result)
	}

	return allResults, nil
}

// executeTestPath executes a single test path with caching
func (r *ParallelTestRunner) executeTestPath(ctx context.Context, testPath string, cfg *config.Config) *ParallelTestResult {
	startTime := time.Now()

	// Check cache first
	if r.cache != nil {
		if cached, exists := r.cache.GetCachedResult(testPath); exists {
			return &ParallelTestResult{
				TestPath:  testPath,
				Suite:     cached.Suite,
				Error:     nil,
				Duration:  time.Since(startTime),
				FromCache: true,
			}
		}
	}

	// Execute the test
	suite, err := r.runSingleTestPath(ctx, testPath, cfg)
	duration := time.Since(startTime)

	// Cache the result if successful
	if err == nil && suite != nil && r.cache != nil {
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
func (r *ParallelTestRunner) runSingleTestPath(ctx context.Context, testPath string, cfg *config.Config) (*models.TestSuite, error) {
	// Create a new test runner instance for this execution to avoid race conditions
	testRunner := NewBasicTestRunner(cfg.Verbosity > 0, true)

	// Apply timeout if configured
	testCtx := ctx
	if cfg.Timeout > 0 {
		var cancel context.CancelFunc
		testCtx, cancel = context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
	}

	// Execute test command using streaming approach
	stream, err := testRunner.RunStream(testCtx, []string{testPath})
	if err != nil {
		return nil, fmt.Errorf("failed to start test stream for %s: %w", testPath, err)
	}
	defer stream.Close()

	// Create a processor for this test execution
	testProcessor := processor.NewTestProcessor(
		&discardWriter{},      // Use discard writer for parallel execution
		&nullColorFormatter{}, // No colors for parallel processing
		&nullIconProvider{},   // No icons for parallel processing
		80,
	)

	// Process the stream
	progress := make(chan models.TestProgress, 10)
	defer close(progress)

	// Start progress monitoring in background (optional)
	go func() {
		for range progress {
			// Consume progress updates without action for parallel execution
		}
	}()

	// Process the stream
	if err := testProcessor.ProcessStream(stream, progress); err != nil {
		return nil, fmt.Errorf("failed to process test stream for %s: %w", testPath, err)
	}

	// Extract the test suite for this path
	if suite, exists := testProcessor.GetSuites()[testPath]; exists {
		return suite, nil
	}

	// If no suite found for exact path, look for first suite (fallback)
	for _, suite := range testProcessor.GetSuites() {
		return suite, nil
	}

	return nil, fmt.Errorf("no test suite found for path %s", testPath)
}

// discardWriter is a writer that discards all writes (for parallel execution)
type discardWriter struct{}

func (d *discardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

// nullColorFormatter implements a basic color formatter interface
type nullColorFormatter struct{}

func (n *nullColorFormatter) Red(text string) string                 { return text }
func (n *nullColorFormatter) Green(text string) string               { return text }
func (n *nullColorFormatter) Yellow(text string) string              { return text }
func (n *nullColorFormatter) Blue(text string) string                { return text }
func (n *nullColorFormatter) Magenta(text string) string             { return text }
func (n *nullColorFormatter) Cyan(text string) string                { return text }
func (n *nullColorFormatter) Gray(text string) string                { return text }
func (n *nullColorFormatter) Bold(text string) string                { return text }
func (n *nullColorFormatter) Dim(text string) string                 { return text }
func (n *nullColorFormatter) White(text string) string               { return text }
func (n *nullColorFormatter) Colorize(text, colorName string) string { return text }

// nullIconProvider implements a basic icon provider interface
type nullIconProvider struct{}

func (n *nullIconProvider) CheckMark() string              { return "✓" }
func (n *nullIconProvider) Cross() string                  { return "✗" }
func (n *nullIconProvider) Skipped() string                { return "-" }
func (n *nullIconProvider) Running() string                { return "..." }
func (n *nullIconProvider) GetIcon(iconType string) string { return "•" }

// MergeResults merges multiple parallel test results into processor suites
func MergeResults(testProcessor *processor.TestProcessor, results []*ParallelTestResult) {
	for _, result := range results {
		if result.Error != nil {
			// Log error but continue with other results
			continue
		}

		if result.Suite != nil {
			testProcessor.AddTestSuite(result.Suite)
		}
	}
}
