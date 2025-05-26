// Package display provides timing formatting for test execution display
package display

import (
	"fmt"
	"time"
)

// TimingFormatter handles formatting durations for test display
type TimingFormatter struct {
	config *TimingConfig
}

// TimingConfig configures timing display behavior
type TimingConfig struct {
	ShowMilliseconds bool // Show milliseconds (e.g., "150ms")
	ShowMicroseconds bool // Show microseconds for very short durations
	IntegerFormat    bool // Use integer format (e.g., "150ms" instead of "150.5ms")
	MinWidth         int  // Minimum width for alignment
}

// NewTimingFormatter creates a new timing formatter with configuration
func NewTimingFormatter(config *TimingConfig) *TimingFormatter {
	if config == nil {
		config = &TimingConfig{
			ShowMilliseconds: true,
			ShowMicroseconds: false,
			IntegerFormat:    true,
			MinWidth:         4,
		}
	}

	return &TimingFormatter{
		config: config,
	}
}

// FormatDuration formats a duration according to the configuration
func (t *TimingFormatter) FormatDuration(duration time.Duration) string {
	if duration == 0 {
		return t.formatWithWidth("0ms")
	}

	// Convert to appropriate unit based on magnitude
	switch {
	case duration >= time.Second:
		return t.formatSeconds(duration)
	case duration >= time.Millisecond || !t.config.ShowMicroseconds:
		return t.formatMilliseconds(duration)
	case duration >= time.Microsecond:
		return t.formatMicroseconds(duration)
	default:
		return t.formatNanoseconds(duration)
	}
}

// formatSeconds formats durations in seconds
func (t *TimingFormatter) formatSeconds(duration time.Duration) string {
	seconds := duration.Seconds()

	if t.config.IntegerFormat && seconds == float64(int(seconds)) {
		return t.formatWithWidth(fmt.Sprintf("%ds", int(seconds)))
	}

	return t.formatWithWidth(fmt.Sprintf("%.1fs", seconds))
}

// formatMilliseconds formats durations in milliseconds
func (t *TimingFormatter) formatMilliseconds(duration time.Duration) string {
	milliseconds := float64(duration.Nanoseconds()) / 1e6

	if t.config.IntegerFormat {
		return t.formatWithWidth(fmt.Sprintf("%dms", int(milliseconds)))
	}

	// Show decimal places for sub-millisecond precision
	if milliseconds < 1.0 {
		return t.formatWithWidth(fmt.Sprintf("%.1fms", milliseconds))
	}

	return t.formatWithWidth(fmt.Sprintf("%.0fms", milliseconds))
}

// formatMicroseconds formats durations in microseconds
func (t *TimingFormatter) formatMicroseconds(duration time.Duration) string {
	microseconds := float64(duration.Nanoseconds()) / 1e3

	if t.config.IntegerFormat {
		return t.formatWithWidth(fmt.Sprintf("%dμs", int(microseconds)))
	}

	return t.formatWithWidth(fmt.Sprintf("%.1fμs", microseconds))
}

// formatNanoseconds formats durations in nanoseconds
func (t *TimingFormatter) formatNanoseconds(duration time.Duration) string {
	nanoseconds := duration.Nanoseconds()
	return t.formatWithWidth(fmt.Sprintf("%dns", nanoseconds))
}

// formatWithWidth ensures the formatted string meets minimum width requirements
func (t *TimingFormatter) formatWithWidth(formatted string) string {
	if len(formatted) >= t.config.MinWidth {
		return formatted
	}

	// Right-align by default for timing values
	padding := t.config.MinWidth - len(formatted)
	return fmt.Sprintf("%*s", padding, "") + formatted
}

// FormatDurationRange formats a range of durations (e.g., "5ms - 150ms")
func (t *TimingFormatter) FormatDurationRange(min, max time.Duration) string {
	minFormatted := t.FormatDuration(min)
	maxFormatted := t.FormatDuration(max)

	if min == max {
		return minFormatted
	}

	return fmt.Sprintf("%s - %s", minFormatted, maxFormatted)
}

// FormatAverageDuration formats an average duration with indication
func (t *TimingFormatter) FormatAverageDuration(duration time.Duration) string {
	formatted := t.FormatDuration(duration)
	return fmt.Sprintf("avg %s", formatted)
}

// FormatTotalDuration formats a total duration with indication
func (t *TimingFormatter) FormatTotalDuration(duration time.Duration) string {
	formatted := t.FormatDuration(duration)
	return fmt.Sprintf("total %s", formatted)
}

// ParseDurationString attempts to parse a duration string back to time.Duration
func (t *TimingFormatter) ParseDurationString(durationStr string) (time.Duration, error) {
	// Try to parse as Go duration format first
	if duration, err := time.ParseDuration(durationStr); err == nil {
		return duration, nil
	}

	// Could add custom parsing logic here if needed
	return 0, fmt.Errorf("unable to parse duration: %s", durationStr)
}

// GetFormattedZero returns the formatted representation of zero duration
func (t *TimingFormatter) GetFormattedZero() string {
	return t.FormatDuration(0)
}

// SetMinWidth updates the minimum width for formatting
func (t *TimingFormatter) SetMinWidth(width int) {
	t.config.MinWidth = width
}

// SetIntegerFormat enables/disables integer formatting
func (t *TimingFormatter) SetIntegerFormat(enabled bool) {
	t.config.IntegerFormat = enabled
}

// SetShowMicroseconds enables/disables microsecond display
func (t *TimingFormatter) SetShowMicroseconds(enabled bool) {
	t.config.ShowMicroseconds = enabled
}

// GetMinWidth returns the current minimum width setting
func (t *TimingFormatter) GetMinWidth() int {
	return t.config.MinWidth
}

// IsIntegerFormat returns whether integer formatting is enabled
func (t *TimingFormatter) IsIntegerFormat() bool {
	return t.config.IntegerFormat
}

// IsShowMicroseconds returns whether microsecond display is enabled
func (t *TimingFormatter) IsShowMicroseconds() bool {
	return t.config.ShowMicroseconds
}
