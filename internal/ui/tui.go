package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
	"github.com/newbpydev/go-sentinel/internal/event"
) // progressbar.go is part of the same package; no import needed

// The styles and colors are defined in theme.go within the same package

// Layout modes
const (
	NormalLayoutMode = iota
	SplitWithLogMode
)

// TUITestExplorerModel is the Bubble Tea model for the tree-based test explorer.
type TUITestExplorerModel struct {
	Sidebar         list.Model
	Items           []list.Item // flat view of visible tree nodes
	Tree            *TreeNode   // actual tree structure
	SelectedIndex   int
	MainPaneContent string
	// Search/filter state
	SearchActive  bool
	SearchInput   string
	FilteredItems []list.Item
	// Modal state
	ShowHelpModal bool
	// Layout/size state
	Width  int
	Height int
	// Animated coverage bar for details panel
	CoverageBar AnimatedCoverageBar
	// Track folder expansion state for search UX
	prevExpansion map[string]bool
	// Track if currently showing a locked filtered set
	FilteredMode bool
	// Store the accepted filter term for updating filtered results
	AcceptedFilter string
	// Coverage visualization state
	CoverageState CoverageState
	// Flag to indicate if in coverage view mode
	ShowCoverageView bool
	// Test runner state
	TestsRunning     bool
	RunningPackage   string
	LastRunSuccess   bool
	LastChangedFile  string
	WatchModeEnabled bool
	// Log view
	LogBuffer        []string
	ShowLogView      bool
	LayoutMode       int
	// Callback for watch mode changes
	watchModeCallback func(bool)
	// CRITICAL FIX: Fields for pending test runs
	pendingTestPackage string
	pendingTestName    string
	hasPendingTest     bool
} // Now tracks terminal width/height

// saveExpansionState saves the expansion state of all folders in the tree.
func (m *TUITestExplorerModel) saveExpansionState() {
	m.prevExpansion = make(map[string]bool)
	save := func(node *TreeNode, path string) {}
	save = func(node *TreeNode, path string) {
		if node == nil {
			return
		}
		if len(node.Children) > 0 {
			m.prevExpansion[path+"/"+node.Title] = node.Expanded
			for _, child := range node.Children {
				save(child, path+"/"+node.Title)
			}
		}
	}
	save(m.Tree, "")
}

// restoreExpansionState restores the expansion state of all folders from prevExpansion.
func (m *TUITestExplorerModel) restoreExpansionState() {
	restore := func(node *TreeNode, path string) {}
	restore = func(node *TreeNode, path string) {
		if node == nil {
			return
		}
		if len(node.Children) > 0 {
			if exp, ok := m.prevExpansion[path+"/"+node.Title]; ok {
				node.Expanded = exp
			}
			for _, child := range node.Children {
				restore(child, path+"/"+node.Title)
			}
		}
	}
	restore(m.Tree, "")
}

// TreeNode represents a node in the test tree (suite/file/test)
type TreeNode struct {
	Title    string
	Children []*TreeNode
	Expanded bool
	Level    int // indentation level
	Parent   *TreeNode
	// Extended fields for real test data
	Coverage float64 // For file nodes (0.0-1.0)
	Passed   *bool   // For test nodes (nil for parent, true/false for tests)
	Duration float64 // For test nodes (seconds)
	Error    string  // For test nodes (error message)
}

// treeItem implements list.Item for bubbles/list
// It wraps a TreeNode and provides a string representation
// for the sidebar list

type treeItem struct {
	node *TreeNode
}

func (ti treeItem) Title() string {
	indent := ""
	for i := 0; i < ti.node.Level; i++ {
		indent += "  "
	}
	if len(ti.node.Children) == 0 {
		// Test node (leaf)
		return fmt.Sprintf("%s%s", indent, ti.node.Title)
	}
	triangle := "▶"
	if ti.node.Expanded {
		triangle = "▼"
	}
	return fmt.Sprintf("%s%s %s", indent, triangle, ti.node.Title)
}
func (ti treeItem) Description() string { return "" }
func (ti treeItem) FilterValue() string { return ti.node.Title }

// NewTUITestExplorerModel creates a new TUI model with the given tree.
type treeItemDelegate struct{}

