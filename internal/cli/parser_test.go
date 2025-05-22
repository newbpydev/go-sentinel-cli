package cli

import (
	"strings"
	"testing"
	"time"
)

// Test 1.1.4: Parse go test output correctly into data structures
func TestParseGoTestOutput(t *testing.T) {
	// Sample JSON output from go test -json
	jsonOutput := `
{"Time":"2023-05-15T12:00:00.1Z","Action":"run","Package":"github.com/user/project/pkg","Test":"TestExample"}
{"Time":"2023-05-15T12:00:00.15Z","Action":"output","Package":"github.com/user/project/pkg","Test":"TestExample","Output":"=== RUN   TestExample\n"}
{"Time":"2023-05-15T12:00:00.2Z","Action":"output","Package":"github.com/user/project/pkg","Test":"TestExample","Output":"    example_test.go:42: Test output\n"}
{"Time":"2023-05-15T12:00:00.25Z","Action":"pass","Package":"github.com/user/project/pkg","Test":"TestExample","Elapsed":0.05}
{"Time":"2023-05-15T12:00:00.3Z","Action":"run","Package":"github.com/user/project/pkg","Test":"TestFailing"}
{"Time":"2023-05-15T12:00:00.35Z","Action":"output","Package":"github.com/user/project/pkg","Test":"TestFailing","Output":"=== RUN   TestFailing\n"}
{"Time":"2023-05-15T12:00:00.4Z","Action":"output","Package":"github.com/user/project/pkg","Test":"TestFailing","Output":"    example_test.go:50: Expected 5, got 10\n"}
{"Time":"2023-05-15T12:00:00.45Z","Action":"fail","Package":"github.com/user/project/pkg","Test":"TestFailing","Elapsed":0.1}
{"Time":"2023-05-15T12:00:00.5Z","Action":"output","Package":"github.com/user/project/pkg","Output":"FAIL\tgithub.com/user/project/pkg\t0.15s\n"}
{"Time":"2023-05-15T12:00:00.55Z","Action":"fail","Package":"github.com/user/project/pkg","Elapsed":0.15}
`

	// Create parser
	parser := NewParser()
	if parser == nil {
		t.Fatal("Expected parser to be created")
	}

	// Parse the JSON output
	reader := strings.NewReader(jsonOutput)
	results, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Error parsing test output: %v", err)
	}

	// Validate results
	if len(results) != 1 {
		t.Fatalf("Expected 1 package, got %d", len(results))
	}

	pkg := results[0]
	if pkg.Package != "github.com/user/project/pkg" {
		t.Errorf("Expected package name to be 'github.com/user/project/pkg', got '%s'", pkg.Package)
	}

	if len(pkg.Tests) != 2 {
		t.Fatalf("Expected 2 tests, got %d", len(pkg.Tests))
	}

	// Find tests by name
	var testExample, testFailing *TestResult
	for _, test := range pkg.Tests {
		if test.Name == "TestExample" {
			testExample = test
		} else if test.Name == "TestFailing" {
			testFailing = test
		}
	}

	// Check first test (passing)
	if testExample == nil {
		t.Fatal("Expected to find TestExample")
	}
	if testExample.Status != StatusPassed {
		t.Errorf("Expected test status to be 'passed', got '%s'", testExample.Status)
	}
	if testExample.Duration != 50*time.Millisecond {
		t.Errorf("Expected duration to be 50ms, got '%v'", testExample.Duration)
	}

	// Check second test (failing)
	if testFailing == nil {
		t.Fatal("Expected to find TestFailing")
	}
	if testFailing.Status != StatusFailed {
		t.Errorf("Expected test status to be 'failed', got '%s'", testFailing.Status)
	}
	if testFailing.Duration != 100*time.Millisecond {
		t.Errorf("Expected duration to be 100ms, got '%v'", testFailing.Duration)
	}
	if testFailing.Error == nil {
		t.Fatal("Expected error to be non-nil")
	}
	if !strings.Contains(testFailing.Error.Message, "Expected 5, got 10") {
		t.Errorf("Expected error message to contain 'Expected 5, got 10', got '%s'", testFailing.Error.Message)
	}
}

