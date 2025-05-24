package display

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/pkg/models"
	"golang.org/x/term"
)

// ErrorFormatterInterface defines the contract for error formatting functionality
type ErrorFormatterInterface interface {
	// FormatClickableLocation formats a clickable file location reference
	FormatClickableLocation(location *models.SourceLocation) string

	// RenderSourceContext renders source code around the error with line numbers
	RenderSourceContext(test *models.TestResult) error

	// RenderErrorPointer renders the ^ pointer at the precise error location
	RenderErrorPointer(location *models.SourceLocation, sourceLine string) error

	// CalculateErrorPosition calculates the precise position of the error in the source line
	CalculateErrorPosition(location *models.SourceLocation, sourceLine string) int

	// GetTerminalWidth returns the current terminal width
	GetTerminalWidth() int
}

// ErrorFormatter handles detailed error message formatting and source context
type ErrorFormatter struct {
	writer    io.Writer
	formatter colors.FormatterInterface
	width     int
}

// NewErrorFormatter creates a new ErrorFormatter
func NewErrorFormatter(writer io.Writer, formatter colors.FormatterInterface, width int) *ErrorFormatter {
	if writer == nil {
		panic("writer cannot be nil")
	}
	if formatter == nil {
		panic("formatter cannot be nil")
	}

	return &ErrorFormatter{
		writer:    writer,
		formatter: formatter,
		width:     width,
	}
}

// NewErrorFormatterWithDefaults creates an ErrorFormatter with auto-detected defaults
func NewErrorFormatterWithDefaults(writer io.Writer, formatter colors.FormatterInterface) *ErrorFormatter {
	return NewErrorFormatter(writer, formatter, 80)
}

// GetTerminalWidth returns the current terminal width or default
func (e *ErrorFormatter) GetTerminalWidth() int {
	if fd := int(os.Stdout.Fd()); term.IsTerminal(fd) {
		if width, _, err := term.GetSize(fd); err == nil && width > 0 {
			return width
		}
	}
	// Fallback to configured width or default
	if e.width > 0 {
		return e.width
	}
	return 80
}

// FormatClickableLocation formats a clickable file location reference
func (e *ErrorFormatter) FormatClickableLocation(location *models.SourceLocation) string {
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
	var linkUrl string

	// Strategy 1: Try Cursor-specific URL scheme (most reliable when it works)
	if e.isCursorInstalled() {
		linkUrl = fmt.Sprintf("cursor://file/%s:%d:%d", urlPath, location.Line, location.Column)
	} else if e.isVSCodeInstalled() {
		// Strategy 2: Use VS Code URL scheme if available
		linkUrl = fmt.Sprintf("vscode://file/%s:%d:%d", urlPath, location.Line, location.Column)
	} else {
		// Strategy 3: Use a pragmatic fallback approach
		linkUrl = fmt.Sprintf("file:///%s#line=%d&column=%d", urlPath, location.Line, location.Column)
	}

	// Use OSC 8 hyperlink escape sequence for terminal link support
	// Format: ESC]8;;URL ESC\\ text ESC]8;; ESC\\
	clickableText := fmt.Sprintf("\033]8;;%s\033\\%s\033]8;;\033\\",
		linkUrl,
		e.formatter.Cyan(displayText))

	return fmt.Sprintf("↳ %s", clickableText)
}

// isCursorInstalled checks if Cursor is installed and available
func (e *ErrorFormatter) isCursorInstalled() bool {
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
func (e *ErrorFormatter) isVSCodeInstalled() bool {
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

// RenderSourceContext renders source code around the error with line numbers and highlighting
func (e *ErrorFormatter) RenderSourceContext(test *models.TestResult) error {
	if test.Error == nil || test.Error.SourceFile == "" {
		return nil
	}

	// Print source context if available
	if len(test.Error.SourceContext) > 0 {
		// Calculate the starting line number
		startLine := test.Error.ContextStartLine
		if startLine <= 0 {
			startLine = test.Error.SourceLine - len(test.Error.SourceContext)/2
		}

		// Format each line of source code
		for i, line := range test.Error.SourceContext {
			// Calculate the actual line number
			lineNum := startLine + i

			// Format line number and source code with improved styling
			var lineStr string
			if lineNum == test.Error.SourceLine {
				// Highlight the error line (red background for the line number, normal text)
				lineStr = fmt.Sprintf("    %s| %s",
					e.formatter.BgRed(e.formatter.White(fmt.Sprintf("%3d", lineNum))),
					line,
				)
			} else {
				// Normal line (gray line number with | separator)
				lineStr = fmt.Sprintf("    %s| %s",
					e.formatter.Gray(fmt.Sprintf("%3d", lineNum)),
					line,
				)
			}

			// Write the line
			if _, err := fmt.Fprintln(e.writer, lineStr); err != nil {
				return fmt.Errorf("failed to write source line: %w", err)
			}

			// If this is the error line, add the enhanced error indicator
			if lineNum == test.Error.SourceLine {
				// Create source location for the pointer
				location := &models.SourceLocation{
					File:   test.Error.SourceFile,
					Line:   test.Error.SourceLine,
					Column: test.Error.SourceColumn,
				}

				// Create the error indicator line with improved positioning
				err := e.RenderErrorPointer(location, line)
				if err != nil {
					return fmt.Errorf("failed to render error pointer: %w", err)
				}
			}
		}
	}

	return nil
}

// RenderErrorPointer renders the ^ pointer at the precise error location
func (e *ErrorFormatter) RenderErrorPointer(location *models.SourceLocation, sourceLine string) error {
	if location == nil {
		return nil
	}

	// Calculate the exact position of the error within the line
	errorPos := e.CalculateErrorPosition(location, sourceLine)

	// Create the error indicator line with | on the left and ^ at the error position
	indicator := fmt.Sprintf("    %s| %s%s",
		e.formatter.Gray("   "), // Space for line number area
		strings.Repeat(" ", errorPos),
		e.formatter.Red("^"),
	)

	_, err := fmt.Fprintln(e.writer, indicator)
	if err != nil {
		return fmt.Errorf("failed to write error pointer: %w", err)
	}
	return nil
}

// CalculateErrorPosition calculates the precise position of the error in the source line
func (e *ErrorFormatter) CalculateErrorPosition(location *models.SourceLocation, sourceLine string) int {
	if location.Column <= 0 {
		// If no column info, try to find a reasonable position based on error context
		return e.inferErrorPosition(sourceLine)
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
func (e *ErrorFormatter) inferErrorPosition(sourceLine string) int {
	// Look for common error patterns and position the pointer appropriately

	// Look for function calls that might be causing the error
	if pos := e.findPatternPosition(sourceLine, []string{"t.Error", "t.Errorf", "t.Fail", "t.Fatal"}); pos >= 0 {
		return pos
	}

	// Look for assertion operators
	if pos := e.findPatternPosition(sourceLine, []string{"!=", "==", "<=", ">=", "<", ">"}); pos >= 0 {
		return pos
	}

	// Look for array/slice access that might cause index errors
	if pos := e.findPatternPosition(sourceLine, []string{"["}); pos >= 0 {
		return pos
	}

	// Look for nil references
	if pos := e.findPatternPosition(sourceLine, []string{"nil"}); pos >= 0 {
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

// findPatternPosition finds the position of the first matching pattern in a line
func (e *ErrorFormatter) findPatternPosition(line string, patterns []string) int {
	earliestPos := -1
	for _, pattern := range patterns {
		if pos := strings.Index(line, pattern); pos >= 0 {
			if earliestPos == -1 || pos < earliestPos {
				earliestPos = pos
			}
		}
	}
	return earliestPos
}
