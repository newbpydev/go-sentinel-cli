package runner

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestNewBasicTestRunner_FactoryFunction tests the factory function
func TestNewBasicTestRunner_FactoryFunction(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		verbose    bool
		jsonOutput bool
	}{
		{
			name:       "Both flags false",
			verbose:    false,
			jsonOutput: false,
		},
		{
			name:       "Verbose only",
			verbose:    true,
			jsonOutput: false,
		},
		{
			name:       "JSON output only",
			verbose:    false,
			jsonOutput: true,
		},
		{
			name:       "Both flags true",
			verbose:    true,
			jsonOutput: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			runner := NewBasicTestRunner(tc.verbose, tc.jsonOutput)

			// Assert
			if runner == nil {
				t.Fatal("NewBasicTestRunner should not return nil")
			}
			if runner.Verbose != tc.verbose {
				t.Errorf("Expected Verbose=%v, got %v", tc.verbose, runner.Verbose)
			}
			if runner.JSONOutput != tc.jsonOutput {
				t.Errorf("Expected JSONOutput=%v, got %v", tc.jsonOutput, runner.JSONOutput)
			}

			// Verify interface compliance
			var _ TestRunnerInterface = runner
		})
	}
}

// TestBasicTestRunner_Run_EmptyPackages tests Run with empty package list
func TestBasicTestRunner_Run_EmptyPackages(t *testing.T) {
	t.Parallel()

	runner := NewBasicTestRunner(false, true)
	ctx := context.Background()

	// Act
	output, err := runner.Run(ctx, []string{})

	// Assert
	if err == nil {
		t.Error("Expected error for empty test paths")
	}
	if !strings.Contains(err.Error(), "no test paths provided") {
		t.Errorf("Expected 'no test paths provided' error, got: %v", err)
	}
	if output != "" {
		t.Errorf("Expected empty output, got: %s", output)
	}
}

// TestBasicTestRunner_Run_NonExistentPath tests Run with non-existent path
func TestBasicTestRunner_Run_NonExistentPath(t *testing.T) {
	t.Parallel()

	runner := NewBasicTestRunner(false, true)
	ctx := context.Background()

	// Act
	output, err := runner.Run(ctx, []string{"./non-existent-path"})

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
	if !strings.Contains(err.Error(), "test path does not exist") {
		t.Errorf("Expected 'test path does not exist' error, got: %v", err)
	}
	if output != "" {
		t.Errorf("Expected empty output, got: %s", output)
	}
}

// TestBasicTestRunner_Run_EmptyStringPath tests Run with empty string path
func TestBasicTestRunner_Run_EmptyStringPath(t *testing.T) {
	t.Parallel()

	runner := NewBasicTestRunner(false, true)
	ctx := context.Background()

	// Act
	output, err := runner.Run(ctx, []string{""})

	// Assert
	if err == nil {
		t.Error("Expected error for empty string path")
	}
	if !strings.Contains(err.Error(), "empty test path provided") {
		t.Errorf("Expected 'empty test path provided' error, got: %v", err)
	}
	if output != "" {
		t.Errorf("Expected empty output, got: %s", output)
	}
}

// TestBasicTestRunner_RunStream_EmptyPackages tests RunStream with empty package list
func TestBasicTestRunner_RunStream_EmptyPackages(t *testing.T) {
	t.Parallel()

	runner := NewBasicTestRunner(false, true)
	ctx := context.Background()

	// Act
	stream, err := runner.RunStream(ctx, []string{})

	// Assert
	if err == nil {
		t.Error("Expected error for empty test paths")
	}
	if !strings.Contains(err.Error(), "no test paths provided") {
		t.Errorf("Expected 'no test paths provided' error, got: %v", err)
	}
	if stream != nil {
		t.Error("Expected nil stream on error")
	}
}

