package models

import (
	"testing"
)

type DummyResult struct {
	Passed   bool
	Duration float64
}

type Summary struct {
	Total    int
	Passed   int
	Failed   int
	Duration float64
}

func AggregateResults(results []DummyResult) Summary {
	sum := Summary{}
	for _, r := range results {
		sum.Total++
		if r.Passed {
			sum.Passed++
		} else {
			sum.Failed++
		}
		sum.Duration += r.Duration
	}
	return sum
}

func TestAggregateResults_SummaryCounts(t *testing.T) {
	results := []DummyResult{
		{Passed: true, Duration: 1.2},
		{Passed: false, Duration: 0.8},
		{Passed: true, Duration: 2.0},
	}
	sum := AggregateResults(results)
	if sum.Total != 3 || sum.Passed != 2 || sum.Failed != 1 {
		t.Errorf("unexpected counts: %+v", sum)
	}
	if sum.Duration != 4.0 {
		t.Errorf("unexpected duration: got %v", sum.Duration)
	}
}

func TestAggregateResults_Empty(t *testing.T) {
	results := []DummyResult{}
	sum := AggregateResults(results)
	if sum.Total != 0 || sum.Passed != 0 || sum.Failed != 0 || sum.Duration != 0 {
		t.Errorf("unexpected summary for empty input: %+v", sum)
	}
}

func TestAggregateResults_AllFail(t *testing.T) {
	results := []DummyResult{
		{Passed: false, Duration: 0.5},
		{Passed: false, Duration: 0.7},
	}
	sum := AggregateResults(results)
	if sum.Passed != 0 || sum.Failed != 2 {
		t.Errorf("all fail: %+v", sum)
	}
}

// MockTestResult implements Aggregatable for testing
type MockTestResult struct {
	passed   bool
	duration float64
}

func (m MockTestResult) IsPassed() bool {
	return m.passed
}

func (m MockTestResult) GetDuration() float64 {
	return m.duration
}

func TestAggregateTestResults_Basic(t *testing.T) {
	results := []Aggregatable{
		MockTestResult{passed: true, duration: 1.2},
		MockTestResult{passed: false, duration: 0.8},
		MockTestResult{passed: true, duration: 2.0},
	}

	summary := AggregateTestResults(results)

	if summary.Total != 3 {
		t.Errorf("expected Total=3, got %d", summary.Total)
	}
	if summary.Passed != 2 {
		t.Errorf("expected Passed=2, got %d", summary.Passed)
	}
	if summary.Failed != 1 {
		t.Errorf("expected Failed=1, got %d", summary.Failed)
	}
	if summary.Duration != 4.0 {
		t.Errorf("expected Duration=4.0, got %f", summary.Duration)
	}
}

// MockResult implements Aggregatable for testing
type MockResult struct {
	passed   bool
	duration float64
}

func (m MockResult) IsPassed() bool {
	return m.passed
}

func (m MockResult) GetDuration() float64 {
	return m.duration
}

func TestAggregateTestResults_Empty(t *testing.T) {
	summary := AggregateTestResults(nil)
	if summary.Total != 0 {
		t.Errorf("expected Total=0, got %d", summary.Total)
	}
	if summary.Passed != 0 {
		t.Errorf("expected Passed=0, got %d", summary.Passed)
	}
	if summary.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", summary.Failed)
	}
	if summary.Duration != 0 {
		t.Errorf("expected Duration=0, got %f", summary.Duration)
	}
}

func TestAggregateTestResults_Mixed(t *testing.T) {
	results := []Aggregatable{
		MockResult{passed: true, duration: 1.5},
		MockResult{passed: false, duration: 0.5},
		MockResult{passed: true, duration: 2.0},
		MockResult{passed: false, duration: 1.0},
	}

	summary := AggregateTestResults(results)

	if summary.Total != 4 {
		t.Errorf("expected Total=4, got %d", summary.Total)
	}
	if summary.Passed != 2 {
		t.Errorf("expected Passed=2, got %d", summary.Passed)
	}
	if summary.Failed != 2 {
		t.Errorf("expected Failed=2, got %d", summary.Failed)
	}
	if summary.Duration != 5.0 {
		t.Errorf("expected Duration=5.0, got %f", summary.Duration)
	}
}

func TestAggregateTestResults_AllPassed(t *testing.T) {
	results := []Aggregatable{
		MockResult{passed: true, duration: 1.0},
		MockResult{passed: true, duration: 2.0},
		MockResult{passed: true, duration: 3.0},
	}

	summary := AggregateTestResults(results)

	if summary.Total != 3 {
		t.Errorf("expected Total=3, got %d", summary.Total)
	}
	if summary.Passed != 3 {
		t.Errorf("expected Passed=3, got %d", summary.Passed)
	}
	if summary.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", summary.Failed)
	}
	if summary.Duration != 6.0 {
		t.Errorf("expected Duration=6.0, got %f", summary.Duration)
	}
}

func TestAggregateTestResults_AllFailed(t *testing.T) {
	results := []Aggregatable{
		MockResult{passed: false, duration: 0.1},
		MockResult{passed: false, duration: 0.2},
		MockResult{passed: false, duration: 0.3},
	}

	summary := AggregateTestResults(results)

	if summary.Total != 3 {
		t.Errorf("expected Total=3, got %d", summary.Total)
	}
	if summary.Passed != 0 {
		t.Errorf("expected Passed=0, got %d", summary.Passed)
	}
	if summary.Failed != 3 {
		t.Errorf("expected Failed=3, got %d", summary.Failed)
	}
	if summary.Duration != 0.6 {
		t.Errorf("expected Duration=0.6, got %f", summary.Duration)
	}
}

func TestAggregateTestResults_SingleResult(t *testing.T) {
	results := []Aggregatable{
		MockResult{passed: true, duration: 1.5},
	}

	summary := AggregateTestResults(results)

	if summary.Total != 1 {
		t.Errorf("expected Total=1, got %d", summary.Total)
	}
	if summary.Passed != 1 {
		t.Errorf("expected Passed=1, got %d", summary.Passed)
	}
	if summary.Failed != 0 {
		t.Errorf("expected Failed=0, got %d", summary.Failed)
	}
	if summary.Duration != 1.5 {
		t.Errorf("expected Duration=1.5, got %f", summary.Duration)
	}
}
