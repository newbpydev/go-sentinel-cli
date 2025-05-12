# Go Sentinel Project Roadmap

This roadmap is the single source of truth for Go Sentinel's development. All work must be reflected here. **We strictly follow Test-Driven Development (TDD): every feature or fix begins with a test.**

---

## Phase 1: Project & Environment Setup

- [x] **1.1. Initialize Project Repository**
  - [x] 1.1.1. Create a new git repository (if not already done)
  - [x] 1.1.2. Add `.gitignore` for Go and editor files
  - [x] 1.1.3. Set up repository on GitHub (public, with license)
  - [x] 1.1.4. Add `README.md`, `ROADMAP.md`, and `LICENSE`
- [x] **1.2. Go Environment Setup**
  - [x] 1.2.1. Install Go (minimum version 1.17+ required for fsnotify)
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
- [x] **1.3. Tooling & CI/CD**
  - [x] 1.3.1. Set up code formatter (`gofmt`, `goimports`)
  - [x] 1.3.2. Set up linter (e.g., `golangci-lint`)
  - [x] 1.3.3. Set up pre-commit hooks
  - [x] 1.3.4. Configure CI (GitHub Actions) for test, lint, build

---

## Phase 2: Core File Watcher & Debouncer (TDD)

- [x] **2.1. Write File Watcher Tests** (scaffolded)
  - [x] 2.1.1. Test: Detect file changes in Go source dirs
  - [x] 2.1.2. Test: Ignore `vendor/` and hidden dirs
  - [x] 2.1.3. Test: Handle file create, write, remove events
- [x] **2.2. Implement File Watcher** (TDD-validated, all watcher tests pass)
  - [x] 2.2.1. Integrate `fsnotify` for recursive watching (TDD-validated)
  - [x] 2.2.2. Correctly skip excluded directories (TDD-validated)
  - [x] 2.2.3. Emit events to channel (TDD-validated)
- [x] **2.3. Write Debouncer Tests** (TDD-validated, all debouncer tests pass)
  - [x] 2.3.1. Test: Buffer rapid events per package
  - [x] 2.3.2. Test: Trigger only after quiet period
- [x] **2.4. Implement Debouncer** (TDD-validated, all debouncer tests pass)
  - [x] 2.4.1. Buffer and coalesce events (TDD-validated)
  - [x] 2.4.2. Trigger test run after debounce interval (TDD-validated)

---

## Phase 3: Test Runner & Output Parser (TDD)

- [x] **3.1. Write Test Runner Tests**
  - [x] 3.1.1. Test: Run `go test -json` in correct pkg
  - [x] 3.1.2. Test: Capture stdout/stderr, handle errors
  - [x] 3.1.3. Test: Handle non-JSON output (build errors)
  - [x] 3.1.4. Test: Pipe stdout/stderr for real-time output
  - [x] 3.1.5. Test: Integration with goroutine pipeline pattern
- [x] **3.2. Implement Test Runner**
  - [x] 3.2.1. Use `os/exec` to run `go test -json`
  - [x] 3.2.2. Stream output to parser through channels
  - [x] 3.2.3. Handle and log command errors properly
  - [x] 3.2.4. Support future per-test reruns with `-run=TestName`
  - [x] 3.2.5. Implement timeout and deadlock protection
    - [x] 3.2.5.1. Set appropriate test timeouts with `-timeout` flag
    - [x] 3.2.5.2. Add context with cancel for graceful termination
    - [x] 3.2.5.3. Detect hanging tests and provide useful feedback
- [ ] **3.3. Write Output Parser Tests**
  - [x] 3.3.1. Test: Parse TestEvent JSON objects from output stream
  - [x] 3.3.2. Test: Track test start/run/pass/fail/output events
  - [x] 3.3.3. Test: Extract file/line information from failure output
  - [x] 3.3.4. Test: Collect test durations and output lines
  - [x] 3.3.5. Test: Handle edge cases (build errors, test panics, timeouts)
