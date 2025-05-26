// Package display provides file summary rendering tests
package display

import (
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

func TestNewFileSummaryRenderer(t *testing.T) {
	tests := []struct {
		name     string
		config   *Config
		options  *FileSummaryRenderOptions
		expected struct {
			showMemoryUsage bool
			showTiming      bool
			maxPathLength   int
			indentLevel     int
		}
	}{
		{
			name:    "with nil options should use defaults",
			config:  &Config{},
			options: nil,
			expected: struct {
				showMemoryUsage bool
				showTiming      bool
				maxPathLength   int
				indentLevel     int
			}{
				showMemoryUsage: true,
				showTiming:      true,
				maxPathLength:   60,
				indentLevel:     0,
			},
		},
		{
			name:   "with custom options",
			config: &Config{},
			options: &FileSummaryRenderOptions{
				ShowMemoryUsage: false,
				ShowTiming:      false,
				MaxPathLength:   40,
				IndentLevel:     2,
			},
			expected: struct {
				showMemoryUsage bool
				showTiming      bool
				maxPathLength   int
				indentLevel     int
			}{
				showMemoryUsage: false,
				showTiming:      false,
				maxPathLength:   40,
				indentLevel:     2,
			},
		},
		{
			name:   "with zero MaxPathLength should use default",
			config: &Config{},
			options: &FileSummaryRenderOptions{
				ShowMemoryUsage: true,
				ShowTiming:      true,
				MaxPathLength:   0, // Should default to 60
				IndentLevel:     1,
			},
			expected: struct {
				showMemoryUsage bool
				showTiming      bool
				maxPathLength   int
				indentLevel     int
			}{
				showMemoryUsage: true,
				showTiming:      true,
				maxPathLength:   60,
				indentLevel:     1,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewFileSummaryRenderer(tt.config, tt.options)

			if renderer == nil {
				t.Fatal("Expected renderer to be created, got nil")
			}

			if renderer.showMemoryUsage != tt.expected.showMemoryUsage {
				t.Errorf("Expected showMemoryUsage %v, got %v", tt.expected.showMemoryUsage, renderer.showMemoryUsage)
			}

			if renderer.showTiming != tt.expected.showTiming {
				t.Errorf("Expected showTiming %v, got %v", tt.expected.showTiming, renderer.showTiming)
			}

			if renderer.maxPathLength != tt.expected.maxPathLength {
				t.Errorf("Expected maxPathLength %d, got %d", tt.expected.maxPathLength, renderer.maxPathLength)
			}

			if renderer.indentLevel != tt.expected.indentLevel {
				t.Errorf("Expected indentLevel %d, got %d", tt.expected.indentLevel, renderer.indentLevel)
			}
		})
	}
}

func TestFileSummaryRenderer_RenderFileSummary(t *testing.T) {
	tests := []struct {
		name     string
		suite    *models.TestSuite
		options  *FileSummaryRenderOptions
		expected struct {
			contains    []string
			notContains []string
		}
	}{
		{
			name: "basic file summary with all tests passed",
			suite: &models.TestSuite{
				FilePath:     "pkg/example_test.go",
				TestCount:    3,
				PassedCount:  3,
				FailedCount:  0,
				SkippedCount: 0,
				Duration:     150 * time.Millisecond,
				MemoryUsage:  5 * 1024 * 1024, // 5 MB
			},
			options: &FileSummaryRenderOptions{
				ShowMemoryUsage: true,
				ShowTiming:      true,
				MaxPathLength:   60,
				IndentLevel:     0,
			},
			expected: struct {
				contains    []string
				notContains []string
			}{
				contains:    []string{"example_test.go", "(3 tests)", "150ms", "5 MB heap used"},
				notContains: []string{"failed"},
			},
		},
		{
			name: "file summary with failed tests",
			suite: &models.TestSuite{
				FilePath:     "pkg/failing_test.go",
				TestCount:    5,
				PassedCount:  3,
				FailedCount:  2,
				SkippedCount: 0,
				Duration:     200 * time.Millisecond,
				MemoryUsage:  10 * 1024 * 1024, // 10 MB
			},
			options: &FileSummaryRenderOptions{
				ShowMemoryUsage: true,
				ShowTiming:      true,
				MaxPathLength:   60,
				IndentLevel:     0,
			},
			expected: struct {
				contains    []string
				notContains []string
			}{
				contains:    []string{"failing_test.go", "(5 tests | 2 failed)", "200ms", "10 MB heap used"},
				notContains: []string{},
			},
		},
		{
			name: "file summary with single test and single failure",
			suite: &models.TestSuite{
				FilePath:     "pkg/single_test.go",
				TestCount:    1,
				PassedCount:  0,
				FailedCount:  1,
				SkippedCount: 0,
				Duration:     50 * time.Millisecond,
				MemoryUsage:  1024 * 1024, // 1 MB
			},
			options: &FileSummaryRenderOptions{
				ShowMemoryUsage: true,
				ShowTiming:      true,
				MaxPathLength:   60,
				IndentLevel:     0,
			},
			expected: struct {
				contains    []string
				notContains []string
			}{
				contains:    []string{"single_test.go", "(1 test | 1 failed)", "50ms", "1 MB heap used"},
				notContains: []string{},
			},
		},
		{
			name: "file summary without timing and memory",
			suite: &models.TestSuite{
				FilePath:     "pkg/minimal_test.go",
				TestCount:    2,
				PassedCount:  2,
				FailedCount:  0,
				SkippedCount: 0,
				Duration:     100 * time.Millisecond,
				MemoryUsage:  2 * 1024 * 1024, // 2 MB
			},
			options: &FileSummaryRenderOptions{
				ShowMemoryUsage: false,
				ShowTiming:      false,
				MaxPathLength:   60,
				IndentLevel:     0,
			},
			expected: struct {
				contains    []string
				notContains []string
			}{
				contains:    []string{"minimal_test.go", "(2 tests)"},
				notContains: []string{"100ms", "2 MB heap used"},
			},
		},
		{
			name: "file summary with long filename truncation",
			suite: &models.TestSuite{
				FilePath:     "pkg/very/long/path/to/a/test/file/with/extremely/long/name_test.go",
				TestCount:    1,
				PassedCount:  1,
				FailedCount:  0,
				SkippedCount: 0,
				Duration:     25 * time.Millisecond,
				MemoryUsage:  512 * 1024, // 0.5 MB
			},
			options: &FileSummaryRenderOptions{
				ShowMemoryUsage: true,
				ShowTiming:      true,
				MaxPathLength:   20,
				IndentLevel:     0,
			},
			expected: struct {
				contains    []string
				notContains []string
			}{
				contains:    []string{"name_test.go", "(1 test)", "25ms", "0.5 MB heap used"},
				notContains: []string{"extremely"},
			},
		},
		{
			name: "file summary with indentation",
			suite: &models.TestSuite{
				FilePath:     "pkg/indented_test.go",
				TestCount:    2,
				PassedCount:  2,
				FailedCount:  0,
				SkippedCount: 0,
				Duration:     75 * time.Millisecond,
				MemoryUsage:  3 * 1024 * 1024, // 3 MB
			},
			options: &FileSummaryRenderOptions{
				ShowMemoryUsage: true,
				ShowTiming:      true,
				MaxPathLength:   60,
				IndentLevel:     2,
			},
			expected: struct {
				contains    []string
				notContains []string
			}{
				contains:    []string{"    indented_test.go", "(2 tests)", "75ms", "3 MB heap used"},
				notContains: []string{},
			},
		},
		{
			name: "file summary with zero memory usage",
			suite: &models.TestSuite{
				FilePath:     "pkg/zero_memory_test.go",
				TestCount:    1,
				PassedCount:  1,
				FailedCount:  0,
				SkippedCount: 0,
				Duration:     10 * time.Millisecond,
				MemoryUsage:  0,
			},
			options: &FileSummaryRenderOptions{
				ShowMemoryUsage: true,
				ShowTiming:      true,
				MaxPathLength:   60,
				IndentLevel:     0,
			},
			expected: struct {
				contains    []string
				notContains []string
			}{
				contains:    []string{"zero_memory_test.go", "(1 test)", "10ms", "0 MB heap used"},
				notContains: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewFileSummaryRenderer(&Config{}, tt.options)
			result := renderer.RenderFileSummary(tt.suite)

			for _, expected := range tt.expected.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', but it didn't. Result: %s", expected, result)
				}
			}

			for _, notExpected := range tt.expected.notContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("Expected result to NOT contain '%s', but it did. Result: %s", notExpected, result)
				}
			}
		})
	}
}

