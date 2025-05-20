package cli

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// watchModel represents the UI state for watch mode
type watchModel struct {
	runner      *Runner
	opts        RunOptions
	spinner     spinner.Model
	keyPrompt   string
	lastOutput  string
	err         error
	quitting    bool
	fileChanged string
}

// newWatchModel creates a new watch mode model
func newWatchModel(runner *Runner, opts RunOptions) watchModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return watchModel{
		runner:    runner,
		opts:      opts,
		spinner:   s,
		keyPrompt: "\nPress 'a' to run all tests\nPress 'f' to run only failed tests\nPress 'q' to quit",
	}
}

// Init implements tea.Model
func (m watchModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.runTests,
	)
}

// Update implements tea.Model
func (m watchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "a":
			m.opts.OnlyFailed = false
			return m, m.runTests
		case "f":
			m.opts.OnlyFailed = true
			return m, m.runTests
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case fileChangeMsg:
		m.fileChanged = msg.path
		return m, m.runTests

	case testResultMsg:
		m.lastOutput = msg.output
		m.err = msg.err
		return m, nil

	case tea.WindowSizeMsg:
		// Handle window resize if needed
		return m, nil
	}

	return m, nil
}

// View implements tea.Model
func (m watchModel) View() string {
	var s string

	// Header
	s += lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#ffffff")).
		Background(lipgloss.Color("#1a1a1a")).
		Padding(0, 1).
		Render(" GO SENTINEL WATCH MODE ")
	s += "\n\n"

	// File change notification
	if m.fileChanged != "" {
		s += lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Render(fmt.Sprintf("File changed: %s\n\n", m.fileChanged))
	}

	// Test output or spinner
	if m.lastOutput != "" {
		s += m.lastOutput
	} else {
		s += fmt.Sprintf("%s Running tests...\n", m.spinner.View())
	}

	// Error output
	if m.err != nil {
		s += lipgloss.NewStyle().
			Foreground(lipgloss.Color("#ff0000")).
			Render(fmt.Sprintf("\nError: %v\n", m.err))
	}

	// Key prompt
	if !m.quitting {
		s += lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Render(m.keyPrompt)
	}

	return s
}

// runTests is a command that runs the tests
func (m watchModel) runTests() tea.Msg {
	output, err := m.runner.RunOnce(m.opts)
	return testResultMsg{output: output, err: err}
}

// Custom messages
type fileChangeMsg struct {
	path string
}

type testResultMsg struct {
	output string
	err    error
}

// StartWatch starts the watch mode UI
func (r *Runner) StartWatch(opts RunOptions) error {
	p := tea.NewProgram(
		newWatchModel(r, opts),
		tea.WithAltScreen(),
	)

	// Create channels for file events and errors
	fileEvents := make(chan string, 100)
	errorEvents := make(chan error, 100)
	done := make(chan struct{})

	// Start file watcher in a goroutine
	go func() {
		defer close(done)
		defer close(fileEvents)
		defer close(errorEvents)

		// Create debounced channel for file events
		debouncedEvents := make(chan string)
		go debounce(250*time.Millisecond, fileEvents, debouncedEvents)

		for {
			select {
			case event, ok := <-r.watcher.Events:
				if !ok {
					return
				}
				if r.shouldRunTests(event.Name) {
					fileEvents <- event.Name
				}
			case err, ok := <-r.watcher.Errors:
				if !ok {
					return
				}
				errorEvents <- fmt.Errorf("watcher error: %w", err)
			case path := <-debouncedEvents:
				p.Send(fileChangeMsg{path: path})
			case err := <-errorEvents:
				p.Send(testResultMsg{err: err})
			case <-done:
				return
			}
		}
	}()

	// Run the UI
	_, err := p.Run()

	// Clean up
	close(done)
	<-done // Wait for goroutine to finish

	return err
}

// Helper function to debounce file changes
func debounce(interval time.Duration, input chan string, output chan string) {
	var item string
	timer := time.NewTimer(interval)
	timer.Stop()

	for {
		select {
		case item = <-input:
			timer.Reset(interval)
		case <-timer.C:
			if item != "" {
				output <- item
				item = ""
			}
		}
	}
}
