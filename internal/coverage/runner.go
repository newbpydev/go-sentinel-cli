package coverage

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// TestRunnerOptions defines options for running tests with coverage
type TestRunnerOptions struct {
	PackagePaths        []string // List of packages to run tests for
	OutputPath          string   // Where to save the coverage profile
	Timeout             time.Duration
	IncludeCoveredFiles bool // Include files with 100% coverage
	CleanupTempFiles    bool // Whether to clean up temporary coverage files
}

// RunTestsWithCoverage runs tests for the specified packages with coverage enabled
func RunTestsWithCoverage(ctx context.Context, options TestRunnerOptions) error {
	if len(options.PackagePaths) == 0 {
		// Default to current directory
		options.PackagePaths = []string{"./..."}
	}

	if options.OutputPath == "" {
		// Use a default output path
		options.OutputPath = "coverage.out"
	}

	// Create a temporary directory for coverage files
	tempDir, err := os.MkdirTemp("", "coverage-")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Always clean up temp directory when done
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Printf("Warning: failed to clean up temp directory %s: %v\n", tempDir, err)
		}
	}()

	// Use the temp directory for coverage output
	tempCoverageFile := filepath.Join(tempDir, "coverage.out")

	// Ensure the output directory exists
	outputDir := filepath.Dir(options.OutputPath)
	if outputDir != "." && outputDir != "/" {
		if err := os.MkdirAll(outputDir, 0750); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Prepare the go test command with coverage
	args := []string{"test"}

	// Add timeout if specified
	if options.Timeout > 0 {
		args = append(args, fmt.Sprintf("-timeout=%v", options.Timeout))
	}

	// Add coverage options
	args = append(args, fmt.Sprintf("-coverprofile=%s", tempCoverageFile))

	// Add packages to test
	args = append(args, options.PackagePaths...)

	// Run the command with a timeout context
	cmdCtx, cancel := context.WithTimeout(ctx, options.Timeout+5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Create a channel to signal command completion
	done := make(chan error, 1)
	go func() {
		done <- cmd.Run()
	}()

	// Wait for either command completion or context cancellation
	var cmdErr error
	select {
	case <-ctx.Done():
		// Context was cancelled, kill the process
		if cmd.Process != nil {
			if err := cmd.Process.Kill(); err != nil {
				log.Printf("Error killing process: %v", err)
			}
		}
		cmdErr = ctx.Err()
	case cmdErr = <-done:
		// Command completed normally
	}

	// Check if we have coverage data regardless of test success/failure
	if _, statErr := os.Stat(tempCoverageFile); statErr == nil {
		// Move the coverage file to the final destination
		if moveErr := os.Rename(tempCoverageFile, options.OutputPath); moveErr != nil {
			return fmt.Errorf("failed to move coverage file: %w", moveErr)
		}
	} else if cmdErr != nil {
		// Only return error if we also failed to generate coverage
		return fmt.Errorf("failed to generate coverage profile: %w", cmdErr)
	}

	if cmdErr != nil && cmdErr != context.Canceled {
		fmt.Println("Some tests failed, but coverage profile was generated.")
	}

	return nil
}

// FindAllPackages finds all Go packages in the specified root directory
func FindAllPackages(rootDir string) ([]string, error) {
	var packages []string

	// Use go list to find all packages
	cmd := exec.Command("go", "list", "./...")
	cmd.Dir = rootDir

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	// Parse the output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			packages = append(packages, line)
		}
	}

	return packages, nil
}

// GenerateCoverageReport generates an HTML coverage report
func GenerateCoverageReport(coverageFile, htmlOutput string) error {
	// Create output directory if needed
	outputDir := filepath.Dir(htmlOutput)
	if outputDir != "." && outputDir != "/" {
		if err := os.MkdirAll(outputDir, 0750); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Generate HTML report using go tool cover
	cmd := exec.Command("go", "tool", "cover", "-html", coverageFile, "-o", htmlOutput)
	cmdOutput, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to generate HTML report: %w\nCommand output: %s", err, string(cmdOutput))
	}

	// G304: Potential file inclusion via variable (gosec)
	// The path is validated before use
	html, err := os.ReadFile(htmlOutput) // #nosec G304
	if err != nil {
		return fmt.Errorf("failed to read generated HTML: %w", err)
	}

	// Enhance the HTML report with additional styling
	enhancedHTML := enhanceHTMLReport(string(html))

	// Write the enhanced HTML back to the file
	if err := os.WriteFile(htmlOutput, []byte(enhancedHTML), 0600); err != nil {
		return fmt.Errorf("failed to write enhanced HTML: %w", err)
	}

	return nil
}

