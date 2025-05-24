# 🚨 CLI REFACTORING CRITICAL ROADMAP
## Completing the Actual Modular Architecture Migration

**CRITICAL STATUS**: We have created modular packages but the monolithic CLI still exists and is being used. This roadmap completes the actual code migration.

**Current Reality**: 23 source files (6,875 lines) still in `internal/cli` that need to be moved to modular architecture.

---

## 📊 File Analysis & Migration Map

### 🏗️ **TIER 1: Core Models & Types (Move First)**
*Foundation files that other components depend on*

- [x] **`types.go` (51 lines)** → `pkg/models/test_types.go`
  - TestEvent, TestResult, TestSuite, TestStatus enums
  - Core data structures used everywhere
  - **Risk**: Very high - all files depend on this
  - **Dependencies**: None (foundation)
  - ✅ **COMPLETED**: Moved to pkg/models/test_types.go with legacy constants for compatibility

- [x] **`models.go` (138 lines)** → `pkg/models/core_models.go`
  - TestProcessor, TestSuite, TestResult models
  - Configuration structures  
  - **Risk**: High - widely used
  - **Dependencies**: types.go
  - ✅ **COMPLETED**: Moved to pkg/models/core_models.go with LegacyTestResult and LegacyTestError types

### 🔧 **TIER 2: Configuration Management**
*Configuration loading and CLI parsing*

- [x] **`config.go` (437 lines)** → `internal/config/loader.go`
  - ConfigLoader interface and implementation
  - File parsing and validation logic
  - Default configuration management
  - **Risk**: Medium - used by app controller
  - **Dependencies**: models.go, types.go
  - ✅ **COMPLETED**: Moved to internal/config/loader.go with full functionality preserved

- [x] **`cli_args.go` (253 lines)** → `internal/config/args.go`
  - Command line argument parsing
  - Args struct and validation
  - CLI flag definitions
  - **Risk**: Medium - used by main entry points
  - **Dependencies**: config.go, models.go
  - ✅ **COMPLETED**: Moved to internal/config/args.go with backward compatibility layer in internal/cli/config_compat.go

### 🧪 **TIER 3: Test Processing Engine**
*Core test execution and result processing*

- [x] **`processor.go` (834 lines)** → **SPLIT INTO**:
  - `internal/test/processor/test_processor.go` (300 lines)
    - Main TestProcessor struct and core methods
    - ProcessJSONOutput, AddTestSuite, RenderResults
  - `internal/test/processor/event_handler.go` (200 lines)
    - onTestRun, onTestPass, onTestFail, onTestSkip methods
  - `internal/test/processor/error_processor.go` (200 lines)
    - createTestError, extractFileLocationFromLine
    - Source code extraction and error context
  - `internal/test/processor/statistics.go` (134 lines)
    - Statistics tracking and phase management
  - **Risk**: VERY HIGH - 834 lines, most complex file
  - **Dependencies**: models.go, types.go, source_extractor.go
  - ✅ **COMPLETED**: Split into modular components with backward compatibility layer

- [x] **`source_extractor.go` (143 lines)** → `internal/test/processor/source_extractor.go`
  - Source code context extraction
  - File location inference
  - **Risk**: Medium - used by processor
  - **Dependencies**: models.go
  - ✅ **COMPLETED**: Moved to internal/test/processor/source_extractor.go

- [x] **`parser.go` (272 lines)** → `internal/test/processor/json_parser.go`
  - JSON output parsing
  - Stream parsing logic
  - **Risk**: Medium - used by processor
  - **Dependencies**: models.go, types.go
  - ✅ **COMPLETED**: Moved to internal/test/processor/json_parser.go

- [x] **`stream.go` (356 lines)** → `internal/test/processor/stream_processor.go`
  - Stream processing logic
  - Real-time test output handling
  - **Risk**: Medium - used by optimized runners
  - **Dependencies**: processor.go, models.go
  - ✅ **COMPLETED**: Moved to internal/test/processor/stream_processor.go

### 🏃 **TIER 4: Test Runners**
*Test execution engines*

- [x] **`test_runner.go` (187 lines)** → `internal/test/runner/basic_runner.go`
  - Basic test execution
  - TestRunner interface implementation
  - **Risk**: Medium - entry point for test execution
  - **Dependencies**: types.go, models.go
  - ✅ **COMPLETED**: Moved to internal/test/runner/basic_runner.go with backward compatibility

