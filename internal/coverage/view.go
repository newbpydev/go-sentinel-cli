package coverage

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color definitions for coverage display
var (
	// Style definitions
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#0366d6")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#333333")).
			Background(lipgloss.Color("#f1f8ff")).
			Padding(0, 1)

	// Coverage percentage styles based on thresholds
	highCoverageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#28a745")) // Green
	medCoverageStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#f1e05a")) // Yellow
	lowCoverageStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#d73a49")) // Red

	// Line annotation styles
	coveredLineStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#28a745"))
	uncoveredLineStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#d73a49"))
	executionStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#0366d6"))
)

// View represents a coverage view for a file or package.
type View struct {
	metrics           *Metrics
	selectedFile      string
	selectedFileIndex int
	sortedFiles       []string // Sorted list of file names for predictable navigation
	width             int
	height            int
	keyHandlers       map[rune]func()
	visible           bool
	collector         *Collector // Reference to the collector for source code access
	showLowCoverage   bool       // Filter to show only files with low coverage
}

// NewCoverageView creates a new coverage view
func NewCoverageView(metrics *Metrics) *View {
	cv := &View{
		metrics:           metrics,
		width:             80, // Default width
		height:            20, // Default height
		keyHandlers:       make(map[rune]func()),
		visible:           true,
		collector:         nil, // Will be set when a coverage file is loaded
		selectedFileIndex: 0,
		showLowCoverage:   false,
	}

	// Initialize the sorted files list if we have metrics
	if metrics != nil && metrics.FileMetrics != nil {
		cv.updateSortedFiles()

		// Initialize selectedFile to first file if available
		if len(cv.sortedFiles) > 0 {
			cv.selectedFile = cv.sortedFiles[0]
		}
	}

	// Set up key handlers
	cv.keyHandlers['v'] = func() { cv.visible = !cv.visible }
	cv.keyHandlers['f'] = cv.ToggleLowCoverageFilter
	cv.keyHandlers['j'] = cv.SelectNextFile
	cv.keyHandlers['k'] = cv.SelectPreviousFile
	cv.keyHandlers['g'] = cv.SelectFirstFile
	cv.keyHandlers['G'] = cv.SelectLastFile

	return cv
}

// SetCollector sets the coverage collector for accessing source code
func (cv *View) SetCollector(collector *Collector) {
	cv.collector = collector
}

// GetSourceCode retrieves the source code for a given file
func (cv *View) GetSourceCode(filePath string) (map[int]string, error) {
	if cv.collector == nil {
		return make(map[int]string), nil
	}

	// Use the collector to get the source code
	return cv.collector.GetSourceCode(filePath)
}

// HasKeyBinding checks if the view has a specific key binding
func (cv *View) HasKeyBinding(key rune) bool {
	_, exists := cv.keyHandlers[key]
	return exists
}

// updateSortedFiles refreshes the sorted list of files from the metrics
func (cv *View) updateSortedFiles() {
	if cv.metrics == nil || cv.metrics.FileMetrics == nil {
		cv.sortedFiles = nil
		return
	}

	// Get all file names
	cv.sortedFiles = make([]string, 0, len(cv.metrics.FileMetrics))
	for filename := range cv.metrics.FileMetrics {
		// If we're showing only low coverage files, filter accordingly
		if cv.showLowCoverage {
			fileMetric := cv.metrics.FileMetrics[filename]
			// Only include files with coverage less than 50%
			if fileMetric.LineCoverage < 50.0 {
				cv.sortedFiles = append(cv.sortedFiles, filename)
			}
		} else {
			cv.sortedFiles = append(cv.sortedFiles, filename)
		}
	}

	// Sort the file names for consistent navigation
	sort.Strings(cv.sortedFiles)
}

// ToggleLowCoverageFilter toggles the filter to show only low coverage files
func (cv *View) ToggleLowCoverageFilter() {
	cv.showLowCoverage = !cv.showLowCoverage

	// Update the file list based on the new filter
	cv.updateSortedFiles()

	// Reset selection to first file
	cv.selectedFileIndex = 0
	if len(cv.sortedFiles) > 0 {
		cv.selectedFile = cv.sortedFiles[0]
	} else {
		cv.selectedFile = ""
	}
}

