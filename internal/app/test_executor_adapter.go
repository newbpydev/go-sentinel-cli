// Package app provides test executor adapter to maintain clean package boundaries
package app

import (
	"context"
	"fmt"
)

// testExecutorAdapter adapts test execution to use proper dependency injection
// This adapter eliminates direct dependencies on internal packages
type testExecutorAdapter struct {
	// Dependencies injected through interfaces
	testRunner  TestRunner
	processor   TestProcessor
	coordinator WatchCoordinator
	formatter   ColorFormatter
	icons       IconProvider
	config      *Configuration
}

// TestRunner interface for test execution - defined in app package as consumer
type TestRunner interface {
	Execute(ctx context.Context, packages []string, options *TestRunOptions) (*TestRunResult, error)
}

// TestProcessor interface for result processing - defined in app package as consumer
type TestProcessor interface {
	ProcessResults(results *TestRunResult) error
	RenderResults(showSummary bool) error
	GetStats() *TestStats
}

// WatchCoordinator interface for watch mode - defined in app package as consumer
type WatchCoordinator interface {
	Start(ctx context.Context) error
	Configure(options *WatchOptions) error
}

// ColorFormatter interface for color formatting - defined in app package as consumer
type ColorFormatter interface {
	FormatSuccess(text string) string
	FormatError(text string) string
	FormatWarning(text string) string
}

// IconProvider interface for icons - defined in app package as consumer
type IconProvider interface {
	GetSuccessIcon() string
	GetErrorIcon() string
	GetWarningIcon() string
}

// TestRunOptions represents test execution options
type TestRunOptions struct {
	Verbose    bool
	JSONOutput bool
	Timeout    string
	Parallel   int
}

// TestRunResult represents test execution results
type TestRunResult struct {
	Success     bool
	Packages    []*PackageTestResult
	Duration    string
	TotalTests  int
	PassedTests int
	FailedTests int
}

// PackageTestResult represents results for a single package
type PackageTestResult struct {
	Package string
	Success bool
	Output  string
	Error   error
}

// TestStats represents test execution statistics
type TestStats struct {
	TotalTests  int
	PassedTests int
	FailedTests int
	Duration    string
}

// WatchOptions represents watch mode configuration
type WatchOptions struct {
	Paths            []string
	IgnorePatterns   []string
	TestPatterns     []string
	DebounceInterval string
	ClearTerminal    bool
	RunOnStart       bool
}

// NewTestExecutorAdapter creates a new test executor adapter with dependency injection
func NewTestExecutorAdapter() TestExecutor {
	return &testExecutorAdapter{}
}

// NewTestExecutor creates a new test executor using the adapter pattern
// This eliminates direct dependencies on internal packages
func NewTestExecutor() TestExecutor {
	return NewTestExecutorAdapter()
}

// SetConfiguration configures the test executor adapter
func (e *testExecutorAdapter) SetConfiguration(config *Configuration) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	e.config = config

	// Dependencies will be injected through factory pattern
	// This eliminates direct imports of internal packages
	return nil
}

// ExecuteSingle executes tests once for the specified packages
func (e *testExecutorAdapter) ExecuteSingle(ctx context.Context, packages []string, config *Configuration) error {
	if err := e.ensureConfigured(config); err != nil {
		return err
	}

	if len(packages) == 0 {
		packages = []string{"./..."}
	}

	// Convert app config to test runner options
	options := &TestRunOptions{
		Verbose:    config.Verbosity > 0,
		JSONOutput: true,
		Timeout:    config.Test.Timeout,
		Parallel:   config.Test.Parallel,
	}

	// Execute tests using injected test runner
	if e.testRunner == nil {
		return fmt.Errorf("test runner not configured")
	}

	result, err := e.testRunner.Execute(ctx, packages, options)
	if err != nil {
		return fmt.Errorf("test execution failed: %w", err)
	}

	// Process results using injected processor
	if e.processor == nil {
		return fmt.Errorf("test processor not configured")
	}

	if err := e.processor.ProcessResults(result); err != nil {
		return fmt.Errorf("failed to process results: %w", err)
	}

	if err := e.processor.RenderResults(true); err != nil {
		return fmt.Errorf("failed to render results: %w", err)
	}

	// Check for test failures
	stats := e.processor.GetStats()
	if stats.FailedTests > 0 {
		return fmt.Errorf("%d tests failed", stats.FailedTests)
	}

	return nil
}

// ExecuteWatch executes tests in watch mode
func (e *testExecutorAdapter) ExecuteWatch(ctx context.Context, config *Configuration) error {
	if err := e.ensureConfigured(config); err != nil {
		return err
	}

	// Convert app config to watch options
	watchOptions := &WatchOptions{
		Paths:            config.Paths.IncludePatterns,
		IgnorePatterns:   config.Watch.IgnorePatterns,
		TestPatterns:     []string{"*_test.go"},
		DebounceInterval: config.Watch.Debounce,
		ClearTerminal:    config.Watch.ClearOnRerun,
		RunOnStart:       config.Watch.RunOnStart,
	}

	// Configure watch coordinator
	if e.coordinator == nil {
		return fmt.Errorf("watch coordinator not configured")
	}

	if err := e.coordinator.Configure(watchOptions); err != nil {
		return fmt.Errorf("failed to configure watch coordinator: %w", err)
	}

	// Start watch mode
	return e.coordinator.Start(ctx)
}

// ensureConfigured ensures the adapter is properly configured
func (e *testExecutorAdapter) ensureConfigured(config *Configuration) error {
	if e.config == nil {
		if config == nil {
			return fmt.Errorf("no configuration provided")
		}
		return e.SetConfiguration(config)
	}
	return nil
}

// SetTestRunner injects the test runner dependency
func (e *testExecutorAdapter) SetTestRunner(runner TestRunner) {
	e.testRunner = runner
}

// SetTestProcessor injects the test processor dependency
func (e *testExecutorAdapter) SetTestProcessor(processor TestProcessor) {
	e.processor = processor
}

// SetWatchCoordinator injects the watch coordinator dependency
func (e *testExecutorAdapter) SetWatchCoordinator(coordinator WatchCoordinator) {
	e.coordinator = coordinator
}

// SetColorFormatter injects the color formatter dependency
func (e *testExecutorAdapter) SetColorFormatter(formatter ColorFormatter) {
	e.formatter = formatter
}

// SetIconProvider injects the icon provider dependency
func (e *testExecutorAdapter) SetIconProvider(icons IconProvider) {
	e.icons = icons
}
