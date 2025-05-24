# Go Sentinel CLI - API Documentation

This document provides comprehensive documentation for all exported symbols in the Go Sentinel CLI project.

## Table of Contents

- [Package Overview](#package-overview)
- [Models Package](#models-package)
- [Events Package](#events-package)
- [Usage Examples](#usage-examples)
- [Error Handling](#error-handling)
- [Event System](#event-system)

## Package Overview

The Go Sentinel CLI is organized into several key packages:

- `pkg/models` - Core data models and value objects
- `pkg/events` - Event system for inter-component communication
- `internal/app` - Application controller and coordination
- `internal/ui` - User interface components and rendering
- `internal/watch` - File watching and change detection
- `internal/cli` - Command-line interface implementation

## Models Package

The `models` package provides shared data structures for test results, error handling, file changes, and configuration.

### Core Types

#### TestResult

Represents the result of a test execution.

```go
type TestResult struct {
    ID       string
    Name     string
    Package  string
    Status   TestStatus
    Duration time.Duration
    // ... additional fields
}
```

**Example Usage:**

```go
// Create a new test result
result := models.NewTestResult("TestUserLogin", "github.com/example/auth")
result.Status = models.TestStatusPassed
result.Duration = 150 * time.Millisecond

// Check test status
if result.IsSuccess() {
    fmt.Printf("Test %s passed in %v\n", result.Name, result.Duration)
}
```

#### PackageResult

Represents the result of testing an entire package.

```go
type PackageResult struct {
    Package     string
    Success     bool
    Duration    time.Duration
    Tests       []*TestResult
    TestCount   int
    PassedCount int
    FailedCount int
    // ... additional fields
}
```

**Example Usage:**

```go
// Create a package result
pkg := models.NewPackageResult("github.com/example/auth")
pkg.AddTest(result)

// Get success rate
successRate := pkg.GetSuccessRate()
fmt.Printf("Package success rate: %.1f%%\n", successRate*100)
```

#### TestSummary

Contains aggregated test statistics across multiple packages.

```go
type TestSummary struct {
    TotalTests         int
    PassedTests        int
    FailedTests        int
    SkippedTests       int
    TotalDuration      time.Duration
    CoveragePercentage float64
    // ... additional fields
}
```

**Example Usage:**

```go
// Create a test summary
summary := models.NewTestSummary()
summary.AddPackageResult(pkg)

// Display overall results
fmt.Printf("Overall: %d/%d tests passed (%.1f%%)\n",
    summary.PassedTests, summary.TotalTests, summary.GetSuccessRate()*100)
```

### Error Handling

#### SentinelError

The base error type for all application errors with comprehensive context.

```go
type SentinelError struct {
    Type     ErrorType
    Severity ErrorSeverity
    Message  string
    Cause    error
    Context  ErrorContext
    Stack    []StackFrame
    UserSafe bool
}
```

**Example Usage:**

```go
// Create a configuration error
err := models.NewConfigError("invalid timeout value", true)

// Check error type
if models.IsErrorType(err, models.ErrorTypeConfig) {
    fmt.Println("Configuration error:", err.UserMessage())
}

// Wrap an existing error
wrappedErr := models.WrapError(originalErr, models.ErrorTypeFileSystem, 
    models.SeverityError, "failed to read config file")
```

#### Error Types

```go
const (
    ErrorTypeConfig        ErrorType = "CONFIG"
    ErrorTypeFileSystem    ErrorType = "FILESYSTEM"
    ErrorTypeTestExecution ErrorType = "TEST_EXECUTION"
    ErrorTypeWatch         ErrorType = "WATCH"
    ErrorTypeValidation    ErrorType = "VALIDATION"
    // ... additional types
)
```

#### Error Severities

```go
const (
    SeverityInfo     ErrorSeverity = "INFO"
    SeverityWarning  ErrorSeverity = "WARNING"
    SeverityError    ErrorSeverity = "ERROR"
    SeverityCritical ErrorSeverity = "CRITICAL"
)
```

### Configuration Types

#### TestConfiguration

Configuration for test execution.

```go
type TestConfiguration struct {
    Packages        []string
    Verbose         bool
    Coverage        bool
    Parallel        int
    Timeout         time.Duration
    Tags            []string
    Environment     map[string]string
    // ... additional fields
}
```

**Example Usage:**

```go
config := &models.TestConfiguration{
    Packages:     []string{"./internal/...", "./pkg/..."},
    Verbose:      true,
    Coverage:     true,
    Parallel:     4,
    Timeout:      5 * time.Minute,
    Environment:  map[string]string{"TEST_ENV": "development"},
}
```

#### WatchConfiguration

Configuration for watch mode.

```go
type WatchConfiguration struct {
    Enabled          bool
    Paths            []string
    IgnorePatterns   []string
    DebounceInterval time.Duration
    RunOnStart       bool
    ClearOnRerun     bool
    // ... additional fields
}
```

### File Change Tracking

#### FileChange

Represents a file system change event.

```go
type FileChange struct {
    FilePath   string
    ChangeType ChangeType
    Timestamp  time.Time
    OldPath    string
    Size       int64
    Checksum   string
}
```

**Example Usage:**

```go
// Create a file change event
change := models.NewFileChange("main.go", models.ChangeTypeModified)
change.Size = 2048

fmt.Printf("File %s was %s at %v\n", 
    change.FilePath, change.ChangeType, change.Timestamp)
```

#### Change Types

```go
const (
    ChangeTypeCreated  ChangeType = "created"
    ChangeTypeModified ChangeType = "modified"
    ChangeTypeDeleted  ChangeType = "deleted"
    ChangeTypeRenamed  ChangeType = "renamed"
    ChangeTypeMoved    ChangeType = "moved"
)
```

### Coverage Information

#### TestCoverage

Coverage information for tests.

```go
type TestCoverage struct {
    Percentage        float64
    CoveredLines      int
    TotalLines        int
    CoveredStatements int
    TotalStatements   int
    Files             map[string]*FileCoverage
}
```

#### PackageCoverage

Coverage information for packages.

```go
type PackageCoverage struct {
    Package           string
    Percentage        float64
    CoveredLines      int
    TotalLines        int
    Files             map[string]*FileCoverage
    Functions         map[string]*FunctionCoverage
}
```

## Events Package

The `events` package provides a comprehensive event system for inter-component communication.

### Core Interfaces

#### EventBus

Central event publishing and subscription management.

```go
type EventBus interface {
    Publish(ctx context.Context, event Event) error
    PublishAsync(ctx context.Context, event Event) error
    Subscribe(eventType string, handler EventHandler) (Subscription, error)
    SubscribeWithFilter(filter EventFilter, handler EventHandler) (Subscription, error)
    Unsubscribe(subscription Subscription) error
    Close() error
    GetMetrics() *EventBusMetrics
}
```

#### Event

Core event interface.

```go
type Event interface {
    ID() string
    Type() string
    Timestamp() time.Time
    Source() string
    Data() interface{}
    Metadata() map[string]interface{}
    String() string
}
```

#### EventHandler

Interface for processing events.

```go
type EventHandler interface {
    Handle(ctx context.Context, event Event) error
    CanHandle(event Event) bool
    Priority() int
}
```

### Event Types

#### Test Events

```go
const (
    EventTypeTestStarted   = "test.started"
    EventTypeTestCompleted = "test.completed"
    EventTypeTestFailed    = "test.failed"
    EventTypeTestSkipped   = "test.skipped"
)
```

#### File Events

```go
const (
    EventTypeFileChanged     = "file.changed"
    EventTypeWatchStarted    = "watch.started"
    EventTypeWatchStopped    = "watch.stopped"
)
```

#### Application Events

```go
const (
    EventTypeAppStarted  = "app.started"
    EventTypeAppStopped  = "app.stopped"
    EventTypeAppError    = "app.error"
)
```

### Event Sources

```go
const (
    SourceTestRunner      = "test.runner"
    SourceFileWatcher     = "file.watcher"
    SourceAppController   = "app.controller"
    SourceConfig          = "config"
)
```

### Concrete Event Types

#### TestStartedEvent

```go
type TestStartedEvent struct {
    *BaseEvent
    TestName    string
    PackageName string
}
```

**Example Usage:**

```go
event := events.NewTestStartedEvent("TestUserAuth", "github.com/example/auth")
fmt.Printf("Test started: %s in %s\n", event.TestName, event.PackageName)
```

#### TestCompletedEvent

```go
type TestCompletedEvent struct {
    *BaseEvent
    TestName    string
    PackageName string
    Duration    time.Duration
    Success     bool
}
```

#### FileChangedEvent

```go
type FileChangedEvent struct {
    *BaseEvent
    FilePath   string
    ChangeType string
}
```

**Example Usage:**

```go
event := events.NewFileChangedEvent("/src/main.go", "modified")
fmt.Printf("File changed: %s (%s)\n", event.FilePath, event.ChangeType)
```

### Event Querying

#### EventQuery

Criteria for querying stored events.

```go
type EventQuery struct {
    EventTypes []string
    Sources    []string
    StartTime  *time.Time
    EndTime    *time.Time
    Limit      int
    Offset     int
    OrderBy    string
    OrderDesc  bool
    Metadata   map[string]interface{}
}
```

**Example Usage:**

```go
query := &events.EventQuery{
    EventTypes: []string{"test.started", "test.completed"},
    StartTime:  &lastHour,
    Limit:      50,
    OrderDesc:  true,
}
```

### Metrics and Statistics

#### EventBusMetrics

```go
type EventBusMetrics struct {
    TotalEvents           int64
    TotalSubscriptions    int
    EventsPerSecond       float64
    AverageProcessingTime time.Duration
    ErrorCount            int64
    LastEventTime         time.Time
}
```

#### SubscriptionStats

```go
type SubscriptionStats struct {
    EventsReceived        int64
    EventsProcessed       int64
    ProcessingErrors      int64
    AverageProcessingTime time.Duration
    LastEventTime         time.Time
    CreatedAt             time.Time
}
```

## Usage Examples

### Basic Test Execution

```go
package main

import (
    "fmt"
    "time"
    "github.com/newbpydev/go-sentinel/pkg/models"
)

func main() {
    // Create test configuration
    config := &models.TestConfiguration{
        Packages: []string{"./..."},
        Verbose:  true,
        Coverage: true,
        Parallel: 4,
    }
    
    // Create test results
    result := models.NewTestResult("TestExample", "example/pkg")
    result.Status = models.TestStatusPassed
    result.Duration = 100 * time.Millisecond
    
    // Create package result
    pkg := models.NewPackageResult("example/pkg")
    pkg.AddTest(result)
    
    // Create summary
    summary := models.NewTestSummary()
    summary.AddPackageResult(pkg)
    
    fmt.Printf("Tests: %d passed, %d failed\n", 
        summary.PassedTests, summary.FailedTests)
}
```

### Error Handling

```go
package main

import (
    "fmt"
    "github.com/newbpydev/go-sentinel/pkg/models"
)

func main() {
    // Create and handle errors
    err := models.NewValidationError("timeout", "must be positive")
    
    if models.IsErrorType(err, models.ErrorTypeValidation) {
        fmt.Printf("Validation error: %s\n", err.UserMessage())
    }
    
    // Sanitize errors for user display
    sanitized := models.SanitizeError(err)
    fmt.Printf("User-safe: %s\n", sanitized.Error())
}
```

### Event System

```go
package main

import (
    "fmt"
    "time"
    "github.com/newbpydev/go-sentinel/pkg/events"
)

func main() {
    // Create events
    testStarted := events.NewTestStartedEvent("TestAuth", "auth/pkg")
    testCompleted := events.NewTestCompletedEvent("TestAuth", "auth/pkg", 
        150*time.Millisecond, true)
    
    fmt.Printf("Test lifecycle:\n")
    fmt.Printf("1. %s\n", testStarted.String())
    fmt.Printf("2. %s (success: %t)\n", testCompleted.String(), testCompleted.Success)
}
```

### File Watching

```go
package main

import (
    "fmt"
    "github.com/newbpydev/go-sentinel/pkg/models"
    "github.com/newbpydev/go-sentinel/pkg/events"
)

func main() {
    // Configure watch mode
    watchConfig := &models.WatchConfiguration{
        Enabled:          true,
        Paths:            []string{"./src", "./tests"},
        IgnorePatterns:   []string{"*.tmp", ".git/"},
        DebounceInterval: 500 * time.Millisecond,
        RunOnStart:      true,
    }
    
    // Handle file changes
    change := models.NewFileChange("src/main.go", models.ChangeTypeModified)
    event := events.NewFileChangedEvent(change.FilePath, string(change.ChangeType))
    
    fmt.Printf("File change detected: %s\n", event.String())
}
```

## Error Handling

The Go Sentinel CLI uses a comprehensive error handling system based on the `SentinelError` type:

### Error Creation

```go
// Configuration errors
err := models.NewConfigError("invalid timeout", true)

// Validation errors (always user-safe)
err := models.NewValidationError("email", "invalid format")

// File system errors
err := models.NewFileSystemError("read", "/path/to/file", originalErr)

// Test execution errors
err := models.NewTestExecutionError("TestExample", originalErr)
```

### Error Checking

```go
// Check error type
if models.IsErrorType(err, models.ErrorTypeConfig) {
    // Handle configuration error
}

// Check error severity
if models.IsErrorSeverity(err, models.SeverityCritical) {
    // Handle critical error
}

// Get error context
if context := models.GetErrorContext(err); context != nil {
    fmt.Printf("Operation: %s, Resource: %s\n", 
        context.Operation, context.Resource)
}
```

### Error Sanitization

```go
// Sanitize errors for user display
sanitized := models.SanitizeError(err)
fmt.Printf("User message: %s\n", sanitized.Error())
```

## Event System

The event system enables decoupled communication between components:

### Publishing Events

```go
// Create and publish events
event := events.NewTestStartedEvent("TestExample", "example/pkg")
err := bus.Publish(ctx, event)
```

### Subscribing to Events

```go
// Subscribe to specific event types
subscription, err := bus.Subscribe("test.completed", handler)
if err != nil {
    return err
}
defer subscription.Cancel()
```

### Event Filtering

```go
// Subscribe with custom filters
filter := &CustomEventFilter{
    PackagePattern: "github.com/example/*",
}
subscription, err := bus.SubscribeWithFilter(filter, handler)
```

### Event Storage and Querying

```go
// Query events
query := &events.EventQuery{
    EventTypes: []string{"test.completed"},
    StartTime:  &yesterday,
    Limit:      100,
}
events, err := store.Retrieve(ctx, query)
```

This API documentation provides comprehensive coverage of all exported symbols with practical examples for effective usage of the Go Sentinel CLI components. 