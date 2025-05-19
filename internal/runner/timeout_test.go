package runner

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// Test 3.2.5.1: Test that timeout flag is properly set
func TestRunnerWithTimeout(t *testing.T) {
	r := NewRunner()
	args := r.buildTestArgs("./...", "", 30*time.Second)

	// Verify timeout flag is included
	hasTimeoutFlag := false
	for i, arg := range args {
		if arg == "-timeout" && i+1 < len(args) {
			hasTimeoutFlag = true
			duration := args[i+1]
			if duration != "30s" {
				t.Errorf("expected timeout of 30s, got %s", duration)
			}
			break
		}
	}

	if !hasTimeoutFlag {
		t.Error("timeout flag not included in test args")
	}
}

// Test 3.2.5.2: Test context cancellation for graceful termination
func TestRunnerWithContextCancellation(t *testing.T) {
	r := NewRunner()

	// Create a context that will be canceled very quickly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start a channel to receive output
	out := make(chan []byte, 10)

	// Run a test that would normally take longer than our context timeout
	go func() {
		err := r.RunWithContext(ctx, "./testdata/timeouttest", "", out)
		if err == nil {
			t.Error("expected error due to context cancellation, got nil")
		}
		// The error could be context.DeadlineExceeded, context canceled, or exit status
		// as the process might terminate in different ways depending on timing
		isValidError := errors.Is(err, context.DeadlineExceeded) ||
			strings.Contains(err.Error(), "context") ||
			strings.Contains(err.Error(), "exit status")
		if !isValidError {
			t.Errorf("expected valid termination error, got %v", err)
		}
		close(out)
	}()

	// Read output until channel closes
	timeout := false
	for line := range out {
		if strings.Contains(string(line), "timeout") || strings.Contains(string(line), "canceled") {
			timeout = true
		}
	}

	if !timeout {
		t.Error("expected timeout message, none received")
	}
}

// Test 3.2.5.3: Test detection of hanging tests
func TestHangingTestDetection(t *testing.T) {
	r := NewRunner()

	// Create output channel
	out := make(chan []byte, 100)

	// Set a quick inactivity threshold for testing
	r.SetInactivityThreshold(500 * time.Millisecond)

	// Start a context with timeout longer than inactivity threshold
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Run the test in a goroutine
	go func() {
		err := r.RunWithContext(ctx, "./testdata/hangingtest", "", out)
		if err == nil {
			// We expect an error due to the context timeout
			t.Error("expected error from hanging test, got nil")
		}
		close(out)
	}()

	// Check for inactivity warning
	inactivityWarning := false
	for line := range out {
		output := string(line)
		if strings.Contains(output, "inactivity") || strings.Contains(output, "hanging") {
			inactivityWarning = true
		}
	}

	if !inactivityWarning {
		t.Error("expected inactivity warning, none received")
	}
}
