package handlers

import (
	"encoding/json"
	"net/http"
	"time"
)

// TestResult represents a single test result for the UI
type TestResult struct {
	Name     string    `json:"name"`
	Status   string    `json:"status"`
	Duration string    `json:"duration"`
	LastRun  time.Time `json:"lastRun"`
	Output   string    `json:"output,omitempty"`
}

// TestResultsHandler handles requests for test results
type TestResultsHandler struct {
	testRunner interface{} // This would be the actual test runner in a real implementation
}

// NewTestResultsHandler creates a new test results handler
func NewTestResultsHandler(testRunner interface{}) *TestResultsHandler {
	return &TestResultsHandler{
		testRunner: testRunner,
	}
}

// GetTestResults returns all test results
func (h *TestResultsHandler) GetTestResults(w http.ResponseWriter, r *http.Request) {
	// Get test results from the runner
	// This is a placeholder - in a real implementation, we would get actual test results
	results := getMockTestResults()

	// For HTMX requests, render HTML
	if r.Header.Get("HX-Request") == "true" {
		h.renderTestResultsHTML(w, results)
		return
	}

	// For API requests, return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
	})
}

// RunTest runs a specific test
func (h *TestResultsHandler) RunTest(w http.ResponseWriter, r *http.Request) {
	testName := r.PathValue("testName")
	if testName == "" {
		http.Error(w, "Test name is required", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual test running
	// This would call the TestRunner to run the specific test
	// h.testRunner.RunTest(testName)

	// For now, return mock data
	result := TestResult{
		Name:     testName,
		Status:   "passed",
		Duration: "0.3s",
		LastRun:  time.Now(),
	}

	// For HTMX requests, render HTML for the single test row
	if r.Header.Get("HX-Request") == "true" {
		h.renderSingleTestHTML(w, result)
		return
	}

	// For API requests, return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// RunAllTests runs all tests
func (h *TestResultsHandler) RunAllTests(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement actual running of all tests
	// This would call the TestRunner to run all tests
	// h.testRunner.RunAllTests()

	// Get results (mock for now)
	results := getMockTestResults()

	// For HTMX requests, render HTML
	if r.Header.Get("HX-Request") == "true" {
		h.renderTestResultsHTML(w, results)
		return
	}

	// For API requests, return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": results,
	})
}

// FilterTestResults filters test results based on criteria
func (h *TestResultsHandler) FilterTestResults(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	// Get results (mock for now)
	allResults := getMockTestResults()
	
	// Filter results if status parameter provided
	var filteredResults []TestResult
	if status != "" {
		for _, result := range allResults {
			if result.Status == status {
				filteredResults = append(filteredResults, result)
			}
		}
	} else {
		filteredResults = allResults
	}

	// For HTMX requests, render HTML
	if r.Header.Get("HX-Request") == "true" {
		h.renderTestResultsHTML(w, filteredResults)
		return
	}

	// For API requests, return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": filteredResults,
	})
}

// renderTestResultsHTML renders HTML for test results (for HTMX)
func (h *TestResultsHandler) renderTestResultsHTML(w http.ResponseWriter, results []TestResult) {
	// In a real implementation, this would use a template engine
	// For now, we'll render a simple HTML table
	w.Header().Set("Content-Type", "text/html")
	
	html := `<table aria-label="Test Results">
		<thead>
			<tr>
				<th scope="col">Test Name</th>
				<th scope="col">Status</th>
				<th scope="col">Duration</th>
				<th scope="col">Last Run</th>
				<th scope="col">Actions</th>
			</tr>
		</thead>
		<tbody>`
	
	for i, test := range results {
		statusClass := "passed"
		if test.Status == "failed" {
			statusClass = "failed"
		}
		
		html += `<tr class="test-row ` + statusClass + `" data-test-id="` + string(rune(i)) + `" tabindex="0">
			<td class="test-name">` + test.Name + `</td>
			<td class="test-status">
				<span class="status-badge ` + test.Status + `" aria-label="Test ` + test.Status + `">
					` + test.Status + `
				</span>
			</td>
			<td class="test-duration">` + test.Duration + `</td>
			<td class="test-last-run">` + formatTimeAgo(test.LastRun) + `</td>
			<td class="test-actions">
				<button class="run-button"
						hx-post="/api/run-test/` + test.Name + `"
						hx-target="closest tr"
						hx-swap="outerHTML"
						aria-label="Run ` + test.Name + `">
					Run
				</button>
			</td>
		</tr>`
	}
	
	html += `</tbody></table>`
	
	w.Write([]byte(html))
}

// renderSingleTestHTML renders HTML for a single test row (for HTMX)
func (h *TestResultsHandler) renderSingleTestHTML(w http.ResponseWriter, test TestResult) {
	// In a real implementation, this would use a template engine
	// For now, we'll render a simple HTML row
	w.Header().Set("Content-Type", "text/html")
	
	statusClass := "passed"
	if test.Status == "failed" {
		statusClass = "failed"
	}
	
	html := `<tr class="test-row ` + statusClass + `" data-test-id="0" tabindex="0">
		<td class="test-name">` + test.Name + `</td>
		<td class="test-status">
			<span class="status-badge ` + test.Status + `" aria-label="Test ` + test.Status + `">
				` + test.Status + `
			</span>
		</td>
		<td class="test-duration">` + test.Duration + `</td>
		<td class="test-last-run">` + formatTimeAgo(test.LastRun) + `</td>
		<td class="test-actions">
			<button class="run-button"
					hx-post="/api/run-test/` + test.Name + `"
					hx-target="closest tr"
					hx-swap="outerHTML"
					aria-label="Run ` + test.Name + `">
				Run
			</button>
		</td>
	</tr>`
	
	w.Write([]byte(html))
}

// Helper functions

// formatTimeAgo formats a time as relative (e.g., "2 min ago")
func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)
	
	if duration.Minutes() < 1 {
		return "just now"
	} else if duration.Minutes() < 60 {
		minutes := int(duration.Minutes())
		return string(rune(minutes)) + " min ago"
	} else if duration.Hours() < 24 {
		hours := int(duration.Hours())
		return string(rune(hours)) + " hours ago"
	}
	
	days := int(duration.Hours() / 24)
	return string(rune(days)) + " days ago"
}

// getMockTestResults returns mock test results for demonstration
func getMockTestResults() []TestResult {
	now := time.Now()
	
	return []TestResult{
		{
			Name:     "TestParseConfig",
			Status:   "passed",
			Duration: "0.8s",
			LastRun:  now.Add(-2 * time.Minute),
		},
		{
			Name:     "TestValidateInput",
			Status:   "passed",
			Duration: "1.2s",
			LastRun:  now.Add(-2 * time.Minute),
		},
		{
			Name:     "TestProcessResults",
			Status:   "failed",
			Duration: "2.1s",
			LastRun:  now.Add(-2 * time.Minute),
			Output:   "Expected 5 results, got 4",
		},
		{
			Name:     "TestExportReport",
			Status:   "passed",
			Duration: "0.9s",
			LastRun:  now.Add(-2 * time.Minute),
		},
	}
}
