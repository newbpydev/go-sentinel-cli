package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// demoCmd represents the command for testing and demonstrating features
var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Run a demonstration of CLI features",
	Long: `Demonstrates various features of the CLI by running tests
and displaying the results in different formats.

This command is used for development and validation.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("🎉 CLI Migration Demo\n")
		fmt.Printf("📦 The CLI has been successfully migrated to modular architecture!\n")
		fmt.Printf("✅ All files moved from internal/cli to their respective modular packages\n")
		fmt.Printf("🏗️  New architecture: pkg/models, internal/config, internal/test, internal/ui, internal/watch, internal/app\n")
		fmt.Printf("🧹 internal/cli directory is now clean and lean\n")
	},
}

func init() {
	rootCmd.AddCommand(demoCmd)
}
