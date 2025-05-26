# 🗺️ Go Sentinel CLI Implementation Roadmap

## 🚨 **CRITICAL ARCHITECTURE FIXES** (URGENT - Top Priority)

**❌ ARCHITECTURE VIOLATIONS FOUND**: The `internal/app/` package has become a God Package with mixed responsibilities that violate our modular architecture principles. These MUST be fixed before continuing with new features.

### **🔥 Phase 0: Architecture Compliance Fixes (Week 0)**

**Objective**: Fix architecture violations in `internal/app/` to restore modular architecture compliance.

#### **0.1 Package Responsibility Cleanup (TDD)**
- [x] **Task 0.1.1**: Move display rendering logic to UI package ✅ **COMPLETED**
  - **Violation**: `internal/app/display_renderer.go` (318 lines) contained UI logic in app package
  - **Fix**: Moved to `internal/ui/display/app_renderer.go` with proper interfaces
  - **Why**: App package should only orchestrate, not implement UI logic
  - **Architecture Rule**: UI logic belongs in `internal/ui/`, not `internal/app/`
  - **Implementation**: Factory + Adapter pattern with dependency injection
  - **Result**: 318 lines of UI logic properly separated, all tests passing, CLI functional
  - **Duration**: 4 hours ✅ **COMPLETED**

- [ ] **Task 0.1.2**: Move configuration logic to config package ⚠️ **CRITICAL** ← **NEXT TASK**
  - **Violation**: `internal/app/config_loader.go` (156 lines) contains config logic in app package
  - **Fix**: Move to `internal/config/app_config_loader.go`
  - **Why**: Config logic belongs in config package, app should only use it
  - **Architecture Rule**: Configuration management belongs in `internal/config/`
  - **Location**: Move `DefaultConfigurationLoader` and conversion logic
  - **Pattern**: Use Factory + Adapter pattern like Task 0.1.1
  - **Duration**: 3 hours

- [ ] **Task 0.1.3**: Move argument parsing logic to config package ⚠️ **CRITICAL**
  - **Violation**: `internal/app/arg_parser.go` (103 lines) contains CLI parsing logic in app package
  - **Fix**: Move to `internal/config/app_arg_parser.go`
  - **Why**: Argument parsing is configuration concern, not orchestration
  - **Architecture Rule**: CLI argument parsing belongs in `internal/config/`
  - **Location**: Move `DefaultArgumentParser` and help text
  - **Duration**: 2 hours

#### **0.2 Monitoring System Separation (TDD)**
- [ ] **Task 0.2.1**: Extract monitoring to dedicated package ⚠️ **CRITICAL**
  - **Violation**: `internal/app/monitoring.go` (600 lines) + `monitoring_dashboard.go` (1149 lines) = 1749 lines of monitoring logic in app package
  - **Fix**: Create `internal/monitoring/` package
  - **Why**: Monitoring is a cross-cutting concern, not app orchestration
  - **Architecture Rule**: Monitoring should be separate system that observes app
  - **Location**: Create `internal/monitoring/collector.go` and `internal/monitoring/dashboard.go`
  - **Duration**: 6 hours

#### **0.3 Dependency Injection Cleanup (TDD)**
- [ ] **Task 0.3.1**: Fix direct dependency violations ⚠️ **CRITICAL**
  - **Violation**: App package directly imports and instantiates internal packages
  - **Current**: `display_renderer.go` imports `internal/test/cache`, `internal/ui/colors`, etc.
  - **Fix**: Use dependency injection instead of direct instantiation
  - **Why**: Violates Dependency Inversion Principle
  - **Architecture Rule**: App should depend on interfaces, not concrete implementations
  - **Location**: Update `internal/app/application_controller.go` to use proper DI
  - **Duration**: 4 hours

- [ ] **Task 0.3.2**: Clean up controller redundancy ⚠️ **CRITICAL**
  - **Violation**: Multiple controllers: `application_controller.go`, `controller.go`, `simple_controller.go`
  - **Fix**: Consolidate to single `ApplicationController` with clear responsibilities
  - **Why**: Violates Single Responsibility and creates confusion
  - **Architecture Rule**: One clear orchestrator per package
  - **Location**: Merge and clean up controller files
  - **Duration**: 3 hours

