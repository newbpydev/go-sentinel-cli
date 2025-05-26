// Package display provides failed tests rendering tests
package display

import (
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

func TestNewFailedTestsRenderer(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		options  *FailedTestsRenderOptions
		expected struct {
			terminalWidth   int
			showCodeContext bool
			maxContextLines int
			indentLevel     int
		}
	}{
		{
			name:    "with nil options should use defaults",
			config:  &Config{},
			options: nil,
			expected: struct {
				terminalWidth   int
				showCodeContext bool
				maxContextLines int
				indentLevel     int
			}{
				terminalWidth:   110,
				showCodeContext: true,
				maxContextLines: 5,
				indentLevel:     0,
			},
		},
		{
			name:   "with partial options should apply defaults to zero values",
			config: &Config{},
			options: &FailedTestsRenderOptions{
				ShowCodeContext: false,
				MaxContextLines: 3,
			},
			expected: struct {
				terminalWidth   int
				showCodeContext bool
				maxContextLines int
				indentLevel     int
			}{
				terminalWidth:   110,
				showCodeContext: false,
				maxContextLines: 3,
				indentLevel:     0,
			},
		},
		{
			name:   "with all options specified",
			config: &Config{},
			options: &FailedTestsRenderOptions{
				TerminalWidth:   120,
				ShowCodeContext: true,
				MaxContextLines: 7,
				IndentLevel:     1,
			},
			expected: struct {
				terminalWidth   int
				showCodeContext bool
				maxContextLines int
				indentLevel     int
			}{
				terminalWidth:   120,
				showCodeContext: true,
				maxContextLines: 7,
				indentLevel:     1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewFailedTestsRenderer(tt.config, tt.options)

			if renderer.terminalWidth != tt.expected.terminalWidth {
				t.Errorf("terminalWidth = %d, expected %d", renderer.terminalWidth, tt.expected.terminalWidth)
			}
			if renderer.showCodeContext != tt.expected.showCodeContext {
				t.Errorf("showCodeContext = %v, expected %v", renderer.showCodeContext, tt.expected.showCodeContext)
			}
			if renderer.maxContextLines != tt.expected.maxContextLines {
				t.Errorf("maxContextLines = %d, expected %d", renderer.maxContextLines, tt.expected.maxContextLines)
			}
			if renderer.indentLevel != tt.expected.indentLevel {
				t.Errorf("indentLevel = %d, expected %d", renderer.indentLevel, tt.expected.indentLevel)
			}
		})
	}
}

