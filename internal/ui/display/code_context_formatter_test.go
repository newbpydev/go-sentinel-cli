// Package display provides code context formatting tests
package display

import (
	"strings"
	"testing"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

func TestNewCodeContextFormatter(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		options  *CodeContextFormatOptions
		expected struct {
			maxContextLines  int
			lineNumberWidth  int
			showLineNumbers  bool
			showErrorPointer bool
			indentLevel      int
		}
	}{
		{
			name:    "with nil options should use defaults",
			config:  &Config{},
			options: nil,
			expected: struct {
				maxContextLines  int
				lineNumberWidth  int
				showLineNumbers  bool
				showErrorPointer bool
				indentLevel      int
			}{
				maxContextLines:  5,
				lineNumberWidth:  4,
				showLineNumbers:  true,
				showErrorPointer: true,
				indentLevel:      0,
			},
		},
		{
			name:   "with partial options should apply defaults to zero values",
			config: &Config{},
			options: &CodeContextFormatOptions{
				ShowLineNumbers:  false,
				ShowErrorPointer: false,
			},
			expected: struct {
				maxContextLines  int
				lineNumberWidth  int
				showLineNumbers  bool
				showErrorPointer bool
				indentLevel      int
			}{
				maxContextLines:  5,
				lineNumberWidth:  4,
				showLineNumbers:  false,
				showErrorPointer: false,
				indentLevel:      0,
			},
		},
		{
			name:   "with all options specified",
			config: &Config{},
			options: &CodeContextFormatOptions{
				MaxContextLines:  7,
				LineNumberWidth:  6,
				ShowLineNumbers:  true,
				ShowErrorPointer: true,
				IndentLevel:      2,
			},
			expected: struct {
				maxContextLines  int
				lineNumberWidth  int
				showLineNumbers  bool
				showErrorPointer bool
				indentLevel      int
			}{
				maxContextLines:  7,
				lineNumberWidth:  6,
				showLineNumbers:  true,
				showErrorPointer: true,
				indentLevel:      2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewCodeContextFormatter(tt.config, tt.options)

			if formatter.maxContextLines != tt.expected.maxContextLines {
				t.Errorf("maxContextLines = %d, expected %d", formatter.maxContextLines, tt.expected.maxContextLines)
			}
			if formatter.lineNumberWidth != tt.expected.lineNumberWidth {
				t.Errorf("lineNumberWidth = %d, expected %d", formatter.lineNumberWidth, tt.expected.lineNumberWidth)
			}
			if formatter.showLineNumbers != tt.expected.showLineNumbers {
				t.Errorf("showLineNumbers = %v, expected %v", formatter.showLineNumbers, tt.expected.showLineNumbers)
			}
			if formatter.showErrorPointer != tt.expected.showErrorPointer {
				t.Errorf("showErrorPointer = %v, expected %v", formatter.showErrorPointer, tt.expected.showErrorPointer)
			}
			if formatter.indentLevel != tt.expected.indentLevel {
				t.Errorf("indentLevel = %d, expected %d", formatter.indentLevel, tt.expected.indentLevel)
			}
		})
	}
}

