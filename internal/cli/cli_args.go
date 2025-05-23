package cli

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

// CLIArgs represents the parsed command line arguments
type CLIArgs struct {
	Watch        bool
	Packages     []string
	TestPattern  string
	Verbosity    int
	FailFast     bool
	Colors       bool
	ConfigFile   string
	Timeout      string
	Parallel     int
	CoverageMode string
}

// ArgParser interface for parsing command line arguments
type ArgParser interface {
	Parse(args []string) (*CLIArgs, error)
	ParseFromCobra(watchFlag, colorFlag, verboseFlag, failFastFlag bool, packages []string, testPattern string) *CLIArgs
}

// DefaultArgParser implements the ArgParser interface
type DefaultArgParser struct{}

// Parse parses command line arguments into CLIArgs structure
func (p *DefaultArgParser) Parse(args []string) (*CLIArgs, error) {
	// First, handle multiple -v flags manually
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

	// Create a new flag set to avoid conflicts with global flags
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
		if level < 0 {
			return nil, errors.New("verbosity level cannot be negative")
		}
		if level > 5 {
			return nil, errors.New("verbosity level too high")
		}
		verbosity = level
	}

	// Determine watch mode
	watch := *watchFlag || *watchLongFlag

	// Determine color mode (default true unless --no-color is specified)
	colors := *colorFlag
	if *noColorFlag {
		colors = false
	}

	// Get test pattern
	pattern := *testPattern
	if *testPatternLong != "" {
		pattern = *testPatternLong
	}

	// Get remaining arguments as packages
	packages := fs.Args()

	return &CLIArgs{
		Watch:        watch,
		Packages:     packages,
		TestPattern:  pattern,
		Verbosity:    verbosity,
		FailFast:     *failFastFlag,
		Colors:       colors,
		ConfigFile:   *configFile,
		Timeout:      *timeout,
		Parallel:     *parallel,
		CoverageMode: *coverage,
	}, nil
}

// ParseFromCobra creates CLIArgs from Cobra command flags
func (p *DefaultArgParser) ParseFromCobra(watchFlag, colorFlag, verboseFlag, failFastFlag bool, packages []string, testPattern string) *CLIArgs {
	verbosity := 0
	if verboseFlag {
		verbosity = 1
	}

	return &CLIArgs{
		Watch:       watchFlag,
		Colors:      colorFlag,
		Verbosity:   verbosity,
		FailFast:    failFastFlag,
		Packages:    packages,
		TestPattern: testPattern,
	}
}

// NewArgParser creates a new argument parser
func NewArgParser() ArgParser {
	return &DefaultArgParser{}
}

// ValidateArgs validates the parsed CLI arguments
func ValidateArgs(args *CLIArgs) error {
	if args.Verbosity < 0 || args.Verbosity > 5 {
		return errors.New("verbosity level must be between 0 and 5")
	}

	if args.Parallel < 0 {
		return errors.New("parallel count cannot be negative")
	}

	// Validate coverage mode if specified
	if args.CoverageMode != "" {
		validModes := []string{"set", "count", "atomic"}
		valid := false
		for _, mode := range validModes {
			if args.CoverageMode == mode {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid coverage mode: %s", args.CoverageMode)
		}
	}

	return nil
}

// GetDefaultArgs returns default CLI arguments
func GetDefaultArgs() *CLIArgs {
	return &CLIArgs{
		Watch:        false,
		Packages:     []string{},
		TestPattern:  "",
		Verbosity:    0,
		FailFast:     false,
		Colors:       true,
		ConfigFile:   "",
		Timeout:      "",
		Parallel:     0,
		CoverageMode: "",
	}
}
