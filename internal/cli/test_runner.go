package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// TestRunnerInterface defines the interface for test runners
type TestRunnerInterface interface {
	Run(ctx context.Context, testPaths []string) (string, error)
	RunStream(ctx context.Context, testPaths []string) (io.ReadCloser, error)
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
	// Validate test paths
	if len(testPaths) == 0 {
		return "", fmt.Errorf("no test paths provided")
	}

	for _, path := range testPaths {
		if path == "" {
			return "", fmt.Errorf("empty test path provided")
		}

		// Check if path exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return "", fmt.Errorf("test path does not exist: %s", path)
		}
	}

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

// RunStream executes the specified tests and returns a stream of JSON output
func (r *TestRunner) RunStream(ctx context.Context, testPaths []string) (io.ReadCloser, error) {
	// Validate test paths
	if len(testPaths) == 0 {
		return nil, fmt.Errorf("no test paths provided")
	}

	for _, path := range testPaths {
		if path == "" {
			return nil, fmt.Errorf("empty test path provided")
		}

		// Check if path exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("test path does not exist: %s", path)
		}
	}

	// Build the command arguments
	args := []string{"test"}

	// Add verbose flag if required
	if r.Verbose {
		args = append(args, "-v")
	}

	// Add JSON output flag - required for streaming
	args = append(args, "-json")

	// Add the test paths
	args = append(args, testPaths...)

	// Create the command
	cmd := exec.CommandContext(ctx, "go", args...)

	// Get stdout pipe for streaming
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Capture stderr for error reporting
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Start the command
	if err := cmd.Start(); err != nil {
		stdout.Close()
		return nil, fmt.Errorf("failed to start test command: %w", err)
	}

	// Return a reader that will close the process when done
	return &streamReader{
		reader: stdout,
		cmd:    cmd,
		stderr: &stderr,
	}, nil
}

// streamReader wraps the stdout pipe and handles process cleanup
type streamReader struct {
	reader io.ReadCloser
	cmd    *exec.Cmd
	stderr *bytes.Buffer
}

func (sr *streamReader) Read(p []byte) (n int, err error) {
	return sr.reader.Read(p)
}

func (sr *streamReader) Close() error {
	// Close the reader first
	sr.reader.Close()

	// Wait for the command to finish
	if err := sr.cmd.Wait(); err != nil {
		// For test failures, this is expected
		if _, ok := err.(*exec.ExitError); ok {
			return nil // Test failures are not stream errors
		}
		// Check if there's stderr output
		if sr.stderr.Len() > 0 {
			return fmt.Errorf("test command error: %w: %s", err, sr.stderr.String())
		}
		return fmt.Errorf("test command error: %w", err)
	}

	return nil
}

// IsGoTestFile returns true if the file is a Go test file
func IsGoTestFile(path string) bool {
	return strings.HasSuffix(path, "_test.go")
}

// IsGoFile returns true if the file is a Go source file
func IsGoFile(path string) bool {
	return strings.HasSuffix(path, ".go")
}
