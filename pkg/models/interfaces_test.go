package models

import (
	"testing"
	"time"
)

// TestNewPackageResult_FactoryFunction tests the NewPackageResult factory function
func TestNewPackageResult_FactoryFunction(t *testing.T) {
	t.Parallel()

	packageName := "github.com/example/pkg"
	result := NewPackageResult(packageName)

	if result == nil {
		t.Fatal("NewPackageResult should not return nil")
	}

	if result.Package != packageName {
		t.Errorf("Expected package %q, got %q", packageName, result.Package)
	}

	// Verify tests slice is initialized
	if result.Tests == nil {
		t.Error("Tests slice should be initialized")
	}

	// Verify metadata is initialized
	if result.Metadata == nil {
		t.Error("Metadata should be initialized")
	}
}

// TestNewTestSummary_FactoryFunction tests the NewTestSummary factory function
func TestNewTestSummary_FactoryFunction(t *testing.T) {
	t.Parallel()

	summary := NewTestSummary()

	if summary == nil {
		t.Fatal("NewTestSummary should not return nil")
	}

	// Verify FailedPackages slice is initialized
	if summary.FailedPackages == nil {
		t.Error("FailedPackages slice should be initialized")
	}

	// Verify metadata is initialized
	if summary.Metadata == nil {
		t.Error("Metadata should be initialized")
	}
}

// TestNewFileChange_FactoryFunction tests the NewFileChange factory function
func TestNewFileChange_FactoryFunction(t *testing.T) {
	t.Parallel()

	path := "/src/test.go"
	changeType := ChangeTypeModified

	change := NewFileChange(path, changeType)

	if change == nil {
		t.Fatal("NewFileChange should not return nil")
	}

	if change.FilePath != path {
		t.Errorf("Expected path %q, got %q", path, change.FilePath)
	}

	if change.ChangeType != changeType {
		t.Errorf("Expected change type %v, got %v", changeType, change.ChangeType)
	}

	// Verify timestamp is set and recent
	if change.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}

	if time.Since(change.Timestamp) > time.Second {
		t.Error("Timestamp should be recent")
	}

	// Verify metadata is initialized
	if change.Metadata == nil {
		t.Error("Metadata should be initialized")
	}
}

// TestPackageResult_GetSuccessRate tests the GetSuccessRate method
func TestPackageResult_GetSuccessRate(t *testing.T) {
	t.Parallel()

	result := NewPackageResult("test.pkg")

	// Test with no tests
	rate := result.GetSuccessRate()
	if rate != 0.0 {
		t.Errorf("Expected zero rate for no tests, got %f", rate)
	}

	// Test with some tests
	result.TestCount = 4
	result.PassedCount = 2
	rate = result.GetSuccessRate()
	expected := 0.5 // 50% as decimal
	if rate != expected {
		t.Errorf("Expected success rate %f, got %f", expected, rate)
	}
}

// TestPackageResult_AddTest tests the AddTest method
func TestPackageResult_AddTest(t *testing.T) {
	t.Parallel()

	result := NewPackageResult("test.pkg")
	test := NewTestResult("TestExample", "test.pkg")
	test.Status = TestStatusPassed

	// Verify initial state
	if len(result.Tests) != 0 {
		t.Errorf("Expected 0 tests initially, got %d", len(result.Tests))
	}

	result.AddTest(test)

	if len(result.Tests) != 1 {
		t.Errorf("Expected 1 test after adding, got %d", len(result.Tests))
	}

	if result.TestCount != 1 {
		t.Errorf("Expected TestCount 1, got %d", result.TestCount)
	}

	if result.PassedCount != 1 {
		t.Errorf("Expected PassedCount 1, got %d", result.PassedCount)
	}
}

// TestTestSummary_GetSuccessRate tests the TestSummary GetSuccessRate method
func TestTestSummary_GetSuccessRate(t *testing.T) {
	t.Parallel()

	summary := NewTestSummary()

	// Test with no tests
	rate := summary.GetSuccessRate()
	if rate != 0.0 {
		t.Errorf("Expected zero rate for no tests, got %f", rate)
	}

	// Test with some tests
	summary.TotalTests = 10
	summary.PassedTests = 7
	rate = summary.GetSuccessRate()
	expected := 0.7 // 70% as decimal
	if rate != expected {
		t.Errorf("Expected success rate %f, got %f", expected, rate)
	}
}

