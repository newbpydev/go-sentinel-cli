package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// SourceExtractor extracts source code context around error locations
type SourceExtractor struct{}

// NewSourceExtractor creates a new SourceExtractor
func NewSourceExtractor() *SourceExtractor {
	return &SourceExtractor{}
}

// ExtractContext extracts lines of context around the specified line in a file
func (e *SourceExtractor) ExtractContext(filePath string, lineNumber int, contextLines int) ([]string, error) {
	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	// Read all lines from the file
	var allLines []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	// Check if line number is valid
	if lineNumber < 1 || lineNumber > len(allLines) {
		return []string{}, nil // Return empty context for invalid line numbers
	}

	// Calculate the range of lines to extract (convert to 0-based indexing)
	lineIndex := lineNumber - 1
	startLine := max(0, lineIndex-contextLines)
	endLine := min(len(allLines), lineIndex+contextLines+1)

	// Extract the context lines
	context := make([]string, 0, endLine-startLine)
	for i := startLine; i < endLine; i++ {
		context = append(context, allLines[i])
	}

	return context, nil
}

// ExtractSourceContext extracts source context for a TestError
func (e *SourceExtractor) ExtractSourceContext(testError *TestError, contextLines int) error {
	if testError.Location == nil {
		return nil // No location to extract context from
	}

	context, err := e.ExtractContext(testError.Location.File, testError.Location.Line, contextLines)
	if err != nil {
		return fmt.Errorf("failed to extract source context: %w", err)
	}

	// Set the source context and highlighted line
	testError.SourceContext = context

	// Calculate the highlighted line index (0-based) within the context
	if len(context) > 0 {
		// The highlighted line should be the line containing the error
		// Calculate its position within the extracted context
		lineIndex := testError.Location.Line - 1
		startLine := max(0, lineIndex-contextLines)
		highlightedIndex := lineIndex - startLine

		if highlightedIndex >= 0 && highlightedIndex < len(context) {
			testError.HighlightedLine = highlightedIndex
		}
	}

	return nil
}

// IsValidSourceFile checks if a file appears to be a valid source file
func (e *SourceExtractor) IsValidSourceFile(filePath string) bool {
	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	// Check if it's a regular file
	if !info.Mode().IsRegular() {
		return false
	}

	// Check file extension
	if !strings.HasSuffix(filePath, ".go") {
		return false
	}

	// Try to read a small portion to check if it's text
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first few bytes to check for binary content
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && n == 0 {
		return false
	}

	// Check for null bytes (indicating binary content)
	for i := 0; i < n; i++ {
		if buffer[i] == 0 {
			return false
		}
	}

	return true
}

// Helper functions for min/max (since Go doesn't have built-in min/max for int)
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
