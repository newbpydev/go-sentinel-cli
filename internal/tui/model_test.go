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

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || (len(s) > len(substr) && (s[0:len(substr)] == substr || contains(s[1:], substr))))
}
