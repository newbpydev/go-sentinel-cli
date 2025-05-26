// Package display provides progress rendering for live test execution updates
package display

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/internal/ui/icons"
)

// ProgressRendererImpl implements ProgressRenderer interface for real-time progress display
type ProgressRendererImpl struct {
	config    *Config
	formatter *colors.ColorFormatter
	icons     icons.IconProvider
	spinner   *SpinnerConfig
	writer    io.Writer

	// Progress state
	mutex      sync.RWMutex
	current    int
	total      int
	startTime  time.Time
	lastUpdate time.Time
	status     string
	isActive   bool
	isFinished bool

	// Display configuration
	width          int
	showSpinner    bool
	showPercent    bool
	showEta        bool
	showThroughput bool
	updateInterval time.Duration

	// Animation state
	spinnerFrame int
	animationCh  chan struct{}
	stopCh       chan struct{}
}

// NewProgressRenderer creates a new progress renderer with configuration
func NewProgressRenderer(config *Config) *ProgressRendererImpl {
	formatter := colors.NewAutoColorFormatter()
	detector := colors.NewTerminalDetector()

	// Use Unicode provider if Unicode is supported, otherwise ASCII
	var iconProvider icons.IconProvider
	if detector.SupportsUnicode() {
		iconProvider = icons.NewUnicodeProvider()
	} else {
		iconProvider = icons.NewASCIIProvider()
	}

	return &ProgressRendererImpl{
		config:         config,
		formatter:      formatter,
		icons:          iconProvider,
		writer:         config.Output,
		width:          40, // Default progress bar width
		showSpinner:    true,
		showPercent:    true,
		showEta:        true,
		showThroughput: false,
		updateInterval: 100 * time.Millisecond,
		animationCh:    make(chan struct{}, 1),
		stopCh:         make(chan struct{}),
	}
}

// StartProgress begins progress rendering with total expected items
func (p *ProgressRendererImpl) StartProgress(ctx context.Context, total int) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.total = total
	p.current = 0
	p.startTime = time.Now()
	p.lastUpdate = p.startTime
	p.isActive = true
	p.isFinished = false
	p.status = "Starting..."

	// Start spinner animation if enabled
	if p.showSpinner {
		go p.animateSpinner(ctx)
	}

	return p.render()
}

// UpdateProgress updates progress with current status
func (p *ProgressRendererImpl) UpdateProgress(current int, status string) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isActive || p.isFinished {
		return nil
	}

	p.current = current
	p.status = status
	p.lastUpdate = time.Now()

	// Trigger animation update
	select {
	case p.animationCh <- struct{}{}:
	default:
	}

	return p.render()
}

// FinishProgress completes progress rendering
func (p *ProgressRendererImpl) FinishProgress() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.isFinished {
		return nil // Already finished
	}

	p.isFinished = true
	p.isActive = false

	// Stop spinner animation (only if not already closed)
	select {
	case <-p.stopCh:
		// Channel already closed
	default:
		close(p.stopCh)
	}

	// Final render
	return p.renderComplete()
}

// SetSpinner configures the progress spinner
func (p *ProgressRendererImpl) SetSpinner(spinner *SpinnerConfig) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.spinner = spinner
	return nil
}

// render displays the current progress state
func (p *ProgressRendererImpl) render() error {
	if !p.isActive {
		return nil
	}

	var output strings.Builder

	// Clear current line and move cursor to beginning
	output.WriteString("\r\033[K")

	// Render spinner if enabled
	if p.showSpinner {
		frame := p.getSpinnerFrame()
		output.WriteString(p.formatter.Cyan(frame))
		output.WriteString(" ")
	}

	// Render progress bar
	progressBar := p.renderProgressBar()
	output.WriteString(progressBar)

	// Render percentage if enabled
	if p.showPercent {
		percentage := p.calculatePercentage()
		output.WriteString(" ")
		output.WriteString(p.formatter.Bold(fmt.Sprintf("%.1f%%", percentage)))
	}

	// Render current/total
	output.WriteString(" ")
	output.WriteString(p.formatter.Dim(fmt.Sprintf("(%d/%d)", p.current, p.total)))

	// Render ETA if enabled
	if p.showEta && p.current > 0 {
		eta := p.calculateETA()
		if eta > 0 {
			output.WriteString(" ETA: ")
			output.WriteString(p.formatter.Cyan(p.formatDuration(eta)))
		}
	}

	// Render status
	if p.status != "" {
		output.WriteString(" • ")
		output.WriteString(p.formatter.Gray(p.status))
	}

	// Write to output
	_, err := p.writer.Write([]byte(output.String()))
	return err
}

