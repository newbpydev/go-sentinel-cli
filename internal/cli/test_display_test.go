package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// Test 2.3.1: Format passed tests with green check and name
func TestFormatPassedTest(t *testing.T) {
	formatter := NewColorFormatter(true)
	icons := NewIconProvider(true)

	// Create a passed test
	result := &TestResult{
		Name:     "TestPassed",
		Status:   StatusPassed,
		Duration: 50 * time.Millisecond,
		Package:  "github.com/user/project/pkg",
		Test:     "TestPassed",
	}

	// Format the test
	var buf bytes.Buffer
	renderer := NewTestRenderer(&buf, formatter, icons)
	err := renderer.RenderTestResult(result, 0)

	// Ensure no error
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check the output
	output := buf.String()

	// Should contain check mark and test name
	if !containsTestIcon(output, icons.CheckMark()) {
		t.Errorf("Expected output to contain check mark, got: %s", output)
	}

	if !strings.Contains(output, "TestPassed") {
		t.Errorf("Expected output to contain test name 'TestPassed', got: %s", output)
	}
}

// Test 2.3.2: Format failed tests with red X and name
func TestFormatFailedTest(t *testing.T) {
	formatter := NewColorFormatter(true)
	icons := NewIconProvider(true)

	// Create a failed test
	result := &TestResult{
		Name:   "TestFailed",
		Status: StatusFailed,
		Error: &TestError{
			Message: "Expected 5, got 10",
			Type:    "AssertionError",
		},
		Duration: 100 * time.Millisecond,
		Package:  "github.com/user/project/pkg",
		Test:     "TestFailed",
	}

	// Format the test
	var buf bytes.Buffer
	renderer := NewTestRenderer(&buf, formatter, icons)
	err := renderer.RenderTestResult(result, 0)

	// Ensure no error
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check the output
	output := buf.String()

	// Should contain X mark and test name
	if !containsTestIcon(output, icons.Cross()) {
		t.Errorf("Expected output to contain cross mark, got: %s", output)
	}

	if !strings.Contains(output, "TestFailed") {
		t.Errorf("Expected output to contain test name 'TestFailed', got: %s", output)
	}

	// Should contain error message
	if !strings.Contains(output, "Expected 5, got 10") {
		t.Errorf("Expected output to contain error message, got: %s", output)
	}
}

// Test 2.3.3: Indent subtests/nested tests correctly
func TestIndentSubtests(t *testing.T) {
	formatter := NewColorFormatter(true)
	icons := NewIconProvider(true)

	// Create a parent test with subtests
	parent := &TestResult{
		Name:     "TestParent",
		Status:   StatusPassed,
		Duration: 150 * time.Millisecond,
		Package:  "github.com/user/project/pkg",
		Test:     "TestParent",
	}

	// Add subtests
	subtest1 := &TestResult{
		Name:     "TestParent/SubTest1",
		Status:   StatusPassed,
		Duration: 50 * time.Millisecond,
		Package:  "github.com/user/project/pkg",
		Test:     "TestParent/SubTest1",
		Parent:   "TestParent",
	}

	subtest2 := &TestResult{
		Name:     "TestParent/SubTest2",
		Status:   StatusFailed,
		Duration: 50 * time.Millisecond,
		Package:  "github.com/user/project/pkg",
		Test:     "TestParent/SubTest2",
		Parent:   "TestParent",
		Error: &TestError{
			Message: "Subtest error",
			Type:    "Error",
		},
	}

	parent.Subtests = append(parent.Subtests, subtest1, subtest2)

	// Format the tests
	var buf bytes.Buffer
	renderer := NewTestRenderer(&buf, formatter, icons)
	err := renderer.RenderTestResult(parent, 0)

	// Ensure no error
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Check the output
	output := buf.String()

	// Parent should appear before subtests
	parentIndex := strings.Index(output, "TestParent")
	subtest1Index := strings.Index(output, "SubTest1")
	subtest2Index := strings.Index(output, "SubTest2")

	if parentIndex == -1 || subtest1Index == -1 || subtest2Index == -1 {
		t.Fatalf("Expected output to contain parent and subtests, got: %s", output)
	}

	if parentIndex > subtest1Index || parentIndex > subtest2Index {
		t.Errorf("Expected parent to appear before subtests in output")
	}

	// Subtests should be indented
	lines := strings.Split(output, "\n")
	foundIndentation := false

	for _, line := range lines {
		if strings.Contains(line, "SubTest") && strings.HasPrefix(line, "  ") {
			foundIndentation = true
			break
		}
	}

	if !foundIndentation {
		t.Errorf("Expected subtests to be indented, got: %s", output)
	}
}

