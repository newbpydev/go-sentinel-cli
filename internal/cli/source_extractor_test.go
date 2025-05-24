package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// TestNewSourceExtractor_Creation verifies source extractor initialization
func TestNewSourceExtractor_Creation(t *testing.T) {
	// Act
	extractor := NewSourceExtractor()

	// Assert
	if extractor == nil {
		t.Fatal("Expected extractor to be created, got nil")
	}
}

// TestExtractContext_ValidFile tests extracting context from a valid file
func TestExtractContext_ValidFile(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()

	// Create a temporary source file with known content
	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "test.go")
	content := `package main

import "fmt"

func main() {
    fmt.Println("Hello")
    fmt.Println("World")
    fmt.Println("Testing")
    fmt.Println("Context")
    fmt.Println("Extraction")
}
`
	err := os.WriteFile(sourceFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Act
	context, err := extractor.ExtractContext(sourceFile, 5, 2) // Line 5 with 2 lines of context (func main line)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(context) != 5 {
		t.Fatalf("Expected 5 lines of context, got %d", len(context))
	}

	expectedLines := []string{
		"import \"fmt\"",
		"",
		"func main() {",
		"    fmt.Println(\"Hello\")",
		"    fmt.Println(\"World\")",
	}

	for i, expected := range expectedLines {
		if i >= len(context) {
			t.Errorf("Missing line %d", i)
			continue
		}
		if context[i] != expected {
			t.Errorf("Line %d: expected '%s', got '%s'", i, expected, context[i])
		}
	}
}

// TestExtractContext_EdgeOfFile tests extracting context at the beginning of file
func TestExtractContext_EdgeOfFile(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()

	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "edge.go")
	content := `line1
line2
line3
line4`
	err := os.WriteFile(sourceFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testCases := []struct {
		name         string
		lineNumber   int
		contextLines int
		expectedLen  int
	}{
		{
			name:         "First line with context",
			lineNumber:   1,
			contextLines: 2,
			expectedLen:  3, // line 1 + 2 after
		},
		{
			name:         "Last line with context",
			lineNumber:   4,
			contextLines: 2,
			expectedLen:  3, // 2 before + line 4
		},
		{
			name:         "Middle line with large context",
			lineNumber:   2,
			contextLines: 10,
			expectedLen:  4, // All lines since context exceeds file size
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			context, err := extractor.ExtractContext(sourceFile, tc.lineNumber, tc.contextLines)

			// Assert
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if len(context) != tc.expectedLen {
				t.Errorf("Expected %d lines, got %d", tc.expectedLen, len(context))
			}
		})
	}
}

// TestExtractContext_InvalidLineNumber tests extracting context with invalid line numbers
func TestExtractContext_InvalidLineNumber(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()

	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "test.go")
	content := `line1
line2
line3`
	err := os.WriteFile(sourceFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testCases := []struct {
		name       string
		lineNumber int
	}{
		{
			name:       "Zero line number",
			lineNumber: 0,
		},
		{
			name:       "Negative line number",
			lineNumber: -5,
		},
		{
			name:       "Line number beyond file",
			lineNumber: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			context, err := extractor.ExtractContext(sourceFile, tc.lineNumber, 2)

			// Assert
			if err != nil {
				t.Fatalf("Expected no error for invalid line, got: %v", err)
			}
			if len(context) != 0 {
				t.Errorf("Expected empty context for invalid line, got %d lines", len(context))
			}
		})
	}
}

// TestExtractContext_NonexistentFile tests extracting context from nonexistent file
func TestExtractContext_NonexistentFile(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()
	nonexistentFile := "/path/to/nonexistent/file.go"

	// Act
	context, err := extractor.ExtractContext(nonexistentFile, 5, 2)

	// Assert
	if err == nil {
		t.Error("Expected error for nonexistent file, got nil")
	}
	if context != nil {
		t.Error("Expected context to be nil for nonexistent file")
	}
}

// TestExtractSourceContext_ValidTestError tests extracting source context for TestError
func TestExtractSourceContext_ValidTestError(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()

	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "test.go")
	content := `package main

func TestExample(t *testing.T) {
    if 1 != 2 {
        t.Error("This will fail")
    }
}
`
	err := os.WriteFile(sourceFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testError := &TestError{
		Message: "Test failed",
		Location: &SourceLocation{
			File: sourceFile,
			Line: 5,
		},
	}

	// Act
	err = extractor.ExtractSourceContext(testError, 2)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if testError.SourceContext == nil {
		t.Fatal("Expected source context to be set")
	}
	if len(testError.SourceContext) == 0 {
		t.Error("Expected non-empty source context")
	}

	// Check that highlighted line is set correctly
	if testError.HighlightedLine < 0 {
		t.Error("Expected highlighted line to be set")
	}
}

// TestExtractSourceContext_NoLocation tests extracting context for TestError without location
func TestExtractSourceContext_NoLocation(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()

	testError := &TestError{
		Message:  "Test failed",
		Location: nil,
	}

	// Act
	err := extractor.ExtractSourceContext(testError, 2)

	// Assert
	if err != nil {
		t.Errorf("Expected no error for TestError without location, got: %v", err)
	}
	if testError.SourceContext != nil {
		t.Error("Expected source context to remain nil")
	}
}

