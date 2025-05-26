// Package config provides application configuration loading components
package config

// AppConfig represents application configuration for the config package.
// This is defined here to avoid circular dependencies with the app package.
type AppConfig struct {
	// Watch configuration
	Watch AppWatchConfig

	// Paths configuration
	Paths AppPathsConfig

	// Visual configuration
	Visual AppVisualConfig

	// Test configuration
	Test AppTestConfig

	// Colors enabled
	Colors bool

	// Verbosity level
	Verbosity int
}

// AppWatchConfig represents watch-specific configuration
type AppWatchConfig struct {
	Enabled        bool
	IgnorePatterns []string
	Debounce       string
	RunOnStart     bool
	ClearOnRerun   bool
}

// AppPathsConfig represents path-specific configuration
type AppPathsConfig struct {
	IncludePatterns []string
	ExcludePatterns []string
}

// AppVisualConfig represents visual/UI configuration
type AppVisualConfig struct {
	Icons         string
	Theme         string
	TerminalWidth int
}

// AppTestConfig represents test execution configuration
type AppTestConfig struct {
	Timeout  string
	Parallel int
	Coverage bool
}

// AppArguments represents parsed command-line arguments
type AppArguments struct {
	Packages         []string
	Watch            bool
	Verbose          bool
	Colors           bool
	Optimized        bool
	OptimizationMode string
}

// AppConfigLoader provides application-specific configuration loading functionality.
// This interface belongs in the config package as it defines config behavior contracts.
type AppConfigLoader interface {
	// LoadFromFile loads configuration from a file
	LoadFromFile(path string) (*AppConfig, error)

	// LoadFromDefaults returns default configuration
	LoadFromDefaults() *AppConfig

	// Merge merges CLI arguments with configuration
	Merge(config *AppConfig, args *AppArguments) *AppConfig

	// Validate validates the final configuration
	Validate(config *AppConfig) error
}

// AppConfigLoaderFactory creates instances of AppConfigLoader with proper dependencies injected.
// This follows the Factory pattern for clean dependency management.
type AppConfigLoaderFactory interface {
	// Create creates a new AppConfigLoader with the specified dependencies
	Create(dependencies AppConfigLoaderDependencies) AppConfigLoader

	// CreateDefault creates a new AppConfigLoader with default dependencies
	CreateDefault() AppConfigLoader
}

// AppConfigLoaderDependencies encapsulates all dependencies needed by the AppConfigLoader.
// This follows dependency injection principles for better testability.
type AppConfigLoaderDependencies struct {
	// CliLoader for loading underlying CLI configuration
	CliLoader ConfigLoader

	// ValidationMode for different validation levels
	ValidationMode ValidationMode
}

// ValidationMode represents different validation levels
type ValidationMode int

const (
	// ValidationModeStrict enforces all validation rules
	ValidationModeStrict ValidationMode = iota

	// ValidationModeLenient allows some validation relaxation
	ValidationModeLenient

	// ValidationModeOff disables validation (for testing)
	ValidationModeOff
)
