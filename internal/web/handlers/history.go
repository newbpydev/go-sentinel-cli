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
	
	html := `<div class="history-list">
		<div class="history-header">
			<h3>Test Run History</h3>
			<div class="history-actions">
				<button class="refresh-button" hx-get="/api/history" hx-target="#history-container">
					<span class="icon">🔄</span> Refresh
				</button>
			</div>
		</div>
		
		<div class="history-timeline">`
	
	// Generate timeline entries
	for i, run := range runs {
		// Format timestamp
		timestamp := run.Timestamp.Format("Jan 02, 15:04")
		
		// Calculate success rate
		successRate := 0
		if run.TotalTests > 0 {
			successRate = (run.PassedTests * 100) / run.TotalTests
		}
		
		// Determine status class
		statusClass := "success"
		if run.FailedTests > 0 {
			statusClass = "warning"
			if successRate < 80 {
				statusClass = "error"
			}
		}
		
		html += fmt.Sprintf(`
			<div class="timeline-entry %s" data-run-id="%s">
				<div class="timeline-marker"></div>
				<div class="timeline-content" hx-get="/api/history/%s" hx-target="#run-details">
					<div class="timeline-header">
						<span class="timeline-date">%s</span>
						<span class="timeline-id">#%s</span>
					</div>
					<div class="timeline-summary">
						<div class="timeline-stats">
							<span class="stat-item">Total: %d</span>
							<span class="stat-item success">Passed: %d</span>
							<span class="stat-item error">Failed: %d</span>
							<span class="stat-item">Time: %s</span>
						</div>
						<div class="timeline-progress">
							<div class="progress-bar">
								<div class="progress-value" style="width: %d%%"></div>
							</div>
							<span class="progress-label">%d%%</span>
						</div>
					</div>
					<div class="timeline-actions">
						<button class="mini-button" aria-label="View details">Details</button>
						%s
					</div>
				</div>
			</div>
		`, statusClass, run.ID, run.ID, timestamp, run.ID[:8],
		   run.TotalTests, run.PassedTests, run.FailedTests, run.TotalTime,
		   successRate, successRate,
		   func() string {
			   if i > 0 {
				   return fmt.Sprintf(`<button class="mini-button compare" hx-get="/api/history/compare?baseRunID=%s&compareRunID=%s" hx-target="#comparison-container">Compare</button>`, runs[0].ID, run.ID)
			   }
			   return ""
		   }())
	}
	
	html += `</div></div>`
	
	w.Write([]byte(html))
}

