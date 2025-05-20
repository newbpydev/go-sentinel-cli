package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunner_RunOnce(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "go-sentinel-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize Go module
	if err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example\n\ngo 1.23\n"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tmpDir, "example_test.go")
	err = os.WriteFile(testFile, []byte(`package example

import "testing"

func TestPass(t *testing.T) {
	// This test should pass
}

func TestFail(t *testing.T) {
	t.Error("This test should fail")
}

func TestSkip(t *testing.T) {
	t.Skip("This test should be skipped")
}`), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a runner
	runner, err := NewRunner(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}
	defer runner.Stop()

	// Run all tests
	ctx := context.Background()
	err = runner.Run(ctx, RunOptions{})

	// We expect an error because one test fails
	if err == nil {
		t.Error("Expected error from failing test, got nil")
	}

	// Run only passing test
	err = runner.Run(ctx, RunOptions{
		Tests: []string{"TestPass"},
	})

	// Should not error when running only passing test
	if err != nil {
		t.Errorf("Expected no error when running passing test, got: %v", err)
	}
}

func TestRunner_ShouldRunTests(t *testing.T) {
	runner, err := NewRunner(".")
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}
	defer runner.Stop()

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "go file",
			path:     "example.go",
			expected: true,
		},
		{
			name:     "go test file",
			path:     "example_test.go",
			expected: true,
		},
		{
			name:     "non-go file",
			path:     "example.txt",
			expected: false,
		},
		{
			name:     "hidden file",
			path:     ".hidden.go",
			expected: true, // We still run tests for hidden Go files
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file
			tmpFile := filepath.Join(t.TempDir(), tt.path)
			if err := os.WriteFile(tmpFile, []byte(""), 0644); err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			got := runner.shouldRunTests(tmpFile)
			if got != tt.expected {
				t.Errorf("shouldRunTests(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}

func TestRunner_WatchMode(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "go-sentinel-watch-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Initialize Go module
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module watch\n\ngo 1.23\n"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create a test file
	testFile := filepath.Join(tmpDir, "watch_test.go")
	err = os.WriteFile(testFile, []byte(`package watch

import "testing"

func TestWatch(t *testing.T) {
	// This test should pass
}`), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a runner
	runner, err := NewRunner(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}
	defer runner.Stop()

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Start watch mode in a goroutine
	errCh := make(chan error)
	go func() {
		errCh <- runner.Run(ctx, RunOptions{Watch: true})
	}()

	// Wait a bit for the initial run
	time.Sleep(500 * time.Millisecond)

	// Modify the test file
	err = os.WriteFile(testFile, []byte(`package watch

import "testing"

func TestWatch(t *testing.T) {
	// Modified test
}`), 0644)
	if err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Wait for the run to complete or timeout
	select {
	case err := <-errCh:
		if err != nil && err != context.DeadlineExceeded {
			t.Errorf("Unexpected error from watch mode: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("Watch mode test timed out")
	}
}

func TestRunner_FailFast(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "go-sentinel-failfast-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a test file with multiple failing tests
	testFile := filepath.Join(tmpDir, "failfast_test.go")
	err = os.WriteFile(testFile, []byte(`package failfast

import "testing"

func TestFail1(t *testing.T) {
	t.Error("First failure")
}

func TestFail2(t *testing.T) {
	t.Error("Second failure")
}`), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a runner
	runner, err := NewRunner(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}
	defer runner.Stop()

	// Run with failfast
	ctx := context.Background()
	err = runner.Run(ctx, RunOptions{FailFast: true})

	// We expect an error because tests fail
	if err == nil {
		t.Error("Expected error from failing tests, got nil")
	}
}
