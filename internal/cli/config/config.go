package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/newbpydev/go-sentinel/internal/cli/core"
)

// CLIArgs represents command line arguments for the go-sentinel CLI tool
type CLIArgs struct {
	// Core options
	Watch     bool `description:"Enable watch mode"`
	Verbose   bool `description:"Enable verbose output"`
	Colors    bool `description:"Use colored output"`
	FailFast  bool `description:"Stop on first failure"`
	Optimized bool `description:"Enable optimized test execution with Go's built-in caching"`

	// Advanced options
	Verbosity        int    `description:"Set verbosity level (0-5)"`
	Parallel         int    `description:"Number of tests to run in parallel"`
	TestPattern      string `description:"Run only tests matching pattern"`
	ConfigFile       string `description:"Path to config file"`
	Timeout          string `description:"Timeout for test execution"`
	CoverageMode     string `description:"Coverage mode"`
	OptimizationMode string `description:"Set optimization mode (conservative, balanced, aggressive)"`

	// Positional arguments
	Packages []string `description:"Packages to test"`
}

// Configuration represents the complete configuration combining CLI args and file config
type Configuration struct {
	// Inherit from core.Config
	core.Config

	// Additional CLI-specific settings
	CLIArgs     *CLIArgs
	ConfigPath  string
	IsWatchMode bool
}

// CLIParser handles command line argument parsing
type CLIParser struct{}

// NewCLIParser creates a new CLI argument parser
func NewCLIParser() *CLIParser {
	return &CLIParser{}
}

