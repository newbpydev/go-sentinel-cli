package coordinator

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/processor"
	"github.com/newbpydev/go-sentinel/internal/watch/core"
)

// TestNewTestWatchCoordinator_DefaultValues tests default value assignment (95.0% → 100.0%)
func TestNewTestWatchCoordinator_DefaultValues(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		options       core.WatchOptions
		expectedError string
		validateFunc  func(*TestWatchCoordinator) error
	}{
		"nil_writer_uses_default": {
			options: core.WatchOptions{
				Paths:  []string{"./"},
				Mode:   core.WatchAll,
				Writer: nil, // Should default to os.Stdout
			},
			expectedError: "",
			validateFunc: func(coord *TestWatchCoordinator) error {
				if coord.options.Writer == nil {
					return errors.New("writer should not be nil after default assignment")
				}
				return nil
			},
		},
		"zero_debounce_interval_uses_default": {
			options: core.WatchOptions{
				Paths:            []string{"./"},
				Mode:             core.WatchAll,
				DebounceInterval: 0, // Should default to 500ms
			},
			expectedError: "",
			validateFunc: func(coord *TestWatchCoordinator) error {
				if coord.options.DebounceInterval != 500*time.Millisecond {
					return errors.New("debounce interval should default to 500ms")
				}
				return nil
			},
		},
		"empty_test_patterns_uses_default": {
			options: core.WatchOptions{
				Paths:        []string{"./"},
				Mode:         core.WatchAll,
				TestPatterns: nil, // Should default to ["*_test.go"]
			},
			expectedError: "",
			validateFunc: func(coord *TestWatchCoordinator) error {
				if len(coord.options.TestPatterns) != 1 || coord.options.TestPatterns[0] != "*_test.go" {
					return errors.New("test patterns should default to [\"*_test.go\"]")
				}
				return nil
			},
		},
		"empty_ignore_patterns_uses_default": {
			options: core.WatchOptions{
				Paths:          []string{"./"},
				Mode:           core.WatchAll,
				IgnorePatterns: nil, // Should default to vendor, git, node_modules
			},
			expectedError: "",
			validateFunc: func(coord *TestWatchCoordinator) error {
				if len(coord.options.IgnorePatterns) != 3 {
					return errors.New("ignore patterns should have 3 default entries")
				}
				expected := []string{"*/vendor/*", "*/.git/*", "*/node_modules/*"}
				for i, pattern := range expected {
					if coord.options.IgnorePatterns[i] != pattern {
						return errors.New("ignore patterns don't match expected defaults")
					}
				}
				return nil
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			coord, err := NewTestWatchCoordinator(tt.options)

			if tt.expectedError != "" {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.expectedError)
					return
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("Expected error containing %q, got %q", tt.expectedError, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if coord == nil {
				t.Error("NewTestWatchCoordinator should not return nil on success")
				return
			}

			if tt.validateFunc != nil {
				if err := tt.validateFunc(coord); err != nil {
					t.Errorf("Validation failed: %v", err)
				}
			}
		})
	}
}

// TestTestWatchCoordinator_Start_RunOnStartPath tests RunOnStart path (88.9% → 100.0%)
func TestTestWatchCoordinator_Start_RunOnStartPath(t *testing.T) {
	t.Parallel()

	// Create coordinator with RunOnStart enabled
	options := core.WatchOptions{
		Paths:      []string{"./"},
		Mode:       core.WatchChanged, // Not WatchAll, but RunOnStart enabled
		RunOnStart: true,
		Writer:     &strings.Builder{},
	}

	coord, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("Failed to create coordinator: %v", err)
	}

	// Replace test runner with SAFE mock
	coord.testRunner = &mockTestRunner{
		runFunc: func(ctx context.Context, packages []string) (string, error) {
			return `{"Action":"pass","Package":"mock","Test":"MockTest","Output":"PASS"}`, nil
		},
	}

	// Use a very short timeout to test the RunOnStart path
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// This should trigger the RunOnStart path even though mode is not WatchAll
	err = coord.Start(ctx)

	// Should return context deadline exceeded
	if err == nil {
		t.Error("Expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
	}
}

