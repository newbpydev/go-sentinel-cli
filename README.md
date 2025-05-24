# 🚀 Go Sentinel CLI

> A modern, Vitest-inspired test runner for Go with beautiful terminal output

[![Build Status](https://github.com/newbpydev/go-sentinel-cli/actions/workflows/test.yml/badge.svg)](https://github.com/newbpydev/go-sentinel-cli/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/newbpydev/go-sentinel-cli)](https://goreportcard.com/report/github.com/newbpydev/go-sentinel-cli)
[![Go Reference](https://pkg.go.dev/badge/github.com/newbpydev/go-sentinel-cli)](https://pkg.go.dev/github.com/newbpydev/go-sentinel-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## ✨ Overview

Go Sentinel CLI transforms the standard `go test` experience into a modern, beautiful test runner with real-time feedback, smart file watching, and comprehensive reporting. Built for Go developers who want the productivity and aesthetic of modern JavaScript testing tools like Vitest, but native to the Go ecosystem.

**Key Features:**
- 🎨 **Beautiful Vitest-style Output** - Clean, colorful test results with intuitive indicators
- 👁️ **Smart Watch Mode** - Intelligent file monitoring with debounced re-runs
- ⚡ **Optimized Execution** - Leverages Go's built-in test caching for faster runs
- 📊 **Rich Test Reporting** - Detailed failure analysis with source code context
- 🎯 **Selective Testing** - Run specific tests, packages, or patterns
- 🔧 **Highly Configurable** - JSON config files and comprehensive CLI options

## 📦 Installation

### Prerequisites
- Go 1.23+ 
- Git

### Install via Go (Recommended)
```bash
go install github.com/newbpydev/go-sentinel-cli/cmd/go-sentinel-cli@latest
```

### Build from Source
```bash
git clone https://github.com/newbpydev/go-sentinel-cli.git
cd go-sentinel-cli
go build -o go-sentinel-cli ./cmd/go-sentinel-cli
```

### Download Pre-built Binaries
Visit the [GitHub Releases](https://github.com/newbpydev/go-sentinel-cli/releases) page.

## 🚦 Quick Start

### Basic Usage
```bash
# Run tests with beautiful output
go-sentinel run

# Run tests in watch mode
go-sentinel run --watch

# Run specific package
go-sentinel run ./internal/cli

# Run tests matching pattern
go-sentinel run --test="TestConfig*"
```

### Common Workflows
```bash
# Development workflow (watch mode)
go-sentinel run -w --color

# CI/CD workflow (no colors, fail fast)
go-sentinel run --no-color --fail-fast

# Debug mode (verbose output)
go-sentinel run -vvv

# Performance testing (parallel execution)
go-sentinel run --parallel=8
```

## 📋 CLI Reference

### Commands

#### `go-sentinel run [packages]`
Run tests with beautiful output and optional watch mode.

**Flags:**
```
  -c, --color              Enable colored output (default: true)
      --no-color           Disable colored output
  -v, --verbose            Increase verbosity (can be repeated: -v, -vv, -vvv)
  -w, --watch              Enable watch mode for file changes
  -t, --test string        Run only tests matching pattern
  -f, --fail-fast          Stop on first test failure
  -j, --parallel int       Number of parallel test processes (default: 4)
      --timeout duration   Test execution timeout (default: 10m)
      --optimized          Enable optimized mode with Go test caching
  -h, --help               Show help information
```

**Examples:**
```bash
# Watch mode with test filtering
go-sentinel run -w --test="TestHandler*" ./api

# Parallel execution with timeout
go-sentinel run --parallel=4 --timeout=30s ./...

# High verbosity for debugging
go-sentinel run -vvv --color ./internal

# CI-friendly mode
go-sentinel run --no-color --fail-fast ./...
```

#### `go-sentinel demo --phase=<1-7>`
View development phase demonstrations.

## ⚙️ Configuration

### Configuration File (`sentinel.config.json`)

Create a `sentinel.config.json` file in your project root:

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
    "excludePatterns": ["vendor/**", ".git/**", "node_modules/**"]
  },
  "watch": {
    "enabled": false,
    "debounce": "500ms",
    "ignorePatterns": ["**/.git/**", "**/vendor/**", "**/*.tmp"],
    "clearOnRerun": true,
    "runOnStart": true
  },
  "testCommand": "go test"
}
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `colors` | boolean | `true` | Enable/disable colored output |
| `verbosity` | number | `1` | Output verbosity level (0-5) |
| `parallel` | number | `4` | Number of parallel test processes |
| `timeout` | string | `"10m"` | Test execution timeout |
| `visual.icons` | string | `"unicode"` | Icon style: `unicode`, `ascii`, `minimal`, `none` |
| `visual.theme` | string | `"dark"` | Color theme |
| `watch.debounce` | string | `"500ms"` | File change debounce interval |
| `watch.clearOnRerun` | boolean | `true` | Clear terminal between test runs |

## 📁 Directory Structure (Post-Refactoring)

```
go-sentinel-cli/
├── cmd/
│   ├── go-sentinel-cli/           # Main CLI entry point
│   │   ├── cmd/                   # Cobra command definitions
│   │   └── main.go               # Application entry point
│   └── go-sentinel-cli-v2/        # V2 CLI (development)
├── internal/
│   ├── app/                       # Core application logic
│   │   ├── controller.go         # Main application controller
│   │   └── controller_test.go    # Application tests
│   ├── watch/                     # File watching functionality
│   │   ├── core/                 # Core watch interfaces
│   │   ├── debouncer/            # File change debouncing
│   │   ├── watcher/              # File system watching
│   │   └── runner/               # Watch-triggered test running
│   ├── test/                      # Test execution & processing
│   │   ├── runner/               # Test execution engines
│   │   ├── processor/            # Test output processing
│   │   └── cache/                # Test result caching
│   ├── ui/                        # User interface components
│   │   ├── display/              # Test result display
│   │   ├── colors/               # Color formatting
│   │   └── icons/                # Icon providers
│   └── config/                    # Configuration management
│       ├── loader.go             # Configuration loading
│       └── validation.go         # Configuration validation
├── pkg/
│   ├── events/                    # Event system for inter-component communication
│   └── models/                    # Shared data models
├── docs/                          # Documentation
├── demo-configs/                  # Example configurations
└── stress_tests/                  # Performance and stress tests
```

## 🧪 Development

### Prerequisites
- Go 1.23+
- Make (optional, for convenience commands)

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/cli

# Run with race detection
go test -race ./...
```

### TDD Workflow
1. **Write failing tests** in `*_test.go` files
2. **Implement minimal code** to pass tests
3. **Run test suite**: `go test ./...`
4. **Refactor** while maintaining test coverage
5. **Validate with linting**: `golangci-lint run`

### Code Quality Standards
- **Test Coverage**: ≥ 90% for all new code
- **Linting**: Must pass `golangci-lint run` without errors
- **Formatting**: All code must be `go fmt` compliant
- **Documentation**: All exported symbols must be documented

### Building
```bash
# Build main CLI
go build -o go-sentinel-cli ./cmd/go-sentinel-cli

# Build with optimizations
go build -ldflags="-s -w" -o go-sentinel-cli ./cmd/go-sentinel-cli

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o go-sentinel-cli-linux ./cmd/go-sentinel-cli
GOOS=windows GOARCH=amd64 go build -o go-sentinel-cli.exe ./cmd/go-sentinel-cli
```

## 🤝 Contributing

### Code Style
- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use meaningful variable and function names
- Keep functions focused and small (≤ 50 lines)
- Prefer composition over inheritance

### Pull Request Process
1. **Fork** the repository
2. **Create feature branch**: `git checkout -b feature/amazing-feature`
3. **Write tests** for new functionality
4. **Ensure tests pass**: `go test ./...`
5. **Run linting**: `golangci-lint run`
6. **Update documentation** if needed
7. **Submit pull request** with clear description

### Adding New Features
1. **Design Phase**: Document the feature in an issue
2. **TDD Phase**: Write tests before implementation
3. **Implementation**: Follow the established patterns
4. **Testing**: Ensure ≥ 90% test coverage
5. **Documentation**: Update README and code docs

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by [Vitest](https://vitest.dev/) - Modern testing framework for Vite
- Built with [Cobra](https://github.com/spf13/cobra) - Modern CLI framework for Go
- File watching powered by [fsnotify](https://github.com/fsnotify/fsnotify)

---

<div align="center">
  <strong>Made with ❤️ for the Go community</strong>
</div>