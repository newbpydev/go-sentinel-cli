# Go Sentinel: Vitest CLI v2 Roadmap

This roadmap outlines the comprehensive development plan for building a Vitest-like command-line interface for the Go Sentinel test runner from scratch. The goal is to provide a modern, user-friendly terminal output that displays real-time test results with clear indicators for passed and failed tests, detailed error reporting, and a cohesive summary view.

## Progress Summary

- Completed Phase 1: Core Architecture & Data Structures (100%)
- Completed Phase 2: Test Suite Display (100%) 
- Completed Phase 1-D: Demonstration of Core Architecture (100%)
- Completed Phase 2-D: Demonstration of Test Suite Display (100%)
- Completed Phase 3: Failed Test Details Section (100%)
- Completed Phase 3-D: Demonstration of Failed Test Details (100%)
- Current code coverage: 70.5%
- Next phase: Phase 4 - Real-time Processing & Summary

## Development Approach

We will follow a prototype-first approach where:
1. Each major feature is first implemented in a demo/prototype version
2. The prototype is validated against actual Vitest output for visual and functional parity
3. After validation, the feature is incorporated into the main application
4. Each phase has a corresponding demonstration phase (marked as Phase X-D)

## Overview

Go Sentinel CLI v2 will offer:
- Real-time, colorful test output in the terminal
- Clear file-based test suite summaries with memory usage metrics
- Detailed error reporting with source code context (showing ~5 lines around failures)
- Distinctive "Failed Tests" section with comprehensive error details
- Watch mode for selective test rerunning on file changes
- Fast feedback loop during development

---

## Phase 1: Core Architecture & Data Structures (TDD)

- [x] **1.1. Core Data Structure Design Tests**
  - [x] 1.1.1. Test: Define TestResult structure with required fields for Vitest-like display
  - [x] 1.1.2. Test: Define TestSuite structure for organizing tests by file
  - [x] 1.1.3. Test: Define FailedTestDetail structure for detailed error reporting
  - [x] 1.1.4. Test: Parse go test output correctly into these structures
  - [x] 1.1.5. Test: Handle edge cases (panics, build failures, timeouts)

- [x] **1.2. Implement Core Data Structures**
  - [x] 1.2.1. Implement TestResult structure with all required fields
  - [x] 1.2.2. Implement TestSuite structure with file path, test count, and duration
  - [x] 1.2.3. Implement FailedTestDetail with error location and context
  - [x] 1.2.4. Implement JSON test output parser for go test -json
  - [x] 1.2.5. Add support for capturing source code context from failed tests

- [x] **1.3. Terminal & Color Handling Tests**
  - [x] 1.3.1. Test: Define color scheme constants matching Vitest style
  - [x] 1.3.2. Test: Generate ANSI color sequences correctly
  - [x] 1.3.3. Test: Handle terminal capability detection
  - [x] 1.3.4. Test: Handle emoji/icon fallbacks for different terminals

- [x] **1.4. Implement Terminal & Color Handling**
  - [x] 1.4.1. Implement color scheme constants for consistency
  - [x] 1.4.2. Create helpers for ANSI color output
  - [x] 1.4.3. Implement terminal capability detection
  - [x] 1.4.4. Add emoji/icon support with fallbacks

## Phase 1-D: Demonstration of Core Architecture

- [x] **1-D.1. Create Core Demo Application**
  - [x] 1-D.1.1. Implement minimal CLI to run basic tests
  - [x] 1-D.1.2. Add test cases that exercise data structures
  - [x] 1-D.1.3. Output raw parsed results to validate data structure correctness

- [x] **1-D.2. Validate Core Architecture**
  - [x] 1-D.2.1. Verify test output is correctly parsed into data structures
  - [x] 1-D.2.2. Validate terminal color support detection
  - [x] 1-D.2.3. Confirm correct emoji/icon display based on terminal capabilities
  - [x] 1-D.2.4. Document any discrepancies or issues found during validation

---

## Phase 2: Test Suite Display (TDD)

- [x] **2.1. Test Suite Header Design Tests**
  - [x] 2.1.1. Test: Format test file path with colorized file name
  - [x] 2.1.2. Test: Display test counts with failed test highlighting
  - [x] 2.1.3. Test: Show accurate test duration with proper formatting
  - [x] 2.1.4. Test: Include memory usage information
  - [x] 2.1.5. Test: Handle multiline headers gracefully

- [x] **2.2. Implement Test Suite Header**
  - [x] 2.2.1. Create function to format file paths consistently
  - [x] 2.2.2. Implement colorized test count display
  - [x] 2.2.3. Add duration formatter with MS precision
  - [x] 2.2.4. Add heap memory usage tracking and display
  - [x] 2.2.5. Handle edge cases in header formatting

