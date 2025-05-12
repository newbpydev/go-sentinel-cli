package tui

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/newbpydev/go-sentinel/internal/debouncer"
	"github.com/newbpydev/go-sentinel/internal/event"
	"github.com/newbpydev/go-sentinel/internal/runner"
	"github.com/newbpydev/go-sentinel/internal/ui"
	"github.com/newbpydev/go-sentinel/internal/watcher"
)

// programSender is an interface for the tea.Program methods we use
// This allows us to mock the tea.Program for testing
type programSender interface {
	Send(tea.Msg)
	Start() error
	Quit()
}

// App is the main TUI application controller
type App struct {
	// Components
	watcher   *watcher.Watcher
	debouncer *debouncer.ActionDebouncer
	runner    *runner.Runner
	
	// Channels
	fileEvents   chan watcher.Event
	testEvents   chan []byte
	errorCh      chan error
	
	// State
	rootPath       string
	program        programSender // Changed from *tea.Program to interface
	model          *ui.TUITestExplorerModel
	watchModeEnabled bool
	
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
	
	// Create channels for Phase 6 implementation (concurrency & error handling)
	fileEvents := make(chan watcher.Event, 32)
	testEvents := make(chan []byte, 100)
	errorCh := make(chan error, 10) // Buffered error channel for all components
	
	return &App{
		watcher:     w,
		debouncer:   d,
		runner:      r,
		fileEvents:  fileEvents,
		testEvents:  testEvents,
		errorCh:     errorCh,
		rootPath:    rootPath,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

// safeGoroutine runs a function in a goroutine with panic recovery
// Implements task 6.3.2: Add recovery handlers for each goroutine
func (a *App) safeGoroutine(name string, fn func()) {
	// Create a wrapper function with panic recovery
	go func() {
		// Set up panic recovery
		defer func() {
			if r := recover(); r != nil {
				// Convert panic value to error
				var err error
				switch x := r.(type) {
				case string:
					err = errors.New(x)
				case error:
					err = x
				default:
					err = fmt.Errorf("unknown panic: %v", r)
				}
				
				// Send error to error channel
				a.errorCh <- fmt.Errorf("panic in %s: %w", name, err)
				
				// Also log to UI immediately
				a.program.Send(ui.LogEntryMsg{
					Content: fmt.Sprintf("[ERROR] Recovered from panic in %s: %v", name, err),
				})
				
				// Show log panel to make sure the error is visible
				a.program.Send(ui.ShowLogViewMsg{Show: true})
			}
		}()
		
		// Call the wrapped function
		fn()
	}()
}

// errorHandlerLoop processes errors from all components
// Implements task 6.3.1: Design error channel architecture
func (a *App) errorHandlerLoop() {
	// CRITICAL FIX: Check for nil program before sending messages
	if a.program != nil {
		a.program.Send(ui.LogEntryMsg{Content: "[SYSTEM] Error handler started"})
	}
	
	for {
		select {
		case <-a.ctx.Done():
			return
		case err, ok := <-a.errorCh:
			// Check if channel is closed
			if !ok {
				return
			}
			
			// CRITICAL FIX: Check for nil program before sending messages
			if a.program == nil {
				// Can't log to UI if program is nil
				fmt.Printf("[ERROR] %v (program is nil, can't send to UI)\n", err)
				continue
			}
			
			// Log the error to the UI
			a.program.Send(ui.LogEntryMsg{
				Content: fmt.Sprintf("[ERROR] %v", err),
			})
			
			// Attempt to restart failed components if needed
			if strings.Contains(err.Error(), "panic in watcherLoop") {
				// Restart watcher loop after a short delay
				time.Sleep(1 * time.Second)
				go a.safeGoroutine("watcherLoop-restart", a.watcherLoop)
				a.program.Send(ui.LogEntryMsg{Content: "[SYSTEM] Watcher loop restarted"})
			} else if strings.Contains(err.Error(), "panic in testResultsLoop") {
				// Restart test results loop after a short delay
				time.Sleep(1 * time.Second)
				go a.safeGoroutine("testResultsLoop-restart", a.testResultsLoop)
				a.program.Send(ui.LogEntryMsg{Content: "[SYSTEM] Test results loop restarted"})
			}
		}
	}
}

// Start starts the TUI application
func (a *App) Start() error {
	// Create & populate initial test tree
	root := a.scanForTestFiles()
	
	// Create a TUI model
	model := ui.NewTUITestExplorerModel(root)
	a.model = &model
	
	// CRITICAL FIX: Initialize program BEFORE starting any goroutines
	// that might try to access it
	a.program = tea.NewProgram(&model, tea.WithAltScreen())
	
	// Add a subscription for running tests (CRITICAL FIX: implementing message handler for RunTestsMsg)
	// This connects the UI's 'a' key press to the actual test execution
	go func() {
		for {
			select {
			case <-a.ctx.Done():
				return
			default:
				// Poll for messages - tea.Program doesn't have a built-in subscription mechanism
				time.Sleep(50 * time.Millisecond)
				
				// Get the current model state to check for RunTestsMsg
				if a.model == nil {
					continue
				}
				
				// Check if there's a pending RunTestsMsg
				if a.model.HasPendingTestRun() {
					// Extract the test run details
					pkg, test := a.model.GetAndClearPendingTestRun()
					
					// Run the tests with the extracted parameters
					a.runTests(pkg, test)
				}
			}
		}
	}()
	
	// Adjust the model to handle watch mode state
	model.SetWatchModeCallback(func(enabled bool) {
		// First, just update the flag without doing anything else
		a.watchModeEnabled = enabled
		
		// Use a goroutine to avoid blocking the UI thread and creating a recursive message loop
		go a.safeGoroutine("watchModeCallback", func() {
			// Small delay to ensure the original message processing completes
			time.Sleep(100 * time.Millisecond)
			
			// Now it's safe to send feedback to the UI
			if a.program != nil {
				a.program.Send(ui.LogEntryMsg{
					Content: fmt.Sprintf("Watch mode set to: %v", enabled),
				})
			}
		})
	})
	
	// Initialize with watch mode enabled by default
	a.watchModeEnabled = true
	
	// Start error handler loop (Phase 6.3.2)
	go a.safeGoroutine("errorHandlerLoop", a.errorHandlerLoop)
	
	// Start components in goroutines with panic recovery (Phase 6.2.1 & 6.3.2)
	go a.safeGoroutine("watcherLoop", a.watcherLoop)
	go a.safeGoroutine("testResultsLoop", a.testResultsLoop)
	
	// Start Bubble Tea program (already initialized above)
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

// scanForTestFiles finds all test files in the project and builds a tree structure
func (a *App) scanForTestFiles() *ui.TreeNode {
	root := &ui.TreeNode{
		Title:    "root",
		Expanded: true,
		Children: []*ui.TreeNode{},
	}

	// Map to track packages for organizing tests
	packageMap := make(map[string]*ui.TreeNode)

	// Walk the directory tree to find test files
	_ = filepath.Walk(a.rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Only look at Go test files
		if !info.IsDir() && strings.HasSuffix(path, "_test.go") {
			// Extract package name from path
			relPath, err := filepath.Rel(a.rootPath, filepath.Dir(path))
			if err != nil {
				relPath = filepath.Dir(path)
			}

			// Convert to package path format
			pkgPath := "." + string(filepath.Separator) + relPath
			if relPath == "." {
				pkgPath = "."
			}

			// Create package node if it doesn't exist
			pkg, exists := packageMap[pkgPath]
			if !exists {
				pkg = &ui.TreeNode{
					Title:    pkgPath,
					Expanded: false,
					Children: []*ui.TreeNode{},
				}
				packageMap[pkgPath] = pkg
				root.Children = append(root.Children, pkg)
			}

			// Extract test names from file
			testNames := extractTestNamesFromFile(path)
			for _, testName := range testNames {
				// Add test node
				pkg.Children = append(pkg.Children, &ui.TreeNode{
					Title: testName,
				})
			}
		}

		return nil
	})

	// If no tests found, return a stub tree
	if len(root.Children) == 0 {
		return createStubTree()
	}

	return root
}

// extractTestNamesFromFile scans a Go file for test functions
func extractTestNamesFromFile(filePath string) []string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	// Simple regex pattern to find test functions
	re := regexp.MustCompile(`func\s+(Test[a-zA-Z0-9_]+)\(`)
	matches := re.FindAllSubmatch(content, -1)

	testNames := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			testNames = append(testNames, string(match[1]))
		}
	}

	return testNames
}

