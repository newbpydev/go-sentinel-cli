package coverage

// Metrics represents coverage metrics for a file or the entire project.
type Metrics struct {
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

	LineExecutionCounts map[int]int  // Maps line number to execution count
	UncoveredLines      []int        // Line numbers with no coverage
	PartialBranches     []BranchInfo // Information about branches with partial coverage
}

// BranchInfo holds information about a branch in the code
type BranchInfo struct {
	Line       int
	Condition  string
	Covered    bool
	Executions int
}
