// Package display provides live terminal updating capabilities for real-time UI
package display

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/newbpydev/go-sentinel/internal/ui/colors"
)

// LiveUpdater manages real-time terminal updates with cursor positioning and screen management
type LiveUpdater struct {
	config    *Config
	formatter *colors.ColorFormatter
	writer    io.Writer

	// Update coordination
	mutex          sync.RWMutex
	updateChannel  chan *UpdateRequest
	stopChannel    chan struct{}
	isActive       bool
	updateInterval time.Duration

	// Terminal state management
	terminalWidth   int
	terminalHeight  int
	currentLines    int
	maxLines        int
	cursorSaved     bool
	alternateScreen bool

	// Component management
	components     map[string]LiveComponent
	componentOrder []string
	lastUpdate     time.Time

	// Buffering
	updateBuffer strings.Builder
	lastContent  map[string]string
}

// LiveComponent represents a component that can be updated in real-time
type LiveComponent interface {
	// Render returns the current display content
	Render() string

	// GetHeight returns the number of lines this component occupies
	GetHeight() int

	// GetPriority returns the display priority (higher = more important)
	GetPriority() int

	// ShouldUpdate indicates if the component needs updating
	ShouldUpdate() bool

	// GetID returns the unique component identifier
	GetID() string
}

// UpdateRequest represents a request to update the live display
type UpdateRequest struct {
	ComponentID string
	Content     string
	ForceUpdate bool
	Timestamp   time.Time
}

// LiveUpdaterConfig configures the live updater behavior
type LiveUpdaterConfig struct {
	// UpdateInterval is how often to refresh the display
	UpdateInterval time.Duration

	// MaxLines is the maximum number of lines to use
	MaxLines int

	// UseAlternateScreen enables alternate screen buffer
	UseAlternateScreen bool

	// ClearOnStart clears the screen when starting
	ClearOnStart bool

	// RestoreOnStop restores the original state when stopping
	RestoreOnStop bool
}

// NewLiveUpdater creates a new live updater with configuration
func NewLiveUpdater(config *Config, updaterConfig *LiveUpdaterConfig) *LiveUpdater {
	formatter := colors.NewAutoColorFormatter()

	if updaterConfig == nil {
		updaterConfig = &LiveUpdaterConfig{
			UpdateInterval:     100 * time.Millisecond,
			MaxLines:           20,
			UseAlternateScreen: false,
			ClearOnStart:       true,
			RestoreOnStop:      true,
		}
	}

	return &LiveUpdater{
		config:          config,
		formatter:       formatter,
		writer:          config.Output,
		updateChannel:   make(chan *UpdateRequest, 100),
		stopChannel:     make(chan struct{}),
		updateInterval:  updaterConfig.UpdateInterval,
		maxLines:        updaterConfig.MaxLines,
		alternateScreen: updaterConfig.UseAlternateScreen,
		components:      make(map[string]LiveComponent),
		componentOrder:  make([]string, 0),
		lastContent:     make(map[string]string),
		terminalWidth:   80, // Default, will be detected
		terminalHeight:  24, // Default, will be detected
	}
}

// Start begins live updating
func (l *LiveUpdater) Start(ctx context.Context) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if l.isActive {
		return fmt.Errorf("live updater is already active")
	}

	l.isActive = true

	// Initialize terminal
	if err := l.initializeTerminal(); err != nil {
		return fmt.Errorf("failed to initialize terminal: %w", err)
	}

	// Start update loop
	go l.updateLoop(ctx)

	return nil
}

// Stop stops live updating and restores terminal state
func (l *LiveUpdater) Stop() error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if !l.isActive {
		return nil
	}

	l.isActive = false

	// Signal stop
	close(l.stopChannel)

	// Restore terminal
	return l.restoreTerminal()
}

// RegisterComponent registers a component for live updates
func (l *LiveUpdater) RegisterComponent(component LiveComponent) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	componentID := component.GetID()
	if _, exists := l.components[componentID]; exists {
		return fmt.Errorf("component %s already registered", componentID)
	}

	l.components[componentID] = component
	l.componentOrder = append(l.componentOrder, componentID)
	l.lastContent[componentID] = ""

	// Sort by priority
	l.sortComponentsByPriority()

	return nil
}

// UnregisterComponent removes a component from live updates
func (l *LiveUpdater) UnregisterComponent(componentID string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if _, exists := l.components[componentID]; !exists {
		return fmt.Errorf("component %s not found", componentID)
	}

	delete(l.components, componentID)
	delete(l.lastContent, componentID)

	// Remove from order
	for i, id := range l.componentOrder {
		if id == componentID {
			l.componentOrder = append(l.componentOrder[:i], l.componentOrder[i+1:]...)
			break
		}
	}

	return nil
}

// RequestUpdate requests an update for a specific component
func (l *LiveUpdater) RequestUpdate(componentID string, forceUpdate bool) {
	if !l.isActive {
		return
	}

	request := &UpdateRequest{
		ComponentID: componentID,
		ForceUpdate: forceUpdate,
		Timestamp:   time.Now(),
	}

	select {
	case l.updateChannel <- request:
	default:
		// Channel full, skip this update
	}
}

// updateLoop handles the main update loop
func (l *LiveUpdater) updateLoop(ctx context.Context) {
	ticker := time.NewTicker(l.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-l.stopChannel:
			return
		case <-ticker.C:
			l.performUpdate(false)
		case request := <-l.updateChannel:
			l.handleUpdateRequest(request)
		}
	}
}