- [x] **3.4. Implement Output Parser**
  - [x] 3.4.1. Parse TestEvent structs from JSON stream
  - [x] 3.4.2. Group events by package/test name
  - [x] 3.4.3. Extract error context and file locations
  - [x] 3.4.4. Provide structured results to UI component

*All output parser implementation tasks are now complete and validated by tests. Ready for UI or further integration.*

*Output parser implementation has begun: TestEvent struct and ParseTestEvents are in place and validated by tests.*


**Next up:**
- Begin Phase 3.3 and 3.4: Implement and test the Output Parser.
- Focus on parsing the JSON output from `go test -json`, tracking all test event types, extracting error/file context, and providing structured results for future UI or reporting integration.


---

## Phase 4: Interactive CLI UI & Controller (TDD)

- [x] **4.1. Write UI/Controller Tests**
  - [x] 4.1.1. Test: Display summary with color (ANSI)
  - [x] 4.1.2. Test: Keybindings (Enter, f, q)
  - [x] 4.1.3. Test: Filter failures mode
  - [x] 4.1.4. Test: Show code context for failed tests
  - [x] 4.1.5. Test: UI updates on each run without exit
  - [x] 4.1.6. Test: Copy failed test information to clipboard ('c' key)
  - [x] 4.1.7. Test: Interactive test selection and copying ('C' key, space for selection)
  
  > **All UI/controller test cases for MVP are implemented and passing.**
