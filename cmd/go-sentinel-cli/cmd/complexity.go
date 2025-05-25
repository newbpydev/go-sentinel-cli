package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/newbpydev/go-sentinel/internal/test/metrics"
	"github.com/spf13/cobra"
)

// complexityCmd represents the complexity analysis command
var complexityCmd = &cobra.Command{
	Use:   "complexity [path]",
	Short: "Analyze code complexity metrics for Go files",
	Long: `Analyze code complexity metrics including cyclomatic complexity, 
maintainability index, and technical debt for Go source files.

This command provides comprehensive code quality analysis with:
- Cyclomatic complexity measurement per function
- Maintainability index calculation using industry standards  
- Technical debt estimation in time units
- Violation detection and severity assessment
- Actionable recommendations for code improvement

Examples:
  go-sentinel complexity                    # Analyze current directory
  go-sentinel complexity ./internal        # Analyze specific package
  go-sentinel complexity --format=json     # JSON output format
  go-sentinel complexity --output=report.html --format=html`,
	Args: cobra.MaximumNArgs(1),
	RunE: runComplexityAnalysis,
}

var (
	complexityFormat     string
	complexityOutput     string
	complexityThresholds complexityThresholdFlags
	complexityVerbose    bool
)

type complexityThresholdFlags struct {
	cyclomaticComplexity int
	maintainabilityIndex float64
	linesOfCode          int
	technicalDebtRatio   float64
	functionLength       int
}

func init() {
	rootCmd.AddCommand(complexityCmd)

	// Output format options
	complexityCmd.Flags().StringVarP(&complexityFormat, "format", "f", "text",
		"Output format: text, json, html")
	complexityCmd.Flags().StringVarP(&complexityOutput, "output", "o", "",
		"Output file path (default: stdout)")

	// Threshold configuration
	complexityCmd.Flags().IntVar(&complexityThresholds.cyclomaticComplexity, "max-complexity", 10,
		"Maximum cyclomatic complexity threshold")
	complexityCmd.Flags().Float64Var(&complexityThresholds.maintainabilityIndex, "min-maintainability", 85.0,
		"Minimum maintainability index threshold")
	complexityCmd.Flags().IntVar(&complexityThresholds.linesOfCode, "max-lines", 500,
		"Maximum lines per file threshold")
	complexityCmd.Flags().Float64Var(&complexityThresholds.technicalDebtRatio, "max-debt-ratio", 5.0,
		"Maximum technical debt ratio threshold")
	complexityCmd.Flags().IntVar(&complexityThresholds.functionLength, "max-function-lines", 50,
		"Maximum lines per function threshold")

	// Verbose output
	complexityCmd.Flags().BoolVarP(&complexityVerbose, "verbose", "v", false,
		"Verbose output with detailed analysis")
}

