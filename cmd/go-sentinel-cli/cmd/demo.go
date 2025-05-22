package cmd

import (
	"fmt"
	"os"

	"github.com/newbpydev/go-sentinel/cmd/go-sentinel-cli/cmd/demo"
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
		// Check which phase demo to run
		phase, _ := cmd.Flags().GetString("phase")

		switch phase {
		case "1d":
			demo.RunPhase1Demo()
		case "2d":
			demo.RunPhase2Demo()
		case "3d":
			demo.RunPhase3Demo()
		case "4d":
			demo.RunPhase4Demo()
		case "5d":
			demo.RunPhase5Demo()
		case "6d":
			demo.RunPhase6DDemo()
		default:
			fmt.Println("Please specify a valid phase to demo (1d, 2d, 3d, 4d, 5d, or 6d)")
			fmt.Println("Example: go-sentinel-cli demo --phase=1d")
		}
	},
}

func init() {
	rootCmd.AddCommand(demoCmd)

	// Add flags
	demoCmd.Flags().StringP("phase", "p", "", "Phase to run (1-6)")
	if err := demoCmd.MarkFlagRequired("phase"); err != nil {
		fmt.Printf("Error marking phase flag as required: %v\n", err)
		os.Exit(1)
	}
}
