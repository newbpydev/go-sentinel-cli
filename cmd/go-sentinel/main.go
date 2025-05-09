package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	clipboard "github.com/atotto/clipboard"
	"github.com/yourusername/go-sentinel/internal/ui"
)

func main() {
	// Simulated static test results
	staticResults := []ui.TestResult{
		{Package: "pkg/foo", Passed: true, Summary: "ok   pkg/foo  0.05s"},
		{Package: "pkg/bar", Passed: false, Summary: "FAIL pkg/bar  0.10s", Message: "expected true, got false", File: "main.go", Line: 42},
		{Package: "pkg/baz", Passed: true, Summary: "ok   pkg/baz  0.02s"},
		{Package: "pkg/qux", Passed: false, Summary: "FAIL pkg/qux  0.20s", Message: "panic: index out of range"},
	}
	resultsCh := make(chan []ui.TestResult, 2)
	requestCh := make(chan struct{}, 2)

	// Simulated runner/controller goroutine
	go func() {
		for range requestCh {
			time.Sleep(400 * time.Millisecond) // Simulate test run delay
			resultsCh <- staticResults
		}
	}()

	uiState := ui.NewUI()
	reader := bufio.NewReader(os.Stdin)
	filterFailures := false
	results := staticResults

	// Initial run
	requestCh <- struct{}{}

	for {
		// Wait for new results from runner
		select {
		case results = <-resultsCh:
			uiState.SetResults(results)
		case <-time.After(10 * time.Millisecond):
			// Allow UI to remain responsive
		}

		// Clear the screen (cross-platform ANSI)
		fmt.Print("\033[H\033[2J")

		output := &bytes.Buffer{}
		fmt.Fprintln(output, "Go Sentinel CLI (MVP)")
		fmt.Fprintln(output, "====================")
		visible := results
		if filterFailures {
			uiState.SetFilterFailures(true)
			visible = uiState.VisibleResults()
		} else {
			uiState.SetFilterFailures(false)
		}
		for _, r := range visible {
			label := "\x1b[42m\x1b[30m PASS \x1b[0m"
			if !r.Passed {
				label = "\x1b[41m\x1b[30m FAIL \x1b[0m"
			}
			fmt.Fprintf(output, "%s %s\n", label, r.Summary)
			if !r.Passed {
				// Print failure message
				if r.Message != "" {
					fmt.Fprintf(output, "    %s\n", r.Message)
				}
				// Print code snippet if file and line are available
				if r.File != "" && r.Line > 0 {
					file, err := os.Open(r.File)
					if err == nil {
						defer file.Close()
						scanner := bufio.NewScanner(file)
						var lines []string
						for scanner.Scan() {
							lines = append(lines, scanner.Text())
						}
						start := r.Line - 3
						if start < 0 { start = 0 }
						end := r.Line + 2
						if end > len(lines) { end = len(lines) }
						for i := start; i < end; i++ {
							lnum := i + 1
							code := lines[i]
							arrow := "  "
							prefix := ""
							reset := ""
							if lnum == r.Line {
								arrow = "> "
								prefix = "\x1b[41m"
								reset = "\x1b[0m"
							}
							fmt.Fprintf(output, "%s%3d |\t%s%s%s\n", arrow, lnum, prefix, code, reset)
						}
					}
				}
			}
		}

		// Failure context when filtering failures or if any failures are visible
		if filterFailures && len(visible) > 0 {
			fmt.Fprintln(output, "\nFailure Context:")
			for _, r := range visible {
				if !r.Passed {
					ctx := uiState.CodeContextForFailure(r)
					fmt.Fprintf(output, "- %s\n", ctx)
				}
			}
		}

		totalSuites := 1
		passedSuites := 1
		totalTests := len(results)
		passedTests := 0
		for _, r := range results {
			if r.Passed {
				passedTests++
			}
		}
		totalTime := 0.37 // seconds, sum of all test times (static for now)

		fmt.Fprintln(output, "")
		fmt.Fprintf(output, "Test Suites: \x1b[32m%d passed\x1b[0m, %d total\n", passedSuites, totalSuites)
		fmt.Fprintf(output, "Tests:       \x1b[32m%d passed\x1b[0m, %d total\n", passedTests, totalTests)
		fmt.Fprintf(output, "Time:        \x1b[33m%.3fs\x1b[0m\n", totalTime)
		fmt.Fprintln(output, "\nRan all test suites.")
		fmt.Fprintln(output, "Done in 0.37s.")

		filterLabel := ""
		if filterFailures {
			filterLabel = "[ON]"
		} else {
			filterLabel = "[OFF]"
		}
		fmt.Fprintf(output, "\n[Enter] Rerun   [f] Filter failures %s   [r] Refresh   [c] Copy all failures   [C] Select failures   [q] Quit\n", filterLabel)
		os.Stdout.Write(output.Bytes())

		fmt.Print(": ")
		input, _ := reader.ReadString('\n')
		if len(input) > 0 {
			switch input[0] {
			case 'C':
				// Always set results before entering selection mode to avoid stale/empty state
				uiState.SetResults(results)
				uiState.DeselectAll()
				for {
					failures := uiState.VisibleResults()
					var selectable []ui.TestResult
					for _, r := range failures {
						if !r.Passed {
							selectable = append(selectable, r)
						}
					}
					if len(selectable) == 0 {
						fmt.Println("No failures to select.")
						fmt.Print("Press Enter to continue...")
						reader.ReadString('\n')
						break
					}
					fmt.Print("\033[H\033[2J") // Clear screen
					fmt.Println("Select test failures (toggle: 1-9, space=all, Enter=copy, q=quit):")
					// Build mapping from visible (filtered) failures to real indices in results
					visibleToReal := make([]int, 0, len(selectable))
					for _, r := range selectable {
						for realIdx, rr := range results {
							if rr.Package == r.Package && rr.Summary == r.Summary {
								visibleToReal = append(visibleToReal, realIdx)
								break
							}
						}
					}
					for idx, r := range selectable {
						selected := "\x1b[2m[ ]\x1b[0m" // dim by default
						_ = visibleToReal[idx] // keep mapping for selection, but not needed for display
						if uiState.SelectedTests() != nil {
							for _, sel := range uiState.SelectedTests() {
								if sel.Package == r.Package && sel.Summary == r.Summary {
									selected = "\x1b[32m[x]\x1b[0m" // green for selected
								}
							}
						}
						label := "\x1b[41m\x1b[30m FAIL \x1b[0m"
						if r.Passed {
							label = "\x1b[42m\x1b[30m PASS \x1b[0m"
						}
						fmt.Printf("%d. %s %s %s\n", idx+1, selected, label, r.Summary)
					}
					fmt.Print("Toggle (1-9), a=all, Enter=copy, q=quit: ")
					inputSel, _ := reader.ReadString('\n')
					if len(inputSel) > 0 {
						// Remove trailing newline from input
						inputSel = strings.TrimSuffix(inputSel, "\n")
						inputSel = strings.TrimSuffix(inputSel, "\r")

						// Handle keys based on cleaned input
						if inputSel == "q" {
							// Quit selection mode
							uiState.ClearSelection()
							fmt.Println("Exited selection mode.")
							fmt.Print("Press Enter to continue...")
							reader.ReadString('\n')
							break
						} else if inputSel == "a" {
							// Toggle all: select all if any unselected, else deselect all
							// First, count how many are currently selected without modifying
							selectedCount := 0
							totalCount := len(selectable)
							for _, sel := range uiState.SelectedTests() {
								for _, r := range selectable {
									if sel.Package == r.Package && sel.Summary == r.Summary {
										selectedCount++
										break
									}
								}
							}
							// If all are selected, deselect all; otherwise select all
							if selectedCount == totalCount {
								// Deselect all
								uiState.DeselectAll()
								fmt.Println("Deselected all failures.")
							} else {
								// Select all
								// First clear current selections to avoid duplicates
								uiState.DeselectAll()
								for idx := range selectable {
									realIdx := visibleToReal[idx]
									uiState.SelectTest(realIdx)
								}
								fmt.Println("Selected all failures.")
							}
							continue
						} else if inputSel == "" { // Enter key (empty after trimming)
							copied := uiState.CopySelectedFailures()
							if copied != "" {
								err := clipboard.WriteAll(copied)
								if err != nil {
									fmt.Println("Failed to copy to clipboard:", err)
								} else {
									fmt.Println("Copied selected failures to clipboard:")
									fmt.Println(copied)
								}
							} else {
								fmt.Println("No failures selected to copy.")
							}
							uiState.ClearSelection()
							fmt.Print("Press Enter to continue...")
							reader.ReadString('\n')
							break
						} else if len(inputSel) == 1 && inputSel[0] >= '1' && inputSel[0] <= '9' {
							// Number key for toggling specific tests
							idx := int(inputSel[0] - '1')
							if idx >= 0 && idx < len(selectable) {
								realIdx := visibleToReal[idx]
								selected := uiState.SelectTest(realIdx)
								if selected {
									fmt.Printf("Selected failure %d.\n", idx+1)
								} else {
									fmt.Printf("Deselected failure %d.\n", idx+1)
								}
							}
							continue
						} else {
							// Unrecognized input
							fmt.Printf("DEBUG - Unrecognized input: '%s'\n", inputSel)
							continue
						}
					}
					break
				}
				continue

			case 'q':
				fmt.Println("Exiting.")
				return
			case 'f':
				filterFailures = !filterFailures
				continue
			case 'r':
				requestCh <- struct{}{} // manual refresh
				continue
			case 'c':
				// Copy all failures (select all visible failures and copy)
				// Always copy all failures from the full results, not just filtered/visible
				var allFailures []ui.TestResult
				for _, r := range results {
					if !r.Passed {
						allFailures = append(allFailures, r)
					}
				}
				if len(allFailures) == 0 {
					fmt.Println("No failures to copy.")
					fmt.Print("Press Enter to continue...")
					reader.ReadString('\n')
					continue
				}
				// Select all failures
				uiState.DeselectAll()
				for _, r := range allFailures {
					for realIdx, rr := range results {
						if rr.Package == r.Package && rr.Summary == r.Summary {
							uiState.SelectTest(realIdx)
							break
						}
					}
				}
				copied := uiState.CopySelectedFailures()
				if copied != "" {
					err := clipboard.WriteAll(copied)
					if err != nil {
						fmt.Println("Failed to copy to clipboard:", err)
					} else {
						fmt.Println("Copied all failures to clipboard:")
						fmt.Println(copied)
					}
				} else {
					fmt.Println("No failures to copy.")
				}
				uiState.ClearSelection()
				fmt.Print("Press Enter to continue...")
				reader.ReadString('\n')
				continue
			case '\n':
				requestCh <- struct{}{} // rerun
				continue
			default:
				continue
			}
		}
	}
}
