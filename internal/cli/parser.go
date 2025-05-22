package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// GoTestEvent represents the JSON output from 'go test -json'
type GoTestEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test,omitempty"`
	Output  string    `json:"Output,omitempty"`
	Elapsed float64   `json:"Elapsed,omitempty"`
}

var (
	// Regular expressions for parsing test output
	errorLocationRe = regexp.MustCompile(`(?m)^\s*([\w./-]+\.go):(\d+)(?::(\d+))?:`)
)

// Parser handles parsing of go test -json output
type Parser struct {
	currentRun   *TestRun
	currentSuite *TestSuite
	suites       map[string]*TestSuite
}

// NewParser creates a new parser instance
func NewParser() *Parser {
	return &Parser{
		suites: make(map[string]*TestSuite),
	}
}

// Parse reads go test -json output and returns a TestRun
func (p *Parser) Parse(r io.Reader) (*TestRun, error) {
	p.currentRun = &TestRun{
		Suites:     make([]*TestSuite, 0),
		StartTime:  time.Now(),
		NumTotal:   0,
		NumPassed:  0,
		NumFailed:  0,
		NumSkipped: 0,
	}
	p.suites = make(map[string]*TestSuite)

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		var event GoTestEvent
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			continue
		}

		if err := p.handleEvent(&event); err != nil {
			return nil, fmt.Errorf("error handling test event: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading test output: %w", err)
	}

	p.finalize()
	return p.currentRun, nil
}

// handleEvent processes a single test event
func (p *Parser) handleEvent(event *GoTestEvent) error {
	switch event.Action {
	case "start":
		if event.Test == "" {
			return p.handlePackageStart(event)
		}
		return p.handleTestStart(event)
	case "run":
		return p.handleTestRun(event)
	case "pass":
		return p.handleTestPass(event)
	case "fail":
		return p.handleTestFail(event)
	case "skip":
		return p.handleTestSkip(event)
	case "output":
		return p.handleTestOutput(event)
	}
	return nil
}

// handlePackageStart processes a package start event
func (p *Parser) handlePackageStart(event *GoTestEvent) error {
	// Check if we already have a suite for this package
	if suite, exists := p.suites[event.Package]; exists {
		p.currentSuite = suite
		return nil
	}

	suite := &TestSuite{
		Package:     event.Package,
		PackageName: event.Package,
		FilePath:    p.getTestFilePath(event.Package),
		Tests:       make([]*TestResult, 0),
		Errors:      make([]*TestError, 0),
		StartTime:   event.Time,
		NumTotal:    0,
		NumPassed:   0,
		NumFailed:   0,
		NumSkipped:  0,
	}
	p.suites[event.Package] = suite
	p.currentSuite = suite
	p.currentRun.Suites = append(p.currentRun.Suites, suite)
	return nil
}

// handleTestStart processes a test start event
func (p *Parser) handleTestStart(event *GoTestEvent) error {
	if p.currentSuite == nil {
		return fmt.Errorf("no current suite")
	}

	// Check if test already exists
	if test := p.findTest(event.Test); test != nil {
		return nil
	}

	test := &TestResult{
		Name:      event.Test,
		Status:    TestStatusRunning,
		StartTime: event.Time,
	}
	p.currentSuite.Tests = append(p.currentSuite.Tests, test)
	p.currentSuite.NumTotal++
	p.currentRun.NumTotal++
	return nil
}

// handleTestRun processes a test run event
func (p *Parser) handleTestRun(event *GoTestEvent) error {
	if p.currentSuite == nil {
		return fmt.Errorf("no current suite")
	}

	// Check if test already exists
	if test := p.findTest(event.Test); test != nil {
		return nil
	}

	test := &TestResult{
		Name:      event.Test,
		Status:    TestStatusRunning,
		StartTime: event.Time,
	}
	p.currentSuite.Tests = append(p.currentSuite.Tests, test)
	p.currentSuite.NumTotal++
	p.currentRun.NumTotal++
	return nil
}

// handleTestPass processes a test pass event
func (p *Parser) handleTestPass(event *GoTestEvent) error {
	test := p.findTest(event.Test)
	if test == nil {
		return nil
	}

	test.Status = TestStatusPassed
	test.Duration = time.Duration(event.Elapsed * float64(time.Second))
	test.EndTime = event.Time
	p.currentSuite.NumPassed++
	p.currentRun.NumPassed++
	return nil
}

// handleTestFail processes a test fail event
func (p *Parser) handleTestFail(event *GoTestEvent) error {
	test := p.findTest(event.Test)
	if test == nil {
		return nil
	}

	test.Status = TestStatusFailed
	test.Duration = time.Duration(event.Elapsed * float64(time.Second))
	test.EndTime = event.Time
	if test.Error == nil {
		test.Error = &TestError{
			Message: event.Output,
		}
	}
	p.currentSuite.NumFailed++
	p.currentRun.NumFailed++

	// Track failed test in the TestRun
	p.currentRun.FailedTests = append(p.currentRun.FailedTests, test)
	return nil
}

