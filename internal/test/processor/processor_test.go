package processor

import (
	"bytes"
	"strings"
	"testing"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// MockColorFormatter for testing
type MockColorFormatter struct{}

func (m *MockColorFormatter) Red(text string) string                 { return text }
func (m *MockColorFormatter) Green(text string) string               { return text }
func (m *MockColorFormatter) Yellow(text string) string              { return text }
func (m *MockColorFormatter) Blue(text string) string                { return text }
func (m *MockColorFormatter) Magenta(text string) string             { return text }
func (m *MockColorFormatter) Cyan(text string) string                { return text }
func (m *MockColorFormatter) Gray(text string) string                { return text }
func (m *MockColorFormatter) Bold(text string) string                { return text }
func (m *MockColorFormatter) Dim(text string) string                 { return text }
func (m *MockColorFormatter) White(text string) string               { return text }
func (m *MockColorFormatter) Colorize(text, colorName string) string { return text }

// MockIconProvider for testing
type MockIconProvider struct{}

func (m *MockIconProvider) CheckMark() string { return "✓" }
func (m *MockIconProvider) Cross() string     { return "✗" }
func (m *MockIconProvider) Skipped() string   { return "○" }
func (m *MockIconProvider) Running() string   { return "●" }
func (m *MockIconProvider) GetIcon(iconType string) string {
	switch iconType {
	case "check":
		return "✓"
	case "cross":
		return "✗"
	case "skip":
		return "○"
	case "run":
		return "●"
	default:
		return "?"
	}
}

// TestNewTestProcessor_Creation verifies processor initialization
func TestNewTestProcessor_Creation(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := &MockColorFormatter{}
	icons := &MockIconProvider{}

	// Act
	processor := NewTestProcessor(&buf, formatter, icons, 80)

	// Assert
	if processor == nil {
		t.Fatal("Expected processor to be created, got nil")
	}
	if processor.GetWriter() != &buf {
		t.Error("Expected writer to be set correctly")
	}
	if len(processor.GetSuites()) != 0 {
		t.Error("Expected suites to be initialized as empty map")
	}
	if processor.GetStats() == nil {
		t.Error("Expected statistics to be initialized")
	}
}

// TestProcessJSONOutput_ValidSingleTest tests processing a single test result
func TestProcessJSONOutput_ValidSingleTest(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	jsonOutput := `{"Time":"2023-10-01T12:00:00Z","Action":"run","Package":"example.com/test","Test":"TestExample"}
{"Time":"2023-10-01T12:00:01Z","Action":"pass","Package":"example.com/test","Test":"TestExample","Elapsed":1.0}`

	// Act
	err := processor.ProcessJSONOutput(jsonOutput)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	stats := processor.GetStats()
	if stats.TotalTests != 1 {
		t.Errorf("Expected 1 total test, got %d", stats.TotalTests)
	}
	if stats.PassedTests != 1 {
		t.Errorf("Expected 1 passed test, got %d", stats.PassedTests)
	}
	if stats.FailedTests != 0 {
		t.Errorf("Expected 0 failed tests, got %d", stats.FailedTests)
	}
}

// TestProcessJSONOutput_ValidFailedTest tests processing a failed test
func TestProcessJSONOutput_ValidFailedTest(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	jsonOutput := `{"Time":"2023-10-01T12:00:00Z","Action":"run","Package":"example.com/test","Test":"TestFail"}
{"Time":"2023-10-01T12:00:01Z","Action":"output","Package":"example.com/test","Test":"TestFail","Output":"    test_file.go:10: assertion failed\n"}
{"Time":"2023-10-01T12:00:01Z","Action":"fail","Package":"example.com/test","Test":"TestFail","Elapsed":1.0}`

	// Act
	err := processor.ProcessJSONOutput(jsonOutput)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	stats := processor.GetStats()
	if stats.TotalTests != 1 {
		t.Errorf("Expected 1 total test, got %d", stats.TotalTests)
	}
	if stats.FailedTests != 1 {
		t.Errorf("Expected 1 failed test, got %d", stats.FailedTests)
	}
	if stats.PassedTests != 0 {
		t.Errorf("Expected 0 passed tests, got %d", stats.PassedTests)
	}
}

// TestProcessJSONOutput_Subtests tests processing tests with subtests
func TestProcessJSONOutput_Subtests(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	jsonOutput := `{"Time":"2023-10-01T12:00:00Z","Action":"run","Package":"example.com/test","Test":"TestParent"}
{"Time":"2023-10-01T12:00:00Z","Action":"run","Package":"example.com/test","Test":"TestParent/subtest1"}
{"Time":"2023-10-01T12:00:01Z","Action":"pass","Package":"example.com/test","Test":"TestParent/subtest1","Elapsed":0.5}
{"Time":"2023-10-01T12:00:01Z","Action":"run","Package":"example.com/test","Test":"TestParent/subtest2"}
{"Time":"2023-10-01T12:00:02Z","Action":"fail","Package":"example.com/test","Test":"TestParent/subtest2","Elapsed":0.5}
{"Time":"2023-10-01T12:00:02Z","Action":"fail","Package":"example.com/test","Test":"TestParent","Elapsed":2.0}`

	// Act
	err := processor.ProcessJSONOutput(jsonOutput)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	stats := processor.GetStats()
	if stats.TotalTests != 3 { // Parent + 2 subtests
		t.Errorf("Expected 3 total tests, got %d", stats.TotalTests)
	}
	if stats.FailedTests != 2 { // Parent test and subtest2
		t.Errorf("Expected 2 failed tests, got %d", stats.FailedTests)
	}
	if stats.PassedTests != 1 { // subtest1
		t.Errorf("Expected 1 passed test, got %d", stats.PassedTests)
	}
}

// TestProcessJSONOutput_SkippedTest tests processing skipped tests
func TestProcessJSONOutput_SkippedTest(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	jsonOutput := `{"Time":"2023-10-01T12:00:00Z","Action":"run","Package":"example.com/test","Test":"TestSkipped"}
{"Time":"2023-10-01T12:00:01Z","Action":"skip","Package":"example.com/test","Test":"TestSkipped","Elapsed":0.1}`

	// Act
	err := processor.ProcessJSONOutput(jsonOutput)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	stats := processor.GetStats()
	if stats.TotalTests != 1 {
		t.Errorf("Expected 1 total test, got %d", stats.TotalTests)
	}
	if stats.SkippedTests != 1 {
		t.Errorf("Expected 1 skipped test, got %d", stats.SkippedTests)
	}
}

// TestProcessJSONOutput_InvalidJSON tests handling of malformed JSON
func TestProcessJSONOutput_InvalidJSON(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	jsonOutput := `{"invalid": "json"` // Missing closing brace

	// Act
	err := processor.ProcessJSONOutput(jsonOutput)

	// Assert
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse JSON") {
		t.Errorf("Expected 'failed to parse JSON' in error message, got: %v", err)
	}
}

