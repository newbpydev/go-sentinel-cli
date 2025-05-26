# 🔄 Phase 3: Watch Mode & File Monitoring Roadmap

## 📋 **PHASE 3: WATCH MODE & FILE MONITORING** ✅ **READY TO PROCEED**

**Objective**: Integrate existing watch system components with CLI for intelligent file monitoring and test execution.

**Visual Standards**: 📋 **MUST FOLLOW** → [Go Sentinel CLI Visual Guidelines](./GO_SENTINEL_CLI_VISUAL_GUIDELINES.md)

**Current Status**: ✅ Watch components exist, ✅ Beautiful output ready, 🎯 **INTEGRATION NEEDED**

---

## 📊 **Current State Analysis**

### **✅ COMPLETED FOUNDATION** (Phase 0-2 delivered)
- ✅ **Watch Components**: `internal/watch/` package fully implemented with all subsystems
- ✅ **File Watcher**: `internal/watch/watcher/fs_watcher.go` (334 lines) with pattern matching
- ✅ **Debouncer**: `internal/watch/debouncer/file_debouncer.go` (179 lines) with intelligent delays
- ✅ **Coordinator**: `internal/watch/coordinator/watch_coordinator.go` (359 lines) with orchestration
- ✅ **UI System**: Ready for live updates and watch mode display
- ✅ **Test Execution**: Working pipeline ready for triggered execution

### **🎯 TARGET WATCH MODE EXPERIENCE**
```
┌─ Watch Mode Active ───────────────────────────────────────────────────┐
│ 👁️  Watching 23 files for changes...                    Memory: 42MB │
│ 🎯 Smart selection: Only testing related files                       │
└───────────────────────────────────────────────────────────────────────┘

📁 Watching: internal/config/, internal/test/, internal/ui/
⚡ Last run: 2.3s ago (triggered by internal/config/loader.go)

📝 internal/config/loader.go changed
🔍 Analyzing dependencies...
🎯 Running related tests: internal/config/config_test.go

📁 internal/config/config_test.go                              ✅ 20 passed

✨ Watch mode: 20 tests passed in 1.2s
💡 Press 'r' to run all tests, 'q' to quit, 'c' to clear
```

### **🔍 EXISTING WATCH ARCHITECTURE**
```
internal/watch/
├── core/
│   ├── interfaces.go (150 lines) - Watch system interfaces ✅
│   └── types.go (225 lines) - Watch data types ✅
├── watcher/
│   ├── fs_watcher.go (334 lines) - File system monitoring ✅
│   └── patterns.go (100 lines) - Pattern matching ✅
├── debouncer/
│   ├── file_debouncer.go (179 lines) - Event debouncing ✅
│   └── debouncer_test.go (399 lines) - Comprehensive tests ✅
├── coordinator/
│   ├── watch_coordinator.go (359 lines) - Watch orchestration ✅
│   └── coordinator.go (242 lines) - Component coordination ✅
└── watcher/
    └── patterns.go (100 lines) - File pattern utilities ✅
```

### **❌ MISSING INTEGRATION POINTS**
- **Issue**: Watch components exist but not integrated with CLI execution flow
- **Root Cause**: App controller has placeholder watch coordination, no real integration
- **Impact**: `--watch` flag exists but doesn't activate file monitoring

---

## 🔧 **Phase 3 Task Breakdown**

### **3.1 Watch System Integration** (16 hours)

#### **Task 3.1.1**: File watcher CLI integration ✅ **COMPONENTS READY**
- **Violation**: App controller has placeholder watch coordinator adapter, no real file monitoring
- **Fix**: Wire existing file watcher system to CLI execution flow with proper configuration
- **Location**: Enhanced `internal/app/watch_coordinator_adapter.go` and app controller integration
- **Why**: Watch mode requires file system monitoring connected to test execution pipeline
- **Architecture Rule**: Watch system should be loosely coupled through event-driven interfaces
- **Implementation Pattern**: Observer pattern for file events + Adapter pattern for CLI integration
- **New Structure**:
  - Enhanced `internal/app/watch_coordinator_adapter.go` - Full watcher integration (200 lines)
  - `internal/app/watch_executor.go` - Watch mode execution logic (180 lines)
  - Enhanced `internal/app/application_controller.go` - Watch mode support (450 lines)
  - `internal/app/file_pattern_matcher.go` - Pattern matching for file changes (150 lines)
- **Result**: File system monitoring integrated with CLI, watch mode functional
- **Duration**: 6 hours

#### **Task 3.1.2**: Event debouncing integration ✅ **DEBOUNCER READY**
- **Violation**: File change events need intelligent debouncing to prevent excessive test runs
- **Fix**: Integrate existing debouncer system with CLI to handle rapid file changes gracefully
- **Location**: Connect `internal/watch/debouncer/` to app controller and test execution
- **Why**: Debouncing prevents overwhelming the system with rapid file changes during development
- **Architecture Rule**: Event debouncing should be configurable and adaptable to different use cases
- **Implementation Pattern**: Strategy pattern for debounce algorithms + Chain of Responsibility for event filtering
- **New Structure**:
  - Enhanced `internal/app/watch_coordinator_adapter.go` - Debouncer integration (250 lines)
  - `internal/app/debounce_config.go` - Debounce configuration management (120 lines)
  - Enhanced debouncer configuration in CLI flags and config files
  - Integration with event system for real-time debounce status
