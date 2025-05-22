package cmd

import (
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

		// TODO: Implement the Vitest-style CLI using the new architecture
		fmt.Println("Running tests with Vitest-style output...")
		fmt.Println("Directory:", dir)
		fmt.Println("Colors:", useColors)
		fmt.Println("Watch mode:", watchMode)
		fmt.Println("Fail fast:", failFast)
		fmt.Println("Verbose:", verbose)

		// TODO: Replace with actual implementation
		if len(args) > 0 {
			fmt.Println("Packages:", args)
		} else {
			fmt.Println("Packages: [current directory]")
		}

		// TODO: Implement actual test runner and renderer
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Add run-specific flags
	runCmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")
	runCmd.Flags().BoolP("fail-fast", "f", false, "Stop on first failure")
	runCmd.Flags().BoolP("color", "c", true, "Use colored output")
	runCmd.Flags().BoolP("watch", "w", false, "Watch for file changes and re-run tests")
}
