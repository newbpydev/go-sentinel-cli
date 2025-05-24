package cli

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"
)

// BenchmarkColorFormatter benchmarks color formatting performance
func BenchmarkColorFormatter(b *testing.B) {
	formatter := NewColorFormatter(true)

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
	providers := []*IconProvider{
		NewIconProvider(true),  // Unicode
		NewIconProvider(false), // ASCII
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

			formatter := NewColorFormatter(false)
			icons := NewIconProvider(false)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				var buf bytes.Buffer
				renderer := NewSuiteRenderer(&buf, formatter, icons, 80)
				_ = renderer.RenderSuite(suite, false)
			}
		})
	}
}

// BenchmarkIncrementalRendering benchmarks incremental rendering performance
func BenchmarkIncrementalRendering(b *testing.B) {
	cache := NewTestResultCache()
	renderer := NewIncrementalRenderer(
		io.Discard,
		NewColorFormatter(false),
		NewIconProvider(false),
		80,
		cache,
	)

	// Create test suites
	suites := make(map[string]*TestSuite)
	for i := 0; i < 10; i++ {
		suites[fmt.Sprintf("test_%d.go", i)] = createBenchmarkTestSuite(50, fmt.Sprintf("test_%d.go", i))
	}

	// Create file changes
	changes := []*FileChange{
		{Path: "test_0.go", Type: ChangeTypeTest},
		{Path: "test_1.go", Type: ChangeTypeTest},
		{Path: "main.go", Type: ChangeTypeSource},
	}

	stats := &TestRunStats{
		TotalTests:  500,
		PassedTests: 450,
		FailedTests: 50,
		TotalFiles:  10,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = renderer.RenderIncrementalResults(suites, stats, changes)
	}
}

// BenchmarkFailedTestRendering benchmarks rendering of failed tests with source context
func BenchmarkFailedTestRendering(b *testing.B) {
	// Create failed tests with various complexities
	testCases := []struct {
		name        string
		contextSize int
		errorLength int
	}{
		{"SimpleError", 5, 50},
		{"MediumError", 10, 200},
		{"ComplexError", 20, 500},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			failedTests := createFailedBenchmarkTests(10, tc.contextSize, tc.errorLength)

			formatter := NewColorFormatter(false)
			icons := NewIconProvider(false)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				var buf bytes.Buffer
				renderer := NewFailedTestRenderer(&buf, formatter, icons, 80)
				_ = renderer.RenderFailedTests(failedTests)
			}
		})
	}
}

// BenchmarkTerminalOutput benchmarks actual terminal output performance
func BenchmarkTerminalOutput(b *testing.B) {
	var buf bytes.Buffer

	// Large output simulation
	suite := createBenchmarkTestSuite(200, "large_output_test.go")
	formatter := NewColorFormatter(true) // With colors
	icons := NewIconProvider(true)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		renderer := NewSuiteRenderer(&buf, formatter, icons, 80)
		_ = renderer.RenderSuite(suite, false)
	}
}

// BenchmarkConcurrentRendering benchmarks concurrent rendering operations
func BenchmarkConcurrentRendering(b *testing.B) {
	suites := make([]*TestSuite, 10)
	for i := 0; i < 10; i++ {
		suites[i] = createBenchmarkTestSuite(50, fmt.Sprintf("concurrent_%d_test.go", i))
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		formatter := NewColorFormatter(false)
		icons := NewIconProvider(false)

		i := 0
		for pb.Next() {
			suite := suites[i%len(suites)]
			var buf bytes.Buffer
			renderer := NewSuiteRenderer(&buf, formatter, icons, 80)
			_ = renderer.RenderSuite(suite, false)
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
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)

	b.Run("Buffered", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			renderer := NewSuiteRenderer(&buf, formatter, icons, 80)
			_ = renderer.RenderSuite(suite, false)
		}
	})

	b.Run("Direct", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			renderer := NewSuiteRenderer(io.Discard, formatter, icons, 80)
			_ = renderer.RenderSuite(suite, false)
		}
	})
}

// Helper functions for benchmark setup

func createBenchmarkTestSuite(testCount int, filePath string) *TestSuite {
	suite := &TestSuite{
		FilePath:     filePath,
		TestCount:    testCount,
		PassedCount:  int(float64(testCount) * 0.9),
		FailedCount:  int(float64(testCount) * 0.1),
		SkippedCount: 0,
		Duration:     time.Duration(testCount) * 10 * time.Millisecond,
		Tests:        make([]*TestResult, 0, testCount),
	}

	for i := 0; i < testCount; i++ {
		status := StatusPassed
		if i%10 == 0 {
			status = StatusFailed
		}

		test := &TestResult{
			Name:     fmt.Sprintf("TestBenchmark%d", i),
			Status:   status,
			Duration: 10 * time.Millisecond,
			Package:  "github.com/test/benchmark",
		}

		if status == StatusFailed {
			test.Error = &TestError{
				Message: "Test assertion failed",
				Type:    "AssertionError",
				Location: &SourceLocation{
					File:   filePath,
					Line:   20 + i,
					Column: 16,
				},
			}
		}

		suite.Tests = append(suite.Tests, test)
	}

	return suite
}

func createFailedBenchmarkTests(count, contextSize, errorLength int) []*TestResult {
	tests := make([]*TestResult, count)

	for i := 0; i < count; i++ {
		// Create source context
		context := make([]string, contextSize)
		for j := 0; j < contextSize; j++ {
			context[j] = fmt.Sprintf("    line %d: some code here", j+1)
		}

		// Create error message
		errorMsg := strings.Repeat(fmt.Sprintf("Error part %d. ", i), errorLength/20)

		tests[i] = &TestResult{
			Name:   fmt.Sprintf("TestFailed%d", i),
			Status: StatusFailed,
			Error: &TestError{
				Type:    "AssertionError",
				Message: errorMsg,
				Location: &SourceLocation{
					File:   "test_file.go",
					Line:   10 + i,
					Column: 16,
				},
				SourceContext:   context,
				HighlightedLine: contextSize / 2,
			},
		}
	}

	return tests
}

// Mock text processing functions for benchmarking
func removeANSIColors(text string) string {
	// Simplified ANSI removal
	return strings.ReplaceAll(text, "\033[", "")
}

func calculateDisplayWidth(text string) int {
	// Simplified width calculation
	return len(text)
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
	// Simplified escape
	return strings.ReplaceAll(text, "&", "&amp;")
}
