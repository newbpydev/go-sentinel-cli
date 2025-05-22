package cli

import (
	"fmt"
	"io"
	"log"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

// Renderer handles the display of test results
type Renderer struct {
	out    io.Writer
	style  *Style
	width  int
	height int
}

// write is a helper method to handle write errors
func (r *Renderer) write(format string, args ...interface{}) {
	if _, err := fmt.Fprintf(r.out, format, args...); err != nil {
		log.Printf("Error writing to output: %v", err)
	}
}

// writeln is a helper method to handle write errors with newline
func (r *Renderer) writeln(format string, args ...interface{}) {
	r.write(format+"\n", args...)
}

// NewRenderer creates a new test result renderer
func NewRenderer(out io.Writer) *Renderer {
	return &Renderer{
		out:   out,
		style: NewStyle(true), // Enable colors by default
	}
}

// NewRendererWithStyle creates a new renderer with a custom style
func NewRendererWithStyle(out io.Writer, useColors bool) *Renderer {
	return &Renderer{
		out:   out,
		style: NewStyle(useColors),
	}
}

// RenderTestRun renders a complete test run
func (r *Renderer) RenderTestRun(run *TestRun) {
	// Header
	r.renderHeader()

	// Test results
	for _, suite := range run.Suites {
		r.renderSuite(suite)
	}

	// Add a newline before summary
	r.writeln("")

	// Render the summary
	r.renderSummary(run)
}

// renderSummary is the single source of truth for rendering test summaries
func (r *Renderer) renderSummary(run *TestRun) {
	// Calculate file statistics
	passedFiles := 0
	failedFiles := 0
	for _, suite := range run.Suites {
		if suite.NumFailed > 0 {
			failedFiles++
		} else if suite.NumTotal > 0 {
			passedFiles++
		}
	}

	// Add a divider line before summary
	r.writeln("")

	// Format summaries with consistent spacing and color
	r.writeln(r.style.FormatTestSummary("Test Files", failedFiles, passedFiles, 0, len(run.Suites)))
	r.writeln(r.style.FormatTestSummary("Tests", run.NumFailed, run.NumPassed, run.NumSkipped, run.NumTotal))

	// Add total duration and (if possible) heap usage
	r.writeln("")
	r.writeln(r.style.FormatTimestamp("Start at", run.StartTime))
	if !run.EndTime.IsZero() {
		r.writeln(r.style.FormatTimestamp("End at", run.EndTime))
	}

	totalDuration := run.Duration
	mainDurationStr := FormatDurationAdaptive(totalDuration)
	formattedMainDuration := r.style.FormatDuration("Duration", mainDurationStr)

	// Add breakdown details
	breakdownParts := []string{}
	if run.SetupDuration > 0 {
		breakdownParts = append(breakdownParts, fmt.Sprintf("setup %s", FormatDurationAdaptive(run.SetupDuration)))
	}
	if run.CollectDuration > 0 {
		breakdownParts = append(breakdownParts, fmt.Sprintf("collect %s", FormatDurationAdaptive(run.CollectDuration)))
	}
	if run.TestsDuration > 0 {
		breakdownParts = append(breakdownParts, fmt.Sprintf("tests %s", FormatDurationAdaptive(run.TestsDuration)))
	}
	if run.ParseDuration > 0 {
		breakdownParts = append(breakdownParts, fmt.Sprintf("parse %s", FormatDurationAdaptive(run.ParseDuration)))
	}
	if len(breakdownParts) > 0 {
		formattedMainDuration += " " + r.style.FormatBreakdownText(fmt.Sprintf("(%s)", strings.Join(breakdownParts, ", ")))
	}
	r.writeln(formattedMainDuration)

	// Show failed tests if any
	if run.NumFailed > 0 {
		r.writeln("")
		r.writeln(r.style.FormatErrorHeader(" FAILED Tests "))
		r.writeln("")
		for _, suite := range run.Suites {
			if suite.NumFailed > 0 {
				r.writeln(r.style.FormatFailedSuite(suite.FilePath))
				for _, test := range suite.Tests {
					if test.Status == TestStatusFailed {
						testName := test.Name
						if strings.Contains(testName, "/") {
							parts := strings.Split(testName, "/")
							testName = parts[len(parts)-1]
						}
						r.writeln("    %s", r.style.FormatFailedTest(testName))
						if test.Error != nil {
							if test.Error.Message != "" {
								msg := strings.TrimSpace(test.Error.Message)
								if idx := strings.Index(msg, "\n"); idx > 0 {
									msg = msg[:idx]
								}
								r.writeln("    %s", r.style.FormatErrorMessage(msg))
							}
							if test.Error.Location != nil {
								r.writeln("    %s", r.style.FormatErrorLocation(test.Error.Location))
							}
						}
						r.writeln("")
					}
				}
			}
		}
	}
}

// renderHeader renders the test run header
func (r *Renderer) renderHeader() {
	header := r.style.FormatHeader(" GO SENTINEL ")
	r.writeln("%s", header)
	r.writeln("")
}

// renderSuite renders a test suite
func (r *Renderer) renderSuite(suite *TestSuite) {
	// Measure heap before
	var memBefore, memAfter runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	// Format file path to be more readable
	filePath := formatFilePath(suite.FilePath)
	// Convert _test.go to .test.ts for visual consistency with Vitest
	filePath = strings.TrimSuffix(filePath, "_test.go") + ".test.ts"

	// Format suite header
	totalTests := suite.NumTotal
	var headerParts []string

	// Add file path
	headerParts = append(headerParts, filePath)

	// Add test count and failed count if any
	testCountStr := fmt.Sprintf("%d tests", totalTests)
	if suite.NumFailed > 0 {
		testCountStr = fmt.Sprintf("%d tests | %d failed", totalTests, suite.NumFailed)
	}
	headerParts = append(headerParts, testCountStr)

	// Add duration
	if suite.Duration > 0 {
		headerParts = append(headerParts, FormatDurationPrecise(suite.Duration))
	}

	// Add heap info if available
	var heapInfo string
	if runtime.GOOS != "js" {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		heapInfo = fmt.Sprintf("%d MB heap used", m.HeapAlloc/1024/1024)
		headerParts = append(headerParts, heapInfo)
	}

	headerText := strings.Join(headerParts, " | ")

	// Determine suite status
	status := TestStatusPassed
	if suite.NumFailed > 0 {
		status = TestStatusFailed
	} else if suite.NumSkipped > 0 && suite.NumPassed == 0 {
		status = TestStatusSkipped
	}

	// Style header based on status
	var headerStyle lipgloss.Style
	switch status {
	case TestStatusFailed:
		headerStyle = errorStyle.Copy()
	case TestStatusSkipped:
		headerStyle = warningStyle.Copy()
	default:
		headerStyle = successStyle.Copy()
	}

	// Add padding and render header
	headerStyle = headerStyle.PaddingLeft(1)
	fmt.Fprintln(r.out, headerStyle.Render(headerText))

	// Render test results
	for _, test := range suite.Tests {
		r.RenderTestResult(test)
	}

	// Add spacing after test results
	if len(suite.Tests) > 0 {
		fmt.Fprintln(r.out)
	}

	// Measure heap after
	runtime.ReadMemStats(&memAfter)
}

// pluralize returns the plural form of a word if count != 1
func pluralize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

// formatBytes formats bytes into human readable string
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// formatFilePath formats a file path to be more readable
func formatFilePath(path string) string {
	// Remove common prefixes
	prefixes := []string{
		"github.com/",
		"internal/",
		"cmd/",
		"pkg/",
		"test/",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			path = strings.TrimPrefix(path, prefix)
		}
	}

	// Split the path into parts
	parts := strings.Split(path, "/")

	// Handle special cases
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]

		// Convert _test.go to .test.ts for better visual alignment with Vitest
		if strings.HasSuffix(lastPart, "_test.go") {
			lastPart = strings.TrimSuffix(lastPart, "_test.go") + ".test.ts"
			parts[len(parts)-1] = lastPart
		}

		// If it's a package path, use the last meaningful part
		if strings.Contains(path, "/") && !strings.HasSuffix(lastPart, ".go") && !strings.HasSuffix(lastPart, ".ts") {
			// Find the last meaningful part (not "pkg", "internal", etc.)
			for i := len(parts) - 1; i >= 0; i-- {
				if !isCommonDir(parts[i]) {
					parts = parts[i:]
					break
				}
			}
		}
	}

	return strings.Join(parts, "/")
}

