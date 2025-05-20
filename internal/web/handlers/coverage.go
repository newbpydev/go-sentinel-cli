package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/newbpydev/go-sentinel/internal/coverage"
)

// CoverageHandler handles coverage report requests
type CoverageHandler struct {
	templates *template.Template
	collector *coverage.Collector
	mu        sync.RWMutex
}

// NewCoverageHandler creates a new coverage handler
func NewCoverageHandler(tmpl *template.Template) *CoverageHandler {
	return &CoverageHandler{
		templates: tmpl,
	}
}

// Initialize sets up the coverage collector
func (h *CoverageHandler) Initialize(coverageFile string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Close existing collector if any
	if h.collector != nil {
		h.collector.Close()
	}

	// Create new collector
	collector, err := coverage.NewCollector(coverageFile)
	if err != nil {
		return fmt.Errorf("failed to create coverage collector: %w", err)
	}

	h.collector = collector
	return nil
}

// Close cleans up resources
func (h *CoverageHandler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.collector != nil {
		if err := h.collector.Close(); err != nil {
			return fmt.Errorf("failed to close coverage collector: %w", err)
		}
		h.collector = nil
	}
	return nil
}

// ServeHTTP handles HTTP requests for coverage data
func (h *CoverageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	collector := h.collector
	h.mu.RUnlock()

	if collector == nil {
		http.Error(w, "Coverage collector not initialized", http.StatusServiceUnavailable)
		return
	}

	// Handle request...
}

// CoverageSummaryData represents the data for the coverage summary
type CoverageSummaryData struct {
	TotalCoverage    float64 `json:"totalCoverage"`
	LineCoverage     float64 `json:"lineCoverage"`
	FunctionCoverage float64 `json:"functionCoverage"`
	BranchCoverage   float64 `json:"branchCoverage"`
	TotalFiles       int     `json:"totalFiles"`
	LowCoverageFiles int     `json:"lowCoverageFiles"`
	LastUpdated      string  `json:"lastUpdated"`
}

// FileListData represents the data for the file list
type FileListData struct {
	Files       []FileCoverageData `json:"files"`
	CurrentPage int                `json:"currentPage"`
	TotalPages  int                `json:"totalPages"`
	Filter      string             `json:"filter"`
	Search      string             `json:"search"`
}

// FileCoverageData represents the coverage data for a single file
type FileCoverageData struct {
	FileID           string  `json:"fileID"`
	FilePath         string  `json:"filePath"`
	LineCoverage     float64 `json:"lineCoverage"`
	FunctionCoverage float64 `json:"functionCoverage"`
	BranchCoverage   float64 `json:"branchCoverage"`
	CoveredLines     int     `json:"coveredLines"`
	TotalLines       int     `json:"totalLines"`
	CoveredFunctions int     `json:"coveredFunctions"`
	TotalFunctions   int     `json:"totalFunctions"`
	CoveredBranches  int     `json:"coveredBranches"`
	TotalBranches    int     `json:"totalBranches"`
}

// FileDetailData represents the detailed coverage data for a file
type FileDetailData struct {
	FileID           string         `json:"fileID"`
	FilePath         string         `json:"filePath"`
	LineCoverage     float64        `json:"lineCoverage"`
	FunctionCoverage float64        `json:"functionCoverage"`
	BranchCoverage   float64        `json:"branchCoverage"`
	CoveredLines     int            `json:"coveredLines"`
	TotalLines       int            `json:"totalLines"`
	CoveredFunctions int            `json:"coveredFunctions"`
	TotalFunctions   int            `json:"totalFunctions"`
	CoveredBranches  int            `json:"coveredBranches"`
	TotalBranches    int            `json:"totalBranches"`
	SourceLines      map[int]string `json:"sourceLines"`
	LineStatuses     map[int]string `json:"lineStatuses"`
}

