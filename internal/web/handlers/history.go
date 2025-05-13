package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
	
	"github.com/go-chi/chi/v5"
)

// TestRunHistory represents a collection of test runs
type TestRunHistory struct {
	// Use RWMutex for concurrent access
	mu       sync.RWMutex
	TestRuns []TestRun `json:"testRuns"`
}

// TestRun represents a single execution of tests
type TestRun struct {
	ID           string       `json:"id"`
	Timestamp    time.Time    `json:"timestamp"`
	TotalTests   int          `json:"totalTests"`
	PassedTests  int          `json:"passedTests"`
	FailedTests  int          `json:"failedTests"`
	TotalTime    string       `json:"totalTime"` // Duration as string for display
	TestResults  []TestResult `json:"testResults"`
	Branch       string       `json:"branch,omitempty"`
	Commit       string       `json:"commit,omitempty"`
	TriggeredBy  string       `json:"triggeredBy,omitempty"`
	BuildVersion string       `json:"buildVersion,omitempty"`
}

// NewTestRunHistory creates a new test history manager
func NewTestRunHistory() *TestRunHistory {
	return &TestRunHistory{
		TestRuns: []TestRun{},
	}
}

// AddTestRun adds a new test run to the history
func (h *TestRunHistory) AddTestRun(run TestRun) {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// Generate ID if not provided
	if run.ID == "" {
		run.ID = fmt.Sprintf("run-%d", time.Now().UnixNano())
	}
	
	// Set timestamp if not provided
	if run.Timestamp.IsZero() {
		run.Timestamp = time.Now()
	}
	
	// Add run to history
	h.TestRuns = append(h.TestRuns, run)
	
	// Sort by timestamp (most recent first)
	sort.Slice(h.TestRuns, func(i, j int) bool {
		return h.TestRuns[i].Timestamp.After(h.TestRuns[j].Timestamp)
	})
	
	// Limit history size (optional)
	const maxHistorySize = 100
	if len(h.TestRuns) > maxHistorySize {
		h.TestRuns = h.TestRuns[:maxHistorySize]
	}
}

// GetTestRuns returns all test runs
func (h *TestRunHistory) GetTestRuns() []TestRun {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	// Return a copy to avoid concurrent access issues
	runs := make([]TestRun, len(h.TestRuns))
	copy(runs, h.TestRuns)
	return runs
}

// GetTestRunByID returns a specific test run by ID
func (h *TestRunHistory) GetTestRunByID(id string) (TestRun, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	for _, run := range h.TestRuns {
		if run.ID == id {
			return run, true
		}
	}
	
	return TestRun{}, false
}

// GetRecentTestRuns returns the most recent test runs
func (h *TestRunHistory) GetRecentTestRuns(limit int) []TestRun {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	if limit <= 0 || limit > len(h.TestRuns) {
		limit = len(h.TestRuns)
	}
	
	runs := make([]TestRun, limit)
	copy(runs, h.TestRuns[:limit])
	return runs
}

// HistoryHandler manages test history endpoints
type HistoryHandler struct {
	history *TestRunHistory
}

// NewHistoryHandler creates a new history handler
func NewHistoryHandler() *HistoryHandler {
	// Create history with some sample data for demo
	history := NewTestRunHistory()
	
	// Add mock test runs for demo
	mockTestRuns := createMockTestRuns()
	for _, run := range mockTestRuns {
		history.AddTestRun(run)
	}
	
	return &HistoryHandler{
		history: history,
	}
}

// GetTestRunHistory returns the test run history
func (h *HistoryHandler) GetTestRunHistory(w http.ResponseWriter, r *http.Request) {
	// Get limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // Default limit
	if limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}
	
	// Get test runs
	runs := h.history.GetRecentTestRuns(limit)
	
	// For HTMX requests, render HTML
	if r.Header.Get("HX-Request") == "true" {
		h.renderHistoryHTML(w, runs)
		return
	}
	
	// For API requests, return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"testRuns": runs,
	})
}

// GetTestRunDetails returns details for a specific test run
func (h *HistoryHandler) GetTestRunDetails(w http.ResponseWriter, r *http.Request) {
	// Get run ID from URL
	runID := chi.URLParam(r, "runID")
	if runID == "" {
		http.Error(w, "Run ID is required", http.StatusBadRequest)
		return
	}
	
	// Get test run
	run, found := h.history.GetTestRunByID(runID)
	if !found {
		http.Error(w, "Test run not found", http.StatusNotFound)
		return
	}
	
	// For HTMX requests, render HTML
	if r.Header.Get("HX-Request") == "true" {
		h.renderTestRunDetailsHTML(w, run)
		return
	}
	
	// For API requests, return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(run)
}

