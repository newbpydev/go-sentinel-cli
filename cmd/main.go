package main

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/go-sentinel/internal/ui"
	"github.com/newbpydev/go-sentinel/internal/parser"
)

func main() {
	// Load real test results from JSON file
	var root *ui.TreeNode

	// Try to load real test data
	if results, err := parser.LoadTestResultsFromJSON("internal/testdata/json-output.json"); err == nil && len(results) > 0 {
		root = parser.ConvertTestResultsToTree(results)
	} else {
		log.Printf("[WARN] Could not load real test results: %v. Using stub data.", err)
		// Fallback: Minimal stub
		root = &ui.TreeNode{
			Title:    "root",
			Expanded: true,
			Children: []*ui.TreeNode{
				{Title: "pkg/foo", Expanded: true, Children: []*ui.TreeNode{{Title: "TestAlpha"}, {Title: "TestBeta"}}},
				{Title: "pkg/bar", Expanded: true, Children: []*ui.TreeNode{{Title: "TestGamma"}}},
			},
		}
	}
	model := ui.NewTUITestExplorerModel(root)
	p := tea.NewProgram(model, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		log.Fatalf("Error running TUI: %v", err)
	}
}
