// Package coordinator provides watch mode orchestration and coordination
package coordinator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/processor"
	"github.com/newbpydev/go-sentinel/internal/test/runner"
	"github.com/newbpydev/go-sentinel/internal/watch/core"
	"github.com/newbpydev/go-sentinel/internal/watch/debouncer"
	"github.com/newbpydev/go-sentinel/internal/watch/watcher"
)

// TestWatchCoordinator runs tests in watch mode and implements core.WatchCoordinator
type TestWatchCoordinator struct {
	options       core.WatchOptions
	fileWatcher   core.FileSystemWatcher
	testRunner    runner.TestRunnerInterface
	testFinder    core.TestFileFinder
	processor     *processor.TestProcessor
	debouncer     core.EventDebouncer
	terminalWidth int
	status        core.WatchStatus
}

// NewTestWatchCoordinator creates a new TestWatchCoordinator
func NewTestWatchCoordinator(options core.WatchOptions) (*TestWatchCoordinator, error) {
	if options.Writer == nil {
		options.Writer = os.Stdout
	}

	if options.DebounceInterval == 0 {
		options.DebounceInterval = 500 * time.Millisecond
	}

	// Set default test patterns if not provided
	if len(options.TestPatterns) == 0 {
		options.TestPatterns = []string{"*_test.go"}
	}

	// Set default ignore patterns if not provided
	if len(options.IgnorePatterns) == 0 {
		options.IgnorePatterns = []string{
			"*/vendor/*",
			"*/.git/*",
			"*/node_modules/*",
		}
	}

	// Create file watcher
	fsWatcher, err := watcher.NewFileSystemWatcher(options.Paths, options.IgnorePatterns)
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Set up test finder
	rootDir := "."
	if len(options.Paths) > 0 {
		rootDir = options.Paths[0]
	}
	finder := watcher.NewTestFileFinder(rootDir)

	// Get terminal width
	terminalWidth := 80

	// Create test runner
	testRunner := runner.NewTestRunner(true, true) // verbose and JSON output

	// Create debouncer
	eventDebouncer := debouncer.NewFileEventDebouncer(options.DebounceInterval)

	// Initialize status
	status := core.WatchStatus{
		IsRunning:     false,
		WatchedPaths:  options.Paths,
		Mode:          options.Mode,
		StartTime:     time.Now(),
		LastEventTime: time.Time{},
		EventCount:    0,
		ErrorCount:    0,
	}

	return &TestWatchCoordinator{
		options:       options,
		fileWatcher:   fsWatcher,
		testRunner:    testRunner,
		testFinder:    finder,
		debouncer:     eventDebouncer,
		terminalWidth: terminalWidth,
		status:        status,
	}, nil
}

// Start begins watching for file changes and implements core.WatchCoordinator.Start
func (w *TestWatchCoordinator) Start(ctx context.Context) error {
	w.status.IsRunning = true
	defer func() { w.status.IsRunning = false }()

	// Display initial status message
	w.printStatus("Watching for file changes...")

	// Create channel for file events
	fileEvents := make(chan core.FileEvent, 100)

	// Start watching for file changes
	watchCtx, watchCancel := context.WithCancel(ctx)
	defer watchCancel()

	go func() {
		if err := w.fileWatcher.Watch(watchCtx, fileEvents); err != nil && err != context.Canceled {
			fmt.Fprintf(w.options.Writer, "Watch error: %v\n", err)
		}
	}()

	// Start event processing
	go w.processFileEvents(fileEvents)

	// Run tests on start if in WatchAll mode or RunOnStart is enabled
	if w.options.Mode == core.WatchAll || w.options.RunOnStart {
		w.runAllTests()
	}

	// Process debounced events
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case debouncedEvents := <-w.debouncer.Events():
			if err := w.HandleFileChanges(debouncedEvents); err != nil {
				fmt.Fprintf(w.options.Writer, "Error handling file changes: %v\n", err)
			}
		}
	}
}

// Stop stops the watcher and implements core.WatchCoordinator.Stop
func (w *TestWatchCoordinator) Stop() error {
	var errs []error

	if err := w.debouncer.Stop(); err != nil {
		errs = append(errs, fmt.Errorf("debouncer stop error: %w", err))
	}

	if err := w.fileWatcher.Close(); err != nil {
		errs = append(errs, fmt.Errorf("file watcher close error: %w", err))
	}

	w.status.IsRunning = false

	if len(errs) > 0 {
		return fmt.Errorf("errors during stop: %v", errs)
	}
	return nil
}

// HandleFileChanges processes a batch of file changes and implements core.WatchCoordinator.HandleFileChanges
func (w *TestWatchCoordinator) HandleFileChanges(changes []core.FileEvent) error {
	if len(changes) == 0 {
		return nil
	}

	// Update status
	w.status.LastEventTime = time.Now()

	// Clear terminal if requested
	if w.options.ClearTerminal {
		w.clearTerminal()
	}

	// Process each change
	for _, change := range changes {
		// Display file change notification
		fileName := filepath.Base(change.Path)
		w.printStatus(fmt.Sprintf("File changed: %s", fileName))

		// Run appropriate tests based on watch mode
		switch w.options.Mode {
		case core.WatchAll:
			w.runAllTests()

		case core.WatchChanged:
			if err := w.runTestsForFile(change.Path); err != nil {
				fmt.Fprintf(w.options.Writer, "Error running tests for %s: %v\n", change.Path, err)
			}

		case core.WatchRelated:
			if err := w.runRelatedTests(change.Path); err != nil {
				fmt.Fprintf(w.options.Writer, "Error running related tests for %s: %v\n", change.Path, err)
			}
		}

		w.status.EventCount++
	}

	// Display watch mode info
	w.printWatchInfo()
	return nil
}

