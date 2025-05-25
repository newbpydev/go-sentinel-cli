package processor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/newbpydev/go-sentinel/pkg/models"
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

import "testing"

func TestExample(t *testing.T) {
    if 1 != 2 {
        t.Errorf("Expected 1 to equal 2")
    }
}
`
	err := os.WriteFile(sourceFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testError := &models.LegacyTestError{
		Message: "Expected 1 to equal 2",
		Location: &models.SourceLocation{
			File: sourceFile,
			Line: 7,
		},
	}

	// Act
	err = extractor.ExtractSourceContext(testError, 2)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(testError.SourceContext) == 0 {
		t.Fatal("Expected context to be non-empty")
	}
	if testError.HighlightedLine < 0 {
		t.Errorf("Expected valid highlight line, got %d", testError.HighlightedLine)
	}
}

// TestExtractSourceContext_NoLocation tests extracting source context when no location is available
func TestExtractSourceContext_NoLocation(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()

	testError := &models.LegacyTestError{
		Message: "Some error without location",
	}

	// Act
	err := extractor.ExtractSourceContext(testError, 2)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(testError.SourceContext) != 0 {
		t.Errorf("Expected empty context when no location, got %d lines", len(testError.SourceContext))
	}
}

// TestIsValidSourceFile_ValidFiles tests valid source file detection
func TestIsValidSourceFile_ValidFiles(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()

	tempDir := t.TempDir()
	validFiles := []string{
		"test.go",
		"main.go",
		"package_test.go",
	}

	for _, filename := range validFiles {
		filepath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filepath, []byte("package main"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}

		// Act
		isValid := extractor.IsValidSourceFile(filepath)

		// Assert
		if !isValid {
			t.Errorf("Expected %s to be valid source file", filename)
		}
	}
}

// TestIsValidSourceFile_InvalidFiles tests invalid source file detection
func TestIsValidSourceFile_InvalidFiles(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()

	tempDir := t.TempDir()
	invalidFiles := []string{
		"test.txt",
		"main.py",
		"package.js",
		"README.md",
	}

	for _, filename := range invalidFiles {
		filepath := filepath.Join(tempDir, filename)
		err := os.WriteFile(filepath, []byte("content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}

		// Act
		isValid := extractor.IsValidSourceFile(filepath)

		// Assert
		if isValid {
			t.Errorf("Expected %s to be invalid source file", filename)
		}
	}
}

// TestIsValidSourceFile_NonexistentFile tests nonexistent file detection
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

// TestIsValidSourceFile_Directory tests directory detection
func TestIsValidSourceFile_Directory(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()
	tempDir := t.TempDir()

	// Act
	isValid := extractor.IsValidSourceFile(tempDir)

	// Assert
	if isValid {
		t.Error("Expected directory to be invalid source file")
	}
}

// TestExtractContext_ZeroContextLines tests extracting context with zero context lines
func TestExtractContext_ZeroContextLines(t *testing.T) {
	// Arrange
	extractor := NewSourceExtractor()

	tempDir := t.TempDir()
	sourceFile := filepath.Join(tempDir, "test.go")
	content := `line1
line2
line3
line4`
	err := os.WriteFile(sourceFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Act
	context, err := extractor.ExtractContext(sourceFile, 2, 0)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(context) != 1 {
		t.Errorf("Expected 1 line (just the target line), got %d", len(context))
	}
	if len(context) > 0 && context[0] != "line2" {
		t.Errorf("Expected 'line2', got '%s'", context[0])
	}
}

// TestExtractSourceContext_HighlightedLineCalculation tests highlighted line calculation
func TestExtractSourceContext_HighlightedLineCalculation(t *testing.T) {
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

	testCases := []struct {
		name              string
		errorLine         int
		contextLines      int
		expectedHighlight int
	}{
		{
			name:              "Middle line",
			errorLine:         3,
			contextLines:      1,
			expectedHighlight: 1, // line3 should be at index 1 in context [line2, line3, line4]
		},
		{
			name:              "First line",
			errorLine:         1,
			contextLines:      2,
			expectedHighlight: 0, // line1 should be at index 0 in context [line1, line2, line3]
		},
		{
			name:              "Last line",
			errorLine:         5,
			contextLines:      2,
			expectedHighlight: 2, // line5 should be at index 2 in context [line3, line4, line5]
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testError := &models.LegacyTestError{
				Location: &models.SourceLocation{
					File: sourceFile,
					Line: tc.errorLine,
				},
			}

			// Act
			err := extractor.ExtractSourceContext(testError, tc.contextLines)

			// Assert
			if err != nil {
				t.Fatalf("Expected no error, got: %v", err)
			}
			if testError.HighlightedLine != tc.expectedHighlight {
				t.Errorf("Expected highlight line %d, got %d", tc.expectedHighlight, testError.HighlightedLine)
			}
			if len(testError.SourceContext) == 0 {
				t.Error("Expected non-empty context")
			}
		})
	}
}
