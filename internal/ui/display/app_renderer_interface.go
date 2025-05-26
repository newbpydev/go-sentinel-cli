// Package display provides application-specific display rendering components
package display

import (
	"context"
	"io"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

// AppConfig represents application configuration for the app renderer.
// This is defined here to avoid circular dependencies with the app package.
type AppConfig struct {
	// Colors indicates whether colored output is enabled
	Colors bool

	// Visual contains visual display configuration
	Visual struct {
		// Icons setting (none, simple, rich)
		Icons string

		// TerminalWidth for display formatting
		TerminalWidth int
	}
}

// AppRenderer provides application-specific display rendering functionality.
// This interface belongs in the UI package as it defines UI behavior contracts.
type AppRenderer interface {
	// RenderResults renders test execution results to the configured output
	RenderResults(ctx context.Context) error

	// RenderTestResults renders individual test results with proper formatting
	RenderTestResults(ctx context.Context, results []*models.TestResult) error

	// RenderIncrementalResults renders results incrementally for watch mode
	RenderIncrementalResults(ctx context.Context, results interface{}) error

	// RenderFailedTests renders only the failed test results with detailed output
	RenderFailedTests(ctx context.Context, failedTests []*models.TestResult) error

	// SetConfiguration configures the renderer with application settings
	SetConfiguration(config *AppConfig) error

	// SetWriter sets the output writer for rendering
	SetWriter(writer io.Writer)

	// GetWriter returns the current output writer
	GetWriter() io.Writer
}

// AppRendererFactory creates instances of AppRenderer with proper dependencies injected.
// This follows the Factory pattern for clean dependency management.
type AppRendererFactory interface {
	// Create creates a new AppRenderer with the specified dependencies
	Create(dependencies AppRendererDependencies) AppRenderer

	// CreateDefault creates a new AppRenderer with default dependencies
	CreateDefault() AppRenderer
}

// AppRendererDependencies encapsulates all dependencies needed by the AppRenderer.
// This follows dependency injection principles for better testability.
type AppRendererDependencies struct {
	// Writer for output (defaults to os.Stdout)
	Writer io.Writer

	// ColorFormatter for colorized output
	ColorFormatter FormatterInterface

	// IconProvider for displaying icons
	IconProvider IconProviderInterface

	// TestDisplay for individual test rendering
	TestDisplay TestDisplayInterface

	// SuiteDisplay for test suite rendering
	SuiteDisplay SuiteDisplayInterface

	// SummaryDisplay for summary rendering
	SummaryDisplay SummaryDisplayInterface

	// FailureDisplay for failure rendering
	FailureDisplay FailureDisplayInterface

	// Cache for result caching (optional)
	Cache CacheInterface

	// TerminalWidth for layout calculations
	TerminalWidth int
}

// CacheInterface represents the cache dependency for the renderer.
// This avoids direct import of cache package from UI package.
type CacheInterface interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Delete(key string)
	Clear()
}

// FormatterInterface represents color formatting capabilities.
// This is already defined in the colors package.
type FormatterInterface interface {
	FormatSuccess(text string) string
	FormatError(text string) string
	FormatWarning(text string) string
	FormatInfo(text string) string
	FormatMuted(text string) string
	IsColorEnabled() bool
}

// IconProviderInterface represents icon provision capabilities.
// This is already defined in the colors package.
type IconProviderInterface interface {
	Success() string
	Error() string
	Warning() string
	Info() string
	Running() string
	Skipped() string
}
