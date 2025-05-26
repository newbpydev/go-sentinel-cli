# Testing Guidelines for Go Sentinel CLI

This document outlines the testing standards and organization for the Go Sentinel CLI project, following Go best practices and architecture principles.

## Test Organization Structure

### Package Naming Convention

We follow the **"underscore test" package** pattern recommended by the Go community:

```go
// ✅ CORRECT: Tests in separate package
package app_test

import (
    "testing"
    "github.com/newbpydev/go-sentinel/internal/app"
)

func TestNewTestExecutor_FactoryFunction(t *testing.T) {
    // Test the public API only
    executor := app.NewTestExecutor()
    // ...
}
```

```go
// ❌ WRONG: Tests in same package
package app

func TestNewTestExecutor_FactoryFunction(t *testing.T) {
    // This accesses internal implementation details
}
```

### Test Types and Locations

#### 1. Unit Tests (`*_test.go` files alongside source)
- **Location**: Same directory as source code
- **Package**: `package packagename_test`
- **Purpose**: Test individual components in isolation
- **Example**: `internal/app/app_test.go`

#### 2. Integration Tests (`tests/integration/`)
- **Location**: Separate `tests/integration/` directory
- **Package**: `package integration`
- **Purpose**: Test component interactions
- **Example**: `tests/integration/app_integration_test.go`

#### 3. End-to-End Tests (future)
- **Location**: `tests/e2e/`
- **Package**: `package e2e`
- **Purpose**: Test complete user workflows

### Test Naming Convention

Follow the `TestFunction_Scenario` pattern:

```go
// ✅ CORRECT: Clear function and scenario names
func TestNewTestExecutor_FactoryFunction(t *testing.T)
func TestTestExecutor_ExecuteSingle_ValidConfiguration(t *testing.T)
func TestTestExecutor_ExecuteSingle_NilConfiguration(t *testing.T)
func TestWatchCoordinator_Configure_ValidOptions(t *testing.T)

// ❌ WRONG: Unclear or generic names
func TestTestExecutor(t *testing.T)
func TestPhase1_ComprehensiveSuite(t *testing.T)
func testFactoryPattern(t *testing.T) // should be Test*, not test*
```

## Architecture Principles in Testing

### 1. Test Public APIs Only

Following the architecture principle of "consumer owns interface", we test the public APIs that consumers will use:

```go
// ✅ CORRECT: Testing public interface
func TestNewTestExecutor_FactoryFunction(t *testing.T) {
    t.Parallel()

    executor := app.NewTestExecutor()
    
    // Test that it implements the expected interface
    _, ok := executor.(app.TestExecutor)
    if !ok {
        t.Error("Should implement TestExecutor interface")
    }
}
```

### 2. Use Dependency Injection for Testing

Our adapter pattern enables easy testing through dependency injection:

```go
// ✅ CORRECT: Test configuration and behavior
func TestTestExecutor_ExecuteSingle_ValidConfiguration(t *testing.T) {
    t.Parallel()

    executor := app.NewTestExecutor()
    config := &app.Configuration{
        Verbosity: 1,
        Test: app.TestConfig{
            Timeout:  "30s",
            Parallel: 1,
        },
        // ... other config
    }

    err := executor.ExecuteSingle(ctx, []string{"."}, config)
    // Test the behavior, not the implementation
}
```

### 3. Parallel Tests by Default

All tests should use `t.Parallel()` unless they have dependencies that prevent it:

```go
func TestFunction_Scenario(t *testing.T) {
    t.Parallel() // Always add this first

    // Test implementation
}
```

## Unit vs Integration Test Guidelines

### Unit Tests: Test Logic, Not External Commands

Unit tests should focus on:
- **Input validation** (empty paths, nil configurations)
- **Error handling** (proper error messages, error conditions)
- **Configuration logic** (verbose flags, JSON output settings)
- **Utility functions** (file type detection, path parsing)
- **Interface compliance** (factory functions return correct types)

```go
// ✅ GOOD: Unit test testing validation logic
func TestBasicTestRunner_Run_EmptyPackages(t *testing.T) {
    r := runner.NewBasicTestRunner(false, false)
    result, err := r.Run(ctx, []string{})
    
    if err == nil {
        t.Error("Run should return error for empty packages")
    }
}
```

### Integration Tests: Test Real Command Execution

Integration tests should:
- **Create isolated environments** (temp directories with test modules)
- **Use timeouts** to prevent hanging
- **Test real command execution** in controlled environments
- **Be skippable** with `testing.Short()`