// SelectNextFile selects the next file in the list
func (cv *View) SelectNextFile() {
	if len(cv.sortedFiles) == 0 {
		return
	}

	// Move to next file, wrapping around if necessary
	cv.selectedFileIndex = (cv.selectedFileIndex + 1) % len(cv.sortedFiles)
	cv.selectedFile = cv.sortedFiles[cv.selectedFileIndex]
}

// SelectPreviousFile selects the previous file in the list
func (cv *View) SelectPreviousFile() {
	if len(cv.sortedFiles) == 0 {
		return
	}

	// Move to previous file, wrapping around if necessary
	cv.selectedFileIndex = (cv.selectedFileIndex - 1 + len(cv.sortedFiles)) % len(cv.sortedFiles)
	cv.selectedFile = cv.sortedFiles[cv.selectedFileIndex]
}

// SelectFirstFile selects the first file in the list
func (cv *View) SelectFirstFile() {
	if len(cv.sortedFiles) == 0 {
		return
	}

	cv.selectedFileIndex = 0
	cv.selectedFile = cv.sortedFiles[0]
}

// SelectLastFile selects the last file in the list
func (cv *View) SelectLastFile() {
	if len(cv.sortedFiles) == 0 {
		return
	}

	cv.selectedFileIndex = len(cv.sortedFiles) - 1
	cv.selectedFile = cv.sortedFiles[cv.selectedFileIndex]
}

// GetSelectedFile returns the currently selected file path
func (cv *View) GetSelectedFile() string {
	return cv.selectedFile
}

// HandleKey processes a key press
func (cv *View) HandleKey(key rune) bool {
	handler, exists := cv.keyHandlers[key]
	if exists {
		handler()
		return true
	}
	return false
}

// SetSize sets the size of the view
func (cv *View) SetSize(width, height int) {
	cv.width = width
	cv.height = height
}

// SelectFile selects a file to show detailed coverage for
func (cv *View) SelectFile(filename string) {
	cv.selectedFile = filename
}

// showOnlyLowCoverage filters view to show only low coverage files
// This is a stub implementation that will be completed in a future update
//
//nolint:unused // Will be implemented in the future
func (cv *View) showOnlyLowCoverage() {
	// Implementation would update internal filter state
	// For now this is a stub to satisfy the tests
}

