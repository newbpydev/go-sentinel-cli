package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

// TestNewIncrementalRenderer_Creation verifies incremental renderer initialization
func TestNewIncrementalRenderer_Creation(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	cache := NewTestResultCache()

	// Act
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cache)

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
	if renderer.cache != cache {
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
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	cache := NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cache)

	var changes []*FileChange
	suites := make(map[string]*TestSuite)
	stats := &TestRunStats{}

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
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	cache := NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cache)

	changes := []*FileChange{
		{
			Path: "pkg/test.go",
			Type: ChangeTypeTest,
		},
		{
			Path: "pkg/source.go",
			Type: ChangeTypeSource,
		},
	}

	suites := map[string]*TestSuite{
		"pkg": {
			FilePath:    "pkg",
			TestCount:   2,
			PassedCount: 2,
			Tests: []*TestResult{
				{Name: "TestExample1", Status: StatusPassed, Duration: 10 * time.Millisecond},
				{Name: "TestExample2", Status: StatusPassed, Duration: 15 * time.Millisecond},
			},
		},
	}

	stats := &TestRunStats{
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
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	cache := NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cache)

	currentSuites := map[string]*TestSuite{
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
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	cache := NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cache)

	// Set up previous results
	suite := &TestSuite{
		FilePath:    "pkg1",
		TestCount:   1,
		PassedCount: 1,
		Tests: []*TestResult{
			{Name: "TestExample", Status: StatusPassed},
		},
	}

	renderer.lastResults["pkg1"] = suite

	// Current suites are identical
	currentSuites := map[string]*TestSuite{
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
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	cache := NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cache)

	lastSuite := &TestSuite{
		TestCount:   2,
		PassedCount: 2,
		FailedCount: 0,
	}

	currentSuite := &TestSuite{
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
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	cache := NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cache)

	lastSuite := &TestSuite{
		TestCount:   1,
		PassedCount: 1,
		Tests: []*TestResult{
			{Name: "TestExample", Status: StatusPassed},
		},
	}

	currentSuite := &TestSuite{
		TestCount:   1,
		FailedCount: 1,
		Tests: []*TestResult{
			{Name: "TestExample", Status: StatusFailed},
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
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	cache := NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cache)

	suite := &TestSuite{TestCount: 1}

	testCases := []struct {
		name         string
		lastSuite    *TestSuite
		currentSuite *TestSuite
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
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	cache := NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cache)

	suite := &TestSuite{
		FilePath: "pkg/test",
		Tests: []*TestResult{
			{Name: "TestExample1", Status: StatusPassed, Duration: 10 * time.Millisecond},
			{Name: "TestExample2", Status: StatusFailed, Duration: 20 * time.Millisecond},
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

// TestGetChangeIcon_AllTypes tests change icon retrieval for all change types
func TestGetChangeIcon_AllTypes(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	cache := NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cache)

	testCases := []struct {
		changeType ChangeType
		name       string
	}{
		{ChangeTypeTest, "test"},
		{ChangeTypeSource, "source"},
		{ChangeTypeConfig, "config"},
		{ChangeTypeDependency, "dependency"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			icon := renderer.getChangeIcon(tc.changeType)

			// Assert
			if icon == "" {
				t.Errorf("Expected non-empty icon for %s change type", tc.name)
			}
		})
	}
}

// TestGetChangeTypeString_AllTypes tests change type string conversion
func TestGetChangeTypeString_AllTypes(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	cache := NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cache)

	testCases := []struct {
		changeType ChangeType
		expected   string
	}{
		{ChangeTypeTest, "test file"},
		{ChangeTypeSource, "source file"},
		{ChangeTypeConfig, "config file"},
		{ChangeTypeDependency, "dependency"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			// Act
			result := renderer.getChangeTypeString(tc.changeType)

			// Assert
			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// TestGetTestStatusIcon_AllStatuses tests test status icon retrieval
func TestGetTestStatusIcon_AllStatuses(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	cache := NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cache)

	testCases := []struct {
		status TestStatus
		name   string
	}{
		{StatusPassed, "passed"},
		{StatusFailed, "failed"},
		{StatusSkipped, "skipped"},
		{StatusRunning, "running"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			icon := renderer.getTestStatusIcon(tc.status)

			// Assert
			if icon == "" {
				t.Errorf("Expected non-empty icon for %s status", tc.name)
			}
		})
	}
}

// TestGetTestStatusColor_AllStatuses tests test status color retrieval
func TestGetTestStatusColor_AllStatuses(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	cache := NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cache)

	testCases := []struct {
		status   TestStatus
		expected string
	}{
		{StatusPassed, "green"},
		{StatusFailed, "red"},
		{StatusSkipped, "yellow"},
		{StatusRunning, "white"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.status), func(t *testing.T) {
			// Act
			color := renderer.getTestStatusColor(tc.status)

			// Assert
			if color != tc.expected {
				t.Errorf("Expected color '%s', got '%s'", tc.expected, color)
			}
		})
	}
}

// TestUpdateLastResults_StoresResults tests updating cached results
func TestUpdateLastResults_StoresResults(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	cache := NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cache)

	suites := map[string]*TestSuite{
		"pkg1": {FilePath: "pkg1", TestCount: 1},
		"pkg2": {FilePath: "pkg2", TestCount: 2},
	}

	stats := &TestRunStats{
		TotalTests:  3,
		PassedTests: 3,
	}

	// Act
	renderer.updateLastResults(suites, stats)

	// Assert
	if len(renderer.lastResults) != 2 {
		t.Errorf("Expected 2 cached suites, got %d", len(renderer.lastResults))
	}

	if renderer.lastResults["pkg1"] == nil {
		t.Error("Expected pkg1 to be cached")
	}
	if renderer.lastResults["pkg2"] == nil {
		t.Error("Expected pkg2 to be cached")
	}

	if renderer.lastStats != stats {
		t.Error("Expected stats to be cached")
	}
}

// TestRenderChangesSummary_EmptyChanges tests rendering with no changes
func TestRenderChangesSummary_EmptyChanges(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	cache := NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cache)

	var changes []*FileChange

	// Act
	err := renderer.renderChangesSummary(changes)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if output != "" {
		t.Error("Expected no output for empty changes")
	}
}

// TestRenderChangesSummary_WithChanges tests rendering changes summary
func TestRenderChangesSummary_WithChanges(t *testing.T) {
	// Arrange
	var buf bytes.Buffer
	formatter := NewColorFormatter(false)
	icons := NewIconProvider(false)
	cache := NewTestResultCache()
	renderer := NewIncrementalRenderer(&buf, formatter, icons, 80, cache)

	changes := []*FileChange{
		{Path: "test.go", Type: ChangeTypeTest},
		{Path: "source.go", Type: ChangeTypeSource},
	}

	// Act
	err := renderer.renderChangesSummary(changes)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "File changes detected") {
		t.Error("Expected 'File changes detected' header")
	}
	if !strings.Contains(output, "test.go") {
		t.Error("Expected test.go in output")
	}
	if !strings.Contains(output, "source.go") {
		t.Error("Expected source.go in output")
	}
}
