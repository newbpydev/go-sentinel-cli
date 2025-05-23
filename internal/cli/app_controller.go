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
	argParser           ArgParser
	configLoader        ConfigLoader
	testRunner          *TestRunner
	processor           *TestProcessor
	watcher             *FileWatcher
	cache               *TestResultCache
	incrementalRenderer *IncrementalRenderer
}

// NewAppController creates a new application controller
func NewAppController() *AppController {
	cache := NewTestResultCache()
	return &AppController{
		argParser:    &DefaultArgParser{},
		configLoader: &DefaultConfigLoader{},
		testRunner:   &TestRunner{JSONOutput: true}, // Enable JSON output for processing
		cache:        cache,
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

	// Step 5.5: Ensure watch paths are properly set from CLI args
	if mergedConfig.Watch.Enabled && len(mergedConfig.Paths.IncludePatterns) == 0 {
		// Convert CLI packages to watch paths if not already set
		watchPaths := convertPackagesToWatchPaths(cliArgs.Packages)
		mergedConfig.Paths.IncludePatterns = watchPaths
	}

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

	// Display detailed watch configuration
	a.displayWatchConfiguration(config)

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

	// Create a buffered channel for file events to handle bursts
	events := make(chan FileEvent, 50)

	// Create error channel for watcher errors
	watcherErrors := make(chan error, 1)

	// Start the watcher in a goroutine
	go func() {
		defer close(events)
		if err := watcher.Watch(events); err != nil {
			watcherErrors <- fmt.Errorf("watcher error: %w", err)
		}
	}()

	// Run tests initially if configured
	if config.Watch.RunOnStart {
		fmt.Printf("🏃 Running initial tests...\n\n")
		if err := a.runSingleMode(config, cliArgs); err != nil {
			fmt.Printf("❌ Initial test run failed: %v\n", err)
		}
	}

	// Display watch mode help
	a.displayWatchModeHelp()

	// Watch for file changes with debouncing
	debouncer := NewFileEventDebouncer(config.Watch.Debounce)
	defer debouncer.Stop()

	for {
		select {
		case event, ok := <-events:
			if !ok {
				fmt.Printf("\n👋 Watch mode stopped\n")
				return nil
			}

			// Send to debouncer
			debouncer.AddEvent(event)

		case debouncedEvents := <-debouncer.Events():
			if err := a.handleDebouncedFileChanges(debouncedEvents, config); err != nil {
				fmt.Printf("❌ Error handling file changes: %v\n", err)
			}

		case err := <-watcherErrors:
			fmt.Printf("🚨 Watcher error: %v\n", err)
			return err
		}
	}
}

// displayWatchConfiguration shows what is being watched and ignored
func (a *AppController) displayWatchConfiguration(config *Config) {
	fmt.Printf("📁 Watching: ")
	if len(config.Paths.IncludePatterns) == 0 {
		fmt.Printf("%s\n", "current directory")
	} else {
		for i, path := range config.Paths.IncludePatterns {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", path)
		}
		fmt.Printf("\n")
	}

	if len(config.Watch.IgnorePatterns) > 0 {
		fmt.Printf("🚫 Ignoring: ")
		for i, pattern := range config.Watch.IgnorePatterns {
			if i > 0 {
				fmt.Printf(", ")
			}
			fmt.Printf("%s", pattern)
		}
		fmt.Printf("\n")
	}

	fmt.Printf("⏱️  Debounce: %v\n", config.Watch.Debounce)
	fmt.Printf("⌨️  Press Ctrl+C to exit\n\n")
}

// displayWatchModeHelp shows available commands in watch mode
func (a *AppController) displayWatchModeHelp() {
	fmt.Printf("📋 Watch mode active. File changes will trigger test runs.\n")
	fmt.Printf("   • .go files: Runs tests for the package\n")
	fmt.Printf("   • *_test.go files: Runs specific test file\n\n")
}

// handleDebouncedFileChanges processes multiple debounced file events efficiently
func (a *AppController) handleDebouncedFileChanges(events []FileEvent, config *Config) error {
	if len(events) == 0 {
		return nil
	}

	// Analyze file changes for impact
	changes := make([]*FileChange, 0, len(events))
	for _, event := range events {
		change, err := a.cache.AnalyzeChange(event.Path)
		if err != nil {
			fmt.Printf("⚠️ Failed to analyze change for %s: %v\n", event.Path, err)
			continue
		}
		changes = append(changes, change)
	}

	if len(changes) == 0 {
		return nil
	}

	// Clear terminal if configured
	if config.Watch.ClearOnRerun {
		clearTerminal()
	}

	// Get stale tests based on change analysis
	staleTests := a.cache.GetStaleTests(changes)

	if len(staleTests) == 0 {
		// Initialize incremental renderer if needed
		if a.incrementalRenderer == nil {
			a.initializeIncrementalRenderer(config)
		}

		// Show that changes were detected but no tests need to run
		if err := a.incrementalRenderer.RenderIncrementalResults(
			make(map[string]*TestSuite),
			&TestRunStats{},
			changes,
		); err != nil {
			return err
		}

		fmt.Printf("👀 Watching for file changes...\n")
		return nil // Don't run any tests!
	}

	fmt.Printf("⚡ Running tests for %d changed file(s) affecting %d test package(s)\n\n",
		len(changes), len(staleTests))

	// Reset processor for new test run
	a.processor = NewTestProcessor(
		os.Stdout,
		NewColorFormatter(config.Colors),
		NewIconProvider(config.Visual.Icons != "none"),
		80,
	)

	// Run tests efficiently - use parallel execution for multiple tests
	if len(staleTests) > 1 && config.Parallel > 1 {
		// Use parallel execution for multiple tests
		parallelRunner := NewParallelTestRunner(config.Parallel, a.testRunner, a.cache)

		results, err := parallelRunner.RunParallel(context.Background(), staleTests, config)
		if err != nil {
			return fmt.Errorf("parallel test execution failed: %w", err)
		}

		// Merge results into processor
		MergeResults(a.processor, results)

		// Cache all results from parallel execution
		for _, result := range results {
			if result.Error == nil && result.Suite != nil {
				a.cache.CacheResult(result.TestPath, result.Suite)
			}
		}

		// Report parallel execution metrics
		var fromCache, fromExecution int
		for _, result := range results {
			if result.FromCache {
				fromCache++
			} else {
				fromExecution++
			}
		}

		if fromCache > 0 {
			fmt.Printf("📋 Parallel execution: %d ran, %d from cache\n", fromExecution, fromCache)
		}
	} else {
		// Sequential execution for single test or when parallel is disabled
		for _, testPath := range staleTests {
			if err := a.runPackageTests(testPath, config); err != nil {
				fmt.Printf("❌ Failed to run test %s: %v\n", testPath, err)
				continue
			}

			// Cache the results for this test - fix the suite key lookup
			for suitePath, suite := range a.processor.suites {
				// Use actual suite path instead of testPath
				a.cache.CacheResult(suitePath, suite)
			}
		}
	}

	// Initialize incremental renderer if needed
	if a.incrementalRenderer == nil {
		a.initializeIncrementalRenderer(config)
	}

	// Use incremental rendering for watch mode
	if err := a.incrementalRenderer.RenderIncrementalResults(
		a.processor.suites,
		a.processor.GetStats(),
		changes,
	); err != nil {
		return err
	}

	// Display performance metrics
	stats := a.processor.GetStats()
	cacheStats := a.cache.GetStats()
	fmt.Printf("⏱️  Completed in %v | Cache: %d results, %d files tracked\n",
		stats.Duration,
		cacheStats["cached_results"],
		cacheStats["tracked_files"])

	fmt.Printf("👀 Watching for file changes...\n")

	return nil
}

// initializeIncrementalRenderer creates and configures the incremental renderer
func (a *AppController) initializeIncrementalRenderer(config *Config) {
	a.incrementalRenderer = NewIncrementalRenderer(
		os.Stdout,
		NewColorFormatter(config.Colors),
		NewIconProvider(config.Visual.Icons != "none"),
		80,
		a.cache,
	)
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
