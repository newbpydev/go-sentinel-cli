// Package app provides application orchestration following modular architecture
package app

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/newbpydev/go-sentinel/internal/config"
	"github.com/newbpydev/go-sentinel/internal/test/runner"
	"github.com/newbpydev/go-sentinel/internal/ui/display"
)

// ApplicationControllerImpl implements application orchestration following modular architecture
type ApplicationControllerImpl struct {
	argParser    config.ArgParser
	testExecutor runner.TestExecutor
	renderer     display.Renderer
}

// NewApplicationController creates a new application controller with modular dependencies
func NewApplicationController() ApplicationController {
	// Create a temporary basic renderer implementation
	return &ApplicationControllerImpl{
		argParser:    config.NewArgParser(),
		testExecutor: runner.NewExecutor(),
		renderer:     NewBasicRenderer(os.Stdout), // Use our own basic renderer
	}
}

// Initialize implements ApplicationController interface
func (c *ApplicationControllerImpl) Initialize() error {
	// Basic initialization - no special setup needed for now
	return nil
}

// Shutdown implements ApplicationController interface
func (c *ApplicationControllerImpl) Shutdown(ctx context.Context) error {
	// Basic shutdown - no special cleanup needed for now
	return nil
}

// Run executes the application with the given arguments following modular architecture
func (c *ApplicationControllerImpl) Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command specified")
	}

	// Validate command
	command := args[0]
	if command != "run" {
		return fmt.Errorf("unknown command: %s", command)
	}

	// Parse arguments using dedicated config package
	parsedArgs, err := c.argParser.Parse(args[1:])
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Convert to execution options
	options, err := c.convertToExecutionOptions(parsedArgs)
	if err != nil {
		return fmt.Errorf("failed to convert execution options: %w", err)
	}

	// Set working directory to project root to ensure relative paths work correctly
	// This is important when running from subdirectories like internal/app/
	if options.WorkingDirectory == "" {
		// Try to find the project root by looking for go.mod
		if wd, err := findProjectRoot(); err == nil {
			options.WorkingDirectory = wd
		}
	}

	// Validate packages
	packages := parsedArgs.Packages
	if len(packages) == 0 {
		packages = []string{"."} // Default to current directory
	}

	// Execute tests using dedicated test runner
	ctx := context.Background()
	result, err := c.testExecutor.Execute(ctx, packages, options)
	if err != nil {
		// Check for package not found errors (should be returned as-is)
		if c.isPackageNotFoundError(err) {
			return fmt.Errorf("package not found: %w", err)
		}
		// Check for common "no tests" scenarios
		if c.isNoTestsError(err) {
			return fmt.Errorf("no tests found in specified packages")
		}
		return fmt.Errorf("test execution failed: %w", err)
	}

	// Check for package-level errors (like directory not found)
	for _, pkg := range result.Packages {
		if pkg.Error != nil {
			// Check the output for error messages since the error might just be "exit status 1"
			if c.isPackageNotFoundErrorInOutput(pkg.Output) || c.isPackageNotFoundError(pkg.Error) {
				return fmt.Errorf("package not found: %w", pkg.Error)
			}
		}
	}

	// Convert and display results using dedicated UI package
	displayResults := c.convertToDisplayResults(result)
	if err := c.renderer.RenderResults(ctx, displayResults); err != nil {
		return fmt.Errorf("failed to render results: %w", err)
	}

	return nil
}

// convertToExecutionOptions converts parsed CLI args to test runner execution options
func (c *ApplicationControllerImpl) convertToExecutionOptions(args *config.Args) (*runner.ExecutionOptions, error) {
	options := &runner.ExecutionOptions{
		JSONOutput: true, // Always use JSON for parsing
		Verbose:    args.Verbosity > 0,
		Coverage:   args.CoverageMode != "",
		Parallel:   args.Parallel,
		Args:       []string{},
		Env:        make(map[string]string),
	}

	// Add additional args if needed
	if args.TestPattern != "" {
		options.Args = append(options.Args, "-run", args.TestPattern)
	}

	return options, nil
}

