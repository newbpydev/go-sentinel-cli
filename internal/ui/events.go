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

// RunTestsMsg is sent when a test should be executed
type RunTestsMsg struct {
	Package string
	Test    string
}

// TestsCompletedMsg is sent when all tests finish
type TestsCompletedMsg struct {
	Success bool
}

// ToggleWatchModeMsg is sent to toggle file watching mode
type ToggleWatchModeMsg struct {}

// WatchStatusChangedMsg is sent when watch status changes
type WatchStatusChangedMsg struct {
	Enabled bool
}

// LogEntryMsg is sent when a new log entry should be added to the UI log panel
type LogEntryMsg struct {
	Content string
}

// ClearLogMsg is sent when the log panel should be cleared
type ClearLogMsg struct{}

// ShowLogViewMsg is sent to explicitly show or hide the log panel
type ShowLogViewMsg struct {
	Show bool
}

// Command to run tests
func runTestsCmd(pkg, test string) tea.Cmd {
	return func() tea.Msg {
		// CRITICAL FIX: Return RunTestsMsg instead of TestsStartedMsg
		// This ensures the request to actually run tests reaches the controller
		return RunTestsMsg{
			Package: pkg,
			Test:    test,
		}
	}
}
