// Package app provides the main application controller implementation
package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/newbpydev/go-sentinel/internal/watch/core"
)

// Controller implements the ApplicationController interface
type Controller struct {
	// Core dependencies
	argParser    ArgumentParser
	configLoader ConfigurationLoader
	lifecycle    LifecycleManager
	container    DependencyContainer
	eventHandler ApplicationEventHandler

	// Component interfaces (to be injected)
	testExecutor     TestExecutor
	watchCoordinator core.WatchCoordinator
	displayRenderer  DisplayRenderer

	// Internal state
	config    *Configuration
	args      *Arguments
	ctx       context.Context
	cancel    context.CancelFunc
	isRunning bool
}

// TestExecutor interface for test execution (will be defined in test package)
type TestExecutor interface {
	ExecuteSingle(ctx context.Context, packages []string, config *Configuration) error
	ExecuteWatch(ctx context.Context, config *Configuration) error
}

// DisplayRenderer interface for result display (will be defined in ui package)
type DisplayRenderer interface {
	RenderResults(ctx context.Context) error
	SetConfiguration(config *Configuration) error
}

// NewController creates a new application controller
func NewController(
	argParser ArgumentParser,
	configLoader ConfigurationLoader,
	lifecycle LifecycleManager,
	container DependencyContainer,
	eventHandler ApplicationEventHandler,
) ApplicationController {
	return &Controller{
		argParser:    argParser,
		configLoader: configLoader,
		lifecycle:    lifecycle,
		container:    container,
		eventHandler: eventHandler,
	}
}

// Initialize implements the ApplicationController interface
func (c *Controller) Initialize() error {
	// Create application context
	c.ctx, c.cancel = context.WithCancel(context.Background())

	// Initialize dependency container
	if err := c.container.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize dependency container: %w", err)
	}

	// Resolve required dependencies
	if err := c.resolveDependencies(); err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Register shutdown hooks
	c.lifecycle.RegisterShutdownHook(func() error {
		return c.cleanup()
	})

	return nil
}

// Run implements the ApplicationController interface
func (c *Controller) Run(args []string) error {
	// Mark as running
	c.isRunning = true
	defer func() { c.isRunning = false }()

	// Notify startup
	if err := c.eventHandler.OnStartup(c.ctx); err != nil {
		return fmt.Errorf("startup event handler failed: %w", err)
	}

	// Step 1: Parse CLI arguments
	parsedArgs, err := c.argParser.Parse(args)
	if err != nil {
		return fmt.Errorf("failed to parse arguments: %w", err)
	}
	c.args = parsedArgs

	// Step 2: Load and merge configuration
	config, err := c.loadAndMergeConfiguration(parsedArgs)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	c.config = config

	// Step 3: Validate configuration
	if err := c.configLoader.Validate(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Step 4: Configure components
	if err := c.configureComponents(config); err != nil {
		return fmt.Errorf("failed to configure components: %w", err)
	}

	// Step 5: Execute based on configuration
	return c.executeApplication(config, parsedArgs)
}

// Shutdown implements the ApplicationController interface
func (c *Controller) Shutdown(ctx context.Context) error {
	// Cancel application context
	if c.cancel != nil {
		c.cancel()
	}

	// Notify shutdown
	if err := c.eventHandler.OnShutdown(ctx); err != nil {
		fmt.Printf("Warning: shutdown event handler failed: %v\n", err)
	}

	// Shutdown lifecycle manager
	return c.lifecycle.Shutdown(ctx)
}

// loadAndMergeConfiguration loads configuration from file or defaults and merges with CLI args
func (c *Controller) loadAndMergeConfiguration(args *Arguments) (*Configuration, error) {
	// Load configuration
	var config *Configuration
	var err error

	// Check for sentinel.config.json in current directory
	configPath := "sentinel.config.json"
	if _, statErr := os.Stat(configPath); statErr == nil {
		config, err = c.configLoader.LoadFromFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load configuration from file: %w", err)
		}
	} else {
		// Use defaults if no file found
		config = c.configLoader.LoadFromDefaults()
	}

	// Merge with CLI arguments
	mergedConfig := c.configLoader.Merge(config, args)

	// Notify configuration changed
	c.eventHandler.OnConfigChanged(mergedConfig)

	return mergedConfig, nil
}

