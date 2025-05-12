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
