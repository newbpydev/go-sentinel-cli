package ui

import (
	"fmt"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/go-sentinel/internal/coverage"
)

// ExportCoverageHTMLReport exports the coverage data as an HTML report
func (m *TUITestExplorerModel) ExportCoverageHTMLReport(coverageFile string) tea.Cmd {
	return func() tea.Msg {
		// Ensure we have valid coverage data
		if m.CoverageState.Metrics == nil {
			return ErrorMsg{Error: fmt.Errorf("no coverage data available to export")}
		}

		// Generate timestamp for the filename
		timestamp := time.Now().Format("20060102-150405")
		htmlOutputPath := filepath.Join("coverage-reports", fmt.Sprintf("coverage-%s.html", timestamp))

		// Create the report options
		options := coverage.CoverageReportOptions{
			CoverageFile:     coverageFile,
			OutputPath:       htmlOutputPath,
			Title:            "Go-Sentinel Coverage Report",
			IncludeTimestamp: true,
		}

		// Generate the enhanced HTML report
		err := coverage.GenerateEnhancedCoverageReport(options)
		if err != nil {
			return ErrorMsg{Error: fmt.Errorf("failed to generate HTML report: %w", err)}
		}

		// Return success message
		return HTMLReportGeneratedMsg{
			ReportPath: htmlOutputPath,
		}
	}
}

// HTMLReportGeneratedMsg indicates an HTML report was successfully generated
type HTMLReportGeneratedMsg struct {
	ReportPath string
}
