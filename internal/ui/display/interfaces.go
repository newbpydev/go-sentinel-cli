// Package display provides test result rendering and formatting interfaces
package display

import (
	"context"
	"io"
	"time"
)

// DisplayRenderer handles rendering of test results and progress
type DisplayRenderer interface {
	// RenderResults renders the final test results
	RenderResults(ctx context.Context, results *DisplayResults) error

	// RenderProgress renders real-time progress updates
	RenderProgress(ctx context.Context, progress *ProgressUpdate) error

	// RenderSummary renders a summary of test execution
	RenderSummary(ctx context.Context, summary *TestSummary) error

	// Clear clears the display/terminal
	Clear() error

	// SetConfiguration configures the renderer
	SetConfiguration(config *DisplayConfig) error
}

// ProgressRenderer handles real-time progress display
type ProgressRenderer interface {
	// StartProgress begins progress rendering
	StartProgress(ctx context.Context, total int) error

	// UpdateProgress updates progress with current status
	UpdateProgress(current int, status string) error

	// FinishProgress completes progress rendering
	FinishProgress() error

	// SetSpinner configures the progress spinner
	SetSpinner(spinner *SpinnerConfig) error
}

// ResultFormatter formats test results for display
type ResultFormatter interface {
	// FormatTestResult formats a single test result
	FormatTestResult(result *TestResult) (string, error)

	// FormatPackageResult formats results for an entire package
	FormatPackageResult(result *PackageResult) (string, error)

	// FormatSummary formats the overall test summary
	FormatSummary(summary *TestSummary) (string, error)

	// FormatError formats error messages for display
	FormatError(err error) (string, error)
}

// LayoutManager handles layout and positioning of display elements
type LayoutManager interface {
	// CreateLayout creates a new layout with specified dimensions
	CreateLayout(width, height int) (*Layout, error)

	// UpdateLayout updates the layout with new content
	UpdateLayout(layout *Layout, content *LayoutContent) error

	// RenderLayout renders the layout to the output
	RenderLayout(layout *Layout, output io.Writer) error

	// GetTerminalSize returns the current terminal dimensions
	GetTerminalSize() (width, height int, err error)
}

// ThemeManager manages display themes and styling
type ThemeManager interface {
	// LoadTheme loads a theme by name
	LoadTheme(name string) (*Theme, error)

	// GetCurrentTheme returns the currently active theme
	GetCurrentTheme() *Theme

	// SetTheme sets the active theme
	SetTheme(theme *Theme) error

	// ListAvailableThemes returns available theme names
	ListAvailableThemes() []string
}

// DisplayResults represents test results to be displayed
type DisplayResults struct {
	// Packages contains results for each package
	Packages []*PackageResult

	// Summary contains overall statistics
	Summary *TestSummary

	// Duration is the total execution time
	Duration time.Duration

	// StartTime indicates when tests started
	StartTime time.Time

	// EndTime indicates when tests finished
	EndTime time.Time

	// Configuration used for the test run
	Configuration *TestConfiguration
}

// PackageResult represents results for a single package
type PackageResult struct {
	// Package is the package name/path
	Package string

	// Tests contains individual test results
	Tests []*TestResult

	// Success indicates if all tests passed
	Success bool

	// Duration is the package execution time
	Duration time.Duration

	// Coverage is the coverage percentage
	Coverage float64

	// Output contains raw output for the package
	Output string

	// Error contains any package-level error
	Error error
}

// TestResult represents the result of a single test
type TestResult struct {
	// Name is the test name
	Name string

	// Package is the containing package
	Package string

	// Status is the test status
	Status TestStatus

	// Duration is the test execution time
	Duration time.Duration

	// Output contains test output lines
	Output []string

	// Error contains error details if failed
	Error *TestError

	// Subtests contains any subtest results
	Subtests []*TestResult

	// Parent is the parent test name (for subtests)
	Parent string
}

// TestSummary contains aggregated test statistics
type TestSummary struct {
	// TotalTests is the total number of tests
	TotalTests int

	// PassedTests is the number of passed tests
	PassedTests int

	// FailedTests is the number of failed tests
	FailedTests int

	// SkippedTests is the number of skipped tests
	SkippedTests int

	// TotalDuration is the total execution time
	TotalDuration time.Duration

	// CoveragePercentage is the overall coverage
	CoveragePercentage float64

	// PackageCount is the number of packages tested
	PackageCount int

	// Success indicates if all tests passed
	Success bool
}

