package display

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// Mock formatter for testing
type mockFormatter struct {
	enabled bool
}

func (m *mockFormatter) Red(text string) string     { return "[RED]" + text + "[/RED]" }
func (m *mockFormatter) Green(text string) string   { return "[GREEN]" + text + "[/GREEN]" }
func (m *mockFormatter) Yellow(text string) string  { return "[YELLOW]" + text + "[/YELLOW]" }
func (m *mockFormatter) Blue(text string) string    { return "[BLUE]" + text + "[/BLUE]" }
func (m *mockFormatter) Gray(text string) string    { return "[GRAY]" + text + "[/GRAY]" }
func (m *mockFormatter) Dim(text string) string     { return "[DIM]" + text + "[/DIM]" }
func (m *mockFormatter) Bold(text string) string    { return "[BOLD]" + text + "[/BOLD]" }
func (m *mockFormatter) Magenta(text string) string { return "[MAGENTA]" + text + "[/MAGENTA]" }
func (m *mockFormatter) Cyan(text string) string    { return "[CYAN]" + text + "[/CYAN]" }
func (m *mockFormatter) White(text string) string   { return "[WHITE]" + text + "[/WHITE]" }
func (m *mockFormatter) BgRed(text string) string   { return "[BGRED]" + text + "[/BGRED]" }
func (m *mockFormatter) Colorize(text, color string) string {
	return "[" + color + "]" + text + "[/" + color + "]"
}
func (m *mockFormatter) IsEnabled() bool { return m.enabled }

// Mock icon provider for testing
type mockIconProvider struct {
	unicode bool
}

func (m *mockIconProvider) CheckMark() string              { return "✓" }
func (m *mockIconProvider) Cross() string                  { return "✗" }
func (m *mockIconProvider) Skipped() string                { return "○" }
func (m *mockIconProvider) Running() string                { return "~" }
func (m *mockIconProvider) GetIcon(iconType string) string { return "[" + iconType + "]" }
func (m *mockIconProvider) SupportsUnicode() bool          { return m.unicode }

func TestNewTestRenderer(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}

	renderer := NewTestRenderer(&buf, formatter, icons)

	if renderer == nil {
		t.Fatal("NewTestRenderer returned nil")
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
	if renderer.indentLevel != 0 {
		t.Error("IndentLevel should be 0 by default")
	}
}

