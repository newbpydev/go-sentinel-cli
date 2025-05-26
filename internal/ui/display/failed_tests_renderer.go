// Package display provides failed tests section rendering for detailed failure analysis
package display

import (
	"strings"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/internal/ui/icons"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// FailedTestsRenderer handles rendering the failed tests section with detailed analysis
type FailedTestsRenderer struct {
	formatter      *colors.ColorFormatter
	icons          icons.IconProvider
	spacingManager *SpacingManager
	config         *Config

	// Display options
	terminalWidth   int
	showCodeContext bool
	maxContextLines int
	indentLevel     int
}

// FailedTestsRenderOptions configures failed tests rendering
type FailedTestsRenderOptions struct {
	TerminalWidth   int
	ShowCodeContext bool
	MaxContextLines int
	IndentLevel     int
}

// NewFailedTestsRenderer creates a new failed tests renderer
func NewFailedTestsRenderer(config *Config, options *FailedTestsRenderOptions) *FailedTestsRenderer {
	formatter := colors.NewAutoColorFormatter()

	// Detect terminal capabilities for icon selection
	detector := colors.NewTerminalDetector()
	var iconProvider icons.IconProvider
	if detector.SupportsUnicode() {
		iconProvider = icons.NewUnicodeProvider()
	} else {
		iconProvider = icons.NewASCIIProvider()
	}

	spacingManager := NewSpacingManager(&SpacingConfig{
		BaseIndent:    0, // No base indent for failed tests section
		TestIndent:    2,
		SubtestIndent: 4,
		ErrorIndent:   4,
	})

	// Set defaults if options not provided or apply defaults to unset fields
	if options == nil {
		options = &FailedTestsRenderOptions{
			TerminalWidth:   110,
			ShowCodeContext: true,
			MaxContextLines: 5,
			IndentLevel:     0,
		}
	} else {
		// Apply defaults to zero-value fields
		if options.TerminalWidth == 0 {
			options.TerminalWidth = 110
		}
		if options.MaxContextLines == 0 {
			options.MaxContextLines = 5
		}
	}

	return &FailedTestsRenderer{
		formatter:       formatter,
		icons:           iconProvider,
		spacingManager:  spacingManager,
		config:          config,
		terminalWidth:   options.TerminalWidth,
		showCodeContext: options.ShowCodeContext,
		maxContextLines: options.MaxContextLines,
		indentLevel:     options.IndentLevel,
	}
}

// RenderFailedTestsSection renders the complete failed tests section
// Format: 110+ ─ separator, centered "Failed Tests X" header, detailed test failures
func (r *FailedTestsRenderer) RenderFailedTestsSection(failedTests []*models.TestResult, failedCount int) string {
	if len(failedTests) == 0 || failedCount == 0 {
		return ""
	}

	var result strings.Builder

	// Render section separator (110+ ─ characters)
	separator := r.renderSectionSeparator()
	result.WriteString(separator)
	result.WriteString("\n")

	// Render centered section header
	header := r.renderSectionHeader("Failed Tests", failedCount)
	result.WriteString(header)
	result.WriteString("\n")

	// Render another separator
	result.WriteString(separator)
	result.WriteString("\n")

	// Render each failed test with details
	for i, test := range failedTests {
		if test == nil || test.Status != models.TestStatusFailed {
			continue
		}

		testDetail := r.renderFailedTestDetail(test)
		result.WriteString(testDetail)

		// Add spacing between tests (except for the last one)
		if i < len(failedTests)-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// renderSectionSeparator renders a line of ─ characters spanning the terminal width
func (r *FailedTestsRenderer) renderSectionSeparator() string {
	// Ensure minimum 110 characters as specified
	width := r.terminalWidth
	if width < 110 {
		width = 110
	}

	return strings.Repeat("─", width)
}

// renderSectionHeader renders a centered section header
func (r *FailedTestsRenderer) renderSectionHeader(title string, count int) string {
	// Format: "Failed Tests 26"
	headerText := title
	if count >= 0 {
		headerText += " " + r.formatCount(count)
	}

	// Center the header within the terminal width
	return r.spacingManager.AlignCenter(headerText, r.terminalWidth)
}

// renderFailedTestDetail renders detailed information for a single failed test
func (r *FailedTestsRenderer) renderFailedTestDetail(test *models.TestResult) string {
	if test == nil || test.Error == nil {
		return ""
	}

	var result strings.Builder

	// Render test failure header: " FAIL  filename > TestName"
	failureHeader := r.renderFailureHeader(test)
	result.WriteString(failureHeader)
	result.WriteString("\n")

	// Render error message with type
	errorMessage := r.renderErrorMessage(test.Error)
	result.WriteString(errorMessage)
	result.WriteString("\n")

	// Render source location with chevron (↳ filename:line:column)
	if test.Error.SourceFile != "" && test.Error.SourceLine > 0 {
		locationLine := r.renderSourceLocation(test.Error)
		result.WriteString(locationLine)
		result.WriteString("\n")

		// Render code context if available and enabled
		if r.showCodeContext && len(test.Error.SourceContext) > 0 {
			codeContext := r.renderCodeContext(test.Error)
			if codeContext != "" {
				result.WriteString(codeContext)
				result.WriteString("\n")
			}
		}
	}

	return result.String()
}

// renderFailureHeader renders the failure header line
// Format: " FAIL  filename > TestName"
func (r *FailedTestsRenderer) renderFailureHeader(test *models.TestResult) string {
	var result strings.Builder

	// Get FAIL icon and include both icon and text for compatibility
	failIcon, _ := r.icons.GetIcon("test_failed")
	coloredFailIcon := r.formatter.Red(failIcon)

	result.WriteString(" ")
	result.WriteString(coloredFailIcon)
	result.WriteString(" FAIL  ")

	// Extract filename from package path (if available)
	filename := test.Package
	if test.Error != nil && test.Error.SourceFile != "" {
		// Use the source file from the error
		filename = test.Error.SourceFile
	}
	if filename == "" {
		filename = "(unknown)"
	}

	result.WriteString(filename)
	result.WriteString(" > ")
	result.WriteString(test.Name)

	return result.String()
}

// renderErrorMessage renders the error type and message
func (r *FailedTestsRenderer) renderErrorMessage(err *models.TestError) string {
	if err == nil {
		return ""
	}

	errorType := err.Type
	if errorType == "" {
		errorType = "Error"
	}

	message := err.Message
	if message == "" {
		message = "Test failed"
	}

	// Format: "AssertionError: Expected 1+1 to equal 3, but got 2"
	return errorType + ": " + message
}

// renderSourceLocation renders the source location with chevron
// Format: "↳ filename:line:column"
func (r *FailedTestsRenderer) renderSourceLocation(err *models.TestError) string {
	if err == nil || err.SourceFile == "" {
		return ""
	}

	var result strings.Builder

	// Get chevron icon
	chevronIcon, _ := r.icons.GetIcon("chevron_right")
	result.WriteString(chevronIcon)
	result.WriteString(" ")

	// Build location string
	result.WriteString(err.SourceFile)
	result.WriteString(":")
	result.WriteString(r.formatLineNumber(err.SourceLine))

	if err.SourceColumn > 0 {
		result.WriteString(":")
		result.WriteString(r.formatLineNumber(err.SourceColumn))
	}

	return result.String()
}

// renderCodeContext renders the 5-line code snippet with right-aligned line numbers and ^ pointer
func (r *FailedTestsRenderer) renderCodeContext(err *models.TestError) string {
	if err == nil || len(err.SourceContext) == 0 {
		return ""
	}

	var result strings.Builder

	// Calculate starting line number
	startLine := err.ContextStartLine
	if startLine <= 0 {
		startLine = err.SourceLine - 2 // Assume error line is in the middle
		if startLine < 1 {
			startLine = 1
		}
	}

	// Determine which line has the error (for ^ pointer)
	errorLineIndex := -1
	if err.SourceLine > 0 {
		errorLineIndex = err.SourceLine - startLine
	}

	// Limit context lines based on configuration
	maxLines := r.maxContextLines
	contextLines := err.SourceContext
	if len(contextLines) > maxLines {
		contextLines = contextLines[:maxLines]
	}

	// Calculate line number width for right alignment
	maxLineNumber := startLine + len(contextLines) - 1
	lineNumberWidth := len(r.formatLineNumber(maxLineNumber))

	// Render each line with right-aligned line numbers
	for i, line := range contextLines {
		currentLineNumber := startLine + i
		lineNumStr := r.formatLineNumber(currentLineNumber)

		// Right-align line number
		alignedLineNum := r.spacingManager.AlignRight(lineNumStr, lineNumberWidth)

		result.WriteString("     ") // 5 spaces for indentation
		result.WriteString(r.formatter.Dim(alignedLineNum))
		result.WriteString("|")

		// Add extra spacing for the code
		if strings.TrimSpace(line) != "" {
			result.WriteString("  ")
			result.WriteString(line)
		}

		result.WriteString("\n")

		// Add ^ pointer on the next line if this is the error line
		if i == errorLineIndex && err.SourceColumn > 0 {
			result.WriteString("     ") // 5 spaces for indentation
			result.WriteString(strings.Repeat(" ", lineNumberWidth))
			result.WriteString("|")
			result.WriteString("  ")

			// Calculate pointer position (accounting for any leading whitespace)
			pointerPos := err.SourceColumn - 1
			if pointerPos > 0 {
				result.WriteString(strings.Repeat(" ", pointerPos))
			}
			result.WriteString(r.formatter.Red("^"))
			result.WriteString("\n")
		}
	}

	return result.String()
}

// formatCount formats a count for display
func (r *FailedTestsRenderer) formatCount(count int) string {
	return r.formatter.Red(r.formatLineNumber(count))
}

// formatLineNumber formats a line number as a string
func (r *FailedTestsRenderer) formatLineNumber(line int) string {
	if line <= 0 {
		return "0"
	}

	return getIntString(line)
}

// Helper function to convert int to string (avoiding fmt import for simplicity)
func getIntString(n int) string {
	if n == 0 {
		return "0"
	}

	var result []byte
	negative := n < 0
	if negative {
		n = -n
	}

	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}

	if negative {
		result = append([]byte{'-'}, result...)
	}

	return string(result)
}

// SetTerminalWidth sets the terminal width for formatting
func (r *FailedTestsRenderer) SetTerminalWidth(width int) {
	if width < 110 {
		width = 110
	}
	r.terminalWidth = width
}

// SetShowCodeContext enables/disables code context display
func (r *FailedTestsRenderer) SetShowCodeContext(show bool) {
	r.showCodeContext = show
}

// SetMaxContextLines sets the maximum number of context lines to show
func (r *FailedTestsRenderer) SetMaxContextLines(lines int) {
	if lines < 1 {
		lines = 1
	}
	r.maxContextLines = lines
}

// GetTerminalWidth returns the current terminal width
func (r *FailedTestsRenderer) GetTerminalWidth() int {
	return r.terminalWidth
}

// IsShowCodeContext returns whether code context display is enabled
func (r *FailedTestsRenderer) IsShowCodeContext() bool {
	return r.showCodeContext
}

// GetMaxContextLines returns the maximum number of context lines
func (r *FailedTestsRenderer) GetMaxContextLines() int {
	return r.maxContextLines
}