- **Result**: Intelligent event debouncing preventing excessive test runs
- **Duration**: 4 hours

#### **Task 3.1.3**: Watch coordination integration ✅ **COORDINATOR READY**
- **Violation**: Watch coordinator exists but not connected to CLI orchestration
- **Fix**: Integrate watch coordinator with app controller for complete watch mode orchestration
- **Location**: Wire `internal/watch/coordinator/` to application lifecycle and test execution
- **Why**: Coordination ensures proper startup, shutdown, and state management for watch mode
- **Architecture Rule**: Watch coordination should manage component lifecycle and handle errors gracefully
- **Implementation Pattern**: Facade pattern for watch subsystem + State pattern for watch mode states
- **New Structure**:
  - Enhanced `internal/app/watch_coordinator_adapter.go` - Full coordinator integration (300 lines)
  - `internal/app/watch_state_manager.go` - Watch mode state management (180 lines)
  - Enhanced lifecycle management for watch mode startup/shutdown
  - Integration with dependency container for watch component management
- **Result**: Complete watch mode orchestration with proper lifecycle management
- **Duration**: 6 hours

### **3.2 Watch Mode CLI Integration** (18 hours)

#### **Task 3.2.1**: Watch flag handling enhancement ✅ **FLAGS EXIST**
- **Violation**: `--watch` flag exists but lacks comprehensive watch configuration options
- **Fix**: Enhance CLI flags and configuration for watch mode with granular control options
- **Location**: Enhance `cmd/go-sentinel-cli/cmd/run.go` and `internal/config/args.go`
- **Why**: Watch mode needs configuration for patterns, debounce timing, and behavior options
- **Architecture Rule**: CLI flags should provide intuitive control over watch mode behavior
- **Implementation Pattern**: Builder pattern for watch configuration + Strategy pattern for different watch modes
- **New Structure**:
  - Enhanced `cmd/go-sentinel-cli/cmd/run.go` - Extended watch flags (180 lines)
  - Enhanced `internal/config/args.go` - Watch configuration handling (200 lines)
  - `internal/config/watch_config.go` - Watch-specific configuration (150 lines)
  - Enhanced configuration validation for watch mode options
- **Result**: Comprehensive watch mode configuration through CLI flags and config files
- **Duration**: 4 hours

#### **Task 3.2.2**: Watch mode display ✅ **UI SYSTEM READY**
- **Violation**: Beautiful output system needs watch mode specific display enhancements
- **Fix**: Implement watch mode UI with file monitoring status, change notifications, and interactive controls
- **Location**: Enhance `internal/ui/display/` for watch mode specific display requirements
- **Why**: Watch mode requires real-time status display and user interaction capabilities
- **Architecture Rule**: Watch mode display should be informative and provide clear user feedback
- **Implementation Pattern**: Observer pattern for watch events + State pattern for display modes
- **New Structure**:
  - `internal/ui/display/watch_renderer.go` - Watch mode display implementation (300 lines)
  - `internal/ui/display/file_monitor_display.go` - File monitoring status display (200 lines)
  - `internal/ui/display/interactive_controls.go` - Keyboard controls for watch mode (180 lines)
  - Enhanced `internal/ui/display/app_renderer.go` - Watch mode integration (900 lines)
- **Result**: Beautiful watch mode display with real-time status and interactive controls
- **Duration**: 6 hours

#### **Task 3.2.3**: Watch mode test execution ✅ **TEST SYSTEM READY**
- **Violation**: Test execution needs watch mode specific behavior and smart test selection
- **Fix**: Implement watch mode test execution with intelligent test selection and incremental runs
- **Location**: Enhance test execution pipeline for watch mode with smart selection algorithms
- **Why**: Watch mode should only run tests related to changed files for efficiency
- **Architecture Rule**: Watch mode test execution should be selective and efficient
- **Implementation Pattern**: Strategy pattern for test selection + Observer pattern for change notifications
- **New Structure**:
  - `internal/app/watch_test_executor.go` - Watch mode test execution (250 lines)
  - `internal/test/runner/smart_executor.go` - Intelligent test selection (200 lines)
  - `internal/test/cache/dependency_tracker.go` - File dependency tracking (180 lines)
  - Enhanced test execution with incremental and related test algorithms
- **Result**: Smart test execution that only runs tests related to changed files
- **Duration**: 8 hours

### **3.3 Smart Test Selection** (14 hours)

