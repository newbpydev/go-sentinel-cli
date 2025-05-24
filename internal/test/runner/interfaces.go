// Package runner provides test execution interfaces and implementations
package runner

import (
	"context"
	"io"
	"time"
)

// TestExecutor handles test execution with different strategies
type TestExecutor interface {
	// Execute runs tests for the specified packages
	Execute(ctx context.Context, packages []string, options *ExecutionOptions) (*ExecutionResult, error)

	// ExecutePackage runs tests for a single package
	ExecutePackage(ctx context.Context, pkg string, options *ExecutionOptions) (*PackageResult, error)

	// Cancel cancels the current test execution
	Cancel() error

	// IsRunning returns whether tests are currently running
	IsRunning() bool
}

// OptimizedExecutor provides optimized test execution with caching
type OptimizedExecutor interface {
	TestExecutor

	// EnableOptimization enables test execution optimization
	EnableOptimization() error

	// DisableOptimization disables test execution optimization
	DisableOptimization() error

	// IsOptimizationEnabled returns whether optimization is enabled
	IsOptimizationEnabled() bool

	// SetOptimizationMode configures the optimization strategy
	SetOptimizationMode(mode OptimizationMode) error
}

// ParallelExecutor provides parallel test execution capabilities
type ParallelExecutor interface {
	TestExecutor

	// SetParallelism configures the number of parallel test processes
	SetParallelism(count int) error

	// GetParallelism returns the current parallelism setting
	GetParallelism() int

	// ExecuteParallel runs multiple packages in parallel
	ExecuteParallel(ctx context.Context, packages []string, options *ExecutionOptions) (*ExecutionResult, error)
}

// ExecutionOptions configures test execution behavior
type ExecutionOptions struct {
	// Verbose enables verbose output
	Verbose bool

	// JSONOutput enables JSON output format
	JSONOutput bool

	// Coverage enables coverage reporting
	Coverage bool

	// CoverageProfile specifies the coverage profile file
	CoverageProfile string

	// Timeout specifies the execution timeout
	Timeout time.Duration

	// Parallel specifies the number of parallel processes
	Parallel int

	// Args contains additional arguments to pass to go test
	Args []string

	// Env contains environment variables for test execution
	Env map[string]string

	// WorkingDirectory specifies the working directory for execution
	WorkingDirectory string

	// Output specifies where to write test output
	Output io.Writer
}

// ExecutionResult represents the result of test execution
type ExecutionResult struct {
	// Packages contains results for each package
	Packages []*PackageResult

	// TotalDuration is the total execution time
	TotalDuration time.Duration

	// Success indicates if all tests passed
	Success bool

	// TotalTests is the total number of tests run
	TotalTests int

	// PassedTests is the number of tests that passed
	PassedTests int

	// FailedTests is the number of tests that failed
	FailedTests int

	// SkippedTests is the number of tests that were skipped
	SkippedTests int

	// Coverage is the overall coverage percentage
	Coverage float64

	// Output contains the raw test output
	Output string

	// StartTime indicates when execution started
	StartTime time.Time

	// EndTime indicates when execution finished
	EndTime time.Time
}

// PackageResult represents the result of testing a single package
type PackageResult struct {
	// Package is the package name/path
	Package string

	// Success indicates if all tests in the package passed
	Success bool

	// Duration is the execution time for this package
	Duration time.Duration

	// Tests contains individual test results
	Tests []*TestResult

	// Coverage is the coverage percentage for this package
	Coverage float64

	// Output contains the raw output for this package
	Output string

	// Error contains any error that occurred during execution
	Error error
}

// TestResult represents the result of a single test
type TestResult struct {
	// Name is the test name
	Name string

	// Package is the package containing the test
	Package string

	// Status is the test status (pass, fail, skip)
	Status TestStatus

	// Duration is the test execution time
	Duration time.Duration

	// Output contains the test output
	Output string

	// Error contains any error message
	Error string
}

// TestStatus represents the status of a test
type TestStatus string

const (
	// TestStatusPass indicates the test passed
	TestStatusPass TestStatus = "pass"

	// TestStatusFail indicates the test failed
	TestStatusFail TestStatus = "fail"

	// TestStatusSkip indicates the test was skipped
	TestStatusSkip TestStatus = "skip"

	// TestStatusRunning indicates the test is currently running
	TestStatusRunning TestStatus = "running"
)

// OptimizationMode represents different optimization strategies
type OptimizationMode string

const (
	// OptimizationModeNone disables optimization
	OptimizationModeNone OptimizationMode = "none"

	// OptimizationModeBasic enables basic optimization
	OptimizationModeBasic OptimizationMode = "basic"

	// OptimizationModeBalanced enables balanced optimization
	OptimizationModeBalanced OptimizationMode = "balanced"

	// OptimizationModeAggressive enables aggressive optimization
	OptimizationModeAggressive OptimizationMode = "aggressive"
)

// ExecutionStrategy represents different execution strategies
type ExecutionStrategy string

const (
	// ExecutionStrategySequential runs tests sequentially
	ExecutionStrategySequential ExecutionStrategy = "sequential"

	// ExecutionStrategyParallel runs tests in parallel
	ExecutionStrategyParallel ExecutionStrategy = "parallel"

	// ExecutionStrategyOptimized uses optimized execution with caching
	ExecutionStrategyOptimized ExecutionStrategy = "optimized"
)
