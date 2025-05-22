package cli

import (
	"fmt"
	"time"
)

// DurationFromSeconds converts a float64 seconds value to time.Duration
func DurationFromSeconds(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
}

// FormatDurationPrecise formats a duration with appropriate units and precision
// For test results display (always shows 2 decimal places)
func FormatDurationPrecise(d time.Duration) string {
	seconds := d.Seconds()
	// Always show seconds with 2 decimal places for consistency
	return fmt.Sprintf("%.2fs", seconds)
}

// FormatDurationAdaptive formats a duration with adaptive precision
// For summary display (shows ms for sub-second, s for >=1s)
func FormatDurationAdaptive(d time.Duration) string {
	if d < 0 {
		d = -d // Use absolute duration for formatting
	}

	if d < time.Millisecond {
		return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1e6)
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

// AggregateDurations sums up a slice of durations
func AggregateDurations(durations []time.Duration) time.Duration {
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total
}
