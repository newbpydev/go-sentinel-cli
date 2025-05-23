package demo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/newbpydev/go-sentinel/internal/cli"
)

// RunPhase7DDemo runs the Phase 7-D demonstration (exported)
func RunPhase7DDemo() {
	if err := runPhase7DDemo(); err != nil {
		fmt.Printf("Error running Phase 7-D demo: %v\n", err)
	}
}

// runPhase7DDemo demonstrates CLI options and configuration functionality
func runPhase7DDemo() error {
	formatter := cli.NewColorFormatter(isColorTerminal())
	icons := cli.NewIconProvider(isColorTerminal())

	fmt.Println(formatter.Bold(formatter.Cyan("🚀 Phase 7-D: CLI Options & Configuration Demonstration")))
	fmt.Println(formatter.Dim("──────────────────────────────────────────────────────────"))
	fmt.Println()

	// Demo 1: CLI Arguments Parsing
	if err := demonstrateCLIArguments(formatter, icons); err != nil {
		return err
	}

	fmt.Println()
	time.Sleep(1 * time.Second)

	// Demo 2: Configuration File Creation and Loading
	if err := demonstrateConfigurationFiles(formatter, icons); err != nil {
		return err
	}

	fmt.Println()
	time.Sleep(1 * time.Second)

	// Demo 3: CLI Options Variations
	if err := demonstrateVariousCLIOptions(formatter, icons); err != nil {
		return err
	}

	fmt.Println()
	time.Sleep(1 * time.Second)

	// Demo 4: Configuration Precedence and Merging
	if err := demonstrateConfigurationPrecedence(formatter, icons); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(formatter.Bold(formatter.Green("✅ Phase 7-D: CLI Options & Configuration Demo Complete!")))
	fmt.Println(formatter.Dim("All CLI argument parsing and configuration features demonstrated successfully."))

	return nil
}

// demonstrateCLIArguments shows CLI argument parsing capabilities
func demonstrateCLIArguments(formatter *cli.ColorFormatter, icons *cli.IconProvider) error {
	fmt.Println(formatter.Bold("📋 CLI Arguments Parsing"))
	fmt.Println(formatter.Dim("Testing various CLI argument combinations and validation"))
	fmt.Println()

	parser := &cli.DefaultArgParser{}

	// Test cases with different argument combinations
	testCases := []struct {
		name        string
		args        []string
		description string
		expectError bool
	}{
		{
			name:        "Basic Watch Mode",
			args:        []string{"-w", "./internal"},
			description: "Simple watch mode with package",
			expectError: false,
		},
		{
			name:        "Verbose Testing",
			args:        []string{"-vvv", "--test=TestUnit*", "./..."},
			description: "High verbosity with test pattern",
			expectError: false,
		},
		{
			name:        "Production Mode",
			args:        []string{"--no-color", "--parallel=8", "--timeout=60s", "--fail-fast"},
			description: "CI/CD optimized settings",
			expectError: false,
		},
		{
			name:        "Invalid Verbosity",
			args:        []string{"--verbosity=10"},
			description: "Invalid verbosity level",
			expectError: true,
		},
	}

	for i, tc := range testCases {
		fmt.Printf("  %s %s\n", formatter.Yellow("⏱️"), formatter.Bold(tc.name))
		fmt.Printf("    %s %s\n", formatter.Cyan("Args:"), formatter.Dim(fmt.Sprintf("%v", tc.args)))
		fmt.Printf("    %s %s\n", formatter.Cyan("Desc:"), tc.description)

		time.Sleep(200 * time.Millisecond)

		cliArgs, err := parser.Parse(tc.args)

		if tc.expectError {
			if err != nil {
				fmt.Printf("    %s Correctly caught validation error\n", formatter.Green(icons.CheckMark()))
				fmt.Printf("      %s %s\n", formatter.Red("Error:"), formatter.Dim(err.Error()))
			} else {
				fmt.Printf("    %s Expected error but got none\n", formatter.Red(icons.Cross()))
			}
		} else {
			if err != nil {
				fmt.Printf("    %s Unexpected error: %v\n", formatter.Red(icons.Cross()), err)
			} else {
				fmt.Printf("    %s Parsed successfully\n", formatter.Green(icons.CheckMark()))
				fmt.Printf("      %s Watch: %t, Colors: %t, Verbosity: %d\n",
					formatter.Cyan("→"), cliArgs.Watch, cliArgs.Colors, cliArgs.Verbosity)
			}
		}

		if i < len(testCases)-1 {
			fmt.Println()
		}
	}

	fmt.Println()
	fmt.Printf("%s %s\n",
		formatter.Green(icons.CheckMark()),
		formatter.Bold("CLI argument parsing demonstration completed"))

	return nil
}

