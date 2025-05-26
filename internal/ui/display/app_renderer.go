// Package display provides application-specific display rendering components
package display

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// DefaultAppRenderer implements the AppRenderer interface using modular UI components.
// This implementation follows dependency injection principles and the Single Responsibility Principle.
type DefaultAppRenderer struct {
	// Output configuration
	writer io.Writer
	width  int

	// UI component dependencies (injected)
	formatter      FormatterInterface
	icons          IconProviderInterface
	testDisplay    TestDisplayInterface
	suiteDisplay   SuiteDisplayInterface
	summaryDisplay SummaryDisplayInterface
	failureDisplay FailureDisplayInterface
	cache          CacheInterface

	// Configuration
	config *AppConfig
}

// NewAppRenderer creates a new DefaultAppRenderer with default dependencies.
// This follows the Factory pattern and dependency injection principles.
func NewAppRenderer() AppRenderer {
	return &DefaultAppRenderer{
		writer: os.Stdout,
		width:  80, // Default terminal width
	}
}

// NewAppRendererWithDependencies creates a new DefaultAppRenderer with injected dependencies.
// This constructor promotes testability and follows dependency inversion principles.
func NewAppRendererWithDependencies(deps AppRendererDependencies) AppRenderer {
	writer := deps.Writer
	if writer == nil {
		writer = os.Stdout
	}

	width := deps.TerminalWidth
	if width <= 0 {
		width = 80
	}

	return &DefaultAppRenderer{
		writer:         writer,
		width:          width,
		formatter:      deps.ColorFormatter,
		icons:          deps.IconProvider,
		testDisplay:    deps.TestDisplay,
		suiteDisplay:   deps.SuiteDisplay,
		summaryDisplay: deps.SummaryDisplay,
		failureDisplay: deps.FailureDisplay,
		cache:          deps.Cache,
	}
}

// SetConfiguration configures the renderer with application settings.
// This method initializes UI components based on configuration.
func (r *DefaultAppRenderer) SetConfiguration(config *AppConfig) error {
	if config == nil {
		return models.WrapError(
			fmt.Errorf("configuration cannot be nil"),
			models.ErrorTypeValidation,
			models.SeverityError,
			"failed to configure app renderer",
		).WithContext("component", "app_renderer")
	}

	r.config = config

	// Update terminal width if configured
	if config.Visual.TerminalWidth > 0 {
		r.width = config.Visual.TerminalWidth
	}

	// Initialize UI components only if not already injected
	if r.formatter == nil {
		r.formatter = &DefaultColorFormatter{enabled: config.Colors}
	}

	if r.icons == nil {
		r.icons = &DefaultIconProvider{enabled: config.Visual.Icons != "none"}
	}

	// Note: Other display components (testDisplay, suiteDisplay, etc.) would be
	// initialized here in a full implementation, but they require imports from
	// the existing display package which would need refactoring to avoid circular deps.
	// For now, we'll use placeholder implementations.

	return nil
}

