// Package display provides status bar rendering for test execution headers
package display

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
	"github.com/newbpydev/go-sentinel/internal/ui/icons"
)

// StatusBarRenderer handles rendering of the header status bar
type StatusBarRenderer struct {
	config    *Config
	formatter *colors.ColorFormatter
	icons     icons.IconProvider

	// Status information
	mutex          sync.RWMutex
	testsPassed    int
	testsFailed    int
	testsSkipped   int
	testsTotal     int
	startTime      time.Time
	duration       time.Duration
	memoryUsage    string
	watchMode      bool
	currentPackage string
	status         StatusBarStatus

	// Display configuration
	width        int
	showTiming   bool
	showMemory   bool
	showProgress bool
	showWatch    bool
}

// StatusBarStatus represents the current execution status
type StatusBarStatus int

const (
	StatusBarIdle StatusBarStatus = iota
	StatusBarRunning
	StatusBarPassing
	StatusBarFailing
	StatusBarComplete
	StatusBarError
)

// StatusBarInfo contains status bar display information
type StatusBarInfo struct {
	TestsPassed    int
	TestsFailed    int
	TestsSkipped   int
	TestsTotal     int
	Duration       time.Duration
	MemoryUsage    string
	WatchMode      bool
	CurrentPackage string
	Status         StatusBarStatus
}

// NewStatusBarRenderer creates a new status bar renderer
func NewStatusBarRenderer(config *Config) *StatusBarRenderer {
	formatter := colors.NewAutoColorFormatter()
	detector := colors.NewTerminalDetector()

	// Use Unicode provider if Unicode is supported, otherwise ASCII
	var iconProvider icons.IconProvider
	if detector.SupportsUnicode() {
		iconProvider = icons.NewUnicodeProvider()
	} else {
		iconProvider = icons.NewASCIIProvider()
	}

	return &StatusBarRenderer{
		config:       config,
		formatter:    formatter,
		icons:        iconProvider,
		width:        80, // Default width
		showTiming:   true,
		showMemory:   true,
		showProgress: true,
		showWatch:    true,
		status:       StatusBarIdle,
	}
}

// Update updates the status bar with new information
func (s *StatusBarRenderer) Update(info *StatusBarInfo) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.testsPassed = info.TestsPassed
	s.testsFailed = info.TestsFailed
	s.testsSkipped = info.TestsSkipped
	s.testsTotal = info.TestsTotal
	s.duration = info.Duration
	s.memoryUsage = info.MemoryUsage
	s.watchMode = info.WatchMode
	s.currentPackage = info.CurrentPackage
	s.status = info.Status
}

// Render generates the status bar display
func (s *StatusBarRenderer) Render() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var parts []string

	// Status icon and text
	statusPart := s.renderStatus()
	if statusPart != "" {
		parts = append(parts, statusPart)
	}

	// Progress information
	if s.showProgress && s.testsTotal > 0 {
		progressPart := s.renderProgress()
		parts = append(parts, progressPart)
	}

	// Timing information
	if s.showTiming && s.duration > 0 {
		timingPart := s.renderTiming()
		parts = append(parts, timingPart)
	}

	// Memory usage
	if s.showMemory && s.memoryUsage != "" {
		memoryPart := s.renderMemory()
		parts = append(parts, memoryPart)
	}

	// Watch mode indicator
	if s.showWatch && s.watchMode {
		watchPart := s.renderWatch()
		parts = append(parts, watchPart)
	}

	// Current package
	if s.currentPackage != "" {
		packagePart := s.renderCurrentPackage()
		parts = append(parts, packagePart)
	}

	// Join parts with separator
	separator := s.formatter.Dim(" • ")
	result := strings.Join(parts, separator)

	// Add padding and borders if needed
	return s.formatWithBorder(result)
}

// renderStatus creates the status icon and text
func (s *StatusBarRenderer) renderStatus() string {
	var icon, text, color string

	switch s.status {
	case StatusBarIdle:
		icon = "⚪"
		text = "Ready"
		color = "gray"
	case StatusBarRunning:
		icon = "🔄"
		text = "Running"
		color = "cyan"
	case StatusBarPassing:
		checkIcon, _ := s.icons.GetIcon("CheckMark")
		icon = checkIcon
		text = "Passing"
		color = "green"
	case StatusBarFailing:
		crossIcon, _ := s.icons.GetIcon("Cross")
		icon = crossIcon
		text = "Failing"
		color = "red"
	case StatusBarComplete:
		checkIcon, _ := s.icons.GetIcon("CheckMark")
		icon = checkIcon
		text = "Complete"
		color = "green"
	case StatusBarError:
		crossIcon, _ := s.icons.GetIcon("Cross")
		icon = crossIcon
		text = "Error"
		color = "red"
	default:
		icon = "❓"
		text = "Unknown"
		color = "gray"
	}

	switch color {
	case "green":
		return s.formatter.Green(icon + " " + text)
	case "red":
		return s.formatter.Red(icon + " " + text)
	case "cyan":
		return s.formatter.Cyan(icon + " " + text)
	case "gray":
		return s.formatter.Gray(icon + " " + text)
	default:
		return icon + " " + text
	}
}