// configureComponents configures all application components with the final configuration
func (c *Controller) configureComponents(config *Configuration) error {
	// Configure display renderer
	if c.displayRenderer != nil {
		if err := c.displayRenderer.SetConfiguration(config); err != nil {
			return fmt.Errorf("failed to configure display renderer: %w", err)
		}
	}

	// Configure watch coordinator if watch mode is enabled
	if config.Watch.Enabled && c.watchCoordinator != nil {
		watchOptions := core.WatchOptions{
			Paths:            config.Paths.IncludePatterns,
			IgnorePatterns:   config.Watch.IgnorePatterns,
			TestPatterns:     []string{"*_test.go"},
			Mode:             core.WatchAll,          // Default mode
			DebounceInterval: 100 * time.Millisecond, // Parse from config.Watch.Debounce
			ClearTerminal:    config.Watch.ClearOnRerun,
			RunOnStart:       config.Watch.RunOnStart,
			Writer:           os.Stdout,
		}

		if err := c.watchCoordinator.Configure(watchOptions); err != nil {
			return fmt.Errorf("failed to configure watch coordinator: %w", err)
		}
	}

	return nil
}

// executeApplication executes the main application logic
func (c *Controller) executeApplication(config *Configuration, args *Arguments) error {
	fmt.Printf("🚀 Running tests with go-sentinel...\n\n")

	if config.Watch.Enabled {
		return c.executeWatchMode(config)
	} else {
		return c.executeSingleMode(config, args)
	}
}

// executeSingleMode runs tests once and exits
func (c *Controller) executeSingleMode(config *Configuration, args *Arguments) error {
	startTime := time.Now()

	// Determine packages to test
	packages := args.Packages
	if len(packages) == 0 {
		packages = []string{"./..."}
	}

	// Execute tests
	if err := c.testExecutor.ExecuteSingle(c.ctx, packages, config); err != nil {
		return fmt.Errorf("test execution failed: %w", err)
	}

	// Render results
	if err := c.displayRenderer.RenderResults(c.ctx); err != nil {
		return fmt.Errorf("failed to render results: %w", err)
	}

	// Display timing
	duration := time.Since(startTime)
	fmt.Printf("\n⏱️  Tests completed in %v\n", duration)

	return nil
}

// executeWatchMode runs tests in watch mode
func (c *Controller) executeWatchMode(config *Configuration) error {
	fmt.Printf("👀 Starting watch mode...\n")

	// Start watch coordinator
	if err := c.watchCoordinator.Start(c.ctx); err != nil {
		return fmt.Errorf("failed to start watch coordinator: %w", err)
	}

	// Wait for context cancellation (shutdown signal)
	<-c.ctx.Done()

	return nil
}

// resolveDependencies resolves required dependencies from the container
func (c *Controller) resolveDependencies() error {
	// Resolve test executor
	if err := c.container.ResolveAs("testExecutor", &c.testExecutor); err != nil {
		return fmt.Errorf("failed to resolve test executor: %w", err)
	}

	// Resolve watch coordinator
	if err := c.container.ResolveAs("watchCoordinator", &c.watchCoordinator); err != nil {
		return fmt.Errorf("failed to resolve watch coordinator: %w", err)
	}

	// Resolve display renderer
	if err := c.container.ResolveAs("displayRenderer", &c.displayRenderer); err != nil {
		return fmt.Errorf("failed to resolve display renderer: %w", err)
	}

	return nil
}

// cleanup performs cleanup operations
func (c *Controller) cleanup() error {
	// Cleanup dependency container
	if err := c.container.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup dependency container: %w", err)
	}

	return nil
}
