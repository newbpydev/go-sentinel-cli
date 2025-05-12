package tui

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/go-sentinel/internal/debouncer"
	"github.com/newbpydev/go-sentinel/internal/event"
	"github.com/newbpydev/go-sentinel/internal/runner"
	"github.com/newbpydev/go-sentinel/internal/ui"
	"github.com/newbpydev/go-sentinel/internal/watcher"
)

// App is the main TUI application controller
type App struct {
	// Components
	watcher   *watcher.Watcher
	debouncer *debouncer.ActionDebouncer
	runner    *runner.Runner
	
	// Channels
	fileEvents   chan watcher.Event
	testEvents   chan []byte
	
	// State
	rootPath       string
	program        *tea.Program
	model          *ui.TUITestExplorerModel
	
	// Context for graceful shutdown
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewApp creates a new TUI application
func NewApp(rootPath string) (*App, error) {
	// Initialize components
	w, err := watcher.NewWatcher(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize watcher: %w", err)
	}
	
	d := debouncer.NewActionDebouncer(300 * time.Millisecond)
	r := runner.NewRunner()
	
	// Create context with cancel for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	
	return &App{
		watcher:     w,
		debouncer:   d,
		runner:      r,
		fileEvents:  make(chan watcher.Event, 32),
		testEvents:  make(chan []byte, 100),
		rootPath:    rootPath,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// Start starts the TUI application
func (a *App) Start() error {
	// For simplicity during initial integration, use a stub tree
	root := createStubTree()
	
	// Create TUI model
	model := ui.NewTUITestExplorerModel(root)
	a.model = &model
	
	// Start components in goroutines
	go a.watcherLoop()
	go a.testResultsLoop()
	
	// Start Bubble Tea program
	a.program = tea.NewProgram(&model, tea.WithAltScreen())
	return a.program.Start()
}

// Stop gracefully stops the application
func (a *App) Stop() {
	// Signal all goroutines to stop
	a.cancel()
	
	// Close the watcher to stop file events
	if a.watcher != nil {
		a.watcher.Close()
	}
	
	// Close channels
	close(a.fileEvents)
	close(a.testEvents)
	
	// Quit the program if running
	if a.program != nil {
		a.program.Quit()
	}
}

// watcherLoop connects the watcher, debouncer, and runner
func (a *App) watcherLoop() {
	// Forward events from watcher to our channel
	go func() {
		for {
			select {
			case <-a.ctx.Done():
				return
			case event, ok := <-a.watcher.Events:
				if !ok {
					return
				}
				a.fileEvents <- event
			}
		}
	}()
	
	// Process debounced file events and trigger test runs
	for {
		select {
		case <-a.ctx.Done():
			return
		case event, ok := <-a.fileEvents:
			if !ok {
				return
			}
			
			// Skip events for non-Go files
			if !strings.HasSuffix(event.Name, ".go") {
				continue
			}
			
			// Debounce events by file path
			a.debouncer.Debounce(event.Name, func() {
				// Extract package from file path
				pkg := "./..."  // Default to all packages
				
				// Try to determine package from file path
				relPath, err := filepath.Rel(a.rootPath, filepath.Dir(event.Name))
				if err == nil && relPath != "." && relPath != ".." {
					pkg = "./" + relPath
				}
				
				// Send file changed event to UI
				a.program.Send(ui.FileChangedMsg{Path: event.Name})
				
				// Run tests for the package
				a.runTests(pkg, "")
			})
		}
	}
}

// runTests runs tests for a package and collects the output
func (a *App) runTests(pkg, test string) {
	// Send test started message to UI
	a.program.Send(ui.TestsStartedMsg{Package: pkg})
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(a.ctx, 2*time.Minute)
	defer cancel()
	
	// Run tests and collect output
	err := a.runner.RunWithContext(ctx, pkg, test, a.testEvents)
	if err != nil {
		a.program.Send(event.ErrorEvent{Err: err})
		a.program.Send(ui.TestsCompletedMsg{Success: false})
	}
}

// testResultsLoop processes test output and updates the UI
func (a *App) testResultsLoop() {
	var results []event.TestResult
	
	for {
		select {
		case <-a.ctx.Done():
			return
		case output, ok := <-a.testEvents:
			if !ok {
				return
			}
			
			// Check if this is the end of output
			if bytes.Equal(output, []byte("DONE")) {
				// For now, create a simple tree from results
				root := createTreeFromResults(results)
				a.program.Send(ui.TestResultsMsg{
					Results: results,
					Tree:    root,
				})
				a.program.Send(ui.TestsCompletedMsg{Success: true})
				
				// Reset results for next run
				results = nil
				continue
			}
			
			// Try to parse the test output
			// For now, just print the output as we haven't fully implemented parser.ParseTestEvent
			fmt.Printf("Test output: %s\n", string(output))
			
			// Create a simple test result with the output
			result := event.TestResult{
				Package: "unknown",
				Test:    "unknown",
				Output:  string(output),
			}
			
			results = append(results, result)
		}
	}
}

// createTreeFromResults creates a UI tree from test results
func createTreeFromResults(results []event.TestResult) *ui.TreeNode {
	// Create the root node
	root := &ui.TreeNode{
		Title:    "root",
		Expanded: true,
		Children: []*ui.TreeNode{},
	}
	
	// Group results by package
	packages := make(map[string][]*ui.TreeNode)
	
	// Create test nodes
	for _, result := range results {
		passed := true
		if result.Error != "" {
			passed = false
		}
		
		testNode := &ui.TreeNode{
			Title:    result.Test,
			Passed:   &passed,
			Duration: result.Elapsed,
			Error:    result.Error,
		}
		
		packages[result.Package] = append(packages[result.Package], testNode)
	}
	
	// Create package nodes and add test nodes as children
	for pkg, tests := range packages {
		pkgNode := &ui.TreeNode{
			Title:    pkg,
			Expanded: true,
			Children: tests,
		}
		
		// Set parent reference
		for _, test := range tests {
			test.Parent = pkgNode
		}
		
		root.Children = append(root.Children, pkgNode)
		pkgNode.Parent = root
	}
	
	return root
}

// Helper to create a stub tree for initial UI
func createStubTree() *ui.TreeNode {
	// Create a simple stub tree for initial UI rendering
	return &ui.TreeNode{
		Title:    "root",
		Expanded: true,
		Children: []*ui.TreeNode{
			{Title: "internal/parser", Expanded: true, Children: []*ui.TreeNode{
				{Title: "TestParseOutput"},
				{Title: "TestConvertToTree"},
			}},
			{Title: "internal/watcher", Expanded: true, Children: []*ui.TreeNode{
				{Title: "TestFileWatch"},
			}},
		},
	}
}