// renderProgress creates the progress display
func (s *StatusBarRenderer) renderProgress() string {
	if s.testsTotal == 0 {
		return ""
	}

	var parts []string

	// Passed tests
	if s.testsPassed > 0 {
		checkIcon, _ := s.icons.GetIcon("CheckMark")
		parts = append(parts, s.formatter.Green(fmt.Sprintf("%s %d", checkIcon, s.testsPassed)))
	}

	// Failed tests
	if s.testsFailed > 0 {
		crossIcon, _ := s.icons.GetIcon("Cross")
		parts = append(parts, s.formatter.Red(fmt.Sprintf("%s %d", crossIcon, s.testsFailed)))
	}

	// Skipped tests
	if s.testsSkipped > 0 {
		skipIcon := "⏭"
		parts = append(parts, s.formatter.Yellow(fmt.Sprintf("%s %d", skipIcon, s.testsSkipped)))
	}

	// Total
	progressText := strings.Join(parts, " ")
	if len(parts) > 0 {
		progressText += s.formatter.Dim(fmt.Sprintf(" (%d total)", s.testsTotal))
	} else {
		progressText = s.formatter.Dim(fmt.Sprintf("0/%d", s.testsTotal))
	}

	return progressText
}

// renderTiming creates the timing display
func (s *StatusBarRenderer) renderTiming() string {
	timerIcon, _ := s.icons.GetIcon("Timer")
	if timerIcon == "" {
		timerIcon = "⏱"
	}

	formattedDuration := s.formatDuration(s.duration)
	return s.formatter.Cyan(timerIcon) + " " + s.formatter.Bold(formattedDuration)
}

// renderMemory creates the memory usage display
func (s *StatusBarRenderer) renderMemory() string {
	memoryIcon := "💾"
	return s.formatter.Blue(memoryIcon) + " " + s.formatter.Dim(s.memoryUsage)
}

// renderWatch creates the watch mode indicator
func (s *StatusBarRenderer) renderWatch() string {
	watchIcon := "👁"
	return s.formatter.Magenta(watchIcon) + " " + s.formatter.Bold("WATCH")
}

// renderCurrentPackage creates the current package display
func (s *StatusBarRenderer) renderCurrentPackage() string {
	packageIcon := "📦"
	maxLength := 30 // Truncate long package names

	packageName := s.currentPackage
	if len(packageName) > maxLength {
		packageName = "..." + packageName[len(packageName)-maxLength+3:]
	}

	return s.formatter.Blue(packageIcon) + " " + s.formatter.Dim(packageName)
}

// formatWithBorder adds padding and optional border to the status bar
func (s *StatusBarRenderer) formatWithBorder(content string) string {
	if content == "" {
		return ""
	}

	// Calculate content length (without ANSI escape sequences)
	contentLength := s.calculateDisplayLength(content)

	// Add padding if needed
	if contentLength < s.width {
		padding := s.width - contentLength
		leftPad := padding / 2
		rightPad := padding - leftPad

		content = strings.Repeat(" ", leftPad) + content + strings.Repeat(" ", rightPad)
	}

	// Add top and bottom borders
	border := s.formatter.Dim(strings.Repeat("─", s.width))

	return fmt.Sprintf("%s\n%s\n%s", border, content, border)
}

// formatDuration formats a duration for display
func (s *StatusBarRenderer) formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	minutes := int(d.Minutes())
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%ds", minutes, seconds)
}

// calculateDisplayLength calculates the display length of text with ANSI codes
func (s *StatusBarRenderer) calculateDisplayLength(text string) int {
	// Simple implementation - in practice, you'd want to strip ANSI codes
	// For now, estimate based on visible characters
	runes := []rune(text)
	length := 0
	inEscape := false

	for _, r := range runes {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		length++
	}

	return length
}

// SetWidth configures the status bar width
func (s *StatusBarRenderer) SetWidth(width int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.width = width
}

// SetShowTiming enables/disables timing display
func (s *StatusBarRenderer) SetShowTiming(show bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.showTiming = show
}

// SetShowMemory enables/disables memory display
func (s *StatusBarRenderer) SetShowMemory(show bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.showMemory = show
}

// SetShowProgress enables/disables progress display
func (s *StatusBarRenderer) SetShowProgress(show bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.showProgress = show
}

// SetShowWatch enables/disables watch mode display
func (s *StatusBarRenderer) SetShowWatch(show bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.showWatch = show
}

// GetCurrentStatus returns the current status
func (s *StatusBarRenderer) GetCurrentStatus() StatusBarStatus {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.status
}

// Clear clears the status bar
func (s *StatusBarRenderer) Clear() string {
	return "\033[2K\r" // Clear line and return to start
}
