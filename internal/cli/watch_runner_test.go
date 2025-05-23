package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWatchOptionsDefaults(t *testing.T) {
	buffer := &bytes.Buffer{}

	options := WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	watcher, err := NewTestWatcher(options)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}

	// Check defaults were applied
	if watcher.options.DebounceInterval != 500*time.Millisecond {
		t.Errorf("expected default debounce interval of 500ms, got %v", watcher.options.DebounceInterval)
	}

	if len(watcher.options.TestPatterns) != 1 || watcher.options.TestPatterns[0] != "*_test.go" {
		t.Errorf("expected default test pattern of *_test.go, got %v", watcher.options.TestPatterns)
	}

	if len(watcher.options.IgnorePatterns) < 3 {
		t.Errorf("expected at least 3 default ignore patterns, got %v", watcher.options.IgnorePatterns)
	}
}

func TestWatcherStatusPrinting(t *testing.T) {
	buffer := &bytes.Buffer{}

	options := WatchOptions{
		Paths:  []string{"."},
		Mode:   WatchAll,
		Writer: buffer,
	}

	watcher, err := NewTestWatcher(options)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}

	// Test status printing
	watcher.printStatus("Test message")
	output := buffer.String()

	if !strings.Contains(output, "Test message") {
		t.Errorf("expected output to contain 'Test message', got %s", output)
	}

	// Clear buffer
	buffer.Reset()

	// Test watch info printing
	watcher.printWatchInfo()
	output = buffer.String()

	if !strings.Contains(output, "Watch mode: all") {
		t.Errorf("expected output to contain watch mode, got %s", output)
	}

	if !strings.Contains(output, "Press 'a' to run all tests") {
		t.Errorf("expected output to contain key commands, got %s", output)
	}
}

func TestWatcherClearTerminal(t *testing.T) {
	buffer := &bytes.Buffer{}

	options := WatchOptions{
		Paths:  []string{"."},
		Writer: buffer,
	}

	watcher, err := NewTestWatcher(options)
	if err != nil {
		t.Fatalf("failed to create watcher: %v", err)
	}

	// Test terminal clearing
	watcher.clearTerminal()
	output := buffer.String()

	// Check for ANSI escape sequence
	if output != "\033[2J\033[H" {
		t.Errorf("expected ANSI escape sequence for clearing terminal, got %q", output)
	}
}

func TestWatcherFileChange(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "watcher-test")
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

	// Create a mock test runner that simulates running tests
	runner := &mockTestRunner{
		output: `{"Time":"2023-11-05T12:34:56.789Z","Action":"run","Test":"TestExample"}
{"Time":"2023-11-05T12:34:56.790Z","Action":"output","Test":"TestExample","Output":"PASS\n"}
{"Time":"2023-11-05T12:34:56.791Z","Action":"pass","Test":"TestExample"}`,
	}

	buffer := &bytes.Buffer{}

	// Create watcher with mocked components
	watcher := &TestWatcher{
		options: WatchOptions{
			Paths:         []string{tempDir},
			Mode:          WatchChanged,
			Writer:        buffer,
			ClearTerminal: false,
		},
		testRunner: runner,
		formatter:  NewColorFormatter(false),
		processor:  NewTestProcessor(buffer, NewColorFormatter(false), NewIconProvider(true), 80),
	}

	// Test runTestsForFile
	watcher.runTestsForFile(testFile)
	output := buffer.String()

	if !strings.Contains(output, "Running tests for example_test.go") {
		t.Errorf("expected output to contain running message, got %s", output)
	}
}

// mockTestRunner is a mock implementation of the test runner for testing
type mockTestRunner struct {
	output string
	err    error
}

// Run implements TestRunnerInterface
func (m *mockTestRunner) Run(ctx context.Context, testPaths []string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.output, nil
}

// RunStream implements TestRunnerInterface
func (m *mockTestRunner) RunStream(ctx context.Context, testPaths []string) (io.ReadCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Return a string reader that implements ReadCloser
	return io.NopCloser(strings.NewReader(m.output)), nil
}
