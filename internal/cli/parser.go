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

// TestEvent represents a single event from go test -json output
type TestEvent struct {
	Time    time.Time // Time when the event occurred
	Action  string    // Action is the action type ("run", "pause", "pass", "fail", etc.)
	Package string    // Package is the package being tested
	Test    string    // Test is the test being run (may be empty for package events)
	Output  string    // Output is the output of the test or package (may be empty)
	Elapsed float64   // Elapsed is the time elapsed for the test or package (in seconds)
}

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
			return events[i].Time.Before(events[j].Time)
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

				// If parent not found yet, create placeholder and add later
				if !found {
					// This is a top-level test for now, will be reorganized
					pkg.Tests = append(pkg.Tests, result)
				}
			} else {
				// This is a top-level test
				pkg.Tests = append(pkg.Tests, result)
			}
		}
	}

	// Reorganize subtests if needed
	for _, pkg := range packages {
		// Create a map of test names to their indices for quick lookup
		testIndices := make(map[string]int)
		for i, test := range pkg.Tests {
			testIndices[test.Name] = i
		}

		// Go through tests again, moving subtests to their parents
		var i int
		for i < len(pkg.Tests) {
			test := pkg.Tests[i]
			if test.Parent != "" && testIndices[test.Parent] != i { // It's a subtest and not parent
				// Find parent and move this test as a subtest
				parentIdx, exists := testIndices[test.Parent]
				if exists {
					// Add as subtest
					pkg.Tests[parentIdx].Subtests = append(pkg.Tests[parentIdx].Subtests, test)
					// Remove from top level
					pkg.Tests = append(pkg.Tests[:i], pkg.Tests[i+1:]...)
					// Update indices
					for j := i; j < len(pkg.Tests); j++ {
						testIndices[pkg.Tests[j].Name] = j
					}
					continue // Don't increment i since we removed a test
				}
			}
			i++
		}

		// Count total tests
		pkg.TestCount = pkg.PassedCount + pkg.FailedCount + pkg.SkippedCount
	}

	// Convert map to slice, preserving package order from the test
	var result []*TestPackage
	var packageNames []string
	for name := range packages {
		packageNames = append(packageNames, name)
	}

	// Sort by package name to ensure consistent ordering
	sort.Strings(packageNames)

	for _, name := range packageNames {
		result = append(result, packages[name])
	}

	return result, nil
}

// determineErrorType determines the type of error from the error message
func (p *Parser) determineErrorType(output string) string {
	if strings.Contains(output, "panic:") {
		return "Panic"
	}
	if strings.Contains(output, "timed out") {
		return "Timeout"
	}
	if strings.Contains(output, "assertion") || strings.Contains(output, "expected") || strings.Contains(output, "Expected") {
		return "AssertionError"
	}
	return "Error"
}

// extractSourceLocation extracts file and line information from error output
func (p *Parser) extractSourceLocation(output string) *SourceLocation {
	// Special case for panic stack traces
	if strings.Contains(output, "panic:") || strings.Contains(output, "PANIC:") {
		// Try to find a stack trace line
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			// Look for specific stack trace patterns
			// Pattern 1: file.go:42 +0x39
			if strings.Contains(line, ".go:") {
				re := regexp.MustCompile(`\s*([^:\s]+\.go):(\d+)`)
				matches := re.FindStringSubmatch(line)
				if len(matches) >= 3 {
					file := matches[1]
					lineNum, err := strconv.Atoi(matches[2])
					if err == nil {
						return &SourceLocation{
							File: file,
							Line: lineNum,
						}
					}
				}
			}

			// Pattern 2: .TestFunc in file.go:42
			if strings.Contains(line, ".Test") && strings.Contains(line, ".go:") {
				parts := strings.Split(line, ".go:")
				if len(parts) >= 2 {
					// Extract the file path
					fileParts := strings.Split(parts[0], " ")
					file := fileParts[len(fileParts)-1] + ".go"

					// Extract line number
					lineStr := ""
					for i, c := range parts[1] {
						if c >= '0' && c <= '9' {
							lineStr += string(c)
						} else if lineStr != "" {
							break
						}

						// Avoid very long lines
						if i > 10 {
							break
						}
					}

					if lineStr != "" {
						lineNum, err := strconv.Atoi(lineStr)
						if err == nil {
							return &SourceLocation{
								File: file,
								Line: lineNum,
							}
						}
					}
				}
			}
		}
	}

	// Standard extraction for other error types
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

	var lineStr string
	if lineMatches[1] != "" {
		lineStr = lineMatches[1]
	} else if len(lineMatches) > 2 && lineMatches[2] != "" {
		lineStr = lineMatches[2]
	} else {
		return nil
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
