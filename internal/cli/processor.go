package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

// NewTestProcessor creates a new TestProcessor
func NewTestProcessor(writer io.Writer, formatter *ColorFormatter, icons *IconProvider, width int) *TestProcessor {
	return &TestProcessor{
		writer:    writer,
		formatter: formatter,
		icons:     icons,
		width:     getTerminalWidthForProcessor(),
		suites:    make(map[string]*TestSuite),
		statistics: &TestRunStats{
			StartTime: time.Now(),
			Phases:    make(map[string]time.Duration),
		},
		startTime: time.Now(),
		// Phase tracking timestamps
		setupStartTime:  time.Now(),
		firstTestTime:   time.Time{}, // Will be set when first test runs
		lastTestTime:    time.Time{}, // Will be set when last test completes
		teardownEndTime: time.Time{}, // Will be set at finalization
	}
}

// getTerminalWidthForProcessor returns the current terminal width or default
func getTerminalWidthForProcessor() int {
	if fd := int(os.Stdout.Fd()); term.IsTerminal(fd) {
		if width, _, err := term.GetSize(fd); err == nil && width > 0 {
			return width
		}
	}
	return 80 // Default fallback
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
			// Test output - add to current test
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
	p.statistics = &TestRunStats{
		StartTime: now,
		Phases:    make(map[string]time.Duration),
	}
	p.suites = make(map[string]*TestSuite)
	p.setupStartTime = now
	p.firstTestTime = time.Time{}   // Will be set when first test runs
	p.lastTestTime = time.Time{}    // Will be set when last test completes
	p.teardownEndTime = time.Time{} // Will be set at finalization
}

// onTestRun handles a test run event
func (p *TestProcessor) onTestRun(event TestEvent) {
	// Track the first test start time (end of setup phase)
	if p.firstTestTime.IsZero() {
		p.firstTestTime = time.Now()
	}

	// Create a new test result
	result := &TestResult{
		Name:     event.Test,
		Package:  event.Package,
		Status:   StatusRunning,
		Duration: 0,
	}

	// Find or create the test suite first
	suitePath := event.Package
	suite, ok := p.suites[suitePath]
	if !ok {
		suite = &TestSuite{
			FilePath: suitePath,
		}
		p.suites[suitePath] = suite
	}

	// Check if this is a subtest
	if strings.Contains(event.Test, "/") {
		// Extract parent test name
		parts := strings.SplitN(event.Test, "/", 2)
		result.Parent = parts[0]

		// Add to parent test if it exists
		for _, test := range suite.Tests {
			if test.Name == result.Parent {
				test.Subtests = append(test.Subtests, result)
				// FIXED: Count subtests in suite.TestCount too
				suite.TestCount++
				return
			}
		}
	}

	// Add the test to the suite
	suite.Tests = append(suite.Tests, result)
	suite.TestCount++
}

// onTestPass handles a test pass event
func (p *TestProcessor) onTestPass(event TestEvent) {
	// Update last test completion time
	p.lastTestTime = time.Now()

	// Find the test
	for _, suite := range p.suites {
		for _, test := range suite.Tests {
			if test.Name == event.Test {
				test.Status = StatusPassed
				test.Duration = time.Duration(event.Elapsed * float64(time.Second))
				suite.PassedCount++
				p.statistics.PassedTests++
				p.statistics.TotalTests++
				return
			}

			// Check subtests
			for _, subtest := range test.Subtests {
				if subtest.Name == event.Test {
					subtest.Status = StatusPassed
					subtest.Duration = time.Duration(event.Elapsed * float64(time.Second))
					suite.PassedCount++
					p.statistics.PassedTests++
					p.statistics.TotalTests++
					return
				}
			}
		}
	}
}

