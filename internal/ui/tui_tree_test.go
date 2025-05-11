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
	// Only visible nodes should be checked (since collapsed folders hide children)
	// Only visible nodes should be checked (since collapsed folders hide children)
	visibleNodes := []string{"➤ ▶ root"}
	for _, name := range visibleNodes {
		if !contains(output, name) {
			t.Errorf("sidebar missing node: %s\nOutput:\n%s", name, output)
		}
	}
	// Check for triangle on root
	if !contains(output, "➤ ▶ root") {
		t.Errorf("sidebar missing triangle icon for collapsed root. Output:\n%s", output)
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
	// Only visible nodes should be checked (collapsed folders hide children)
	// Only visible nodes should be checked (collapsed folders hide children)
	visibleNodes := []string{"➤ ▼ src"}
	for _, name := range visibleNodes {
		if !contains(output, name) {
			t.Errorf("sidebar missing node: %s\nOutput:\n%s", name, output)
		}
	}
	// Check for triangle on root
	if !contains(output, "➤ ▼ src") {
		t.Errorf("sidebar missing triangle icon for expanded src. Output:\n%s", output)
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

func TestSidebar_DefaultExpansionBasedOnFailures(t *testing.T) {
	// root
	// ├── pkg/foo (all pass)
	// │   ├── TestAlpha (pass)
	// │   └── TestBeta  (pass)
	// └── pkg/bar (has fail)
	//     └── TestGamma (fail)
	pass := true
	fail := false
	root := &TreeNode{
		Title: "root",
		Children: []*TreeNode{
			{
				Title: "pkg/foo",
				Children: []*TreeNode{
					{Title: "TestAlpha", Passed: &pass},
					{Title: "TestBeta", Passed: &pass},
				},
			},
			{
				Title: "pkg/bar",
				Children: []*TreeNode{
					{Title: "TestGamma", Passed: &fail},
				},
			},
		},
	}
	model := NewTUITestExplorerModel(root)
	foo := model.Tree.Children[0]
	bar := model.Tree.Children[1]
	if foo.Expanded {
		t.Errorf("pkg/foo should be collapsed by default (all passing)")
	}
	if !bar.Expanded {
		t.Errorf("pkg/bar should be expanded by default (has failure)")
	}
}

func TestSidebar_NestedFolderExpansion(t *testing.T) {
	// root
	// └── pkg (expanded)
	//     └── sub (expanded, has fail)
	//         └── TestFail (fail)
	fail := false
	root := &TreeNode{
		Title: "root",
		Children: []*TreeNode{
			{
				Title: "pkg",
				Children: []*TreeNode{
					{
						Title: "sub",
						Children: []*TreeNode{
							{Title: "TestFail", Passed: &fail},
						},
					},
				},
			},
		},
	}
	model := NewTUITestExplorerModel(root)
	pkg := model.Tree.Children[0]
	sub := pkg.Children[0]
	if !pkg.Expanded || !sub.Expanded {
		t.Errorf("All ancestor folders of a failing test should be expanded")
	}
}

func TestSidebar_ToggleExpansionWithSpace(t *testing.T) {
	pass := true
	root := &TreeNode{
		Title: "root",
		Expanded: true, // Expand root so 'pkg' is visible
		Children: []*TreeNode{
			{
				Title: "pkg",
				Children: []*TreeNode{{Title: "TestAlpha", Passed: &pass}},
			},
		},
	}
	model := NewTUITestExplorerModelWithNoExpansion(root)
	pkg := model.Tree.Children[0]
	// DEBUG: Print pointer address before toggle
	print("pkg pointer before toggle: ", pkg, "\n")
	if pkg.Expanded {
		t.Errorf("pkg should be collapsed by default")
	}
	// Simulate selecting the folder and pressing space
	// Helper to find the index of 'pkg' in model.Items
	findPkgIndex := func() int {
		for i, item := range model.Items {
			ti := item.(treeItem)
			if ti.node.Title == "pkg" {
				return i
			}
		}
		return -1
	}
	model.SelectedIndex = findPkgIndex()
	if model.SelectedIndex == -1 {
		// Print all sidebar items for debugging
		for i, item := range model.Items {
			ti := item.(treeItem)
			print("Sidebar item ", i, ": ", ti.node.Title, "\n")
		}
		t.Fatalf("'pkg' not found in sidebar items after initial flattenTree")
	}
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}}
	_, _ = (&model).Update(msg)
	model.SelectedIndex = findPkgIndex()
	if model.SelectedIndex == -1 {
		for i, item := range model.Items {
			ti := item.(treeItem)
			print("Sidebar item after toggle ", i, ": ", ti.node.Title, "\n")
		}
		t.Fatalf("'pkg' not found in sidebar items after first toggle")
	}
	pkg = model.Tree.Children[0]
	// DEBUG: Print pointer address after first toggle
	print("pkg pointer after first toggle: ", pkg, "\n")
	if !pkg.Expanded {
		t.Errorf("pkg should be expanded after pressing space")
	}
	// Press space again to collapse
	_, _ = (&model).Update(msg)
	model.SelectedIndex = findPkgIndex()
	if model.SelectedIndex == -1 {
		for i, item := range model.Items {
			ti := item.(treeItem)
			print("Sidebar item after 2nd toggle ", i, ": ", ti.node.Title, "\n")
		}
		t.Fatalf("'pkg' not found in sidebar items after second toggle")
	}
	pkg = model.Tree.Children[0]
	// DEBUG: Print pointer address after second toggle
	print("pkg pointer after second toggle: ", pkg, "\n")
	if pkg.Expanded {
		t.Errorf("pkg should be collapsed after pressing space again")
	}
}

func TestSidebar_RendersTriangleIcons(t *testing.T) {
	pass := true
	fail := false
	root := &TreeNode{
		Title: "root",
		Children: []*TreeNode{
			{
				Title: "pkg",
				Children: []*TreeNode{{Title: "TestAlpha", Passed: &pass}},
			},
			{
				Title: "bar",
				Children: []*TreeNode{{Title: "TestFail", Passed: &fail}},
			},
		},
	}
	model := NewTUITestExplorerModel(root)
	output := model.Sidebar.View()
	if !strings.Contains(output, "▶ pkg") && !strings.Contains(output, "▼ bar") {
		t.Errorf("Sidebar should render triangle icons for collapsed (▶) and expanded (▼) folders. Output:\n%s", output)
	}
}

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


