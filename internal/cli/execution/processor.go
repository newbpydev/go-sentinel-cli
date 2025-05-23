package execution

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/newbpydev/go-sentinel/internal/cli/core"
	"golang.org/x/term"
)

// TestProcessor processes JSON output from Go test command
type TestProcessor struct {
	suites     map[string]*core.TestSuite
	statistics *core.TestRunStats
	startTime  time.Time

	// Phase tracking
	setupStartTime  time.Time
	firstTestTime   time.Time
	lastTestTime    time.Time
	teardownEndTime time.Time
}

// TestEvent represents a single JSON event from Go test output
type TestEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test,omitempty"`
	Elapsed float64   `json:"Elapsed,omitempty"`
	Output  string    `json:"Output,omitempty"`
}

// NewTestProcessor creates a new TestProcessor
func NewTestProcessor() *TestProcessor {
	return &TestProcessor{
		suites: make(map[string]*core.TestSuite),
		statistics: &core.TestRunStats{
			StartTime: time.Now(),
		},
		startTime:       time.Now(),
		setupStartTime:  time.Now(),
		firstTestTime:   time.Time{},
		lastTestTime:    time.Time{},
		teardownEndTime: time.Time{},
	}
}

// ProcessJSONOutput processes the JSON output from Go test and updates the processor state
func (p *TestProcessor) ProcessJSONOutput(output string) error {
	// Reset the processor state
	p.Reset()

	// Split the output into lines
	lines := strings.Split(output, "\n")

	// Process each line as a JSON object
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var event TestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}

		// Process the event based on its action
		switch event.Action {
		case "run":
			// A test is about to run
			p.onTestRun(event)
		case "pass":
			// A test has passed
			p.onTestPass(event)
		case "fail":
			// A test has failed
			p.onTestFail(event)
		case "skip":
			// A test was skipped
			p.onTestSkip(event)
		case "output":
			// Test output - accumulate for later use
			p.onTestOutput(event)
		}
	}

	// Finalize the state
	p.finalize()

	return nil
}

// Reset resets the processor state for a new test run
func (p *TestProcessor) Reset() {
	now := time.Now()
	p.statistics = &core.TestRunStats{
		StartTime: now,
	}
	p.suites = make(map[string]*core.TestSuite)
	p.setupStartTime = now
	p.firstTestTime = time.Time{}
	p.lastTestTime = time.Time{}
	p.teardownEndTime = time.Time{}
}

// onTestRun handles a test run event
func (p *TestProcessor) onTestRun(event TestEvent) {
	// Track the first test start time (end of setup phase)
	if p.firstTestTime.IsZero() {
		p.firstTestTime = time.Now()
	}

	// Find or create the test suite
	suitePath := event.Package
	suite, ok := p.suites[suitePath]
	if !ok {
		suite = &core.TestSuite{
			Name:     event.Package,
			FilePath: suitePath,
			Status:   core.StatusRunning,
		}
		p.suites[suitePath] = suite
	}

	// Count the test
	suite.TestCount++
}

// onTestPass handles a test pass event
func (p *TestProcessor) onTestPass(event TestEvent) {
	// Update last test completion time
	p.lastTestTime = time.Now()

	// Find the test suite
	suite, ok := p.suites[event.Package]
	if !ok {
		return
	}

	// Update suite counts
	suite.PassedCount++
	suite.Duration += time.Duration(event.Elapsed * float64(time.Second))

	// Update overall statistics
	p.statistics.PassedTests++
	p.statistics.TotalTests++

	// Update suite status if all tests are done
	p.updateSuiteStatus(suite)
}

// onTestFail handles a test fail event
func (p *TestProcessor) onTestFail(event TestEvent) {
	// Update last test completion time
	p.lastTestTime = time.Now()

	// Find the test suite
	suite, ok := p.suites[event.Package]
	if !ok {
		return
	}

	// Update suite counts
	suite.FailedCount++
	suite.Duration += time.Duration(event.Elapsed * float64(time.Second))
	suite.Status = core.StatusFailed // Any failure marks the suite as failed

	// Update overall statistics
	p.statistics.FailedTests++
	p.statistics.TotalTests++

	// Set error on suite
	if suite.Error == nil {
		suite.Error = fmt.Errorf("test failures in package %s", event.Package)
	}
}

