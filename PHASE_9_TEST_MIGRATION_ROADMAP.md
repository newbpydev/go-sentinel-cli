# PHASE 9: Test Migration and Error Resolution Roadmap

## 🎯 Objective
Complete the CLI refactoring by migrating all remaining tests from `internal/cli/` to their respective modules and fix all build errors.

## 🚨 Current Issues to Fix

### 1. Build Error - Missing Main Function ✅ COMPLETED
- **Issue**: `runtime.main_main·f: function main is undeclared in the main package`
- **Root Cause**: Module expects main package but we have cmd structure
- **Solution**: Update go.mod or create proper main.go

### 2. Orphaned Tests in internal/cli/ 🔄 IN PROGRESS
- **Issue**: 37 test files remain in `internal/cli/` after code migration
- **Impact**: Tests are disconnected from their migrated code
- **Solution**: Migrate tests to match new modular structure

## 📋 PHASE 9 TIER BREAKDOWN

### TIER 9.1: Fix Build Errors (Priority: CRITICAL) ✅ COMPLETED
**Estimated Time**: 30 minutes

#### Task 9.1.1: Resolve Main Package Issue ✅ COMPLETED
- [x] Check if we need a root main.go or update module structure
- [x] Fix the main function declaration error
- [x] Verify build passes: `go build ./...`

#### Task 9.1.2: Verify Module Structure ✅ COMPLETED
- [x] Ensure go.mod is correctly configured
- [x] Check all package declarations are correct
- [x] Run `go mod tidy` to clean dependencies

### TIER 9.2: Test Migration Analysis (Priority: HIGH) ✅ COMPLETED
**Estimated Time**: 45 minutes

#### Task 9.2.1: Categorize Remaining Tests ✅ COMPLETED
Analyze and categorize all 37 test files in `internal/cli/`:

**Core Model Tests** → `pkg/models/` ✅ COMPLETED
- [x] `types_test.go` (387 lines) → Migrated to `pkg/models/types_test.go`
- [x] `models_test.go` (199 lines) → Migrated to `pkg/models/models_test.go`

**Configuration Tests** → `internal/config/` ✅ COMPLETED
- [x] `config_test.go` (381 lines) → Migrated to `internal/config/config_test.go`
- [x] `cli_args_test.go` (159 lines) → Migrated to `internal/config/cli_args_test.go`

**Test Processing Tests** → `internal/test/processor/`
- [ ] `processor_test.go` (386 lines)
- [ ] `parser_test.go` (219 lines)
- [ ] `source_extractor_test.go` (528 lines)

**Test Runner Tests** → `internal/test/runner/`
- [ ] `test_runner_test.go` (100 lines)
- [ ] `parallel_runner_test.go` (495 lines)

**Watch System Tests** → `internal/watch/`
- [ ] `intelligent_watch_test.go` (213 lines)
- [ ] `intelligent_watch_stress_test.go` (234 lines)

**UI/Display Tests** → `internal/ui/`
- [ ] `display_test.go` (233 lines)
- [ ] `test_display_test.go` (316 lines)
- [ ] `suite_display_test.go` (308 lines)
- [ ] `colors_test.go` (150 lines)
- [ ] `failed_tests_test.go` (290 lines)
- [ ] `incremental_renderer_test.go` (580 lines)
- [ ] `watch_ui_test.go` (148 lines)

**App Controller Tests** → `internal/app/`
- [ ] `app_controller_test.go` (472 lines)
- [ ] `integration_test.go` (220 lines)

**Performance/Benchmark Tests** → `internal/test/benchmarks/`
- [ ] `execution_bench_test.go` (452 lines)
- [ ] `integration_bench_test.go` (421 lines)
- [ ] `filesystem_bench_test.go` (342 lines)
- [ ] `rendering_bench_test.go` (364 lines)
- [ ] `performance_test.go` (369 lines)

**Error Recovery Tests** → `internal/test/recovery/`
- [ ] `error_recovery_test.go` (990 lines)

**Stream/Summary Tests** → `internal/test/processor/`
- [ ] `stream_test.go` (192 lines)
- [ ] `summary_test.go` (176 lines)

