# Go Sentinel Project Roadmap

This roadmap is the single source of truth for Go Sentinel's development. All work must be reflected here. **We strictly follow Test-Driven Development (TDD): every feature or fix begins with a test.**

---

## Phase 1: Project & Environment Setup

- [ ] **1.1. Initialize Project Repository**
  - [x] 1.1.1. Create a new git repository (if not already done)
  - [x] 1.1.2. Add `.gitignore` for Go and editor files
  - [ ] 1.1.3. Set up repository on GitHub (public, with license)
  - [x] 1.1.4. Add `README.md`, `ROADMAP.md`, and `LICENSE`
- [ ] **1.2. Go Environment Setup**
  - [ ] 1.2.1. Install Go (minimum version 1.17+ required for fsnotify)
  - [x] 1.2.2. Initialize Go module (`go mod init`)
  - [x] 1.2.3. Set up package structure following Go best practices:
    > Project structure scaffolded as per README.md guidance.
    ```
    cmd/go-sentinel/main.go        # CLI entrypoint (flag parsing, setup)
    internal/watcher/watcher.go    # fsnotify logic, recursive watching
    internal/debouncer/debounce.go # Event debouncing logic
    internal/runner/runner.go      # Executes go test -json, streams output
    internal/parser/parser.go      # Parses JSON into test results structs
    internal/ui/ui.go              # Rendering logic, key handling
    internal/config/config.go      # Reads flags/yaml, holds options
    internal/event/event.go        # Defines event/result types shared across packages
    ```
- [ ] **1.3. Tooling & CI/CD**
  - [x] 1.3.1. Set up code formatter (`gofmt`, `goimports`)
  - [x] 1.3.2. Set up linter (e.g., `golangci-lint`)
  - [x] 1.3.3. Set up pre-commit hooks
  - [x] 1.3.4. Configure CI (GitHub Actions) for test, lint, build

---

## Phase 2: Core File Watcher & Debouncer (TDD)

- [x] **2.1. Write File Watcher Tests** (scaffolded)
  - [ ] 2.1.1. Test: Detect file changes in Go source dirs
  - [ ] 2.1.2. Test: Ignore `vendor/` and hidden dirs
  - [ ] 2.1.3. Test: Handle file create, write, remove events
- [ ] **2.2. Implement File Watcher**
  - [ ] 2.2.1. Integrate `fsnotify` for recursive watching
  - [ ] 2.2.2. Correctly skip excluded directories
  - [ ] 2.2.3. Emit events to channel
- [x] **2.3. Write Debouncer Tests** (scaffolded)
  - [ ] 2.3.1. Test: Buffer rapid events per package
  - [ ] 2.3.2. Test: Trigger only after quiet period
- [ ] **2.4. Implement Debouncer**
  - [ ] 2.4.1. Buffer and coalesce events
  - [ ] 2.4.2. Trigger test run after debounce interval

---

## Phase 3: Test Runner & Output Parser (TDD)

- [ ] **3.1. Write Test Runner Tests**
  - [ ] 3.1.1. Test: Run `go test -json` in correct pkg
  - [ ] 3.1.2. Test: Capture stdout/stderr, handle errors
  - [ ] 3.1.3. Test: Handle non-JSON output (build errors)
  - [ ] 3.1.4. Test: Pipe stdout/stderr for real-time output
  - [ ] 3.1.5. Test: Integration with goroutine pipeline pattern
- [ ] **3.2. Implement Test Runner**
  - [ ] 3.2.1. Use `os/exec` to run `go test -json`
  - [ ] 3.2.2. Stream output to parser through channels
  - [ ] 3.2.3. Handle and log command errors properly
  - [ ] 3.2.4. Support future per-test reruns with `-run=TestName`
- [ ] **3.3. Write Output Parser Tests**
  - [ ] 3.3.1. Test: Parse TestEvent JSON objects from output stream
  - [ ] 3.3.2. Test: Track test start/run/pass/fail/output events
  - [ ] 3.3.3. Test: Extract file/line information from failure output
  - [ ] 3.3.4. Test: Collect test durations and output lines
  - [ ] 3.3.5. Test: Handle edge cases (build errors, test panics, timeouts)
- [ ] **3.4. Implement Output Parser**
  - [ ] 3.4.1. Parse TestEvent structs from JSON stream
  - [ ] 3.4.2. Group events by package/test name
  - [ ] 3.4.3. Extract error context and file locations
  - [ ] 3.4.4. Provide structured results to UI component

---

## Phase 4: Interactive CLI UI & Controller (TDD)

- [ ] **4.1. Write UI/Controller Tests**
  - [ ] 4.1.1. Test: Display summary with color (ANSI)
  - [ ] 4.1.2. Test: Keybindings (Enter, f, q)
  - [ ] 4.1.3. Test: Filter failures mode
  - [ ] 4.1.4. Test: Show code context for failed tests
  - [ ] 4.1.5. Test: UI updates on each run without exit
  - [ ] 4.1.6. Test: Copy failed test information to clipboard ('c' key)
  - [ ] 4.1.7. Test: Interactive test selection and copying ('C' key, space for selection)
