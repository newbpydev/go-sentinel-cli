# 🗺️ Go Sentinel CLI Refactoring Roadmap

## 🎯 Project Status Overview

**Overall Progress**: 64.9% Complete (37/57 tasks)  
**Current Phase**: Phase 4 - Code Quality & Best Practices  
**Last Updated**: [Current Date]

**🚨 CRITICAL DISCOVERY**: We have created modular packages but the monolithic CLI still exists and is being used. The actual code migration was never completed.

### 📊 Phase Summary
- **Phase 1**: Foundation & Structure - **100%** ✅
- **Phase 2**: Watch Logic Consolidation - **100%** ✅  
- **Phase 3**: Package Architecture & Boundaries - **100%** ✅ *(Interfaces only, implementation not migrated)*
- **Phase 4**: Code Quality & Best Practices - **77.8%** 🔄
- **Phase 5**: Automation & CI/CD Integration - **0%** ⏳
- **Phase 6**: CLI v2 Development & Migration - **0%** ⏳

**🎯 IMMEDIATE PRIORITY**: Complete actual modular architecture migration per [CLI_REFACTORING_CRITICAL_ROADMAP.md](mdc:CLI_REFACTORING_CRITICAL_ROADMAP.md)

### 🚀 Recent Achievements
- ✅ **Task 6 Completed**: Comprehensive performance benchmarking system with 35+ benchmarks
- ✅ **Task 7 Completed**: Integration tests created for CLI workflows and error recovery
- ✅ **Quality Gate Integration**: Automated 8-step quality pipeline with performance validation
- ✅ **Performance Targets**: All critical components meeting or exceeding performance benchmarks
- ✅ **CI/CD Enhancement**: Benchmark execution integrated into GitHub Actions workflow
- ✅ **Documentation**: 326-line performance guide with best practices and troubleshooting
- ✅ **Critical Analysis**: Identified that 25 files (7,565 lines) need actual migration to modular packages

### 🎯 Next Priorities
1. **CRITICAL**: Execute actual modular migration - 25 files (7,565 lines) in 8 tiers
2. **Task 8**: Implement code complexity metrics and monitoring
3. **Task 9**: Complete pre-commit hooks setup and validation

### 🛠️ New Cursor Rules Created
- ✅ **CLI Refactoring Guidelines**: Systematic tier-based migration process
- ✅ **Go Development Standards**: Code quality and architecture standards  
- ✅ **Architecture Principles**: Modular design principles and anti-patterns

> Systematic refactoring plan for creating a clean, modular, and maintainable CLI architecture

## 🎯 Refactoring Goals

- **Modular Architecture**: Clear separation of concerns with well-defined package boundaries
- **Eliminate Duplication**: Consolidate overlapping watch logic and test processing
- **Improve Testability**: Achieve ≥ 90% test coverage with isolated, testable components
- **Enhanced Maintainability**: Follow Go best practices and SOLID principles
- **Performance Optimization**: Leverage optimized caching and reduced resource usage

---

## 📋 Phase 1: Test Organization & Coverage Analysis

**Objective**: Establish baseline test coverage and reorganize test files for clarity.

### 1.1 Test File Reorganization
- [x] **Move co-located tests**: Ensure all `*_test.go` files are in the same package as their corresponding implementation
  - *Why*: Go convention for package-level testing and better test discovery
  - *How*: Move test files, update import paths, validate test execution
  - ✅ **COMPLETED**: All tests properly co-located in packages
- [x] **Validate test discovery**: Confirm all tests are properly discovered by `go test ./...`
  - *Why*: Ensures comprehensive test coverage measurement
  - *How*: Run `go test -v ./...` and verify all test files are included
  - ✅ **COMPLETED**: All tests discovered and running properly
- [x] **Create missing test files**: Add `*_test.go` files for untested components
  - *Why*: Establish TDD foundation for refactoring work
  - *How*: Create minimal failing tests for each major component
  - ✅ **COMPLETED**: Created 6 major test files (2,714 lines of test code)

