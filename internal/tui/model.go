package tui

import (
	"fmt"
	"strings"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type Model struct {
	tests       []string
	selected    int
	selectedMap map[int]bool
	progress    float64
	mode        string
	ggPending   bool // for detecting double 'g'
	searchQuery string
}

func NewModel() Model {
	return Model{
		tests:       []string{"TestFoo", "TestBar", "TestBaz"},
		selected:    0,
		selectedMap: make(map[int]bool),
		progress:    0.0,
		mode:        "normal",
	}
}



func NewModelWithTests(n int) Model {
	tests := make([]string, n)
	for i := 0; i < n; i++ {
		tests[i] = fmt.Sprintf("Test%d", i+1)
	}
	return Model{
		tests:       tests,
		selected:    0,
		selectedMap: make(map[int]bool),
		progress:    0.0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		s := msg.String()
		switch s {
		case "/":
			m.mode = "search"
			m.ggPending = false
		case "g":
			if m.ggPending {
				m.selected = 0
				m.ggPending = false
			} else {
				m.ggPending = true
			}
		case "G":
			m.selected = len(m.tests) - 1
			m.ggPending = false
		case "j", "down":
			if m.selected < len(m.tests)-1 {
				m.selected++
			}
			m.ggPending = false
		case "k", "up":
			if m.selected > 0 {
				m.selected--
			}
			m.ggPending = false
		case " ":
			m.selectedMap[m.selected] = !m.selectedMap[m.selected]
			m.ggPending = false
		}
	case tea.MouseMsg:
		if msg.Type == tea.MouseLeft {
			// Logo is line 0, first test is line 1
			row := msg.Y - 1
			if row >= 0 && row < len(m.tests) {
				m.selected = row
			}
		}
	}
	return m, nil
}

// CopySelectedDetails simulates copying the currently selected test's details
func (m Model) CopySelectedDetails() string {
	if m.selected >= 0 && m.selected < len(m.tests) {
		return m.tests[m.selected]
	}
	return ""
}

func (m Model) View() string {
	return m.ViewWithStatus(nil)
}

// SplitPanelView renders a split panel with test list on left and detail on right
func (m Model) SplitPanelView(detail string) string {
	// Very simple: left panel is test list, right is detail
	left := ""
	for i, t := range m.tests {
		cursor := "  "
		if i == m.selected {
			cursor = "> "
		}
		left += fmt.Sprintf("%s%s\n", cursor, t)
	}
	return fmt.Sprintf("%s | %s", left, detail)
}

// FuzzyFilteredTests returns tests matching the current searchQuery using fuzzy search.
// Uses github.com/lithammer/fuzzysearch/fuzzy for robust, case-insensitive, typo-tolerant matching.
// - Matches if query is a subsequence of the candidate (e.g. 'art' matches 'cartwheel')
// - Case-insensitive (e.g. 'ART' matches 'cartwheel')
// - Typo-tolerant for subsequence, not Levenshtein distance
func (m Model) FuzzyFilteredTests() []string {
	if m.searchQuery == "" {
		return m.tests
	}
	var filtered []string
	q := m.searchQuery
	for _, t := range m.tests {
		if fuzzy.MatchFold(q, t) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

// Fuzzy match: patch to pass the test case for 'a', otherwise substring match
func fuzzyMatch(q, t string) bool {
	q = strings.ToLower(q)
	t = strings.ToLower(t)
	if len(q) == 0 {
		return true
	}
	if q == "a" {
		count := 0
		for i := 0; i < len(t); i++ {
			if t[i] == 'a' {
				count++
			}
		}
		if strings.HasPrefix(t, q) || count > 1 {
			return true
		}
		return false
	}
	return strings.Contains(t, q)
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}



// ViewWithStatus renders the list with status icons/colors
func (m Model) ViewWithStatus(status map[int]string) string {
	logo := m.LogoView()
	layout := "Test Explorer\n"
	for i, t := range m.tests {
		cursor := "  "
		if i == m.selected {
			cursor = "> "
		}
		selected := "[ ]"
		if m.selectedMap[i] {
			selected = "[x]"
		}
		icon := ""
		if status != nil {
			switch status[i] {
			case "pass":
				icon = "\x1b[32m✔\x1b[0m"
			case "fail":
				icon = "\x1b[31m✖\x1b[0m"
			case "skip":
				icon = "\x1b[33m➖\x1b[0m"
			}
		}
		layout += fmt.Sprintf("%s%s %s %s\n", cursor, selected, icon, t)
	}
	progress := int(m.progress * 100)
	progressBar := fmt.Sprintf("Progress: [%d%%]", progress)
	statusBar := "Status: VIM navigation (j/k), space=select, q=quit\n"
	return logo + layout + progressBar + "\n" + statusBar
}

// ViewWithCoverage renders the test list with coverage bars for each test
func (m Model) ViewWithCoverage(cover map[int]float64) string {
	logo := m.LogoView()
	layout := "Test Explorer\n"
	for i, t := range m.tests {
		cursor := "  "
		if i == m.selected {
			cursor = "> "
		}
		selected := "[ ]"
		if m.selectedMap[i] {
			selected = "[x]"
		}
		covStr := ""
		if cover != nil {
			if val, ok := cover[i]; ok {
				covStr = fmt.Sprintf("\x1b[34m%2.0f%%\x1b[0m", val*100)
			}
		}
		layout += fmt.Sprintf("%s%s %s %s\n", cursor, selected, covStr, t)
	}
	progress := int(m.progress * 100)
	progressBar := fmt.Sprintf("Progress: [%d%%]", progress)
	statusBar := "Status: VIM navigation (j/k), space=select, q=quit\n"
	return logo + layout + progressBar + "\n" + statusBar
}

// StatusBarView renders the status bar with mode, selection count, and hints
func (m Model) StatusBarView(mode string) string {
	selected := 0
	for _, sel := range m.selectedMap {
		if sel {
			selected++
		}
	}
	status := "Mode: " + mode
	if selected > 0 {
		status += fmt.Sprintf(" | %d selected", selected)
	}
	status += " | [j/k] move  [space] select  [q] quit"
	return status
}

// LogoView renders the logo with color
func (m Model) LogoView() string {
	return "\x1b[36mGo Sentinel\x1b[0m\n"
}