// performUpdate performs a display update
func (l *LiveUpdater) performUpdate(force bool) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if !l.isActive {
		return
	}

	var needsUpdate bool
	var totalHeight int

	// Check if any component needs updating
	for _, componentID := range l.componentOrder {
		component := l.components[componentID]
		if component == nil {
			continue
		}

		height := component.GetHeight()
		if totalHeight+height > l.maxLines {
			break // Skip components that don't fit
		}

		if force || component.ShouldUpdate() {
			needsUpdate = true
		}

		totalHeight += height
	}

	if !needsUpdate && !force {
		return
	}

	// Generate display content
	content := l.generateDisplayContent(totalHeight)

	// Update display
	if err := l.updateDisplay(content); err != nil {
		// Handle error silently to avoid disrupting display
		return
	}

	l.lastUpdate = time.Now()
}

// handleUpdateRequest handles a specific update request
func (l *LiveUpdater) handleUpdateRequest(request *UpdateRequest) {
	l.mutex.RLock()
	component := l.components[request.ComponentID]
	l.mutex.RUnlock()

	if component == nil {
		return
	}

	// Perform targeted update
	l.performUpdate(request.ForceUpdate)
}

// generateDisplayContent generates the complete display content
func (l *LiveUpdater) generateDisplayContent(totalHeight int) string {
	l.updateBuffer.Reset()

	var currentHeight int

	for _, componentID := range l.componentOrder {
		component := l.components[componentID]
		if component == nil {
			continue
		}

		height := component.GetHeight()
		if currentHeight+height > l.maxLines {
			break
		}

		content := component.Render()
		l.lastContent[componentID] = content

		if content != "" {
			l.updateBuffer.WriteString(content)
			if currentHeight+height < totalHeight {
				l.updateBuffer.WriteString("\n")
			}
		}

		currentHeight += height
	}

	return l.updateBuffer.String()
}

// updateDisplay updates the terminal display
func (l *LiveUpdater) updateDisplay(content string) error {
	// Move cursor to saved position or top
	if l.cursorSaved {
		_, err := l.writer.Write([]byte("\033[u")) // Restore cursor position
		if err != nil {
			return err
		}
	} else {
		_, err := l.writer.Write([]byte(fmt.Sprintf("\033[%dA", l.currentLines))) // Move up
		if err != nil {
			return err
		}
	}

	// Clear from cursor to end of screen
	_, err := l.writer.Write([]byte("\033[J"))
	if err != nil {
		return err
	}

	// Write new content
	_, err = l.writer.Write([]byte(content))
	if err != nil {
		return err
	}

	// Update line count
	l.currentLines = strings.Count(content, "\n")

	return nil
}

// initializeTerminal sets up the terminal for live updates
func (l *LiveUpdater) initializeTerminal() error {
	// Detect terminal size
	l.detectTerminalSize()

	// Save cursor position
	_, err := l.writer.Write([]byte("\033[s"))
	if err != nil {
		return err
	}
	l.cursorSaved = true

	// Use alternate screen if configured
	if l.alternateScreen {
		_, err = l.writer.Write([]byte("\033[?1049h"))
		if err != nil {
			return err
		}
	}

	// Hide cursor
	_, err = l.writer.Write([]byte("\033[?25l"))
	if err != nil {
		return err
	}

	return nil
}

// restoreTerminal restores the terminal to its original state
func (l *LiveUpdater) restoreTerminal() error {
	// Show cursor
	_, err := l.writer.Write([]byte("\033[?25h"))
	if err != nil {
		return err
	}

	// Restore alternate screen if used
	if l.alternateScreen {
		_, err = l.writer.Write([]byte("\033[?1049l"))
		if err != nil {
			return err
		}
	}

	// Restore cursor position
	if l.cursorSaved {
		_, err = l.writer.Write([]byte("\033[u"))
		if err != nil {
			return err
		}
	}

	return nil
}

// detectTerminalSize detects the current terminal dimensions
func (l *LiveUpdater) detectTerminalSize() {
	// Simple fallback implementation
	// In a real implementation, you'd use syscalls or environment variables
	l.terminalWidth = 80
	l.terminalHeight = 24
}

// sortComponentsByPriority sorts components by their priority
func (l *LiveUpdater) sortComponentsByPriority() {
	// Simple bubble sort by priority (descending)
	for i := 0; i < len(l.componentOrder)-1; i++ {
		for j := 0; j < len(l.componentOrder)-i-1; j++ {
			comp1 := l.components[l.componentOrder[j]]
			comp2 := l.components[l.componentOrder[j+1]]

			if comp1 != nil && comp2 != nil && comp1.GetPriority() < comp2.GetPriority() {
				l.componentOrder[j], l.componentOrder[j+1] = l.componentOrder[j+1], l.componentOrder[j]
			}
		}
	}
}

// GetRegisteredComponents returns the list of registered component IDs
func (l *LiveUpdater) GetRegisteredComponents() []string {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	result := make([]string, len(l.componentOrder))
	copy(result, l.componentOrder)
	return result
}

// IsActive returns whether the live updater is currently active
func (l *LiveUpdater) IsActive() bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.isActive
}

// GetTerminalSize returns the current terminal dimensions
func (l *LiveUpdater) GetTerminalSize() (width, height int) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.terminalWidth, l.terminalHeight
}

// SetTerminalSize manually sets the terminal dimensions
func (l *LiveUpdater) SetTerminalSize(width, height int) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.terminalWidth = width
	l.terminalHeight = height
}

// ClearScreen clears the entire screen
func (l *LiveUpdater) ClearScreen() error {
	_, err := l.writer.Write([]byte("\033[2J\033[H"))
	return err
}

// GetLastUpdateTime returns the timestamp of the last update
func (l *LiveUpdater) GetLastUpdateTime() time.Time {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.lastUpdate
}
