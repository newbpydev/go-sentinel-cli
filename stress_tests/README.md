# Go Sentinel CLI - Stress Test Suite

## Purpose

This directory contains **intentional test scenarios** designed to stress test the go-sentinel CLI application. These tests are meant to **fail, hang, panic, or exhibit problematic behavior** to validate that the CLI handles edge cases gracefully.

⚠️ **WARNING**: These tests are designed to fail and exhibit problematic behavior. They are NOT indicative of broken functionality.

## Test Categories

### 1. Basic Failure Scenarios (`basic_failures_test.go`)

**Purpose**: Test fundamental failure modes and edge cases

**Test Cases**:
- `TestBasicFail` - Simple assertion failure
- `TestMixedSubtests` - Subtests with mixed pass/fail/skip results
- `TestPanic` - Test that panics due to index out of bounds
- `TestAssertionFailures` - Table-driven tests with deliberate failures
- `TestComplexErrors` - Tests with multi-line error messages
- `TestNilPointer` - Controlled nil pointer scenarios
- `TestWithCleanup` - Tests that fail after cleanup operations

**Expected Behavior**: 
- CLI should capture all failure types
- Error messages should be properly formatted
- Panics should be recovered and reported
- Cleanup functions should still execute

### 2. Extreme Scenarios (`extreme_scenarios_test.go`)

**Purpose**: Test CLI behavior under extreme conditions

**Test Cases**:
- `TestHangingTest` - Simulates deadlocked/hanging tests (requires `ENABLE_HANGING_TEST` env var)
- `TestVeryLongRunningTest` - Tests with extended execution time
- `TestMassiveOutput` - Tests generating excessive log output (100+ lines)
- `TestDeeplyNestedSubtests` - 4-level deep subtest nesting
- `TestTableDrivenMixed` - Large table-driven tests with mixed results
- `TestGoroutineLeaks` - Tests that spawn goroutines without cleanup
- `TestMutexDeadlock` - Potential deadlock scenarios (requires `ENABLE_DEADLOCK_TEST` env var)
- `TestMultiplePanics` - Multiple panic recovery scenarios
- `TestFileSystemOperations` - File operations that may fail
- `TestEnvironmentDependencies` - Tests dependent on environment variables

**Expected Behavior**:
- CLI should handle long-running tests gracefully
- Massive output should not crash the renderer
- Nested subtests should be properly formatted
- Deadlock detection should work (with timeouts)
- Goroutine leaks should be contained

## Running Stress Tests

### Safe Mode (Default)
```bash
go test ./stress_tests/... -v
```
Most dangerous tests are disabled by default.

### Full Stress Mode
```bash
# Enable hanging tests (use with caution)
ENABLE_HANGING_TEST=1 go test ./stress_tests/... -v -timeout 30s

# Enable deadlock tests
ENABLE_DEADLOCK_TEST=1 go test ./stress_tests/... -v -timeout 10s

# Both (not recommended)
ENABLE_HANGING_TEST=1 ENABLE_DEADLOCK_TEST=1 go test ./stress_tests/... -v -timeout 60s
```

### Short Mode (Skip long-running tests)
```bash
go test ./stress_tests/... -v -short
```

## CLI Testing Scenarios

### What the CLI Should Handle:

1. **Graceful Failure Reporting**
   - All failures should be captured and formatted
   - No crashes from test panics
   - Proper error message rendering

2. **Performance Under Load**
   - Handle tests with massive output
   - Manage memory efficiently with many failures
   - Responsive UI during long-running tests

3. **Edge Case Handling**
   - Deeply nested subtest reporting
   - Mixed pass/fail/skip scenarios
   - Timeout and cancellation handling

4. **Resource Management**
   - Proper cleanup after panics
   - Goroutine leak detection/prevention
   - File handle management

## Expected Output Patterns

When running with go-sentinel CLI, you should see:

### Color-coded Results
- 🔴 Failed tests clearly highlighted
- 🟡 Skipped tests appropriately marked
- 🟢 Passing tests (few in this suite)

### Structured Error Display
- Clear test hierarchy (nested subtests)
- Detailed error messages for failures
- Panic recovery with stack traces

### Summary Statistics
- Total tests run
- Pass/Fail/Skip counts
- Execution times
- Any performance warnings

## Performance Benchmarks

These tests also serve as benchmarks for CLI performance:

- **Rendering Speed**: How fast can the CLI render 100+ log lines?
- **Memory Usage**: Memory efficiency with large test suites
- **Responsiveness**: UI updates during long-running tests
- **Error Handling**: Performance impact of error processing

## Integration with CI/CD

### Recommended CI Usage:
```yaml
# In CI pipeline
- name: "Stress Test - Safe Mode"
  run: go test ./stress_tests/... -v -short -timeout 5m

# Optional full stress test (separate job)
- name: "Stress Test - Full Mode" 
  run: |
    ENABLE_DEADLOCK_TEST=1 go test ./stress_tests/... -v -timeout 10m
  # Don't enable ENABLE_HANGING_TEST in CI
```

## Contributing

When adding new stress tests:

1. **Document the Purpose**: What specific edge case does it test?
2. **Use Environment Guards**: Protect dangerous tests with env vars
3. **Set Appropriate Timeouts**: Don't hang CI indefinitely
4. **Expected Behavior**: Document what the CLI should do

## Troubleshooting

### Common Issues:

1. **Tests Hanging**: 
   - Check if `ENABLE_HANGING_TEST` is set
   - Use shorter timeouts: `-timeout 30s`

2. **Memory Issues**:
   - Run with `-short` to skip heavy tests
   - Monitor memory usage during massive output tests

3. **CI Failures**:
   - These tests are supposed to fail
   - Check that CLI handles failures gracefully, not that tests pass

## Related Documentation

- [Main README](../README.md) - CLI usage and features
- [Architecture](../ARCHITECTURE.md) - CLI design principles
- [Testing Guide](../docs/testing.md) - Normal test suite documentation 