- [x] **2.3. Individual Test Result Display Tests**
  - [x] 2.3.1. Test: Format passed tests with green check and name
  - [x] 2.3.2. Test: Format failed tests with red X and name
  - [x] 2.3.3. Test: Indent subtests/nested tests correctly
  - [x] 2.3.4. Test: Handle test names with special characters
  - [x] 2.3.5. Test: Show appropriate error messages for failed tests

- [x] **2.4. Implement Individual Test Result Display**
  - [x] 2.4.1. Create passed test formatter with appropriate icons
  - [x] 2.4.2. Create failed test formatter with error message
  - [x] 2.4.3. Implement subtest/nested test indentation
  - [x] 2.4.4. Add special character handling in test names
  - [x] 2.4.5. Implement error message formatting

- [x] **2.5. Test Suite Collapsing/Expanding Tests**
  - [x] 2.5.1. Test: Collapse passing test suites by default
  - [x] 2.5.2. Test: Expand test suites with failing tests
  - [x] 2.5.3. Test: Properly indent and format nested tests
  - [x] 2.5.4. Test: Handle edge cases like empty suites or all skipped tests

- [x] **2.6. Implement Suite Collapsing/Expanding**
  - [x] 2.6.1. Create logic to determine if suite should be collapsed
  - [x] 2.6.2. Implement nested test indentation formatting
  - [x] 2.6.3. Add icons for test status indication
  - [x] 2.6.4. Ensure proper padding and spacing for all components

## Phase 2-D: Demonstration of Test Suite Display

- [x] **2-D.1. Create Test Suite Display Demo**
  - [x] 2-D.1.1. Implement CLI command to run and display test suites
  - [x] 2-D.1.2. Create sample test suites with various test outcomes (pass, fail, skip)
  - [x] 2-D.1.3. Ensure demo accurately displays file paths with correct coloring
  - [x] 2-D.1.4. Validate test count display with proper highlighting of failed tests
  - [x] 2-D.1.5. Verify duration and memory usage formatting

- [x] **2-D.2. Visual Validation Against Vitest**
  - [x] 2-D.2.1. Compare output side-by-side with Vitest screenshot
  - [x] 2-D.2.2. Verify color scheme matches Vitest style (green checks, red Xs)
  - [x] 2-D.2.3. Confirm indentation and spacing match Vitest output
  - [x] 2-D.2.4. Ensure collapse/expand behavior matches Vitest expectations
  - [x] 2-D.2.5. Document any visual discrepancies and adjust as needed

---

## Phase 3: Failed Test Details Section (TDD)

- [x] **3.1. Failed Test Section Header Tests**
  - [x] 3.1.1. Test: Create distinctive "Failed Tests" header
  - [x] 3.1.2. Test: Display accurate count of failed tests
  - [x] 3.1.3. Test: Format header with appropriate styling
  - [x] 3.1.4. Test: Handle case when there are no failed tests

- [x] **3.2. Implement Failed Test Section Header**
  - [x] 3.2.1. Create header formatter function
  - [x] 3.2.2. Add counter for failed tests
  - [x] 3.2.3. Implement visual separation between sections
  - [x] 3.2.4. Add conditional rendering based on test results

- [x] **3.3. Individual Test Failure Display Tests**
  - [x] 3.3.1. Test: Show file path and failing test name
  - [x] 3.3.2. Test: Display error type and error message
  - [x] 3.3.3. Test: Show code location with line numbers
  - [x] 3.3.4. Test: Include 5 lines of code context with line highlighting
  - [x] 3.3.5. Test: Format the error section with appropriate spacing

- [x] **3.4. Implement Individual Test Failure Display**
  - [x] 3.4.1. Create test failure formatter
  - [x] 3.4.2. Add error message and type display
  - [x] 3.4.3. Implement source code extraction with proper context
  - [x] 3.4.4. Add line highlighting for the exact error location
  - [x] 3.4.5. Create consistent spacing and formatting

- [x] **3.5. Error Context & Source Code Display Tests**
  - [x] 3.5.1. Test: Extract 5 lines of context around error location
  - [x] 3.5.2. Test: Highlight specific error line
  - [x] 3.5.3. Test: Format stack traces in a readable way
  - [x] 3.5.4. Test: Handle missing source files gracefully
  - [x] 3.5.5. Test: Format TypeErrors and other common error types