// TestProcessJSONOutput_EmptyOutput tests handling of empty output
func TestProcessJSONOutput_EmptyOutput(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	// Act
	err := processor.ProcessJSONOutput("")

	// Assert
	if err != nil {
		t.Fatalf("Expected no error for empty output, got: %v", err)
	}

	stats := processor.GetStats()
	if stats.TotalTests != 0 {
		t.Errorf("Expected 0 total tests for empty output, got %d", stats.TotalTests)
	}
}

// TestProcessJSONOutput_MultiplePackages tests processing multiple test packages
func TestProcessJSONOutput_MultiplePackages(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	jsonOutput := `{"Time":"2023-10-01T12:00:00Z","Action":"run","Package":"pkg1","Test":"TestPkg1"}
{"Time":"2023-10-01T12:00:01Z","Action":"pass","Package":"pkg1","Test":"TestPkg1","Elapsed":1.0}
{"Time":"2023-10-01T12:00:02Z","Action":"run","Package":"pkg2","Test":"TestPkg2"}
{"Time":"2023-10-01T12:00:03Z","Action":"pass","Package":"pkg2","Test":"TestPkg2","Elapsed":1.0}`

	// Act
	err := processor.ProcessJSONOutput(jsonOutput)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	stats := processor.GetStats()
	if stats.TotalTests != 2 {
		t.Errorf("Expected 2 total tests, got %d", stats.TotalTests)
	}
	if stats.PassedTests != 2 {
		t.Errorf("Expected 2 passed tests, got %d", stats.PassedTests)
	}
}

// TestReset_ClearsState tests that Reset clears the processor state
func TestReset_ClearsState(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	// Add some test data
	suite := &models.TestSuite{
		FilePath:    "test.go",
		TestCount:   1,
		PassedCount: 1,
	}
	processor.AddTestSuite(suite)

	// Act
	processor.Reset()

	// Assert
	stats := processor.GetStats()
	if stats.TotalTests != 0 {
		t.Errorf("Expected 0 total tests after reset, got %d", stats.TotalTests)
	}
	if len(processor.GetSuites()) != 0 {
		t.Errorf("Expected 0 suites after reset, got %d", len(processor.GetSuites()))
	}
}

// TestGetTerminalWidthForProcessor_DefaultFallback tests terminal width fallback
func TestGetTerminalWidthForProcessor_DefaultFallback(t *testing.T) {
	// Act
	width := getTerminalWidthForProcessor()

	// Assert
	if width <= 0 {
		t.Errorf("Expected positive width, got %d", width)
	}
}

// TestAddTestSuite_AddsToSuites tests adding test suites
func TestAddTestSuite_AddsToSuites(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	suite := &models.TestSuite{
		FilePath:     "test.go",
		TestCount:    3,
		PassedCount:  2,
		FailedCount:  1,
		SkippedCount: 0,
	}

	// Act
	processor.AddTestSuite(suite)

	// Assert
	suites := processor.GetSuites()
	if len(suites) != 1 {
		t.Errorf("Expected 1 suite, got %d", len(suites))
	}
	if suites["test.go"] != suite {
		t.Error("Expected suite to be added correctly")
	}

	stats := processor.GetStats()
	if stats.TotalTests != 3 {
		t.Errorf("Expected 3 total tests, got %d", stats.TotalTests)
	}
	if stats.PassedTests != 2 {
		t.Errorf("Expected 2 passed tests, got %d", stats.PassedTests)
	}
	if stats.FailedTests != 1 {
		t.Errorf("Expected 1 failed test, got %d", stats.FailedTests)
	}
}

// TestRenderResults_ShowsSummary tests result rendering
func TestRenderResults_ShowsSummary(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	// Add some test statistics
	processor.statistics.PassedTests = 5
	processor.statistics.FailedTests = 2
	processor.statistics.SkippedTests = 1

	// Act
	err := processor.RenderResults(true)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "5 passed") {
		t.Error("Expected output to contain '5 passed'")
	}
	if !strings.Contains(output, "2 failed") {
		t.Error("Expected output to contain '2 failed'")
	}
	if !strings.Contains(output, "1 skipped") {
		t.Error("Expected output to contain '1 skipped'")
	}
}
