package cli

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTestResultSerialization(t *testing.T) {
	// Create a sample test result with all fields populated
	now := time.Now()
	result := &TestResult{
		Name:     "TestExample",
		Status:   TestStatusFailed,
		Duration: 2 * time.Second,
		Error: &TestError{
			Message: "expected 42 but got 24",
			Location: &SourceLocation{
				File:      "internal/example/example_test.go",
				Line:      42,
				Column:    12,
				Snippet:   "func TestExample(t *testing.T) {\n\tresult := compute()\n\tassert.Equal(t, 42, result)\n}",
				StartLine: 40,
			},
			Expected: "42",
			Actual:   "24",
		},
		StartTime: now,
		EndTime:   now.Add(2 * time.Second),
	}

	// Test JSON serialization
	t.Run("JSON marshaling", func(t *testing.T) {
		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("Failed to marshal TestResult: %v", err)
		}

		// Unmarshal back into a new struct
		var decoded TestResult
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal TestResult: %v", err)
		}

		// Verify fields were preserved
		if decoded.Name != result.Name {
			t.Errorf("Name mismatch: got %s, want %s", decoded.Name, result.Name)
		}
		if decoded.Status != result.Status {
			t.Errorf("Status mismatch: got %v, want %v", decoded.Status, result.Status)
		}
		if decoded.Duration != result.Duration {
			t.Errorf("Duration mismatch: got %v, want %v", decoded.Duration, result.Duration)
		}
		if decoded.Error == nil {
			t.Fatal("Expected Error to be non-nil")
		}
		if decoded.Error.Message != result.Error.Message {
			t.Errorf("Error message mismatch: got %s, want %s", decoded.Error.Message, result.Error.Message)
		}
	})
}

func TestTestSuiteSummary(t *testing.T) {
	// Create a sample test suite with mixed results
	now := time.Now()
	suite := &TestSuite{
		Package:     "example",
		PackageName: "github.com/newbpydev/go-sentinel/internal/example",
		FilePath:    "internal/example/example_test.go",
		Tests: []*TestResult{
			{
				Name:      "TestPass",
				Status:    TestStatusPassed,
				Duration:  1 * time.Second,
				StartTime: now,
				EndTime:   now.Add(1 * time.Second),
			},
			{
				Name:      "TestFail",
				Status:    TestStatusFailed,
				Duration:  2 * time.Second,
				StartTime: now.Add(1 * time.Second),
				EndTime:   now.Add(3 * time.Second),
				Error: &TestError{
					Message: "test failed",
				},
			},
			{
				Name:      "TestSkip",
				Status:    TestStatusSkipped,
				Duration:  100 * time.Millisecond,
				StartTime: now.Add(3 * time.Second),
				EndTime:   now.Add(3100 * time.Millisecond),
			},
		},
		StartTime: now,
		EndTime:   now.Add(3100 * time.Millisecond),
	}

	// Calculate summary
	suite.NumTotal = len(suite.Tests)
	for _, test := range suite.Tests {
		switch test.Status {
		case TestStatusPassed:
			suite.NumPassed++
		case TestStatusFailed:
			suite.NumFailed++
		case TestStatusSkipped:
			suite.NumSkipped++
		}
	}
	suite.Duration = suite.EndTime.Sub(suite.StartTime)

	// Verify summary calculations
	t.Run("summary calculations", func(t *testing.T) {
		if suite.NumTotal != 3 {
			t.Errorf("NumTotal: got %d, want 3", suite.NumTotal)
		}
		if suite.NumPassed != 1 {
			t.Errorf("NumPassed: got %d, want 1", suite.NumPassed)
		}
		if suite.NumFailed != 1 {
			t.Errorf("NumFailed: got %d, want 1", suite.NumFailed)
		}
		if suite.NumSkipped != 1 {
			t.Errorf("NumSkipped: got %d, want 1", suite.NumSkipped)
		}
		if suite.Duration != 3100*time.Millisecond {
			t.Errorf("Duration: got %v, want %v", suite.Duration, 3100*time.Millisecond)
		}
	})
}

func TestTestRunAggregation(t *testing.T) {
	now := time.Now()
	run := &TestRun{
		Suites: []*TestSuite{
			{
				Package:  "pkg1",
				FilePath: "pkg1/pkg1_test.go",
				Tests: []*TestResult{
					{Status: TestStatusPassed},
					{Status: TestStatusPassed},
				},
				NumTotal:  2,
				NumPassed: 2,
				Duration:  1 * time.Second,
			},
			{
				Package:  "pkg2",
				FilePath: "pkg2/pkg2_test.go",
				Tests: []*TestResult{
					{Status: TestStatusFailed},
					{Status: TestStatusSkipped},
				},
				NumTotal:   2,
				NumFailed:  1,
				NumSkipped: 1,
				Duration:   2 * time.Second,
			},
		},
		StartTime: now,
		EndTime:   now.Add(3 * time.Second),
	}

	// Calculate totals
	for _, suite := range run.Suites {
		run.NumTotal += suite.NumTotal
		run.NumPassed += suite.NumPassed
		run.NumFailed += suite.NumFailed
		run.NumSkipped += suite.NumSkipped
	}
	run.Duration = run.EndTime.Sub(run.StartTime)

	// Verify aggregation
	t.Run("run aggregation", func(t *testing.T) {
		if run.NumTotal != 4 {
			t.Errorf("NumTotal: got %d, want 4", run.NumTotal)
		}
		if run.NumPassed != 2 {
			t.Errorf("NumPassed: got %d, want 2", run.NumPassed)
		}
		if run.NumFailed != 1 {
			t.Errorf("NumFailed: got %d, want 1", run.NumFailed)
		}
		if run.NumSkipped != 1 {
			t.Errorf("NumSkipped: got %d, want 1", run.NumSkipped)
		}
		if run.Duration != 3*time.Second {
			t.Errorf("Duration: got %v, want %v", run.Duration, 3*time.Second)
		}
	})
}
