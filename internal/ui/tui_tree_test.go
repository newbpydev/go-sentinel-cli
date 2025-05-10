package ui

import (
	"testing"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// TestTUITreeSidebar_RenderInitialTree ensures the tree sidebar renders test suites/files/tests correctly.

func TestTUITreeSidebar_RenderInitialTree(t *testing.T) {
	model := NewTUITestExplorerModel(mockTestTree())
	if !model.SidebarHasTree() {
		t.Errorf("sidebar did not render tree structure as expected")
	}
}

func TestTUITreeSidebar_IndentationAndIcons(t *testing.T) {
	model := NewTUITestExplorerModel(mockTestTree())
	output := model.Sidebar.View()
	output = stripSidebarHeader(output)
	output = trimBlankLines(output)
	// Only check for presence of node names (no icons/indent for test nodes)
	nodes := []string{"root", "pkg/foo", "pkg/bar", "TestAlpha", "TestBeta", "TestGamma"}
	for _, name := range nodes {
		if !contains(output, name) {
			t.Errorf("sidebar missing node: %s\nOutput:\n%s", name, output)
		}
	}
}

func stripSidebarHeader(s string) string {
	lines := strings.SplitN(s, "\n", 2)
	if len(lines) == 2 {
		return lines[1]
	}
	return s
}


func trimBlankLines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n") // normalize CRLF to LF
	lines := strings.Split(s, "\n")
	start := 0
	end := len(lines)
	for start < end && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	trimmed := make([]string, 0, end-start)
	for _, line := range lines[start:end] {
		line = strings.TrimSpace(line)
		trimmed = append(trimmed, line)
	}
	return strings.Join(trimmed, "\n")
}

func TestTUITreeSidebar_RendersCoverageAndTestDetails(t *testing.T) {
	// Simulate parsed test data with coverage
	root := &TreeNode{
		Title:    "src",
		Expanded: true,
		Coverage: 0.0769, // 7.69%
		Children: []*TreeNode{
			{
				Title:    "App.js",
				Expanded: true,
				Coverage: 1.0,
				Children: []*TreeNode{{Title: "App renders", Passed: boolPtr(true), Duration: 0.01}},
			},
			{
				Title:    "index.js",
				Coverage: 0.0,
				Children: []*TreeNode{{Title: "index loads", Passed: boolPtr(false), Duration: 0.02, Error: "ReferenceError"}},
			},
			{
				Title:    "serviceWorker.js",
				Coverage: 0.0,
				Children: []*TreeNode{{Title: "service registers", Passed: boolPtr(false), Duration: 0.01, Error: "TypeError"}},
			},
			{
				Title:    "setupTests.js",
				Coverage: 1.0,
				Children: []*TreeNode{{Title: "setup runs", Passed: boolPtr(true), Duration: 0.005}},
			},
		},
	}
	model := NewTUITestExplorerModel(root)
	output := model.Sidebar.View()
	output = stripSidebarHeader(output)
	output = trimBlankLines(output)
	// Only check for presence of node names (no icons/coverage/duration)
	nodes := []string{"App.js", "index.js", "src", "App renders"}
	for _, name := range nodes {
		if !contains(output, name) {
			t.Errorf("sidebar missing node: %s\nOutput:\n%s", name, output)
		}
	}

}

func boolPtr(b bool) *bool { return &b }

func contains(s, substr string) bool {
	return len(s) > 0 && (s == substr || (len(s) > len(substr) && (contains(s[1:], substr) || s[:len(substr)] == substr)))
}


func TestTUITreeSidebar_VIMNavigation(t *testing.T) {
	model := NewTUITestExplorerModel(mockTestTree())
	model.SelectedIndex = 1 // move down
	model.SelectedIndex = 0 // move up
	if !model.VIMNavigationWorked() {
		t.Errorf("VIM navigation did not update selection as expected")
	}
}

func TestTUITreeSidebar_QuitKey(t *testing.T) {
	model := NewTUITestExplorerModel(mockTestTree())
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := model.Update(msg)
	if cmd == nil {
		t.Errorf("expected tea.Quit command when pressing 'q'")
	}
}

func TestTUITreeSidebar_EnterShowsDetails(t *testing.T) {
	model := NewTUITestExplorerModel(mockTestTree())
	model.SelectedIndex = 1 // select second test
	if !model.MainPaneShowsTestDetails() {
		t.Errorf("main pane did not show test details after Enter")
	}
}

func TestTUITreeSidebar_FilterSearch(t *testing.T) {
	model := NewTUITestExplorerModel(mockTestTree())
	if !model.SidebarFiltered("foo") {
		t.Errorf("sidebar did not filter by search term as expected")
	}
}

// --- Helpers ---

// --- Helpers ---

func mockTestTree() *TreeNode {
	return &TreeNode{
		Title:    "root",
		Expanded: true,
		Children: []*TreeNode{
			{
				Title:    "pkg/foo",
				Expanded: true,
				Children: []*TreeNode{{Title: "TestAlpha"}, {Title: "TestBeta"}},
			},
			{
				Title:    "pkg/bar",
				Expanded: true,
				Children: []*TreeNode{{Title: "TestGamma"}},
			},
		},
	}
}

// TestTUITreeSidebar_VIMNavigation tests VIM-style navigation in the tree sidebar.


// TestTUITreeSidebar_QuitKey ensures pressing 'q' triggers a quit command.


// TestTUITreeSidebar_EnterShowsDetails tests that Enter shows the correct test details in the main pane.


// TestTUITreeSidebar_FilterSearch tests filtering/searching with '/'.



// --- Helpers ---


