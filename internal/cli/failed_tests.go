package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
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

// getTerminalWidth returns the current terminal width or default
func (r *FailedTestRenderer) getTerminalWidth() int {
	if fd := int(os.Stdout.Fd()); term.IsTerminal(fd) {
		if width, _, err := term.GetSize(fd); err == nil && width > 0 {
			return width
		}
	}
	// Fallback to configured width or default
	if r.width > 0 {
		return r.width
	}
	return 80
}

// RenderFailedTestsHeader renders the header for the failed tests section
func (r *FailedTestRenderer) RenderFailedTestsHeader(failCount int) error {
	// If no failed tests, don't show header
	if failCount <= 0 {
		return nil
	}

	// Create separator line before failed tests header
	width := r.getTerminalWidth()
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

	// Display error type and message (no extra spacing after this)
	errorTypeLine := r.formatter.Red(test.Error.Type) + ": " + r.formatter.Red(test.Error.Message)
	_, err = fmt.Fprintln(r.writer, errorTypeLine)
	if err != nil {
		return err
	}

	// If we have source location information, show it immediately (no spacing before)
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

		// Display source code context immediately if available (no spacing before)
		if len(test.Error.SourceContext) > 0 {
			err = r.renderSourceContext(test)
			if err != nil {
				return err
			}
		}
	}

	// Add spacing after each failed test for better readability
	_, err = fmt.Fprintln(r.writer)
	return err
}

// renderSourceContext renders source code around the error with line numbers and highlighting
func (r *FailedTestRenderer) renderSourceContext(test *TestResult) error {
	if test.Error == nil || test.Error.Location == nil {
		return nil
	}

	// Print source context if available
	if len(test.Error.SourceContext) > 0 {
		// Calculate the starting line number
		startLine := test.Error.Location.Line - test.Error.HighlightedLine

		// Format each line of source code
		for i, line := range test.Error.SourceContext {
			// Calculate the actual line number
			lineNum := startLine + i

			// Format line number and source code (fix duplicate line numbers)
			var lineStr string
			if i == test.Error.HighlightedLine {
				// Highlight the error line (red background for the line number, red text)
				lineStr = fmt.Sprintf("    %s| %s",
					r.formatter.BgRed(r.formatter.White(fmt.Sprintf("%3d", lineNum))),
					r.formatter.Red(line),
				)
			} else {
				// Normal line (gray line number)
				lineStr = fmt.Sprintf("    %s| %s",
					r.formatter.Gray(fmt.Sprintf("%3d", lineNum)),
					line,
				)
			}

			// Write the line
			if _, err := fmt.Fprintln(r.writer, lineStr); err != nil {
				return err
			}

			// If this is the error line, add the error indicator
			if i == test.Error.HighlightedLine {
				// Find position of the error within the line
				errorPos := 0
				if test.Error.Location.Column > 0 {
					errorPos = test.Error.Location.Column - 1
				}

				// Create the error indicator line with proper spacing
				indicatorSpacing := 8 + errorPos // 4 for line number + 2 for "| " + column position
				indicator := fmt.Sprintf("%s%s",
					strings.Repeat(" ", indicatorSpacing),
					r.formatter.Red("^"),
				)
				if _, err := fmt.Fprintln(r.writer, indicator); err != nil {
					return err
				}
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

// formatTestHeader formats the header for a failed test
func (r *FailedTestRenderer) formatTestHeader(test *TestResult) string {
	// Create a red "FAIL" badge
	failBadge := r.formatter.Red("FAIL")

	// Format the test path and name
	testPath := test.Package
	testName := test.Name

	// Return the formatted header
	return fmt.Sprintf("%s %s > %s", failBadge, testPath, testName)
}

// RenderFailedTests renders a section with all failed tests
func (r *FailedTestRenderer) RenderFailedTests(tests []*TestResult) error {
	// Skip if no failed tests
	if len(tests) == 0 {
		return nil
	}

	// Count failed tests
	failedCount := 0
	for _, test := range tests {
		if test.Status == StatusFailed {
			failedCount++
		}
	}

	// Render the header
	if err := r.RenderFailedTestsHeader(failedCount); err != nil {
		return err
	}

	// Print each failed test using the enhanced method with separators
	testNumber := 1
	for _, test := range tests {
		// Skip if not failed
		if test.Status != StatusFailed {
			continue
		}

		// Add separator line with embedded test number (except for first test)
		if testNumber > 1 {
			if err := r.renderTestSeparator(testNumber, failedCount); err != nil {
				return err
			}
		}

		// Use the enhanced RenderFailedTest method that includes source context
		if err := r.RenderFailedTest(test); err != nil {
			return err
		}

		testNumber++
	}

	return nil
}

// renderTestSeparator renders a separator line with embedded test number
func (r *FailedTestRenderer) renderTestSeparator(testNumber, totalTests int) error {
	width := r.getTerminalWidth()

	// Create the test indicator text
	testText := fmt.Sprintf(" Test %d/%d ", testNumber, totalTests)
	textLen := len(testText)

	// Calculate padding for centering
	padding := (width - textLen) / 2
	if padding < 3 {
		padding = 3 // Minimum padding
	}

	// Create left and right parts of the separator
	leftSeparator := strings.Repeat("─", padding)
	rightSeparator := strings.Repeat("─", width-padding-textLen)

	// Combine into full separator with darker red color
	fullSeparator := leftSeparator + testText + rightSeparator

	// Use a darker red color (RGB: 139,0,0 - DarkRed)
	darkRed := "\033[38;2;139;0;0m" // Dark red color
	reset := "\033[0m"

	_, err := fmt.Fprintln(r.writer, darkRed+fullSeparator+reset)
	return err
}