// isCommonDir checks if a directory name is a common one that should be trimmed
func isCommonDir(dir string) bool {
	commonDirs := map[string]bool{
		"pkg":      true,
		"internal": true,
		"cmd":      true,
		"test":     true,
		"tests":    true,
		"src":      true,
	}
	return commonDirs[dir]
}

// RenderTestResult renders a single test result
func (r *Renderer) RenderTestResult(result *TestResult) {
	// Format test name with icon and color
	icon := r.style.StatusIcon(result.Status)

	// Format the test name
	name := formatTestName(result.Name)

	// Format duration with right alignment
	duration := ""
	if result.Status != TestStatusRunning && result.Status != TestStatusPending {
		duration = FormatDurationPrecise(result.Duration)
	}

	// Choose color for test name and icon
	var style lipgloss.Style
	switch result.Status {
	case TestStatusPassed:
		style = successStyle.Copy()
	case TestStatusFailed:
		style = errorStyle.Copy()
	case TestStatusSkipped:
		style = warningStyle.Copy()
	default:
		style = dimStyle.Copy()
	}

	// Calculate indentation based on test hierarchy
	indent := "  "
	if strings.Contains(result.Name, "/") {
		parts := strings.Split(result.Name, "/")
		indent = strings.Repeat("  ", len(parts))
	}

	// Format the line with proper spacing and indentation
	line := fmt.Sprintf("%s%s %s %s", indent, icon, name, duration)
	r.out.Write([]byte(style.Render(line) + "\n"))

	// Format error if present
	if result.Error != nil {
		r.renderError(result.Error, strings.Count(result.Name, "/")+1)
	}
}