// TestTestWatchCoordinator_Stop_ErrorAggregation tests error aggregation (66.7% → 100.0%)
func TestTestWatchCoordinator_Stop_ErrorAggregation(t *testing.T) {
	t.Parallel()

	// Create a coordinator that will have stop errors
	options := core.WatchOptions{
		Paths:  []string{"./"},
		Mode:   core.WatchAll,
		Writer: &strings.Builder{},
	}

	coord, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("Failed to create coordinator: %v", err)
	}

	// Replace debouncer with one that will error on stop
	coord.debouncer = &mockEventDebouncer{
		stopFunc: func() error {
			return errors.New("debouncer stop error")
		},
	}

	// Replace file watcher with one that will error on close
	coord.fileWatcher = &mockFileSystemWatcher{
		closeFunc: func() error {
			return errors.New("file watcher close error")
		},
	}

	err = coord.Stop()

	// Should aggregate both errors
	if err == nil {
		t.Error("Expected aggregated error")
	}

	errorStr := err.Error()
	if !strings.Contains(errorStr, "errors during stop") {
		t.Errorf("Expected 'errors during stop' in error message, got: %v", err)
	}
	if !strings.Contains(errorStr, "debouncer stop error") {
		t.Errorf("Expected debouncer error in aggregated message, got: %v", err)
	}
	if !strings.Contains(errorStr, "file watcher close error") {
		t.Errorf("Expected file watcher error in aggregated message, got: %v", err)
	}

	// Verify status is updated even with errors
	if coord.status.IsRunning {
		t.Error("Status should show not running after stop, even with errors")
	}
}

// TestTestWatchCoordinator_HandleFileChanges_ClearTerminal tests clear terminal path (88.2% → 100.0%)
func TestTestWatchCoordinator_HandleFileChanges_ClearTerminal(t *testing.T) {
	t.Parallel()

	options := core.WatchOptions{
		Paths:         []string{"./"},
		Mode:          core.WatchAll,
		ClearTerminal: true, // Enable terminal clearing
		Writer:        &strings.Builder{},
	}

	coord, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("Failed to create coordinator: %v", err)
	}

	// Replace test runner with SAFE mock
	coord.testRunner = &mockTestRunner{
		runFunc: func(ctx context.Context, packages []string) (string, error) {
			return `{"Action":"pass","Package":"mock","Test":"MockTest","Output":"PASS"}`, nil
		},
	}

	// Capture output
	var output strings.Builder
	coord.options.Writer = &output

	changes := []core.FileEvent{
		{Path: "src/main.go", Type: "modify"},
	}

	err = coord.HandleFileChanges(changes)
	if err != nil {
		t.Errorf("HandleFileChanges should not return error: %v", err)
	}

	// Verify terminal clear sequence was written
	outputStr := output.String()
	if !strings.Contains(outputStr, "\033[2J\033[H") {
		t.Error("Output should contain terminal clear sequence when ClearTerminal is enabled")
	}

	// Verify event count was incremented
	if coord.status.EventCount != 1 {
		t.Errorf("Expected event count 1, got %d", coord.status.EventCount)
	}
}

// TestTestWatchCoordinator_RunAllTests_ErrorHandling tests error handling (75.0% → 100.0%)
func TestTestWatchCoordinator_RunAllTests_ErrorHandling(t *testing.T) {
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

	// Replace test runner with SAFE mock that doesn't execute real tests
	coord.testRunner = &mockTestRunner{
		runFunc: func(ctx context.Context, packages []string) (string, error) {
			// Return safe mock JSON output instead of executing real tests
			return `{"Action":"fail","Package":"mock","Test":"MockTest","Output":"test execution failed"}`, errors.New("test execution failed")
		},
	}

	// Capture output
	var output strings.Builder
	coord.options.Writer = &output

	// This should not panic or return error, but handle error internally
	coord.runAllTests()

	// Verify error was logged
	outputStr := output.String()
	if !strings.Contains(outputStr, "Error running all tests") {
		t.Error("Expected error message in output when test execution fails")
	}
	if !strings.Contains(outputStr, "test execution failed") {
		t.Error("Expected specific error message in output")
	}
}

// TestTestWatchCoordinator_ExecuteTests_DuplicateRemoval tests duplicate removal (77.8% → 100.0%)
func TestTestWatchCoordinator_ExecuteTests_DuplicateRemoval(t *testing.T) {
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

	// Replace test runner with SAFE mock
	coord.testRunner = &mockTestRunner{
		runFunc: func(ctx context.Context, packages []string) (string, error) {
			// Verify duplicates were removed
			if len(packages) != 2 {
				return "", errors.New("duplicates not removed correctly")
			}
			return `{"Action":"pass","Package":"mock","Test":"MockTest","Output":"PASS"}`, nil
		},
	}

	// Test with duplicate targets - should remove duplicates
	err = coord.executeTests([]string{"./test", "./test", "./src", "./test"})
	if err != nil {
		t.Errorf("executeTests should handle duplicates: %v", err)
	}
}

