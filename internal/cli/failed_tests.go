package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
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
		// Format the file:line reference with chevron and make it clickable
		locationRef := r.formatClickableLocation(test.Error.Location)
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

// formatClickableLocation formats a clickable file location reference
func (r *FailedTestRenderer) formatClickableLocation(location *SourceLocation) string {
	if location == nil {
		return ""
	}

	// Create absolute file path for clickable link
	absolutePath := location.File

	// Handle different cases for path resolution
	if !filepath.IsAbs(absolutePath) {
		// Get current working directory
		if wd, err := os.Getwd(); err == nil {
			// If the file path doesn't contain directory separators, check if it's in stress_tests
			if !strings.Contains(absolutePath, string(filepath.Separator)) && !strings.Contains(absolutePath, "/") {
				// Check if file exists in stress_tests directory
				stressTestPath := filepath.Join(wd, "stress_tests", absolutePath)
				if _, err := os.Stat(stressTestPath); err == nil {
					absolutePath = stressTestPath
				} else {
					// Default to current directory
					absolutePath = filepath.Join(wd, absolutePath)
				}
			} else {
				// File path has directory info, resolve relative to working directory
				absolutePath = filepath.Join(wd, absolutePath)
			}
		}
	}

	// Clean the path and ensure proper format
	absolutePath = filepath.Clean(absolutePath)

	// Create display text (show relative path for readability)
	displayText := fmt.Sprintf("%s:%d:%d", location.File, location.Line, location.Column)

	// Convert backslashes to forward slashes for URL format
	urlPath := strings.ReplaceAll(absolutePath, "\\", "/")

	// Multi-layered approach for maximum compatibility:
	// Try different URL schemes based on what's available and most likely to work
	var linkUrl string

	// Strategy 1: Try Cursor-specific URL scheme (most reliable when it works)
	if r.isCursorInstalled() {
		// Use the format recommended in Cursor forum: cursor://file/path:line:column
		linkUrl = fmt.Sprintf("cursor://file/%s:%d:%d", urlPath, location.Line, location.Column)
	} else if r.isVSCodeInstalled() {
		// Strategy 2: Use VS Code URL scheme if available
		linkUrl = fmt.Sprintf("vscode://file/%s:%d:%d", urlPath, location.Line, location.Column)
	} else {
		// Strategy 3: Use a pragmatic fallback approach
		// Create a command-style URL that could be processed by system handlers
		// This format works with many systems and can be extended
		linkUrl = fmt.Sprintf("file:///%s#line=%d&column=%d", urlPath, location.Line, location.Column)
	}

	// Use OSC 8 hyperlink escape sequence for terminal link support
	// Format: ESC]8;;URL ESC\\ text ESC]8;; ESC\\
	clickableText := fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\",
		linkUrl,
		r.formatter.Cyan(displayText))

	return fmt.Sprintf("↳ %s", clickableText)
}

// isCursorInstalled checks if Cursor is installed and available
func (r *FailedTestRenderer) isCursorInstalled() bool {
	// Check if cursor command is available in PATH first (most reliable)
	if _, err := exec.LookPath("cursor"); err == nil {
		return true
	}

	// Check common Cursor installation paths on Windows
	username := os.Getenv("USERNAME")
	if username != "" {
		commonPaths := []string{
			"C:\\Users\\" + username + "\\AppData\\Local\\Programs\\cursor\\Cursor.exe",
			"C:\\Users\\" + username + "\\AppData\\Local\\Programs\\cursor\\resources\\app\\bin\\cursor.cmd",
		}

		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				return true
			}
		}
	}

	// Fallback paths
	fallbackPaths := []string{
		"C:\\Program Files\\Cursor\\Cursor.exe",
		"C:\\Program Files (x86)\\Cursor\\Cursor.exe",
	}

	for _, path := range fallbackPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}

