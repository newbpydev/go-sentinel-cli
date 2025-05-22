# Go Sentinel CLI Revamp Roadmap

This roadmap outlines the plan for revamping the Go Sentinel CLI to match the desired output format. The goal is to provide a modern, Vitest-like CLI interface with clear suite information, detailed test failure reporting, and a comprehensive summary.

## Overview

The revamped CLI will display information in the following order:
1. **Suite Information** - Test suite details with pass/fail counts and memory usage
2. **Failed Test Details** - Detailed information about failed tests with code context
3. **Summary** - Overall test run statistics and timing information

The CLI will process test output in real-time as it streams from the `go test` command.

---

## Phase 1: Test Suite Display (TDD)

- [ ] **1.1. Test Suite Header Design Tests**
  - [ ] 1.1.1. Test: Format test file path with colorized file name
  - [ ] 1.1.2. Test: Display test counts with failed test highlighting
  - [ ] 1.1.3. Test: Show accurate test duration with proper formatting
  - [ ] 1.1.4. Test: Include memory usage information

- [ ] **1.2. Implement Test Suite Header**
  - [ ] 1.2.1. Create function to format file paths consistently
  - [ ] 1.2.2. Implement colorized test count display
  - [ ] 1.2.3. Add duration formatter with MS precision
  - [ ] 1.2.4. Add heap memory usage tracking and display

- [ ] **1.3. Test Suite Collapsing/Expanding Tests**
  - [ ] 1.3.1. Test: Collapse passing test suites by default
  - [ ] 1.3.2. Test: Expand test suites with failing tests
  - [ ] 1.3.3. Test: Properly indent and format nested tests
  - [ ] 1.3.4. Test: Handle edge cases like empty suites or all skipped tests

- [ ] **1.4. Implement Suite Collapsing/Expanding**
  - [ ] 1.4.1. Create logic to determine if suite should be collapsed
  - [ ] 1.4.2. Implement nested test indentation formatting
  - [ ] 1.4.3. Add icons for test status indication
  - [ ] 1.4.4. Ensure proper padding and spacing for all components

---

## Phase 2: Failed Test Details (TDD)

- [ ] **2.1. Failed Test Section Header Tests**
  - [ ] 2.1.1. Test: Create distinctive "Failed Tests" header
  - [ ] 2.1.2. Test: Display accurate count of failed tests
  - [ ] 2.1.3. Test: Format header with appropriate styling

- [ ] **2.2. Implement Failed Test Section Header**
  - [ ] 2.2.1. Create header formatter function
  - [ ] 2.2.2. Add counter for failed tests
  - [ ] 2.2.3. Implement visual separation between sections

- [ ] **2.3. Individual Test Failure Display Tests**
  - [ ] 2.3.1. Test: Show file path and failing test name
  - [ ] 2.3.2. Test: Display error type and error message
  - [ ] 2.3.3. Test: Show code location with line numbers
  - [ ] 2.3.4. Test: Include 5 lines of code context with line highlighting

- [ ] **2.4. Implement Individual Test Failure Display**
  - [ ] 2.4.1. Create test failure formatter
  - [ ] 2.4.2. Add error message and type display
  - [ ] 2.4.3. Implement source code extraction with proper context
  - [ ] 2.4.4. Add line highlighting for the exact error location

- [ ] **2.5. Failure Code Context Tests**
  - [ ] 2.5.1. Test: Extract code context from source files
  - [ ] 2.5.2. Test: Format line numbers with proper alignment
  - [ ] 2.5.3. Test: Highlight the failing line with color
  - [ ] 2.5.4. Test: Handle missing source files gracefully

- [ ] **2.6. Implement Failure Code Context**
  - [ ] 2.6.1. Create source code reader with line number tracking
  - [ ] 2.6.2. Implement line number padding and alignment
  - [ ] 2.6.3. Add syntax highlighting for error line
  - [ ] 2.6.4. Create fallback for missing source files

---

## Phase 3: Real-time Processing (TDD)

- [ ] **3.1. Stream Processing Tests**
  - [ ] 3.1.1. Test: Process test output as it arrives
  - [ ] 3.1.2. Test: Update test suite display in real-time
  - [ ] 3.1.3. Test: Handle partial test results correctly
  - [ ] 3.1.4. Test: Manage concurrent updates from multiple test packages

- [ ] **3.2. Implement Stream Processing**
  - [ ] 3.2.1. Create streaming JSON parser for test output
  - [ ] 3.2.2. Implement incremental test result updates
  - [ ] 3.2.3. Add buffering for incomplete test data
  - [ ] 3.2.4. Create concurrent update handler

- [ ] **3.3. Progressive Rendering Tests**
  - [ ] 3.3.1. Test: Render suite headers as they become available
  - [ ] 3.3.2. Test: Update test status in real-time
  - [ ] 3.3.3. Test: Defer detailed failure rendering until test completion
  - [ ] 3.3.4. Test: Show running progress indicator

- [ ] **3.4. Implement Progressive Rendering**
  - [ ] 3.4.1. Create incremental suite renderer
  - [ ] 3.4.2. Implement test status updater
  - [ ] 3.4.3. Add deferred failure collector
  - [ ] 3.4.4. Create running indicator with test counts

