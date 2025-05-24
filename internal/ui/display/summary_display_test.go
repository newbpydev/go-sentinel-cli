package display

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
)

func TestSummaryRenderer(t *testing.T) {
	tests := []struct {
		name          string
		testFiles     int
		filesPassed   int
		filesFailed   int
		testsTotal    int
		testsPassed   int
		testsFailed   int
		startTime     time.Time
		duration      time.Duration
		wantTestFiles bool
		wantTests     bool
		wantStartAt   bool
		wantDuration  bool
	}{
		{
			name:          "renders complete summary",
			testFiles:     8,
			filesPassed:   7,
			filesFailed:   1,
			testsTotal:    78,
			testsPassed:   70,
			testsFailed:   8,
			startTime:     time.Date(2023, 1, 1, 11, 39, 32, 0, time.Local),
			duration:      26*time.Second + 170*time.Millisecond,
			wantTestFiles: true,
			wantTests:     true,
			wantStartAt:   true,
			wantDuration:  true,
		},
		{
			name:          "handles zero tests",
			testFiles:     0,
			filesPassed:   0,
			filesFailed:   0,
			testsTotal:    0,
			testsPassed:   0,
			testsFailed:   0,
			startTime:     time.Date(2023, 1, 1, 11, 39, 32, 0, time.Local),
			duration:      0,
			wantTestFiles: true,
			wantTests:     true,
			wantStartAt:   true,
			wantDuration:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			var buf bytes.Buffer
			formatter := colors.NewColorFormatter(false) // No colors for testing
			icons := colors.NewIconProvider(false)
			renderer := NewSummaryRenderer(&buf, formatter, icons, 80)

			// Create test stats
			stats := &TestRunStats{
				TotalFiles:  tt.testFiles,
				PassedFiles: tt.filesPassed,
				FailedFiles: tt.filesFailed,
				TotalTests:  tt.testsTotal,
				PassedTests: tt.testsPassed,
				FailedTests: tt.testsFailed,
				StartTime:   tt.startTime,
				Duration:    tt.duration,
			}

			// Execute
			err := renderer.RenderSummary(stats)

			// Assert
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()

			// Check test files line
			if tt.wantTestFiles {
				expected := "Test Files"
				if !strings.Contains(output, expected) {
					t.Errorf("expected %q in output, but not found in: %s", expected, output)
				}

				if tt.filesFailed > 0 {
					expectedFailed := formatTestCount(tt.filesFailed, "failed")
					if !strings.Contains(output, expectedFailed) {
						t.Errorf("expected %q in output, but not found in: %s", expectedFailed, output)
					}
				}

				if tt.filesPassed > 0 {
					expectedPassed := formatTestCount(tt.filesPassed, "passed")
					if !strings.Contains(output, expectedPassed) {
						t.Errorf("expected %q in output, but not found in: %s", expectedPassed, output)
					}
				}
			}

			// Check tests line
			if tt.wantTests {
				expected := "Tests"
				if !strings.Contains(output, expected) {
					t.Errorf("expected %q in output, but not found in: %s", expected, output)
				}

				if tt.testsFailed > 0 {
					expectedFailed := formatTestCount(tt.testsFailed, "failed")
					if !strings.Contains(output, expectedFailed) {
						t.Errorf("expected %q in output, but not found in: %s", expectedFailed, output)
					}
				}

				if tt.testsPassed > 0 {
					expectedPassed := formatTestCount(tt.testsPassed, "passed")
					if !strings.Contains(output, expectedPassed) {
						t.Errorf("expected %q in output, but not found in: %s", expectedPassed, output)
					}
				}
			}

			// Check start time
			if tt.wantStartAt {
				expected := "Start at"
				if !strings.Contains(output, expected) {
					t.Errorf("expected %q in output, but not found in: %s", expected, output)
				}

				expectedTime := "11:39:32"
				if !strings.Contains(output, expectedTime) {
					t.Errorf("expected time %q in output, but not found in: %s", expectedTime, output)
				}
			}

			// Check duration
			if tt.wantDuration {
				expected := "Duration"
				if !strings.Contains(output, expected) {
					t.Errorf("expected %q in output, but not found in: %s", expected, output)
				}

				if tt.duration > 0 {
					expectedDuration := formatExpectedDuration(tt.duration)
					if !strings.Contains(output, expectedDuration) {
						t.Errorf("expected duration %q in output, but not found in: %s", expectedDuration, output)
					}
				}
			}
		})
	}
}

func TestSummaryRenderer_NilStats(t *testing.T) {
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	renderer := NewSummaryRenderer(&buf, formatter, icons, 80)

	err := renderer.RenderSummary(nil)
	if err == nil {
		t.Error("expected error for nil stats")
	}
	if !strings.Contains(err.Error(), "stats cannot be nil") {
		t.Errorf("expected 'stats cannot be nil' error, got: %v", err)
	}
}

func TestSummaryRenderer_NilConstructorArgs(t *testing.T) {
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)

	// Test nil writer
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil writer")
		}
	}()
	NewSummaryRenderer(nil, formatter, icons, 80)
}