- [ ] **4.2. Implement CLI UI**
  - [x] 4.2.1. Render summary, color output, icons
  - [x] 4.2.2. Implement interactive controls (keypresses)
  - [x] 4.2.3. Display code context for failures (extract lines from source)
  - [x] 4.2.4. Channel communication between components
    > CLI UI and runner/controller now communicate via Go channels and goroutines; event-driven updates and responsive UI loop implemented.

  - [x] 4.2.5. Implement clipboard integration for test failures
  - [x] 4.2.6. Create interactive test selection mode with visual indicators
    > Interactive selection mode with visual indicators, keybindings, and clipboard integration implemented and tested.
  - [ ] 4.2.7. Full-Screen TUI Framework Integration (Bubble Tea + Lipgloss)
    1. TUI Layout Foundation (TDD)
      1.1. Layout Testing
        1.1.1. Test: 4-pane layout renders at various terminal sizes (min, max, odd sizes)
        1.1.2. Test: Each pane (header, footer, test list, details) is rendered and non-overlapping
        1.1.3. Test: Layout adapts to terminal resizing (increasing, decreasing, edge cases)
        1.1.4. Test: Handles extremely small terminal sizes gracefully (panes collapse/scroll/warning)
        1.1.5. Test: Handles extremely large terminal sizes (no empty gaps, layout stretches appropriately)
      1.2. Layout Implementation
        1.2.1. Define theme constants for colors, borders, padding
        1.2.2. Implement Lipgloss styles for each pane
        1.2.3. Compose layout using JoinVertical/JoinHorizontal
        1.2.4. Ensure model tracks width/height and passes to layout
        1.2.5. Implement minimum/maximum size logic for panes
      1.3. Window Resize Handling
        1.3.1. Listen for WindowSizeMsg and update model
        1.3.2. Recalculate pane sizes and trigger redraw
        1.3.3. Test: No visual artifacts or overlaps on rapid resize
        1.3.4. Test: Pane content is never truncated mid-character (UTF-8 safe)

    2. Header Component (Div 1)
      2.1. Header Testing
        2.1.1. Test: Logo, app name, version render at all widths
        2.1.2. Test: Stats (pass/fail/total) and refresh time update live
        2.1.3. Test: Progress bar and coverage display correct values
        2.1.4. Test: Header truncates or wraps gracefully on overflow
        2.1.5. Test: Header does not overlap with other panes
      2.2. Header Implementation
        2.2.1. Create ASCII/Unicode logo and style
        2.2.2. Implement stats area: pass/fail/total, last refresh, coverage
        2.2.3. Add progress bar (Bubble Tea progress bubble or Unicode)
        2.2.4. Implement responsive header layout (truncation, ellipsis, wrapping)
        2.2.5. Add error handling for missing stats/coverage (shows N/A or warning)

    3. Test List Panel (Div 3)
      3.1. Test List Testing
        3.1.1. Test: Tree renders icons (suite, file, test) and indentation
        3.1.2. Test: Test status (pass/fail/running/skipped) is shown with color/icon
        3.1.3. Test: Selection highlighting is visible and accessible
        3.1.4. Test: Filtering/search updates list and preserves selection
        3.1.5. Test: Handles empty test lists (shows message)
        3.1.6. Test: Handles very large test lists (performance, scrolling)
        3.1.7. Test: Handles deeply nested test suites/files
        3.1.8. Test: Handles test names with non-ASCII/emoji characters
      3.2. Test List Implementation
        3.2.1. Style tree items and status icons with Lipgloss
        3.2.2. Implement fuzzy search and filter controls
        3.2.3. Add selection mode: toggling, select all/deselect all, copying
        3.2.4. Show selection state visually (checkbox, highlight)
        3.2.5. Clipboard logic for copying selected tests
        3.2.6. Ensure selection state maps correctly to filtered/visible items
        3.2.7. Handle input edge cases (rapid keypress, invalid keys)
        3.2.8. Test: Selection/copy logic with empty/large/filtered lists

    4. Details Panel (Div 4)
      4.1. Details Panel Testing
        4.1.1. Test: Selecting a test shows details (output, error, code context)
        4.1.2. Test: Error messages, stack traces, and code context are formatted and highlighted
        4.1.3. Test: Tabs (output/source/coverage) switch content
        4.1.4. Test: Handles missing output, missing source, missing coverage
        4.1.5. Test: Handles very large outputs (scrolling, truncation)
        4.1.6. Test: Handles non-UTF8 or binary output gracefully
      4.2. Details Panel Implementation
        4.2.1. Implement tabbed interface for details
        4.2.2. Syntax highlighting for code context
        4.2.3. Scroll support for long outputs
        4.2.4. Highlight error lines, show coverage if available
        4.2.5. Show placeholder or warning if details unavailable

    5. Footer Component (Div 2)
      5.1. Footer Testing
        5.1.1. Test: Footer displays context-aware keybindings for all modes
        5.1.2. Test: Status messages (info, warning, error) display and auto-clear
        5.1.3. Test: Footer remains visible at minimal terminal height
        5.1.4. Test: Handles overflow of keybindings (truncates, scrolls, or wraps)
      5.2. Footer Implementation
        5.2.1. Dynamic keybinding bar updates by selection/mode
        5.2.2. Message queue for status messages (timeouts, severity coloring)
        5.2.3. Footer always visible and styled
        5.2.4. Graceful handling of unknown/unsupported keybindings

    6. Event System & Integration
      6.1. Event System Testing
        6.1.1. Test: File watcher/test runner events update UI in real time
        6.1.2. Test: User actions (rerun, filter, select) trigger correct updates in all panes
        6.1.3. Test: Handles rapid/frequent events without UI lag or missed updates
        6.1.4. Test: Handles event errors (invalid data, lost connection)
      6.2. Event System Implementation
        6.2.1. Define custom Bubble Tea messages for all relevant events
        6.2.2. Event handlers in model update state and trigger re-renders
        6.2.3. All panes subscribe to and reflect state changes
        6.2.4. Error handling for event routing failures

    7. Performance & Polish
      7.1. Performance Testing
        7.1.1. Test: Rendering/interaction with large test suites (hundreds/thousands)
        7.1.2. Test: Viewport/lazy loading logic prevents lag or flicker
        7.1.3. Test: UI remains responsive under heavy load
        7.1.4. Test: Handles rapid user input (keypress spamming)
      7.2. Performance Implementation
        7.2.1. Viewport rendering for large lists (Bubble Tea viewport bubble)
        7.2.2. Lazy loading for details panel
        7.2.3. Subtle transitions/animations for state changes
        7.2.4. Polish help overlay and keyboard shortcut reference
        7.2.5. Defensive coding for all known edge cases (panics, nils, out-of-bounds, etc.)

