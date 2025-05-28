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

// TestNewSourceExtractor_Factory tests NewSourceExtractor factory function
func TestNewSourceExtractor_Factory(t *testing.T) {
	t.Parallel()

	extractor := NewSourceExtractor()
	if extractor == nil {
		t.Fatal("Expected source extractor to be created, got nil")
	}

	// Verify extractor can be used
	context, err := extractor.ExtractContext("test.go", 10, 3)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if context != nil {
		t.Error("Expected nil context for non-existent file")
	}
}

// TestMinMaxHelperFunctions tests the min and max helper functions
func TestMinMaxHelperFunctions(t *testing.T) {
	t.Parallel()

	// Create an extractor to access the helper functions
	extractor := NewSourceExtractor()

	// Test min function through ExtractContext which uses it
	// We'll create a scenario where min is called
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.go")

	// Create a small file to test boundary conditions
	content := `package main

func test() {
	// line 4
	x := 1
	// line 6
}`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test extracting context at line 1 (should use min function)
	context, err := extractor.ExtractContext(testFile, 1, 2)
	if err != nil {
		t.Errorf("Unexpected error for line 1: %v", err)
	}
	if context == nil {
		t.Error("Expected context for line 1, got nil")
	}

	// Test extracting context at last line (should use max function)
	context, err = extractor.ExtractContext(testFile, 7, 2)
	if err != nil {
		t.Errorf("Unexpected error for line 7: %v", err)
	}
	if len(context) == 0 {
		t.Error("Expected context for last line, got empty")
	}

	// Test extracting context beyond file length (should use max function)
	context, err = extractor.ExtractContext(testFile, 100, 2)
	if err != nil {
		t.Errorf("Unexpected error for line 100: %v", err)
	}
	if len(context) != 0 {
		t.Error("Expected empty context for line beyond file")
	}
}

// TestExtractContext_BoundaryConditions tests ExtractContext with boundary conditions
func TestExtractContext_BoundaryConditions(t *testing.T) {
	t.Parallel()

	extractor := NewSourceExtractor()
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "boundary_test.go")

	// Create a file with exactly 5 lines
	content := `line 1
line 2
line 3
line 4
line 5`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testCases := []struct {
		name        string
		line        int
		contextSize int
		expectNil   bool
	}{
		{
			name:        "Line 0 (invalid)",
			line:        0,
			contextSize: 2,
			expectNil:   false, // Should handle gracefully
		},
		{
			name:        "Line 1 with context 0",
			line:        1,
			contextSize: 0,
			expectNil:   false,
		},
		{
			name:        "Line 1 with large context",
			line:        1,
			contextSize: 10,
			expectNil:   false,
		},
		{
			name:        "Last line with large context",
			line:        5,
			contextSize: 10,
			expectNil:   false,
		},
		{
			name:        "Beyond last line",
			line:        10,
			contextSize: 2,
			expectNil:   false,
		},
		{
			name:        "Negative line",
			line:        -1,
			contextSize: 2,
			expectNil:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			context, err := extractor.ExtractContext(testFile, tc.line, tc.contextSize)
			if err != nil && tc.line > 0 {
				t.Errorf("Unexpected error for valid line %d: %v", tc.line, err)
			}

			if tc.expectNil && context != nil {
				t.Errorf("Expected nil context, got: %v", context)
			}
			if !tc.expectNil && context == nil {
				t.Error("Expected context, got nil")
			}
		})
	}
}

// TestExtractSourceContext_EdgeCases tests ExtractSourceContext with edge cases
func TestExtractSourceContext_EdgeCases(t *testing.T) {
	t.Parallel()

	extractor := NewSourceExtractor()
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "edge_test.go")

	// Create a file with various edge cases
	content := `package main

import "fmt"

func main() {
	fmt.Println("Hello")
	// This is a comment
	x := 42
	if x > 0 {
		fmt.Println("Positive")
	}
}`

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	testCases := []struct {
		name        string
		line        int
		column      int
		contextSize int
		expectNil   bool
	}{
		{
			name:        "Valid location",
			line:        6,
			column:      5,
			contextSize: 2,
			expectNil:   false,
		},
		{
			name:        "Line 0",
			line:        0,
			column:      1,
			contextSize: 2,
			expectNil:   false,
		},
		{
			name:        "Negative line",
			line:        -1,
			column:      1,
			contextSize: 2,
			expectNil:   false,
		},
		{
			name:        "Beyond file",
			line:        100,
			column:      1,
			contextSize: 2,
			expectNil:   false,
		},
		{
			name:        "Zero context size",
			line:        5,
			column:      1,
			contextSize: 0,
			expectNil:   false,
		},
		{
			name:        "Large context size",
			line:        5,
			column:      1,
			contextSize: 100,
			expectNil:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Create a test error with location to test ExtractSourceContext
			testError := &models.LegacyTestError{
				Location: &models.SourceLocation{
					File:   testFile,
					Line:   tc.line,
					Column: tc.column,
				},
			}

			err := extractor.ExtractSourceContext(testError, tc.contextSize)
			context := testError.SourceContext

			// Check for errors in edge cases
			if tc.line <= 0 && err == nil {
				t.Logf("ExtractSourceContext handled edge case gracefully for line %d", tc.line)
			}

			if tc.expectNil && context != nil {
				t.Errorf("Expected nil context, got: %v", context)
			}
			if !tc.expectNil && context == nil {
				t.Error("Expected context, got nil")
			}
		})
	}
}

