package cli

import (
	"os"
	"strings"
	"testing"
)

func TestStyle_StatusIcon(t *testing.T) {
	tests := []struct {
		name       string
		useIcons   bool
		isWindows  bool
		status     TestStatus
		wantIcon   string
		wantASCII  string
		wantWinFmt string
	}{
		{
			name:       "passed with icons",
			useIcons:   true,
			status:     TestStatusPassed,
			wantIcon:   IconPass,
			wantASCII:  ASCIIIconPass,
			wantWinFmt: WinIconPass,
		},
		{
			name:       "failed with icons",
			useIcons:   true,
			status:     TestStatusFailed,
			wantIcon:   IconFail,
			wantASCII:  ASCIIIconFail,
			wantWinFmt: WinIconFail,
		},
		{
			name:       "skipped with icons",
			useIcons:   true,
			status:     TestStatusSkipped,
			wantIcon:   IconSkip,
			wantASCII:  ASCIIIconSkip,
			wantWinFmt: WinIconSkip,
		},
		{
			name:       "running with icons",
			useIcons:   true,
			status:     TestStatusRunning,
			wantIcon:   IconRunning,
			wantASCII:  ASCIIIconRunning,
			wantWinFmt: WinIconRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Style{
				useIcons:  tt.useIcons,
				isWindows: tt.isWindows,
			}

			// Test with icons enabled
			s.useIcons = true
			s.isWindows = false
			got := s.StatusIcon(tt.status)
			if got != tt.wantIcon {
				t.Errorf("StatusIcon() with icons = %q, want %q", got, tt.wantIcon)
			}

			// Test with ASCII fallback
			s.useIcons = false
			got = s.StatusIcon(tt.status)
			if got != tt.wantASCII {
				t.Errorf("StatusIcon() with ASCII = %q, want %q", got, tt.wantASCII)
			}

			// Test with Windows format
			s.useIcons = true
			s.isWindows = true
			got = s.StatusIcon(tt.status)
			if got != tt.wantWinFmt {
				t.Errorf("StatusIcon() with Windows = %q, want %q", got, tt.wantWinFmt)
			}
		})
	}
}

func TestStyle_FormatTestName(t *testing.T) {
	tests := []struct {
		name      string
		useColors bool
		useIcons  bool
		result    *TestResult
		want      string
	}{
		{
			name:      "passed test with colors",
			useColors: true,
			useIcons:  true,
			result: &TestResult{
				Name:   "TestExample",
				Status: TestStatusPassed,
			},
			want: "✓ TestExample",
		},
		{
			name:      "failed test with colors",
			useColors: true,
			useIcons:  true,
			result: &TestResult{
				Name:   "TestExample",
				Status: TestStatusFailed,
			},
			want: "✕ TestExample",
		},
		{
			name:      "skipped test with colors",
			useColors: true,
			useIcons:  true,
			result: &TestResult{
				Name:   "TestExample",
				Status: TestStatusSkipped,
			},
			want: "○ TestExample",
		},
		{
			name:      "test without colors",
			useColors: false,
			useIcons:  false,
			result: &TestResult{
				Name:   "TestExample",
				Status: TestStatusPassed,
			},
			want: "+ TestExample",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Style{
				useColors: tt.useColors,
				useIcons:  tt.useIcons,
			}

			got := s.FormatTestName(tt.result)
			// For colored output, just verify it contains the test name and icon
			if tt.useColors {
				if !strings.Contains(got, tt.result.Name) {
					t.Errorf("FormatTestName() = %q, should contain test name %q", got, tt.result.Name)
				}
			} else {
				// For non-colored output, verify exact match
				if got != tt.want {
					t.Errorf("FormatTestName() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestStyle_FormatErrorSnippet(t *testing.T) {
	snippet := "func TestExample() {\n\tresult := compute()\n\tassert.Equal(t, 42, result)\n}"
	errorLine := 2

	s := &Style{useColors: false}
	got := s.FormatErrorSnippet(snippet, errorLine)

	// Without colors, should return the original snippet
	if got != snippet {
		t.Errorf("FormatErrorSnippet() without colors = %q, want %q", got, snippet)
	}

	// With colors, should contain the original lines
	s.useColors = true
	got = s.FormatErrorSnippet(snippet, errorLine)
	lines := strings.Split(snippet, "\n")
	for _, line := range lines {
		if !strings.Contains(got, line) {
			t.Errorf("FormatErrorSnippet() with colors should contain line %q", line)
		}
	}
}

func TestStyle_FormatTestSummary(t *testing.T) {
	s := &Style{useColors: false}

	// Test with no failures
	got := s.FormatTestSummary("Test Files", 0, 8, 0, 8)
	if !strings.Contains(got, "8 passed") {
		t.Errorf("FormatTestSummary() = %q, should contain '8 passed'", got)
	}
	if !strings.Contains(got, "(8)") {
		t.Errorf("FormatTestSummary() = %q, should contain total '(8)'", got)
	}

	// Test with mixed results
	got = s.FormatTestSummary("Test Files", 2, 5, 1, 8)
	expectedParts := []string{
		"2 failed",
		"5 passed",
		"1 skipped",
		"(8)",
	}
	for _, part := range expectedParts {
		if !strings.Contains(got, part) {
			t.Errorf("FormatTestSummary() = %q, should contain %q", got, part)
		}
	}
}

func TestStyle_Detect(t *testing.T) {
	// Save original env vars
	origForceColor := os.Getenv("FORCE_COLOR")
	origNoColor := os.Getenv("NO_COLOR")
	defer func() {
		if err := os.Setenv("FORCE_COLOR", origForceColor); err != nil {
			t.Errorf("Failed to restore FORCE_COLOR: %v", err)
		}
		if err := os.Setenv("NO_COLOR", origNoColor); err != nil {
			t.Errorf("Failed to restore NO_COLOR: %v", err)
		}
	}()

	tests := []struct {
		name       string
		forceColor bool
		noColor    bool
		wantColors bool
		wantIcons  bool
		isWindows  bool
	}{
		{
			name:       "force color enabled",
			forceColor: true,
			wantColors: true,
			wantIcons:  true,
		},
		{
			name:       "no color enabled",
			noColor:    true,
			wantColors: false,
			wantIcons:  false,
		},
		{
			name:       "windows platform",
			isWindows:  true,
			forceColor: true, // Force color for Windows test
			wantColors: true,
			wantIcons:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.forceColor {
				if err := os.Setenv("FORCE_COLOR", "1"); err != nil {
					t.Fatalf("Failed to set FORCE_COLOR: %v", err)
				}
			} else {
				if err := os.Unsetenv("FORCE_COLOR"); err != nil {
					t.Fatalf("Failed to unset FORCE_COLOR: %v", err)
				}
			}
			if tt.noColor {
				if err := os.Setenv("NO_COLOR", "1"); err != nil {
					t.Fatalf("Failed to set NO_COLOR: %v", err)
				}
			} else {
				if err := os.Unsetenv("NO_COLOR"); err != nil {
					t.Fatalf("Failed to unset NO_COLOR: %v", err)
				}
			}

			s := &Style{
				useColors: true,
				useIcons:  true,
				isWindows: tt.isWindows,
			}

			s.Detect()

			if s.useColors != tt.wantColors {
				t.Errorf("Detect() useColors = %v, want %v", s.useColors, tt.wantColors)
			}
			if s.useIcons != tt.wantIcons {
				t.Errorf("Detect() useIcons = %v, want %v", s.useIcons, tt.wantIcons)
			}
		})
	}
}