- [x] **3.6. Implement Error Context & Source Display**
  - [x] 3.6.1. Implement source code context extractor
  - [x] 3.6.2. Create syntax highlighting for error lines
  - [x] 3.6.3. Implement stack trace formatter
  - [x] 3.6.4. Add fallback handling for missing source files
  - [x] 3.6.5. Create special formatting for common error types

## Phase 3-D: Demonstration of Failed Test Details

- [x] **3-D.1. Create Failed Test Details Demo**
  - [x] 3-D.1.1. Develop test suite with various failure types
  - [x] 3-D.1.2. Implement CLI command to display detailed failure information
  - [x] 3-D.1.3. Validate error message formatting against Vitest style
  - [x] 3-D.1.4. Verify source code context display with line highlighting

- [x] **3-D.2. Validate Error Reporting**
  - [x] 3-D.2.1. Compare failed test display with Vitest screenshot
  - [x] 3-D.2.2. Verify stack trace formatting is readable and helpful
  - [x] 3-D.2.3. Confirm error types are displayed correctly
  - [x] 3-D.2.4. Test with various error scenarios (assertion errors, panics, timeouts)
  - [x] 3-D.2.5. Document any formatting improvements needed

---

## Phase 4: Real-time Processing & Summary (TDD)

- [ ] **4.1. Stream Processing Tests**
  - [ ] 4.1.1. Test: Process test output as it arrives
  - [ ] 4.1.2. Test: Update test suite display in real-time
  - [ ] 4.1.3. Test: Handle partial test results correctly
  - [ ] 4.1.4. Test: Manage concurrent updates from multiple test packages
  - [ ] 4.1.5. Test: Display progress indicators during test execution

- [ ] **4.2. Implement Stream Processing**
  - [ ] 4.2.1. Create streaming JSON parser for test output
  - [ ] 4.2.2. Implement incremental test result updates
  - [ ] 4.2.3. Add buffering for incomplete test data
  - [ ] 4.2.4. Create concurrent update handler
  - [ ] 4.2.5. Implement spinner or progress indicator

- [ ] **4.3. Summary Section Tests**
  - [ ] 4.3.1. Test: Display overall test statistics (passed/failed/total)
  - [ ] 4.3.2. Test: Show test run duration information
  - [ ] 4.3.3. Test: Format timing information clearly
  - [ ] 4.3.4. Test: Handle different test result scenarios

- [ ] **4.4. Implement Summary Section**
  - [ ] 4.4.1. Create summary generator with statistics
  - [ ] 4.4.2. Implement timing information display
  - [ ] 4.4.3. Add colorized status indicators
  - [ ] 4.4.4. Handle edge cases (all tests passed, all failed, etc.)

- [ ] **4.5. Progress Indicators Tests**
  - [ ] 4.5.1. Test: Display spinner during test execution
  - [ ] 4.5.2. Test: Show real-time test count updates
  - [ ] 4.5.3. Test: Indicate test run progress
  - [ ] 4.5.4. Test: Handle terminal resizing during test run

- [ ] **4.6. Implement Progress Indicators**
  - [ ] 4.6.1. Create animated spinner component
  - [ ] 4.6.2. Implement real-time counter updates
  - [ ] 4.6.3. Add progress bar/indicator
  - [ ] 4.6.4. Implement terminal resize handling

## Phase 4-D: Demonstration of Real-time Processing

- [ ] **4-D.1. Create Real-time Processing Demo**
  - [ ] 4-D.1.1. Implement interactive CLI with real-time updates
  - [ ] 4-D.1.2. Run test suites with varying execution times
  - [ ] 4-D.1.3. Validate progress indicators and spinners
  - [ ] 4-D.1.4. Test with large test suites to verify performance

- [ ] **4-D.2. Validate User Experience**
  - [ ] 4-D.2.1. Compare real-time updates with Vitest behavior
  - [ ] 4-D.2.2. Verify summary section accuracy and formatting
  - [ ] 4-D.2.3. Assess readability and clarity of real-time information
  - [ ] 4-D.2.4. Document any UX improvements needed

---

## Phase 5: Watch Mode & Integration (TDD)

- [ ] **5.1. File Change Detection Tests**
  - [ ] 5.1.1. Test: Detect changes to test files
  - [ ] 5.1.2. Test: Detect changes to implementation files
  - [ ] 5.1.3. Test: Identify related test files for changed implementation files
  - [ ] 5.1.4. Test: Handle file system events properly

- [ ] **5.2. Implement File Change Detection**
  - [ ] 5.2.1. Implement file watcher component
  - [ ] 5.2.2. Add path filtering for test files
  - [ ] 5.2.3. Create mapping between implementation and test files
  - [ ] 5.2.4. Implement change-to-test-file resolution logic

