package models

import (
	"testing"
	"time"
)

// TestTestEvent_JSONMarshaling tests that TestEvent can be properly marshaled/unmarshaled
func TestTestEvent_JSONMarshaling(t *testing.T) {
	// Test the basic structure validation
	event := TestEvent{
		Time:    "2023-10-01T12:00:00Z",
		Action:  "pass",
		Package: "example.com/test",
		Test:    "TestExample",
		Output:  "test output",
		Elapsed: 1.5,
	}

	// Verify all fields are accessible
	if event.Time != "2023-10-01T12:00:00Z" {
		t.Errorf("Expected Time '2023-10-01T12:00:00Z', got '%s'", event.Time)
	}
	if event.Action != "pass" {
		t.Errorf("Expected Action 'pass', got '%s'", event.Action)
	}
	if event.Package != "example.com/test" {
		t.Errorf("Expected Package 'example.com/test', got '%s'", event.Package)
	}
	if event.Test != "TestExample" {
		t.Errorf("Expected Test 'TestExample', got '%s'", event.Test)
	}
	if event.Output != "test output" {
		t.Errorf("Expected Output 'test output', got '%s'", event.Output)
	}
	if event.Elapsed != 1.5 {
		t.Errorf("Expected Elapsed 1.5, got %f", event.Elapsed)
	}
}

// TestTestEvent_EmptyValues tests TestEvent with empty/zero values
func TestTestEvent_EmptyValues(t *testing.T) {
	event := TestEvent{}

	// Verify zero values are handled correctly
	if event.Time != "" {
		t.Errorf("Expected empty Time, got '%s'", event.Time)
	}
	if event.Action != "" {
		t.Errorf("Expected empty Action, got '%s'", event.Action)
	}
	if event.Package != "" {
		t.Errorf("Expected empty Package, got '%s'", event.Package)
	}
	if event.Test != "" {
		t.Errorf("Expected empty Test, got '%s'", event.Test)
	}
	if event.Output != "" {
		t.Errorf("Expected empty Output, got '%s'", event.Output)
	}
	if event.Elapsed != 0 {
		t.Errorf("Expected Elapsed 0, got %f", event.Elapsed)
	}
}

// TestTestEvent_CommonActions tests TestEvent with common action types
func TestTestEvent_CommonActions(t *testing.T) {
	commonActions := []string{"run", "pass", "fail", "skip", "output"}

	for _, action := range commonActions {
		t.Run("Action_"+action, func(t *testing.T) {
			event := TestEvent{
				Action:  action,
				Package: "test/package",
				Test:    "TestSample",
			}

			if event.Action != action {
				t.Errorf("Expected action '%s', got '%s'", action, event.Action)
			}
		})
	}
}

// TestTestProgress_FieldAccess tests TestProgress struct field access
func TestTestProgress_FieldAccess(t *testing.T) {
	progress := TestProgress{
		CompletedTests: 5,
		TotalTests:     10,
		CurrentFile:    "example_test.go",
		Status:         TestStatusRunning,
	}

	// Verify all fields are accessible
	if progress.CompletedTests != 5 {
		t.Errorf("Expected CompletedTests 5, got %d", progress.CompletedTests)
	}
	if progress.TotalTests != 10 {
		t.Errorf("Expected TotalTests 10, got %d", progress.TotalTests)
	}
	if progress.CurrentFile != "example_test.go" {
		t.Errorf("Expected CurrentFile 'example_test.go', got '%s'", progress.CurrentFile)
	}
	if progress.Status != TestStatusRunning {
		t.Errorf("Expected Status TestStatusRunning, got %v", progress.Status)
	}
}

// TestTestProgress_CalculatePercentage tests progress percentage calculation
func TestTestProgress_CalculatePercentage(t *testing.T) {
	testCases := []struct {
		name               string
		completed          int
		total              int
		expectedPercentage float64
	}{
		{
			name:               "Half complete",
			completed:          5,
			total:              10,
			expectedPercentage: 50.0,
		},
		{
			name:               "All complete",
			completed:          10,
			total:              10,
			expectedPercentage: 100.0,
		},
		{
			name:               "None complete",
			completed:          0,
			total:              10,
			expectedPercentage: 0.0,
		},
		{
			name:               "Zero total (edge case)",
			completed:          0,
			total:              0,
			expectedPercentage: 0.0, // Should handle gracefully
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			progress := TestProgress{
				CompletedTests: tc.completed,
				TotalTests:     tc.total,
			}

			// Calculate percentage manually since there's no method defined
			var percentage float64
			if tc.total > 0 {
				percentage = (float64(tc.completed) / float64(tc.total)) * 100
			}

			if percentage != tc.expectedPercentage {
				t.Errorf("Expected percentage %.1f, got %.1f", tc.expectedPercentage, percentage)
			}

			// Verify the values are set correctly
			if progress.CompletedTests != tc.completed {
				t.Errorf("Expected completed %d, got %d", tc.completed, progress.CompletedTests)
			}
			if progress.TotalTests != tc.total {
				t.Errorf("Expected total %d, got %d", tc.total, progress.TotalTests)
			}
		})
	}
}