func TestNewTestRendererWithDefaults(t *testing.T) {
	var buf bytes.Buffer

	renderer := NewTestRendererWithDefaults(&buf)

	if renderer == nil {
		t.Fatal("NewTestRendererWithDefaults returned nil")
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
}

func TestSetIndentLevel(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewTestRenderer(&buf, &mockFormatter{}, &mockIconProvider{})

	renderer.SetIndentLevel(2)
	if renderer.indentLevel != 2 {
		t.Errorf("Expected indent level 2, got %d", renderer.indentLevel)
	}

	indent := renderer.GetCurrentIndent()
	expected := "    " // 2 levels * 2 spaces each
	if indent != expected {
		t.Errorf("Expected indent '%s', got '%s'", expected, indent)
	}
}

func TestRenderTestResult_PassedTest(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewTestRenderer(&buf, formatter, icons)

	result := &models.LegacyTestResult{
		Name:     "TestExample",
		Status:   models.StatusPassed,
		Duration: 10 * time.Millisecond,
	}

	err := renderer.RenderTestResult(result, 0)
	if err != nil {
		t.Fatalf("RenderTestResult failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[GREEN]✓[/GREEN]") {
		t.Error("Expected green checkmark for passed test")
	}
	if !strings.Contains(output, "TestExample") {
		t.Error("Expected test name in output")
	}
	if !strings.Contains(output, "[DIM]10ms[/DIM]") {
		t.Error("Expected duration in output")
	}
}

func TestRenderTestResult_FailedTest(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewTestRenderer(&buf, formatter, icons)

	result := &models.LegacyTestResult{
		Name:     "TestFailed",
		Status:   models.StatusFailed,
		Duration: 15 * time.Millisecond,
		Error: &models.LegacyTestError{
			Message: "assertion failed",
			Type:    "AssertionError",
		},
	}

	err := renderer.RenderTestResult(result, 0)
	if err != nil {
		t.Fatalf("RenderTestResult failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[RED]✗[/RED]") {
		t.Error("Expected red cross for failed test")
	}
	if !strings.Contains(output, "TestFailed") {
		t.Error("Expected test name in output")
	}
	if !strings.Contains(output, "→ [RED]assertion failed[/RED]") {
		t.Error("Expected error message in output")
	}
}

func TestRenderTestResult_SkippedTest(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewTestRenderer(&buf, formatter, icons)

	result := &models.LegacyTestResult{
		Name:     "TestSkipped",
		Status:   models.StatusSkipped,
		Duration: 0,
	}

	err := renderer.RenderTestResult(result, 0)
	if err != nil {
		t.Fatalf("RenderTestResult failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[YELLOW]○[/YELLOW]") {
		t.Error("Expected yellow skip icon for skipped test")
	}
	if !strings.Contains(output, "TestSkipped") {
		t.Error("Expected test name in output")
	}
}

func TestRenderTestResult_RunningTest(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewTestRenderer(&buf, formatter, icons)

	result := &models.LegacyTestResult{
		Name:     "TestRunning",
		Status:   models.StatusRunning,
		Duration: 5 * time.Millisecond,
	}

	err := renderer.RenderTestResult(result, 0)
	if err != nil {
		t.Fatalf("RenderTestResult failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[BLUE]~[/BLUE]") {
		t.Error("Expected blue running icon for running test")
	}
	if !strings.Contains(output, "TestRunning") {
		t.Error("Expected test name in output")
	}
}

func TestRenderTestResult_WithSubtests(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewTestRenderer(&buf, formatter, icons)

	subtest := &models.LegacyTestResult{
		Name:     "TestParent/SubTest",
		Status:   models.StatusPassed,
		Duration: 5 * time.Millisecond,
		Parent:   "TestParent",
	}

	result := &models.LegacyTestResult{
		Name:     "TestParent",
		Status:   models.StatusPassed,
		Duration: 10 * time.Millisecond,
		Subtests: []*models.LegacyTestResult{subtest},
	}

	err := renderer.RenderTestResult(result, 0)
	if err != nil {
		t.Fatalf("RenderTestResult failed: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 lines for parent and subtest, got %d", len(lines))
	}

	// Parent test should not be indented
	if !strings.HasPrefix(lines[0], "[GREEN]✓[/GREEN] TestParent") {
		t.Error("Parent test should not be indented")
	}

	// Subtest should be indented
	if !strings.HasPrefix(lines[1], "  [GREEN]✓[/GREEN] SubTest") {
		t.Error("Subtest should be indented and show only the subtest name")
	}
}

func TestRenderTestResult_AssertionError(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewTestRenderer(&buf, formatter, icons)

	result := &models.LegacyTestResult{
		Name:     "TestAssertion",
		Status:   models.StatusFailed,
		Duration: 10 * time.Millisecond,
		Error: &models.LegacyTestError{
			Message:  "values not equal",
			Type:     "AssertionError",
			Expected: "42",
			Actual:   "24",
		},
	}

	err := renderer.RenderTestResult(result, 0)
	if err != nil {
		t.Fatalf("RenderTestResult failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[DIM]Expected:[/DIM] [GREEN]42[/GREEN]") {
		t.Error("Expected value should be formatted in green")
	}
	if !strings.Contains(output, "[DIM]Received:[/DIM] [RED]24[/RED]") {
		t.Error("Actual value should be formatted in red")
	}
}

func TestRenderTestResult_WithLocation(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewTestRenderer(&buf, formatter, icons)

	result := &models.LegacyTestResult{
		Name:     "TestWithLocation",
		Status:   models.StatusFailed,
		Duration: 10 * time.Millisecond,
		Error: &models.LegacyTestError{
			Message: "test failed",
			Type:    "Error",
			Location: &models.SourceLocation{
				File: "test.go",
				Line: 42,
			},
		},
	}

	err := renderer.RenderTestResult(result, 0)
	if err != nil {
		t.Fatalf("RenderTestResult failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[DIM]at[/DIM] test.go:42") {
		t.Error("Expected location information in output")
	}
}

func TestRenderTestResult_PanicWithStack(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewTestRenderer(&buf, formatter, icons)

	stackTrace := "goroutine 1 [running]:\npackage.function()\n\tfile.go:10 +0x123\nmain.main()\n\tmain.go:15 +0x456\nmore lines here"

	result := &models.LegacyTestResult{
		Name:     "TestPanic",
		Status:   models.StatusFailed,
		Duration: 5 * time.Millisecond,
		Error: &models.LegacyTestError{
			Message: "panic occurred",
			Type:    "Panic",
			Stack:   stackTrace,
		},
	}

	err := renderer.RenderTestResult(result, 0)
	if err != nil {
		t.Fatalf("RenderTestResult failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[DIM]Stack trace:[/DIM]") {
		t.Error("Expected stack trace header")
	}
	if !strings.Contains(output, "[DIM]...[/DIM]") {
		t.Error("Expected ellipsis for truncated stack trace")
	}
}

func TestRenderTestResult_NilResult(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewTestRenderer(&buf, formatter, icons)

	err := renderer.RenderTestResult(nil, 0)
	if err == nil {
		t.Error("Expected error for nil test result")
	}
	if !strings.Contains(err.Error(), "test result cannot be nil") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestRenderTestResults(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewTestRenderer(&buf, formatter, icons)

	results := []*models.LegacyTestResult{
		{
			Name:     "Test1",
			Status:   models.StatusPassed,
			Duration: 5 * time.Millisecond,
		},
		{
			Name:     "Test2",
			Status:   models.StatusFailed,
			Duration: 10 * time.Millisecond,
			Error:    &models.LegacyTestError{Message: "failed"},
		},
	}

	err := renderer.RenderTestResults(results)
	if err != nil {
		t.Fatalf("RenderTestResults failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Test1") {
		t.Error("Expected Test1 in output")
	}
	if !strings.Contains(output, "Test2") {
		t.Error("Expected Test2 in output")
	}
}

func TestRenderTestSummaryLine(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewTestRenderer(&buf, formatter, icons)

	suite := &models.TestSuite{
		FilePath:     "test/example_test.go",
		TestCount:    5,
		PassedCount:  4,
		FailedCount:  1,
		SkippedCount: 0,
		Duration:     100 * time.Millisecond,
	}

	err := renderer.RenderTestSummaryLine(suite)
	if err != nil {
		t.Fatalf("RenderTestSummaryLine failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[RED]✗[/RED]") {
		t.Error("Expected red cross for suite with failures")
	}
	if !strings.Contains(output, "[BOLD]test/example_test.go[/BOLD]") {
		t.Error("Expected bold file path")
	}
	if !strings.Contains(output, "(4/5 passed)") {
		t.Error("Expected test counts")
	}
	if !strings.Contains(output, "[DIM]100ms[/DIM]") {
		t.Error("Expected duration")
	}
}

func TestRenderTestSummaryLine_AllPassed(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewTestRenderer(&buf, formatter, icons)

	suite := &models.TestSuite{
		FilePath:     "test/success_test.go",
		TestCount:    3,
		PassedCount:  3,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     50 * time.Millisecond,
	}

	err := renderer.RenderTestSummaryLine(suite)
	if err != nil {
		t.Fatalf("RenderTestSummaryLine failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[GREEN]✓[/GREEN]") {
		t.Error("Expected green checkmark for suite with all tests passed")
	}
}

func TestRenderTestSummaryLine_NilSuite(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewTestRenderer(&buf, formatter, icons)

	err := renderer.RenderTestSummaryLine(nil)
	if err == nil {
		t.Error("Expected error for nil test suite")
	}
	if !strings.Contains(err.Error(), "test suite cannot be nil") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestFormatTestName(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewTestRenderer(&buf, formatter, icons)

	tests := []struct {
		name     string
		result   *models.LegacyTestResult
		expected string
	}{
		{
			name: "regular test",
			result: &models.LegacyTestResult{
				Name:   "TestExample",
				Parent: "",
			},
			expected: "TestExample",
		},
		{
			name: "subtest with parent",
			result: &models.LegacyTestResult{
				Name:   "TestParent/SubTest",
				Parent: "TestParent",
			},
			expected: "SubTest",
		},
		{
			name: "subtest without slash",
			result: &models.LegacyTestResult{
				Name:   "TestSubTest",
				Parent: "TestParent",
			},
			expected: "TestSubTest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := renderer.formatTestName(tt.result)
			if actual != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

func TestIndentationLevels(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewTestRenderer(&buf, formatter, icons)

	// Set base indent level
	renderer.SetIndentLevel(1)

	result := &models.LegacyTestResult{
		Name:     "TestIndented",
		Status:   models.StatusPassed,
		Duration: 5 * time.Millisecond,
	}

	err := renderer.RenderTestResult(result, 2) // Additional 2 levels
	if err != nil {
		t.Fatalf("RenderTestResult failed: %v", err)
	}

	output := buf.String()
	// Should have 3 levels total (1 base + 2 additional) = 6 spaces
	expectedPrefix := "      [GREEN]✓[/GREEN]"
	if !strings.HasPrefix(output, expectedPrefix) {
		t.Errorf("Expected output to start with %q, got %q", expectedPrefix, output[:len(expectedPrefix)])
	}
}
