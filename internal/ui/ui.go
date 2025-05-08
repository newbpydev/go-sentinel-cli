package ui

import (
	"bytes"
	"fmt"
)

type TestResult struct {
	Package string
	Passed  bool
	Summary string
	File    string
	Line    int
	Message string
}

type UI struct {
	results         []TestResult
	filterFailures  bool
	selected        map[int]bool
}

func NewUI() *UI {
	return &UI{selected: make(map[int]bool)}
}

func (ui *UI) SetResults(results []TestResult) {
	ui.results = results
}

func (ui *UI) SetFilterFailures(on bool) {
	ui.filterFailures = on
}

func (ui *UI) VisibleResults() []TestResult {
	if !ui.filterFailures {
		return ui.results
	}
	var filtered []TestResult
	for _, r := range ui.results {
		if !r.Passed {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

func DisplaySummary(output *bytes.Buffer, results []TestResult) {
	for _, r := range results {
		if r.Passed {
			output.WriteString("\x1b[32m" + r.Summary + "\x1b[0m\n") // Green
		} else {
			output.WriteString("\x1b[31m" + r.Summary + "\x1b[0m\n") // Red
		}
	}
}

func (ui *UI) HandleKey(key rune) string {
	switch key {
	case '\n':
		return "rerun"
	case 'f':
		ui.filterFailures = !ui.filterFailures
		return "filter"
	case 'q':
		return "quit"
	default:
		return ""
	}
}

func (ui *UI) CodeContextForFailure(result TestResult) string {
	if result.File != "" && result.Line > 0 {
		return fmt.Sprintf("%s:%d: %s", result.File, result.Line, result.Message)
	}
	return result.Message
}

func (ui *UI) CopyFailure() string {
	for _, r := range ui.results {
		if !r.Passed {
			return r.Message
		}
	}
	return ""
}

// Toggle selection for a given index
func (ui *UI) SelectTest(idx int) bool {
	if ui.selected[idx] {
		delete(ui.selected, idx)
		return false // deselected
	} else {
		ui.selected[idx] = true
		return true // selected
	}
}

// Select all given indices
func (ui *UI) SelectAll(indices []int) {
	for _, idx := range indices {
		ui.selected[idx] = true
	}
}

// Deselect all
func (ui *UI) DeselectAll() {
	for k := range ui.selected {
		delete(ui.selected, k)
	}
}


func (ui *UI) SelectedTests() []TestResult {
	var sel []TestResult
	for i, r := range ui.results {
		if ui.selected[i] {
			sel = append(sel, r)
		}
	}
	return sel
}

// Clear all selection
func (ui *UI) ClearSelection() {
	ui.selected = make(map[int]bool)
}

func (ui *UI) CopySelectedFailures() string {
	var buf bytes.Buffer
	for i, r := range ui.results {
		if ui.selected[i] && !r.Passed {
			buf.WriteString(r.Package + ": " + r.Message + "\n")
		}
	}
	return buf.String()
}
