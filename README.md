<div align="center">
  <h1>Go Sentinel</h1>
  <p>
    <strong>Accelerate your Go test-driven development workflow with real-time feedback</strong>
  </p>
  <p>
    <a href="https://github.com/newbpydev/go-sentinel/actions">
      <img src="https://github.com/newbpydev/go-sentinel/actions/workflows/test.yml/badge.svg" alt="Build Status">
    </a>
    <a href="https://goreportcard.com/report/github.com/newbpydev/go-sentinel">
      <img src="https://goreportcard.com/badge/github.com/newbpydev/go-sentinel" alt="Go Report Card">
    </a>
    <a href="https://pkg.go.dev/github.com/newbpydev/go-sentinel">
      <img src="https://pkg.go.dev/badge/github.com/newbpydev/go-sentinel" alt="Go Reference">
    </a>
    <a href="LICENSE">
      <img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License: MIT">
    </a>
  </p>
</div>

## 🚀 Overview

Go Sentinel is an open-source, Go-native CLI utility that supercharges your test-driven development (TDD) workflow. It automatically watches your Go source files, reruns tests on changes, and presents concise, actionable feedback in your terminal. Built with concurrency and resilience at its core, Go Sentinel helps you maintain an uninterrupted TDD flow.

## ✨ Features

- **Real-time Test Execution**: Automatically runs tests when files change
- **Smart Debouncing**: Coalesces rapid file system events to prevent redundant test runs
- **Rich Terminal UI**: Color-coded output with clear pass/fail indicators
- **Interactive Controls**:
  - `Enter`: Rerun all tests
  - `f`: Filter to show only failing tests
  - `c/C`: Copy test information (current/all failures)
  - `q`: Quit the application
- **Robust Error Handling**:
  - Automatic test timeouts (configurable, default: 2m)
  - Deadlock detection and reporting
  - Structured logging for debugging