// convertToDisplayResults converts test runner results to display format
func (c *ApplicationControllerImpl) convertToDisplayResults(result *runner.ExecutionResult) *display.Results {
	displayResults := &display.Results{
		Packages:  make([]*display.PackageResult, 0, len(result.Packages)),
		Duration:  result.TotalDuration,
		StartTime: result.StartTime,
		EndTime:   result.EndTime,
		Summary: &display.TestSummary{
			TotalTests:         result.TotalTests,
			PassedTests:        result.PassedTests,
			FailedTests:        result.FailedTests,
			SkippedTests:       result.SkippedTests,
			TotalDuration:      result.TotalDuration,
			CoveragePercentage: result.Coverage,
			PackageCount:       len(result.Packages),
			Success:            result.Success,
		},
	}

	// Convert package results
	for _, pkg := range result.Packages {
		displayPkg := &display.PackageResult{
			Package:  pkg.Package,
			Success:  pkg.Success,
			Duration: pkg.Duration,
			Coverage: pkg.Coverage,
			Output:   pkg.Output,
			Tests:    make([]*display.TestResult, 0, len(pkg.Tests)),
		}

		// Convert individual test results
		for _, test := range pkg.Tests {
			displayTest := &display.TestResult{
				Name:     test.Name,
				Package:  test.Package,
				Status:   display.TestStatus(test.Status),
				Duration: test.Duration,
				Output:   []string{test.Output},
			}

			if test.Error != "" {
				displayTest.Error = &display.TestError{
					Message: test.Error,
				}
			}

			displayPkg.Tests = append(displayPkg.Tests, displayTest)
		}

		displayResults.Packages = append(displayResults.Packages, displayPkg)
	}

	return displayResults
}

// isPackageNotFoundError checks if the error indicates a package was not found
func (c *ApplicationControllerImpl) isPackageNotFoundError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "package not found") ||
		contains(errStr, "cannot find package") ||
		contains(errStr, "no such file or directory") ||
		contains(errStr, "directory not found")
}

// isPackageNotFoundErrorInOutput checks if the output contains package not found messages
func (c *ApplicationControllerImpl) isPackageNotFoundErrorInOutput(output string) bool {
	// Look for specific patterns that indicate the package directory doesn't exist
	return (contains(output, "directory not found") && contains(output, "stat ")) ||
		contains(output, "package not found") ||
		contains(output, "cannot find package") ||
		contains(output, "no such file or directory")
}

// isNoTestsError checks if the error indicates no tests were found (but package exists)
func (c *ApplicationControllerImpl) isNoTestsError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "no test files") ||
		contains(errStr, "no tests to run") ||
		contains(errStr, "no tests found")
}

// contains checks if a string contains a substring (case-insensitive helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// findProjectRoot finds the project root directory by looking for go.mod
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree looking for go.mod
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached the root directory
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("go.mod not found")
}

// BasicRenderer provides a simple implementation of display.Renderer interface
type BasicRenderer struct {
	output io.Writer
}

// NewBasicRenderer creates a new BasicRenderer
func NewBasicRenderer(output io.Writer) display.Renderer {
	return &BasicRenderer{output: output}
}

// RenderResults implements display.Renderer interface
func (r *BasicRenderer) RenderResults(ctx context.Context, results *display.Results) error {
	if results == nil || results.Summary == nil {
		fmt.Fprintln(r.output, "❌ No test results available")
		return nil
	}

	summary := results.Summary
	fmt.Fprintln(r.output, "🚀 Test Execution Summary")
	fmt.Fprintf(r.output, "📦 Packages: %d\n", summary.PackageCount)
	fmt.Fprintf(r.output, "🧪 Total Tests: %d\n", summary.TotalTests)
	fmt.Fprintf(r.output, "✅ Passed: %d\n", summary.PassedTests)
	fmt.Fprintf(r.output, "❌ Failed: %d\n", summary.FailedTests)
	fmt.Fprintf(r.output, "⏭️  Skipped: %d\n", summary.SkippedTests)
	fmt.Fprintf(r.output, "⏱️  Duration: %v\n", summary.TotalDuration)

	if summary.TotalTests == 0 {
		fmt.Fprintln(r.output, "ℹ️  No tests found")
	} else if summary.Success {
		fmt.Fprintln(r.output, "🎉 All tests passed!")
	} else {
		fmt.Fprintln(r.output, "💥 Some tests failed!")
	}

	return nil
}

// RenderProgress implements display.Renderer interface
func (r *BasicRenderer) RenderProgress(ctx context.Context, progress *display.ProgressUpdate) error {
	// Simple progress implementation - just print status
	if progress != nil && progress.Status != "" {
		fmt.Fprintf(r.output, "🔄 %s\n", progress.Status)
	}
	return nil
}

// RenderSummary implements display.Renderer interface
func (r *BasicRenderer) RenderSummary(ctx context.Context, summary *display.TestSummary) error {
	// Already handled in RenderResults
	return nil
}

// Clear implements display.Renderer interface
func (r *BasicRenderer) Clear() error {
	// Simple implementation - just print newline
	fmt.Fprintln(r.output)
	return nil
}

// SetConfiguration implements display.Renderer interface
func (r *BasicRenderer) SetConfiguration(config *display.Config) error {
	// Basic renderer doesn't need configuration
	return nil
}
