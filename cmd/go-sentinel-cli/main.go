package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/newbpydev/go-sentinel/internal/cli"
	"github.com/spf13/cobra"
)

var (
	watchMode  bool
	failFast   bool
	onlyFailed bool
	packages   []string
	tests      []string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "go-sentinel",
		Short: "A modern test runner for Go",
		Long: `A beautiful and feature-rich test runner for Go
that aims to make testing delightful.

Features:
- Beautiful terminal output with colors
- Watch mode for automatic test reruns
- File change detection
- Detailed error reporting with source context
- Progress indicators and summaries
- Cross-platform compatibility`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get working directory
			workDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("error getting working directory: %v", err)
			}

			// Create runner
			runner, err := cli.NewRunner(workDir)
			if err != nil {
				return fmt.Errorf("error creating runner: %v", err)
			}
			defer runner.Stop()

			// Configure options
			opts := cli.RunOptions{
				OnlyFailed: onlyFailed,
				FailFast:   failFast,
				Packages:   packages,
				Tests:      tests,
			}

			// Run tests
			if watchMode {
				return runner.Watch(cmd.Context(), opts)
			}

			output, err := runner.RunOnce(opts)
			if err != nil {
				return fmt.Errorf("error running tests: %v", err)
			}

			// Parse and display results
			parser := cli.NewParser()
			run, err := parser.Parse(strings.NewReader(output))
			if err != nil {
				return fmt.Errorf("error parsing test output: %v", err)
			}

			renderer := cli.NewRenderer(os.Stdout)
			renderer.RenderTestRun(run)

			return nil
		},
	}

	// Add flags
	rootCmd.Flags().BoolVarP(&watchMode, "watch", "w", false, "Watch for file changes and rerun tests")
	rootCmd.Flags().BoolVarP(&failFast, "fail-fast", "f", false, "Stop on first test failure")
	rootCmd.Flags().BoolVarP(&onlyFailed, "only-failed", "o", false, "Only run failed tests")
	rootCmd.Flags().StringSliceVarP(&packages, "packages", "p", nil, "Specific packages to test (comma-separated)")
	rootCmd.Flags().StringSliceVarP(&tests, "tests", "t", nil, "Specific tests to run (comma-separated)")

	// Execute
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
