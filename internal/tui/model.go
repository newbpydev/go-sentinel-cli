package tui

import (
	fmt "fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	tests       []string
	selected    int
	selectedMap map[int]bool
	progress    float64
	mode        string
	ggPending   bool // for detecting double 'g'
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
