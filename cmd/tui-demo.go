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
	detail   string
	quit     bool
	progress progress.Model
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
	sb.WriteString(m.Model.LogoView())
	sb.WriteString("\n")
	sb.WriteString("[q] quit  [/] search  [esc] clear search  [j/k/up/down] move  Live updates: progress bar below\n")
	sb.WriteString("\n")

	// Split panel: left = filtered test list, right = details
	left := ""
	filtered := m.FuzzyFilteredTests()
	for i, t := range filtered {
		cursor := "  "
		if i == m.GetSelected() {
			cursor = "> "
		}
		left += fmt.Sprintf("%s%s\n", cursor, t)
	}
	sb.WriteString(fmt.Sprintf("%s | %s\n", left, m.detail))

	// Fancy animated progress bar
	sb.WriteString(m.progress.ViewAs(m.GetProgress()))
	sb.WriteString("\n")
	return sb.String()
}

func main() {
	model := tui.NewModel()
	m := DemoModel{
		Model:    &model,
		progress: progress.New(progress.WithDefaultGradient()),
	}
	m.SetTests([]string{"Alpha", "Beta", "Gamma", "TestFoo", "TestBar", "TestBaz", "Delta", "Omega", "Lambda"})
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error running TUI demo:", err)
		os.Exit(1)
	}
}
