// Package display provides tests for the progress renderer implementation
package display

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestNewProgressRenderer(t *testing.T) {
	config := &Config{
		Output: &bytes.Buffer{},
		Colors: true,
	}

	renderer := NewProgressRenderer(config)

	if renderer == nil {
		t.Fatal("Expected renderer to be created")
	}

	if renderer.config != config {
		t.Error("Expected config to be set")
	}

	if renderer.formatter == nil {
		t.Error("Expected formatter to be initialized")
	}

	if renderer.icons == nil {
		t.Error("Expected icons to be initialized")
	}

	if renderer.width != 40 {
		t.Errorf("Expected default width 40, got %d", renderer.width)
	}

	if !renderer.showSpinner {
		t.Error("Expected spinner to be enabled by default")
	}

	if !renderer.showPercent {
		t.Error("Expected percentage to be enabled by default")
	}

	if !renderer.showEta {
		t.Error("Expected ETA to be enabled by default")
	}
}

func TestProgressRenderer_StartProgress(t *testing.T) {
	buffer := &bytes.Buffer{}
	config := &Config{
		Output: buffer,
		Colors: true,
	}

	renderer := NewProgressRenderer(config)
	ctx := context.Background()

	err := renderer.StartProgress(ctx, 10)
	if err != nil {
		t.Fatalf("StartProgress failed: %v", err)
	}

	if !renderer.IsActive() {
		t.Error("Expected renderer to be active")
	}

	if renderer.total != 10 {
		t.Errorf("Expected total 10, got %d", renderer.total)
	}

	if renderer.current != 0 {
		t.Errorf("Expected current 0, got %d", renderer.current)
	}

	if renderer.status != "Starting..." {
		t.Errorf("Expected status 'Starting...', got '%s'", renderer.status)
	}

	// Check that output was written
	output := buffer.String()
	if output == "" {
		t.Error("Expected output to be written")
	}
}

func TestProgressRenderer_UpdateProgress(t *testing.T) {
	buffer := &bytes.Buffer{}
	config := &Config{
		Output: buffer,
		Colors: true,
	}

	renderer := NewProgressRenderer(config)
	ctx := context.Background()

	// Start progress
	err := renderer.StartProgress(ctx, 10)
	if err != nil {
		t.Fatalf("StartProgress failed: %v", err)
	}

	// Clear buffer to test update output
	buffer.Reset()

	// Update progress
	err = renderer.UpdateProgress(5, "Processing items")
	if err != nil {
		t.Fatalf("UpdateProgress failed: %v", err)
	}

	if renderer.current != 5 {
		t.Errorf("Expected current 5, got %d", renderer.current)
	}

	if renderer.status != "Processing items" {
		t.Errorf("Expected status 'Processing items', got '%s'", renderer.status)
	}

	// Check that output was written
	output := buffer.String()
	if output == "" {
		t.Error("Expected output to be written")
	}

	// Check for progress bar elements
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Error("Expected progress bar brackets in output")
	}
}

func TestProgressRenderer_FinishProgress(t *testing.T) {
	buffer := &bytes.Buffer{}
	config := &Config{
		Output: buffer,
		Colors: true,
	}

	renderer := NewProgressRenderer(config)
	ctx := context.Background()

	// Start progress
	err := renderer.StartProgress(ctx, 10)
	if err != nil {
		t.Fatalf("StartProgress failed: %v", err)
	}

	// Update to completion
	err = renderer.UpdateProgress(10, "All items processed")
	if err != nil {
		t.Fatalf("UpdateProgress failed: %v", err)
	}

	// Clear buffer to test finish output
	buffer.Reset()

	// Finish progress
	err = renderer.FinishProgress()
	if err != nil {
		t.Fatalf("FinishProgress failed: %v", err)
	}

	if renderer.IsActive() {
		t.Error("Expected renderer to be inactive after finish")
	}

	if !renderer.IsFinished() {
		t.Error("Expected renderer to be marked as finished")
	}

	// Check that completion output was written
	output := buffer.String()
	if output == "" {
		t.Error("Expected completion output to be written")
	}

	if !strings.Contains(output, "Complete") {
		t.Error("Expected 'Complete' message in output")
	}
}