func runComplexityAnalysis(cmd *cobra.Command, args []string) error {
	// Determine analysis path
	analysisPath := "."
	if len(args) > 0 {
		analysisPath = args[0]
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(analysisPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path %s: %w", analysisPath, err)
	}

	// Verify path exists
	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", absPath)
	}

	// Create complexity analyzer
	analyzer := metrics.NewComplexityAnalyzer()

	// Configure custom thresholds if provided
	thresholds := metrics.ComplexityThresholds{
		CyclomaticComplexity: complexityThresholds.cyclomaticComplexity,
		MaintainabilityIndex: complexityThresholds.maintainabilityIndex,
		LinesOfCode:          complexityThresholds.linesOfCode,
		TechnicalDebtRatio:   complexityThresholds.technicalDebtRatio,
		FunctionLength:       complexityThresholds.functionLength,
	}
	analyzer.SetThresholds(thresholds)

	// Determine if analyzing single file or project
	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if complexityVerbose {
		fmt.Printf("🔍 Analyzing complexity for: %s\n", absPath)
		if fileInfo.IsDir() {
			fmt.Printf("📁 Scanning directory for Go packages...\n")
		} else {
			fmt.Printf("📄 Analyzing single file...\n")
		}
	}

	var projectComplexity *metrics.ProjectComplexity

	if fileInfo.IsDir() {
		// Analyze entire project/directory
		projectComplexity, err = analyzer.AnalyzeProject(absPath)
		if err != nil {
			return fmt.Errorf("failed to analyze project: %w", err)
		}
	} else if strings.HasSuffix(absPath, ".go") {
		// Analyze single file - convert to project format for consistency
		fileComplexity, err := analyzer.AnalyzeFile(absPath)
		if err != nil {
			return fmt.Errorf("failed to analyze file: %w", err)
		}

		// Wrap in project structure
		packagePath := filepath.Dir(absPath)
		projectComplexity = &metrics.ProjectComplexity{
			ProjectRoot: packagePath,
			Packages: []metrics.PackageComplexity{
				{
					PackagePath: packagePath,
					Files:       []metrics.FileComplexity{*fileComplexity},
				},
			},
		}

		// Calculate package and project metrics
		analyzer.SetThresholds(thresholds) // Ensure thresholds are set
		// Note: We need access to internal methods, so we'll create a simple calculation
		pkg := &projectComplexity.Packages[0]
		pkg.TotalLinesOfCode = fileComplexity.LinesOfCode
		pkg.TotalFunctions = len(fileComplexity.Functions)
		if len(fileComplexity.Functions) > 0 {
			pkg.AverageCyclomaticComplexity = fileComplexity.AverageCyclomaticComplexity
		}
		pkg.MaintainabilityIndex = fileComplexity.MaintainabilityIndex
		pkg.TechnicalDebtHours = float64(fileComplexity.TechnicalDebtMinutes) / 60.0
		pkg.Violations = fileComplexity.Violations

		// Calculate project summary
		summary := &projectComplexity.Summary
		summary.TotalFiles = 1
		summary.TotalFunctions = pkg.TotalFunctions
		summary.TotalLinesOfCode = pkg.TotalLinesOfCode
		summary.AverageCyclomaticComplexity = pkg.AverageCyclomaticComplexity
		summary.MaintainabilityIndex = pkg.MaintainabilityIndex
		summary.TechnicalDebtDays = pkg.TechnicalDebtHours / 8.0
		summary.ViolationCount = len(pkg.Violations)
	} else {
		return fmt.Errorf("path must be a directory or .go file: %s", absPath)
	}

	if complexityVerbose {
		fmt.Printf("✅ Analysis complete: %d packages, %d functions, %d violations\n",
			len(projectComplexity.Packages),
			projectComplexity.Summary.TotalFunctions,
			projectComplexity.Summary.ViolationCount)
	}

	// Determine output destination
	var outputFile *os.File
	if complexityOutput != "" {
		outputFile, err = os.Create(complexityOutput)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputFile.Close()
	} else {
		outputFile = os.Stdout
	}

	// Generate report in requested format
	switch strings.ToLower(complexityFormat) {
	case "json":
		err = analyzer.GenerateJSONReport(projectComplexity, outputFile)
	case "html":
		err = analyzer.GenerateHTMLReport(projectComplexity, outputFile)
	case "text", "":
		err = analyzer.GenerateReport(projectComplexity, outputFile)
	default:
		return fmt.Errorf("unsupported format: %s (supported: text, json, html)", complexityFormat)
	}

	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	if complexityOutput != "" {
		fmt.Printf("📊 Complexity report saved to: %s\n", complexityOutput)
	}

	// Exit with non-zero code if critical violations found
	criticalViolations := 0
	for _, pkg := range projectComplexity.Packages {
		for _, violation := range pkg.Violations {
			if violation.Severity == "Critical" {
				criticalViolations++
			}
		}
	}

	if criticalViolations > 0 {
		fmt.Fprintf(os.Stderr, "⚠️  Found %d critical complexity violations\n", criticalViolations)
		os.Exit(1)
	}

	// Show quick summary for successful analysis
	if complexityFormat == "text" && complexityOutput == "" {
		summary := projectComplexity.Summary
		fmt.Printf("\n🎯 Quality Grade: %s | Complexity: %.2f | Maintainability: %.2f | Debt: %.2f days\n",
			summary.QualityGrade,
			summary.AverageCyclomaticComplexity,
			summary.MaintainabilityIndex,
			summary.TechnicalDebtDays)
	}

	return nil
}
