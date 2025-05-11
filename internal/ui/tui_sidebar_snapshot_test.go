package ui

import (
	"testing"
)




func TestSidebarVisualSnapshot(t *testing.T) {
	root := &TreeNode{
		Title:    "root",
		Expanded: true,
		Children: []*TreeNode{
			{
				Title:    "pkg/foo",
				Expanded: true,
				Children: []*TreeNode{
					{Title: "TestAlpha", Passed: boolPtr(true)},
					{Title: "TestBeta", Passed: boolPtr(false)},
				},
			},
			{
				Title:    "pkg/bar",
				Expanded: true,
				Children: []*TreeNode{
					{Title: "TestGamma", Passed: boolPtr(true)},
				},
			},
		},
	}
	model := NewTUITestExplorerModel(root)
	output := model.Sidebar.View()
	output = stripSidebarHeader(output)
	expected := `➤ ▼ root
  ▼ pkg/foo
    TestAlpha
    TestBeta
  ▶ pkg/bar`
	actualClean := trimBlankLines(output)
	expectedClean := trimBlankLines(expected)
	if actualClean != expectedClean {
		t.Errorf("Sidebar visual snapshot mismatch.\nExpected:\n%s\nGot:\n%s", expectedClean, actualClean)
	}
}

