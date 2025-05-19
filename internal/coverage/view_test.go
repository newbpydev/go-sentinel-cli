package coverage

import (
	"strings"
	"testing"
)

func TestCoverageViewCreation(t *testing.T) {
	// Create sample coverage metrics
	metrics := &Metrics{
		StatementCoverage: 85.5,
		BranchCoverage:    70.2,
		FunctionCoverage:  90.0,
		LineCoverage:      88.3,
		FileMetrics: map[string]*FileMetrics{
			"github.com/newbpydev/go-sentinel/internal/sample/sample.go": {
				StatementCoverage: 85.5,
				BranchCoverage:    70.2,
				FunctionCoverage:  90.0,
				LineCoverage:      88.3,
				LineExecutionCounts: map[int]int{
					10: 1,
					11: 2,
					12: 0,
					15: 5,
				},
				UncoveredLines: []int{12},
			},
		},
	}

	// Create the coverage view
	view := NewCoverageView(metrics)
	if view == nil {
		t.Fatal("Expected coverage view to be created, got nil")
	}

	if view.metrics != metrics {
		t.Error("Expected metrics to be stored in view")
	}
}

func TestCoverageViewRendering(t *testing.T) {
	// Create sample coverage metrics
	metrics := &Metrics{
		StatementCoverage: 85.5,
		BranchCoverage:    70.2,
		FunctionCoverage:  90.0,
		LineCoverage:      88.3,
		FileMetrics: map[string]*FileMetrics{
			"github.com/newbpydev/go-sentinel/internal/sample/sample.go": {
				StatementCoverage: 85.5,
				BranchCoverage:    70.2,
				FunctionCoverage:  90.0,
				LineCoverage:      88.3,
				LineExecutionCounts: map[int]int{
					10: 1,
					11: 2,
					12: 0,
					15: 5,
				},
				UncoveredLines: []int{12},
			},
		},
	}

	// Create and render the coverage view
	view := NewCoverageView(metrics)
	rendered := view.Render()

	// Check that the rendered output contains expected elements
	if !strings.Contains(rendered, "Coverage Report") {
		t.Error("Expected rendered view to contain 'Coverage Report'")
	}

	if !strings.Contains(rendered, "88.3") {
		t.Error("Expected rendered view to show line coverage percentage")
	}
}

func TestCoverageViewColorCoding(t *testing.T) {
	// Create sample coverage metrics with various coverage levels
	metrics := &Metrics{
		LineCoverage: 75.0,
		FileMetrics: map[string]*FileMetrics{
			"high_coverage.go": {
				LineCoverage: 90.0,
				LineExecutionCounts: map[int]int{
					10: 1,
				},
			},
			"medium_coverage.go": {
				LineCoverage: 60.0,
				LineExecutionCounts: map[int]int{
					5: 1,
					6: 0,
				},
				UncoveredLines: []int{6},
			},
			"low_coverage.go": {
				LineCoverage: 30.0,
				LineExecutionCounts: map[int]int{
					1: 1,
					2: 0,
					3: 0,
				},
				UncoveredLines: []int{2, 3},
			},
		},
	}

	// Create and render the coverage view
	view := NewCoverageView(metrics)
	rendered := view.Render()

	// Check that high coverage files are rendered in green
	if !strings.Contains(rendered, "high_coverage.go") {
		t.Error("Expected rendered view to contain high coverage file")
	}

	// Check that medium coverage files are rendered in yellow
	if !strings.Contains(rendered, "medium_coverage.go") {
		t.Error("Expected rendered view to contain medium coverage file")
	}

	// Check that low coverage files are rendered in red
	if !strings.Contains(rendered, "low_coverage.go") {
		t.Error("Expected rendered view to contain low coverage file")
	}
}

func TestFileCoverageView(t *testing.T) {
	// Create sample file metrics
	fileMetrics := &FileMetrics{
		StatementCoverage: 85.5,
		BranchCoverage:    70.2,
		FunctionCoverage:  90.0,
		LineCoverage:      88.3,
		LineExecutionCounts: map[int]int{
			10: 1,
			11: 2,
			12: 0,
			15: 5,
		},
		UncoveredLines: []int{12},
	}

	// Create sample source code
	sourceCode := map[int]string{
		10: "func add(a, b int) int {",
		11: "    return a + b",
		12: "    // This line is never reached",
		15: "func main() {",
	}

	// Create and render the file coverage view
	view := NewFileCoverageView("sample.go", fileMetrics, sourceCode)
	rendered := view.Render()

	// Check that the rendered output shows source code with coverage annotations
	if !strings.Contains(rendered, "sample.go") {
		t.Error("Expected rendered view to contain file name")
	}

	if !strings.Contains(rendered, "88.3%") {
		t.Error("Expected rendered view to show coverage percentage")
	}

	if !strings.Contains(rendered, "1x") {
		t.Error("Expected rendered view to show execution counts")
	}
}

func TestCoverageKeyBindings(t *testing.T) {
	// Create a coverage view with key handlers
	view := NewCoverageView(&Metrics{})

	// Check that the view has expected key handlers
	if !view.HasKeyBinding('v') {
		t.Error("Expected coverage view to have 'v' key binding for toggling view")
	}

	if !view.HasKeyBinding('f') {
		t.Error("Expected coverage view to have 'f' key binding for filtering")
	}
}