```go
// ✅ GOOD: Integration test with proper isolation
func TestBasicTestRunner_Integration_ActualExecution(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }
    
    tempDir := t.TempDir() // Isolated environment
    // Create test module in tempDir
    // Change to tempDir
    // Run with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
}
```

## Test Quality Standards

### 1. Given-When-Then Structure

Organize test logic clearly:

```go
func TestTestExecutor_ExecuteSingle_ValidConfiguration(t *testing.T) {
    t.Parallel()

    // Arrange (Given)
    executor := app.NewTestExecutor()
    config := &app.Configuration{...}
    ctx := context.Background()

    // Act (When)
    err := executor.ExecuteSingle(ctx, []string{"."}, config)

    // Assert (Then)
    if err != nil {
        // Handle expected test environment issues
        if strings.Contains(err.Error(), "not configured") {
            t.Errorf("Configuration issue: %v", err)
        } else {
            t.Logf("Expected failure in test environment: %v", err)
        }
    }
}
```

### 2. Error Testing

Test both success and error cases:

```go
func TestFunction_ErrorCase(t *testing.T) {
    t.Parallel()

    // Test nil input
    err := function(nil)
    if err == nil {
        t.Error("Should return error for nil input")
    }

    expectedMessage := "configuration cannot be nil"
    if !strings.Contains(err.Error(), expectedMessage) {
        t.Errorf("Expected error containing %q, got: %v", expectedMessage, err)
    }
}
```

### 3. Integration Test Patterns

Integration tests should skip when running in short mode:

```go
func TestAppIntegration_FullWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    t.Parallel()

    // Test component interactions
}
```

## Running Tests

### Unit Tests
```bash
# Run all unit tests
go test ./...

# Run specific package
go test ./internal/app

# Run with verbose output
go test ./internal/app -v

# Run with parallel execution
go test ./internal/app -parallel 4
```

### Integration Tests
```bash
# Run integration tests
go test ./tests/integration

# Skip integration tests (short mode)
go test ./... -short
```

### Coverage
```bash
# Run with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Anti-Patterns to Avoid

### ❌ Wrong Test Organization

```go
// DON'T: Mix different test types in same file
func TestUnit_Something(t *testing.T) {}
func TestIntegration_Something(t *testing.T) {}
func TestE2E_Something(t *testing.T) {}

// DON'T: Test internal implementation details
func TestInternalAdapter_PrivateMethod(t *testing.T) {}

// DON'T: Use same package name
package app // Should be app_test
```

### ❌ Executing External Commands in Unit Tests

```go
// DON'T: Execute actual go test commands in unit tests
func TestBasicTestRunner_Run_ValidPackage(t *testing.T) {
    r := runner.NewBasicTestRunner(false, false)
    result, err := r.Run(ctx, []string{"."}) // This can hang!
    // This executes actual go test which can:
    // 1. Hang indefinitely 
    // 2. Create circular test execution
    // 3. Have unpredictable test environment issues
}

// DO: Test the logic, not the external command execution
func TestBasicTestRunner_Run_NonExistentPath(t *testing.T) {
    r := runner.NewBasicTestRunner(false, false)
    result, err := r.Run(ctx, []string{"./non-existent-path"})
    // Test validation logic, error handling, configuration
    if err == nil {
        t.Error("Should return error for non-existent path")
    }
}
```

### ❌ Poor Test Naming

```go
// DON'T: Generic or unclear names
func TestBasicTest(t *testing.T) {}
func TestSomething(t *testing.T) {}
func TestPhase1(t *testing.T) {}

// DON'T: Helper functions with Test prefix
func TestHelper() {} // Should be helper() or testHelper()
```

### ❌ Package Pollution

```go
// DON'T: Create test files that test adapters directly
// internal/app/test_executor_adapter_test.go ❌
// internal/app/watch_coordinator_adapter_test.go ❌

// DO: Test public interfaces instead
// internal/app/app_test.go ✅
```

## References

This testing approach follows established Go best practices:

- [Ben Johnson's Testing Structure Guide](https://medium.com/@benbjohnson/structuring-tests-in-go-46ddee7a25c)
- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Architecture Principles](./architecture-principles.md)

## Architecture Compliance

Our testing approach maintains:

1. **Consumer owns interface**: Tests are written from the consumer perspective
2. **Package boundaries**: No direct internal package dependencies in tests
3. **Adapter pattern**: Tests work through public interfaces, not adapter internals
4. **Dependency injection**: Components can be easily tested through configuration
5. **Single responsibility**: Each test has a clear, focused purpose 