// formatTestName formats a test name to be more readable
func formatTestName(name string) string {
	// Remove Test prefix if present
	name = strings.TrimPrefix(name, "Test")

	// Handle subtests
	if strings.Contains(name, "/") {
		parts := strings.Split(name, "/")
		// Format each part
		for i, part := range parts {
			// For subtest parts, keep the original casing if it looks intentional
			if i > 0 && (strings.Contains(part, "_") || strings.Contains(part, " ")) {
				parts[i] = formatTestPart(part)
			} else if i > 0 {
				// For clean subtest names, just trim Test prefix
				parts[i] = strings.TrimPrefix(part, "Test")
			} else {
				// Format the main test name
				parts[i] = formatTestPart(part)
			}
		}
		// Join with Vitest-style separator
		return strings.Join(parts, " › ")
	}

	return formatTestPart(name)
}

// formatTestPart formats a single part of a test name
func formatTestPart(part string) string {
	// Handle empty parts
	if part == "" {
		return ""
	}

	// Remove Test prefix if present
	part = strings.TrimPrefix(part, "Test")

	// Split on underscores, spaces, and numbers
	words := splitTestName(part)

	// Format each word
	for i, word := range words {
		if word == "" {
			continue
		}

		// Keep common abbreviations uppercase
		if isCommonAbbreviation(word) {
			words[i] = strings.ToUpper(word)
			continue
		}

		// Handle numbers
		if isNumeric(word) {
			continue
		}

		// Convert first word to title case, rest to lower
		if i == 0 {
			words[i] = strings.Title(strings.ToLower(word))
		} else {
			words[i] = strings.ToLower(word)
		}
	}

	// Join words with a single space
	return strings.Join(words, " ")
}