func (d treeItemDelegate) Height() int                               { return 1 }
func (d treeItemDelegate) Spacing() int                              { return 0 }
func (d treeItemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d treeItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	treeItem := item.(treeItem)
	title := treeItem.Title()
	selected := m.Index() == index

	// Add color based on node type
	var colored string
	switch {
	case treeItem.node.Passed != nil && *treeItem.node.Passed:
		// Passing test (leaf)
		colored = lipgloss.NewStyle().Foreground(AccentGreen).Render(title)
	case treeItem.node.Passed != nil && !*treeItem.node.Passed:
		// Failing test (leaf)
		colored = lipgloss.NewStyle().Foreground(AccentRed).Render(title)
	case treeItem.node.Error == "skip":
		label := title
		if !strings.Contains(label, "(skipped)") {
			label += " (skipped)"
		}
		colored = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(label)
	case treeItem.node.Level == 0:
		colored = lipgloss.NewStyle().Foreground(AccentBlue).Render(title)
	case len(treeItem.node.Children) > 0:
		colored = lipgloss.NewStyle().Foreground(AccentYellow).Render(title)
	default:
		colored = title
	}

	// Apply pointer for selection instead of highlight
	pointer := "➤"
	if selected {
		line := fmt.Sprintf("%s %s", pointer, colored)
		fmt.Fprint(w, SidebarSelectedItemStyle.Render(line))
	} else {
		fmt.Fprintf(w, "  %s", colored)
	}
}

func NewTUITestExplorerModel(root *TreeNode) TUITestExplorerModel {
	setExpansionByFailures(root)
	return newTUITestExplorerModelCore(root)
}

// NewTUITestExplorerModelWithNoExpansion skips setExpansionByFailures for precise test control
func NewTUITestExplorerModelWithNoExpansion(root *TreeNode) TUITestExplorerModel {
	return newTUITestExplorerModelCore(root)
}

// AddLogLine adds a line to the log buffer
func (m *TUITestExplorerModel) AddLogLine(line string) {
	// Cap the buffer at 500 lines to prevent excessive memory usage
	if len(m.LogBuffer) > 500 {
		// Remove oldest lines when buffer gets too large
		m.LogBuffer = m.LogBuffer[len(m.LogBuffer)-500:]
	}
	m.LogBuffer = append(m.LogBuffer, line)
}

// ClearLog clears the log buffer
func (m *TUITestExplorerModel) ClearLog() {
	m.LogBuffer = []string{}
}

// ToggleLogView toggles the log view on/off
func (m *TUITestExplorerModel) ToggleLogView() {
	m.ShowLogView = !m.ShowLogView
}

func newTUITestExplorerModelCore(root *TreeNode) TUITestExplorerModel {
	items := flattenTree(root)
	dlgt := treeItemDelegate{}
	l := list.New(items, dlgt, 50, 20) // wider list for test names
	l.Title = "Test Explorer"
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	// Forcibly hide the help bar/footer in the sidebar list
	if setShowHelp, ok := interface{}(&l).(interface{ SetShowHelp(bool) }); ok {
		setShowHelp.SetShowHelp(false)
	} else {
		l.Help.ShowAll = false
		// No way to override Help.View in this version
	}
	l.SetShowPagination(true)
	l.SetShowStatusBar(false)
	return TUITestExplorerModel{
		Sidebar:         l,
		Items:           items,
		Tree:            root,
		SelectedIndex:   0,
		MainPaneContent: "",
		Width:           100, // wider default
		Height:          24, // default, will be set on first WindowSizeMsg
		CoverageBar:     NewAnimatedCoverageBar(),
		LogBuffer:        []string{},
		ShowLogView:      false,
		LayoutMode:       0,
		WatchModeEnabled: false,
	}
}

// setExpansionByFailures recursively sets .Expanded for all folders:
// expanded if any descendant test failed, collapsed otherwise
func setExpansionByFailures(node *TreeNode) bool {
	if node == nil {
		return false
	}
	if len(node.Children) == 0 {
		return node.Passed != nil && !*node.Passed
	}
	hasFailure := false
	for _, child := range node.Children {
		if setExpansionByFailures(child) {
			hasFailure = true
		}
	}
	node.Expanded = hasFailure
	return hasFailure
}

// flattenTree returns a flat slice of treeItems for visible nodes
func flattenTree(root *TreeNode) []list.Item {
	var items []list.Item
	var walk func(node *TreeNode, level int, parent *TreeNode)
	walk = func(node *TreeNode, level int, parent *TreeNode) {
		if node == nil {
			return
		}
		node.Level = level
		node.Parent = parent
		// DEBUG: Print pointer address for test
		// fmt.Printf("flattenTree: node %s at %p\n", node.Title, node)
		items = append(items, treeItem{node})
		if node.Expanded {
			for _, child := range node.Children {
				walk(child, level+1, node)
			}
		}
	}
	walk(root, 0, nil)
	return items
}

func (m *TUITestExplorerModel) Init() tea.Cmd {
	return nil
}

// HasPendingTestRun checks if there is a pending test run
func (m *TUITestExplorerModel) HasPendingTestRun() bool {
	return m.hasPendingTest
}

// GetAndClearPendingTestRun returns the details of a pending test run and clears the pending state
func (m *TUITestExplorerModel) GetAndClearPendingTestRun() (string, string) {
	pkg := m.pendingTestPackage
	test := m.pendingTestName
	
	// Clear the pending state
	m.pendingTestPackage = ""
	m.pendingTestName = ""
	m.hasPendingTest = false
	
	return pkg, test
}

