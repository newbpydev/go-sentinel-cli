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

// TestReset_ClearsState tests that Reset properly clears processor state
func TestReset_ClearsState(t *testing.T) {
	t.Parallel()

	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	// Add some test data first
	suite := &models.TestSuite{
		FilePath:     "test.go",
		TestCount:    5,
		PassedCount:  3,
		FailedCount:  2,
		SkippedCount: 0,
	}
	processor.AddTestSuite(suite)

	// Verify data exists
	if len(processor.GetSuites()) == 0 {
		t.Fatal("Expected suite to be added before reset")
	}
	if processor.GetStats().TotalTests == 0 {
		t.Fatal("Expected stats to have data before reset")
	}

	// Act
	processor.Reset()

	// Assert
	suites := processor.GetSuites()
	if len(suites) != 0 {
		t.Errorf("Expected suites to be cleared after reset, got %d", len(suites))
	}

	stats := processor.GetStats()
	if stats.TotalTests != 0 {
		t.Errorf("Expected total tests to be 0 after reset, got %d", stats.TotalTests)
	}
	if stats.PassedTests != 0 {
		t.Errorf("Expected passed tests to be 0 after reset, got %d", stats.PassedTests)
	}
	if stats.FailedTests != 0 {
		t.Errorf("Expected failed tests to be 0 after reset, got %d", stats.FailedTests)
	}
	if stats.SkippedTests != 0 {
		t.Errorf("Expected skipped tests to be 0 after reset, got %d", stats.SkippedTests)
	}

	// Verify timestamps are reset
	if stats.StartTime.IsZero() {
		t.Error("Expected start time to be set after reset")
	}
	if stats.Phases == nil {
		t.Error("Expected phases map to be initialized after reset")
	}
}

// TestGetStats_ReturnsCurrentStats tests GetStats method
func TestGetStats_ReturnsCurrentStats(t *testing.T) {
	t.Parallel()

	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	// Act
	stats := processor.GetStats()

	// Assert
	if stats == nil {
		t.Fatal("Expected stats to be returned, got nil")
	}

	// Verify initial state
	if stats.TotalTests != 0 {
		t.Errorf("Expected initial total tests to be 0, got %d", stats.TotalTests)
	}
	if stats.StartTime.IsZero() {
		t.Error("Expected start time to be set")
	}
	if stats.Phases == nil {
		t.Error("Expected phases map to be initialized")
	}
}

// TestGetWriter_ReturnsWriter tests GetWriter method
func TestGetWriter_ReturnsWriter(t *testing.T) {
	t.Parallel()

	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	// Act
	writer := processor.GetWriter()

	// Assert
	if writer != &buf {
		t.Error("Expected GetWriter to return the same writer passed to constructor")
	}
}

// TestGetSuites_ReturnsCurrentSuites tests GetSuites method
func TestGetSuites_ReturnsCurrentSuites(t *testing.T) {
	t.Parallel()

	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	// Act
	suites := processor.GetSuites()

	// Assert
	if suites == nil {
		t.Fatal("Expected suites map to be returned, got nil")
	}

	if len(suites) != 0 {
		t.Errorf("Expected initial suites to be empty, got %d", len(suites))
	}

	// Add a suite and verify it's returned
	suite := &models.TestSuite{
		FilePath:    "test.go",
		TestCount:   1,
		PassedCount: 1,
	}
	processor.AddTestSuite(suite)

	suites = processor.GetSuites()
	if len(suites) != 1 {
		t.Errorf("Expected 1 suite after adding, got %d", len(suites))
	}

	if _, exists := suites["test.go"]; !exists {
		t.Error("Expected suite with key 'test.go' to exist")
	}
}

// TestRenderResults_WithSummary tests RenderResults with summary enabled
func TestRenderResults_WithSummary(t *testing.T) {
	t.Parallel()

	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	// Add some test data
	processor.GetStats().PassedTests = 5
	processor.GetStats().FailedTests = 2
	processor.GetStats().SkippedTests = 1

	// Act
	err := processor.RenderResults(true)

	// Assert
	if err != nil {
		t.Errorf("Expected no error from RenderResults, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Tests completed") {
		t.Error("Expected output to contain 'Tests completed'")
	}
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

// TestRenderResults_WithoutSummary tests RenderResults with summary disabled
func TestRenderResults_WithoutSummary(t *testing.T) {
	t.Parallel()

	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	// Add some test data
	processor.GetStats().PassedTests = 3
	processor.GetStats().FailedTests = 1

	// Act
	err := processor.RenderResults(false)

	// Assert
	if err != nil {
		t.Errorf("Expected no error from RenderResults, got: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Errorf("Expected no output when summary is disabled, got: %s", output)
	}
}

// TestAddTestSuite_EmptyFilePath tests AddTestSuite with empty file path
func TestAddTestSuite_EmptyFilePath(t *testing.T) {
	t.Parallel()

	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	suite := &models.TestSuite{
		FilePath:     "", // Empty file path
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

	// Should use "unknown" as default file path
	if _, exists := suites["unknown"]; !exists {
		t.Error("Expected suite with key 'unknown' to exist for empty file path")
	}

	stats := processor.GetStats()
	if stats.TotalFiles != 1 {
		t.Errorf("Expected 1 total file, got %d", stats.TotalFiles)
	}
	if stats.FailedFiles != 1 {
		t.Errorf("Expected 1 failed file (has failed tests), got %d", stats.FailedFiles)
	}
	if stats.PassedFiles != 0 {
		t.Errorf("Expected 0 passed files (has failed tests), got %d", stats.PassedFiles)
	}
}

// TestAddTestSuite_PassedFile tests AddTestSuite with all tests passing
func TestAddTestSuite_PassedFile(t *testing.T) {
	t.Parallel()

	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, &MockColorFormatter{}, &MockIconProvider{}, 80)

	suite := &models.TestSuite{
		FilePath:     "passed_test.go",
		TestCount:    5,
		PassedCount:  5,
		FailedCount:  0, // No failed tests
		SkippedCount: 0,
	}

	// Act
	processor.AddTestSuite(suite)

	// Assert
	stats := processor.GetStats()
	if stats.PassedFiles != 1 {
		t.Errorf("Expected 1 passed file (no failed tests), got %d", stats.PassedFiles)
	}
	if stats.FailedFiles != 0 {
		t.Errorf("Expected 0 failed files (no failed tests), got %d", stats.FailedFiles)
	}
	if stats.TotalTests != 5 {
		t.Errorf("Expected 5 total tests, got %d", stats.TotalTests)
	}
	if stats.PassedTests != 5 {
		t.Errorf("Expected 5 passed tests, got %d", stats.PassedTests)
	}
}

// TestGetTerminalWidthForProcessor_DefaultFallback tests terminal width detection
func TestGetTerminalWidthForProcessor_DefaultFallback(t *testing.T) {
	t.Parallel()

	// Act
	width := getTerminalWidthForProcessor()

	// Assert
	// Should return either detected width or default fallback (80)
	if width <= 0 {
		t.Errorf("Expected positive width, got %d", width)
	}

	// In test environment, likely to get the default fallback
	// but we can't guarantee it, so just check it's reasonable
	if width < 20 || width > 1000 {
		t.Errorf("Expected reasonable width (20-1000), got %d", width)
	}
}