// Configure updates the watch system configuration and implements core.WatchCoordinator.Configure
func (w *TestWatchCoordinator) Configure(options core.WatchOptions) error {
	w.options = options

	// Update debouncer interval
	w.debouncer.SetInterval(options.DebounceInterval)

	// TODO: Update file watcher paths if needed
	// This would require stopping and restarting the watcher

	return nil
}

// GetStatus returns the current status and implements core.WatchCoordinator.GetStatus
func (w *TestWatchCoordinator) GetStatus() core.WatchStatus {
	return w.status
}

// processFileEvents processes incoming file events through the debouncer
func (w *TestWatchCoordinator) processFileEvents(fileEvents <-chan core.FileEvent) {
	for event := range fileEvents {
		w.debouncer.AddEvent(event)
	}
}

// runAllTests runs all tests in the workspace
func (w *TestWatchCoordinator) runAllTests() {
	w.printStatus("Running all tests...")

	// Run tests for all packages
	packages := []string{"./..."}
	if err := w.executeTests(packages); err != nil {
		fmt.Fprintf(w.options.Writer, "Error running all tests: %v\n", err)
	}
}

// runTestsForFile runs tests for a specific file
func (w *TestWatchCoordinator) runTestsForFile(filePath string) error {
	var testTargets []string

	if w.testFinder.IsTestFile(filePath) {
		// It's a test file, run it directly
		testTargets = []string{filepath.Dir(filePath)}
	} else {
		// Find corresponding test file
		testFile, err := w.testFinder.FindTestFile(filePath)
		if err != nil {
			// No specific test file found, run all tests in the package
			testTargets = []string{filepath.Dir(filePath)}
		} else {
			testTargets = []string{filepath.Dir(testFile)}
		}
	}

	w.printStatus(fmt.Sprintf("Running tests for: %s", filepath.Base(filePath)))
	return w.executeTests(testTargets)
}

// runRelatedTests runs tests related to the changed file
func (w *TestWatchCoordinator) runRelatedTests(filePath string) error {
	var testTargets []string

	if w.testFinder.IsTestFile(filePath) {
		// It's a test file, find its implementation and run all package tests
		implFile, err := w.testFinder.FindImplementationFile(filePath)
		if err == nil {
			// Run tests for both the test file and its package
			testTargets = append(testTargets, filepath.Dir(filePath))
			if implDir := filepath.Dir(implFile); implDir != filepath.Dir(filePath) {
				testTargets = append(testTargets, implDir)
			}
		} else {
			// Just run the test file
			testTargets = []string{filepath.Dir(filePath)}
		}
	} else {
		// It's a source file, find all related tests
		packageTests, err := w.testFinder.FindPackageTests(filePath)
		if err != nil {
			// No tests found, run package tests
			testTargets = []string{filepath.Dir(filePath)}
		} else {
			// Add unique directories
			dirSet := make(map[string]bool)
			for _, testFile := range packageTests {
				dirSet[filepath.Dir(testFile)] = true
			}
			for dir := range dirSet {
				testTargets = append(testTargets, dir)
			}
		}
	}

	w.printStatus(fmt.Sprintf("Running related tests for: %s", filepath.Base(filePath)))
	return w.executeTests(testTargets)
}

// executeTests runs tests for the specified targets
func (w *TestWatchCoordinator) executeTests(targets []string) error {
	if len(targets) == 0 {
		return nil
	}

	// Remove duplicates
	uniqueTargets := make([]string, 0, len(targets))
	seen := make(map[string]bool)
	for _, target := range targets {
		if !seen[target] {
			uniqueTargets = append(uniqueTargets, target)
			seen[target] = true
		}
	}

	// Create context for test execution
	ctx := context.Background()

	// Execute tests using the test runner
	output, err := w.testRunner.Run(ctx, uniqueTargets)
	if err != nil {
		return fmt.Errorf("test execution failed: %w", err)
	}

	// Process the JSON output if we have a processor configured and there's output
	if w.processor != nil && output != "" {
		if err := w.processor.ProcessJSONOutput(output); err != nil {
			fmt.Fprintf(w.options.Writer, "Error processing test output: %v\n", err)
		}

		if err := w.processor.RenderResults(true); err != nil {
			fmt.Fprintf(w.options.Writer, "Error rendering results: %v\n", err)
		}
	}

	return nil
}

// printStatus displays a status message
func (w *TestWatchCoordinator) printStatus(message string) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Fprintf(w.options.Writer, "[%s] %s\n", timestamp, message)
}

// printWatchInfo displays information about the watch mode
func (w *TestWatchCoordinator) printWatchInfo() {
	fmt.Fprintf(w.options.Writer, "\n👀 Watching for changes... (mode: %s)\n", w.options.Mode)
	fmt.Fprintf(w.options.Writer, "   Press Ctrl+C to exit\n\n")
}

// clearTerminal clears the terminal screen
func (w *TestWatchCoordinator) clearTerminal() {
	fmt.Fprint(w.options.Writer, "\033[2J\033[H") // ANSI escape codes
}

// Ensure TestWatchCoordinator implements the WatchCoordinator interface
var _ core.WatchCoordinator = (*TestWatchCoordinator)(nil)
