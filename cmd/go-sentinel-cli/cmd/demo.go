// Package cmd provides the command-line interface for go-sentinel.
package cmd

// import (
// 	"fmt"
// 	"os"
// 	"time"

// 	"github.com/newbpydev/go-sentinel/internal/cli"
// 	"github.com/spf13/cobra"
// )

// var demoCmd = &cobra.Command{
// 	Use:   "demo",
// 	Short: "Show a demo of the Vitest-style output",
// 	Long:  `Demonstrates the beautiful Vitest-style test output format.`,
// 	Run: func(cmd *cobra.Command, _ []string) {
// 		// Create a renderer
// 		useColors, _ := cmd.Flags().GetBool("color")
// 		renderer := cli.NewRendererWithStyle(os.Stdout, useColors)

// 		// Create a sample test run with pre-populated timing data
// 		run := &cli.TestRun{
// 			StartTime:         time.Now().Add(-26 * time.Second),
// 			EndTime:           time.Now(),
// 			Duration:          26*time.Second + 170*time.Millisecond,
// 			TransformDuration: 859 * time.Millisecond,
// 			SetupDuration:     34*time.Second + 480*time.Millisecond,
// 			CollectDuration:   1*time.Second + 290*time.Millisecond,
// 			TestsDuration:     1 * time.Second,
// 			EnvDuration:       78*time.Second + 910*time.Millisecond,
// 			PrepareDuration:   3*time.Second + 690*time.Millisecond,
// 			NumTotal:          78,
// 			NumPassed:         70,
// 			NumFailed:         8,
// 			NumSkipped:        0,
// 		}

// 		// Create a suite with a failed test
// 		failedSuite := &cli.TestSuite{
// 			Package:    "test/websocket.test.ts",
// 			FilePath:   "test/websocket.test.ts",
// 			NumTotal:   1,
// 			NumPassed:  0,
// 			NumFailed:  1,
// 			NumSkipped: 0,
// 		}

// 		// Add a failed test
// 		failedTest := &cli.TestResult{
// 			Name:   "WebSocketClient > disconnect method > should close the WebSocket connection",
// 			Status: cli.TestStatusFailed,
// 			Error: &cli.TestError{
// 				Message: "TypeError: wsClient.connect is not a function",
// 				Location: &cli.SourceLocation{
// 					File: "test/websocket.test.ts",
// 					Line: 203,
// 				},
// 			},
// 		}
// 		failedSuite.Tests = append(failedSuite.Tests, failedTest)
// 		run.Suites = append(run.Suites, failedSuite)

// 		// Create some passing suites
// 		for i := 0; i < 7; i++ {
// 			passSuite := &cli.TestSuite{
// 				Package:    fmt.Sprintf("test/module%d.test.ts", i+1),
// 				FilePath:   fmt.Sprintf("test/module%d.test.ts", i+1),
// 				NumTotal:   10,
// 				NumPassed:  10,
// 				NumFailed:  0,
// 				NumSkipped: 0,
// 			}
// 			run.Suites = append(run.Suites, passSuite)
// 		}

// 		// Render the final summary
// 		fmt.Print("\n\n=== Demo of Vitest-style Summary ===\n\n")
// 		renderer.RenderFinalSummary(run)
// 	},
// }

// func init() {
// 	rootCmd.AddCommand(demoCmd)
// }
