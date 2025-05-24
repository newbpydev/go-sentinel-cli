package coordinator

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/watch/core"
)

func TestWatchCoordinatorDefaults(t *testing.T) {
	buffer := &bytes.Buffer{}

	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Check defaults were applied
	if coordinator.options.DebounceInterval != 500*time.Millisecond {
		t.Errorf("expected default debounce interval of 500ms, got %v", coordinator.options.DebounceInterval)
	}

	if len(coordinator.options.TestPatterns) != 1 || coordinator.options.TestPatterns[0] != "*_test.go" {
		t.Errorf("expected default test pattern of *_test.go, got %v", coordinator.options.TestPatterns)
	}

	if len(coordinator.options.IgnorePatterns) < 3 {
		t.Errorf("expected at least 3 default ignore patterns, got %v", coordinator.options.IgnorePatterns)
	}
}

func TestCoordinatorStatusPrinting(t *testing.T) {
	buffer := &bytes.Buffer{}

	options := core.WatchOptions{
		Paths:  []string{"."},
		Mode:   core.WatchAll,
		Writer: buffer,
	}

	coordinator, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Test status printing
	coordinator.printStatus("Test message")
	output := buffer.String()

	if !strings.Contains(output, "Test message") {
		t.Errorf("expected output to contain 'Test message', got %s", output)
	}

	// Clear buffer
	buffer.Reset()

	// Test watch info printing
	coordinator.printWatchInfo()
	output = buffer.String()

	if !strings.Contains(output, "mode: all") {
		t.Errorf("expected output to contain watch mode, got %s", output)
	}

	if !strings.Contains(output, "Press Ctrl+C to exit") {
		t.Errorf("expected output to contain exit instructions, got %s", output)
	}
}

func TestCoordinatorClearTerminal(t *testing.T) {
	buffer := &bytes.Buffer{}

	options := core.WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	coordinator, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Test terminal clearing
	coordinator.clearTerminal()
	output := buffer.String()

	// Check for ANSI escape sequence
	if output != "\033[2J\033[H" {
		t.Errorf("expected ANSI escape sequence for clearing terminal, got %q", output)
	}
}

func TestCoordinatorFileChange(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "coordinator-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Logf("failed to remove temp dir: %v", err)
		}
	}()

	// Create a test file
	testFile := filepath.Join(tempDir, "example_test.go")
	// #nosec G306 - Test file, permissions not important
	if err := os.WriteFile(testFile, []byte("package example_test\n\nfunc TestExample(t *testing.T) {}\n"), 0600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	buffer := &bytes.Buffer{}

	options := core.WatchOptions{
		Paths:         []string{tempDir},
		Mode:          core.WatchChanged,
		Writer:        buffer,
		ClearTerminal: false,
	}

	coordinator, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Test runTestsForFile
	if err := coordinator.runTestsForFile(testFile); err != nil {
		t.Errorf("runTestsForFile failed: %v", err)
	}

	output := buffer.String()
	if !strings.Contains(output, "Running tests for: example_test.go") {
		t.Errorf("expected output to contain running message, got %s", output)
	}
}

func TestCoordinatorHandleFileChanges(t *testing.T) {
	buffer := &bytes.Buffer{}

	options := core.WatchOptions{
		Paths:  []string{"."},
		Mode:   core.WatchChanged,
		Writer: buffer,
	}

	coordinator, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Test handling empty file changes
	if err := coordinator.HandleFileChanges([]core.FileEvent{}); err != nil {
		t.Errorf("HandleFileChanges with empty slice failed: %v", err)
	}

	// Check that status was not updated for empty changes
	status := coordinator.GetStatus()
	if status.EventCount != 0 {
		t.Errorf("expected event count to be 0 for empty changes, got %d", status.EventCount)
	}

	// Test that the coordinator was created with correct initial status
	if status.IsRunning {
		t.Error("expected IsRunning to be false initially")
	}

	if len(status.WatchedPaths) != 1 || status.WatchedPaths[0] != "." {
		t.Errorf("expected WatchedPaths to be ['.'], got %v", status.WatchedPaths)
	}
}

func TestCoordinatorConfigure(t *testing.T) {
	buffer := &bytes.Buffer{}

	options := core.WatchOptions{
		Paths:            []string{"."},
		DebounceInterval: 100 * time.Millisecond,
		Writer:           buffer,
	}

	coordinator, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	// Test configuration update
	newOptions := core.WatchOptions{
		Paths:            []string{"./src"},
		DebounceInterval: 200 * time.Millisecond,
		Writer:           buffer,
	}

	if err := coordinator.Configure(newOptions); err != nil {
		t.Errorf("Configure failed: %v", err)
	}

	// Check that options were updated
	if coordinator.options.DebounceInterval != 200*time.Millisecond {
		t.Errorf("expected debounce interval to be 200ms, got %v", coordinator.options.DebounceInterval)
	}
}

func TestCoordinatorGetStatus(t *testing.T) {
	buffer := &bytes.Buffer{}

	options := core.WatchOptions{
		Paths:  []string{"."},
		Mode:   core.WatchAll,
		Writer: buffer,
	}

	coordinator, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("failed to create coordinator: %v", err)
	}

	status := coordinator.GetStatus()

	// Check initial status
	if status.IsRunning {
		t.Error("expected IsRunning to be false initially")
	}

	if len(status.WatchedPaths) != 1 || status.WatchedPaths[0] != "." {
		t.Errorf("expected WatchedPaths to be ['.'], got %v", status.WatchedPaths)
	}

	if status.Mode != core.WatchAll {
		t.Errorf("expected Mode to be WatchAll, got %v", status.Mode)
	}
}
