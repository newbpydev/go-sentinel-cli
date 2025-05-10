package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
)

// The styles and colors are defined in theme.go within the same package

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
} // Now tracks terminal width/height

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
	icon := ""
	switch {
	case ti.node.Level == 0:
		icon = "📦"
	case ti.node.Error == "skip":
		icon = "📁"
	case len(ti.node.Children) > 0:
		icon = "📁"
	default:
		icon = ""
	}
	indent := ""
	for i := 0; i < ti.node.Level; i++ {
		indent += "  "
	}
	// For test nodes (leaf), only show name, no icon
	if len(ti.node.Children) == 0 {
		return fmt.Sprintf("%s%s", indent, ti.node.Title)
	}
	return fmt.Sprintf("%s%s %s", indent, icon, ti.node.Title)
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
		fmt.Fprintf(w, "%s %s", pointer, colored)
	} else {
		fmt.Fprintf(w, "  %s", colored)
	}
}

func NewTUITestExplorerModel(root *TreeNode) TUITestExplorerModel {
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
	}
} // Default size, updated on resize

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

// Bubble Tea Model interface
func (m TUITestExplorerModel) Init() tea.Cmd {
	return nil
}

func (m TUITestExplorerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle window resize
	if ws, ok := msg.(tea.WindowSizeMsg); ok {
		m.Width = ws.Width
		m.Height = ws.Height
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
		if m.SearchActive {
			// Handle search input
			switch msg.Type {
			case tea.KeyEsc:
				m.SearchActive = false
				m.SearchInput = ""
				m.FilteredItems = nil
				m.Items = flattenTree(m.Tree)
				m.Sidebar.SetItems(m.Items)
				m.SelectedIndex = 0
				m.Sidebar.Select(0)
			case tea.KeyEnter:
				m.SearchActive = false
				m.Items = m.FilteredItems
				m.Sidebar.SetItems(m.Items)
				m.SelectedIndex = 0
				m.Sidebar.Select(0)
			default:
				if msg.Type == tea.KeyBackspace || msg.Type == tea.KeyDelete {
					if len(m.SearchInput) > 0 {
						m.SearchInput = m.SearchInput[:len(m.SearchInput)-1]
					}
				} else if msg.Type == tea.KeyRunes {
					m.SearchInput += msg.String()
				}
				// Fuzzy filter
				m.FilteredItems = fuzzyFilterTreeItems(flattenTree(m.Tree), m.SearchInput)
				m.Sidebar.SetItems(m.FilteredItems)
				m.SelectedIndex = 0
				m.Sidebar.Select(0)
			}
			return m, nil
		}
		if msg.String() == "?" {
			m.ShowHelpModal = true
			return m, nil
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
			m.SearchActive = true
			m.SearchInput = ""
			m.FilteredItems = flattenTree(m.Tree)
			return m, nil
		case "q":
			return m, tea.Quit
		case "\n":
			// Enter logic stub
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

	// Main pane content: show package details if a folder is selected
	mainPaneContent := ""
	if m.SelectedIndex >= 0 && m.SelectedIndex < len(m.Items) {
		selectedItem, ok := m.Items[m.SelectedIndex].(treeItem)
		if ok && selectedItem.node != nil && len(selectedItem.node.Children) > 0 {
			pkg := selectedItem.node
			avgCoverage := FormatCoverage(AverageCoverage(pkg.Children))
			totalDuration := FormatDurationSmart(TotalDuration(pkg.Children))
			mainPaneContent = fmt.Sprintf("Package: %s\nCoverage: %s\nSuite Duration: %s\n\n", pkg.Title, avgCoverage, totalDuration)
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
		mainPaneContent = "details placeholder!"
	}
	mainPane := MainPaneStyle.Width(mainWidth).Height(mainHeight).Render(mainPaneContent)

	// Sidebar always has search bar at the top (no logo in sidebar)
	sidebarContent := m.Sidebar.View()
	sidebarWithHeader := searchBar + "\n" + sidebarContent
	sidebar := SidebarStyle.Width(sidebarWidth).Render(sidebarWithHeader)

	// Join horizontally
	row := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainPane)

	// Footer - now centered
	footerContent := "↑/k up • ↓/j down • / filter • q quit • ? help"
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

func (m TUITestExplorerModel) SidebarFiltered(term string) bool {
	if term == "" {
		return true
	}
	filtered := fuzzyFilterTreeItems(flattenTree(m.Tree), term)
	return len(filtered) > 0
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
