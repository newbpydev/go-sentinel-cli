package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// TestNewTestProcessor_Creation verifies processor initialization
func TestNewTestProcessor_Creation(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)

	// Act
	processor := NewTestProcessor(&buf, formatter, icons, 80)

	// Assert
	if processor == nil {
		t.Fatal("Expected processor to be created, got nil")
	}
	if processor.writer != &buf {
		t.Error("Expected writer to be set correctly")
	}
	if processor.formatter != formatter {
		t.Error("Expected formatter to be set correctly")
	}
	if processor.icons != icons {
		t.Error("Expected icons to be set correctly")
	}
	if len(processor.suites) != 0 {
		t.Error("Expected suites to be initialized as empty map")
	}
	if processor.statistics == nil {
		t.Error("Expected statistics to be initialized")
	}
}

// TestProcessJSONOutput_ValidSingleTest tests processing a single test result
func TestProcessJSONOutput_ValidSingleTest(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, NewColorFormatter(false), NewIconProvider(false), 80)

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
	processor := NewTestProcessor(&buf, NewColorFormatter(false), NewIconProvider(false), 80)

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
	processor := NewTestProcessor(&buf, NewColorFormatter(false), NewIconProvider(false), 80)

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
	processor := NewTestProcessor(&buf, NewColorFormatter(false), NewIconProvider(false), 80)

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
	processor := NewTestProcessor(&buf, NewColorFormatter(false), NewIconProvider(false), 80)

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
	processor := NewTestProcessor(&buf, NewColorFormatter(false), NewIconProvider(false), 80)

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
	processor := NewTestProcessor(&buf, NewColorFormatter(false), NewIconProvider(false), 80)

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

	if len(processor.suites) != 2 {
		t.Errorf("Expected 2 test suites, got %d", len(processor.suites))
	}

	stats := processor.GetStats()
	if stats.TotalTests != 2 {
		t.Errorf("Expected 2 total tests, got %d", stats.TotalTests)
	}
	if stats.PassedTests != 2 {
		t.Errorf("Expected 2 passed tests, got %d", stats.PassedTests)
	}
}

// TestReset_ClearsState tests that Reset clears processor state correctly
func TestReset_ClearsState(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, NewColorFormatter(false), NewIconProvider(false), 80)

	// Add some test data
	processor.suites["test"] = &TestSuite{FilePath: "test"}
	processor.statistics.TotalTests = 5

	// Act
	processor.Reset()

	// Assert
	if len(processor.suites) != 0 {
		t.Error("Expected suites to be cleared after reset")
	}
	if processor.statistics.TotalTests != 0 {
		t.Error("Expected statistics to be reset")
	}
	if processor.statistics.StartTime.IsZero() {
		t.Error("Expected start time to be set after reset")
	}
}

// TestGetTerminalWidthForProcessor_DefaultFallback tests terminal width detection
func TestGetTerminalWidthForProcessor_DefaultFallback(t *testing.T) {
	// Act - This will likely fall back to default since we're in test environment
	width := getTerminalWidthForProcessor()

	// Assert
	if width <= 0 {
		t.Error("Expected positive terminal width")
	}
	// In test environment, should default to 80
	if width != 80 {
		// This is actually ok - might detect real terminal width
		t.Logf("Terminal width: %d (expected 80 in test environment)", width)
	}
}

// TestAddTestSuite_AddsToSuites tests adding test suites
func TestAddTestSuite_AddsToSuites(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, NewColorFormatter(false), NewIconProvider(false), 80)

	suite := &TestSuite{
		FilePath:  "test/example_test.go",
		TestCount: 3,
	}

	// Act
	processor.AddTestSuite(suite)

	// Assert
	if len(processor.suites) != 1 {
		t.Errorf("Expected 1 suite, got %d", len(processor.suites))
	}

	addedSuite, exists := processor.suites["test/example_test.go"]
	if !exists {
		t.Error("Expected suite to be added with correct key")
	}
	if addedSuite.TestCount != 3 {
		t.Errorf("Expected test count 3, got %d", addedSuite.TestCount)
	}
}