// splitTestName splits a test name into words based on common patterns
func splitTestName(name string) []string {
	var words []string
	var current string
	var lastType rune

	// Character types
	const (
		lower = 'a'
		upper = 'A'
		digit = '0'
		other = '_'
	)

	getType := func(r rune) rune {
		switch {
		case unicode.IsLower(r):
			return lower
		case unicode.IsUpper(r):
			return upper
		case unicode.IsDigit(r):
			return digit
		default:
			return other
		}
	}

	for i, r := range name {
		t := getType(r)

		// Start a new word on type changes
		if i > 0 {
			newWord := false

			switch {
			case t == other:
				// Always split on special characters
				newWord = true
			case lastType == lower && t == upper:
				// camelCase to Camel
				newWord = true
			case lastType == upper && t == lower && len(current) > 1:
				// HTTPRequest to HTTP Request
				words = append(words, current[:len(current)-1])
				current = string(name[i-1]) + string(r)
			case lastType != digit && t == digit:
				// Word2 to Word 2
				newWord = true
			case lastType == digit && t != digit:
				// 2Word to 2 Word
				newWord = true
			}

			if newWord && current != "" {
				words = append(words, current)
				current = ""
			}
		}

		current += string(r)
		lastType = t
	}

	if current != "" {
		words = append(words, current)
	}

	return words
}

