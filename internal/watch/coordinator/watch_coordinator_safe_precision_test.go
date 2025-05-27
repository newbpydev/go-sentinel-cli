package coordinator

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/processor"
	"github.com/newbpydev/go-sentinel/internal/test/runner"
	"github.com/newbpydev/go-sentinel/internal/watch/core"
)

// TestNewTestWatchCoordinator_ExactPathsEdgeCase tests the exact paths edge case for 100% coverage
func TestNewTestWatchCoordinator_ExactPathsEdgeCase(t *testing.T) {
	t.Parallel()

	// Test with nil paths slice (edge case for rootDir assignment)
	options := core.WatchOptions{
		Paths:  nil, // This should trigger rootDir = "." assignment
		Mode:   core.WatchAll,
		Writer: &strings.Builder{},
	}

	coord, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Errorf("NewTestWatchCoordinator should handle nil paths: %v", err)
	}
	if coord == nil {
		t.Error("NewTestWatchCoordinator should not return nil with nil paths")
	}
}

// TestNewTestWatchCoordinator_TestFinderCreationError tests test finder creation error path
func TestNewTestWatchCoordinator_TestFinderCreationError(t *testing.T) {
	t.Parallel()

	// Test with invalid path that could cause test finder creation to fail
	options := core.WatchOptions{
		Paths:  []string{"/dev/null/invalid"}, // Invalid path that might cause creation failure
		Mode:   core.WatchAll,
		Writer: &strings.Builder{},
	}

	// This tests the error path in test finder creation
	coord, err := NewTestWatchCoordinator(options)
	// We expect this to either succeed or fail gracefully
	if err != nil {
		// Error path is acceptable - we're testing error handling
		if coord != nil {
			t.Error("NewTestWatchCoordinator should return nil when error occurs")
		}
	} else {
		// Success path is also acceptable - we're testing the creation
		if coord == nil {
			t.Error("NewTestWatchCoordinator should not return nil when no error")
		}
	}
}

// TestStart_ExactCancelledContextPath tests the exact cancelled context handling
func TestStart_ExactCancelledContextPath(t *testing.T) {
	t.Parallel()

	options := core.WatchOptions{
		Paths:  []string{"./"},
		Mode:   core.WatchAll,
		Writer: &strings.Builder{},
	}

	coord, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("Failed to create coordinator: %v", err)
	}

	// Replace with SAFE mock that doesn't trigger real processes
	coord.testRunner = &SafeMockTestRunner{
		RunFunc: func(ctx context.Context, packages []string) (string, error) {
			// SAFE: Return mock data without executing anything
			return `{"Action":"pass","Package":"mock","Test":"MockTest","Output":"PASS"}`, nil
		},
	}

	// Create a pre-cancelled context to test the exact cancellation path
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// This should handle the cancelled context path specifically
	err = coord.Start(ctx)
	if err == nil {
		t.Error("Start should return error with cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}
}

// TestStart_ProcessorConfigurationPath tests the exact processor configuration code path
func TestStart_ProcessorConfigurationPath(t *testing.T) {
	t.Parallel()

	output := &strings.Builder{}
	options := core.WatchOptions{
		Paths:  []string{"./"},
		Mode:   core.WatchAll,
		Writer: output,
	}

	coord, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("Failed to create coordinator: %v", err)
	}

	// Configure with specific processor to test processor configuration path
	coord.processor = processor.NewTestProcessor(output, &mockColorFormatter{}, &mockIconProvider{}, 80)

	// Replace with SAFE mock that doesn't trigger real processes
	coord.testRunner = &SafeMockTestRunner{
		RunFunc: func(ctx context.Context, packages []string) (string, error) {
			// SAFE: Return controlled mock data
			return `{"Action":"pass","Package":"safe","Test":"SafeTest","Output":"PASS"}`, nil
		},
	}

	// Create very short timeout to trigger timeout path safely
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// This tests the processor configuration path in Start
	err = coord.Start(ctx)
	if err == nil {
		t.Error("Expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
	}
}

// TestExecuteTests_ExactProcessorCallPath tests the exact processor call path in executeTests
func TestExecuteTests_ExactProcessorCallPath(t *testing.T) {
	t.Parallel()

	output := &strings.Builder{}
	options := core.WatchOptions{
		Paths:  []string{"./"},
		Mode:   core.WatchAll,
		Writer: output,
	}

	coord, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("Failed to create coordinator: %v", err)
	}

	// Replace with SAFE mock that returns specific JSON for processor testing
	coord.testRunner = &SafeMockTestRunner{
		RunFunc: func(ctx context.Context, packages []string) (string, error) {
			// SAFE: Return controlled JSON that will trigger processor path
			return `{"Action":"run","Package":"exact","Test":"ExactTest"}
{"Action":"output","Package":"exact","Test":"ExactTest","Output":"Running test\n"}
{"Action":"pass","Package":"exact","Test":"ExactTest","Output":"PASS\n"}`, nil
		},
	}

	// Set processor to trigger the exact processor call path
	coord.processor = processor.NewTestProcessor(output, &mockColorFormatter{}, &mockIconProvider{}, 80)

	// This should trigger the exact processor call path in executeTests
	err = coord.executeTests([]string{"./safe-test"})
	if err != nil {
		t.Errorf("executeTests should not error with safe mock: %v", err)
	}

	// Verify processor was called (output should contain processed results)
	outputStr := output.String()
	if len(outputStr) == 0 {
		t.Error("Expected processor to generate output")
	}
}

