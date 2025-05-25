package display

import (
	"bytes"
	"strings"
	"testing"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

func TestFailureRenderer_Creation(t *testing.T) {
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	errorFormatter := NewErrorFormatterWithDefaults(&buf, formatter)

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
		t.Errorf("Expected width 80, got %d", renderer.width)
	}
}

func TestFailureRenderer_WithDefaults(t *testing.T) {
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

func TestRenderFailedTestsHeader_WithFailures(t *testing.T) {
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false) // No colors for easier testing
	icons := colors.NewIconProvider(false)
	errorFormatter := NewErrorFormatterWithDefaults(&buf, formatter)
	renderer := NewFailureRenderer(&buf, formatter, icons, errorFormatter, 80)

	err := renderer.RenderFailedTestsHeader(3)
	if err != nil {
		t.Fatalf("RenderFailedTestsHeader failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Failed Tests 3") {
		t.Errorf("Expected 'Failed Tests 3' in output, got: %s", output)
	}
	if !strings.Contains(output, "─") {
		t.Error("Expected separator characters in output")
	}
}

func TestRenderFailedTestsHeader_NoFailures(t *testing.T) {
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	errorFormatter := NewErrorFormatterWithDefaults(&buf, formatter)
	renderer := NewFailureRenderer(&buf, formatter, icons, errorFormatter, 80)

	err := renderer.RenderFailedTestsHeader(0)
	if err != nil {
		t.Fatalf("RenderFailedTestsHeader failed: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("Expected no output for 0 failures, got: %s", output)
	}
}

func TestRenderFailedTest_FailedTest(t *testing.T) {
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	errorFormatter := NewErrorFormatterWithDefaults(&buf, formatter)
	renderer := NewFailureRenderer(&buf, formatter, icons, errorFormatter, 80)

	failedTest := &models.TestResult{
		Name:   "TestExample",
		Status: models.StatusFailed,
		Error: &models.TestError{
			Type:    "AssertionError",
			Message: "Expected 5, got 10",
		},
	}

	err := renderer.RenderFailedTest(failedTest)
	if err != nil {
		t.Fatalf("RenderFailedTest failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "FAIL") {
		t.Error("Expected 'FAIL' badge in output")
	}
	if !strings.Contains(output, "TestExample") {
		t.Error("Expected test name in output")
	}
	if !strings.Contains(output, "AssertionError") {
		t.Error("Expected error type in output")
	}
	if !strings.Contains(output, "Expected 5, got 10") {
		t.Error("Expected error message in output")
	}
}

func TestRenderFailedTest_PassedTest(t *testing.T) {
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	errorFormatter := NewErrorFormatterWithDefaults(&buf, formatter)
	renderer := NewFailureRenderer(&buf, formatter, icons, errorFormatter, 80)

	passedTest := &models.TestResult{
		Name:   "TestPassed",
		Status: models.StatusPassed,
	}

	err := renderer.RenderFailedTest(passedTest)
	if err != nil {
		t.Fatalf("RenderFailedTest failed: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("Expected no output for passed test, got: %s", output)
	}
}

func TestRenderFailedTest_WithLocation(t *testing.T) {
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	errorFormatter := NewErrorFormatterWithDefaults(&buf, formatter)
	renderer := NewFailureRenderer(&buf, formatter, icons, errorFormatter, 80)

	failedTest := &models.TestResult{
		Name:   "TestWithLocation",
		Status: models.StatusFailed,
		Error: &models.TestError{
			Type:       "TypeError",
			Message:    "Function not found",
			SourceFile: "test.go",
			SourceLine: 42,
		},
	}

	err := renderer.RenderFailedTest(failedTest)
	if err != nil {
		t.Fatalf("RenderFailedTest failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test.go") {
		t.Error("Expected source file in output")
	}
	if !strings.Contains(output, "42") {
		t.Error("Expected line number in output")
	}
}

func TestRenderFailedTests_MultipleTests(t *testing.T) {
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	errorFormatter := NewErrorFormatterWithDefaults(&buf, formatter)
	renderer := NewFailureRenderer(&buf, formatter, icons, errorFormatter, 80)

	tests := []*models.TestResult{
		{
			Name:   "TestPassed",
			Status: models.StatusPassed,
		},
		{
			Name:   "TestFailed1",
			Status: models.StatusFailed,
			Error: &models.TestError{
				Type:    "Error",
				Message: "First failure",
			},
		},
		{
			Name:   "TestFailed2",
			Status: models.StatusFailed,
			Error: &models.TestError{
				Type:    "Error",
				Message: "Second failure",
			},
		},
	}

	err := renderer.RenderFailedTests(tests)
	if err != nil {
		t.Fatalf("RenderFailedTests failed: %v", err)
	}

	output := buf.String()

	// Should show header for 2 failed tests
	if !strings.Contains(output, "Failed Tests 2") {
		t.Error("Expected header for 2 failed tests")
	}

	// Should show both failed tests
	if !strings.Contains(output, "TestFailed1") {
		t.Error("Expected first failed test in output")
	}
	if !strings.Contains(output, "TestFailed2") {
		t.Error("Expected second failed test in output")
	}
	if !strings.Contains(output, "First failure") {
		t.Error("Expected first error message in output")
	}
	if !strings.Contains(output, "Second failure") {
		t.Error("Expected second error message in output")
	}

	// Should not show passed test
	if strings.Contains(output, "TestPassed") {
		t.Error("Should not show passed test in failed tests output")
	}
}

func TestRenderFailedTests_NoFailures(t *testing.T) {
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	errorFormatter := NewErrorFormatterWithDefaults(&buf, formatter)
	renderer := NewFailureRenderer(&buf, formatter, icons, errorFormatter, 80)

	tests := []*models.TestResult{
		{
			Name:   "TestPassed1",
			Status: models.StatusPassed,
		},
		{
			Name:   "TestPassed2",
			Status: models.StatusPassed,
		},
	}

	err := renderer.RenderFailedTests(tests)
	if err != nil {
		t.Fatalf("RenderFailedTests failed: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("Expected no output when no tests failed, got: %s", output)
	}
}

func TestSetAndGetWidth(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewFailureRendererWithDefaults(&buf)

	// Test default width
	defaultWidth := renderer.GetTerminalWidth()
	if defaultWidth != 80 {
		t.Errorf("Expected default width 80, got %d", defaultWidth)
	}

	// Test setting width
	newWidth := 120
	renderer.SetWidth(newWidth)
	actualWidth := renderer.GetTerminalWidth()
	if actualWidth != newWidth {
		t.Errorf("Expected width %d, got %d", newWidth, actualWidth)
	}
}