// isNumeric checks if a string is a number
func isNumeric(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// isCommonAbbreviation checks if a string is a common abbreviation
func isCommonAbbreviation(s string) bool {
	// Add more common abbreviations as needed
	commonAbbreviations := map[string]bool{
		"HTTP":      true,
		"HTTPS":     true,
		"API":       true,
		"REST":      true,
		"URL":       true,
		"URI":       true,
		"SQL":       true,
		"XML":       true,
		"JSON":      true,
		"HTML":      true,
		"CSS":       true,
		"JS":        true,
		"ID":        true,
		"UUID":      true,
		"TCP":       true,
		"UDP":       true,
		"IP":        true,
		"FTP":       true,
		"SMTP":      true,
		"SSH":       true,
		"SSL":       true,
		"TLS":       true,
		"JWT":       true,
		"OAuth":     true,
		"OAuth2":    true,
		"WebSocket": true,
		"WS":        true,
		"WSS":       true,
	}
	return commonAbbreviations[strings.ToUpper(s)]
}

// renderErrors renders a list of test errors
func (r *Renderer) renderErrors(errors []*TestError) {
	for _, err := range errors {
		r.renderError(err, 0)
	}
}

// renderError renders a test error in Vitest style
func (r *Renderer) renderError(err *TestError, depth int) {
	indent := strings.Repeat("  ", depth)

	// Format error message with arrow
	if err.Message != "" {
		msg := strings.TrimSpace(err.Message)
		// Split on newlines and format each line
		for _, line := range strings.Split(msg, "\n") {
			if line != "" {
				errorLine := fmt.Sprintf("%s→ %s", indent, line)
				r.out.Write([]byte(errorStyle.Render(errorLine) + "\n"))
			}
		}
	}

	// Show location with file and line
	if err.Location != nil {
		// Format location in Vitest style
		locLine := fmt.Sprintf("%s  at %s:%d", indent, err.Location.File, err.Location.Line)
		r.out.Write([]byte(dimStyle.Render(locLine) + "\n"))

		// Show code snippet if available
		if err.Location.Snippet != "" {
			// Format snippet with line numbers and highlighting
			snippetLines := strings.Split(strings.TrimSpace(err.Location.Snippet), "\n")
			startLine := err.Location.StartLine
			for i, line := range snippetLines {
				lineNum := startLine + i
				// Highlight the error line
				if lineNum == err.Location.Line {
					snippetLine := fmt.Sprintf("%s    %d │ %s", indent, lineNum, line)
					r.out.Write([]byte(errorStyle.Render(snippetLine) + "\n"))
				} else {
					snippetLine := fmt.Sprintf("%s    %d │ %s", indent, lineNum, line)
					r.out.Write([]byte(dimStyle.Render(snippetLine) + "\n"))
				}
			}
		}
	}

	// Show expected/actual values in a clean format
	if err.Expected != "" || err.Actual != "" {
		r.out.Write([]byte("\n")) // Add spacing
		if err.Expected != "" {
			r.out.Write([]byte(dimStyle.Render(fmt.Sprintf("%s  Expected", indent)) + "\n"))
			r.out.Write([]byte(errorStyle.Render(fmt.Sprintf("%s    %s", indent, err.Expected)) + "\n"))
		}
		if err.Actual != "" {
			r.out.Write([]byte(dimStyle.Render(fmt.Sprintf("%s  Actual", indent)) + "\n"))
			r.out.Write([]byte(errorStyle.Render(fmt.Sprintf("%s    %s", indent, err.Actual)) + "\n"))
		}
		r.out.Write([]byte("\n")) // Add spacing
	}
}

// RenderTestStart renders the start of a test run
func (r *Renderer) RenderTestStart(_ *TestRun) {
	// Add a blank line before test output
	r.writeln("")
}

// SetDimensions sets the terminal dimensions
func (r *Renderer) SetDimensions(width, height int) {
	r.width = width
	r.height = height
}

// RenderWatchHeader displays the watch mode header
func (r *Renderer) RenderWatchHeader() {
	r.writeln("%s", r.style.FormatHeader(" WATCH MODE "))
	r.writeln(" Press 'a' to run all tests")
	r.writeln(" Press 'f' to run only failed tests")
	r.writeln(" Press 'q' to quit")
	r.writeln("")
}

// RenderFileChange displays a file change notification
func (r *Renderer) RenderFileChange(path string) {
	r.writeln("\nFile changed: %s\n", path)
}

// Helper functions

// RenderFinalSummary renders the final test summary
func (r *Renderer) RenderFinalSummary(run *TestRun) {
	// Use the consolidated summary rendering
	r.renderSummary(run)
}

// RenderTestSummary is deprecated and should not be used
func (r *Renderer) RenderTestSummary(run *TestRun) {
	// This function is deprecated and should not be used
	// Use RenderFinalSummary instead
}

// RenderProgress renders the current test progress
func (r *Renderer) RenderProgress(run *TestRun) {
	completed := run.NumPassed + run.NumFailed + run.NumSkipped
	if run.NumTotal == 0 {
		return
	}

	percentage := float64(completed) / float64(run.NumTotal) * 100
	r.write("Running tests... %.0f%% (%d/%d)\n", percentage, completed, run.NumTotal)
}

// RenderSuiteSummary renders a test suite summary
func (r *Renderer) RenderSuiteSummary(suite *TestSuite) {
	// Only show summary for suites with failures
	if suite.NumFailed == 0 {
		return
	}

	r.writeln("Suite")
	r.writeln("  %s", suite.FilePath)
	r.writeln("  Total: %d", suite.NumTotal)
	r.writeln("  Passed: %d", suite.NumPassed)
	r.writeln("  Failed: %d", suite.NumFailed)
	r.writeln("  Skipped: %d", suite.NumSkipped)
	r.writeln("  Time: %s", FormatDurationPrecise(suite.Duration))
	r.writeln("")
}

// RenderSuite renders a test suite
func (r *Renderer) RenderSuite(suite *TestSuite) {
	// Print suite header
	if _, err := fmt.Fprintf(r.out, "%s\n", r.style.FormatHeader(fmt.Sprintf(" %s ", suite.Package))); err != nil {
		log.Printf("Error writing suite header: %v", err)
	}

	// Test results
	for _, result := range suite.Tests {
		r.RenderTestResult(result)
	}

	// Suite errors
	if len(suite.Errors) > 0 {
		r.renderErrors(suite.Errors)
	}

	r.writeln("")
}

// RenderTest renders a test result
func (r *Renderer) RenderTest(test *TestResult, indent string) {
	// Print test name
	name := r.style.FormatTestName(test)
	r.write("%s%s\n", indent, name)

	// Print error if test failed
	if test.Error != nil {
		r.write("%sError:\n", indent)
		// ... rest of the error handling ...
	}
}

// formatLocation formats a location to be more readable
func formatLocation(loc *SourceLocation) string {
	if loc.File == "" {
		return ""
	}
	return fmt.Sprintf("%s:%d", formatFilePath(loc.File), loc.Line)
}