// CompareTestRuns compares two test runs
func (h *HistoryHandler) CompareTestRuns(w http.ResponseWriter, r *http.Request) {
	// Get run IDs from URL
	baseRunID := r.URL.Query().Get("baseRunID")
	compareRunID := r.URL.Query().Get("compareRunID")
	
	if baseRunID == "" || compareRunID == "" {
		http.Error(w, "Both base and compare run IDs are required", http.StatusBadRequest)
		return
	}
	
	// Get test runs
	baseRun, foundBase := h.history.GetTestRunByID(baseRunID)
	compareRun, foundCompare := h.history.GetTestRunByID(compareRunID)
	
	if !foundBase || !foundCompare {
		http.Error(w, "One or both test runs not found", http.StatusNotFound)
		return
	}
	
	// For HTMX requests, render HTML
	if r.Header.Get("HX-Request") == "true" {
		h.renderComparisonHTML(w, baseRun, compareRun)
		return
	}
	
	// For API requests, return JSON
	comparison := generateComparison(baseRun, compareRun)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comparison)
}

// renderHistoryHTML renders HTML for test history
func (h *HistoryHandler) renderHistoryHTML(w http.ResponseWriter, runs []TestRun) {
	w.Header().Set("Content-Type", "text/html")
	
	// Start with empty content to replace the loading spinner
	html := ""
	
	// Generate history items using the new card-based styling
	for i, run := range runs {
		// Format timestamp
		timestamp := run.Timestamp.Format("Jan 02, 15:04")
		runNumber := i + 1
		
		// Calculate success rate
		successRate := 0
		if run.TotalTests > 0 {
			successRate = (run.PassedTests * 100) / run.TotalTests
		}
		
		// Generate compare button if not the most recent run
		compareButton := ""
		if i > 0 && len(runs) > 0 {
			compareButton = fmt.Sprintf(`<button class="btn btn-warning" hx-get="/api/history/compare?baseRunID=%s&compareRunID=%s" hx-target="#comparison-container">Compare</button>`, runs[0].ID, run.ID)
		}
		
		// Create a history item card with the new styling
		html += fmt.Sprintf(`
		<div class="history-item">
			<div class="history-item-header">
				<div class="history-item-title">Run #%d - %s</div>
				<span class="status-badge passed">%d%%</span>
			</div>
			<div class="history-item-stats">
				<div>Total: %d</div>
				<div>Passed: %d</div>
				<div>Failed: %d</div>
				<div>Time: %s</div>
			</div>
			<div class="history-item-actions">
				<button class="btn btn-info" hx-get="/api/history/%s" hx-target="#run-details">Details</button>
				%s
			</div>
		</div>
		`, runNumber, timestamp, successRate, run.TotalTests, run.PassedTests, run.FailedTests, run.TotalTime, run.ID, compareButton)
	}
	
	// If no runs, show a message
	if len(runs) == 0 {
		html = `<div class="content-padding"><p class="text-center">No test runs available yet.</p></div>`
	}
	
	w.Write([]byte(html))
}

