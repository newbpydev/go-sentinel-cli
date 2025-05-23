package cmd

import (
	"fmt"

	"github.com/newbpydev/go-sentinel/internal/cli"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [flags] [packages]",
	Short: "Run tests with beautiful output",
	Long: `Run Go tests with beautiful, Vitest-style output.
If no packages are specified, runs tests in the current directory and subdirectories.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create the application controller
		controller := cli.NewAppController()

		// Collect command-line arguments for the controller
		cliArgs := buildCLIArgs(cmd, args)

		// Run the application
		return controller.Run(cliArgs)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Add run-specific flags
	runCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	runCmd.Flags().CountP("verbosity", "q", "Verbosity level (can be repeated: -v, -vv, -vvv)")
	runCmd.Flags().BoolP("fail-fast", "f", false, "Stop on first failure")
	runCmd.Flags().BoolP("color", "c", true, "Use colored output")
	runCmd.Flags().BoolP("no-color", "", false, "Disable colored output")
	runCmd.Flags().BoolP("watch", "w", false, "Watch for file changes and re-run tests")
	runCmd.Flags().StringP("test", "t", "", "Run only tests matching pattern")
	runCmd.Flags().IntP("parallel", "j", 0, "Number of tests to run in parallel")
	runCmd.Flags().DurationP("timeout", "", 0, "Timeout for test execution")
}

// buildCLIArgs converts cobra command flags and args to CLI args format
func buildCLIArgs(cmd *cobra.Command, args []string) []string {
	var cliArgs []string

	// Handle verbosity levels
	if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
		cliArgs = append(cliArgs, "-v")
	}

	if verbosity, _ := cmd.Flags().GetCount("verbosity"); verbosity > 0 {
		for i := 0; i < verbosity; i++ {
			cliArgs = append(cliArgs, "-v")
		}
	}

	// Handle color flags
	if noColor, _ := cmd.Flags().GetBool("no-color"); noColor {
		cliArgs = append(cliArgs, "--no-color")
	} else if color, _ := cmd.Flags().GetBool("color"); color {
		cliArgs = append(cliArgs, "--color")
	}

	// Handle watch mode
	if watch, _ := cmd.Flags().GetBool("watch"); watch {
		cliArgs = append(cliArgs, "--watch")
	}

	// Handle test pattern
	if testPattern, _ := cmd.Flags().GetString("test"); testPattern != "" {
		cliArgs = append(cliArgs, "--test="+testPattern)
	}

	// Handle parallel execution
	if parallel, _ := cmd.Flags().GetInt("parallel"); parallel > 0 {
		cliArgs = append(cliArgs, fmt.Sprintf("--parallel=%d", parallel))
	}

	// Handle timeout
	if timeout, _ := cmd.Flags().GetDuration("timeout"); timeout > 0 {
		cliArgs = append(cliArgs, fmt.Sprintf("--timeout=%v", timeout))
	}

	// Handle fail fast
	if failFast, _ := cmd.Flags().GetBool("fail-fast"); failFast {
		cliArgs = append(cliArgs, "--fail-fast")
	}

	// Add package arguments
	cliArgs = append(cliArgs, args...)

	return cliArgs
}
