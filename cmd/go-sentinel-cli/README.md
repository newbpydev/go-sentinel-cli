# Go Sentinel CLI

A modern test runner for Go with beautiful, Vitest-style output.

## Features

- 🎨 Beautiful test output with colorized test results
- ⏱ Real-time test progress with animated spinners
- 🔍 Detailed test failure information with code context
- 🧠 Memory usage tracking
- 👀 Watch mode for continuous testing
- ⌨️ Interactive keyboard shortcuts for filtering and navigation

## Usage

```bash
# Run tests in the current directory and subdirectories
go-sentinel-cli test

# Run tests with watch mode (auto-rerun on file changes)
go-sentinel-cli test -w

# Run tests for specific packages
go-sentinel-cli test ./pkg/...

# Run only failed tests
go-sentinel-cli test -o

# Run tests and stop on first failure
go-sentinel-cli test -f
```

## Keyboard Shortcuts

When running tests, especially in watch mode, you can use these keyboard shortcuts:

- `f` - Clear all filters
- `p` - Toggle display of passing tests
- `x` - Toggle display of failing tests
- `s` - Toggle display of skipped tests
- `q` - Quit (in watch mode)

## CLI Output

The CLI displays test results in three main sections:

1. **Test Suite Information**
   - File paths with pass/fail status
   - Test counts and statistics
   - Duration and memory usage

2. **Failed Test Details** (shown when tests fail)
   - Error messages with type information
   - Source code context with line highlighting
   - File and line number information

3. **Summary**
   - Overall test run statistics
   - Timing information for various phases

## Installation

```bash
# Install directly
go install github.com/newbpydev/go-sentinel/cmd/go-sentinel-cli@latest

# Or clone the repository and build
git clone https://github.com/newbpydev/go-sentinel.git
cd go-sentinel
go build ./cmd/go-sentinel-cli
```

## Configuration

The CLI can be configured through command-line flags:

- `-c, --color`: Enable/disable colored output (default: true)
- `-w, --watch`: Enable watch mode (default: false)
- `-v, --verbose`: Enable verbose output (default: false)
- `-f, --fail-fast`: Stop on first failure (default: false)
- `-o, --only-failed`: Only run previously failed tests (default: false)

## Test Commands

The CLI offers several test command options:

1. **`test`** - The fully revamped CLI with beautiful output
   - May have stability issues in some environments
   - Shows complete test information with code context
   ```bash
   go-sentinel-cli test ./...
   ```

2. **`minimaltest`** - A simplified version of the revamped CLI
   - More stable with basic styling
   - Shows test results with colored output
   ```bash
   go-sentinel-cli minimaltest ./...
   ```

3. **`basictest`** - A basic test command with minimal output
   - Very stable, just shows basic pass/fail information
   ```bash
   go-sentinel-cli basictest ./...
   ```

4. **`simpletest`** - A direct wrapper for `go test`
   - The most stable option, uses standard Go test output
   ```bash
   go-sentinel-cli simpletest ./...
   ``` 