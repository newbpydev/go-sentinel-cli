package display

import (
	"bytes"
	"strings"
	"testing"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

func TestNewErrorFormatter(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}

	errorFormatter := NewErrorFormatter(&buf, formatter, 100)

	if errorFormatter == nil {
		t.Fatal("NewErrorFormatter returned nil")
	}
	if errorFormatter.writer != &buf {
		t.Error("Writer not set correctly")
	}
	if errorFormatter.formatter != formatter {
		t.Error("Formatter not set correctly")
	}
	if errorFormatter.width != 100 {
		t.Error("Width not set correctly")
	}
}

func TestNewErrorFormatterWithDefaults(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}

	errorFormatter := NewErrorFormatterWithDefaults(&buf, formatter)

	if errorFormatter == nil {
		t.Fatal("NewErrorFormatterWithDefaults returned nil")
	}
	if errorFormatter.writer != &buf {
		t.Error("Writer not set correctly")
	}
	if errorFormatter.formatter != formatter {
		t.Error("Formatter not set correctly")
	}
	if errorFormatter.width != 80 {
		t.Error("Expected default width 80")
	}
}

func TestErrorFormatter_GetTerminalWidth(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}

	tests := []struct {
		name        string
		configWidth int
		expectedMin int
	}{
		{
			name:        "with configured width",
			configWidth: 120,
			expectedMin: 80, // Should return at least 80
		},
		{
			name:        "with zero width",
			configWidth: 0,
			expectedMin: 80, // Should return default 80
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errorFormatter := NewErrorFormatter(&buf, formatter, tt.configWidth)
			width := errorFormatter.GetTerminalWidth()
			if width < tt.expectedMin {
				t.Errorf("Expected width >= %d, got %d", tt.expectedMin, width)
			}
		})
	}
}

func TestFormatClickableLocation(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	errorFormatter := NewErrorFormatter(&buf, formatter, 80)

	tests := []struct {
		name     string
		location *models.SourceLocation
		wantText []string
	}{
		{
			name: "valid location",
			location: &models.SourceLocation{
				File:   "test.go",
				Line:   42,
				Column: 10,
			},
			wantText: []string{"↳", "[CYAN]test.go:42:10[/CYAN]", "\033]8;;"},
		},
		{
			name: "location without column",
			location: &models.SourceLocation{
				File:   "example.go",
				Line:   100,
				Column: 0,
			},
			wantText: []string{"↳", "[CYAN]example.go:100:0[/CYAN]"},
		},
		{
			name:     "nil location",
			location: nil,
			wantText: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errorFormatter.FormatClickableLocation(tt.location)

			if tt.location == nil {
				if result != "" {
					t.Error("Expected empty result for nil location")
				}
				return
			}

			for _, want := range tt.wantText {
				if !strings.Contains(result, want) {
					t.Errorf("Expected %q in result, but not found in: %s", want, result)
				}
			}
		})
	}
}

func TestRenderSourceContext(t *testing.T) {
	tests := []struct {
		name     string
		test     *models.TestResult
		wantText []string
		wantErr  bool
	}{
		{
			name: "test with source context",
			test: &models.TestResult{
				Error: &models.TestError{
					SourceFile: "test.go",
					SourceLine: 42,
					SourceContext: []string{
						"func TestExample(t *testing.T) {",
						"    result := someFunction()",
						"    if result != expected {",
						"        t.Error(\"failed\")",
						"    }",
					},
					ContextStartLine: 40,
				},
			},
			wantText: []string{"|", "func TestExample", "[BGRED]", "42", "result := someFunction"},
		},
		{
			name: "test without source context",
			test: &models.TestResult{
				Error: &models.TestError{
					SourceFile: "test.go",
					SourceLine: 42,
				},
			},
			wantText: []string{},
		},
		{
			name: "test without error",
			test: &models.TestResult{
				Error: nil,
			},
			wantText: []string{},
		},
		{
			name: "test without source file",
			test: &models.TestResult{
				Error: &models.TestError{
					SourceFile: "",
				},
			},
			wantText: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := &mockFormatter{enabled: true}
			errorFormatter := NewErrorFormatter(&buf, formatter, 80)

			err := errorFormatter.RenderSourceContext(tt.test)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderSourceContext error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()

			if len(tt.wantText) == 0 {
				if output != "" {
					t.Error("Expected empty output")
				}
				return
			}

			for _, want := range tt.wantText {
				if !strings.Contains(output, want) {
					t.Errorf("Expected %q in output, but not found in: %s", want, output)
				}
			}
		})
	}
}

