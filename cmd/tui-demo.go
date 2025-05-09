package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/atotto/clipboard"
	"github.com/yourusername/go-sentinel/internal/tui"
)

type demoMsg struct{}

// DemoModel embeds tui.Model so all fields/methods are accessible
// and adds a detail string for the right panel.
// DemoModel embeds *tui.Model so all fields/methods are accessible via pointer
// and adds a detail string for the right panel.
type DemoModel struct {
	*tui.Model
	detail      string
	quit        bool
	progress    progress.Model
	logs        map[string]string
	selectedMap map[string]bool
	statusMsg   string
} // progress.Model for animated/fancy progress bar

func (m DemoModel) Init() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(time.Time) tea.Msg { return demoMsg{} })
} // periodic tick for live updates

func (m DemoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case demoMsg:
		// Simulate live updates: progress and status
		p := m.GetProgress() + 0.1
		if p > 1.0 {
			p = 1.0
		}
		m.SetProgress(p)
		cmd := m.progress.SetPercent(p)
		return m, tea.Batch(tea.Tick(time.Millisecond*500, func(time.Time) tea.Msg { return demoMsg{} }), cmd)

	// Handle progress bar animation frames
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd
	case tea.KeyMsg:
		if m.GetMode() == "search" {
			switch msg.String() {
			case "/":
				// Toggle out of search mode
				m.SetMode("normal")
				m.detail = ""
				return m, nil
			case "esc":
				// Clear search and exit
				if m.GetSearchQuery() != "" {
					m.SetSearchQuery("")
					m.SetMode("normal")
					m.detail = ""
				}
				return m, nil
			case "backspace", "delete":
				q := m.GetSearchQuery()
				if len(q) > 0 {
					q = q[:len(q)-1]
					m.SetSearchQuery(q)
					m.detail = "Search: " + m.GetSearchQuery()
					m.SetSelected(0)
				}
				return m, nil
			default:
				if msg.Type == tea.KeyRunes {
					q := m.GetSearchQuery() + msg.String()
					m.SetSearchQuery(q)
					m.detail = "Search: " + m.GetSearchQuery()
					m.SetSelected(0)
				}
				return m, nil
			}
		}
		switch msg.String() {
		case "q", "ctrl+c":
			m.quit = true
			return m, tea.Quit
		case "/":
			if m.GetMode() == "search" {
				m.SetMode("normal")
				m.detail = ""
			} else {
				m.SetMode("search")
				m.detail = "Type to filter tests. Press Esc to exit search."
			}
			return m, nil
		case "esc":
			if m.GetSearchQuery() != "" {
				m.SetSearchQuery("")
				m.SetMode("normal")
				m.detail = ""
			}
			return m, nil
		case "up", "k":
			if m.GetSelected() > 0 {
				m.SetSelected(m.GetSelected() - 1)
			}
			return m, nil
		case "down", "j":
			filtered := m.FuzzyFilteredTests()
			if m.GetSelected() < len(filtered)-1 {
				m.SetSelected(m.GetSelected() + 1)
			}
			return m, nil
		case " ", "enter":
			filtered := m.FuzzyFilteredTests()
			if len(filtered) > 0 && m.GetSelected() < len(filtered) {
				testName := filtered[m.GetSelected()]
				if m.selectedMap == nil {
					m.selectedMap = make(map[string]bool)
				}
				m.selectedMap[testName] = !m.selectedMap[testName]
				if m.selectedMap[testName] {
					m.statusMsg = "Selected: " + testName
				} else {
					m.statusMsg = "Deselected: " + testName
				}
			}
			return m, nil
		case "a":
			// Only allow select all in normal mode
			if m.GetMode() != "normal" {
				return m, nil
			}
			filtered := m.FuzzyFilteredTests()
			if len(filtered) == 0 {
				return m, nil
			}
			if m.selectedMap == nil {
				m.selectedMap = make(map[string]bool)
			}
			allSelected := true
			for _, t := range filtered {
				if !m.selectedMap[t] {
					allSelected = false
					break
				}
			}
			if allSelected {
				for _, t := range filtered {
					delete(m.selectedMap, t)
				}
				m.statusMsg = "Deselected all"
			} else {
				for _, t := range filtered {
					m.selectedMap[t] = true
				}
				m.statusMsg = "Selected all"
			}
			return m, nil
		case "c":
			// Copy selected test names to clipboard
			var selected []string
			for test := range m.selectedMap {
				if m.selectedMap[test] {
					selected = append(selected, test)
				}
			}
			if len(selected) == 0 {
				m.statusMsg = "No tests selected to copy"
				return m, nil
			}
			clipText := strings.Join(selected, "\n")
			err := clipboard.WriteAll(clipText)
			if err != nil {
				m.statusMsg = "Clipboard error: " + err.Error()
			} else {
				m.statusMsg = fmt.Sprintf("Copied %d tests!", len(selected))
			}
			return m, nil
		}
		return m, nil
	}
	return m, nil
}

