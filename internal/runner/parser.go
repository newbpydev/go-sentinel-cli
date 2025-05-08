package runner

import (
	"encoding/json"
	"fmt"
	"io"
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

// ExtractErrorContext scans output events for file:line info and error message.
func ExtractErrorContext(events []TestEvent) *ErrorContext {
	for _, ev := range events {
		if ev.Action == "output" {
			// Example output: "main_test.go:42: expected true, got false"
			msg := ev.Output
			var file string
			var line int
			// Try to match file:line: prefix
			fmtScan, err := fmt.Sscanf(msg, "%s:%d:", &file, &line)
			if err == nil && fmtScan == 2 {
				return &ErrorContext{
					Message:      msg,
					FileLocation: &FileLocation{File: file, Line: line},
				}
			}
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
