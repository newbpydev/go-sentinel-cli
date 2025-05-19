package handlers

import (
	"time"
)

// GetMetricsData returns mock metrics data for the dashboard
// This is a utility function used by multiple handlers
func GetMetricsData() map[string]interface{} {
	return map[string]interface{}{
		"TotalTests":     "128",
		"TotalChange":    "+3 since yesterday",
		"Passing":        "119",
		"PassingRate":    "93% success rate",
		"Failing":        "9",
		"FailingChange":  "-2 since yesterday",
		"Duration":       "1.2s",
		"DurationChange": "-0.3s from last run",
		"LastUpdated":    time.Now(),
	}
}

// GetMockTestResults returns mock test results for the dashboard
// This is a utility function used by multiple handlers
func GetMockTestResults() []TestResult {
	now := time.Now()

	return []TestResult{
		{
			Name:     "TestParseConfig",
			Status:   "passed",
			Duration: "0.8s",
			LastRun:  now.Add(-2 * time.Minute),
		},
		{
			Name:     "TestValidateInput",
			Status:   "passed",
			Duration: "1.2s",
			LastRun:  now.Add(-2 * time.Minute),
		},
		{
			Name:     "TestProcessResults",
			Status:   "failed",
			Duration: "2.1s",
			LastRun:  now.Add(-2 * time.Minute),
			Output:   "Expected 5 results, got 4",
		},
		{
			Name:     "TestExportReport",
			Status:   "passed",
			Duration: "0.9s",
			LastRun:  now.Add(-2 * time.Minute),
		},
	}
}
