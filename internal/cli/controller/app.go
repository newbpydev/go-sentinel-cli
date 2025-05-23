package controller

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/newbpydev/go-sentinel/internal/cli/core"
	"github.com/newbpydev/go-sentinel/internal/cli/execution"
	"github.com/newbpydev/go-sentinel/internal/cli/rendering"
)

// AppController coordinates the overall application flow using the new modular architecture
type AppController struct {
	testRunner core.TestRunner
	cache      core.CacheManager
	factory    *execution.StrategyFactory
	renderer   *rendering.StructuredRenderer
}

// NewAppController creates a new application controller with default dependencies
func NewAppController() *AppController {
	// Initialize cache with reasonable defaults
	cache := execution.NewInMemoryCacheManager(1000)

	// Initialize strategy factory
	factory := execution.NewStrategyFactory()

	// Create default strategy (aggressive for development)
	strategy := factory.CreateStrategy("aggressive")

	// Initialize test runner
	testRunner := execution.NewSmartTestRunner(cache, strategy)

	// Initialize renderer with default settings
	renderer := rendering.NewStructuredRenderer(os.Stdout, true, false)

	return &AppController{
		testRunner: testRunner,
		cache:      cache,
		factory:    factory,
		renderer:   renderer,
	}
}

// Run executes the application with the given configuration
func (a *AppController) Run(ctx context.Context, config *core.Config) error {
	// Update renderer settings based on config
	a.updateRenderer(config)

	// Render startup message with original style
	a.renderer.RenderStartup(config.UseCache, config.CacheStrategy)

	if config.WatchMode {
		return a.RunWatch(ctx, config)
	}
	return a.RunOnce(ctx, config)
}

// RunOnce executes tests once without watching
func (a *AppController) RunOnce(ctx context.Context, config *core.Config) error {
	startTime := time.Now()

	// Get execution strategy based on configuration
	strategy := a.factory.CreateStrategy(config.CacheStrategy)

	// Determine what files to test
	changes := a.determineInitialChanges(config)

	// Execute tests
	result, err := a.testRunner.RunTests(ctx, changes, strategy)
	if err != nil {
		a.renderer.RenderError(err)
		return fmt.Errorf("failed to run tests: %w", err)
	}

	// Render results using original style
	a.renderer.RenderTestResult(result)

	// Render cache statistics if verbose
	if config.Verbose {
		stats := a.cache.GetStats()
		a.renderer.RenderCacheStats(stats)
	}

	// Render completion timing
	a.renderer.RenderCompletion(time.Since(startTime))

	return nil
}

// RunWatch starts watch mode
func (a *AppController) RunWatch(ctx context.Context, config *core.Config) error {
	// Render watch mode startup with original style
	a.renderer.RenderWatchStart()

	// For now, implement a basic watch mode that runs once
	// In the full implementation, this would set up file watchers
	err := a.RunOnce(ctx, config)
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Initial test run complete.")
	a.renderer.RenderWatchModeInfo()

	// For now, just return immediately instead of blocking
	// In a real implementation, this would set up file watchers and block
	fmt.Println("⚠️  Watch mode is not fully implemented yet. Exiting...")

	return nil
}

// updateRenderer updates renderer settings based on configuration
func (a *AppController) updateRenderer(config *core.Config) {
	// Create new renderer with updated settings
	useColors := !config.NoColor
	a.renderer = rendering.NewStructuredRenderer(os.Stdout, useColors, config.Verbose)
}

// determineInitialChanges determines what files to test based on configuration
func (a *AppController) determineInitialChanges(config *core.Config) []core.FileChange {
	var changes []core.FileChange

	// If specific test patterns are provided, treat them as test changes
	if config.TestPattern != "" {
		// Find matching test files
		matches := a.findTestFiles(config.TestPattern)
		for _, match := range matches {
			changes = append(changes, core.FileChange{
				Path:      match,
				Type:      core.ChangeTypeTest,
				IsNew:     true,
				Timestamp: time.Now(),
			})
		}
	} else {
		// Default: treat current directory as a source change to run all tests
		changes = append(changes, core.FileChange{
			Path:      ".",
			Type:      core.ChangeTypeSource,
			IsNew:     false,
			Timestamp: time.Now(),
		})
	}

	return changes
}

// findTestFiles finds test files matching the pattern
func (a *AppController) findTestFiles(pattern string) []string {
	var matches []string

	// Walk the current directory to find test files
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor and other common directories
		if info.IsDir() {
			name := info.Name()
			if name == "vendor" || name == ".git" || name == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if it's a test file matching the pattern
		if strings.HasSuffix(path, "_test.go") {
			if pattern == "" || strings.Contains(path, pattern) {
				matches = append(matches, path)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Warning: Error finding test files: %v\n", err)
	}

	return matches
}

// Legacy compatibility functions for the old CLI interface

// RunLegacy provides compatibility with the old string slice argument format
func (a *AppController) RunLegacy(args []string) error {
	// Parse legacy arguments into configuration
	config := a.parseLegacyArgs(args)

	// Create context
	ctx := context.Background()

	// Run with new interface
	return a.Run(ctx, config)
}

// parseLegacyArgs converts old string slice format to new Config
func (a *AppController) parseLegacyArgs(args []string) *core.Config {
	config := &core.Config{
		UseCache:      true,
		CacheStrategy: "aggressive",
		ShowProgress:  true,
		ShowSummary:   true,
	}

	// Parse flags from args
	for i, arg := range args {
		switch {
		case arg == "--watch" || arg == "-w":
			config.WatchMode = true
		case arg == "--verbose" || arg == "-v":
			config.Verbose = true
		case arg == "--no-color":
			config.NoColor = true
		case arg == "--optimized" || arg == "-o":
			config.UseCache = true
		case strings.HasPrefix(arg, "--optimization="):
			config.CacheStrategy = strings.TrimPrefix(arg, "--optimization=")
		case strings.HasPrefix(arg, "--test="):
			config.TestPattern = strings.TrimPrefix(arg, "--test=")
		case arg == "--fail-fast":
			config.FailFast = true
		case !strings.HasPrefix(arg, "-"):
			// This is a package argument
			// For simplicity, we'll just store the first one as test pattern
			if config.TestPattern == "" && !strings.Contains(arg, "/") {
				config.TestPattern = arg
			}
		}

		// Handle arguments that come after flags
		if i > 0 && args[i-1] == "--test" {
			config.TestPattern = arg
		}
	}

	return config
}
