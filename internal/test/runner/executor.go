// Package runner provides test execution implementation
package runner

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// DefaultExecutor implements the TestExecutor interface
type DefaultExecutor struct {
	mu        sync.RWMutex
	isRunning bool
	cancel    context.CancelFunc
	options   *ExecutionOptions
}

// NewExecutor creates a new test executor
func NewExecutor() TestExecutor {
	return &DefaultExecutor{
		isRunning: false,
	}
}

// Execute implements the TestExecutor interface
func (e *DefaultExecutor) Execute(ctx context.Context, packages []string, options *ExecutionOptions) (*ExecutionResult, error) {
	e.mu.Lock()
	if e.isRunning {
		e.mu.Unlock()
		return nil, fmt.Errorf("executor is already running")
	}
	e.isRunning = true
	e.options = options
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		e.isRunning = false
		e.mu.Unlock()
	}()

	startTime := time.Now()
	result := &ExecutionResult{
		Packages:    make([]*PackageResult, 0, len(packages)),
		StartTime:   startTime,
		Success:     true,
		TotalTests:  0,
		PassedTests: 0,
		FailedTests: 0,
	}

	// Execute tests for each package
	for _, pkg := range packages {
		packageResult, err := e.ExecutePackage(ctx, pkg, options)
		if err != nil {
			return nil, fmt.Errorf("failed to execute tests for package %s: %w", pkg, err)
		}

		result.Packages = append(result.Packages, packageResult)
		result.TotalTests += len(packageResult.Tests)

		// Update success status and counts
		if !packageResult.Success {
			result.Success = false
		}

		for _, test := range packageResult.Tests {
			switch test.Status {
			case TestStatusPass:
				result.PassedTests++
			case TestStatusFail:
				result.FailedTests++
			case TestStatusSkip:
				result.SkippedTests++
			}
		}
	}

	result.EndTime = time.Now()
	result.TotalDuration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// ExecutePackage implements the TestExecutor interface
func (e *DefaultExecutor) ExecutePackage(ctx context.Context, pkg string, options *ExecutionOptions) (*PackageResult, error) {
	startTime := time.Now()

	// Build the go test command
	args := []string{"test"}

	// Add JSON output if requested
	if options.JSONOutput {
		args = append(args, "-json")
	}

	// Add verbose flag if requested
	if options.Verbose {
		args = append(args, "-v")
	}

	// Add coverage if requested
	if options.Coverage {
		args = append(args, "-cover")
		if options.CoverageProfile != "" {
			args = append(args, "-coverprofile="+options.CoverageProfile)
		}
	}

	// Add parallel setting if specified
	if options.Parallel > 0 {
		args = append(args, fmt.Sprintf("-parallel=%d", options.Parallel))
	}

	// Add timeout if specified
	if options.Timeout > 0 {
		args = append(args, "-timeout="+options.Timeout.String())
	}

	// Add additional arguments
	args = append(args, options.Args...)

	// Add the package
	args = append(args, pkg)

	// Create the command
	cmd := exec.CommandContext(ctx, "go", args...)

	// Set working directory if specified
	if options.WorkingDirectory != "" {
		cmd.Dir = options.WorkingDirectory
	}

	// Set environment variables
	if len(options.Env) > 0 {
		env := os.Environ()
		for key, value := range options.Env {
			env = append(env, key+"="+value)
		}
		cmd.Env = env
	}

	// Execute the command and capture output
	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	// Parse the results
	result := &PackageResult{
		Package:  pkg,
		Success:  err == nil,
		Duration: time.Since(startTime),
		Output:   outputStr,
		Tests:    make([]*TestResult, 0),
	}

	if err != nil {
		result.Error = err
	}

	// Parse test results from output
	result.Tests = e.parseTestResults(outputStr, pkg)

	// Update package success based on individual test results
	for _, test := range result.Tests {
		if test.Status == TestStatusFail {
			result.Success = false
		}
	}

	return result, nil
}

// Cancel implements the TestExecutor interface
func (e *DefaultExecutor) Cancel() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.isRunning {
		return fmt.Errorf("no test execution is currently running")
	}

	if e.cancel != nil {
		e.cancel()
	}

	return nil
}

// IsRunning implements the TestExecutor interface
func (e *DefaultExecutor) IsRunning() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.isRunning
}

// parseTestResults parses test results from go test output
func (e *DefaultExecutor) parseTestResults(output, pkg string) []*TestResult {
	var tests []*TestResult
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Parse different line formats
		if strings.Contains(line, "--- PASS:") {
			test := e.parseTestLine(line, pkg, TestStatusPass)
			if test != nil {
				tests = append(tests, test)
			}
		} else if strings.Contains(line, "--- FAIL:") {
			test := e.parseTestLine(line, pkg, TestStatusFail)
			if test != nil {
				tests = append(tests, test)
			}
		} else if strings.Contains(line, "--- SKIP:") {
			test := e.parseTestLine(line, pkg, TestStatusSkip)
			if test != nil {
				tests = append(tests, test)
			}
		}
	}

	return tests
}

// parseTestLine parses a single test result line
func (e *DefaultExecutor) parseTestLine(line, pkg string, status TestStatus) *TestResult {
	// Example: "--- PASS: TestName (0.00s)"
	parts := strings.Fields(line)
	if len(parts) < 3 {
		return nil
	}

	testName := parts[2]

	// Parse duration if present
	var duration time.Duration
	if len(parts) >= 4 {
		durationStr := strings.Trim(parts[3], "()")
		if d, err := time.ParseDuration(strings.Replace(durationStr, "s", "s", 1)); err == nil {
			duration = d
		}
	}

	return &TestResult{
		Name:     testName,
		Package:  pkg,
		Status:   status,
		Duration: duration,
		Output:   line,
	}
}