// renderTestRunDetailsHTML renders HTML for a single test run
func (h *HistoryHandler) renderTestRunDetailsHTML(w http.ResponseWriter, run TestRun) {
	w.Header().Set("Content-Type", "text/html")
	
	// Calculate success rate
	successRate := 0
	if run.TotalTests > 0 {
		successRate = (run.PassedTests * 100) / run.TotalTests
	}
	
	// Format timestamp
	timestamp := run.Timestamp.Format("Jan 02, 2006 15:04:05")
	
	// Generate build info
	buildInfo := renderBuildInfo(run)
	
	// Start building the HTML
	html := fmt.Sprintf(`
	<div class="card-content">
		<div class="history-item-header">
			<div class="history-item-title">Test Run Details: %s</div>
			<span class="status-badge passed">%d%%</span>
		</div>
		
		<!-- Stats Cards for Quick Metrics -->
		<div class="stats-cards content-padding-y">
			<!-- Total Tests Card -->
			<div class="stat-card" role="region" aria-label="Total Tests">
				<div class="stat-card-content">
					<div class="stat-title">Total Tests</div>
					<div class="stat-value">%d</div>
				</div>
			</div>
			
			<!-- Passing Tests Card -->
			<div class="stat-card success" role="region" aria-label="Passing Tests">
				<div class="stat-card-content">
					<div class="stat-title">Passing</div>
					<div class="stat-value">%d</div>
					<div class="stat-change">%d%% success rate</div>
				</div>
			</div>
			
			<!-- Failing Tests Card -->
			<div class="stat-card error" role="region" aria-label="Failing Tests">
				<div class="stat-card-content">
					<div class="stat-title">Failing</div>
					<div class="stat-value">%d</div>
				</div>
			</div>
			
			<!-- Average Duration Card -->
			<div class="stat-card" role="region" aria-label="Test Duration">
				<div class="stat-card-content">
					<div class="stat-title">Duration</div>
					<div class="stat-value">%s</div>
				</div>
			</div>
		</div>
		
		<!-- Build Information Section -->
		<div class="content-padding">
			<div class="detail-grid">
				<div class="detail-item">
					<span class="detail-label">Run Date:</span>
					<span class="detail-value">%s</span>
				</div>
				<div class="detail-item">
					<span class="detail-label">ID:</span>
					<span class="detail-value">%s</span>
				</div>
			</div>
			%s
		</div>
		
		<!-- Test Results Table Section -->
		<div class="content-padding">
			<div class="section-header">
				<h4 class="section-title">Test Results</h4>
				<div class="filter-controls">
					<button class="btn btn-info active" data-filter="all">All (%d)</button>
					<button class="btn btn-success" data-filter="passed">Passed (%d)</button>
					<button class="btn btn-error" data-filter="failed">Failed (%d)</button>
				</div>
			</div>
			
			<div class="test-table-container content-padding-y">
				<table aria-label="Test Results" class="test-table">
					<thead>
						<tr>
							<th scope="col">Test Name</th>
							<th scope="col">Status</th>
							<th scope="col">Duration</th>
							<th scope="col">Actions</th>
						</tr>
					</thead>
					<tbody>`,
		getShortID(run.ID), successRate, 
		run.TotalTests, run.PassedTests, successRate, run.FailedTests, run.TotalTime,
		timestamp, run.ID, buildInfo,
		run.TotalTests, run.PassedTests, run.FailedTests)
	
	// Generate test result rows
	for _, test := range run.TestResults {
		// Determine status class
		statusClass := "error"
		statusText := "Failed"
		if test.Status == "passed" {
			statusClass = "success"
			statusText = "Passed"
		}
		
		// Determine if output should be shown
		outputHTML := ""
		if test.Status == "failed" && test.Output != "" {
			outputHTML = fmt.Sprintf(`<tr class="test-output-row">
				<td colspan="4" class="test-output">
					<pre>%s</pre>
				</td>
			</tr>`, test.Output)
		}
		
		html += fmt.Sprintf(`
						<tr class="test-row %s" data-test-name="%s" data-test-status="%s">
							<td class="test-name">%s</td>
							<td class="test-status">
								<span class="status-badge %s">%s</span>
							</td>
							<td class="test-duration">%s</td>
							<td class="test-actions">
								<button class="btn btn-sm btn-info">View</button>
							</td>
						</tr>
						%s`,
			test.Status, test.Name, test.Status, test.Name, statusClass, statusText, test.Duration, outputHTML)
	}
	
	// Close all HTML tags properly
	html += `
					</tbody>
				</table>
			</div>
		</div>
	</div>`
	
	w.Write([]byte(html))
}

