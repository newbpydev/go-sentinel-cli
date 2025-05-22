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
// For test results display (uses ms for small durations, s for larger ones)
func FormatDurationPrecise(d time.Duration) string {
	seconds := d.Seconds()

	if seconds < 0.01 && d.Microseconds() > 0 {
		// For very small durations (< 10ms), show microseconds
		return fmt.Sprintf("%dµs", d.Microseconds())
	} else if seconds < 1 {
		// For durations less than 1 second, show milliseconds without decimal
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	// For durations >= 1 second, show seconds with 2 decimal places
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