// onTestFail handles a test fail event
func (p *TestProcessor) onTestFail(event TestEvent) {
	// Update last test completion time
	p.lastTestTime = time.Now()

	// Find the test
	for _, suite := range p.suites {
		for _, test := range suite.Tests {
			if test.Name == event.Test {
				test.Status = StatusFailed
				test.Duration = time.Duration(event.Elapsed * float64(time.Second))

				// Create error details from accumulated output
				test.Error = p.createTestError(test, event)

				suite.FailedCount++
				p.statistics.FailedTests++
				p.statistics.TotalTests++
				return
			}

			// Check subtests
			for _, subtest := range test.Subtests {
				if subtest.Name == event.Test {
					subtest.Status = StatusFailed
					subtest.Duration = time.Duration(event.Elapsed * float64(time.Second))

					// Create error details from accumulated output
					subtest.Error = p.createTestError(subtest, event)

					suite.FailedCount++
					p.statistics.FailedTests++
					p.statistics.TotalTests++
					return
				}
			}
		}
	}
}

// onTestSkip handles a test skip event
func (p *TestProcessor) onTestSkip(event TestEvent) {
	// Update last test completion time
	p.lastTestTime = time.Now()

	// Find the test
	for _, suite := range p.suites {
		for _, test := range suite.Tests {
			if test.Name == event.Test {
				test.Status = StatusSkipped
				test.Duration = time.Duration(event.Elapsed * float64(time.Second))
				suite.SkippedCount++
				p.statistics.SkippedTests++
				p.statistics.TotalTests++
				return
			}

			// Check subtests
			for _, subtest := range test.Subtests {
				if subtest.Name == event.Test {
					subtest.Status = StatusSkipped
					subtest.Duration = time.Duration(event.Elapsed * float64(time.Second))
					suite.SkippedCount++
					p.statistics.SkippedTests++
					p.statistics.TotalTests++
					return
				}
			}
		}
	}
}

// onTestOutput handles test output
func (p *TestProcessor) onTestOutput(event TestEvent) {
	// Find the test
	for _, suite := range p.suites {
		for _, test := range suite.Tests {
			if test.Name == event.Test {
				test.Output += event.Output
				return
			}

			// Check subtests
			for _, subtest := range test.Subtests {
				if subtest.Name == event.Test {
					subtest.Output += event.Output
					return
				}
			}
		}
	}
}

// finalize updates final statistics
func (p *TestProcessor) finalize() {
	// Set end time and calculate duration
	p.statistics.EndTime = time.Now()
	p.statistics.Duration = p.statistics.EndTime.Sub(p.statistics.StartTime)
	p.teardownEndTime = p.statistics.EndTime

	// Calculate real phase durations based on actual timestamps
	if !p.firstTestTime.IsZero() {
		// Setup phase: from start until first test begins
		setupDuration := p.firstTestTime.Sub(p.setupStartTime)
		if setupDuration > 0 {
			p.statistics.Phases["setup"] = setupDuration
		}

		// Tests phase: from first test until last test completes
		if !p.lastTestTime.IsZero() {
			testsDuration := p.lastTestTime.Sub(p.firstTestTime)
			if testsDuration > 0 {
				p.statistics.Phases["tests"] = testsDuration
			}

			// Teardown phase: from last test completion until end
			teardownDuration := p.teardownEndTime.Sub(p.lastTestTime)
			if teardownDuration > 0 {
				p.statistics.Phases["teardown"] = teardownDuration
			}
		}
	}

	// Calculate suite statistics
	for _, suite := range p.suites {
		if suite.FailedCount > 0 {
			p.statistics.FailedFiles++
		} else if suite.TestCount > 0 {
			p.statistics.PassedFiles++
		}
		p.statistics.TotalFiles++

		// Calculate total duration (simple sum for now)
		for _, test := range suite.Tests {
			suite.Duration += test.Duration
		}
	}
}

// GetStats returns the current test run statistics
func (p *TestProcessor) GetStats() *TestRunStats {
	return p.statistics
}