// renderComparisonHTML renders HTML comparing two test runs
func (h *HistoryHandler) renderComparisonHTML(w http.ResponseWriter, baseRun, compareRun TestRun) {
	w.Header().Set("Content-Type", "text/html")
	
	comparison := generateComparison(baseRun, compareRun)
	
	// Prepare formatted values
	totalTestsClass := "neutral"
	if comparison.TotalTestsDiff > 0 {
		totalTestsClass = "positive"
	} else if comparison.TotalTestsDiff < 0 {
		totalTestsClass = "negative"
	}
	
	totalTestsDiffStr := fmt.Sprint(comparison.TotalTestsDiff)
	if comparison.TotalTestsDiff > 0 {
		totalTestsDiffStr = "+" + totalTestsDiffStr
	}
	
	passedTestsClass := "neutral"
	if comparison.PassedTestsDiff > 0 {
		passedTestsClass = "positive"
	} else if comparison.PassedTestsDiff < 0 {
		passedTestsClass = "negative"
	}
	
	passedTestsDiffStr := fmt.Sprint(comparison.PassedTestsDiff)
	if comparison.PassedTestsDiff > 0 {
		passedTestsDiffStr = "+" + passedTestsDiffStr
	}
	
	failedTestsClass := "neutral"
	if comparison.FailedTestsDiff > 0 {
		failedTestsClass = "negative"
	} else if comparison.FailedTestsDiff < 0 {
		failedTestsClass = "positive"
	}
	
	failedTestsDiffStr := fmt.Sprint(comparison.FailedTestsDiff)
	if comparison.FailedTestsDiff > 0 {
		failedTestsDiffStr = "+" + failedTestsDiffStr
	}
	
	successRateClass := "neutral"
	if comparison.SuccessRateDiff > 0 {
		successRateClass = "positive"
	} else if comparison.SuccessRateDiff < 0 {
		successRateClass = "negative"
	}
	
	successRateDiffStr := fmt.Sprintf("%.1f%%", comparison.SuccessRateDiff)
	if comparison.SuccessRateDiff > 0 {
		successRateDiffStr = "+" + successRateDiffStr
	}
	
	html := fmt.Sprintf(`
	<div class="card-content">
		<div class="history-item-header">
			<div class="history-item-title">Test Run Comparison</div>
		</div>
		
		<div class="content-padding">
			<div class="comparison-runs">
				<div class="comparison-run base">
					<span class="comparison-label">Base:</span>
					<span class="comparison-value">%s (%s)</span>
				</div>
				<div class="comparison-run compare">
					<span class="comparison-label">Compare:</span>
					<span class="comparison-value">%s (%s)</span>
				</div>
			</div>
		</div>
		
		<!-- Stats Cards for Metrics Comparison -->
		<div class="stats-cards content-padding-y">
			<!-- Total Tests Card -->
			<div class="stat-card" role="region" aria-label="Total Tests">
				<div class="stat-card-content">
					<div class="stat-title">Total Tests</div>
					<div class="stat-value">%d → %d</div>
					<div class="stat-change %s">%s</div>
				</div>
			</div>
			
			<!-- Passing Tests Card -->
			<div class="stat-card success" role="region" aria-label="Passing Tests">
				<div class="stat-card-content">
					<div class="stat-title">Passing</div>
					<div class="stat-value">%d → %d</div>
					<div class="stat-change %s">%s</div>
				</div>
			</div>
			
			<!-- Failing Tests Card -->
			<div class="stat-card error" role="region" aria-label="Failing Tests">
				<div class="stat-card-content">
					<div class="stat-title">Failing</div>
					<div class="stat-value">%d → %d</div>
					<div class="stat-change %s">%s</div>
				</div>
			</div>
			
			<!-- Success Rate Card -->
			<div class="stat-card" role="region" aria-label="Success Rate">
				<div class="stat-card-content">
					<div class="stat-title">Success Rate</div>
					<div class="stat-value">%d%% → %d%%</div>
					<div class="stat-change %s">%s</div>
				</div>
			</div>
		</div>
		
		<!-- Test Changes Table -->
		<div class="test-table-container">
			<table class="test-table w-full">
				<thead>
					<tr>
						<th>Test Name</th>
						<th>Status</th>
					</tr>
				</thead>
				<tbody>`,
		getShortID(baseRun.ID), baseRun.Timestamp.Format("Jan 02, 15:04"),
		getShortID(compareRun.ID), compareRun.Timestamp.Format("Jan 02, 15:04"),
		baseRun.TotalTests, compareRun.TotalTests, totalTestsClass, totalTestsDiffStr,
		baseRun.PassedTests, compareRun.PassedTests, passedTestsClass, passedTestsDiffStr,
		baseRun.FailedTests, compareRun.FailedTests, failedTestsClass, failedTestsDiffStr,
		comparison.BaseSuccessRate, comparison.CompareSuccessRate, successRateClass, successRateDiffStr)

	// Add fixed tests
	for _, test := range comparison.Fixed {
		html += fmt.Sprintf(`
					<tr class="test-row success">
						<td>%s</td>
						<td><span class="badge success">Fixed</span></td>
					</tr>`, test)
	}

	// Add newly failed tests
	for _, test := range comparison.NewlyFailed {
		html += fmt.Sprintf(`
					<tr class="test-row error">
						<td>%s</td>
						<td><span class="badge error">Newly Failed</span></td>
					</tr>`, test)
	}

	// Add new tests
	for _, test := range comparison.New {
		html += fmt.Sprintf(`
					<tr class="test-row info">
						<td>%s</td>
						<td><span class="badge info">New</span></td>
					</tr>`, test)
	}

	// Add removed tests
	for _, test := range comparison.Removed {
		html += fmt.Sprintf(`
					<tr class="test-row warning">
						<td>%s</td>
						<td><span class="badge warning">Removed</span></td>
					</tr>`, test)
	}

	html += `
				</tbody>
			</table>
		</div>
	</div>
</div>`
	
	w.Write([]byte(html))
}

