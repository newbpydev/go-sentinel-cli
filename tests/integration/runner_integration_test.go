package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/runner"
)

// TestBasicTestRunner_Integration_ActualExecution tests real command execution
func TestBasicTestRunner_Integration_ActualExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Parallel()

	// Create a temporary test directory with a simple test
	tempDir := t.TempDir()

	// Create a simple Go module
	goModContent := `module testmod

go 1.21
`
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a simple test file
	testContent := `package main

import "testing"

func TestSimple(t *testing.T) {
	// Simple passing test
	if 1+1 != 2 {
		t.Error("math is broken")
	}
}
`
	if err := os.WriteFile(filepath.Join(tempDir, "simple_test.go"), []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Change to temp directory for test execution
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create runner
	r := runner.NewBasicTestRunner(true, false)

	// Create context with timeout to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Act
	result, err := r.Run(ctx, []string{"."})

	// Assert
	if err != nil {
		t.Logf("Test execution returned error (may be expected): %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result from test execution")
	}

	// Should contain some indication of test execution
	if !strings.Contains(result, "TestSimple") && !strings.Contains(result, "PASS") {
		t.Logf("Test result: %s", result)
		t.Error("Expected result to contain test execution information")
	}

	t.Logf("Integration test completed successfully with result length: %d", len(result))
}

// TestBasicTestRunner_Integration_JSONOutput tests JSON output mode
func TestBasicTestRunner_Integration_JSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Parallel()

	// Create a temporary test directory
	tempDir := t.TempDir()

	// Create a simple Go module
	goModContent := `module testmod

go 1.21
`
	if err := os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a simple test file
	testContent := `package main

import "testing"

func TestJSON(t *testing.T) {
	// Simple passing test for JSON output
	t.Log("This is a test log")
}
`
	if err := os.WriteFile(filepath.Join(tempDir, "json_test.go"), []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Change to temp directory for test execution
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(oldDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create runner with JSON output
	r := runner.NewBasicTestRunner(false, true)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Act
	result, err := r.Run(ctx, []string{"."})

	// Assert
	if err != nil {
		t.Logf("Test execution returned error (may be expected): %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result from JSON test execution")
	}

	// JSON output should contain specific fields
	if !strings.Contains(result, `"Action"`) && !strings.Contains(result, `"Test"`) {
		t.Logf("JSON result: %s", result)
		t.Error("Expected JSON output to contain Action and Test fields")
	}

	t.Logf("JSON integration test completed successfully")
}

// TestBasicTestRunner_Integration_Timeout tests command timeout handling
func TestBasicTestRunner_Integration_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	t.Parallel()

	// Create runner
	r := runner.NewBasicTestRunner(false, false)

	// Create context with very short timeout (longer than 1ms to ensure command starts)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// Act - this should timeout quickly
	result, err := r.Run(ctx, []string{"."})

	// Assert - timeout might happen before error is returned, which is also valid
	if err == nil {
		t.Logf("Command completed before timeout (very fast system)")
	}

	if result != "" {
		t.Error("Expected empty result on timeout")
	}

	// Should contain timeout-related error
	if err != nil && (!strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "timeout")) {
		t.Logf("Timeout error: %v", err)
		t.Error("Expected timeout-related error message")
	}

	t.Log("Timeout integration test completed successfully")
}