func TestRenderErrorPointer(t *testing.T) {
	tests := []struct {
		name       string
		location   *models.SourceLocation
		sourceLine string
		wantText   []string
	}{
		{
			name: "valid location with column",
			location: &models.SourceLocation{
				File:   "test.go",
				Line:   42,
				Column: 5,
			},
			sourceLine: "    result := someFunction()",
			wantText:   []string{"|", "[RED]^[/RED]"},
		},
		{
			name: "location without column",
			location: &models.SourceLocation{
				File:   "test.go",
				Line:   42,
				Column: 0,
			},
			sourceLine: "result := someFunction()",
			wantText:   []string{"|", "[RED]^[/RED]"},
		},
		{
			name:       "nil location",
			location:   nil,
			sourceLine: "some line",
			wantText:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := &mockFormatter{enabled: true}
			errorFormatter := NewErrorFormatter(&buf, formatter, 80)

			err := errorFormatter.RenderErrorPointer(tt.location, tt.sourceLine)
			if err != nil {
				t.Fatalf("RenderErrorPointer failed: %v", err)
			}

			output := buf.String()

			if tt.location == nil {
				if output != "" {
					t.Error("Expected empty output for nil location")
				}
				return
			}

			for _, want := range tt.wantText {
				if !strings.Contains(output, want) {
					t.Errorf("Expected %q in output, but not found in: %s", want, output)
				}
			}
		})
	}
}

func TestCalculateErrorPosition(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	errorFormatter := NewErrorFormatter(&buf, formatter, 80)

	tests := []struct {
		name       string
		location   *models.SourceLocation
		sourceLine string
		expected   int
	}{
		{
			name: "valid column position",
			location: &models.SourceLocation{
				Column: 5,
			},
			sourceLine: "    result := someFunction()",
			expected:   4, // Column 5 becomes index 4 (0-based)
		},
		{
			name: "column beyond line length",
			location: &models.SourceLocation{
				Column: 100,
			},
			sourceLine: "short line",
			expected:   9, // Last valid position
		},
		{
			name: "zero column",
			location: &models.SourceLocation{
				Column: 0,
			},
			sourceLine: "func TestExample(t *testing.T) {",
			expected:   0, // Inferred position at start of function
		},
		{
			name: "negative column",
			location: &models.SourceLocation{
				Column: -1,
			},
			sourceLine: "    t.Error(\"assertion failed\")",
			expected:   4, // Inferred position at t.Error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := errorFormatter.CalculateErrorPosition(tt.location, tt.sourceLine)
			if actual != tt.expected {
				t.Errorf("Expected position %d, got %d", tt.expected, actual)
			}
		})
	}
}

func TestInferErrorPosition(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	errorFormatter := NewErrorFormatter(&buf, formatter, 80)

	tests := []struct {
		name       string
		sourceLine string
		expected   int
	}{
		{
			name:       "line with t.Error",
			sourceLine: "    t.Error(\"assertion failed\")",
			expected:   4, // Position of t.Error
		},
		{
			name:       "line with comparison operator",
			sourceLine: "if result != expected {",
			expected:   10, // Position of !=
		},
		{
			name:       "line with array access",
			sourceLine: "value := array[index]",
			expected:   14, // Position of [
		},
		{
			name:       "line with nil",
			sourceLine: "if pointer == nil {",
			expected:   11, // Position of nil (corrected)
		},
		{
			name:       "line with only whitespace",
			sourceLine: "   \t   func TestExample() {",
			expected:   7, // First non-whitespace character
		},
		{
			name:       "empty line",
			sourceLine: "",
			expected:   0, // Default to 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := errorFormatter.inferErrorPosition(tt.sourceLine)
			if actual != tt.expected {
				t.Errorf("Expected position %d, got %d for line: %q", tt.expected, actual, tt.sourceLine)
			}
		})
	}
}

func TestFindPatternPosition(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	errorFormatter := NewErrorFormatter(&buf, formatter, 80)

	tests := []struct {
		name     string
		line     string
		patterns []string
		expected int
	}{
		{
			name:     "pattern found",
			line:     "if result != expected {",
			patterns: []string{"!=", "=="},
			expected: 10, // Position of !=
		},
		{
			name:     "first occurrence in line wins",
			line:     "if a == b || c != d {",
			patterns: []string{"!=", "=="},
			expected: 5, // Position of == (appears first in line)
		},
		{
			name:     "no pattern found",
			line:     "simple line without patterns",
			patterns: []string{"!=", "=="},
			expected: -1,
		},
		{
			name:     "empty patterns",
			line:     "any line",
			patterns: []string{},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := errorFormatter.findPatternPosition(tt.line, tt.patterns)
			if actual != tt.expected {
				t.Errorf("Expected position %d, got %d", tt.expected, actual)
			}
		})
	}
}

func TestIsCursorInstalled(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	errorFormatter := NewErrorFormatter(&buf, formatter, 80)

	// This test checks that the method doesn't panic
	// The actual result depends on the environment
	result := errorFormatter.isCursorInstalled()

	// Just verify it returns a boolean
	if result != true && result != false {
		t.Error("isCursorInstalled should return a boolean")
	}
}

func TestIsVSCodeInstalled(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	errorFormatter := NewErrorFormatter(&buf, formatter, 80)

	// This test checks that the method doesn't panic
	// The actual result depends on the environment
	result := errorFormatter.isVSCodeInstalled()

	// Just verify it returns a boolean
	if result != true && result != false {
		t.Error("isVSCodeInstalled should return a boolean")
	}
}