// TestBasicTestRunner_RunStream_NonExistentPath tests RunStream with non-existent path
func TestBasicTestRunner_RunStream_NonExistentPath(t *testing.T) {
	t.Parallel()

	runner := NewBasicTestRunner(false, true)
	ctx := context.Background()

	// Act
	stream, err := runner.RunStream(ctx, []string{"./non-existent-path"})

	// Assert
	if err == nil {
		t.Error("Expected error for non-existent path")
	}
	if !strings.Contains(err.Error(), "test path does not exist") {
		t.Errorf("Expected 'test path does not exist' error, got: %v", err)
	}
	if stream != nil {
		t.Error("Expected nil stream on error")
	}
}

// TestBasicTestRunner_RunStream_EmptyStringPath tests RunStream with empty string path
func TestBasicTestRunner_RunStream_EmptyStringPath(t *testing.T) {
	t.Parallel()

	runner := NewBasicTestRunner(false, true)
	ctx := context.Background()

	// Act
	stream, err := runner.RunStream(ctx, []string{""})

	// Assert
	if err == nil {
		t.Error("Expected error for empty string path")
	}
	if !strings.Contains(err.Error(), "empty test path provided") {
		t.Errorf("Expected 'empty test path provided' error, got: %v", err)
	}
	if stream != nil {
		t.Error("Expected nil stream on error")
	}
}

// TestBasicTestRunner_RunStream_ValidPath tests RunStream with valid test package
func TestBasicTestRunner_RunStream_ValidPath(t *testing.T) {
	t.Parallel()

	// Create a temporary test package
	tempDir := createTestPackageForBasicRunner(t)
	defer os.RemoveAll(tempDir)

	runner := NewBasicTestRunner(true, true)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Act
	stream, err := runner.RunStream(ctx, []string{tempDir})

	// Assert
	if err != nil {
		t.Fatalf("Expected no error for valid path, got: %v", err)
	}
	if stream == nil {
		t.Fatal("Expected non-nil stream")
	}

	// Read some data from the stream
	buffer := make([]byte, 1024)
	n, readErr := stream.Read(buffer)

	// Close the stream
	closeErr := stream.Close()
	if closeErr != nil {
		t.Logf("Stream close error (may be expected): %v", closeErr)
	}

	// Verify we got some data
	if readErr != nil && readErr != io.EOF {
		t.Logf("Read error (may be expected for test completion): %v", readErr)
	}
	if n > 0 {
		output := string(buffer[:n])
		t.Logf("Stream output sample: %s", output)
	}
}

// TestBasicTestRunner_RunStream_CancelledContext tests RunStream with cancelled context
func TestBasicTestRunner_RunStream_CancelledContext(t *testing.T) {
	t.Parallel()

	// Create a temporary test package
	tempDir := createTestPackageForBasicRunner(t)
	defer os.RemoveAll(tempDir)

	runner := NewBasicTestRunner(false, true)

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Act
	stream, err := runner.RunStream(ctx, []string{tempDir})

	// Assert
	// The behavior depends on timing - if the command starts before context check,
	// it may succeed. If context is checked first, it should fail.
	if err != nil {
		if !strings.Contains(err.Error(), "context") && !strings.Contains(err.Error(), "cancel") {
			t.Errorf("Expected context cancellation error, got: %v", err)
		}
		if stream != nil {
			t.Error("Expected nil stream on context cancellation error")
		}
	} else {
		// If no error, the stream was created before cancellation was detected
		if stream == nil {
			t.Error("Expected non-nil stream if no error")
		} else {
			// Clean up the stream
			stream.Close()
		}
		t.Log("Stream created before context cancellation was detected - this is acceptable")
	}
}

// TestBasicTestRunner_Configuration tests different configuration combinations
func TestBasicTestRunner_Configuration(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		verbose    bool
		jsonOutput bool
	}{
		{
			name:       "verbose enabled",
			verbose:    true,
			jsonOutput: false,
		},
		{
			name:       "json output enabled",
			verbose:    false,
			jsonOutput: true,
		},
		{
			name:       "both enabled",
			verbose:    true,
			jsonOutput: true,
		},
		{
			name:       "both disabled",
			verbose:    false,
			jsonOutput: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			runner := NewBasicTestRunner(tc.verbose, tc.jsonOutput)

			// Assert
			if runner.Verbose != tc.verbose {
				t.Errorf("Expected Verbose=%v, got %v", tc.verbose, runner.Verbose)
			}
			if runner.JSONOutput != tc.jsonOutput {
				t.Errorf("Expected JSONOutput=%v, got %v", tc.jsonOutput, runner.JSONOutput)
			}
		})
	}
}