---

## Phase 4: UI Refinement & Integration (TDD)

- [ ] **4.1. Color Scheme & Style Tests**
  - [ ] 4.1.1. Test: Define consistent color scheme matching screenshots
  - [ ] 4.1.2. Test: Create proper spacing and alignment
  - [ ] 4.1.3. Test: Handle terminal width constraints
  - [ ] 4.1.4. Test: Support different terminal types

- [ ] **4.2. Implement Color Scheme & Style**
  - [ ] 4.2.1. Define color constants for consistent usage
  - [ ] 4.2.2. Create style helper functions
  - [ ] 4.2.3. Implement terminal width detection
  - [ ] 4.2.4. Add terminal type detection and fallbacks

- [ ] **4.3. Interaction Model Tests**
  - [ ] 4.3.1. Test: Support terminal clearing between runs
  - [ ] 4.3.2. Test: Handle keyboard inputs for control
  - [ ] 4.3.3. Test: Support filtering and focusing tests
  - [ ] 4.3.4. Test: Allow expanding/collapsing test sections

- [ ] **4.4. Implement Interaction Model**
  - [ ] 4.4.1. Create terminal clearing function
  - [ ] 4.4.2. Add keyboard input handler
  - [ ] 4.4.3. Implement test filtering mechanism
  - [ ] 4.4.4. Add section expansion controls

---

## Phase 5: Integration & Performance (TDD)

- [ ] **5.1. CLI Integration Tests**
  - [ ] 5.1.1. Test: Combine all components into a unified output
  - [ ] 5.1.2. Test: Maintain correct ordering of sections
  - [ ] 5.1.3. Test: Ensure all sections update appropriately
  - [ ] 5.1.4. Test: Verify output matches desired screenshots

- [ ] **5.2. Implement CLI Integration**
  - [ ] 5.2.1. Create main CLI renderer
  - [ ] 5.2.2. Implement section ordering logic
  - [ ] 5.2.3. Add section update coordinator
  - [ ] 5.2.4. Final styling adjustments

- [ ] **5.3. Performance Optimization Tests**
  - [ ] 5.3.1. Test: Measure and optimize rendering performance
  - [ ] 5.3.2. Test: Reduce memory usage for large test suites
  - [ ] 5.3.3. Test: Minimize CPU usage during idle periods
  - [ ] 5.3.4. Test: Handle very large test output efficiently

- [ ] **5.4. Implement Performance Optimizations**
  - [ ] 5.4.1. Add rendering performance optimizations
  - [ ] 5.4.2. Implement memory usage improvements
  - [ ] 5.4.3. Create idle mode optimizations
  - [ ] 5.4.4. Add large output handling mechanisms

---

## Phase 6: Testing & Documentation

- [ ] **6.1. Comprehensive Testing**
  - [ ] 6.1.1. Add unit tests for all new components
  - [ ] 6.1.2. Create integration tests for end-to-end verification
  - [ ] 6.1.3. Test on different terminal types and sizes
  - [ ] 6.1.4. Verify handling of edge cases and error conditions

- [ ] **6.2. Documentation**
  - [ ] 6.2.1. Update code documentation for all new functions
  - [ ] 6.2.2. Create usage examples in README
  - [ ] 6.2.3. Document keyboard shortcuts and features
  - [ ] 6.2.4. Add screenshots of new CLI output

---

## Detailed Implementation Notes

### Suite Display Format

Based on the screenshots, the suite display should:
1. Show the file path in a readable format (e.g., `test/websocket.test.ts`)
2. Display test count information (e.g., `(8 tests | 8 failed)`)
3. Show execution time (e.g., `21ms`)
4. Include memory usage information (e.g., `32 MB heap used`)
5. Use color coding for pass/fail status
6. Properly indent and format test names 
7. For failing tests, include error messages with proper indentation

### Failed Test Detail Format

The failed test detail section should:
1. Display a distinctive "Failed Tests" header with count
2. For each failed test:
   - Show file path and test name
   - Display error type and message
   - Include line numbers for the error location
   - Show ~5 lines of code around the error
   - Highlight the exact line where the error occurred
   - Use proper indentation and formatting

### Utility Functions Needed

1. `formatFilePath(path string) string` - Format file paths consistently
2. `formatTestName(name string) string` - Format test names for display
3. `extractCodeContext(file string, line int) string` - Extract code around an error
4. `formatLineNumber(num int) string` - Format line numbers with consistent padding
5. `measureHeapUsage() string` - Measure and format heap memory usage
6. `colorizeStatus(status TestStatus, text string) string` - Apply color based on test status

### Data Structure Updates

1. Enhance `TestResult` to track code context and error details
2. Update `TestSuite` to support collapsing/expanding
3. Create a dedicated `FailedTestDetail` structure for the failed test section

---

This roadmap will guide the implementation of the revamped CLI interface. Each phase follows TDD principles, ensuring that tests are written before implementation. The final result will match the desired output format shown in the screenshots. 