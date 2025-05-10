package runner

import "fmt"

// FormatMillis converts seconds (float64) to milliseconds string, e.g. 0.0123 -> "12ms"
func FormatMillis(seconds float64) string {
	return fmt.Sprintf("%dms", int(seconds*1000))
}

// FormatCoverage takes a float64 (0.0-1.0 or 0-100) and returns a string with 2 decimal places and a percent sign.
func FormatCoverage(coverage float64) string {
	if coverage > 1.0 {
		return fmt.Sprintf("%.2f%%", coverage)
	}
	return fmt.Sprintf("%.2f%%", coverage*100)
}
