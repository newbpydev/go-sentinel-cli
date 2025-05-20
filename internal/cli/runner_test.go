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
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

	// Initialize Go module
	if err = os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example\n\ngo 1.23\n"), 0600); err != nil {
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
}`), 0600)
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
			if err := os.WriteFile(tmpFile, []byte(""), 0600); err != nil {
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
	dir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(dir, "test.go")
	err := os.WriteFile(testFile, []byte(`package test
func Add(a, b int) int { return a + b }`), 0600)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Create a test runner with watch mode enabled
	runner, err := NewRunner(dir)
	if err != nil {
		t.Fatalf("Failed to create runner: %v", err)
	}

	// Start the runner in a goroutine
	done := make(chan bool)
	go func() {
		if err := runner.Run(context.Background(), RunOptions{Watch: true}); err != nil {
			t.Errorf("Failed to run tests in watch mode: %v", err)
		}
		done <- true
	}()

	// Wait for a short time to let the runner start
	time.Sleep(100 * time.Millisecond)

	// Modify the test file
	err = os.WriteFile(testFile, []byte(`package test
func Add(a, b int) int { return a + b + 1 }`), 0600)
	if err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Wait for up to 10 seconds for the runner to complete
	select {
	case <-done:
		// Success
	case <-time.After(10 * time.Second):
		t.Fatal("Watch mode test timed out")
	}
}

func TestRunner_FailFast(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "go-sentinel-failfast-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

	// Create a test file with multiple failing tests
	testFile := filepath.Join(tmpDir, "failfast_test.go")
	err = os.WriteFile(testFile, []byte(`package failfast

import "testing"

func TestFail1(t *testing.T) {
	t.Error("First failure")
}

func TestFail2(t *testing.T) {
	t.Error("Second failure")
}`), 0600)
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

func TestRunner_Run(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "go-sentinel-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temp dir: %v", err)
		}
	}()

	// Create go.mod
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example\n\ngo 1.23\n"), 0600); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}

	// Create test file
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
}`), 0600)
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
