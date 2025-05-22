package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// WatchMode represents the possible watch modes
type WatchMode string

const (
	// WatchAll runs all tests when any file changes
	WatchAll WatchMode = "all"

	// WatchChanged runs tests only for changed files
	WatchChanged WatchMode = "changed"

	// WatchRelated runs tests for changed files and related files
	WatchRelated WatchMode = "related"
)

// WatchOptions configures the watch mode behavior
type WatchOptions struct {
	// Paths to watch for changes
	Paths []string

	// Patterns to ignore
	IgnorePatterns []string

	// Test file patterns
	TestPatterns []string

	// Watch mode type
	Mode WatchMode

	// Debounce interval to avoid running tests too frequently
	DebounceInterval time.Duration

	// Clear terminal between test runs
	ClearTerminal bool

	// Writer for console output
	Writer io.Writer
}

// TestWatcher runs tests in watch mode
type TestWatcher struct {
	options       WatchOptions
	fileWatcher   *FileWatcher
	testRunner    TestRunnerInterface
	testFinder    *TestFileFinder
	processor     *TestProcessor
	formatter     *ColorFormatter
	icons         *IconProvider
	terminalWidth int
}

// NewTestWatcher creates a new TestWatcher
func NewTestWatcher(options WatchOptions) (*TestWatcher, error) {
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
	watcher, err := NewFileWatcher(options.Paths, options.IgnorePatterns)
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	// Set up test finder
	rootDir := "."
	if len(options.Paths) > 0 {
		rootDir = options.Paths[0]
	}
	finder := NewTestFileFinder(rootDir)

	// Set up formatters
	formatter := NewColorFormatter(true)
	icons := NewIconProvider(true)

	// Get terminal width
	terminalWidth := 80

	// Create test runner
	runner := &TestRunner{
		Verbose:    true,
		JSONOutput: true,
	}

	// Create test processor
	processor := NewTestProcessor(options.Writer, formatter, icons, terminalWidth)

	return &TestWatcher{
		options:       options,
		fileWatcher:   watcher,
		testRunner:    runner,
		testFinder:    finder,
		processor:     processor,
		formatter:     formatter,
		icons:         icons,
		terminalWidth: terminalWidth,
	}, nil
}

// Start begins watching for file changes and running tests
func (w *TestWatcher) Start(ctx context.Context) error {
	// Display initial status message
	w.printStatus("Watching for file changes...")

	// Create channels for file events and debounced events
	fileEvents := make(chan FileEvent, 100)
	debouncedEvents := make(chan string, 10)

	// Start watching for file changes
	go func() {
		if err := w.fileWatcher.Watch(fileEvents); err != nil {
			fmt.Fprintf(w.options.Writer, "Watch error: %v\n", err)
		}
	}()

	// Set up debouncer
	go w.debounceEvents(fileEvents, debouncedEvents)

	// Run tests on start if in WatchAll mode
	if w.options.Mode == WatchAll {
		w.runAllTests()
	}

	// Process debounced events
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case filePath := <-debouncedEvents:
			// Clear terminal if requested
			if w.options.ClearTerminal {
				w.clearTerminal()
			}

			// Display file change notification
			fileName := filepath.Base(filePath)
			w.printStatus(fmt.Sprintf("File changed: %s", fileName))

			// Run appropriate tests based on watch mode
			switch w.options.Mode {
			case WatchAll:
				w.runAllTests()

			case WatchChanged:
				w.runTestsForFile(filePath)

			case WatchRelated:
				w.runRelatedTests(filePath)
			}

			// Display watch mode info
			w.printWatchInfo()
		}
	}
}

// Stop stops the watcher
func (w *TestWatcher) Stop() error {
	return w.fileWatcher.Close()
}

// debounceEvents collects file events and debounces them to avoid multiple test runs
func (w *TestWatcher) debounceEvents(fileEvents <-chan FileEvent, debouncedEvents chan<- string) {
	// Map to track most recent change for each file
	recentChanges := make(map[string]time.Time)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case event := <-fileEvents:
			// Record this file change
			recentChanges[event.Path] = event.Timestamp

		case <-ticker.C:
			now := time.Now()
			for file, timestamp := range recentChanges {
				// If enough time has passed, send this event and remove from tracking
				if now.Sub(timestamp) >= w.options.DebounceInterval {
					debouncedEvents <- file
					delete(recentChanges, file)
				}
			}
		}
	}
}