// Render renders the coverage view
func (cv *View) Render() string {
	if !cv.visible || cv.metrics == nil {
		return ""
	}

	var sb strings.Builder

	// Render title
	sb.WriteString(titleStyle.Render("Coverage Report"))
	sb.WriteString("\n\n")

	// Render overall metrics
	sb.WriteString(headerStyle.Render("Overall Coverage Metrics:"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("Statement Coverage: %s\n", formatCoveragePercent(cv.metrics.StatementCoverage)))
	sb.WriteString(fmt.Sprintf("Branch Coverage:    %s\n", formatCoveragePercent(cv.metrics.BranchCoverage)))
	sb.WriteString(fmt.Sprintf("Function Coverage:  %s\n", formatCoveragePercent(cv.metrics.FunctionCoverage)))
	sb.WriteString(fmt.Sprintf("Line Coverage:      %s\n", formatCoveragePercent(cv.metrics.LineCoverage)))
	sb.WriteString("\n")

	// Render file list with coverage
	sb.WriteString(headerStyle.Render("File Coverage:"))
	sb.WriteString("\n")

	// Get sorted file names
	fileNames := make([]string, 0, len(cv.metrics.FileMetrics))
	for filename := range cv.metrics.FileMetrics {
		fileNames = append(fileNames, filename)
	}
	sort.Strings(fileNames)

	// Render each file with its coverage
	for _, filename := range fileNames {
		metrics := cv.metrics.FileMetrics[filename]
		coverageStyle := getCoverageStyle(metrics.LineCoverage)
		shortFileName := getShortFileName(filename)

		// Highlight the currently selected file
		isSelected := filename == cv.selectedFile
		linePrefix := "  "
		if isSelected {
			linePrefix = "▶ "
		}

		// Format line with index, selection marker, and coverage
		fileDisplay := fmt.Sprintf("%s%s: %s",
			linePrefix,
			shortFileName,
			coverageStyle.Render(fmt.Sprintf("%.1f%%", metrics.LineCoverage)))

		// If selected, use a highlighted style
		if isSelected {
			fileDisplay = lipgloss.NewStyle().
				Bold(true).
				Background(lipgloss.Color("#f1f8ff")).
				Render(fileDisplay)
		}

		sb.WriteString(fileDisplay + "\n")

		// If this is the selected file and we have detailed metrics, show them
		if cv.selectedFile == filename {
			sb.WriteString(fmt.Sprintf("  Statement: %.1f%%, Branch: %.1f%%, Function: %.1f%%\n",
				metrics.StatementCoverage,
				metrics.BranchCoverage,
				metrics.FunctionCoverage))
		}
	}

	return sb.String()
}

// FileCoverageView represents a view for detailed file coverage
type FileCoverageView struct {
	filename   string
	metrics    *FileMetrics
	sourceCode map[int]string
	width      int
	height     int
}

// NewFileCoverageView creates a new detailed file coverage view
func NewFileCoverageView(filename string, metrics *FileMetrics, sourceCode map[int]string) *FileCoverageView {
	return &FileCoverageView{
		filename:   filename,
		metrics:    metrics,
		sourceCode: sourceCode,
		width:      80,
		height:     20,
	}
}

// SetSize sets the size of the view
func (fcv *FileCoverageView) SetSize(width, height int) {
	fcv.width = width
	fcv.height = height
}

// Render renders the file coverage view
func (fcv *FileCoverageView) Render() string {
	if fcv.metrics == nil {
		return ""
	}

	var sb strings.Builder

	// Render file header
	sb.WriteString(titleStyle.Render(fmt.Sprintf("File Coverage: %s", getShortFileName(fcv.filename))))
	sb.WriteString("\n\n")

	// Render coverage summary
	coverageStyle := getCoverageStyle(fcv.metrics.LineCoverage)
	sb.WriteString(fmt.Sprintf("Coverage: %s\n\n",
		coverageStyle.Render(fmt.Sprintf("%.1f%%", fcv.metrics.LineCoverage))))

	// Render source code with coverage annotations
	if len(fcv.sourceCode) > 0 {
		// Get sorted line numbers
		lineNumbers := make([]int, 0, len(fcv.sourceCode))
		for line := range fcv.sourceCode {
			lineNumbers = append(lineNumbers, line)
		}
		sort.Ints(lineNumbers)

		// Render each line with annotations
		for _, lineNum := range lineNumbers {
			code := fcv.sourceCode[lineNum]
			execCount, hasExec := fcv.metrics.LineExecutionCounts[lineNum]

			// Determine line style based on coverage
			lineStyle := coveredLineStyle
			executionAnnotation := ""

			if hasExec {
				if execCount > 0 {
					// Covered line
					executionAnnotation = executionStyle.Render(fmt.Sprintf("%dx", execCount))
				} else {
					// Uncovered line
					lineStyle = uncoveredLineStyle
					executionAnnotation = uncoveredLineStyle.Render("0x")
				}
			}

			// Format the line with annotations
			sb.WriteString(fmt.Sprintf("%4d | %s %s\n",
				lineNum,
				lineStyle.Render(code),
				executionAnnotation))
		}
	} else {
		sb.WriteString("Source code not available")
	}

	return sb.String()
}

// Helper functions

// formatCoveragePercent formats a coverage percentage with appropriate styling
func formatCoveragePercent(percent float64) string {
	return getCoverageStyle(percent).Render(fmt.Sprintf("%.1f%%", percent))
}

// getCoverageStyle returns the appropriate style based on coverage percentage
func getCoverageStyle(percent float64) lipgloss.Style {
	if percent >= 80 {
		return highCoverageStyle
	} else if percent >= 50 {
		return medCoverageStyle
	}
	return lowCoverageStyle
}

// getShortFileName returns a shortened file name for display
func getShortFileName(fullPath string) string {
	parts := strings.Split(fullPath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return fullPath
}