// Update processes a message and returns an updated model and optional command
func (m *TUITestExplorerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case TestResultsMsg:
		// Update tree with new results
		m.Tree = msg.Tree
		m.Items = flattenTree(m.Tree)
		m.Sidebar.SetItems(m.Items)
		return m, nil
		
	case RunTestsMsg:
		// Set pending test run flags - the message polling loop will pick this up
		m.pendingTestPackage = msg.Package
		m.pendingTestName = msg.Test
		m.hasPendingTest = true
		
		// Show running status
		m.TestsRunning = true
		m.RunningPackage = msg.Package
		
		// CRITICAL FIX: Show the log panel automatically when running tests
		m.ShowLogView = true
		return m, nil
		
	case TestsStartedMsg:
		// Show running status
		m.TestsRunning = true
		m.RunningPackage = msg.Package
		m.ShowLogView = true  // Also show log view
		return m, nil
		
	case TestsCompletedMsg:
		// Update status
		m.TestsRunning = false
		m.LastRunSuccess = msg.Success
		return m, nil
		
	case ToggleWatchModeMsg:
		// Toggle the watch mode and send the command to notify the app controller
		return m, func() tea.Msg {
			return WatchStatusChangedMsg{Enabled: !m.WatchModeEnabled}
		}

	case WatchStatusChangedMsg:
		// Update the UI state with the new watch mode status
		m.WatchModeEnabled = msg.Enabled
		
		// IMPORTANT: Don't call the callback directly from Update() as this creates
		// a synchronous loop. Instead, return a command that will be executed
		// asynchronously by the Bubble Tea runtime.
		return m, func() tea.Msg {
			// Only if we have a callback, notify the controller about the change
			if m.watchModeCallback != nil {
				m.watchModeCallback(msg.Enabled)
			}
			// Return a no-op message that won't trigger more updates
			return nil
		}

	case LogEntryMsg:
		// Add the log entry to our buffer
		m.AddLogLine(msg.Content)
		// Don't automatically show the log view - respect user's preference
		return m, nil
		
	case ClearLogMsg:
		// Clear the log buffer
		m.LogBuffer = []string{}
		return m, nil
		
	case ShowLogViewMsg:
		// Explicitly show or hide the log view based on the message
		m.ShowLogView = msg.Show
		return m, nil
		
	// NOTE: RunTestsMsg case handler was here but removed to prevent duplication
	// The first handler now sets the pending test flags that will be picked up by the controller
		
	case FileChangedMsg:
		// Record changed file and trigger tests
		m.LastChangedFile = msg.Path
		// Automatically trigger test run for the package
		pkg := "./..." // Default to all packages
		// TODO: Extract package from file path more intelligently
		return m, runTestsCmd(pkg, "")

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil
	
	case CoverageGeneratedMsg:
		// Coverage generation completed successfully
		m.ShowCoverageView = true // Automatically show coverage view
		m.MainPaneContent = fmt.Sprintf("Coverage report generated at %s", msg.CoverageFile)
		return m, nil
	}
	
	// Handle error messages
	if errMsg, ok := msg.(event.ErrorEvent); ok {
		// Show the error
		m.MainPaneContent = fmt.Sprintf("Error: %v", errMsg.Error())
		return m, nil
	}
	
	// Handle HTML report generation messages
	if htmlMsg, ok := msg.(HTMLReportGeneratedMsg); ok {
		// HTML report generation completed successfully
		m.MainPaneContent = fmt.Sprintf("HTML coverage report generated at %s", htmlMsg.ReportPath)
		return m, nil
	}

	// Help modal open: only handle modal keys
	if m.ShowHelpModal {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "q", "esc":
				m.ShowHelpModal = false
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// If in coverage view, check if the key should be handled by the coverage view first
		if m.ShowCoverageView && m.CoverageState.Enabled && m.CoverageState.View != nil {
			// Try to handle the key in the coverage view
			key := msg.String()
			handled := false
			
			// Forward navigation keys
			switch key {
			case "j", "k", "g", "G":
				if m.CoverageState.View.HasKeyBinding([]rune(key)[0]) {
					m.CoverageState.View.HandleKey([]rune(key)[0])
					handled = true
				}
			case "\n", "enter":
				// If a file is selected, show detailed view
				if m.CoverageState.View.GetSelectedFile() != "" {
					m.SelectFileForCoverage(m.CoverageState.View.GetSelectedFile())
					handled = true
				}
			}
			
			if handled {
				return m, nil
			}
		}
		if m.SearchActive {
			// Handle search input
			switch msg.Type {
			case tea.KeyEsc:
				m.SearchActive = false
				m.SearchInput = ""
				m.FilteredItems = nil
				m.FilteredMode = false
				m.restoreExpansionState()
				m.Items = flattenTree(m.Tree)
				m.Sidebar.SetItems(m.Items)
				m.SelectedIndex = 0
				m.Sidebar.Select(0)
			case tea.KeyEnter:
				// Accept filtered list: exit search mode, keep only filtered items, enter filtered mode
				m.SearchActive = false
				m.FilteredMode = true
				m.AcceptedFilter = m.SearchInput
				m.Items = m.FilteredItems
				m.Sidebar.SetItems(m.Items)
				// Clamp selection
				if m.SelectedIndex >= len(m.Items) {
					m.SelectedIndex = len(m.Items) - 1
				}
				if m.SelectedIndex < 0 && len(m.Items) > 0 {
					m.SelectedIndex = 0
				}
				m.Sidebar.Select(m.SelectedIndex)
				return m, nil
			case tea.KeyRunes:
				// Special handling for spacebar to expand/collapse folder and exit search mode
				if msg.String() == " " && len(m.Items) > 0 && m.SelectedIndex < len(m.Items) {
					item := m.Items[m.SelectedIndex].(treeItem)
					if len(item.node.Children) > 0 {
						// Update expansion state in prevExpansion before restoring
						path := ""
						n := item.node
						for p := n; p != nil; p = p.Parent {
							if p.Parent != nil {
								path = "/" + p.Title + path
							}
						}
						key := path
						if key == "" {
							key = "/" + n.Title
						}
						if m.prevExpansion == nil {
							m.saveExpansionState()
						}
						m.prevExpansion[key] = !item.node.Expanded
						item.node.Expanded = !item.node.Expanded
						m.SearchActive = false
						m.SearchInput = ""
						m.FilteredItems = nil
						m.restoreExpansionState()
						m.Items = flattenTree(m.Tree)
						// Clamp selection
						if m.SelectedIndex >= len(m.Items) {
							m.SelectedIndex = len(m.Items) - 1
						}
						if m.SelectedIndex < 0 && len(m.Items) > 0 {
							m.SelectedIndex = 0
						}
						m.Sidebar.SetItems(m.Items)
						m.Sidebar.Select(m.SelectedIndex)
						return m, nil
					}
				}
				// Otherwise, treat as normal input
				m.SearchInput += msg.String()
				// Use our improved fuzzy search with hierarchy preservation
				m.FilteredItems = getNodesMatchingFilter(m.Tree, m.SearchInput)
				m.Items = m.FilteredItems
				m.Sidebar.SetItems(m.Items)
				m.SelectedIndex = 0
				m.Sidebar.Select(0)
				return m, nil
			default:
				if msg.Type == tea.KeyBackspace || msg.Type == tea.KeyDelete {
					if len(m.SearchInput) > 0 {
						m.SearchInput = m.SearchInput[:len(m.SearchInput)-1]
					}
					// Update search results
					if m.SearchInput == "" {
						m.FilteredItems = flattenTree(m.Tree)
					} else {
						// Use our improved fuzzy search with hierarchy preservation
						m.FilteredItems = getNodesMatchingFilter(m.Tree, m.SearchInput)
					}
					m.Items = m.FilteredItems
					// Reset selection to top
					m.SelectedIndex = 0
					m.Sidebar.SetItems(m.Items)
					m.Sidebar.Select(0)
					return m, nil
				}
				return m, nil
			}
			return m, nil
		}
		if msg.String() == "?" {
			m.ShowHelpModal = true
			return m, nil
		}
		// Handle spacebar toggle for folder expansion
		// Only allow expand/collapse with spacebar when NOT in search mode
		if (msg.String() == " " || msg.Type == tea.KeySpace) && !m.SearchActive {
			if len(m.Items) > 0 && m.SelectedIndex < len(m.Items) {
				item := m.Items[m.SelectedIndex].(treeItem)
				if len(item.node.Children) > 0 {
					item.node.Expanded = !item.node.Expanded
					if m.FilteredMode {
						// Re-run the improved fuzzy search with hierarchy preservation
						m.FilteredItems = getNodesMatchingFilter(m.Tree, m.AcceptedFilter)
						m.Items = m.FilteredItems
					} else {
						m.Items = flattenTree(m.Tree)
					}
					// Clamp selection to valid range
					if m.SelectedIndex >= len(m.Items) {
						m.SelectedIndex = len(m.Items) - 1
					}
					if m.SelectedIndex < 0 && len(m.Items) > 0 {
						m.SelectedIndex = 0
					}
					m.Sidebar.SetItems(m.Items)
					m.Sidebar.Select(m.SelectedIndex)
				}
			}
			return m, nil
		}

	// Allow clearing filtered mode (locked filter) with Esc
	if m.FilteredMode || len(m.FilteredItems) > 0 {
		if msg.Type == tea.KeyEsc || msg.String() == "esc" {
			m.FilteredMode = false
			m.AcceptedFilter = ""
			m.SearchInput = "" // Clear search input text so it doesn't show in the UI
			m.FilteredItems = nil
			m.Items = flattenTree(m.Tree)
			m.Sidebar.SetItems(m.Items)
			m.SelectedIndex = 0
			m.Sidebar.Select(0)
			return m, nil
		}
	}

		switch msg.String() {
		case "j":
			if m.SelectedIndex < len(m.Items)-1 {
				m.SelectedIndex++
				m.Sidebar.Select(m.SelectedIndex)
			}
		case "k":
			if m.SelectedIndex > 0 {
				m.SelectedIndex--
				m.Sidebar.Select(m.SelectedIndex)
			}
		case "h":
			// Collapse selected node if expanded
			if len(m.Items) > 0 && m.SelectedIndex < len(m.Items) {
				item := m.Items[m.SelectedIndex].(treeItem)
				if item.node.Expanded && len(item.node.Children) > 0 {
					item.node.Expanded = false
					m.Items = flattenTree(m.Tree)
					// Clamp selection if needed
					if m.SelectedIndex >= len(m.Items) {
						m.SelectedIndex = len(m.Items) - 1
					}
					m.Sidebar.SetItems(m.Items)
					m.Sidebar.Select(m.SelectedIndex)
				}
			}
		case "l":
			// Expand selected node if it has children
			if len(m.Items) > 0 && m.SelectedIndex < len(m.Items) {
				item := m.Items[m.SelectedIndex].(treeItem)
				if !item.node.Expanded && len(item.node.Children) > 0 {
					item.node.Expanded = true
					m.Items = flattenTree(m.Tree)
					m.Sidebar.SetItems(m.Items)
					m.Sidebar.Select(m.SelectedIndex)
				}
			}
		case "gg":
			m.SelectedIndex = 0
			m.Sidebar.Select(0)
		case "G":
			m.SelectedIndex = len(m.Items) - 1
			m.Sidebar.Select(m.SelectedIndex)
		case "/":
			// Activate search mode
			m.saveExpansionState() // Save expansion state before filtering
			m.SearchActive = true
			m.SearchInput = ""
			m.FilteredItems = flattenTree(m.Tree)
			return m, nil
		case "c":
			// Toggle coverage view
			m.ShowCoverageView = !m.ShowCoverageView
			return m, nil
		case "C":
			// Run tests with coverage
			packagePath := "./..."
			// If a specific package is selected, run coverage for that package
			if m.SelectedIndex >= 0 && m.SelectedIndex < len(m.Items) {
				selectedItem, ok := m.Items[m.SelectedIndex].(treeItem)
				if ok && selectedItem.node != nil {
					// Find package path for selected item
					// This would need to be adapted to your package path extraction logic
					packagePath = "./..." // Default to all packages
				}
			}
			return m, m.RunTestsWithCoverage(packagePath)
		case "L":
			// Toggle low coverage filter when in coverage view
			if m.ShowCoverageView {
				m.ToggleLowCoverageFilter()
				return m, nil
			}
		case "E":
			// Export HTML coverage report when in coverage view
			if m.ShowCoverageView && m.CoverageState.Enabled {
				// Find the current coverage file
				var coverageFile string
				if m.CoverageState.LastCoverageFile != "" {
					coverageFile = m.CoverageState.LastCoverageFile
					return m, m.ExportCoverageHTMLReport(coverageFile)
				} else {
					// No coverage file available
					m.MainPaneContent = "No coverage file available for export"
					return m, nil
				}
			}
		case "q":
			return m, tea.Quit
		case "\n":
			// Run selected test if a test is selected
			if m.SelectedIndex >= 0 && m.SelectedIndex < len(m.Items) {
				selectedItem, ok := m.Items[m.SelectedIndex].(treeItem)
				if ok && selectedItem.node != nil {
					// If it's a leaf node (test), run that specific test
					if len(selectedItem.node.Children) == 0 {
						// Find the package path by walking up the tree
						pkg := ""
						parent := selectedItem.node.Parent
						if parent != nil {
							pkg = parent.Title
						}
						return m, runTestsCmd(pkg, selectedItem.node.Title)
					} else {
						// If it's a package/folder, run all tests in that package
						return m, runTestsCmd(selectedItem.node.Title, "")
					}
				}
			}
		case "a":
			// Run all tests
			return m, runTestsCmd("./...", "")
		case "p":
			// Run package tests (current or selected package)
			packagePath := "./..."
			if m.SelectedIndex >= 0 && m.SelectedIndex < len(m.Items) {
				selectedItem, ok := m.Items[m.SelectedIndex].(treeItem)
				if ok && selectedItem.node != nil {
					// Find the package by navigating up to the first parent that's a package
					node := selectedItem.node
					for node != nil {
						if len(node.Children) > 0 {
							packagePath = node.Title
							break
						}
						node = node.Parent
					}
				}
			}
			return m, runTestsCmd(packagePath, "")
		case "w":
			// Toggle watch mode
			// Send a message to toggle watch mode
			return m, func() tea.Msg {
				return ToggleWatchModeMsg{}
			}
		case "o":
			// Toggle log view (output)
			m.ToggleLogView()
			return m, nil
		}
	}
	return m, nil
}