// RenderResults renders the current test results
func (p *TestProcessor) RenderResults(showSummary bool) error {
	// Render each test suite
	suiteRenderer := NewSuiteRenderer(p.writer, p.formatter, p.icons, p.width)

	for _, suite := range p.suites {
		// Only expand suites with failures; collapse passing suites by default
		shouldExpand := suite.FailedCount > 0
		if err := suiteRenderer.RenderSuite(suite, !shouldExpand); err != nil {
			return err
		}
		_, _ = fmt.Fprintln(p.writer)
	}

	// Collect failed tests
	var failedTests []*TestResult
	for _, suite := range p.suites {
		for _, test := range suite.Tests {
			if test.Status == StatusFailed {
				failedTests = append(failedTests, test)
			}
		}
	}

	// Render failed tests section if any
	if len(failedTests) > 0 {
		failedRenderer := NewFailedTestRenderer(p.writer, p.formatter, p.icons, p.width)
		if err := failedRenderer.RenderFailedTests(failedTests); err != nil {
			return err
		}
	}

	// Render summary if requested
	if showSummary {
		summaryRenderer := NewSummaryRenderer(p.writer, p.formatter, p.icons, p.width)
		if err := summaryRenderer.RenderSummary(p.statistics); err != nil {
			return err
		}
	}

	return nil
}

// AddTestSuite adds a test suite to the processor's state
func (p *TestProcessor) AddTestSuite(suite *TestSuite) {
	if suite == nil {
		return
	}

	p.suites[suite.FilePath] = suite

	// Update statistics
	p.statistics.TotalTests += suite.TestCount
	p.statistics.PassedTests += suite.PassedCount
	p.statistics.FailedTests += suite.FailedCount
	p.statistics.SkippedTests += suite.SkippedCount

	// Update file statistics
	p.statistics.TotalFiles++
	if suite.FailedCount > 0 {
		p.statistics.FailedFiles++
	} else {
		p.statistics.PassedFiles++
	}
}

