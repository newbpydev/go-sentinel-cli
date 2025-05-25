package renderer

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/internal/test/cache"
	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// TestNewIncrementalRenderer_Creation verifies incremental renderer initialization
func TestNewIncrementalRenderer_Creation(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	cacheImpl := cache.NewTestResultCache()

	// Act
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cacheImpl)

	// Assert
	if renderer == nil {
		t.Fatal("Expected renderer to be created, got nil")
	}
	if renderer.writer != &buf {
		t.Error("Expected writer to be set correctly")
	}
	if renderer.formatter != formatter {
		t.Error("Expected formatter to be set correctly")
	}
	if renderer.icons != icons {
		t.Error("Expected icons to be set correctly")
	}
	if renderer.width != 80 {
		t.Errorf("Expected width 80, got %d", renderer.width)
	}
	if renderer.cache != cacheImpl {
		t.Error("Expected cache to be set correctly")
	}
	if renderer.lastResults == nil {
		t.Error("Expected lastResults to be initialized")
	}
	if len(renderer.lastResults) != 0 {
		t.Error("Expected lastResults to be empty initially")
	}
}

// TestRenderIncrementalResults_NoChanges tests rendering when no changes are detected
func TestRenderIncrementalResults_NoChanges(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	cacheImpl := cache.NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cacheImpl)

	var changes []*cache.FileChange
	suites := make(map[string]*models.TestSuite)
	stats := &models.TestRunStats{}

	// Act
	err := renderer.RenderIncrementalResults(suites, stats, changes)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No test changes detected") {
		t.Error("Expected 'No test changes detected' message")
	}
}

// TestRenderIncrementalResults_WithChanges tests rendering with file changes
func TestRenderIncrementalResults_WithChanges(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	cacheImpl := cache.NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cacheImpl)

	changes := []*cache.FileChange{
		{
			Path: "pkg/test.go",
			Type: cache.ChangeTypeTest,
		},
		{
			Path: "pkg/source.go",
			Type: cache.ChangeTypeSource,
		},
	}

	suites := map[string]*models.TestSuite{
		"pkg": {
			FilePath:    "pkg",
			TestCount:   2,
			PassedCount: 2,
			Tests: []*models.LegacyTestResult{
				{Name: "TestExample1", Status: models.TestStatusPassed, Duration: 10 * time.Millisecond},
				{Name: "TestExample2", Status: models.TestStatusPassed, Duration: 15 * time.Millisecond},
			},
		},
	}

	stats := &models.TestRunStats{
		TotalTests:  2,
		PassedTests: 2,
		Duration:    25 * time.Millisecond,
	}

	// Act
	err := renderer.RenderIncrementalResults(suites, stats, changes)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "File changes detected") {
		t.Error("Expected 'File changes detected' message")
	}
	if !strings.Contains(output, "pkg/test.go") {
		t.Error("Expected test file path in output")
	}
	if !strings.Contains(output, "pkg/source.go") {
		t.Error("Expected source file path in output")
	}
}

// TestIdentifyChangedSuites_NewSuite tests identifying new test suites
func TestIdentifyChangedSuites_NewSuite(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	cacheImpl := cache.NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cacheImpl)

	currentSuites := map[string]*models.TestSuite{
		"pkg1": {FilePath: "pkg1", TestCount: 1},
		"pkg2": {FilePath: "pkg2", TestCount: 2},
	}

	// Act
	changed := renderer.identifyChangedSuites(currentSuites)

	// Assert
	if len(changed) != 2 {
		t.Errorf("Expected 2 changed suites, got %d", len(changed))
	}

	// Check that both suites are identified as changed (since they're new)
	changedMap := make(map[string]bool)
	for _, suite := range changed {
		changedMap[suite] = true
	}

	if !changedMap["pkg1"] {
		t.Error("Expected pkg1 to be identified as changed")
	}
	if !changedMap["pkg2"] {
		t.Error("Expected pkg2 to be identified as changed")
	}
}

// TestIdentifyChangedSuites_NoChanges tests when no suites have changed
func TestIdentifyChangedSuites_NoChanges(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	cacheImpl := cache.NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cacheImpl)

	// Set up previous results
	suite := &models.TestSuite{
		FilePath:    "pkg1",
		TestCount:   1,
		PassedCount: 1,
		Tests: []*models.LegacyTestResult{
			{Name: "TestExample", Status: models.TestStatusPassed},
		},
	}

	renderer.lastResults["pkg1"] = suite

	// Current suites are identical
	currentSuites := map[string]*models.TestSuite{
		"pkg1": suite,
	}

	// Act
	changed := renderer.identifyChangedSuites(currentSuites)

	// Assert
	if len(changed) != 0 {
		t.Errorf("Expected 0 changed suites, got %d", len(changed))
	}
}