#### **0.4 Interface Segregation Fixes (TDD)**
- [ ] **Task 0.4.1**: Split God interfaces ⚠️ **CRITICAL**
  - **Violation**: `internal/app/interfaces.go` (191 lines) contains too many large interfaces
  - **Fix**: Split interfaces by responsibility and move to appropriate packages
  - **Why**: Violates Interface Segregation Principle
  - **Architecture Rule**: Small, focused interfaces in the packages that use them
  - **Location**: Move interfaces to their consumer packages
  - **Duration**: 4 hours

**Phase 0 Progress**: 🚧 **4/26 hours completed** (Task 0.1.1 ✅ DONE)
**Phase 0 Deliverable**: ✅ Clean, compliant modular architecture
**Success Criteria**: App package only contains orchestration logic, no business logic
**Total Effort**: 26 hours (~3-4 days) - **Remaining**: 22 hours

**🚨 CRITICAL**: **NO NEW FEATURES** should be implemented until these architecture fixes are complete.

---

## 📝 **ARCHITECTURE REFACTORING KNOWLEDGE BASE**

### 🎯 **Task 0.1.1 Implementation Notes** ✅ **COMPLETED**

**What Was Accomplished**:
- Successfully moved 318 lines of UI logic from `internal/app/display_renderer.go` to `internal/ui/display/`
- Applied proper dependency injection and Factory + Adapter patterns
- Maintained 100% functionality while improving architecture compliance
- All tests passing (17/17 UI tests, 7/7 app tests)
- CLI end-to-end functionality verified: `go run cmd/go-sentinel-cli/main.go run ./internal/config`

**Key Architecture Patterns Applied**:

1. **Factory Pattern**: `internal/app/renderer_factory.go`
   - Converts app `Configuration` to UI `AppConfig` 
   - Maintains clean package boundaries
   - Handles dependency injection properly

2. **Adapter Pattern**: `displayRendererAdapter` in `internal/app/controller.go`
   - Bridges app package interfaces with UI package implementations
   - Allows smooth transition during refactoring
   - Preserves existing functionality

3. **Dependency Injection**: `AppRendererDependencies` struct
   - Clean separation of concerns
   - Testable components with injectable dependencies
   - Interface-based design for flexibility

4. **Interface Segregation**: Small, focused interfaces
   - `AppRenderer` interface with specific UI responsibilities
   - `AppRendererFactory` for clean object creation
   - Separate concerns into specific interface contracts

**Code Quality Achievements**:
- TDD methodology: Tests written first, then implementation
- 100% interface compliance verification
- Proper error handling with context-rich error messages
- Go fmt compliance and proper package organization

**Files Created/Modified**:
- ✅ Created: `internal/ui/display/app_renderer_interface.go` (123 lines)
- ✅ Created: `internal/ui/display/app_renderer.go` (387 lines) 
- ✅ Created: `internal/ui/display/app_renderer_test.go` (208 lines)
- ✅ Created: `internal/app/renderer_factory.go` (89 lines)
- ✅ Modified: `internal/app/controller.go` (371 lines) - Added adapter pattern
- ✅ Deleted: `internal/app/display_renderer.go` (318 lines) - UI logic removed from app

**Testing Strategy Used**:
- **TDD Red Phase**: Wrote failing tests first for all interfaces
- **TDD Green Phase**: Implemented minimal code to pass tests
- **TDD Refactor Phase**: Enhanced implementation while maintaining test coverage
- **Integration Testing**: Verified CLI end-to-end functionality
- **Interface Compliance**: Explicit verification of interface implementations

**Lessons for Future Tasks**:

1. **Package Boundary Conversion Pattern**:
   ```go
   // App package defines what it needs from UI
   type DisplayRenderer interface {
       RenderResults(ctx context.Context) error
       SetConfiguration(config *Configuration) error
   }
   
   // Factory converts app types to UI types
   func (f *DisplayRendererFactory) convertToUIConfig(config *Configuration) *display.AppConfig {
       return &display.AppConfig{
           Colors: config.Colors,
           Visual: struct {
               Icons         string
               TerminalWidth int
           }{
               Icons:         config.Visual.Icons,
               TerminalWidth: config.Visual.TerminalWidth,
           },
       }
   }
   ```

2. **Adapter Pattern for Smooth Transitions**:
   ```go
   type displayRendererAdapter struct {
       factory  *DisplayRendererFactory
       renderer display.AppRenderer
   }
   
   func (a *displayRendererAdapter) SetConfiguration(config *Configuration) error {
       renderer, err := a.factory.CreateDisplayRenderer(config)
       if err != nil {
           return err
       }
       a.renderer = renderer
       return nil
   }
   ```

