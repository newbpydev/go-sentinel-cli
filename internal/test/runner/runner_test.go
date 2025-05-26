package runner_test

import (
	"context"
	"strings"
	"testing"

	"github.com/newbpydev/go-sentinel/internal/test/runner"
)

// TestNewBasicTestRunner_FactoryFunction tests the factory function
func TestNewBasicTestRunner_FactoryFunction(t *testing.T) {
	t.Parallel()

	// Act
	r := runner.NewBasicTestRunner(false, false)

	// Assert
	if r == nil {
		t.Fatal("NewBasicTestRunner should not return nil")
	}
}

// TestBasicTestRunner_Run_EmptyPackages tests running with empty package list
func TestBasicTestRunner_Run_EmptyPackages(t *testing.T) {
	t.Parallel()

	// Arrange
	r := runner.NewBasicTestRunner(false, false)
	ctx := context.Background()
	packages := []string{}

	// Act
	result, err := r.Run(ctx, packages)

	// Assert
	if err == nil {
		t.Error("Run should return error for empty packages")
	}

	if result != "" {
		t.Error("Run should return empty result on error")
	}
}

// TestBasicTestRunner_Run_NonExistentPath tests running with non-existent path
func TestBasicTestRunner_Run_NonExistentPath(t *testing.T) {
	t.Parallel()

	// Arrange
	r := runner.NewBasicTestRunner(false, false)
	ctx := context.Background()
	packages := []string{"./non-existent-path"}

	// Act
	result, err := r.Run(ctx, packages)

	// Assert
	if err == nil {
		t.Error("Run should return error for non-existent path")
	}

	if result != "" {
		t.Error("Run should return empty result on error")
	}

	expectedMessage := "test path does not exist"
	if !strings.Contains(err.Error(), expectedMessage) {
		t.Errorf("Expected error containing %q, got: %v", expectedMessage, err)
	}
}

// TestBasicTestRunner_Run_EmptyStringPath tests running with empty string path
func TestBasicTestRunner_Run_EmptyStringPath(t *testing.T) {
	t.Parallel()

	// Arrange
	r := runner.NewBasicTestRunner(false, false)
	ctx := context.Background()
	packages := []string{""}

	// Act
	result, err := r.Run(ctx, packages)

	// Assert
	if err == nil {
		t.Error("Run should return error for empty string path")
	}

	if result != "" {
		t.Error("Run should return empty result on error")
	}

	expectedMessage := "empty test path provided"
	if !strings.Contains(err.Error(), expectedMessage) {
		t.Errorf("Expected error containing %q, got: %v", expectedMessage, err)
	}
}

// TestBasicTestRunner_Configuration tests runner configuration options
func TestBasicTestRunner_Configuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		verbose    bool
		jsonOutput bool
	}{
		{"verbose enabled", true, false},
		{"json output enabled", false, true},
		{"both enabled", true, true},
		{"both disabled", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			r := runner.NewBasicTestRunner(tt.verbose, tt.jsonOutput)

			// Assert
			if r == nil {
				t.Fatal("NewBasicTestRunner should not return nil")
			}

			// Test that the runner has the expected configuration
			// Note: We're not testing private fields, but the behavior they should produce
			// This test validates the constructor works with different configurations
		})
	}
}

// TestBasicTestRunner_UtilityFunctions tests the utility functions
func TestBasicTestRunner_UtilityFunctions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		expected bool
		function func(string) bool
	}{
		{"go test file", "example_test.go", true, runner.IsGoTestFile},
		{"go source file", "example.go", false, runner.IsGoTestFile},
		{"non-go file", "example.txt", false, runner.IsGoTestFile},
		{"go source file", "example.go", true, runner.IsGoFile},
		{"go test file", "example_test.go", true, runner.IsGoFile},
		{"non-go file", "example.txt", false, runner.IsGoFile},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			result := tt.function(tt.path)

			// Assert
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for path %q", tt.expected, result, tt.path)
			}
		})
	}
}
