package display

import (
	"bytes"
	"strings"
	"testing"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

func TestNewFailureRenderer(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	errorFormatter := &mockErrorFormatter{}

	renderer := NewFailureRenderer(&buf, formatter, icons, errorFormatter, 80)

	if renderer == nil {
		t.Fatal("NewFailureRenderer returned nil")
	}
	if renderer.writer != &buf {
		t.Error("Writer not set correctly")
	}
	if renderer.formatter != formatter {
		t.Error("Formatter not set correctly")
	}
	if renderer.icons != icons {
		t.Error("Icons not set correctly")
	}
	if renderer.errorFormatter != errorFormatter {
		t.Error("ErrorFormatter not set correctly")
	}
	if renderer.width != 80 {
		t.Error("Width not set correctly")
	}
}

func TestNewFailureRendererWithDefaults(t *testing.T) {
	var buf bytes.Buffer

	renderer := NewFailureRendererWithDefaults(&buf)

	if renderer == nil {
		t.Fatal("NewFailureRendererWithDefaults returned nil")
	}
	if renderer.writer != &buf {
		t.Error("Writer not set correctly")
	}
	if renderer.formatter == nil {
		t.Error("Default formatter not created")
	}
	if renderer.icons == nil {
		t.Error("Default icons not created")
	}
	if renderer.errorFormatter == nil {
		t.Error("Default error formatter not created")
	}
}

func TestFailureRenderer_SetWidth(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewFailureRenderer(&buf, &mockFormatter{}, &mockIconProvider{}, &mockErrorFormatter{}, 80)

	renderer.SetWidth(120)
	if renderer.width != 120 {
		t.Errorf("Expected width 120, got %d", renderer.width)
	}

	if renderer.GetTerminalWidth() != 120 {
		t.Errorf("Expected terminal width 120, got %d", renderer.GetTerminalWidth())
	}
}

func TestFailureRenderer_GetTerminalWidth(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewFailureRenderer(&buf, &mockFormatter{}, &mockIconProvider{}, &mockErrorFormatter{}, 0)

	// Should return default when width is 0
	width := renderer.GetTerminalWidth()
	if width != 80 {
		t.Errorf("Expected default width 80, got %d", width)
	}

	// Should return configured width when set
	renderer.SetWidth(100)
	width = renderer.GetTerminalWidth()
	if width != 100 {
		t.Errorf("Expected configured width 100, got %d", width)
	}
}

func TestRenderFailedTestsHeader(t *testing.T) {
	tests := []struct {
		name      string
		failCount int
		wantText  []string
		wantEmpty bool
	}{
		{
			name:      "with failed tests",
			failCount: 3,
			wantText:  []string{"Failed Tests 3", "[RED]", "[BGRED]", "[WHITE]"},
		},
		{
			name:      "no failed tests",
			failCount: 0,
			wantEmpty: true,
		},
		{
			name:      "negative count",
			failCount: -1,
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := &mockFormatter{enabled: true}
			icons := &mockIconProvider{unicode: true}
			errorFormatter := &mockErrorFormatter{}
			renderer := NewFailureRenderer(&buf, formatter, icons, errorFormatter, 80)

			err := renderer.RenderFailedTestsHeader(tt.failCount)
			if err != nil {
				t.Fatalf("RenderFailedTestsHeader failed: %v", err)
			}

			output := buf.String()

			if tt.wantEmpty {
				if output != "" {
					t.Error("Expected empty output for zero/negative fail count")
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

func TestRenderFailedTest(t *testing.T) {
	tests := []struct {
		name     string
		test     *models.TestResult
		wantText []string
		wantErr  bool
	}{
		{
			name: "failed test with error",
			test: &models.TestResult{
				Name:   "TestExample",
				Status: models.StatusFailed,
				Error: &models.TestError{
					Message: "assertion failed",
					Type:    "AssertionError",
				},
			},
			wantText: []string{"[BGRED]", "[WHITE] FAIL [/WHITE]", "[/BGRED]", "TestExample", "[RED]AssertionError[/RED]", "[RED]assertion failed[/RED]"},
		},
		{
			name: "failed test with source location",
			test: &models.TestResult{
				Name:   "TestWithLocation",
				Status: models.StatusFailed,
				Error: &models.TestError{
					Message:    "test failed",
					Type:       "Error",
					SourceFile: "test.go",
					SourceLine: 42,
				},
			},
			wantText: []string{"TestWithLocation", "[RED]Error[/RED]", "[RED]test failed[/RED]", "clickable_location_called"},
		},
		{
			name: "passed test",
			test: &models.TestResult{
				Name:   "TestPassed",
				Status: models.StatusPassed,
			},
			wantText: []string{}, // Should not render anything
		},
		{
			name: "failed test without error",
			test: &models.TestResult{
				Name:   "TestFailedNoError",
				Status: models.StatusFailed,
				Error:  nil,
			},
			wantText: []string{}, // Should not render anything
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := &mockFormatter{enabled: true}
			icons := &mockIconProvider{unicode: true}
			errorFormatter := &mockErrorFormatter{}
			renderer := NewFailureRenderer(&buf, formatter, icons, errorFormatter, 80)

			err := renderer.RenderFailedTest(tt.test)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderFailedTest error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			output := buf.String()

			for _, want := range tt.wantText {
				if !strings.Contains(output, want) {
					t.Errorf("Expected %q in output, but not found in: %s", want, output)
				}
			}
		})
	}
}

func TestRenderFailedTests(t *testing.T) {
	tests := []struct {
		name     string
		tests    []*models.TestResult
		wantText []string
	}{
		{
			name: "multiple failed tests",
			tests: []*models.TestResult{
				{
					Name:   "Test1",
					Status: models.StatusFailed,
					Error: &models.TestError{
						Message: "error 1",
						Type:    "Error",
					},
				},
				{
					Name:   "Test2",
					Status: models.StatusFailed,
					Error: &models.TestError{
						Message: "error 2",
						Type:    "Error",
					},
				},
			},
			wantText: []string{"Failed Tests 2", "Test1", "Test2", "error 1", "error 2"},
		},
		{
			name: "mixed test results",
			tests: []*models.TestResult{
				{
					Name:   "TestPassed",
					Status: models.StatusPassed,
				},
				{
					Name:   "TestFailed",
					Status: models.StatusFailed,
					Error: &models.TestError{
						Message: "test failed",
						Type:    "Error",
					},
				},
			},
			wantText: []string{"Failed Tests 1", "TestFailed", "test failed"},
		},
		{
			name:     "empty test list",
			tests:    []*models.TestResult{},
			wantText: []string{},
		},
		{
			name: "no failed tests",
			tests: []*models.TestResult{
				{
					Name:   "TestPassed",
					Status: models.StatusPassed,
				},
			},
			wantText: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := &mockFormatter{enabled: true}
			icons := &mockIconProvider{unicode: true}
			errorFormatter := &mockErrorFormatter{}
			renderer := NewFailureRenderer(&buf, formatter, icons, errorFormatter, 80)

			err := renderer.RenderFailedTests(tt.tests)
			if err != nil {
				t.Fatalf("RenderFailedTests failed: %v", err)
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

func TestFormatFailHeader(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	errorFormatter := &mockErrorFormatter{}
	renderer := NewFailureRenderer(&buf, formatter, icons, errorFormatter, 80)

	tests := []struct {
		name     string
		test     *models.TestResult
		expected string
	}{
		{
			name: "normal test",
			test: &models.TestResult{
				Name: "TestExample",
			},
			expected: "[BGRED][WHITE] FAIL [/WHITE][/BGRED] TestExample",
		},
		{
			name: "test with empty name",
			test: &models.TestResult{
				Name: "",
			},
			expected: "[BGRED][WHITE] FAIL [/WHITE][/BGRED] Unknown Test",
		},
		{
			name:     "nil test",
			test:     nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := renderer.formatFailHeader(tt.test)
			if actual != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

func TestRenderTestSeparator(t *testing.T) {
	tests := []struct {
		name       string
		width      int
		testNumber int
		totalTests int
		wantText   []string
	}{
		{
			name:       "wide terminal",
			width:      120,
			testNumber: 2,
			totalTests: 3,
			wantText:   []string{"[GRAY]", "─"},
		},
		{
			name:       "narrow terminal",
			width:      30,
			testNumber: 1,
			totalTests: 2,
			wantText:   []string{"[GRAY]────────────────[/GRAY]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := &mockFormatter{enabled: true}
			icons := &mockIconProvider{unicode: true}
			errorFormatter := &mockErrorFormatter{}
			renderer := NewFailureRenderer(&buf, formatter, icons, errorFormatter, tt.width)

			err := renderer.renderTestSeparator(tt.testNumber, tt.totalTests)
			if err != nil {
				t.Fatalf("renderTestSeparator failed: %v", err)
			}

			output := buf.String()
			for _, want := range tt.wantText {
				if !strings.Contains(output, want) {
					t.Errorf("Expected %q in output, but not found in: %s", want, output)
				}
			}
		})
	}
}

// Mock error formatter for testing
type mockErrorFormatter struct{}

func (m *mockErrorFormatter) FormatClickableLocation(location *models.SourceLocation) string {
	return "clickable_location_called"
}

func (m *mockErrorFormatter) RenderSourceContext(test *models.TestResult) error {
	return nil
}

func (m *mockErrorFormatter) RenderErrorPointer(location *models.SourceLocation, sourceLine string) error {
	return nil
}

func (m *mockErrorFormatter) CalculateErrorPosition(location *models.SourceLocation, sourceLine string) int {
	return 0
}

func (m *mockErrorFormatter) GetTerminalWidth() int {
	return 80
}
