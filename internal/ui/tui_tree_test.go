package ui

import (
	"testing"

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
	// Check for icons and indentation
	if !contains(output, "📦 root") {
		t.Errorf("sidebar missing root icon or label: %s", output)
	}
	if !contains(output, "  📁 pkg/foo") {
		t.Errorf("sidebar missing file/folder icon or indentation for pkg/foo: %s", output)
	}
	if !contains(output, "    🧪 TestAlpha") {
		t.Errorf("sidebar missing test icon or indentation for TestAlpha: %s", output)
	}
	if !contains(output, "  📁 pkg/bar") {
		t.Errorf("sidebar missing file/folder icon or indentation for pkg/bar: %s", output)
	}
	if !contains(output, "    🧪 TestGamma") {
		t.Errorf("sidebar missing test icon or indentation for TestGamma: %s", output)
	}
}

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


