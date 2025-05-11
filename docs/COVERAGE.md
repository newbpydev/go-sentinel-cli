Coverage Reporting Implementation Plan for Go-Sentinel
I've studied the Jest coverage reporting system and can now outline a detailed plan to implement similar functionality in Go-Sentinel. This will enhance the existing testing capabilities by providing clear coverage metrics.

Understanding Required Coverage Metrics
Based on the Jest coverage approach, we'll implement four key coverage metrics:

Statement Coverage: Percentage of executable statements executed during tests
Branch Coverage: Percentage of code branches (if/else/switch) that were tested
Function Coverage: Percentage of functions executed during testing
Line Coverage: Percentage of code lines executed during tests
Implementation Plan
Phase 1: Leverage Go's Native Coverage Tools
go
CopyInsert
// Phase 1: Using Go's built-in coverage tools
Research and Integration:
Go already provides built-in coverage analysis with go test -cover
We'll leverage the -coverprofile flag to generate coverage data files
Integrate with the existing Go-Sentinel test execution flow
Coverage Data Collection:
Execute tests with: go test -coverprofile=coverage.out ./...
Parse the generated coverage profile using Go's cover package
Store parsed coverage data for visualization
Phase 2: Coverage Data Analysis
go
CopyInsert
// Phase 2: Process and analyze coverage data
Coverage Metrics Calculation:
Calculate statement coverage from Go's native coverage data
For branch coverage, we'll need to analyze AST (Abstract Syntax Tree)
Function coverage can be determined by tracking function entries
Line coverage is directly available from Go's coverage profile
Create Coverage Processor:
Build a CoverageProcessor struct to parse Go coverage output
Extract file-specific coverage information
Implement metrics calculations for each coverage type
Phase 3: UI Integration
go
CopyInsert
// Phase 3: Integrate with Go-Sentinel's TUI
Coverage Visualization:
Add a new coverage view to Go-Sentinel's TUI
Implement color-coding similar to Jest (yellow for partial, red for uncovered)
Display execution counts (1x, 2x) next to code lines
Interactive Coverage Explorer:
Allow users to navigate through files and view coverage details
Implement filtering for uncovered or partially covered code
Enable drilling down into specific functions or branches
Phase 4: Commands and User Experience
go
CopyInsert
// Phase 4: Add user commands and improve UX
User Commands:
Add a [v] key to toggle coverage visualization
Implement [f] to focus on files with low coverage
Create [b] command to highlight uncovered branches
Coverage Reports:
Generate HTML reports similar to Jest's lcov output
Implement coverage summary in terminal view
Add options to export coverage data
Detailed Technical Implementation
Coverage Data Collection
go
CopyInsert
// coverage/collector.go
package coverage

import (
    "go/ast"
    "go/parser"
    "go/token"
    "golang.org/x/tools/cover"
    "io/ioutil"
)

// CoverageCollector handles the collection of coverage data
type CoverageCollector struct {
    CoverageFile string
    Profiles     []*cover.Profile
}

// NewCollector creates a new coverage collector
func NewCollector(coverageFile string) (*CoverageCollector, error) {
    profiles, err := cover.ParseProfiles(coverageFile)
    if err != nil {
        return nil, err
    }
    
    return &CoverageCollector{
        CoverageFile: coverageFile,
        Profiles:     profiles,
    }, nil
}

// CalculateMetrics processes coverage data and returns metrics
func (c *CoverageCollector) CalculateMetrics() (*CoverageMetrics, error) {
    // Implementation for calculating coverage metrics
    // ...
}
Coverage Metrics
go
CopyInsert
// coverage/metrics.go
package coverage

// CoverageMetrics holds the coverage information
type CoverageMetrics struct {
    StatementCoverage float64
    BranchCoverage    float64
    FunctionCoverage  float64
    LineCoverage      float64
    
    FileMetrics map[string]*FileMetrics
}

// FileMetrics holds per-file coverage information
type FileMetrics struct {
    StatementCoverage float64
    BranchCoverage    float64
    FunctionCoverage  float64
    LineCoverage      float64
    
    LineExecutionCounts map[int]int
    UncoveredLines      []int
    PartialBranches     []BranchInfo
}

// BranchInfo holds information about a branch
type BranchInfo struct {
    Line       int
    Condition  string
    Covered    bool
    Executions int
}
UI Integration
go
CopyInsert
// ui/coverage_view.go
package ui

import (
    "github.com/newbpydev/go-sentinel/coverage"
    "github.com/charmbracelet/lipgloss"
)

// CoverageView represents the TUI component for displaying coverage
type CoverageView struct {
    metrics *coverage.CoverageMetrics
    // Other fields for UI state
}

// NewCoverageView creates a new coverage view
func NewCoverageView(metrics *coverage.CoverageMetrics) *CoverageView {
    return &CoverageView{
        metrics: metrics,
    }
}

// Render renders the coverage view
func (cv *CoverageView) Render() string {
    // Implementation for rendering coverage information
    // ...
}
Command Line Integration
go
CopyInsert
// cmd/coverage.go
package cmd

import (
    "github.com/newbpydev/go-sentinel/coverage"
    "github.com/spf13/cobra"
)

// AddCoverageCommands adds coverage-related commands to the root command
func AddCoverageCommands(rootCmd *cobra.Command) {
    coverageCmd := &cobra.Command{
        Use:   "coverage",
        Short: "Manage and view test coverage",
        Run:   runCoverage,
    }
    
    // Add subcommands
    // ...
    
    rootCmd.AddCommand(coverageCmd)
}

// runCoverage runs coverage analysis
func runCoverage(cmd *cobra.Command, args []string) {
    // Implementation for coverage command
    // ...
}
Integration with Existing Test Selection Mode
We'll enhance the current interactive test selection mode by adding coverage information:

Add coverage indicators next to failing tests
Color code tests based on their coverage metrics
Allow selecting tests with low coverage using the [l] key
Implement coverage trend tracking to show improvement/regression
Timeline and Milestones
Research & Design (1-2 days)
Study Go's coverage tools in depth
Design data structures for coverage metrics
Plan UI integration
Core Implementation (3-5 days)
Implement coverage data collection
Build metrics calculation
Create basic coverage visualization
UI Enhancement (2-3 days)
Add coverage view to TUI
Implement color coding and indicators
Add interactive navigation
Testing & Documentation (2-3 days)
Test the coverage functionality
Document usage and implementation
Write examples
Next Steps
To begin implementation, we should:

Set up a development branch for the coverage feature
Add the necessary dependencies for AST parsing and coverage analysis
Create the initial structure for the coverage package
Implement basic coverage data collection as a proof of