// createTestError creates a TestError from test output and event information
func (p *TestProcessor) createTestError(test *TestResult, event TestEvent) *TestError {
	// Extract error message from test output
	output := strings.TrimSpace(test.Output)
	lines := strings.Split(output, "\n")

	var errorMessage string
	var errorType string = "TestFailure"
	var sourceFile string
	var sourceLine int

	// Check if this is a parent test that failed due to subtest failures
	isParentTestFailure := p.isParentTestFailure(test, output)

	// Parse output to extract error information
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for error messages (lines that start with test file and line number)
		if strings.Contains(trimmed, ".go:") && strings.Contains(trimmed, ": ") {
			// This looks like an error line: filename.go:123: error message
			parts := strings.SplitN(trimmed, ": ", 2)
			if len(parts) == 2 {
				// Extract file and line info
				fileLinePart := parts[0]
				errorMessage = parts[1]

				// Try to extract file name and line number
				if lastSpace := strings.LastIndex(fileLinePart, " "); lastSpace != -1 {
					fileLinePart = fileLinePart[lastSpace+1:]
				}

				if colonIndex := strings.LastIndex(fileLinePart, ":"); colonIndex != -1 {
					sourceFile = fileLinePart[:colonIndex]
					if lineStr := fileLinePart[colonIndex+1:]; lineStr != "" {
						if line, err := strconv.Atoi(lineStr); err == nil {
							sourceLine = line
						}
					}
				}
				break
			}
		}

		// Look for test function calls in output (alternative approach)
		if sourceFile == "" && (strings.Contains(trimmed, "t.Error") ||
			strings.Contains(trimmed, "t.Fail") ||
			strings.Contains(trimmed, "panic")) {
			// Try to find a reference to the test file in surrounding context
			for i, contextLine := range lines {
				if strings.Contains(contextLine, ".go:") {
					// Found a file reference, extract it
					if fileMatch := extractFileLocationFromLine(contextLine); fileMatch != nil {
						sourceFile = fileMatch.file
						sourceLine = fileMatch.line
						break
					}
				}
				// Don't search too far
				if i > 10 {
					break
				}
			}
		}

		// If no structured error found, use the first non-empty line as error message
		if errorMessage == "" && trimmed != "" && !strings.HasPrefix(trimmed, "=== ") &&
			!strings.HasPrefix(trimmed, "--- ") {
			errorMessage = trimmed
		}
	}

	// For parent test failures, always try to infer source location
	if sourceFile == "" || isParentTestFailure {
		inferredFile := p.inferSourceFileFromTest(test, event)
		if inferredFile != "" {
			sourceFile = inferredFile
			sourceLine = p.findTestFunctionLine(sourceFile, test.Name)
		}
	}

	// Improve error message for parent test failures
	if isParentTestFailure && (errorMessage == "" || errorMessage == "Test failed") {
		failedSubtestCount := p.countFailedSubtests(test)
		if failedSubtestCount > 0 {
			if failedSubtestCount == 1 {
				errorMessage = "Test failed due to 1 failed subtest"
			} else {
				errorMessage = fmt.Sprintf("Test failed due to %d failed subtests", failedSubtestCount)
			}
			errorType = "SubtestFailure"
		}
	}

	// Default error message if none found
	if errorMessage == "" {
		errorMessage = "Test failed"
	}

	// Detect error type from message content
	if strings.Contains(errorMessage, "panic") {
		errorType = "Panic"
	} else if strings.Contains(errorMessage, "timeout") {
		errorType = "Timeout"
	} else if strings.Contains(errorMessage, "Expected") || strings.Contains(errorMessage, "expected") {
		errorType = "AssertionError"
	}

	// Create location and extract source context
	var location *SourceLocation
	var sourceContext []string
	var highlightedLine int

	if sourceFile != "" && sourceLine > 0 {
		// Try to determine column position from error message context
		column := p.determineErrorColumn(errorMessage, sourceFile, sourceLine)

		location = &SourceLocation{
			File:     sourceFile,
			Line:     sourceLine,
			Column:   column,
			Function: test.Name,
		}

		// Extract source context from the actual file
		sourceContext, highlightedLine = p.extractSourceContext(sourceFile, sourceLine)
	}

	return &TestError{
		Message:         errorMessage,
		Type:            errorType,
		Stack:           output, // Store full output as stack trace
		Location:        location,
		SourceContext:   sourceContext,
		HighlightedLine: highlightedLine,
	}
}

// isParentTestFailure determines if a test failed because of subtest failures rather than its own error
func (p *TestProcessor) isParentTestFailure(test *TestResult, output string) bool {
	// Check if this test has subtests
	hasSubtests := len(test.Subtests) > 0 || strings.Contains(test.Name, "/")

	// Check if the output indicates subtest failures rather than direct failures
	lines := strings.Split(output, "\n")
	hasDirectError := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Look for direct error indicators (t.Error, t.Fail, panic, etc.)
		if strings.Contains(trimmed, "t.Error") ||
			strings.Contains(trimmed, "t.Fail") ||
			strings.Contains(trimmed, "panic") ||
			(strings.Contains(trimmed, ".go:") && strings.Contains(trimmed, ": ")) {
			hasDirectError = true
			break
		}
	}

	// It's a parent test failure if it has subtests but no direct error
	return hasSubtests && !hasDirectError
}

// countFailedSubtests counts how many subtests failed for a given test
func (p *TestProcessor) countFailedSubtests(test *TestResult) int {
	failedCount := 0

	// Count direct subtests
	for _, subtest := range test.Subtests {
		if subtest.Status == StatusFailed {
			failedCount++
		}
	}

	// For nested tests, also search through all suites for related failures
	testPrefix := test.Name + "/"
	for _, suite := range p.suites {
		for _, suiteTest := range suite.Tests {
			if strings.HasPrefix(suiteTest.Name, testPrefix) && suiteTest.Status == StatusFailed {
				failedCount++
			}
		}
	}

	return failedCount
}

