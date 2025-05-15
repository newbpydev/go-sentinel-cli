# Go Sentinel

Go Sentinel is an open source, Go-native CLI utility that supercharges your test-driven development (TDD) workflow by automatically watching your Go source files, rerunning tests on changes, and presenting concise, actionable feedback in your terminal. Built with concurrency and resilience at its core, Go Sentinel helps you keep your TDD flow uninterrupted and productive.

## What Problem Does Go Sentinel Solve?

Manual test execution slows down TDD and feedback loops. Existing tools are often language-agnostic, slow, or lack Go-native ergonomics. Go Sentinel solves this by:
- **Automatically detecting file changes** in your Go project (excluding vendor and generated files)
- **Debouncing rapid events** to avoid redundant test runs
- **Running `go test -json`** per package for accurate, real-time results
- **Parsing and summarizing test output** with color-coded, human-friendly CLI feedback
- **Providing keyboard shortcuts** for rerun, filtering failures, and quitting
- **Ensuring stability** with robust error handling and structured logging

## Key Features
- Fast, recursive file watching using [fsnotify](https://github.com/fsnotify/fsnotify) (Go 1.17+)
- Intelligent debouncing that coalesces rapid events per package (~100ms quiet period)
- Real-time, colored summary with test durations and contextual error information
- Robust timeout and deadlock protection:
  - Automatic test timeouts with configurable duration (default: 2m)
  - Context-based cancellation for graceful termination
  - Deadlock detection and meaningful error reporting
  - Inactivity monitoring to detect and report hanging tests
- Minimal, intuitive keybindings (Enter: rerun tests, f: filter failures only, c/C: copy test information, q: quit)
- Structured logging with [zap](https://github.com/uber-go/zap) for reliable diagnostics
- Resilient architecture: pipeline pattern with panic recovery in each goroutine
- Configurable via CLI flags and/or `watcher.yaml` file
- Extensible plugin architecture for custom integrations (planned)
- Test reruns at package or individual test level (future versions)

## Project Structure

```
/go-sentinel
│
├── cmd/
│   ├── go-sentinel/         # CLI entrypoint (flag parsing, setup)
│   └── go-sentinel-api/     # API server entrypoint (for web/frontend integration)
│
├── internal/
│   ├── api/                 # API server, OpenAPI config, handlers
│   ├── watcher/             # fsnotify logic, recursive directory watching
│   ├── debouncer/           # Event debouncing to coalesce rapid changes
│   ├── runner/              # Executes go test -json, streams output
│   ├── parser/              # Parses JSON stream into test results
│   ├── ui/                  # Rendering logic, key handling
│   ├── config/              # Configuration management
│   └── event/               # Event types shared across packages
│
├── docs/
│   ├── RESEARCH.md          # High-level & detailed design, rationale, technical notes
│   └── assets/              # Architecture diagrams and images
│
├── README.md                # This file
├── ROADMAP.md               # Development roadmap and task tracking
└── LICENSE                  # MIT License
```

> **Note:** The package structure follows Go best practices with clear separation of concerns.

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

## Open Source & Contribution Guidelines
Go Sentinel is an open source project and welcomes contributions! To get involved:
- **Read the [docs/RESEARCH.md](docs/RESEARCH.md)** for design context and technical approach
- **Open issues** for bugs, feature requests, or questions
- **Submit pull requests** with clear descriptions and tests
- **Follow Go best practices** for code style and documentation
- **Add/update tests** for any code changes
- **Document public flags and configuration**
- **Use semantic versioning** ([semver.org](https://semver.org/))

## Best Practices (from the Community)
- Always provide `--help` and `--version` flags
- Document all public flags and configuration files
- Use clear, predictable exit codes
- Prefer human-friendly output by default, but support machine-readable formats (e.g., JSON)
- Keep output concise and support color toggling for accessibility
- Support easy updates (planned: `go-sentinel update`)
- Follow [CLI best practices](https://github.com/arturtamborski/cli-best-practices)

## License
MIT License. See [LICENSE](LICENSE) for details.

## Acknowledgments
- Inspired by Go TDD workflows and community feedback
- Uses [fsnotify](https://github.com/fsnotify/fsnotify), [zap](https://github.com/uber-go/zap), and other open source libraries

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

We follow strict Test-Driven Development throughout all phases.

---

For detailed design, see [`docs/RESEARCH.md`](docs/RESEARCH.md). Contributions and feedback are welcome!
