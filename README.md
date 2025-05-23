<div align="center">
  <h1>🚀 Go Sentinel CLI</h1>
  <p>
    <strong>A modern, Vitest-inspired test runner for Go with beautiful terminal output</strong>
  </p>
  <p>
    <a href="https://github.com/newbpydev/go-sentinel-cli/actions">
      <img src="https://github.com/newbpydev/go-sentinel-cli/actions/workflows/test.yml/badge.svg" alt="Build Status">
    </a>
    <a href="https://goreportcard.com/report/github.com/newbpydev/go-sentinel-cli">
      <img src="https://goreportcard.com/badge/github.com/newbpydev/go-sentinel-cli" alt="Go Report Card">
    </a>
    <a href="https://pkg.go.dev/github.com/newbpydev/go-sentinel-cli">
      <img src="https://pkg.go.dev/badge/github.com/newbpydev/go-sentinel-cli" alt="Go Reference">
    </a>
    <a href="LICENSE">
      <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT">
    </a>
  </p>
</div>

## ✨ Overview

Go Sentinel CLI brings the modern, beautiful test runner experience from Vitest to the Go ecosystem. It provides real-time test execution with gorgeous terminal output, smart file watching, and a rich developer experience that makes testing in Go a joy.

Born from the need for better Go testing UX, Go Sentinel CLI transforms standard `go test` output into beautiful, informative displays with clear test suite summaries, detailed failure reporting, and comprehensive statistics.

## 🎨 Beautiful Output

```
🚀 Running tests with go-sentinel...

github.com/myproject/pkg/utils (15 tests | 2 failed) 1240ms 2.1 MB heap used
  ✓ TestStringHelper 45ms
  ✗ TestValidation 230ms
  ✓ TestFormatter 12ms
  ✓ TestConfig 89ms
  ...

────────────────────────────────────────────────────────────────────────────────
                                 Failed Tests 2
────────────────────────────────────────────────────────────────────────────────
FAIL github.com/myproject/pkg/utils > TestValidation

    validation_test.go:25
    Expected validation to pass but got error: invalid input

────────────────────────────────────────────────────────────────────────────────
Test Summary:
Test Files: 3 passed, 1 failed (total: 4)
Tests: 28 passed, 2 failed (total: 30)
Start at: 14:32:15
Duration: 1.2s
────────────────────────────────────────────────────────────────────────────────

⏱️  Tests completed in 1.24s
```

## 🚀 Features

### **🎯 Core Features**
- **🎨 Beautiful Vitest-style Output**: Clean, colorful test results with clear pass/fail indicators
- **📊 Test Suite Display**: File-based organization with test counts, timing, and memory usage
- **❌ Detailed Error Reporting**: Failed tests section with source code context and line numbers
- **📈 Comprehensive Summary**: Overall statistics with timing breakdown
- **🎪 Real-time Processing**: Live updates as tests execute

### **👁️ Watch Mode**
- **📁 Smart File Watching**: Automatically detects file changes and runs relevant tests
- **🎯 Selective Test Running**: Only runs tests affected by changed files
- **⚡ Debounced Updates**: Intelligent handling of rapid file changes
- **🧹 Clean Re-runs**: Optional terminal clearing between test runs

### **⚙️ Configuration System**
- **📄 JSON Configuration**: Flexible `sentinel.config.json` support
- **🎛️ CLI Arguments**: Comprehensive command-line flag system
- **🎨 Visual Customization**: Icons (unicode/ascii/minimal/none), themes, colors
- **📂 Path Management**: Include/exclude patterns for files and directories

### **🔧 Advanced Options**
- **🔄 Parallel Execution**: Configurable parallel test execution
- **⏱️ Timeout Control**: Customizable test timeouts
- **🎯 Pattern Filtering**: Run specific tests by name pattern
- **📊 Verbosity Levels**: Multiple levels of output detail
- **🎨 Color Control**: Enable/disable colored output

## 📦 Installation

### Prerequisites
- Go 1.23 or higher
- Git

### Using Go Install (Recommended)
```bash
go install github.com/newbpydev/go-sentinel-cli/cmd/go-sentinel-cli@latest
```

### Building from Source
```bash
git clone https://github.com/newbpydev/go-sentinel-cli.git
cd go-sentinel-cli
go build -o go-sentinel-cli ./cmd/go-sentinel-cli
```