// renderComplete displays the final completion state
func (p *ProgressRendererImpl) renderComplete() error {
	var output strings.Builder

	// Clear current line and move cursor to beginning
	output.WriteString("\r\033[K")

	// Success indicator
	checkIcon, _ := p.icons.GetIcon("CheckMark")
	output.WriteString(p.formatter.Green(checkIcon))
	output.WriteString(" ")

	// Completion message
	output.WriteString(p.formatter.Green("Complete"))

	// Final stats
	duration := time.Since(p.startTime)
	output.WriteString(" ")
	output.WriteString(p.formatter.Dim(fmt.Sprintf("(%d items in %s)", p.total, p.formatDuration(duration))))

	// Final status
	if p.status != "" {
		output.WriteString(" • ")
		output.WriteString(p.formatter.Gray(p.status))
	}

	output.WriteString("\n")

	// Write to output
	_, err := p.writer.Write([]byte(output.String()))
	return err
}

// renderProgressBar creates the visual progress bar
func (p *ProgressRendererImpl) renderProgressBar() string {
	if p.total == 0 {
		return p.formatter.Dim("[" + strings.Repeat(" ", p.width) + "]")
	}

	percentage := p.calculatePercentage()
	filled := int(percentage / 100.0 * float64(p.width))

	var bar strings.Builder
	bar.WriteString("[")

	// Filled portion (success color)
	if filled > 0 {
		bar.WriteString(p.formatter.Green(strings.Repeat("=", filled)))
	}

	// Current position indicator
	if filled < p.width && p.current < p.total {
		bar.WriteString(p.formatter.Yellow(">"))
		filled++
	}

	// Empty portion
	remaining := p.width - filled
	if remaining > 0 {
		bar.WriteString(p.formatter.Dim(strings.Repeat("-", remaining)))
	}

	bar.WriteString("]")
	return bar.String()
}

// animateSpinner handles spinner animation in a separate goroutine
func (p *ProgressRendererImpl) animateSpinner(ctx context.Context) {
	ticker := time.NewTicker(p.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopCh:
			return
		case <-ticker.C:
			p.mutex.Lock()
			if p.isActive && !p.isFinished {
				p.spinnerFrame++
				p.mutex.Unlock()
				// Trigger re-render
				select {
				case p.animationCh <- struct{}{}:
				default:
				}
			} else {
				p.mutex.Unlock()
			}
		case <-p.animationCh:
			// Re-render triggered by progress update
			if p.isActive && !p.isFinished {
				_ = p.render()
			}
		}
	}
}

// getSpinnerFrame returns the current spinner frame
func (p *ProgressRendererImpl) getSpinnerFrame() string {
	var frames []string
	if p.spinner != nil && len(p.spinner.Frames) > 0 {
		frames = p.spinner.Frames
	} else {
		// Default spinner frames
		frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	}

	frameIndex := p.spinnerFrame % len(frames)
	return frames[frameIndex]
}

// calculatePercentage returns the current completion percentage
func (p *ProgressRendererImpl) calculatePercentage() float64 {
	if p.total == 0 {
		return 0.0
	}
	return float64(p.current) / float64(p.total) * 100.0
}

// calculateETA estimates time remaining based on current progress
func (p *ProgressRendererImpl) calculateETA() time.Duration {
	if p.current == 0 || p.total == 0 {
		return 0
	}

	elapsed := time.Since(p.startTime)
	rate := float64(p.current) / elapsed.Seconds()

	if rate == 0 {
		return 0
	}

	remaining := p.total - p.current
	etaSeconds := float64(remaining) / rate

	return time.Duration(etaSeconds) * time.Second
}

// formatDuration formats a duration for display
func (p *ProgressRendererImpl) formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%ds", minutes, seconds)
}

// SetWidth configures the progress bar width
func (p *ProgressRendererImpl) SetWidth(width int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.width = width
}

// SetShowSpinner enables/disables spinner animation
func (p *ProgressRendererImpl) SetShowSpinner(show bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.showSpinner = show
}

// SetShowPercent enables/disables percentage display
func (p *ProgressRendererImpl) SetShowPercent(show bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.showPercent = show
}

// SetShowETA enables/disables ETA display
func (p *ProgressRendererImpl) SetShowETA(show bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.showEta = show
}

// GetProgress returns current progress information
func (p *ProgressRendererImpl) GetProgress() (current, total int, percentage float64) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.current, p.total, p.calculatePercentage()
}

// IsActive returns whether progress rendering is currently active
func (p *ProgressRendererImpl) IsActive() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.isActive
}

// IsFinished returns whether progress rendering has finished
func (p *ProgressRendererImpl) IsFinished() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.isFinished
}

// Ensure ProgressRendererImpl implements ProgressRenderer interface
var _ ProgressRenderer = (*ProgressRendererImpl)(nil)
