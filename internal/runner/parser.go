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
	var testOrder []string
	var runEvents = make(map[string]TestEvent)
	var statusEvents = make(map[string]TestEvent)
	var outputs = make(map[string][]string)
	var seenTest = make(map[string]bool)
	var lastTime = time.Now().UTC()

	scanner := bufio.NewScanner(r)
	flushEvents := func(pkg string) {
		// Build a map of parent to subtests
		parentToSubtests := make(map[string][]string)
		isSubtest := make(map[string]bool)
		for _, testName := range testOrder {
			if idx := strings.LastIndex(testName, "/"); idx != -1 {
				parent := testName[:idx]
				parentToSubtests[parent] = append(parentToSubtests[parent], testName)
				isSubtest[testName] = true
			}
		}
		visited := make(map[string]bool)
		for _, testName := range testOrder {
			if visited[testName] {
				continue
			}
			if isSubtest[testName] {
				continue // subtests are handled with their parent
			}
			re, hasRun := runEvents[testName]
			if hasRun {
				if re.Package == "" {
					re.Package = pkg
				}
				lastTime = lastTime.Add(time.Millisecond)
				re.Time = lastTime.Format(time.RFC3339Nano)
				events = append(events, re)
			}
			// Emit subtests (run, status) in order
			if subtests, ok := parentToSubtests[testName]; ok {
				for _, sub := range subtests {
					sre, shasRun := runEvents[sub]
					if shasRun {
						if sre.Package == "" {
							sre.Package = pkg
						}
						lastTime = lastTime.Add(time.Millisecond)
						sre.Time = lastTime.Format(time.RFC3339Nano)
						events = append(events, sre)
					}
					ss, shasStatus := statusEvents[sub]
					if shasStatus {
						if ss.Package == "" {
							ss.Package = pkg
						}
						if outs, ok := outputs[sub]; ok && len(outs) > 0 {
							ss.Output = strings.Join(outs, "\n")
						}
						lastTime = lastTime.Add(time.Millisecond)
						ss.Time = lastTime.Format(time.RFC3339Nano)
						events = append(events, ss)
					}
					visited[sub] = true
				}
			}
			// Emit parent status after subtests
			se, hasStatus := statusEvents[testName]
			if hasStatus {
				// Build the status event to append from the latest values
				output := ""
				if outs, ok := outputs[testName]; ok && len(outs) > 0 {
					if se.Action == "fail" || se.Action == "skip" {
						output = strings.Join(outs, "\n")
					}
					if se.Action == "pass" {
						for _, l := range outs {
							if strings.Contains(l, "coverage:") {
								output = strings.Join(outs, "\n")
								break
							}
						}
					}
				}
				finalStatus := TestEvent{
					Action:  se.Action,
					Test:    se.Test,
					Elapsed: se.Elapsed,
					Package: se.Package,
					Output:  output,
				}
				if finalStatus.Package == "" {
					finalStatus.Package = pkg
				}
				lastTime = lastTime.Add(time.Millisecond)
				finalStatus.Time = lastTime.Format(time.RFC3339Nano)
				events = append(events, finalStatus)
			}
			visited[testName] = true
		}
		testOrder = nil
		runEvents = make(map[string]TestEvent)
		statusEvents = make(map[string]TestEvent)
		outputs = make(map[string][]string)
		seenTest = make(map[string]bool)
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if trimmed == "PASS" || trimmed == "FAIL" || strings.HasPrefix(trimmed, "exit status") {
			continue
		}

		if (strings.HasPrefix(trimmed, "ok  ") || strings.HasPrefix(trimmed, "FAIL")) && len(strings.Fields(trimmed)) >= 2 {
			parts := strings.Fields(trimmed)
			pkg := parts[1]
			flushEvents(pkg)
			currentPkg = pkg
			continue
		}

		if strings.HasPrefix(trimmed, "=== RUN") {
			testName := strings.TrimSpace(strings.TrimPrefix(trimmed, "=== RUN"))
			if !seenTest[testName] {
				testOrder = append(testOrder, testName)
				seenTest[testName] = true
			}
			re := TestEvent{
				Action: "run",
				Test:   testName,
			}
			runEvents[testName] = re
			continue
		}

		if strings.HasPrefix(line, "    --- ") {
			parts := strings.Fields(trimmed)
			if len(parts) < 4 {
				continue
			}
			status := strings.ToLower(parts[1])
			status = strings.TrimSuffix(status, ":")
			status = strings.TrimSpace(status)
			testName := parts[2]
			duration := 0.0
			if len(parts) > 3 && strings.HasPrefix(parts[3], "(") && strings.HasSuffix(parts[3], ")") {
				durStr := strings.Trim(parts[3], "()s")
				if _, err := fmt.Sscanf(durStr, "%f", &duration); err != nil {
					duration = 0.0
				}
			}
			if !seenTest[testName] {
				parent := ""
				if idx := strings.LastIndex(testName, "/"); idx != -1 {
					parent = testName[:idx]
				}
				insertIdx := len(testOrder)
				for i, name := range testOrder {
					if name == parent {
						insertIdx = i + 1
						break
					}
				}
				testOrder = append(testOrder[:insertIdx], append([]string{testName}, testOrder[insertIdx:]...)...)
				seenTest[testName] = true
			}
			se := TestEvent{
				Action:  status,
				Test:    testName,
				Elapsed: duration,
			}
			statusEvents[testName] = se
			continue
		}

		// Indented output line (e.g., error/skip message)
		if strings.HasPrefix(line, "    ") {
			msg := strings.TrimSpace(line)
			attached := false
			// Attach to the most recent subtest in testOrder
			for i := len(testOrder) - 1; i >= 0; i-- {
				if strings.Contains(testOrder[i], "/") {
					outputs[testOrder[i]] = append(outputs[testOrder[i]], msg)
					attached = true
					break
				}
			}
			// If no subtest found, attach to the most recent test (parent)
			if !attached && len(testOrder) > 0 {
				outputs[testOrder[len(testOrder)-1]] = append(outputs[testOrder[len(testOrder)-1]], msg)
			}
			continue
		}

		if strings.HasPrefix(trimmed, "--- ") {
			parts := strings.Fields(trimmed)
			if len(parts) < 4 {
				continue
			}
			status := strings.ToLower(parts[1])
			status = strings.TrimSuffix(status, ":")
			status = strings.TrimSpace(status)
			testName := parts[2]
			duration := 0.0
			if len(parts) > 3 && strings.HasPrefix(parts[3], "(") && strings.HasSuffix(parts[3], ")") {
				durStr := strings.Trim(parts[3], "()s")
				if _, err := fmt.Sscanf(durStr, "%f", &duration); err != nil {
					duration = 0.0
				}
			}
			if !seenTest[testName] {
				testOrder = append(testOrder, testName)
				seenTest[testName] = true
			}
			se := TestEvent{
				Action:  status,
				Test:    testName,
				Elapsed: duration,
			}
			statusEvents[testName] = se
			continue
		}

		if strings.Contains(trimmed, "coverage:") {
			if len(testOrder) > 0 {
				lastTest := testOrder[len(testOrder)-1]
				outputs[lastTest] = append(outputs[lastTest], trimmed)
			}
			continue
		}

		// Non-indented output line (attach to most recent test in testOrder)
		if len(testOrder) > 0 && !strings.HasPrefix(trimmed, "ok  ") && !strings.HasPrefix(trimmed, "FAIL") {
			lastTest := testOrder[len(testOrder)-1]
			outputs[lastTest] = append(outputs[lastTest], trimmed)
		}
	}

	flushEvents(currentPkg)

	return events
}
