package cli

import (
	"strings"
	"testing"
	"time"
)

func TestParser_Parse(t *testing.T) {
	// Sample go test -json output
	input := `
{"Time":"2024-01-20T10:00:00Z","Action":"start","Package":"example.com/pkg/foo"}
{"Time":"2024-01-20T10:00:00.1Z","Action":"start","Package":"example.com/pkg/foo","Test":"TestExample"}
{"Time":"2024-01-20T10:00:00.2Z","Action":"output","Package":"example.com/pkg/foo","Test":"TestExample","Output":"=== RUN   TestExample\n"}
{"Time":"2024-01-20T10:00:00.3Z","Action":"output","Package":"example.com/pkg/foo","Test":"TestExample","Output":"    foo_test.go:42: Error: expected 42 but got 24\n"}
{"Time":"2024-01-20T10:00:00.4Z","Action":"fail","Package":"example.com/pkg/foo","Test":"TestExample","Elapsed":0.3}
{"Time":"2024-01-20T10:00:00.5Z","Action":"start","Package":"example.com/pkg/foo","Test":"TestPass"}
{"Time":"2024-01-20T10:00:00.6Z","Action":"pass","Package":"example.com/pkg/foo","Test":"TestPass","Elapsed":0.1}
{"Time":"2024-01-20T10:00:00.7Z","Action":"start","Package":"example.com/pkg/foo","Test":"TestSkip"}
{"Time":"2024-01-20T10:00:00.8Z","Action":"skip","Package":"example.com/pkg/foo","Test":"TestSkip","Elapsed":0.1}
`

	parser := NewParser()
	run, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Verify test run statistics
	if run.NumTotal != 3 {
		t.Errorf("NumTotal = %d, want 3", run.NumTotal)
	}
	if run.NumPassed != 1 {
		t.Errorf("NumPassed = %d, want 1", run.NumPassed)
	}
	if run.NumFailed != 1 {
		t.Errorf("NumFailed = %d, want 1", run.NumFailed)
	}
	if run.NumSkipped != 1 {
		t.Errorf("NumSkipped = %d, want 1", run.NumSkipped)
	}

	// Verify test suites
	if len(run.Suites) != 1 {
		t.Fatalf("got %d suites, want 1", len(run.Suites))
	}

	suite := run.Suites[0]
	if suite.PackageName != "example.com/pkg/foo" {
		t.Errorf("PackageName = %s, want example.com/pkg/foo", suite.PackageName)
	}
	if suite.FilePath != "pkg/foo/foo_test.go" {
		t.Errorf("FilePath = %s, want pkg/foo/foo_test.go", suite.FilePath)
	}

	// Verify individual tests
	if len(suite.Tests) != 3 {
		t.Fatalf("got %d tests, want 3", len(suite.Tests))
	}

	// Find and verify the failed test
	var failedTest *TestResult
	for _, test := range suite.Tests {
		if test.Name == "TestExample" {
			failedTest = test
			break
		}
	}

	if failedTest == nil {
		t.Fatal("TestExample not found")
	}
	if failedTest.Status != TestStatusFailed {
		t.Errorf("TestExample status = %v, want failed", failedTest.Status)
	}
	if failedTest.Error == nil {
		t.Fatal("TestExample error is nil")
	}
	if failedTest.Error.Location.File != "foo_test.go" {
		t.Errorf("Error location file = %s, want foo_test.go", failedTest.Error.Location.File)
	}
	if failedTest.Error.Location.Line != 42 {
		t.Errorf("Error location line = %d, want 42", failedTest.Error.Location.Line)
	}
}

func TestParser_ExtractSourceLocation(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		wantFile string
		wantLine int
		wantCol  int
	}{
		{
			name:     "basic error",
			output:   "foo_test.go:42: test failed",
			wantFile: "foo_test.go",
			wantLine: 42,
			wantCol:  0,
		},
		{
			name:     "error with column",
			output:   "foo_test.go:42:12: test failed",
			wantFile: "foo_test.go",
			wantLine: 42,
			wantCol:  12,
		},
		{
			name:     "full path",
			output:   "/path/to/foo_test.go:42: test failed",
			wantFile: "/path/to/foo_test.go",
			wantLine: 42,
			wantCol:  0,
		},
		{
			name:     "with indentation",
			output:   "    foo_test.go:42: test failed",
			wantFile: "foo_test.go",
			wantLine: 42,
			wantCol:  0,
		},
	}

	parser := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := parser.extractSourceLocation(tt.output)
			if loc == nil {
				t.Fatal("extractSourceLocation returned nil")
			}
			if loc.File != tt.wantFile {
				t.Errorf("File = %s, want %s", loc.File, tt.wantFile)
			}
			if loc.Line != tt.wantLine {
				t.Errorf("Line = %d, want %d", loc.Line, tt.wantLine)
			}
			if loc.Column != tt.wantCol {
				t.Errorf("Column = %d, want %d", loc.Column, tt.wantCol)
			}
		})
	}
}

func TestParser_GetTestFilePath(t *testing.T) {
	tests := []struct {
		pkg  string
		want string
	}{
		{
			pkg:  "example.com/pkg/foo",
			want: "pkg/foo/foo_test.go",
		},
		{
			pkg:  "foo",
			want: "foo_test.go",
		},
		{
			pkg:  "github.com/user/project/internal/pkg/bar",
			want: "pkg/bar/bar_test.go",
		},
	}

	parser := NewParser()
	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			got := parser.getTestFilePath(tt.pkg)
			if got != tt.want {
				t.Errorf("getTestFilePath(%q) = %q, want %q", tt.pkg, got, tt.want)
			}
		})
	}
}

func TestParser_ProcessEvent_Timing(t *testing.T) {
	parser := NewParser()
	now := time.Now()

	// Initialize test run
	parser.currentRun = &TestRun{
		Suites:     make([]*TestSuite, 0),
		StartTime:  now,
		NumTotal:   0,
		NumPassed:  0,
		NumFailed:  0,
		NumSkipped: 0,
	}

	// Start package test
	if err := parser.processEvent(&GoTestEvent{
		Time:    now,
		Action:  "start",
		Package: "example/pkg",
	}); err != nil {
		t.Fatalf("Failed to process start event: %v", err)
	}

	// Start test
	if err := parser.processEvent(&GoTestEvent{
		Time:    now.Add(100 * time.Millisecond),
		Action:  "run",
		Package: "example/pkg",
		Test:    "TestExample",
	}); err != nil {
		t.Fatalf("Failed to process run event: %v", err)
	}

	// Pass test
	if err := parser.processEvent(&GoTestEvent{
		Time:    now.Add(200 * time.Millisecond),
		Action:  "pass",
		Package: "example/pkg",
		Test:    "TestExample",
		Elapsed: 0.1,
	}); err != nil {
		t.Fatalf("Failed to process pass event: %v", err)
	}

	// Verify timing
	test := parser.findTest("TestExample")
	if test == nil {
		t.Fatal("Test not found")
	}

	if test.StartTime != now.Add(100*time.Millisecond) {
		t.Errorf("StartTime = %v, want %v", test.StartTime, now.Add(100*time.Millisecond))
	}

	if test.EndTime != now.Add(200*time.Millisecond) {
		t.Errorf("EndTime = %v, want %v", test.EndTime, now.Add(200*time.Millisecond))
	}

	if test.Duration != 100*time.Millisecond {
		t.Errorf("Duration = %v, want %v", test.Duration, 100*time.Millisecond)
	}
}