func TestSummaryRenderer_NilFormatter(t *testing.T) {
	icons := colors.NewIconProvider(false)

	// Test nil formatter
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil formatter")
		}
	}()
	NewSummaryRenderer(&bytes.Buffer{}, nil, icons, 80)
}

func TestSummaryRenderer_NilIcons(t *testing.T) {
	formatter := colors.NewColorFormatter(false)

	// Test nil icons
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil icons")
		}
	}()
	NewSummaryRenderer(&bytes.Buffer{}, formatter, nil, 80)
}

func TestNewSummaryRendererWithDefaults(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewSummaryRendererWithDefaults(&buf)

	if renderer == nil {
		t.Fatal("expected non-nil renderer")
	}
	if renderer.writer != &buf {
		t.Error("expected writer to be set correctly")
	}
	if renderer.formatter == nil {
		t.Error("expected formatter to be set")
	}
	if renderer.icons == nil {
		t.Error("expected icons to be set")
	}
	if renderer.width != 80 {
		t.Errorf("expected width 80, got %d", renderer.width)
	}
}

func TestRenderTestFilesSummary(t *testing.T) {
	tests := []struct {
		name        string
		totalFiles  int
		passedFiles int
		failedFiles int
		wantText    []string
	}{
		{
			name:        "mixed results",
			totalFiles:  8,
			passedFiles: 7,
			failedFiles: 1,
			wantText:    []string{"Test Files", "1 failed", "7 passed", "(8)"},
		},
		{
			name:        "all passed",
			totalFiles:  5,
			passedFiles: 5,
			failedFiles: 0,
			wantText:    []string{"Test Files", "5 passed", "(5)"},
		},
		{
			name:        "all failed",
			totalFiles:  3,
			passedFiles: 0,
			failedFiles: 3,
			wantText:    []string{"Test Files", "3 failed", "(3)"},
		},
		{
			name:        "no files",
			totalFiles:  0,
			passedFiles: 0,
			failedFiles: 0,
			wantText:    []string{"Test Files"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := colors.NewColorFormatter(false)
			icons := colors.NewIconProvider(false)
			renderer := NewSummaryRenderer(&buf, formatter, icons, 80)

			err := renderer.RenderTestFilesSummary(tt.totalFiles, tt.passedFiles, tt.failedFiles)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()
			for _, want := range tt.wantText {
				if !strings.Contains(output, want) {
					t.Errorf("expected %q in output, but not found in: %s", want, output)
				}
			}
		})
	}
}

func TestRenderTestsSummary(t *testing.T) {
	tests := []struct {
		name        string
		totalTests  int
		passedTests int
		failedTests int
		wantText    []string
	}{
		{
			name:        "mixed results",
			totalTests:  78,
			passedTests: 70,
			failedTests: 8,
			wantText:    []string{"Tests", "8 failed", "70 passed", "(78)"},
		},
		{
			name:        "all passed",
			totalTests:  50,
			passedTests: 50,
			failedTests: 0,
			wantText:    []string{"Tests", "50 passed", "(50)"},
		},
		{
			name:        "no tests",
			totalTests:  0,
			passedTests: 0,
			failedTests: 0,
			wantText:    []string{"Tests"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := colors.NewColorFormatter(false)
			icons := colors.NewIconProvider(false)
			renderer := NewSummaryRenderer(&buf, formatter, icons, 80)

			err := renderer.RenderTestsSummary(tt.totalTests, tt.passedTests, tt.failedTests)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()
			for _, want := range tt.wantText {
				if !strings.Contains(output, want) {
					t.Errorf("expected %q in output, but not found in: %s", want, output)
				}
			}
		})
	}
}

func TestRenderTimingSummary(t *testing.T) {
	tests := []struct {
		name      string
		startTime time.Time
		duration  time.Duration
		wantText  []string
	}{
		{
			name:      "normal timing",
			startTime: time.Date(2023, 1, 1, 11, 39, 32, 0, time.Local),
			duration:  26*time.Second + 170*time.Millisecond,
			wantText:  []string{"Start at", "11:39:32", "Duration", "26.17s"},
		},
		{
			name:      "millisecond duration",
			startTime: time.Date(2023, 1, 1, 14, 25, 10, 0, time.Local),
			duration:  859 * time.Millisecond,
			wantText:  []string{"Start at", "14:25:10", "Duration", "859ms"},
		},
		{
			name:      "zero duration",
			startTime: time.Date(2023, 1, 1, 9, 0, 0, 0, time.Local),
			duration:  0,
			wantText:  []string{"Start at", "09:00:00", "Duration", "0ms"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := colors.NewColorFormatter(false)
			icons := colors.NewIconProvider(false)
			renderer := NewSummaryRenderer(&buf, formatter, icons, 80)

			err := renderer.RenderTimingSummary(tt.startTime, tt.duration)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			output := buf.String()
			for _, want := range tt.wantText {
				if !strings.Contains(output, want) {
					t.Errorf("expected %q in output, but not found in: %s", want, output)
				}
			}
		})
	}
}

// Helper to format count strings like "1 failed" or "7 passed"
func formatTestCount(count int, status string) string {
	return fmt.Sprintf("%d %s", count, status)
}

// Helper to format duration like "26.17s" or "859ms"
func formatExpectedDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}
