package ui

import "fmt"

// FormatDurationSmart returns ms if <1000ms, otherwise seconds with 2 decimals (e.g. 1320ms -> 1.32s)
func FormatDurationSmart(seconds float64) string {
	ms := int(seconds * 1000)
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.2fs", float64(ms)/1000)
}

// FormatCoverage takes a float64 (0.0-1.0 or 0-100) and returns a string with 2 decimal places and a percent sign.
func FormatCoverage(coverage float64) string {
	if coverage > 1.0 {
		return fmt.Sprintf("%.2f%%", coverage)
	}
	return fmt.Sprintf("%.2f%%", coverage*100)
}

// AverageCoverage computes average coverage for a slice of test nodes (0-1 float), including zeros
func AverageCoverage(nodes []*TreeNode) float64 {
	if len(nodes) == 0 {
		return 0
	}
	sum := 0.0
	for _, n := range nodes {
		sum += n.Coverage
	}
	return sum / float64(len(nodes))
}

// TotalDuration sums durations for a slice of test nodes (seconds)
func TotalDuration(nodes []*TreeNode) float64 {
	var total float64
	for _, n := range nodes {
		if len(n.Children) == 0 {
			total += n.Duration
		}
	}
	return total
}
