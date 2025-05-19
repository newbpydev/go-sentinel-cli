package runner

import "testing"

func TestFormatMillis_Conversion(t *testing.T) {
	tests := []struct {
		name     string
		seconds  float64
		expected string
	}{
		{"zero", 0.0, "0ms"},
		{"one_second", 1.0, "1000ms"},
		{"fraction", 0.0123, "12ms"},
		{"small_fraction", 0.001, "1ms"},
		{"large_number", 123.456, "123456ms"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatMillis(tc.seconds)
			if result != tc.expected {
				t.Errorf("FormatMillis(%v) = %v, want %v", tc.seconds, result, tc.expected)
			}
		})
	}
}

func TestFormatCoverage_Percentage(t *testing.T) {
	tests := []struct {
		name     string
		coverage float64
		expected string
	}{
		{"zero", 0.0, "0.00%"},
		{"hundred_percent", 1.0, "100.00%"},
		{"fifty_percent", 0.5, "50.00%"},
		{"decimal_percent", 0.7523, "75.23%"},
		{"already_percentage", 75.23, "75.23%"},
		{"over_hundred", 150.0, "150.00%"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatCoverage(tc.coverage)
			if result != tc.expected {
				t.Errorf("FormatCoverage(%v) = %v, want %v", tc.coverage, result, tc.expected)
			}
		})
	}
}
