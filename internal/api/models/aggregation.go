package models

// TestResultSummary represents aggregated statistics for a set of test results.
type TestResultSummary struct {
	Total    int     `json:"total"`
	Passed   int     `json:"passed"`
	Failed   int     `json:"failed"`
	Duration float64 `json:"duration"`
}

// Aggregatable is an interface for types that expose pass/fail and duration.
type Aggregatable interface {
	IsPassed() bool
	GetDuration() float64
}

// AggregateTestResults computes a summary from a slice of Aggregatable test results.
func AggregateTestResults(results []Aggregatable) TestResultSummary {
	sum := TestResultSummary{}
	for _, r := range results {
		sum.Total++
		if r.IsPassed() {
			sum.Passed++
		} else {
			sum.Failed++
		}
		sum.Duration += r.GetDuration()
	}
	return sum
}
