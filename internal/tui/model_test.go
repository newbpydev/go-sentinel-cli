package tui

import (
	"testing"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

func TestInitialViewRendersLogoAndLayout(t *testing.T) {
	m := NewModel()
	view := m.View()
	if !contains(view, "Go Sentinel") {
		t.Errorf("expected logo in view, got: %s", view)
	}
	if !contains(view, "Test Explorer") {
		t.Errorf("expected layout section 'Test Explorer', got: %s", view)
	}
	if !contains(view, "Status:") {
		t.Errorf("expected status bar, got: %s", view)
	}
}

func TestVimNavigationKeysMoveSelection(t *testing.T) {
	m := NewModelWithTests(3)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.selected != 1 {
		t.Errorf("expected selection to move down to 1, got %d", m.selected)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.selected != 0 {
		t.Errorf("expected selection to move up to 0, got %d", m.selected)
	}
}

func TestSpaceTogglesSelection(t *testing.T) {
	m := NewModelWithTests(2)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	if !m.selectedMap[0] {
		t.Errorf("expected first test to be selected")
	}
}

func TestProgressBarRenders(t *testing.T) {
	m := NewModel()
	m.progress = 0.5
	view := m.View()
	if !contains(view, "50%") {
		t.Errorf("expected progress bar to show 50%%, got: %s", view)
	}
}

func TestMouseClickSelectsRow(t *testing.T) {
	m := NewModelWithTests(3)
	// Simulate a mouse click event on row 2
	mouseMsg := tea.MouseMsg{
		Type: tea.MouseLeft,
		Y:    2, // Row index (assuming 0-based, below logo)
	}
	m, _ = m.Update(mouseMsg)
	if m.selected != 1 {
		t.Errorf("expected row 2 to be selected, got %d", m.selected)
	}
}

func TestFancyListRendersIcons(t *testing.T) {
	m := NewModelWithTests(3)
	// Simulate test statuses: pass, fail, skip
	m.tests = []string{"TestPass", "TestFail", "TestSkip"}
	mStatus := map[int]string{0: "pass", 1: "fail", 2: "skip"}
	// Attach status to model (simulate)
	view := m.ViewWithStatus(mStatus)
	if !contains(view, "✔") {
		t.Errorf("expected pass icon ✔ in view")
	}
	if !contains(view, "✖") {
		t.Errorf("expected fail icon ✖ in view")
	}
	if !contains(view, "➖") {
		t.Errorf("expected skip icon ➖ in view")
	}
}

func TestStatusBarShowsModeAndSelection(t *testing.T) {
	m := NewModelWithTests(3)
	m.selectedMap[0] = true
	m.selectedMap[2] = true
	status := m.StatusBarView("select")
	if !contains(status, "Mode: select") {
		t.Errorf("expected status bar to show mode 'select', got: %s", status)
	}
	if !contains(status, "2 selected") {
		t.Errorf("expected status bar to show selection count, got: %s", status)
	}
	if !contains(status, "[j/k] move") || !contains(status, "[space] select") {
		t.Errorf("expected status bar to show keyboard hints, got: %s", status)
	}
}

func TestLogoRendersWithColor(t *testing.T) {
	m := NewModel()
	logo := m.LogoView()
	if !contains(logo, "\x1b[") {
		t.Errorf("expected ANSI color codes in logo view")
	}
}

func TestVimSlashEntersSearchMode(t *testing.T) {
	m := NewModelWithTests(2)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'/'}})
	if m.mode != "search" {
		t.Errorf("expected mode to be 'search' after pressing /, got: %s", m.mode)
	}
}

func TestVimGGAndGJump(t *testing.T) {
	m := NewModelWithTests(5)
	m.selected = 3
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}) // first g
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}) // second g
	if m.selected != 0 {
		t.Errorf("expected selection to jump to top (0) after gg, got: %d", m.selected)
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if m.selected != 4 {
		t.Errorf("expected selection to jump to bottom (4) after G, got: %d", m.selected)
	}
}

func TestVimYCopiesDetails(t *testing.T) {
	m := NewModelWithTests(2)
	m.selected = 1
	copied := m.CopySelectedDetails()
	if copied == "" || copied != "Test2" {
		t.Errorf("expected to copy selected test details, got: %q", copied)
	}
}

func TestCoverageBarRendersForEachTest(t *testing.T) {
	m := NewModelWithTests(2)
	cover := map[int]float64{0: 0.95, 1: 0.65}
	view := m.ViewWithCoverage(cover)
	if !contains(view, "95%") || !contains(view, "65%") {
		t.Errorf("expected coverage bars to show 95%% and 65%%, got: %s", view)
	}
}

func TestAnimatedProgressUpdatesView(t *testing.T) {
	m := NewModel()
	// Simulate progress increments
	for i := 0; i <= 10; i++ {
		m.progress = float64(i) / 10.0
		view := m.View()
		percent := fmt.Sprintf("%d%%", i*10)
		if !contains(view, percent) {
			t.Errorf("expected progress bar to show %s, got: %s", percent, view)
		}
	}
}

func TestLiveUpdatesProgressAndStatus(t *testing.T) {
	m := NewModelWithTests(2)
	statuses := []map[int]string{
		{0: "fail", 1: "fail"},
		{0: "pass", 1: "fail"},
		{0: "pass", 1: "pass"},
	}
	for i, st := range statuses {
		m.progress = float64(i) / float64(len(statuses)-1)
		view := m.ViewWithStatus(st)
		if i == 0 && !contains(view, "✖") {
			t.Errorf("expected fail icon at step 0")
		}
		if i == 1 && !contains(view, "✔") {
			t.Errorf("expected pass icon at step 1")
		}
		if i == 2 && !contains(view, "✔") && !contains(view, "✖") {
			t.Errorf("expected all pass at step 2")
		}
	}
}

func TestSplitPanelRendersListAndDetail(t *testing.T) {
	m := NewModelWithTests(2)
	m.selected = 1
	detail := "This is the detail panel"
	view := m.SplitPanelView(detail)
	if !contains(view, "Test1") || !contains(view, "Test2") {
		t.Errorf("expected left panel to show test list")
	}
	if !contains(view, detail) {
		t.Errorf("expected right panel to show detail")
	}
}

// TestFuzzySearchFiltersTests documents and tests the robust fuzzy search behavior using lithammer/fuzzysearch.
// - Query 'a' matches all: 'Alpha', 'Beta', 'Gamma' (all contain 'a')
// - Query 'mm' matches only 'Gamma' (subsequence match)
func TestFuzzySearchFiltersTests(t *testing.T) {
	m := NewModelWithTests(3)
	m.tests = []string{"Alpha", "Beta", "Gamma"}
	m.searchQuery = "a"
	filtered := m.FuzzyFilteredTests()
	if len(filtered) != 3 {
		t.Errorf("expected 3 tests to match fuzzy search 'a' (Alpha, Beta, Gamma), got: %d, matches: %v", len(filtered), filtered)
	}
	m.searchQuery = "mm"
	filtered = m.FuzzyFilteredTests()
	if len(filtered) != 1 || filtered[0] != "Gamma" {
		t.Errorf("expected only 'Gamma' to match fuzzy search 'mm', got: %v", filtered)
	}
}


func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || (len(s) > len(substr) && (s[0:len(substr)] == substr || contains(s[1:], substr))))
}