### TIER 9.3: Create Missing Test Directories (Priority: HIGH) ✅ COMPLETED
**Estimated Time**: 15 minutes

#### Task 9.3.1: Create Test Package Structure ✅ COMPLETED
- [x] `mkdir -p internal/test/benchmarks`
- [x] `mkdir -p internal/test/recovery`
- [x] `mkdir -p internal/ui/display/tests`
- [x] `mkdir -p internal/ui/renderer/tests`
- [x] `mkdir -p internal/watch/tests`

### TIER 9.4: Migrate Core Tests (Priority: HIGH) ✅ COMPLETED
**Estimated Time**: 2 hours

#### Task 9.4.1: Migrate Model Tests ✅ COMPLETED
- [x] Move `types_test.go` → `pkg/models/types_test.go`
- [x] Move `models_test.go` → `pkg/models/models_test.go`
- [x] Update import paths in test files
- [x] Run: `go test ./pkg/models/...`

#### Task 9.4.2: Migrate Config Tests ✅ COMPLETED
- [x] Move `config_test.go` → `internal/config/config_test.go`
- [x] Move `cli_args_test.go` → `internal/config/cli_args_test.go`
- [x] Update import paths and test references
- [x] Run: `go test ./internal/config/...`

#### Task 9.4.3: Migrate Processor Tests 🔄 NEXT
- [ ] Move `processor_test.go` → `internal/test/processor/processor_test.go`
- [ ] Move `parser_test.go` → `internal/test/processor/parser_test.go`
- [ ] Move `source_extractor_test.go` → `internal/test/processor/source_extractor_test.go`
- [ ] Move `stream_test.go` → `internal/test/processor/stream_test.go`
- [ ] Move `summary_test.go` → `internal/test/processor/summary_test.go`
- [ ] Update import paths and test references
- [ ] Run: `go test ./internal/test/processor/...`

### TIER 9.5: Migrate Runner Tests (Priority: HIGH)
**Estimated Time**: 1 hour

#### Task 9.5.1: Migrate Test Runner Tests
- [ ] Move `test_runner_test.go` → `internal/test/runner/test_runner_test.go`
- [ ] Move `parallel_runner_test.go` → `internal/test/runner/parallel_runner_test.go`
- [ ] Update import paths and test references
- [ ] Run: `go test ./internal/test/runner/...`

### TIER 9.6: Migrate Watch System Tests (Priority: MEDIUM)
**Estimated Time**: 45 minutes

#### Task 9.6.1: Migrate Watch Tests
- [ ] Move `intelligent_watch_test.go` → `internal/watch/intelligent_watch_test.go`
- [ ] Move `intelligent_watch_stress_test.go` → `internal/watch/stress_test.go`
- [ ] Update import paths and test references
- [ ] Run: `go test ./internal/watch/...`

### TIER 9.7: Migrate UI Tests (Priority: MEDIUM)
**Estimated Time**: 1.5 hours

#### Task 9.7.1: Migrate Display Tests
- [ ] Move `display_test.go` → `internal/ui/display/display_test.go`
- [ ] Move `test_display_test.go` → `internal/ui/display/test_display_test.go`
- [ ] Move `suite_display_test.go` → `internal/ui/display/suite_display_test.go`
- [ ] Move `colors_test.go` → `internal/ui/colors/colors_test.go`
- [ ] Move `failed_tests_test.go` → `internal/ui/display/failed_tests_test.go`
- [ ] Move `watch_ui_test.go` → `internal/ui/display/watch_ui_test.go`

#### Task 9.7.2: Migrate Renderer Tests
- [ ] Move `incremental_renderer_test.go` → `internal/ui/renderer/incremental_renderer_test.go`
- [ ] Update import paths and test references
- [ ] Run: `go test ./internal/ui/...`

### TIER 9.8: Migrate Benchmark Tests (Priority: LOW)
**Estimated Time**: 1 hour

