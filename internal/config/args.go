package config

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

// Args represents command line arguments for the go-sentinel CLI tool
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

// ArgParser interface for parsing command line arguments
type ArgParser interface {
	Parse(args []string) (*Args, error)
	ParseFromCobra(watchFlag, colorFlag, verboseFlag, failFastFlag, optimizedFlag bool, packages []string, testPattern, optimizationMode string) *Args
}

// DefaultArgParser implements the ArgParser interface
type DefaultArgParser struct{}

// Parse parses command line arguments into Args structure
func (p *DefaultArgParser) Parse(args []string) (*Args, error) {
	// Strip known command names from the beginning of arguments
	filteredArgs := p.stripCommandName(args)

	// Handle multiple -v flags manually
	verbosity, finalArgs := p.extractVerbosityFlags(filteredArgs)

	// Create flag set and define flags
	fs, flagValues := p.createFlagSet()

	// Parse the filtered arguments
	if err := fs.Parse(finalArgs); err != nil {
		return nil, fmt.Errorf("failed to parse arguments: %w", err)
	}

	// Build and validate Args structure
	parsedArgs, err := p.buildArgsFromFlags(flagValues, verbosity, fs.Args())
	if err != nil {
		return nil, err
	}

	return parsedArgs, nil
}

// stripCommandName removes known command names from the beginning of arguments
func (p *DefaultArgParser) stripCommandName(args []string) []string {
	if len(args) == 0 {
		return args
	}

	// Known command names that should be stripped
	knownCommands := []string{"run", "demo", "help", "version"}

	// Check if first argument is a known command
	for _, cmd := range knownCommands {
		if args[0] == cmd {
			// Return arguments without the command name
			if len(args) > 1 {
				return args[1:]
			}
			return []string{}
		}
	}

	// No known command found, return original arguments
	return args
}

// extractVerbosityFlags handles multiple -v flags and returns verbosity level and filtered args
func (p *DefaultArgParser) extractVerbosityFlags(args []string) (int, []string) {
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

	return verbosity, filteredArgs
}

// flagValues holds pointers to all flag values for easy access
type flagValues struct {
	watchFlag        *bool
	watchLongFlag    *bool
	colorFlag        *bool
	noColorFlag      *bool
	verboseFlag      *bool
	verboseLongFlag  *bool
	failFastFlag     *bool
	testPattern      *string
	testPatternLong  *string
	verbosityLevel   *string
	configFile       *string
	timeout          *string
	parallel         *int
	coverage         *string
	optimized        *bool
	optimizationMode *string
}

// createFlagSet creates a new flag set with all required flags and returns flag values
func (p *DefaultArgParser) createFlagSet() (*flag.FlagSet, *flagValues) {
	fs := flag.NewFlagSet("go-sentinel", flag.ContinueOnError)

	flags := &flagValues{
		watchFlag:        fs.Bool("w", false, "Enable watch mode"),
		watchLongFlag:    fs.Bool("watch", false, "Enable watch mode"),
		colorFlag:        fs.Bool("color", true, "Enable colored output"),
		noColorFlag:      fs.Bool("no-color", false, "Disable colored output"),
		verboseFlag:      fs.Bool("v", false, "Enable verbose output"),
		verboseLongFlag:  fs.Bool("verbose", false, "Enable verbose output"),
		failFastFlag:     fs.Bool("fail-fast", false, "Stop on first failure"),
		testPattern:      fs.String("t", "", "Run only tests matching pattern"),
		testPatternLong:  fs.String("test", "", "Run only tests matching pattern"),
		verbosityLevel:   fs.String("verbosity", "0", "Set verbosity level (0-5)"),
		configFile:       fs.String("config", "", "Path to configuration file"),
		timeout:          fs.String("timeout", "", "Test timeout duration"),
		parallel:         fs.Int("parallel", 0, "Number of parallel test executions"),
		coverage:         fs.String("covermode", "", "Set coverage mode"),
		optimized:        fs.Bool("optimized", false, "Enable optimized test execution with Go's built-in caching"),
		optimizationMode: fs.String("optimization", "", "Set optimization mode (conservative, balanced, aggressive)"),
	}

	return fs, flags
}

// buildArgsFromFlags constructs Args structure from parsed flag values
func (p *DefaultArgParser) buildArgsFromFlags(flags *flagValues, baseVerbosity int, packages []string) (*Args, error) {
	// Handle verbosity level parsing
	verbosity := baseVerbosity
	if *flags.verboseLongFlag || *flags.verboseFlag {
		verbosity = 1
	}

	// Parse verbosity level if specified
	if *flags.verbosityLevel != "0" {
		level, err := strconv.Atoi(*flags.verbosityLevel)
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

	// Determine boolean flags
	watch := *flags.watchFlag || *flags.watchLongFlag
	colors := *flags.colorFlag && !*flags.noColorFlag

	// Get test pattern (prefer long form)
	pattern := *flags.testPattern
	if *flags.testPatternLong != "" {
		pattern = *flags.testPatternLong
	}

	return &Args{
		Watch:            watch,
		Packages:         packages,
		TestPattern:      pattern,
		Verbosity:        verbosity,
		FailFast:         *flags.failFastFlag,
		Colors:           colors,
		ConfigFile:       *flags.configFile,
		Timeout:          *flags.timeout,
		Parallel:         *flags.parallel,
		CoverageMode:     *flags.coverage,
		Optimized:        *flags.optimized,
		OptimizationMode: *flags.optimizationMode,
	}, nil
}

// ParseFromCobra creates Args from Cobra command flags
func (p *DefaultArgParser) ParseFromCobra(watchFlag, colorFlag, verboseFlag, failFastFlag, optimizedFlag bool, packages []string, testPattern, optimizationMode string) *Args {
	verbosity := 0
	if verboseFlag {
		verbosity = 1
	}

	return &Args{
		Watch:            watchFlag,
		Colors:           colorFlag,
		Verbosity:        verbosity,
		FailFast:         failFastFlag,
		Packages:         packages,
		TestPattern:      testPattern,
		Optimized:        optimizedFlag,
		OptimizationMode: optimizationMode,
	}
}

// NewArgParser creates a new argument parser
func NewArgParser() ArgParser {
	return &DefaultArgParser{}
}

// ValidateArgs validates the parsed CLI arguments
func ValidateArgs(args *Args) error {
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

	// Validate optimization mode if specified
	if args.OptimizationMode != "" {
		validModes := []string{"conservative", "balanced", "aggressive"}
		valid := false
		for _, mode := range validModes {
			if args.OptimizationMode == mode {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid optimization mode: %s (valid options: conservative, balanced, aggressive)", args.OptimizationMode)
		}
	}

	return nil
}

// GetDefaultArgs returns default CLI arguments
func GetDefaultArgs() *Args {
	return &Args{
		Watch:            false,
		Packages:         []string{},
		TestPattern:      "",
		Verbosity:        0,
		FailFast:         false,
		Colors:           true,
		ConfigFile:       "",
		Timeout:          "",
		Parallel:         0,
		CoverageMode:     "",
		Optimized:        false,
		OptimizationMode: "",
	}
}