// highlightQuery colors all characters of query as a fuzzy subsequence in target (case-insensitive, cyan)
func highlightQuery(target, query string) string {
	tRunes := []rune(target)
	qRunes := []rune(query)
	if len(qRunes) == 0 {
		return target
	}
	res := ""
	j := 0
	for i := 0; i < len(tRunes); i++ {
		if j < len(qRunes) && (tRunes[i] == qRunes[j] || strings.ToLower(string(tRunes[i])) == strings.ToLower(string(qRunes[j]))) {
			// Color match cyan
			res += "\x1b[36m" + string(tRunes[i]) + "\x1b[0m"
			j++
		} else {
			res += string(tRunes[i])
		}
	}
	return res
}

func (m DemoModel) View() string {
	var sb strings.Builder
	// Header with search mode indicator
	header := "[q] quit  [/] search  [j/k/up/down] move  Live updates: progress bar below"
	if m.GetMode() == "search" {
		header = "[SEARCH MODE] " + header
	}
	if m.GetSearchQuery() != "" {
		header = header[:strings.Index(header, "[")] + "[esc] clear search  " + header[strings.Index(header, "["):]
	}
	sb.WriteString(m.Model.LogoView())
	sb.WriteString("\n")
	sb.WriteString(header + "\n\n")

	// Split panel: left = filtered test list, right = details/logs for selected test
	// Render left and right panels as slices of lines
	leftLines := []string{}
	filtered := m.FuzzyFilteredTests()
	query := m.GetSearchQuery()
	for i, t := range filtered {
		cursor := "  "
		if i == m.GetSelected() {
			cursor = "> "
		}
		check := "[ ]"
		if m.selectedMap != nil && m.selectedMap[t] {
			check = "[x]"
		}
		name := t
		if m.GetMode() == "search" && query != "" {
			name = highlightQuery(t, query)
		}
		leftLines = append(leftLines, fmt.Sprintf("%s%s %s", cursor, check, name))
	}
	if len(leftLines) == 0 {
		leftLines = append(leftLines, "(no tests)")
	} 

	rightLines := []string{"No test selected."}
	if len(filtered) > 0 && m.GetSelected() < len(filtered) {
		sel := filtered[m.GetSelected()]
		if log, ok := m.logs[sel]; ok {
			rightLines = []string{log}
		} else {
			rightLines = []string{"---"}
		}
	}

	// Pad panels to same height
	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}
	for len(leftLines) < maxLines {
		leftLines = append(leftLines, "")
	}
	for len(rightLines) < maxLines {
		rightLines = append(rightLines, "")
	}

	// Always render search bar above the list (avoid layout shift)
	searchBar := fmt.Sprintf("\x1b[36mSearch [/]\x1b[0m: %s\n", m.GetSearchQuery())

	// Render side by side
	sb.WriteString(searchBar)
	for i := 0; i < maxLines; i++ {
		sb.WriteString(fmt.Sprintf("%-30s │ %s\n", leftLines[i], rightLines[i]))
	}

	// Fancy animated progress bar
	sb.WriteString(m.progress.ViewAs(m.GetProgress()))
	// Status message (if any)
	if m.statusMsg != "" {
		sb.WriteString("\n" + m.statusMsg + "\n")
	}
	return sb.String()
}

func main() {
	model := tui.NewModel()
	// Mock logs for each test
	logs := map[string]string{
		"Alpha":   "Alpha passed. No issues detected.",
		"Beta":    "Beta failed: expected 42, got 41.",
		"Gamma":   "Gamma skipped due to config.",
		"TestFoo": "TestFoo: all assertions passed.",
		"TestBar": "TestBar: warning - slow execution.",
		"TestBaz": "TestBaz: failed at step 2.\nStacktrace...",
		"Delta":   "Delta: flaky, rerun advised.",
		"Omega":   "Omega: passed.",
		"Lambda":  "Lambda: not implemented.",
	}
	m := DemoModel{
		Model:    &model,
		progress: progress.New(progress.WithDefaultGradient()),
		detail:   "",
		logs:     logs,
	}
	m.SetTests([]string{"Alpha", "Beta", "Gamma", "TestFoo", "TestBar", "TestBaz", "Delta", "Omega", "Lambda"})
	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Println("Error running TUI demo:", err)
		os.Exit(1)
	}
}
