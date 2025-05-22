package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

// Test 2.5.1: Collapse passing test suites by default
func TestCollapsePassingSuites(t *testing.T) {
	formatter := NewColorFormatter(true)
	icons := NewIconProvider(true)

	// Create a test suite with all passing tests
	suite := &TestSuite{
		FilePath:     "github.com/user/project/pkg/passing_test.go",
		TestCount:    5,
		PassedCount:  5,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     100 * time.Millisecond,
		MemoryUsage:  1024 * 1024,
	}

	// Add passing tests
	for i := 1; i <= 5; i++ {
		test := &TestResult{
			Name:     fmt.Sprintf("TestPassing%d", i),
			Status:   StatusPassed,
			Duration: 20 * time.Millisecond,
			Package:  "github.com/user/project/pkg",
		}
		suite.Tests = append(suite.Tests, test)
	}

	// Render the suite
	var buf bytes.Buffer
	renderer := NewSuiteRenderer(&buf, formatter, icons, 80)

	// Test collapsed mode
	err := renderer.RenderSuite(suite, true)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()

	// Should contain header
	if !strings.Contains(output, "passing_test.go") {
		t.Errorf("Expected output to contain file name, got: %s", output)
	}

	// Should not show individual test details in collapsed mode
	for i := 1; i <= 5; i++ {
		testName := fmt.Sprintf("TestPassing%d", i)
		if strings.Contains(output, testName) {
			t.Errorf("Expected collapsed output to NOT contain test '%s', but it does", testName)
		}
	}

	// Should contain summary showing number of tests
	if !strings.Contains(output, "Suite passed") && !strings.Contains(output, "5 tests") {
		t.Errorf("Expected output to contain summary of passed tests, got: %s", output)
	}
}

// Test 2.5.2: Expand test suites with failing tests
func TestExpandFailingSuites(t *testing.T) {
	formatter := NewColorFormatter(true)
	icons := NewIconProvider(true)

	// Create a test suite with a failing test
	suite := &TestSuite{
		FilePath:     "github.com/user/project/pkg/failing_test.go",
		TestCount:    5,
		PassedCount:  4,
		FailedCount:  1,
		SkippedCount: 0,
		Duration:     100 * time.Millisecond,
		MemoryUsage:  1024 * 1024,
	}

	// Add passing tests
	for i := 1; i <= 4; i++ {
		test := &TestResult{
			Name:     fmt.Sprintf("TestPassing%d", i),
			Status:   StatusPassed,
			Duration: 20 * time.Millisecond,
			Package:  "github.com/user/project/pkg",
		}
		suite.Tests = append(suite.Tests, test)
	}

	// Add one failing test
	failingTest := &TestResult{
		Name:     "TestFailing",
		Status:   StatusFailed,
		Duration: 20 * time.Millisecond,
		Package:  "github.com/user/project/pkg",
		Error: &TestError{
			Message: "Failed assertion",
			Type:    "AssertionError",
		},
	}
	suite.Tests = append(suite.Tests, failingTest)

	// Render the suite
	var buf bytes.Buffer
	renderer := NewSuiteRenderer(&buf, formatter, icons, 80)

	// Test auto-expand mode
	err := renderer.RenderSuite(suite, true)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()

	// Should contain header
	if !strings.Contains(output, "failing_test.go") {
		t.Errorf("Expected output to contain file name, got: %s", output)
	}

	// Should show failing test in expanded mode
	if !strings.Contains(output, "TestFailing") {
		t.Errorf("Expected expanded output to contain failing test, got: %s", output)
	}

	// Should show error details
	if !strings.Contains(output, "Failed assertion") {
		t.Errorf("Expected output to contain error message, got: %s", output)
	}
}

