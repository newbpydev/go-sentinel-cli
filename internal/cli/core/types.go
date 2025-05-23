package core

import (
	"time"
)

// ChangeType represents the type of file change
type ChangeType int

const (
	ChangeTypeTest ChangeType = iota
	ChangeTypeSource
	ChangeTypeConfig
	ChangeTypeDependency
	ChangeTypeUnknown
)

func (ct ChangeType) String() string {
	switch ct {
	case ChangeTypeTest:
		return "test"
	case ChangeTypeSource:
		return "source"
	case ChangeTypeConfig:
		return "config"
	case ChangeTypeDependency:
		return "dependency"
	default:
		return "unknown"
	}
}

// TestStatus represents the status of a test
type TestStatus int

const (
	StatusPending TestStatus = iota
	StatusRunning
	StatusPassed
	StatusFailed
	StatusSkipped
	StatusError
)

func (ts TestStatus) String() string {
	switch ts {
	case StatusPending:
		return "pending"
	case StatusRunning:
		return "running"
	case StatusPassed:
		return "passed"
	case StatusFailed:
		return "failed"
	case StatusSkipped:
		return "skipped"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// IconType represents different types of icons
type IconType int

const (
	IconTypeSuccess IconType = iota
	IconTypeError
	IconTypeWarning
	IconTypeInfo
	IconTypeLoading
	IconTypeTest
	IconTypeSource
	IconTypeConfig
	IconTypeCache
	IconTypeWatch
)

// FileChange represents a detected file change
type FileChange struct {
	Path          string
	Type          ChangeType
	IsNew         bool
	IsDeleted     bool
	Hash          string
	Timestamp     time.Time
	AffectedTests []string
}

// FileEvent represents a raw file system event
type FileEvent struct {
	Path      string
	Operation string // "create", "write", "remove", "rename"
	Timestamp time.Time
}

// TestTarget represents a target for test execution
type TestTarget struct {
	Path              string        // File or package path
	Type              string        // "package", "file", "function"
	Functions         []string      // Specific test functions (if applicable)
	Priority          int           // Execution priority
	EstimatedDuration time.Duration // Estimated execution time
}

// TestResult represents the result of test execution
type TestResult struct {
	Target       TestTarget
	Status       TestStatus
	Output       string
	Duration     time.Duration
	StartTime    time.Time
	EndTime      time.Time
	TestCount    int
	PassedCount  int
	FailedCount  int
	SkippedCount int
	CacheHit     bool
	Error        error
}

// CachedResult represents a cached test result
type CachedResult struct {
	Result       *TestResult
	CacheTime    time.Time
	Dependencies []string
	Hash         string
	IsValid      bool
}

// TestProgress represents ongoing test progress
type TestProgress struct {
	CurrentTarget      TestTarget
	Completed          int
	Total              int
	ElapsedTime        time.Duration
	EstimatedRemaining time.Duration
}

// TestSummary represents a summary of test execution
type TestSummary struct {
	TotalTargets int
	TotalTests   int
	Passed       int
	Failed       int
	Skipped      int
	CacheHits    int
	Duration     time.Duration
	CacheHitRate float64
	Efficiency   float64
}

// CacheStats represents cache statistics
type CacheStats struct {
	TotalEntries   int
	ValidEntries   int
	InvalidEntries int
	HitRate        float64
	MemoryUsage    int64
	OldestEntry    time.Time
	NewestEntry    time.Time
}

// RunnerCapabilities represents what a test runner can do
type RunnerCapabilities struct {
	SupportsCaching    bool
	SupportsParallel   bool
	SupportsWatchMode  bool
	SupportsFiltering  bool
	MaxConcurrency     int
	SupportedFileTypes []string
}

// Config represents application configuration
type Config struct {
	// General settings
	Verbose bool
	Debug   bool
	NoColor bool
	NoIcons bool

	// Watch mode settings
	WatchMode        bool
	WatchPaths       []string
	DebounceInterval time.Duration

	// Test execution settings
	UseCache       bool
	CacheStrategy  string
	MaxConcurrency int
	TestTimeout    time.Duration
	FailFast       bool

	// Output settings
	OutputFormat string
	OutputFile   string
	ShowProgress bool
	ShowSummary  bool

	// Filter settings
	TestPattern    string
	ExcludePattern string
	TestFunctions  []string
}

// ExecutionMode represents different execution modes
type ExecutionMode int

const (
	ModeStandard ExecutionMode = iota
	ModeWatch
	ModeOnce
	ModeDryRun
)

func (em ExecutionMode) String() string {
	switch em {
	case ModeStandard:
		return "standard"
	case ModeWatch:
		return "watch"
	case ModeOnce:
		return "once"
	case ModeDryRun:
		return "dry-run"
	default:
		return "unknown"
	}
}

// TestSuite represents a collection of tests
type TestSuite struct {
	Name         string
	FilePath     string
	TestCount    int
	PassedCount  int
	FailedCount  int
	SkippedCount int
	Duration     time.Duration
	Status       TestStatus
	Error        error
}

// TestRunStats represents statistics about a test run
type TestRunStats struct {
	TotalSuites  int
	TotalTests   int
	PassedTests  int
	FailedTests  int
	SkippedTests int
	CachedSuites int
	Duration     time.Duration
	CacheHitRate float64
	StartTime    time.Time
	EndTime      time.Time
}