// runAllTests runs all tests
func (w *TestWatcher) runAllTests() {
	w.printStatus("Running all tests...")

	// Reset the processor
	w.processor.Reset()

	// Set up a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Run all tests with JSON output
	testOutput, err := w.testRunner.Run(ctx, []string{"./..."})
	if err != nil {
		fmt.Fprintf(w.options.Writer, "Error running tests: %v\n", err)
		return
	}

	// Process the results
	w.processTestResults(testOutput)
}

// runTestsForFile runs tests for a specific file
func (w *TestWatcher) runTestsForFile(filePath string) {
	if !strings.HasSuffix(filePath, ".go") {
		// Skip non-Go files
		return
	}

	w.printStatus(fmt.Sprintf("Running tests for %s", filepath.Base(filePath)))

	// Reset the processor
	w.processor.Reset()

	// If it's a test file, run it directly
	var testPaths []string
	if strings.HasSuffix(filePath, "_test.go") {
		testPaths = []string{filePath}
	} else {
		// Try to find the corresponding test file
		testFile, err := w.testFinder.FindTestFile(filePath)
		if err != nil {
			w.printStatus(fmt.Sprintf("No test file found for %s", filepath.Base(filePath)))
			return
		}
		testPaths = []string{testFile}
	}

	// Set up a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Run the tests
	testOutput, err := w.testRunner.Run(ctx, testPaths)
	if err != nil {
		fmt.Fprintf(w.options.Writer, "Error running tests: %v\n", err)
		return
	}

	// Process the results
	w.processTestResults(testOutput)
}

// runRelatedTests runs tests for a file and its related files
func (w *TestWatcher) runRelatedTests(filePath string) {
	if !strings.HasSuffix(filePath, ".go") {
		// Skip non-Go files
		return
	}

	w.printStatus(fmt.Sprintf("Running related tests for %s", filepath.Base(filePath)))

	// Reset the processor
	w.processor.Reset()

	// If it's a test file, run the package tests
	var testPaths []string
	if strings.HasSuffix(filePath, "_test.go") {
		// Find all tests in the same package
		packageTests, err := w.testFinder.FindPackageTests(filePath)
		if err != nil {
			// Fall back to just running this test file
			testPaths = []string{filePath}
		} else {
			testPaths = packageTests
		}
	} else {
		// Try to find all test files in the package
		packageTests, err := w.testFinder.FindPackageTests(filePath)
		if err != nil {
			// Try to find just the corresponding test file
			testFile, err := w.testFinder.FindTestFile(filePath)
			if err != nil {
				w.printStatus(fmt.Sprintf("No test files found for %s", filepath.Base(filePath)))
				return
			}
			testPaths = []string{testFile}
		} else {
			testPaths = packageTests
		}
	}

	// Set up a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	// Run the tests
	testOutput, err := w.testRunner.Run(ctx, testPaths)
	if err != nil {
		fmt.Fprintf(w.options.Writer, "Error running tests: %v\n", err)
		return
	}

	// Process the results
	w.processTestResults(testOutput)
}

// processTestResults processes the test output and displays results
func (w *TestWatcher) processTestResults(output string) {
	err := w.processor.ProcessJSONOutput(output)
	if err != nil {
		fmt.Fprintf(w.options.Writer, "Error processing test output: %v\n", err)
		return
	}

	// Render the results
	err = w.processor.RenderResults(true)
	if err != nil {
		fmt.Fprintf(w.options.Writer, "Error rendering results: %v\n", err)
	}
}

// printStatus prints a status message
func (w *TestWatcher) printStatus(message string) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Fprintf(w.options.Writer, "%s %s\n",
		w.formatter.Dim(timestamp),
		message)
}

// printWatchInfo displays watch mode information
func (w *TestWatcher) printWatchInfo() {
	fmt.Fprintln(w.options.Writer)
	fmt.Fprintf(w.options.Writer, "Watch mode: %s\n", w.options.Mode)
	fmt.Fprintln(w.options.Writer)
	fmt.Fprintln(w.options.Writer, "Press 'a' to run all tests")
	fmt.Fprintln(w.options.Writer, "Press 'c' to run only changed tests")
	fmt.Fprintln(w.options.Writer, "Press 'r' to run related tests")
	fmt.Fprintln(w.options.Writer, "Press 'q' to quit")
}

// clearTerminal clears the terminal screen
func (w *TestWatcher) clearTerminal() {
	// ANSI escape sequence to clear the screen and move cursor to top-left
	fmt.Fprint(w.options.Writer, "\033[2J\033[H")
}
