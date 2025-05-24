package processor

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/term"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// TestProcessor processes test output and tracks statistics
type TestProcessor struct {
	writer     io.Writer
	formatter  ColorFormatter // Will be refactored to interface later
	icons      IconProvider   // Will be refactored to interface later
	width      int
	suites     map[string]*models.TestSuite
	statistics *models.TestRunStats
	startTime  time.Time

	// Phase tracking timestamps for real duration measurement
	setupStartTime  time.Time // When processing started
	firstTestTime   time.Time // When first test began running
	lastTestTime    time.Time // When last test completed
	teardownEndTime time.Time // When processing ended
}

// ColorFormatter interface matching the existing CLI ColorFormatter
type ColorFormatter interface {
	Red(text string) string
	Green(text string) string
	Yellow(text string) string
	Blue(text string) string
	Magenta(text string) string
	Cyan(text string) string
	Gray(text string) string
	Bold(text string) string
	Dim(text string) string
	White(text string) string
	Colorize(text, colorName string) string
}

// IconProvider interface matching the existing CLI IconProvider
type IconProvider interface {
	CheckMark() string
	Cross() string
	Skipped() string
	Running() string
	GetIcon(iconType string) string
}

// NewTestProcessor creates a new TestProcessor
func NewTestProcessor(writer io.Writer, formatter ColorFormatter, icons IconProvider, width int) *TestProcessor {
	return &TestProcessor{
		writer:    writer,
		formatter: formatter,
		icons:     icons,
		width:     getTerminalWidthForProcessor(),
		suites:    make(map[string]*models.TestSuite),
		statistics: &models.TestRunStats{
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

		var event models.TestEvent
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
	p.statistics = &models.TestRunStats{
		StartTime: now,
		Phases:    make(map[string]time.Duration),
	}
	p.suites = make(map[string]*models.TestSuite)
	p.setupStartTime = now
	p.firstTestTime = time.Time{}   // Will be set when first test runs
	p.lastTestTime = time.Time{}    // Will be set when last test completes
	p.teardownEndTime = time.Time{} // Will be set at finalization
}

// GetStats returns the current test run statistics
func (p *TestProcessor) GetStats() *models.TestRunStats {
	return p.statistics
}

// GetWriter returns the writer for backward compatibility
func (p *TestProcessor) GetWriter() io.Writer {
	return p.writer
}

// GetSuites returns the test suites for backward compatibility
func (p *TestProcessor) GetSuites() map[string]*models.TestSuite {
	return p.suites
}

// RenderResults renders the final test results
func (p *TestProcessor) RenderResults(showSummary bool) error {
	// This will be moved to UI layer later, for now just placeholder
	if showSummary {
		fmt.Fprintf(p.writer, "Tests completed: %d passed, %d failed, %d skipped\n",
			p.statistics.PassedTests, p.statistics.FailedTests, p.statistics.SkippedTests)
	}
	return nil
}

// AddTestSuite adds a test suite to the processor
func (p *TestProcessor) AddTestSuite(suite *models.TestSuite) {
	if suite.FilePath == "" {
		suite.FilePath = "unknown"
	}
	p.suites[suite.FilePath] = suite

	// Update statistics
	p.statistics.TotalFiles++
	p.statistics.TotalTests += suite.TestCount
	p.statistics.PassedTests += suite.PassedCount
	p.statistics.FailedTests += suite.FailedCount
	p.statistics.SkippedTests += suite.SkippedCount

	// Update file-level statistics
	if suite.FailedCount > 0 {
		p.statistics.FailedFiles++
	} else {
		p.statistics.PassedFiles++
	}
}

// ProcessStream processes a stream of test output with real-time updates
func (p *TestProcessor) ProcessStream(r io.Reader, progress chan<- models.TestProgress) error {
	// Create a stream parser
	parser := NewStreamParser()

	// Create a channel for test results
	resultCh := make(chan *models.LegacyTestResult, 10)

	// Start a goroutine to parse the stream
	go func() {
		if err := parser.Parse(r, resultCh); err != nil && err != io.EOF {
			_, _ = fmt.Fprintf(p.writer, "Error parsing test output: %v\n", err)
		}
		close(resultCh)
	}()

	// Process test results as they arrive
	totalTests := 0
	completedTests := 0

	for result := range resultCh {
		totalTests++

		// Track timing
		if p.firstTestTime.IsZero() && result.Status != models.StatusRunning {
			p.firstTestTime = time.Now()
		}

		if result.Status != models.StatusRunning {
			p.lastTestTime = time.Now()
			completedTests++
		}

		// Send progress update
		if progress != nil {
			progress <- models.TestProgress{
				CompletedTests: completedTests,
				TotalTests:     totalTests,
				CurrentFile:    result.Package,
				Status:         result.Status,
			}
		}
	}

	return nil
}

// finalize will be defined in statistics.go
// Temporary placeholder to make the file compile
func (p *TestProcessor) finalize() {
	// Set end time and calculate duration
	p.statistics.EndTime = time.Now()
	p.statistics.Duration = p.statistics.EndTime.Sub(p.statistics.StartTime)
	p.teardownEndTime = p.statistics.EndTime
}