#### Task 9.8.1: Migrate Performance Tests
- [ ] Move `execution_bench_test.go` → `internal/test/benchmarks/execution_bench_test.go`
- [ ] Move `integration_bench_test.go` → `internal/test/benchmarks/integration_bench_test.go`
- [ ] Move `filesystem_bench_test.go` → `internal/test/benchmarks/filesystem_bench_test.go`
- [ ] Move `rendering_bench_test.go` → `internal/test/benchmarks/rendering_bench_test.go`
- [ ] Move `performance_test.go` → `internal/test/benchmarks/performance_test.go`
- [ ] Update import paths and test references
- [ ] Run: `go test ./internal/test/benchmarks/...`

### TIER 9.9: Migrate Error Recovery Tests (Priority: MEDIUM)
**Estimated Time**: 45 minutes

#### Task 9.9.1: Migrate Error Recovery
- [ ] Move `error_recovery_test.go` → `internal/test/recovery/error_recovery_test.go`
- [ ] Update import paths and test references
- [ ] Run: `go test ./internal/test/recovery/...`

### TIER 9.10: Migrate App Controller Tests (Priority: HIGH)
**Estimated Time**: 1 hour

#### Task 9.10.1: Migrate App Tests
- [ ] Move `app_controller_test.go` → `internal/app/app_controller_test.go`
- [ ] Move `integration_test.go` → `internal/app/integration_test.go`
- [ ] Update import paths and test references
- [ ] Run: `go test ./internal/app/...`

### TIER 9.11: Final Cleanup and Validation (Priority: CRITICAL)
**Estimated Time**: 30 minutes

#### Task 9.11.1: Clean Up Old CLI Directory
- [ ] Verify all tests have been migrated
- [ ] Remove empty test files from `internal/cli/`
- [ ] Update any remaining references

#### Task 9.11.2: Full System Validation
- [ ] Run: `go build ./...`
- [ ] Run: `go test ./...`
- [ ] Run: `go test ./... -race`
- [ ] Run: `go test ./... -bench=.`
- [ ] Verify test coverage: `go test ./... -cover`

## 📊 Success Metrics

### Build Health ✅ COMPLETED
- [x] `go build ./...` passes without errors
- [x] All packages compile successfully
- [x] No import cycle errors

### Test Health 🔄 IN PROGRESS
- [x] Core model tests pass in their new locations
- [x] Config tests pass in their new locations
- [ ] All migrated tests pass in their new locations
- [ ] Test coverage maintained or improved (target: ≥90%)
- [ ] No orphaned test files remain in `internal/cli/`
- [ ] Benchmark tests run successfully

### Code Quality
- [ ] All linting passes: `golangci-lint run`
- [ ] No race conditions: `go test ./... -race`
- [ ] Performance benchmarks stable

## 🚨 Risk Mitigation

### High-Risk Areas
1. **Import Path Updates**: Many tests will have broken imports
2. **Test Dependencies**: Tests may depend on unexported functions
3. **Package Visibility**: Some tests may need to be internal to packages

### Mitigation Strategies
1. **Incremental Migration**: Migrate one tier at a time
2. **Continuous Testing**: Run tests after each migration
3. **Backup Strategy**: Keep original files until migration confirmed
4. **Interface Extraction**: Create test interfaces for unexported dependencies

## 📈 Estimated Timeline
- **Total Estimated Time**: 8-10 hours
- **Critical Path**: TIER 9.1 → 9.4 → 9.10 → 9.11
- **Parallel Work**: TIER 9.6, 9.7, 9.8, 9.9 can be done in parallel after 9.4

## 🎉 Completion Criteria
- [ ] Zero build errors
- [ ] All tests pass in new locations
- [ ] No test files remain in `internal/cli/`
- [ ] Test coverage ≥90%
- [ ] All benchmarks run successfully
- [ ] Documentation updated

## 📊 Progress Summary
**COMPLETED**: TIER 9.1, 9.2, 9.3, 9.4 (Core models and config tests)
**NEXT**: TIER 9.4.3 (Processor tests migration)
**REMAINING**: 33 test files to migrate

This roadmap ensures a systematic migration of all tests while maintaining code quality and test coverage throughout the process. 