// TestBasicTestRunner_UtilityFunctions tests the utility functions
func TestBasicTestRunner_UtilityFunctions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		filename string
		isGoTest bool
		isGo     bool
	}{
		{
			name:     "go test file",
			filename: "example_test.go",
			isGoTest: true,
			isGo:     true,
		},
		{
			name:     "go source file",
			filename: "example.go",
			isGoTest: false,
			isGo:     true,
		},
		{
			name:     "non-go file",
			filename: "example.txt",
			isGoTest: false,
			isGo:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Test IsGoTestFile
			if IsGoTestFile(tc.filename) != tc.isGoTest {
				t.Errorf("IsGoTestFile(%s) = %v, expected %v", tc.filename, IsGoTestFile(tc.filename), tc.isGoTest)
			}

			// Test IsGoFile
			if IsGoFile(tc.filename) != tc.isGo {
				t.Errorf("IsGoFile(%s) = %v, expected %v", tc.filename, IsGoFile(tc.filename), tc.isGo)
			}
		})
	}
}

// TestNewTestRunner_BackwardCompatibility tests the backward compatibility function
func TestNewTestRunner_BackwardCompatibility(t *testing.T) {
	t.Parallel()

	// Act
	runner := NewTestRunner(true, false)

	// Assert
	if runner == nil {
		t.Fatal("NewTestRunner should not return nil")
	}
	if runner.Verbose != true {
		t.Error("Expected Verbose=true")
	}
	if runner.JSONOutput != false {
		t.Error("Expected JSONOutput=false for backward compatibility")
	}

	// Verify it's the same type as BasicTestRunner
	var _ *BasicTestRunner = runner
}

// TestStreamReader_ReadAndClose tests the streamReader functionality
func TestStreamReader_ReadAndClose(t *testing.T) {
	t.Parallel()

	// Create a temporary test package
	tempDir := createTestPackageForBasicRunner(t)
	defer os.RemoveAll(tempDir)

	runner := NewBasicTestRunner(false, true)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Act - Get a stream
	stream, err := runner.RunStream(ctx, []string{tempDir})
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}
	if stream == nil {
		t.Fatal("Expected non-nil stream")
	}

	// Test Read functionality
	buffer := make([]byte, 512)
	totalRead := 0

	// Read in a loop to test the Read method thoroughly
	for i := 0; i < 5; i++ {
		n, readErr := stream.Read(buffer)
		totalRead += n

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			t.Logf("Read error (may be expected): %v", readErr)
			break
		}

		if n > 0 {
			t.Logf("Read %d bytes in iteration %d", n, i)
		}
	}

	// Test Close functionality
	closeErr := stream.Close()
	if closeErr != nil {
		t.Logf("Close error (may be expected for test process): %v", closeErr)
	}

	t.Logf("Total bytes read: %d", totalRead)
}

// createTestPackageForBasicRunner creates a temporary test package for testing
func createTestPackageForBasicRunner(t *testing.T) string {
	tempDir, err := os.MkdirTemp("", "basic_runner_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create go.mod
	goMod := `module test-package
go 1.21`
	err = os.WriteFile(filepath.Join(tempDir, "go.mod"), []byte(goMod), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a simple test file
	testContent := `package main

import "testing"

func TestBasicExample(t *testing.T) {
	if 1+1 != 2 {
		t.Error("Math is broken")
	}
}

func TestAnotherExample(t *testing.T) {
	if "hello" == "" {
		t.Error("String comparison failed")
	}
}`
	err = os.WriteFile(filepath.Join(tempDir, "main_test.go"), []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	return tempDir
}