func TestFailedTestsRenderer_RenderFailedTestsSection(t *testing.T) {
	tests := []struct {
		name        string
		failedTests []*models.TestResult
		failedCount int
		options     *FailedTestsRenderOptions
		expected    struct {
			contains    []string
			notContains []string
		}
	}{
		{
			name:        "empty failed tests should return empty string",
			failedTests: []*models.TestResult{},
			failedCount: 0,
			options:     nil,
			expected: struct {
				contains    []string
				notContains []string
			}{
				contains:    []string{},
				notContains: []string{"Failed Tests", "─"},
			},
		},
		{
			name: "single failed test with error details",
			failedTests: []*models.TestResult{
				{
					Name:     "TestCalculation",
					Package:  "pkg/calculator",
					Status:   models.TestStatusFailed,
					Duration: 15 * time.Millisecond,
					Error: &models.TestError{
						Type:         "AssertionError",
						Message:      "Expected 1+1 to equal 3, but got 2",
						SourceFile:   "calculator_test.go",
						SourceLine:   42,
						SourceColumn: 15,
						SourceContext: []string{
							"func TestCalculation(t *testing.T) {",
							"    result := Add(1, 1)",
							"    assert.Equal(t, 3, result)",
							"}",
						},
						ContextStartLine: 40,
					},
				},
			},
			failedCount: 1,
			options: &FailedTestsRenderOptions{
				TerminalWidth:   110,
				ShowCodeContext: true,
				MaxContextLines: 5,
			},
			expected: struct {
				contains    []string
				notContains []string
			}{
				contains: []string{
					"─", "Failed Tests 1", "FAIL", "TestCalculation",
					"AssertionError: Expected 1+1 to equal 3, but got 2",
					"calculator_test.go:42:15", "assert.Equal",
				},
				notContains: []string{"PASS", "SKIP"},
			},
		},
		{
			name: "multiple failed tests",
			failedTests: []*models.TestResult{
				{
					Name:     "TestFirstFailure",
					Package:  "pkg/math",
					Status:   models.TestStatusFailed,
					Duration: 10 * time.Millisecond,
					Error: &models.TestError{
						Type:       "TimeoutError",
						Message:    "Test timed out after 5 seconds",
						SourceFile: "math_test.go",
						SourceLine: 25,
					},
				},
				{
					Name:     "TestSecondFailure",
					Package:  "pkg/string",
					Status:   models.TestStatusFailed,
					Duration: 5 * time.Millisecond,
					Error: &models.TestError{
						Type:         "PanicError",
						Message:      "runtime error: index out of range",
						SourceFile:   "string_test.go",
						SourceLine:   18,
						SourceColumn: 8,
					},
				},
			},
			failedCount: 2,
			options: &FailedTestsRenderOptions{
				TerminalWidth:   120,
				ShowCodeContext: false,
			},
			expected: struct {
				contains    []string
				notContains []string
			}{
				contains: []string{
					"Failed Tests 2", "TestFirstFailure", "TestSecondFailure",
					"TimeoutError", "PanicError", "math_test.go:25", "string_test.go:18:8",
				},
				notContains: []string{"PASS", "SKIP"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewFailedTestsRenderer(&Config{}, tt.options)
			result := renderer.RenderFailedTestsSection(tt.failedTests, tt.failedCount)

			// If expected empty, verify empty result
			if len(tt.expected.contains) == 0 && len(tt.expected.notContains) > 0 {
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

			// Verify separator line has minimum 110 characters
			if tt.failedCount > 0 {
				lines := strings.Split(result, "\n")
				foundSeparator := false
				for _, line := range lines {
					if strings.Contains(line, "─") && len(line) >= 110 {
						foundSeparator = true
						break
					}
				}
				if !foundSeparator {
					t.Errorf("Expected to find separator line with at least 110 ─ characters")
				}
			}
		})
	}
}

func TestFailedTestsRenderer_renderSectionSeparator(t *testing.T) {
	renderer := NewFailedTestsRenderer(&Config{}, &FailedTestsRenderOptions{
		TerminalWidth: 120,
	})

	separator := renderer.renderSectionSeparator()

	// Should be exactly the terminal width in ─ characters (UTF-8 encoded)
	// Each ─ character is 3 bytes in UTF-8, so 120 characters = 360 bytes
	if len(separator) != 360 {
		t.Errorf("Expected separator length to be 360, got %d", len(separator))
	}

	// Should only contain ─ characters
	for _, char := range separator {
		if char != '─' {
			t.Errorf("Expected separator to only contain ─ characters, found: %c", char)
		}
	}

	// Test minimum width enforcement
	renderer.SetTerminalWidth(50) // Below minimum
	separatorMin := renderer.renderSectionSeparator()
	// 110 ─ characters = 330 bytes
	if len(separatorMin) != 330 {
		t.Errorf("Expected minimum separator length to be 330, got %d", len(separatorMin))
	}
}

func TestFailedTestsRenderer_renderSectionHeader(t *testing.T) {
	renderer := NewFailedTestsRenderer(&Config{}, &FailedTestsRenderOptions{
		TerminalWidth: 110,
	})

	tests := []struct {
		name     string
		title    string
		count    int
		expected struct {
			contains    []string
			totalLength int
		}
	}{
		{
			name:  "failed tests header with count",
			title: "Failed Tests",
			count: 5,
			expected: struct {
				contains    []string
				totalLength int
			}{
				contains:    []string{"Failed Tests", "5"},
				totalLength: 110,
			},
		},
		{
			name:  "zero count",
			title: "Failed Tests",
			count: 0,
			expected: struct {
				contains    []string
				totalLength int
			}{
				contains:    []string{"Failed Tests", "0"},
				totalLength: 110,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := renderer.renderSectionHeader(tt.title, tt.count)

			for _, expected := range tt.expected.contains {
				if !strings.Contains(header, expected) {
					t.Errorf("Expected header to contain %q, got: %q", expected, header)
				}
			}

			if len(header) != tt.expected.totalLength {
				t.Errorf("Expected header length to be %d, got %d", tt.expected.totalLength, len(header))
			}
		})
	}
}

func TestFailedTestsRenderer_renderFailureHeader(t *testing.T) {
	renderer := NewFailedTestsRenderer(&Config{}, nil)

	tests := []struct {
		name     string
		test     *models.TestResult
		expected struct {
			contains []string
		}
	}{
		{
			name: "basic failure header",
			test: &models.TestResult{
				Name:    "TestExample",
				Package: "pkg/example",
				Status:  models.TestStatusFailed,
				Error: &models.TestError{
					SourceFile: "example_test.go",
				},
			},
			expected: struct {
				contains []string
			}{
				contains: []string{"FAIL", "example_test.go", ">", "TestExample"},
			},
		},
		{
			name: "failure header with unknown source",
			test: &models.TestResult{
				Name:    "TestUnknown",
				Package: "pkg/unknown",
				Status:  models.TestStatusFailed,
				Error:   &models.TestError{},
			},
			expected: struct {
				contains []string
			}{
				contains: []string{"FAIL", "pkg/unknown", ">", "TestUnknown"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := renderer.renderFailureHeader(tt.test)

			for _, expected := range tt.expected.contains {
				if !strings.Contains(header, expected) {
					t.Errorf("Expected header to contain %q, got: %q", expected, header)
				}
			}
		})
	}
}

func TestFailedTestsRenderer_renderErrorMessage(t *testing.T) {
	renderer := NewFailedTestsRenderer(&Config{}, nil)

	tests := []struct {
		name     string
		err      *models.TestError
		expected string
	}{
		{
			name: "complete error",
			err: &models.TestError{
				Type:    "AssertionError",
				Message: "Values do not match",
			},
			expected: "AssertionError: Values do not match",
		},
		{
			name: "error without type",
			err: &models.TestError{
				Message: "Something went wrong",
			},
			expected: "Error: Something went wrong",
		},
		{
			name: "error without message",
			err: &models.TestError{
				Type: "RuntimeError",
			},
			expected: "RuntimeError: Test failed",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.renderErrorMessage(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFailedTestsRenderer_renderSourceLocation(t *testing.T) {
	renderer := NewFailedTestsRenderer(&Config{}, nil)

	tests := []struct {
		name     string
		err      *models.TestError
		expected struct {
			contains []string
		}
	}{
		{
			name: "complete source location",
			err: &models.TestError{
				SourceFile:   "test.go",
				SourceLine:   42,
				SourceColumn: 15,
			},
			expected: struct {
				contains []string
			}{
				contains: []string{"test.go:42:15"},
			},
		},
		{
			name: "source location without column",
			err: &models.TestError{
				SourceFile: "test.go",
				SourceLine: 25,
			},
			expected: struct {
				contains []string
			}{
				contains: []string{"test.go:25"},
			},
		},
		{
			name: "nil error",
			err:  nil,
			expected: struct {
				contains []string
			}{
				contains: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.renderSourceLocation(tt.err)

			if len(tt.expected.contains) == 0 {
				if result != "" {
					t.Errorf("Expected empty result, got %q", result)
				}
				return
			}

			for _, expected := range tt.expected.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain %q, got: %q", expected, result)
				}
			}
		})
	}
}

func TestFailedTestsRenderer_SettersAndGetters(t *testing.T) {
	renderer := NewFailedTestsRenderer(&Config{}, nil)

	// Test SetTerminalWidth
	renderer.SetTerminalWidth(150)
	if renderer.GetTerminalWidth() != 150 {
		t.Errorf("Expected terminal width 150, got %d", renderer.GetTerminalWidth())
	}

	// Test minimum width enforcement
	renderer.SetTerminalWidth(50)
	if renderer.GetTerminalWidth() != 110 {
		t.Errorf("Expected minimum terminal width 110, got %d", renderer.GetTerminalWidth())
	}

	// Test SetShowCodeContext
	renderer.SetShowCodeContext(false)
	if renderer.IsShowCodeContext() != false {
		t.Errorf("Expected ShowCodeContext to be false")
	}

	renderer.SetShowCodeContext(true)
	if renderer.IsShowCodeContext() != true {
		t.Errorf("Expected ShowCodeContext to be true")
	}

	// Test SetMaxContextLines
	renderer.SetMaxContextLines(10)
	if renderer.GetMaxContextLines() != 10 {
		t.Errorf("Expected MaxContextLines 10, got %d", renderer.GetMaxContextLines())
	}

	// Test minimum lines enforcement
	renderer.SetMaxContextLines(0)
	if renderer.GetMaxContextLines() != 1 {
		t.Errorf("Expected minimum MaxContextLines 1, got %d", renderer.GetMaxContextLines())
	}
}