- [ ] **`optimized_test_runner.go` (400 lines)** → `internal/test/runner/optimized_runner.go`
  - Optimized test execution with caching
  - Intelligent test selection
  - **Risk**: High - complex optimization logic
  - **Dependencies**: test_runner.go, test_cache.go, stream.go

- [x] **`parallel_runner.go` (195 lines)** → `internal/test/runner/parallel_runner.go`
  - Parallel test execution
  - Concurrency management
  - **Risk**: Medium - parallel execution complexity
  - **Dependencies**: test_runner.go, models.go
  - ✅ **COMPLETED**: Moved to internal/test/runner/parallel_runner.go with interface adaptation

- [ ] **`performance_optimizations.go` (356 lines)** → `internal/test/runner/performance_optimizer.go`
  - Performance optimization strategies
  - Resource usage optimization
  - **Risk**: Medium - performance critical
  - **Dependencies**: test_runner.go, optimized_test_runner.go

### 💾 **TIER 5: Caching System**
*Test result caching and optimization*

- [ ] **`test_cache.go` (340 lines)** → `internal/test/cache/result_cache.go`
  - Test result caching implementation
  - Cache invalidation logic
  - **Risk**: Medium - used by optimized runners
  - **Dependencies**: models.go, types.go

### 👁️ **TIER 6: Watch System**
*File watching and change detection*

- [ ] **`debouncer.go` (136 lines)** → `internal/watch/debouncer/file_debouncer.go`
  - File change debouncing
  - Event deduplication
  - **Risk**: Low - self-contained
  - **Dependencies**: models.go

- [ ] **`watcher.go` (347 lines)** → `internal/watch/watcher/fs_watcher.go`
  - File system watching
  - Pattern matching
  - **Risk**: Medium - file system interactions
  - **Dependencies**: debouncer.go, models.go

- [ ] **`watch_runner.go` (372 lines)** → `internal/watch/coordinator/watch_coordinator.go`
  - Watch mode orchestration
  - Test triggering logic
  - **Risk**: High - coordinates watch system
  - **Dependencies**: watcher.go, debouncer.go, test_runner.go

- [ ] **`optimization_integration.go` (333 lines)** → `internal/watch/coordinator/optimization_coordinator.go`
  - Watch optimization logic
  - Intelligent change detection
  - **Risk**: Medium - optimization complexity
  - **Dependencies**: watch_runner.go, optimized_test_runner.go

### 🎨 **TIER 7: UI Components**
*User interface and display logic*

- [ ] **`colors.go` (385 lines)** → `internal/ui/colors/color_formatter.go`
  - Color formatting and themes
  - Terminal detection
  - **Risk**: Low - presentation only
  - **Dependencies**: None

- [ ] **`display.go` (166 lines)** → `internal/ui/display/basic_display.go`
  - Basic display formatting
  - Core display logic
  - **Risk**: Low - presentation only
  - **Dependencies**: colors.go

- [ ] **`test_display.go` (159 lines)** → `internal/ui/display/test_display.go`
  - Individual test result display
  - Test formatting logic
  - **Risk**: Low - presentation only
  - **Dependencies**: display.go, colors.go

- [ ] **`suite_display.go` (103 lines)** → `internal/ui/display/suite_display.go`
  - Test suite display formatting
  - Suite summary logic
  - **Risk**: Low - presentation only
  - **Dependencies**: display.go, colors.go

- [ ] **`failed_tests.go` (508 lines)** → **SPLIT INTO**:
  - `internal/ui/display/failure_display.go` (300 lines)
    - Failed test rendering and formatting
  - `internal/ui/display/error_formatter.go` (208 lines)
    - Error message formatting and context
  - **Risk**: Medium - complex failure display logic
  - **Dependencies**: display.go, colors.go, source_extractor.go

- [ ] **`incremental_renderer.go` (351 lines)** → `internal/ui/renderer/incremental_renderer.go`
  - Progressive result rendering
  - Real-time display updates
  - **Risk**: Medium - complex rendering logic
  - **Dependencies**: display.go, colors.go

- [ ] **`summary.go` (190 lines)** → `internal/ui/display/summary_display.go`
  - Test run summary display
  - Statistics formatting
  - **Risk**: Low - presentation only
  - **Dependencies**: display.go, colors.go

### 🎯 **TIER 8: Application Orchestration**
*Main application controller*

- [ ] **`