// Parse parses command line arguments into CLIArgs structure
func (p *CLIParser) Parse(args []string) (*CLIArgs, error) {
	// Handle multiple -v flags manually for verbosity
	verbosity := 0
	filteredArgs := []string{}

	for _, arg := range args {
		if strings.HasPrefix(arg, "-v") && len(arg) > 2 && !strings.Contains(arg, "=") {
			// Handle -vvv style flags
			verbosity = len(arg) - 1
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}

	// Create a new flag set to avoid conflicts
	fs := flag.NewFlagSet("go-sentinel", flag.ContinueOnError)

	// Define flags
	watchFlag := fs.Bool("w", false, "Enable watch mode")
	watchLongFlag := fs.Bool("watch", false, "Enable watch mode")
	colorFlag := fs.Bool("color", true, "Enable colored output")
	noColorFlag := fs.Bool("no-color", false, "Disable colored output")
	verboseFlag := fs.Bool("v", false, "Enable verbose output")
	verboseLongFlag := fs.Bool("verbose", false, "Enable verbose output")
	failFastFlag := fs.Bool("fail-fast", false, "Stop on first failure")
	testPattern := fs.String("t", "", "Run only tests matching pattern")
	testPatternLong := fs.String("test", "", "Run only tests matching pattern")
	verbosityLevel := fs.String("verbosity", "0", "Set verbosity level (0-5)")
	configFile := fs.String("config", "", "Path to configuration file")
	timeout := fs.String("timeout", "", "Test timeout duration")
	parallel := fs.Int("parallel", 0, "Number of parallel test executions")
	coverage := fs.String("covermode", "", "Set coverage mode")
	optimized := fs.Bool("optimized", false, "Enable optimized test execution")
	optimizationMode := fs.String("optimization", "", "Set optimization mode (conservative, balanced, aggressive)")

	// Parse the filtered arguments
	err := fs.Parse(filteredArgs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Handle verbosity level parsing
	if *verboseLongFlag || *verboseFlag {
		verbosity = 1
	}

	// Parse verbosity level if specified
	if *verbosityLevel != "0" {
		level, err := strconv.Atoi(*verbosityLevel)
		if err != nil {
			return nil, errors.New("invalid verbosity level")
		}
		if level < 0 || level > 5 {
			return nil, errors.New("verbosity level must be between 0 and 5")
		}
		verbosity = level
	}

	// Determine watch mode
	watch := *watchFlag || *watchLongFlag

	// Determine color mode (default true unless --no-color is specified)
	colors := *colorFlag && !*noColorFlag

	// Get test pattern
	pattern := *testPattern
	if *testPatternLong != "" {
		pattern = *testPatternLong
	}

	// Get remaining arguments as packages
	packages := fs.Args()

	return &CLIArgs{
		Watch:            watch,
		Verbose:          verbosity > 0,
		Colors:           colors,
		FailFast:         *failFastFlag,
		Optimized:        *optimized,
		Verbosity:        verbosity,
		Parallel:         *parallel,
		TestPattern:      pattern,
		ConfigFile:       *configFile,
		Timeout:          *timeout,
		CoverageMode:     *coverage,
		OptimizationMode: *optimizationMode,
		Packages:         packages,
	}, nil
}

// ConfigLoader handles loading configuration from files
type ConfigLoader struct{}

// NewConfigLoader creates a new configuration loader
func NewConfigLoader() *ConfigLoader {
	return &ConfigLoader{}
}

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
func (l *ConfigLoader) LoadFromFile(path string) (*core.Config, error) {
	// If file doesn't exist, return default config
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return l.GetDefaultConfig(), nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var fileData configFileData
	if unmarshalErr := json.Unmarshal(data, &fileData); unmarshalErr != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", unmarshalErr)
	}

	config, err := l.parseConfigData(&fileData)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

// parseConfigData converts the file data structure to core.Config structure
func (l *ConfigLoader) parseConfigData(data *configFileData) (*core.Config, error) {
	config := l.GetDefaultConfig()

	// Parse basic settings
	if data.Colors != nil {
		config.NoColor = !*data.Colors
	}

	if data.Verbosity != nil {
		if *data.Verbosity < 0 || *data.Verbosity > 5 {
			return nil, errors.New("verbosity level must be between 0 and 5")
		}
		config.Verbose = *data.Verbosity > 0
		config.Debug = *data.Verbosity > 2
	}

	if data.Parallel != nil {
		if *data.Parallel < 0 {
			return nil, errors.New("parallel count cannot be negative")
		}
		config.MaxConcurrency = *data.Parallel
	}

	// Parse timeout
	if data.Timeout != "" {
		timeout, err := time.ParseDuration(data.Timeout)
		if err != nil {
			return nil, fmt.Errorf("invalid timeout format: %w", err)
		}
		config.TestTimeout = timeout
	}

	// Parse watch configuration
	if data.WatchMode != nil {
		config.WatchMode = *data.WatchMode
	}

	if data.WatchDebounce != "" {
		debounce, err := time.ParseDuration(data.WatchDebounce)
		if err != nil {
			return nil, fmt.Errorf("invalid watch debounce format: %w", err)
		}
		config.DebounceInterval = debounce
	}

	// Parse watch paths and patterns
	if len(data.WatchIgnore) > 0 {
		config.WatchPaths = append(config.WatchPaths, data.WatchIgnore...)
	}

	// Handle icons setting
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
		config.NoIcons = (data.Icons == "none")
	}

	return config, nil
}

// GetDefaultConfig returns a default configuration using core types
func (l *ConfigLoader) GetDefaultConfig() *core.Config {
	return &core.Config{
		// General settings
		Verbose: false,
		Debug:   false,
		NoColor: false,
		NoIcons: false,

		// Watch mode settings
		WatchMode:        false,
		WatchPaths:       []string{},
		DebounceInterval: 300 * time.Millisecond,

		// Test execution settings
		UseCache:       true,
		CacheStrategy:  "aggressive",
		MaxConcurrency: 4,
		TestTimeout:    30 * time.Second,
		FailFast:       false,

		// Output settings
		OutputFormat: "structured",
		ShowProgress: true,
		ShowSummary:  true,

		// Test filtering
		TestPattern:    "",
		ExcludePattern: "",
		TestFunctions:  []string{},
	}
}

// ConfigManager manages the complete configuration state
type ConfigManager struct {
	cliParser    *CLIParser
	configLoader *ConfigLoader
	config       *Configuration
}

// NewConfigManager creates a new configuration manager
func NewConfigManager() *ConfigManager {
	return &ConfigManager{
		cliParser:    NewCLIParser(),
		configLoader: NewConfigLoader(),
	}
}

// LoadConfiguration loads and merges configuration from CLI args and config file
func (m *ConfigManager) LoadConfiguration(args []string) (*Configuration, error) {
	// Parse CLI arguments first
	cliArgs, err := m.cliParser.Parse(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CLI arguments: %w", err)
	}

	// Determine config file path
	configPath := cliArgs.ConfigFile
	if configPath == "" {
		configPath = "sentinel.config.json"
	}

	// Load file configuration
	fileConfig, err := m.configLoader.LoadFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	// Merge CLI args with file config
	finalConfig := m.mergeCLIWithFileConfig(cliArgs, fileConfig)

	// Validate the final configuration
	if err := m.validateConfiguration(finalConfig); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	m.config = &Configuration{
		Config:      *finalConfig,
		CLIArgs:     cliArgs,
		ConfigPath:  configPath,
		IsWatchMode: cliArgs.Watch,
	}

	return m.config, nil
}

// mergeCLIWithFileConfig merges CLI arguments with file configuration (CLI takes precedence)
func (m *ConfigManager) mergeCLIWithFileConfig(cliArgs *CLIArgs, fileConfig *core.Config) *core.Config {
	config := *fileConfig // Copy file config

	// CLI args override file config
	if cliArgs.Verbose {
		config.Verbose = true
	}

	if !cliArgs.Colors {
		config.NoColor = true
	}

	if cliArgs.Watch {
		config.WatchMode = true
	}

	if cliArgs.FailFast {
		config.FailFast = true
	}

	if cliArgs.TestPattern != "" {
		config.TestPattern = cliArgs.TestPattern
	}

	if cliArgs.Parallel > 0 {
		config.MaxConcurrency = cliArgs.Parallel
	}

	if cliArgs.Timeout != "" {
		if timeout, err := time.ParseDuration(cliArgs.Timeout); err == nil {
			config.TestTimeout = timeout
		}
	}

	// Handle optimization settings
	if cliArgs.Optimized {
		config.UseCache = true
		if cliArgs.OptimizationMode != "" {
			config.CacheStrategy = cliArgs.OptimizationMode
		}
	}

	return &config
}

// validateConfiguration validates the final merged configuration
func (m *ConfigManager) validateConfiguration(config *core.Config) error {
	if config.MaxConcurrency < 0 {
		return errors.New("max concurrency cannot be negative")
	}

	if config.TestTimeout <= 0 {
		return errors.New("test timeout must be positive")
	}

	if config.DebounceInterval < 0 {
		return errors.New("debounce interval cannot be negative")
	}

	// Validate cache strategy
	if config.CacheStrategy != "" {
		validStrategies := []string{"conservative", "balanced", "aggressive"}
		valid := false
		for _, strategy := range validStrategies {
			if config.CacheStrategy == strategy {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid cache strategy: %s (valid options: conservative, balanced, aggressive)", config.CacheStrategy)
		}
	}

	return nil
}

// GetConfiguration returns the current configuration
func (m *ConfigManager) GetConfiguration() *Configuration {
	return m.config
}

// GetCoreConfig returns just the core configuration
func (m *ConfigManager) GetCoreConfig() *core.Config {
	if m.config == nil {
		return m.configLoader.GetDefaultConfig()
	}
	return &m.config.Config
}