func TestProgressRenderer_CalculatePercentage(t *testing.T) {
	buffer := &bytes.Buffer{}
	config := &Config{
		Output: buffer,
		Colors: true,
	}

	renderer := NewProgressRenderer(config)

	// Test with zero total
	renderer.total = 0
	renderer.current = 5
	percentage := renderer.calculatePercentage()
	if percentage != 0.0 {
		t.Errorf("Expected 0%% for zero total, got %.1f%%", percentage)
	}

	// Test normal percentage
	renderer.total = 100
	renderer.current = 25
	percentage = renderer.calculatePercentage()
	if percentage != 25.0 {
		t.Errorf("Expected 25%%, got %.1f%%", percentage)
	}

	// Test 100% completion
	renderer.current = 100
	percentage = renderer.calculatePercentage()
	if percentage != 100.0 {
		t.Errorf("Expected 100%%, got %.1f%%", percentage)
	}
}

func TestProgressRenderer_CalculateETA(t *testing.T) {
	buffer := &bytes.Buffer{}
	config := &Config{
		Output: buffer,
		Colors: true,
	}

	renderer := NewProgressRenderer(config)

	// Test with zero current (no ETA possible)
	renderer.total = 100
	renderer.current = 0
	eta := renderer.calculateETA()
	if eta != 0 {
		t.Errorf("Expected 0 ETA for zero current, got %v", eta)
	}

	// Test with zero total
	renderer.total = 0
	renderer.current = 5
	eta = renderer.calculateETA()
	if eta != 0 {
		t.Errorf("Expected 0 ETA for zero total, got %v", eta)
	}

	// Test normal ETA calculation
	renderer.total = 100
	renderer.current = 50
	renderer.startTime = time.Now().Add(-10 * time.Second)
	eta = renderer.calculateETA()

	// ETA should be approximately 10 seconds (same time as elapsed)
	if eta < 5*time.Second || eta > 15*time.Second {
		t.Errorf("Expected ETA around 10s, got %v", eta)
	}
}

func TestProgressRenderer_SetSpinner(t *testing.T) {
	buffer := &bytes.Buffer{}
	config := &Config{
		Output: buffer,
		Colors: true,
	}

	renderer := NewProgressRenderer(config)

	spinnerConfig := &SpinnerConfig{
		Frames:   []string{"◐", "◓", "◑", "◒"},
		Interval: 200 * time.Millisecond,
	}

	err := renderer.SetSpinner(spinnerConfig)
	if err != nil {
		t.Fatalf("SetSpinner failed: %v", err)
	}

	if renderer.spinner != spinnerConfig {
		t.Error("Expected spinner config to be set")
	}
}

func TestProgressRenderer_GetSpinnerFrame(t *testing.T) {
	buffer := &bytes.Buffer{}
	config := &Config{
		Output: buffer,
		Colors: true,
	}

	renderer := NewProgressRenderer(config)

	// Test default spinner frames
	frame1 := renderer.getSpinnerFrame()
	if frame1 == "" {
		t.Error("Expected non-empty spinner frame")
	}

	renderer.spinnerFrame = 1
	frame2 := renderer.getSpinnerFrame()
	if frame2 == frame1 {
		t.Error("Expected different frame after incrementing")
	}

	// Test custom spinner
	customSpinner := &SpinnerConfig{
		Frames: []string{"A", "B", "C"},
	}
	renderer.SetSpinner(customSpinner)

	renderer.spinnerFrame = 0
	frame := renderer.getSpinnerFrame()
	if frame != "A" {
		t.Errorf("Expected 'A', got '%s'", frame)
	}

	renderer.spinnerFrame = 1
	frame = renderer.getSpinnerFrame()
	if frame != "B" {
		t.Errorf("Expected 'B', got '%s'", frame)
	}

	renderer.spinnerFrame = 3 // Should wrap around
	frame = renderer.getSpinnerFrame()
	if frame != "A" {
		t.Errorf("Expected 'A' (wrapped), got '%s'", frame)
	}
}