// isVSCodeInstalled checks if VS Code is installed and available
func (r *FailedTestRenderer) isVSCodeInstalled() bool {
	// Check if code command is available in PATH first (most reliable)
	if _, err := exec.LookPath("code"); err == nil {
		return true
	}

	// Check common VS Code installation paths on Windows
	username := os.Getenv("USERNAME")
	if username != "" {
		commonPaths := []string{
			"C:\\Users\\" + username + "\\AppData\\Local\\Programs\\Microsoft VS Code\\Code.exe",
		}

		for _, path := range commonPaths {
			if _, err := os.Stat(path); err == nil {
				return true
			}
		}
	}

	// Fallback paths
	fallbackPaths := []string{
		"C:\\Program Files\\Microsoft VS Code\\Code.exe",
		"C:\\Program Files (x86)\\Microsoft VS Code\\Code.exe",
	}

	for _, path := range fallbackPaths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
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

			// Format line number and source code with improved styling
			var lineStr string
			if i == test.Error.HighlightedLine {
				// Highlight the error line (red background for the line number, normal text)
				lineStr = fmt.Sprintf("    %s| %s",
					r.formatter.BgRed(r.formatter.White(fmt.Sprintf("%3d", lineNum))),
					line,
				)
			} else {
				// Normal line (gray line number with | separator)
				lineStr = fmt.Sprintf("    %s| %s",
					r.formatter.Gray(fmt.Sprintf("%3d", lineNum)),
					line,
				)
			}

			// Write the line
			if _, err := fmt.Fprintln(r.writer, lineStr); err != nil {
				return err
			}

			// If this is the error line, add the enhanced error indicator
			if i == test.Error.HighlightedLine {
				// Create the error indicator line with improved positioning
				err := r.renderErrorPointer(test.Error.Location, line)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// renderErrorPointer renders the ^ pointer at the precise error location
func (r *FailedTestRenderer) renderErrorPointer(location *SourceLocation, sourceLine string) error {
	if location == nil {
		return nil
	}

	// Calculate the exact position of the error within the line
	errorPos := r.calculateErrorPosition(location, sourceLine)

	// Create the error indicator line with | on the left and ^ at the error position
	indicator := fmt.Sprintf("    %s| %s%s",
		r.formatter.Gray("   "), // Space for line number area
		strings.Repeat(" ", errorPos),
		r.formatter.Red("^"),
	)

	_, err := fmt.Fprintln(r.writer, indicator)
	return err
}

// calculateErrorPosition calculates the precise position of the error in the source line
func (r *FailedTestRenderer) calculateErrorPosition(location *SourceLocation, sourceLine string) int {
	if location.Column <= 0 {
		// If no column info, try to find a reasonable position based on error context
		return r.inferErrorPosition(sourceLine)
	}

	// Use the provided column, but ensure it's within bounds
	column := location.Column - 1 // Convert to 0-based indexing
	if column < 0 {
		column = 0
	}
	if column >= len(sourceLine) {
		column = len(sourceLine) - 1
		if column < 0 {
			column = 0
		}
	}

	return column
}

// inferErrorPosition tries to infer the best position to point to in a source line
func (r *FailedTestRenderer) inferErrorPosition(sourceLine string) int {
	// Look for common error patterns and position the pointer appropriately

	// Look for function calls that might be causing the error
	if pos := r.findPatternPosition(sourceLine, []string{"t.Error", "t.Errorf", "t.Fail", "t.Fatal"}); pos >= 0 {
		return pos
	}

	// Look for assertion operators
	if pos := r.findPatternPosition(sourceLine, []string{"!=", "==", "<=", ">=", "<", ">"}); pos >= 0 {
		return pos
	}

	// Look for array/slice access that might cause index errors
	if pos := r.findPatternPosition(sourceLine, []string{"["}); pos >= 0 {
		return pos
	}

	// Look for nil references
	if pos := r.findPatternPosition(sourceLine, []string{"nil"}); pos >= 0 {
		return pos
	}

	// Default: point to the first non-whitespace character
	for i, char := range sourceLine {
		if char != ' ' && char != '\t' {
			return i
		}
	}

	return 0
}

// findPatternPosition finds the position of the first matching pattern in the line
func (r *FailedTestRenderer) findPatternPosition(line string, patterns []string) int {
	for _, pattern := range patterns {
		if pos := strings.Index(line, pattern); pos >= 0 {
			return pos
		}
	}
	return -1
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