// GetCoverageSummary handles requests for the coverage summary
func (h *CoverageHandler) GetCoverageSummary(w http.ResponseWriter, _ *http.Request) {
	// In a real implementation, this would fetch actual coverage data
	// For now, we'll use mock data
	summary := CoverageSummaryData{
		TotalCoverage:    78.5,
		LineCoverage:     82.3,
		FunctionCoverage: 75.0,
		BranchCoverage:   68.7,
		TotalFiles:       42,
		LowCoverageFiles: 8,
		LastUpdated:      time.Now().Format("Jan 2, 2006 15:04:05"),
	}

	// Render the coverage summary partial
	if err := h.templates.ExecuteTemplate(w, "partials/coverage-summary", summary); err != nil {
		http.Error(w, "Failed to render coverage summary", http.StatusInternalServerError)
		return
	}
}

// GetCoverageFiles handles requests for the coverage file list
func (h *CoverageHandler) GetCoverageFiles(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	filter := r.URL.Query().Get("filter")
	search := r.URL.Query().Get("search")

	// In a real implementation, this would fetch actual coverage data
	// For now, we'll use mock data
	files := []FileCoverageData{
		{
			FileID:           "file1",
			FilePath:         "internal/parser/parser.go",
			LineCoverage:     92.5,
			FunctionCoverage: 100.0,
			BranchCoverage:   85.7,
			CoveredLines:     185,
			TotalLines:       200,
			CoveredFunctions: 12,
			TotalFunctions:   12,
			CoveredBranches:  18,
			TotalBranches:    21,
		},
		{
			FileID:           "file2",
			FilePath:         "internal/runner/runner.go",
			LineCoverage:     78.3,
			FunctionCoverage: 80.0,
			BranchCoverage:   75.0,
			CoveredLines:     130,
			TotalLines:       166,
			CoveredFunctions: 8,
			TotalFunctions:   10,
			CoveredBranches:  15,
			TotalBranches:    20,
		},
		{
			FileID:           "file3",
			FilePath:         "internal/web/server/server.go",
			LineCoverage:     65.2,
			FunctionCoverage: 70.0,
			BranchCoverage:   60.0,
			CoveredLines:     120,
			TotalLines:       184,
			CoveredFunctions: 7,
			TotalFunctions:   10,
			CoveredBranches:  12,
			TotalBranches:    20,
		},
		{
			FileID:           "file4",
			FilePath:         "internal/coverage/collector.go",
			LineCoverage:     88.7,
			FunctionCoverage: 90.0,
			BranchCoverage:   85.0,
			CoveredLines:     110,
			TotalLines:       124,
			CoveredFunctions: 9,
			TotalFunctions:   10,
			CoveredBranches:  17,
			TotalBranches:    20,
		},
		{
			FileID:           "file5",
			FilePath:         "internal/tui/tui.go",
			LineCoverage:     45.8,
			FunctionCoverage: 50.0,
			BranchCoverage:   40.0,
			CoveredLines:     55,
			TotalLines:       120,
			CoveredFunctions: 5,
			TotalFunctions:   10,
			CoveredBranches:  8,
			TotalBranches:    20,
		},
	}

	// Apply filtering based on coverage threshold
	var filteredFiles []FileCoverageData
	if filter != "" && filter != "all" {
		for _, file := range files {
			switch filter {
			case "low":
				if file.LineCoverage < 50.0 {
					filteredFiles = append(filteredFiles, file)
				}
			case "medium":
				if file.LineCoverage >= 50.0 && file.LineCoverage < 80.0 {
					filteredFiles = append(filteredFiles, file)
				}
			case "high":
				if file.LineCoverage >= 80.0 {
					filteredFiles = append(filteredFiles, file)
				}
			}
		}
	} else {
		filteredFiles = files
	}

	// Apply search filtering if provided
	if search != "" {
		var searchResults []FileCoverageData
		for _, file := range filteredFiles {
			// Simple case-insensitive substring search
			if containsIgnoreCase(file.FilePath, search) {
				searchResults = append(searchResults, file)
			}
		}
		filteredFiles = searchResults
	}

	// Pagination
	itemsPerPage := 10
	totalPages := (len(filteredFiles) + itemsPerPage - 1) / itemsPerPage

	// Ensure page is within bounds
	if page > totalPages && totalPages > 0 {
		page = totalPages
	}

	// Calculate slice bounds for pagination
	startIdx := (page - 1) * itemsPerPage
	endIdx := startIdx + itemsPerPage
	if endIdx > len(filteredFiles) {
		endIdx = len(filteredFiles)
	}

	// Get the current page of files
	var pageFiles []FileCoverageData
	if startIdx < len(filteredFiles) {
		pageFiles = filteredFiles[startIdx:endIdx]
	}

	// Prepare the response data
	data := FileListData{
		Files:       pageFiles,
		CurrentPage: page,
		TotalPages:  totalPages,
		Filter:      filter,
		Search:      search,
	}

	// Render the file list partial
	if err := h.templates.ExecuteTemplate(w, "partials/coverage-file-list", data); err != nil {
		http.Error(w, "Failed to render coverage file list", http.StatusInternalServerError)
		return
	}
}

