# Commands Reference

Complete reference for Go Sentinel CLI commands and features.

## Main Commands

### `go-sentinel run`

Run tests with beautiful, Vitest-style output.

```bash
go-sentinel run [flags] [packages]
```

**Arguments:**
- `packages` - Go packages to test (default: current directory)

**Flags:**

#### Output Control
- `-c, --color` - Use colored output (default: true)
- `--no-color` - Disable colored output
- `-v, --verbose` - Enable verbose output
- `-vv, -vvv` - Increase verbosity level (can be repeated)

#### Test Execution
- `-w, --watch` - Enable watch mode for file changes
- `-t, --test string` - Run only tests matching pattern
- `-f, --fail-fast` - Stop on first test failure
- `-j, --parallel int` - Number of tests to run in parallel
- `--timeout duration` - Timeout for test execution

#### Examples

```bash
# Basic usage
go-sentinel run

# Run with watch mode
go-sentinel run --watch

# Run specific package with verbose output
go-sentinel run -v ./pkg/utils

# Run tests matching pattern
go-sentinel run --test="TestConfig*" ./...

# High performance with parallel execution
go-sentinel run --parallel=8 --timeout=5m ./...

# CI/CD friendly (no colors, fail fast)
go-sentinel run --no-color --fail-fast ./...

# Maximum verbosity for debugging
go-sentinel run -vvv --test="TestProblem*"
```

### `go-sentinel demo`

View interactive demonstrations of development phases.

```bash
go-sentinel demo --phase=<1-7>
```

**Flags:**
- `--phase` - Phase number to demonstrate (1-7)

**Available Phases:**
- `1` - Core Architecture & Data Structures
- `2` - Test Suite Display
- `3` - Failed Test Details Section
- `4` - Real-time Processing & Summary
- `5` - Watch Mode & Integration
- `6` - Performance & Error Handling
- `7` - CLI Options & Configuration

**Examples:**

```bash
# View core architecture demo
go-sentinel demo --phase=1

# Experience watch mode functionality
go-sentinel demo --phase=5

# Try CLI options and configuration
go-sentinel demo --phase=7
```

## Features

### Watch Mode

Automatically runs tests when files change.

**Activation:**
```bash
go-sentinel run --watch
```

**Behavior:**
- Monitors file system for changes
- Runs only tests affected by changed files
- Debounces rapid changes to prevent excessive runs
- Optional terminal clearing between runs

**Configuration:**
```json
{
  "watch": {
    "enabled": true,
    "debounce": "100ms",
    "ignorePatterns": ["**/*.log", "**/*.tmp"],
    "clearOnRerun": true,
    "runOnStart": true
  }
}
```

### Test Filtering

Run specific subsets of tests.

**By Name Pattern:**
```bash
go-sentinel run --test="TestConfig*"
go-sentinel run --test="Test*Handler"
```

**By Package:**
```bash
go-sentinel run ./pkg/utils
go-sentinel run ./internal/...
```

**Examples:**
```bash
# Run all config-related tests
go-sentinel run --test="*Config*"

# Run only unit tests (by convention)
go-sentinel run --test="TestUnit*"

# Run integration tests
go-sentinel run --test="TestIntegration*"
```

### Parallel Execution

Run tests in parallel for better performance.

**Usage:**
```bash
go-sentinel run --parallel=4
```

**Configuration:**
```json
{
  "parallel": 4
}
```

**Recommendations:**
- **Development**: 2-4 parallel processes
- **CI/CD**: Number of CPU cores available
- **Performance Testing**: 8+ parallel processes
- **Debugging**: 1 (sequential execution)

### Output Customization

Control visual appearance and verbosity.

**Verbosity Levels:**
- `0` - Minimal output (default)
- `1` - Standard output (`-v`)
- `2` - Detailed output (`-vv`)
- `3` - Maximum output (`-vvv`)

**Color Themes:**
- `dark` - Dark terminal backgrounds
- `light` - Light terminal backgrounds  
- `auto` - Automatic detection

**Icon Styles:**
- `unicode` - Full Unicode symbols (✓, ✗, 🚀)
- `ascii` - ASCII-only symbols (+, -, >)
- `minimal` - Reduced symbols
- `none` - No icons

**Examples:**
```bash
# High verbosity with colors
go-sentinel run -vvv --color

# ASCII-only output for compatibility
go-sentinel run --config='{"visual":{"icons":"ascii"}}'

# Minimal output for scripts
go-sentinel run --no-color -q
```

### Error Handling

Comprehensive error reporting and recovery.

**Features:**
- Source code context for failed tests
- Line number highlighting
- Stack trace formatting
- Timeout detection
- Build error reporting

**Failed Test Display:**
```
────────────────────────────────────────────────────────────────────────────────
                                 Failed Tests 2
────────────────────────────────────────────────────────────────────────────────
FAIL github.com/myproject/pkg/utils > TestValidation

    validation_test.go:25
    Expected validation to pass but got error: invalid input
```

### Performance Features

Optimizations for large test suites.

**Features:**
- Lazy rendering for large outputs
- Memory leak prevention
- Concurrent test execution
- Efficient JSON parsing
- Progress indicators

**Benchmarks:**
- JSON parsing: ~147µs per operation
- Suite rendering: ~60µs per operation
- Memory usage: <1MB per 1000 tests

## Configuration

### File-based Configuration

Create `sentinel.config.json` in your project root:

```json
{
  "colors": true,
  "verbosity": 1,
  "parallel": 4,
  "timeout": "2m",
  "visual": {
    "colors": true,
    "icons": "unicode",
    "theme": "dark"
  },
  "paths": {
    "includePatterns": ["**/*.go"],
    "excludePatterns": ["vendor/**", ".git/**"]
  },
  "watch": {
    "enabled": false,
    "debounce": "100ms",
    "clearOnRerun": true,
    "runOnStart": true
  }
}
```

### Environment Variables

Override configuration with environment variables:

```bash
export SENTINEL_COLORS=false
export SENTINEL_PARALLEL=8
export SENTINEL_TIMEOUT=5m
```

### CLI Precedence

Configuration precedence (highest to lowest):
1. CLI flags
2. Environment variables
3. Configuration file
4. Default values

## Exit Codes

- `0` - All tests passed
- `1` - One or more tests failed
- `2` - Build or compilation errors
- `3` - Configuration or argument errors

## Integration

### CI/CD Integration

**GitHub Actions:**
```yaml
- name: Run tests with Go Sentinel
  run: go-sentinel run --no-color --fail-fast --parallel=4 ./...
```

**GitLab CI:**
```yaml
test:
  script:
    - go-sentinel run --no-color --timeout=10m ./...
```

**Jenkins:**
```groovy
sh 'go-sentinel run --no-color --parallel=$BUILD_NUMBER ./...'
```

### IDE Integration

**VS Code Task:**
```json
{
  "label": "Go Sentinel Watch",
  "type": "shell",
  "command": "go-sentinel run --watch",
  "group": "test"
}
```

**IntelliJ/GoLand External Tool:**
- Program: `go-sentinel`
- Arguments: `run --watch`
- Working Directory: `$ProjectFileDir$` 