// demonstrateConfigurationFiles shows configuration file creation and loading
func demonstrateConfigurationFiles(formatter *cli.ColorFormatter, icons *cli.IconProvider) error {
	fmt.Println(formatter.Bold("📁 Configuration File Management"))
	fmt.Println(formatter.Dim("Creating sample configurations and testing file loading"))
	fmt.Println()

	// Create demo-configs directory
	configDir := "demo-configs"
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Sample configurations for different scenarios
	configurations := map[string]map[string]interface{}{
		"development.json": {
			"colors":      true,
			"icons":       "unicode",
			"verbosity":   1,
			"watchMode":   true,
			"parallel":    4,
			"timeout":     "30s",
			"testCommand": "go test -race",
		},
		"ci-cd.json": {
			"colors":    false,
			"icons":     "ascii",
			"verbosity": 0,
			"watchMode": false,
			"parallel":  8,
			"timeout":   "120s",
			"failFast":  true,
		},
		"performance.json": {
			"colors":      true,
			"icons":       "minimal",
			"verbosity":   0,
			"parallel":    16,
			"timeout":     "300s",
			"testCommand": "go test -benchmem -bench=.",
		},
	}

	fmt.Println(formatter.Yellow("📝 Creating sample configuration files..."))
	time.Sleep(300 * time.Millisecond)

	for filename, config := range configurations {
		configPath := filepath.Join(configDir, filename)

		data, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal config %s: %w", filename, err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write config %s: %w", filename, err)
		}

		fmt.Printf("  %s Created: %s\n", formatter.Green(icons.CheckMark()), formatter.Cyan(configPath))
	}

	fmt.Println()
	fmt.Println(formatter.Yellow("🔍 Testing configuration loading..."))
	time.Sleep(300 * time.Millisecond)

	// Test loading each configuration
	loader := &cli.DefaultConfigLoader{}
	for filename := range configurations {
		configPath := filepath.Join(configDir, filename)

		config, err := loader.LoadFromFile(configPath)
		if err != nil {
			fmt.Printf("  %s Failed to load %s: %v\n", formatter.Red(icons.Cross()), filename, err)
		} else {
			fmt.Printf("  %s Loaded %s successfully\n", formatter.Green(icons.CheckMark()), formatter.Cyan(filename))
			fmt.Printf("    %s Colors: %t, Icons: %s, Parallel: %d\n",
				formatter.Cyan("→"), config.Colors, config.Visual.Icons, config.Parallel)
		}
	}

	fmt.Println()
	fmt.Printf("%s %s\n",
		formatter.Green(icons.CheckMark()),
		formatter.Bold("Configuration file management demonstration completed"))

	return nil
}

// demonstrateVariousCLIOptions shows different CLI option combinations
func demonstrateVariousCLIOptions(formatter *cli.ColorFormatter, icons *cli.IconProvider) error {
	fmt.Println(formatter.Bold("🎯 CLI Options Variations"))
	fmt.Println(formatter.Dim("Testing various CLI option combinations for different use cases"))
	fmt.Println()

	parser := &cli.DefaultArgParser{}

	// Realistic usage scenarios
	scenarios := []struct {
		name string
		args []string
		desc string
	}{
		{
			name: "Development Workflow",
			args: []string{"-w", "-v", "--color", "./internal", "./cmd"},
			desc: "Watch mode with verbose output for active development",
		},
		{
			name: "Quick Test Run",
			args: []string{"--test=TestConfig*", "./pkg"},
			desc: "Run specific test patterns on a package",
		},
		{
			name: "CI Pipeline",
			args: []string{"--no-color", "--parallel=8", "--timeout=120s", "--fail-fast", "./..."},
			desc: "Production CI/CD configuration",
		},
		{
			name: "Performance Testing",
			args: []string{"--parallel=16", "--timeout=300s", "./benchmarks"},
			desc: "High-performance testing configuration",
		},
		{
			name: "Debug Mode",
			args: []string{"-vvv", "--parallel=1", "--timeout=60s", "./internal"},
			desc: "Sequential execution with maximum verbosity for debugging",
		},
	}

	for i, scenario := range scenarios {
		fmt.Printf("  %s %s\n", formatter.Yellow("⚡"), formatter.Bold(scenario.name))
		fmt.Printf("    %s %s\n", formatter.Cyan("Command:"), formatter.Dim(fmt.Sprintf("go-sentinel %v", scenario.args)))
		fmt.Printf("    %s %s\n", formatter.Cyan("Use case:"), scenario.desc)

		time.Sleep(200 * time.Millisecond)

		cliArgs, err := parser.Parse(scenario.args)
		if err != nil {
			fmt.Printf("    %s Error: %v\n", formatter.Red(icons.Cross()), err)
		} else {
			fmt.Printf("    %s Parsed successfully\n", formatter.Green(icons.CheckMark()))

			// Show key parsed values
			details := []string{}
			if cliArgs.Watch {
				details = append(details, "watch")
			}
			if cliArgs.Colors {
				details = append(details, "colors")
			}
			if cliArgs.Verbosity > 0 {
				details = append(details, fmt.Sprintf("v%d", cliArgs.Verbosity))
			}
			if cliArgs.Parallel > 0 {
				details = append(details, fmt.Sprintf("parallel=%d", cliArgs.Parallel))
			}
			if len(cliArgs.Packages) > 0 {
				details = append(details, fmt.Sprintf("packages=%d", len(cliArgs.Packages)))
			}

			if len(details) > 0 {
				fmt.Printf("      %s %s\n", formatter.Cyan("→"), formatter.Dim(fmt.Sprintf("[%s]", joinStrings(details, ", "))))
			}
		}

		if i < len(scenarios)-1 {
			fmt.Println()
		}
	}

	fmt.Println()
	fmt.Printf("%s %s\n",
		formatter.Green(icons.CheckMark()),
		formatter.Bold("CLI options variations demonstration completed"))

	return nil
}

