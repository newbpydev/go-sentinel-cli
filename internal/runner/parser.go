package runner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

// TestEvent represents a single event from 'go test -json' output.
type TestEvent struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test,omitempty"`
	Output  string  `json:"Output,omitempty"`
	Elapsed float64 `json:"Elapsed,omitempty"`
}

// ParseTestEvents reads a stream of JSON lines from r and returns parsed TestEvents.
func ParseTestEvents(r io.Reader) ([]TestEvent, error) {
	var events []TestEvent
	dec := json.NewDecoder(r)
	for {
		var ev TestEvent
		if err := dec.Decode(&ev); err != nil {
			if err == io.EOF {
				break
			}
			return events, err
		}
		events = append(events, ev)
	}
	return events, nil
}

// GroupTestEvents groups events by package and test name.
func GroupTestEvents(events []TestEvent) map[string]map[string][]TestEvent {
	grouped := make(map[string]map[string][]TestEvent)
	for _, ev := range events {
		pkg := ev.Package
		test := ev.Test
		if _, ok := grouped[pkg]; !ok {
			grouped[pkg] = make(map[string][]TestEvent)
		}
		grouped[pkg][test] = append(grouped[pkg][test], ev)
	}
	return grouped
}

// FileLocation represents a file:line location extracted from test output.
type FileLocation struct {
	File string
	Line int
}

// ErrorContext holds error message and file location information for a failed test.
type ErrorContext struct {
	Message      string
	FileLocation *FileLocation
}

// ExtractErrorContext scans output events for file:line information and error messages.
// It parses test output to extract file locations and error details from failed tests,
// which can be used to provide more context in the UI or for debugging purposes.
func ExtractErrorContext(events []TestEvent) *ErrorContext {
	if len(events) == 0 {
		return nil
	}

	fileLineRe := regexp.MustCompile(`(?m)^\s*([\w./-]+):(\d+):?\s*(.*)$`)
	var lastOutput string

	for _, ev := range events {
		if ev.Action == "output" {
			msg := ev.Output
			lastOutput = msg
			if matches := fileLineRe.FindStringSubmatch(msg); matches != nil {
				file := matches[1]
				lineStr := matches[2]
				var line int
				_, err := fmt.Sscanf(lineStr, "%d", &line)
				if err == nil {
					return &ErrorContext{
						Message:      msg,
						FileLocation: &FileLocation{File: file, Line: line},
					}
				}
			}
		}
	}

	// If there is any output, return a context with the last output as the message
	if lastOutput != "" {
		return &ErrorContext{
			Message: lastOutput,
		}
	}

	return nil
}

// TestResult is a UI-facing summary of a test run.
type TestResult struct {
	Package      string
	Test         string
	Passed       bool
	Duration     float64
	ErrorContext *ErrorContext
	OutputLines  []string
}

// SummarizeTestResults converts grouped events into structured results for UI.
func SummarizeTestResults(grouped map[string]map[string][]TestEvent) []TestResult {
	var results []TestResult
	for pkg, tests := range grouped {
		for test, events := range tests {
			var passed bool
			var duration float64
			var outputLines []string
			var errCtx *ErrorContext
			for _, ev := range events {
				if ev.Action == "pass" {
					passed = true
					duration = ev.Elapsed
				}
				if ev.Action == "fail" {
					passed = false
					duration = ev.Elapsed
				}
				if ev.Action == "output" && ev.Output != "" {
					outputLines = append(outputLines, ev.Output)
				}
			}
			errCtx = ExtractErrorContext(events)
			results = append(results, TestResult{
				Package:      pkg,
				Test:         test,
				Passed:       passed,
				Duration:     duration,
				ErrorContext: errCtx,
				OutputLines:  outputLines,
			})
		}
	}
	return results
}