// renderTestRunDetailsHTML renders HTML for a single test run
func (h *HistoryHandler) renderTestRunDetailsHTML(w http.ResponseWriter, run TestRun) {
	w.Header().Set("Content-Type", "text/html")
	
	// Format timestamp
	timestamp := run.Timestamp.Format("Jan 02, 2006 15:04:05")
	
	html := fmt.Sprintf(`
		<div class="run-details-container">
			<div class="run-details-header">
				<h3>Test Run Details</h3>
				<span class="run-timestamp">%s</span>
			</div>
			
			<div class="run-details-summary">
				<div class="detail-section">
					<div class="detail-item">
						<span class="detail-label">ID:</span>
						<span class="detail-value">%s</span>
					</div>
					<div class="detail-item">
						<span class="detail-label">Total Tests:</span>
						<span class="detail-value">%d</span>
					</div>
					<div class="detail-item">
						<span class="detail-label">Passed:</span>
						<span class="detail-value success">%d</span>
					</div>
					<div class="detail-item">
						<span class="detail-label">Failed:</span>
						<span class="detail-value error">%d</span>
					</div>
					<div class="detail-item">
						<span class="detail-label">Duration:</span>
						<span class="detail-value">%s</span>
					</div>
				</div>
				
				<div class="detail-section">
					%s
				</div>
			</div>
			
			<div class="run-details-tests">
				<h4>Test Results</h4>
				<div class="test-filter-controls">
					<button class="filter-button active" data-filter="all">All (%d)</button>
					<button class="filter-button" data-filter="passed">Passed (%d)</button>
					<button class="filter-button" data-filter="failed">Failed (%d)</button>
				</div>
				
				<div class="test-results-list">`,
		timestamp, run.ID, run.TotalTests, run.PassedTests, run.FailedTests, run.TotalTime,
		renderBuildInfo(run),
		run.TotalTests, run.PassedTests, run.FailedTests)
	
	// Generate test result rows
	for _, test := range run.TestResults {
		// Determine status text
		statusText := "Failed"
		if test.Status == "passed" {
			statusText = "Passed"
		}
		
		// Determine if output should be shown
		outputHTML := ""
		if test.Status == "failed" && test.Output != "" {
			outputHTML = fmt.Sprintf(`<div class="test-output"><pre>%s</pre></div>`, test.Output)
		}
		
		html += fmt.Sprintf(`
			<div class="test-result-item %s">
				<div class="test-result-header">
					<span class="test-name">%s</span>
					<span class="test-badge %s">%s</span>
					<span class="test-duration">%s</span>
				</div>
				%s
			</div>
		`, test.Status, test.Name, test.Status, statusText, test.Duration, outputHTML)
	}
	
	html += `</div></div></div>`
	
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
		<div class="comparison-container">
			<div class="comparison-header">
				<h3>Test Run Comparison</h3>
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
			
			<div class="comparison-summary">
				<div class="metric-comparison">
					<div class="metric-group">
						<div class="metric">
							<div class="metric-name">Total Tests</div>
							<div class="metric-values">
								<div class="value base">%d</div>
								<div class="value change %s">%s</div>
								<div class="value compare">%d</div>
							</div>
						</div>
						
						<div class="metric">
							<div class="metric-name">Passed Tests</div>
							<div class="metric-values">
								<div class="value base">%d</div>
								<div class="value change %s">%s</div>
								<div class="value compare">%d</div>
							</div>
						</div>
						
						<div class="metric">
							<div class="metric-name">Failed Tests</div>
							<div class="metric-values">
								<div class="value base">%d</div>
								<div class="value change %s">%s</div>
								<div class="value compare">%d</div>
							</div>
						</div>
						
						<div class="metric">
							<div class="metric-name">Success Rate</div>
							<div class="metric-values">
								<div class="value base">%d%%</div>
								<div class="value change %s">%s</div>
								<div class="value compare">%d%%</div>
							</div>
						</div>
					</div>
				</div>
			</div>
			
			<div class="comparison-details">
				<h4>Test Status Changes</h4>
				<div class="status-changes">`,
		baseRun.ID[:8], baseRun.Timestamp.Format("Jan 02, 15:04"),
		compareRun.ID[:8], compareRun.Timestamp.Format("Jan 02, 15:04"),
		baseRun.TotalTests, 
		totalTestsClass,
		totalTestsDiffStr,
		compareRun.TotalTests,
		baseRun.PassedTests,
		passedTestsClass,
		passedTestsDiffStr,
		compareRun.PassedTests,
		baseRun.FailedTests,
		failedTestsClass,
		failedTestsDiffStr,
		compareRun.FailedTests,
		comparison.BaseSuccessRate,
		successRateClass,
		successRateDiffStr,
		comparison.CompareSuccessRate)
	
	// Show fixed tests
	if len(comparison.Fixed) > 0 {
		html += `<div class="change-group positive">
			<h5>Fixed Tests</h5>
			<ul class="change-list">`
		
		for _, test := range comparison.Fixed {
			html += fmt.Sprintf(`<li>%s</li>`, test)
		}
		
		html += `</ul></div>`
	}
	
	// Show newly failed tests
	if len(comparison.NewlyFailed) > 0 {
		html += `<div class="change-group negative">
			<h5>Newly Failed Tests</h5>
			<ul class="change-list">`
		
		for _, test := range comparison.NewlyFailed {
			html += fmt.Sprintf(`<li>%s</li>`, test)
		}
		
		html += `</ul></div>`
	}
	
	// Show new tests
	if len(comparison.New) > 0 {
		html += `<div class="change-group neutral">
			<h5>New Tests</h5>
			<ul class="change-list">`
		
		for _, test := range comparison.New {
			html += fmt.Sprintf(`<li>%s</li>`, test)
		}
		
		html += `</ul></div>`
	}
	
	// Show removed tests
	if len(comparison.Removed) > 0 {
		html += `<div class="change-group neutral">
			<h5>Removed Tests</h5>
			<ul class="change-list">`
		
		for _, test := range comparison.Removed {
			html += fmt.Sprintf(`<li>%s</li>`, test)
		}
		
		html += `</ul></div>`
	}
	
	html += `</div></div></div>`
	
	w.Write([]byte(html))
}

// renderBuildInfo renders HTML for build information
func renderBuildInfo(run TestRun) string {
	// Skip if no build info is available
	if run.Branch == "" && run.Commit == "" && run.TriggeredBy == "" && run.BuildVersion == "" {
		return ""
	}
	
	html := `<div class="build-info">`
	
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
