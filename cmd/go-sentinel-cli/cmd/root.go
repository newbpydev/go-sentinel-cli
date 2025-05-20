package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "go-sentinel",
	Short: "A modern test runner for Go with beautiful output",
	Long: `go-sentinel is a modern test runner for Go that provides beautiful, 
Vitest-style output and a great developer experience.

Features:
- Beautiful test output with colors and icons
- Watch mode for continuous testing
- Detailed test summaries and statistics
- Support for parallel test execution`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings
	rootCmd.PersistentFlags().BoolP("color", "c", true, "Enable/disable colored output")
	rootCmd.PersistentFlags().BoolP("watch", "w", false, "Enable watch mode")
}
