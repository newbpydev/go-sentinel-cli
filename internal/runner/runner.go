// Package runner provides functionality for executing Go tests and processing their results.
// It includes tools for running tests, parsing output, and handling test events.
package runner

import (
	"bytes"
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// Runner manages the execution of Go tests and provides configuration options
// for timeouts and inactivity detection to prevent hanging tests.
type Runner struct{
	// Default timeout for test execution
	defaultTimeout time.Duration
	// Threshold for detecting inactive/hanging tests
	inactivityThreshold time.Duration
}

// NewRunner creates a new test runner with default timeout and inactivity threshold settings.
// By default, tests will timeout after 2 minutes, and tests showing no activity for 30 seconds
// will be considered potentially hanging.
func NewRunner() *Runner { 
	return &Runner{
		defaultTimeout: 2 * time.Minute, // Default 2 minute timeout
		inactivityThreshold: 30 * time.Second, // Default 30 second inactivity threshold
	} 
}

// startGoTest runs `go test -json` in the given pkg, optionally for a specific testName.
// Returns the exec.Cmd, a buffer that will contain the output, and any startup error.
// The caller is responsible for calling Wait() on the command.
// Note: Currently testName parameter is not used but kept for future implementation.
func (r *Runner) startGoTest(pkg string, testName string) (*exec.Cmd, *bytes.Buffer, error) {
	// Validate package path to prevent command injection
	if !isValidPackagePath(pkg) {
		return nil, nil, fmt.Errorf("invalid package path: %s", pkg)
	}
	
	args := []string{"test", "-json"}
	// Uncomment when testName implementation is needed
	// if testName != "" {
	// 	args = append(args, "-run", testName)
	// }
	args = append(args, pkg)
	cmd := exec.Command("go", args...)
	cmd.Dir = findProjectRoot()

	// Create a buffer to capture the output
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, &buf, fmt.Errorf("failed to start command: %w", err)
	}

	return cmd, &buf, nil
}

// SetTimeout sets the default timeout for test execution
func (r *Runner) SetTimeout(timeout time.Duration) {
	r.defaultTimeout = timeout
}

// SetInactivityThreshold sets the threshold for detecting inactive/hanging tests
func (r *Runner) SetInactivityThreshold(threshold time.Duration) {
	r.inactivityThreshold = threshold
}

// buildTestArgs creates the argument slice for the go test command
func (r *Runner) buildTestArgs(pkg string, testName string, timeout time.Duration) []string {
	// CRITICAL FIX: Use both -json for machine parsing AND -v for human-readable verbose output
	args := []string{"test", "-json", "-v"}

	// Add timeout flag
	if timeout > 0 {
		args = append(args, "-timeout", timeout.String())
	}

	// Add test filter if specified
	if testName != "" {
		args = append(args, "-run", testName)
	}

	// Add package path
	args = append(args, pkg)
	return args
}

// Run executes go test -json and sends each output line to the channel.
// For backward compatibility, this calls RunWithContext using the default timeout.
func (r *Runner) Run(pkg string, testName string, out chan<- []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), r.defaultTimeout)
	defer cancel()
	return r.RunWithContext(ctx, pkg, testName, out)
}

// RunWithContext executes go test -json with context control and sends each output line to the channel.
func (r *Runner) RunWithContext(ctx context.Context, pkg string, testName string, out chan<- []byte) error {
	// Build command args with timeout
	timeout, hasTimeout := ctx.Deadline()
	var cmdTimeout time.Duration
	if hasTimeout {
		cmdTimeout = time.Until(timeout)
		if cmdTimeout <= 0 {
			return context.DeadlineExceeded
		}
	} else {
		cmdTimeout = r.defaultTimeout
	}

	// Create command with args and context
	args := r.buildTestArgs(pkg, testName, cmdTimeout)
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = findProjectRoot()
	
	// Log the exact command being executed - this helps in debugging
	cmdString := "go " + strings.Join(args, " ")
	out <- []byte(fmt.Sprintf("\n[COMMAND] %s\n", cmdString))
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr io.ReadCloser
	stderr, err = cmd.StderrPipe()
	if err != nil {
		return err
	}
	if startErr := cmd.Start(); startErr != nil {
		return startErr
	}

	// Set up channels for detecting inactivity
	activityCh := make(chan struct{}, 10)

	// Set up inactivity detection
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel() // Ensure cleanup of resources

	// Start inactivity detector
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		lastActivity := time.Now()
		for {
			select {
			case <-ctxWithCancel.Done():
				return
			case <-activityCh:
				lastActivity = time.Now()
			case <-ticker.C:
				inactiveDuration := time.Since(lastActivity)
				if inactiveDuration > r.inactivityThreshold {
					// Log the inactivity warning
					warning := fmt.Sprintf("[runner warning] No activity detected for %v, possible hanging test in package %s", inactiveDuration.Round(time.Second), pkg)
					out <- []byte(warning)
				}
			}
		}
	}()

	var wg sync.WaitGroup
	stream := func(reader io.Reader) {
		defer wg.Done()
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			// Check if context is canceled
			select {
			case <-ctxWithCancel.Done():
				return
			default:
				// Continue with processing
			}

			// Signal activity
			activityCh <- struct{}{}

			line := scanner.Bytes()
			if len(bytes.TrimSpace(line)) > 0 {
				out <- append([]byte{}, line...)
			}
		}
		if scanErr := scanner.Err(); scanErr != nil {
			out <- []byte("[runner debug] scanner error: " + scanErr.Error())
		}
	}

	wg.Add(2)
	go stream(stdout)
	go stream(stderr)

	// Wait for streams to complete or context to be canceled
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Streams completed normally
	case <-ctx.Done():
		// Context was canceled
		cancel() // Cancel our own context to stop routines
		out <- []byte(fmt.Sprintf("[runner] Test execution stopped: %v", ctx.Err()))
		return ctx.Err()
	}

	err = cmd.Wait()
	if err != nil {
		// Check for context cancellation
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			out <- []byte(fmt.Sprintf("[runner] Test timed out after %v", r.defaultTimeout))
		} else {
			// Log other errors for debugging
			out <- []byte("[runner debug] cmd.Wait() error: " + err.Error())
		}
	}
	return err
}


// findProjectRoot returns the absolute path to the project root directory.
func findProjectRoot() string {
	// Hardcoded for now; in a real implementation, this could be dynamic
	return "c:/Users/Remym/pythonProject/__personal-projects/go-sentinel"
}
