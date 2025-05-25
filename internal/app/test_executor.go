// Package app provides test execution bridging to the modular test system
package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/processor"
	"github.com/newbpydev/go-sentinel/internal/test/runner"
	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/internal/watch/coordinator"
	"github.com/newbpydev/go-sentinel/internal/watch/core"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// DefaultTestExecutor implements the TestExecutor interface using modular components
type DefaultTestExecutor struct {
	testRunner  runner.TestRunnerInterface
	processor   *processor.TestProcessor
	coordinator *coordinator.TestWatchCoordinator
	formatter   colors.FormatterInterface
	icons       colors.IconProviderInterface
	config      *Configuration
}

// NewTestExecutor creates a new test executor with modular components
func NewTestExecutor() TestExecutor {
	return &DefaultTestExecutor{}
}

// SetConfiguration configures the test executor with the application configuration
func (e *DefaultTestExecutor) SetConfiguration(config *Configuration) error {
	if config == nil {
		return models.WrapError(
			fmt.Errorf("configuration cannot be nil"),
			models.ErrorTypeValidation,
			models.SeverityError,
			"failed to configure test executor",
		).WithContext("component", "test_executor")
	}

	e.config = config

	// Initialize UI components with configuration
	e.formatter = colors.NewColorFormatter(config.Colors)
	e.icons = colors.NewIconProvider(config.Visual.Icons != "none")

	// Initialize test runner with configuration
	e.testRunner = runner.NewTestRunner(config.Verbosity > 0, true) // JSON output enabled

	// Initialize test processor for result processing and display
	e.processor = processor.NewTestProcessor(
		os.Stdout,
		e.formatter,
		e.icons,
		80, // Terminal width - could be auto-detected
	)

	return nil
}

// ExecuteSingle executes tests once for the specified packages
func (e *DefaultTestExecutor) ExecuteSingle(ctx context.Context, packages []string, config *Configuration) error {
	if err := e.ensureConfigured(config); err != nil {
		return err
	}

	if len(packages) == 0 {
		packages = []string{"./..."} // Default to current directory
	}

	fmt.Printf("🚀 Running tests with go-sentinel...\n\n")

	// Start timing
	startTime := time.Now()

	// Execute tests for each package
	for _, pkg := range packages {
		if err := e.executePackageTests(ctx, pkg); err != nil {
			return models.WrapError(
				err,
				models.ErrorTypeTestExecution,
				models.SeverityError,
				fmt.Sprintf("failed to run tests for package %s", pkg),
			).WithContext("package", pkg).WithContext("operation", "execute_single")
		}
	}

	// Add separator before final results
	fmt.Fprintln(os.Stdout)

	// Render final summary
	if err := e.processor.RenderResults(true); err != nil {
		return models.WrapError(
			err,
			models.ErrorTypeInternal,
			models.SeverityError,
			"failed to render test results",
		).WithContext("operation", "render_results")
	}

	// Calculate and display timing
	stats := e.processor.GetStats()
	actualDuration := time.Since(startTime)

	fmt.Printf("\n⏱️  Tests completed in %v\n", actualDuration)

	// Exit with appropriate code based on test results
	if stats.FailedTests > 0 {
		return models.NewTestExecutionError(fmt.Sprintf("%d tests failed", stats.FailedTests), fmt.Errorf("%d tests failed", stats.FailedTests)).
			WithContext("failed_tests", fmt.Sprintf("%d", stats.FailedTests)).
			WithContext("total_tests", fmt.Sprintf("%d", stats.TotalTests))
	}

	return nil
}

// ExecuteWatch executes tests in watch mode with file monitoring
func (e *DefaultTestExecutor) ExecuteWatch(ctx context.Context, config *Configuration) error {
	if err := e.ensureConfigured(config); err != nil {
		return err
	}

	fmt.Printf("👀 Starting watch mode...\n")

	// Parse debounce duration
	debounceInterval := 100 * time.Millisecond
	if config.Watch.Debounce != "" {
		if parsed, err := time.ParseDuration(config.Watch.Debounce); err == nil {
			debounceInterval = parsed
		}
	}

	// Create watch options from configuration
	watchOptions := core.WatchOptions{
		Paths:            config.Paths.IncludePatterns,
		IgnorePatterns:   config.Watch.IgnorePatterns,
		TestPatterns:     []string{"*_test.go"},
		Mode:             core.WatchAll,
		DebounceInterval: debounceInterval,
		ClearTerminal:    config.Watch.ClearOnRerun,
		RunOnStart:       config.Watch.RunOnStart,
		Writer:           os.Stdout,
	}

	// Initialize watch coordinator if not already done
	if e.coordinator == nil {
		coordinator, err := coordinator.NewTestWatchCoordinator(watchOptions)
		if err != nil {
			return models.WrapError(
				err,
				models.ErrorTypeDependency,
				models.SeverityError,
				"failed to create watch coordinator",
			).WithContext("operation", "initialize_watch")
		}
		e.coordinator = coordinator
	} else {
		// Configure existing coordinator
		if err := e.coordinator.Configure(watchOptions); err != nil {
			return models.WrapError(
				err,
				models.ErrorTypeConfig,
				models.SeverityError,
				"failed to configure watch coordinator",
			).WithContext("operation", "configure_watch")
		}
	}

	// Run initial tests if configured
	if config.Watch.RunOnStart {
		fmt.Printf("🏃 Running initial tests...\n\n")
		packages := config.Paths.IncludePatterns
		if len(packages) == 0 {
			packages = []string{"./..."}
		}

		if err := e.ExecuteSingle(ctx, packages, config); err != nil {
			fmt.Printf("❌ Initial test run failed: %v\n", err)
		}
	}

	// Start watch mode
	fmt.Printf("👀 Watching for file changes...\n")
	fmt.Printf("Press Ctrl+C to stop\n\n")

	// Start watching - the coordinator will handle file changes and test execution
	return e.coordinator.Start(ctx)
}

// ensureConfigured ensures the test executor is properly configured
func (e *DefaultTestExecutor) ensureConfigured(config *Configuration) error {
	if e.config == nil {
		if config == nil {
			return models.WrapError(
				fmt.Errorf("no configuration provided"),
				models.ErrorTypeConfig,
				models.SeverityError,
				"test executor not configured",
			).WithContext("component", "test_executor")
		}
		return e.SetConfiguration(config)
	}
	return nil
}

// executePackageTests executes tests for a specific package
func (e *DefaultTestExecutor) executePackageTests(ctx context.Context, pkg string) error {
	// Parse timeout from config if provided
	var testCtx context.Context = ctx
	if config := e.config; config != nil {
		if config.Test.Timeout != "" {
			if timeout, err := time.ParseDuration(config.Test.Timeout); err == nil {
				var cancel context.CancelFunc
				testCtx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}
		}
	}

	// Execute test command using the test runner
	testPaths := []string{pkg}
	output, err := e.testRunner.Run(testCtx, testPaths)
	if err != nil {
		return err
	}

	// Process the string output through the processor to parse test results
	if output != "" {
		if err := e.processor.ProcessJSONOutput(output); err != nil {
			return fmt.Errorf("failed to process test output: %w", err)
		}
	}

	return nil
}

// Ensure DefaultTestExecutor implements TestExecutor interface
var _ TestExecutor = (*DefaultTestExecutor)(nil)
