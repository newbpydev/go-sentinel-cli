package ui_test

import (
	"strings"
	"testing"

	"github.com/newbpydev/go-sentinel/internal/event"
	"github.com/newbpydev/go-sentinel/internal/ui"
)




func TestSidebarVisualSnapshot(t *testing.T) {
	root := &ui.TreeNode{
		Title:    "root",
		Expanded: true,
		Children: []*ui.TreeNode{
			{
				Title:    "pkg/foo",
				Expanded: true,
				Children: []*ui.TreeNode{
					{Title: "TestAlpha", Passed: event.BoolPtr(true)},
					{Title: "TestBeta", Passed: event.BoolPtr(false)},
				},
			},
			{
				Title:    "pkg/bar",
				Expanded: true,
				Children: []*ui.TreeNode{
					{Title: "TestGamma", Passed: event.BoolPtr(true)},
				},
			},
		},
	}
	model := ui.NewTUITestExplorerModel(root)
	output := model.Sidebar.View()
	// Strip search bar from sidebar header
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		output = strings.Join(lines[1:], "\n")
	}
	expected := `➤ ▼ root
  ▼ pkg/foo
    TestAlpha
    TestBeta
  ▶ pkg/bar`
	actualClean := strings.TrimSpace(output)
	expectedClean := strings.TrimSpace(expected)
	if actualClean != expectedClean {
		t.Errorf("Sidebar visual snapshot mismatch.\nExpected:\n%s\nGot:\n%s", expectedClean, actualClean)
	}
}