// Test 2.5.3: Properly indent and format nested tests
func TestNestedTestIndentation(t *testing.T) {
	formatter := NewColorFormatter(true)
	icons := NewIconProvider(true)

	// Create a test suite with nested tests
	suite := &TestSuite{
		FilePath:     "github.com/user/project/pkg/nested_test.go",
		TestCount:    3,
		PassedCount:  2,
		FailedCount:  1,
		SkippedCount: 0,
		Duration:     100 * time.Millisecond,
		MemoryUsage:  1024 * 1024,
	}

	// Add parent test
	parentTest := &TestResult{
		Name:     "TestParent",
		Status:   StatusPassed,
		Duration: 50 * time.Millisecond,
		Package:  "github.com/user/project/pkg",
	}

	// Add subtests
	passingSubtest := &TestResult{
		Name:     "TestParent/Subtest1",
		Status:   StatusPassed,
		Duration: 20 * time.Millisecond,
		Package:  "github.com/user/project/pkg",
		Parent:   "TestParent",
	}

	failingSubtest := &TestResult{
		Name:     "TestParent/Subtest2",
		Status:   StatusFailed,
		Duration: 20 * time.Millisecond,
		Package:  "github.com/user/project/pkg",
		Parent:   "TestParent",
		Error: &TestError{
			Message: "Subtest failure",
			Type:    "Error",
		},
	}

	// Add subtests to parent
	parentTest.Subtests = append(parentTest.Subtests, passingSubtest, failingSubtest)

	// Add to suite
	suite.Tests = append(suite.Tests, parentTest)

	// Render the suite
	var buf bytes.Buffer
	renderer := NewSuiteRenderer(&buf, formatter, icons, 80)

	// Test expanded mode
	err := renderer.RenderSuite(suite, false) // Force expanded mode
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()

	// Split into lines to check indentation
	lines := strings.Split(output, "\n")

	// Find parent and subtest lines
	var parentLine, subtest1Line, subtest2Line string
	for _, line := range lines {
		if strings.Contains(line, "TestParent") && !strings.Contains(line, "Subtest") {
			parentLine = line
		} else if strings.Contains(line, "Subtest1") {
			subtest1Line = line
		} else if strings.Contains(line, "Subtest2") {
			subtest2Line = line
		}
	}

	// Check parent line exists
	if parentLine == "" {
		t.Fatalf("Expected output to contain parent test line, got: %s", output)
	}

	// Check subtest lines exist
	if subtest1Line == "" || subtest2Line == "" {
		t.Fatalf("Expected output to contain subtest lines, got: %s", output)
	}

	// Check subtests are indented
	if !strings.Contains(subtest1Line, "  ") {
		t.Errorf("Expected Subtest1 to be indented, got parent: '%s', subtest: '%s'", parentLine, subtest1Line)
	}

	if !strings.Contains(subtest2Line, "  ") {
		t.Errorf("Expected Subtest2 to be indented, got parent: '%s', subtest: '%s'", parentLine, subtest2Line)
	}
}

// Test 2.5.4: Handle edge cases like empty suites or all skipped tests
func TestEdgeCaseSuites(t *testing.T) {
	formatter := NewColorFormatter(true)
	icons := NewIconProvider(true)

	// Case 1: Empty suite
	emptySuite := &TestSuite{
		FilePath:  "github.com/user/project/pkg/empty_test.go",
		TestCount: 0,
		Duration:  10 * time.Millisecond,
	}

	// Case 2: All skipped tests
	skippedSuite := &TestSuite{
		FilePath:     "github.com/user/project/pkg/skipped_test.go",
		TestCount:    3,
		PassedCount:  0,
		FailedCount:  0,
		SkippedCount: 3,
		Duration:     20 * time.Millisecond,
	}

	// Add skipped tests
	for i := 1; i <= 3; i++ {
		test := &TestResult{
			Name:     fmt.Sprintf("TestSkipped%d", i),
			Status:   StatusSkipped,
			Duration: 5 * time.Millisecond,
			Package:  "github.com/user/project/pkg",
		}
		skippedSuite.Tests = append(skippedSuite.Tests, test)
	}

	// Test empty suite
	var emptyBuf bytes.Buffer
	emptyRenderer := NewSuiteRenderer(&emptyBuf, formatter, icons, 80)

	err := emptyRenderer.RenderSuite(emptySuite, true)
	if err != nil {
		t.Fatalf("Expected no error for empty suite, got: %v", err)
	}

	emptyOutput := emptyBuf.String()

	// Should contain header and empty indication
	if !strings.Contains(emptyOutput, "empty_test.go") {
		t.Errorf("Expected output to contain file name for empty suite, got: %s", emptyOutput)
	}

	if !strings.Contains(emptyOutput, "0 test") {
		t.Errorf("Expected output to indicate 0 tests, got: %s", emptyOutput)
	}

	// Test skipped suite
	var skippedBuf bytes.Buffer
	skippedRenderer := NewSuiteRenderer(&skippedBuf, formatter, icons, 80)

	err = skippedRenderer.RenderSuite(skippedSuite, true)
	if err != nil {
		t.Fatalf("Expected no error for skipped suite, got: %v", err)
	}

	skippedOutput := skippedBuf.String()

	// Should contain header and skipped indication
	if !strings.Contains(skippedOutput, "skipped_test.go") {
		t.Errorf("Expected output to contain file name for skipped suite, got: %s", skippedOutput)
	}

	if !strings.Contains(skippedOutput, "All tests skipped") && !strings.Contains(skippedOutput, "3 tests") {
		t.Errorf("Expected output to indicate 3 skipped tests, got: %s", skippedOutput)
	}
}