// watcherLoop connects the watcher, debouncer, and runner
func (a *App) watcherLoop() {
	// A cache of recently changed files for debouncing
	changedFileTime := make(map[string]time.Time)
	
	for {
		select {
		case <-a.ctx.Done():
			return
		case event := <-a.watcher.Events:
			// Only watch .go files
			if filepath.Ext(event.Name) != ".go" {
				continue
			}
			
			// Simple debouncing - skip if this file was just processed
			now := time.Now()
			if lastTime, exists := changedFileTime[event.Name]; exists {
				if now.Sub(lastTime) < 500*time.Millisecond {
					// Skip this event if it's too soon after the last one
					continue
				}
			}
			changedFileTime[event.Name] = now

			// Notify UI about the file change
			a.program.Send(ui.FileChangedMsg{
				Path: event.Name,
			})

			// Log the file change detection
			a.program.Send(ui.LogEntryMsg{
				Content: fmt.Sprintf("File changed: %s. Watch mode: %v", 
					event.Name, a.watchModeEnabled),
			})

			// Only run tests if watch mode is enabled
			if a.watchModeEnabled {
				// Use a goroutine to prevent blocking the main loop
				go func(filePath string) {
					// Add a small delay to avoid UI freezing
					time.Sleep(500 * time.Millisecond)
					
					// Run tests for the package containing this file
					pkg := ""
					// Try to determine package from file path
					dir := filepath.Dir(filePath)
					relPath, err := filepath.Rel(a.rootPath, dir)
					if err == nil && relPath != "." && relPath != ".." {
						pkg = "." + string(filepath.Separator) + relPath
					}
					
					// If we couldn't determine the package, run all tests
					if pkg == "" {
						pkg = "./..."
					}
					
					a.runTests(pkg, "")
				}(event.Name)
			}
		}
	}
}