// ProgressUpdate represents a progress update during test execution
type ProgressUpdate struct {
	// Current is the current progress value
	Current int

	// Total is the total expected value
	Total int

	// Status is the current status message
	Status string

	// Package is the current package being processed
	Package string

	// Test is the current test being executed
	Test string

	// ElapsedTime is the time elapsed since start
	ElapsedTime time.Duration

	// EstimatedRemaining is the estimated time remaining
	EstimatedRemaining time.Duration
}

// DisplayConfig configures display behavior
type DisplayConfig struct {
	// Output is where to write display output
	Output io.Writer

	// Colors indicates if colors should be used
	Colors bool

	// Theme is the display theme to use
	Theme string

	// Icons indicates the icon set to use
	Icons string

	// Verbose enables verbose output
	Verbose bool

	// ShowProgress enables progress display
	ShowProgress bool

	// ShowCoverage enables coverage display
	ShowCoverage bool

	// TerminalWidth is the terminal width
	TerminalWidth int

	// TerminalHeight is the terminal height
	TerminalHeight int

	// UpdateInterval is how often to update progress
	UpdateInterval time.Duration
}

// Layout represents a display layout
type Layout struct {
	// Width is the layout width
	Width int

	// Height is the layout height
	Height int

	// Sections contains layout sections
	Sections []*LayoutSection

	// Theme is the theme used for rendering
	Theme *Theme
}

// LayoutSection represents a section within a layout
type LayoutSection struct {
	// Name is the section identifier
	Name string

	// X is the horizontal position
	X int

	// Y is the vertical position
	Y int

	// Width is the section width
	Width int

	// Height is the section height
	Height int

	// Content is the section content
	Content string

	// Style is the section styling
	Style *SectionStyle
}

// LayoutContent represents content to be placed in a layout
type LayoutContent struct {
	// Sections contains content for each section
	Sections map[string]string

	// Metadata contains additional layout metadata
	Metadata map[string]interface{}
}

// SectionStyle defines styling for a layout section
type SectionStyle struct {
	// ForegroundColor is the text color
	ForegroundColor string

	// BackgroundColor is the background color
	BackgroundColor string

	// Bold indicates if text should be bold
	Bold bool

	// Italic indicates if text should be italic
	Italic bool

	// Underline indicates if text should be underlined
	Underline bool

	// Border indicates if section should have a border
	Border bool
}

// Theme defines a display theme
type Theme struct {
	// Name is the theme name
	Name string

	// Colors contains color definitions
	Colors map[string]string

	// Styles contains style definitions
	Styles map[string]*SectionStyle

	// Icons contains icon definitions
	Icons map[string]string
}

// SpinnerConfig configures progress spinners
type SpinnerConfig struct {
	// Frames contains spinner animation frames
	Frames []string

	// Interval is the animation interval
	Interval time.Duration

	// Color is the spinner color
	Color string

	// Style is additional styling
	Style *SectionStyle
}

// TestConfiguration represents test execution configuration
type TestConfiguration struct {
	// Packages contains packages to test
	Packages []string

	// Watch indicates if watch mode is enabled
	Watch bool

	// Coverage indicates if coverage is enabled
	Coverage bool

	// Verbose indicates if verbose output is enabled
	Verbose bool

	// Parallel indicates the number of parallel tests
	Parallel int
}

// TestError contains detailed error information
type TestError struct {
	// Message is the error message
	Message string

	// StackTrace contains stack trace lines
	StackTrace []string

	// SourceFile is the source file containing the error
	SourceFile string

	// SourceLine is the line number of the error
	SourceLine int

	// SourceColumn is the column number of the error
	SourceColumn int
}

// TestStatus represents the status of a test
type TestStatus string

const (
	// StatusRunning indicates the test is currently running
	StatusRunning TestStatus = "running"

	// StatusPassed indicates the test passed
	StatusPassed TestStatus = "passed"

	// StatusFailed indicates the test failed
	StatusFailed TestStatus = "failed"

	// StatusSkipped indicates the test was skipped
	StatusSkipped TestStatus = "skipped"
)

// DisplayMode represents different display modes
type DisplayMode string

const (
	// DisplayModeCompact uses compact display format
	DisplayModeCompact DisplayMode = "compact"

	// DisplayModeStandard uses standard display format
	DisplayModeStandard DisplayMode = "standard"

	// DisplayModeVerbose uses verbose display format
	DisplayModeVerbose DisplayMode = "verbose"

	// DisplayModeQuiet uses minimal display format
	DisplayModeQuiet DisplayMode = "quiet"
)