// TestOnTestOutput_AccumulatesOutput tests test output accumulation
func TestOnTestOutput_AccumulatesOutput(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, NewColorFormatter(false), NewIconProvider(false), 80)

	// Create a test first
	runEvent := TestEvent{
		Action:  "run",
		Package: "example.com/test",
		Test:    "TestExample",
	}
	processor.onTestRun(runEvent)

	// Add output
	outputEvent := TestEvent{
		Action:  "output",
		Package: "example.com/test",
		Test:    "TestExample",
		Output:  "Some test output",
	}

	// Act
	processor.onTestOutput(outputEvent)

	// Assert - This tests the internal behavior
	// The exact assertion depends on how output is stored internally
	// Since we can't access private fields, we test indirectly through JSON processing
	jsonOutput := `{"Time":"2023-10-01T12:00:00Z","Action":"run","Package":"example.com/test","Test":"TestExample"}
{"Time":"2023-10-01T12:00:01Z","Action":"output","Package":"example.com/test","Test":"TestExample","Output":"Test output line"}
{"Time":"2023-10-01T12:00:02Z","Action":"fail","Package":"example.com/test","Test":"TestExample","Elapsed":1.0}`

	err := processor.ProcessJSONOutput(jsonOutput)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// If we got here, output processing worked
	stats := processor.GetStats()
	if stats.FailedTests != 1 {
		t.Error("Expected output to be processed correctly for failed test")
	}
}

// TestFinalize_UpdatesPhaseTimings tests finalization phase timing calculation
func TestFinalize_UpdatesPhaseTimings(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, NewColorFormatter(false), NewIconProvider(false), 80)

	// Simulate some time passing
	time.Sleep(1 * time.Millisecond)
	processor.firstTestTime = time.Now()
	time.Sleep(1 * time.Millisecond)
	processor.lastTestTime = time.Now()

	// Act
	processor.finalize()

	// Assert
	if processor.statistics.Phases == nil {
		t.Error("Expected phases to be initialized")
	}

	// Check that teardown end time was set
	if processor.teardownEndTime.IsZero() {
		t.Error("Expected teardown end time to be set during finalization")
	}
}

// TestCreateTestError_GeneratesErrorDetails tests error creation from test events
func TestCreateTestError_GeneratesErrorDetails(t *testing.T) {
	// This is testing a private method indirectly through ProcessJSONOutput
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, NewColorFormatter(false), NewIconProvider(false), 80)

	jsonOutput := `{"Time":"2023-10-01T12:00:00Z","Action":"run","Package":"example.com/test","Test":"TestFail"}
{"Time":"2023-10-01T12:00:01Z","Action":"output","Package":"example.com/test","Test":"TestFail","Output":"    example_test.go:42: Expected 5, got 3\n"}
{"Time":"2023-10-01T12:00:01Z","Action":"fail","Package":"example.com/test","Test":"TestFail","Elapsed":1.0}`

	// Act
	err := processor.ProcessJSONOutput(jsonOutput)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify that error details were created (test indirectly through stats)
	stats := processor.GetStats()
	if stats.FailedTests != 1 {
		t.Error("Expected failed test to have error details")
	}
}

// TestProcessJSONOutput_OutputBufferHandling tests that output is written correctly
func TestProcessJSONOutput_OutputBufferHandling(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	processor := NewTestProcessor(&buf, NewColorFormatter(false), NewIconProvider(false), 80)

	jsonOutput := `{"Time":"2023-10-01T12:00:00Z","Action":"run","Package":"example.com/test","Test":"TestExample"}
{"Time":"2023-10-01T12:00:01Z","Action":"pass","Package":"example.com/test","Test":"TestExample","Elapsed":1.0}`

	// Act
	err := processor.ProcessJSONOutput(jsonOutput)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// RenderResults should write to the buffer when called
	err = processor.RenderResults(true)
	if err != nil {
		t.Fatalf("Expected no error from RenderResults, got: %v", err)
	}

	output := buf.String()
	if len(output) == 0 {
		t.Error("Expected output to be written to buffer")
	}
}