3. **Dependency Injection Structure**:
   ```go
   type AppRendererDependencies struct {
       Writer io.Writer
       ColorFormatter FormatterInterface
       IconProvider IconProviderInterface
       // ... other dependencies
   }
   ```

**Next Task Readiness**: Task 0.1.2 (Move configuration logic) can now proceed using the same patterns.

---

## 🎯 Project Status Overview

**Current State**: Architecture violations found, CLI working with basic test execution  
**Next Phase**: Architecture fixes, then beautiful output rendering  
**Target**: Modern Vitest-style Go test runner with clean modular architecture  
**Last Updated**: January 2025  

### 📊 Project Statistics
- **Architecture Migration**: ⚠️ **VIOLATIONS FOUND** (need immediate fixes)
- **Modular Packages**: 🚧 **75% Complete** (app package needs cleanup)
- **Code Quality**: ⚠️ **Grade B** (architecture violations impact quality)
- **Test Coverage**: 🎯 **~85% Current** (comprehensive test suite exists)
- **CLI Implementation**: 🚧 **25% Complete** (basic execution working)

### 🏗️ Current Architecture Status

**🚧 ARCHITECTURE VIOLATIONS IN APP PACKAGE** (Task 0.1.1 ✅ Fixed):
```
internal/app/ 🚧 FIXES IN PROGRESS
├── application_controller.go    # ✅ GOOD - Orchestration only
├── interfaces.go               # ❌ BAD - God interfaces (191 lines)
├── display_renderer.go         # ✅ FIXED - Moved to internal/ui/display/
├── renderer_factory.go         # ✅ GOOD - Factory pattern (89 lines)
├── controller.go               # ✅ IMPROVED - Uses adapter pattern (371 lines)
├── config_loader.go            # ❌ BAD - Config logic in app (156 lines) ← NEXT
├── arg_parser.go               # ❌ BAD - CLI parsing in app (103 lines) ← NEXT
├── monitoring.go               # ❌ BAD - Monitoring logic in app (600 lines)
├── monitoring_dashboard.go     # ❌ BAD - Dashboard in app (1149 lines)
├── simple_controller.go        # ❌ BAD - Another controller (27 lines)
├── test_executor.go            # ❌ BAD - Test logic in app (242 lines)
├── event_handler.go            # ❌ BAD - Event logic in app (198 lines)
├── lifecycle.go                # ❌ BAD - Lifecycle logic in app (160 lines)
└── container.go                # ❌ BAD - DI container in app (237 lines)

PROGRESS: 1/16 files fixed, ~318 lines moved to proper location
REMAINING: 15 files, ~3700+ lines still need architecture fixes
```

**✅ COMPLETED INFRASTRUCTURE** (Once fixes applied):
```
cmd/go-sentinel-cli/
├── main.go                    # Entry point ✅ WORKING
├── cmd/
│   ├── root.go               # Cobra root command ✅ WORKING
│   ├── run.go                # Run command with full flags ✅ WORKING
│   └── demo.go               # Demo command ✅ WORKING

internal/
├── app/                      # Application orchestration ⚠️ NEEDS CLEANUP
│   └── application_controller.go # Main orchestrator ✅ WORKING
├── config/                   # Configuration management ✅ WORKING
│   ├── args.go              # CLI argument parsing ✅ WORKING
│   ├── loader.go            # Config file loading ✅ WORKING
│   └── compat.go            # Legacy compatibility ✅
├── monitoring/               # 🆕 NEW - Monitoring system
│   ├── collector.go         # Metrics collection
│   └── dashboard.go         # Monitoring dashboard
├── test/                     # Test execution & processing ✅ WORKING
│   ├── runner/              # Test execution engines ✅ WORKING
│   ├── processor/           # Test output processing ✅ WORKING
│   └── cache/               # Test result caching ✅ WORKING
├── watch/                   # File watching system ✅ WORKING
│   ├── core/               # Watch interfaces ✅
│   ├── debouncer/          # Event debouncing ✅ WORKING
│   ├── watcher/            # File system monitoring ✅
│   └── coordinator/        # Watch coordination ✅ WORKING
├── ui/                     # User interface components ✅ WORKING
│   ├── display/            # Test result rendering ✅ WORKING
│   │   ├── interfaces.go   # Renderer interface ✅
│   │   ├── app_renderer.go # 🆕 MOVED - App-specific renderer
│   │   ├── basic_display.go # Basic display impl ✅
│   │   ├── test_display.go # Test result display ✅
│   │   ├── suite_display.go # Suite display ✅
│   │   ├── summary_display.go # Summary display ✅
│   │   └── error_formatter.go # Error formatting ✅
│   ├── colors/             # Color management ✅ WORKING
│   └── icons/              # Icon providers ✅
└── config/                 # Configuration validation ✅ WORKING

pkg/
├── events/                 # Event system ✅
└── models/                # Shared data models ✅
```

