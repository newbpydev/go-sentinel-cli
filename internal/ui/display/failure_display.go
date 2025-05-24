package display

import (
	"fmt"
	"io"
	"strings"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// FailureDisplayInterface defines the contract for failure display functionality
type FailureDisplayInterface interface {
	// RenderFailedTestsHeader renders the header for the failed tests section
	RenderFailedTestsHeader(failCount int) error

	// RenderFailedTest renders a detailed view of a failed test
	RenderFailedTest(test *models.TestResult) error

	// RenderFailedTests renders multiple failed tests with separators
	RenderFailedTests(tests []*models.TestResult) error

	// SetWidth sets the terminal width for formatting
	SetWidth(width int)

	// GetTerminalWidth returns the current terminal width
	GetTerminalWidth() int
}

// FailureRenderer renders detailed failure information
type FailureRenderer struct {
	writer         io.Writer
	formatter      colors.FormatterInterface
	icons          colors.IconProviderInterface
	errorFormatter ErrorFormatterInterface
	width          int
}

// NewFailureRenderer creates a new FailureRenderer
func NewFailureRenderer(writer io.Writer, formatter colors.FormatterInterface, icons colors.IconProviderInterface, errorFormatter ErrorFormatterInterface, width int) *FailureRenderer {
	if writer == nil {
		panic("writer cannot be nil")
	}
	if formatter == nil {
		panic("formatter cannot be nil")
	}
	if icons == nil {
		panic("icons cannot be nil")
	}
	if errorFormatter == nil {
		panic("errorFormatter cannot be nil")
	}

	return &FailureRenderer{
		writer:         writer,
		formatter:      formatter,
		icons:          icons,
		errorFormatter: errorFormatter,
		width:          width,
	}
}

// NewFailureRendererWithDefaults creates a FailureRenderer with auto-detected defaults
func NewFailureRendererWithDefaults(writer io.Writer) *FailureRenderer {
	formatter := colors.NewAutoColorFormatter()
	icons := colors.NewAutoIconProvider()
	errorFormatter := NewErrorFormatterWithDefaults(writer, formatter)
	return NewFailureRenderer(writer, formatter, icons, errorFormatter, 80)
}

// SetWidth sets the terminal width for formatting
func (r *FailureRenderer) SetWidth(width int) {
	r.width = width
}

// GetTerminalWidth returns the current terminal width or default
func (r *FailureRenderer) GetTerminalWidth() int {
	if r.width > 0 {
		return r.width
	}
	return 80
}

// RenderFailedTestsHeader renders the header for the failed tests section
func (r *FailureRenderer) RenderFailedTestsHeader(failCount int) error {
	// If no failed tests, don't show header
	if failCount <= 0 {
		return nil
	}

	// Create separator line before failed tests header
	width := r.GetTerminalWidth()
	separator := strings.Repeat("─", width)

	// Write separator
	_, err := fmt.Fprintln(r.writer, r.formatter.Red(separator))
	if err != nil {
		return fmt.Errorf("failed to write separator: %w", err)
	}

	// Format "Failed Tests X" header with background color
	headerText := fmt.Sprintf(" Failed Tests %d ", failCount)

	// Calculate padding to center the header
	padding := (width - len(headerText)) / 2
	if padding < 0 {
		padding = 0
	}

	// Create padded header
	paddedHeader := strings.Repeat(" ", padding) + headerText + strings.Repeat(" ", padding)
	if len(paddedHeader) > width {
		paddedHeader = paddedHeader[:width]
	}

	// Style with background color
	styledHeader := r.formatter.BgRed(r.formatter.White(paddedHeader))

	// Write the header
	_, err = fmt.Fprintln(r.writer, styledHeader)
	if err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write closing separator
	_, err = fmt.Fprintln(r.writer, r.formatter.Red(separator))
	if err != nil {
		return fmt.Errorf("failed to write closing separator: %w", err)
	}

	return nil
}

// RenderFailedTest renders a detailed view of a failed test with source context
func (r *FailureRenderer) RenderFailedTest(test *models.TestResult) error {
	// Only proceed if this is a failed test with error information
	if test.Status != models.StatusFailed || test.Error == nil {
		return nil
	}

	// Format and write the failure header (FAIL badge + test name)
	failHeader := r.formatFailHeader(test)
	_, err := fmt.Fprintln(r.writer, failHeader)
	if err != nil {
		return fmt.Errorf("failed to write fail header: %w", err)
	}

	// Display error type and message
	errorTypeLine := r.formatter.Red(test.Error.Type) + ": " + r.formatter.Red(test.Error.Message)
	_, err = fmt.Fprintln(r.writer, errorTypeLine)
	if err != nil {
		return fmt.Errorf("failed to write error type: %w", err)
	}

	// If we have source location information, show it
	if test.Error.SourceFile != "" && test.Error.SourceLine > 0 {
		// Create source location for the error formatter
		location := &models.SourceLocation{
			File:   test.Error.SourceFile,
			Line:   test.Error.SourceLine,
			Column: test.Error.SourceColumn,
		}

		// Format the file:line reference with chevron and make it clickable
		locationRef := r.errorFormatter.FormatClickableLocation(location)
		_, err = fmt.Fprintln(r.writer, locationRef)
		if err != nil {
			return fmt.Errorf("failed to write location: %w", err)
		}

		// Display source code context if available
		if len(test.Error.SourceContext) > 0 {
			err = r.errorFormatter.RenderSourceContext(test)
			if err != nil {
				return fmt.Errorf("failed to render source context: %w", err)
			}
		}
	}

	// Add spacing after each failed test for better readability
	_, err = fmt.Fprintln(r.writer)
	if err != nil {
		return fmt.Errorf("failed to write spacing: %w", err)
	}

	return nil
}

// formatFailHeader formats the failure header with FAIL badge and test name
func (r *FailureRenderer) formatFailHeader(test *models.TestResult) string {
	if test == nil {
		return ""
	}

	// Create FAIL badge with red background
	failBadge := r.formatter.BgRed(r.formatter.White(" FAIL "))

	// Format test name
	testName := test.Name
	if testName == "" {
		testName = "Unknown Test"
	}

	return fmt.Sprintf("%s %s", failBadge, testName)
}

// RenderFailedTests renders multiple failed tests with separators
func (r *FailureRenderer) RenderFailedTests(tests []*models.TestResult) error {
	if len(tests) == 0 {
		return nil
	}

	// Count actual failed tests
	failedTests := make([]*models.TestResult, 0, len(tests))
	for _, test := range tests {
		if test != nil && test.Status == models.StatusFailed {
			failedTests = append(failedTests, test)
		}
	}

	if len(failedTests) == 0 {
		return nil
	}

	// Render header
	if err := r.RenderFailedTestsHeader(len(failedTests)); err != nil {
		return fmt.Errorf("failed to render header: %w", err)
	}

	// Render each failed test with separators
	for i, test := range failedTests {
		// Render test separator (except for the first test)
		if i > 0 {
			if err := r.renderTestSeparator(i+1, len(failedTests)); err != nil {
				return fmt.Errorf("failed to render separator: %w", err)
			}
		}

		// Render the failed test
		if err := r.RenderFailedTest(test); err != nil {
			return fmt.Errorf("failed to render test %d: %w", i+1, err)
		}
	}

	return nil
}

// renderTestSeparator renders a separator between failed tests
func (r *FailureRenderer) renderTestSeparator(testNumber, totalTests int) error {
	// Create a subtle separator line
	width := r.GetTerminalWidth()
	if width > 40 {
		// Use a shorter separator for readability
		separatorLength := width / 3
		separator := strings.Repeat("─", separatorLength)
		centeredSeparator := fmt.Sprintf("%s %s %s",
			strings.Repeat(" ", (width-separatorLength-6)/2),
			r.formatter.Gray(separator),
			strings.Repeat(" ", (width-separatorLength-6)/2),
		)
		_, err := fmt.Fprintln(r.writer, centeredSeparator)
		return err
	}

	// Fallback for narrow terminals
	_, err := fmt.Fprintln(r.writer, r.formatter.Gray("────────────────"))
	return err
}
