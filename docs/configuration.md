# Go Sentinel CLI Configuration Guide

This guide explains how to configure Go Sentinel CLI for optimal testing experience across different environments.

## Quick Start

Create a `sentinel.config.json` file in your project root:

```json
{
  "colors": true,
  "icons": "unicode",
  "verbosity": 1,
  "watchMode": true,
  "parallel": 4,
  "timeout": "30s"
}
```

## Configuration File Location

Go Sentinel looks for configuration in the following order:
1. Path specified by `--config` flag
2. `sentinel.config.json` in current directory
3. Default configuration (if no file found)

## Configuration Options

### Visual Settings

```json
{
  "colors": true,           // Enable colored output
  "icons": "unicode",       // Icon style: "unicode", "ascii", "minimal", "none"
  "theme": "dark"          // Theme: "dark", "light", "auto"
}
```

**Icons Options:**
- `unicode`: 🟢 ❌ ⚡ (modern terminals)
- `ascii`: [PASS] [FAIL] [RUN] (compatible)
- `minimal`: + - ~ (simple)
- `none`: no icons

### Execution Settings

```json
{
  "verbosity": 1,           // Verbosity level (0-5)
  "parallel": 4,            // Number of parallel test processes
  "timeout": "30s",         // Test timeout (e.g., "30s", "2m", "1h")
  "testCommand": "go test", // Custom test command
  "failFast": false         // Stop on first failure
}
```

### Watch Mode Settings

```json
{
  "watchMode": true,           // Enable watch mode by default
  "watchDebounce": "250ms",    // Delay before re-running tests
  "clearOnRerun": true,        // Clear terminal on each run
  "runOnStart": true           // Run tests immediately when starting
}
```

### Path Configuration

```json
{
  "includePatterns": [
    "./internal/...",
    "./pkg/...",
    "./cmd/..."
  ],
  "excludePatterns": [
    "./vendor",
    "./tmp",
    ".*_test.go"
  ],
  "watchIgnore": [
    "*.log",
    "*.tmp",
    ".git/*",
    "node_modules/*"
  ]
}
```

## CLI Arguments Override

CLI arguments always take precedence over configuration file settings:

```bash
# Configuration file has colors: true
go-sentinel --no-color test    # Colors disabled via CLI
```

## Environment-Specific Configurations

### Development Configuration

Optimized for local development with immediate feedback:

```json
{
  "colors": true,
  "icons": "unicode",
  "theme": "dark",
  "verbosity": 1,
  "parallel": 4,
  "watchMode": true,
  "watchDebounce": "250ms",
  "clearOnRerun": true,
  "runOnStart": true,
  "timeout": "30s"
}
```

### CI/CD Configuration

Optimized for continuous integration environments:

```json
{
  "colors": false,
  "icons": "ascii",
  "verbosity": 0,
  "parallel": 8,
  "watchMode": false,
  "timeout": "120s",
  "failFast": true
}
```

### Performance Testing Configuration

Optimized for benchmark and performance testing:

```json
{
  "colors": true,
  "icons": "minimal",
  "verbosity": 0,
  "parallel": 16,
  "watchMode": false,
  "timeout": "300s",
  "testCommand": "go test -benchmem -bench=."
}
```

### Debugging Configuration

Optimized for debugging failed tests:

```json
{
  "colors": true,
  "icons": "unicode",
  "verbosity": 3,
  "parallel": 1,
  "watchMode": true,
  "clearOnRerun": false,
  "timeout": "60s"
}
```

## Terminal Compatibility

Go Sentinel automatically adjusts to your terminal capabilities:

### Modern Terminals (VSCode, iTerm2, Windows Terminal)
- Full Unicode support
- 256 colors
- Rich icons and formatting

### Legacy Terminals (cmd.exe, basic terminals)
- ASCII-only icons
- Basic color support
- Simplified formatting

### CI/CD Environments
- No colors (unless explicitly enabled)
- ASCII icons only
- Minimal visual elements

## Configuration Validation

Go Sentinel validates your configuration and provides helpful error messages:

```bash
$ go-sentinel test
❌ Configuration Error: invalid timeout format "invalid"
   Expected format: "30s", "2m", "1h"
```

## Examples

### Basic Watch Mode
```bash
go-sentinel -w ./internal
```

### Verbose Testing with Pattern
```bash
go-sentinel -vvv --test=TestUnit ./pkg
```

### Production CI Run
```bash
go-sentinel --no-color --parallel=8 --timeout=120s --fail-fast ./...
```

### Custom Configuration File
```bash
go-sentinel --config=./configs/ci.json test
```

## Migration from Legacy Config

Old flat configuration structure is automatically migrated:

```json
// Old format (still supported)
{
  "colors": true,
  "icons": "unicode",
  "verbosity": 2
}

// New nested format
{
  "colors": true,
  "visual": {
    "icons": "unicode",
    "colors": true
  },
  "verbosity": 2
}
```

## Troubleshooting

### Colors Not Working
1. Check terminal color support: `go-sentinel --help` should show colors
2. Verify `colors: true` in configuration
3. Ensure no `--no-color` flag is used

### Watch Mode Not Detecting Changes
1. Check `watchIgnore` patterns in configuration
2. Verify file permissions in watched directories
3. Increase `watchDebounce` for slower filesystems

### Performance Issues
1. Reduce `parallel` count for slower systems
2. Add exclusion patterns for large directories
3. Use `icons: "minimal"` or `"none"` for faster rendering

## Best Practices

1. **Use watch mode during development** for immediate feedback
2. **Disable colors in CI/CD** to avoid log pollution
3. **Adjust parallel count** based on system resources
4. **Use appropriate verbosity levels** (0 for CI, 1-2 for development, 3+ for debugging)
5. **Configure timeout values** based on test complexity
6. **Use exclusion patterns** to avoid watching unnecessary files

For more examples, run: `go-sentinel config-demo` 