### Download Releases
Download pre-built binaries from the [GitHub Releases](https://github.com/newbpydev/go-sentinel-cli/releases) page.

## 🚦 Quick Start

### Basic Usage
   ```bash
# Run tests with beautiful output
go-sentinel run

# Run tests in watch mode
go-sentinel run --watch

# Run specific package with verbose output
go-sentinel run -v ./internal/cli

# Run tests matching a pattern
go-sentinel run --test="TestConfig*"
```

### Common Workflows
   ```bash
# Development workflow (watch mode with colors)
go-sentinel run -w --color

# CI/CD workflow (no colors, fail fast)
go-sentinel run --no-color --fail-fast

# Debug mode (maximum verbosity)
go-sentinel run -vvv

# Performance testing (parallel execution)
go-sentinel run --parallel=8 --timeout=5m
```

## 📋 CLI Commands

### Main Commands
- `go-sentinel run [packages]` - Run tests with beautiful output
- `go-sentinel demo --phase=<1-7>` - View development phase demonstrations

### Run Command Flags
   ```bash
go-sentinel run [flags] [packages]

Flags:
  -c, --color              Use colored output (default true)
      --no-color           Disable colored output
  -v, --verbose            Enable verbose output
  -vv, -vvv               Verbosity levels (can be repeated)
  -w, --watch              Enable watch mode for file changes
  -t, --test string        Run only tests matching pattern
  -f, --fail-fast          Stop on first test failure
  -j, --parallel int       Number of tests to run in parallel
      --timeout duration   Timeout for test execution
  -h, --help               Show help information
```

### Examples
   ```bash
# Watch mode with test filtering
go-sentinel run -w --test="TestHandler*" ./api

# Parallel execution with timeout
go-sentinel run --parallel=4 --timeout=30s ./...

# High verbosity for debugging
go-sentinel run -vvv --color ./internal

# CI-friendly mode
go-sentinel run --no-color --fail-fast --parallel=2 ./...
```

## ⚙️ Configuration

### Configuration File (`sentinel.config.json`)

Create a `sentinel.config.json` file in your project root for persistent settings:

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
    "ignorePatterns": ["**/*_test.go"],
    "clearOnRerun": true,
    "runOnStart": true
  }
}
```

### Configuration Options

#### Visual Configuration
- **`colors`** (bool): Enable/disable colored output
- **`icons`** (string): Icon style - `unicode`, `ascii`, `minimal`, `none`
- **`theme`** (string): Color theme - `dark`, `light`, `auto`

#### Execution Configuration
- **`verbosity`** (int): Output verbosity level (0-3)
- **`parallel`** (int): Number of parallel test executions
- **`timeout`** (duration): Maximum test execution time
- **`testPattern`** (string): Pattern to filter tests

#### Path Configuration
- **`includePatterns`** ([]string): File patterns to include
- **`excludePatterns`** ([]string): File patterns to exclude

#### Watch Configuration
- **`enabled`** (bool): Enable watch mode by default
- **`debounce`** (duration): Delay before running tests after file changes
- **`ignorePatterns`** ([]string): File patterns to ignore during watching
- **`clearOnRerun`** (bool): Clear terminal before re-running tests
- **`runOnStart`** (bool): Run tests immediately when watch mode starts

### CLI Arguments Override Configuration
CLI arguments always take precedence over configuration file settings, allowing for flexible per-run customization.

## 🏗️ Project Structure

```
go-sentinel-cli/
├── cmd/
│   └── go-sentinel-cli/
│       ├── main.go              # Application entry point
│       └── cmd/
│           ├── root.go          # Root command configuration
│           ├── run.go           # Main test runner command
│           └── demo/            # Development phase demos
│               ├── demo.go      # Demo command handler
│               ├── phase1.go    # Core architecture demo
│               ├── phase2.go    # Test suite display demo
│               ├── phase3.go    # Failed test details demo
│               ├── phase4.go    # Real-time processing demo
│               ├── phase5.go    # Watch mode demo
│               ├── phase6d.go   # Performance demo
│               └── phase7d.go   # CLI options demo
│
├── internal/cli/                # Core CLI implementation
│   ├── app_controller.go        # Main application controller
│   ├── cli_args.go             # CLI argument parsing
│   ├── config.go               # Configuration system
│   ├── colors.go               # Color and formatting
│   ├── display.go              # Test suite display
│   ├── failed_tests.go         # Failed test rendering
│   ├── models.go               # Core data structures
│   ├── parser.go               # Test output parsing
│   ├── processor.go            # Test result processing
│   ├── summary.go              # Summary generation
│   ├── test_runner.go          # Test execution
│   ├── watcher.go              # File watching
│   └── performance_optimizations.go # Performance features
│
├── docs/                       # Documentation
│   ├── configuration.md        # Configuration guide
│   └── assets/                 # Documentation assets
│
├── demo-configs/               # Example configurations
├── .golangci.yml              # Linting configuration
├── go.mod                     # Go module definition
├── go.sum                     # Dependency checksums
├── README.md                  # This file
└── ROADMAP-CLI-VITEST-V2.md   # Development roadmap
```

## 🎬 Demonstrations

Go Sentinel CLI includes interactive demonstrations of each development phase:

```bash
# View core architecture (data structures, parsing)
go-sentinel demo --phase=1

