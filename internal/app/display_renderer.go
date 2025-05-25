// Package app provides display rendering bridging to the modular UI system
package app

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/newbpydev/go-sentinel/internal/test/cache"
	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/internal/ui/display"
	"github.com/newbpydev/go-sentinel/internal/ui/renderer"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// DefaultDisplayRenderer implements the DisplayRenderer interface using modular UI components
type DefaultDisplayRenderer struct {
	writer         io.Writer
	formatter      colors.FormatterInterface
	icons          colors.IconProviderInterface
	renderer       renderer.IncrementalRendererInterface
	testDisplay    display.TestDisplayInterface
	suiteDisplay   display.SuiteDisplayInterface
	summaryDisplay display.SummaryDisplayInterface
	failureDisplay display.FailureDisplayInterface
	cache          cache.CacheInterface
	config         *Configuration
	width          int
}

// NewDisplayRenderer creates a new display renderer with modular UI components
func NewDisplayRenderer() DisplayRenderer {
	return &DefaultDisplayRenderer{
		writer: os.Stdout,
		width:  80,                         // Default terminal width
		cache:  cache.NewTestResultCache(), // Default cache
	}
}

// SetConfiguration configures the display renderer with the application configuration
func (r *DefaultDisplayRenderer) SetConfiguration(config *Configuration) error {
	if config == nil {
		return models.WrapError(
			fmt.Errorf("configuration cannot be nil"),
			models.ErrorTypeValidation,
			models.SeverityError,
			"failed to configure display renderer",
		).WithContext("component", "display_renderer")
	}

	r.config = config

	// Initialize color formatter and icon provider
	r.formatter = colors.NewColorFormatter(config.Colors)
	r.icons = colors.NewIconProvider(config.Visual.Icons != "none")

	// Initialize terminal width if configured
	if config.Visual.TerminalWidth > 0 {
		r.width = config.Visual.TerminalWidth
	}

	// Initialize display components
	r.testDisplay = display.NewTestRenderer(r.writer, r.formatter, r.icons)
	r.suiteDisplay = display.NewSuiteRenderer(r.writer, r.formatter, r.icons, r.width)
	r.summaryDisplay = display.NewSummaryRenderer(r.writer, r.formatter, r.icons, r.width)

	// Initialize failure display with error formatter
	errorFormatter := display.NewErrorFormatter(r.writer, r.formatter, r.width)
	r.failureDisplay = display.NewFailureRenderer(r.writer, r.formatter, r.icons, errorFormatter, r.width)

	// Initialize incremental renderer with cache
	r.renderer = renderer.NewIncrementalRenderer(r.writer, r.formatter, r.icons, r.width, r.cache)

	return nil
}

