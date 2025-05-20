# Go Sentinel: Vitest-like CLI Roadmap

This roadmap outlines the development plan for creating a Vitest-like command-line interface for the Go Sentinel test runner. The goal is to provide a modern, user-friendly terminal output similar to Vitest, with support for watching file changes and running tests selectively.

## Overview

Go Sentinel CLI will offer:
- Beautiful, colorful test output in the terminal
- Clear indicators for passed and failed tests
- Code context display for failed tests (showing ~5 lines around the failure)
- File-based test summaries
- Watch mode for selective test rerunning on file changes
- Fast feedback loop during development

---

## Phase 1: Core CLI Output Design (TDD)

- [ ] **1.1. Design Output Format & Data Structure Tests**
  - [ ] 1.1.1. Test: Define TestResult structure with required fields for Vitest-like display
  - [ ] 1.1.2. Test: Parse go test output correctly into the structure
  - [ ] 1.1.3. Test: Handle edge cases (panics, build failures, timeouts)

- [ ] **1.2. Implement Output Format & Data Structures**
  - [ ] 1.2.1. Implement TestResult structure with all required fields
  - [ ] 1.2.2. Implement JSON test output parser for go test -json
  - [ ] 1.2.3. Add support for capturing source code context from failed tests

- [ ] **1.3. Color Scheme & Visual Elements Tests**
  - [ ] 1.3.1. Test: Define color scheme constants matching Vitest style
  - [ ] 1.3.2. Test: Generate ANSI color sequences correctly
  - [ ] 1.3.3. Test: Handle terminal capability detection
  - [ ] 1.3.4. Test: Handle emoji/icon fallbacks for different terminals

- [ ] **1.4. Implement Color & Visual Elements**
  - [ ] 1.4.1. Implement color scheme constants
  - [ ] 1.4.2. Create helpers for ANSI color output
  - [ ] 1.4.3. Implement terminal capability detection
  - [ ] 1.4.4. Add emoji/icon support with fallbacks

---

## Phase 2: Test Results Formatting & Display (TDD)

- [ ] **2.1. Test Header/Summary Design Tests**
  - [ ] 2.1.1. Test: Format test file/suite header with file path and status
  - [ ] 2.1.2. Test: Generate test run summary with pass/fail counts
  - [ ] 2.1.3. Test: Format timing information correctly
  - [ ] 2.1.4. Test: Handle multiline headers gracefully

- [ ] **2.2. Implement Test Header/Summary**
  - [ ] 2.2.1. Implement file/suite header formatter
  - [ ] 2.2.2. Create summary generator with statistics
  - [ ] 2.2.3. Implement timing formatter with appropriate precision
  - [ ] 2.2.4. Add multiline header support

- [ ] **2.3. Individual Test Result Display Tests**
  - [ ] 2.3.1. Test: Format passed tests with green check and name
  - [ ] 2.3.2. Test: Format failed tests with red X and name
  - [ ] 2.3.3. Test: Indent subtests/nested tests correctly
  - [ ] 2.3.4. Test: Handle test names with special characters

- [ ] **2.4. Implement Individual Test Result Display**
  - [ ] 2.4.1. Create passed test formatter
  - [ ] 2.4.2. Create failed test formatter
  - [ ] 2.4.3. Implement subtest/nested test indentation
  - [ ] 2.4.4. Add special character handling in test names

- [ ] **2.5. Error Context & Source Code Display Tests**
  - [ ] 2.5.1. Test: Extract 5 lines of context around error location
  - [ ] 2.5.2. Test: Highlight specific error line
  - [ ] 2.5.3. Test: Format stack traces in a readable way
  - [ ] 2.5.4. Test: Handle missing source files gracefully

- [ ] **2.6. Implement Error Context & Source Display**
  - [ ] 2.6.1. Implement source code context extractor
  - [ ] 2.6.2. Create syntax highlighting for error lines
  - [ ] 2.6.3. Implement stack trace formatter
  - [ ] 2.6.4. Add fallback handling for missing source files

---

## Phase 3: Watch Mode & Selective Testing (TDD)

- [ ] **3.1. File Change Detection Integration Tests**
  - [ ] 3.1.1. Test: Integrate with existing file watcher
  - [ ] 3.1.2. Test: Detect changes to test files
  - [ ] 3.1.3. Test: Detect changes to implementation files
  - [ ] 3.1.4. Test: Identify related test files for changed implementation files

- [ ] **3.2. Implement File Change Detection Integration**
  - [ ] 3.2.1. Connect to existing file watcher component
  - [ ] 3.2.2. Add path filtering for test files
  - [ ] 3.2.3. Create mapping between implementation and test files
  - [ ] 3.2.4. Implement change-to-test-file resolution logic

- [ ] **3.3. Selective Test Running Tests**
  - [ ] 3.3.1. Test: Run only tests from modified test files
  - [ ] 3.3.2. Test: Run related tests when implementation files change
  - [ ] 3.3.3. Test: Support running all tests on demand
  - [ ] 3.3.4. Test: Handle package-level changes correctly

- [ ] **3.4. Implement Selective Test Running**
  - [ ] 3.4.1. Create test file selector
  - [ ] 3.4.2. Implement related test finder
  - [ ] 3.4.3. Add full test suite runner
  - [ ] 3.4.4. Implement package-level dependency analysis

- [ ] **3.5. Watch Mode UI Tests**
  - [ ] 3.5.1. Test: Display watch mode status line
  - [ ] 3.5.2. Test: Handle terminal clearing between runs
  - [ ] 3.5.3. Test: Show file change notifications
  - [ ] 3.5.4. Test: Display watch mode key commands

