package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// AppController orchestrates the main application flow
type AppController struct {
	argParser    ArgParser
	configLoader ConfigLoader
	testRunner   *TestRunner
	processor    *TestProcessor
	watcher      *FileWatcher
}

// NewAppController creates a new application controller
func NewAppController() *AppController {
	return &AppController{
		argParser:    &DefaultArgParser{},
		configLoader: &DefaultConfigLoader{},
		testRunner:   &TestRunner{JSONOutput: true}, // Enable JSON output for processing
	}
}

// Run executes the main application flow
func (a *AppController) Run(args []string) error {
	// Step 1: Parse CLI arguments
	cliArgs, err := a.argParser.Parse(args)
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Step 2: Load configuration
	config, err := a.loadConfiguration()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Step 3: Merge CLI arguments with configuration
	mergedConfig := config.MergeWithCLIArgs(cliArgs)

	// Step 4: Validate final configuration
	if err := ValidateConfig(mergedConfig); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Step 5: Initialize test processor
	a.processor = NewTestProcessor(
		os.Stdout,
		NewColorFormatter(mergedConfig.Colors),
		NewIconProvider(mergedConfig.Visual.Icons != "none"),
		80, // Terminal width, could be detected
	)

	// Step 6: Execute tests based on configuration
	if mergedConfig.Watch.Enabled {
		return a.runWatchMode(mergedConfig, cliArgs)
	} else {
		return a.runSingleMode(mergedConfig, cliArgs)
	}
}

// loadConfiguration loads configuration from file or returns defaults
func (a *AppController) loadConfiguration() (*Config, error) {
	// Check for sentinel.config.json in current directory
	configPath := "sentinel.config.json"
	if _, err := os.Stat(configPath); err == nil {
		return a.configLoader.LoadFromFile(configPath)
	}

	// Return default configuration if no file found
	return GetDefaultConfig(), nil
}

// runSingleMode executes tests once and exits
func (a *AppController) runSingleMode(config *Config, cliArgs *Args) error {
	fmt.Printf("🚀 Running tests with go-sentinel...\n\n")

	// Start timing
	startTime := time.Now()

	// Determine packages to test
	packages := cliArgs.Packages
	if len(packages) == 0 {
		// Default to current directory
		packages = []string{"./..."}
	}

	// Execute tests for each package
	for _, pkg := range packages {
		if err := a.runPackageTests(pkg, config); err != nil {
			return fmt.Errorf("failed to run tests for package %s: %w", pkg, err)
		}
	}

	// Add separator before final results
	fmt.Fprintln(a.processor.writer)

	// Render final summary (no need to call finalize again as ProcessStream does it)
	if err := a.processor.RenderResults(true); err != nil {
		return fmt.Errorf("failed to render results: %w", err)
	}

	// Calculate and display timing using our actual duration
	stats := a.processor.GetStats()
	actualDuration := time.Since(startTime)

	fmt.Printf("\n⏱️  Tests completed in %v\n", actualDuration)

	// Exit with appropriate code
	if stats.FailedTests > 0 {
		os.Exit(1)
	}

	return nil
}

// runWatchMode executes tests in watch mode
func (a *AppController) runWatchMode(config *Config, cliArgs *Args) error {
	fmt.Printf("👀 Starting watch mode...\n")
	fmt.Printf("📁 Watching for changes in: %v\n", config.Paths.IncludePatterns)
	fmt.Printf("🚫 Ignoring: %v\n", config.Paths.ExcludePatterns)
	fmt.Printf("⌨️  Press Ctrl+C to exit\n\n")

	// Initialize file watcher
	watcher, err := NewFileWatcher(
		config.Paths.IncludePatterns,
		config.Watch.IgnorePatterns,
	)
	if err != nil {
		return fmt.Errorf("failed to create file watcher: %w", err)
	}
	defer watcher.Close()

	a.watcher = watcher

	// Create a channel for file events
	events := make(chan FileEvent, 10)

	// Start the watcher in a goroutine
	go func() {
		if err := watcher.Watch(events); err != nil {
			fmt.Printf("🚨 Watcher error: %v\n", err)
		}
	}()

	// Run tests initially if configured
	if config.Watch.RunOnStart {
		if err := a.runSingleMode(config, cliArgs); err != nil {
			fmt.Printf("❌ Initial test run failed: %v\n", err)
		}
	}

	// Watch for file changes
	for event := range events {
		if err := a.handleFileChange(event, config); err != nil {
			fmt.Printf("❌ Error handling file change: %v\n", err)
		}
	}

	return nil
}