// renderBuildInfo renders HTML for build information
func renderBuildInfo(run TestRun) string {
	// Skip if no build info is available
	if run.Branch == "" && run.Commit == "" && run.TriggeredBy == "" && run.BuildVersion == "" {
		return ""
	}
	
	html := `<div class="build-info card-content-sm">
		<h5 class="text-md font-medium mb-2">Build Information</h5>`
	
	if run.Branch != "" {
		html += fmt.Sprintf(`
			<div class="detail-item">
				<span class="detail-label">Branch:</span>
				<span class="detail-value">%s</span>
			</div>`, run.Branch)
	}
	
	if run.Commit != "" {
		html += fmt.Sprintf(`
			<div class="detail-item">
				<span class="detail-label">Commit:</span>
				<span class="detail-value">%s</span>
			</div>`, run.Commit)
	}
	
	if run.TriggeredBy != "" {
		html += fmt.Sprintf(`
			<div class="detail-item">
				<span class="detail-label">Triggered By:</span>
				<span class="detail-value">%s</span>
			</div>`, run.TriggeredBy)
	}
	
	if run.BuildVersion != "" {
		html += fmt.Sprintf(`
			<div class="detail-item">
				<span class="detail-label">Version:</span>
				<span class="detail-value">%s</span>
			</div>`, run.BuildVersion)
	}
	
	html += `</div>`
	
	return html
}

// TestRunComparison holds differences between two test runs
type TestRunComparison struct {
	BaseRunID         string   `json:"baseRunId"`
	CompareRunID      string   `json:"compareRunId"`
	TotalTestsDiff    int      `json:"totalTestsDiff"`    // Positive = more tests in compare
	PassedTestsDiff   int      `json:"passedTestsDiff"`   // Positive = more passes in compare
	FailedTestsDiff   int      `json:"failedTestsDiff"`   // Positive = more failures in compare
	BaseSuccessRate   int      `json:"baseSuccessRate"`   // As percentage
	CompareSuccessRate int     `json:"compareSuccessRate"` // As percentage
	SuccessRateDiff   float64  `json:"successRateDiff"`   // Positive = better success rate in compare
	Fixed             []string `json:"fixed"`             // Tests that were fixed (failed in base, passed in compare)
	NewlyFailed       []string `json:"newlyFailed"`       // Tests that newly failed (passed in base, failed in compare)
	New               []string `json:"new"`               // Tests that are new in compare
	Removed           []string `json:"removed"`           // Tests that were removed (in base but not in compare)
}

// generateComparison compares two test runs
func generateComparison(baseRun, compareRun TestRun) TestRunComparison {
	comparison := TestRunComparison{
		BaseRunID:    baseRun.ID,
		CompareRunID: compareRun.ID,
	}
	
	// Calculate basic metrics
	comparison.TotalTestsDiff = compareRun.TotalTests - baseRun.TotalTests
	comparison.PassedTestsDiff = compareRun.PassedTests - baseRun.PassedTests
	comparison.FailedTestsDiff = compareRun.FailedTests - baseRun.FailedTests
	
	// Calculate success rates
	comparison.BaseSuccessRate = 0
	if baseRun.TotalTests > 0 {
		comparison.BaseSuccessRate = (baseRun.PassedTests * 100) / baseRun.TotalTests
	}
	
	comparison.CompareSuccessRate = 0
	if compareRun.TotalTests > 0 {
		comparison.CompareSuccessRate = (compareRun.PassedTests * 100) / compareRun.TotalTests
	}
	
	comparison.SuccessRateDiff = float64(comparison.CompareSuccessRate - comparison.BaseSuccessRate)
	
	// Build maps of test status
	baseTests := make(map[string]string)
	compareTests := make(map[string]string)
	
	for _, test := range baseRun.TestResults {
		baseTests[test.Name] = test.Status
	}
	
	for _, test := range compareRun.TestResults {
		compareTests[test.Name] = test.Status
	}
	
	// Find fixed, newly failed, new, and removed tests
	for name, status := range baseTests {
		compareStatus, exists := compareTests[name]
		
		if !exists {
			comparison.Removed = append(comparison.Removed, name)
			continue
		}
		
		if status == "failed" && compareStatus == "passed" {
			comparison.Fixed = append(comparison.Fixed, name)
		} else if status == "passed" && compareStatus == "failed" {
			comparison.NewlyFailed = append(comparison.NewlyFailed, name)
		}
	}
	
	for name := range compareTests {
		if _, exists := baseTests[name]; !exists {
			comparison.New = append(comparison.New, name)
		}
	}
	
	// Sort results for consistency
	sort.Strings(comparison.Fixed)
	sort.Strings(comparison.NewlyFailed)
	sort.Strings(comparison.New)
	sort.Strings(comparison.Removed)
	
	return comparison
}

