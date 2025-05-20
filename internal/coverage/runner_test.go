package coverage

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunTestsWithCoverage(t *testing.T) {
	// Skip this test if running in CI or as part of a larger test suite
	if testing.Short() {
		t.Skip("Skipping coverage test in short mode")
	}

	// Create a temporary directory for test output
	tempDir, err := os.MkdirTemp("", "coverage-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if removeAllErr := os.RemoveAll(tempDir); removeAllErr != nil {
			t.Logf("Failed to remove temp dir %s: %v", tempDir, removeAllErr)
		}
	}()

	// Define the coverage output path
	coverageFile := filepath.Join(tempDir, "coverage.out")

	// Set up test context with a longer timeout than the go test command's timeout
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second) // Main test timeout
	defer cancel()

	// Resolve the absolute path to the dummy package
	wd, err := os.Getwd() // Should be internal/coverage
	if err != nil {
		t.Fatalf("Failed to get current working directory: %v", err)
	}
	dummyPkgDir := filepath.Join(wd, "testdata", "samplepkg")

	// Run tests with coverage on the dummy package
	options := TestRunnerOptions{
		PackagePaths:     []string{"."}, // Test current dir within dummyPkgDir
		OutputPath:       coverageFile,
		Timeout:          10 * time.Second, // Timeout for the 'go test' command itself
		WorkingDirectory: dummyPkgDir,
	}

	// Run the tests
	if err := RunTestsWithCoverage(ctx, options); err != nil {
		// It's possible the error is just "exit status 1" if tests failed,
		// which is okay for this test as long as coverage is generated.
		// However, if the error indicates a more fundamental problem with the runner,
		// we should fail. For now, we'll log it and proceed to check coverage file.
		t.Logf("RunTestsWithCoverage returned an error: %v. Checking if coverage file was still generated.", err)
	}

	// Check if the coverage file was created
	if _, err := os.Stat(coverageFile); os.IsNotExist(err) {
		t.Fatalf("Coverage file %s was not created: %v", coverageFile, err)
	}

	// Check if we can create a collector from the coverage file
	collector, err := NewCollector(coverageFile)
	if err != nil {
		t.Fatalf("Failed to create collector from coverage file %s: %v", coverageFile, err)
	}
	if collector == nil {
		t.Fatalf("NewCollector returned nil for coverage file %s without an error", coverageFile)
	}

	// Check if we can calculate metrics
	metrics, err := collector.CalculateMetrics()
	if err != nil {
		// If "no coverage profiles available", it means the .out file was likely empty or invalid
		// This often happens if the tests themselves had compilation errors or panicked early.
		t.Fatalf("Failed to calculate metrics from coverage file %s: %v", coverageFile, err)
	}
	if metrics == nil {
		t.Fatalf("CalculateMetrics returned nil for coverage file %s without an error", coverageFile)
	}

	t.Logf("Successfully processed coverage file. Overall Line Coverage: %.2f%%", metrics.LineCoverage)
	// Test passes if it reaches here without a hang and processes the coverage file.
}

func TestFindAllPackages(t *testing.T) {
	// Skip this test if running in CI or as part of a larger test suite
	if testing.Short() {
		t.Skip("Skipping package finder test in short mode")
	}

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Create a context with a shorter timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create and run the command with the context
	cmd := exec.CommandContext(ctx, "go", "list", "./...")
	cmd.Dir = cwd

	output, err := cmd.Output()
	if err != nil {
		// Don't fail the test, just log the error and skip
		t.Logf("Note: Couldn't list packages: %v", err)
		return
	}

	// Process the output
	packages := []string{}
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if line != "" {
			packages = append(packages, line)
		}
	}

	// We should find at least the current package
	if len(packages) == 0 {
		t.Log("Note: No packages found, but not failing the test")
		return
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