// TestSuiteHasChanged_CountChanges tests suite change detection based on counts
func TestSuiteHasChanged_CountChanges(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	cacheImpl := cache.NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cacheImpl)

	lastSuite := &models.TestSuite{
		TestCount:   2,
		PassedCount: 2,
		FailedCount: 0,
	}

	currentSuite := &models.TestSuite{
		TestCount:   2,
		PassedCount: 1,
		FailedCount: 1,
	}

	// Act
	hasChanged := renderer.suiteHasChanged(lastSuite, currentSuite)

	// Assert
	if !hasChanged {
		t.Error("Expected suite to be detected as changed due to count differences")
	}
}

// TestSuiteHasChanged_StatusChanges tests suite change detection based on test status
func TestSuiteHasChanged_StatusChanges(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	cacheImpl := cache.NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cacheImpl)

	lastSuite := &models.TestSuite{
		TestCount:   1,
		PassedCount: 1,
		Tests: []*models.LegacyTestResult{
			{Name: "TestExample", Status: models.TestStatusPassed},
		},
	}

	currentSuite := &models.TestSuite{
		TestCount:   1,
		FailedCount: 1,
		Tests: []*models.LegacyTestResult{
			{Name: "TestExample", Status: models.TestStatusFailed},
		},
	}

	// Act
	hasChanged := renderer.suiteHasChanged(lastSuite, currentSuite)

	// Assert
	if !hasChanged {
		t.Error("Expected suite to be detected as changed due to status change")
	}
}

// TestSuiteHasChanged_NilSuites tests suite change detection with nil suites
func TestSuiteHasChanged_NilSuites(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	cacheImpl := cache.NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cacheImpl)

	suite := &models.TestSuite{TestCount: 1}

	testCases := []struct {
		name         string
		lastSuite    *models.TestSuite
		currentSuite *models.TestSuite
		expectChange bool
	}{
		{
			name:         "Both nil",
			lastSuite:    nil,
			currentSuite: nil,
			expectChange: true,
		},
		{
			name:         "Last nil, current exists",
			lastSuite:    nil,
			currentSuite: suite,
			expectChange: true,
		},
		{
			name:         "Last exists, current nil",
			lastSuite:    suite,
			currentSuite: nil,
			expectChange: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			hasChanged := renderer.suiteHasChanged(tc.lastSuite, tc.currentSuite)

			// Assert
			if hasChanged != tc.expectChange {
				t.Errorf("Expected hasChanged to be %v, got %v", tc.expectChange, hasChanged)
			}
		})
	}
}

// TestRenderNewSuite_ValidSuite tests rendering a new test suite
func TestRenderNewSuite_ValidSuite(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	cacheImpl := cache.NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cacheImpl)

	suite := &models.TestSuite{
		FilePath: "pkg/test",
		Tests: []*models.LegacyTestResult{
			{Name: "TestExample1", Status: models.TestStatusPassed, Duration: 10 * time.Millisecond},
			{Name: "TestExample2", Status: models.TestStatusFailed, Duration: 20 * time.Millisecond},
		},
	}

	// Act
	err := renderer.renderNewSuite("pkg/test", suite)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "pkg/test") {
		t.Error("Expected suite path in output")
	}
	if !strings.Contains(output, "TestExample1") {
		t.Error("Expected first test name in output")
	}
	if !strings.Contains(output, "TestExample2") {
		t.Error("Expected second test name in output")
	}
}

// TestRenderIncrementalResults_Integration tests the complete incremental rendering workflow
func TestRenderIncrementalResults_Integration(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := colors.NewColorFormatter(false)
	icons := colors.NewIconProvider(false)
	cacheImpl := cache.NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cacheImpl)

	changes := []*cache.FileChange{
		{
			Path: "pkg/test.go",
			Type: cache.ChangeTypeTest,
		},
	}

	suites := map[string]*models.TestSuite{
		"pkg": {
			FilePath:    "pkg",
			TestCount:   1,
			PassedCount: 1,
			Tests: []*models.LegacyTestResult{
				{Name: "TestExample", Status: models.TestStatusPassed, Duration: 10 * time.Millisecond},
			},
		},
	}

	stats := &models.TestRunStats{
		TotalTests:  1,
		PassedTests: 1,
		Duration:    10 * time.Millisecond,
	}

	// Act
	err := renderer.RenderIncrementalResults(suites, stats, changes)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("Expected some output to be generated")
	}
}
