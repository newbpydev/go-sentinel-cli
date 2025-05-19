package handlers

import (
	"testing"
	"time"
)

func TestGetMetricsData_StructureAndTypes(t *testing.T) {
	data := GetMetricsData()
	if data == nil {
		t.Fatal("expected non-nil map")
	}
	fields := []string{"TotalTests", "TotalChange", "Passing", "PassingRate", "Failing", "FailingChange", "Duration", "DurationChange", "LastUpdated"}
	for _, f := range fields {
		if _, ok := data[f]; !ok {
			t.Errorf("missing field: %s", f)
		}
	}
	if _, ok := data["LastUpdated"].(time.Time); !ok {
		t.Errorf("LastUpdated should be time.Time, got %T", data["LastUpdated"])
	}
}

func TestGetMockTestResults_Contents(t *testing.T) {
	results := GetMockTestResults()
	if len(results) != 4 {
		t.Fatalf("expected 4 mock test results, got %d", len(results))
	}
	for _, r := range results {
		if r.Name == "" {
			t.Error("test result missing Name")
		}
		if r.Status != "passed" && r.Status != "failed" {
			t.Errorf("unexpected Status: %s", r.Status)
		}
		if r.Duration == "" {
			t.Error("test result missing Duration")
		}
		if r.LastRun.IsZero() {
			t.Error("test result LastRun is zero")
		}
	}
	// Edge: check failed test has Output
	foundFail := false
	for _, r := range results {
		if r.Status == "failed" {
			foundFail = true
			if r.Output == "" {
				t.Error("failed test should have Output")
			}
		}
	}
	if !foundFail {
		t.Error("expected at least one failed test result")
	}
}
