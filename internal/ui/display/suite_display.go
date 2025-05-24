// Package display provides test suite display formatting
package display

import (
	"fmt"
	"io"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// SuiteDisplayInterface defines the interface for test suite display
type SuiteDisplayInterface interface {
	RenderSuite(suite *models.TestSuite, autoCollapse bool) error
	SetAutoCollapse(enabled bool)
	SetWidth(width int)
}

// SuiteRenderer renders test suites with collapsing/expanding and implements SuiteDisplayInterface
type SuiteRenderer struct {
	writer       io.Writer
	formatter    colors.FormatterInterface
	icons        colors.IconProviderInterface
	width        int
	autoCollapse bool
	header       *HeaderRenderer
	test         *TestRenderer
}

// NewSuiteRenderer creates a new SuiteRenderer
func NewSuiteRenderer(writer io.Writer, formatter colors.FormatterInterface, icons colors.IconProviderInterface, width int) *SuiteRenderer {
	return &SuiteRenderer{
		writer:       writer,
		formatter:    formatter,
		icons:        icons,
		width:        width,
		autoCollapse: false,
		header:       NewHeaderRenderer(writer, formatter, icons, width),
		test:         NewTestRenderer(writer, formatter, icons),
	}
}

// NewSuiteRendererWithDefaults creates a SuiteRenderer with auto-detected color and icon support
func NewSuiteRendererWithDefaults(writer io.Writer, width int) *SuiteRenderer {
	formatter := colors.NewAutoColorFormatter()
	icons := colors.NewAutoIconProvider()

	return &SuiteRenderer{
		writer:       writer,
		formatter:    formatter,
		icons:        icons,
		width:        width,
		autoCollapse: false,
		header:       NewHeaderRenderer(writer, formatter, icons, width),
		test:         NewTestRenderer(writer, formatter, icons),
	}
}

// SetAutoCollapse sets whether suites should auto-collapse when all tests pass
func (r *SuiteRenderer) SetAutoCollapse(enabled bool) {
	r.autoCollapse = enabled
}

// SetWidth sets the display width for formatting
func (r *SuiteRenderer) SetWidth(width int) {
	r.width = width
	if r.header != nil {
		r.header.SetWidth(width)
	}
}

// RenderSuite renders a test suite to the output writer
func (r *SuiteRenderer) RenderSuite(suite *models.TestSuite, autoCollapse bool) error {
	if suite == nil {
		return fmt.Errorf("test suite cannot be nil")
	}

	// Use parameter or instance setting for auto-collapse
	shouldCollapse := autoCollapse || (r.autoCollapse && suite.FailedCount == 0)

	// Format and write the suite header
	header := r.formatSuiteHeader(suite)
	if _, err := fmt.Fprintln(r.writer, header); err != nil {
		return fmt.Errorf("failed to write suite header: %w", err)
	}

	// If collapsed, just show a summary line
	if shouldCollapse {
		return r.renderCollapsedSummary(suite)
	}

	// If not collapsed, render each test
	return r.renderExpandedSuite(suite)
}

// renderCollapsedSummary renders a collapsed suite summary
func (r *SuiteRenderer) renderCollapsedSummary(suite *models.TestSuite) error {
	statusIcon := r.formatter.Green(r.icons.CheckMark())
	summaryText := r.formatter.Green(fmt.Sprintf("Suite passed (%d tests)", suite.TestCount))

	_, err := fmt.Fprintf(r.writer, "  %s %s\n", statusIcon, summaryText)
	if err != nil {
		return fmt.Errorf("failed to write collapsed summary: %w", err)
	}

	return nil
}

// renderExpandedSuite renders all tests in the suite
func (r *SuiteRenderer) renderExpandedSuite(suite *models.TestSuite) error {
	testRenderer := NewTestRenderer(r.writer, r.formatter, r.icons)

	for i, test := range suite.Tests {
		if err := testRenderer.RenderTestResult(test, 1); err != nil {
			return fmt.Errorf("failed to render test %d: %w", i, err)
		}
	}

	return nil
}

// formatSuiteHeader creates a header line for a test suite
func (r *SuiteRenderer) formatSuiteHeader(suite *models.TestSuite) string {
	// Get formatted file path
	formattedPath := FormatFilePath(r.formatter, suite.FilePath)

	// Format test counts with appropriate colors
	testCountStr := r.formatTestCounts(suite)

	// Format duration and memory
	durationStr := r.formatter.Dim(fmt.Sprintf("%dms", suite.Duration.Milliseconds()))
	memoryStr := r.formatter.Dim(fmt.Sprintf("%d MB heap used", suite.MemoryUsage/(1024*1024)))

	// Combine all parts
	return fmt.Sprintf("%s %s %s %s",
		formattedPath,
		testCountStr,
		durationStr,
		memoryStr,
	)
}

// formatTestCounts formats the test count information with appropriate colors
func (r *SuiteRenderer) formatTestCounts(suite *models.TestSuite) string {
	if suite.FailedCount > 0 {
		// Red if any tests failed
		return fmt.Sprintf("(%d tests | %s)",
			suite.TestCount,
			r.formatter.Red(fmt.Sprintf("%d failed", suite.FailedCount)),
		)
	}

	if suite.SkippedCount > 0 && suite.SkippedCount == suite.TestCount {
		// Yellow if all tests skipped
		return fmt.Sprintf("(%d tests | %s)",
			suite.TestCount,
			r.formatter.Yellow(fmt.Sprintf("%d skipped", suite.SkippedCount)),
		)
	}

	// Green for passed tests, with optional skipped count
	skippedPart := ""
	if suite.SkippedCount > 0 {
		skippedPart = " | " + r.formatter.Yellow(fmt.Sprintf("%d skipped", suite.SkippedCount))
	}

	return fmt.Sprintf("(%d tests%s)", suite.TestCount, skippedPart)
}

// RenderSuiteWithOptions renders a suite with specific display options
func (r *SuiteRenderer) RenderSuiteWithOptions(suite *models.TestSuite, options SuiteDisplayOptions) error {
	if suite == nil {
		return fmt.Errorf("test suite cannot be nil")
	}

	// Apply options
	oldCollapse := r.autoCollapse
	oldWidth := r.width

	if options.AutoCollapse != nil {
		r.autoCollapse = *options.AutoCollapse
	}
	if options.Width > 0 {
		r.SetWidth(options.Width)
	}

	// Render the suite
	err := r.RenderSuite(suite, false) // Use instance settings

	// Restore original settings
	r.autoCollapse = oldCollapse
	r.SetWidth(oldWidth)

	return err
}

// SuiteDisplayOptions contains options for suite display
type SuiteDisplayOptions struct {
	AutoCollapse *bool // Use pointer to detect if it was set
	Width        int
	ShowMemory   bool
	ShowTiming   bool
}

// RenderSuiteSummary renders just the suite header without individual tests
func (r *SuiteRenderer) RenderSuiteSummary(suite *models.TestSuite) error {
	if suite == nil {
		return fmt.Errorf("test suite cannot be nil")
	}

	header := r.formatSuiteHeader(suite)
	if _, err := fmt.Fprintln(r.writer, header); err != nil {
		return fmt.Errorf("failed to write suite summary: %w", err)
	}

	return nil
}

// RenderMultipleSuites renders multiple test suites
func (r *SuiteRenderer) RenderMultipleSuites(suites []*models.TestSuite, autoCollapse bool) error {
	for i, suite := range suites {
		if err := r.RenderSuite(suite, autoCollapse); err != nil {
			return fmt.Errorf("failed to render suite %d: %w", i, err)
		}

		// Add spacing between suites (except for the last one)
		if i < len(suites)-1 {
			if _, err := fmt.Fprintln(r.writer); err != nil {
				return fmt.Errorf("failed to write suite separator: %w", err)
			}
		}
	}

	return nil
}
