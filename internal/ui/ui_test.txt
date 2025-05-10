package ui

// import (
// 	"bytes"
// 	"testing"
// )

// // 4.1.1: Test: Display summary with color (ANSI)
// func TestDisplaySummaryWithColor(t *testing.T) {
// 	output := &bytes.Buffer{}
// 	results := []TestResult{
// 		{Package: "pkg/foo", Passed: true, Summary: "ok   pkg/foo  0.05s"},
// 		{Package: "pkg/bar", Passed: false, Summary: "FAIL pkg/bar  0.10s"},
// 	}
// 	DisplaySummary(output, results)
// 	outStr := output.String()
// 	if !containsANSI(outStr) {
// 		t.Errorf("expected ANSI color codes in output, got: %q", outStr)
// 	}
// 	if !bytes.Contains(output.Bytes(), []byte("pkg/foo")) || !bytes.Contains(output.Bytes(), []byte("pkg/bar")) {
// 		t.Errorf("expected both package summaries in output, got: %q", outStr)
// 	}
// }

// // 4.1.2: Test: Keybindings (Enter, f, q)
// func TestKeybindings(t *testing.T) {
// 	ui := NewUI()
// 	// Simulate keypresses: Enter, 'f', 'q'
// 	keys := []rune{'\n', 'f', 'q'}
// 	var handled []string
// 	for _, k := range keys {
// 		action := ui.HandleKey(k)
// 		handled = append(handled, action)
// 	}
// 	if handled[0] != "rerun" || handled[1] != "filter" || handled[2] != "quit" {
// 		t.Errorf("unexpected key actions: %v", handled)
// 	}
// }

// // 4.1.3: Test: Filter failures mode
// func TestFilterFailuresMode(t *testing.T) {
// 	ui := NewUI()
// 	results := []TestResult{
// 		{Package: "pkg/foo", Passed: true},
// 		{Package: "pkg/bar", Passed: false},
// 	}
// 	ui.SetResults(results)
// 	ui.SetFilterFailures(true)
// 	filtered := ui.VisibleResults()
// 	if len(filtered) != 1 || filtered[0].Package != "pkg/bar" {
// 		t.Errorf("expected only failing package, got: %v", filtered)
// 	}
// }

// // 4.1.4: Test: Show code context for failed tests
// func TestShowCodeContextForFailures(t *testing.T) {
// 	ui := NewUI()
// 	failResult := TestResult{
// 		Package: "pkg/bar", Passed: false, File: "main.go", Line: 42,
// 		Message: "expected true, got false",
// 	}
// 	ctx := ui.CodeContextForFailure(failResult)
// 	if ctx == "" {
// 		t.Error("expected code context for failure, got empty string")
// 	}
// }

// // 4.1.5: Test: UI updates on each run without exit
// func TestUIUpdatesOnRun(t *testing.T) {
// 	ui := NewUI()
// 	results1 := []TestResult{{Package: "pkg/foo", Passed: true}}
// 	results2 := []TestResult{{Package: "pkg/bar", Passed: false}}
// 	ui.SetResults(results1)
// 	if ui.VisibleResults()[0].Package != "pkg/foo" {
// 		t.Errorf("expected foo in first update")
// 	}
// 	ui.SetResults(results2)
// 	if ui.VisibleResults()[0].Package != "pkg/bar" {
// 		t.Errorf("expected bar in second update")
// 	}
// }

// // 4.1.6: Test: Copy failed test information to clipboard ('c' key)
// func TestCopyFailureToClipboard(t *testing.T) {
// 	ui := NewUI()
// 	failResult := TestResult{Package: "pkg/bar", Passed: false, Message: "failure"}
// 	ui.SetResults([]TestResult{failResult})
// 	copied := ui.CopyFailure()
// 	if copied == "" || copied != failResult.Message {
// 		t.Errorf("expected to copy failure message, got: %q", copied)
// 	}
// }

// // 4.1.7: Test: Interactive test selection and copying ('C' key, space for selection)
// func TestInteractiveTestSelectionAndCopying(t *testing.T) {
// 	ui := NewUI()
// 	results := []TestResult{{Package: "pkg/foo", Passed: false}, {Package: "pkg/bar", Passed: false}}
// 	ui.SetResults(results)
// 	ui.SelectTest(0)
// 	ui.SelectTest(1)
// 	selected := ui.SelectedTests()
// 	if len(selected) != 2 {
// 		t.Errorf("expected 2 selected tests, got %d", len(selected))
// 	}
// 	copied := ui.CopySelectedFailures()
// 	if copied == "" || !bytes.Contains([]byte(copied), []byte("pkg/foo")) || !bytes.Contains([]byte(copied), []byte("pkg/bar")) {
// 		t.Errorf("expected copied string to include both failures, got: %q", copied)
// 	}
// }

// // --- Helpers and stubs ---
// func containsANSI(s string) bool {
// 	return bytes.Contains([]byte(s), []byte("\x1b["))
// }
