package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/newbpydev/go-sentinel/internal/test/benchmarks"
	"github.com/spf13/cobra"
)

var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Run performance benchmarks with regression detection",
	Long: `Run comprehensive performance benchmarks and detect performance regressions.

This command runs the full benchmark suite, compares results with baseline,
and generates detailed performance reports with actionable recommendations.

Examples:
  # Run benchmarks and compare with baseline
  go-sentinel benchmark

  # Run benchmarks and save as new baseline
  go-sentinel benchmark --save-baseline

  # Run benchmarks with custom thresholds
  go-sentinel benchmark --max-slowdown=15 --max-memory-increase=20

  # Generate detailed performance report
  go-sentinel benchmark --format=json --output=performance-report.json

  # Run specific benchmark packages
  go-sentinel benchmark --packages="./internal/test/processor,./internal/ui/display"`,
	RunE: runBenchmark,
}

var (
	benchmarkFormat       string
	benchmarkOutput       string
	benchmarkPackages     string
	benchmarkSaveBaseline bool
	benchmarkMaxSlowdown  float64
	benchmarkMaxMemory    float64
	benchmarkBenchtime    string
	benchmarkCount        int
	benchmarkVerbose      bool
	benchmarkBaselineFile string
)

func init() {
	rootCmd.AddCommand(benchmarkCmd)

	benchmarkCmd.Flags().StringVar(&benchmarkFormat, "format", "text", "Output format (text, json)")
	benchmarkCmd.Flags().StringVar(&benchmarkOutput, "output", "", "Output file (default: stdout)")
	benchmarkCmd.Flags().StringVar(&benchmarkPackages, "packages", "", "Comma-separated list of packages to benchmark (default: all)")
	benchmarkCmd.Flags().BoolVar(&benchmarkSaveBaseline, "save-baseline", false, "Save current results as new baseline")
	benchmarkCmd.Flags().Float64Var(&benchmarkMaxSlowdown, "max-slowdown", 20.0, "Maximum acceptable slowdown percentage")
	benchmarkCmd.Flags().Float64Var(&benchmarkMaxMemory, "max-memory-increase", 25.0, "Maximum acceptable memory increase percentage")
	benchmarkCmd.Flags().StringVar(&benchmarkBenchtime, "benchtime", "1s", "Benchmark duration per test")
	benchmarkCmd.Flags().IntVar(&benchmarkCount, "count", 5, "Number of benchmark iterations")
	benchmarkCmd.Flags().BoolVar(&benchmarkVerbose, "verbose", false, "Verbose benchmark output")
	benchmarkCmd.Flags().StringVar(&benchmarkBaselineFile, "baseline-file", "build/benchmarks/baseline.json", "Baseline file path")
}

func runBenchmark(cmd *cobra.Command, args []string) error {
	// Create performance monitor
	monitor := benchmarks.NewPerformanceMonitor(benchmarkBaselineFile)

	// Set custom thresholds if provided
	thresholds := benchmarks.RegressionThresholds{
		MaxSlowdownPercent: benchmarkMaxSlowdown,
		MaxMemoryIncrease:  benchmarkMaxMemory,
		MinSampleSize:      3,
	}
	monitor.SetThresholds(thresholds)

	// Determine packages to benchmark
	packages := getBenchmarkPackages()

	fmt.Printf("🚀 Running performance benchmarks...\n")
	fmt.Printf("📦 Packages: %s\n", strings.Join(packages, ", "))
	fmt.Printf("⏱️  Benchtime: %s, Count: %d\n", benchmarkBenchtime, benchmarkCount)
	fmt.Printf("📊 Thresholds: %.1f%% slowdown, %.1f%% memory increase\n\n",
		benchmarkMaxSlowdown, benchmarkMaxMemory)

	// Run benchmarks
	benchmarkOutput, err := runBenchmarkSuite(packages)
	if err != nil {
		return fmt.Errorf("failed to run benchmarks: %w", err)
	}

	// Parse benchmark results
	results, err := monitor.ParseBenchmarkOutput(benchmarkOutput)
	if err != nil {
		return fmt.Errorf("failed to parse benchmark results: %w", err)
	}

	if len(results) == 0 {
		fmt.Println("⚠️  No benchmark results found. Make sure benchmark functions exist in the specified packages.")
		return nil
	}

	fmt.Printf("✅ Parsed %d benchmark results\n\n", len(results))

	// Save baseline if requested
	if benchmarkSaveBaseline {
		if err := monitor.SaveBaseline(results); err != nil {
			return fmt.Errorf("failed to save baseline: %w", err)
		}
		fmt.Printf("💾 Saved baseline to %s\n", benchmarkBaselineFile)
		return nil
	}

	// Compare with baseline and generate report
	report, err := monitor.CompareWithBaseline(results)
	if err != nil {
		return fmt.Errorf("failed to compare with baseline: %w", err)
	}

	// Generate and output report
	return generateBenchmarkReport(monitor, report)
}