// ParseTestOutput parses test output into structured test results.
// It handles various test output formats including passing, failing, skipped tests,
// subtests, timestamps, and coverage information.
func ParseTestOutput(r io.Reader) []TestEvent {
	var events []TestEvent
	var currentPkg string
	var testStack []string
	var lastTime = time.Now().UTC()
	var lastStatusTest string
	var bufferedOrder []string
	var bufferedRunEvents = make(map[string]TestEvent)
	var bufferedStatusEvents = make(map[string]TestEvent)
	var bufferedOutputs = make(map[string][]string)
	var seenTest = make(map[string]bool)
	var lastStatusEvent *TestEvent // NEW: track last status event for indented output

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Skip summary lines
		if trimmed == "PASS" || trimmed == "FAIL" || strings.HasPrefix(trimmed, "exit status") {
			continue
		}

		// Package status line
		if (strings.HasPrefix(trimmed, "ok  ") || strings.HasPrefix(trimmed, "FAIL")) && len(strings.Fields(trimmed)) >= 2 {
			parts := strings.Fields(trimmed)
			currentPkg = parts[1]
			// Assign package to all buffered events and flush them in order
			for _, testName := range bufferedOrder {
				re, hasRun := bufferedRunEvents[testName]
				se, hasStatus := bufferedStatusEvents[testName]
				if hasRun {
					re.Package = currentPkg
					lastTime = lastTime.Add(time.Nanosecond)
					re.Time = lastTime.Format(time.RFC3339)
					events = append(events, re)
				}
				if hasStatus {
					se.Package = currentPkg
					if outs, ok := bufferedOutputs[testName]; ok && len(outs) > 0 {
						if se.Output != "" {
							se.Output += "\n" + strings.Join(outs, "\n")
						} else {
							se.Output = strings.Join(outs, "\n")
						}
					}
					lastTime = lastTime.Add(time.Nanosecond)
					se.Time = lastTime.Format(time.RFC3339)
					// Only set Output if non-empty for status events
					if se.Output == "" {
						// Omit Output field by using struct literal
						events = append(events, TestEvent{
							Time:    se.Time,
							Action:  se.Action,
							Package: se.Package,
							Test:    se.Test,
							Elapsed: se.Elapsed,
						})
					} else {
						events = append(events, se)
					}
				}
			}
			bufferedOrder = nil
			bufferedRunEvents = make(map[string]TestEvent)
			bufferedStatusEvents = make(map[string]TestEvent)
			bufferedOutputs = make(map[string][]string)
			seenTest = make(map[string]bool)
			continue
		}

		// RUN marker
		if strings.HasPrefix(trimmed, "=== RUN") {
			testName := strings.TrimSpace(strings.TrimPrefix(trimmed, "=== RUN"))
			lastTime = lastTime.Add(time.Millisecond)
			re := TestEvent{
				Time:   lastTime.Format(time.RFC3339),
				Action: "run",
				Test:   testName,
			}
			bufferedRunEvents[testName] = re
			if !seenTest[testName] {
				bufferedOrder = append(bufferedOrder, testName)
				seenTest[testName] = true
			}
			// Maintain testStack for subtest ordering
			if len(testStack) == 0 || !strings.HasPrefix(testName, testStack[len(testStack)-1]+"/") {
				testStack = append(testStack, testName)
			}
			continue
		}

		// Indented subtest status line (e.g., '    --- PASS: TestParent/SubtestA (0.10s)')
		if strings.HasPrefix(line, "    --- ") {
			parts := strings.Fields(trimmed)
			if len(parts) < 4 {
				continue
			}
			status := strings.ToLower(parts[1])
			testName := parts[2]
			duration := 0.0
			if len(parts) > 3 && strings.HasPrefix(parts[3], "(") && strings.HasSuffix(parts[3], ")") {
				durStr := strings.Trim(parts[3], "()s")
				if _, err := fmt.Sscanf(durStr, "%f", &duration); err != nil {
					duration = 0.0
				}
			}
			lastTime = lastTime.Add(time.Millisecond)
			se := TestEvent{
				Time:    lastTime.Format(time.RFC3339),
				Action:  status,
				Test:    testName,
				Elapsed: duration,
			}
			bufferedStatusEvents[testName] = se
			lastStatusTest = testName
			lastStatusEvent = &se
			if !seenTest[testName] {
				// Insert subtest run/status after parent run
				parent := ""
				if idx := strings.LastIndex(testName, "/"); idx != -1 {
					parent = testName[:idx]
				}
				insertIdx := 0
				for i, name := range bufferedOrder {
					if name == parent {
						insertIdx = i + 1
						break
					}
				}
				// Insert run and status for subtest in order
				bufferedOrder = append(bufferedOrder[:insertIdx], append([]string{testName}, bufferedOrder[insertIdx:]...)...)
				seenTest[testName] = true
			}
			continue
		}

		// Indented output line (e.g., error/skip message)
		if strings.HasPrefix(line, "    ") && lastStatusEvent != nil {
			msg := strings.TrimSpace(line)
			if lastStatusEvent.Output != "" {
				lastStatusEvent.Output += "\n" + msg
			} else {
				lastStatusEvent.Output = msg
			}
			bufferedStatusEvents[lastStatusEvent.Test] = *lastStatusEvent
			continue
		}

		// PASS/FAIL/SKIP marker
		if strings.HasPrefix(trimmed, "--- ") {
			parts := strings.Fields(trimmed)
			if len(parts) < 4 {
				continue
			}
			status := strings.ToLower(parts[1])
			testName := parts[2]
			duration := 0.0
			if len(parts) > 3 && strings.HasPrefix(parts[3], "(") && strings.HasSuffix(parts[3], ")") {
				durStr := strings.Trim(parts[3], "()s")
				if _, err := fmt.Sscanf(durStr, "%f", &duration); err != nil {
					duration = 0.0
				}
			}
			lastTime = lastTime.Add(time.Millisecond)
			se := TestEvent{
				Time:    lastTime.Format(time.RFC3339),
				Action:  status,
				Test:    testName,
				Elapsed: duration,
			}
			// Attach output to status event and clear buffer
			if outs, ok := bufferedOutputs[testName]; ok && len(outs) > 0 {
				se.Output = strings.Join(outs, "\n")
				bufferedOutputs[testName] = nil
			}
			bufferedStatusEvents[testName] = se
			lastStatusTest = testName
			lastStatusEvent = &se
			if !seenTest[testName] {
				bufferedOrder = append(bufferedOrder, testName)
				seenTest[testName] = true
			}
			// Remove from stack if present
			if len(testStack) > 0 && testStack[len(testStack)-1] == testName {
				testStack = testStack[:len(testStack)-1]
			}
			continue
		}

		// Coverage line
		if strings.Contains(trimmed, "coverage:") {
			// Attach to last status event for the package
			if lastStatusTest != "" {
				if bufferedOutputs[lastStatusTest] == nil {
					bufferedOutputs[lastStatusTest] = []string{trimmed}
				} else {
					bufferedOutputs[lastStatusTest] = append(bufferedOutputs[lastStatusTest], trimmed)
				}
			}
			continue
		}

		// Output line (attach to last test in stack)
		if len(testStack) > 0 {
			lastTest := testStack[len(testStack)-1]
			bufferedOutputs[lastTest] = append(bufferedOutputs[lastTest], trimmed)
		}
	}

	// Flush any remaining buffered events (for the last package)
	for _, testName := range bufferedOrder {
		re, hasRun := bufferedRunEvents[testName]
		se, hasStatus := bufferedStatusEvents[testName]
		if hasRun {
			if re.Package == "" {
				re.Package = currentPkg
			}
			events = append(events, re)
		}
		if hasStatus {
			if se.Package == "" {
				se.Package = currentPkg
			}
			if outs, ok := bufferedOutputs[testName]; ok && len(outs) > 0 {
				if se.Output != "" {
					se.Output += "\n" + strings.Join(outs, "\n")
				} else {
					se.Output = strings.Join(outs, "\n")
				}
			}
			// Only set Output if non-empty for status events
			if se.Output == "" {
				// Omit Output field by using struct literal
				events = append(events, TestEvent{
					Time:    se.Time,
					Action:  se.Action,
					Package: se.Package,
					Test:    se.Test,
					Elapsed: se.Elapsed,
				})
			} else {
				events = append(events, se)
			}
		}
	}

	return events
}
