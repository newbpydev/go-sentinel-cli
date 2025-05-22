package cli

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
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
		phaseTiming   map[string]time.Duration
		wantTestFiles bool
		wantTests     bool
		wantStartAt   bool
		wantDuration  bool
		wantPhases    bool
	}{
		{
			name:        "renders complete summary",
			testFiles:   8,
			filesPassed: 7,
			filesFailed: 1,
			testsTotal:  78,
			testsPassed: 70,
			testsFailed: 8,
			startTime:   time.Date(2023, 1, 1, 11, 39, 32, 0, time.Local),
			duration:    26*time.Second + 170*time.Millisecond,
			phaseTiming: map[string]time.Duration{
				"transform":   859 * time.Millisecond,
				"setup":       34*time.Second + 400*time.Millisecond,
				"collect":     1*time.Second + 290*time.Millisecond,
				"tests":       1 * time.Second,
				"environment": 70*time.Second + 910*time.Millisecond,
				"prepare":     3*time.Second + 690*time.Millisecond,
			},
			wantTestFiles: true,
			wantTests:     true,
			wantStartAt:   true,
			wantDuration:  true,
			wantPhases:    true,
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
			phaseTiming:   map[string]time.Duration{},
			wantTestFiles: true,
			wantTests:     true,
			wantStartAt:   true,
			wantDuration:  true,
			wantPhases:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			var buf bytes.Buffer
			formatter := NewColorFormatter(false) // No colors for testing
			renderer := NewSummaryRenderer(&buf, formatter)

			// Create test stats
			stats := &TestRunStats{
				TotalFiles:    tt.testFiles,
				PassedFiles:   tt.filesPassed,
				FailedFiles:   tt.filesFailed,
				TotalTests:    tt.testsTotal,
				PassedTests:   tt.testsPassed,
				FailedTests:   tt.testsFailed,
				StartTime:     tt.startTime,
				Duration:      tt.duration,
				PhaseDuration: tt.phaseTiming,
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

			// Check phases
			if tt.wantPhases && len(tt.phaseTiming) > 0 {
				for phase, duration := range tt.phaseTiming {
					if !strings.Contains(output, phase) {
						t.Errorf("expected phase %q in output, but not found in: %s", phase, output)
					}

					durationStr := formatExpectedDuration(duration)
					if !strings.Contains(output, durationStr) {
						t.Errorf("expected duration %q for phase %q, but not found in: %s", durationStr, phase, output)
					}
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
	seconds := float64(d) / float64(time.Second)
	return fmt.Sprintf("%.2fs", seconds)
}