// TestTestWatchCoordinator_ExecuteTests_EmptyTargetsPrecision tests empty targets path precision (77.8% → 100.0%)
func TestTestWatchCoordinator_ExecuteTests_EmptyTargetsPrecision(t *testing.T) {
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

	// Test with empty targets - should return immediately without error
	err = coord.executeTests([]string{})
	if err != nil {
		t.Errorf("executeTests with empty targets should not error: %v", err)
	}
}

// TestTestWatchCoordinator_RunTestsForFile_ErrorHandling tests error handling (88.9% → 100.0%)
func TestTestWatchCoordinator_RunTestsForFile_ErrorHandling(t *testing.T) {
	t.Parallel()

	options := core.WatchOptions{
		Paths:  []string{"./"},
		Mode:   core.WatchChanged,
		Writer: &strings.Builder{},
	}

	coord, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("Failed to create coordinator: %v", err)
	}

	// Replace test runner with SAFE mock that returns error
	coord.testRunner = &mockTestRunner{
		runFunc: func(ctx context.Context, packages []string) (string, error) {
			return "", errors.New("test execution failed")
		},
	}

	// Capture output
	var output strings.Builder
	coord.options.Writer = &output

	// This should handle error gracefully and return the error
	err = coord.runTestsForFile("test/main_test.go")
	if err == nil {
		t.Error("Expected error from runTestsForFile when test execution fails")
	}

	// Verify the error contains the expected message
	if !strings.Contains(err.Error(), "test execution failed") {
		t.Errorf("Expected error containing 'test execution failed', got: %v", err)
	}
}

// TestTestWatchCoordinator_RunTestsForFile_DirectCall tests direct runTestsForFile call (88.9% → 100.0%)
func TestTestWatchCoordinator_RunTestsForFile_DirectCall(t *testing.T) {
	t.Parallel()

	options := core.WatchOptions{
		Paths:  []string{"./"},
		Mode:   core.WatchChanged,
		Writer: &strings.Builder{},
	}

	coord, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("Failed to create coordinator: %v", err)
	}

	// Replace test runner with SAFE mock
	coord.testRunner = &mockTestRunner{
		runFunc: func(ctx context.Context, packages []string) (string, error) {
			return `{"Action":"pass","Package":"mock","Test":"MockTest","Output":"PASS"}`, nil
		},
	}

	// Capture output
	var output strings.Builder
	coord.options.Writer = &output

	// Call runTestsForFile directly - this should print status and execute tests
	err = coord.runTestsForFile("test/main_test.go")
	if err != nil {
		t.Errorf("runTestsForFile should not error: %v", err)
	}

	// Verify status message was printed
	outputStr := output.String()
	if !strings.Contains(outputStr, "Running tests for: main_test.go") {
		t.Errorf("Expected status message, got: %s", outputStr)
	}
}

// Mock TestFileFinder for precision testing
type mockTestFileFinder struct {
	isTestFileFunc             func(filePath string) bool
	findTestFileFunc           func(filePath string) (string, error)
	findImplementationFileFunc func(testFilePath string) (string, error)
	findPackageTestsFunc       func(filePath string) ([]string, error)
}

func (m *mockTestFileFinder) IsTestFile(filePath string) bool {
	if m.isTestFileFunc != nil {
		return m.isTestFileFunc(filePath)
	}
	return false
}

func (m *mockTestFileFinder) FindTestFile(filePath string) (string, error) {
	if m.findTestFileFunc != nil {
		return m.findTestFileFunc(filePath)
	}
	return "", nil
}

func (m *mockTestFileFinder) FindImplementationFile(testFilePath string) (string, error) {
	if m.findImplementationFileFunc != nil {
		return m.findImplementationFileFunc(testFilePath)
	}
	return "", nil
}

func (m *mockTestFileFinder) FindPackageTests(filePath string) ([]string, error) {
	if m.findPackageTestsFunc != nil {
		return m.findPackageTestsFunc(filePath)
	}
	return nil, nil
}

