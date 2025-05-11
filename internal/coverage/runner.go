package coverage

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// TestRunnerOptions defines options for running tests with coverage
type TestRunnerOptions struct {
	PackagePaths   []string // List of packages to run tests for
	OutputPath     string   // Where to save the coverage profile
	Timeout        time.Duration
	IncludeCoveredFiles bool // Include files with 100% coverage
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

	// Ensure the output directory exists
	outputDir := filepath.Dir(options.OutputPath)
	if outputDir != "." && outputDir != "/" {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
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
	args = append(args, fmt.Sprintf("-coverprofile=%s", options.OutputPath))
	
	// Add packages to test
	args = append(args, options.PackagePaths...)
	
	// Run the command
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// Execute the command
	err := cmd.Run()
	if err != nil {
		// Even if tests fail, we might still have coverage data
		// Check if the output file was created
		if _, statErr := os.Stat(options.OutputPath); statErr != nil {
			return fmt.Errorf("failed to generate coverage profile: %w", err)
		}
		
		// If the file exists, continue with analysis despite test failures
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
	cmd := exec.Command("go", "tool", "cover", "-html", coverageFile, "-o", htmlOutput)
	return cmd.Run()
}
