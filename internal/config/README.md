# 📋 Config Package

[![Test Coverage](https://img.shields.io/badge/coverage-100.0%25-brightgreen.svg)](https://github.com/newbpydev/go-sentinel/tree/main/internal/config)
[![Go Report Card](https://goreportcard.com/badge/github.com/newbpydev/go-sentinel/internal/config)](https://goreportcard.com/report/github.com/newbpydev/go-sentinel/internal/config)
[![Go Reference](https://pkg.go.dev/badge/github.com/newbpydev/go-sentinel/internal/config.svg)](https://pkg.go.dev/github.com/newbpydev/go-sentinel/internal/config)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## 📖 Overview

The `config` package provides comprehensive application configuration management for the Go Sentinel CLI. It handles configuration loading from files, command-line argument parsing, and configuration validation with support for multiple sources and precedence rules.

### 🎯 Key Features

- **Multi-source Configuration**: Load from files, CLI arguments, and defaults
- **Configuration Precedence**: CLI args override file config, file config overrides defaults
- **Validation Modes**: Strict, lenient, and disabled validation options
- **Type Safety**: Strongly typed configuration structures with validation
- **Factory Pattern**: Clean dependency injection and object creation
- **Adapter Pattern**: Seamless integration with app package interfaces

## 🏗️ Architecture

This package follows clean architecture principles:

- **Single Responsibility**: Focuses only on configuration management
- **Dependency Inversion**: App package depends on config interfaces, not implementations
- **Interface Segregation**: Small, focused interfaces for specific concerns
- **Factory Pattern**: Clean object creation with dependency injection

### 📦 Package Structure

```
internal/config/
├── app_config_loader.go          # Main configuration loader implementation
├── app_config_loader_interface.go # Configuration loader interfaces
├── app_arg_parser.go             # Command-line argument parser
├── app_arg_parser_interface.go   # Argument parser interfaces
├── args.go                       # CLI argument definitions
├── loader.go                     # Core configuration loading logic
├── compat.go                     # Compatibility utilities
└── *_test.go                     # Comprehensive test suite (100% coverage)
```

## 🚀 Quick Start

### Basic Configuration Loading

```go
package main

import (
    "github.com/newbpydev/go-sentinel/internal/config"
)

func main() {
    // Create factory with default dependencies
    factory := config.NewAppConfigLoaderFactory()
    loader := factory.CreateDefault()
    
    // Load configuration from file
    cfg, err := loader.LoadFromFile("config.yaml")
    if err != nil {
        // Handle error
    }
    
    // Validate configuration
    if err := loader.Validate(cfg); err != nil {
        // Handle validation error
    }
}
```

### Command-Line Argument Parsing

```go
package main

import (
    "os"
    "github.com/newbpydev/go-sentinel/internal/config"
)

func main() {
    // Create argument parser factory
    factory := config.NewAppArgParserFactory()
    parser := factory.CreateDefault()
    
    // Parse command-line arguments
    args, err := parser.Parse(os.Args[1:])
    if err != nil {
        // Handle parsing error
    }
    
    // Use parsed arguments
    fmt.Printf("Packages: %v\n", args.Packages)
    fmt.Printf("Watch mode: %v\n", args.Watch)
}
```

### Configuration Merging

```go
package main

import (
    "github.com/newbpydev/go-sentinel/internal/config"
)

func main() {
    factory := config.NewAppConfigLoaderFactory()
    loader := factory.CreateDefault()
    
    // Load base configuration
    cfg := loader.LoadFromDefaults()
    
    // Parse CLI arguments
    argFactory := config.NewAppArgParserFactory()
    argParser := argFactory.CreateDefault()
    args, _ := argParser.Parse(os.Args[1:])
    
    // Merge CLI args with configuration (CLI takes precedence)
    finalCfg := loader.Merge(cfg, args)
    
    // Validate final configuration
    if err := loader.Validate(finalCfg); err != nil {
        log.Fatal("Configuration validation failed:", err)
    }
}
```

## 🔧 Configuration Structure

### AppConfig

The main configuration structure supporting all application settings:

```go
type AppConfig struct {
    Watch    AppWatchConfig    // Watch mode configuration
    Paths    AppPathsConfig    // Path patterns and filters
    Visual   AppVisualConfig   // UI and display settings
    Test     AppTestConfig     // Test execution settings
    Colors   bool              // Enable colored output
    Verbosity int              // Logging verbosity level
}
```

### Watch Configuration

```go
type AppWatchConfig struct {
    Enabled        bool     // Enable watch mode
    IgnorePatterns []string // Patterns to ignore
    Debounce       string   // Debounce interval (e.g., "500ms")
    RunOnStart     bool     // Run tests on startup
    ClearOnRerun   bool     // Clear terminal between runs
}
```

### Path Configuration

```go
type AppPathsConfig struct {
    IncludePatterns []string // Patterns to include
    ExcludePatterns []string // Patterns to exclude
}
```

### Visual Configuration

```go
type AppVisualConfig struct {
    Icons         string // Icon set to use
    Theme         string // Color theme
    TerminalWidth int    // Terminal width for formatting
}
```

### Test Configuration

```go
type AppTestConfig struct {
    Timeout  string // Test timeout (e.g., "30s")
    Parallel int    // Number of parallel test processes
    Coverage bool   // Enable coverage reporting
}
```

## 🎛️ Validation Modes

The package supports three validation modes:

### Strict Mode (Default)
```go
dependencies := config.AppConfigLoaderDependencies{
    ValidationMode: config.ValidationModeStrict,
}
loader := factory.Create(dependencies)
```

- Enforces all validation rules
- Fails on any configuration inconsistency
- Recommended for production use

### Lenient Mode
```go
dependencies := config.AppConfigLoaderDependencies{
    ValidationMode: config.ValidationModeLenient,
}
loader := factory.Create(dependencies)
```

- Allows some validation relaxation
- Warns on minor issues but continues
- Useful for development environments

### Disabled Mode
```go
dependencies := config.AppConfigLoaderDependencies{
    ValidationMode: config.ValidationModeOff,
}
loader := factory.Create(dependencies)
```

- Disables validation entirely
- Used primarily for testing
- Not recommended for production

## 🔄 Integration Patterns

### Factory Pattern Usage

```go
// Create factory
factory := config.NewAppConfigLoaderFactory()

// Create with custom dependencies
dependencies := config.AppConfigLoaderDependencies{
    CliLoader:      customLoader,
    ValidationMode: config.ValidationModeStrict,
}
loader := factory.Create(dependencies)

// Or use defaults
loader := factory.CreateDefault()
```

### Adapter Pattern Integration

The package integrates with the app package through adapter patterns:

```go
// App package uses config through adapters
type ConfigLoaderAdapter struct {
    factory *config.AppConfigLoaderFactory
    loader  config.AppConfigLoader
}

func (a *ConfigLoaderAdapter) LoadConfiguration(path string) (*app.Configuration, error) {
    // Delegate to config package
    appConfig, err := a.loader.LoadFromFile(path)
    if err != nil {
        return nil, err
    }
    
    // Convert to app package types
    return a.factory.ConvertToAppConfiguration(appConfig), nil
}
```

## 🧪 Testing

The package achieves **100% test coverage** with comprehensive test suites:

### Running Tests

```bash
# Run all tests
go test ./internal/config/...

# Run with coverage
go test ./internal/config/... -coverprofile=coverage.out

# View coverage report
go tool cover -html=coverage.out
```

### Test Categories

- **Unit Tests**: Individual component testing
- **Integration Tests**: Multi-component workflows
- **Comprehensive Tests**: Edge cases and error conditions
- **Compatibility Tests**: Backward compatibility validation

### Example Test Structure

```go
func TestAppConfigLoader_LoadFromFile_Success(t *testing.T) {
    t.Parallel()
    
    factory := NewAppConfigLoaderFactory()
    loader := factory.CreateDefault()
    
    // Test successful file loading
    config, err := loader.LoadFromFile("testdata/valid-config.yaml")
    assert.NoError(t, err)
    assert.NotNil(t, config)
    assert.True(t, config.Watch.Enabled)
}
```

## 📊 Performance

The package is optimized for performance:

- **Lazy Loading**: Configuration loaded only when needed
- **Caching**: Parsed configurations cached for reuse
- **Minimal Allocations**: Efficient memory usage patterns
- **Fast Validation**: Optimized validation algorithms

### Benchmarks

```bash
# Run performance benchmarks
go test ./internal/config/... -bench=.

# Example results:
BenchmarkAppConfigLoader_LoadFromFile-8     1000    1.2ms/op    256B/op
BenchmarkAppArgParser_Parse-8               5000    0.3ms/op    128B/op
```

## 🔍 Error Handling

The package provides comprehensive error handling:

### Error Types

```go
// Configuration validation errors
type ValidationError struct {
    Field   string
    Value   interface{}
    Message string
}

// File loading errors
type LoadError struct {
    Path    string
    Cause   error
    Message string
}

// Argument parsing errors
type ParseError struct {
    Argument string
    Value    string
    Message  string
}
```

### Error Handling Examples

```go
config, err := loader.LoadFromFile("config.yaml")
if err != nil {
    switch e := err.(type) {
    case *config.LoadError:
        log.Printf("Failed to load config from %s: %v", e.Path, e.Cause)
    case *config.ValidationError:
        log.Printf("Invalid config field %s: %s", e.Field, e.Message)
    default:
        log.Printf("Unexpected error: %v", err)
    }
    return
}
```

## 🔧 Configuration Examples

### YAML Configuration File

```yaml
# config.yaml
watch:
  enabled: true
  ignorePatterns:
    - "*.tmp"
    - "node_modules/**"
  debounce: "500ms"
  runOnStart: true
  clearOnRerun: true

paths:
  includePatterns:
    - "./internal/**"
    - "./pkg/**"
  excludePatterns:
    - "./vendor/**"

visual:
  icons: "emoji"
  theme: "dark"
  terminalWidth: 120

test:
  timeout: "30s"
  parallel: 4
  coverage: true

colors: true
verbosity: 2
```

### JSON Configuration File

```json
{
  "watch": {
    "enabled": true,
    "ignorePatterns": ["*.tmp", "node_modules/**"],
    "debounce": "500ms",
    "runOnStart": true,
    "clearOnRerun": true
  },
  "paths": {
    "includePatterns": ["./internal/**", "./pkg/**"],
    "excludePatterns": ["./vendor/**"]
  },
  "visual": {
    "icons": "emoji",
    "theme": "dark",
    "terminalWidth": 120
  },
  "test": {
    "timeout": "30s",
    "parallel": 4,
    "coverage": true
  },
  "colors": true,
  "verbosity": 2
}
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](../../CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone the repository
git clone https://github.com/newbpydev/go-sentinel.git

# Navigate to the config package
cd go-sentinel/internal/config

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

- [`internal/app`](../app/README.md) - Application orchestration layer
- [`internal/lifecycle`](../lifecycle/README.md) - Application lifecycle management
- [`internal/events`](../events/README.md) - Event handling system
- [`pkg/models`](../../pkg/models/README.md) - Shared data models

---

**Package Version**: v1.0.0  
**Go Version**: 1.21+  
**Last Updated**: January 2025  
**Maintainer**: Go Sentinel CLI Team 