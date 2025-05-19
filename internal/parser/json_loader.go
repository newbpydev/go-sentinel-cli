// Package parser provides functionality for parsing and processing test results.
// It includes tools for loading test data from JSON files and converting them into structured formats.
package parser

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"github.com/newbpydev/go-sentinel/internal/event"
)

// TestResult represents a single Go test result entry as output by 'go test -json'.
// It contains all the fields that may be present in the JSON output of Go's test runner.
type TestResult struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test,omitempty"`
	Elapsed float64 `json:"Elapsed,omitempty"`
	Output  string  `json:"Output,omitempty"`
	Summary string  `json:"Summary,omitempty"`
	File    string  `json:"File,omitempty"`
	Line    int     `json:"Line,omitempty"`
	Message string  `json:"Message,omitempty"`
}

// LoadTestResultsFromJSON loads []TestResult from a JSON file
// It safely handles the file opening and closing, and returns parsed test results.
func LoadTestResultsFromJSON(path string) ([]TestResult, error) {
	// Validate path to prevent directory traversal attacks
	cleanPath := filepath.Clean(path)
	if !filepath.IsAbs(cleanPath) {
		return nil, fmt.Errorf("path must be absolute: %s", path)
	}
	
	f, err := os.Open(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Printf("Error closing file: %v", closeErr)
		}
	}()
	
	var results []TestResult
	dec := json.NewDecoder(f)
	if err := dec.Decode(&results); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return results, nil
}

// ConvertTestResultsToTree converts flat results to a TreeNode hierarchy
// looksLikeTestName returns true if the summary or test field looks like a test name
func looksLikeTestName(s string) bool {
	return strings.HasPrefix(s, "Test") || 
		strings.HasPrefix(s, "Example") ||
		strings.HasPrefix(s, "Benchmark")
}

