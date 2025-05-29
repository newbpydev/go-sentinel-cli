// Package runner provides test execution implementation
package runner

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
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

	// Create cancellable context for proper cleanup
	executionCtx, cancel := context.WithCancel(ctx)
	e.cancel = cancel
	e.mu.Unlock()

	defer func() {
		e.mu.Lock()
		e.isRunning = false
		if e.cancel != nil {
			e.cancel()
			e.cancel = nil
		}
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

	// CRITICAL FIX: Always execute packages individually to maintain full process control
	// This prevents the resource leak issue caused by `go test ./...` spawning uncontrolled child processes

	// Expand ./... patterns manually to get individual packages
	expandedPackages, err := e.expandPackagePatterns(executionCtx, packages)
	if err != nil {
		return nil, fmt.Errorf("failed to expand package patterns: %w", err)
	}

	// Execute each package individually for complete process control
	for _, pkg := range expandedPackages {
		packageResult, err := e.ExecutePackage(executionCtx, pkg, options)
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

	// Create the command with context for proper cancellation
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

	// CRITICAL FIX: Set process group to ensure child processes are cleaned up
	// This prevents orphaned processes when the parent is terminated
	setProcessGroup(cmd)

	// CRITICAL FIX: Execute command with proper output capture and process cleanup
	// This eliminates goroutine leaks and ensures processes are properly terminated
	output, err := func() ([]byte, error) {
		// Set up pipes to capture output
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("failed to start command: %w", err)
		}

		// Read output in a separate goroutine
		outputChan := make(chan []byte, 1)
		errorChan := make(chan error, 1)

		go func() {
			defer func() {
				stdout.Close()
				stderr.Close()
			}()

			// Read stdout and stderr
			stdoutBytes, readErr := io.ReadAll(stdout)
			if readErr != nil {
				errorChan <- fmt.Errorf("failed to read stdout: %w", readErr)
				return
			}

			stderrBytes, readErr := io.ReadAll(stderr)
			if readErr != nil {
				errorChan <- fmt.Errorf("failed to read stderr: %w", readErr)
				return
			}

			// Combine stdout and stderr
			combined := append(stdoutBytes, stderrBytes...)
			outputChan <- combined
		}()

		// Wait for completion or cancellation
		waitChan := make(chan error, 1)
		go func() {
			waitChan <- cmd.Wait()
		}()

		select {
		case waitErr := <-waitChan:
			// Process completed, get the output
			// CRITICAL FIX: Even on successful completion, ensure all child processes are cleaned up
			// This addresses the Windows issue where go test spawns child processes that don't get cleaned up
			if cmd.Process != nil {
				killProcessGroup(cmd.Process)
			}

			select {
			case output := <-outputChan:
				return output, waitErr
			case readErr := <-errorChan:
				return nil, readErr
			case <-time.After(1 * time.Second):
				// Timeout reading output, return what we have
				return []byte{}, waitErr
			}
		case <-ctx.Done():
			// Context cancelled, kill the process and wait for cleanup
			if cmd.Process != nil {
				killProcessGroup(cmd.Process)
			}
			// Wait for process termination with timeout
			select {
			case <-waitChan:
				// Process terminated
			case <-time.After(2 * time.Second):
				// Force kill if still running
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
			}
			return nil, fmt.Errorf("test execution cancelled: %w", ctx.Err())
		}
	}()

	outputStr := string(output)

	// Parse test results from output first
	tests := e.parseTestResults(outputStr, pkg)

	// Determine if this is a real package error or just test failures
	isPackageError := false
	if err != nil {
		// Check if this is just exit status 1 (test failures) with valid test results
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			// Exit status 1 with parsed tests means test failures, not package error
			if len(tests) > 0 {
				isPackageError = false // This is normal test failure, not package error
			} else {
				// Exit status 1 with no parsed tests might be a real package error
				isPackageError = true
			}
		} else {
			// Other errors (compilation, missing packages, etc.) are real package errors
			isPackageError = true
		}
	}

	// Parse the results
	result := &PackageResult{
		Package:  pkg,
		Success:  err == nil,
		Duration: time.Since(startTime),
		Output:   outputStr,
		Tests:    tests,
	}

	// Only set Error for real package errors, not test failures
	if isPackageError {
		result.Error = err
	}

	// Update package success based on individual test results
	for _, test := range result.Tests {
		if test.Status == TestStatusFail {
			result.Success = false
		}
	}

	return result, nil
}