// TestTestSummary_AddPackageResult tests the AddPackageResult method
func TestTestSummary_AddPackageResult(t *testing.T) {
	t.Parallel()

	summary := NewTestSummary()
	pkg := NewPackageResult("pkg1")
	pkg.TestCount = 2
	pkg.PassedCount = 1
	pkg.FailedCount = 1
	pkg.Duration = 100 * time.Millisecond

	summary.AddPackageResult(pkg)

	if summary.PackageCount != 1 {
		t.Errorf("Expected PackageCount 1, got %d", summary.PackageCount)
	}

	if summary.TotalTests != 2 {
		t.Errorf("Expected TotalTests 2, got %d", summary.TotalTests)
	}

	if summary.PassedTests != 1 {
		t.Errorf("Expected PassedTests 1, got %d", summary.PassedTests)
	}
}

// TestChangeTypeConstants tests the ChangeType constants
func TestChangeTypeConstants(t *testing.T) {
	t.Parallel()

	expectedTypes := map[ChangeType]string{
		ChangeTypeCreated:  "created",
		ChangeTypeModified: "modified",
		ChangeTypeDeleted:  "deleted",
		ChangeTypeRenamed:  "renamed",
		ChangeTypeMoved:    "moved",
	}

	for changeType, expectedString := range expectedTypes {
		if string(changeType) != expectedString {
			t.Errorf("Expected ChangeType %v to equal %q, got %q", changeType, expectedString, string(changeType))
		}
	}
}

// TestTestStatusConstants tests the TestStatus constants
func TestTestStatusConstants(t *testing.T) {
	t.Parallel()

	expectedStatuses := map[TestStatus]string{
		TestStatusPending: "pending",
		TestStatusRunning: "running",
		TestStatusPassed:  "passed",
		TestStatusFailed:  "failed",
		TestStatusSkipped: "skipped",
		TestStatusTimeout: "timeout",
		TestStatusError:   "error",
	}

	for status, expectedString := range expectedStatuses {
		if string(status) != expectedString {
			t.Errorf("Expected TestStatus %v to equal %q, got %q", status, expectedString, string(status))
		}
	}
}

// TestPackageResult_AddTest_AllStatuses tests AddTest with all possible test statuses
func TestPackageResult_AddTest_AllStatuses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		status          TestStatus
		expectedPassed  int
		expectedFailed  int
		expectedSkipped int
		expectedSuccess bool
	}{
		{
			name:            "passed_test",
			status:          TestStatusPassed,
			expectedPassed:  1,
			expectedFailed:  0,
			expectedSkipped: 0,
			expectedSuccess: false, // Package success defaults to false and is only set to false on failures
		},
		{
			name:            "failed_test",
			status:          TestStatusFailed,
			expectedPassed:  0,
			expectedFailed:  1,
			expectedSkipped: 0,
			expectedSuccess: false,
		},
		{
			name:            "timeout_test",
			status:          TestStatusTimeout,
			expectedPassed:  0,
			expectedFailed:  1,
			expectedSkipped: 0,
			expectedSuccess: false,
		},
		{
			name:            "error_test",
			status:          TestStatusError,
			expectedPassed:  0,
			expectedFailed:  1,
			expectedSkipped: 0,
			expectedSuccess: false,
		},
		{
			name:            "skipped_test",
			status:          TestStatusSkipped,
			expectedPassed:  0,
			expectedFailed:  0,
			expectedSkipped: 1,
			expectedSuccess: false, // Package success defaults to false
		},
		{
			name:            "pending_test",
			status:          TestStatusPending,
			expectedPassed:  0,
			expectedFailed:  0,
			expectedSkipped: 0,
			expectedSuccess: false, // Package success defaults to false
		},
		{
			name:            "running_test",
			status:          TestStatusRunning,
			expectedPassed:  0,
			expectedFailed:  0,
			expectedSkipped: 0,
			expectedSuccess: false, // Package success defaults to false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := NewPackageResult("test.pkg")
			test := NewTestResult("TestExample", "test.pkg")
			test.Status = tt.status

			// Verify initial state
			if result.Success != false {
				t.Error("Package should start as unsuccessful by default")
			}

			result.AddTest(test)

			if len(result.Tests) != 1 {
				t.Errorf("Expected 1 test after adding, got %d", len(result.Tests))
			}

			if result.TestCount != 1 {
				t.Errorf("Expected TestCount 1, got %d", result.TestCount)
			}

			if result.PassedCount != tt.expectedPassed {
				t.Errorf("Expected PassedCount %d, got %d", tt.expectedPassed, result.PassedCount)
			}

			if result.FailedCount != tt.expectedFailed {
				t.Errorf("Expected FailedCount %d, got %d", tt.expectedFailed, result.FailedCount)
			}

			if result.SkippedCount != tt.expectedSkipped {
				t.Errorf("Expected SkippedCount %d, got %d", tt.expectedSkipped, result.SkippedCount)
			}

			if result.Success != tt.expectedSuccess {
				t.Errorf("Expected Success %t, got %t", tt.expectedSuccess, result.Success)
			}
		})
	}
}