// Helper struct for file location extraction
type fileLocation struct {
	file string
	line int
}

// extractFileLocationFromLine extracts file and line number from a line
func extractFileLocationFromLine(line string) *fileLocation {
	// Look for patterns like "filename.go:123"
	parts := strings.Fields(line)
	for _, part := range parts {
		if strings.Contains(part, ".go:") {
			if colonIndex := strings.LastIndex(part, ":"); colonIndex != -1 {
				file := part[:colonIndex]
				if lineStr := part[colonIndex+1:]; lineStr != "" {
					// Remove any trailing characters that aren't digits
					var numStr string
					for _, char := range lineStr {
						if char >= '0' && char <= '9' {
							numStr += string(char)
						} else {
							break
						}
					}
					if line, err := strconv.Atoi(numStr); err == nil {
						return &fileLocation{file: file, line: line}
					}
				}
			}
		}
	}
	return nil
}

// inferSourceFileFromTest attempts to infer the source file from test information
func (p *TestProcessor) inferSourceFileFromTest(test *TestResult, event TestEvent) string {
	// If we have package information, try to construct the test file name
	if event.Package != "" {
		// For command-line-arguments, look for common test file patterns
		if event.Package == "command-line-arguments" {
			// Try common test file names in the current directory
			possibleFiles := []string{
				"basic_failures_test.go",
				"extreme_scenarios_test.go",
				test.Name + "_test.go",
			}

			for _, filename := range possibleFiles {
				if _, err := os.Stat(filename); err == nil {
					return filename
				}
				// Also try in stress_tests directory
				stressPath := filepath.Join("stress_tests", filename)
				if _, err := os.Stat(stressPath); err == nil {
					return stressPath
				}
			}
		} else {
			// For named packages, construct the likely test file name
			packageParts := strings.Split(event.Package, "/")
			if len(packageParts) > 0 {
				lastPart := packageParts[len(packageParts)-1]
				return lastPart + "_test.go"
			}
		}
	}

	return ""
}

// findTestFunctionLine finds the line number where a test function is defined
func (p *TestProcessor) findTestFunctionLine(filename string, testName string) int {
	content, err := os.ReadFile(filename)
	if err != nil {
		return 0
	}

	lines := strings.Split(string(content), "\n")

	// Extract the base test name (remove subtest parts)
	baseTestName := testName
	if slashIndex := strings.Index(testName, "/"); slashIndex != -1 {
		baseTestName = testName[:slashIndex]
	}

	// Look for the test function definition
	functionPattern := "func " + baseTestName + "("

	for i, line := range lines {
		if strings.Contains(line, functionPattern) {
			return i + 1 // Convert to 1-based line number
		}
	}

	return 0
}

// extractSourceContext reads the source file and extracts context around the error line
func (p *TestProcessor) extractSourceContext(filename string, errorLine int) ([]string, int) {
	// Try multiple possible paths for the source file
	possiblePaths := []string{
		filename,                                // Direct filename
		filepath.Join("stress_tests", filename), // In stress_tests directory
		filepath.Join(".", filename),            // Current directory
	}

	// Also try checking known suite paths
	for _, suite := range p.suites {
		if suite.FilePath != "" {
			dir := filepath.Dir(suite.FilePath)
			if dir != "" && dir != "." {
				possiblePaths = append(possiblePaths, filepath.Join(dir, filename))
			}
		}
	}

	var content []byte
	var err error

	// Try each possible path
	for _, path := range possiblePaths {
		content, err = os.ReadFile(path)
		if err == nil {
			break // Found the file
		}
	}

	if err != nil {
		// File not found or not readable, return empty context
		return nil, 0
	}

	// Split into lines
	lines := strings.Split(string(content), "\n")

	// Calculate context range (2 lines before and after)
	contextSize := 2
	startLine := errorLine - contextSize - 1 // Convert to 0-based index
	endLine := errorLine + contextSize - 1   // Convert to 0-based index

	// Ensure bounds
	if startLine < 0 {
		startLine = 0
	}
	if endLine >= len(lines) {
		endLine = len(lines) - 1
	}

	// Extract context lines
	var context []string
	for i := startLine; i <= endLine; i++ {
		if i < len(lines) {
			context = append(context, lines[i])
		}
	}

	// Calculate which line in the context is the error line
	highlightedLine := (errorLine - 1) - startLine
	if highlightedLine < 0 {
		highlightedLine = 0
	}
	if highlightedLine >= len(context) {
		highlightedLine = len(context) - 1
	}

	return context, highlightedLine
}