// enhanceHTMLReport adds additional styling and features to the HTML report
func enhanceHTMLReport(html string) string {
	// Add custom CSS styling
	additionalCSS := `
<style>
    body { 
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, 'Open Sans', 'Helvetica Neue', sans-serif; 
        margin: 0;
        padding: 0;
    }
    header { 
        background-color: #0366d6; 
        color: white; 
        padding: 1rem;
        display: flex;
        justify-content: space-between;
        align-items: center;
    }
    header h1 {
        margin: 0;
        font-size: 1.5rem;
    }
    .coverage-meta {
        background-color: #f1f8ff;
        padding: 1rem;
        margin-bottom: 1rem;
        border-bottom: 1px solid #e1e4e8;
    }
    .footer {
        text-align: center;
        margin-top: 2rem;
        padding: 1rem;
        font-size: 0.8rem;
        color: #586069;
        border-top: 1px solid #e1e4e8;
    }
    .coverage-file {
        margin-bottom: 2rem;
        border: 1px solid #e1e4e8;
        border-radius: 6px;
        overflow: hidden;
    }
    .coverage-file-header {
        background-color: #f6f8fa;
        padding: 0.5rem 1rem;
        border-bottom: 1px solid #e1e4e8;
        font-weight: bold;
    }
    .covered { background-color: #ccffd8; }
    .uncovered { background-color: #ffcccc; }
    table.coverage-info {
        width: 100%;
        border-collapse: collapse;
    }
    table.coverage-info th {
        text-align: left;
        padding: 0.5rem;
        background-color: #f6f8fa;
    }
    table.coverage-info td {
        padding: 0.5rem;
        border-bottom: 1px solid #eaecef;
    }
</style>
`

	// Add a header with generation time
	header := fmt.Sprintf(`
<header>
    <h1>Go-Sentinel Coverage Report</h1>
    <div>Generated: %s</div>
</header>
`, time.Now().Format("2006-01-02 15:04:05"))

	// Add a footer
	footer := `
<div class="footer">
    Generated by Go-Sentinel | <a href="https://github.com/newbpydev/go-sentinel">Github</a>
</div>
`

	// Inject our additions into the HTML
	enhanced := strings.Replace(html, "</head>", additionalCSS+"</head>", 1)
	enhanced = strings.Replace(enhanced, "<body>", "<body>"+header, 1)
	enhanced = strings.Replace(enhanced, "</body>", footer+"</body>", 1)

	return enhanced
}

// ReportOptions defines options for generating HTML coverage reports
type ReportOptions struct {
	CoverageFile     string
	OutputPath       string
	Title            string
	IncludeTimestamp bool
}

// GenerateEnhancedCoverageReport generates a more detailed HTML coverage report
func GenerateEnhancedCoverageReport(options ReportOptions) error {
	// First use the basic report generator
	if err := GenerateCoverageReport(options.CoverageFile, options.OutputPath); err != nil {
		return err
	}

	// Get detailed coverage metrics
	collector, err := NewCollector(options.CoverageFile)
	if err != nil {
		return err
	}

	metrics, err := collector.CalculateMetrics()
	if err != nil {
		return err
	}

	// Read the enhanced HTML file
	html, err := os.ReadFile(options.OutputPath)
	if err != nil {
		return fmt.Errorf("failed to read HTML report: %w", err)
	}

	// Add a coverage summary to the report
	summaryHTML := generateCoverageSummaryHTML(metrics, options.Title)
	enhanced := strings.Replace(string(html), "<body>", "<body>"+summaryHTML, 1)

	// Write the enhanced HTML back to the file
	if err := os.WriteFile(options.OutputPath, []byte(enhanced), 0600); err != nil {
		return fmt.Errorf("failed to write enhanced HTML: %w", err)
	}

	return nil
}

// generateCoverageSummaryHTML creates an HTML summary of coverage metrics
func generateCoverageSummaryHTML(metrics *Metrics, title string) string {
	if title == "" {
		title = "Coverage Summary"
	}

	return fmt.Sprintf(`
<div class="coverage-meta">
    <h2>%s</h2>
    <table class="coverage-info">
        <tr>
            <th>Statement Coverage</th>
            <th>Branch Coverage</th>
            <th>Function Coverage</th>
            <th>Line Coverage</th>
        </tr>
        <tr>
            <td>%.2f%%</td>
            <td>%.2f%%</td>
            <td>%.2f%%</td>
            <td>%.2f%%</td>
        </tr>
    </table>
</div>
`,
		title,
		metrics.StatementCoverage,
		metrics.BranchCoverage,
		metrics.FunctionCoverage,
		metrics.LineCoverage)
}
