// Package app provides the main application controller implementation
package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/newbpydev/go-sentinel/internal/watch/core"
	"github.com/newbpydev/go-sentinel/pkg/models"
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

	// Register components in the dependency container
	if err := c.registerComponents(); err != nil {
		return fmt.Errorf("failed to register components: %w", err)
	}

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

// registerComponents registers all required components in the dependency container
func (c *Controller) registerComponents() error {
	// Register test executor
	testExecutor := NewTestExecutor()
	if err := c.container.Register("testExecutor", testExecutor); err != nil {
		return fmt.Errorf("failed to register testExecutor: %w", err)
	}

	// Register display renderer
	displayRenderer := NewDisplayRenderer()
	if err := c.container.Register("displayRenderer", displayRenderer); err != nil {
		return fmt.Errorf("failed to register displayRenderer: %w", err)
	}

	// Register watch coordinator (create a factory function)
	watchCoordinatorFactory := func() (interface{}, error) {
		// This will be created when needed based on configuration
		return nil, fmt.Errorf("watch coordinator not yet implemented")
	}
	if err := c.container.Register("watchCoordinator", watchCoordinatorFactory); err != nil {
		return fmt.Errorf("failed to register watchCoordinator factory: %w", err)
	}

	return nil
}

// Run implements the ApplicationController interface
func (c *Controller) Run(args []string) error {
	// Mark as running
	c.isRunning = true
	defer func() { c.isRunning = false }()

	// Notify startup
	if err := c.eventHandler.OnStartup(c.ctx); err != nil {
		return models.NewLifecycleError("startup", err).
			WithContext("component", "event_handler")
	}

	// Step 1: Parse CLI arguments
	parsedArgs, err := c.argParser.Parse(args)
	if err != nil {
		return models.WrapError(err, models.ErrorTypeValidation, models.SeverityWarning, "failed to parse command line arguments").
			WithContext("operation", "argument_parsing").
			WithContext("args", fmt.Sprintf("%v", args))
	}
	c.args = parsedArgs

	// Step 2: Load and merge configuration
	config, err := c.loadAndMergeConfiguration(parsedArgs)
	if err != nil {
		return models.WrapError(err, models.ErrorTypeConfig, models.SeverityError, "failed to load configuration").
			WithContext("operation", "config_loading")
	}
	c.config = config

	// Step 3: Validate configuration
	if err := c.configLoader.Validate(config); err != nil {
		return models.WrapError(err, models.ErrorTypeValidation, models.SeverityWarning, "configuration validation failed").
			WithContext("operation", "config_validation")
	}

	// Step 4: Configure components
	if err := c.configureComponents(config); err != nil {
		return models.WrapError(err, models.ErrorTypeDependency, models.SeverityError, "failed to configure application components").
			WithContext("operation", "component_configuration")
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
		// Log warning but don't fail shutdown
		fmt.Printf("Warning: %s\n", models.SanitizeError(err).Error())
	}

	// Shutdown lifecycle manager
	if err := c.lifecycle.Shutdown(ctx); err != nil {
		return models.NewLifecycleError("shutdown", err).
			WithContext("component", "lifecycle_manager")
	}

	return nil
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
			return nil, models.NewFileSystemError("read_config", configPath, err).
				WithContext("config_type", "file")
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
			return models.NewDependencyError("displayRenderer", err).
				WithContext("operation", "configure_display")
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
			return models.NewDependencyError("watchCoordinator", err).
				WithContext("operation", "configure_watch").
				WithContext("watch_mode", "enabled")
		}
	}

	return nil
}

// executeApplication executes the main application logic
func (c *Controller) executeApplication(config *Configuration, args *Arguments) error {
	fmt.Printf("🚀 Running tests with go-sentinel...\n\n")

	if config.Watch.Enabled {
		return c.executeWatchMode()
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
		return models.NewTestExecutionError(fmt.Sprintf("%v", packages), err).
			WithContext("mode", "single").
			WithContext("package_count", fmt.Sprintf("%d", len(packages)))
	}

	// Render results
	if err := c.displayRenderer.RenderResults(c.ctx); err != nil {
		return models.WrapError(err, models.ErrorTypeInternal, models.SeverityError, "failed to render test results").
			WithContext("operation", "render_results").
			WithContext("mode", "single")
	}

	// Display timing
	duration := time.Since(startTime)
	fmt.Printf("\n⏱️  Tests completed in %v\n", duration)

	return nil
}

// executeWatchMode runs tests in watch mode
func (c *Controller) executeWatchMode() error {
	fmt.Printf("👀 Starting watch mode...\n")

	// Start watch coordinator
	if err := c.watchCoordinator.Start(c.ctx); err != nil {
		return models.NewWatchError("start_coordinator", "", err).
			WithContext("mode", "watch")
	}

	// Wait for context cancellation (shutdown signal)
	<-c.ctx.Done()

	return nil
}

// resolveDependencies resolves required dependencies from the container
func (c *Controller) resolveDependencies() error {
	// Resolve test executor
	if err := c.container.ResolveAs("testExecutor", &c.testExecutor); err != nil {
		return models.NewDependencyError("testExecutor", err).
			WithContext("operation", "resolve_dependency")
	}

	// Resolve watch coordinator
	if err := c.container.ResolveAs("watchCoordinator", &c.watchCoordinator); err != nil {
		return models.NewDependencyError("watchCoordinator", err).
			WithContext("operation", "resolve_dependency")
	}

	// Resolve display renderer
	if err := c.container.ResolveAs("displayRenderer", &c.displayRenderer); err != nil {
		return models.NewDependencyError("displayRenderer", err).
			WithContext("operation", "resolve_dependency")
	}

	return nil
}

// cleanup performs cleanup operations
func (c *Controller) cleanup() error {
	// Cleanup dependency container
	if err := c.container.Cleanup(); err != nil {
		return models.NewLifecycleError("cleanup", err).
			WithContext("component", "dependency_container")
	}

	return nil
}