// demonstrateConfigurationPrecedence shows how CLI args override config files
func demonstrateConfigurationPrecedence(formatter *cli.ColorFormatter, icons *cli.IconProvider) error {
	fmt.Println(formatter.Bold("🔄 Configuration Precedence & Merging"))
	fmt.Println(formatter.Dim("Testing CLI argument precedence over configuration files"))
	fmt.Println()

	// Load a configuration file
	configPath := "demo-configs/development.json"
	loader := &cli.DefaultConfigLoader{}

	fmt.Println(formatter.Yellow("📄 Loading base configuration..."))
	time.Sleep(300 * time.Millisecond)

	config, err := loader.LoadFromFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	fmt.Printf("  %s Loaded configuration from %s\n", formatter.Green(icons.CheckMark()), formatter.Cyan(configPath))
	fmt.Printf("    %s Colors: %t, Verbosity: %d, Parallel: %d\n",
		formatter.Cyan("→"), config.Colors, config.Verbosity, config.Parallel)
	fmt.Printf("    %s Watch: %t, Icons: %s\n",
		formatter.Cyan("→"), config.Watch.Enabled, config.Visual.Icons)

	fmt.Println()
	fmt.Println(formatter.Yellow("⚙️  Testing CLI argument overrides..."))
	time.Sleep(300 * time.Millisecond)

	// Create CLI args that override config values
	cliArgs := &cli.Args{
		Watch:       false, // Override config (was true)
		Colors:      false, // Override config (was true)
		Verbosity:   3,     // Override config (was 1)
		Parallel:    2,     // Override config (was 4)
		TestPattern: "TestOverride",
		FailFast:    true,
	}

	fmt.Printf("  %s CLI arguments (overrides):\n", formatter.Green(icons.CheckMark()))
	fmt.Printf("    %s Colors: %t, Verbosity: %d, Parallel: %d\n",
		formatter.Cyan("→"), cliArgs.Colors, cliArgs.Verbosity, cliArgs.Parallel)
	fmt.Printf("    %s Watch: %t, Pattern: %s, FailFast: %t\n",
		formatter.Cyan("→"), cliArgs.Watch, cliArgs.TestPattern, cliArgs.FailFast)

	fmt.Println()
	fmt.Println(formatter.Yellow("🔀 Merging configurations..."))
	time.Sleep(300 * time.Millisecond)

	// Merge configuration with CLI args
	merged := config.MergeWithCLIArgs(cliArgs)

	fmt.Printf("  %s Merged configuration (CLI takes precedence):\n", formatter.Green(icons.CheckMark()))
	fmt.Printf("    %s Colors: %t %s\n",
		formatter.Cyan("→"), merged.Colors, formatter.Dim("(from CLI)"))
	fmt.Printf("    %s Verbosity: %d %s\n",
		formatter.Cyan("→"), merged.Verbosity, formatter.Dim("(from CLI)"))
	fmt.Printf("    %s Parallel: %d %s\n",
		formatter.Cyan("→"), merged.Parallel, formatter.Dim("(from CLI)"))
	fmt.Printf("    %s Watch: %t %s\n",
		formatter.Cyan("→"), merged.Watch.Enabled, formatter.Dim("(from CLI)"))
	fmt.Printf("    %s Icons: %s %s\n",
		formatter.Cyan("→"), merged.Visual.Icons, formatter.Dim("(from config)"))
	fmt.Printf("    %s Timeout: %v %s\n",
		formatter.Cyan("→"), merged.Timeout, formatter.Dim("(from config)"))

	fmt.Println()
	fmt.Println(formatter.Yellow("✅ Validating merged configuration..."))
	time.Sleep(300 * time.Millisecond)

	// Validate the merged configuration
	if err := cli.ValidateConfig(merged); err != nil {
		fmt.Printf("  %s Validation failed: %v\n", formatter.Red(icons.Cross()), err)
	} else {
		fmt.Printf("  %s Merged configuration is valid\n", formatter.Green(icons.CheckMark()))
	}

	fmt.Println()
	fmt.Printf("%s %s\n",
		formatter.Green(icons.CheckMark()),
		formatter.Bold("Configuration precedence demonstration completed"))

	return nil
}

// Helper function to join strings (since strings.Join might not be available)
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