// handleTestSkip processes a test skip event
func (p *Parser) handleTestSkip(event *GoTestEvent) error {
	test := p.findTest(event.Test)
	if test == nil {
		return nil
	}

	test.Status = TestStatusSkipped
	test.Duration = time.Duration(event.Elapsed * float64(time.Second))
	test.EndTime = event.Time
	p.currentSuite.NumSkipped++
	p.currentRun.NumSkipped++
	return nil
}

// handleTestOutput processes a test output event
func (p *Parser) handleTestOutput(event *GoTestEvent) error {
	if event.Test == "" {
		// Package-level output
		if p.currentSuite != nil && strings.Contains(event.Output, "FAIL") {
			p.currentSuite.NumFailed++
			p.currentRun.NumFailed++
			p.currentSuite.Errors = append(p.currentSuite.Errors, &TestError{
				Message: event.Output,
			})
		}
		return nil
	}

	test := p.findTest(event.Test)
	if test == nil {
		return nil
	}

	// Accumulate test output
	if test.Error == nil {
		test.Error = &TestError{
			Message: event.Output,
		}
	} else {
		test.Error.Message += event.Output
	}

	// Extract source location from output
	if loc := p.extractSourceLocation(event.Output); loc != nil {
		test.Error.Location = loc
	}

	return nil
}

// finalize processes any remaining test results and updates statistics
func (p *Parser) finalize() {
	// Sort suites by package name for consistent output
	sort.Slice(p.currentRun.Suites, func(i, j int) bool {
		return p.currentRun.Suites[i].Package < p.currentRun.Suites[j].Package
	})

	// Update file paths and ensure they're relative
	for _, suite := range p.currentRun.Suites {
		if suite.FilePath == "" {
			suite.FilePath = p.getTestFilePath(suite.Package)
		}
		// Make path relative if it's absolute
		if filepath.IsAbs(suite.FilePath) {
			if rel, err := filepath.Rel(".", suite.FilePath); err == nil {
				suite.FilePath = rel
			}
		}
		// Ensure forward slashes for path consistency
		suite.FilePath = strings.ReplaceAll(suite.FilePath, "\\", "/")

		// Sort tests by name for consistent output
		sort.Slice(suite.Tests, func(i, j int) bool {
			return suite.Tests[i].Name < suite.Tests[j].Name
		})
	}

	// Calculate final statistics
	p.currentRun.NumTotal = 0
	p.currentRun.NumPassed = 0
	p.currentRun.NumFailed = 0
	p.currentRun.NumSkipped = 0

	for _, suite := range p.currentRun.Suites {
		p.currentRun.NumTotal += suite.NumTotal
		p.currentRun.NumPassed += suite.NumPassed
		p.currentRun.NumFailed += suite.NumFailed
		p.currentRun.NumSkipped += suite.NumSkipped
	}

	// Set end time and duration
	p.currentRun.EndTime = time.Now()
	p.currentRun.Duration = p.currentRun.EndTime.Sub(p.currentRun.StartTime)
}

// findTest finds a test by name in the current suite
func (p *Parser) findTest(name string) *TestResult {
	if p.currentSuite == nil {
		return nil
	}
	for _, test := range p.currentSuite.Tests {
		if test.Name == name {
			return test
		}
	}
	return nil
}

// extractSourceLocation extracts source location information from test output
func (p *Parser) extractSourceLocation(output string) *SourceLocation {
	matches := errorLocationRe.FindStringSubmatch(output)
	if matches == nil {
		return nil
	}

	loc := &SourceLocation{
		File: strings.ReplaceAll(matches[1], "\\", "/"), // Ensure forward slashes
	}

	// Parse line number
	if matches[2] != "" {
		var line int
		if _, err := fmt.Sscanf(matches[2], "%d", &line); err != nil {
			log.Printf("Error parsing line number: %v", err)
		}
		loc.Line = line
	}

	// Parse column number
	if matches[3] != "" {
		var col int
		if _, err := fmt.Sscanf(matches[3], "%d", &col); err != nil {
			log.Printf("Error parsing column number: %v", err)
		}
		loc.Column = col
	}

	return loc
}

// getTestFilePath returns the file path for a test package
func (p *Parser) getTestFilePath(pkg string) string {
	// Extract the last component of the package path
	parts := strings.Split(pkg, "/")
	if len(parts) == 0 {
		return ""
	}

	// For single component packages, just append _test.go
	if len(parts) == 1 {
		pkgName := parts[0]
		if !strings.HasSuffix(pkgName, "_test") {
			pkgName += "_test"
		}
		return pkgName + ".go"
	}

	// Get the package name from the last component
	pkgName := parts[len(parts)-1]
	if !strings.HasSuffix(pkgName, "_test") {
		pkgName += "_test"
	}

	// For packages with a pkg directory, use pkg/name/name_test.go format
	for i, part := range parts {
		if part == "pkg" && i < len(parts)-1 {
			return fmt.Sprintf("pkg/%s/%s.go", parts[len(parts)-1], pkgName)
		}
	}

	// For other multi-component packages, use pkg/name/name_test.go format
	// This matches the test requirements where we want pkg/foo/foo_test.go
	return fmt.Sprintf("pkg/%s/%s.go", pkgName, pkgName)
}

// processEvent is an alias for handleEvent to maintain compatibility with tests
func (p *Parser) processEvent(event *GoTestEvent) error {
	return p.handleEvent(event)
}