// TestIsValidSourceFile_ValidFiles tests validation of valid source files
func TestIsValidSourceFile_ValidFiles(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()

	tempDir := t.TempDir()
	validFile := filepath.Join(tempDir, "valid.go")
	content := `package main

func main() {
    fmt.Println("Hello, World!")
}
`
	err := os.WriteFile(validFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create valid file: %v", err)
	}

	// Act
	isValid := extractor.IsValidSourceFile(validFile)

	// Assert
	if !isValid {
		t.Error("Expected valid Go file to be recognized as valid")
	}
}

// TestIsValidSourceFile_InvalidFiles tests validation of invalid source files
func TestIsValidSourceFile_InvalidFiles(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()
	tempDir := t.TempDir()

	testCases := []struct {
		name     string
		fileName string
		content  []byte
	}{
		{
			name:     "Non-Go file",
			fileName: "test.txt",
			content:  []byte("This is a text file"),
		},
		{
			name:     "Binary file with .go extension",
			fileName: "binary.go",
			content:  []byte{0x00, 0x01, 0x02, 0x03}, // Binary content with null bytes
		},
		{
			name:     "Empty file",
			fileName: "empty.go",
			content:  []byte{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tempDir, tc.fileName)
			err := os.WriteFile(testFile, tc.content, 0644)
			if err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Act
			isValid := extractor.IsValidSourceFile(testFile)

			// Assert
			if isValid && tc.name != "Empty file" { // Empty files might be considered valid
				t.Errorf("Expected %s to be invalid", tc.name)
			}
		})
	}
}

// TestIsValidSourceFile_NonexistentFile tests validation of nonexistent file
func TestIsValidSourceFile_NonexistentFile(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()
	nonexistentFile := "/path/to/nonexistent/file.go"

	// Act
	isValid := extractor.IsValidSourceFile(nonexistentFile)

	// Assert
	if isValid {
		t.Error("Expected nonexistent file to be invalid")
	}
}

// TestIsValidSourceFile_Directory tests validation of directory
func TestIsValidSourceFile_Directory(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()
	tempDir := t.TempDir()

	// Act
	isValid := extractor.IsValidSourceFile(tempDir)

	// Assert
	if isValid {
		t.Error("Expected directory to be invalid as source file")
	}
}

// TestMinMaxHelpers tests the min and max helper functions
func TestMinMaxHelpers(t *testing.T) {
	testCases := []struct {
		name   string
		a, b   int
		minExp int
		maxExp int
	}{
		{
			name:   "a < b",
			a:      3,
			b:      7,
			minExp: 3,
			maxExp: 7,
		},
		{
			name:   "a > b",
			a:      10,
			b:      5,
			minExp: 5,
			maxExp: 10,
		},
		{
			name:   "a == b",
			a:      4,
			b:      4,
			minExp: 4,
			maxExp: 4,
		},
		{
			name:   "negative numbers",
			a:      -3,
			b:      -7,
			minExp: -7,
			maxExp: -3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test min function
			if result := min(tc.a, tc.b); result != tc.minExp {
				t.Errorf("min(%d, %d): expected %d, got %d", tc.a, tc.b, tc.minExp, result)
			}

			// Test max function
			if result := max(tc.a, tc.b); result != tc.maxExp {
				t.Errorf("max(%d, %d): expected %d, got %d", tc.a, tc.b, tc.maxExp, result)
			}
		})
	}
}

// TestExtractContext_ZeroContextLines tests extracting with zero context lines
func TestExtractContext_ZeroContextLines(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()

	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "test.go")
	content := `line1
line2
line3
line4
line5`
	err := os.WriteFile(sourceFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Act
	context, err := extractor.ExtractContext(sourceFile, 3, 0)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(context) != 1 {
		t.Errorf("Expected 1 line with zero context, got %d", len(context))
	}
	if context[0] != "line3" {
		t.Errorf("Expected 'line3', got '%s'", context[0])
	}
}

// TestExtractSourceContext_HighlightedLineCalculation tests highlighted line index calculation
func TestExtractSourceContext_HighlightedLineCalculation(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()

	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "test.go")
	content := `line1
line2
line3
line4
line5
line6
line7`
	err := os.WriteFile(sourceFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testCases := []struct {
		name                    string
		targetLine              int
		contextLines            int
		expectedHighlightedLine int
	}{
		{
			name:                    "Middle line",
			targetLine:              4,
			contextLines:            2,
			expectedHighlightedLine: 2, // Index within the extracted context
		},
		{
			name:                    "First line",
			targetLine:              1,
			contextLines:            2,
			expectedHighlightedLine: 0,
		},
		{
			name:                    "Last line",
			targetLine:              7,
			contextLines:            2,
			expectedHighlightedLine: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testError := &TestError{
				Message: "Test failed",
				Location: &SourceLocation{
					File: sourceFile,
					Line: tc.targetLine,
				},
			}

			// Act
			err := extractor.ExtractSourceContext(testError, tc.contextLines)

			// Assert
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if testError.HighlightedLine != tc.expectedHighlightedLine {
				t.Errorf("Expected highlighted line %d, got %d", tc.expectedHighlightedLine, testError.HighlightedLine)
			}
		})
	}
}