func TestProgressRenderer_RenderProgressBar(t *testing.T) {
	buffer := &bytes.Buffer{}
	config := &Config{
		Output: buffer,
		Colors: true,
	}

	renderer := NewProgressRenderer(config)
	renderer.width = 10 // Small width for testing

	// Test with zero total
	renderer.total = 0
	renderer.current = 5
	bar := renderer.renderProgressBar()
	if !strings.Contains(bar, "[") || !strings.Contains(bar, "]") {
		t.Error("Expected progress bar brackets")
	}

	// Test normal progress
	renderer.total = 10
	renderer.current = 5
	bar = renderer.renderProgressBar()
	if !strings.Contains(bar, "=") {
		t.Error("Expected filled portion in progress bar")
	}
	if !strings.Contains(bar, ">") {
		t.Error("Expected current position indicator")
	}

	// Test completed progress
	renderer.current = 10
	bar = renderer.renderProgressBar()
	if !strings.Contains(bar, "=") {
		t.Error("Expected filled progress bar")
	}
}

func TestProgressRenderer_FormatDuration(t *testing.T) {
	buffer := &bytes.Buffer{}
	config := &Config{
		Output: buffer,
		Colors: true,
	}

	renderer := NewProgressRenderer(config)

	tests := []struct {
		duration time.Duration
		expected string
	}{
		{500 * time.Millisecond, "0.5s"},
		{1500 * time.Millisecond, "1.5s"},
		{30 * time.Second, "30.0s"},
		{90 * time.Second, "1m30s"},
		{3661 * time.Second, "61m1s"},
	}

	for _, test := range tests {
		result := renderer.formatDuration(test.duration)
		if result != test.expected {
			t.Errorf("Expected '%s' for %v, got '%s'", test.expected, test.duration, result)
		}
	}
}

func TestProgressRenderer_Configuration(t *testing.T) {
	buffer := &bytes.Buffer{}
	config := &Config{
		Output: buffer,
		Colors: true,
	}

	renderer := NewProgressRenderer(config)

	// Test width configuration
	renderer.SetWidth(20)
	if renderer.width != 20 {
		t.Errorf("Expected width 20, got %d", renderer.width)
	}

	// Test show spinner configuration
	renderer.SetShowSpinner(false)
	if renderer.showSpinner {
		t.Error("Expected spinner to be disabled")
	}

	// Test show percentage configuration
	renderer.SetShowPercent(false)
	if renderer.showPercent {
		t.Error("Expected percentage to be disabled")
	}

	// Test show ETA configuration
	renderer.SetShowETA(false)
	if renderer.showEta {
		t.Error("Expected ETA to be disabled")
	}
}

func TestProgressRenderer_GetProgress(t *testing.T) {
	buffer := &bytes.Buffer{}
	config := &Config{
		Output: buffer,
		Colors: true,
	}

	renderer := NewProgressRenderer(config)
	renderer.total = 100
	renderer.current = 25

	current, total, percentage := renderer.GetProgress()

	if current != 25 {
		t.Errorf("Expected current 25, got %d", current)
	}

	if total != 100 {
		t.Errorf("Expected total 100, got %d", total)
	}

	if percentage != 25.0 {
		t.Errorf("Expected percentage 25.0, got %.1f", percentage)
	}
}

func TestProgressRenderer_UpdateWithoutStart(t *testing.T) {
	buffer := &bytes.Buffer{}
	config := &Config{
		Output: buffer,
		Colors: true,
	}

	renderer := NewProgressRenderer(config)

	// Try to update without starting
	err := renderer.UpdateProgress(5, "Test")
	if err != nil {
		t.Errorf("UpdateProgress should not fail when inactive: %v", err)
	}

	// Should not update internal state when inactive
	if renderer.current != 0 {
		t.Errorf("Expected current to remain 0, got %d", renderer.current)
	}
}

func TestProgressRenderer_DoubleFinish(t *testing.T) {
	buffer := &bytes.Buffer{}
	config := &Config{
		Output: buffer,
		Colors: true,
	}

	renderer := NewProgressRenderer(config)
	ctx := context.Background()

	// Start progress
	err := renderer.StartProgress(ctx, 10)
	if err != nil {
		t.Fatalf("StartProgress failed: %v", err)
	}

	// First finish
	err = renderer.FinishProgress()
	if err != nil {
		t.Fatalf("First FinishProgress failed: %v", err)
	}

	// Second finish should not cause issues
	err = renderer.FinishProgress()
	if err != nil {
		t.Errorf("Second FinishProgress should not fail: %v", err)
	}
}