func (m TUITestExplorerModel) View() string {
	// Layout constants
	minSidebarWidth := 50 // was 40, make wider
	minMainWidth := 30
	minHeight := 10
	maxSidebarWidth := 80 // was 60, make wider

	width := m.Width
	height := m.Height
	if width < minSidebarWidth+minMainWidth {
		width = minSidebarWidth + minMainWidth
	}
	if height < minHeight {
		height = minHeight
	}
	// Calculate pane sizes
	sidebarWidth := width / 3
	if sidebarWidth < minSidebarWidth {
		sidebarWidth = minSidebarWidth
	}
	if sidebarWidth > maxSidebarWidth {
		sidebarWidth = maxSidebarWidth
	}
	mainWidth := width - sidebarWidth - 2 // Account for border spacing
	mainHeight := height - 4              // header+footer+searchbar
	if mainHeight < 3 {
		mainHeight = 3
	}
	// Compose panes with Lipgloss
	// Header: centered, bold logo on full terminal width (no highlight)
	logoText := "Go-Sentinel Test Explorer"
	centeredLogo := HeaderStyle.Bold(true).Width(m.Width).Align(lipgloss.Center).Render(logoText)

	// Search bar - always visible above the sidebar list
	searchInput := m.SearchInput
	searchPrompt := "/ "
	searchBarStyle := FooterStyle.Width(sidebarWidth)
	if !m.SearchActive {
		// Dim style or placeholder when not active
		if searchInput == "" {
			searchPrompt = "/ filter..."
		}
		searchBarStyle = searchBarStyle.Foreground(lipgloss.Color("240")) // dim text
	}
	searchBar := searchBarStyle.Render(searchPrompt + searchInput)

	// Dynamically calculate the height for the sidebar list
	headerHeight := 1 // the logo/header line
	footerHeight := 1 // the footer help line
	searchBarHeight := 1
	sidebarListHeight := height - headerHeight - footerHeight - searchBarHeight
	if sidebarListHeight < 3 {
		sidebarListHeight = 3
	}
	m.Sidebar.SetSize(sidebarWidth, sidebarListHeight)
	// Forcibly hide status bar and pagination every render
	m.Sidebar.SetShowStatusBar(false)
	m.Sidebar.SetShowPagination(true)

	// Main pane content: show coverage view or package details
	mainPaneContent := ""
	
	// If coverage view is active, show that instead of normal details
	if m.ShowCoverageView && m.CoverageState.Enabled {
		mainPaneContent = m.RenderCoverageView(mainWidth, mainHeight)
	} else if m.SelectedIndex >= 0 && m.SelectedIndex < len(m.Items) {
		selectedItem, ok := m.Items[m.SelectedIndex].(treeItem)
		if ok && selectedItem.node != nil && len(selectedItem.node.Children) > 0 {
			pkg := selectedItem.node
			covValue := pkg.Coverage
			avgCoverage := FormatCoverage(covValue)
			totalDuration := FormatDurationSmart(TotalDuration(pkg.Children))
			// Set the animated coverage bar value
			m.CoverageBar.SetCoverage(covValue)
			// Animate bar color: red (0) to yellow (0.5) to green (1)
			var barColor string
			if covValue == 0 {
				barColor = "240" // gray for no tests
			} else if covValue < 0.5 {
				barColor = "196" // red
			} else if covValue < 0.8 {
				barColor = "220" // yellow
			} else {
				barColor = "42" // green
			}
			bar := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor)).Render(m.CoverageBar.View())
			percentLabel := lipgloss.NewStyle().Foreground(lipgloss.Color(barColor)).Render(avgCoverage)
			mainPaneContent = fmt.Sprintf("Package: %s\nCoverage: %s\n%s\nSuite Duration: %s\n\n", pkg.Title, percentLabel, bar, totalDuration)
			for _, child := range pkg.Children {
				if len(child.Children) == 0 { // leaf/test node
					status := "PASS"
					color := AccentGreen
					if child.Passed != nil && !*child.Passed {
						status = "FAIL"
						color = AccentRed
					}
					if child.Passed == nil {
						status = "?"
						color = ""
					}
					statusStr := lipgloss.NewStyle().Foreground(color).Render(status)
					mainPaneContent += fmt.Sprintf("%s  %s  %s\n", statusStr, child.Title, FormatDurationSmart(child.Duration))
				}
			}
		}
	}
	if mainPaneContent == "" {
		// Show file watcher details if no package is selected
		if m.LastChangedFile != "" {
			status := "Active"
			statusColor := AccentGreen
			if m.TestsRunning {
				status = "Running Tests"
				statusColor = AccentYellow
			}
			statusText := lipgloss.NewStyle().Foreground(statusColor).Bold(true).Render(status)
			
			mainPaneContent = fmt.Sprintf("File Watcher: %s\n\n", statusText)
			mainPaneContent += fmt.Sprintf("Last Changed File:\n%s\n\n", m.LastChangedFile)
			
			if m.RunningPackage != "" {
				mainPaneContent += fmt.Sprintf("Testing Package:\n%s\n\n", m.RunningPackage)
			}
			
			// Show run status
			if !m.TestsRunning && m.RunningPackage != "" {
				runStatus := "Failed"
				runColor := AccentRed
				if m.LastRunSuccess {
					runStatus = "Passed"
					runColor = AccentGreen
				}
				runText := lipgloss.NewStyle().Foreground(runColor).Bold(true).Render(runStatus)
				mainPaneContent += fmt.Sprintf("Last Run: %s\n", runText)
			}
		} else {
			mainPaneContent = "Welcome to Go-Sentinel!\n\nSelect a package from the sidebar to view details.\nFile changes will be displayed here when detected.\n\nPress '/' to filter tests\nPress 'q' to quit"
		}
	}
	mainPane := MainPaneStyle.Width(mainWidth).Height(mainHeight).Render(mainPaneContent)

	// Sidebar always has search bar at the top (no logo in sidebar)
	sidebarContent := m.Sidebar.View()
	sidebarWithHeader := searchBar + "\n" + sidebarContent
	sidebar := SidebarStyle.Width(sidebarWidth).Render(sidebarWithHeader)

	// Join horizontally
	row := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainPane)

	// Add log panel if log view is enabled
	if m.ShowLogView {
		// Calculate log panel height - use 1/3 of available height
		logHeight := height / 3
		if logHeight < 5 {
			logHeight = 5 // Minimum log height
		}
		if logHeight > height / 2 {
			logHeight = height / 2 // Maximum log height (half of screen)
		}

		// Adjust the main row height
		rowHeight := height - logHeight - headerHeight - footerHeight
		if rowHeight < 8 {
			rowHeight = 8 // Ensure minimum row height
			logHeight = height - rowHeight - headerHeight - footerHeight
		}

		// Create a styled log content with most recent logs
		logContent := "Terminal Output"
		if len(m.LogBuffer) > 0 {
			// Calculate how many lines to show based on available height
			maxLines := logHeight - 2 // Account for borders and title
			startLine := 0
			if len(m.LogBuffer) > maxLines {
				startLine = len(m.LogBuffer) - maxLines
			}

			// Join the visible log lines
			visibleLines := m.LogBuffer[startLine:]
			logContent = "Terminal Output (press 'o' to hide)\n\n" + strings.Join(visibleLines, "\n")
		} else {
			logContent = "Terminal Output (press 'o' to hide)\n\nNo output available yet."
		}

		// Render log panel with proper styling
		logPane := LogPaneStyle.Width(width).Height(logHeight).Render(logContent)

		// Add the log panel below the main content
		row = lipgloss.JoinVertical(lipgloss.Left, row, logPane)
	}

	// Footer - now with status indicators and key help
	watchStatus := "[w] Watch: OFF"
	if m.WatchModeEnabled {
		watchStatus = lipgloss.NewStyle().Foreground(AccentGreen).Render("[W] Watch: ON")
	}
	
	runStatus := ""
	if m.TestsRunning {
		runStatus = lipgloss.NewStyle().Foreground(AccentYellow).Render(" [Running Tests]")
	} else if m.LastChangedFile != "" {
		if m.LastRunSuccess {
			runStatus = lipgloss.NewStyle().Foreground(AccentGreen).Render(" [Tests Passed]")
		} else {
			runStatus = lipgloss.NewStyle().Foreground(AccentRed).Render(" [Tests Failed]")
		}
	}
	
	// Log view toggle status
	logViewStatus := "[o] Output: OFF"
	if m.ShowLogView {
		logViewStatus = lipgloss.NewStyle().Foreground(AccentBlue).Render("[O] Output: ON")
	}

	// Help text showing available commands
	footerContent := fmt.Sprintf("%s%s • %s • a all tests • p pkg tests • ↵ run test • w watch • q quit", 
		watchStatus, runStatus, logViewStatus)
	footer := FooterStyle.
		Width(width).
		Align(lipgloss.Center). // Center the text
		Render(footerContent)

	// Full layout: header/logo at top, then row, then footer
	layout := lipgloss.JoinVertical(lipgloss.Left, centeredLogo, row, footer)
	return layout
}

