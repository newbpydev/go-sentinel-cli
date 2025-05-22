package cli

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// FailedTestRenderer renders detailed failure information
type FailedTestRenderer struct {
	writer    io.Writer
	formatter *ColorFormatter
	icons     *IconProvider
	width     int
}

// NewFailedTestRenderer creates a new FailedTestRenderer
func NewFailedTestRenderer(writer io.Writer, formatter *ColorFormatter, icons *IconProvider, width int) *FailedTestRenderer {
	return &FailedTestRenderer{
		writer:    writer,
		formatter: formatter,
		icons:     icons,
		width:     width,
	}
}

// RenderFailedTestsHeader renders the header for the failed tests section
func (r *FailedTestRenderer) RenderFailedTestsHeader(failCount int) error {
	// If no failed tests, don't show header
	if failCount <= 0 {
		return nil
	}

	// Create separator line before failed tests header
	width := r.width
	if width <= 0 {
		width = 80 // Default width
	}
	separator := strings.Repeat("─", width)

	// Write separator
	_, err := fmt.Fprintln(r.writer, r.formatter.Red(separator))
	if err != nil {
		return err
	}

	// Format "Failed Tests X" header with background color
	// Use a centered format with background color for the entire header
	// This mimics the Vitest output
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
		return err
	}

	// Write closing separator
	_, err = fmt.Fprintln(r.writer, r.formatter.Red(separator))

	return err
}

// RenderFailedTest renders a detailed view of a failed test with source context
func (r *FailedTestRenderer) RenderFailedTest(test *TestResult) error {
	// Only proceed if this is a failed test with error information
	if test.Status != StatusFailed || test.Error == nil {
		return nil
	}

	// Format and write the failure header (FAIL badge + test name)
	failHeader := r.formatFailHeader(test)
	_, err := fmt.Fprintln(r.writer, failHeader)
	if err != nil {
		return err
	}

	// Display error type and message
	errorTypeLine := r.formatter.Red(test.Error.Type) + ": " + r.formatter.Red(test.Error.Message)
	_, err = fmt.Fprintln(r.writer, errorTypeLine)
	if err != nil {
		return err
	}

	// If we have source location information, show it
	if test.Error.Location != nil {
		// Format the file:line reference with chevron
		locationRef := fmt.Sprintf("↳ %s:%d:%d",
			r.formatter.Cyan(test.Error.Location.File),
			test.Error.Location.Line,
			test.Error.Location.Column,
		)
		_, err = fmt.Fprintln(r.writer, locationRef)
		if err != nil {
			return err
		}

		// Display source code context if available
		if test.Error.SourceContext != nil && len(test.Error.SourceContext) > 0 {
			err = r.renderSourceContext(test.Error)
			if err != nil {
				return err
			}
		}
	}

	// Add a blank line after each failed test for better readability
	_, err = fmt.Fprintln(r.writer)
	return err
}

// renderSourceContext renders source code around the error with line numbers and highlighting
func (r *FailedTestRenderer) renderSourceContext(err *TestError) error {
	// Only proceed if we have source context and location
	if err.SourceContext == nil || len(err.SourceContext) == 0 || err.Location == nil {
		return nil
	}

	// Calculate the starting line number
	startLine := err.Location.Line - err.HighlightedLine

	// Format each line of source code
	for i, line := range err.SourceContext {
		// Calculate the actual line number
		lineNum := startLine + i

		// Format line number and source code
		var lineStr string
		if i == err.HighlightedLine {
			// Highlight the error line (red)
			lineStr = fmt.Sprintf("  %s| %s",
				r.formatter.Dim(fmt.Sprintf("%2d", lineNum)),
				r.formatter.Red(line),
			)
		} else {
			// Regular line (dimmed)
			lineStr = fmt.Sprintf("  %s| %s",
				r.formatter.Dim(fmt.Sprintf("%2d", lineNum)),
				line,
			)
		}

		// Write the line
		_, err2 := fmt.Fprintln(r.writer, lineStr)
		if err2 != nil {
			return err2
		}

		// If this is the error line, add the error indicator
		if i == err.HighlightedLine {
			// Find position of the error within the line
			errorPos := 0
			if err.Location.Column > 0 {
				errorPos = err.Location.Column - 1
			}

			// Create the indicator line with an arrow pointing to the error
			indicator := fmt.Sprintf("  %s%s",
				strings.Repeat(" ", 3+errorPos), // Padding + line number width
				r.formatter.Red("^"),            // Error indicator arrow
			)
			_, err2 := fmt.Fprintln(r.writer, indicator)
			if err2 != nil {
				return err2
			}
		}
	}

	return nil
}

// formatFailHeader formats the header for a single failed test
func (r *FailedTestRenderer) formatFailHeader(test *TestResult) string {
	// Create a "FAIL" badge
	failBadge := r.formatter.BgRed(r.formatter.White(" FAIL "))

	// Format test file path and name
	testPath := ""
	if test.Error != nil && test.Error.Location != nil {
		testPath = test.Error.Location.File
	}

	// Format the full test name with path components
	testName := test.Name

	return fmt.Sprintf("%s %s > %s", failBadge, testPath, testName)
}

// Helper function to convert int to string
func intToString(i int) string {
	return strconv.Itoa(i)
}

// RenderFailedTests renders a section with all failed tests
func (r *FailedTestRenderer) RenderFailedTests(tests []*TestResult) error {
	// Filter out non-failed tests
	var failedTests []*TestResult
	for _, test := range tests {
		if test.Status == StatusFailed && test.Error != nil {
			failedTests = append(failedTests, test)
		}
	}

	// If no failed tests, don't show anything
	if len(failedTests) == 0 {
		return nil
	}

	// Render the header
	err := r.RenderFailedTestsHeader(len(failedTests))
	if err != nil {
		return err
	}

	// Render each failed test
	for _, test := range failedTests {
		err = r.RenderFailedTest(test)
		if err != nil {
			return err
		}
	}

	return nil
}
