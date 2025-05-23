package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// Config represents the complete configuration for the sentinel CLI
type Config struct {
	Colors      bool          `json:"colors"`
	Verbosity   int           `json:"verbosity"`
	Parallel    int           `json:"parallel"`
	Timeout     time.Duration `json:"timeout"`
	TestPattern string        `json:"testPattern"`
	TestCommand string        `json:"testCommand"`
	Visual      VisualConfig  `json:"visual"`
	Paths       PathsConfig   `json:"paths"`
	Watch       WatchConfig   `json:"watch"`

	// Legacy fields for backward compatibility
	Icons string `json:"icons"`
}

// VisualConfig contains configuration for visual appearance
type VisualConfig struct {
	Colors bool   `json:"colors"`
	Icons  string `json:"icons"`
	Theme  string `json:"theme"`
}

// PathsConfig contains configuration for path patterns
type PathsConfig struct {
	IncludePatterns []string `json:"includePatterns"`
	ExcludePatterns []string `json:"excludePatterns"`
}

// WatchConfig contains configuration for watch mode behavior
type WatchConfig struct {
	Enabled        bool          `json:"enabled"`
	Debounce       time.Duration `json:"debounce"`
	IgnorePatterns []string      `json:"ignorePatterns"`
	ClearOnRerun   bool          `json:"clearOnRerun"`
	RunOnStart     bool          `json:"runOnStart"`
}

// ConfigLoader interface for loading configuration from files
type ConfigLoader interface {
	LoadFromFile(path string) (*Config, error)
	LoadFromDefault() (*Config, error)
}

// DefaultConfigLoader implements the ConfigLoader interface
type DefaultConfigLoader struct{}

// configFileData represents the JSON structure for configuration files
type configFileData struct {
	Colors          *bool    `json:"colors"`
	Icons           string   `json:"icons"`
	Theme           string   `json:"theme"`
	WatchMode       *bool    `json:"watchMode"`
	Verbosity       *int     `json:"verbosity"`
	Timeout         string   `json:"timeout"`
	IncludePatterns []string `json:"includePatterns"`
	ExcludePatterns []string `json:"excludePatterns"`
	WatchIgnore     []string `json:"watchIgnore"`
	WatchDebounce   string   `json:"watchDebounce"`
	ClearOnRerun    *bool    `json:"clearOnRerun"`
	RunOnStart      *bool    `json:"runOnStart"`
	TestCommand     string   `json:"testCommand"`
	Parallel        *int     `json:"parallel"`
}