// Test 1.1.5: Handle edge cases (panics, build failures, timeouts)
func TestParseGoTestEdgeCases(t *testing.T) {
	// Sample JSON output with edge cases
	jsonOutput := `
{"Time":"2023-05-15T12:00:00.1Z","Action":"run","Package":"github.com/user/project/pkg1","Test":"TestPanic"}
{"Time":"2023-05-15T12:00:00.15Z","Action":"output","Package":"github.com/user/project/pkg1","Test":"TestPanic","Output":"=== RUN   TestPanic\n"}
{"Time":"2023-05-15T12:00:00.2Z","Action":"output","Package":"github.com/user/project/pkg1","Test":"TestPanic","Output":"panic: runtime error: index out of range [1] with length 1\n"}
{"Time":"2023-05-15T12:00:00.25Z","Action":"output","Package":"github.com/user/project/pkg1","Test":"TestPanic","Output":"goroutine 8 [running]:\n"}
{"Time":"2023-05-15T12:00:00.3Z","Action":"output","Package":"github.com/user/project/pkg1","Test":"TestPanic","Output":"github.com/user/project/pkg1.TestPanic(0xc00012a000)\n\tpanic_test.go:15 +0x39\n"}
{"Time":"2023-05-15T12:00:00.33Z","Action":"output","Package":"github.com/user/project/pkg1","Test":"TestPanic","Output":"testing.tRunner(0xc00012a000, 0x10be430)\n\ttesting.go:1446 +0x10b\n"}
{"Time":"2023-05-15T12:00:00.35Z","Action":"fail","Package":"github.com/user/project/pkg1","Test":"TestPanic","Elapsed":0.2}
{"Time":"2023-05-15T12:00:00.4Z","Action":"output","Package":"github.com/user/project/pkg1","Output":"FAIL\tgithub.com/user/project/pkg1\t0.2s\n"}
{"Time":"2023-05-15T12:00:00.45Z","Action":"fail","Package":"github.com/user/project/pkg1","Elapsed":0.2}

{"Time":"2023-05-15T12:00:00.5Z","Action":"output","Package":"github.com/user/project/pkg2","Output":"# github.com/user/project/pkg2\npkg2/example.go:10:1: syntax error: unexpected semicolon or newline\n"}
{"Time":"2023-05-15T12:00:00.55Z","Action":"skip","Package":"github.com/user/project/pkg2","Elapsed":0.01}

{"Time":"2023-05-15T12:00:00.6Z","Action":"run","Package":"github.com/user/project/pkg3","Test":"TestTimeout"}
{"Time":"2023-05-15T12:00:00.65Z","Action":"output","Package":"github.com/user/project/pkg3","Test":"TestTimeout","Output":"=== RUN   TestTimeout\n"}
{"Time":"2023-05-15T12:00:10.65Z","Action":"output","Package":"github.com/user/project/pkg3","Test":"TestTimeout","Output":"panic: test timed out after 10s\n"}
{"Time":"2023-05-15T12:00:10.7Z","Action":"fail","Package":"github.com/user/project/pkg3","Test":"TestTimeout","Elapsed":10}
{"Time":"2023-05-15T12:00:10.75Z","Action":"output","Package":"github.com/user/project/pkg3","Output":"FAIL\tgithub.com/user/project/pkg3\t10s\n"}
{"Time":"2023-05-15T12:00:10.8Z","Action":"fail","Package":"github.com/user/project/pkg3","Elapsed":10}
`

	// Create parser
	parser := NewParser()
	if parser == nil {
		t.Fatal("Expected parser to be created")
	}

	// Parse the JSON output
	reader := strings.NewReader(jsonOutput)
	results, err := parser.Parse(reader)
	if err != nil {
		t.Fatalf("Error parsing test output: %v", err)
	}

	// Validate results
	if len(results) != 3 {
		t.Fatalf("Expected 3 packages, got %d", len(results))
	}

	// Check first package (panic)
	pkg1 := findPackage(results, "github.com/user/project/pkg1")
	if pkg1 == nil {
		t.Fatal("Could not find package github.com/user/project/pkg1")
	}

	if len(pkg1.Tests) != 1 {
		t.Fatalf("Expected 1 test in package 1, got %d", len(pkg1.Tests))
	}

	test1 := pkg1.Tests[0]
	if test1.Name != "TestPanic" {
		t.Errorf("Expected test name to be 'TestPanic', got '%s'", test1.Name)
	}

	if test1.Status != StatusFailed {
		t.Errorf("Expected test status to be 'failed', got '%s'", test1.Status)
	}

	if test1.Error == nil {
		t.Fatal("Expected error to be non-nil for panic test")
	}

	if !strings.Contains(test1.Error.Message, "panic: runtime error") {
		t.Errorf("Expected panic error message, got '%s'", test1.Error.Message)
	}

	// Check for location in all output
	panicLocation := strings.Contains(test1.Output, "panic_test.go:15")
	if !panicLocation {
		t.Errorf("Expected to find panic_test.go:15 in test output")
	}

	// Check second package (build failure)
	pkg2 := findPackage(results, "github.com/user/project/pkg2")
	if pkg2 == nil {
		t.Fatal("Could not find package github.com/user/project/pkg2")
	}

	if !pkg2.BuildFailed {
		t.Errorf("Expected BuildFailed to be true for package 2")
	}

	if !strings.Contains(pkg2.BuildError, "syntax error") {
		t.Errorf("Expected build error message, got '%s'", pkg2.BuildError)
	}

	// Check third package (timeout)
	pkg3 := findPackage(results, "github.com/user/project/pkg3")
	if pkg3 == nil {
		t.Fatal("Could not find package github.com/user/project/pkg3")
	}

	if len(pkg3.Tests) != 1 {
		t.Fatalf("Expected 1 test in package 3, got %d", len(pkg3.Tests))
	}

	test3 := pkg3.Tests[0]
	if test3.Name != "TestTimeout" {
		t.Errorf("Expected test name to be 'TestTimeout', got '%s'", test3.Name)
	}

	if test3.Status != StatusFailed {
		t.Errorf("Expected test status to be 'failed', got '%s'", test3.Status)
	}

	if test3.Error == nil || !strings.Contains(test3.Error.Message, "test timed out") {
		t.Errorf("Expected timeout error message, got '%+v'", test3.Error)
	}

	if test3.Duration != 10*time.Second {
		t.Errorf("Expected duration to be 10s, got '%v'", test3.Duration)
	}
}

// Helper function to find a package by name in the test results
func findPackage(packages []*TestPackage, name string) *TestPackage {
	for _, pkg := range packages {
		if pkg.Package == name {
			return pkg
		}
	}
	return nil
}