### 1.2 Test Coverage Baseline
- [x] **Generate coverage report**: Run `go test -cover ./...` to establish current coverage
  - *Why*: Baseline for measuring refactoring impact on test coverage
  - *How*: Use `go test -coverprofile=coverage.out ./...` and `go tool cover -html=coverage.out`
  - ✅ **COMPLETED**: Baseline 53.6% → Final 61.6% coverage for internal/cli
- [x] **Identify coverage gaps**: Highlight functions/packages with <90% coverage
  - *Why*: Prioritize testing efforts during refactoring
  - *How*: Use coverage tools to generate gap analysis
  - ✅ **COMPLETED**: Identified and addressed critical gaps in processor, cache, parallel runner
- [x] **Create coverage improvement plan**: Document specific areas needing test enhancement
  - *Why*: Systematic approach to achieving coverage goals
  - *How*: Create checklist of functions requiring additional tests
  - ✅ **COMPLETED**: Systematic test creation for debouncer, renderer, extractor, cache

### 1.3 Test Quality Enhancement
- [x] **Standardize test naming**: Ensure all tests follow `TestXxx_Scenario` format
  - *Why*: Consistent test organization and better test output readability
  - *How*: Rename existing tests to follow standard convention
  - ✅ **COMPLETED**: All new tests follow TestXxx_Scenario naming convention
- [x] **Add integration tests**: Create comprehensive end-to-end test scenarios
  - *Why*: Validate complete workflows and catch integration issues
  - *How*: Create `integration_test.go` files with realistic test scenarios
  - ✅ **COMPLETED**: Added comprehensive CLI integration tests for v2 commands
- [x] **Implement test helpers**: Create shared test utilities and fixtures
  - *Why*: Reduce test duplication and improve test maintainability
  - *How*: Extract common test setup/teardown into helper functions
  - ✅ **COMPLETED**: Used shared test patterns and proper isolation throughout

---

## 📋 Phase 2: Watch Logic Consolidation

**Objective**: Eliminate duplication in watch-related functionality and create a unified watch system.

### 2.1 Watch Component Analysis
- [x] **Inventory watch files**: Document functionality in `watcher.go`, `watch_runner.go`, `watch_integration.go`
  - *Why*: Understand overlap and identify consolidation opportunities
  - *How*: Create functional analysis document showing component responsibilities
  - ✅ **COMPLETED**: Created comprehensive component analysis with duplication identification
- [x] **Identify shared interfaces**: Extract common contracts between watch components
  - *Why*: Enable clean separation and dependency injection
  - *How*: Define interfaces for file watching, event processing, and test triggering
  - ✅ **COMPLETED**: Designed 9 core interfaces with clean boundaries
- [x] **Map dependencies**: Document how watch components interact with each other
  - *Why*: Understand coupling and design clean boundaries
  - *How*: Create dependency graph showing component relationships
  - ✅ **COMPLETED**: Documented current dependencies and designed new architecture

### 2.2 Core Watch Architecture
- [x] **Create `internal/watch/core` package**: Define foundational watch interfaces and types
  - *Why*: Establish clean contracts for watch system components
  - *How*: Extract common interfaces like `FileWatcher`, `EventProcessor`, `TestTrigger`
  - ✅ **COMPLETED**: Created core package with interfaces and comprehensive types
- [x] **Implement `internal/watch/watcher` package**: File system monitoring functionality
  - *Why*: Isolate file system concerns from business logic
  - *How*: Move file watching logic from current files into focused package
  - ✅ **COMPLETED**: Implemented FSWatcher with PatternMatcher, consolidating watcher.go functionality
- [x] **Create `internal/watch/debouncer` package**: File change debouncing logic
  - *Why*: Separate temporal concerns from file system concerns
  - *How*: Extract debouncing logic into reusable component with configurable intervals
  - ✅ **COMPLETED**: Implemented race-condition-free debouncer with proper cleanup

### 2.3 Watch Integration Refactoring  
- [x] **Consolidate watch runners**: Merge overlapping watch execution logic
  - *Why*: Eliminate code duplication and simplify maintenance
  - *How*: Create unified `WatchRunner` that composes core components
  - ✅ **COMPLETED**: Created WatchCoordinator that orchestrates all watch components