func getBenchmarkPackages() []string {
	if benchmarkPackages != "" {
		return strings.Split(benchmarkPackages, ",")
	}

	// Default benchmark packages
	return []string{
		"./internal/test/benchmarks/...",
		"./internal/test/processor",
		"./internal/test/runner",
		"./internal/ui/display",
		"./internal/watch/core",
	}
}

func runBenchmarkSuite(packages []string) (string, error) {
	// Prepare benchmark command
	args := []string{"test", "-bench=.", "-benchmem"}
	args = append(args, fmt.Sprintf("-benchtime=%s", benchmarkBenchtime))
	args = append(args, fmt.Sprintf("-count=%d", benchmarkCount))
	args = append(args, "-run=^$") // Don't run regular tests

	if benchmarkVerbose {
		args = append(args, "-v")
	}

	// Add packages
	args = append(args, packages...)

	// Run benchmark command
	cmd := exec.Command("go", args...)
	output, err := cmd.CombinedOutput()

	if benchmarkVerbose {
		fmt.Printf("🔧 Command: go %s\n", strings.Join(args, " "))
		fmt.Printf("📝 Raw output:\n%s\n", string(output))
	}

	return string(output), err
}

func generateBenchmarkReport(monitor *benchmarks.PerformanceMonitor, report *benchmarks.PerformanceReport) error {
	// Determine output destination
	var outputFile *os.File
	var err error

	if benchmarkOutput != "" {
		// Ensure output directory exists
		dir := filepath.Dir(benchmarkOutput)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		outputFile, err = os.Create(benchmarkOutput)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer outputFile.Close()
	} else {
		outputFile = os.Stdout
	}

	// Generate report based on format
	switch benchmarkFormat {
	case "json":
		err = monitor.GenerateJSONReport(report, outputFile)
	case "text":
		err = monitor.GenerateTextReport(report, outputFile)
	default:
		return fmt.Errorf("unsupported format: %s (supported: text, json)", benchmarkFormat)
	}

	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Print summary to stderr if outputting to file
	if benchmarkOutput != "" {
		printBenchmarkSummary(report)
	}

	// Check for critical regressions and exit with error code
	if report.Summary.CriticalRegressions > 0 {
		fmt.Fprintf(os.Stderr, "\n❌ Critical performance regressions detected! (%d critical)\n",
			report.Summary.CriticalRegressions)
		os.Exit(1)
	}

	return nil
}

func printBenchmarkSummary(report *benchmarks.PerformanceReport) {
	fmt.Fprintf(os.Stderr, "\n📊 Performance Benchmark Summary\n")
	fmt.Fprintf(os.Stderr, "================================\n")
	fmt.Fprintf(os.Stderr, "Total Benchmarks: %d\n", report.TotalBenchmarks)
	fmt.Fprintf(os.Stderr, "Regressions: %d (%d critical)\n",
		report.Summary.TotalRegressions, report.Summary.CriticalRegressions)
	fmt.Fprintf(os.Stderr, "Improvements: %d\n", report.Summary.TotalImprovements)
	fmt.Fprintf(os.Stderr, "Overall Trend: %s\n", report.Summary.OverallTrend)

	if report.Summary.TotalRegressions > 0 {
		fmt.Fprintf(os.Stderr, "\n⚠️  Performance regressions detected!\n")
		for _, regression := range report.Regressions {
			if regression.Severity == "CRITICAL" {
				fmt.Fprintf(os.Stderr, "  🔴 %s: %.1f%% slower (%s)\n",
					regression.BenchmarkName, regression.SlowdownPercent, regression.Severity)
			}
		}
	}

	if report.Summary.TotalImprovements > 0 {
		fmt.Fprintf(os.Stderr, "\n✅ Performance improvements detected!\n")
		for _, improvement := range report.Improvements {
			fmt.Fprintf(os.Stderr, "  🟢 %s: %.1f%% faster\n",
				improvement.BenchmarkName, improvement.ImprovementPercent)
		}
	}

	if benchmarkOutput != "" {
		fmt.Fprintf(os.Stderr, "\n📄 Detailed report saved to: %s\n", benchmarkOutput)
	}
}
