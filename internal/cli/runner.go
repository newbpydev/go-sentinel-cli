package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

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
	OnlyFailed bool
	FailFast   bool
	Watch      bool
	Tests      []string
	Packages   []string
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

	// Build the go test command
	args := []string{"test"}

	// Add -json flag for parsing
	args = append(args, "-json")

	// Add fail-fast flag if enabled
	if opts.FailFast {
		args = append(args, "-failfast")
	}

	// Add test filter if specified
	if len(opts.Tests) > 0 {
		args = append(args, "-run", strings.Join(opts.Tests, "|"))
	}

	// Add packages to test
	if len(opts.Packages) > 0 {
		args = append(args, opts.Packages...)
	} else {
		args = append(args, "./...")
	}

	// Create and configure the command
	cmd := exec.Command("go", args...)
	cmd.Dir = r.workDir
	cmd.Env = os.Environ()

	// Capture output
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

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
	defer r.watcher.Close()

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
	r.watcher.Close()
}