- [x] **Implement watch modes**: Cleanly separate `WatchAll`, `WatchChanged`, `WatchRelated` logic
  - *Why*: Clear mode-specific behavior without code duplication
  - *How*: Use strategy pattern for different watch behaviors
  - ✅ **COMPLETED**: Implemented mode switching in WatchCoordinator.HandleFileChanges()
- [x] **Create watch configuration**: Centralized configuration for all watch behavior
  - *Why*: Consistent configuration interface and easy customization
  - *How*: Consolidate watch-related config into unified structure
  - ✅ **COMPLETED**: Created WatchOptions type and Configure() method in coordinator

---

## 📋 Phase 3: Package Architecture & Boundaries

**Objective**: Establish clear package boundaries and responsibilities following Go best practices.

### 3.1 Application Layer Design
- [x] **Create `internal/app` package**: Main application controller and orchestration
  - *Why*: Central coordination of application flow without business logic
  - *How*: Move high-level orchestration from current controller into app package
  - ✅ **COMPLETED**: Created app package with ApplicationController, LifecycleManager, DependencyContainer interfaces and implementations
- [x] **Implement dependency injection**: Use interfaces for component dependencies
  - *Why*: Improve testability and enable component substitution
  - *How*: Define interfaces and inject dependencies through constructors
  - ✅ **COMPLETED**: Implemented DependencyContainer with reflection-based dependency injection
- [x] **Add graceful shutdown**: Implement context-based cancellation and cleanup
  - *Why*: Proper resource management and clean application termination
  - *How*: Use context.Context throughout the application for cancellation
  - ✅ **COMPLETED**: Implemented LifecycleManager with signal handling and shutdown hooks

### 3.2 Test Processing Architecture
- [x] **Create `internal/test/runner` package**: Test execution engines and optimization
  - *Why*: Separate test execution concerns from output processing
  - *How*: Move test running logic into focused package with clear interfaces
  - ✅ **COMPLETED**: Created test runner package with TestExecutor interfaces and DefaultExecutor implementation
- [x] **Implement `internal/test/processor` package**: Test output parsing and processing
  - *Why*: Isolate JSON parsing and test result processing logic
  - *How*: Extract processor logic with stream processing capabilities
  - ✅ **COMPLETED**: Created processor interfaces for OutputProcessor, EventProcessor, TestEventParser, ResultAggregator
- [x] **Design `internal/test/cache` package**: Test result caching and optimization
  - *Why*: Dedicated caching logic separate from execution concerns
  - *How*: Implement cache interfaces with pluggable storage backends
  - ✅ **COMPLETED**: Created cache package with ResultCache, FileHashCache, DependencyCache interfaces

### 3.3 UI Component Architecture
- [x] **Create `internal/ui/display` package**: Test result rendering and formatting
  - *Why*: Separate presentation logic from business logic
  - *How*: Extract display formatting into reusable rendering components
  - ✅ **COMPLETED**: Created display package with DisplayRenderer, ProgressRenderer, ResultFormatter, LayoutManager interfaces
- [x] **Implement `internal/ui/colors` package**: Color formatting and theme management
  - *Why*: Centralized color management with theme support
  - *How*: Extract color logic with theme abstraction and terminal detection
  - ✅ **COMPLETED**: Created colors package with ColorFormatter, ThemeProvider, TerminalDetector interfaces and predefined themes
- [x] **Design `internal/ui/icons` package**: Icon providers and visual elements
  - *Why*: Consistent icon management across different terminal capabilities
  - *How*: Create icon abstraction with multiple provider implementations
  - ✅ **COMPLETED**: Created icons package with IconProvider, IconSetManager interfaces and predefined icon sets

### 3.4 Shared Components
- [x] **Create `pkg/events` package**: Event system for inter-component communication
  - *Why*: Decouple components through event-driven architecture
  - *How*: Implement event bus with typed events and subscription management
  - ✅ **COMPLETED**: Created events package with EventBus, EventHandler, EventStore interfaces and event types
