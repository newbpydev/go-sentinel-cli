package parser

import (
	"encoding/json"
	"os"
	"fmt"
	"github.com/newbpydev/go-sentinel/internal/ui"
)

type TestResult struct {
	Package string  `json:"Package"`
	Passed  bool    `json:"Passed"`
	Summary string  `json:"Summary"`
	File    string  `json:"File"`
	Line    int     `json:"Line"`
	Message string  `json:"Message"`
}

// LoadTestResultsFromJSON loads []TestResult from a JSON file
func LoadTestResultsFromJSON(path string) ([]TestResult, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer f.Close()
	var results []TestResult
	dec := json.NewDecoder(f)
	if err := dec.Decode(&results); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return results, nil
}

// ConvertTestResultsToTree converts flat results to a TreeNode hierarchy
func ConvertTestResultsToTree(results []TestResult) *ui.TreeNode {
	root := &ui.TreeNode{Title: "root", Expanded: true}
	pkgs := map[string]*ui.TreeNode{}
	for _, r := range results {
		pkg, ok := pkgs[r.Package]
		if !ok {
			pkg = &ui.TreeNode{Title: r.Package, Expanded: true}
			pkgs[r.Package] = pkg
			root.Children = append(root.Children, pkg)
		}
		testNode := &ui.TreeNode{Title: r.Summary, Passed: &r.Passed, Error: r.Message}
		pkg.Children = append(pkg.Children, testNode)
	}
	return root
}