- **Modern Tech Stack**:
  - Built with Go 1.17+
  - Uses [fsnotify](https://github.com/fsnotify/fsnotify) for efficient file watching
  - [zap](https://github.com/uber-go/zap) for high-performance structured logging

## 🏗️ Project Structure

```
/go-sentinel
├── cmd/                  # Command-line applications
│   ├── go-sentinel-api/  # Web API server with WebSocket support
│   └── go-sentinel-web/  # Web interface server
│
├── internal/           # Private application code
│   ├── api/             # API server implementation
│   ├── config/          # Configuration management
│   ├── debouncer/       # Event debouncing logic
│   ├── event/           # Event types and handling
│   ├── parser/          # Test output parsing
│   ├── runner/          # Test execution
│   ├── watcher/         # File system watching
│   └── web/             # Web interface implementation
│
├── web/                # Web interface assets
│   ├── static/          # Static assets
│   │   ├── css/         # Stylesheets
│   │   ├── images/      # Image assets
│   │   └── js/          # JavaScript files
│   └── templates/       # Server-side templates
│       ├── layouts/     # Base templates
│       ├── pages/       # Page-specific templates
│       └── partials/    # Reusable template components
│
├── docs/               # Project documentation
│   ├── assets/          # Documentation assets
│   ├── COVERAGE.md      # Test coverage information
│   ├── IMPLEMENTATION_PLAN.md
│   ├── RESEARCH-API.md
│   └── RESEARCH.md
│
├── .github/            # GitHub configurations
├── testdata/            # Test fixtures and data
├── CHANGELOG.md         # Release history
├── CHANGES.md           # Detailed change log
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── LICENSE              # MIT License
├── README.md            # This file
├── ROADMAP.md           # Main project roadmap
├── ROADMAP-API.md       # API development roadmap
├── ROADMAP-FRONTEND.md  # Frontend development roadmap
└── ROADMAP-INTEGRATION.md # Integration roadmap
```

## 📦 Installation

### Prerequisites
- Go 1.17 or higher
- Git

### Using Go Install
```bash
go install github.com/newbpydev/go-sentinel/cmd/go-sentinel-cli@latest
```

### Building from Source
```bash
git clone https://github.com/newbpydev/go-sentinel-cli.git
cd go-sentinel
make build
```

## 🚦 Quick Start

1. Navigate to your Go project directory
2. Run:
   ```bash
   go-sentinel-cli
   ```
3. Start editing your files - tests will run automatically on save

## ⚙️ Configuration

Create a `watcher.yaml` file in your project root:

```yaml
# Default configuration for Go Sentinel CLI
watch:
  # Directories to watch (default: ["."])
  dirs: ["."]
  
  # File patterns to include (default: ["*.go"])
  include: ["*.go"]
  
  # Directories to exclude (default: ["vendor", ".git"])
  exclude: ["vendor", ".git"]
  
  # Debounce interval in milliseconds (default: 100)
  debounce: 100

test:
  # Test timeout duration (default: 2m)
  timeout: 2m
  
  # Enable/disable test caching (default: true)
  cache: true
  
  # Additional go test flags
  args: ["-v", "-race"]

log:
  # Log level (debug, info, warn, error)
  level: "info"
  
  # Log format (console, json)
  format: "console"
```

## 📚 Documentation

- [Getting Started](docs/getting-started.md)
- [Configuration Guide](docs/configuration.md)
- [Development Guide](docs/development.md)
- [API Reference](docs/api.md)

## 🤝 Contributing

Go Sentinel is an open source project and welcomes contributions! To get involved:
- **Read the [docs/RESEARCH.md](docs/RESEARCH.md)** for design context and technical approach
- **Open issues** for bugs, feature requests, or questions
- **Submit pull requests** with clear descriptions and tests
- **Follow Go best practices** for code style and documentation
- **Add/update tests** for any code changes
- **Document public flags and configuration**
- **Use semantic versioning** ([semver.org](https://semver.org/))

### Development Best Practices
- Always provide `--help` and `--version` flags
- Document all public flags and configuration files
- Use clear, predictable exit codes
- Prefer human-friendly output by default, but support machine-readable formats (e.g., JSON)
- Keep output concise and support color toggling for accessibility
- Follow [CLI best practices](https://github.com/arturtamborski/cli-best-practices)

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by [modd](https://github.com/cortesi/modd) and [reflex](https://github.com/cespare/reflex)
- Built with [cobra](https://github.com/spf13/cobra) and [viper](https://github.com/spf13/viper)
- Uses [bubbletea](https://github.com/charmbracelet/bubbletea) for TUI
- Uses [fsnotify](https://github.com/fsnotify/fsnotify), [zap](https://github.com/uber-go/zap), and other open source libraries

---

<div align="center">
  Made with ❤️ by the Go Sentinel Team
</div>

## Development Roadmap

Our detailed [ROADMAP.md](ROADMAP.md) outlines the full development plan in phases:

1. **Project & Environment Setup** - Git, package structure, CI/CD
2. **Core File Watcher & Debouncer** - fsnotify integration, event coalescing
3. **Test Runner & Output Parser** - JSON stream processing, structured results
4. **Interactive CLI UI & Controller** - ANSI color, keybindings, code context
5. **Concurrency & Resilience** - Pipeline pattern, panic recovery
6. **Configuration & Validation** - CLI flags, config file support
7. **Extensibility & Integrations** - Plugins, per-test reruns, coverage tools
8. **Documentation, Packaging, Release** - Binaries, installation
9. **Maintenance & Community** - Issue triage, continuous improvement

## Automatic Server Restart with air

For a smoother, test-driven development experience, Go Sentinel supports automatic server restarts and test execution on file changes using [`air`](https://github.com/cosmtrek/air).

### Setup

1. **Install air**
   ```bash
   go install github.com/cosmtrek/air@latest
   ```
   Ensure your Go bin directory (e.g., `$GOPATH/bin` or `$HOME/go/bin`) is in your system PATH.

2. **Verify `.air.toml`**
   The project root contains a preconfigured `.air.toml`:
   ```toml
   [build]
     cmd = "go run ./cmd/go-sentinel-web/main.go"
     bin = "tmp/main"
     full_bin = "false"
     # Run tests before rebuilding
     before_build = "go test ./..."

   [watch]
     dirs = ["./cmd", "./internal"]
     include_ext = ["go", "tmpl", "html", "css", "js"]

   [log]
     color = "true"
     time = "true"
   ```

### Usage

- From the project root, run:
  ```bash
  air
  ```
- The development server will automatically restart whenever Go code, templates, or static files change.
- All Go tests (`go test ./...`) are run before each rebuild. If any test fails, the server will not restart until tests pass.

### Customization
- Edit `.air.toml` to watch additional directories or file types as needed.
- See [air documentation](https://github.com/cosmtrek/air) for advanced configuration.

### Troubleshooting
- If `air` is not found, verify your Go bin directory is in your PATH and restart your terminal.
- For Windows, you may need to restart your shell or log out/in after installing `air`.

> This workflow is aligned with Go Sentinel's systematic, TDD-first, and roadmap-driven approach. For more, see `.windsurf/workflows/dev-auto-restart.md`.

## API Server & Frontend Integration

Go Sentinel provides a RESTful API and WebSocket server for frontend dashboards and automation. The API server is a separate executable with its own entrypoint.

### Running the API Server

From your project root:

```sh
# Start the API server (default port 8080)
go run ./cmd/go-sentinel-api/main.go
```

- The server will listen on `http://localhost:8080` by default. Set `API_PORT` to override.
- OpenAPI documentation is available at `http://localhost:8080/docs`.
- Interactive Swagger UI is available at `http://localhost:8080/docs/ui`.
- WebSocket endpoint: `ws://localhost:8080/ws`

### Connecting a Frontend

- Point your frontend HTTP and WebSocket requests to the API server URL (default: `http://localhost:8080`).
- Supports CORS for local development (customize as needed).
- See `ROADMAP-FRONTEND.md` for frontend project structure and integration steps.

---

## TUI Sidebar Conventions

The Go Sentinel TUI sidebar displays a minimal, clean tree of your test suite:

- **Sidebar shows only node names** for packages, files, and tests.
- **No icons, durations, or coverage** are shown for test nodes in the sidebar.
- **All test status, durations, and coverage** are shown in the details pane when a node is selected.
- **Whitespace and formatting** in the sidebar are robustly handled by the test suite, ensuring consistent output across platforms.

This approach keeps the sidebar uncluttered and focused, while providing full details in the main pane for selected items.

## Usage (Planned)

### Installation
```bash
# Install from source
go install github.com/your-org/go-sentinel/cmd/go-sentinel@latest

# Or download binary from releases (future)
```

### Basic Usage
```bash
# Run in your Go project root
go-sentinel

# With custom options
go-sentinel --debounce 200ms --no-color --exclude "vendor,generated"
```

### Configuration File
Create `watcher.yaml` in your project:
```yaml
exclude: ["vendor", "testdata", ".git"]
debounce: 200ms
color: true
verbosity: info
```

### Keyboard Controls
- `Enter`: Rerun the last test(s)
- `f`: Toggle failure-only mode
- `c`: Quick copy of all failed test information to clipboard
- `C`: Enter selection mode to choose specific test failures to copy
  - `Space`: Toggle selection of a test under cursor
  - `↑/↓`: Navigate between tests
  - `Enter`: Copy selected tests and return to main view
  - `Esc`: Cancel and return to main view
- `q`: Quit the watcher

See `go-sentinel --help` for all options.
