# 🎨 Go Sentinel CLI Visual Guidelines

## 📋 **VISUAL OUTPUT STANDARDS** - **REQUIRED FOR ALL PHASES**

**Objective**: Establish standardized visual output ensuring consistent three-part structure across all CLI modes (normal, watch, verbose, etc.) following Vitest aesthetic patterns.

**Reference Implementation**: This document defines the **AUTHORITATIVE** visual standards based on EXACT terminal output analysis. All phases must implement this EXACT formatting without deviation.

## ⚠️ **CRITICAL IMPLEMENTATION NOTE**

**This specification is based on PIXEL-PERFECT analysis of reference terminal output. Every character, space, and symbol has been documented exactly as shown. NO creative interpretation is allowed.**

**Key Requirements:**
- ✅ **Character-perfect matching** - Every icon, space, separator must match exactly
- ✅ **Spacing precision** - 2-space indents, 4-space error details, exact alignment
- ✅ **Unicode exactness** - Use specified Unicode points, not similar-looking characters
- ✅ **Format strings** - Follow documented format patterns without modification
- ✅ **Section headers** - 110+ ─ characters with exact centering
- ✅ **Error context** - 5-line code snippets with right-aligned line numbers + ^ pointer

---

## 🎯 **CORE THREE-PART STRUCTURE STANDARD**

**ALL modes must implement this EXACT structure (based on reference terminal output):**

### **Part 1: Individual Test Execution** (Real-time test results)
```
  ✓ TestFileSystemOperations/read_nonexistent_file 0ms
  ✓ TestFileSystemOperations/permission_denied 0ms
  ✓ TestFileSystemOperations 0ms
  ✗ TestEnvironmentDependencies 0ms
  ✓ TestFileWatcher/detects_changes_to_implementation_files 100ms
  ✓ TestFileWatcher/respects_ignore_patterns 2120ms
  ✓ TestFileWatcher 2330ms
  ⃠ TestConditionalSkip 0ms
```

### **Part 2: File Summary & Detailed Results**
```
cli_test.go (127 tests) 0ms 0 MB heap used
  ✓ Suite passed (127 tests)

stress_tests_test.go (48 tests | 26 failed) 0ms 0 MB heap used
  ✓ TestBasicPass 0ms
  ✗ TestBasicFail 0ms
    → Expected 1+1 to equal 3, but got 2
    at basic_failures_test.go:20
  ⃠ TestSkipped 0ms
  ✓ TestMixedSubtests/passing_subtest 0ms
  ✗ TestMixedSubtests/failing_subtest 0ms
    → This subtest is designed to fail
    at basic_failures_test.go:39
```

### **Part 3: Failed Tests Detail Section & Summary**
```
──────────────────────────────────────────────────────────────────────────────────────────────────────────────────
                                                 Failed Tests 26
──────────────────────────────────────────────────────────────────────────────────────────────────────────────────
 FAIL  basic_failures_test.go > TestBasicFail
AssertionError: Expected 1+1 to equal 3, but got 2
↳ basic_failures_test.go:20:1
     18|                t.Log("This should not happen")
     19|        } else {
     20|                t.Errorf("Expected 1+1 to equal 3, but got %d", 1+1)
       | ^
     21|        }
     22| }

────────────────────────────────────────────────── Test Summary ──────────────────────────────────────────────────

Test Files: 1 passed | 1 failed (2)
Tests: 142 passed | 26 failed | 7 skipped (175)
Start at: 12:17:10
End at: 12:17:23
Duration: 12.96s (setup 7.61s, tests 4.36s, teardown 979ms)

──────────────────────────────────────────────────────────────────────────────────────────────────────────────────

⏱️  Tests completed in 13.1472234s
```

---

## 🎨 **VISUAL DESIGN STANDARDS**

### **Color Scheme** (Vitest-inspired)
```
Success:   #10b981 (emerald-500)    ✅ Passed tests, success messages
Error:     #ef4444 (red-500)        ❌ Failed tests, error messages  
Warning:   #f59e0b (amber-500)      ⚠️  Warnings, skipped tests
Info:      #3b82f6 (blue-500)       ℹ️  General information
Muted:     #6b7280 (gray-500)       📝 Secondary text, metadata
Accent:    #8b5cf6 (violet-500)     🎯 Highlights, special status
```

