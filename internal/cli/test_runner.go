package cli

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// TestRunnerInterface defines the interface for test runners
type TestRunnerInterface interface {
	Run(ctx context.Context, testPaths []string) (string, error)
}

// TestRunner executes Go tests
type TestRunner struct {
	// Verbose enables verbose output
	Verbose bool

	// JSONOutput enables JSON output format
	JSONOutput bool
}

// Run executes the specified tests and returns the output
func (r *TestRunner) Run(ctx context.Context, testPaths []string) (string, error) {
	// Build the command arguments
	args := []string{"test"}

	// Add verbose flag if required
	if r.Verbose {
		args = append(args, "-v")
	}

	// Add JSON output flag if required
	if r.JSONOutput {
		args = append(args, "-json")
	}

	// Add the test paths
	args = append(args, testPaths...)

	// Create the command
	cmd := exec.CommandContext(ctx, "go", args...)

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()
	if err != nil {
		// Return stderr if there was an error running the command itself
		if stderr.Len() > 0 {
			return "", fmt.Errorf("error running tests: %w: %s", err, stderr.String())
		}

		// For test failures, we still want to return the stdout to process the results
		if _, ok := err.(*exec.ExitError); ok {
			return stdout.String(), nil
		}

		return "", fmt.Errorf("error running tests: %w", err)
	}

	return stdout.String(), nil
}

// IsGoTestFile returns true if the file is a Go test file
func IsGoTestFile(path string) bool {
	return strings.HasSuffix(path, "_test.go")
}

// IsGoFile returns true if the file is a Go source file
func IsGoFile(path string) bool {
	return strings.HasSuffix(path, ".go")
}
