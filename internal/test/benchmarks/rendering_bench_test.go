package benchmarks

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// BenchmarkColorFormatter benchmarks color formatting performance
func BenchmarkColorFormatter(b *testing.B) {
	formatter := colors.NewColorFormatter(true)

	testStrings := []string{
		"PASS",
		"FAIL",
		"SKIP",
		"package github.com/test/example",
		"TestExample",
		"Error: something went wrong",
		"Warning: this is a warning",
		"Info: informational message",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, str := range testStrings {
			_ = formatter.Green(str)
			_ = formatter.Red(str)
			_ = formatter.Yellow(str)
			_ = formatter.Blue(str)
			_ = formatter.Bold(str)
			_ = formatter.Dim(str)
		}
	}
}

// BenchmarkIconProvider benchmarks icon rendering performance
func BenchmarkIconProvider(b *testing.B) {
	providers := []*colors.IconProvider{
		colors.NewIconProvider(true),  // Unicode
		colors.NewIconProvider(false), // ASCII
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, provider := range providers {
			_ = provider.CheckMark()
			_ = provider.Cross()
			_ = provider.Skipped()
			_ = provider.Running()
			_ = provider.GetIcon("package")
			_ = provider.GetIcon("test")
			_ = provider.GetIcon("watch")
			_ = provider.GetIcon("info")
			_ = provider.GetIcon("summary")
		}
	}
}

// BenchmarkSuiteRendering benchmarks rendering performance for different suite sizes
func BenchmarkSuiteRendering(b *testing.B) {
	suiteSizes := []int{10, 50, 100, 500}

	for _, size := range suiteSizes {
		b.Run(fmt.Sprintf("SuiteSize_%d", size), func(b *testing.B) {
			suite := createBenchmarkTestSuite(size, fmt.Sprintf("suite_%d_test.go", size))

			formatter := colors.NewColorFormatter(false)
			icons := colors.NewIconProvider(false)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				var buf bytes.Buffer
				// Simulate suite rendering by formatting each test
				renderTestSuite(&buf, formatter, icons, suite, 80)
			}
		})
	}
}

// BenchmarkFailedTestRendering benchmarks rendering of failed tests
func BenchmarkFailedTestRendering(b *testing.B) {
	// Create failed tests with various complexities
	testCases := []struct {
		name        string
		errorLength int
	}{
		{"SimpleError", 50},
		{"MediumError", 200},
		{"ComplexError", 500},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			failedTests := createFailedBenchmarkTests(10, tc.errorLength)

			formatter := colors.NewColorFormatter(false)
			icons := colors.NewIconProvider(false)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				var buf bytes.Buffer
				// Simulate failed test rendering
				renderFailedTests(&buf, formatter, icons, failedTests)
			}
		})
	}
}

// BenchmarkTerminalOutput benchmarks actual terminal output performance
func BenchmarkTerminalOutput(b *testing.B) {
	var buf bytes.Buffer

	// Large output simulation
	suite := createBenchmarkTestSuite(200, "large_output_test.go")
	formatter := colors.NewColorFormatter(true) // With colors
	icons := colors.NewIconProvider(true)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		renderTestSuite(&buf, formatter, icons, suite, 80)
	}
}

// BenchmarkConcurrentRendering benchmarks concurrent rendering operations
func BenchmarkConcurrentRendering(b *testing.B) {
	suites := make([]*models.TestSuite, 10)
	for i := 0; i < 10; i++ {
		suites[i] = createBenchmarkTestSuite(50, fmt.Sprintf("concurrent_%d_test.go", i))
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		formatter := colors.NewColorFormatter(false)
		icons := colors.NewIconProvider(false)

		i := 0
		for pb.Next() {
			suite := suites[i%len(suites)]
			var buf bytes.Buffer
			renderTestSuite(&buf, formatter, icons, suite, 80)
			i++
		}
	})
}

// BenchmarkTextProcessing benchmarks text processing operations
func BenchmarkTextProcessing(b *testing.B) {
	// Various text processing scenarios
	texts := []string{
		"Simple text",
		"Text with ANSI \033[31mcolor\033[0m codes",
		"Very long text that contains multiple words and punctuation marks, designed to test wrapping and processing performance under realistic conditions",
		"Text\nwith\nmultiple\nlines\nand\nbreaks",
		"Mixed content: numbers 123456, symbols !@#$%^&*(), and Unicode: 🎯📝✅❌",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		for _, text := range texts {
			// Simulate common text processing operations
			_ = removeANSIColors(text)
			_ = calculateDisplayWidth(text)
			_ = truncateAtWordBoundary(text, 50)
			_ = escapeSpecialChars(text)
		}
	}
}