- [ ] **4.2. Implement CLI UI**
  - [ ] 4.2.1. Render summary, color output, icons
  - [ ] 4.2.2. Implement interactive controls (keypresses)
  - [ ] 4.2.3. Display code context for failures (extract lines from source)
  - [ ] 4.2.4. Channel communication between components
  - [ ] 4.2.5. Implement clipboard integration for test failures
  - [ ] 4.2.6. Create interactive test selection mode with visual indicators
  - [ ] 4.2.7. (Future) Integrate TUI framework (Bubble Tea/tview)

---

## Phase 5: Concurrency & Resilience (TDD)

- [ ] **5.1. Write Pipeline/Recovery Tests**
  - [ ] 5.1.1. Test: Watcher, Debouncer, Runner, Parser, UI as goroutines with channels
  - [ ] 5.1.2. Test: Pipeline pattern (input/output channels)
  - [ ] 5.1.3. Test: Each goroutine recovers from panic and logs error
  - [ ] 5.1.4. Test: Program never crashes on test/compile errors
- [ ] **5.2. Implement Concurrency Pipeline**
  - [ ] 5.2.1. Implement each stage as a goroutine with channel communication
  - [ ] 5.2.2. Add panic recovery and error logging to each goroutine
  - [ ] 5.2.3. Ensure resilience: watcher stays alive on errors

---

## Phase 6: Configuration & Validation (TDD)

- [ ] **6.1. Write Config Tests**
  - [ ] 6.1.1. Test: Parse CLI flags and YAML/JSON config file
  - [ ] 6.1.2. Test: Validate config (includes/excludes, debounce, color, verbosity)
  - [ ] 6.1.3. Test: Config precedence (flags override file)
- [ ] **6.2. Implement Config Support**
  - [ ] 6.2.1. Implement CLI flags and watcher.yaml support
  - [ ] 6.2.2. Validate config at startup, error on invalid
  - [ ] 6.2.3. Mirror CLI flags in config file (viper/cobra)

---

## Phase 7: Extensibility & Integrations (TDD)

- [ ] **7.1. Write Extensibility Tests**
  - [ ] 7.1.1. Test: Plugin/event hooks for test/file events
  - [ ] 7.1.2. Test: Per-test rerun logic
  - [ ] 7.1.3. Test: Coverage and lint integration
  - [ ] 7.1.4. Test: Editor/IDE integration API
  - [ ] 7.1.5. Test: Custom output reporters
- [ ] **7.2. Implement Extensibility & Integrations**
  - [ ] 7.2.1. Implement plugin/event hook interfaces
  - [ ] 7.2.2. Add per-test rerun support (`go test -run=Name`)
  - [ ] 7.2.3. Integrate coverage and lint tools (e.g., golangci-lint)
  - [ ] 7.2.4. Provide API/protocol for editor/IDE integration
  - [ ] 7.2.5. Implement custom reporter/output format support

---

## Phase 8: Documentation, Packaging, and Release

- [ ] **8.1. Documentation**
  - [ ] 8.1.1. Update README with comprehensive usage and examples
  - [ ] 8.1.2. Document all flags and config options in watcher.yaml format
  - [ ] 8.1.3. Add man page and help output following standard conventions
  - [ ] 8.1.4. Create sample configuration files with common settings
- [ ] **8.2. Packaging**
  - [ ] 8.2.1. Build single-binary executable for all major platforms
  - [ ] 8.2.2. Set up installation via `go install` command
  - [ ] 8.2.3. Test installation instructions on all target platforms
  - [ ] 8.2.4. Create update mechanism (`go-sentinel update`) for easy upgrades
- [ ] **8.3. Release**
  - [ ] 8.3.1. Tag and create GitHub release with semantic versioning
  - [ ] 8.3.2. Publish release notes and changelog
  - [ ] 8.3.3. Announce release and collect initial feedback

---

## Phase 9: Maintenance & Community

- [ ] **9.1. Issue Triage and PR Review**
  - [ ] 9.1.1. Set up issue templates for bugs, features, etc.
  - [ ] 9.1.2. Establish PR review process with TDD validation
  - [ ] 9.1.3. Update roadmap and docs per community feedback
- [ ] **9.2. Continuous Improvement**
  - [ ] 9.2.1. Refactor for performance and reliability
  - [ ] 9.2.2. Expand test coverage to edge cases
  - [ ] 9.2.3. Respond to user feedback and add frequently requested features
  - [ ] 9.2.4. Monitor and improve error handling and recovery

---

**Remember:**
- Every implementation task is preceded by a test task.
- Update this roadmap as you progress. Check off tasks/subtasks as you complete them.
- All code must be covered by tests before merging.

---

This roadmap is your guide. Stick to it, improve it, and let it drive Go Sentinel toward a robust, error-free, and community-friendly release.
