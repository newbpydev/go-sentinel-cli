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
	if d == 0 {
		return "0ms" // Display 0ms instead of 0µs for zero durations
	}

	seconds := d.Seconds()
	milliseconds := d.Milliseconds()
	microseconds := d.Microseconds()

	// Use Vitest-style formatting:
	// - Very short: use ms without decimal (1ms)
	// - Medium: use ms without decimal (123ms)
	// - Longer: use seconds with decimal (1.2s)
	if seconds >= 1 {
		// For >= 1 second, use decimal seconds (like 1.2s)
		if seconds < 10 {
			return fmt.Sprintf("%.1fs", seconds)
		}
		return fmt.Sprintf("%.0fs", seconds)
	} else if milliseconds > 0 {
		// For milliseconds range
		return fmt.Sprintf("%dms", milliseconds)
	} else {
		// For microseconds range
		return fmt.Sprintf("%dµs", microseconds)
	}
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