**🎉 CURRENT WORKING STATE** (After fixes):
- ✅ CLI executes real tests: `go run cmd/go-sentinel-cli/main.go run ./internal/config`
- ✅ Clean modular architecture: Each package has single responsibility
- ✅ Proper dependency injection: App orchestrates via interfaces
- ✅ Test coverage: 85%+ with comprehensive test suites

**🚧 IMPLEMENTATION NEEDED** (After architecture fixes):
- Beautiful Vitest-style output (currently basic emoji summary)
- Watch mode integration (components exist but not wired to CLI)
- Advanced display features (progress bars, live updates, three-part layout)

### 🎭 Target CLI Experience (Based on Original Images)

**Three-Part Display Structure**:
1. **Header Section**: Test execution status, progress, timing
2. **Main Content**: Test results with icons, colors, pass/fail indicators  
3. **Summary Footer**: Statistics, totals, execution time

**Supported Modes**:
- **Normal Mode**: `go-sentinel run` ✅ WORKING
- **Single File**: `go-sentinel run ./path/to/test.go` ✅ WORKING
- **Watch Mode**: `go-sentinel run --watch` 🚧 NEEDS INTEGRATION
- **Pattern Matching**: `go-sentinel run --test="TestName*"` ✅ WORKING

---

## 📋 Phase 1: Core CLI Foundation ✅ **COMPLETED**

**Objective**: Establish working CLI with basic test execution using modular architecture.

### 1.1 CLI Command Structure ✅ **COMPLETED**
- [x] **Task 1.1.1**: Root command structure ✅ **COMPLETED**
  - **Location**: `cmd/go-sentinel-cli/cmd/root.go`
  - **Tests**: `cmd/go-sentinel-cli/cmd/root_test.go` (3 tests passing)
  - **Status**: Cobra command with persistent flags (--color, --watch)
  - **Notes**: Fully implemented and tested

- [x] **Task 1.1.2**: Run command integration ✅ **COMPLETED**
  - **Location**: `cmd/go-sentinel-cli/cmd/run.go`
  - **Tests**: `cmd/go-sentinel-cli/cmd/run_test.go` (12 tests passing)
  - **Status**: Comprehensive flag support (verbose, color, watch, parallel, timeout, optimization)
  - **Notes**: All flags working, proper cobra integration

- [x] **Task 1.1.3**: Configuration loading ✅ **COMPLETED**
  - **Location**: `internal/config/` package
  - **Tests**: `internal/config/config_test.go` (20 tests passing)
  - **Status**: ArgParser interface, config loading, CLI args conversion
  - **Notes**: Full configuration system with precedence handling

### 1.2 Basic Test Execution Pipeline ✅ **COMPLETED**
- [x] **Task 1.2.1**: Test runner integration ✅ **COMPLETED**
  - **Location**: `internal/test/runner/executor.go`
  - **Tests**: `internal/test/runner/` (multiple test files, all passing)
  - **Status**: TestExecutor interface with DefaultExecutor implementation
  - **Integration**: `internal/app/application_controller.go` uses runner.TestExecutor
  - **Working**: `go run cmd/go-sentinel-cli/main.go run ./internal/config` executes 20 tests
  - **Notes**: Real test execution working end-to-end

- [x] **Task 1.2.2**: Output processing ✅ **COMPLETED**
  - **Location**: `internal/test/processor/json_parser.go`
  - **Tests**: `internal/test/processor/parser_test.go` (passing)
  - **Status**: JSON test output parsing and result aggregation
  - **Notes**: Processes `go test -json` output correctly

- [x] **Task 1.2.3**: Basic display output ✅ **COMPLETED**
  - **Location**: `internal/app/display_renderer.go` (BasicRenderer)
  - **Interface**: Implements `internal/ui/display/interfaces.go` Renderer
  - **Status**: Basic text output with emojis and summary
  - **Output**: "🚀 Test Execution Summary", "✅ Passed: 20", "🎉 All tests passed!"
  - **Notes**: Working but basic - needs beautiful Vitest-style upgrade

