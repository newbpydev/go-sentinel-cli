# 🔄 Migration Guide: Legacy to Refactored Architecture

## 📋 Overview

This guide helps developers migrate from the old monolithic structure to the new modular architecture. It provides exact mappings of old code to new locations and usage patterns.

## 🗺️ File Migration Map

### Direct Migrations
| Old File | New Location | Notes |
|----------|--------------|-------|
| `optimized_test_runner.go` | `execution/runner.go` | Renamed to SmartTestRunner |
| `test_cache.go` | `execution/cache.go` | Enhanced with interfaces |
| `optimization_integration.go` | Split across modules | See breakdown below |
| `optimized_test_runner_test.go` | `testing/complexity/unit/runner_unit_test.go` | Better organized |

### `optimization_integration.go` Breakdown
| Old Functionality | New Location |
|-------------------|---------------|
| File change handling | `watch/analyzer.go` (planned) |
| Output rendering | `rendering/renderer.go` (planned) |
| Strategy coordination | `execution/strategy.go` |
| Configuration | `config/config.go` (planned) |

## 🔧 API Migration

### Test Runner Creation

#### Old Way
```go
// Before
runner := NewOptimizedTestRunner()
runner.SetOptimizationMode("aggressive")
```

#### New Way
```go
// After
cache := execution.NewInMemoryCacheManager(1000)
factory := execution.NewStrategyFactory()
strategy := factory.CreateStrategy("aggressive")
runner := execution.NewSmartTestRunner(cache, strategy)
```

### Running Tests

#### Old Way
```go
// Before
result, err := runner.RunOptimized(ctx, changes)
```

#### New Way
```go
// After
result, err := runner.RunTests(ctx, changes, strategy)
```

### Cache Management

#### Old Way
```go
// Before - mixed in TestResultCache
cache := NewTestResultCache()
cache.CacheResult("pkg/test", suite)
cached, exists := cache.GetCachedResult("pkg/test")
```

#### New Way
```go
// After - clean interface
cache := execution.NewInMemoryCacheManager(1000)
target := core.TestTarget{Path: "pkg/test", Type: "package"}
cache.StoreResult(target, result)
cached, exists := cache.GetCachedResult(target)
```

## 🎯 Import Path Changes

### Old Imports
```go
// These will no longer work
import (
    "github.com/newbpydev/go-sentinel/internal/cli"
)

// Direct struct access
runner := cli.OptimizedTestRunner{}
```

### New Imports
```go
// New modular imports
import (
    "github.com/newbpydev/go-sentinel/internal/cli/core"
    "github.com/newbpydev/go-sentinel/internal/cli/execution"
)

// Interface-based usage
var runner core.TestRunner = execution.NewSmartTestRunner(cache, strategy)
```

## 🧪 Testing Migration

### Old Test Structure
```go
// Before - monolithic tests
func TestOptimizedTestRunner(t *testing.T) {
    runner := NewOptimizedTestRunner()
    // Test everything together
}
```

### New Test Structure
```go
// After - organized by complexity
package unit // for simple tests

import (
    "github.com/newbpydev/go-sentinel/internal/cli/testing/helpers"
)

func TestSmartTestRunner_BasicFunctionality(t *testing.T) {
    cache := helpers.NewMockCacheManager()
    strategy := helpers.NewMockStrategy("test")
    runner := execution.NewSmartTestRunner(cache, strategy)
    // Test with mocked dependencies
}
```

## 🔌 Interface Usage Patterns

### Old Pattern (Concrete Dependencies)
```go
// Before - tightly coupled
type WatchMode struct {
    runner *OptimizedTestRunner // concrete type
    cache  *TestResultCache     // concrete type
}
```

### New Pattern (Interface Dependencies)
```go
// After - loosely coupled
type WatchMode struct {
    runner core.TestRunner      // interface
    cache  core.CacheManager    // interface
}

// Constructor with dependency injection
func NewWatchMode(runner core.TestRunner, cache core.CacheManager) *WatchMode {
    return &WatchMode{
        runner: runner,
        cache:  cache,
    }
}
```

## 🏗️ Configuration Migration

### Old Configuration
```go
// Before - mixed configuration
type Config struct {
    OptimizedMode bool
    CacheEnabled  bool
    Strategy      string
}
```

### New Configuration
```go
// After - structured configuration
type Config struct {
    // Execution settings
    UseCache        bool
    CacheStrategy   string
    MaxConcurrency  int
    
    // Watch mode settings
    WatchMode       bool
    DebounceInterval time.Duration
    
    // Output settings
    OutputFormat    string
    ShowProgress    bool
}
```

## 🔄 Error Handling Migration

### Old Error Handling
```go
// Before - generic errors
if err != nil {
    return fmt.Errorf("test execution failed: %v", err)
}
```

### New Error Handling
```go
// After - typed errors with context
if err != nil {
    var execErr *core.TestExecutionError
    if errors.As(err, &execErr) {
        log.Printf("Test failed for target %s after %v", 
                  execErr.Target.Path, execErr.Duration)
    }
    return err
}
```