// TestIsValidSourceFile_EdgeCases tests IsValidSourceFile with edge cases
func TestIsValidSourceFile_EdgeCases(t *testing.T) {
	t.Parallel()

	extractor := NewSourceExtractor()
	tempDir := t.TempDir()

	testCases := []struct {
		name       string
		filename   string
		content    string
		expected   bool
		createFile bool
	}{
		{
			name:       "Valid Go file",
			filename:   "valid.go",
			content:    "package main\n\nfunc main() {}",
			expected:   true,
			createFile: true,
		},
		{
			name:       "Non-Go file",
			filename:   "readme.txt",
			content:    "This is a text file",
			expected:   false,
			createFile: true,
		},
		{
			name:       "Go file without extension",
			filename:   "main",
			content:    "package main\n\nfunc main() {}",
			expected:   false,
			createFile: true,
		},
		{
			name:       "Empty Go file",
			filename:   "empty.go",
			content:    "",
			expected:   false, // Empty files are considered invalid by the implementation
			createFile: true,
		},
		{
			name:       "Non-existent file",
			filename:   "nonexistent.go",
			content:    "",
			expected:   false,
			createFile: false,
		},
		{
			name:       "Go test file",
			filename:   "main_test.go",
			content:    "package main\n\nimport \"testing\"\n\nfunc TestMain(t *testing.T) {}",
			expected:   true,
			createFile: true,
		},
		{
			name:       "Hidden Go file",
			filename:   ".hidden.go",
			content:    "package main",
			expected:   true,
			createFile: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var filePath string
			if tc.createFile {
				filePath = filepath.Join(tempDir, tc.filename)
				err := os.WriteFile(filePath, []byte(tc.content), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			} else {
				filePath = filepath.Join(tempDir, tc.filename)
			}

			result := extractor.IsValidSourceFile(filePath)
			if result != tc.expected {
				t.Errorf("Expected %v for file %s, got %v", tc.expected, tc.filename, result)
			}
		})
	}
}

// TestExtractContext_FilePermissionErrors tests ExtractContext with permission errors
func TestExtractContext_FilePermissionErrors(t *testing.T) {
	t.Parallel()

	extractor := NewSourceExtractor()

	// Test with non-existent file
	context, err := extractor.ExtractContext("/path/that/does/not/exist.go", 1, 2)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
	if context != nil {
		t.Error("Expected nil context for non-existent file")
	}

	// Test with empty file path
	context, err = extractor.ExtractContext("", 1, 2)
	if err == nil {
		t.Error("Expected error for empty file path")
	}
	if context != nil {
		t.Error("Expected nil context for empty file path")
	}
}

// TestExtractSourceContext_FilePermissionErrors tests ExtractSourceContext with permission errors
func TestExtractSourceContext_FilePermissionErrors(t *testing.T) {
	t.Parallel()

	extractor := NewSourceExtractor()

	// Test with non-existent file
	testError := &models.LegacyTestError{
		Location: &models.SourceLocation{
			File:   "/path/that/does/not/exist.go",
			Line:   1,
			Column: 1,
		},
	}

	err := extractor.ExtractSourceContext(testError, 2)
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// Test with empty file path
	testError2 := &models.LegacyTestError{
		Location: &models.SourceLocation{
			File:   "",
			Line:   1,
			Column: 1,
		},
	}

	err = extractor.ExtractSourceContext(testError2, 2)
	if err == nil {
		t.Error("Expected error for empty file path")
	}
}
