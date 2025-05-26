// Package display provides tests for the test execution renderer
package display

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/models"
)

func TestNewTestExecutionRenderer(t *testing.T) {
	config := &Config{
		Output: &bytes.Buffer{},
		Colors: true,
	}

	renderer := NewTestExecutionRenderer(config, nil)

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

	if renderer.spacingManager == nil {
		t.Error("Expected spacing manager to be initialized")
	}

	if renderer.timingFormatter == nil {
		t.Error("Expected timing formatter to be initialized")
	}

	// Test default options
	if !renderer.showTiming {
		t.Error("Expected timing to be shown by default")
	}

	if !renderer.showSubtests {
		t.Error("Expected subtests to be shown by default")
	}

	if renderer.maxNameLength != 80 {
		t.Errorf("Expected max name length 80, got %d", renderer.maxNameLength)
	}

	if renderer.indentLevel != 0 {
		t.Errorf("Expected indent level 0, got %d", renderer.indentLevel)
	}
}

func TestTestExecutionRenderer_RenderTestExecution_PassedTest(t *testing.T) {
	config := &Config{
		Output: &bytes.Buffer{},
		Colors: false, // Disable colors for easier testing
	}

	renderer := NewTestExecutionRenderer(config, nil)

	test := &models.TestResult{
		Name:     "TestExample",
		Status:   models.TestStatusPassed,
		Duration: 150 * time.Millisecond,
	}

	result := renderer.RenderTestExecution(test)

	// Should start with 2-space indentation
	if !strings.HasPrefix(result, "  ") {
		t.Errorf("Expected 2-space indentation, got: %q", result)
	}

	// Should contain test name
	if !strings.Contains(result, "TestExample") {
		t.Errorf("Expected test name in result: %q", result)
	}

	// Should contain timing
	if !strings.Contains(result, "ms") {
		t.Errorf("Expected timing in result: %q", result)
	}

	// Should contain status icon (check mark or equivalent)
	lines := strings.Split(result, "\n")
	firstLine := lines[0]

	// The format should be: "  <icon> TestExample <timing>"
	parts := strings.Fields(firstLine)
	if len(parts) < 3 {
		t.Errorf("Expected at least 3 parts in output, got %d: %q", len(parts), firstLine)
	}
}

func TestTestExecutionRenderer_RenderTestExecution_FailedTest(t *testing.T) {
	config := &Config{
		Output: &bytes.Buffer{},
		Colors: false,
	}

	renderer := NewTestExecutionRenderer(config, nil)

	test := &models.TestResult{
		Name:     "TestFailure",
		Status:   models.TestStatusFailed,
		Duration: 75 * time.Millisecond,
	}

	result := renderer.RenderTestExecution(test)

	// Should start with 2-space indentation
	if !strings.HasPrefix(result, "  ") {
		t.Errorf("Expected 2-space indentation, got: %q", result)
	}

	// Should contain test name
	if !strings.Contains(result, "TestFailure") {
		t.Errorf("Expected test name in result: %q", result)
	}

	// Should contain timing
	if !strings.Contains(result, "ms") {
		t.Errorf("Expected timing in result: %q", result)
	}
}

func TestTestExecutionRenderer_RenderTestExecution_SkippedTest(t *testing.T) {
	config := &Config{
		Output: &bytes.Buffer{},
		Colors: false,
	}

	renderer := NewTestExecutionRenderer(config, nil)

	test := &models.TestResult{
		Name:     "TestSkipped",
		Status:   models.TestStatusSkipped,
		Duration: 0,
	}

	result := renderer.RenderTestExecution(test)

	// Should start with 2-space indentation
	if !strings.HasPrefix(result, "  ") {
		t.Errorf("Expected 2-space indentation, got: %q", result)
	}

	// Should contain test name
	if !strings.Contains(result, "TestSkipped") {
		t.Errorf("Expected test name in result: %q", result)
	}

	// Should show 0ms for skipped test
	if !strings.Contains(result, "0ms") {
		t.Errorf("Expected 0ms timing in result: %q", result)
	}
}

func TestTestExecutionRenderer_DisabledTiming(t *testing.T) {
	config := &Config{
		Output: &bytes.Buffer{},
		Colors: false,
	}

	options := &TestExecutionRenderOptions{
		ShowTiming: false,
	}

	renderer := NewTestExecutionRenderer(config, options)

	test := &models.TestResult{
		Name:     "TestExample",
		Status:   models.TestStatusPassed,
		Duration: 100 * time.Millisecond,
	}

	result := renderer.RenderTestExecution(test)

	// Should not contain timing when disabled
	if strings.Contains(result, "ms") {
		t.Errorf("Should not contain timing when disabled: %q", result)
	}

	// Should still contain test name
	if !strings.Contains(result, "TestExample") {
		t.Errorf("Expected test name in result: %q", result)
	}
}
