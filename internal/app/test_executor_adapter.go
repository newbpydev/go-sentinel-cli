// Package app provides test executor adapter to maintain clean package boundaries
package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/processor"
	"github.com/newbpydev/go-sentinel/internal/test/runner"
	"github.com/newbpydev/go-sentinel/internal/ui/colors"
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

// NewTestExecutorAdapter creates a new test executor adapter with dependency injection
func NewTestExecutorAdapter() TestExecutor {
	return &testExecutorAdapter{}
}

// NewTestExecutor creates a new test executor using the adapter pattern
// This eliminates direct dependencies on internal packages
func NewTestExecutor() TestExecutor {
	adapter := &testExecutorAdapter{}

	// Wire real implementations following architecture principles
	// TestRunner interface is defined in app package (consumer owns interface)
	testRunner := &testRunnerAdapter{
		executor: runner.NewExecutor(),
	}
	adapter.SetTestRunner(testRunner)

	// TestProcessor interface is defined in app package
	testProcessor := &testProcessorAdapter{
		processor: nil, // Will be created when needed
	}
	adapter.SetTestProcessor(testProcessor)

	// ColorFormatter interface is defined in app package
	colorFormatter := &colorFormatterAdapter{
		formatter: colors.NewAutoColorFormatter(),
	}
	adapter.SetColorFormatter(colorFormatter)

	// IconProvider interface is defined in app package
	iconProvider := &iconProviderAdapter{
		provider: colors.NewAutoIconProvider(),
	}
	adapter.SetIconProvider(iconProvider)

	return adapter
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

	// Normalize package paths - ensure relative paths start with "./"
	normalizedPackages := make([]string, len(packages))
	for i, pkg := range packages {
		normalizedPackages[i] = e.normalizePackagePath(pkg)
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

	result, err := e.testRunner.Execute(ctx, normalizedPackages, options)
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

	// Check for package-level errors (e.g., package not found)
	for _, pkg := range result.Packages {
		if pkg.Error != nil {
			// Check if it's a package not found error
			errorStr := pkg.Error.Error()
			if strings.Contains(errorStr, "no such file or directory") ||
				strings.Contains(errorStr, "cannot find package") ||
				strings.Contains(errorStr, "package not found") ||
				strings.Contains(errorStr, "no Go files") {
				return fmt.Errorf("package not found: %s", pkg.Package)
			}
			return fmt.Errorf("package %s failed: %w", pkg.Package, pkg.Error)
		}
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

// normalizePackagePath ensures package paths are in the correct format for Go
func (e *testExecutorAdapter) normalizePackagePath(pkg string) string {
	// If it's already a relative path starting with "./" or an absolute module path, keep it as is
	if strings.HasPrefix(pkg, "./") || strings.HasPrefix(pkg, "../") || strings.Contains(pkg, "/") && !strings.HasPrefix(pkg, ".") {
		// Check if it looks like a module path (contains domain or starts with known patterns)
		if strings.Contains(pkg, ".") || strings.HasPrefix(pkg, "github.com/") || strings.HasPrefix(pkg, "golang.org/") {
			return pkg
		}
		// It's a relative path without "./" prefix, add it
		return "./" + pkg
	}

	// If it's just a package name like "internal/test/runner", make it relative
	if !strings.HasPrefix(pkg, ".") && strings.Contains(pkg, "/") {
		return "./" + pkg
	}

	// For simple package names or already correct paths, return as is
	return pkg
}

// testRunnerAdapter adapts internal test runner to app interface
type testRunnerAdapter struct {
	executor runner.TestExecutor
}

func (a *testRunnerAdapter) Execute(ctx context.Context, packages []string, options *TestRunOptions) (*TestRunResult, error) {
	// Convert app options to internal runner options
	runnerOptions := &runner.ExecutionOptions{
		Verbose:    options.Verbose,
		JSONOutput: options.JSONOutput,
		Parallel:   options.Parallel,
	}

	// Parse timeout
	if options.Timeout != "" {
		if duration, err := time.ParseDuration(options.Timeout); err == nil {
			runnerOptions.Timeout = duration
		}
	}

	// Execute using internal runner
	result, err := a.executor.Execute(ctx, packages, runnerOptions)
	if err != nil {
		return nil, err
	}

	// Convert internal result to app result
	appResult := &TestRunResult{
		Success:     result.Success,
		TotalTests:  result.TotalTests,
		PassedTests: result.PassedTests,
		FailedTests: result.FailedTests,
		Duration:    result.TotalDuration.String(),
		Packages:    make([]*PackageTestResult, len(result.Packages)),
	}

	for i, pkg := range result.Packages {
		appResult.Packages[i] = &PackageTestResult{
			Package: pkg.Package,
			Success: pkg.Success,
			Output:  pkg.Output,
			Error:   pkg.Error,
		}
	}

	return appResult, nil
}

// testProcessorAdapter adapts internal test processor to app interface
type testProcessorAdapter struct {
	processor processor.OutputProcessor
	stats     *TestStats
}

func (a *testProcessorAdapter) ProcessResults(results *TestRunResult) error {
	// Update stats
	a.stats = &TestStats{
		TotalTests:  results.TotalTests,
		PassedTests: results.PassedTests,
		FailedTests: results.FailedTests,
		Duration:    results.Duration,
	}
	return nil
}

func (a *testProcessorAdapter) RenderResults(showSummary bool) error {
	if a.stats == nil {
		return fmt.Errorf("no results to render")
	}

	if showSummary {
		fmt.Printf("📊 Test Summary:\n")
		fmt.Printf("   Total: %d\n", a.stats.TotalTests)
		fmt.Printf("   Passed: %d\n", a.stats.PassedTests)
		fmt.Printf("   Failed: %d\n", a.stats.FailedTests)
		fmt.Printf("   Duration: %s\n", a.stats.Duration)
	}

	return nil
}

func (a *testProcessorAdapter) GetStats() *TestStats {
	if a.stats == nil {
		return &TestStats{}
	}
	return a.stats
}

// colorFormatterAdapter adapts internal color formatter to app interface
type colorFormatterAdapter struct {
	formatter colors.FormatterInterface
}

func (a *colorFormatterAdapter) FormatSuccess(text string) string {
	return a.formatter.Green(text)
}

func (a *colorFormatterAdapter) FormatError(text string) string {
	return a.formatter.Red(text)
}

func (a *colorFormatterAdapter) FormatWarning(text string) string {
	return a.formatter.Yellow(text)
}

// iconProviderAdapter adapts internal icon provider to app interface
type iconProviderAdapter struct {
	provider colors.IconProviderInterface
}

func (a *iconProviderAdapter) GetSuccessIcon() string {
	return a.provider.CheckMark()
}

func (a *iconProviderAdapter) GetErrorIcon() string {
	return a.provider.Cross()
}

func (a *iconProviderAdapter) GetWarningIcon() string {
	return a.provider.GetIcon("warning")
}