// RenderResults renders test execution results to the configured output.
// This method coordinates with UI components to provide structured output.
func (r *DefaultAppRenderer) RenderResults(ctx context.Context) error {
	if err := r.ensureConfigured(); err != nil {
		return err
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Render a basic summary for now
	// In a full implementation, this would coordinate with test processors
	// to get actual results and render them using the display components
	return r.renderBasicSummary()
}

// RenderTestResults renders individual test results with proper formatting.
func (r *DefaultAppRenderer) RenderTestResults(ctx context.Context, results []*models.TestResult) error {
	if err := r.ensureConfigured(); err != nil {
		return err
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Render each test result
	for _, result := range results {
		if err := r.renderSingleTestResult(result); err != nil {
			return models.WrapError(
				err,
				models.ErrorTypeInternal,
				models.SeverityError,
				"failed to render test result",
			).WithContext("test_name", result.Name)
		}
	}

	return nil
}

// RenderIncrementalResults renders results incrementally for watch mode.
func (r *DefaultAppRenderer) RenderIncrementalResults(ctx context.Context, results interface{}) error {
	if err := r.ensureConfigured(); err != nil {
		return err
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Placeholder implementation for incremental rendering
	fmt.Fprintf(r.writer, "%s Incremental update received\n", r.icons.Info())
	return nil
}

// RenderFailedTests renders only the failed test results with detailed output.
func (r *DefaultAppRenderer) RenderFailedTests(ctx context.Context, failedTests []*models.TestResult) error {
	if err := r.ensureConfigured(); err != nil {
		return err
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if len(failedTests) == 0 {
		return nil
	}

	// Render header for failed tests
	fmt.Fprintf(r.writer, "\n%s Failed Tests:\n", r.icons.Error())

	// Render each failed test
	for _, test := range failedTests {
		if err := r.renderFailedTestDetail(test); err != nil {
			return models.WrapError(
				err,
				models.ErrorTypeInternal,
				models.SeverityError,
				"failed to render failed test",
			).WithContext("test_name", test.Name)
		}
	}

	return nil
}

// SetWriter sets the output writer for rendering.
func (r *DefaultAppRenderer) SetWriter(writer io.Writer) {
	if writer != nil {
		r.writer = writer
	}
}

// GetWriter returns the current output writer.
func (r *DefaultAppRenderer) GetWriter() io.Writer {
	return r.writer
}

// ensureConfigured verifies that the renderer is properly configured.
func (r *DefaultAppRenderer) ensureConfigured() error {
	if r.config == nil {
		return models.WrapError(
			fmt.Errorf("renderer not configured"),
			models.ErrorTypeValidation,
			models.SeverityError,
			"app renderer requires configuration before use",
		).WithContext("component", "app_renderer")
	}
	return nil
}

// renderBasicSummary renders a basic test execution summary.
func (r *DefaultAppRenderer) renderBasicSummary() error {
	fmt.Fprintf(r.writer, "%s Test Execution Summary\n", r.icons.Info())
	fmt.Fprintf(r.writer, "%s All configured and ready\n", r.icons.Success())
	return nil
}

// renderSingleTestResult renders an individual test result.
func (r *DefaultAppRenderer) renderSingleTestResult(result *models.TestResult) error {
	icon := r.getIconForStatus(result.Status)
	name := r.formatter.FormatInfo(result.Name)

	fmt.Fprintf(r.writer, "%s %s\n", icon, name)

	// Show output if verbose or failed
	if result.Status == models.TestStatusFailed && len(result.Output) > 0 {
		for _, line := range result.Output {
			fmt.Fprintf(r.writer, "    %s\n", r.formatter.FormatMuted(line))
		}
	}

	return nil
}

// renderFailedTestDetail renders detailed information for a failed test.
func (r *DefaultAppRenderer) renderFailedTestDetail(test *models.TestResult) error {
	// Render test name with error styling
	testName := r.formatter.FormatError(fmt.Sprintf("  %s %s", r.icons.Error(), test.Name))
	fmt.Fprintf(r.writer, "%s\n", testName)

	// Render package info
	if test.Package != "" {
		pkg := r.formatter.FormatMuted(fmt.Sprintf("    Package: %s", test.Package))
		fmt.Fprintf(r.writer, "%s\n", pkg)
	}

	// Render test output
	if len(test.Output) > 0 {
		fmt.Fprintf(r.writer, "    Output:\n")
		for _, line := range test.Output {
			fmt.Fprintf(r.writer, "      %s\n", r.formatter.FormatMuted(line))
		}
	}

	fmt.Fprintf(r.writer, "\n")
	return nil
}

// getIconForStatus returns the appropriate icon for a test status.
func (r *DefaultAppRenderer) getIconForStatus(status models.TestStatus) string {
	switch status {
	case models.TestStatusPassed:
		return r.icons.Success()
	case models.TestStatusFailed:
		return r.icons.Error()
	case models.TestStatusSkipped:
		return r.icons.Skipped()
	default:
		return r.icons.Running()
	}
}

// DefaultColorFormatter provides basic color formatting.
// This is a placeholder implementation that would be replaced by proper color formatting.
type DefaultColorFormatter struct {
	enabled bool
}

func (f *DefaultColorFormatter) FormatSuccess(text string) string {
	if f.enabled {
		return fmt.Sprintf("\033[32m%s\033[0m", text) // Green
	}
	return text
}

func (f *DefaultColorFormatter) FormatError(text string) string {
	if f.enabled {
		return fmt.Sprintf("\033[31m%s\033[0m", text) // Red
	}
	return text
}

func (f *DefaultColorFormatter) FormatWarning(text string) string {
	if f.enabled {
		return fmt.Sprintf("\033[33m%s\033[0m", text) // Yellow
	}
	return text
}

func (f *DefaultColorFormatter) FormatInfo(text string) string {
	if f.enabled {
		return fmt.Sprintf("\033[36m%s\033[0m", text) // Cyan
	}
	return text
}

func (f *DefaultColorFormatter) FormatMuted(text string) string {
	if f.enabled {
		return fmt.Sprintf("\033[90m%s\033[0m", text) // Gray
	}
	return text
}

func (f *DefaultColorFormatter) IsColorEnabled() bool {
	return f.enabled
}

// DefaultIconProvider provides basic icon support.
// This is a placeholder implementation that would be replaced by proper icon handling.
type DefaultIconProvider struct {
	enabled bool
}

func (p *DefaultIconProvider) Success() string {
	if p.enabled {
		return "✅"
	}
	return "[PASS]"
}

func (p *DefaultIconProvider) Error() string {
	if p.enabled {
		return "❌"
	}
	return "[FAIL]"
}

func (p *DefaultIconProvider) Warning() string {
	if p.enabled {
		return "⚠️"
	}
	return "[WARN]"
}

func (p *DefaultIconProvider) Info() string {
	if p.enabled {
		return "ℹ️"
	}
	return "[INFO]"
}

func (p *DefaultIconProvider) Running() string {
	if p.enabled {
		return "🔄"
	}
	return "[RUN]"
}

func (p *DefaultIconProvider) Skipped() string {
	if p.enabled {
		return "⏭️"
	}
	return "[SKIP]"
}

// Ensure DefaultAppRenderer implements AppRenderer interface
var _ AppRenderer = (*DefaultAppRenderer)(nil)
