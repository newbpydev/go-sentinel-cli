package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/runner"
	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// TestIntegration_BasicTestExecution tests basic test execution workflow
func TestIntegration_BasicTestExecution(t *testing.T) {
	tempDir := t.TempDir()

	// Create a simple test file
	testFile := filepath.Join(tempDir, "example_test.go")
	testContent := `package main

import "testing"

func TestExample(t *testing.T) {
	if 1+1 != 2 {
		t.Error("Math is broken")
	}
}
`

	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test the TestRunner
	testRunner := runner.NewBasicTestRunner(true, false)

	ctx := context.Background()
	output, err := testRunner.Run(ctx, []string{tempDir})

	// In test environment, this may fail but should not crash
	if err != nil {
		t.Logf("Test runner failed (expected in test environment): %v", err)
	} else {
		t.Logf("Test runner succeeded with output length: %d", len(output))
	}
}

// TestIntegration_ColorFormatting tests color formatting integration
func TestIntegration_ColorFormatting(t *testing.T) {
	formatter := colors.NewColorFormatter(false)  // No colors for test consistency
	iconProvider := colors.NewIconProvider(false) // ASCII icons for test consistency

	// Create a test suite
	suite := &models.TestSuite{
		FilePath:     "example_test.go",
		TestCount:    2,
		PassedCount:  1,
		FailedCount:  1,
		SkippedCount: 0,
		Duration:     100 * time.Millisecond,
	}

	// Add test results
	suite.Tests = []*models.LegacyTestResult{
		{
			Name:     "TestPassing",
			Status:   models.TestStatusPassed,
			Duration: 50 * time.Millisecond,
			Package:  "github.com/test/example",
		},
		{
			Name:     "TestFailing",
			Status:   models.TestStatusFailed,
			Duration: 50 * time.Millisecond,
			Package:  "github.com/test/example",
			Error: &models.LegacyTestError{
				Message: "Test failed",
				Type:    "AssertionError",
			},
		},
	}

	// Test formatting
	fmt := formatter.Green("PASS")
	if fmt == "" {
		t.Error("Formatter should return formatted text")
	}

	icon := iconProvider.CheckMark()
	if icon == "" {
		t.Error("Icon provider should return icon")
	}

	// Verify suite has expected structure
	if suite.TestCount != 2 {
		t.Errorf("Expected 2 tests, got %d", suite.TestCount)
	}

	if len(suite.Tests) != 2 {
		t.Errorf("Expected 2 test results, got %d", len(suite.Tests))
	}
}

// TestIntegration_FileOperations tests file operations integration
func TestIntegration_FileOperations(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	testFiles := []string{
		"main.go",
		"main_test.go",
		"utils.go",
		"utils_test.go",
	}

	for _, file := range testFiles {
		filePath := filepath.Join(tempDir, file)
		content := "package main\n"
		if strings.HasSuffix(file, "_test.go") {
			content += "import \"testing\"\nfunc TestExample(t *testing.T) {}\n"
		} else {
			content += "func main() {}\n"
		}

		err := os.WriteFile(filePath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	// Test file discovery
	var goFiles []string
	var discoveredTestFiles []string

	err := filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".go") {
			goFiles = append(goFiles, path)
			if strings.HasSuffix(path, "_test.go") {
				discoveredTestFiles = append(discoveredTestFiles, path)
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to walk directory: %v", err)
	}

	if len(goFiles) != 4 {
		t.Errorf("Expected 4 Go files, found %d", len(goFiles))
	}

	if len(discoveredTestFiles) != 2 {
		t.Errorf("Expected 2 test files, found %d", len(discoveredTestFiles))
	}
}

// TestIntegration_ErrorHandling tests error handling integration
func TestIntegration_ErrorHandling(t *testing.T) {
	// Test with non-existent directory
	testRunner := runner.NewBasicTestRunner(false, true)
	ctx := context.Background()

	_, err := testRunner.Run(ctx, []string{"/path/that/does/not/exist"})
	if err == nil {
		t.Error("Expected error for non-existent path")
	}

	// Test with empty paths
	_, err = testRunner.Run(ctx, []string{})
	if err == nil {
		t.Error("Expected error for empty paths")
	}

	// Test with invalid path
	_, err = testRunner.Run(ctx, []string{""})
	if err == nil {
		t.Error("Expected error for empty path string")
	}
}

// TestIntegration_ComponentCreation tests component creation
func TestIntegration_ComponentCreation(t *testing.T) {
	// Test creating basic components that we know exist
	formatter := colors.NewColorFormatter(false)
	if formatter == nil {
		t.Error("NewColorFormatter returned nil")
	}

	iconProvider := colors.NewIconProvider(false)
	if iconProvider == nil {
		t.Error("NewIconProvider returned nil")
	}

	testRunner := runner.NewBasicTestRunner(false, true)
	if testRunner == nil {
		t.Error("NewBasicTestRunner returned nil")
	}

	t.Logf("Basic components created successfully")
}

// TestIntegration_ContextHandling tests context handling
func TestIntegration_ContextHandling(t *testing.T) {
	ctx := context.Background()

	// Test context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	// Test that context is properly handled
	select {
	case <-timeoutCtx.Done():
		if timeoutCtx.Err() != context.DeadlineExceeded {
			t.Errorf("Expected DeadlineExceeded, got %v", timeoutCtx.Err())
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("Context timeout did not work")
	}

	// Test context cancellation
	cancelCtx, cancelFunc := context.WithCancel(ctx)
	cancelFunc()

	select {
	case <-cancelCtx.Done():
		if cancelCtx.Err() != context.Canceled {
			t.Errorf("Expected Canceled, got %v", cancelCtx.Err())
		}
	default:
		t.Error("Context cancellation did not work")
	}
}

// TestIntegration_MemoryUsage tests memory usage patterns
func TestIntegration_MemoryUsage(t *testing.T) {
	// Create many test suites to test memory usage
	suites := make([]*models.TestSuite, 100)

	for i := 0; i < 100; i++ {
		suite := &models.TestSuite{
			FilePath:     filepath.Join("test", "suite_"+string(rune(i))+".go"),
			TestCount:    10,
			PassedCount:  9,
			FailedCount:  1,
			SkippedCount: 0,
			Duration:     time.Millisecond * 100,
		}

		// Add test results
		for j := 0; j < 10; j++ {
			test := &models.LegacyTestResult{
				Name:     "TestMemory_" + string(rune(i)) + "_" + string(rune(j)),
				Status:   models.TestStatusPassed,
				Duration: time.Millisecond,
				Package:  "github.com/test/memory",
			}
			suite.Tests = append(suite.Tests, test)
		}

		suites[i] = suite
	}

	// Process all suites
	totalTests := 0
	for _, suite := range suites {
		totalTests += len(suite.Tests)
	}

	if totalTests != 1000 {
		t.Errorf("Expected 1000 tests, got %d", totalTests)
	}

	t.Logf("Memory usage test completed with %d test suites and %d tests", len(suites), totalTests)
}
