package cli

import (
	"io"
	"time"
)

// TestEvent represents a JSON event from Go test output
type TestEvent struct {
	Time    string  `json:"Time"`    // RFC3339Nano formatted timestamp
	Action  string  `json:"Action"`  // run, pass, fail, skip, output
	Package string  `json:"Package"` // Package being tested
	Test    string  `json:"Test"`    // Name of the test
	Output  string  `json:"Output"`  // Test output (stdout/stderr)
	Elapsed float64 `json:"Elapsed"` // Test duration in seconds
}

// TestProcessor processes test output and tracks statistics
type TestProcessor struct {
	writer     io.Writer
	formatter  *ColorFormatter
	icons      *IconProvider
	width      int
	suites     map[string]*TestSuite
	statistics *TestRunStats
	startTime  time.Time

	// Phase tracking timestamps for real duration measurement
	setupStartTime  time.Time // When processing started
	firstTestTime   time.Time // When first test began running
	lastTestTime    time.Time // When last test completed
	teardownEndTime time.Time // When processing ended
}

// TestProcessorInterface defines the interface for test processors
type TestProcessorInterface interface {
	ProcessJSONOutput(output string) error
	ProcessStream(r io.Reader, progress chan<- TestProgress) error
	Reset()
	GetStats() *TestRunStats
	RenderResults(showSummary bool) error
	AddTestSuite(suite *TestSuite)
}

// TestProgress represents real-time progress information
type TestProgress struct {
	CompletedTests int
	TotalTests     int
	CurrentFile    string
	Status         TestStatus
}