- [ ] **3.6. Implement Watch Mode UI**
  - [ ] 3.6.1. Create watch mode status line
  - [ ] 3.6.2. Implement terminal clearing functionality
  - [ ] 3.6.3. Add file change notification display
  - [ ] 3.6.4. Implement key command display and handling

---

## Phase 4: Performance & User Experience (TDD)

- [ ] **4.1. Progress Indicators Tests**
  - [ ] 4.1.1. Test: Display spinner during test execution
  - [ ] 4.1.2. Test: Show real-time test count updates
  - [ ] 4.1.3. Test: Indicate test run progress
  - [ ] 4.1.4. Test: Handle terminal resizing during test run

- [ ] **4.2. Implement Progress Indicators**
  - [ ] 4.2.1. Create animated spinner component
  - [ ] 4.2.2. Implement real-time counter updates
  - [ ] 4.2.3. Add progress bar/indicator
  - [ ] 4.2.4. Implement terminal resize handling

- [ ] **4.3. Performance Optimization Tests**
  - [ ] 4.3.1. Test: Measure and optimize parsing performance
  - [ ] 4.3.2. Test: Optimize rendering for large test suites
  - [ ] 4.3.3. Test: Benchmark parallel vs. sequential test execution
  - [ ] 4.3.4. Test: Memory usage optimization for long-running sessions

- [ ] **4.4. Implement Performance Optimizations**
  - [ ] 4.4.1. Optimize parser for speed
  - [ ] 4.4.2. Implement lazy rendering for large test suites
  - [ ] 4.4.3. Add parallel execution support
  - [ ] 4.4.4. Implement memory leak prevention for watch mode

- [ ] **4.5. Error Recovery & Stability Tests**
  - [ ] 4.5.1. Test: Recover from test runner crashes
  - [ ] 4.5.2. Test: Handle filesystem permission errors
  - [ ] 4.5.3. Test: Recover from syntax errors in tests
  - [ ] 4.5.4. Test: Stable behavior with corrupted/inconsistent Go files

- [ ] **4.6. Implement Error Recovery & Stability**
  - [ ] 4.6.1. Add test runner crash recovery
  - [ ] 4.6.2. Implement filesystem error handling
  - [ ] 4.6.3. Create syntax error recovery mechanism
  - [ ] 4.6.4. Add corrupted file detection and handling

---

## Phase 5: CLI Options & Configuration (TDD)

- [ ] **5.1. Command Line Arguments Tests**
  - [ ] 5.1.1. Test: Parse watch flag correctly
  - [ ] 5.1.2. Test: Handle package/file patterns as arguments
  - [ ] 5.1.3. Test: Support filtering by test name pattern
  - [ ] 5.1.4. Test: Process verbosity level flags

- [ ] **5.2. Implement Command Line Arguments**
  - [ ] 5.2.1. Add watch mode flag
  - [ ] 5.2.2. Implement package/file pattern support
  - [ ] 5.2.3. Create test name pattern filtering
  - [ ] 5.2.4. Add verbosity level control

- [ ] **5.3. Configuration File Tests**
  - [ ] 5.3.1. Test: Load configuration from sentinel.config.json
  - [ ] 5.3.2. Test: Support configuration for colors, icons, formatting
  - [ ] 5.3.3. Test: Handle path inclusion/exclusion patterns
  - [ ] 5.3.4. Test: Configure watch mode behavior

- [ ] **5.4. Implement Configuration File Support**
  - [ ] 5.4.1. Create configuration file loader
  - [ ] 5.4.2. Implement visual style configuration
  - [ ] 5.4.3. Add path pattern processor
  - [ ] 5.4.4. Implement watch behavior configuration

---

## Phase 6: Integration & Documentation

- [ ] **6.1. Integration with Existing Components**
  - [ ] 6.1.1. Integrate with file watcher module
  - [ ] 6.1.2. Connect with test runner module
  - [ ] 6.1.3. Ensure compatibility with parser module
  - [ ] 6.1.4. Test complete system integration

- [ ] **6.2. Documentation**
  - [ ] 6.2.1. Update README with CLI usage instructions
  - [ ] 6.2.2. Create example configurations
  - [ ] 6.2.3. Document key commands and features
  - [ ] 6.2.4. Add screenshots of CLI output

- [ ] **6.3. Final User Experience Polish**
  - [ ] 6.3.1. Collect user feedback on early versions
  - [ ] 6.3.2. Refine visual elements based on feedback
  - [ ] 6.3.3. Optimize performance for common use cases
  - [ ] 6.3.4. Final QA testing across different environments

---

## Technologies & Dependencies

- **Color Output**: Will use Go's [github.com/fatih/color](https://github.com/fatih/color) package for ANSI colors
- **Terminal Handling**: [github.com/mattn/go-isatty](https://github.com/mattn/go-isatty) for terminal detection
- **Source Code Parsing**: Standard Go AST packages for extracting code context
- **File Watching**: Leveraging Go Sentinel's existing file watcher built on fsnotify
- **Test Execution**: Using Go Sentinel's existing test runner for go test execution and output parsing

---

This roadmap will be our source of truth for implementing the Vitest-like CLI for Go Sentinel. Each feature will be developed following TDD principles, creating failing tests first and then implementing the code to pass those tests. The roadmap is designed to be iterative, allowing early versions of the CLI to be usable while additional features are developed. 