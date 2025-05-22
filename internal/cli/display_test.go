package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// Test 2.1.1: Format test file path with colorized file name
func TestFormatTestFilePath(t *testing.T) {
	formatter := NewColorFormatter(true)

	// Test basic file path formatting
	filePath := "github.com/user/project/pkg/example_test.go"
	formatted := FormatFilePath(formatter, filePath)

	// Should contain the original path
	if formatted == "" {
		t.Fatal("Formatted file path is empty")
	}

	// The file name should be highlighted in a different color than the directory
	if formatted == filePath {
		t.Errorf("Expected file path to be formatted with colors, got the original path")
	}

	// Test with Windows-style path
	winPath := `C:\Users\User\project\pkg\example_test.go`
	winFormatted := FormatFilePath(formatter, winPath)

	// Should contain the original file name
	if !containsSubstringWithCase(winFormatted, "example_test.go") {
		t.Errorf("Formatted Windows path does not contain the file name")
	}

	// Test with relative path
	relPath := "./pkg/example_test.go"
	relFormatted := FormatFilePath(formatter, relPath)

	// Should contain the original file name
	if !containsSubstringWithCase(relFormatted, "example_test.go") {
		t.Errorf("Formatted relative path does not contain the file name")
	}

	// Test with colors disabled
	noColorFormatter := NewColorFormatter(false)
	plainFormatted := FormatFilePath(noColorFormatter, filePath)

	// Should be the same as the original path or at least contain all parts
	if !containsSubstringWithCase(plainFormatted, "example_test.go") {
		t.Errorf("Formatted path without colors does not contain the file name")
	}
}

// Test 2.1.2: Display test counts with failed test highlighting
func TestDisplayTestCounts(t *testing.T) {
	formatter := NewColorFormatter(true)

	// Test with all tests passing
	allPassed := formatTestCounts(formatter, 10, 10, 0, 0)
	if !containsSubstringWithCase(allPassed, "10") || !containsSubstringWithCase(allPassed, "pass") {
		t.Errorf("Expected test count to show 10 passed tests, got: %s", allPassed)
	}

	// Test with some tests failing
	someFailed := formatTestCounts(formatter, 10, 7, 3, 0)
	if !containsSubstringWithCase(someFailed, "7") || !containsSubstringWithCase(someFailed, "pass") ||
		!containsSubstringWithCase(someFailed, "3") || !containsSubstringWithCase(someFailed, "fail") {
		t.Errorf("Expected test count to show 7 passed and 3 failed tests, got: %s", someFailed)
	}

	// Test with skipped tests
	withSkipped := formatTestCounts(formatter, 12, 8, 2, 2)
	if !containsSubstringWithCase(withSkipped, "8") || !containsSubstringWithCase(withSkipped, "pass") ||
		!containsSubstringWithCase(withSkipped, "2") || !containsSubstringWithCase(withSkipped, "fail") ||
		!containsSubstringWithCase(withSkipped, "skip") {
		t.Errorf("Expected test count to show passed, failed, and skipped tests, got: %s", withSkipped)
	}

	// Test with colors disabled
	noColorFormatter := NewColorFormatter(false)
	plainFormatted := formatTestCounts(noColorFormatter, 10, 7, 3, 0)
	if !containsSubstringWithCase(plainFormatted, "7") || !containsSubstringWithCase(plainFormatted, "pass") ||
		!containsSubstringWithCase(plainFormatted, "3") || !containsSubstringWithCase(plainFormatted, "fail") {
		t.Errorf("Expected plain test count to show passed and failed tests, got: %s", plainFormatted)
	}
}

// Test 2.1.3: Show accurate test duration with proper formatting
func TestFormatTestDuration(t *testing.T) {
	formatter := NewColorFormatter(true)

	// Test millisecond formatting
	ms := FormatDuration(formatter, 50*time.Millisecond)
	if !containsSubstringWithCase(ms, "50ms") {
		t.Errorf("Expected duration to be formatted as '50ms', got: %s", ms)
	}

	// Test second formatting
	sec := FormatDuration(formatter, 1500*time.Millisecond)
	if !containsSubstringWithCase(sec, "1.5s") {
		t.Errorf("Expected duration to be formatted as '1.5s', got: %s", sec)
	}

	// Test minute formatting
	min := FormatDuration(formatter, 90*time.Second)
	if !containsString(min, "1m 30s") {
		t.Errorf("Expected duration to be formatted as '1m 30s', got: %s", min)
	}

	// Test with very small duration
	tiny := FormatDuration(formatter, 50*time.Nanosecond)
	if !containsString(tiny, "0ms") {
		t.Errorf("Expected tiny duration to be formatted as '0ms', got: %s", tiny)
	}

	// Test with colors disabled
	noColorFormatter := NewColorFormatter(false)
	plainFormatted := FormatDuration(noColorFormatter, 50*time.Millisecond)
	if !strings.Contains(plainFormatted, "50ms") {
		t.Errorf("Expected plain duration to contain '50ms', got: %s", plainFormatted)
	}
}