// LoadFromFile loads configuration from a specified file path
func (l *DefaultConfigLoader) LoadFromFile(path string) (*Config, error) {
	// If file doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return GetDefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var fileData configFileData
	if err := json.Unmarshal(data, &fileData); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	config, err := l.parseConfigData(&fileData)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// LoadFromDefault loads the default configuration file (sentinel.config.json)
func (l *DefaultConfigLoader) LoadFromDefault() (*Config, error) {
	return l.LoadFromFile("sentinel.config.json")
}

// parseConfigData converts the file data structure to internal Config structure
func (l *DefaultConfigLoader) parseConfigData(data *configFileData) (*Config, error) {
	config := GetDefaultConfig()

	// Parse basic settings
	if data.Colors != nil {
		config.Colors = *data.Colors
		config.Visual.Colors = *data.Colors
	}

	if data.Verbosity != nil {
		if *data.Verbosity < 0 || *data.Verbosity > 5 {
			return nil, errors.New("verbosity level must be between 0 and 5")
		}
		config.Verbosity = *data.Verbosity
	}

	if data.Parallel != nil {
		if *data.Parallel < 0 {
			return nil, errors.New("parallel count cannot be negative")
		}
		config.Parallel = *data.Parallel
	}

	// Parse timeout
	if data.Timeout != "" {
		timeout, err := time.ParseDuration(data.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout format: %w", err)
		}
		config.Timeout = timeout
	}

	// Parse visual configuration
	if data.Icons != "" {
		validIcons := []string{"unicode", "ascii", "minimal", "none"}
		valid := false
		for _, icon := range validIcons {
			if data.Icons == icon {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("invalid icons type: %s (must be one of: unicode, ascii, minimal, none)", data.Icons)
		}
		config.Icons = data.Icons
		config.Visual.Icons = data.Icons
	}

	if data.Theme != "" {
		config.Visual.Theme = data.Theme
	}

	// Parse paths configuration
	if len(data.IncludePatterns) > 0 {
		config.Paths.IncludePatterns = data.IncludePatterns
	}

	if len(data.ExcludePatterns) > 0 {
		config.Paths.ExcludePatterns = data.ExcludePatterns
	}

	// Parse watch configuration
	if data.WatchMode != nil {
		config.Watch.Enabled = *data.WatchMode
	}

	if len(data.WatchIgnore) > 0 {
		config.Watch.IgnorePatterns = data.WatchIgnore
	}

	if data.WatchDebounce != "" {
		debounce, err := time.ParseDuration(data.WatchDebounce)
		if err != nil {
			return nil, fmt.Errorf("invalid watch debounce format: %w", err)
		}
		config.Watch.Debounce = debounce
	}

	if data.ClearOnRerun != nil {
		config.Watch.ClearOnRerun = *data.ClearOnRerun
	}

	if data.RunOnStart != nil {
		config.Watch.RunOnStart = *data.RunOnStart
	}

	// Parse test command
	if data.TestCommand != "" {
		config.TestCommand = data.TestCommand
	}

	return config, nil
}

// MergeWithCLIArgs merges configuration with CLI arguments, with CLI args taking precedence
func (c *Config) MergeWithCLIArgs(args *CLIArgs) *Config {
	merged := &Config{
		// Copy config values
		Colors:      c.Colors,
		Verbosity:   c.Verbosity,
		Parallel:    c.Parallel,
		Timeout:     c.Timeout,
		TestPattern: c.TestPattern,
		TestCommand: c.TestCommand,
		Visual:      c.Visual,
		Paths:       c.Paths,
		Watch:       c.Watch,
		Icons:       c.Icons,
	}

	// Override with CLI args
	if args.Colors != c.Colors { // CLI explicitly set colors
		merged.Colors = args.Colors
		merged.Visual.Colors = args.Colors
	}

	if args.Verbosity > 0 {
		merged.Verbosity = args.Verbosity
	}

	if args.Parallel > 0 {
		merged.Parallel = args.Parallel
	}

	if args.Watch {
		merged.Watch.Enabled = true
	}

	if args.TestPattern != "" {
		merged.TestPattern = args.TestPattern
	}

	if args.FailFast {
		// FailFast affects behavior but doesn't have a direct config equivalent
		// Could be added to config in the future
	}

	if args.ConfigFile != "" {
		// ConfigFile is used for loading, not stored in final config
	}

	if args.Timeout != "" {
		if timeout, err := time.ParseDuration(args.Timeout); err == nil {
			merged.Timeout = timeout
		}
	}

	if args.CoverageMode != "" {
		// CoverageMode could be added to config structure in the future
	}

	return merged
}

// GetDefaultConfig returns the default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Colors:      true,
		Verbosity:   0,
		Parallel:    0, // 0 means use Go's default
		Timeout:     30 * time.Second,
		TestPattern: "",
		TestCommand: "go test",
		Visual: VisualConfig{
			Colors: true,
			Icons:  "unicode",
			Theme:  "dark",
		},
		Paths: PathsConfig{
			IncludePatterns: []string{},
			ExcludePatterns: []string{},
		},
		Watch: WatchConfig{
			Enabled:        false,
			Debounce:       250 * time.Millisecond,
			IgnorePatterns: []string{"*.log", "*.tmp", ".git/*", "node_modules/*"},
			ClearOnRerun:   true,
			RunOnStart:     true,
		},
		Icons: "unicode", // Legacy field
	}
}

// ValidateConfig validates a configuration for consistency and correctness
func ValidateConfig(config *Config) error {
	if config.Verbosity < 0 || config.Verbosity > 5 {
		return errors.New("verbosity level must be between 0 and 5")
	}

	if config.Parallel < 0 {
		return errors.New("parallel count cannot be negative")
	}

	if config.Timeout <= 0 {
		return errors.New("timeout must be positive")
	}

	validIcons := []string{"unicode", "ascii", "minimal", "none"}
	iconValid := false
	for _, icon := range validIcons {
		if config.Visual.Icons == icon {
			iconValid = true
			break
		}
	}
	if !iconValid {
		return fmt.Errorf("invalid icons type: %s", config.Visual.Icons)
	}

	validThemes := []string{"dark", "light", "auto"}
	themeValid := false
	for _, theme := range validThemes {
		if config.Visual.Theme == theme {
			themeValid = true
			break
		}
	}
	if !themeValid {
		return fmt.Errorf("invalid theme: %s", config.Visual.Theme)
	}

	if config.Watch.Debounce < 0 {
		return errors.New("watch debounce cannot be negative")
	}

	return nil
}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader() ConfigLoader {
	return &DefaultConfigLoader{}
}
