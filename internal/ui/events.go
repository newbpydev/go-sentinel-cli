package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/go-sentinel/internal/event"
)

// TestResultsMsg is sent when new test results are available
type TestResultsMsg struct {
	Results []event.TestResult
	Tree    *TreeNode
}

// FileChangedMsg is sent when a file changes
type FileChangedMsg struct {
	Path string
}

// TestsStartedMsg is sent when tests start running
type TestsStartedMsg struct {
	Package string
}

// TestsCompletedMsg is sent when all tests finish
type TestsCompletedMsg struct {
	Success bool
}

// RunTestsMsg is sent to request test execution
type RunTestsMsg struct {
	Package string
	Test    string
}

// Command to run tests
func runTestsCmd(pkg, test string) tea.Cmd {
	return func() tea.Msg {
		return TestsStartedMsg{Package: pkg}
	}
}