// ExecuteMultiplePackages executes tests for multiple packages in a single command
// This prevents the resource leak caused by spawning many separate go test processes
func (e *DefaultExecutor) ExecuteMultiplePackages(ctx context.Context, packages []string, options *ExecutionOptions) (*ExecutionResult, error) {
	startTime := time.Now()

	// Build the go test command for multiple packages
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

	// Add ALL packages to the single command
	args = append(args, packages...)

	// Create the command with context for proper cancellation
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

	// Set process group to ensure child processes are cleaned up
	setProcessGroup(cmd)

	// Execute the single command for all packages
	output, err := func() ([]byte, error) {
		// Set up pipes to capture output
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("failed to start command: %w", err)
		}

		// Read output in a separate goroutine
		outputChan := make(chan []byte, 1)
		errorChan := make(chan error, 1)

		go func() {
			defer func() {
				stdout.Close()
				stderr.Close()
			}()

			// Read stdout and stderr
			stdoutBytes, readErr := io.ReadAll(stdout)
			if readErr != nil {
				errorChan <- fmt.Errorf("failed to read stdout: %w", readErr)
				return
			}

			stderrBytes, readErr := io.ReadAll(stderr)
			if readErr != nil {
				errorChan <- fmt.Errorf("failed to read stderr: %w", readErr)
				return
			}

			// Combine stdout and stderr
			combined := append(stdoutBytes, stderrBytes...)
			outputChan <- combined
		}()

		// Wait for completion or cancellation
		waitChan := make(chan error, 1)
		go func() {
			waitChan <- cmd.Wait()
		}()

		select {
		case waitErr := <-waitChan:
			// Process completed, get the output
			// CRITICAL FIX: Even on successful completion, ensure all child processes are cleaned up
			// This addresses the Windows issue where go test ./... spawns child processes that don't get cleaned up
			if cmd.Process != nil {
				killProcessGroup(cmd.Process)
			}

			select {
			case output := <-outputChan:
				return output, waitErr
			case readErr := <-errorChan:
				return nil, readErr
			case <-time.After(1 * time.Second):
				// Timeout reading output, return what we have
				return []byte{}, waitErr
			}
		case <-ctx.Done():
			// Context cancelled, kill the process and wait for cleanup
			if cmd.Process != nil {
				killProcessGroup(cmd.Process)
			}
			// Wait for process termination with timeout
			select {
			case <-waitChan:
				// Process terminated
			case <-time.After(2 * time.Second):
				// Force kill if still running
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
			}
			return nil, fmt.Errorf("test execution cancelled: %w", ctx.Err())
		}
	}()

	outputStr := string(output)

	// Parse the combined results
	result := &ExecutionResult{
		Packages:    make([]*PackageResult, 0),
		StartTime:   startTime,
		Success:     err == nil,
		TotalTests:  0,
		PassedTests: 0,
		FailedTests: 0,
	}

	if err != nil {
		// Check if this is just exit status 1 (test failures) or a real package error
		isPackageError := true
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			// Try to parse results - if we get valid test results, it's just test failures
			parsedResult := e.parseMultiplePackageResults(outputStr, packages, startTime)
			if parsedResult.TotalTests > 0 {
				// We have valid test results, so this is just test failures, not package error
				result = parsedResult
				isPackageError = false
			}
		}

		if isPackageError {
			// Create a single package result with the real package error
			result.Packages = append(result.Packages, &PackageResult{
				Package:  strings.Join(packages, ", "),
				Success:  false,
				Duration: time.Since(startTime),
				Output:   outputStr,
				Error:    err,
				Tests:    make([]*TestResult, 0),
			})
		}
	} else {
		// Parse results for all packages from the combined output
		result = e.parseMultiplePackageResults(outputStr, packages, startTime)
	}

	result.EndTime = time.Now()
	result.TotalDuration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// parseMultiplePackageResults parses test results from combined output of multiple packages
