package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// PanelOptions configures the panel's appearance and layout.
type PanelOptions struct {
	Title       string
	Width       int // 0 = auto
	Padding     int // spaces around content
	Margin      int // lines above/below
	Border      bool
	BorderStyle lipgloss.Border // e.g. lipgloss.NormalBorder(), lipgloss.RoundedBorder()
	Style       lipgloss.Style  // full style control
	BorderColor lipgloss.Color  // border color
	TitleStyle  lipgloss.Style  // style for the title
}

// Panel is a reusable container for TUI content.
type Panel struct {
	Content  []string // lines of content
	Options  PanelOptions
}

// Render returns the panel as a slice of strings, each representing a line.
func isZeroStyle(s lipgloss.Style) bool {
	return s.GetWidth() == 0
}

func isZeroBorder(b lipgloss.Border) bool {
	return b.Top == "" && b.Bottom == "" && b.Left == "" && b.Right == "" && b.TopLeft == "" && b.TopRight == "" && b.BottomLeft == "" && b.BottomRight == ""
}

func isZeroColor(c lipgloss.Color) bool {
	r, g, b, a := c.RGBA()
	return r == 0 && g == 0 && b == 0 && a == 0
}

func (p Panel) Render() []string {
	opt := p.Options
	content := strings.Join(p.Content, "\n")

	// Always start with a base style
	style := opt.Style
	if isZeroStyle(style) {
		style = lipgloss.NewStyle()
	}
	if opt.Width > 0 {
		style = style.Width(opt.Width)
	}
	if opt.Padding > 0 {
		style = style.Padding(opt.Padding)
	}
	if opt.Margin > 0 {
		style = style.Margin(opt.Margin)
	}
	if opt.Border {
		border := opt.BorderStyle
		if isZeroBorder(border) {
			border = lipgloss.NormalBorder()
		}
		style = style.BorderStyle(border)
		if !isZeroColor(opt.BorderColor) {
			style = style.BorderForeground(opt.BorderColor)
		}
	}
	if opt.Title != "" {
		title := opt.Title
		if !isZeroStyle(opt.TitleStyle) {
			title = opt.TitleStyle.Render(opt.Title)
		}
		content = title + "\n" + strings.Join(p.Content, "\n")
	}
	panel := style.Render(content)
	return strings.Split(panel, "\n")
}