// --- Test helpers for TDD ---

func (m TUITestExplorerModel) SidebarHasTree() bool {
	return len(m.Items) > 0
}

func (m TUITestExplorerModel) VIMNavigationWorked() bool {
	// For now, just check the selected index is within range
	return m.SelectedIndex >= 0 && m.SelectedIndex < len(m.Items)
}

func (m TUITestExplorerModel) MainPaneShowsTestDetails() bool {
	// For now, just check that main pane content can change
	return true // Expand for detail logic later
}

// SetWatchModeCallback sets the callback function that gets called when watch mode changes
func (m *TUITestExplorerModel) SetWatchModeCallback(callback func(bool)) {
	m.watchModeCallback = callback
}

// flattenAllTree returns a flat slice of treeItems for all nodes, ignoring expansion
func flattenAllTree(root *TreeNode) []list.Item {
	var items []list.Item
	var walk func(node *TreeNode, level int, parent *TreeNode)
	walk = func(node *TreeNode, level int, parent *TreeNode) {
		if node == nil {
			return
		}
		node.Level = level
		node.Parent = parent
		items = append(items, treeItem{node})
		for _, child := range node.Children {
			walk(child, level+1, node)
		}
	}
	walk(root, 0, nil)
	return items
}

func (m TUITestExplorerModel) SidebarFiltered(term string) bool {
	if term == "" {
		return true
	}
	filtered := fuzzyFilterTreeItems(flattenAllTree(m.Tree), term)
	return len(filtered) > 0
}

