package ui

import (
	"github.com/newbpydev/go-sentinel/internal/coverage"
)

// CoverageState holds the state for coverage visualization
type CoverageState struct {
	Enabled           bool                    // Whether coverage view is enabled
	Metrics           *coverage.CoverageMetrics // Coverage metrics
	SelectedFile      string                    // Currently selected file for detailed view
	View              *coverage.CoverageView    // Main coverage view
	FileView          *coverage.FileCoverageView // Detailed file view
	ShowLowCoverageOnly bool                     // Filter to show only low coverage
}

// InitCoverage initializes the coverage state
func (m *TUITestExplorerModel) InitCoverage() {
	m.CoverageState = CoverageState{
		Enabled: false,
	}
}

// LoadCoverageData loads coverage data from a profile
func (m *TUITestExplorerModel) LoadCoverageData(profilePath string) error {
	// Create a collector to load coverage data
	collector, err := coverage.NewCollector(profilePath)
	if err != nil {
		return err
	}

	// Calculate coverage metrics
	metrics, err := collector.CalculateMetrics()
	if err != nil {
		return err
	}

	// Store metrics and create views
	m.CoverageState.Metrics = metrics
	m.CoverageState.View = coverage.NewCoverageView(metrics)
	
	// Set collector for source code access
	m.CoverageState.View.SetCollector(collector)
	
	// Enable coverage view
	m.CoverageState.Enabled = true
	m.ShowCoverageView = true
	
	return nil
}

// ToggleCoverageView toggles the coverage view on/off
func (m *TUITestExplorerModel) ToggleCoverageView() {
	m.CoverageState.Enabled = !m.CoverageState.Enabled
}

// SelectFileForCoverage selects a file to show detailed coverage information
func (m *TUITestExplorerModel) SelectFileForCoverage(filename string) error {
	if m.CoverageState.Metrics == nil {
		return nil // No metrics loaded
	}

	// Find file metrics
	fileMetrics, ok := m.CoverageState.Metrics.FileMetrics[filename]
	if !ok {
		return nil // File not found in metrics
	}

	// Create file view
	m.CoverageState.SelectedFile = filename
	
	// Try to get source code
	sourceCode, err := m.CoverageState.View.GetSourceCode(filename)
	if err != nil {
		// Just use empty source code if we can't get it
		sourceCode = make(map[int]string)
	}

	// Create file view
	m.CoverageState.FileView = coverage.NewFileCoverageView(filename, fileMetrics, sourceCode)

	return nil
}

// RenderCoverageView renders the coverage view or file view based on current state
func (m *TUITestExplorerModel) RenderCoverageView(width, height int) string {
	if !m.CoverageState.Enabled || m.CoverageState.Metrics == nil {
		return ""
	}

	// Set view sizes
	m.CoverageState.View.SetSize(width, height)

	// If a file is selected, show file view
	if m.CoverageState.SelectedFile != "" && m.CoverageState.FileView != nil {
		m.CoverageState.FileView.SetSize(width, height)
		return m.CoverageState.FileView.Render()
	}

	// Otherwise show the main coverage view
	return m.CoverageState.View.Render()
}

// ToggleLowCoverageFilter toggles showing only low coverage files
func (m *TUITestExplorerModel) ToggleLowCoverageFilter() {
	m.CoverageState.ShowLowCoverageOnly = !m.CoverageState.ShowLowCoverageOnly
}
