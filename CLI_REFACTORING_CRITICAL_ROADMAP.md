# 🚨 CLI REFACTORING CRITICAL ROADMAP
## Completing the Actual Modular Architecture Migration

**CRITICAL STATUS**: We have created modular packages but the monolithic CLI still exists and is being used. This roadmap completes the actual code migration.

**Current Reality**: 25 source files (7,565 lines) still in `internal/cli` that need to be moved to modular architecture.

---

## 📊 File Analysis & Migration Map

### 🏗️ **TIER 1: Core Models & Types (Move First)**
*Foundation files that other components depend on*

- [ ] **`types.go` (51 lines)** → `pkg/models/test_types.go`
  - TestEvent, TestResult, TestSuite, TestStatus enums
  - Core data structures used everywhere
  - **Risk**: Very high - all files depend on this
  - **Dependencies**: None (foundation)

- [ ] **`models.go` (138 lines)** → `pkg/models/core_models.go`
  - TestProcessor, TestSuite, TestResult models
  - Configuration structures  
  - **Risk**: High - widely used
  - **Dependencies**: types.go

### 🔧 **TIER 2: Configuration Management**
*Configuration loading and CLI parsing*

- [ ] **`config.go` (437 lines)** → `internal/config/loader.go`
  - ConfigLoader interface and implementation
  - File parsing and validation logic
  - Default configuration management
  - **Risk**: Medium - used by app controller
  - **Dependencies**: models.go, types.go

- [ ] **`cli_args.go` (253 lines)** → `internal/config/args_parser.go`
  - Command line argument parsing
  - Args struct and validation
  - CLI flag definitions
  - **Risk**: Medium - used by main entry points
  - **Dependencies**: config.go, models.go

### 🧪 **TIER 3: Test Processing Engine**
*Core test execution and result processing*

- [ ] **`processor.go` (834 lines)** → **SPLIT INTO**:
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

- [ ] **`source_extractor.go` (143 lines)** → `internal/test/processor/source_extractor.go`
  - Source code context extraction
  - File location inference
  - **Risk**: Medium - used by processor
  - **Dependencies**: models.go

- [ ] **`parser.go` (272 lines)** → `internal/test/processor/json_parser.go`
  - JSON output parsing
  - Stream parsing logic
  - **Risk**: Medium - used by processor
  - **Dependencies**: models.go, types.go

- [ ] **`stream.go` (356 lines)** → `internal/test/processor/stream_processor.go`
  - Stream processing logic
  - Real-time test output handling
  - **Risk**: Medium - used by optimized runners
  - **Dependencies**: processor.go, models.go

### 🏃 **TIER 4: Test Runners**
*Test execution engines*

- [ ] **`test_runner.go` (187 lines)** → `internal/test/runner/basic_runner.go`
  - Basic test execution
  - TestRunner interface implementation
  - **Risk**: Medium - entry point for test execution
  - **Dependencies**: types.go, models.go

- [ ] **`optimized_test_runner.go` (400 lines)** → `internal/test/runner/optimized_runner.go`
  - Optimized test execution with caching
  - Intelligent test selection
  - **Risk**: High - complex optimization logic
  - **Dependencies**: test_runner.go, test_cache.go, stream.go

- [ ] **`parallel_runner.go` (195 lines)** → `internal/test/runner/parallel_runner.go`
  - Parallel test execution
  - Concurrency management
  - **Risk**: Medium - parallel execution complexity
  - **Dependencies**: test_runner.go, models.go

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

- [ ] **`app_controller.go` (553 lines)** → **REFACTOR TO USE NEW PACKAGES**:
  - Keep in `internal/app/` but update to use new modular components
  - Remove direct file dependencies, use interfaces
  - **Risk**: VERY HIGH - main orchestration point
  - **Dependencies**: ALL OTHER COMPONENTS

---

## 🛣️ **MIGRATION EXECUTION PLAN**

### **Phase A: Foundation Migration (Week 1)**
- [ ] Create temporary bridge interfaces to avoid breaking changes
- [ ] Move Tier 1 (types, models) to `pkg/models/`
- [ ] Update all imports across codebase
- [ ] Run full test suite to verify no regressions

### **Phase B: Configuration & Processing (Week 2)**
- [ ] Move Tier 2 (config, cli_args) to `internal/config/`
- [ ] **CRITICAL**: Split processor.go into 4 files (Tier 3)
- [ ] Move source_extractor, parser, stream to test/processor
- [ ] Update interfaces and dependencies
- [ ] Test configuration loading and test processing

### **Phase C: Test Execution (Week 3)**
- [ ] Move Tier 4 (all test runners) to `internal/test/runner/`
- [ ] Move Tier 5 (caching) to `internal/test/cache/`
- [ ] Update runner interfaces and coordination
- [ ] Test all execution modes (basic, optimized, parallel)

### **Phase D: Watch System (Week 4)**
- [ ] Move Tier 6 (watch components) to `internal/watch/`
- [ ] Update watch coordination and file monitoring
- [ ] Test watch mode functionality
- [ ] Verify debouncing and pattern matching

### **Phase E: UI Components (Week 5)**
- [ ] Move Tier 7 (UI components) to `internal/ui/`
- [ ] Split failed_tests.go into 2 files
- [ ] Update display interfaces and rendering
- [ ] Test all display modes and color themes

### **Phase F: Final Integration (Week 6)**
- [ ] Refactor app_controller.go to use new modular packages
- [ ] Remove old CLI package dependencies
- [ ] Update main entry points to use modular architecture
- [ ] Complete end-to-end testing
- [ ] Update documentation and examples

---

## ⚠️ **CRITICAL RISK MITIGATION**

### **High-Risk Files Requiring Special Attention**
1. **`processor.go` (834 lines)** - Most complex, needs careful splitting
2. **`app_controller.go` (553 lines)** - Main orchestration, affects everything
3. **`optimized_test_runner.go` (400 lines)** - Complex dependencies
4. **`watch_runner.go` (372 lines)** - Watch system coordination

### **Risk Mitigation Strategies**
- [ ] Create comprehensive integration tests before migration
- [ ] Use feature flags to enable gradual migration
- [ ] Maintain parallel implementations during transition
- [ ] Create rollback plan for each tier migration
- [ ] Continuous testing after each file migration

### **Dependencies That Must Be Handled Carefully**
- [ ] Types/Models → Everything depends on these
- [ ] Processor → Used by all runners and display components
- [ ] App Controller → Orchestrates everything
- [ ] Watch Runner → Coordinates watch system with test execution

### **Testing Strategy During Migration**
- [ ] Run full test suite after each tier migration
- [ ] Create integration tests for cross-package interfaces
- [ ] Test all CLI commands and watch modes
- [ ] Performance benchmarking to catch regressions
- [ ] Manual testing of critical user workflows

---

## 📋 **SUCCESS CRITERIA**

- [ ] **Zero Files in `internal/cli`** except minimal entry point
- [ ] **All Tests Passing** with ≥ 90% coverage maintained
- [ ] **No Performance Regression** in execution speed
- [ ] **Clean Package Boundaries** with proper interfaces
- [ ] **Documentation Updated** for new architecture
- [ ] **Examples Working** with new package structure

---

**ESTIMATED EFFORT**: 6 weeks (assuming 1 person, full-time focus)
**CRITICAL SUCCESS FACTOR**: Systematic tier-by-tier migration with continuous testing

*This roadmap ensures we complete the actual modular refactoring that was promised in Phases 1-3 but never actually executed.* 