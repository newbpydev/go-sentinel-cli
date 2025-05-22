package cli

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Runner handles test execution and watch mode
type Runner struct {
	workDir string
	watcher *fsnotify.Watcher
	mu      sync.Mutex
}

// RunOptions configures how tests are run
type RunOptions struct {
	OnlyFailed bool      // Only run previously failed tests
	FailFast   bool      // Stop on first failure
	Watch      bool      // Enable watch mode
	Tests      []string  // Specific tests to run
	Packages   []string  // Specific packages to test
	Renderer   *Renderer // Custom renderer for test output
}

// NewRunner creates a new test runner
func NewRunner(workDir string) (*Runner, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create file watcher: %w", err)
	}

	return &Runner{
		workDir: workDir,
		watcher: watcher,
	}, nil
}

// Run executes tests with the given options
func (r *Runner) Run(ctx context.Context, opts RunOptions) error {
	// Use default renderer if none provided
	if opts.Renderer == nil {
		opts.Renderer = NewRenderer(os.Stdout)
	}

	if opts.Watch {
		return r.Watch(ctx, opts)
	}
	_, err := r.RunOnce(opts)
	return err
}

// RunOnce executes tests once with the given options
func (r *Runner) RunOnce(opts RunOptions) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	startTime := time.Now()

	// Show test start message
	if opts.Renderer != nil {
		opts.Renderer.RenderTestStart(nil)
	}

	// Transform phase
	transformStart := time.Now()
	args := []string{"test"}
	args = append(args, "-json", "-v") // Add -v for verbose output
	if opts.FailFast {
		args = append(args, "-failfast")
	}
	if len(opts.Tests) > 0 {
		args = append(args, "-run", strings.Join(opts.Tests, "|"))
	}
	if len(opts.Packages) > 0 {
		args = append(args, opts.Packages...)
	} else {
		args = append(args, "./...")
	}
	transformDuration := time.Since(transformStart)

	// Setup phase
	setupStart := time.Now()
	cmd := exec.Command("go", args...)
	cmd.Dir = r.workDir
	cmd.Env = os.Environ()
	setupDuration := time.Since(setupStart)

	// Collection phase
	collectStart := time.Now()
	output, err := cmd.CombinedOutput()
	outputStr := string(output)
	collectDuration := time.Since(collectStart)

	// Test execution phase (parsing)
	parseStart := time.Now()
	parser := NewParser()
	run, parseErr := parser.Parse(strings.NewReader(outputStr))
	parseDuration := time.Since(parseStart)

	if run != nil {
		run.StartTime = startTime
		run.EndTime = time.Now()
		run.Duration = run.EndTime.Sub(startTime)
		run.TransformDuration = transformDuration
		run.SetupDuration = setupDuration
		run.CollectDuration = collectDuration
		run.ParseDuration = parseDuration

		// Render test results as they come in
		if opts.Renderer != nil {
			for _, suite := range run.Suites {
				opts.Renderer.RenderSuite(suite)
			}
		}
	}

	// Prepare phase
	prepareStart := time.Now()
	if parseErr == nil && opts.Renderer != nil && run != nil {
		opts.Renderer.RenderFinalSummary(run)
	}
	if run != nil {
		run.PrepareDuration = time.Since(prepareStart)
	}

	// Return error for test failures
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Test failures have exit code 1
			if exitErr.ExitCode() == 1 {
				return outputStr, fmt.Errorf("tests failed: %s", outputStr)
			}
			return outputStr, fmt.Errorf("test execution failed with code %d: %s", exitErr.ExitCode(), outputStr)
		}
		return outputStr, fmt.Errorf("failed to run tests: %w", err)
	}

	return outputStr, nil
}

// Watch starts watching for file changes and runs tests
func (r *Runner) Watch(ctx context.Context, opts RunOptions) error {
	// Add watch paths
	if err := r.addWatchPaths(); err != nil {
		return err
	}
	defer func() {
		if err := r.watcher.Close(); err != nil {
			log.Printf("Error closing watcher: %v", err)
		}
	}()

	// Show watch mode header
	if opts.Renderer != nil {
		opts.Renderer.RenderWatchHeader()
	}

	// Run tests initially
	if _, err := r.RunOnce(opts); err != nil {
		return err
	}

	// Watch for changes
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-r.watcher.Events:
			if !ok {
				return nil
			}
			if r.shouldRunTests(event.Name) {
				// Show file change notification
				if opts.Renderer != nil {
					opts.Renderer.RenderFileChange(event.Name)
				}
				if _, err := r.RunOnce(opts); err != nil {
					return err
				}
			}
		case err, ok := <-r.watcher.Errors:
			if !ok {
				return nil
			}
			return fmt.Errorf("watcher error: %w", err)
		}
	}
}

// shouldRunTests determines if tests should be run for a file change
func (r *Runner) shouldRunTests(path string) bool {
	// Only run tests for Go files
	return strings.HasSuffix(path, ".go")
}

// addWatchPaths adds Go source files to the watcher
func (r *Runner) addWatchPaths() error {
	return filepath.Walk(r.workDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip hidden directories and vendor
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only watch Go files
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		return r.watcher.Add(path)
	})
}

// Stop stops the test runner
func (r *Runner) Stop() {
	if err := r.watcher.Close(); err != nil {
		log.Printf("Error closing watcher: %v", err)
	}
}
