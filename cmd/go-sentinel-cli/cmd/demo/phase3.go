package demo

import (
	"fmt"
	"os"

	"github.com/newbpydev/go-sentinel/internal/cli"
)

// RunPhase3Demo runs the Phase 3-D demonstration (Failed Test Renderer)
func RunPhase3Demo() {
	fmt.Println("=== Phase 3-D: Failed Test Details Section Demonstration ===")
	fmt.Println()

	// Get terminal properties
	isColorSupported := isColorTerminal()
	terminalWidth := 80 // Fixed width for demo purposes

	// Create formatters
	formatter := cli.NewColorFormatter(isColorSupported)
	icons := cli.NewIconProvider(true) // Always use Unicode icons for demo

	// Create a renderer for detailed failed test rendering
	failedRenderer := cli.NewFailedTestRenderer(os.Stdout, formatter, icons, terminalWidth)

	// Create sample failed tests with source context
	failedTests := createMockFailedTestsWithSourceContext()

	// Render the failed tests section
	err := failedRenderer.RenderFailedTests(failedTests)
	if err != nil {
		fmt.Printf("Error rendering failed tests: %v\n", err)
	}

	// Add a summary and comparison
	fmt.Println("\nComparison with Vitest Failed Tests Display:")
	fmt.Println("1. Distinctive red separator lines above and below failed tests section ✓")
	fmt.Println("2. Red background with white 'Failed Tests X' header ✓")
	fmt.Println("3. Red FAIL badge at the beginning of each test ✓")
	fmt.Println("4. Error type and message displayed in red ✓")
	fmt.Println("5. File path and line numbers displayed with detailed context ✓")
	fmt.Println("6. Source code displayed with line numbers and highlighted error line ✓")
	fmt.Println("7. Error position marked with ^ character under the error location ✓")
}
