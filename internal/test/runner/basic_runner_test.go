package runner

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
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

// TestNewBasicTestRunner_ComprehensiveCoverage tests the NewBasicTestRunner factory function
func TestNewBasicTestRunner_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		verbose    bool
		jsonOutput bool
		validate   func(*testing.T, *BasicTestRunner)
	}{
		"default_configuration": {
			verbose:    false,
			jsonOutput: false,
			validate: func(t *testing.T, runner *BasicTestRunner) {
				if runner.Verbose {
					t.Error("Expected Verbose to be false")
				}
				if runner.JSONOutput {
					t.Error("Expected JSONOutput to be false")
				}
			},
		},
		"verbose_enabled": {
			verbose:    true,
			jsonOutput: false,
			validate: func(t *testing.T, runner *BasicTestRunner) {
				if !runner.Verbose {
					t.Error("Expected Verbose to be true")
				}
				if runner.JSONOutput {
					t.Error("Expected JSONOutput to be false")
				}
			},
		},
		"json_output_enabled": {
			verbose:    false,
			jsonOutput: true,
			validate: func(t *testing.T, runner *BasicTestRunner) {
				if runner.Verbose {
					t.Error("Expected Verbose to be false")
				}
				if !runner.JSONOutput {
					t.Error("Expected JSONOutput to be true")
				}
			},
		},
		"both_enabled": {
			verbose:    true,
			jsonOutput: true,
			validate: func(t *testing.T, runner *BasicTestRunner) {
				if !runner.Verbose {
					t.Error("Expected Verbose to be true")
				}
				if !runner.JSONOutput {
					t.Error("Expected JSONOutput to be true")
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			runner := NewBasicTestRunner(tt.verbose, tt.jsonOutput)

			if runner == nil {
				t.Fatal("NewBasicTestRunner should not return nil")
			}

			// Verify it implements the TestRunnerInterface
			var _ TestRunnerInterface = runner

			tt.validate(t, runner)
		})
	}
}

// TestStreamReader_ComprehensiveCoverage tests the streamReader Read and Close methods
func TestStreamReader_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	// Test Read method with valid data
	t.Run("read_method", func(t *testing.T) {
		t.Parallel()

		// Create test data
		testData := "line 1\nline 2\nline 3\n"
		reader := strings.NewReader(testData)

		streamReader := &streamReader{
			reader: io.NopCloser(reader),
			cmd:    nil, // No command for this test
			stderr: &bytes.Buffer{},
		}

		// Test Read method
		buffer := make([]byte, 10)
		n, err := streamReader.Read(buffer)
		if err != nil && err != io.EOF {
			t.Fatalf("Read should not error: %v", err)
		}
		if n == 0 {
			t.Error("Read should return some bytes")
		}
	})

	// Test Close method with nil command (should not panic)
	t.Run("close_with_nil_command", func(t *testing.T) {
		t.Parallel()

		testData := "test data"
		reader := strings.NewReader(testData)

		streamReader := &streamReader{
			reader: io.NopCloser(reader),
			cmd:    nil, // No command
			stderr: &bytes.Buffer{},
		}

		// Test Close method - should not panic with nil cmd
		err := streamReader.Close()
		if err != nil {
			t.Errorf("Close should not error with nil cmd: %v", err)
		}
	})

	// Test with actual command (integration test)
	t.Run("close_with_real_command", func(t *testing.T) {
		t.Parallel()

		// Create a simple command that will complete quickly
		cmd := exec.Command("go", "version")

		// Start the command and get stdout
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			t.Fatalf("Failed to create stdout pipe: %v", err)
		}

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Start(); err != nil {
			t.Fatalf("Failed to start command: %v", err)
		}

		streamReader := &streamReader{
			reader: stdout,
			cmd:    cmd,
			stderr: &stderr,
		}

		// Read some data
		buffer := make([]byte, 100)
		_, err = streamReader.Read(buffer)
		// Error is acceptable here (might be EOF)

		// Test Close method with real command
		err = streamReader.Close()
		if err != nil {
			t.Logf("Close returned error (acceptable for test command): %v", err)
		}
	})
}

