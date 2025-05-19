package handlers

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/newbpydev/go-sentinel/internal/coverage"
)

// CoverageHandler handles requests related to test coverage visualization
type CoverageHandler struct {
	templates *template.Template
	collector *coverage.CoverageCollector
}

// NewCoverageHandler creates a new coverage handler
func NewCoverageHandler(tmpl *template.Template) *CoverageHandler {
	// Create a default collector with a mock coverage file path
	// In a real implementation, this would be configured from settings
	collector, _ := coverage.NewCollector("coverage.out")
	return &CoverageHandler{
		templates: tmpl,
		collector: collector,
	}
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
	fileDetail := FileDetailData{
		FileID:           fileID,
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
		SourceLines: map[int]string{
			1:  "package parser",
			2:  "",
			3:  "import (",
			4:  "    \"fmt\"",
			5:  "    \"strings\"",
			6:  ")",
			7:  "",
			8:  "// Parser handles parsing of test output",
			9:  "type Parser struct {",
			10: "    buffer string",
			11: "}",
			12: "",
			13: "// NewParser creates a new parser instance",
			14: "func NewParser() *Parser {",
			15: "    return &Parser{}",
			16: "}",
			17: "",
			18: "// Parse parses the test output",
			19: "func (p *Parser) Parse(input string) ([]TestResult, error) {",
			20: "    if input == \"\" {",
			21: "        return nil, fmt.Errorf(\"empty input\")",
			22: "    }",
			23: "",
			24: "    lines := strings.Split(input, \"\\n\")",
			25: "    results := make([]TestResult, 0)",
			26: "",
			27: "    // Process lines",
			28: "    for _, line := range lines {",
			29: "        if strings.HasPrefix(line, \"--- PASS\") {",
			30: "            // Handle passing test",
			31: "            results = append(results, TestResult{Status: \"pass\"})",
			32: "        } else if strings.HasPrefix(line, \"--- FAIL\") {",
			33: "            // Handle failing test",
			34: "            results = append(results, TestResult{Status: \"fail\"})",
			35: "        }",
			36: "    }",
			37: "",
			38: "    return results, nil",
			39: "}",
		},
		LineStatuses: map[int]string{
			1:  "covered",
			2:  "not-executable",
			3:  "covered",
			4:  "covered",
			5:  "covered",
			6:  "covered",
			7:  "not-executable",
			8:  "not-executable",
			9:  "covered",
			10: "covered",
			11: "covered",
			12: "not-executable",
			13: "not-executable",
			14: "covered",
			15: "covered",
			16: "covered",
			17: "not-executable",
			18: "not-executable",
			19: "covered",
			20: "covered",
			21: "covered",
			22: "covered",
			23: "not-executable",
			24: "covered",
			25: "covered",
			26: "not-executable",
			27: "not-executable",
			28: "covered",
			29: "covered",
			30: "not-executable",
			31: "covered",
			32: "covered",
			33: "not-executable",
			34: "covered",
			35: "covered",
			36: "covered",
			37: "not-executable",
			38: "covered",
			39: "covered",
		},
	}

	// Render the file detail partial
	if err := h.templates.ExecuteTemplate(w, "partials/coverage-file-detail", fileDetail); err != nil {
		http.Error(w, "Failed to render file detail", http.StatusInternalServerError)
		return
	}
}

// SearchCoverage handles search requests for coverage files
func (h *CoverageHandler) SearchCoverage(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		// If no query, redirect to get all files
		http.Redirect(w, r, "/api/coverage/files", http.StatusSeeOther)
		return
	}

	// Create a new URL with the search parameter
	q := r.URL.Query()
	q.Set("search", query)

	// Create a new request with the updated query
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