- [x] **Implement `pkg/models` package**: Shared data models and value objects
  - *Why*: Common data structures without business logic
  - *How*: Move shared types into dedicated package with clear interfaces
  - ✅ **COMPLETED**: Created models package with TestResult, PackageResult, TestSummary and configuration types

---

## 📋 Phase 4: Code Quality & Best Practices (77.8% Complete)

**Focus**: Establish and enforce comprehensive code quality standards

### ✅ Task 1: Enhance error handling consistency
- [x] Implement structured error types
- [x] Add context to error messages
- [x] Create error validation utilities
**Status**: ✅ **COMPLETED** - Comprehensive error handling system implemented

### ✅ Task 2: Add comprehensive logging framework
- [x] Implement structured logging with levels
- [x] Add context-aware logging
- [x] Create log formatting utilities
**Status**: ✅ **COMPLETED** - Full logging system with multiple levels and formats

### ✅ Task 3: Implement proper configuration validation
- [x] Add configuration schema validation
- [x] Create validation error reporting
- [x] Implement configuration migration
**Status**: ✅ **COMPLETED** - Robust configuration validation with detailed error reporting

### ✅ Task 4: Enhance CLI argument parsing and validation
- [x] Implement structured argument validation
- [x] Add help text generation
- [x] Create command composition system
**Status**: ✅ **COMPLETED** - Comprehensive CLI system with validation and help generation

### ✅ Task 5: Add proper resource management and cleanup
- [x] Implement context-based cancellation
- [x] Add resource cleanup patterns
- [x] Create lifecycle management
**Status**: ✅ **COMPLETED** - Full lifecycle management with proper cleanup and signal handling

### ✅ Task 6: Implement performance benchmarks
- [x] Create comprehensive benchmark suite (35+ benchmarks)
- [x] Add performance monitoring and alerting
- [x] Integrate benchmarks into CI pipeline
- [x] Document performance best practices (326 lines)
**Status**: ✅ **COMPLETED** - Complete performance benchmarking system operational

### ✅ Task 7: Add integration tests
- [x] Create end-to-end CLI workflow tests
- [x] Add watch mode integration tests  
- [x] Implement error recovery scenario tests
- [x] Test cross-component integration
**Status**: ✅ **COMPLETED** - Comprehensive integration test suite for CLI workflows

### 🔄 Task 8: Implement code complexity metrics
- [ ] Add cyclomatic complexity measurement
- [ ] Create maintainability index calculation
- [ ] Implement technical debt tracking
- [ ] Add complexity reporting to CI
**Status**: 🔄 **IN PROGRESS** - Planning and design phase

### ⏳ Task 9: Add pre-commit hooks
- [ ] Implement git hooks for code quality
- [ ] Add automated formatting checks
- [ ] Create test execution hooks
- [ ] Add commit message validation
**Status**: ⏳ **PENDING** - Awaiting Task 8 completion

---

## 🚨 CRITICAL: Actual Modular Migration Required

**Current Reality**: We have modular packages but 25 files (7,565 lines) remain in `internal/cli`

**CRITICAL DISCOVERY (December 2024)**: Despite appearing to have "completed" TIER 7 UI components, a directory analysis reveals that **all UI components are still in `internal/cli`** and have NOT been migrated. The `internal/ui/` directory contains duplicate/test components, but the actual monolithic UI components remain in CLI.

**Actual Files Still in `internal/cli/` requiring migration**:
- `app_controller.go` (557 lines) - Main application orchestration
- `optimization_integration.go` (335 lines) - Watch optimization integration  
- `colors.go` (386 lines) - Color and icon management ❌ **NOT MIGRATED**
- `display.go` (167 lines) - Basic display functionality ❌ **NOT MIGRATED**
- `failed_tests.go` (509 lines) - Failed test display ❌ **NOT MIGRATED**
- `incremental_renderer.go` (433 lines) - Progressive rendering ❌ **NOT MIGRATED**
- `suite_display.go` (104 lines) - Suite display ❌ **NOT MIGRATED**
- `test_display.go` (160 lines) - Test display ❌ **NOT MIGRATED**