func TestCodeContextFormatter_FormatCodeContext(t *testing.T) {
	tests := []struct {
		name     string
		err      *models.TestError
		options  *CodeContextFormatOptions
		expected struct {
			contains    []string
			notContains []string
			lineCount   int
		}
	}{
		{
			name: "nil error should return empty string",
			err:  nil,
			expected: struct {
				contains    []string
				notContains []string
				lineCount   int
			}{
				contains:    []string{},
				notContains: []string{},
				lineCount:   0,
			},
		},
		{
			name: "error without source context should return empty string",
			err: &models.TestError{
				Type:    "Error",
				Message: "Test failed",
			},
			expected: struct {
				contains    []string
				notContains []string
				lineCount   int
			}{
				contains:    []string{},
				notContains: []string{},
				lineCount:   0,
			},
		},
		{
			name: "basic code context with line numbers",
			err: &models.TestError{
				SourceLine:   42,
				SourceColumn: 15,
				SourceContext: []string{
					"func TestExample(t *testing.T) {",
					"    result := Calculate(5, 3)",
					"    assert.Equal(t, 8, result)",
					"    // This should pass",
					"}",
				},
				ContextStartLine: 40,
			},
			options: &CodeContextFormatOptions{
				MaxContextLines:  5,
				ShowLineNumbers:  true,
				ShowErrorPointer: true,
			},
			expected: struct {
				contains    []string
				notContains []string
				lineCount   int
			}{
				contains: []string{
					"40|", "41|", "42|", "43|", "44|",
					"func TestExample", "Calculate(5, 3)", "assert.Equal",
					"^", // Error pointer
				},
				notContains: []string{},
				lineCount:   6, // 5 code lines + 1 pointer line
			},
		},
		{
			name: "code context without line numbers",
			err: &models.TestError{
				SourceLine:   25,
				SourceColumn: 8,
				SourceContext: []string{
					"x := 10",
					"y := 20",
					"result := x + y",
				},
				ContextStartLine: 23,
			},
			options: &CodeContextFormatOptions{
				MaxContextLines:  3,
				ShowLineNumbers:  false,
				ShowErrorPointer: true,
			},
			expected: struct {
				contains    []string
				notContains []string
				lineCount   int
			}{
				contains: []string{
					"x := 10", "y := 20", "result := x + y",
					"^", // Error pointer
				},
				notContains: []string{"23|", "24|", "25|"},
				lineCount:   4, // 3 code lines + 1 pointer line
			},
		},
		{
			name: "code context without error pointer",
			err: &models.TestError{
				SourceLine:   15,
				SourceColumn: 0, // No column means no pointer
				SourceContext: []string{
					"package main",
					"",
					"import \"fmt\"",
				},
				ContextStartLine: 1,
			},
			options: &CodeContextFormatOptions{
				MaxContextLines:  3,
				ShowLineNumbers:  true,
				ShowErrorPointer: true,
			},
			expected: struct {
				contains    []string
				notContains []string
				lineCount   int
			}{
				contains: []string{
					"1|", "2|", "3|",
					"package main", "import \"fmt\"",
				},
				notContains: []string{"^"}, // No pointer due to column 0
				lineCount:   3,             // Just 3 code lines
			},
		},
		{
			name: "long code context gets limited",
			err: &models.TestError{
				SourceLine:   50,
				SourceColumn: 5,
				SourceContext: []string{
					"line1", "line2", "line3", "line4", "line5",
					"line6", "line7", "line8", "line9", "line10",
				},
				ContextStartLine: 45,
			},
			options: &CodeContextFormatOptions{
				MaxContextLines:  3, // Limit to 3 lines
				ShowLineNumbers:  true,
				ShowErrorPointer: true,
			},
			expected: struct {
				contains    []string
				notContains []string
				lineCount   int
			}{
				contains: []string{
					"45|", "46|", "47|",
					"line1", "line2", "line3",
				},
				notContains: []string{"line4", "line5", "line6", "48|", "^"}, // No pointer since error line is outside displayed context
				lineCount:   3,                                               // 3 limited lines only (no pointer)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewCodeContextFormatter(&Config{}, tt.options)
			result := formatter.FormatCodeContext(tt.err)

			// If expected empty, verify empty result
			if len(tt.expected.contains) == 0 {
				if result != "" {
					t.Errorf("Expected empty result, got: %q", result)
				}
				return
			}

			// Check that all expected strings are present
			for _, expected := range tt.expected.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, but it didn't. Result:\n%s", expected, result)
				}
			}

			// Check that unwanted strings are not present
			for _, notExpected := range tt.expected.notContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("Expected result to not contain %q, but it did. Result:\n%s", notExpected, result)
				}
			}

			// Verify line count
			lines := strings.Split(strings.TrimRight(result, "\n"), "\n")
			actualLineCount := len(lines)
			if result == "" {
				actualLineCount = 0
			}
			if actualLineCount != tt.expected.lineCount {
				t.Errorf("Expected %d lines, got %d. Result:\n%s", tt.expected.lineCount, actualLineCount, result)
			}
		})
	}
}

func TestCodeContextFormatter_FormatSingleLine(t *testing.T) {
	formatter := NewCodeContextFormatter(&Config{}, &CodeContextFormatOptions{
		ShowLineNumbers: true,
		LineNumberWidth: 4,
	})

	tests := []struct {
		name       string
		line       string
		lineNumber int
		expected   struct {
			contains []string
		}
	}{
		{
			name:       "simple code line",
			line:       "fmt.Println(\"Hello, World!\")",
			lineNumber: 15,
			expected: struct {
				contains []string
			}{
				contains: []string{"15|", "fmt.Println", "Hello, World!"},
			},
		},
		{
			name:       "empty line",
			line:       "",
			lineNumber: 5,
			expected: struct {
				contains []string
			}{
				contains: []string{"5|"},
			},
		},
		{
			name:       "line with indentation",
			line:       "    return result",
			lineNumber: 42,
			expected: struct {
				contains []string
			}{
				contains: []string{"42|", "return result"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.FormatSingleLine(tt.line, tt.lineNumber)

			for _, expected := range tt.expected.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %q", expected, result)
				}
			}
		})
	}
}