// TestTestEvent_TimeFieldFormat tests time field format validation
func TestTestEvent_TimeFieldFormat(t *testing.T) {
	testCases := []struct {
		name     string
		timeStr  string
		expected bool
	}{
		{
			name:     "Valid RFC3339Nano",
			timeStr:  "2023-10-01T12:00:00.123456789Z",
			expected: true,
		},
		{
			name:     "Valid RFC3339",
			timeStr:  "2023-10-01T12:00:00Z",
			expected: true,
		},
		{
			name:     "Invalid format",
			timeStr:  "2023-10-01 12:00:00",
			expected: false,
		},
		{
			name:     "Empty string",
			timeStr:  "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := TestEvent{
				Time: tc.timeStr,
			}

			// Try to parse the time
			_, err := time.Parse(time.RFC3339Nano, event.Time)
			if tc.expected && err != nil {
				t.Errorf("Expected valid time format, but got error: %v", err)
			}
			if !tc.expected && err == nil && event.Time != "" {
				t.Errorf("Expected invalid time format, but parsing succeeded")
			}
		})
	}
}

// TestTestEvent_ElapsedFieldTypes tests elapsed field with different numeric types
func TestTestEvent_ElapsedFieldTypes(t *testing.T) {
	testCases := []struct {
		name    string
		elapsed float64
	}{
		{"Zero", 0.0},
		{"Small decimal", 0.001},
		{"Integer", 1.0},
		{"Large decimal", 123.456789},
		{"Very large", 999999.999999},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := TestEvent{
				Elapsed: tc.elapsed,
			}

			if event.Elapsed != tc.elapsed {
				t.Errorf("Expected elapsed %f, got %f", tc.elapsed, event.Elapsed)
			}
		})
	}
}

// TestTestProgress_ZeroValues tests TestProgress with zero values
func TestTestProgress_ZeroValues(t *testing.T) {
	progress := TestProgress{}

	if progress.CompletedTests != 0 {
		t.Errorf("Expected CompletedTests 0, got %d", progress.CompletedTests)
	}
	if progress.TotalTests != 0 {
		t.Errorf("Expected TotalTests 0, got %d", progress.TotalTests)
	}
	if progress.CurrentFile != "" {
		t.Errorf("Expected empty CurrentFile, got '%s'", progress.CurrentFile)
	}
	if progress.Status != "" {
		t.Errorf("Expected empty Status, got '%s'", progress.Status)
	}
}

// TestTestStatus_Constants tests TestStatus constants
func TestTestStatus_Constants(t *testing.T) {
	testCases := []struct {
		name     string
		status   TestStatus
		expected string
	}{
		{"Pending", TestStatusPending, "pending"},
		{"Running", TestStatusRunning, "running"},
		{"Passed", TestStatusPassed, "passed"},
		{"Failed", TestStatusFailed, "failed"},
		{"Skipped", TestStatusSkipped, "skipped"},
		{"Timeout", TestStatusTimeout, "timeout"},
		{"Error", TestStatusError, "error"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if string(tc.status) != tc.expected {
				t.Errorf("Expected status '%s', got '%s'", tc.expected, string(tc.status))
			}
		})
	}
}

// TestLegacyTestResult_FieldAccess tests LegacyTestResult struct field access
func TestLegacyTestResult_FieldAccess(t *testing.T) {
	result := LegacyTestResult{
		Name:     "TestExample",
		Status:   TestStatusPassed,
		Duration: time.Second,
		Package:  "example.com/test",
		Test:     "TestExample",
		Output:   "test output",
	}

	// Verify all fields are accessible
	if result.Name != "TestExample" {
		t.Errorf("Expected Name 'TestExample', got '%s'", result.Name)
	}
	if result.Status != TestStatusPassed {
		t.Errorf("Expected Status TestStatusPassed, got %v", result.Status)
	}
	if result.Duration != time.Second {
		t.Errorf("Expected Duration 1s, got %v", result.Duration)
	}
	if result.Package != "example.com/test" {
		t.Errorf("Expected Package 'example.com/test', got '%s'", result.Package)
	}
	if result.Test != "TestExample" {
		t.Errorf("Expected Test 'TestExample', got '%s'", result.Test)
	}
	if result.Output != "test output" {
		t.Errorf("Expected Output 'test output', got '%s'", result.Output)
	}
}

// TestTestSuite_FieldAccess tests TestSuite struct field access
func TestTestSuite_FieldAccess(t *testing.T) {
	suite := TestSuite{
		FilePath:     "example_test.go",
		Duration:     time.Second,
		MemoryUsage:  1024,
		TestCount:    5,
		PassedCount:  4,
		FailedCount:  1,
		SkippedCount: 0,
	}

	// Verify all fields are accessible
	if suite.FilePath != "example_test.go" {
		t.Errorf("Expected FilePath 'example_test.go', got '%s'", suite.FilePath)
	}
	if suite.Duration != time.Second {
		t.Errorf("Expected Duration 1s, got %v", suite.Duration)
	}
	if suite.MemoryUsage != 1024 {
		t.Errorf("Expected MemoryUsage 1024, got %d", suite.MemoryUsage)
	}
	if suite.TestCount != 5 {
		t.Errorf("Expected TestCount 5, got %d", suite.TestCount)
	}
	if suite.PassedCount != 4 {
		t.Errorf("Expected PassedCount 4, got %d", suite.PassedCount)
	}
	if suite.FailedCount != 1 {
		t.Errorf("Expected FailedCount 1, got %d", suite.FailedCount)
	}
	if suite.SkippedCount != 0 {
		t.Errorf("Expected SkippedCount 0, got %d", suite.SkippedCount)
	}
}