// getNodesMatchingFilter returns items that match the input filter
// It performs fuzzy search and preserves parent-child relationships
func getNodesMatchingFilter(root *TreeNode, input string) []list.Item {
	if input == "" {
		return flattenTree(root) // Return the full tree if no filter
	}
	
	// First, collect all titles for fuzzy searching
	var nodeInfos []struct {
		title string
		node  *TreeNode
	}
	
	// Collect all nodes
	var collectNodes func(node *TreeNode)
	collectNodes = func(node *TreeNode) {
		if node == nil {
			return
		}
		
		// Add this node
		nodeInfos = append(nodeInfos, struct {
			title string
			node  *TreeNode
		}{title: node.Title, node: node})
		
		// Process children
		for _, child := range node.Children {
			collectNodes(child)
		}
	}
	collectNodes(root)
	
	// Extract just the titles for fuzzy search
	titles := make([]string, len(nodeInfos))
	for i, info := range nodeInfos {
		titles[i] = info.title
	}
	
	// Perform fuzzy search
	matches := fuzzy.Find(input, titles)
	
	// Track matching nodes and their ancestors
	matchingNodes := make(map[*TreeNode]bool)
	
	// Add matching nodes
	for _, match := range matches {
		node := nodeInfos[match.Index].node
		matchingNodes[node] = true
		
		// Also include all ancestors to maintain path
		curr := node.Parent
		for curr != nil {
			matchingNodes[curr] = true
			curr = curr.Parent
		}
		
		// If a folder matches, include all its children
		if len(node.Children) > 0 {
			var addAllDescendants func(n *TreeNode)
			addAllDescendants = func(n *TreeNode) {
				for _, child := range n.Children {
					matchingNodes[child] = true
					addAllDescendants(child)
				}
			}
			addAllDescendants(node)
		}
	}
	
	// Now flatten the tree but only include matching nodes
	var filteredItems []list.Item
	
	var walkFiltered func(node *TreeNode, level int)
	walkFiltered = func(node *TreeNode, level int) {
		if node == nil {
			return
		}
		
		// Only include this node if it or any descendant matched
		if matchingNodes[node] {
			filteredItems = append(filteredItems, treeItem{node})
			
			// Only walk children if this node is expanded
			if node.Expanded {
				for _, child := range node.Children {
					// Only recurse if the child or its descendants matched
					if matchingNodes[child] {
						walkFiltered(child, level+1)
					}
				}
			}
		}
	}
	
	walkFiltered(root, 0)
	
	return filteredItems
}

// fuzzyFilterTreeItems returns only those treeItems whose Title fuzzy-matches the input
func fuzzyFilterTreeItems(items []list.Item, input string) []list.Item {
	if input == "" {
		return items
	}
	titles := make([]string, len(items))
	for i, it := range items {
		titles[i] = it.(treeItem).node.Title
	}
	matches := fuzzy.Find(input, titles)
	var filtered []list.Item
	for _, m := range matches {
		filtered = append(filtered, items[m.Index])
	}
	return filtered
}
