# Go Sentinel Project Roadmap

This roadmap is the single source of truth for Go Sentinel's development. All work must be reflected here. **We strictly follow Test-Driven Development (TDD): every feature or fix begins with a test.**

---

## Phase 1: Project & Environment Setup

- [ ] **1.1. Initialize Project Repository**
  - [ ] 1.1.1. Create a new git repository (if not already done)
  - [ ] 1.1.2. Add `.gitignore` for Go and editor files
  - [ ] 1.1.3. Set up repository on GitHub (public, with license)
  - [ ] 1.1.4. Add `README.md`, `ROADMAP.md`, and `LICENSE`
- [ ] **1.2. Go Environment Setup**
  - [ ] 1.2.1. Install Go (minimum version per design)
  - [ ] 1.2.2. Initialize Go module (`go mod init`)
  - [ ] 1.2.3. Set up directory structure (`cmd/`, `internal/`, `pkg/`, `test/`)
- [ ] **1.3. Tooling & CI/CD**
  - [ ] 1.3.1. Set up code formatter (`gofmt`, `goimports`)
  - [ ] 1.3.2. Set up linter (e.g., `golangci-lint`)
  - [ ] 1.3.3. Set up pre-commit hooks
  - [ ] 1.3.4. Configure CI (GitHub Actions) for test, lint, build

---

## Phase 2: Core File Watcher & Debouncer (TDD)

- [ ] **2.1. Write File Watcher Tests**
  - [ ] 2.1.1. Test: Detect file changes in Go source dirs
  - [ ] 2.1.2. Test: Ignore `vendor/` and hidden dirs
  - [ ] 2.1.3. Test: Handle file create, write, remove events
- [ ] **2.2. Implement File Watcher**
  - [ ] 2.2.1. Integrate `fsnotify` for recursive watching
  - [ ] 2.2.2. Correctly skip excluded directories
  - [ ] 2.2.3. Emit events to channel
- [ ] **2.3. Write Debouncer Tests**
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
- [ ] **3.2. Implement Test Runner**
  - [ ] 3.2.1. Use `os/exec` to run `go test -json`
  - [ ] 3.2.2. Stream output to parser
- [ ] **3.3. Write Output Parser Tests**
  - [ ] 3.3.1. Test: Parse JSON events from test output
  - [ ] 3.3.2. Test: Summarize results, handle edge cases
- [ ] **3.4. Implement Output Parser**
  - [ ] 3.4.1. Parse and structure test results
  - [ ] 3.4.2. Handle errors and malformed output

---

## Phase 4: Interactive CLI UI & Controller (TDD)

- [ ] **4.1. Write UI/Controller Tests**
  - [ ] 4.1.1. Test: Display summary with color (ANSI)
  - [ ] 4.1.2. Test: Keybindings (Enter, f, q)
  - [ ] 4.1.3. Test: Filter failures mode
- [ ] **4.2. Implement CLI UI**
  - [ ] 4.2.1. Render summary, color output
  - [ ] 4.2.2. Implement interactive controls
  - [ ] 4.2.3. Channel communication between components

---

## Phase 5: Configuration, Logging, and Extensibility (TDD)

- [ ] **5.1. Write Config/Logging Tests**
  - [ ] 5.1.1. Test: Parse flags and config files (YAML/JSON)
  - [ ] 5.1.2. Test: Logging at various levels
  - [ ] 5.1.3. Test: User toggles (color, debounce, verbosity)
- [ ] **5.2. Implement Config/Logging**
  - [ ] 5.2.1. Integrate config parser (e.g., viper)
  - [ ] 5.2.2. Integrate zap logger

---

## Phase 6: Documentation, Packaging, and Release

- [ ] **6.1. Documentation**
  - [ ] 6.1.1. Update README with usage and examples
  - [ ] 6.1.2. Document all flags and config options
  - [ ] 6.1.3. Add man page and help output
- [ ] **6.2. Packaging**
  - [ ] 6.2.1. Build binaries for major platforms
  - [ ] 6.2.2. Test installation instructions
- [ ]  **6.3. Release**
  - [ ] 6.3.1. Tag and create GitHub release
  - [ ] 6.3.2. Announce and gather feedback

---

## Phase 7: Maintenance & Community

- [ ] **7.1. Issue Triage and PR Review**
  - [ ] 7.1.1. Review issues and pull requests
  - [ ] 7.1.2. Update roadmap and docs as needed
- [ ] **7.2. Continuous Improvement**
  - [ ] 7.2.1. Refactor for performance and reliability
  - [ ] 7.2.2. Expand test coverage
  - [ ] 7.2.3. Respond to user feedback

---

**Remember:**
- Every implementation task is preceded by a test task.
- Update this roadmap as you progress. Check off tasks/subtasks as you complete them.
- All code must be covered by tests before merging.

---

This roadmap is your guide. Stick to it, improve it, and let it drive Go Sentinel toward a robust, error-free, and community-friendly release.
