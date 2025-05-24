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

		// Extract flags directly from cobra command and create Args struct
		watchFlag, _ := cmd.Flags().GetBool("watch")
		colorFlag, _ := cmd.Flags().GetBool("color")
		verboseFlag, _ := cmd.Flags().GetBool("verbose")
		failFastFlag, _ := cmd.Flags().GetBool("fail-fast")
		optimizedFlag, _ := cmd.Flags().GetBool("optimized")
		testPattern, _ := cmd.Flags().GetString("test")
		optimizationMode, _ := cmd.Flags().GetString("optimization")

		// Handle no-color flag
		if noColor, _ := cmd.Flags().GetBool("no-color"); noColor {
			colorFlag = false
		}

		// Create Args struct
		parser := cli.NewArgParser()
		cliArgs := parser.ParseFromCobra(
			watchFlag,
			colorFlag,
			verboseFlag,
			failFastFlag,
			optimizedFlag,
			args,
			testPattern,
			optimizationMode,
		)

		// Convert to string slice for compatibility with existing Run method
		cliArgsSlice := convertArgsToSlice(cliArgs)

		// Run the application
		return controller.Run(cliArgsSlice)
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

	// Add optimization flags
	runCmd.Flags().BoolP("optimized", "o", false, "Enable optimized test execution with Go's built-in caching")
	runCmd.Flags().String("optimization", "", "Set optimization mode (conservative, balanced, aggressive)")
}

// buildCLIArgs converts cobra command flags and args to CLI args format
func buildCLIArgs(cmd *cobra.Command, args []string) []string {
	var cliArgs []string

	// Add different types of flags
	cliArgs = append(cliArgs, buildVerbosityArgs(cmd)...)
	cliArgs = append(cliArgs, buildColorArgs(cmd)...)
	cliArgs = append(cliArgs, buildModeArgs(cmd)...)
	cliArgs = append(cliArgs, buildOptimizationArgs(cmd)...)
	cliArgs = append(cliArgs, buildTestArgs(cmd)...)
	cliArgs = append(cliArgs, buildExecutionArgs(cmd)...)

	// Add package arguments
	cliArgs = append(cliArgs, args...)

	return cliArgs
}

// buildVerbosityArgs builds verbosity-related CLI arguments
func buildVerbosityArgs(cmd *cobra.Command) []string {
	var args []string

	if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
		args = append(args, "-v")
	}

	if verbosity, _ := cmd.Flags().GetCount("verbosity"); verbosity > 0 {
		for i := 0; i < verbosity; i++ {
			args = append(args, "-v")
		}
	}

	return args
}

// buildColorArgs builds color-related CLI arguments
func buildColorArgs(cmd *cobra.Command) []string {
	var args []string

	if noColor, _ := cmd.Flags().GetBool("no-color"); noColor {
		args = append(args, "--no-color")
	} else if color, _ := cmd.Flags().GetBool("color"); color {
		args = append(args, "--color")
	}

	return args
}

// buildModeArgs builds mode-related CLI arguments
func buildModeArgs(cmd *cobra.Command) []string {
	var args []string

	if watch, _ := cmd.Flags().GetBool("watch"); watch {
		args = append(args, "--watch")
	}

	if failFast, _ := cmd.Flags().GetBool("fail-fast"); failFast {
		args = append(args, "--fail-fast")
	}

	return args
}

// buildOptimizationArgs builds optimization-related CLI arguments
func buildOptimizationArgs(cmd *cobra.Command) []string {
	var args []string

	if optimized, _ := cmd.Flags().GetBool("optimized"); optimized {
		args = append(args, "--optimized")
	}

	if optimization, _ := cmd.Flags().GetString("optimization"); optimization != "" {
		args = append(args, "--optimization="+optimization)
	}

	return args
}

// buildTestArgs builds test-related CLI arguments
func buildTestArgs(cmd *cobra.Command) []string {
	var args []string

	if testPattern, _ := cmd.Flags().GetString("test"); testPattern != "" {
		args = append(args, "--test="+testPattern)
	}

	return args
}

// buildExecutionArgs builds execution-related CLI arguments
func buildExecutionArgs(cmd *cobra.Command) []string {
	var args []string

	if parallel, _ := cmd.Flags().GetInt("parallel"); parallel > 0 {
		args = append(args, fmt.Sprintf("--parallel=%d", parallel))
	}

	if timeout, _ := cmd.Flags().GetDuration("timeout"); timeout > 0 {
		args = append(args, fmt.Sprintf("--timeout=%v", timeout))
	}

	return args
}

// convertArgsToSlice converts Args struct back to string slice for compatibility
func convertArgsToSlice(args *cli.Args) []string {
	var result []string

	// Handle verbosity levels
	if args.Verbosity > 0 {
		for i := 0; i < args.Verbosity; i++ {
			result = append(result, "-v")
		}
	}

	// Handle color flags
	if !args.Colors {
		result = append(result, "--no-color")
	} else {
		result = append(result, "--color")
	}

	// Handle watch mode
	if args.Watch {
		result = append(result, "--watch")
	}

	// Handle optimization flags
	if args.Optimized {
		result = append(result, "--optimized")
	}

	if args.OptimizationMode != "" {
		result = append(result, "--optimization="+args.OptimizationMode)
	}

	// Handle test pattern
	if args.TestPattern != "" {
		result = append(result, "--test="+args.TestPattern)
	}

	// Handle fail fast
	if args.FailFast {
		result = append(result, "--fail-fast")
	}

	// Add package arguments
	result = append(result, args.Packages...)

	return result
}
