package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// Theme constants for colors, borders, padding
var (
	// Inspired by dark dashboard UI
	BaseBackground = lipgloss.Color("16")       // Very dark blue-black (#121328)
	DarkerBackground = lipgloss.Color("16")    // Even darker blue-black
	PanelBackground = lipgloss.Color("233")     // Dark panel (#191c2c)
	
	HeaderStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).         // White
		Background(BaseBackground).              // Very dark blue-black
		Padding(0, 2)

	FooterStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).         // White
		Background(BaseBackground).              // Very dark blue-black
		Padding(0, 2)

	SidebarStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).        // Very light gray text
		Background(BaseBackground).              // Very dark blue-black 
		Padding(0, 2)

	MainPaneStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).         // White
		Background(PanelBackground).             // Dark panel color
		Padding(1, 2).                           
		Border(lipgloss.RoundedBorder()).         // Rounded borders for panels
		BorderForeground(lipgloss.Color("237"))  // Dark border

	// Accent colors for various elements
	AccentRed = lipgloss.Color("196")           // Red (#E53935)
	AccentPink = lipgloss.Color("198")          // Pink (#EC407A)
	AccentGreen = lipgloss.Color("42")          // Green (#43A047)
	AccentBlue = lipgloss.Color("33")           // Blue (#1E88E5)
	AccentYellow = lipgloss.Color("220")        // Gold/Yellow for progress
	
	// Selection styles
	SidebarSelectedItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).         // White text
		Background(lipgloss.Color("238")).        // Darker gray background
		Padding(0, 0)

	SelectedItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).         // White text
		Background(lipgloss.Color("90"))          // Highlighted background

	// Log panel style
	LogPaneStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("252")).		 // Light gray text
		Background(DarkerBackground).			 // Dark background
		Padding(1, 2).						 // Same padding as main pane
		Border(lipgloss.RoundedBorder()).		 // Rounded borders
		BorderForeground(lipgloss.Color("237"))  // Dark border

	BorderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).          // Rounded borders
		BorderForeground(lipgloss.Color("237"))   // Dark subtle border
)