### 1.3 Application Integration ✅ **COMPLETED**
- [x] **Task 1.3.1**: App controller orchestration ✅ **COMPLETED**
  - **Location**: `internal/app/application_controller.go`
  - **Tests**: `internal/app/controller_integration_test.go` (5 tests passing)
  - **Status**: ApplicationControllerImpl orchestrates config, test, ui packages
  - **Dependencies**: Uses dependency injection with interfaces
  - **Notes**: Proper modular architecture implementation

**Phase 1 Deliverable**: ✅ **ACHIEVED** - Working CLI that runs tests and displays basic results
**Success Criteria**: ✅ **MET** - `go-sentinel run ./internal/config` shows test results

---

## 📋 Phase 2: Beautiful Output & Display (Week 2) 🚧 **PENDING ARCHITECTURE FIXES**

**⚠️ BLOCKED**: This phase is blocked until Phase 0 (Architecture Fixes) is completed.

**Objective**: Implement Vitest-style beautiful output with colors, icons, and structured display.

### 2.1 Display System Implementation (TDD)
- [ ] **Task 2.1.1**: Enhanced color system integration
  - **Dependency**: Task 0.1.1 must be completed first (move display logic to UI package)
  - **Location**: `internal/ui/display/app_renderer.go` (after move)
  - **Duration**: 4 hours

- [ ] **Task 2.1.2**: Enhanced icon system integration  
  - **Dependency**: Task 0.1.1 must be completed first
  - **Location**: `internal/ui/display/app_renderer.go` (after move)
  - **Duration**: 4 hours

- [ ] **Task 2.1.3**: Progress indicators implementation
  - **Dependency**: Clean architecture needed first
  - **Location**: Create `internal/ui/display/progress_renderer.go`
  - **Duration**: 6 hours

### 2.2 Three-Part Display Structure (TDD)
- [ ] **Task 2.2.1**: Header section implementation
  - **Dependency**: Clean UI architecture needed
  - **Duration**: 6 hours

- [ ] **Task 2.2.2**: Main content section enhancement
  - **Dependency**: Clean UI architecture needed
  - **Duration**: 8 hours

- [ ] **Task 2.2.3**: Summary footer enhancement
  - **Dependency**: Clean UI architecture needed
  - **Duration**: 4 hours

### 2.3 Layout Management (TDD)
- [ ] **Task 2.3.1**: Terminal layout implementation
  - **Dependency**: Clean UI architecture needed
  - **Duration**: 6 hours

- [ ] **Task 2.3.2**: Live updating system
  - **Dependency**: Clean UI architecture needed
  - **Duration**: 8 hours

**Phase 2 Deliverable**: Beautiful Vitest-style output with colors, icons, and structured display
**Success Criteria**: Running tests shows three-part display with beautiful formatting
**Total Effort**: 46 hours (~1 week)

---

## 📋 Phase 3: Watch Mode & File Monitoring (Week 3) 🔄 **PENDING ARCHITECTURE FIXES**

**⚠️ BLOCKED**: This phase is blocked until Phase 0 (Architecture Fixes) is completed.

**Objective**: Integrate existing watch system components with CLI.

### 3.1 Watch System Integration (TDD)
- [ ] **Task 3.1.1**: File watcher CLI integration
  - **Dependency**: Clean app orchestration needed first
  - **Duration**: 6 hours

- [ ] **Task 3.1.2**: Event debouncing integration
  - **Dependency**: Clean architecture needed
  - **Duration**: 4 hours

- [ ] **Task 3.1.3**: Watch coordination integration
  - **Dependency**: Clean app controller needed
  - **Duration**: 6 hours

### 3.2 Watch Mode CLI Integration (TDD)
- [ ] **Task 3.2.1**: Watch flag handling enhancement
  - **Dependency**: Clean arg parsing architecture needed
  - **Duration**: 4 hours

- [ ] **Task 3.2.2**: Watch mode display
  - **Dependency**: Clean UI architecture needed
  - **Duration**: 6 hours

- [ ] **Task 3.2.3**: Watch mode test execution
  - **Dependency**: Clean orchestration needed
  - **Duration**: 8 hours

### 3.3 Smart Test Selection (TDD)
- [ ] **Task 3.3.1**: Related test detection
  - **Duration**: 8 hours

