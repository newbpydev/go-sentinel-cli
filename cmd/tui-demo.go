package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/progress"
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
		switch msg.String() {
		case "q", "ctrl+c":
			m.quit = true
			return m, tea.Quit
		case "/":
			m.SetMode("search")
			m.detail = "Type to filter tests. Press Esc to exit search."
			return m, nil
		case "esc":
			m.SetMode("normal")
			m.SetSearchQuery("")
			m.detail = ""
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
			filtered := m.FuzzyFilteredTests()
			if len(filtered) == 0 {
				return m, nil
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
		}
		if m.GetMode() == "search" {
			q := m.GetSearchQuery()
			switch msg.Type {
			case tea.KeyBackspace, tea.KeyDelete:
				if len(q) > 0 {
					q = q[:len(q)-1]
				}
			case tea.KeyRunes:
				q += msg.String()
			}
			m.SetSearchQuery(q)
			m.detail = "Search: " + m.GetSearchQuery()
			// Reset selection to top of filtered list
			m.SetSelected(0)
			return m, nil
		}
	}
	return m, nil
}

func (m DemoModel) View() string {
	var sb strings.Builder
	// Header with search mode indicator
	header := "[q] quit  [/] search  [esc] clear search  [j/k/up/down] move  Live updates: progress bar below"
	if m.GetMode() == "search" {
		header = "[SEARCH MODE] " + header
	}
	sb.WriteString(m.Model.LogoView())
	sb.WriteString("\n")
	sb.WriteString(header + "\n\n")

	// Split panel: left = filtered test list, right = details/logs for selected test
	// Render left and right panels as slices of lines
	leftLines := []string{}
	filtered := m.FuzzyFilteredTests()
	for i, t := range filtered {
		cursor := "  "
		if i == m.GetSelected() {
			cursor = "> "
		}
		check := "[ ]"
		if m.selectedMap != nil && m.selectedMap[t] {
			check = "[x]"
		}
		leftLines = append(leftLines, fmt.Sprintf("%s%s %s", cursor, check, t))
	}
	if len(leftLines) == 0 {
		leftLines = append(leftLines, "(no tests)")
	} 

	rightLines := []string{"No test selected."}
	if len(filtered) > 0 && m.GetSelected() < len(filtered) {
		testName := filtered[m.GetSelected()]
		log, ok := m.logs[testName]
		if !ok {
			rightLines = []string{fmt.Sprintf("No log for %s", testName)}
		} else {
			rightLines = append([]string{fmt.Sprintf("Details for %s", testName), "---"}, strings.Split(log, "\n")...)
		}
	}

	// Panel width config
	leftWidth := 18
	rightWidth := 40
	panelGap := "  "
	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}

	// Render both panels side by side
	for i := 0; i < maxLines; i++ {
		var left, right string
		if i < len(leftLines) {
			left = leftLines[i]
		} else {
			left = ""
		}
		if i < len(rightLines) {
			right = rightLines[i]
		} else {
			right = ""
		}
		// Pad left panel to leftWidth
		if len(left) < leftWidth {
			left = left + strings.Repeat(" ", leftWidth-len(left))
		} else if len(left) > leftWidth {
			left = left[:leftWidth]
		}
		// Pad right panel to rightWidth (optional)
		if len(right) < rightWidth {
			right = right + strings.Repeat(" ", rightWidth-len(right))
		} else if len(right) > rightWidth {
			right = right[:rightWidth]
		}
		sb.WriteString(left + panelGap + right + "\n")
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
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error running TUI demo:", err)
		os.Exit(1)
	}
}