// handleFileChange processes a file change event and runs relevant tests
func (a *AppController) handleFileChange(event FileEvent, config *Config) error {
	fmt.Printf("📝 File changed: %s\n", event.Path)

	// Wait for debounce period to handle rapid successive changes
	time.Sleep(config.Watch.Debounce)

	// Clear terminal if configured
	if config.Watch.ClearOnRerun {
		clearTerminal()
	}

	// Determine which tests to run based on the changed file
	testsToRun := a.determineTestsToRun(event.Path)

	if len(testsToRun) == 0 {
		fmt.Printf("🔍 No tests found for changed file\n")
		return nil
	}

	fmt.Printf("🏃 Running tests: %v\n\n", testsToRun)

	// Reset processor for new test run
	a.processor = NewTestProcessor(
		os.Stdout,
		NewColorFormatter(config.Colors),
		NewIconProvider(config.Visual.Icons != "none"),
		80,
	)

	// Run the determined tests
	for _, test := range testsToRun {
		if err := a.runPackageTests(test, config); err != nil {
			fmt.Printf("❌ Failed to run test %s: %v\n", test, err)
		}
	}

	// Finalize and render results
	a.processor.finalize()
	return a.processor.RenderResults(true)
}

// runPackageTests executes tests for a specific package
func (a *AppController) runPackageTests(pkg string, config *Config) error {
	// Prepare test paths
	testPaths := []string{pkg}

	// Configure the test runner
	a.testRunner.Verbose = config.Verbosity > 0
	a.testRunner.JSONOutput = true

	// Execute test command using context with timeout
	ctx := context.Background()
	if config.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, config.Timeout)
		defer cancel()
	}

	// Use streaming approach for real-time output
	stream, err := a.testRunner.RunStream(ctx, testPaths)
	if err != nil {
		return fmt.Errorf("failed to start test stream: %w", err)
	}
	defer stream.Close()

	// Create progress channel
	progress := make(chan TestProgress, 10)
	defer close(progress)

	// Start progress monitoring in background (optional)
	go func() {
		for p := range progress {
			// Could add progress bar or other indicators here
			_ = p // Currently just consume the progress updates
		}
	}()

	// Process the stream in real-time
	return a.processor.ProcessStream(stream, progress)
}

// determineTestsToRun determines which tests should run based on a changed file
func (a *AppController) determineTestsToRun(changedFile string) []string {
	// Simple heuristic: if it's a test file, run it directly
	if isTestFile(changedFile) {
		// Get the package directory
		dir := filepath.Dir(changedFile)
		return []string{dir}
	}

	// If it's a source file, find related test files
	testFile := getCorrespondingTestFile(changedFile)
	if testFile != "" {
		if _, err := os.Stat(testFile); err == nil {
			dir := filepath.Dir(testFile)
			return []string{dir}
		}
	}

	// Fallback: run tests in the same directory
	dir := filepath.Dir(changedFile)
	return []string{dir}
}

// Helper functions
func isTestFile(filename string) bool {
	return strings.HasSuffix(filename, "_test.go")
}

func getCorrespondingTestFile(filename string) string {
	if filepath.Ext(filename) != ".go" {
		return ""
	}

	base := filename[:len(filename)-3] // Remove .go
	return base + "_test.go"
}

func clearTerminal() {
	fmt.Print("\033[2J\033[H") // ANSI escape codes to clear screen and move cursor to top
}