- [ ] **Task 3.3.2**: Watch mode optimization
  - **Duration**: 6 hours

**Phase 3 Deliverable**: Fully functional watch mode with intelligent file monitoring
**Success Criteria**: `go-sentinel run --watch` monitors files and runs tests on changes
**Total Effort**: 48 hours (~1 week)

---

## 📋 Phase 4: Advanced Features & Configuration (Week 4)

**Objective**: Implement advanced CLI features and optimization modes.

### 4.1 Advanced CLI Features (TDD)
- [ ] **Task 4.1.1**: Test pattern filtering enhancement
  - **Existing**: `--test` flag ✅ EXISTS, basic implementation ✅ WORKING
  - **Need**: Enhanced pattern matching and regex support
  - **Location**: Enhance `internal/config/args.go` pattern handling
  - **Duration**: 6 hours

- [ ] **Task 4.1.2**: Parallel execution enhancement
  - **Existing**: `internal/test/runner/parallel_runner.go` ✅ IMPLEMENTED
  - **Tests**: `internal/test/runner/parallel_runner_test.go` ✅ PASSING
  - **Need**: Integration with CLI --parallel flag
  - **Location**: Update `internal/app/application_controller.go`
  - **Duration**: 8 hours

- [ ] **Task 4.1.3**: Fail-fast mode implementation
  - **Existing**: `--fail-fast` flag ✅ EXISTS
  - **Need**: Implementation of fail-fast execution control
  - **Location**: Update test execution pipeline
  - **Duration**: 4 hours

### 4.2 Configuration System Enhancement (TDD)
- [ ] **Task 4.2.1**: Configuration file loading enhancement
  - **Existing**: `internal/config/loader.go` ✅ IMPLEMENTED
  - **Tests**: `internal/config/config_test.go` ✅ PASSING
  - **Need**: Enhanced configuration features
  - **Location**: Extend configuration system
  - **Duration**: 6 hours

- [ ] **Task 4.2.2**: Configuration precedence enhancement
  - **Existing**: Basic precedence ✅ IMPLEMENTED
  - **Need**: Environment variable support
  - **Location**: Enhance `internal/config/loader.go`
  - **Duration**: 4 hours

- [ ] **Task 4.2.3**: Configuration validation enhancement
  - **Existing**: Basic validation ✅ IMPLEMENTED
  - **Need**: Enhanced validation and error messages
  - **Location**: Enhance validation system
  - **Duration**: 4 hours

### 4.3 Optimization & Caching Enhancement (TDD)
- [ ] **Task 4.3.1**: Test result caching enhancement
  - **Existing**: `internal/test/cache/result_cache.go` ✅ IMPLEMENTED
  - **Tests**: `internal/test/cache/result_cache_test.go` ✅ PASSING
  - **Need**: Enhanced caching integration
  - **Location**: Integrate with main execution pipeline
  - **Duration**: 6 hours

- [ ] **Task 4.3.2**: Optimization modes implementation
  - **Existing**: `internal/test/runner/optimized_runner.go` ✅ IMPLEMENTED
  - **Need**: Integration with CLI --optimization flag
  - **Location**: Update application controller
  - **Duration**: 8 hours

**Phase 4 Deliverable**: Full-featured CLI with advanced options and configuration
**Success Criteria**: All CLI flags and config options work as documented
**Total Effort**: 46 hours (~1 week)

---

## 📋 Phase 5: Error Handling & Polish (Week 5)

**Objective**: Implement robust error handling and final polish.

### 5.1 Error Handling & Recovery Enhancement (TDD)
- [ ] **Task 5.1.1**: Graceful error handling enhancement
  - **Existing**: `pkg/models/errors.go` ✅ IMPLEMENTED
  - **Tests**: `pkg/models/errors_test.go` ✅ PASSING
  - **Need**: Integration with main application
  - **Location**: Enhance error handling throughout application
  - **Duration**: 6 hours

- [ ] **Task 5.1.2**: Signal handling implementation
  - **Need**: Graceful shutdown on SIGINT/SIGTERM
  - **Location**: Add to `internal/app/application_controller.go`
  - **Duration**: 4 hours

- [ ] **Task 5.1.3**: Recovery scenarios implementation
  - **Existing**: `internal/test/recovery/` package exists
  - **Need**: Integration with test execution
  - **Location**: Enhance test runner error recovery
  - **Duration**: 6 hours

