// Package core provides foundational types for the watch system
package core

import (
	"io"
	"time"
)

// WatchOptions configures the watch system behavior
type WatchOptions struct {
	// Paths to monitor for changes
	Paths []string

	// Patterns to ignore during monitoring
	IgnorePatterns []string

	// Test file patterns to identify test files
	TestPatterns []string

	// Watch mode type
	Mode WatchMode

	// Debounce interval to avoid running tests too frequently
	DebounceInterval time.Duration

	// Clear terminal between test runs
	ClearTerminal bool

	// Run tests on startup
	RunOnStart bool

	// Writer for console output
	Writer io.Writer
}

// WatchStatus represents the current status of the watch system
type WatchStatus struct {
	// IsRunning indicates if the watch system is active
	IsRunning bool

	// WatchedPaths lists the currently monitored paths
	WatchedPaths []string

	// Mode indicates the current watch mode
	Mode WatchMode

	// StartTime indicates when watching began
	StartTime time.Time

	// LastEventTime indicates when the last file event was processed
	LastEventTime time.Time

	// EventCount tracks the total number of events processed
	EventCount int

	// ErrorCount tracks the number of errors encountered
	ErrorCount int
}

// ChangeType represents the type of file change
type ChangeType string

const (
	// ChangeTypeModified indicates a file was modified
	ChangeTypeModified ChangeType = "modified"

	// ChangeTypeAdded indicates a file was added
	ChangeTypeAdded ChangeType = "added"

	// ChangeTypeDeleted indicates a file was deleted
	ChangeTypeDeleted ChangeType = "deleted"

	// ChangeTypeRenamed indicates a file was renamed
	ChangeTypeRenamed ChangeType = "renamed"
)

// ChangeImpact represents the impact of a single file change
type ChangeImpact struct {
	// FilePath is the path of the changed file
	FilePath string

	// Type indicates the type of change
	Type ChangeType

	// IsTest indicates if the changed file is a test file
	IsTest bool

	// AffectedTests lists test files that should be run due to this change
	AffectedTests []string

	// IsNew indicates if this is a newly created file
	IsNew bool

	// Timestamp indicates when the change was detected
	Timestamp time.Time

	// Priority indicates the relative importance of this change
	Priority ChangePriority
}

// ChangePriority represents the priority level of a file change
type ChangePriority int

const (
	// PriorityLow for minor changes that may not require immediate test runs
	PriorityLow ChangePriority = iota

	// PriorityMedium for standard code changes
	PriorityMedium

	// PriorityHigh for critical changes that should trigger immediate test runs
	PriorityHigh

	// PriorityCritical for emergency changes requiring full test suite
	PriorityCritical
)

// BatchImpact represents the combined impact of multiple file changes
type BatchImpact struct {
	// Changes lists all individual change impacts
	Changes []*ChangeImpact

	// TotalFiles indicates the number of files changed
	TotalFiles int

	// UniqueTestFiles lists all test files that should be run
	UniqueTestFiles []string

	// ShouldRunAllTests indicates if all tests should be run
	ShouldRunAllTests bool

	// HighestPriority indicates the highest priority among all changes
	HighestPriority ChangePriority

	// ProcessingTime indicates how long the analysis took
	ProcessingTime time.Duration
}

// WatchEvent represents an internal watch system event
type WatchEvent struct {
	// Type indicates the type of watch event
	Type WatchEventType

	// Message provides details about the event
	Message string

	// Timestamp indicates when the event occurred
	Timestamp time.Time

	// Data contains additional event-specific data
	Data interface{}
}

// WatchEventType represents the type of watch system event
type WatchEventType string

const (
	// WatchEventStarted indicates the watch system has started
	WatchEventStarted WatchEventType = "started"

	// WatchEventStopped indicates the watch system has stopped
	WatchEventStopped WatchEventType = "stopped"

	// WatchEventError indicates an error occurred
	WatchEventError WatchEventType = "error"

	// WatchEventFileChanged indicates a file was changed
	WatchEventFileChanged WatchEventType = "file_changed"

	// WatchEventTestsTriggered indicates tests were triggered
	WatchEventTestsTriggered WatchEventType = "tests_triggered"

	// WatchEventConfigUpdated indicates configuration was updated
	WatchEventConfigUpdated WatchEventType = "config_updated"
)

// TestExecutionResult represents the result of a test execution
type TestExecutionResult struct {
	// TestPaths lists the test paths that were executed
	TestPaths []string

	// Success indicates if all tests passed
	Success bool

	// Duration indicates how long the tests took to run
	Duration time.Duration

	// Output contains the test execution output
	Output string

	// ErrorMessage contains any error message if execution failed
	ErrorMessage string

	// Timestamp indicates when the execution completed
	Timestamp time.Time
}

// FilePattern represents a file pattern with metadata
type FilePattern struct {
	// Pattern is the glob or regex pattern
	Pattern string

	// Type indicates if this is a glob or regex pattern
	Type PatternType

	// Recursive indicates if the pattern should match recursively
	Recursive bool

	// CaseSensitive indicates if the pattern matching should be case sensitive
	CaseSensitive bool
}

// PatternType represents the type of pattern matching
type PatternType string

const (
	// PatternTypeGlob for shell-style glob patterns
	PatternTypeGlob PatternType = "glob"

	// PatternTypeRegex for regular expression patterns
	PatternTypeRegex PatternType = "regex"

	// PatternTypeExact for exact string matching
	PatternTypeExact PatternType = "exact"
)