func (e *DefaultExecutor) parseMultiplePackageResults(output string, packages []string, startTime time.Time) *ExecutionResult {
	result := &ExecutionResult{
		Packages:    make([]*PackageResult, 0),
		StartTime:   startTime,
		Success:     true,
		TotalTests:  0,
		PassedTests: 0,
		FailedTests: 0,
	}

	// Create a map to track package results
	packageResults := make(map[string]*PackageResult)
	for _, pkg := range packages {
		packageResults[pkg] = &PackageResult{
			Package:  pkg,
			Success:  true,
			Duration: 0,
			Output:   "",
			Tests:    make([]*TestResult, 0),
		}
	}

	// Parse the output line by line
	scanner := bufio.NewScanner(strings.NewReader(output))
	currentPackage := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Try to detect package information from JSON output or test lines
		if strings.Contains(line, "\"Package\":") {
			// JSON format - extract package name
			if packageName := e.extractPackageFromJSON(line); packageName != "" {
				currentPackage = packageName
			}
		} else if strings.Contains(line, "--- PASS:") || strings.Contains(line, "--- FAIL:") || strings.Contains(line, "--- SKIP:") {
			// Test result line - add to current package
			if currentPackage != "" {
				if pkgResult, exists := packageResults[currentPackage]; exists {
					test := e.parseTestLineForPackage(line, currentPackage)
					if test != nil {
						pkgResult.Tests = append(pkgResult.Tests, test)
						result.TotalTests++

						switch test.Status {
						case TestStatusPass:
							result.PassedTests++
						case TestStatusFail:
							result.FailedTests++
							pkgResult.Success = false
							result.Success = false
						case TestStatusSkip:
							result.SkippedTests++
						}
					}
				}
			}
		}

		// Add line to current package output
		if currentPackage != "" {
			if pkgResult, exists := packageResults[currentPackage]; exists {
				if pkgResult.Output == "" {
					pkgResult.Output = line
				} else {
					pkgResult.Output += "\n" + line
				}
			}
		}
	}

	// Convert map to slice
	for _, pkg := range packages {
		if pkgResult, exists := packageResults[pkg]; exists {
			result.Packages = append(result.Packages, pkgResult)
		}
	}

	return result
}

// extractPackageFromJSON extracts package name from JSON output line
func (e *DefaultExecutor) extractPackageFromJSON(line string) string {
	// Simple JSON parsing to extract package name
	if start := strings.Index(line, "\"Package\":\""); start != -1 {
		start += len("\"Package\":\"")
		if end := strings.Index(line[start:], "\""); end != -1 {
			return line[start : start+end]
		}
	}
	return ""
}

// parseTestLineForPackage parses a test line for a specific package
func (e *DefaultExecutor) parseTestLineForPackage(line, pkg string) *TestResult {
	var status TestStatus
	if strings.Contains(line, "--- PASS:") {
		status = TestStatusPass
	} else if strings.Contains(line, "--- FAIL:") {
		status = TestStatusFail
	} else if strings.Contains(line, "--- SKIP:") {
		status = TestStatusSkip
	} else {
		return nil
	}

	return e.parseTestLine(line, pkg, status)
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

// killProcessGroup terminates the process and its children
func killProcessGroup(process *os.Process) {
	if runtime.GOOS == "windows" {
		// On Windows, we need to kill the entire process tree
		// Use taskkill to kill the process and all its children
		killCmd := exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprintf("%d", process.Pid))
		killCmd.CombinedOutput() // Ignore output and errors - process may already be gone

		// Also try direct process kill as fallback
		process.Kill() // Ignore error - process may already be gone
	} else {
		// On Unix-like systems, kill the process group
		// Send SIGTERM first for graceful shutdown
		process.Signal(syscall.SIGTERM)

		// Wait a bit and force kill if necessary
		time.Sleep(100 * time.Millisecond)
		process.Kill()
	}
}

// expandPackagePatterns expands package patterns like "./..." into individual package paths
func (e *DefaultExecutor) expandPackagePatterns(ctx context.Context, packages []string) ([]string, error) {
	var expandedPackages []string

	for _, pkg := range packages {
		if strings.Contains(pkg, "...") {
			// Use go list to expand the pattern
			cmd := exec.CommandContext(ctx, "go", "list", pkg)
			output, err := cmd.Output()
			if err != nil {
				return nil, fmt.Errorf("failed to expand package pattern %s: %w", pkg, err)
			}

			// Parse the output to get individual packages
			scanner := bufio.NewScanner(strings.NewReader(string(output)))
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					expandedPackages = append(expandedPackages, line)
				}
			}

			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("error reading go list output for %s: %w", pkg, err)
			}
		} else {
			// Not a pattern, add as-is
			expandedPackages = append(expandedPackages, pkg)
		}
	}

	return expandedPackages, nil
}
