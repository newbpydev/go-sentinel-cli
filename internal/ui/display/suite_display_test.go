package display

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

func TestNewSuiteRenderer(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	width := 80

	renderer := NewSuiteRenderer(&buf, formatter, icons, width)

	if renderer == nil {
		t.Fatal("NewSuiteRenderer returned nil")
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
	if renderer.width != width {
		t.Error("Width not set correctly")
	}
	if renderer.autoCollapse {
		t.Error("AutoCollapse should be false by default")
	}
	if renderer.header == nil {
		t.Error("Header renderer not initialized")
	}
	if renderer.test == nil {
		t.Error("Test renderer not initialized")
	}
}

func TestNewSuiteRendererWithDefaults(t *testing.T) {
	var buf bytes.Buffer
	width := 100

	renderer := NewSuiteRendererWithDefaults(&buf, width)

	if renderer == nil {
		t.Fatal("NewSuiteRendererWithDefaults returned nil")
	}
	if renderer.writer != &buf {
		t.Error("Writer not set correctly")
	}
	if renderer.width != width {
		t.Error("Width not set correctly")
	}
	if renderer.formatter == nil {
		t.Error("Default formatter not created")
	}
	if renderer.icons == nil {
		t.Error("Default icons not created")
	}
}

func TestSetAutoCollapse(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewSuiteRenderer(&buf, &mockFormatter{}, &mockIconProvider{}, 80)

	if renderer.autoCollapse {
		t.Error("AutoCollapse should start as false")
	}

	renderer.SetAutoCollapse(true)
	if !renderer.autoCollapse {
		t.Error("AutoCollapse should be true after setting")
	}

	renderer.SetAutoCollapse(false)
	if renderer.autoCollapse {
		t.Error("AutoCollapse should be false after setting")
	}
}

func TestSetWidth(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewSuiteRenderer(&buf, &mockFormatter{}, &mockIconProvider{}, 80)

	newWidth := 120
	renderer.SetWidth(newWidth)

	if renderer.width != newWidth {
		t.Errorf("Expected width %d, got %d", newWidth, renderer.width)
	}
}

func TestRenderSuite_PassedTests(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewSuiteRenderer(&buf, formatter, icons, 80)

	suite := &models.TestSuite{
		FilePath:     "test/example_test.go",
		TestCount:    3,
		PassedCount:  3,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     100 * time.Millisecond,
		MemoryUsage:  5 * 1024 * 1024, // 5 MB
		Tests: []*models.LegacyTestResult{
			{
				Name:     "Test1",
				Status:   models.StatusPassed,
				Duration: 30 * time.Millisecond,
			},
			{
				Name:     "Test2",
				Status:   models.StatusPassed,
				Duration: 40 * time.Millisecond,
			},
			{
				Name:     "Test3",
				Status:   models.StatusPassed,
				Duration: 30 * time.Millisecond,
			},
		},
	}

	err := renderer.RenderSuite(suite, false)
	if err != nil {
		t.Fatalf("RenderSuite failed: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "test/example_test.go") && !strings.Contains(output, "example_test.go") {
		t.Error("Expected file path in output")
	}
	if !strings.Contains(output, "(3 tests)") {
		t.Error("Expected test count in output")
	}
	if !strings.Contains(output, "100ms") {
		t.Error("Expected duration in output")
	}
	if !strings.Contains(output, "5 MB heap used") && !strings.Contains(output, "5.0 MB heap used") {
		t.Error("Expected memory usage in output")
	}
	if !strings.Contains(output, "Test1") {
		t.Error("Expected individual tests to be rendered")
	}
}

func TestRenderSuite_WithFailures(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewSuiteRenderer(&buf, formatter, icons, 80)

	suite := &models.TestSuite{
		FilePath:     "test/failed_test.go",
		TestCount:    2,
		PassedCount:  1,
		FailedCount:  1,
		SkippedCount: 0,
		Duration:     150 * time.Millisecond,
		MemoryUsage:  3 * 1024 * 1024,
		Tests: []*models.LegacyTestResult{
			{
				Name:     "TestPassed",
				Status:   models.StatusPassed,
				Duration: 50 * time.Millisecond,
			},
			{
				Name:     "TestFailed",
				Status:   models.StatusFailed,
				Duration: 100 * time.Millisecond,
				Error:    &models.LegacyTestError{Message: "assertion failed"},
			},
		},
	}

	err := renderer.RenderSuite(suite, false)
	if err != nil {
		t.Fatalf("RenderSuite failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[RED]1 failed[/RED]") {
		t.Error("Expected red failed count in output")
	}
	if !strings.Contains(output, "TestFailed") {
		t.Error("Expected failed test to be rendered")
	}
}

func TestRenderSuite_AllSkipped(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewSuiteRenderer(&buf, formatter, icons, 80)

	suite := &models.TestSuite{
		FilePath:     "test/skipped_test.go",
		TestCount:    2,
		PassedCount:  0,
		FailedCount:  0,
		SkippedCount: 2,
		Duration:     0,
		MemoryUsage:  0,
		Tests: []*models.LegacyTestResult{
			{
				Name:     "TestSkipped1",
				Status:   models.StatusSkipped,
				Duration: 0,
			},
			{
				Name:     "TestSkipped2",
				Status:   models.StatusSkipped,
				Duration: 0,
			},
		},
	}

	err := renderer.RenderSuite(suite, false)
	if err != nil {
		t.Fatalf("RenderSuite failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[YELLOW]2 skipped[/YELLOW]") {
		t.Error("Expected yellow skipped count in output")
	}
}

func TestRenderSuite_MixedWithSkipped(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewSuiteRenderer(&buf, formatter, icons, 80)

	suite := &models.TestSuite{
		FilePath:     "test/mixed_test.go",
		TestCount:    3,
		PassedCount:  2,
		FailedCount:  0,
		SkippedCount: 1,
		Duration:     80 * time.Millisecond,
		MemoryUsage:  2 * 1024 * 1024,
		Tests: []*models.LegacyTestResult{
			{Name: "Test1", Status: models.StatusPassed, Duration: 30 * time.Millisecond},
			{Name: "Test2", Status: models.StatusPassed, Duration: 50 * time.Millisecond},
			{Name: "Test3", Status: models.StatusSkipped, Duration: 0},
		},
	}

	err := renderer.RenderSuite(suite, false)
	if err != nil {
		t.Fatalf("RenderSuite failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "(3 tests") && !strings.Contains(output, "[YELLOW]1 skipped[/YELLOW]") {
		t.Error("Expected test count with skipped information")
	}
}

func TestRenderSuite_Collapsed(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewSuiteRenderer(&buf, formatter, icons, 80)

	suite := &models.TestSuite{
		FilePath:     "test/passed_test.go",
		TestCount:    2,
		PassedCount:  2,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     50 * time.Millisecond,
		MemoryUsage:  1 * 1024 * 1024,
		Tests: []*models.LegacyTestResult{
			{Name: "Test1", Status: models.StatusPassed, Duration: 25 * time.Millisecond},
			{Name: "Test2", Status: models.StatusPassed, Duration: 25 * time.Millisecond},
		},
	}

	// Render with auto-collapse enabled for passed suites
	err := renderer.RenderSuite(suite, true)
	if err != nil {
		t.Fatalf("RenderSuite failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[GREEN]Suite passed (2 tests)[/GREEN]") {
		t.Error("Expected collapsed suite summary")
	}
	if strings.Contains(output, "Test1") || strings.Contains(output, "Test2") {
		t.Error("Individual tests should not be shown when collapsed")
	}
}

func TestRenderSuite_AutoCollapseInstance(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewSuiteRenderer(&buf, formatter, icons, 80)

	// Set auto-collapse on the instance
	renderer.SetAutoCollapse(true)

	suite := &models.TestSuite{
		FilePath:     "test/auto_collapse_test.go",
		TestCount:    1,
		PassedCount:  1,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     25 * time.Millisecond,
		MemoryUsage:  512 * 1024,
		Tests: []*models.LegacyTestResult{
			{Name: "TestAuto", Status: models.StatusPassed, Duration: 25 * time.Millisecond},
		},
	}

	err := renderer.RenderSuite(suite, false) // Don't pass auto-collapse parameter
	if err != nil {
		t.Fatalf("RenderSuite failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[GREEN]Suite passed (1 tests)[/GREEN]") {
		t.Error("Expected suite to be auto-collapsed based on instance setting")
	}
}

func TestRenderSuite_NilSuite(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewSuiteRenderer(&buf, formatter, icons, 80)

	err := renderer.RenderSuite(nil, false)
	if err == nil {
		t.Error("Expected error for nil suite")
	}
	if !strings.Contains(err.Error(), "test suite cannot be nil") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestRenderSuiteSummary(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewSuiteRenderer(&buf, formatter, icons, 80)

	suite := &models.TestSuite{
		FilePath:     "test/summary_test.go",
		TestCount:    5,
		PassedCount:  4,
		FailedCount:  1,
		SkippedCount: 0,
		Duration:     200 * time.Millisecond,
		MemoryUsage:  10 * 1024 * 1024,
	}

	err := renderer.RenderSuiteSummary(suite)
	if err != nil {
		t.Fatalf("RenderSuiteSummary failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test/summary_test.go") && !strings.Contains(output, "summary_test.go") {
		t.Error("Expected file path in summary")
	}
	if !strings.Contains(output, "200ms") {
		t.Error("Expected duration in summary")
	}
	if !strings.Contains(output, "10 MB heap used") && !strings.Contains(output, "10.0 MB heap used") {
		t.Error("Expected memory usage in summary")
	}
}

func TestRenderSuiteSummary_NilSuite(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewSuiteRenderer(&buf, &mockFormatter{}, &mockIconProvider{}, 80)

	err := renderer.RenderSuiteSummary(nil)
	if err == nil {
		t.Error("Expected error for nil suite")
	}
	if !strings.Contains(err.Error(), "test suite cannot be nil") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestRenderMultipleSuites(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewSuiteRenderer(&buf, formatter, icons, 80)

	suites := []*models.TestSuite{
		{
			FilePath:     "test/suite1_test.go",
			TestCount:    1,
			PassedCount:  1,
			FailedCount:  0,
			SkippedCount: 0,
			Duration:     30 * time.Millisecond,
			MemoryUsage:  1024 * 1024,
			Tests: []*models.LegacyTestResult{
				{Name: "Test1", Status: models.StatusPassed, Duration: 30 * time.Millisecond},
			},
		},
		{
			FilePath:     "test/suite2_test.go",
			TestCount:    1,
			PassedCount:  0,
			FailedCount:  1,
			SkippedCount: 0,
			Duration:     50 * time.Millisecond,
			MemoryUsage:  2 * 1024 * 1024,
			Tests: []*models.LegacyTestResult{
				{
					Name:     "Test2",
					Status:   models.StatusFailed,
					Duration: 50 * time.Millisecond,
					Error:    &models.LegacyTestError{Message: "failed"},
				},
			},
		},
	}

	err := renderer.RenderMultipleSuites(suites, false)
	if err != nil {
		t.Fatalf("RenderMultipleSuites failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "suite1_test.go") {
		t.Error("Expected first suite in output")
	}
	if !strings.Contains(output, "suite2_test.go") {
		t.Error("Expected second suite in output")
	}

	// Check for spacing between suites
	lines := strings.Split(strings.TrimSpace(output), "\n")
	foundEmptyLine := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			foundEmptyLine = true
			break
		}
	}
	if !foundEmptyLine {
		t.Error("Expected empty line between suites")
	}
}

func TestRenderSuiteWithOptions(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewSuiteRenderer(&buf, formatter, icons, 80)

	suite := &models.TestSuite{
		FilePath:     "test/options_test.go",
		TestCount:    1,
		PassedCount:  1,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     20 * time.Millisecond,
		MemoryUsage:  512 * 1024,
		Tests: []*models.LegacyTestResult{
			{Name: "TestOptions", Status: models.StatusPassed, Duration: 20 * time.Millisecond},
		},
	}

	// Test with auto-collapse option
	autoCollapse := true
	options := SuiteDisplayOptions{
		AutoCollapse: &autoCollapse,
		Width:        120,
	}

	err := renderer.RenderSuiteWithOptions(suite, options)
	if err != nil {
		t.Fatalf("RenderSuiteWithOptions failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[GREEN]Suite passed (1 tests)[/GREEN]") {
		t.Error("Expected suite to be collapsed based on options")
	}

	// Verify that original settings are restored
	if renderer.width != 80 {
		t.Error("Original width should be restored")
	}
	if renderer.autoCollapse {
		t.Error("Original auto-collapse setting should be restored")
	}
}

func TestRenderSuiteWithOptions_NilSuite(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewSuiteRenderer(&buf, &mockFormatter{}, &mockIconProvider{}, 80)

	options := SuiteDisplayOptions{}
	err := renderer.RenderSuiteWithOptions(nil, options)
	if err == nil {
		t.Error("Expected error for nil suite")
	}
	if !strings.Contains(err.Error(), "test suite cannot be nil") {
		t.Errorf("Expected specific error message, got: %v", err)
	}
}

func TestFormatTestCounts(t *testing.T) {
	var buf bytes.Buffer
	formatter := &mockFormatter{enabled: true}
	icons := &mockIconProvider{unicode: true}
	renderer := NewSuiteRenderer(&buf, formatter, icons, 80)

	tests := []struct {
		name     string
		suite    *models.TestSuite
		expected string
	}{
		{
			name: "all passed",
			suite: &models.TestSuite{
				TestCount:    3,
				PassedCount:  3,
				FailedCount:  0,
				SkippedCount: 0,
			},
			expected: "(3 tests)",
		},
		{
			name: "with failures",
			suite: &models.TestSuite{
				TestCount:    5,
				PassedCount:  3,
				FailedCount:  2,
				SkippedCount: 0,
			},
			expected: "(5 tests | [RED]2 failed[/RED])",
		},
		{
			name: "all skipped",
			suite: &models.TestSuite{
				TestCount:    2,
				PassedCount:  0,
				FailedCount:  0,
				SkippedCount: 2,
			},
			expected: "(2 tests | [YELLOW]2 skipped[/YELLOW])",
		},
		{
			name: "passed with some skipped",
			suite: &models.TestSuite{
				TestCount:    4,
				PassedCount:  3,
				FailedCount:  0,
				SkippedCount: 1,
			},
			expected: "(4 tests | [YELLOW]1 skipped[/YELLOW])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := renderer.formatTestCounts(tt.suite)
			if actual != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, actual)
			}
		})
	}
}
