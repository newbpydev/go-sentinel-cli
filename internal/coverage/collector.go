// Package coverage provides functionality for collecting and analyzing test coverage data.
// It includes tools for parsing coverage profiles, generating reports, and visualizing coverage information.
package coverage

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/cover"
)

// CoverageCollector handles the collection and analysis of coverage data
type CoverageCollector struct {
	CoverageFile string
	Profiles     []*cover.Profile
}

// NewCollector creates a new coverage collector from a coverage profile file
func NewCollector(coverageFile string) (*CoverageCollector, error) {
	// Check if the file exists
	if _, err := os.Stat(coverageFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("coverage file does not exist: %s", coverageFile)
	}

	// Parse the coverage profiles
	profiles, err := cover.ParseProfiles(coverageFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse coverage profiles: %w", err)
	}

	return &CoverageCollector{
		CoverageFile: coverageFile,
		Profiles:     profiles,
	}, nil
}

// CalculateMetrics processes coverage data and returns metrics
func (c *CoverageCollector) CalculateMetrics() (*CoverageMetrics, error) {
	if len(c.Profiles) == 0 {
		return nil, fmt.Errorf("no coverage profiles available")
	}

	metrics := &CoverageMetrics{
		FileMetrics: make(map[string]*FileMetrics),
	}

	var totalStatements, coveredStatements int64
	var totalBranches, coveredBranches int64
	var totalFunctions, coveredFunctions int64
	var totalLines, coveredLines int64

	// Process each file in the coverage profile
	for _, profile := range c.Profiles {
		fileMetrics := &FileMetrics{
			LineExecutionCounts: make(map[int]int),
			UncoveredLines:      []int{},
			PartialBranches:     []BranchInfo{},
		}

		var fileStatements, fileCoveredStatements int64
		var fileBranches, fileCoveredBranches int64
		var fileFunctions, fileCoveredFunctions int64
		var fileLines, fileCoveredLines int64

		// Process each block in the file
		for _, block := range profile.Blocks {
			// Count statements
			fileStatements += int64(block.NumStmt)
			totalStatements += int64(block.NumStmt)

			if block.Count > 0 {
				fileCoveredStatements += int64(block.NumStmt)
				coveredStatements += int64(block.NumStmt)
			}

			// Track line execution
			// Each block counts as a line for coverage calculation
			fileLines++
			totalLines++

			// Record execution count for all lines in the block
			for line := block.StartLine; line <= block.EndLine; line++ {
				fileMetrics.LineExecutionCounts[line] = block.Count
			}

			// Track covered/uncovered blocks (treating each block as a logical line)
			if block.Count > 0 {
				fileCoveredLines++
				coveredLines++
			} else {
				// Add all lines in the uncovered block to uncovered lines list
				for line := block.StartLine; line <= block.EndLine; line++ {
					fileMetrics.UncoveredLines = append(fileMetrics.UncoveredLines, line)
				}
			}

			// Simple branch detection based on block structure
			// This is a simplified approach; real branch detection would use AST analysis
			if block.NumStmt > 1 {
				fileBranches++
				totalBranches++

				if block.Count > 0 {
					fileCoveredBranches++
					coveredBranches++
				} else {
					// Add partial branch info
					fileMetrics.PartialBranches = append(fileMetrics.PartialBranches, BranchInfo{
						Line:       block.StartLine,
						Condition:  "branch",
						Covered:    block.Count > 0,
						Executions: block.Count,
					})
				}
			}
		}

		// Calculate file-level metrics
		if fileStatements > 0 {
			fileMetrics.StatementCoverage = float64(fileCoveredStatements) / float64(fileStatements) * 100
		}
		if fileBranches > 0 {
			fileMetrics.BranchCoverage = float64(fileCoveredBranches) / float64(fileBranches) * 100
		}
		if fileFunctions > 0 {
			fileMetrics.FunctionCoverage = float64(fileCoveredFunctions) / float64(fileFunctions) * 100
		}
		if fileLines > 0 {
			fileMetrics.LineCoverage = float64(fileCoveredLines) / float64(fileLines) * 100
		}

		// Add file metrics to overall metrics
		metrics.FileMetrics[profile.FileName] = fileMetrics
	}

	// Calculate overall metrics
	if totalStatements > 0 {
		metrics.StatementCoverage = float64(coveredStatements) / float64(totalStatements) * 100
	}
	if totalBranches > 0 {
		metrics.BranchCoverage = float64(coveredBranches) / float64(totalBranches) * 100
	}
	if totalFunctions > 0 {
		metrics.FunctionCoverage = float64(coveredFunctions) / float64(totalFunctions) * 100
	}
	if totalLines > 0 {
		metrics.LineCoverage = float64(coveredLines) / float64(totalLines) * 100
	}

	return metrics, nil
}

// validatePath checks if the path is safe to read
func validatePath(path string) error {
	// Add path validation logic here
	if path == "" {
		return fmt.Errorf("empty path")
	}
	// Add more validation as needed
	return nil
}

// ReadFile reads a file with validation
func readFileWithValidation(absPath string) ([]byte, error) {
	if err := validatePath(absPath); err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return content, nil
}

// GetSourceCode retrieves the source code for a given file with coverage annotations
func (c *CoverageCollector) GetSourceCode(filePath string) (map[int]string, error) {
	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Read file content
	content, err := readFileWithValidation(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Split into lines
	lines := strings.Split(string(content), "\n")
	result := make(map[int]string)

	// Add each line to the result with line number as key
	for i, line := range lines {
		result[i+1] = line // Line numbers are 1-based
	}

	return result, nil
}

// AnalyzeBranches performs deeper branch analysis using AST
func (c *CoverageCollector) AnalyzeBranches(filePath string) ([]BranchInfo, error) {
	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Parse the file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, absPath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	// Extract branch information
	var branches []BranchInfo

	// Visit all nodes in the AST
	ast.Inspect(node, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.IfStmt:
			pos := fset.Position(stmt.If)
			branches = append(branches, BranchInfo{
				Line:      pos.Line,
				Condition: "if",
				// Coverage will be added later
			})
		case *ast.SwitchStmt:
			pos := fset.Position(stmt.Switch)
			branches = append(branches, BranchInfo{
				Line:      pos.Line,
				Condition: "switch",
				// Coverage will be added later
			})
		}
		return true
	})

	return branches, nil
}