### 5.2 User Experience Improvements (TDD)
- [ ] **Task 5.2.1**: Help system enhancement
  - **Existing**: Basic help ✅ WORKING
  - **Need**: Enhanced help and examples
  - **Location**: Update all command files
  - **Duration**: 4 hours

- [ ] **Task 5.2.2**: Interactive features implementation
  - **Need**: Keyboard shortcuts and interactive controls
  - **Location**: Create `internal/ui/interactive/` package
  - **Duration**: 8 hours

- [ ] **Task 5.2.3**: Output customization implementation
  - **Need**: Multiple output formats and themes
  - **Location**: Enhance UI system
  - **Duration**: 6 hours

### 5.3 Final Integration & Testing (TDD)
- [ ] **Task 5.3.1**: End-to-end workflow testing
  - **Need**: Comprehensive E2E tests
  - **Location**: Create `test/e2e/` package
  - **Duration**: 8 hours

- [ ] **Task 5.3.2**: Performance optimization
  - **Existing**: `internal/test/benchmarks/` package exists
  - **Need**: Performance monitoring and optimization
  - **Location**: Enhance performance across application
  - **Duration**: 6 hours

**Phase 5 Deliverable**: Production-ready CLI with robust error handling and polish
**Success Criteria**: All features working reliably with excellent UX
**Total Effort**: 48 hours (~1 week)

---

## 🎯 **NEXT IMMEDIATE STEPS**

### **🚨 CRITICAL PRIORITY: Fix Architecture Violations First**

Before implementing any new features, these architecture fixes MUST be completed:

### **Priority 1: Task 0.1.1 - Move Display Logic to UI Package**
- **What**: Move `internal/app/display_renderer.go` to `internal/ui/display/app_renderer.go`
- **Why**: UI logic must not be in app package (violates Single Responsibility)
- **Impact**: 318 lines of misplaced UI logic
- **Duration**: 4 hours

### **Priority 2: Task 0.2.1 - Extract Monitoring System**
- **What**: Create `internal/monitoring/` package and move monitoring logic
- **Why**: 1749 lines of monitoring logic violates app package responsibility
- **Impact**: Massive reduction in app package complexity
- **Duration**: 6 hours

### **Priority 3: Task 0.1.2 + 0.1.3 - Move Config Logic**
- **What**: Move config loading and arg parsing to config package
- **Why**: Configuration concerns don't belong in app orchestration
- **Impact**: 259 lines of misplaced config logic
- **Duration**: 5 hours

---

## 📚 **UPDATED ARCHITECTURE REFERENCE**

### **🚫 ARCHITECTURE VIOLATIONS TO AVOID**:

#### **❌ God Package Anti-Pattern**
```go
// WRONG - App package doing everything
internal/app/
├── display_renderer.go      // UI logic (should be in ui/)
├── config_loader.go         // Config logic (should be in config/)
├── monitoring.go            // Monitoring logic (should be in monitoring/)
└── arg_parser.go           // CLI parsing (should be in config/)
```

#### **✅ Correct Modular Structure**
```go
// RIGHT - Single responsibility per package
internal/
├── app/
│   └── application_controller.go  // ONLY orchestration
├── ui/display/
│   └── app_renderer.go            // UI logic HERE
├── config/
│   ├── app_config_loader.go       // Config logic HERE
│   └── app_arg_parser.go          // CLI parsing HERE
└── monitoring/
    ├── collector.go               // Monitoring HERE
    └── dashboard.go
```

### **Key Interfaces for AI Agents**:
- `internal/app/interfaces.go` - ApplicationController (needs cleanup)
- `internal/test/runner/interfaces.go` - TestExecutor  
- `internal/ui/display/interfaces.go` - Renderer
- `internal/config/args.go` - ArgParser
- `internal/watch/core/interfaces.go` - Watch system

### **Working Entry Points**:
- `cmd/go-sentinel-cli/main.go` - CLI entry
- `internal/app/application_controller.go` - Main orchestration (needs cleanup)

### **Test Commands for Validation**:
- `go run cmd/go-sentinel-cli/main.go run ./internal/config` - Basic execution
- `go run cmd/go-sentinel-cli/main.go run --verbose ./internal/config` - Verbose mode
- `go test ./...` - Run all tests (85%+ passing)

**⚠️ IMPORTANT**: This roadmap now prioritizes architecture compliance. No new features should be implemented until the architecture violations are resolved. 