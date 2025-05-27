# 📊 Models Package (pkg/models)

[![Test Coverage](https://img.shields.io/badge/coverage-100.0%25-brightgreen.svg)](https://github.com/newbpydev/go-sentinel/tree/main/pkg/models)
[![Go Report Card](https://goreportcard.com/badge/github.com/newbpydev/go-sentinel/pkg/models)](https://goreportcard.com/report/github.com/newbpydev/go-sentinel/pkg/models)
[![Go Reference](https://pkg.go.dev/badge/github.com/newbpydev/go-sentinel/pkg/models.svg)](https://pkg.go.dev/github.com/newbpydev/go-sentinel/pkg/models)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 📖 Overview

The `pkg/models` package provides shared data models and value objects for the Go Sentinel CLI. It defines core data structures for test results, package results, coverage information, error handling, and configuration management that are used across all components of the application.

### 🎯 Key Features

- **Test Result Models**: Comprehensive test execution result structures
- **Package Result Models**: Package-level test aggregation and statistics
- **Coverage Models**: Detailed code coverage information and metrics
- **Error Models**: Rich error information with stack traces and context
- **Configuration Models**: Test and watch configuration structures
- **File Change Models**: File system change tracking and metadata
- **Type Safety**: Strongly typed enums and constants
- **Validation**: Built-in validation methods and helper functions
- **Serialization**: JSON-compatible structures for persistence and API usage

## 🏗️ Architecture

This package follows clean architecture principles:

- **Value Objects**: Immutable data structures with behavior
- **Domain Models**: Rich domain objects with business logic
- **Data Transfer Objects**: Serializable structures for communication
- **Type Safety**: Strongly typed enums and constants

### 📦 Package Structure

```
pkg/models/
├── interfaces.go       # Core model interfaces and contracts
├── core_models.go     # Fundamental data structures
├── errors.go          # Error models and types
├── test_types.go      # Test-specific type definitions
├── examples.go        # Usage examples and factory functions
└── *_test.go         # Comprehensive test suite (100% coverage)
```

## 🚀 Quick Start

### Basic Test Result Usage

```go
package main

import (
    "time"
    "github.com/newbpydev/go-sentinel/pkg/models"
)

func main() {
    // Create a new test result
    testResult := models.NewTestResult("TestExample", "mypackage")
    testResult.Status = models.TestStatusPass
    testResult.Duration = 150 * time.Millisecond
    testResult.StartTime = time.Now().Add(-testResult.Duration)
    testResult.EndTime = time.Now()
    
    // Add test output
    testResult.Output = []string{
        "=== RUN   TestExample",
        "--- PASS: TestExample (0.15s)",
    }
    
    // Add coverage information
    testResult.Coverage = &models.TestCoverage{
        Percentage:        85.5,
        CoveredLines:      342,
        TotalLines:        400,
        CoveredStatements: 156,
        TotalStatements:   182,
    }
    
    // Check test status
    if testResult.IsSuccess() {
        fmt.Printf("Test %s passed in %v\n", testResult.Name, testResult.Duration)
    }
    
    // Add subtest
    subtest := models.NewTestResult("TestExample/SubTest", "mypackage")
    subtest.Status = models.TestStatusPass
    subtest.Parent = testResult.Name
    testResult.AddSubtest(subtest)
}
```

### Package Result Aggregation

```go
package main

import (
    "github.com/newbpydev/go-sentinel/pkg/models"
)

func main() {
    // Create package result
    packageResult := models.NewPackageResult("mypackage")
    packageResult.StartTime = time.Now()
    
    // Add individual test results
    test1 := models.NewTestResult("TestA", "mypackage")
    test1.Status = models.TestStatusPass
    test1.Duration = 100 * time.Millisecond
    packageResult.AddTest(test1)
    
    test2 := models.NewTestResult("TestB", "mypackage")
    test2.Status = models.TestStatusFail
    test2.Duration = 50 * time.Millisecond
    test2.Error = &models.TestError{
        Message:    "assertion failed",
        Type:       "AssertionError",
        SourceFile: "test.go",
        SourceLine: 42,
    }
    packageResult.AddTest(test2)
    
    // Finalize package result
    packageResult.EndTime = time.Now()
    packageResult.Duration = packageResult.EndTime.Sub(packageResult.StartTime)
    
    // Get statistics
    successRate := packageResult.GetSuccessRate()
    fmt.Printf("Package %s: %d/%d tests passed (%.1f%%)\n", 
        packageResult.Package, 
        packageResult.PassedCount, 
        packageResult.TestCount, 
        successRate*100)
}
```

### Test Summary and Statistics

```go
package main

import (
    "github.com/newbpydev/go-sentinel/pkg/models"
)

func main() {
    // Create test summary
    summary := models.NewTestSummary()
    summary.StartTime = time.Now()
    
    // Add package results
    pkg1 := models.NewPackageResult("package1")
    pkg1.TestCount = 5
    pkg1.PassedCount = 5
    pkg1.FailedCount = 0
    pkg1.Duration = 500 * time.Millisecond
    summary.AddPackageResult(pkg1)
    
    pkg2 := models.NewPackageResult("package2")
    pkg2.TestCount = 3
    pkg2.PassedCount = 2
    pkg2.FailedCount = 1
    pkg2.Duration = 300 * time.Millisecond
    summary.AddPackageResult(pkg2)
    
    // Finalize summary
    summary.EndTime = time.Now()
    summary.TotalDuration = summary.EndTime.Sub(summary.StartTime)
    
    // Display summary
    fmt.Printf("Test Summary:\n")
    fmt.Printf("  Total Tests: %d\n", summary.TotalTests)
    fmt.Printf("  Passed: %d\n", summary.PassedTests)
    fmt.Printf("  Failed: %d\n", summary.FailedTests)
    fmt.Printf("  Success Rate: %.1f%%\n", summary.GetSuccessRate()*100)
    fmt.Printf("  Total Duration: %v\n", summary.TotalDuration)
    fmt.Printf("  Average Duration: %v\n", summary.AverageDuration)
}
```

## 🔧 Core Data Models

### TestResult

Represents the result of a single test execution:

```go
type TestResult struct {
    ID        string                 // Unique test result identifier
    Name      string                 // Test name
    Package   string                 // Package containing the test
    Status    TestStatus             // Test execution status
    Duration  time.Duration          // Test execution time
    StartTime time.Time              // When the test started
    EndTime   time.Time              // When the test finished
    Output    []string               // Test output lines
    Error     *TestError             // Error details if failed
    Coverage  *TestCoverage          // Coverage information
    Subtests  []*TestResult          // Subtest results
    Parent    string                 // Parent test name (for subtests)
    Metadata  map[string]interface{} // Additional metadata
}
```

### PackageResult

Represents the result of testing an entire package:

```go
type PackageResult struct {
    Package      string                 // Package name/path
    Success      bool                   // All tests in package passed
    Duration     time.Duration          // Total execution time
    StartTime    time.Time              // When package testing started
    EndTime      time.Time              // When package testing finished
    Tests        []*TestResult          // Individual test results
    Coverage     *PackageCoverage       // Package coverage information
    TestCount    int                    // Total number of tests
    PassedCount  int                    // Number of passed tests
    FailedCount  int                    // Number of failed tests
    SkippedCount int                    // Number of skipped tests
    Output       string                 // Raw package output
    Error        error                  // Package-level error
    Metadata     map[string]interface{} // Additional metadata
}
```

### TestSummary

Aggregated test statistics across all packages:

```go
type TestSummary struct {
    TotalTests         int                    // Total number of tests executed
    PassedTests        int                    // Number of tests that passed
    FailedTests        int                    // Number of tests that failed
    SkippedTests       int                    // Number of tests that were skipped
    TotalDuration      time.Duration          // Total execution time
    AverageDuration    time.Duration          // Average test execution time
    PackageCount       int                    // Number of packages tested
    CoveragePercentage float64                // Overall coverage percentage
    Success            bool                   // All tests passed
    StartTime          time.Time              // When testing started
    EndTime            time.Time              // When testing finished
    FailedPackages     []string               // Names of packages with failed tests
    Metadata           map[string]interface{} // Additional metadata
}
```

## 📊 Coverage Models

### TestCoverage

Coverage information for individual tests:

```go
type TestCoverage struct {
    Percentage        float64                // Coverage percentage
    CoveredLines      int                    // Number of covered lines
    TotalLines        int                    // Total number of lines
    CoveredStatements int                    // Number of covered statements
    TotalStatements   int                    // Total number of statements
    Files             map[string]*FileCoverage // Per-file coverage
    Metadata          map[string]interface{} // Additional metadata
}
```

### PackageCoverage

Coverage information for entire packages:

```go
type PackageCoverage struct {
    Package           string                      // Package name
    Percentage        float64                     // Overall coverage percentage
    CoveredLines      int                         // Total covered lines
    TotalLines        int                         // Total lines in package
    CoveredStatements int                         // Total covered statements
    TotalStatements   int                         // Total statements in package
    Files             map[string]*FileCoverage    // Coverage for each file
    Functions         map[string]*FunctionCoverage // Coverage for each function
    Metadata          map[string]interface{}      // Additional metadata
}
```

### FileCoverage

Coverage information for individual files:

```go
type FileCoverage struct {
    FilePath          string                 // Path to the file
    Percentage        float64                // Coverage percentage for this file
    CoveredLines      int                    // Number of covered lines
    TotalLines        int                    // Total number of lines
    CoveredStatements int                    // Number of covered statements
    TotalStatements   int                    // Total number of statements
    LinesCovered      []int                  // Specific lines that are covered
    LinesUncovered    []int                  // Specific lines that are not covered
    Metadata          map[string]interface{} // Additional metadata
}
```

## 🚨 Error Models

### TestError

Detailed error information for failed tests:

```go
type TestError struct {
    Message          string                 // Primary error message
    Type             string                 // Error type or category
    StackTrace       []string               // Stack trace lines
    SourceFile       string                 // Source file where error occurred
    SourceLine       int                    // Line number where error occurred
    SourceColumn     int                    // Column number where error occurred
    SourceContext    []string               // Surrounding source code lines
    ContextStartLine int                    // Starting line number for context
    Expected         string                 // Expected value (for assertion errors)
    Actual           string                 // Actual value (for assertion errors)
    Metadata         map[string]interface{} // Additional error metadata
}
```

## 🔧 Configuration Models

### TestConfiguration

Configuration for test execution:

```go
type TestConfiguration struct {
    Packages         []string               // Packages to test
    TestFiles        []string               // Specific test files to run
    TestPatterns     []string               // Test name patterns to match
    Verbose          bool                   // Enable verbose output
    Coverage         bool                   // Enable coverage reporting
    CoverageProfile  string                 // Coverage profile file
    JSONOutput       bool                   // Enable JSON output format
    Parallel         int                    // Number of parallel test processes
    Timeout          time.Duration          // Test execution timeout
    Tags             []string               // Build tags to use
    Args             []string               // Additional arguments
    Environment      map[string]string      // Environment variables
    WorkingDirectory string                 // Working directory for execution
    Metadata         map[string]interface{} // Additional metadata
}
```

### WatchConfiguration

Configuration for watch mode:

```go
type WatchConfiguration struct {
    Enabled           bool                   // Enable watch mode
    Paths             []string               // Paths to watch
    IgnorePatterns    []string               // Patterns to ignore
    TestPatterns      []string               // Test file patterns
    DebounceInterval  time.Duration          // Debounce interval for changes
    RunOnStart        bool                   // Run tests on startup
    ClearOnRerun      bool                   // Clear terminal between runs
    NotifyOnSuccess   bool                   // Send notifications on success
    NotifyOnFailure   bool                   // Send notifications on failure
    Metadata          map[string]interface{} // Additional metadata
}
```

## 🔄 File Change Models

### FileChange

Represents a file system change:

```go
type FileChange struct {
    FilePath   string                 // Path to the changed file
    ChangeType ChangeType             // Type of change
    Timestamp  time.Time              // When the change occurred
    OldPath    string                 // Old path (for rename operations)
    Size       int64                  // File size after the change
    Checksum   string                 // File checksum after the change
    Metadata   map[string]interface{} // Additional change metadata
}
```

### ChangeType

Enumeration of file change types:

```go
type ChangeType string

const (
    ChangeTypeCreated  ChangeType = "created"
    ChangeTypeModified ChangeType = "modified"
    ChangeTypeDeleted  ChangeType = "deleted"
    ChangeTypeRenamed  ChangeType = "renamed"
)
```

## 🎯 Type Definitions

### TestStatus

Enumeration of test execution statuses:

```go
type TestStatus string

const (
    TestStatusPass    TestStatus = "PASS"
    TestStatusFail    TestStatus = "FAIL"
    TestStatusSkip    TestStatus = "SKIP"
    TestStatusPending TestStatus = "PENDING"
    TestStatusRunning TestStatus = "RUNNING"
)
```

## 🧪 Testing

The package achieves **100% test coverage** with comprehensive test suites:

### Running Tests

```bash
# Run all tests
go test ./pkg/models/...

# Run with coverage
go test ./pkg/models/... -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out
```

### Test Categories

- **Unit Tests**: Individual model validation and behavior
- **Integration Tests**: Model interaction and composition
- **Serialization Tests**: JSON marshaling and unmarshaling
- **Validation Tests**: Data validation and constraint checking
- **Performance Tests**: Memory usage and allocation efficiency

### Example Test Structure

```go
func TestTestResult_IsSuccess(t *testing.T) {
    t.Parallel()
    
    testCases := []struct {
        name     string
        status   models.TestStatus
        expected bool
    }{
        {"Pass status", models.TestStatusPass, true},
        {"Fail status", models.TestStatusFail, false},
        {"Skip status", models.TestStatusSkip, false},
        {"Pending status", models.TestStatusPending, false},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := models.NewTestResult("test", "pkg")
            result.Status = tc.status
            
            assert.Equal(t, tc.expected, result.IsSuccess())
        })
    }
}

func TestPackageResult_GetSuccessRate(t *testing.T) {
    t.Parallel()
    
    pkg := models.NewPackageResult("testpkg")
    pkg.TestCount = 10
    pkg.PassedCount = 8
    pkg.FailedCount = 2
    
    successRate := pkg.GetSuccessRate()
    assert.Equal(t, 0.8, successRate)
}
```

## 📊 Performance

The package is optimized for performance:

- **Memory Efficient**: Minimal memory allocation for model objects
- **Fast Serialization**: Optimized JSON marshaling/unmarshaling
- **Efficient Aggregation**: Fast statistics calculation
- **Lazy Loading**: Computed properties calculated on demand

### Benchmarks

```bash
# Run performance benchmarks
go test ./pkg/models/... -bench=.

# Example results:
BenchmarkTestResult_Creation-8         2000000    0.8μs/op    128B/op
BenchmarkPackageResult_AddTest-8       1000000    1.2μs/op     64B/op
BenchmarkTestSummary_AddPackage-8       500000    2.1μs/op    256B/op
BenchmarkJSON_Marshal-8                 300000    4.5μs/op    512B/op
```

## 🔍 Validation and Helper Methods

### TestResult Methods

```go
// Status checking methods
func (tr *TestResult) IsSuccess() bool
func (tr *TestResult) IsFailure() bool
func (tr *TestResult) IsComplete() bool

// Subtest management
func (tr *TestResult) AddSubtest(subtest *TestResult)

// Validation
func (tr *TestResult) Validate() error
```

### PackageResult Methods

```go
// Statistics calculation
func (pr *PackageResult) GetSuccessRate() float64

// Test management
func (pr *PackageResult) AddTest(test *TestResult)

// Validation
func (pr *PackageResult) Validate() error
```

### TestSummary Methods

```go
// Statistics calculation
func (ts *TestSummary) GetSuccessRate() float64

// Package management
func (ts *TestSummary) AddPackageResult(pkg *PackageResult)

// Validation
func (ts *TestSummary) Validate() error
```

## 🎯 Best Practices

### Model Creation

```go
// Use factory functions for consistent initialization
testResult := models.NewTestResult("TestName", "package")
packageResult := models.NewPackageResult("package")
summary := models.NewTestSummary()

// Set required fields immediately
testResult.Status = models.TestStatusPass
testResult.Duration = 100 * time.Millisecond
```

### Error Handling

```go
// Create rich error information
testError := &models.TestError{
    Message:      "assertion failed: expected 5, got 3",
    Type:         "AssertionError",
    SourceFile:   "calculator_test.go",
    SourceLine:   42,
    SourceColumn: 15,
    Expected:     "5",
    Actual:       "3",
    StackTrace:   []string{
        "calculator_test.go:42",
        "testing.go:123",
    },
}

testResult.Error = testError
```

### Metadata Usage

```go
// Add custom metadata for extensibility
testResult.Metadata = map[string]interface{}{
    "tags":        []string{"unit", "fast"},
    "author":      "developer@example.com",
    "complexity":  "low",
    "retry_count": 0,
}

// Access metadata safely
if tags, ok := testResult.Metadata["tags"].([]string); ok {
    fmt.Printf("Test tags: %v\n", tags)
}
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](../../CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/newbpydev/go-sentinel.git

# Navigate to the models package
cd go-sentinel/pkg/models

# Run tests
go test ./...

# Run tests with coverage
go test ./... -coverprofile=coverage.out

# View coverage
go tool cover -html=coverage.out
```

### Code Quality Standards

- **Test Coverage**: Maintain 100% test coverage
- **Documentation**: All exported symbols must have documentation
- **Linting**: Code must pass `golangci-lint` checks
- **Formatting**: Use `go fmt` for consistent formatting

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../../LICENSE) file for details.

## 🔗 Related Packages

- [`pkg/events`](../events/README.md) - Event system interfaces and types
- [`internal/app`](../../internal/app/README.md) - Application orchestration layer
- [`internal/config`](../../internal/config/README.md) - Configuration management

---

**Package Version**: v1.0.0  
**Go Version**: 1.21+  
**Last Updated**: January 2025  
**Maintainer**: Go Sentinel CLI Team 