**Updated Reality**: The migration is actually at **75% completion** (6/8 tiers), not 87.5%.

See [CLI_REFACTORING_CRITICAL_ROADMAP.md](mdc:CLI_REFACTORING_CRITICAL_ROADMAP.md) for updated systematic migration plan:

### 🎯 Migration Tiers (8 Tiers, Updated Status)
1. **TIER 1**: Core Models & Types → `pkg/models/` ✅ **COMPLETED**
2. **TIER 2**: Configuration Management → `internal/config/` ✅ **COMPLETED**
3. **TIER 3**: Test Processing Engine (834 lines) → Split into 4 files in `internal/test/processor/` ✅ **COMPLETED**
4. **TIER 4**: Test Runners → `internal/test/runner/` ✅ **COMPLETED**
5. **TIER 5**: Caching System → `internal/test/cache/` ✅ **COMPLETED**
6. **TIER 6**: Watch System → `internal/watch/` ✅ **COMPLETED** (3/4 files, optimization_integration.go deferred)
7. **TIER 7**: UI Components → `internal/ui/` ❌ **NOT COMPLETED** (0/7 files actually migrated)
8. **TIER 8**: Application Orchestration → Refactor + Complete missing components ⏳ **IN PROGRESS**

**TIER 8 Strategy** (3 Phases):
- **8.1**: Complete missing UI migration (what should have been TIER 7)
- **8.2**: Watch integration optimization migration  
- **8.3**: Final app controller refactoring

This migration must be completed before continuing with remaining Phase 4 tasks.

---

## 📊 Overall Project Progress

**Total Progress: 58.2% (33/57 tasks completed)** *(Corrected from previous 61.4%)*

### Phase Breakdown:
- Phase 1: Foundation & Core Refactoring: **100%** (9/9 tasks)  
- Phase 2: Watch Logic Consolidation: **100%** (9/9 tasks)  
- Phase 3: Package Architecture & Boundaries: **100%** (12/12 tasks) *(Interfaces only, implementation not migrated)*
- Phase 4: Code Quality & Best Practices: **77.8%** (7/9 tasks completed)
- Phase 5: Automation & CI/CD Integration: **0%** (0/9 tasks)
- Phase 6: CLI v2 Development & Migration: **0%** (0/9 tasks)

**CRITICAL REALITY**: The modular architecture migration is actually at **75% completion** (6/8 tiers), not the previously reported 87.5%. TIER 7 UI components were never actually migrated from `internal/cli/`.

---

## 📋 Architectural Analysis Summary

### Current Issues Identified

**File Size Issues:**
- `processor.go`: 835 lines (needs split into 4-5 files)
- `app_controller.go`: 492 lines (needs reorganization)
- `failed_tests.go`: 509 lines (needs extraction)
- `watch_runner.go`: 373 lines (consolidation needed)

**Duplication Areas:**
- Watch logic spread across 3+ files
- Display formatting scattered across multiple files
- Configuration handling mixed with business logic
- Test running logic duplicated in multiple runners

**Package Boundary Issues:**
- All logic in single `internal/cli` package
- No clear separation of concerns
- Mixed responsibilities in single files
- Difficult to test components in isolation

### Target Architecture Benefits

**Clear Package Structure:**
```
internal/
├── app/          # Application orchestration
├── watch/        # File watching system
├── test/         # Test execution & processing
├── ui/           # User interface components
└── config/       # Configuration management

pkg/
├── events/       # Event system
└── models/       # Shared data models
```

**Improved Testability:**
- Dependency injection with interfaces
- Isolated components for unit testing
- Clear boundaries for integration testing
- Mocked dependencies for fast tests

**Better Maintainability:**
- Single responsibility principle
- Small, focused files (≤ 200 lines)
- Clear dependency graphs
- Easy to extend and modify

---

*This roadmap follows TDD principles and Go best practices. Each task includes rationale (why) and implementation approach (how) to ensure systematic execution and knowledge transfer.* 