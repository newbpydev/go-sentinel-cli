# Go-Sentinel Bubble Tea TUI Integration Plan

## Overview

This document outlines the step-by-step implementation plan to integrate the existing Go-Sentinel components with the new Bubble Tea TUI. The implementation follows TDD principles and aligns with the project roadmap.

## Current Status

- ✅ UI (Bubble Tea): A tree-based test explorer model with interactive features
- ✅ Parser: Loads JSON test results and converts them to a tree structure for display
- ✅ Runner: Executes tests with timeout protection and streaming output
- ✅ Watcher: Monitors file changes using fsnotify
- ✅ Coverage: Analyzes and visualizes test coverage

Some components are partially connected:
- ✅ UI <-> Parser: Test results are loaded from a JSON file and displayed in the TUI
- ✅ UI <-> Coverage: Integration exists for displaying coverage data

Missing connections:
- ❌ Watcher <-> Debouncer <-> Runner <-> TUI: The live file-watching and auto-test-running pipeline
- ❌ Main application controller: Coordinating all components in an event-driven flow

## Implementation Tasks

### Phase 1: Tests for TUI Controller and Events

- [ ] **1.1 Create TUI Controller Tests**
  - [ ] 1.1.1 Test: Controller initialization with all components
  - [ ] 1.1.2 Test: File event to test run pipeline
  - [ ] 1.1.3 Test: Test output parsing and UI model updates
  - [ ] 1.1.4 Test: User-triggered test runs
  - [ ] 1.1.5 Test: Graceful shutdown and cleanup

- [x] **1.2 Create UI Event Tests**
  - [x] 1.2.1 Test: TestResultsMsg handling
  - [x] 1.2.2 Test: TestsStartedMsg handling
  - [x] 1.2.3 Test: TestsCompletedMsg handling
  - [x] 1.2.4 Test: RunTestsMsg handling
  - [x] 1.2.5 Test: FileChangedMsg handling

### Phase 2: Define Event Types and Messages

- [x] **2.1 Create Event Definition File**
  - [x] 2.1.1 Define TestResultsMsg type
  - [x] 2.1.2 Define FileChangedMsg type
  - [x] 2.1.3 Define TestsStartedMsg type
  - [x] 2.1.4 Define TestsCompletedMsg type
  - [x] 2.1.5 Define RunTestsMsg type

- [x] **2.2 Update TUI Model Update Function**
  - [x] 2.2.1 Add TestResultsMsg handler
  - [x] 2.2.2 Add TestsStartedMsg handler
  - [x] 2.2.3 Add TestsCompletedMsg handler
  - [x] 2.2.4 Add RunTestsMsg handler
  - [x] 2.2.5 Add FileChangedMsg handler

### Phase 3: Implement TUI Controller

- [x] **3.1 Create App Structure**
  - [x] 3.1.1 Define the App struct with component fields
  - [x] 3.1.2 Implement NewApp constructor
  - [x] 3.1.3 Implement cleanup method

- [x] **3.2 Implement Event Loops**
  - [x] 3.2.1 Implement watcherLoop for file changes
  - [x] 3.2.2 Implement testResultsLoop for parsing output
  - [x] 3.2.3 Connect debouncer to file events
  - [x] 3.2.4 Add UI update commands via tea.Program.Send

- [x] **3.3 Implement App Management Methods**
  - [x] 3.3.1 Implement Start method
  - [x] 3.3.2 Implement Stop method
  - [x] 3.3.3 Add package-to-file mapping
  - [x] 3.3.4 Implement test filtering by file

### Phase 4: Update Main Entry Point

- [x] **4.1 Refactor main.go**
  - [x] 4.1.1 Create new App instance
  - [x] 4.1.2 Add flag parsing for customization
  - [x] 4.1.3 Handle proper error reporting
  - [x] 4.1.4 Add graceful shutdown

### Phase 5: Command Handling

- [x] **5.1 Implement User Commands**
  - [x] 5.1.1 Add "run all tests" command
  - [x] 5.1.2 Add "run selected test" command
  - [x] 5.1.3 Add "run package tests" command
  - [x] 5.1.4 Add "toggle watch mode" command

- [x] **5.2 Implement Status Feedback**
  - [x] 5.2.1 Add test running indicator
  - [x] 5.2.2 Add file watch status indicator
  - [x] 5.2.3 Implement error notification display

### Phase 6: Integration Tests

- [ ] **6.1 Create End-to-End Tests**
  - [ ] 6.1.1 Test full pipeline from file change to UI update
  - [ ] 6.1.2 Test user command execution
  - [ ] 6.1.3 Test coverage generation and visualization
  - [ ] 6.1.4 Test proper resource cleanup

## Testing Guidelines

1. Each component should have isolated unit tests
2. Integration tests should verify proper component interaction
3. Use mock implementations where appropriate to isolate test scope
4. Test error conditions and edge cases thoroughly
5. Follow table-driven test patterns for similar test cases

## Implementation Notes

- Follow the existing style and patterns in the codebase
- Use contexts for cancellation and timeout control
- Maintain a clear separation of concerns between components
- Document all public APIs with proper comments
- Consider performance implications, especially for watch events

This implementation plan aligns with the project roadmap and focuses on completing the core functionality while maintaining the existing TDD approach.