// determineErrorColumn attempts to determine the column position from error message context
func (p *TestProcessor) determineErrorColumn(errorMessage, sourceFile string, sourceLine int) int {
	// Read the source file to analyze the error line
	content, err := os.ReadFile(sourceFile)
	if err != nil {
		return 1 // Default to column 1 if file can't be read
	}

	lines := strings.Split(string(content), "\n")
	if sourceLine <= 0 || sourceLine > len(lines) {
		return 1 // Default to column 1 if line is out of bounds
	}

	// Get the source line (convert to 0-based index)
	sourceLiteral := lines[sourceLine-1]

	// Try to find error-related keywords in the line and position the pointer
	errorLower := strings.ToLower(errorMessage)

	// For assertion errors, try to point to the assertion call
	if strings.Contains(errorLower, "expected") || strings.Contains(errorLower, "assert") {
		// Look for test function calls first
		if pos := strings.Index(sourceLiteral, "t.Errorf"); pos != -1 {
			return pos + 1 // Point at t.Errorf
		}
		if pos := strings.Index(sourceLiteral, "t.Error"); pos != -1 {
			return pos + 1 // Point at t.Error
		}
		if pos := strings.Index(sourceLiteral, "t.Fatalf"); pos != -1 {
			return pos + 1 // Point at t.Fatalf
		}
		if pos := strings.Index(sourceLiteral, "t.Fatal"); pos != -1 {
			return pos + 1 // Point at t.Fatal
		}

		// Look for assertion operators as fallback
		if pos := strings.Index(sourceLiteral, "!="); pos != -1 {
			return pos + 1 // Point at the operator
		}
		if pos := strings.Index(sourceLiteral, "=="); pos != -1 {
			return pos + 1 // Point at the operator
		}
		if pos := strings.Index(sourceLiteral, "if "); pos != -1 {
			return pos + 4 // Point after "if "
		}
	}

	// For panic errors, try to point to the problematic operation
	if strings.Contains(errorLower, "panic") || strings.Contains(errorLower, "index out of range") {
		// Look for test function calls that might be the source
		if pos := strings.Index(sourceLiteral, "t.Errorf"); pos != -1 {
			return pos + 1 // Point at the reporting function
		}
		if pos := strings.Index(sourceLiteral, "t.Error"); pos != -1 {
			return pos + 1 // Point at the reporting function
		}

		// Look for array/slice access
		if pos := strings.Index(sourceLiteral, "["); pos != -1 {
			return pos + 1 // Point at the bracket
		}
		// Look for nil pointer dereference
		if pos := strings.Index(sourceLiteral, "nil"); pos != -1 {
			return pos + 1 // Point at nil
		}
	}

	// Look for any test function call as a general fallback
	testFunctions := []string{"t.Errorf", "t.Error", "t.Fatalf", "t.Fatal", "t.Logf", "t.Log"}
	for _, fn := range testFunctions {
		if pos := strings.Index(sourceLiteral, fn); pos != -1 {
			return pos + 1 // Point at the test function
		}
	}

	// Default: try to find the first non-whitespace character after indentation
	for i, char := range sourceLiteral {
		if char != ' ' && char != '\t' {
			return i + 1 // Convert to 1-based indexing
		}
	}

	return 1 // Default to column 1
}