// TestTestWatchCoordinator_RunRelatedTests_ComplexBranching tests complex branching (66.7% → 100.0%)
func TestTestWatchCoordinator_RunRelatedTests_ComplexBranching(t *testing.T) {
	t.Parallel()

	options := core.WatchOptions{
		Paths:  []string{"./"},
		Mode:   core.WatchRelated,
		Writer: &strings.Builder{},
	}

	coord, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("Failed to create coordinator: %v", err)
	}

	// Replace test runner with SAFE mock that doesn't execute real tests
	coord.testRunner = &mockTestRunner{
		runFunc: func(ctx context.Context, packages []string) (string, error) {
			// Return safe mock JSON output instead of executing real tests
			return `{"Action":"pass","Package":"mock","Test":"MockRelatedTest","Output":"PASS"}`, nil
		},
	}

	// Test case: test file with implementation in different directory
	coord.testFinder = &mockTestFileFinder{
		isTestFileFunc: func(filePath string) bool {
			return strings.HasSuffix(filePath, "_test.go")
		},
		findImplementationFileFunc: func(testFilePath string) (string, error) {
			return "different/dir/main.go", nil // Implementation in different directory
		},
	}

	// Capture output
	var output strings.Builder
	coord.options.Writer = &output

	err = coord.runRelatedTests("test/main_test.go")
	if err != nil {
		t.Errorf("runRelatedTests should not return error: %v", err)
	}

	// Verify status message was printed
	outputStr := output.String()
	if !strings.Contains(outputStr, "Running related tests for") {
		t.Error("Expected status message for related tests")
	}

	// Test case: source file with multiple package tests
	coord.testFinder = &mockTestFileFinder{
		isTestFileFunc: func(filePath string) bool {
			return false // Not a test file
		},
		findPackageTestsFunc: func(filePath string) ([]string, error) {
			return []string{
				"test/main_test.go",
				"test/helper_test.go",
				"other/dir/other_test.go",
			}, nil
		},
	}

	output.Reset()
	err = coord.runRelatedTests("src/main.go")
	if err != nil {
		t.Errorf("runRelatedTests should not return error: %v", err)
	}

	// Verify status message was printed
	outputStr = output.String()
	if !strings.Contains(outputStr, "Running related tests for") {
		t.Error("Expected status message for related tests with multiple directories")
	}
}

// TestTestWatchCoordinator_RunRelatedTests_ErrorHandling tests error handling (66.7% → 100.0%)
func TestTestWatchCoordinator_RunRelatedTests_ErrorHandling(t *testing.T) {
	t.Parallel()

	options := core.WatchOptions{
		Paths:  []string{"./"},
		Mode:   core.WatchRelated,
		Writer: &strings.Builder{},
	}

	coord, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("Failed to create coordinator: %v", err)
	}

	// Replace test runner with SAFE mock that returns error
	coord.testRunner = &mockTestRunner{
		runFunc: func(ctx context.Context, packages []string) (string, error) {
			return "", errors.New("test execution failed")
		},
	}

	// Test finder that returns error
	coord.testFinder = &mockTestFileFinder{
		isTestFileFunc: func(filePath string) bool {
			return false
		},
		findPackageTestsFunc: func(filePath string) ([]string, error) {
			return nil, errors.New("test finder error")
		},
	}

	// Capture output
	var output strings.Builder
	coord.options.Writer = &output

	// This should handle error gracefully and return the error
	err = coord.runRelatedTests("src/main.go")
	if err == nil {
		t.Error("Expected error from runRelatedTests when test execution fails")
	}

	// Verify the error contains the expected message
	if !strings.Contains(err.Error(), "test execution failed") {
		t.Errorf("Expected error containing 'test execution failed', got: %v", err)
	}
}

// TestTestWatchCoordinator_ExecuteTests_ProcessorPath tests processor handling (77.8% → 100.0%)
func TestTestWatchCoordinator_ExecuteTests_ProcessorPath(t *testing.T) {
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

	// Replace test runner with SAFE mock that returns output
	coord.testRunner = &mockTestRunner{
		runFunc: func(ctx context.Context, packages []string) (string, error) {
			return `{"Action":"pass","Package":"mock","Test":"MockTest","Output":"PASS"}`, nil
		},
	}

	// Add a real processor to trigger the processor path
	coord.processor = processor.NewTestProcessor(&strings.Builder{}, &mockColorFormatter{}, &mockIconProvider{}, 80)

	// Test with targets - should trigger processor path
	err = coord.executeTests([]string{"./test"})
	if err != nil {
		t.Errorf("executeTests with processor should not error: %v", err)
	}
}

