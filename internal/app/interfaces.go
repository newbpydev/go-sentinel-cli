// Package app provides application orchestration and lifecycle management
package app

import (
	"context"
	"io"
)

// ApplicationController orchestrates the main application flow
type ApplicationController interface {
	// Run executes the main application flow with the given arguments
	Run(args []string) error

	// Initialize sets up the application with dependencies
	Initialize() error

	// Shutdown gracefully shuts down the application
	Shutdown(ctx context.Context) error
}

// LifecycleManager manages application startup and shutdown
type LifecycleManager interface {
	// Startup initializes all application components
	Startup(ctx context.Context) error

	// Shutdown gracefully stops all application components
	Shutdown(ctx context.Context) error

	// IsRunning returns whether the application is currently running
	IsRunning() bool

	// RegisterShutdownHook adds a function to be called during shutdown
	RegisterShutdownHook(hook func() error)
}

// DependencyContainer manages component dependencies and injection
type DependencyContainer interface {
	// Register registers a component with the container
	Register(name string, component interface{}) error

	// Resolve retrieves a component from the container
	Resolve(name string) (interface{}, error)

	// ResolveAs retrieves a component and casts it to the specified type
	ResolveAs(name string, target interface{}) error

	// Initialize initializes all registered components
	Initialize() error

	// Cleanup cleans up all registered components
	Cleanup() error
}

// ArgumentParser handles command-line argument parsing
type ArgumentParser interface {
	// Parse parses command-line arguments into a structured format
	Parse(args []string) (*Arguments, error)

	// Help returns help text for the application
	Help() string

	// Version returns version information
	Version() string
}

// ConfigurationLoader handles application configuration loading
type ConfigurationLoader interface {
	// LoadFromFile loads configuration from a file
	LoadFromFile(path string) (*Configuration, error)

	// LoadFromDefaults returns default configuration
	LoadFromDefaults() *Configuration

	// Merge merges CLI arguments with configuration
	Merge(config *Configuration, args *Arguments) *Configuration

	// Validate validates the final configuration
	Validate(config *Configuration) error
}

// ApplicationEventHandler handles application-level events
type ApplicationEventHandler interface {
	// OnStartup is called when the application starts
	OnStartup(ctx context.Context) error

	// OnShutdown is called when the application shuts down
	OnShutdown(ctx context.Context) error

	// OnError is called when an error occurs
	OnError(err error)

	// OnConfigChanged is called when configuration changes
	OnConfigChanged(config *Configuration)
}

// Arguments represents parsed command-line arguments
type Arguments struct {
	// Packages to test
	Packages []string

	// Watch mode enabled
	Watch bool

	// Verbose output
	Verbose bool

	// Colors enabled
	Colors bool

	// Optimization enabled
	Optimized bool

	// Optimization mode
	OptimizationMode string

	// Output writer
	Writer io.Writer
}

// Configuration represents application configuration
type Configuration struct {
	// Watch configuration
	Watch WatchConfig

	// Paths configuration
	Paths PathsConfig

	// Visual configuration
	Visual VisualConfig

	// Test configuration
	Test TestConfig

	// Colors enabled
	Colors bool

	// Verbosity level
	Verbosity int
}

// WatchConfig represents watch-specific configuration
type WatchConfig struct {
	// Enabled indicates if watch mode is enabled
	Enabled bool

	// IgnorePatterns lists patterns to ignore
	IgnorePatterns []string

	// Debounce duration for file events
	Debounce string

	// RunOnStart runs tests on startup
	RunOnStart bool

	// ClearOnRerun clears screen between runs
	ClearOnRerun bool
}

// PathsConfig represents path-specific configuration
type PathsConfig struct {
	// IncludePatterns lists patterns to include
	IncludePatterns []string

	// ExcludePatterns lists patterns to exclude
	ExcludePatterns []string
}

// VisualConfig represents visual/UI configuration
type VisualConfig struct {
	// Icons setting (none, simple, rich)
	Icons string

	// Theme setting
	Theme string

	// TerminalWidth for display formatting
	TerminalWidth int
}

// TestConfig represents test execution configuration
type TestConfig struct {
	// Timeout for test execution
	Timeout string

	// Parallel execution settings
	Parallel int

	// Coverage settings
	Coverage bool
}