// onTestSkip handles a test skip event
func (p *TestProcessor) onTestSkip(event TestEvent) {
	// Update last test completion time
	p.lastTestTime = time.Now()

	// Find the test suite
	suite, ok := p.suites[event.Package]
	if !ok {
		return
	}

	// Update suite counts
	suite.SkippedCount++
	suite.Duration += time.Duration(event.Elapsed * float64(time.Second))

	// Update overall statistics
	p.statistics.SkippedTests++
	p.statistics.TotalTests++

	// Update suite status if all tests are done
	p.updateSuiteStatus(suite)
}

// onTestOutput handles test output events
func (p *TestProcessor) onTestOutput(event TestEvent) {
	// For now, we just accumulate output but don't process it
	// In the future, this could be used for detailed test output processing
}

// updateSuiteStatus updates the status of a test suite based on its results
func (p *TestProcessor) updateSuiteStatus(suite *core.TestSuite) {
	if suite.FailedCount > 0 {
		suite.Status = core.StatusFailed
	} else if suite.PassedCount == suite.TestCount {
		suite.Status = core.StatusPassed
	} else if suite.SkippedCount == suite.TestCount {
		suite.Status = core.StatusSkipped
	} else {
		suite.Status = core.StatusRunning
	}
}

// finalize finalizes the processor state after all events are processed
func (p *TestProcessor) finalize() {
	now := time.Now()
	p.teardownEndTime = now

	// Calculate total duration
	p.statistics.EndTime = now
	p.statistics.Duration = now.Sub(p.statistics.StartTime)

	// Count total suites
	p.statistics.TotalSuites = len(p.suites)

	// Update final suite statuses
	for _, suite := range p.suites {
		p.updateSuiteStatus(suite)
	}
}

// GetSuites returns the processed test suites
func (p *TestProcessor) GetSuites() map[string]*core.TestSuite {
	return p.suites
}

// GetStats returns the test run statistics
func (p *TestProcessor) GetStats() *core.TestRunStats {
	return p.statistics
}

// GetTerminalWidth returns the current terminal width or default
func GetTerminalWidth() int {
	if fd := int(os.Stdout.Fd()); term.IsTerminal(fd) {
		if width, _, err := term.GetSize(fd); err == nil && width > 0 {
			return width
		}
	}
	return 80 // Default fallback
}

// Helper functions for test analysis

// IsGoTestFile returns true if the file is a Go test file
func IsGoTestFile(path string) bool {
	return strings.HasSuffix(path, "_test.go")
}

// IsGoFile returns true if the file is a Go source file
func IsGoFile(path string) bool {
	return strings.HasSuffix(path, ".go")
}

// GetPackageFromPath extracts the package path from a file path
func GetPackageFromPath(filePath string) string {
	dir := filepath.Dir(filePath)
	return strings.ReplaceAll(dir, "\\", "/")
}

// ExtractTestFunctionName extracts the test function name from output
func ExtractTestFunctionName(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "FAIL:") || strings.Contains(line, "PASS:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				return parts[1]
			}
		}
	}
	return ""
}

// FileLocation represents a file location with line number
type FileLocation struct {
	File string
	Line int
}

// ExtractFileLocationFromLine extracts file location from test output line
func ExtractFileLocationFromLine(line string) *FileLocation {
	// Look for patterns like "file.go:123" in the line
	parts := strings.Fields(line)
	for _, part := range parts {
		if colonIndex := strings.LastIndex(part, ":"); colonIndex != -1 {
			filePart := part[:colonIndex]
			linePart := part[colonIndex+1:]

			// Check if it looks like a file path
			if strings.HasSuffix(filePart, ".go") {
				if lineNum, err := strconv.Atoi(linePart); err == nil {
					return &FileLocation{
						File: filePart,
						Line: lineNum,
					}
				}
			}
		}
	}
	return nil
}

// InferSourceFileFromTest tries to infer the source file from test package and name
func InferSourceFileFromTest(testPackage, testName string) string {
	// Remove "Test" prefix if present
	sourceName := testName
	if strings.HasPrefix(sourceName, "Test") {
		sourceName = sourceName[4:]
	}

	// Convert CamelCase to snake_case and add .go extension
	var result strings.Builder
	for i, r := range sourceName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}

	filename := strings.ToLower(result.String()) + ".go"
	return filepath.Join(testPackage, filename)
}
