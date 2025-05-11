package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/progress"
)

// AnimatedCoverageBar wraps a Bubbles progress.Model for animated coverage bars
// in the details panel.
type AnimatedCoverageBar struct {
	Progress progress.Model
	Value    float64 // 0.0 - 1.0
}

func NewAnimatedCoverageBar() AnimatedCoverageBar {
	p := progress.New(
		progress.WithGradient(
			"#ff2b2b", // red
			"#00ff00", // green
		),
	)
	return AnimatedCoverageBar{
		Progress: p,
		Value:    0,
	}
}

func (b *AnimatedCoverageBar) SetCoverage(v float64) tea.Cmd {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	b.Value = v
	return b.Progress.SetPercent(v)
}

func (b *AnimatedCoverageBar) View() string {
	return b.Progress.ViewAs(b.Value)
}
