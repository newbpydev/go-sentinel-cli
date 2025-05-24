// Package core provides foundational interfaces and types for the watch system
package core

import (
	"context"
	"time"
)

// FileEvent represents a file system event with metadata
type FileEvent struct {
	Path      string    // Full path to the changed file
	Type      string    // Type of event (create, write, remove, rename, chmod)
	Timestamp time.Time // When the event occurred
	IsTest    bool      // Whether this is a test file
}

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

// FileSystemWatcher provides file system monitoring capabilities
type FileSystemWatcher interface {
	// Watch starts monitoring for file changes and sends events to the channel
	// The method blocks until the context is cancelled or an error occurs
	Watch(ctx context.Context, events chan<- FileEvent) error

	// AddPath adds a new path to be monitored
	AddPath(path string) error

	// RemovePath removes a path from monitoring
	RemovePath(path string) error

	// Close releases all resources used by the watcher
	Close() error
}

// EventProcessor processes file system events with filtering capabilities
type EventProcessor interface {
	// ProcessEvent processes a single file event
	ProcessEvent(event FileEvent) error

	// ProcessBatch processes multiple file events efficiently
	ProcessBatch(events []FileEvent) error

	// SetFilters configures the patterns to ignore during processing
	SetFilters(ignorePatterns []string) error

	// ShouldProcess determines if an event should be processed based on filters
	ShouldProcess(event FileEvent) bool
}

// EventDebouncer manages temporal grouping of file events to avoid excessive triggering
type EventDebouncer interface {
	// AddEvent adds a file event to the debouncer
	AddEvent(event FileEvent)

	// Events returns a channel that emits debounced event batches
	Events() <-chan []FileEvent

	// SetInterval configures the debounce interval
	SetInterval(interval time.Duration)

	// Stop stops the debouncer and closes all channels
	Stop() error
}

// TestTrigger handles triggering test execution based on file changes
type TestTrigger interface {
	// TriggerTestsForFile triggers tests for a specific file
	TriggerTestsForFile(ctx context.Context, filePath string) error

	// TriggerAllTests triggers all tests in the workspace
	TriggerAllTests(ctx context.Context) error

	// TriggerRelatedTests triggers tests related to the changed file
	TriggerRelatedTests(ctx context.Context, filePath string) error

	// GetTestTargets determines which tests should be run for given changes
	GetTestTargets(changes []FileEvent) ([]string, error)
}

// WatchCoordinator orchestrates the entire watch system
type WatchCoordinator interface {
	// Start begins watching for file changes with the provided configuration
	Start(ctx context.Context) error

	// Stop gracefully stops all watch operations
	Stop() error

	// HandleFileChanges processes a batch of file changes
	HandleFileChanges(changes []FileEvent) error

	// Configure updates the watch system configuration
	Configure(options WatchOptions) error

	// GetStatus returns the current status of the watch system
	GetStatus() WatchStatus
}

// PatternMatcher provides pattern matching capabilities for file paths
type PatternMatcher interface {
	// MatchesAny checks if a path matches any of the provided patterns
	MatchesAny(path string, patterns []string) bool

	// MatchesPattern checks if a path matches a specific pattern
	MatchesPattern(path string, pattern string) bool

	// AddPattern adds a new pattern to the matcher
	AddPattern(pattern string) error

	// RemovePattern removes a pattern from the matcher
	RemovePattern(pattern string) error
}

// TestFileFinder provides capabilities to find test files related to implementation files
type TestFileFinder interface {
	// FindTestFile finds the test file corresponding to the given implementation file
	FindTestFile(filePath string) (string, error)

	// FindImplementationFile finds the implementation file for a given test file
	FindImplementationFile(testPath string) (string, error)

	// FindPackageTests finds all test files in the same package as the given file
	FindPackageTests(filePath string) ([]string, error)

	// IsTestFile determines if the given file is a test file
	IsTestFile(filePath string) bool
}

// ChangeAnalyzer analyzes file changes for test impact assessment
type ChangeAnalyzer interface {
	// AnalyzeChange determines the impact of a file change
	AnalyzeChange(filePath string) (*ChangeImpact, error)

	// AnalyzeBatch analyzes multiple file changes for overall impact
	AnalyzeBatch(changes []FileEvent) (*BatchImpact, error)

	// ShouldRunTests determines if tests should be run based on changes
	ShouldRunTests(changes []FileEvent) (bool, []string)
}
