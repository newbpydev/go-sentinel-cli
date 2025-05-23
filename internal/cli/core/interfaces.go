package core

import (
	"context"
	"io"
	"time"
)

// TestRunner defines the interface for test execution engines
type TestRunner interface {
	// RunTests executes tests based on the provided changes and strategy
	RunTests(ctx context.Context, changes []FileChange, strategy ExecutionStrategy) (*TestResult, error)

	// GetCapabilities returns the capabilities of this test runner
	GetCapabilities() RunnerCapabilities
}

// ChangeAnalyzer analyzes file changes and determines impact
type ChangeAnalyzer interface {
	// AnalyzeChange examines a file change and determines its impact
	AnalyzeChange(filePath string) (*FileChange, error)

	// DetermineAffectedTests finds tests affected by the given changes
	DetermineAffectedTests(changes []FileChange) ([]TestTarget, error)

	// IsRelevantChange determines if a change should trigger test execution
	IsRelevantChange(change FileChange) bool
}

// CacheManager handles test result caching and invalidation
type CacheManager interface {
	// GetCachedResult retrieves a cached test result if available and valid
	GetCachedResult(target TestTarget) (*CachedResult, bool)

	// StoreResult caches a test result
	StoreResult(target TestTarget, result *TestResult)

	// InvalidateCache invalidates cache entries based on file changes
	InvalidateCache(changes []FileChange)

	// Clear removes all cached results
	Clear()

	// GetStats returns cache statistics
	GetStats() CacheStats
}

// FileWatcher monitors file system changes
type FileWatcher interface {
	// Watch starts monitoring the specified paths for changes
	Watch(paths []string) (<-chan FileEvent, error)

	// Stop stops the file watcher
	Stop() error

	// AddPath adds a new path to watch
	AddPath(path string) error

	// RemovePath removes a path from watching
	RemovePath(path string) error
}

// EventDebouncer manages rapid file change events
type EventDebouncer interface {
	// ProcessEvents takes raw file events and returns debounced changes
	ProcessEvents(events <-chan FileEvent) <-chan []FileChange

	// SetDebounceInterval configures the debounce timing
	SetDebounceInterval(interval time.Duration)
}

// ResultRenderer handles test result output formatting
type ResultRenderer interface {
	// RenderResults formats and displays test results
	RenderResults(result *TestResult, writer io.Writer) error

	// RenderProgress displays ongoing test progress
	RenderProgress(progress TestProgress, writer io.Writer) error

	// RenderSummary displays a summary of test execution
	RenderSummary(summary TestSummary, writer io.Writer) error
}

// OutputFormatter provides formatting capabilities
type OutputFormatter interface {
	// FormatDuration formats time durations for display
	FormatDuration(duration time.Duration) string

	// FormatPercentage formats percentages for display
	FormatPercentage(value float64) string

	// FormatCount formats counts with appropriate units
	FormatCount(count int, singular, plural string) string
}

// ColorProvider provides color formatting
type ColorProvider interface {
	// Success returns success color formatting
	Success(text string) string

	// Error returns error color formatting
	Error(text string) string

	// Warning returns warning color formatting
	Warning(text string) string

	// Info returns info color formatting
	Info(text string) string

	// Dim returns dimmed text formatting
	Dim(text string) string
}

// IconProvider provides icon characters
type IconProvider interface {
	// GetIcon returns the appropriate icon for the given context
	GetIcon(iconType IconType) string

	// IsSupported returns whether icons are supported in current environment
	IsSupported() bool
}

// ConfigLoader handles configuration loading and validation
type ConfigLoader interface {
	// LoadConfig loads configuration from various sources
	LoadConfig(args []string) (*Config, error)

	// ValidateConfig validates the configuration
	ValidateConfig(config *Config) error

	// GetDefaults returns default configuration values
	GetDefaults() *Config
}

// ExecutionStrategy defines how tests should be executed
type ExecutionStrategy interface {
	// ShouldRunTest determines if a test should be executed
	ShouldRunTest(target TestTarget, cache CacheManager) bool

	// GetExecutionOrder determines the order of test execution
	GetExecutionOrder(targets []TestTarget) []TestTarget

	// GetName returns the strategy name
	GetName() string
}

// Controller coordinates the overall application flow
type Controller interface {
	// Run starts the application with the given configuration
	Run(ctx context.Context, config *Config) error

	// RunOnce executes tests once without watching
	RunOnce(ctx context.Context, config *Config) error

	// RunWatch starts watch mode
	RunWatch(ctx context.Context, config *Config) error
}
