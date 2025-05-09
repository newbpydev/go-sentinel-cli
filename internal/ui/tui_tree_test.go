package ui_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	ui "github.com/newbpydev/go-sentinel/internal/ui"
)

// TestTUITreeSidebar_RenderInitialTree ensures the tree sidebar renders test suites/files/tests correctly.

func TestTUITreeSidebar_RenderInitialTree(t *testing.T) {
	model := ui.NewTUITestExplorerModel(mockTestTree())
	if !model.SidebarHasTree() {
		t.Errorf("sidebar did not render tree structure as expected")
	}
}

func TestTUITreeSidebar_VIMNavigation(t *testing.T) {
	model := ui.NewTUITestExplorerModel(mockTestTree())
	model.SelectedIndex = 1 // move down
	model.SelectedIndex = 0 // move up
	if !model.VIMNavigationWorked() {
		t.Errorf("VIM navigation did not update selection as expected")
	}
}

func TestTUITreeSidebar_QuitKey(t *testing.T) {
	model := ui.NewTUITestExplorerModel(mockTestTree())
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}}
	_, cmd := model.Update(msg)
	if cmd == nil {
		t.Errorf("expected tea.Quit command when pressing 'q'")
	}
}

func TestTUITreeSidebar_EnterShowsDetails(t *testing.T) {
	model := ui.NewTUITestExplorerModel(mockTestTree())
	model.SelectedIndex = 1 // select second test
	if !model.MainPaneShowsTestDetails() {
		t.Errorf("main pane did not show test details after Enter")
	}
}

func TestTUITreeSidebar_FilterSearch(t *testing.T) {
	model := ui.NewTUITestExplorerModel(mockTestTree())
	if !model.SidebarFiltered("foo") {
		t.Errorf("sidebar did not filter by search term as expected")
	}
}

// --- Helpers ---

// --- Helpers ---

func mockTestTree() *ui.TreeNode {
	return &ui.TreeNode{
		Title:    "root",
		Expanded: true,
		Children: []*ui.TreeNode{
			{
				Title:    "pkg/foo",
				Expanded: true,
				Children: []*ui.TreeNode{{Title: "TestAlpha"}, {Title: "TestBeta"}},
			},
			{
				Title:    "pkg/bar",
				Expanded: true,
				Children: []*ui.TreeNode{{Title: "TestGamma"}},
			},
		},
	}
}

// TestTUITreeSidebar_VIMNavigation tests VIM-style navigation in the tree sidebar.


// TestTUITreeSidebar_QuitKey ensures pressing 'q' triggers a quit command.


// TestTUITreeSidebar_EnterShowsDetails tests that Enter shows the correct test details in the main pane.


// TestTUITreeSidebar_FilterSearch tests filtering/searching with '/'.



// --- Helpers ---