func TestFileSummaryRenderer_RenderFileSummary_NilSuite(t *testing.T) {
	renderer := NewFileSummaryRenderer(&Config{}, nil)
	result := renderer.RenderFileSummary(nil)

	if result != "" {
		t.Errorf("Expected empty string for nil suite, got: %s", result)
	}
}

func TestFileSummaryRenderer_RenderFileSummaries(t *testing.T) {
	suite1 := &models.TestSuite{
		FilePath:     "pkg/test1.go",
		TestCount:    2,
		PassedCount:  2,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     100 * time.Millisecond,
		MemoryUsage:  2 * 1024 * 1024,
	}

	suite2 := &models.TestSuite{
		FilePath:     "pkg/test2.go",
		TestCount:    3,
		PassedCount:  2,
		FailedCount:  1,
		SkippedCount: 0,
		Duration:     150 * time.Millisecond,
		MemoryUsage:  3 * 1024 * 1024,
	}

	tests := []struct {
		name     string
		suites   []*models.TestSuite
		expected struct {
			contains    []string
			lineCount   int
			notContains []string
		}
	}{
		{
			name:   "multiple suites",
			suites: []*models.TestSuite{suite1, suite2},
			expected: struct {
				contains    []string
				lineCount   int
				notContains []string
			}{
				contains:    []string{"test1.go", "(2 tests)", "test2.go", "(3 tests | 1 failed)"},
				lineCount:   2,
				notContains: []string{},
			},
		},
		{
			name:   "empty suites",
			suites: []*models.TestSuite{},
			expected: struct {
				contains    []string
				lineCount   int
				notContains []string
			}{
				contains:    []string{},
				lineCount:   0,
				notContains: []string{},
			},
		},
		{
			name:   "suites with nil",
			suites: []*models.TestSuite{suite1, nil, suite2},
			expected: struct {
				contains    []string
				lineCount   int
				notContains []string
			}{
				contains:    []string{"test1.go", "test2.go"},
				lineCount:   2,
				notContains: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewFileSummaryRenderer(&Config{}, nil)
			result := renderer.RenderFileSummaries(tt.suites)

			for _, expected := range tt.expected.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', but it didn't. Result: %s", expected, result)
				}
			}

			for _, notExpected := range tt.expected.notContains {
				if strings.Contains(result, notExpected) {
					t.Errorf("Expected result to NOT contain '%s', but it did. Result: %s", notExpected, result)
				}
			}

			if tt.expected.lineCount > 0 {
				lines := strings.Split(strings.TrimSpace(result), "\n")
				if len(lines) != tt.expected.lineCount {
					t.Errorf("Expected %d lines, got %d. Result: %s", tt.expected.lineCount, len(lines), result)
				}
			} else if result != "" {
				t.Errorf("Expected empty result for empty suites, got: %s", result)
			}
		})
	}
}