// runTests runs tests for a package and collects the output
func (a *App) runTests(pkg, test string) {
	// Clear the log panel first for a fresh run
	a.program.Send(ui.ClearLogMsg{})
	
	// Send test started message to UI
	a.program.Send(ui.TestsStartedMsg{Package: pkg})
	
	// Add a run marker to the log
	a.program.Send(ui.LogEntryMsg{Content: fmt.Sprintf("\n=== Running tests for %s ===", pkg)})
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(a.ctx, 2*time.Minute)
	defer cancel()

	// Show the log panel automatically when running tests
	a.program.Send(ui.ShowLogViewMsg{Show: true})
	
	// Run tests and collect output
	err := a.runner.RunWithContext(ctx, pkg, test, a.testEvents)
	if err != nil {
		a.program.Send(event.ErrorEvent{Err: err})
		a.program.Send(ui.TestsCompletedMsg{Success: false})
		a.program.Send(ui.LogEntryMsg{Content: fmt.Sprintf("\n=== Test run failed: %v ===", err)})
	} else {
		// No error, ensure we still complete the test cycle to update the UI
		a.program.Send(ui.LogEntryMsg{Content: "\n=== Test run completed successfully ==="})
		
		// Adding an explicit DONE message to trigger result processing
		// This is needed because some test runs may not produce output
		a.testEvents <- []byte("DONE")
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
			
			// First, always add the output to the log panel
			if len(output) > 0 {
				logLine := string(output)
				// Send log line to UI for display in the log panel
				a.program.Send(ui.LogEntryMsg{Content: logLine})
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
			
			// Parse the test output to extract test information
			// Look for test output in the format: 
			// {"Time":"2023-05-12T12:39:54.30","Action":"output","Package":"github.com/newbpydev/go-sentinel/internal/runner","Test":"TestRunWithContext","Output":"=== RUN   TestRunWithContext\n"}
			
			outputStr := string(output)
			
			// Only try to parse JSON output lines
			if strings.HasPrefix(outputStr, "{") && strings.HasSuffix(outputStr, "}") {
				var testEvent struct {
					Time    string `json:"Time"`
					Action  string `json:"Action"`
					Package string `json:"Package"`
					Test    string `json:"Test"`
					Output  string `json:"Output"`
					Elapsed string `json:"Elapsed,omitempty"`
				}
				
				if err := json.Unmarshal(output, &testEvent); err == nil {
					// Only process relevant events
					if testEvent.Action == "pass" || testEvent.Action == "fail" || testEvent.Action == "run" {
						// Check if we already have this test in our results
						found := false
						for i, r := range results {
							if r.Package == testEvent.Package && r.Test == testEvent.Test {
								// Update existing result
								found = true
								
								// Add this output to the existing output
								results[i].Output += testEvent.Output
								
								// Update status based on action
								if testEvent.Action == "pass" {
									results[i].Error = ""
									elapsed, _ := time.ParseDuration(testEvent.Elapsed + "s")
									results[i].Elapsed = elapsed.Seconds() // Convert to float64 seconds
								} else if testEvent.Action == "fail" {
									results[i].Error = "Test failed"
									elapsed, _ := time.ParseDuration(testEvent.Elapsed + "s")
									results[i].Elapsed = elapsed.Seconds() // Convert to float64 seconds
								}
								
								// Send intermediate updates to keep UI fresh during long test runs
								if testEvent.Action == "pass" || testEvent.Action == "fail" {
									root := createTreeFromResults(results)
									a.program.Send(ui.TestResultsMsg{
										Results: results,
										Tree:    root,
									})
								}
								break
							}
						}
						
						// If this is a new test, add it to results
						if !found && testEvent.Test != "" {
							result := event.TestResult{
								Package: testEvent.Package,
								Test:    testEvent.Test,
								Output:  testEvent.Output,
							}
							
							// For failing tests, set error message
							if testEvent.Action == "fail" {
								result.Error = "Test failed"
								elapsed, _ := time.ParseDuration(testEvent.Elapsed + "s")
								result.Elapsed = elapsed.Seconds() // Convert to float64 seconds
							} else if testEvent.Action == "pass" {
								elapsed, _ := time.ParseDuration(testEvent.Elapsed + "s")
								result.Elapsed = elapsed.Seconds() // Convert to float64 seconds
							}
							
							results = append(results, result)
							
							// Update UI immediately for a new test
							if testEvent.Action == "pass" || testEvent.Action == "fail" {
								root := createTreeFromResults(results)
								a.program.Send(ui.TestResultsMsg{
									Results: results,
									Tree:    root,
								})
							}
						}
					}
				}
			}
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