# See test suite display formatting
go-sentinel demo --phase=2

# Explore failed test detail rendering
go-sentinel demo --phase=3

# Watch real-time processing and summary
go-sentinel demo --phase=4

# Experience watch mode functionality
go-sentinel demo --phase=5

# Test performance optimizations
go-sentinel demo --phase=6

# Try CLI options and configuration
go-sentinel demo --phase=7
```

## 🔧 Development

### Building
```bash
# Build the application
go build -o go-sentinel-cli ./cmd/go-sentinel-cli

# Run tests
go test ./...

# Run tests with coverage
go test ./... -cover

# Format code
go fmt ./...

# Lint code (requires golangci-lint)
golangci-lint run
```

### Testing
```bash
# Run the CLI on itself
./go-sentinel-cli run ./internal/cli

# Test with watch mode
./go-sentinel-cli run -w ./internal/cli

# Test with different configurations
./go-sentinel-cli run -vv --parallel=2 ./...
```

## 🎯 Use Cases

### **👨‍💻 Development Workflow**
```bash
go-sentinel run -w --color           # Watch mode with colors
```
Perfect for TDD with immediate feedback on file changes.

### **🏭 CI/CD Pipeline**
```bash
go-sentinel run --no-color --fail-fast --parallel=4 ./...
```
Fast, parallel execution with clean output for automation.

### **🐛 Debugging Tests**
```bash
go-sentinel run -vvv --test="TestProblem*"
```
Maximum verbosity with focused test execution.

### **⚡ Performance Testing**
```bash
go-sentinel run --parallel=8 --timeout=5m ./...
```
High-performance execution for large test suites.

## 🤝 Contributing

We welcome contributions! Here's how to get involved:

1. **Fork the repository** and clone your fork
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes** and add tests
4. **Run the test suite**: `go test ./...`
5. **Format your code**: `go fmt ./...`
6. **Commit your changes**: `git commit -m 'Add amazing feature'`
7. **Push to your branch**: `git push origin feature/amazing-feature`
8. **Open a Pull Request**

### Development Setup
```bash
git clone https://github.com/newbpydev/go-sentinel-cli.git
cd go-sentinel-cli
go mod download
go build -o go-sentinel-cli ./cmd/go-sentinel-cli
./go-sentinel-cli run ./internal/cli
```

## 📖 Documentation

- **[Configuration Guide](docs/configuration.md)** - Detailed configuration options
- **[Development Roadmap](ROADMAP-CLI-VITEST-V2.md)** - Project development history
- **[API Documentation](https://pkg.go.dev/github.com/newbpydev/go-sentinel-cli)** - Go package documentation

## 🗺️ Roadmap

Go Sentinel CLI follows a structured development approach with completed phases:

- ✅ **Phase 1**: Core Architecture & Data Structures  
- ✅ **Phase 2**: Test Suite Display
- ✅ **Phase 3**: Failed Test Details Section
- ✅ **Phase 4**: Real-time Processing & Summary
- ✅ **Phase 5**: Watch Mode & Integration
- ✅ **Phase 6**: Performance & Error Handling
- ✅ **Phase 7**: CLI Options & Configuration
- ✅ **Phase 8.1**: Main Application Integration
- 🚧 **Phase 8.2**: Final Documentation (In Progress)
- 📋 **Phase 8.3**: Final Testing & Validation

See [ROADMAP-CLI-VITEST-V2.md](ROADMAP-CLI-VITEST-V2.md) for detailed development history.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **[Vitest](https://vitest.dev/)** - Inspiration for the beautiful test output format
- **[Go Team](https://golang.org/)** - For the excellent testing tools and ecosystem
- **[Cobra](https://github.com/spf13/cobra)** - For the CLI framework
- **[fsnotify](https://github.com/fsnotify/fsnotify)** - For efficient file watching

---

<div align="center">
  <p>
    <strong>Made with ❤️ for the Go community</strong>
  </p>
  <p>
    Give us a ⭐ if Go Sentinel CLI makes your testing experience better!
  </p>
</div>