### **Icon System** (EXACT Unicode characters from reference output)
```
Status Icons (Primary):
✓         - Passed test (Unicode U+2713 CHECK MARK)
✗         - Failed test (Unicode U+2717 BALLOT X) 
⃠         - Skipped test (Unicode U+20E0 COMBINING ENCLOSING CIRCLE BACKSLASH)
→         - Error detail pointer (Unicode U+2192 RIGHTWARDS ARROW)
↳         - File location indicator (Unicode U+21B3 DOWNWARDS ARROW WITH TIP RIGHTWARDS)
^         - Error line pointer (Caret symbol)
|         - Line separator in code context
⏱️        - Final timing summary (Unicode U+23F1 STOPWATCH)

ASCII Fallbacks (Limited terminals):
[P] / ✓   - Passed test
[F] / ✗   - Failed test  
[S] / ⃠   - Skipped test
->  / →   - Error detail pointer
\-> / ↳   - File location indicator
^   / ^   - Error line pointer (same)
|   / |   - Line separator (same)
[T] / ⏱️  - Timing summary
```

### **Typography & Spacing** (EXACT formatting from reference)
```
Individual Tests:    2-space indent + icon + space + test name + space + timing
                    "  ✓ TestName 0ms"

File Summary:       filename + space + (stats) + space + timing + space + memory
                    "cli_test.go (127 tests) 0ms 0 MB heap used"

Suite Status:       2-space indent + icon + space + "Suite passed" + space + (count)
                    "  ✓ Suite passed (127 tests)"

Test Details:       2-space indent + icon + space + test name + space + timing
                    "  ✗ TestBasicFail 0ms"

Error Details:      4-space indent + → + space + error message
                    "    → Expected 1+1 to equal 3, but got 2"

File Location:      4-space indent + "at" + space + filename:line
                    "    at basic_failures_test.go:20"

Section Headers:    ─ characters (Unicode U+2500 BOX DRAWINGS LIGHT HORIZONTAL)
                    110+ characters wide for full terminal width

Failed Test Header: " FAIL  filename > TestName"
Error Type:         "AssertionError:", "TestFailure:", "Panic:"
Location Pointer:   "↳ filename:line:column"

Code Context:       Right-aligned line numbers with | separator
                    "     18|                t.Log(...)"
                    "       | ^" (caret points to error column)
```

---

## 📋 **MODE-SPECIFIC IMPLEMENTATIONS** (Reference-Based)

### **Normal Mode** (EXACT format from reference)
```
  ✓ TestFileSystemOperations/read_nonexistent_file 0ms
  ✓ TestFileSystemOperations/permission_denied 0ms
  ✓ TestFileSystemOperations 0ms
  ✗ TestEnvironmentDependencies 0ms

cli_test.go (127 tests) 0ms 0 MB heap used
  ✓ Suite passed (127 tests)

stress_tests_test.go (48 tests | 26 failed) 0ms 0 MB heap used
  ✓ TestBasicPass 0ms
  ✗ TestBasicFail 0ms
    → Expected 1+1 to equal 3, but got 2
    at basic_failures_test.go:20

──────────────────────────────────────────────────────────────────────────────────────────────────────────────────
                                                 Failed Tests 26
──────────────────────────────────────────────────────────────────────────────────────────────────────────────────
 FAIL  basic_failures_test.go > TestBasicFail
AssertionError: Expected 1+1 to equal 3, but got 2
↳ basic_failures_test.go:20:1
     18|                t.Log("This should not happen")
     19|        } else {
     20|                t.Errorf("Expected 1+1 to equal 3, but got %d", 1+1)
       | ^
     21|        }
     22| }

────────────────────────────────────────────────── Test Summary ──────────────────────────────────────────────────

Test Files: 1 passed | 1 failed (2)
Tests: 142 passed | 26 failed | 7 skipped (175)
Start at: 12:17:10
End at: 12:17:23
Duration: 12.96s (setup 7.61s, tests 4.36s, teardown 979ms)

──────────────────────────────────────────────────────────────────────────────────────────────────────────────────

⏱️  Tests completed in 13.1472234s
```