// BenchmarkBufferedOutput benchmarks buffered vs unbuffered output
func BenchmarkBufferedOutput(b *testing.B) {
	suite := createBenchmarkTestSuite(100, "buffered_test.go")
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)

	b.Run("Buffered", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			renderTestSuite(&buf, formatter, icons, suite, 80)
		}
	})

	b.Run("Direct", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			renderTestSuite(io.Discard, formatter, icons, suite, 80)
		}
	})
}

// Helper functions for benchmark setup

func createBenchmarkTestSuite(testCount int, filePath string) *models.TestSuite {
	suite := &models.TestSuite{
		FilePath:     filePath,
		TestCount:    testCount,
		PassedCount:  int(float64(testCount) * 0.9),
		FailedCount:  int(float64(testCount) * 0.1),
		SkippedCount: 0,
		Duration:     time.Duration(testCount) * 10 * time.Millisecond,
		Tests:        make([]*models.LegacyTestResult, 0, testCount),
	}

	for i := 0; i < testCount; i++ {
		status := models.TestStatusPassed
		if i%10 == 0 {
			status = models.TestStatusFailed
		}

		test := &models.LegacyTestResult{
			Name:     fmt.Sprintf("TestBenchmark%d", i),
			Status:   status,
			Duration: 10 * time.Millisecond,
			Package:  "github.com/test/benchmark",
		}

		if status == models.TestStatusFailed {
			test.Error = &models.LegacyTestError{
				Message: "Test assertion failed",
				Type:    "AssertionError",
			}
		}

		suite.Tests = append(suite.Tests, test)
	}

	return suite
}

func createFailedBenchmarkTests(count, errorLength int) []*models.LegacyTestResult {
	tests := make([]*models.LegacyTestResult, count)

	for i := 0; i < count; i++ {
		// Create error message
		errorMsg := strings.Repeat(fmt.Sprintf("Error part %d. ", i), errorLength/20)

		tests[i] = &models.LegacyTestResult{
			Name:   fmt.Sprintf("TestFailed%d", i),
			Status: models.TestStatusFailed,
			Error: &models.LegacyTestError{
				Type:    "AssertionError",
				Message: errorMsg,
			},
		}
	}

	return tests
}

// Mock rendering functions for benchmarking
func renderTestSuite(writer io.Writer, formatter *colors.ColorFormatter, icons *colors.IconProvider, suite *models.TestSuite, width int) {
	// Simulate suite rendering
	fmt.Fprintf(writer, "%s %s\n", icons.GetIcon("package"), formatter.Bold(suite.FilePath))

	for _, test := range suite.Tests {
		icon := icons.CheckMark()
		color := formatter.Green

		if test.Status == models.TestStatusFailed {
			icon = icons.Cross()
			color = formatter.Red
		} else if test.Status == models.TestStatusSkipped {
			icon = icons.Skipped()
			color = formatter.Yellow
		}

		fmt.Fprintf(writer, "  %s %s (%v)\n", icon, color(test.Name), test.Duration)

		if test.Error != nil {
			fmt.Fprintf(writer, "    %s\n", formatter.Red(test.Error.Message))
		}
	}

	// Summary
	fmt.Fprintf(writer, "\n%s Summary: %d total, %d passed, %d failed, %d skipped (%v)\n",
		icons.GetIcon("summary"),
		suite.TestCount,
		suite.PassedCount,
		suite.FailedCount,
		suite.SkippedCount,
		suite.Duration,
	)
}

func renderFailedTests(writer io.Writer, formatter *colors.ColorFormatter, icons *colors.IconProvider, tests []*models.LegacyTestResult) {
	fmt.Fprintf(writer, "%s Failed Tests:\n", icons.Cross())

	for _, test := range tests {
		fmt.Fprintf(writer, "\n%s %s\n", icons.Cross(), formatter.Red(test.Name))
		if test.Error != nil {
			fmt.Fprintf(writer, "  %s: %s\n",
				formatter.Bold("Error"),
				formatter.Red(test.Error.Message))
		}
	}
}

// Mock text processing functions for benchmarking
func removeANSIColors(text string) string {
	// Simplified ANSI removal for benchmarking
	result := text
	for strings.Contains(result, "\033[") {
		start := strings.Index(result, "\033[")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "m")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return result
}

func calculateDisplayWidth(text string) int {
	// Simplified width calculation - count visible characters
	clean := removeANSIColors(text)
	return len(clean)
}

func truncateAtWordBoundary(text string, maxWidth int) string {
	if len(text) <= maxWidth {
		return text
	}

	// Find last word boundary
	for i := maxWidth; i > 0; i-- {
		if text[i] == ' ' {
			return text[:i]
		}
	}
	return text[:maxWidth]
}

func escapeSpecialChars(text string) string {
	// Simplified escape for common characters
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	return text
}