// GetFileDetail handles requests for detailed file coverage
func (h *CoverageHandler) GetFileDetail(w http.ResponseWriter, r *http.Request) {
	// Extract file ID from the URL
	fileID := r.URL.Query().Get("id")
	if fileID == "" {
		http.Error(w, "File ID is required", http.StatusBadRequest)
		return
	}

	// In a real implementation, this would fetch actual file coverage data
	// For now, we'll use mock data
	mockFiles := map[string]FileDetailData{
		"file1": {
			FileID:           "file1",
			FilePath:         "internal/parser/parser.go",
			LineCoverage:     92.5,
			FunctionCoverage: 100.0,
			BranchCoverage:   85.7,
			CoveredLines:     185,
			TotalLines:       200,
			CoveredFunctions: 12,
			TotalFunctions:   12,
			CoveredBranches:  18,
			TotalBranches:    21,
			SourceLines:      map[int]string{1: "package parser"},
			LineStatuses:     map[int]string{1: "covered"},
		},
	}
	fileDetail, ok := mockFiles[fileID]
	if !ok {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if err := h.templates.ExecuteTemplate(w, "partials/coverage-file-detail", fileDetail); err != nil {
		http.Error(w, "Failed to render file detail", http.StatusInternalServerError)
		return
	}
}

// SearchCoverage handles search requests for coverage files
func (h *CoverageHandler) SearchCoverage(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	search := r.URL.Query().Get("search")
	if query == "" && search == "" {
		// If no query or search, redirect to get all files
		http.Redirect(w, r, "/api/coverage/files", http.StatusSeeOther)
		return
	}

	// Prefer 'query' if present, else use 'search'
	searchVal := query
	if searchVal == "" {
		searchVal = search
	}

	q := r.URL.Query()
	q.Set("search", searchVal)

	newURL := *r.URL
	newURL.RawQuery = q.Encode()
	newReq := r.Clone(r.Context())
	newReq.URL = &newURL

	h.GetCoverageFiles(w, newReq)
}

// FilterCoverage handles filter requests for coverage files
func (h *CoverageHandler) FilterCoverage(w http.ResponseWriter, r *http.Request) {
	filter := r.URL.Query().Get("filter")
	if filter == "" {
		filter = "all"
	}

	// Create a new URL with the filter parameter
	q := r.URL.Query()
	q.Set("filter", filter)

	// Create a new request with the updated query
	newURL := *r.URL
	newURL.RawQuery = q.Encode()
	newReq := r.Clone(r.Context())
	newReq.URL = &newURL

	h.GetCoverageFiles(w, newReq)
}

// Helper function for case-insensitive substring search
func containsIgnoreCase(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}
