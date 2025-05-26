// Package display provides code context formatting for source code display
package display

import (
	"strings"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// CodeContextFormatter handles formatting source code context with line numbers and pointers
type CodeContextFormatter struct {
	formatter      *colors.ColorFormatter
	spacingManager *SpacingManager
	config         *Config

	// Display options
	maxContextLines  int
	lineNumberWidth  int
	showLineNumbers  bool
	showErrorPointer bool
	indentLevel      int
}

// CodeContextFormatOptions configures code context formatting
type CodeContextFormatOptions struct {
	MaxContextLines  int
	LineNumberWidth  int
	ShowLineNumbers  bool
	ShowErrorPointer bool
	IndentLevel      int
}

// NewCodeContextFormatter creates a new code context formatter
func NewCodeContextFormatter(config *Config, options *CodeContextFormatOptions) *CodeContextFormatter {
	formatter := colors.NewAutoColorFormatter()

	spacingManager := NewSpacingManager(&SpacingConfig{
		BaseIndent:    0,
		TestIndent:    2,
		SubtestIndent: 4,
		ErrorIndent:   4,
	})

	// Set defaults if options not provided or apply defaults to unset fields
	if options == nil {
		options = &CodeContextFormatOptions{
			MaxContextLines:  5,
			LineNumberWidth:  4,
			ShowLineNumbers:  true,
			ShowErrorPointer: true,
			IndentLevel:      0,
		}
	} else {
		// Apply defaults to zero-value fields
		if options.MaxContextLines == 0 {
			options.MaxContextLines = 5
		}
		if options.LineNumberWidth == 0 {
			options.LineNumberWidth = 4
		}
	}

	return &CodeContextFormatter{
		formatter:        formatter,
		spacingManager:   spacingManager,
		config:           config,
		maxContextLines:  options.MaxContextLines,
		lineNumberWidth:  options.LineNumberWidth,
		showLineNumbers:  options.ShowLineNumbers,
		showErrorPointer: options.ShowErrorPointer,
		indentLevel:      options.IndentLevel,
	}
}

// FormatCodeContext formats source code context with line numbers and error pointer
func (f *CodeContextFormatter) FormatCodeContext(err *models.TestError) string {
	if err == nil || len(err.SourceContext) == 0 {
		return ""
	}

	var result strings.Builder

	// Calculate starting line number
	startLine := f.calculateStartLine(err)

	// Determine which line has the error (for ^ pointer)
	errorLineIndex := f.calculateErrorLineIndex(err, startLine)

	// Limit context lines based on configuration
	contextLines := f.limitContextLines(err.SourceContext)

	// Calculate actual line number width based on content
	actualLineNumberWidth := f.calculateLineNumberWidth(startLine, len(contextLines))

	// Render each line with right-aligned line numbers
	for i, line := range contextLines {
		currentLineNumber := startLine + i

		// Format the code line
		codeLine := f.formatCodeLine(line, currentLineNumber, actualLineNumberWidth)
		result.WriteString(codeLine)
		result.WriteString("\n")

		// Add ^ pointer on the next line if this is the error line
		// Only show pointer if the error line is within the displayed context
		if f.showErrorPointer && i == errorLineIndex && err.SourceColumn > 0 && errorLineIndex >= 0 && errorLineIndex < len(contextLines) {
			pointerLine := f.formatErrorPointer(err.SourceColumn, actualLineNumberWidth)
			result.WriteString(pointerLine)
			result.WriteString("\n")
		}
	}

	return result.String()
}

// FormatSingleLine formats a single line of code with line number
func (f *CodeContextFormatter) FormatSingleLine(line string, lineNumber int) string {
	return f.formatCodeLine(line, lineNumber, f.lineNumberWidth)
}

// calculateStartLine determines the starting line number for the context
func (f *CodeContextFormatter) calculateStartLine(err *models.TestError) int {
	startLine := err.ContextStartLine
	if startLine <= 0 {
		// Center the error line in the context if possible
		startLine = err.SourceLine - (f.maxContextLines / 2)
		if startLine < 1 {
			startLine = 1
		}
	}
	return startLine
}

// calculateErrorLineIndex determines which line in the context contains the error
func (f *CodeContextFormatter) calculateErrorLineIndex(err *models.TestError, startLine int) int {
	if err.SourceLine <= 0 {
		return -1
	}
	return err.SourceLine - startLine
}

// limitContextLines limits the context lines based on configuration
func (f *CodeContextFormatter) limitContextLines(contextLines []string) []string {
	if len(contextLines) <= f.maxContextLines {
		return contextLines
	}
	return contextLines[:f.maxContextLines]
}

// calculateLineNumberWidth calculates the width needed for line numbers
func (f *CodeContextFormatter) calculateLineNumberWidth(startLine, lineCount int) int {
	maxLineNumber := startLine + lineCount - 1
	width := len(f.intToString(maxLineNumber))

	// Ensure minimum width
	if width < f.lineNumberWidth {
		width = f.lineNumberWidth
	}

	return width
}

// formatCodeLine formats a single line of code with line number
func (f *CodeContextFormatter) formatCodeLine(line string, lineNumber int, lineNumberWidth int) string {
	var result strings.Builder

	// Add base indentation
	baseIndent := f.spacingManager.GetIndentString(f.indentLevel * 2)
	result.WriteString(baseIndent)

	// Add fixed indentation for code context (5 spaces as per spec)
	result.WriteString("     ")

	if f.showLineNumbers {
		// Format and right-align line number
		lineNumStr := f.intToString(lineNumber)
		alignedLineNum := f.spacingManager.AlignRight(lineNumStr, lineNumberWidth)

		result.WriteString(f.formatter.Dim(alignedLineNum))
		result.WriteString("|")
	}

	// Add the code content with proper spacing
	if strings.TrimSpace(line) != "" {
		if f.showLineNumbers {
			result.WriteString("  ") // 2 spaces after | separator
		}
		result.WriteString(line)
	}

	return result.String()
}

// formatErrorPointer formats the ^ pointer line to indicate the error location
func (f *CodeContextFormatter) formatErrorPointer(columnNumber int, lineNumberWidth int) string {
	var result strings.Builder

	// Add base indentation
	baseIndent := f.spacingManager.GetIndentString(f.indentLevel * 2)
	result.WriteString(baseIndent)

	// Add fixed indentation for code context (5 spaces as per spec)
	result.WriteString("     ")

	if f.showLineNumbers {
		// Add spacing to align with line numbers
		result.WriteString(strings.Repeat(" ", lineNumberWidth))
		result.WriteString("|")
		result.WriteString("  ") // 2 spaces after | separator
	}

	// Calculate pointer position (accounting for any leading whitespace)
	pointerPos := columnNumber - 1
	if pointerPos > 0 {
		result.WriteString(strings.Repeat(" ", pointerPos))
	}
	result.WriteString(f.formatter.Red("^"))

	return result.String()
}

// FormatMultipleErrors formats multiple error contexts
func (f *CodeContextFormatter) FormatMultipleErrors(errors []*models.TestError) string {
	if len(errors) == 0 {
		return ""
	}

	var result strings.Builder

	for i, err := range errors {
		if err == nil {
			continue
		}

		context := f.FormatCodeContext(err)
		if context != "" {
			result.WriteString(context)

			// Add spacing between contexts (except for the last one)
			if i < len(errors)-1 {
				result.WriteString("\n")
			}
		}
	}

	return result.String()
}

// intToString converts an integer to string without importing fmt
func (f *CodeContextFormatter) intToString(n int) string {
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

// SetMaxContextLines sets the maximum number of context lines to show
func (f *CodeContextFormatter) SetMaxContextLines(lines int) {
	if lines < 1 {
		lines = 1
	}
	f.maxContextLines = lines
}

// SetLineNumberWidth sets the minimum width for line numbers
func (f *CodeContextFormatter) SetLineNumberWidth(width int) {
	if width < 1 {
		width = 1
	}
	f.lineNumberWidth = width
}

// SetShowLineNumbers enables/disables line number display
func (f *CodeContextFormatter) SetShowLineNumbers(show bool) {
	f.showLineNumbers = show
}

// SetShowErrorPointer enables/disables error pointer display
func (f *CodeContextFormatter) SetShowErrorPointer(show bool) {
	f.showErrorPointer = show
}

// SetIndentLevel sets the base indentation level
func (f *CodeContextFormatter) SetIndentLevel(level int) {
	f.indentLevel = level
}

// GetMaxContextLines returns the maximum number of context lines
func (f *CodeContextFormatter) GetMaxContextLines() int {
	return f.maxContextLines
}

// GetLineNumberWidth returns the minimum line number width
func (f *CodeContextFormatter) GetLineNumberWidth() int {
	return f.lineNumberWidth
}

// IsShowLineNumbers returns whether line numbers are displayed
func (f *CodeContextFormatter) IsShowLineNumbers() bool {
	return f.showLineNumbers
}

// IsShowErrorPointer returns whether error pointers are displayed
func (f *CodeContextFormatter) IsShowErrorPointer() bool {
	return f.showErrorPointer
}

// GetIndentLevel returns the current indentation level
func (f *CodeContextFormatter) GetIndentLevel() int {
	return f.indentLevel
}