#### **Task 3.3.1**: Related test detection ✅ **CACHE SYSTEM READY**
- **Violation**: Watch mode needs intelligent detection of which tests to run based on file changes
- **Fix**: Implement dependency analysis and test relationship detection for smart test selection
- **Location**: Create test dependency analysis system and integrate with watch mode
- **Why**: Running only related tests makes watch mode efficient and provides fast feedback
- **Architecture Rule**: Test selection should be based on static analysis and runtime dependency tracking
- **Implementation Pattern**: Visitor pattern for dependency analysis + Graph algorithms for relationship detection
- **New Structure**:
  - `internal/test/analysis/dependency_analyzer.go` - Code dependency analysis (300 lines)
  - `internal/test/analysis/test_mapper.go` - Test-to-code relationship mapping (250 lines)
  - `internal/test/cache/relationship_cache.go` - Dependency relationship caching (200 lines)
  - Enhanced watch mode with intelligent test selection
- **Result**: Accurate detection of which tests should run based on file changes
- **Duration**: 8 hours

#### **Task 3.3.2**: Watch mode optimization ✅ **CACHING READY**
- **Violation**: Watch mode needs performance optimization for large codebases and frequent changes
- **Fix**: Implement caching, incremental analysis, and optimization strategies for watch mode
- **Location**: Enhance watch system with performance optimizations and caching strategies
- **Why**: Watch mode must be responsive and efficient even with large projects
- **Architecture Rule**: Watch mode should use caching and incremental processing for performance
- **Implementation Pattern**: Cache-aside pattern for dependency data + Incremental processing for file analysis
- **New Structure**:
  - Enhanced `internal/test/cache/` - Watch mode specific caching (300 lines additional)
  - `internal/watch/optimization/incremental_analyzer.go` - Incremental dependency analysis (200 lines)
  - `internal/watch/optimization/performance_tracker.go` - Watch mode performance monitoring (150 lines)
  - Enhanced watch coordination with optimization strategies
- **Result**: Optimized watch mode performance with intelligent caching and incremental processing
- **Duration**: 6 hours

---

## 📋 **Phase 3 Deliverable Requirements**

### **Success Criteria**:
- ✅ **Watch Mode Functional**: `--watch` flag activates file monitoring and test execution
- ✅ **Intelligent Selection**: Only runs tests related to changed files
- ✅ **Beautiful Display**: Watch mode has informative UI with real-time status
- ✅ **Interactive Controls**: Keyboard shortcuts for manual test runs and controls
- ✅ **Performance**: Responsive even with large codebases and frequent changes

### **Acceptance Tests**:
```bash
# Must activate watch mode:
go run cmd/go-sentinel-cli/main.go run --watch ./internal/config
# Expected: File monitoring active, tests run on file changes

# Must detect related tests:
# Change internal/config/loader.go -> Should run internal/config/config_test.go
# Change internal/test/runner/executor.go -> Should run related test files

# Must handle rapid changes:
# Rapid file changes should be debounced and handled gracefully

# Must provide interactive controls:
# Press 'r' to run all tests, 'q' to quit, 'c' to clear screen
```

### **Quality Gates**:
- ✅ All existing tests pass (127/127 tests)
- ✅ Watch mode functionality working correctly
- ✅ Smart test selection accurate and efficient
- ✅ No performance degradation
- ✅ Proper error handling and recovery

---

## 🎯 **Implementation Strategy**

### **Phase 3.1: Watch System Integration** (16 hours)
1. **File Watcher Integration** (6 hours) - Connect file monitoring to CLI
2. **Event Debouncing** (4 hours) - Intelligent event handling
3. **Watch Coordination** (6 hours) - Complete orchestration

### **Phase 3.2: CLI Integration** (18 hours)
1. **Enhanced Flags** (4 hours) - Comprehensive watch configuration
2. **Watch Mode Display** (6 hours) - Beautiful UI for watch mode
3. **Watch Test Execution** (8 hours) - Smart test execution pipeline

### **Phase 3.3: Smart Features** (14 hours)
1. **Related Test Detection** (8 hours) - Dependency analysis and test mapping
2. **Watch Mode Optimization** (6 hours) - Performance and caching enhancements

### **Validation After Each Task**:
```bash
# Verify watch mode functionality:
go run cmd/go-sentinel-cli/main.go run --watch ./internal/config
# Touch files and verify test execution
go test ./internal/watch/... -v
go build ./cmd/go-sentinel-cli/...
```

---

## 🚀 **Phase 3 to Phase 4 Transition**

**Once Phase 3 Complete**:
- ✅ Full watch mode functionality with intelligent file monitoring
- ✅ Smart test selection based on file changes
- ✅ Beautiful interactive watch mode display
- ✅ Foundation ready for advanced features and optimization

**Phase 4 Ready**: Advanced features and configuration can begin
- Watch mode ready for advanced pattern matching
- Performance monitoring ready for optimization features
- Configuration system ready for advanced options

**Expected Timeline**: 48 hours (~1 week) to complete Phase 3, then Phase 4 can proceed immediately. 