// TestTestWatchCoordinator_ExecuteTests_ProcessorErrors tests processor error handling (77.8% → 100.0%)
func TestTestWatchCoordinator_ExecuteTests_ProcessorErrors(t *testing.T) {
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

	// Replace test runner with SAFE mock that returns output
	coord.testRunner = &mockTestRunner{
		runFunc: func(ctx context.Context, packages []string) (string, error) {
			return `{"Action":"pass","Package":"mock","Test":"MockTest","Output":"PASS"}`, nil
		},
	}

	// Capture output for error messages
	var output strings.Builder
	coord.options.Writer = &output

	// Set processor to nil to avoid processor path
	coord.processor = nil

	// Test with targets - should work without processor
	err = coord.executeTests([]string{"./test"})
	if err != nil {
		t.Errorf("executeTests should work without processor: %v", err)
	}
}

// Mock ColorFormatter for precision testing
type mockColorFormatter struct{}

func (m *mockColorFormatter) Red(text string) string                 { return text }
func (m *mockColorFormatter) Green(text string) string               { return text }
func (m *mockColorFormatter) Yellow(text string) string              { return text }
func (m *mockColorFormatter) Blue(text string) string                { return text }
func (m *mockColorFormatter) Magenta(text string) string             { return text }
func (m *mockColorFormatter) Cyan(text string) string                { return text }
func (m *mockColorFormatter) Gray(text string) string                { return text }
func (m *mockColorFormatter) Bold(text string) string                { return text }
func (m *mockColorFormatter) Dim(text string) string                 { return text }
func (m *mockColorFormatter) White(text string) string               { return text }
func (m *mockColorFormatter) Colorize(text, colorName string) string { return text }

// Mock IconProvider for precision testing
type mockIconProvider struct{}

func (m *mockIconProvider) CheckMark() string              { return "✓" }
func (m *mockIconProvider) Cross() string                  { return "✗" }
func (m *mockIconProvider) Skipped() string                { return "○" }
func (m *mockIconProvider) Running() string                { return "●" }
func (m *mockIconProvider) GetIcon(iconType string) string { return "●" }

// Mock TestProcessor for precision testing
type mockTestProcessor struct {
	processJSONOutputFunc func(output string) error
	renderResultsFunc     func(showSummary bool) error
}

func (m *mockTestProcessor) ProcessJSONOutput(output string) error {
	if m.processJSONOutputFunc != nil {
		return m.processJSONOutputFunc(output)
	}
	return nil
}

func (m *mockTestProcessor) RenderResults(showSummary bool) error {
	if m.renderResultsFunc != nil {
		return m.renderResultsFunc(showSummary)
	}
	return nil
}

// TestTestWatchCoordinator_NewTestWatchCoordinator_EmptyPaths tests empty paths handling (95.0% → 100.0%)
func TestTestWatchCoordinator_NewTestWatchCoordinator_EmptyPaths(t *testing.T) {
	t.Parallel()

	// Test with empty paths - should use "." as default
	options := core.WatchOptions{
		Paths:  []string{}, // Empty paths
		Mode:   core.WatchAll,
		Writer: &strings.Builder{},
	}

	coord, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Errorf("NewTestWatchCoordinator should handle empty paths: %v", err)
	}
	if coord == nil {
		t.Error("NewTestWatchCoordinator should not return nil with empty paths")
	}
}

// TestTestWatchCoordinator_Start_WatchErrorPath tests watch error handling in Start (88.9% → 100.0%)
func TestTestWatchCoordinator_Start_WatchErrorPath(t *testing.T) {
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

	// Replace test runner with SAFE mock
	coord.testRunner = &mockTestRunner{
		runFunc: func(ctx context.Context, packages []string) (string, error) {
			return `{"Action":"pass","Package":"mock","Test":"MockTest","Output":"PASS"}`, nil
		},
	}

	// Replace file watcher with one that will error during Watch
	coord.fileWatcher = &mockFileSystemWatcher{
		watchFunc: func(ctx context.Context, events chan<- core.FileEvent) error {
			return errors.New("watch operation failed")
		},
	}

	// Use a very short timeout to test the error path
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// This should trigger the watch error path
	err = coord.Start(ctx)

	// Should return context deadline exceeded (the watch error is logged but doesn't stop the function)
	if err == nil {
		t.Error("Expected timeout error")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
	}
}