---

## Phase 5: Coverage Enhancements (TDD)

- [ ] **5.1. Coverage Visualization Improvements**
  - [ ] 5.1.1. User Testing & Feedback
    - [ ] Test: Gather data on coverage view usability with different codebase sizes
    - [ ] Test: Evaluate navigation patterns and key binding efficiency
  - [ ] 5.1.2. Performance Optimization
    - [ ] Test: Profile and benchmark coverage rendering with large codebases
    - [ ] Test: Validate memory usage optimization for large coverage files
    - [ ] Test: Validate incremental parsing and display of large coverage reports

- [ ] **5.2. Coverage Trends & History**
  - [ ] 5.2.1. Historical Comparison
    - [ ] Test: Store and retrieve historical coverage data
    - [ ] Test: Visualize coverage trends over time
    - [ ] Test: Generate comparative reports between runs
  - [ ] 5.2.2. Threshold Alerts
    - [ ] Test: Configure coverage threshold settings
    - [ ] Test: Alert on coverage regression below thresholds
    - [ ] Test: Export regression reports

- [ ] **5.3. CI/CD Integration**
  - [ ] 5.3.1. Export Formats
    - [ ] Test: Generate XML coverage reports for CI integration
    - [ ] Test: Generate JSON coverage data for dashboard consumption
    - [ ] Test: Support multiple export formats with configurable options
  - [ ] 5.3.2. Coverage Badge Generation
    - [ ] Test: Generate coverage badges for inclusion in README
    - [ ] Test: Update badges based on latest coverage data
    - [ ] Test: Support different badge styles and formats

---

## Phase 6: Concurrency & Resilience (TDD)

- [ ] **6.1. Write Pipeline/Recovery Tests**
  - [ ] 6.1.1. Test: Watcher, Debouncer, Runner, Parser, UI as goroutines with channels
  - [ ] 6.1.2. Test: Pipeline pattern (input/output channels)
  - [ ] 6.1.3. Test: Each goroutine recovers from panic and logs error
  - [ ] 6.1.4. Test: Program never crashes on test/compile errors
- [x] **6.2. Implement Concurrency Pipeline**
  - [x] 6.2.1. Implement each stage as a goroutine with channel communication
  - [x] 6.2.2. Use buffered and unbuffered channels as appropriate
  - [x] 6.2.3. Implement context passing for graceful shutdown
  - [x] 6.2.4. Implement error handling and recovery using `recover()`
- [ ] **6.3. Error Handling & Recovery**
  - [x] 6.3.1. Design error channel architecture
  - [x] 6.3.2. Add recovery handlers for each goroutine
  - [x] 6.3.3. Implement graceful shutdown and restart for individual components
  - [x] 6.3.4. Add detailed logging for errors and recovery
  - [ ] 6.3.5. Unit test panic recovery system

{{ ... }}

- [ ] **6.1. Write Config Tests**
  - [ ] 6.1.1. Test: Parse CLI flags and YAML/JSON config file
  - [ ] 6.1.2. Test: Validate config (includes/excludes, debounce, color, verbosity)
  - [ ] 6.1.3. Test: Config precedence (flags override file)
- [ ] **6.2. Implement Config Support**
  - [ ] 6.2.1. Implement CLI flags and watcher.yaml support
  - [ ] 6.2.2. Validate config at startup, error on invalid
  - [ ] 6.2.3. Mirror CLI flags in config file (viper/cobra)

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
