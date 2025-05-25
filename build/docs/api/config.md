package config // import "github.com/newbpydev/go-sentinel/internal/config"


FUNCTIONS

func ConvertPackagesToWatchPaths(packages []string) []string
    convertPackagesToWatchPaths converts Go package patterns to file system
    paths for watching This is a helper function to support Go package syntax in
    watch mode

func ValidateArgs(args *Args) error
    ValidateArgs validates the parsed CLI arguments

func ValidateConfig(config *Config) error
    ValidateConfig validates a configuration for consistency and correctness


TYPES

type ArgParser interface {
	Parse(args []string) (*Args, error)
	ParseFromCobra(watchFlag, colorFlag, verboseFlag, failFastFlag, optimizedFlag bool, packages []string, testPattern, optimizationMode string) *Args
}
    ArgParser interface for parsing command line arguments

func NewArgParser() ArgParser
    NewArgParser creates a new argument parser

type Args struct {
	Colors           bool     `short:"c" long:"color" description:"Use colored output"`
	Verbosity        int      `short:"v" long:"verbosity" description:"Set verbosity level (0-5)" default:"0"`
	Watch            bool     `short:"w" long:"watch" description:"Enable watch mode"`
	Parallel         int      `short:"j" long:"parallel" description:"Number of tests to run in parallel" default:"0"`
	TestPattern      string   `short:"t" long:"test" description:"Run only tests matching pattern"`
	FailFast         bool     `short:"f" long:"fail-fast" description:"Stop on first failure"`
	ConfigFile       string   `long:"config" description:"Path to config file"`
	Timeout          string   `long:"timeout" description:"Timeout for test execution"`
	CoverageMode     string   `long:"coverage" description:"Coverage mode"`
	Optimized        bool     `long:"optimized" description:"Enable optimized test execution with Go's built-in caching"`
	OptimizationMode string   `long:"optimization" description:"Set optimization mode (conservative, balanced, aggressive)"`
	Packages         []string `positional-arg-name:"packages" description:"Packages to test"`
}
    Args represents command line arguments for the go-sentinel CLI tool

func GetDefaultArgs() *Args
    GetDefaultArgs returns default CLI arguments

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
    Config represents the complete configuration for the sentinel CLI

func GetDefaultConfig() *Config
    GetDefaultConfig returns the default configuration

func (c *Config) MergeWithCLIArgs(args *Args) *Config
    MergeWithCLIArgs merges configuration with CLI arguments, with CLI args
    taking precedence

type ConfigLoader interface {
	LoadFromFile(path string) (*Config, error)
	LoadFromDefault() (*Config, error)
}
    ConfigLoader interface for loading configuration from files

func NewConfigLoader() ConfigLoader
    NewConfigLoader creates a new configuration loader

type DefaultArgParser struct{}
    DefaultArgParser implements the ArgParser interface

func (p *DefaultArgParser) Parse(args []string) (*Args, error)
    Parse parses command line arguments into Args structure

func (p *DefaultArgParser) ParseFromCobra(watchFlag, colorFlag, verboseFlag, failFastFlag, optimizedFlag bool, packages []string, testPattern, optimizationMode string) *Args
    ParseFromCobra creates Args from Cobra command flags

type DefaultConfigLoader struct{}
    DefaultConfigLoader implements the ConfigLoader interface

func (l *DefaultConfigLoader) LoadFromDefault() (*Config, error)
    LoadFromDefault loads the default configuration file (sentinel.config.json)

func (l *DefaultConfigLoader) LoadFromFile(path string) (*Config, error)
    LoadFromFile loads configuration from a specified file path

type PathsConfig struct {
	IncludePatterns []string `json:"includePatterns"`
	ExcludePatterns []string `json:"excludePatterns"`
}
    PathsConfig contains configuration for path patterns

type VisualConfig struct {
	Colors bool   `json:"colors"`
	Icons  string `json:"icons"`
	Theme  string `json:"theme"`
}
    VisualConfig contains configuration for visual appearance

type WatchConfig struct {
	Enabled        bool          `json:"enabled"`
	Debounce       time.Duration `json:"debounce"`
	IgnorePatterns []string      `json:"ignorePatterns"`
	ClearOnRerun   bool          `json:"clearOnRerun"`
	RunOnStart     bool          `json:"runOnStart"`
}
    WatchConfig contains configuration for watch mode behavior