// TestExecuteTests_ExactRunnerErrorPath tests the exact runner error handling path
func TestExecuteTests_ExactRunnerErrorPath(t *testing.T) {
	t.Parallel()

	options := core.WatchOptions{
		Paths:  []string{"./"},
		Mode:   core.WatchAll,
		Writer: &strings.Builder{},
	}

	coord, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("Failed to create coordinator: %v", err)
	}

	// Replace with SAFE mock that returns specific error to test error path
	coord.testRunner = &SafeMockTestRunner{
		RunFunc: func(ctx context.Context, packages []string) (string, error) {
			// SAFE: Return controlled error without executing anything
			return "", errors.New("safe mock error for testing")
		},
	}

	// This should trigger the exact error handling path in executeTests
	err = coord.executeTests([]string{"./safe-error-test"})
	if err == nil {
		t.Error("executeTests should return error when runner fails")
	}
	if !strings.Contains(err.Error(), "safe mock error") {
		t.Errorf("Expected 'safe mock error', got: %v", err)
	}
}

// TestStart_ExactRunOnStartPath tests the exact RunOnStart code path
func TestStart_ExactRunOnStartPath(t *testing.T) {
	t.Parallel()

	options := core.WatchOptions{
		Paths:      []string{"./"},
		Mode:       core.WatchAll,
		Writer:     &strings.Builder{},
		RunOnStart: true, // This should trigger the RunOnStart path
	}

	coord, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("Failed to create coordinator: %v", err)
	}

	// Replace with SAFE mock that doesn't trigger real processes
	coord.testRunner = &SafeMockTestRunner{
		RunFunc: func(ctx context.Context, packages []string) (string, error) {
			// SAFE: Return mock data for RunOnStart test
			return `{"Action":"pass","Package":"runonstart","Test":"RunOnStartTest","Output":"PASS"}`, nil
		},
	}

	// Use very short timeout to test RunOnStart path and then exit
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	// This should trigger the RunOnStart path in Start method
	err = coord.Start(ctx)
	if err == nil {
		t.Error("Expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
	}
}

// SafeMockTestRunner - Completely safe mock that never executes real processes
type SafeMockTestRunner struct {
	RunFunc       func(ctx context.Context, testPaths []string) (string, error)
	RunStreamFunc func(ctx context.Context, testPaths []string) (io.ReadCloser, error)
}

func (m *SafeMockTestRunner) Run(ctx context.Context, testPaths []string) (string, error) {
	if m.RunFunc != nil {
		return m.RunFunc(ctx, testPaths)
	}
	// SAFE: Always return controlled mock data
	return `{"Action":"pass","Package":"safe","Test":"SafeTest","Output":"PASS"}`, nil
}

func (m *SafeMockTestRunner) RunStream(ctx context.Context, testPaths []string) (io.ReadCloser, error) {
	if m.RunStreamFunc != nil {
		return m.RunStreamFunc(ctx, testPaths)
	}
	// SAFE: Return nil without triggering real streams
	return nil, nil
}

// Ensure SafeMockTestRunner implements the interface
var _ runner.TestRunnerInterface = (*SafeMockTestRunner)(nil)
