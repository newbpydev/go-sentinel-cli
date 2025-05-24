package demo

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/newbpydev/go-sentinel/internal/cli"
)

// RunPhase5Demo runs the Phase 5 demonstration (Watch Mode)
func RunPhase5Demo() {
	fmt.Println("=== Phase 5: Watch Mode Demonstration ===")
	fmt.Println()

	// Create formatters for proper styling
	isColorSupported := isColorTerminal()
	formatter := cli.NewColorFormatter(isColorSupported)
	icons := cli.NewIconProvider(true) // Use Unicode icons

	// Create a temporary test project
	projectDir, err := createDemoProject()
	if err != nil {
		fmt.Printf("Error creating demo project: %v\n", err)
		return
	}
	defer func() {
		if err := os.RemoveAll(projectDir); err != nil {
			fmt.Printf("Error removing demo project: %v\n", err)
		}
	}()

	fmt.Printf("Created demo project in %s\n", formatter.Dim(projectDir))
	fmt.Println()

	// Simulate watch mode with proper styling
	fmt.Printf("%s %s\n",
		formatter.Cyan("Watch mode started"),
		formatter.Dim("- watching for file changes"))
	fmt.Printf("Press %s to quit, %s to run all tests, %s to run changed tests\n",
		formatter.Bold("'q'"),
		formatter.Bold("'a'"),
		formatter.Bold("'c'"))
	fmt.Println()

	// Simulate file changes with styled output
	fmt.Printf("%s %s\n",
		formatter.Yellow("→"),
		"Detected changes in project files")
	fmt.Printf("%s %s\n",
		formatter.Blue("Running tests..."),
		formatter.Dim("(2 files)"))
	fmt.Println()

	// Simulate test run with proper file formatting and icons
	time.Sleep(500 * time.Millisecond)

	// Format like Vitest test suite output
	fmt.Printf(" %s %s %s %s %s\n",
		formatter.Green(icons.CheckMark()),
		formatter.Bold(formatter.Cyan("pkg/math_test.go")),
		formatter.Green("(2 tests)"),
		formatter.Dim("119ms"),
		formatter.Dim("12 MB heap used"))

	time.Sleep(300 * time.Millisecond)
	fmt.Printf(" %s %s %s %s %s\n",
		formatter.Green(icons.CheckMark()),
		formatter.Bold(formatter.Cyan("pkg/string_test.go")),
		formatter.Green("(1 test)"),
		formatter.Dim("45ms"),
		formatter.Dim("8 MB heap used"))
	fmt.Println()

	// Simulate another file change
	fmt.Printf("%s %s\n",
		formatter.Yellow("→"),
		"Detected changes in math.go")
	fmt.Printf("%s %s\n",
		formatter.Blue("Running related tests..."),
		formatter.Dim("(1 file)"))
	fmt.Println()

	// Simulate test run
	time.Sleep(500 * time.Millisecond)
	fmt.Printf(" %s %s %s %s %s\n",
		formatter.Green(icons.CheckMark()),
		formatter.Bold(formatter.Cyan("pkg/math_test.go")),
		formatter.Green("(3 tests)"),
		formatter.Dim("156ms"),
		formatter.Dim("14 MB heap used"))
	fmt.Println()

	// Simulate test failure with proper styling
	fmt.Printf("%s %s\n",
		formatter.Yellow("→"),
		"Detected changes in string_test.go")
	fmt.Printf("%s %s\n",
		formatter.Blue("Running tests..."),
		formatter.Dim("(1 file)"))
	fmt.Println()

	time.Sleep(500 * time.Millisecond)
	fmt.Printf(" %s %s %s %s %s\n",
		formatter.Red(icons.Cross()),
		formatter.Bold(formatter.Cyan("pkg/string_test.go")),
		formatter.Red("(1 failed)")+" | "+formatter.Green("1 passed"),
		formatter.Dim("73ms"),
		formatter.Dim("9 MB heap used"))

	// Show the failing test detail
	fmt.Printf("   %s %s\n",
		formatter.Red(icons.Cross()),
		formatter.Red("TestReverse/failing_test"))
	fmt.Printf("     %s %s\n",
		formatter.Dim("→"),
		formatter.Red("Expected 'tset', got 'test'"))
	fmt.Println()

	// Add separator line before summary (like Vitest)
	fmt.Println(formatter.Dim(strings.Repeat("─", 60)))

	// Summary with proper styling
	fmt.Println(formatter.Bold("Test Summary:"))

	// Test Files line
	fmt.Printf("Test Files: %s, %s %s\n",
		formatter.Green("2 passed"),
		formatter.Red("1 failed"),
		formatter.Dim("(total: 3)"))

	// Tests line
	fmt.Printf("Tests: %s, %s %s\n",
		formatter.Green("6 passed"),
		formatter.Red("1 failed"),
		formatter.Dim("(total: 7)"))

	// Time line
	fmt.Printf("Start at: %s\n", formatter.Dim(time.Now().Format("15:04:05")))
	fmt.Printf("Duration: %s\n", formatter.Dim("1.35s"))
	fmt.Println()

	fmt.Printf("%s %s\n",
		formatter.Cyan("Watch mode demonstration completed."),
		formatter.Dim("Press Ctrl+C to exit watch mode"))
}

// createDemoProject creates a temporary project with test files
func createDemoProject() (string, error) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "watch-demo")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Create a package directory
	pkgDir := filepath.Join(tempDir, "pkg")
	if err := os.Mkdir(pkgDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create package dir: %w", err)
	}

	// Create implementation file
	mathFile := filepath.Join(pkgDir, "math.go")
	mathContent := `package pkg

// Add adds two integers and returns the result
func Add(a, b int) int {
	return a + b
}

// Subtract subtracts b from a and returns the result
func Subtract(a, b int) int {
	return a - b
}
`
	if err := os.WriteFile(mathFile, []byte(mathContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create math.go: %w", err)
	}

	// Create test file
	mathTestFile := filepath.Join(pkgDir, "math_test.go")
	mathTestContent := `package pkg

import "testing"

func TestAdd(t *testing.T) {
	tests := []struct {
		name     string
		a, b     int
		expected int
	}{
		{"positive numbers", 2, 3, 5},
		{"negative numbers", -2, -3, -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Add(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}
`
	if err := os.WriteFile(mathTestFile, []byte(mathTestContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create math_test.go: %w", err)
	}

	// Create another implementation file
	stringFile := filepath.Join(pkgDir, "string.go")
	stringContent := `package pkg

// Reverse returns the string reversed
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
`
	if err := os.WriteFile(stringFile, []byte(stringContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create string.go: %w", err)
	}

	// Create test file for string
	stringTestFile := filepath.Join(pkgDir, "string_test.go")
	stringTestContent := `package pkg

import "testing"

func TestReverse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"single character", "a", "a"},
		{"normal string", "hello", "olleh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reverse(tt.input)
			if result != tt.expected {
				t.Errorf("Reverse(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
`
	if err := os.WriteFile(stringTestFile, []byte(stringTestContent), 0644); err != nil {
		return "", fmt.Errorf("failed to create string_test.go: %w", err)
	}

	return tempDir, nil
}