// RenderResults renders the test results using the modular UI components
func (r *DefaultDisplayRenderer) RenderResults(ctx context.Context) error {
	if err := r.ensureConfigured(); err != nil {
		return err
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// For now, we'll render a simple message indicating the display renderer is working
	// In a full implementation, this would coordinate with the test processor to get results
	// and render them using the appropriate display components

	// This is a placeholder implementation that demonstrates the modular UI usage
	if r.summaryDisplay != nil {
		// Create a simple test run stats for demonstration
		stats := &display.TestRunStats{
			TotalTests:  0,
			PassedTests: 0,
			FailedTests: 0,
			TotalFiles:  0,
			PassedFiles: 0,
			FailedFiles: 0,
		}

		if err := r.summaryDisplay.RenderSummary(stats); err != nil {
			return models.WrapError(
				err,
				models.ErrorTypeInternal,
				models.SeverityError,
				"failed to render test summary",
			).WithContext("component", "summary_display")
		}
	}

	return nil
}

// RenderIncrementalResults renders results incrementally for watch mode
func (r *DefaultDisplayRenderer) RenderIncrementalResults(ctx context.Context, results interface{}) error {
	if err := r.ensureConfigured(); err != nil {
		return err
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Use the incremental renderer for watch mode updates
	if r.renderer != nil {
		// This would render incremental updates
		// For now, it's a placeholder that shows the modular architecture is working
		return nil
	}

	return models.WrapError(
		fmt.Errorf("incremental renderer not initialized"),
		models.ErrorTypeInternal,
		models.SeverityError,
		"incremental renderer not available",
	).WithContext("component", "incremental_renderer")
}

// RenderTestResults renders individual test results
func (r *DefaultDisplayRenderer) RenderTestResults(ctx context.Context, results []*models.TestResult) error {
	if err := r.ensureConfigured(); err != nil {
		return err
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Use the test display component to render individual test results
	if r.testDisplay != nil {
		for _, result := range results {
			// Convert models.TestResult to models.LegacyTestResult for compatibility
			legacyResult := r.convertToLegacyTestResult(result)
			indentLevel := 0 // Default indent level for top-level tests

			if err := r.testDisplay.RenderTestResult(legacyResult, indentLevel); err != nil {
				return models.WrapError(
					err,
					models.ErrorTypeInternal,
					models.SeverityError,
					"failed to render test result",
				).WithContext("test_name", result.Name).WithContext("component", "test_display")
			}
		}
	}

	return nil
}

// convertToLegacyTestResult converts a models.TestResult to models.LegacyTestResult for UI compatibility
func (r *DefaultDisplayRenderer) convertToLegacyTestResult(result *models.TestResult) *models.LegacyTestResult {
	if result == nil {
		return nil
	}

	// Convert output from []string to string
	var outputStr string
	if len(result.Output) > 0 {
		// Join output lines with newlines
		outputStr = ""
		for _, line := range result.Output {
			outputStr += line + "\n"
		}
	}

	// Create a new LegacyTestResult with converted fields
	legacyResult := &models.LegacyTestResult{
		Name:     result.Name,
		Package:  result.Package,
		Status:   result.Status,
		Duration: result.Duration,
		Output:   outputStr,
		Test:     result.Name, // Use Name for Test field
		Parent:   result.Parent,
	}

	// Convert error if present
	if result.Error != nil {
		legacyResult.Error = &models.LegacyTestError{
			Message: result.Error.Message,
			Type:    "TestError", // Default type
			Stack:   "",          // Initialize empty stack
		}

		// Convert stack trace if present
		if len(result.Error.StackTrace) > 0 {
			stackStr := ""
			for _, line := range result.Error.StackTrace {
				stackStr += line + "\n"
			}
			legacyResult.Error.Stack = stackStr
		}

		// Convert source location if present
		if result.Error.SourceFile != "" {
			legacyResult.Error.Location = &models.SourceLocation{
				File: result.Error.SourceFile,
				Line: result.Error.SourceLine,
			}
		}
	}

	// Convert subtests if present
	if len(result.Subtests) > 0 {
		legacyResult.Subtests = make([]*models.LegacyTestResult, len(result.Subtests))
		for i, subtest := range result.Subtests {
			legacyResult.Subtests[i] = r.convertToLegacyTestResult(subtest)
		}
	}

	return legacyResult
}

// RenderFailedTests renders failed test results with detailed error information
func (r *DefaultDisplayRenderer) RenderFailedTests(ctx context.Context, failedTests []*models.TestResult) error {
	if err := r.ensureConfigured(); err != nil {
		return err
	}

	// Check if context is cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Use the failure display component to render failed tests
	if r.failureDisplay != nil && len(failedTests) > 0 {
		if err := r.failureDisplay.RenderFailedTests(failedTests); err != nil {
			return models.WrapError(
				err,
				models.ErrorTypeInternal,
				models.SeverityError,
				"failed to render failed tests",
			).WithContext("failed_count", fmt.Sprintf("%d", len(failedTests))).WithContext("component", "failure_display")
		}
	}

	return nil
}

// ensureConfigured ensures the display renderer is properly configured
func (r *DefaultDisplayRenderer) ensureConfigured() error {
	if r.config == nil {
		return models.WrapError(
			fmt.Errorf("display renderer not configured"),
			models.ErrorTypeConfig,
			models.SeverityError,
			"display renderer requires configuration",
		).WithContext("component", "display_renderer")
	}
	return nil
}

// GetWriter returns the current output writer
func (r *DefaultDisplayRenderer) GetWriter() io.Writer {
	return r.writer
}

// SetWriter sets the output writer
func (r *DefaultDisplayRenderer) SetWriter(writer io.Writer) {
	if writer != nil {
		r.writer = writer

		// Update all display components with the new writer
		if r.config != nil {
			// Re-initialize components with new writer
			r.testDisplay = display.NewTestRenderer(r.writer, r.formatter, r.icons)
			r.suiteDisplay = display.NewSuiteRenderer(r.writer, r.formatter, r.icons, r.width)
			r.summaryDisplay = display.NewSummaryRenderer(r.writer, r.formatter, r.icons, r.width)

			if r.failureDisplay != nil {
				errorFormatter := display.NewErrorFormatter(r.writer, r.formatter, r.width)
				r.failureDisplay = display.NewFailureRenderer(r.writer, r.formatter, r.icons, errorFormatter, r.width)
			}

			if r.renderer != nil {
				r.renderer = renderer.NewIncrementalRenderer(r.writer, r.formatter, r.icons, r.width, r.cache)
			}
		}
	}
}

// Ensure DefaultDisplayRenderer implements DisplayRenderer interface
var _ DisplayRenderer = (*DefaultDisplayRenderer)(nil)