## 📦 Dependency Injection Examples

### Before (Hard Dependencies)
```go
// Before - can't test easily
func processChanges(changes []FileChange) error {
    runner := NewOptimizedTestRunner() // hard dependency
    return runner.RunOptimized(context.Background(), changes)
}
```

### After (Injected Dependencies)
```go
// After - easily testable
func processChanges(runner core.TestRunner, changes []core.FileChange) error {
    strategy := execution.NewAggressiveStrategy()
    _, err := runner.RunTests(context.Background(), changes, strategy)
    return err
}

// In tests
func TestProcessChanges(t *testing.T) {
    mockRunner := helpers.NewMockTestRunner()
    err := processChanges(mockRunner, []core.FileChange{})
    // Test with controlled behavior
}
```

## 🎭 Mock Usage for Testing

### Creating Mocks
```go
// Create mocks for testing
cache := helpers.NewMockCacheManager()
strategy := helpers.NewMockStrategy("test-strategy")
runner := helpers.NewMockTestRunner()

// Configure mock behavior
mockCache := cache.(*helpers.MockCacheManager)
mockCache.SetShouldFail(false)

mockStrategy := strategy.(*helpers.MockStrategy)
mockStrategy.SetShouldRunTest(true)
```

### Verifying Mock Calls
```go
// Verify interactions
if !mockStrategy.WasCalled("ShouldRunTest") {
    t.Error("Expected ShouldRunTest to be called")
}

callCount := mockCache.GetCallCount("GetCachedResult")
if callCount != 2 {
    t.Errorf("Expected 2 cache calls, got %d", callCount)
}
```

## 🚨 Common Migration Issues

### Issue 1: Direct Struct Access
```go
// ❌ This will break
runner := &OptimizedTestRunner{}
runner.enableGoCache = true

// ✅ Use constructor and methods instead
runner := execution.NewSmartTestRunner(cache, strategy)
// Configuration through strategy pattern
```

### Issue 2: Mixed Concerns
```go
// ❌ Old mixed approach
func handleFileChange(path string) {
    // File watching + test execution + output rendering all mixed
}

// ✅ New separated concerns
func handleFileChange(path string) {
    change := analyzer.AnalyzeChange(path)      // watch module
    result := runner.RunTests(ctx, []*change)  // execution module
    renderer.RenderResults(result)             // rendering module
}
```

### Issue 3: Hard-coded Strategies
```go
// ❌ Hard-coded behavior
runner.SetOptimizationMode("aggressive")

// ✅ Strategy injection
factory := execution.NewStrategyFactory()
strategy := factory.CreateStrategy("aggressive")
runner := execution.NewSmartTestRunner(cache, strategy)
```

## 📈 Performance Migration Notes

### Cache Size Configuration
```go
// Before - fixed size
cache := NewTestResultCache()

// After - configurable size
cache := execution.NewInMemoryCacheManager(1000) // 1000 entries max
```

### Strategy Selection
```go
// Choose strategy based on use case
strategies := map[string]core.ExecutionStrategy{
    "development": execution.NewAggressiveStrategy(),    // 5-min cache
    "ci":         execution.NewNoCacheStrategy(),        // Always run
    "watch":      execution.NewWatchModeStrategy(),      // 2-min cache
}
```

## ✅ Migration Checklist

### Phase 1: Update Imports
- [ ] Change import paths to new modules
- [ ] Update type references
- [ ] Fix compilation errors

### Phase 2: Refactor API Usage
- [ ] Replace direct constructors with factory methods
- [ ] Use interface types instead of concrete types
- [ ] Implement dependency injection patterns

### Phase 3: Update Tests
- [ ] Move tests to appropriate complexity directories
- [ ] Use mock helpers for unit tests
- [ ] Add integration tests for component interaction

### Phase 4: Configuration
- [ ] Update configuration structures
- [ ] Implement new error handling patterns
- [ ] Add proper logging and metrics

### Phase 5: Validation
- [ ] Run full test suite
- [ ] Verify performance characteristics
- [ ] Check documentation updates

## 🆘 Getting Help

### Compilation Issues
1. Check import paths match new module structure
2. Verify interface implementations
3. Use `go mod tidy` to clean dependencies

### Runtime Issues
1. Check dependency injection setup
2. Verify mock configurations in tests
3. Review error handling patterns

### Performance Issues
1. Verify cache configuration
2. Check strategy selection
3. Review concurrency settings

## 📞 Support Resources

- **Architecture Documentation**: `internal/cli/ARCHITECTURE.md`
- **Module Documentation**: Each module's `README.md`
- **Test Examples**: `internal/cli/testing/complexity/unit/`
- **Mock Usage**: `internal/cli/testing/helpers/mocks.go`

This migration guide ensures smooth transition from the legacy codebase to the new modular architecture while maintaining all functionality and improving maintainability. 