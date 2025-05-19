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

	// Support both 'file.go:123: message' and 'file.go:123:message'
	fileLineRe := regexp.MustCompile(`(?m)^\s*([\w./-]+):(\d+):?\s*(.*)$`)
	var lastOutput string
	var hasError bool

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
		} else if ev.Action == "fail" {
			hasError = true
		}
	}

	// Always return a context if there is any output and a failing test
	if lastOutput != "" && hasError {
		return &ErrorContext{
			Message: lastOutput,
		}
	}

	// If we have a failing test but no error message, return a context with just the failure status
	if hasError {
		return &ErrorContext{
			Message: "test failed",
		}
	}

	// If there is any output at all, return a context with the last output as message
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
	var currentTest string
	var currentPkg string
	var lastTime time.Time
	seenLines := make(map[string]bool)
	var buffered []TestEvent

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || seenLines[line] {
			continue
		}
		seenLines[line] = true

		// Skip summary lines
		if line == "PASS" || line == "FAIL" || strings.HasPrefix(line, "exit status") {
			continue
		}

		// Parse package status lines first to set the package name, but do not emit events for them
		if (strings.HasPrefix(line, "ok  ") || strings.HasPrefix(line, "FAIL")) && len(strings.Fields(line)) >= 2 {
			parts := strings.Fields(line)
			currentPkg = parts[1]
			// Flush any buffered events with the now-known package
			for i := range buffered {
				buffered[i].Package = currentPkg
				events = append(events, buffered[i])
			}
			buffered = nil
			continue
		}

		// Parse RUN marker
		if strings.HasPrefix(line, "=== RUN") {
			testName := strings.TrimSpace(strings.TrimPrefix(line, "=== RUN"))
			lastTime = lastTime.Add(time.Millisecond)
			event := TestEvent{
				Time:    lastTime.Format(time.RFC3339),
				Action:  "run",
				Package: currentPkg,
				Test:    testName,
			}
			if currentPkg == "" {
				buffered = append(buffered, event)
			} else {
				events = append(events, event)
			}
			currentTest = testName
			continue
		}

		// Parse PASS/FAIL/SKIP markers
		if strings.HasPrefix(line, "--- ") {
			parts := strings.Fields(line)
			if len(parts) < 4 {
				continue
			}

			status := strings.TrimPrefix(parts[1], "")
			testName := parts[2]
			duration := 0.0

			// Parse duration (e.g., "(0.00s)")
			if len(parts) > 3 && strings.HasPrefix(parts[3], "(") && strings.HasSuffix(parts[3], ")") {
				durStr := strings.Trim(parts[3], "()")
				durStr = strings.TrimSuffix(durStr, "s")
				fmt.Sscanf(durStr, "%f", &duration)
			}

			action := strings.ToLower(status)
			lastTime = lastTime.Add(time.Millisecond)
			event := TestEvent{
				Time:    lastTime.Format(time.RFC3339),
				Action:  action,
				Package: currentPkg,
				Test:    testName,
				Elapsed: duration,
			}
			if currentPkg == "" {
				buffered = append(buffered, event)
			} else {
				events = append(events, event)
			}
			continue
		}

		// Parse coverage information and other output
		if strings.Contains(line, "coverage:") {
			var coverage float64
			if _, err := fmt.Sscanf(line, "coverage: %f%% of statements", &coverage); err == nil {
				lastTime = lastTime.Add(time.Millisecond)
				event := TestEvent{
					Time:    lastTime.Format(time.RFC3339),
					Action:  "output",
					Package: currentPkg,
					Test:    currentTest,
					Output:  line,
				}
				if currentPkg == "" {
					buffered = append(buffered, event)
				} else {
					events = append(events, event)
				}
			}
			continue
		}

		// Add all non-empty output lines that are not already seen
		if line != "" && !strings.HasPrefix(line, "=== RUN") && !strings.HasPrefix(line, "--- ") && !strings.HasPrefix(line, "ok  ") && !strings.HasPrefix(line, "FAIL") {
			lastTime = lastTime.Add(time.Millisecond)
			event := TestEvent{
				Time:    lastTime.Format(time.RFC3339),
				Action:  "output",
				Package: currentPkg,
				Test:    currentTest,
				Output:  line,
			}
			if currentPkg == "" {
				buffered = append(buffered, event)
			} else {
				events = append(events, event)
			}
		}
	}

	// If any buffered events remain, assign them to the last known package (if any)
	for i := range buffered {
		buffered[i].Package = currentPkg
		events = append(events, buffered[i])
	}

	// Filter out duplicate events while preserving order
	seen := make(map[string]bool)
	filtered := make([]TestEvent, 0, len(events))
	for _, ev := range events {
		// Skip empty output events
		if ev.Action == "output" && ev.Output == "" {
			continue
		}

		// Create a unique key for the event
		key := fmt.Sprintf("%s:%s:%s:%s:%s", ev.Time, ev.Action, ev.Package, ev.Test, ev.Output)
		if !seen[key] {
			seen[key] = true
			filtered = append(filtered, ev)
		}
	}

	return filtered
}