// Test 2.3.4: Handle test names with special characters
func TestSpecialCharactersInTestNames(t *testing.T) {
	formatter := NewColorFormatter(true)
	icons := NewIconProvider(true)

	// Create tests with special characters
	specialTests := []*TestResult{
		{
			Name:     "Test with spaces",
			Status:   StatusPassed,
			Duration: 50 * time.Millisecond,
		},
		{
			Name:     "Test_with_underscores",
			Status:   StatusPassed,
			Duration: 50 * time.Millisecond,
		},
		{
			Name:     "Test-with-hyphens",
			Status:   StatusPassed,
			Duration: 50 * time.Millisecond,
		},
		{
			Name:     "Test:with:colons",
			Status:   StatusPassed,
			Duration: 50 * time.Millisecond,
		},
		{
			Name:     "Test.with.dots",
			Status:   StatusPassed,
			Duration: 50 * time.Millisecond,
		},
	}

	// Format each test
	for _, test := range specialTests {
		var buf bytes.Buffer
		renderer := NewTestRenderer(&buf, formatter, icons)
		err := renderer.RenderTestResult(test, 0)

		// Ensure no error
		if err != nil {
			t.Fatalf("Expected no error for test '%s', got: %v", test.Name, err)
		}

		// Check the output
		output := buf.String()

		// Should contain the test name
		if !strings.Contains(output, test.Name) {
			t.Errorf("Expected output to contain test name '%s', got: %s", test.Name, output)
		}
	}
}

// Test 2.3.5: Show appropriate error messages for failed tests
func TestErrorMessagesForFailedTests(t *testing.T) {
	formatter := NewColorFormatter(true)
	icons := NewIconProvider(true)

	// Create different types of errors
	errorTests := []*TestResult{
		{
			Name:   "TestAssertionError",
			Status: StatusFailed,
			Error: &TestError{
				Message:  "Expected 5, got 10",
				Type:     "AssertionError",
				Expected: "5",
				Actual:   "10",
			},
		},
		{
			Name:   "TestPanic",
			Status: StatusFailed,
			Error: &TestError{
				Message: "panic: runtime error: index out of bounds",
				Type:    "Panic",
				Stack:   "goroutine 1 [running]:\npanic(0x123456)\n...",
			},
		},
		{
			Name:   "TestTimeout",
			Status: StatusFailed,
			Error: &TestError{
				Message: "test timed out after 10s",
				Type:    "Timeout",
			},
		},
	}

	// Format each test
	for _, test := range errorTests {
		var buf bytes.Buffer
		renderer := NewTestRenderer(&buf, formatter, icons)
		err := renderer.RenderTestResult(test, 0)

		// Ensure no error
		if err != nil {
			t.Fatalf("Expected no error for test '%s', got: %v", test.Name, err)
		}

		// Check the output
		output := buf.String()

		// Should contain the error message
		if !strings.Contains(output, test.Error.Message) {
			t.Errorf("Expected output to contain error message '%s', got: %s", test.Error.Message, output)
		}

		// For assertion errors, should show expected/actual
		if test.Error.Type == "AssertionError" &&
			(!strings.Contains(output, test.Error.Expected) || !strings.Contains(output, test.Error.Actual)) {
			t.Errorf("Expected output to contain expected/actual values for assertion error, got: %s", output)
		}
	}
}

// Helper function to check if output contains a test icon
// This is needed because the icons may be part of ANSI color sequences
func containsTestIcon(output, icon string) bool {
	// Remove ANSI color codes for comparison
	cleanOutput := stripAnsiCodes(output)
	return strings.Contains(cleanOutput, icon)
}

// stripAnsiCodes removes ANSI color codes from a string
func stripAnsiCodes(s string) string {
	r := strings.NewReplacer(
		"\033[0m", "",
		"\033[1m", "",
		"\033[2m", "",
		"\033[31m", "",
		"\033[32m", "",
		"\033[33m", "",
		"\033[34m", "",
		"\033[35m", "",
		"\033[36m", "",
		"\033[90m", "",
	)
	return r.Replace(s)
}