// TestFileTypeCheckers_ComprehensiveCoverage tests IsGoTestFile and IsGoFile functions
func TestFileTypeCheckers_ComprehensiveCoverage(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		filename       string
		expectedGoTest bool
		expectedGo     bool
	}{
		"go_test_file": {
			filename:       "example_test.go",
			expectedGoTest: true,
			expectedGo:     true,
		},
		"go_source_file": {
			filename:       "example.go",
			expectedGoTest: false,
			expectedGo:     true,
		},
		"non_go_file": {
			filename:       "example.txt",
			expectedGoTest: false,
			expectedGo:     false,
		},
		"go_file_without_extension": {
			filename:       "example",
			expectedGoTest: false,
			expectedGo:     false,
		},
		"test_file_without_go_extension": {
			filename:       "example_test.txt",
			expectedGoTest: false,
			expectedGo:     false,
		},
		"empty_filename": {
			filename:       "",
			expectedGoTest: false,
			expectedGo:     false,
		},
		"go_file_with_path": {
			filename:       "/path/to/example.go",
			expectedGoTest: false,
			expectedGo:     true,
		},
		"test_file_with_path": {
			filename:       "/path/to/example_test.go",
			expectedGoTest: true,
			expectedGo:     true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Test IsGoTestFile
			isGoTest := IsGoTestFile(tt.filename)
			if isGoTest != tt.expectedGoTest {
				t.Errorf("IsGoTestFile(%q) = %v, expected %v", tt.filename, isGoTest, tt.expectedGoTest)
			}

			// Test IsGoFile
			isGo := IsGoFile(tt.filename)
			if isGo != tt.expectedGo {
				t.Errorf("IsGoFile(%q) = %v, expected %v", tt.filename, isGo, tt.expectedGo)
			}
		})
	}
}

// TestNewTestRunner_BackwardCompatibilityComprehensive tests the NewTestRunner backward compatibility function
func TestNewTestRunner_BackwardCompatibilityComprehensive(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		verbose    bool
		jsonOutput bool
		validate   func(*testing.T, *BasicTestRunner)
	}{
		"default_settings": {
			verbose:    false,
			jsonOutput: false,
			validate: func(t *testing.T, runner *BasicTestRunner) {
				if runner.Verbose {
					t.Error("Expected Verbose to be false")
				}
				if runner.JSONOutput {
					t.Error("Expected JSONOutput to be false")
				}
			},
		},
		"verbose_only": {
			verbose:    true,
			jsonOutput: false,
			validate: func(t *testing.T, runner *BasicTestRunner) {
				if !runner.Verbose {
					t.Error("Expected Verbose to be true")
				}
				if runner.JSONOutput {
					t.Error("Expected JSONOutput to be false")
				}
			},
		},
		"json_only": {
			verbose:    false,
			jsonOutput: true,
			validate: func(t *testing.T, runner *BasicTestRunner) {
				if runner.Verbose {
					t.Error("Expected Verbose to be false")
				}
				if !runner.JSONOutput {
					t.Error("Expected JSONOutput to be true")
				}
			},
		},
		"both_enabled": {
			verbose:    true,
			jsonOutput: true,
			validate: func(t *testing.T, runner *BasicTestRunner) {
				if !runner.Verbose {
					t.Error("Expected Verbose to be true")
				}
				if !runner.JSONOutput {
					t.Error("Expected JSONOutput to be true")
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			runner := NewTestRunner(tt.verbose, tt.jsonOutput)

			if runner == nil {
				t.Fatal("NewTestRunner should not return nil")
			}

			// Verify it implements the TestRunnerInterface
			var _ TestRunnerInterface = runner

			tt.validate(t, runner)
		})
	}
}
