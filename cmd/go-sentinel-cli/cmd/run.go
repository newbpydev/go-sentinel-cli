package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [flags] [packages]",
	Short: "Run tests with beautiful output",
	Long: `Run Go tests with beautiful, Vitest-style output.
If no packages are specified, runs tests in the current directory and subdirectories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get working directory
		dir, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error getting current directory: %v", err)
		}

		// Get flags
		useColors, _ := cmd.Flags().GetBool("color")
		watchMode, _ := cmd.Flags().GetBool("watch")
		failFast, _ := cmd.Flags().GetBool("fail-fast")
		verbose, _ := cmd.Flags().GetBool("verbose")

		// Create renderer with color setting
		renderer := cli.NewRendererWithStyle(os.Stdout, useColors)

		// Create and configure runner
		runner, err := cli.NewRunner(dir)
		if err != nil {
			return fmt.Errorf("error creating runner: %v", err)
		}
		defer runner.Stop()

		// Set up run options
		opts := cli.RunOptions{
			Watch:    watchMode,
			FailFast: failFast,
			Renderer: renderer,
		}

		// If packages were specified, add them to options
		if len(args) > 0 {
			opts.Packages = args
		}

		// Run tests
		ctx := context.Background()
		if err := runner.Run(ctx, opts); err != nil {
			if verbose {
				return fmt.Errorf("error running tests: %v", err)
			}
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Add run-specific flags
	runCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	runCmd.Flags().BoolP("fail-fast", "f", false, "Stop on first failure")
}