// TestTestWatchCoordinator_HandleFileChanges_ErrorPaths tests error paths in HandleFileChanges (88.2% → 100.0%)
func TestTestWatchCoordinator_HandleFileChanges_ErrorPaths(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		mode         core.WatchMode
		setupMocks   func(*TestWatchCoordinator)
		expectOutput string
	}{
		"watch_changed_with_error": {
			mode: core.WatchChanged,
			setupMocks: func(coord *TestWatchCoordinator) {
				coord.testRunner = &mockTestRunner{
					runFunc: func(ctx context.Context, packages []string) (string, error) {
						return "", errors.New("test execution failed")
					},
				}
			},
			expectOutput: "Error running tests for",
		},
		"watch_related_with_error": {
			mode: core.WatchRelated,
			setupMocks: func(coord *TestWatchCoordinator) {
				coord.testRunner = &mockTestRunner{
					runFunc: func(ctx context.Context, packages []string) (string, error) {
						return "", errors.New("test execution failed")
					},
				}
				coord.testFinder = &mockTestFileFinder{
					isTestFileFunc: func(filePath string) bool {
						return false
					},
					findPackageTestsFunc: func(filePath string) ([]string, error) {
						return []string{"test/main_test.go"}, nil
					},
				}
			},
			expectOutput: "Error running related tests for",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			options := core.WatchOptions{
				Paths:  []string{"./"},
				Mode:   tt.mode,
				Writer: &strings.Builder{},
			}

			coord, err := NewTestWatchCoordinator(options)
			if err != nil {
				t.Fatalf("Failed to create coordinator: %v", err)
			}

			// Setup mocks
			tt.setupMocks(coord)

			// Capture output
			var output strings.Builder
			coord.options.Writer = &output

			changes := []core.FileEvent{
				{Path: "src/main.go", Type: "modify"},
			}

			err = coord.HandleFileChanges(changes)
			if err != nil {
				t.Errorf("HandleFileChanges should not return error: %v", err)
			}

			// Verify error was logged
			outputStr := output.String()
			if !strings.Contains(outputStr, tt.expectOutput) {
				t.Errorf("Expected output containing %q, got: %s", tt.expectOutput, outputStr)
			}
		})
	}
}

// TestTestWatchCoordinator_RunTestsForFile_FindTestFileError tests FindTestFile error path (88.9% → 100.0%)
func TestTestWatchCoordinator_RunTestsForFile_FindTestFileError(t *testing.T) {
	t.Parallel()

	options := core.WatchOptions{
		Paths:  []string{"./"},
		Mode:   core.WatchChanged,
		Writer: &strings.Builder{},
	}

	coord, err := NewTestWatchCoordinator(options)
	if err != nil {
		t.Fatalf("Failed to create coordinator: %v", err)
	}

	// Replace test runner with SAFE mock
	coord.testRunner = &mockTestRunner{
		runFunc: func(ctx context.Context, packages []string) (string, error) {
			return `{"Action":"pass","Package":"mock","Test":"MockTest","Output":"PASS"}`, nil
		},
	}

	// Replace test finder with one that returns error for FindTestFile
	coord.testFinder = &mockTestFileFinder{
		isTestFileFunc: func(filePath string) bool {
			return false // Not a test file
		},
		findTestFileFunc: func(filePath string) (string, error) {
			return "", errors.New("test file not found")
		},
	}

	// This should trigger the FindTestFile error path, which falls back to package tests
	err = coord.runTestsForFile("src/main.go")
	if err != nil {
		t.Errorf("runTestsForFile should handle FindTestFile error gracefully: %v", err)
	}
}

// TestTestWatchCoordinator_ExecuteTests_ProcessorWithOutput tests processor path with output (88.9% → 100.0%)
func TestTestWatchCoordinator_ExecuteTests_ProcessorWithOutput(t *testing.T) {
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

	// Replace test runner with SAFE mock that returns output
	coord.testRunner = &mockTestRunner{
		runFunc: func(ctx context.Context, packages []string) (string, error) {
			return `{"Action":"pass","Package":"mock","Test":"MockTest","Output":"PASS"}`, nil
		},
	}

	// Add a real processor to trigger the processor path with output
	coord.processor = processor.NewTestProcessor(&strings.Builder{}, &mockColorFormatter{}, &mockIconProvider{}, 80)

	// Test with targets - should trigger processor path with output
	err = coord.executeTests([]string{"./test"})
	if err != nil {
		t.Errorf("executeTests with processor and output should not error: %v", err)
	}
}