- [ ] **5.3. Selective Test Running Tests**
  - [ ] 5.3.1. Test: Run only tests from modified test files
  - [ ] 5.3.2. Test: Run related tests when implementation files change
  - [ ] 5.3.3. Test: Support running all tests on demand
  - [ ] 5.3.4. Test: Handle package-level changes correctly

- [ ] **5.4. Implement Selective Test Running**
  - [ ] 5.4.1. Create test file selector
  - [ ] 5.4.2. Implement related test finder
  - [ ] 5.4.3. Add full test suite runner
  - [ ] 5.4.4. Implement package-level dependency analysis

- [ ] **5.5. Watch Mode UI Tests**
  - [ ] 5.5.1. Test: Display watch mode status line
  - [ ] 5.5.2. Test: Handle terminal clearing between runs
  - [ ] 5.5.3. Test: Show file change notifications
  - [ ] 5.5.4. Test: Display watch mode key commands

- [ ] **5.6. Implement Watch Mode UI**
  - [ ] 5.6.1. Create watch mode status line
  - [ ] 5.6.2. Implement terminal clearing functionality
  - [ ] 5.6.3. Add file change notification display
  - [ ] 5.6.4. Implement key command display and handling

## Phase 5-D: Demonstration of Watch Mode

- [ ] **5-D.1. Create Watch Mode Demo**
  - [ ] 5-D.1.1. Implement interactive watch mode CLI
  - [ ] 5-D.1.2. Develop test cases with files that can be modified
  - [ ] 5-D.1.3. Create demonstration script that modifies files
  - [ ] 5-D.1.4. Validate selective test running behavior

- [ ] **5-D.2. Validate Developer Experience**
  - [ ] 5-D.2.1. Assess file change detection performance
  - [ ] 5-D.2.2. Verify accuracy of related test identification
  - [ ] 5-D.2.3. Evaluate UX of watch mode interface
  - [ ] 5-D.2.4. Compare with Vitest watch mode behavior
  - [ ] 5-D.2.5. Document any DX improvements needed

---

## Phase 6: Performance & Error Handling (TDD)

- [ ] **6.1. Performance Optimization Tests**
  - [ ] 6.1.1. Test: Measure and optimize parsing performance
  - [ ] 6.1.2. Test: Optimize rendering for large test suites
  - [ ] 6.1.3. Test: Benchmark parallel vs. sequential test execution
  - [ ] 6.1.4. Test: Memory usage optimization for long-running sessions

- [ ] **6.2. Implement Performance Optimizations**
  - [ ] 6.2.1. Optimize parser for speed
  - [ ] 6.2.2. Implement lazy rendering for large test suites
  - [ ] 6.2.3. Add parallel execution support
  - [ ] 6.2.4. Implement memory leak prevention for watch mode

- [ ] **6.3. Error Recovery & Stability Tests**
  - [ ] 6.3.1. Test: Recover from test runner crashes
  - [ ] 6.3.2. Test: Handle filesystem permission errors
  - [ ] 6.3.3. Test: Recover from syntax errors in tests
  - [ ] 6.3.4. Test: Stable behavior with corrupted/inconsistent Go files

- [ ] **6.4. Implement Error Recovery & Stability**
  - [ ] 6.4.1. Add test runner crash recovery
  - [ ] 6.4.2. Implement filesystem error handling
  - [ ] 6.4.3. Create syntax error recovery mechanism
  - [ ] 6.4.4. Add corrupted file detection and handling

## Phase 6-D: Demonstration of Performance & Stability

- [ ] **6-D.1. Create Performance & Stability Demo**
  - [ ] 6-D.1.1. Develop benchmark suite with large number of tests
  - [ ] 6-D.1.2. Implement error simulation and recovery demonstrations
  - [ ] 6-D.1.3. Create long-running test to validate memory usage
  - [ ] 6-D.1.4. Compare performance with and without optimizations

- [ ] **6-D.2. Validate Production Readiness**
  - [ ] 6-D.2.1. Measure performance metrics against baseline
  - [ ] 6-D.2.2. Verify stability under error conditions
  - [ ] 6-D.2.3. Assess memory usage over extended runtime
  - [ ] 6-D.2.4. Document performance characteristics and limits

---

## Phase 7: CLI Options & Configuration (TDD)

- [ ] **7.1. Command Line Arguments Tests**
  - [ ] 7.1.1. Test: Parse watch flag correctly
  - [ ] 7.1.2. Test: Handle package/file patterns as arguments
  - [ ] 7.1.3. Test: Support filtering by test name pattern
  - [ ] 7.1.4. Test: Process verbosity level flags

