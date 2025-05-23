package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Parser parses go test output into structured test results
type Parser struct {
	// testEvents maps package+test to slice of events
	testEvents map[string][]*TestEvent
	// packageEvents maps package to slice of events
	packageEvents map[string][]*TestEvent
	// filePath is a regexp to extract file paths from output
	filePath *regexp.Regexp
	// lineNumber is a regexp to extract line numbers from output
	lineNumber *regexp.Regexp
}

// NewParser creates a new Parser instance
func NewParser() *Parser {
	return &Parser{
		testEvents:    make(map[string][]*TestEvent),
		packageEvents: make(map[string][]*TestEvent),
		// Match file paths like /path/to/file.go or C:\path\to\file.go
		filePath: regexp.MustCompile(`(?:\/[\w.-]+)+\.go|\w:\\(?:[\w.-]+\\)+[\w.-]+\.go`),
		// Match line numbers like :42 or line 42
		lineNumber: regexp.MustCompile(`(?::(\d+))|(?:line (\d+))`),
	}
}

// Parse parses the go test -json output and returns TestPackage results
func (p *Parser) Parse(r io.Reader) ([]*TestPackage, error) {
	scanner := bufio.NewScanner(r)

	// First pass: collect all events
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var event TestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return nil, fmt.Errorf("error parsing JSON: %v", err)
		}

		// Store event by package and test
		if event.Test != "" {
			key := event.Package + ":" + event.Test
			p.testEvents[key] = append(p.testEvents[key], &event)
		} else {
			p.packageEvents[event.Package] = append(p.packageEvents[event.Package], &event)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %v", err)
	}

	// Second pass: process events into test results
	return p.processEvents()
}

// processEvents processes all events into test results
func (p *Parser) processEvents() ([]*TestPackage, error) {
	packages := make(map[string]*TestPackage)

	// Process package events first
	for pkgName, events := range p.packageEvents {
		pkg := &TestPackage{
			Package: pkgName,
			Tests:   make([]*TestResult, 0),
		}
		packages[pkgName] = pkg

		for _, event := range events {
			switch event.Action {
			case "fail":
				pkg.Passed = false
				pkg.Duration = time.Duration(event.Elapsed * float64(time.Second))
			case "pass":
				pkg.Passed = true
				pkg.Duration = time.Duration(event.Elapsed * float64(time.Second))
			case "skip":
				// Package was skipped due to build failure
				pkg.BuildFailed = true
			case "output":
				// Check for build failures
				if strings.Contains(event.Output, "# "+pkgName) ||
					strings.Contains(event.Output, "syntax error") ||
					strings.Contains(event.Output, "undefined:") {
					pkg.BuildFailed = true
					pkg.BuildError += event.Output
				}
			}
		}
	}

	// Process test events
	for key, events := range p.testEvents {
		parts := strings.SplitN(key, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid event key: %s", key)
		}

		pkgName, testName := parts[0], parts[1]
		pkg, exists := packages[pkgName]
		if !exists {
			pkg = &TestPackage{
				Package: pkgName,
				Tests:   make([]*TestResult, 0),
			}
			packages[pkgName] = pkg
		}

		result := &TestResult{
			Name:    testName,
			Package: pkgName,
			Test:    testName,
		}

		var output strings.Builder

		// Sort events by time to ensure correct processing order
		sort.Slice(events, func(i, j int) bool {
			timeI, _ := time.Parse(time.RFC3339Nano, events[i].Time)
			timeJ, _ := time.Parse(time.RFC3339Nano, events[j].Time)
			return timeI.Before(timeJ)
		})

		for _, event := range events {
			switch event.Action {
			case "run":
				result.Status = StatusRunning
			case "pass":
				result.Status = StatusPassed
				result.Duration = time.Duration(event.Elapsed * float64(time.Second))
				pkg.PassedCount++
			case "fail":
				result.Status = StatusFailed
				result.Duration = time.Duration(event.Elapsed * float64(time.Second))
				pkg.FailedCount++
			case "skip":
				result.Status = StatusSkipped
				pkg.SkippedCount++
			case "output":
				output.WriteString(event.Output)

				// Check for errors in output
				if strings.Contains(event.Output, "FAIL") ||
					strings.Contains(event.Output, "panic:") ||
					strings.Contains(event.Output, "Expected") {
					if result.Error == nil {
						result.Error = &TestError{
							Message: strings.TrimSpace(event.Output),
							Type:    p.determineErrorType(event.Output),
						}
					} else {
						result.Error.Message += "\n" + strings.TrimSpace(event.Output)
					}

					// Try to extract location information
					if loc := p.extractSourceLocation(event.Output); loc != nil && result.Error.Location == nil {
						result.Error.Location = loc
					}
				}
			}
		}

		result.Output = output.String()

		// Only add the test if it was actually run
		if result.Status != "" {
			// Check if this is a subtest
			if strings.Contains(testName, "/") {
				parentName := testName[:strings.Index(testName, "/")]
				result.Parent = parentName

				// Find parent test and add this as a subtest
				var found bool
				for i, t := range pkg.Tests {
					if t.Name == parentName {
						pkg.Tests[i].Subtests = append(pkg.Tests[i].Subtests, result)
						found = true
						break
					}
				}

				if !found {
					// Parent test hasn't been processed yet, create a placeholder
					parent := &TestResult{
						Name:     parentName,
						Package:  pkgName,
						Test:     parentName,
						Status:   StatusRunning,
						Subtests: []*TestResult{result},
					}
					pkg.Tests = append(pkg.Tests, parent)
				}
			} else {
				// This is a top-level test
				pkg.Tests = append(pkg.Tests, result)
			}
		}
	}

	// Convert map to slice
	results := make([]*TestPackage, 0, len(packages))
	for _, pkg := range packages {
		results = append(results, pkg)
	}

	// Sort packages by name for consistent output
	sort.Slice(results, func(i, j int) bool {
		return results[i].Package < results[j].Package
	})

	return results, nil
}

// determineErrorType determines the type of error from the test output
func (p *Parser) determineErrorType(output string) string {
	if strings.Contains(output, "panic:") {
		return "panic"
	}
	if strings.Contains(output, "Expected") {
		return "assertion"
	}
	return "error"
}

// extractSourceLocation extracts file and line information from test output
func (p *Parser) extractSourceLocation(output string) *SourceLocation {
	// Extract file path
	fileMatches := p.filePath.FindStringSubmatch(output)
	if len(fileMatches) == 0 {
		return nil
	}
	file := fileMatches[0]

	// Extract line number
	lineMatches := p.lineNumber.FindStringSubmatch(output)
	if len(lineMatches) == 0 {
		return nil
	}

	// Line number could be in group 1 or 2
	var lineStr string
	if lineMatches[1] != "" {
		lineStr = lineMatches[1]
	} else {
		lineStr = lineMatches[2]
	}

	line, err := strconv.Atoi(lineStr)
	if err != nil {
		return nil
	}

	return &SourceLocation{
		File: file,
		Line: line,
	}
}