### **Watch Mode** (Same format with watch-specific context)
```
[Watch Context Header - if needed]
📁 Changed: internal/config/loader.go
⚡ Re-running affected tests...

  ✓ TestLoadConfig_ValidFile 45ms
  ✓ TestLoadConfig_InvalidPath 12ms
  ✓ TestValidateConfig_Success 8ms

config_test.go (3 tests) 65ms 0 MB heap used
  ✓ Suite passed (3 tests)

────────────────────────────────────────────────── Test Summary ──────────────────────────────────────────────────

Test Files: 1 passed | 0 failed (1)
Tests: 3 passed | 0 failed | 0 skipped (3)
Duration: 65ms (setup 12ms, tests 45ms, teardown 8ms)

──────────────────────────────────────────────────────────────────────────────────────────────────────────────────

⏱️  Watch run completed in 0.0652134s | 👀 Still watching...
```

### **Verbose Mode** (Enhanced detail but same format)
```
  ✓ TestFileSystemOperations/create_temp_file 5ms
  ✓ TestFileSystemOperations/read_nonexistent_file 0ms
  ✓ TestFileSystemOperations/permission_denied 0ms
  ✓ TestFileSystemOperations 5ms
    → Subtest execution summary
    → create_temp_file: Created file successfully
    → read_nonexistent_file: Handled error correctly
    → permission_denied: Proper permission handling

operations_test.go (4 tests) 5ms 0 MB heap used
  ✓ Suite passed (4 tests)
    → Memory usage: 0 allocations
    → Performance: 0.8 tests/ms average

────────────────────────────────────────────────── Test Summary ──────────────────────────────────────────────────

Test Files: 1 passed | 0 failed (1)
Tests: 4 passed | 0 failed | 0 skipped (4)
Duration: 5ms (setup 1ms, tests 3ms, teardown 1ms)
Performance: 800 tests/second average

──────────────────────────────────────────────────────────────────────────────────────────────────────────────────

⏱️  Tests completed in 0.0051234s
```

---

## 🔧 **PROGRESSIVE ENHANCEMENT STANDARDS**

### **Terminal Capability Detection**
```go
type TerminalCapability int

const (
    ASCII_ONLY TerminalCapability = iota    // No colors, ASCII icons only
    BASIC_COLOR                             // 8/16 colors, basic icons
    EXTENDED_COLOR                          // 256 colors, extended icons
    TRUE_COLOR                              // 24-bit RGB, full Unicode
)
```

### **Fallback Hierarchy**
1. **TrueColor + Unicode**: Full Vitest experience
2. **256 Color + Unicode**: Rich colors with full icons
3. **Basic Color + Unicode**: Limited colors, full icons
4. **ASCII Only**: Monochrome with ASCII character fallbacks

---

## 📐 **EXACT FORMATTING STANDARDS** (Based on Reference Output)

### **1. Individual Test Execution Format**
```
Format:  {2_spaces}{icon} {test_name} {timing}
Example: "  ✓ TestFileSystemOperations/read_nonexistent_file 0ms"
         "  ✗ TestEnvironmentDependencies 0ms" 
         "  ⃠ TestConditionalSkip 0ms"

Subtest Indentation: Same as parent (no additional indent)
Timing Format: Integer + "ms" (no decimals, no parentheses)
```

### **2. File Summary Format**
```
Format:  {filename} ({test_count} tests[ | {failed_count} failed]) {timing} {memory}
Example: "cli_test.go (127 tests) 0ms 0 MB heap used"
         "stress_tests_test.go (48 tests | 26 failed) 0ms 0 MB heap used"

Suite Status: "  ✓ Suite passed ({count} tests)"
Memory Format: "0 MB heap used" (always 0 MB in examples)
```

### **3. Detailed Test Results Within File**
```
Format:  {2_spaces}{icon} {test_name} {timing}
Example: "  ✓ TestBasicPass 0ms"
         "  ✗ TestBasicFail 0ms"
         "  ⃠ TestSkipped 0ms"

Error Details:    "    → {error_message}"
File Location:    "    at {filename}:{line}"
Example: "    → Expected 1+1 to equal 3, but got 2"
         "    at basic_failures_test.go:20"
```

