package coverage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunTestsWithCoverage(t *testing.T) {
	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "coverage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Define the coverage output path
	coverageFile := filepath.Join(tempDir, "coverage.out")

	// Set up test context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run tests with coverage on the current package
	options := TestRunnerOptions{
		PackagePaths: []string{"./..."},
		OutputPath:   coverageFile,
		Timeout:      10 * time.Second,
	}

	// Run the tests
	err = RunTestsWithCoverage(ctx, options)
	if err != nil {
		t.Logf("Error running tests with coverage: %v", err)
		// Don't fail the test because the target code might have failing tests
	}

	// Check if the coverage file was created
	if _, err := os.Stat(coverageFile); os.IsNotExist(err) {
		t.Fatal("Expected coverage file to be created, but it doesn't exist")
	}

	// Check if we can create a collector from the coverage file
	collector, err := NewCollector(coverageFile)
	if err != nil {
		t.Fatalf("Failed to create collector from coverage file: %v", err)
	}

	// Check if we can calculate metrics
	metrics, err := collector.CalculateMetrics()
	if err != nil {
		t.Fatalf("Failed to calculate metrics: %v", err)
	}

	if metrics == nil {
		t.Fatal("Expected metrics to be non-nil")
	}
}

func TestFindAllPackages(t *testing.T) {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Find all packages
	packages, err := FindAllPackages(cwd)
	if err != nil {
		t.Fatalf("Failed to find packages: %v", err)
	}

	// We should find at least the current package
	if len(packages) == 0 {
		t.Error("Expected to find at least one package")
	}

	// Check that the current package is included
	foundSelf := false
	for _, pkg := range packages {
		if pkg == "github.com/newbpydev/go-sentinel/internal/coverage" {
			foundSelf = true
			break
		}
	}

	if !foundSelf {
		t.Error("Expected to find the current package")
	}
}