// createMockTestRuns generates mock test runs for demo
func createMockTestRuns() []TestRun {
	now := time.Now()
	
	runs := []TestRun{
		{
			ID:           "run-1",
			Timestamp:    now.Add(-72 * time.Hour),
			TotalTests:   120,
			PassedTests:  110,
			FailedTests:  10,
			TotalTime:    "1.5s",
			Branch:       "main",
			Commit:       "a1b2c3d4",
			TriggeredBy:  "scheduled",
			BuildVersion: "1.0.0",
			TestResults:  createMockTestResults(120, 10),
		},
		{
			ID:           "run-2",
			Timestamp:    now.Add(-48 * time.Hour),
			TotalTests:   122,
			PassedTests:  112,
			FailedTests:  10,
			TotalTime:    "1.6s",
			Branch:       "feature/new-widget",
			Commit:       "e5f6g7h8",
			TriggeredBy:  "pull-request",
			BuildVersion: "1.0.0",
			TestResults:  createMockTestResults(122, 10),
		},
		{
			ID:           "run-3",
			Timestamp:    now.Add(-24 * time.Hour),
			TotalTests:   125,
			PassedTests:  118,
			FailedTests:  7,
			TotalTime:    "1.4s",
			Branch:       "feature/new-widget",
			Commit:       "i9j0k1l2",
			TriggeredBy:  "pull-request",
			BuildVersion: "1.0.0",
			TestResults:  createMockTestResults(125, 7),
		},
		{
			ID:           "run-4",
			Timestamp:    now.Add(-12 * time.Hour),
			TotalTests:   125,
			PassedTests:  120,
			FailedTests:  5,
			TotalTime:    "1.3s",
			Branch:       "main",
			Commit:       "m3n4o5p6",
			TriggeredBy:  "merge",
			BuildVersion: "1.0.1",
			TestResults:  createMockTestResults(125, 5),
		},
		{
			ID:           "run-5",
			Timestamp:    now.Add(-2 * time.Hour),
			TotalTests:   128,
			PassedTests:  119,
			FailedTests:  9,
			TotalTime:    "1.2s",
			Branch:       "fix/widget-bugs",
			Commit:       "q7r8s9t0",
			TriggeredBy:  "manual",
			BuildVersion: "1.0.1",
			TestResults:  createMockTestResults(128, 9),
		},
	}
	
	return runs
}

// getShortID safely truncates an ID to 8 characters or returns the full ID if shorter
func getShortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

// createMockTestResults generates mock test results
func createMockTestResults(total, failed int) []TestResult {
	passed := total - failed
	results := make([]TestResult, 0, total)
	
	// Generate passed tests
	for i := 0; i < passed; i++ {
		results = append(results, TestResult{
			Name:     fmt.Sprintf("TestSuccess%d", i+1),
			Status:   "passed",
			Duration: fmt.Sprintf("0.%ds", (i%5)+1),
		})
	}
	
	// Generate failed tests
	for i := 0; i < failed; i++ {
		results = append(results, TestResult{
			Name:     fmt.Sprintf("TestFailed%d", i+1),
			Status:   "failed",
			Duration: fmt.Sprintf("0.%ds", (i%5)+1),
			Output:   fmt.Sprintf("Expected value to be %d, got %d", i+10, i+5),
		})
	}
	
	return results
}