### **4. Failed Tests Detail Section**
```
Section Header: 110+ ─ characters, centered text
"──────────────────────────────────────────────────────────────────────────────────────────────────────────────────"
"                                                 Failed Tests 26"
"──────────────────────────────────────────────────────────────────────────────────────────────────────────────────"

Test Entry Format:
" FAIL  {filename} > {test_name}"
"{ErrorType}: {error_message}"
"↳ {filename}:{line}:{column}"

Code Context (5 lines shown):
"     {line-2}|                {code}"
"     {line-1}|        } else {"
"     {line}  |                {error_line}"
"       | ^"
"     {line+1}|        }"

Test Counter: "─────────────────────────────────────────────────── Test {N}/{Total} ────────────────────────────────────────────────────"
```

### **5. Test Summary Section**
```
Header: "────────────────────────────────────────────────── Test Summary ──────────────────────────────────────────────────"

Statistics Format:
"Test Files: {passed} passed | {failed} failed ({total})"
"Tests: {passed} passed | {failed} failed | {skipped} skipped ({total})"
"Start at: {HH:MM:SS}"
"End at: {HH:MM:SS}"
"Duration: {X.XXs} (setup {X.XXs}, tests {X.XXs}, teardown {XXXms})"

Final separator: 110+ ─ characters
Final timing: "⏱️  Tests completed in {X.XXXXXXX}s"
```

### **6. Spacing Rules**
```
Between sections: 1 empty line
Between test files: 1 empty line  
Between failed test details: 1 empty line
Before/after section headers: No empty lines
Code context indentation: Right-aligned line numbers with | separator
Error pointer: Spaces to align ^ under error column
```

---

## 🎯 **CONSISTENCY VALIDATION CHECKLIST**

### **Required for ALL Modes (Based on Reference Output):**
- [ ] **EXACT three-part structure**: Individual Tests → File Summaries → Failed Tests Detail + Summary
- [ ] **Precise icon usage**: ✓ ✗ ⃠ → ↳ ^ | ⏱️ (exact Unicode characters)
- [ ] **Exact spacing**: 2-space indents, 4-space error details, right-aligned line numbers
- [ ] **File summary format**: "filename (X tests[ | Y failed]) Zms 0 MB heap used"
- [ ] **Failed tests section**: 110+ ─ characters, centered headers, detailed code context
- [ ] **Summary format**: Pipe-separated statistics, precise timing breakdown
- [ ] **Progressive enhancement**: Full Unicode → ASCII fallbacks for limited terminals
- [ ] **Error formatting**: → for details, ↳ for locations, ^ for line pointers
- [ ] **Code context**: 5-line snippets with | separators and ^ pointer alignment
- [ ] **Timing precision**: Integer ms for tests, decimal seconds for summaries

---

## 🔗 **IMPLEMENTATION REFERENCE**

### **Required Components (For Reference-Based Implementation)**
- `internal/ui/icons/reference_icons.go` - EXACT Unicode icons (✓ ✗ ⃠ → ↳ ^ |)
- `internal/ui/display/test_execution_renderer.go` - Individual test output formatting
- `internal/ui/display/file_summary_renderer.go` - File summary with heap usage
- `internal/ui/display/failed_tests_renderer.go` - Detailed failure section with code context
- `internal/ui/display/summary_renderer.go` - Final summary with pipe-separated stats
- `internal/ui/display/spacing_manager.go` - Precise spacing and indentation control
- `internal/ui/display/terminal_detector.go` - Unicode capability detection
- `internal/ui/display/code_context_formatter.go` - 5-line code snippets with line numbers

### **Phase Implementation Requirements**
- **Phase 2**: Implement foundation components and three-part structure
- **Phase 3**: Apply guidelines to watch mode with live updates
- **Phase 4**: Extend to advanced features while maintaining consistency
- **Phase 5**: Production hardening with full fallback support

---

**ALL PHASES MUST REFERENCE AND IMPLEMENT THESE GUIDELINES WITH PIXEL-PERFECT PRECISION**

This document serves as the **single source of truth** for visual output standards across the entire Go Sentinel CLI project. The formatting specifications are based on detailed analysis of reference terminal output and must be implemented EXACTLY as documented - no approximations, no creative interpretations, no "close enough" implementations.

**Implementation Verification**: Every component must produce output that is character-for-character identical to the reference patterns shown in this document. 