func TestCodeContextFormatter_calculateStartLine(t *testing.T) {
	formatter := NewCodeContextFormatter(&Config{}, &CodeContextFormatOptions{
		MaxContextLines: 5,
	})

	tests := []struct {
		name     string
		err      *models.TestError
		expected int
	}{
		{
			name: "explicit context start line",
			err: &models.TestError{
				SourceLine:       25,
				ContextStartLine: 20,
			},
			expected: 20,
		},
		{
			name: "calculate from source line",
			err: &models.TestError{
				SourceLine:       25,
				ContextStartLine: 0, // Will be calculated
			},
			expected: 23, // 25 - (5/2) = 23
		},
		{
			name: "calculate with minimum boundary",
			err: &models.TestError{
				SourceLine:       2,
				ContextStartLine: 0, // Will be calculated
			},
			expected: 1, // Cannot go below 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatter.calculateStartLine(tt.err)
			if result != tt.expected {
				t.Errorf("Expected start line %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestCodeContextFormatter_FormatMultipleErrors(t *testing.T) {
	formatter := NewCodeContextFormatter(&Config{}, &CodeContextFormatOptions{
		MaxContextLines:  3,
		ShowLineNumbers:  true,
		ShowErrorPointer: true,
	})

	errors := []*models.TestError{
		{
			SourceLine:   10,
			SourceColumn: 5,
			SourceContext: []string{
				"first error line 1",
				"first error line 2",
				"first error line 3",
			},
			ContextStartLine: 8,
		},
		{
			SourceLine:   20,
			SourceColumn: 10,
			SourceContext: []string{
				"second error line 1",
				"second error line 2",
			},
			ContextStartLine: 19,
		},
	}

	result := formatter.FormatMultipleErrors(errors)

	// Should contain content from both errors
	expectedContains := []string{
		"first error line 1", "first error line 2",
		"second error line 1", "second error line 2",
		"8|", "9|", "10|", "19|", "20|",
	}

	for _, expected := range expectedContains {
		if !strings.Contains(result, expected) {
			t.Errorf("Expected result to contain %q, but it didn't. Result:\n%s", expected, result)
		}
	}

	// Should have multiple ^ pointers
	pointerCount := strings.Count(result, "^")
	if pointerCount != 2 {
		t.Errorf("Expected 2 error pointers, got %d", pointerCount)
	}
}

func TestCodeContextFormatter_SettersAndGetters(t *testing.T) {
	formatter := NewCodeContextFormatter(&Config{}, nil)

	// Test SetMaxContextLines
	formatter.SetMaxContextLines(10)
	if formatter.GetMaxContextLines() != 10 {
		t.Errorf("Expected MaxContextLines 10, got %d", formatter.GetMaxContextLines())
	}

	// Test minimum lines enforcement
	formatter.SetMaxContextLines(0)
	if formatter.GetMaxContextLines() != 1 {
		t.Errorf("Expected minimum MaxContextLines 1, got %d", formatter.GetMaxContextLines())
	}

	// Test SetLineNumberWidth
	formatter.SetLineNumberWidth(8)
	if formatter.GetLineNumberWidth() != 8 {
		t.Errorf("Expected LineNumberWidth 8, got %d", formatter.GetLineNumberWidth())
	}

	// Test minimum width enforcement
	formatter.SetLineNumberWidth(0)
	if formatter.GetLineNumberWidth() != 1 {
		t.Errorf("Expected minimum LineNumberWidth 1, got %d", formatter.GetLineNumberWidth())
	}

	// Test SetShowLineNumbers
	formatter.SetShowLineNumbers(false)
	if formatter.IsShowLineNumbers() != false {
		t.Errorf("Expected ShowLineNumbers to be false")
	}

	formatter.SetShowLineNumbers(true)
	if formatter.IsShowLineNumbers() != true {
		t.Errorf("Expected ShowLineNumbers to be true")
	}

	// Test SetShowErrorPointer
	formatter.SetShowErrorPointer(false)
	if formatter.IsShowErrorPointer() != false {
		t.Errorf("Expected ShowErrorPointer to be false")
	}

	formatter.SetShowErrorPointer(true)
	if formatter.IsShowErrorPointer() != true {
		t.Errorf("Expected ShowErrorPointer to be true")
	}

	// Test SetIndentLevel
	formatter.SetIndentLevel(3)
	if formatter.GetIndentLevel() != 3 {
		t.Errorf("Expected IndentLevel 3, got %d", formatter.GetIndentLevel())
	}
}

func TestCodeContextFormatter_intToString(t *testing.T) {
	formatter := NewCodeContextFormatter(&Config{}, nil)

	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{123, "123"},
		{-5, "-5"},
		{-42, "-42"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := formatter.intToString(tt.input)
			if result != tt.expected {
				t.Errorf("intToString(%d) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