- [ ] **7.2. Implement Command Line Arguments**
  - [ ] 7.2.1. Add watch mode flag
  - [ ] 7.2.2. Implement package/file pattern support
  - [ ] 7.2.3. Create test name pattern filtering
  - [ ] 7.2.4. Add verbosity level control

- [ ] **7.3. Configuration File Tests**
  - [ ] 7.3.1. Test: Load configuration from sentinel.config.json
  - [ ] 7.3.2. Test: Support configuration for colors, icons, formatting
  - [ ] 7.3.3. Test: Handle path inclusion/exclusion patterns
  - [ ] 7.3.4. Test: Configure watch mode behavior

- [ ] **7.4. Implement Configuration File Support**
  - [ ] 7.4.1. Create configuration file loader
  - [ ] 7.4.2. Implement visual style configuration
  - [ ] 7.4.3. Add path pattern processor
  - [ ] 7.4.4. Implement watch behavior configuration

## Phase 7-D: Demonstration of CLI Options & Configuration

- [ ] **7-D.1. Create CLI Options & Configuration Demo**
  - [ ] 7-D.1.1. Implement CLI with all supported arguments
  - [ ] 7-D.1.2. Create sample configuration files
  - [ ] 7-D.1.3. Develop demonstration script showing various CLI options
  - [ ] 7-D.1.4. Validate configuration file loading and precedence

- [ ] **7-D.2. Validate User Configuration Experience**
  - [ ] 7-D.2.1. Assess CLI argument usability
  - [ ] 7-D.2.2. Verify configuration file documentation clarity
  - [ ] 7-D.2.3. Test configuration with various terminal types
  - [ ] 7-D.2.4. Document configuration recommendations

---

## Phase 8: Integration & Final Implementation

- [ ] **8.1. Main Application Integration**
  - [ ] 8.1.1. Merge all validated components into main application
  - [ ] 8.1.2. Ensure consistent behavior between demo and production
  - [ ] 8.1.3. Implement any remaining edge cases identified during demos
  - [ ] 8.1.4. Resolve any integration issues between components

- [ ] **8.2. Final Documentation**
  - [ ] 8.2.1. Create comprehensive README with CLI usage instructions
  - [ ] 8.2.2. Document example configurations
  - [ ] 8.2.3. Document key commands and features
  - [ ] 8.2.4. Add screenshots of CLI output
  - [ ] 8.2.5. Create examples for common use cases

- [ ] **8.3. Final Testing & Validation**
  - [ ] 8.3.1. Conduct end-to-end testing of all features
  - [ ] 8.3.2. Verify cross-platform support (Windows, macOS, Linux)
  - [ ] 8.3.3. Validate performance with large codebases
  - [ ] 8.3.4. Collect and incorporate user feedback
  - [ ] 8.3.5. Final QA testing across different environments

---

## Implementation Details

### Display Format Requirements

Based on the screenshots, the CLI display should:

1. **Test Suite Display:**
   - Show file path with clear formatting (e.g., `test/websocket.test.ts`)
   - Display test count with pass/fail information (e.g., `(8 tests | 8 failed)`)
   - Show execution time with millisecond precision (e.g., `21ms`)
   - Include memory usage information (e.g., `32 MB heap used`)
   - Use green for passing tests and red for failing tests
   - Properly indent and format test names
   - Show error messages for failing tests

2. **Failed Test Details:**
   - Display a distinctive "Failed Tests" header with count
   - Show file path and failing test name
   - Display error type and message clearly
   - Include line numbers for error location
   - Show approximately 5 lines of code around the error
   - Highlight the exact line causing the error
   - Use consistent spacing and formatting

3. **Summary Section:**
   - Show total test files passed/failed
   - Display total tests passed/failed
   - Include test run start time
   - Show total test run duration
   - Break down duration by phase (transform, setup, etc.)

### Core Components Needed

1. **Data Processing:**
   - Go test output parser (JSON format)
   - Test result aggregator
   - Source code extractor

2. **Display Components:**
   - Terminal color and formatting manager
   - Test suite renderer
   - Failed test detail renderer
   - Summary renderer
   - Progress indicator

3. **Real-time Functionality:**
   - File watcher
   - Incremental test runner
   - Real-time display updater

4. **Configuration:**
   - CLI argument parser
   - Configuration file loader
   - Default settings provider

---

This roadmap will guide the implementation of the Vitest-like CLI for Go Sentinel. Each feature will be developed following Test-Driven Development principles, creating failing tests first and then implementing the code to pass those tests. The roadmap is designed to be iterative, allowing early versions of the CLI to be usable while additional features are developed. 