// Test 2.1.4: Include memory usage information
func TestFormatMemoryUsage(t *testing.T) {
	formatter := NewColorFormatter(true)

	// Test KB formatting
	kb := FormatMemoryUsage(formatter, 1024)
	if !containsString(kb, "1 KB") {
		t.Errorf("Expected memory usage to contain '1 KB', got: %s", kb)
	}

	// Test MB formatting
	mb := FormatMemoryUsage(formatter, 1024*1024)
	if !containsString(mb, "1 MB") {
		t.Errorf("Expected memory usage to contain '1 MB', got: %s", mb)
	}

	// Test GB formatting
	gb := FormatMemoryUsage(formatter, 1024*1024*1024)
	if !containsString(gb, "1.00 GB") || !containsString(gb, "heap used") {
		t.Errorf("Expected memory usage to contain '1.00 GB' and 'heap used', got: %s", gb)
	}

	// Test decimal formatting
	decimal := FormatMemoryUsage(formatter, uint64(float64(1024*1024)*1.5))
	if !containsString(decimal, "1.5 MB") {
		t.Errorf("Expected memory usage to contain '1.5 MB', got: %s", decimal)
	}

	// Test with colors disabled
	noColorFormatter := NewColorFormatter(false)
	plainFormatted := FormatMemoryUsage(noColorFormatter, 1024*1024)
	if !strings.Contains(plainFormatted, "1 MB") {
		t.Errorf("Expected plain memory usage to contain '1 MB', got: %s", plainFormatted)
	}
}

// Test 2.1.5: Handle multiline headers gracefully
func TestMultilineHeaders(t *testing.T) {
	formatter := NewColorFormatter(true)
	icons := NewIconProvider(true)

	// Create a test suite with a very long file path
	suite := &TestSuite{
		FilePath:    "github.com/very/long/path/to/module/with/many/nested/directories/that/will/likely/cause/wrapping/in/the/terminal/example_test.go",
		TestCount:   10,
		PassedCount: 7,
		FailedCount: 3,
		Duration:    1500 * time.Millisecond,
		MemoryUsage: 1024 * 1024, // 1MB
	}

	// Format the header
	var buf bytes.Buffer
	renderer := NewHeaderRenderer(&buf, formatter, icons, 40) // With narrow width to force wrapping
	err := renderer.RenderSuiteHeader(suite)

	// Check that it rendered without error
	if err != nil {
		t.Errorf("Expected header rendering to succeed, got error: %v", err)
	}

	// Check that output is not empty
	output := buf.String()
	if output == "" {
		t.Error("Expected header output to be non-empty")
	}

	// Check that it contains all the important information
	if !containsString(output, "example_test.go") {
		t.Error("Header doesn't contain the file name")
	}

	if !containsString(output, "7") || !containsString(output, "pass") ||
		!containsString(output, "3") || !containsString(output, "fail") {
		t.Error("Header doesn't properly display test counts")
	}

	// Test with a very narrow width
	var buf2 bytes.Buffer
	narrowRenderer := NewHeaderRenderer(&buf2, formatter, icons, 10) // Extremely narrow
	err = narrowRenderer.RenderSuiteHeader(suite)

	// Should still render without error
	if err != nil {
		t.Errorf("Expected narrow header rendering to succeed, got error: %v", err)
	}

	// Should still contain minimal information
	narrowOutput := buf2.String()
	if !containsString(narrowOutput, "example_test.go") {
		t.Error("Narrow header doesn't contain the file name")
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return s != "" && substr != "" &&
		s != substr &&
		strings.Contains(s, substr)
}

// Helper function to check if a string contains a substring
func containsSubstringWithCase(s, substr string) bool {
	return s != "" && substr != "" &&
		s != substr &&
		strings.Contains(s, substr)
}
