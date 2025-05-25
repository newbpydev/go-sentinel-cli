package runner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTestRunnerFunctions(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		testFile bool
		goFile   bool
	}{
		{
			name:     "go test file",
			path:     "example_test.go",
			testFile: true,
			goFile:   true,
		},
		{
			name:     "go implementation file",
			path:     "example.go",
			testFile: false,
			goFile:   true,
		},
		{
			name:     "non-go file",
			path:     "example.txt",
			testFile: false,
			goFile:   false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if IsGoTestFile(tc.path) != tc.testFile {
				t.Errorf("IsGoTestFile(%s) = %v, want %v", tc.path, IsGoTestFile(tc.path), tc.testFile)
			}

			if IsGoFile(tc.path) != tc.goFile {
				t.Errorf("IsGoFile(%s) = %v, want %v", tc.path, IsGoFile(tc.path), tc.goFile)
			}
		})
	}
}

func TestTestRunner(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "testrunner-test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Logf("failed to remove temp dir: %v", removeErr)
		}
	}()

	// Create a simple test file
	testFile := filepath.Join(tempDir, "simple_test.go")
	testContent := `package simple_test

import "testing"

func TestPassing(t *testing.T) {
	// This test always passes
}
`
	// #nosec G306 - Test file, permissions not important
	if writeErr := os.WriteFile(testFile, []byte(testContent), 0600); writeErr != nil {
		t.Fatalf("failed to create test file: %v", writeErr)
	}

	// Create a runner
	runner := &TestRunner{
		Verbose:    true,
		JSONOutput: true,
	}

	// Run the test
	ctx := context.Background()
	output, err := runner.Run(ctx, []string{testFile})
	if err != nil {
		t.Fatalf("failed to run test: %v", err)
	}

	// Verify output contains JSON
	if !strings.Contains(output, `"Action":"pass"`) {
		t.Errorf("expected JSON output with pass action, got: %s", output)
	}

	if !strings.Contains(output, `"Test":"TestPassing"`) {
		t.Errorf("expected output for TestPassing, got: %s", output)
	}
}
