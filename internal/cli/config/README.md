# Configuration Module

The configuration module provides a unified system for managing CLI arguments and file-based configuration in go-sentinel.

## Overview

This module combines two sources of configuration:
1. **CLI Arguments** - Command line flags and options
2. **Config Files** - JSON-based configuration files

CLI arguments take precedence over file configuration, allowing users to override file settings on a per-run basis.

## Components

### CLIParser
Handles parsing of command line arguments using Go's `flag` package.

```go
parser := config.NewCLIParser()
cliArgs, err := parser.Parse(os.Args[1:])
```

### ConfigLoader
Handles loading and parsing of JSON configuration files.

```go
loader := config.NewConfigLoader()
fileConfig, err := loader.LoadFromFile("sentinel.config.json")
```

### ConfigManager
Coordinates CLI parsing and file loading, merging them into a unified configuration.

```go
manager := config.NewConfigManager()
configuration, err := manager.LoadConfiguration(os.Args[1:])
```

## Supported CLI Flags

### Core Flags
- `-w, --watch` - Enable watch mode
- `-v, --verbose` - Enable verbose output
- `--color` / `--no-color` - Control colored output
- `--fail-fast` - Stop on first failure
- `--optimized` - Enable optimized test execution

### Advanced Flags
- `--verbosity=N` - Set verbosity level (0-5)
- `--parallel=N` - Number of parallel test executions
- `-t, --test=PATTERN` - Run only tests matching pattern
- `--config=PATH` - Path to configuration file
- `--timeout=DURATION` - Test timeout duration
- `--optimization=MODE` - Optimization mode (conservative, balanced, aggressive)

### Multiple Verbosity
The parser supports multiple `-v` flags for increased verbosity:
- `-v` = verbosity level 1
- `-vv` = verbosity level 2
- `-vvv` = verbosity level 3

## Configuration File Format

Configuration files use JSON format:

```json
{
  "colors": true,
  "icons": "unicode",
  "watchMode": false,
  "verbosity": 1,
  "timeout": "30s",
  "watchDebounce": "300ms",
  "parallel": 4,
  "testCommand": "go test",
  "includePatterns": ["**/*_test.go"],
  "excludePatterns": ["vendor/**"],
  "watchIgnore": [".git/**", "node_modules/**"]
}
```

### Supported Fields
- `colors` (bool) - Enable colored output
- `icons` (string) - Icon set: "unicode", "ascii", "minimal", "none"
- `watchMode` (bool) - Default watch mode setting
- `verbosity` (int) - Verbosity level (0-5)
- `timeout` (string) - Test timeout duration
- `watchDebounce` (string) - Watch debounce interval
- `parallel` (int) - Number of parallel executions
- `includePatterns` ([]string) - File patterns to include
- `excludePatterns` ([]string) - File patterns to exclude
- `watchIgnore` ([]string) - Patterns to ignore in watch mode

## Configuration Precedence

1. **CLI Arguments** (highest priority)
2. **Configuration File**
3. **Default Values** (lowest priority)

Example:
```bash
# File has "verbosity": 2, CLI overrides with -v
go-sentinel -v --config=custom.json
# Result: verbosity = 1 (from CLI -v flag)
```

## Integration with Core Types

The configuration module integrates seamlessly with `internal/cli/core` types:

```go
// Get core configuration for use with other modules
coreConfig := configManager.GetCoreConfig()

// Pass to controller
controller := controller.NewAppController(coreConfig)
```

## Error Handling

The configuration system provides detailed error messages:

```go
config, err := manager.LoadConfiguration(args)
if err != nil {
    // Errors include context about what failed:
    // - "failed to parse CLI arguments: invalid verbosity level"
    // - "failed to load config file: file not found"
    // - "invalid configuration: timeout must be positive"
}
```

## Thread Safety

The `ConfigManager` stores configuration state and should be created once per application instance. Individual parsers and loaders are stateless and can be used concurrently.

## Migration from Legacy

This module replaces the legacy `cli_args.go` and `config.go` files while maintaining full compatibility with existing CLI interfaces and configuration file formats. 