func TestFileSummaryRenderer_formatFilename(t *testing.T) {
	renderer := NewFileSummaryRenderer(&Config{}, &FileSummaryRenderOptions{
		MaxPathLength: 20,
	})

	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "short filename",
			filename: "test.go",
			expected: "test.go",
		},
		{
			name:     "filename with path",
			filename: "pkg/example/test.go",
			expected: "test.go",
		},
		{
			name:     "long filename gets truncated",
			filename: "very_long_test_filename.go",
			expected: "very_long_test_fi...",
		},
		{
			name:     "empty filename",
			filename: "",
			expected: "(unknown file)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.formatFilename(tt.filename)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFileSummaryRenderer_formatMemoryUsage(t *testing.T) {
	renderer := NewFileSummaryRenderer(&Config{}, nil)

	tests := []struct {
		name        string
		memoryBytes uint64
		expected    string
	}{
		{
			name:        "zero memory",
			memoryBytes: 0,
			expected:    "0 MB heap used",
		},
		{
			name:        "small memory (less than 0.1 MB)",
			memoryBytes: 50 * 1024, // 50 KB
			expected:    "0 MB heap used",
		},
		{
			name:        "fractional MB",
			memoryBytes: 512 * 1024, // 0.5 MB
			expected:    "0.5 MB heap used",
		},
		{
			name:        "exactly 1 MB",
			memoryBytes: 1024 * 1024,
			expected:    "1 MB heap used",
		},
		{
			name:        "multiple MB",
			memoryBytes: 5 * 1024 * 1024,
			expected:    "5 MB heap used",
		},
		{
			name:        "large memory",
			memoryBytes: 1024 * 1024 * 1024, // 1 GB
			expected:    "1024 MB heap used",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderer.formatMemoryUsage(tt.memoryBytes)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestFileSummaryRenderer_SettersAndGetters(t *testing.T) {
	renderer := NewFileSummaryRenderer(&Config{}, nil)

	// Test SetIndentLevel and GetIndentLevel
	renderer.SetIndentLevel(3)
	if renderer.GetIndentLevel() != 3 {
		t.Errorf("Expected indent level 3, got %d", renderer.GetIndentLevel())
	}

	// Test SetShowTiming and IsShowTiming
	renderer.SetShowTiming(false)
	if renderer.IsShowTiming() != false {
		t.Errorf("Expected show timing false, got %v", renderer.IsShowTiming())
	}

	// Test SetShowMemoryUsage and IsShowMemoryUsage
	renderer.SetShowMemoryUsage(false)
	if renderer.IsShowMemoryUsage() != false {
		t.Errorf("Expected show memory usage false, got %v", renderer.IsShowMemoryUsage())
	}

	// Test SetMaxPathLength and GetMaxPathLength
	renderer.SetMaxPathLength(30)
	if renderer.GetMaxPathLength() != 30 {
		t.Errorf("Expected max path length 30, got %d", renderer.GetMaxPathLength())
	}
}

func TestFileSummaryRenderer_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		options *FileSummaryRenderOptions
		suite   *models.TestSuite
		test    func(t *testing.T, renderer *FileSummaryRenderer, result string)
	}{
		{
			name: "very small max path length",
			options: &FileSummaryRenderOptions{
				MaxPathLength: 3,
			},
			suite: &models.TestSuite{
				FilePath:     "test.go",
				TestCount:    1,
				PassedCount:  1,
				FailedCount:  0,
				SkippedCount: 0,
			},
			test: func(t *testing.T, renderer *FileSummaryRenderer, result string) {
				if !strings.Contains(result, "...") {
					t.Errorf("Expected truncation with '...', got: %s", result)
				}
			},
		},
		{
			name: "max path length of 1",
			options: &FileSummaryRenderOptions{
				MaxPathLength: 1,
			},
			suite: &models.TestSuite{
				FilePath:     "test.go",
				TestCount:    1,
				PassedCount:  1,
				FailedCount:  0,
				SkippedCount: 0,
			},
			test: func(t *testing.T, renderer *FileSummaryRenderer, result string) {
				if !strings.Contains(result, ".") {
					t.Errorf("Expected single dot for max length 1, got: %s", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewFileSummaryRenderer(&Config{}, tt.options)
			result := renderer.RenderFileSummary(tt.suite)
			tt.test(t, renderer, result)
		})
	}
}
