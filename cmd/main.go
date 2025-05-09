package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/go-sentinel/internal/ui"
)

func main() {
	// Minimal mock tree for demonstration
	root := &ui.TreeNode{
		Title:    "root",
		Expanded: true,
		Children: []*ui.TreeNode{
			{Title: "pkg/foo", Expanded: true, Children: []*ui.TreeNode{{Title: "TestAlpha"}, {Title: "TestBeta"}}},
			{Title: "pkg/bar", Expanded: true, Children: []*ui.TreeNode{{Title: "TestGamma"}}},
		},
	}
	model := ui.NewTUITestExplorerModel(root)
	p := tea.NewProgram(model)
	if err := p.Start(); err != nil {
		log.Fatalf("Error running TUI: %v", err)
	}
}