// ConvertTestResultsToTree converts flat results to a TreeNode hierarchy (no root node)
func ConvertTestResultsToTree(results []TestResult) *event.TreeNode {
	modulePrefix := getModulePrefix()

	// Map of package path to directory node
	dirNodes := map[string]*event.TreeNode{}
	// Map of package path to total elapsed seconds
	packageElapsed := map[string]float64{}
	// Map of package path to whether it has tests
	packageHasTests := map[string]bool{}
	// Map of package path to whether it is skipped (no test files)
	packageIsSkipped := map[string]bool{}

	// First pass: collect test nodes and elapsed times
	for _, r := range results {
		pkgPath := r.Package
		if modulePrefix != "" && strings.HasPrefix(pkgPath, modulePrefix+"/") {
			pkgPath = strings.TrimPrefix(pkgPath, modulePrefix+"/")
		}
		if pkgPath == "" {
			continue
		}
		// Mark as skipped if Action==skip and no Test field
		if r.Action == "skip" && r.Test == "" {
			packageIsSkipped[pkgPath] = true
		}
		// Track elapsed time for package: Action==pass, no Test field, Elapsed present
		if r.Action == "pass" && r.Test == "" && r.Elapsed > 0 {
			packageElapsed[pkgPath] = r.Elapsed
		}
		// Mark as having tests if Action is pass/fail/skip and Test is present
		if (r.Action == "pass" || r.Action == "fail" || r.Action == "skip") && r.Test != "" && looksLikeTestName(r.Test) {
			packageHasTests[pkgPath] = true
		}
	}


	// Build tree, include packages with tests or skipped
	topNodes := []*event.TreeNode{}
	for pkgPath := range packageHasTests {
		segments := splitPath(pkgPath)
		var parent *event.TreeNode
		var pathSoFar string
		for i, seg := range segments {
			if i > 0 {
				pathSoFar += "/"
			}
			pathSoFar += seg
			node, ok := dirNodes[pathSoFar]
			if !ok {
				node = &event.TreeNode{Title: seg, Expanded: true, Level: i}
				dirNodes[pathSoFar] = node
				if parent != nil {
					parent.Children = append(parent.Children, node)
					node.Parent = parent
				} else {
					topNodes = append(topNodes, node)
				}
			}
			parent = node
		}
		// Attach elapsed to package node
		if elapsed, ok := packageElapsed[pkgPath]; ok {
			parent.Duration = elapsed
		}
	}
	// Add skipped packages (no test files)
	for pkgPath := range packageIsSkipped {
		if packageHasTests[pkgPath] {
			continue
		}
		segments := splitPath(pkgPath)
		var parent *event.TreeNode
		var pathSoFar string
		for i, seg := range segments {
			if i > 0 {
				pathSoFar += "/"
			}
			pathSoFar += seg
			node, ok := dirNodes[pathSoFar]
			if !ok {
				label := seg + " [skip]"
				node = &event.TreeNode{Title: label, Expanded: true, Level: i, Error: "skip"}
				dirNodes[pathSoFar] = node
				if parent != nil {
					parent.Children = append(parent.Children, node)
					node.Parent = parent
				} else {
					topNodes = append(topNodes, node)
				}
			}
			parent = node
		}
	}


	// Add test nodes under their respective package
	for _, r := range results {
		pkgPath := r.Package
		if modulePrefix != "" && strings.HasPrefix(pkgPath, modulePrefix+"/") {
			pkgPath = strings.TrimPrefix(pkgPath, modulePrefix+"/")
		}
		if pkgPath == "" || !packageHasTests[pkgPath] {
			continue
		}
		if r.Test == "" || !looksLikeTestName(r.Test) {
			continue
		}
		if r.Action != "pass" && r.Action != "fail" && r.Action != "skip" {
			continue
		}
		segments := splitPath(pkgPath)
		var pathSoFar string
		for i, seg := range segments {
			if i > 0 {
				pathSoFar += "/"
			}
			pathSoFar += seg
		}
		parent, ok := dirNodes[pathSoFar]
		if !ok {
			continue
		}
		testNode := &event.TreeNode{
			Title:    r.Test,
			Passed:   event.BoolPtr(r.Action == "pass"),
			Level:    len(segments),
			Duration: r.Elapsed,
			Parent:   parent,
		}
		parent.Children = append(parent.Children, testNode)
	}

	// --- Coverage Calculation ---
	var calcCoverage func(node *event.TreeNode) (passed, total int)
	calcCoverage = func(node *event.TreeNode) (passed, total int) {
		if len(node.Children) == 0 {
			if node.Passed != nil {
				total = 1
				if *node.Passed {
					passed = 1
				}
			}
			return
		}
		aggPassed, aggTotal := 0, 0
		for _, child := range node.Children {
			p, t := calcCoverage(child)
			aggPassed += p
			aggTotal += t
		}
		if aggTotal > 0 {
			node.Coverage = float64(aggPassed) / float64(aggTotal)
		} else {
			node.Coverage = 0.0
		}
		return aggPassed, aggTotal
	}
	for _, node := range topNodes {
		calcCoverage(node)
	}
	// Wrap all top-level nodes in a dummy parent for TUI compatibility
	return &event.TreeNode{Title: "", Children: topNodes, Expanded: true}
}

// Use event.BoolPtr now
// func boolPtr(b bool) *bool {
// 	return &b
// }

// getModulePrefix reads go.mod and returns the module name prefix, or "" if not found
func getModulePrefix() string {
	f, err := os.Open("go.mod")
	if err != nil {
		return ""
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			log.Printf("Error closing go.mod file: %v", closeErr)
		}
	}()
	// Instead of using JSON, just scan the file
	if _, err := f.Seek(0, 0); err != nil {
		log.Printf("Error seeking in go.mod file: %v", err)
		return ""
	}
	buf := make([]byte, 4096)
	if n, err := f.Read(buf); err == nil && n > 0 {
		lines := strings.Split(string(buf[:n]), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "module ") {
				return strings.TrimSpace(strings.TrimPrefix(line, "module "))
			}
		}
	}
	return ""
}


// splitPath splits a package path into segments (e.g., "a/b/c" -> ["a","b","c"])
func splitPath(pkg string) []string {
	var out []string
	for _, seg := range strings.Split(pkg, "/") {
		if seg != "" {
			out = append(out, seg)
		}
	}
	return out
}

// extractTestName tries to extract the test name from the summary (e.g., "ok   pkg/foo/TestAlpha" -> "TestAlpha")
//nolint:unused // Will be used in future implementation for test result processing
func extractTestName(summary string) string {
	if summary == "" {
		return ""
	}
	parts := strings.Split(summary, "/")
	if len(parts) > 1 {
		return parts[len(parts)-1]
	}
	return summary
}
