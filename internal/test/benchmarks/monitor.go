package benchmarks

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// PerformanceMonitor tracks benchmark results and detects regressions
type PerformanceMonitor struct {
	baselineFile string
	thresholds   RegressionThresholds
}

// RegressionThresholds defines acceptable performance degradation limits
type RegressionThresholds struct {
	MaxSlowdownPercent float64 // Maximum acceptable slowdown percentage
	MaxMemoryIncrease  float64 // Maximum acceptable memory increase percentage
	MinSampleSize      int     // Minimum number of samples for reliable comparison
}

// BenchmarkResult represents a single benchmark result
type BenchmarkResult struct {
	Name        string    `json:"name"`
	Iterations  int       `json:"iterations"`
	NsPerOp     float64   `json:"ns_per_op"`
	BytesPerOp  int64     `json:"bytes_per_op"`
	AllocsPerOp int64     `json:"allocs_per_op"`
	MBPerSec    float64   `json:"mb_per_sec,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	GitCommit   string    `json:"git_commit,omitempty"`
	GoVersion   string    `json:"go_version,omitempty"`
	OS          string    `json:"os,omitempty"`
	Arch        string    `json:"arch,omitempty"`
}

// PerformanceReport contains analysis of benchmark results
type PerformanceReport struct {
	GeneratedAt     time.Time                   `json:"generated_at"`
	TotalBenchmarks int                         `json:"total_benchmarks"`
	Regressions     []RegressionAlert           `json:"regressions"`
	Improvements    []ImprovementAlert          `json:"improvements"`
	Summary         PerformanceSummary          `json:"summary"`
	Trends          map[string]PerformanceTrend `json:"trends"`
}

// RegressionAlert indicates a performance regression
type RegressionAlert struct {
	BenchmarkName   string  `json:"benchmark_name"`
	Severity        string  `json:"severity"` // "CRITICAL", "MAJOR", "MINOR"
	SlowdownPercent float64 `json:"slowdown_percent"`
	MemoryIncrease  float64 `json:"memory_increase_percent"`
	PreviousNsPerOp float64 `json:"previous_ns_per_op"`
	CurrentNsPerOp  float64 `json:"current_ns_per_op"`
	PreviousMemory  int64   `json:"previous_memory"`
	CurrentMemory   int64   `json:"current_memory"`
	Recommendation  string  `json:"recommendation"`
}

// ImprovementAlert indicates a performance improvement
type ImprovementAlert struct {
	BenchmarkName      string  `json:"benchmark_name"`
	ImprovementPercent float64 `json:"improvement_percent"`
	MemoryReduction    float64 `json:"memory_reduction_percent"`
	CurrentNsPerOp     float64 `json:"current_ns_per_op"`
	PreviousNsPerOp    float64 `json:"previous_ns_per_op"`
}

// PerformanceSummary provides overall performance metrics
type PerformanceSummary struct {
	AverageSlowdown     float64 `json:"average_slowdown_percent"`
	AverageImprovement  float64 `json:"average_improvement_percent"`
	TotalRegressions    int     `json:"total_regressions"`
	CriticalRegressions int     `json:"critical_regressions"`
	TotalImprovements   int     `json:"total_improvements"`
	OverallTrend        string  `json:"overall_trend"` // "IMPROVING", "STABLE", "DEGRADING"
}

// PerformanceTrend tracks performance over time for a specific benchmark
type PerformanceTrend struct {
	BenchmarkName  string      `json:"benchmark_name"`
	DataPoints     []float64   `json:"data_points"`
	Timestamps     []time.Time `json:"timestamps"`
	TrendDirection string      `json:"trend_direction"` // "UP", "DOWN", "STABLE"
	ChangePercent  float64     `json:"change_percent"`
	Volatility     float64     `json:"volatility"`
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(baselineFile string) *PerformanceMonitor {
	return &PerformanceMonitor{
		baselineFile: baselineFile,
		thresholds: RegressionThresholds{
			MaxSlowdownPercent: 20.0, // 20% slowdown threshold
			MaxMemoryIncrease:  25.0, // 25% memory increase threshold
			MinSampleSize:      3,    // Minimum 3 samples for comparison
		},
	}
}

// SetThresholds updates the regression detection thresholds
func (pm *PerformanceMonitor) SetThresholds(thresholds RegressionThresholds) {
	pm.thresholds = thresholds
}

// ParseBenchmarkOutput parses Go benchmark output and extracts results
func (pm *PerformanceMonitor) ParseBenchmarkOutput(output string) ([]BenchmarkResult, error) {
	var results []BenchmarkResult
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "Benchmark") && strings.Contains(line, "ns/op") {
			result, err := pm.parseBenchmarkLine(line)
			if err != nil {
				continue // Skip malformed lines
			}
			result.Timestamp = time.Now()
			results = append(results, result)
		}
	}

	return results, nil
}

// parseBenchmarkLine parses a single benchmark result line
func (pm *PerformanceMonitor) parseBenchmarkLine(line string) (BenchmarkResult, error) {
	// Example: BenchmarkProcessorParse-8    1000000    1234 ns/op    456 B/op    7 allocs/op
	parts := strings.Fields(line)
	if len(parts) < 4 {
		return BenchmarkResult{}, fmt.Errorf("invalid benchmark line format")
	}

	result := BenchmarkResult{
		Name: parts[0],
	}

	// Parse iterations
	if iterations, err := strconv.Atoi(parts[1]); err == nil {
		result.Iterations = iterations
	}

	// Parse ns/op
	if nsPerOp, err := strconv.ParseFloat(parts[2], 64); err == nil {
		result.NsPerOp = nsPerOp
	}

	// Parse additional metrics if present
	for i := 4; i < len(parts); i += 2 {
		if i+1 < len(parts) {
			value := parts[i]
			unit := parts[i+1]

			switch unit {
			case "B/op":
				if bytes, err := strconv.ParseInt(value, 10, 64); err == nil {
					result.BytesPerOp = bytes
				}
			case "allocs/op":
				if allocs, err := strconv.ParseInt(value, 10, 64); err == nil {
					result.AllocsPerOp = allocs
				}
			case "MB/s":
				if mbPerSec, err := strconv.ParseFloat(value, 64); err == nil {
					result.MBPerSec = mbPerSec
				}
			}
		}
	}

	return result, nil
}

// CompareWithBaseline compares current results with baseline and generates a report
func (pm *PerformanceMonitor) CompareWithBaseline(currentResults []BenchmarkResult) (*PerformanceReport, error) {
	baselineResults, err := pm.loadBaseline()
	if err != nil {
		// No baseline exists, create one
		return pm.createInitialReport(currentResults), nil
	}

	report := &PerformanceReport{
		GeneratedAt:     time.Now(),
		TotalBenchmarks: len(currentResults),
		Trends:          make(map[string]PerformanceTrend),
	}

	// Create lookup map for baseline results
	baselineMap := make(map[string]BenchmarkResult)
	for _, result := range baselineResults {
		baselineMap[result.Name] = result
	}

	var totalSlowdown, totalImprovement float64
	var regressionCount, improvementCount int

	// Compare each current result with baseline
	for _, current := range currentResults {
		baseline, exists := baselineMap[current.Name]
		if !exists {
			continue // New benchmark, skip comparison
		}

		// Calculate performance change
		slowdownPercent := ((current.NsPerOp - baseline.NsPerOp) / baseline.NsPerOp) * 100
		memoryChangePercent := float64(0)
		if baseline.BytesPerOp > 0 {
			memoryChangePercent = ((float64(current.BytesPerOp) - float64(baseline.BytesPerOp)) / float64(baseline.BytesPerOp)) * 100
		}

		// Check for regressions
		if slowdownPercent > pm.thresholds.MaxSlowdownPercent || memoryChangePercent > pm.thresholds.MaxMemoryIncrease {
			severity := pm.calculateSeverity(slowdownPercent, memoryChangePercent)
			regression := RegressionAlert{
				BenchmarkName:   current.Name,
				Severity:        severity,
				SlowdownPercent: slowdownPercent,
				MemoryIncrease:  memoryChangePercent,
				PreviousNsPerOp: baseline.NsPerOp,
				CurrentNsPerOp:  current.NsPerOp,
				PreviousMemory:  baseline.BytesPerOp,
				CurrentMemory:   current.BytesPerOp,
				Recommendation:  pm.generateRecommendation(severity, slowdownPercent, memoryChangePercent),
			}
			report.Regressions = append(report.Regressions, regression)
			totalSlowdown += slowdownPercent
			regressionCount++
		} else if slowdownPercent < -5.0 { // 5% improvement threshold
			improvement := ImprovementAlert{
				BenchmarkName:      current.Name,
				ImprovementPercent: -slowdownPercent,
				MemoryReduction:    -memoryChangePercent,
				CurrentNsPerOp:     current.NsPerOp,
				PreviousNsPerOp:    baseline.NsPerOp,
			}
			report.Improvements = append(report.Improvements, improvement)
			totalImprovement += -slowdownPercent
			improvementCount++
		}
	}

	// Calculate summary
	report.Summary = PerformanceSummary{
		TotalRegressions:    len(report.Regressions),
		TotalImprovements:   len(report.Improvements),
		CriticalRegressions: pm.countCriticalRegressions(report.Regressions),
		OverallTrend:        pm.determineOverallTrend(totalSlowdown, totalImprovement, regressionCount, improvementCount),
	}

	if regressionCount > 0 {
		report.Summary.AverageSlowdown = totalSlowdown / float64(regressionCount)
	}
	if improvementCount > 0 {
		report.Summary.AverageImprovement = totalImprovement / float64(improvementCount)
	}

	return report, nil
}

// calculateSeverity determines the severity of a performance regression
func (pm *PerformanceMonitor) calculateSeverity(slowdownPercent, memoryChangePercent float64) string {
	maxChange := math.Max(slowdownPercent, memoryChangePercent)

	if maxChange > 50.0 {
		return "CRITICAL"
	} else if maxChange > 30.0 {
		return "MAJOR"
	}
	return "MINOR"
}

// generateRecommendation provides actionable recommendations for performance issues
func (pm *PerformanceMonitor) generateRecommendation(severity string, slowdownPercent, memoryChangePercent float64) string {
	if slowdownPercent > memoryChangePercent {
		switch severity {
		case "CRITICAL":
			return "URGENT: Profile the function to identify CPU bottlenecks. Consider algorithmic improvements."
		case "MAJOR":
			return "Review recent changes for performance impact. Run CPU profiler to identify hotspots."
		default:
			return "Monitor trend. Consider micro-optimizations if pattern continues."
		}
	} else {
		switch severity {
		case "CRITICAL":
			return "URGENT: Memory leak detected. Review memory allocations and object lifecycle."
		case "MAJOR":
			return "Investigate memory usage patterns. Check for unnecessary allocations."
		default:
			return "Monitor memory usage trend. Consider object pooling if appropriate."
		}
	}
}

// countCriticalRegressions counts the number of critical regressions
func (pm *PerformanceMonitor) countCriticalRegressions(regressions []RegressionAlert) int {
	count := 0
	for _, regression := range regressions {
		if regression.Severity == "CRITICAL" {
			count++
		}
	}
	return count
}

// determineOverallTrend determines the overall performance trend
func (pm *PerformanceMonitor) determineOverallTrend(totalSlowdown, totalImprovement float64, regressionCount, improvementCount int) string {
	if regressionCount > improvementCount*2 {
		return "DEGRADING"
	} else if improvementCount > regressionCount*2 {
		return "IMPROVING"
	}
	return "STABLE"
}

// createInitialReport creates a report for the first benchmark run
func (pm *PerformanceMonitor) createInitialReport(results []BenchmarkResult) *PerformanceReport {
	return &PerformanceReport{
		GeneratedAt:     time.Now(),
		TotalBenchmarks: len(results),
		Regressions:     []RegressionAlert{},
		Improvements:    []ImprovementAlert{},
		Summary: PerformanceSummary{
			OverallTrend: "BASELINE",
		},
		Trends: make(map[string]PerformanceTrend),
	}
}

// SaveBaseline saves current results as the new baseline
func (pm *PerformanceMonitor) SaveBaseline(results []BenchmarkResult) error {
	// Ensure directory exists
	dir := filepath.Dir(pm.baselineFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create baseline directory: %w", err)
	}

	file, err := os.Create(pm.baselineFile)
	if err != nil {
		return fmt.Errorf("failed to create baseline file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(results)
}

// loadBaseline loads the baseline results from file
func (pm *PerformanceMonitor) loadBaseline() ([]BenchmarkResult, error) {
	file, err := os.Open(pm.baselineFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []BenchmarkResult
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&results)
	return results, err
}

// GenerateTextReport generates a human-readable text report
func (pm *PerformanceMonitor) GenerateTextReport(report *PerformanceReport, output io.Writer) error {
	fmt.Fprintf(output, "# Performance Monitoring Report\n")
	fmt.Fprintf(output, "Generated: %s\n\n", report.GeneratedAt.Format(time.RFC3339))

	// Summary
	fmt.Fprintf(output, "## Summary\n")
	fmt.Fprintf(output, "- **Total Benchmarks**: %d\n", report.TotalBenchmarks)
	fmt.Fprintf(output, "- **Regressions**: %d (%d critical)\n", report.Summary.TotalRegressions, report.Summary.CriticalRegressions)
	fmt.Fprintf(output, "- **Improvements**: %d\n", report.Summary.TotalImprovements)
	fmt.Fprintf(output, "- **Overall Trend**: %s\n\n", report.Summary.OverallTrend)

	// Regressions
	if len(report.Regressions) > 0 {
		fmt.Fprintf(output, "## ⚠️ Performance Regressions\n\n")

		// Sort by severity and slowdown
		sort.Slice(report.Regressions, func(i, j int) bool {
			if report.Regressions[i].Severity != report.Regressions[j].Severity {
				// Define severity order: CRITICAL > MAJOR > MINOR
				severityOrder := map[string]int{
					"CRITICAL": 0,
					"MAJOR":    1,
					"MINOR":    2,
				}
				return severityOrder[report.Regressions[i].Severity] < severityOrder[report.Regressions[j].Severity]
			}
			return report.Regressions[i].SlowdownPercent > report.Regressions[j].SlowdownPercent
		})

		for _, regression := range report.Regressions {
			fmt.Fprintf(output, "### %s - %s\n", regression.BenchmarkName, regression.Severity)
			fmt.Fprintf(output, "- **Slowdown**: %.1f%%\n", regression.SlowdownPercent)
			if regression.MemoryIncrease > 0 {
				fmt.Fprintf(output, "- **Memory Increase**: %.1f%%\n", regression.MemoryIncrease)
			}
			fmt.Fprintf(output, "- **Previous**: %.0f ns/op\n", regression.PreviousNsPerOp)
			fmt.Fprintf(output, "- **Current**: %.0f ns/op\n", regression.CurrentNsPerOp)
			fmt.Fprintf(output, "- **Recommendation**: %s\n\n", regression.Recommendation)
		}
	}

	// Improvements
	if len(report.Improvements) > 0 {
		fmt.Fprintf(output, "## ✅ Performance Improvements\n\n")

		for _, improvement := range report.Improvements {
			fmt.Fprintf(output, "### %s\n", improvement.BenchmarkName)
			fmt.Fprintf(output, "- **Improvement**: %.1f%%\n", improvement.ImprovementPercent)
			if improvement.MemoryReduction > 0 {
				fmt.Fprintf(output, "- **Memory Reduction**: %.1f%%\n", improvement.MemoryReduction)
			}
			fmt.Fprintf(output, "- **Previous**: %.0f ns/op\n", improvement.PreviousNsPerOp)
			fmt.Fprintf(output, "- **Current**: %.0f ns/op\n\n", improvement.CurrentNsPerOp)
		}
	}

	return nil
}

// GenerateJSONReport generates a JSON report
func (pm *PerformanceMonitor) GenerateJSONReport(report *PerformanceReport, output io.Writer) error {
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
