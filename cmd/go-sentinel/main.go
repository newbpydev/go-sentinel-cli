package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"time"

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
	resultsCh := make(chan []ui.TestResult)
	requestCh := make(chan struct{})

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
			icon := "✔"
			if !r.Passed {
				icon = "✖"
			}
			colorStart := "\x1b[32m"
			if !r.Passed {
				colorStart = "\x1b[31m"
			}
			fmt.Fprintf(output, "%s %s%s\x1b[0m\n", icon, colorStart, r.Summary)
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
		fmt.Fprintf(output, "\n[Enter] Rerun   [f] Filter failures %s   [r] Refresh   [q] Quit\n", filterLabel)
		os.Stdout.Write(output.Bytes())

		fmt.Print(": ")
		input, _ := reader.ReadString('\n')
		if len(input) > 0 {
			switch input[0] {
			case 'q':
				fmt.Println("Exiting.")
				return
			case 'f':
				filterFailures = !filterFailures
				continue
			case 'r':
				requestCh <- struct{}{} // manual refresh
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

