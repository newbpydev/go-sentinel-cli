package ui

import (
	"testing"
)

func TestFormatDurationSmart(t *testing.T) {
	cases := []struct {
		seconds float64
		expect  string
	}{
		{0, "0ms"},
		{0.001, "1ms"},
		{0.0123, "12ms"},
		{0.099, "99ms"},
		{0.999, "999ms"},
		{1.0, "1.00s"},
		{1.234, "1.23s"},
		{12.345, "12.35s"},
		{123.456, "123.46s"},
	}
	for _, c := range cases {
		got := FormatDurationSmart(c.seconds)
		if got != c.expect {
			t.Errorf("FormatDurationSmart(%v) = %q, want %q", c.seconds, got, c.expect)
		}
	}
}

func TestFormatCoverage(t *testing.T) {
	cases := []struct {
		coverage float64
		expect   string
	}{
		{0, "0.00%"},
		{0.1234, "12.34%"},
		{1, "100.00%"},
		{99.5, "99.50%"},
	}
	for _, c := range cases {
		got := FormatCoverage(c.coverage)
		if got != c.expect {
			t.Errorf("FormatCoverage(%v) = %q, want %q", c.coverage, got, c.expect)
		}
	}
}

func TestAverageCoverage(t *testing.T) {
	nodes := []*TreeNode{
		{Coverage: 0.5},
		{Coverage: 1.0},
		{Coverage: 0.0},
		{Coverage: 0.25},
	}
	avg := AverageCoverage(nodes)
	if avg < 0.4374 || avg > 0.4376 {
		t.Errorf("AverageCoverage got %v, want ~0.4375", avg)
	}
	// Empty or all zero
	if got := AverageCoverage([]*TreeNode{}); got != 0 {
		t.Errorf("AverageCoverage([]) = %v, want 0", got)
	}
	if got := AverageCoverage([]*TreeNode{{Coverage: 0}, {Coverage: 0}}); got != 0 {
		t.Errorf("AverageCoverage(all zero) = %v, want 0", got)
	}
}

func TestTotalDuration(t *testing.T) {
	nodes := []*TreeNode{
		{Duration: 0.1, Children: nil},
		{Duration: 1.5, Children: nil},
		{Duration: 0.4, Children: nil},
		{Duration: 0.0, Children: nil},
		{Duration: 0.2, Children: []*TreeNode{{}}}, // not leaf, should skip
	}
	total := TotalDuration(nodes)
	if total < 1.999 || total > 2.001 {
		t.Errorf("TotalDuration got %v, want ~2.0", total)
	}
	if got := TotalDuration([]*TreeNode{}); got != 0 {
		t.Errorf("TotalDuration([]) = %v, want 0", got)
	}
}

func BenchmarkFormatDurationSmart(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = FormatDurationSmart(float64(i%2000) / 100.0)
	}
}

func BenchmarkAverageCoverage(b *testing.B) {
	nodes := make([]*TreeNode, 10000)
	for i := range nodes {
		nodes[i] = &TreeNode{Coverage: float64(i%100) / 100}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AverageCoverage(nodes)
	}
}

func BenchmarkTotalDuration(b *testing.B) {
	nodes := make([]*TreeNode, 10000)
	for i := range nodes {
		nodes[i] = &TreeNode{Duration: float64(i%100) / 100}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TotalDuration(nodes)
	}
}
