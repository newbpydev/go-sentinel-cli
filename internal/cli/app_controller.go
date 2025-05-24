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
	optimizedRunner     *OptimizedTestRunner
	processor           *TestProcessor
	watcher             *FileWatcher
	cache               *TestResultCache
	incrementalRenderer *IncrementalRenderer
	optimizedMode       *OptimizedWatchMode
}

// NewAppController creates a new application controller
func NewAppController() *AppController {
	cache := NewTestResultCache()
	return &AppController{
		argParser:       &DefaultArgParser{},
		configLoader:    &DefaultConfigLoader{},
		testRunner:      &TestRunner{JSONOutput: true}, // Enable JSON output for processing
		optimizedRunner: NewOptimizedTestRunner(),
		cache:           cache,
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

	// Step 6: Configure optimization if enabled
	a.configureOptimization(cliArgs)

	// Step 7: Execute tests based on configuration
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

// configureOptimization sets up optimization based on CLI arguments
func (a *AppController) configureOptimization(cliArgs *Args) {
	// Check if optimization is enabled via CLI flags or environment variable
	optimizationEnabled := cliArgs.Optimized || os.Getenv("GO_SENTINEL_OPTIMIZED") == "true"

	if optimizationEnabled {
		// Initialize optimized watch mode
		a.optimizedMode = NewOptimizedWatchMode()
		a.optimizedMode.EnableOptimization()

		// Set optimization mode if specified
		optimizationMode := cliArgs.OptimizationMode
		if optimizationMode == "" {
			// Check environment variable
			if envMode := os.Getenv("GO_SENTINEL_OPTIMIZATION_MODE"); envMode != "" {
				optimizationMode = envMode
			} else {
				// Default to balanced mode
				optimizationMode = "balanced"
			}
		}

		a.optimizedMode.SetOptimizationMode(optimizationMode)
		fmt.Printf("🚀 Optimized mode enabled (%s) - leveraging Go's built-in caching!\n", optimizationMode)
	}
}

// runSingleMode executes tests once and exits
func (a *AppController) runSingleMode(config *Config, cliArgs *Args) error {
	fmt.Printf("🚀 Running tests with go-sentinel...\n")

	// Show optimization status
	if a.optimizedMode != nil && a.optimizedMode.IsEnabled() {
		fmt.Printf("⚡ Optimization enabled - leveraging Go's built-in caching\n")
	}
	fmt.Printf("\n")

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
	fmt.Fprintln(a.processor.GetWriter())

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

	// Initialize and start file watcher
	watcher, events, watcherErrors, err := a.initializeWatcher(config)
	if err != nil {
		return err
	}
	defer watcher.Close()

	// Run initial tests if configured
	if err := a.runInitialTests(config, cliArgs); err != nil {
		fmt.Printf("❌ Initial test run failed: %v\n", err)
	}

	// Display watch mode help
	a.displayWatchModeHelp()

	// Start watch loop
	return a.runWatchLoop(config, events, watcherErrors)
}

// initializeWatcher creates and starts the file watcher
func (a *AppController) initializeWatcher(config *Config) (*FileWatcher, chan FileEvent, chan error, error) {
	// Initialize file watcher
	watcher, err := NewFileWatcher(
		config.Paths.IncludePatterns,
		config.Watch.IgnorePatterns,
	)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	a.watcher = watcher

	// Create channels for file events and errors
	events := make(chan FileEvent, 50)
	watcherErrors := make(chan error, 1)

	// Start the watcher in a goroutine
	go func() {
		defer close(events)
		if err := watcher.Watch(events); err != nil {
			watcherErrors <- fmt.Errorf("watcher error: %w", err)
		}
	}()

	return watcher, events, watcherErrors, nil
}

// runInitialTests runs tests initially if configured
func (a *AppController) runInitialTests(config *Config, cliArgs *Args) error {
	if config.Watch.RunOnStart {
		fmt.Printf("🏃 Running initial tests...\n\n")
		return a.runSingleMode(config, cliArgs)
	}
	return nil
}

// runWatchLoop runs the main watch loop handling file events
func (a *AppController) runWatchLoop(config *Config, events chan FileEvent, watcherErrors chan error) error {
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
	changes := a.analyzeFileChanges(events)
	if len(changes) == 0 {
		return nil
	}

	// Clear terminal if configured
	if config.Watch.ClearOnRerun {
		clearTerminal()
	}

	// Use optimized mode if enabled, otherwise fall back to standard processing
	if a.optimizedMode != nil && a.optimizedMode.IsEnabled() {
		return a.optimizedMode.HandleFileChanges(events, config)
	}

	// Execute tests and handle results
	return a.executeOptimizedTests(changes, config)
}

// analyzeFileChanges analyzes file events and returns meaningful changes
func (a *AppController) analyzeFileChanges(events []FileEvent) []*FileChange {
	changes := make([]*FileChange, 0, len(events))
	for _, event := range events {
		change, err := a.cache.AnalyzeChange(event.Path)
		if err != nil {
			fmt.Printf("⚠️ Failed to analyze change for %s: %v\n", event.Path, err)
			continue
		}
		changes = append(changes, change)
	}
	return changes
}

// executeOptimizedTests runs tests using the optimized runner and handles results
func (a *AppController) executeOptimizedTests(changes []*FileChange, config *Config) error {
	optimizedResult, err := a.optimizedRunner.RunOptimized(context.Background(), changes)
	if err != nil {
		return fmt.Errorf("optimized test execution failed: %w", err)
	}

	// Handle case where no tests need to run
	if optimizedResult.TestsRun == 0 {
		return a.handleNoTestsNeeded(optimizedResult, changes, config)
	}

	// Handle case where tests were executed
	return a.handleTestsExecuted(optimizedResult, changes, config)
}

// handleNoTestsNeeded handles the case where changes were detected but no tests need to run
func (a *AppController) handleNoTestsNeeded(result *OptimizedTestResult, changes []*FileChange, config *Config) error {
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

	stats := result.GetEfficiencyStats()
	fmt.Printf("💨 %s (%.1f%% cache hit rate)\n",
		result.Message,
		stats["cache_hit_rate"].(float64))
	fmt.Printf("👀 Watching for file changes...\n")
	return nil
}

// handleTestsExecuted handles the case where tests were actually executed
func (a *AppController) handleTestsExecuted(result *OptimizedTestResult, changes []*FileChange, config *Config) error {
	stats := result.GetEfficiencyStats()
	fmt.Printf("⚡ Running %d tests, %d from cache (%.1f%% efficiency)\n",
		result.TestsRun,
		result.CacheHits,
		stats["cache_hit_rate"].(float64))

	// Process test output if available
	if err := a.processTestOutput(result.Output, config); err != nil {
		return err
	}

	// Render incremental results
	if err := a.renderIncrementalResults(changes, config); err != nil {
		return err
	}

	// Display performance metrics
	a.displayPerformanceMetrics(result, stats)
	return nil
}

// processTestOutput processes test output through the processor for consistent rendering
func (a *AppController) processTestOutput(output string, config *Config) error {
	if output == "" {
		return nil
	}

	// Reset processor for new test run
	a.processor = NewTestProcessor(
		os.Stdout,
		NewColorFormatter(config.Colors),
		NewIconProvider(config.Visual.Icons != "none"),
		80,
	)

	// Process the test output through our processor for consistent rendering
	reader := strings.NewReader(output)
	progress := make(chan TestProgress, 10)
	defer close(progress)

	// Start progress monitoring in background
	go func() {
		for range progress {
			// Consume progress updates
		}
	}()

	// Process the output
	if err := a.processor.ProcessStream(reader, progress); err != nil {
		fmt.Printf("⚠️ Warning: failed to process test output: %v\n", err)
	}
	return nil
}

// renderIncrementalResults renders incremental results for watch mode
func (a *AppController) renderIncrementalResults(changes []*FileChange, config *Config) error {
	// Initialize incremental renderer if needed
	if a.incrementalRenderer == nil {
		a.initializeIncrementalRenderer(config)
	}

	// Use incremental rendering for watch mode
	return a.incrementalRenderer.RenderIncrementalResults(
		a.processor.GetSuites(),
		a.processor.GetStats(),
		changes,
	)
}

// displayPerformanceMetrics displays performance and efficiency information
func (a *AppController) displayPerformanceMetrics(result *OptimizedTestResult, stats map[string]interface{}) {
	cacheStats := a.cache.GetStats()
	fmt.Printf("⏱️  Completed in %v | Efficiency: %.1f%% | Cache: %d results\n",
		result.Duration,
		stats["cache_hit_rate"].(float64),
		cacheStats["cached_results"])
	fmt.Printf("👀 Watching for file changes...\n")
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
