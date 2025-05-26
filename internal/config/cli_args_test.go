package config

import (
	"strings"
	"testing"
)

func TestCLIArgs_ParseWatchFlag(t *testing.T) {
	parser := &DefaultArgParser{}

	// Test watch flag short form
	cliArgs, err := parser.Parse([]string{"-w"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cliArgs.Watch {
		t.Errorf("expected watch=true, got watch=false")
	}

	// Test watch flag long form
	cliArgs, err = parser.Parse([]string{"--watch"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cliArgs.Watch {
		t.Errorf("expected watch=true, got watch=false")
	}

	// Test no watch flag
	cliArgs, err = parser.Parse([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cliArgs.Watch {
		t.Errorf("expected watch=false, got watch=true")
	}
}

func TestCLIArgs_ParsePackagePatterns(t *testing.T) {
	parser := &DefaultArgParser{}

	// Test single package
	cliArgs, err := parser.Parse([]string{"./internal"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cliArgs.Packages) != 1 || cliArgs.Packages[0] != "./internal" {
		t.Errorf("expected packages=[./internal], got packages=%v", cliArgs.Packages)
	}

	// Test multiple packages
	cliArgs, err = parser.Parse([]string{"./internal", "./cmd"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cliArgs.Packages) != 2 {
		t.Errorf("expected 2 packages, got %d", len(cliArgs.Packages))
	}
}

func TestCLIArgs_ParseTestNamePattern(t *testing.T) {
	parser := &DefaultArgParser{}

	// Test pattern short form
	cliArgs, err := parser.Parse([]string{"-t", "TestFunction"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cliArgs.TestPattern != "TestFunction" {
		t.Errorf("expected test pattern='TestFunction', got '%s'", cliArgs.TestPattern)
	}

	// Test pattern long form
	cliArgs, err = parser.Parse([]string{"--test", "TestSuite"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cliArgs.TestPattern != "TestSuite" {
		t.Errorf("expected test pattern='TestSuite', got '%s'", cliArgs.TestPattern)
	}
}

func TestCLIArgs_ParseVerbosityLevel(t *testing.T) {
	parser := &DefaultArgParser{}

	// Test single verbose flag
	cliArgs, err := parser.Parse([]string{"-v"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cliArgs.Verbosity != 1 {
		t.Errorf("expected verbosity=1, got %d", cliArgs.Verbosity)
	}

	// Test multiple verbose flags
	cliArgs, err = parser.Parse([]string{"-vvv"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cliArgs.Verbosity != 3 {
		t.Errorf("expected verbosity=3, got %d", cliArgs.Verbosity)
	}

	// Test verbosity level with number
	cliArgs, err = parser.Parse([]string{"--verbosity=2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cliArgs.Verbosity != 2 {
		t.Errorf("expected verbosity=2, got %d", cliArgs.Verbosity)
	}
}

func TestCLIArgs_ParseFromCobra(t *testing.T) {
	parser := &DefaultArgParser{}

	cliArgs := parser.ParseFromCobra(true, true, false, false, true, []string{"./internal"}, "TestUnit", "aggressive")

	if !cliArgs.Watch {
		t.Errorf("expected watch=true, got watch=false")
	}
	if !cliArgs.Colors {
		t.Errorf("expected colors=true, got colors=false")
	}
	if cliArgs.Verbosity != 0 {
		t.Errorf("expected verbosity=0, got verbosity=%d", cliArgs.Verbosity)
	}
	if cliArgs.TestPattern != "TestUnit" {
		t.Errorf("expected test pattern='TestUnit', got '%s'", cliArgs.TestPattern)
	}
	if !cliArgs.Optimized {
		t.Errorf("expected optimized=true, got optimized=false")
	}
	if cliArgs.OptimizationMode != "aggressive" {
		t.Errorf("expected optimization mode='aggressive', got '%s'", cliArgs.OptimizationMode)
	}
}

func TestCLIArgs_ErrorHandling(t *testing.T) {
	parser := &DefaultArgParser{}

	// Test invalid verbosity level
	_, err := parser.Parse([]string{"--verbosity=invalid"})
	if err == nil {
		t.Errorf("expected error but got none")
	}

	// Test negative verbosity
	_, err = parser.Parse([]string{"--verbosity=-1"})
	if err == nil {
		t.Errorf("expected error but got none")
	}

	// Test verbosity too high
	_, err = parser.Parse([]string{"--verbosity=10"})
	if err == nil {
		t.Errorf("expected error but got none")
	}
}

// TestNewArgParser tests the creation of a new argument parser
func TestNewArgParser(t *testing.T) {
	t.Parallel()

	parser := NewArgParser()
	if parser == nil {
		t.Fatal("NewArgParser should not return nil")
	}

	// Verify interface compliance
	_, ok := parser.(ArgParser)
	if !ok {
		t.Fatal("NewArgParser should return ArgParser interface")
	}

	// Verify it's the correct implementation
	_, ok = parser.(*DefaultArgParser)
	if !ok {
		t.Fatal("NewArgParser should return *DefaultArgParser")
	}
}

// TestValidateArgs tests argument validation
func TestValidateArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        *Args
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid_args",
			args: &Args{
				Verbosity:        1,
				Parallel:         4,
				CoverageMode:     "set",
				OptimizationMode: "balanced",
			},
			expectError: false,
		},
		{
			name: "negative_verbosity",
			args: &Args{
				Verbosity: -1,
			},
			expectError: true,
			errorMsg:    "verbosity level must be between 0 and 5",
		},
		{
			name: "high_verbosity",
			args: &Args{
				Verbosity: 6,
			},
			expectError: true,
			errorMsg:    "verbosity level must be between 0 and 5",
		},
		{
			name: "negative_parallel",
			args: &Args{
				Parallel: -1,
			},
			expectError: true,
			errorMsg:    "parallel count cannot be negative",
		},
		{
			name: "invalid_coverage_mode",
			args: &Args{
				CoverageMode: "invalid",
			},
			expectError: true,
			errorMsg:    "invalid coverage mode: invalid",
		},
		{
			name: "invalid_optimization_mode",
			args: &Args{
				OptimizationMode: "invalid",
			},
			expectError: true,
			errorMsg:    "invalid optimization mode: invalid",
		},
		{
			name: "valid_coverage_modes",
			args: &Args{
				CoverageMode: "count",
			},
			expectError: false,
		},
		{
			name: "valid_optimization_modes",
			args: &Args{
				OptimizationMode: "aggressive",
			},
			expectError: false,
		},
		{
			name: "empty_coverage_mode",
			args: &Args{
				CoverageMode: "",
			},
			expectError: false,
		},
		{
			name: "empty_optimization_mode",
			args: &Args{
				OptimizationMode: "",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateArgs(tt.args)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.expectError && err != nil && tt.errorMsg != "" {
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got: %v", tt.errorMsg, err)
				}
			}
		})
	}
}

// TestGetDefaultArgs tests default argument creation
func TestGetDefaultArgs(t *testing.T) {
	t.Parallel()

	args := GetDefaultArgs()
	if args == nil {
		t.Fatal("GetDefaultArgs should not return nil")
	}

	// Verify default values
	if args.Verbosity != 0 {
		t.Errorf("expected default verbosity 0, got %d", args.Verbosity)
	}

	if args.Parallel != 0 {
		t.Errorf("expected default parallel 0, got %d", args.Parallel)
	}

	if args.Colors != true {
		t.Errorf("expected default colors true, got %v", args.Colors)
	}

	if args.Watch != false {
		t.Errorf("expected default watch false, got %v", args.Watch)
	}

	if args.FailFast != false {
		t.Errorf("expected default fail-fast false, got %v", args.FailFast)
	}

	if args.Optimized != false {
		t.Errorf("expected default optimized false, got %v", args.Optimized)
	}

	if len(args.Packages) != 0 {
		t.Errorf("expected empty packages, got %v", args.Packages)
	}

	if args.TestPattern != "" {
		t.Errorf("expected empty test pattern, got %q", args.TestPattern)
	}

	if args.ConfigFile != "" {
		t.Errorf("expected empty config file, got %q", args.ConfigFile)
	}

	if args.Timeout != "" {
		t.Errorf("expected empty timeout, got %q", args.Timeout)
	}

	if args.CoverageMode != "" {
		t.Errorf("expected empty coverage mode, got %q", args.CoverageMode)
	}

	if args.OptimizationMode != "" {
		t.Errorf("expected empty optimization mode, got %q", args.OptimizationMode)
	}
}

// TestDefaultArgParser_ExtractVerbosityFlags tests verbosity flag extraction
func TestDefaultArgParser_ExtractVerbosityFlags(t *testing.T) {
	t.Parallel()

	parser := &DefaultArgParser{}

	tests := []struct {
		name              string
		args              []string
		expectedVerbosity int
		expectedFiltered  []string
	}{
		{
			name:              "no_verbosity_flags",
			args:              []string{"--watch", "package1"},
			expectedVerbosity: 0,
			expectedFiltered:  []string{"--watch", "package1"},
		},
		{
			name:              "single_v_flag",
			args:              []string{"-v", "--watch"},
			expectedVerbosity: 0, // -v alone doesn't count as verbosity level
			expectedFiltered:  []string{"-v", "--watch"},
		},
		{
			name:              "double_v_flag",
			args:              []string{"-vv", "--watch"},
			expectedVerbosity: 2,
			expectedFiltered:  []string{"--watch"},
		},
		{
			name:              "triple_v_flag",
			args:              []string{"-vvv", "package1"},
			expectedVerbosity: 3,
			expectedFiltered:  []string{"package1"},
		},
		{
			name:              "v_with_equals_ignored",
			args:              []string{"-v=true", "--watch"},
			expectedVerbosity: 0,
			expectedFiltered:  []string{"-v=true", "--watch"},
		},
		{
			name:              "mixed_flags",
			args:              []string{"--watch", "-vvvv", "package1", "--color"},
			expectedVerbosity: 4,
			expectedFiltered:  []string{"--watch", "package1", "--color"},
		},
		{
			name:              "verbose_long_form_ignored",
			args:              []string{"--verbose", "package1"},
			expectedVerbosity: 0,
			expectedFiltered:  []string{"--verbose", "package1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			verbosity, filtered := parser.extractVerbosityFlags(tt.args)

			if verbosity != tt.expectedVerbosity {
				t.Errorf("Expected verbosity %d, got %d", tt.expectedVerbosity, verbosity)
			}

			if len(filtered) != len(tt.expectedFiltered) {
				t.Errorf("Expected %d filtered args, got %d", len(tt.expectedFiltered), len(filtered))
			}

			for i, expected := range tt.expectedFiltered {
				if i >= len(filtered) || filtered[i] != expected {
					t.Errorf("Expected filtered arg %d to be %q, got %q", i, expected, filtered[i])
				}
			}
		})
	}
}

// TestDefaultArgParser_StripCommandName tests command name stripping
func TestDefaultArgParser_StripCommandName(t *testing.T) {
	t.Parallel()

	parser := &DefaultArgParser{}

	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "empty_args",
			args:     []string{},
			expected: []string{},
		},
		{
			name:     "known_command_run",
			args:     []string{"run"},
			expected: []string{},
		},
		{
			name:     "known_command_with_flags",
			args:     []string{"run", "--watch", "package1"},
			expected: []string{"--watch", "package1"},
		},
		{
			name:     "known_command_demo",
			args:     []string{"demo", "--verbose"},
			expected: []string{"--verbose"},
		},
		{
			name:     "known_command_help",
			args:     []string{"help"},
			expected: []string{},
		},
		{
			name:     "known_command_version",
			args:     []string{"version"},
			expected: []string{},
		},
		{
			name:     "unknown_command_preserved",
			args:     []string{"go-sentinel", "--watch", "package1"},
			expected: []string{"go-sentinel", "--watch", "package1"},
		},
		{
			name:     "no_command_name",
			args:     []string{"--watch", "package1"},
			expected: []string{"--watch", "package1"},
		},
		{
			name:     "different_command_name_preserved",
			args:     []string{"other-command", "--flag"},
			expected: []string{"other-command", "--flag"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := parser.stripCommandName(tt.args)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d args, got %d", len(tt.expected), len(result))
			}

			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected arg %d to be %q, got %q", i, expected, result[i])
				}
			}
		})
	}
}

// TestDefaultArgParser_CreateFlagSet tests flag set creation
func TestDefaultArgParser_CreateFlagSet(t *testing.T) {
	t.Parallel()

	parser := &DefaultArgParser{}
	fs, flags := parser.createFlagSet()

	if fs == nil {
		t.Fatal("createFlagSet should return non-nil flag set")
	}
	if flags == nil {
		t.Fatal("createFlagSet should return non-nil flag values")
	}

	// Verify flag set name
	if fs.Name() != "go-sentinel" {
		t.Errorf("Expected flag set name 'go-sentinel', got %q", fs.Name())
	}

	// Verify all flag pointers are not nil
	if flags.watchFlag == nil {
		t.Error("watchFlag should not be nil")
	}
	if flags.watchLongFlag == nil {
		t.Error("watchLongFlag should not be nil")
	}
	if flags.colorFlag == nil {
		t.Error("colorFlag should not be nil")
	}
	if flags.noColorFlag == nil {
		t.Error("noColorFlag should not be nil")
	}
	if flags.verboseFlag == nil {
		t.Error("verboseFlag should not be nil")
	}
	if flags.verboseLongFlag == nil {
		t.Error("verboseLongFlag should not be nil")
	}
	if flags.failFastFlag == nil {
		t.Error("failFastFlag should not be nil")
	}
	if flags.testPattern == nil {
		t.Error("testPattern should not be nil")
	}
	if flags.testPatternLong == nil {
		t.Error("testPatternLong should not be nil")
	}
	if flags.verbosityLevel == nil {
		t.Error("verbosityLevel should not be nil")
	}
	if flags.configFile == nil {
		t.Error("configFile should not be nil")
	}
	if flags.timeout == nil {
		t.Error("timeout should not be nil")
	}
	if flags.parallel == nil {
		t.Error("parallel should not be nil")
	}
	if flags.coverage == nil {
		t.Error("coverage should not be nil")
	}
	if flags.optimized == nil {
		t.Error("optimized should not be nil")
	}
	if flags.optimizationMode == nil {
		t.Error("optimizationMode should not be nil")
	}

	// Test default values
	if !*flags.colorFlag {
		t.Error("colorFlag should default to true")
	}
	if *flags.verbosityLevel != "0" {
		t.Errorf("verbosityLevel should default to '0', got %q", *flags.verbosityLevel)
	}
}

// TestDefaultArgParser_BuildArgsFromFlags tests building Args from flags
func TestDefaultArgParser_BuildArgsFromFlags(t *testing.T) {
	t.Parallel()

	parser := &DefaultArgParser{}

	tests := []struct {
		name          string
		setupFlags    func() *flagValues
		baseVerbosity int
		packages      []string
		expectError   bool
		errorMsg      string
		validateArgs  func(*testing.T, *Args)
	}{
		{
			name: "default_flags",
			setupFlags: func() *flagValues {
				_, flags := parser.createFlagSet()
				return flags
			},
			baseVerbosity: 0,
			packages:      []string{"package1"},
			expectError:   false,
			validateArgs: func(t *testing.T, args *Args) {
				if args.Verbosity != 0 {
					t.Errorf("Expected verbosity 0, got %d", args.Verbosity)
				}
				if args.Colors != true {
					t.Error("Expected colors to be true")
				}
				if len(args.Packages) != 1 || args.Packages[0] != "package1" {
					t.Errorf("Expected packages [package1], got %v", args.Packages)
				}
			},
		},
		{
			name: "verbose_flag_set",
			setupFlags: func() *flagValues {
				_, flags := parser.createFlagSet()
				*flags.verboseFlag = true
				return flags
			},
			baseVerbosity: 0,
			packages:      []string{},
			expectError:   false,
			validateArgs: func(t *testing.T, args *Args) {
				if args.Verbosity != 1 {
					t.Errorf("Expected verbosity 1, got %d", args.Verbosity)
				}
			},
		},
		{
			name: "verbosity_level_set",
			setupFlags: func() *flagValues {
				_, flags := parser.createFlagSet()
				*flags.verbosityLevel = "3"
				return flags
			},
			baseVerbosity: 0,
			packages:      []string{},
			expectError:   false,
			validateArgs: func(t *testing.T, args *Args) {
				if args.Verbosity != 3 {
					t.Errorf("Expected verbosity 3, got %d", args.Verbosity)
				}
			},
		},
		{
			name: "invalid_verbosity_level",
			setupFlags: func() *flagValues {
				_, flags := parser.createFlagSet()
				*flags.verbosityLevel = "invalid"
				return flags
			},
			baseVerbosity: 0,
			packages:      []string{},
			expectError:   true,
			errorMsg:      "invalid verbosity level",
		},
		{
			name: "negative_verbosity_level",
			setupFlags: func() *flagValues {
				_, flags := parser.createFlagSet()
				*flags.verbosityLevel = "-1"
				return flags
			},
			baseVerbosity: 0,
			packages:      []string{},
			expectError:   true,
			errorMsg:      "verbosity level cannot be negative",
		},
		{
			name: "high_verbosity_level",
			setupFlags: func() *flagValues {
				_, flags := parser.createFlagSet()
				*flags.verbosityLevel = "6"
				return flags
			},
			baseVerbosity: 0,
			packages:      []string{},
			expectError:   true,
			errorMsg:      "verbosity level too high",
		},
		{
			name: "watch_flags",
			setupFlags: func() *flagValues {
				_, flags := parser.createFlagSet()
				*flags.watchFlag = true
				return flags
			},
			baseVerbosity: 0,
			packages:      []string{},
			expectError:   false,
			validateArgs: func(t *testing.T, args *Args) {
				if !args.Watch {
					t.Error("Expected watch to be true")
				}
			},
		},
		{
			name: "no_color_flag",
			setupFlags: func() *flagValues {
				_, flags := parser.createFlagSet()
				*flags.noColorFlag = true
				return flags
			},
			baseVerbosity: 0,
			packages:      []string{},
			expectError:   false,
			validateArgs: func(t *testing.T, args *Args) {
				if args.Colors {
					t.Error("Expected colors to be false when no-color is set")
				}
			},
		},
		{
			name: "test_pattern_long_form",
			setupFlags: func() *flagValues {
				_, flags := parser.createFlagSet()
				*flags.testPattern = "short"
				*flags.testPatternLong = "long"
				return flags
			},
			baseVerbosity: 0,
			packages:      []string{},
			expectError:   false,
			validateArgs: func(t *testing.T, args *Args) {
				if args.TestPattern != "long" {
					t.Errorf("Expected test pattern 'long', got %q", args.TestPattern)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			flags := tt.setupFlags()
			args, err := parser.buildArgsFromFlags(flags, tt.baseVerbosity, tt.packages)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if args == nil {
					t.Fatal("Expected non-nil args")
				}
				if tt.validateArgs != nil {
					tt.validateArgs(t, args)
				}
			}
		})
	}
}

// TestDefaultArgParser_Parse tests the Parse method comprehensively
func TestDefaultArgParser_Parse(t *testing.T) {
	t.Parallel()

	parser := &DefaultArgParser{}

	tests := []struct {
		name        string
		args        []string
		expectError bool
		expected    *Args
	}{
		{
			name:        "empty_args",
			args:        []string{},
			expectError: false,
			expected: &Args{
				Colors:    true,
				Verbosity: 0,
				Packages:  []string{},
			},
		},
		{
			name:        "watch_flag",
			args:        []string{"--watch"},
			expectError: false,
			expected: &Args{
				Watch:     true,
				Colors:    true,
				Verbosity: 0,
			},
		},
		{
			name:        "color_flag",
			args:        []string{"--color"},
			expectError: false,
			expected: &Args{
				Colors:    true,
				Verbosity: 0,
			},
		},
		{
			name:        "verbose_flag",
			args:        []string{"--verbose"},
			expectError: false,
			expected: &Args{
				Colors:    true,
				Verbosity: 1,
			},
		},
		{
			name:        "verbosity_level",
			args:        []string{"--verbosity=3"},
			expectError: false,
			expected: &Args{
				Colors:    true,
				Verbosity: 3,
			},
		},
		{
			name:        "multiple_v_flags",
			args:        []string{"-vvv"},
			expectError: false,
			expected: &Args{
				Colors:    true,
				Verbosity: 3,
			},
		},
		{
			name:        "packages",
			args:        []string{"./pkg", "./cmd"},
			expectError: false,
			expected: &Args{
				Colors:   true,
				Packages: []string{"./pkg", "./cmd"},
			},
		},
		{
			name:        "test_pattern",
			args:        []string{"--test=TestExample"},
			expectError: false,
			expected: &Args{
				Colors:      true,
				TestPattern: "TestExample",
			},
		},
		{
			name:        "fail_fast",
			args:        []string{"--fail-fast"},
			expectError: false,
			expected: &Args{
				Colors:   true,
				FailFast: true,
			},
		},
		{
			name:        "optimized",
			args:        []string{"--optimized"},
			expectError: false,
			expected: &Args{
				Colors:    true,
				Optimized: true,
			},
		},
		{
			name:        "optimization_mode",
			args:        []string{"--optimization=aggressive"},
			expectError: false,
			expected: &Args{
				Colors:           true,
				OptimizationMode: "aggressive",
			},
		},
		{
			name:        "parallel",
			args:        []string{"--parallel=4"},
			expectError: false,
			expected: &Args{
				Colors:   true,
				Parallel: 4,
			},
		},
		{
			name:        "config_file",
			args:        []string{"--config=sentinel.json"},
			expectError: false,
			expected: &Args{
				Colors:     true,
				ConfigFile: "sentinel.json",
			},
		},
		{
			name:        "timeout",
			args:        []string{"--timeout=30s"},
			expectError: false,
			expected: &Args{
				Colors:  true,
				Timeout: "30s",
			},
		},
		{
			name:        "coverage",
			args:        []string{"--covermode=set"},
			expectError: false,
			expected: &Args{
				Colors:       true,
				CoverageMode: "set",
			},
		},
		{
			name:        "strip_run_command",
			args:        []string{"run", "--watch"},
			expectError: false,
			expected: &Args{
				Colors: true,
				Watch:  true,
			},
		},
		{
			name:        "strip_demo_command",
			args:        []string{"demo", "--verbose"},
			expectError: false,
			expected: &Args{
				Colors:    true,
				Verbosity: 1,
			},
		},
		{
			name:        "strip_help_command",
			args:        []string{"help"},
			expectError: false,
			expected: &Args{
				Colors: true,
			},
		},
		{
			name:        "strip_version_command",
			args:        []string{"version"},
			expectError: false,
			expected: &Args{
				Colors: true,
			},
		},
		{
			name:        "complex_combination",
			args:        []string{"--watch", "--verbose", "--color", "--optimized", "./pkg", "./cmd"},
			expectError: false,
			expected: &Args{
				Watch:     true,
				Colors:    true,
				Verbosity: 1,
				Optimized: true,
				Packages:  []string{"./pkg", "./cmd"},
			},
		},
		{
			name:        "invalid_verbosity",
			args:        []string{"--verbosity=invalid"},
			expectError: true,
		},
		{
			name:        "negative_verbosity",
			args:        []string{"--verbosity=-1"},
			expectError: true,
		},
		{
			name:        "high_verbosity",
			args:        []string{"--verbosity=6"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := parser.Parse(tt.args)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError && result == nil {
				t.Error("expected non-nil result")
			}

			if tt.expected != nil && result != nil {
				if result.Watch != tt.expected.Watch {
					t.Errorf("expected watch %v, got %v", tt.expected.Watch, result.Watch)
				}

				if result.Colors != tt.expected.Colors {
					t.Errorf("expected colors %v, got %v", tt.expected.Colors, result.Colors)
				}

				if result.Verbosity != tt.expected.Verbosity {
					t.Errorf("expected verbosity %v, got %v", tt.expected.Verbosity, result.Verbosity)
				}

				if result.FailFast != tt.expected.FailFast {
					t.Errorf("expected fail-fast %v, got %v", tt.expected.FailFast, result.FailFast)
				}

				if result.Optimized != tt.expected.Optimized {
					t.Errorf("expected optimized %v, got %v", tt.expected.Optimized, result.Optimized)
				}

				if result.TestPattern != tt.expected.TestPattern {
					t.Errorf("expected test pattern %q, got %q", tt.expected.TestPattern, result.TestPattern)
				}

				if result.OptimizationMode != tt.expected.OptimizationMode {
					t.Errorf("expected optimization mode %q, got %q", tt.expected.OptimizationMode, result.OptimizationMode)
				}

				if result.Parallel != tt.expected.Parallel {
					t.Errorf("expected parallel %v, got %v", tt.expected.Parallel, result.Parallel)
				}

				if result.ConfigFile != tt.expected.ConfigFile {
					t.Errorf("expected config file %q, got %q", tt.expected.ConfigFile, result.ConfigFile)
				}

				if result.Timeout != tt.expected.Timeout {
					t.Errorf("expected timeout %q, got %q", tt.expected.Timeout, result.Timeout)
				}

				if result.CoverageMode != tt.expected.CoverageMode {
					t.Errorf("expected coverage mode %q, got %q", tt.expected.CoverageMode, result.CoverageMode)
				}

				if len(tt.expected.Packages) > 0 {
					if len(result.Packages) != len(tt.expected.Packages) {
						t.Errorf("expected packages %v, got %v", tt.expected.Packages, result.Packages)
					} else {
						for i, pkg := range tt.expected.Packages {
							if result.Packages[i] != pkg {
								t.Errorf("expected package[%d] %q, got %q", i, pkg, result.Packages[i])
							}
						}
					}
				}
			}
		})
	}
}

// TestDefaultArgParser_ParseFromCobra tests the ParseFromCobra method
func TestDefaultArgParser_ParseFromCobra(t *testing.T) {
	t.Parallel()

	parser := &DefaultArgParser{}

	tests := []struct {
		name             string
		watchFlag        bool
		colorFlag        bool
		verboseFlag      bool
		failFastFlag     bool
		optimizedFlag    bool
		packages         []string
		testPattern      string
		optimizationMode string
		expected         *Args
	}{
		{
			name:             "all_false",
			watchFlag:        false,
			colorFlag:        false,
			verboseFlag:      false,
			failFastFlag:     false,
			optimizedFlag:    false,
			packages:         []string{},
			testPattern:      "",
			optimizationMode: "",
			expected: &Args{
				Watch:            false,
				Colors:           false,
				Verbosity:        0,
				FailFast:         false,
				Optimized:        false,
				Packages:         []string{},
				TestPattern:      "",
				OptimizationMode: "",
			},
		},
		{
			name:             "all_true",
			watchFlag:        true,
			colorFlag:        true,
			verboseFlag:      true,
			failFastFlag:     true,
			optimizedFlag:    true,
			packages:         []string{"./pkg", "./cmd"},
			testPattern:      "TestExample",
			optimizationMode: "aggressive",
			expected: &Args{
				Watch:            true,
				Colors:           true,
				Verbosity:        1,
				FailFast:         true,
				Optimized:        true,
				Packages:         []string{"./pkg", "./cmd"},
				TestPattern:      "TestExample",
				OptimizationMode: "aggressive",
			},
		},
		{
			name:             "mixed_flags",
			watchFlag:        true,
			colorFlag:        false,
			verboseFlag:      true,
			failFastFlag:     false,
			optimizedFlag:    true,
			packages:         []string{"./internal"},
			testPattern:      "TestConfig",
			optimizationMode: "balanced",
			expected: &Args{
				Watch:            true,
				Colors:           false,
				Verbosity:        1,
				FailFast:         false,
				Optimized:        true,
				Packages:         []string{"./internal"},
				TestPattern:      "TestConfig",
				OptimizationMode: "balanced",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := parser.ParseFromCobra(
				tt.watchFlag,
				tt.colorFlag,
				tt.verboseFlag,
				tt.failFastFlag,
				tt.optimizedFlag,
				tt.packages,
				tt.testPattern,
				tt.optimizationMode,
			)

			if result == nil {
				t.Fatal("ParseFromCobra should not return nil")
			}

			if result.Watch != tt.expected.Watch {
				t.Errorf("expected watch %v, got %v", tt.expected.Watch, result.Watch)
			}

			if result.Colors != tt.expected.Colors {
				t.Errorf("expected colors %v, got %v", tt.expected.Colors, result.Colors)
			}

			if result.Verbosity != tt.expected.Verbosity {
				t.Errorf("expected verbosity %v, got %v", tt.expected.Verbosity, result.Verbosity)
			}

			if result.FailFast != tt.expected.FailFast {
				t.Errorf("expected fail-fast %v, got %v", tt.expected.FailFast, result.FailFast)
			}

			if result.Optimized != tt.expected.Optimized {
				t.Errorf("expected optimized %v, got %v", tt.expected.Optimized, result.Optimized)
			}

			if result.TestPattern != tt.expected.TestPattern {
				t.Errorf("expected test pattern %q, got %q", tt.expected.TestPattern, result.TestPattern)
			}

			if result.OptimizationMode != tt.expected.OptimizationMode {
				t.Errorf("expected optimization mode %q, got %q", tt.expected.OptimizationMode, result.OptimizationMode)
			}

			if len(result.Packages) != len(tt.expected.Packages) {
				t.Errorf("expected packages %v, got %v", tt.expected.Packages, result.Packages)
			} else {
				for i, pkg := range tt.expected.Packages {
					if result.Packages[i] != pkg {
						t.Errorf("expected package[%d] %q, got %q", i, pkg, result.Packages[i])
					}
				}
			}
		})
	}
}
