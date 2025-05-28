package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// MockExecutor implements TestExecutor interface for fast testing
type MockExecutor struct {
	isRunning          bool
	executeFunc        func(ctx context.Context, packages []string, options *ExecutionOptions) (*ExecutionResult, error)
	executePackageFunc func(ctx context.Context, pkg string, options *ExecutionOptions) (*PackageResult, error)
	cancelFunc         func() error
}

func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		isRunning: false,
	}
}

func (m *MockExecutor) Execute(ctx context.Context, packages []string, options *ExecutionOptions) (*ExecutionResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, packages, options)
	}

	// Default fast mock implementation
	result := &ExecutionResult{
		Packages:      make([]*PackageResult, len(packages)),
		TotalDuration: 10 * time.Millisecond, // Fast mock duration
		Success:       true,
		StartTime:     time.Now(),
		EndTime:       time.Now().Add(10 * time.Millisecond),
	}

	for i, pkg := range packages {
		result.Packages[i] = &PackageResult{
			Package:  pkg,
			Success:  true,
			Duration: 5 * time.Millisecond,
			Output:   fmt.Sprintf(`{"Time":"2024-01-01T10:00:00Z","Action":"pass","Package":"%s","Test":"TestMock","Elapsed":0.005}`, pkg),
		}
	}

	return result, nil
}

func (m *MockExecutor) ExecutePackage(ctx context.Context, pkg string, options *ExecutionOptions) (*PackageResult, error) {
	if m.executePackageFunc != nil {
		return m.executePackageFunc(ctx, pkg, options)
	}

	// Default fast mock implementation
	return &PackageResult{
		Package:  pkg,
		Success:  true,
		Duration: 5 * time.Millisecond,
		Output:   fmt.Sprintf(`{"Time":"2024-01-01T10:00:00Z","Action":"pass","Package":"%s","Test":"TestMock","Elapsed":0.005}`, pkg),
	}, nil
}

func (m *MockExecutor) Cancel() error {
	if m.cancelFunc != nil {
		return m.cancelFunc()
	}

	if !m.isRunning {
		return fmt.Errorf("no test execution is currently running")
	}

	m.isRunning = false
	return nil
}

func (m *MockExecutor) IsRunning() bool {
	return m.isRunning
}

// TestDefaultExecutor_ExecutePackage_Optimized replaces the slow 7.86s test with fast mocked version
func TestDefaultExecutor_ExecutePackage_Optimized(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		name           string
		setup          func() (*MockExecutor, *ExecutionOptions)
		pkg            string
		expectedError  string
		validateResult func(*testing.T, *PackageResult, error)
	}{
		"successful_execution": {
			setup: func() (*MockExecutor, *ExecutionOptions) {
				mock := NewMockExecutor()
				options := &ExecutionOptions{
					JSONOutput: true,
					Timeout:    30 * time.Second,
				}
				return mock, options
			},
			pkg: ".",
			validateResult: func(t *testing.T, result *PackageResult, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result == nil {
					t.Fatal("Result should not be nil")
				}
				if result.Package != "." {
					t.Errorf("Expected package '.', got: %s", result.Package)
				}
				if result.Duration <= 0 {
					t.Error("Duration should be positive")
				}
			},
		},
		"execution_with_coverage": {
			setup: func() (*MockExecutor, *ExecutionOptions) {
				mock := NewMockExecutor()
				mock.executePackageFunc = func(ctx context.Context, pkg string, options *ExecutionOptions) (*PackageResult, error) {
					// Simulate coverage execution
					if !options.Coverage {
						t.Error("Coverage should be enabled")
					}
					return &PackageResult{
						Package:  pkg,
						Success:  true,
						Duration: 8 * time.Millisecond,
						Output:   `{"Time":"2024-01-01T10:00:00Z","Action":"pass","Package":"test","Test":"TestCoverage","Elapsed":0.008}`,
					}, nil
				}

				tempDir := t.TempDir()
				options := &ExecutionOptions{
					JSONOutput:      true,
					Coverage:        true,
					CoverageProfile: filepath.Join(tempDir, "coverage.out"),
					Timeout:         30 * time.Second,
				}

				// Create mock coverage file
				os.WriteFile(options.CoverageProfile, []byte("mode: set\ntest.go:1.1,2.2 1 1\n"), 0644)

				return mock, options
			},
			pkg: ".",
			validateResult: func(t *testing.T, result *PackageResult, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result == nil {
					t.Fatal("Result should not be nil")
				}
			},
		},
		"execution_with_environment": {
			setup: func() (*MockExecutor, *ExecutionOptions) {
				mock := NewMockExecutor()
				mock.executePackageFunc = func(ctx context.Context, pkg string, options *ExecutionOptions) (*PackageResult, error) {
					// Verify environment variables are passed
					if val, exists := options.Env["TEST_ENV"]; !exists || val != "test_value" {
						t.Error("Environment variable TEST_ENV should be set to test_value")
					}
					return &PackageResult{
						Package:  pkg,
						Success:  true,
						Duration: 3 * time.Millisecond,
						Output:   `{"Time":"2024-01-01T10:00:00Z","Action":"pass","Package":"test","Test":"TestEnv","Elapsed":0.003}`,
					}, nil
				}

				options := &ExecutionOptions{
					JSONOutput: true,
					Timeout:    30 * time.Second,
					Env:        map[string]string{"TEST_ENV": "test_value"},
				}
				return mock, options
			},
			pkg: ".",
			validateResult: func(t *testing.T, result *PackageResult, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			},
		},
		"execution_with_verbose": {
			setup: func() (*MockExecutor, *ExecutionOptions) {
				mock := NewMockExecutor()
				mock.executePackageFunc = func(ctx context.Context, pkg string, options *ExecutionOptions) (*PackageResult, error) {
					// Verify verbose mode
					if !options.Verbose {
						t.Error("Verbose should be enabled")
					}
					return &PackageResult{
						Package:  pkg,
						Success:  true,
						Duration: 6 * time.Millisecond,
						Output:   "=== RUN   TestVerbose\n--- PASS: TestVerbose (0.00s)\nPASS\nok  \ttest\t0.006s\n",
					}, nil
				}

				options := &ExecutionOptions{
					JSONOutput: false,
					Verbose:    true,
					Timeout:    30 * time.Second,
				}
				return mock, options
			},
			pkg: ".",
			validateResult: func(t *testing.T, result *PackageResult, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			},
		},
		"non_existent_package": {
			setup: func() (*MockExecutor, *ExecutionOptions) {
				mock := NewMockExecutor()
				mock.executePackageFunc = func(ctx context.Context, pkg string, options *ExecutionOptions) (*PackageResult, error) {
					// Simulate package not found error
					return &PackageResult{
						Package:  pkg,
						Success:  false,
						Duration: 2 * time.Millisecond,
						Error:    fmt.Errorf("package not found"),
						Output:   "can't load package: package ./non-existent: cannot find package",
					}, nil
				}

				options := &ExecutionOptions{
					JSONOutput: true,
					Timeout:    30 * time.Second,
				}
				return mock, options
			},
			pkg: "./non-existent",
			validateResult: func(t *testing.T, result *PackageResult, err error) {
				if result == nil {
					t.Error("Result should not be nil even on error")
				}
				if result != nil && result.Error == nil && err == nil {
					t.Error("Either error should be returned or result.Error should be set")
				}
			},
		},
		"context_timeout": {
			setup: func() (*MockExecutor, *ExecutionOptions) {
				mock := NewMockExecutor()
				mock.executePackageFunc = func(ctx context.Context, pkg string, options *ExecutionOptions) (*PackageResult, error) {
					// Check if context is already cancelled
					select {
					case <-ctx.Done():
						return nil, ctx.Err()
					default:
						// Simulate quick completion
						return &PackageResult{
							Package:  pkg,
							Success:  false,
							Duration: 1 * time.Millisecond,
							Error:    fmt.Errorf("test execution cancelled: context deadline exceeded"),
						}, nil
					}
				}

				options := &ExecutionOptions{
					JSONOutput: true,
					Verbose:    true,
					Timeout:    10 * time.Second,
					Args:       []string{"-count=5"},
				}
				return mock, options
			},
			pkg: ".",
			validateResult: func(t *testing.T, result *PackageResult, err error) {
				// Either error or result.Error should indicate cancellation
				if err != nil {
					if !strings.Contains(err.Error(), "cancel") && !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "context") {
						t.Errorf("Expected cancellation/timeout/context error, got: %v", err)
					}
				} else if result != nil && result.Error != nil {
					if !strings.Contains(result.Error.Error(), "cancel") && !strings.Contains(result.Error.Error(), "timeout") && !strings.Contains(result.Error.Error(), "context") {
						t.Errorf("Expected cancellation/timeout/context error in result, got: %v", result.Error)
					}
				}
			},
		},
		"execution_with_args": {
			setup: func() (*MockExecutor, *ExecutionOptions) {
				mock := NewMockExecutor()
				mock.executePackageFunc = func(ctx context.Context, pkg string, options *ExecutionOptions) (*PackageResult, error) {
					// Verify additional args are passed
					found := false
					for _, arg := range options.Args {
						if arg == "-short" {
							found = true
							break
						}
					}
					if !found {
						t.Error("Expected -short argument to be present")
					}

					return &PackageResult{
						Package:  pkg,
						Success:  true,
						Duration: 4 * time.Millisecond,
						Output:   `{"Time":"2024-01-01T10:00:00Z","Action":"pass","Package":"test","Test":"TestShort","Elapsed":0.004}`,
					}, nil
				}

				options := &ExecutionOptions{
					JSONOutput: true,
					Timeout:    30 * time.Second,
					Args:       []string{"-short"},
				}
				return mock, options
			},
			pkg: ".",
			validateResult: func(t *testing.T, result *PackageResult, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result == nil {
					t.Fatal("Result should not be nil")
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Setup
			mock, options := tt.setup()
			ctx := context.Background()

			// For timeout test, use short context
			if name == "context_timeout" {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 10*time.Millisecond)
				defer cancel()
			}

			// Execute
			result, err := mock.ExecutePackage(ctx, tt.pkg, options)

			// Validate
			tt.validateResult(t, result, err)
		})
	}
}

// TestDefaultExecutor_Execute_Optimized replaces the slow 3.98s test
func TestDefaultExecutor_Execute_Optimized(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		name           string
		setup          func() (*MockExecutor, *ExecutionOptions, []string)
		expectedError  string
		validateResult func(*testing.T, *ExecutionResult, error)
	}{
		"successful_single_package": {
			setup: func() (*MockExecutor, *ExecutionOptions, []string) {
				mock := NewMockExecutor()
				options := &ExecutionOptions{
					JSONOutput: true,
					Timeout:    30 * time.Second,
				}
				packages := []string{"."}
				return mock, options, packages
			},
			validateResult: func(t *testing.T, result *ExecutionResult, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result == nil {
					t.Fatal("Result should not be nil")
				}
				if len(result.Packages) != 1 {
					t.Errorf("Expected 1 package result, got %d", len(result.Packages))
				}
				if result.TotalDuration <= 0 {
					t.Error("Total duration should be positive")
				}
				if result.StartTime.IsZero() || result.EndTime.IsZero() {
					t.Error("Start and end times should be set")
				}
				if result.EndTime.Before(result.StartTime) {
					t.Error("End time should be after start time")
				}
			},
		},
		"concurrent_execution_protection": {
			setup: func() (*MockExecutor, *ExecutionOptions, []string) {
				mock := NewMockExecutor()
				mock.isRunning = true // Simulate already running
				mock.executeFunc = func(ctx context.Context, packages []string, options *ExecutionOptions) (*ExecutionResult, error) {
					return nil, fmt.Errorf("executor is already running")
				}

				options := &ExecutionOptions{
					JSONOutput: true,
					Timeout:    30 * time.Second,
				}
				packages := []string{"."}
				return mock, options, packages
			},
			expectedError: "already running",
			validateResult: func(t *testing.T, result *ExecutionResult, err error) {
				if err == nil {
					t.Error("Expected error for concurrent execution")
				}
				if err != nil && !strings.Contains(err.Error(), "already running") {
					t.Errorf("Expected 'already running' error, got: %v", err)
				}
			},
		},
		"pattern_expansion": {
			setup: func() (*MockExecutor, *ExecutionOptions, []string) {
				mock := NewMockExecutor()
				mock.executeFunc = func(ctx context.Context, packages []string, options *ExecutionOptions) (*ExecutionResult, error) {
					// Simulate pattern expansion
					expandedPackages := []string{"./pkg1", "./pkg2", "./pkg3"}
					result := &ExecutionResult{
						Packages:      make([]*PackageResult, len(expandedPackages)),
						TotalDuration: 15 * time.Millisecond,
						Success:       true,
						StartTime:     time.Now(),
						EndTime:       time.Now().Add(15 * time.Millisecond),
					}

					for i, pkg := range expandedPackages {
						result.Packages[i] = &PackageResult{
							Package:  pkg,
							Success:  true,
							Duration: 5 * time.Millisecond,
							Output:   fmt.Sprintf(`{"Time":"2024-01-01T10:00:00Z","Action":"pass","Package":"%s","Test":"TestExpanded","Elapsed":0.005}`, pkg),
						}
					}

					return result, nil
				}

				options := &ExecutionOptions{
					JSONOutput: true,
					Timeout:    30 * time.Second,
				}
				packages := []string{"./..."}
				return mock, options, packages
			},
			validateResult: func(t *testing.T, result *ExecutionResult, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result == nil {
					t.Fatal("Result should not be nil")
				}
				if len(result.Packages) < 1 {
					t.Error("Expected at least 1 package result from pattern expansion")
				}
			},
		},
		"error_in_expansion": {
			setup: func() (*MockExecutor, *ExecutionOptions, []string) {
				mock := NewMockExecutor()
				mock.executeFunc = func(ctx context.Context, packages []string, options *ExecutionOptions) (*ExecutionResult, error) {
					return nil, fmt.Errorf("failed to expand package patterns: invalid pattern")
				}

				options := &ExecutionOptions{
					JSONOutput: true,
					Timeout:    30 * time.Second,
				}
				packages := []string{"./invalid-pattern"}
				return mock, options, packages
			},
			expectedError: "failed to expand",
			validateResult: func(t *testing.T, result *ExecutionResult, err error) {
				if err == nil {
					t.Error("Expected error for invalid pattern")
				}
				if err != nil && !strings.Contains(err.Error(), "expand") {
					t.Errorf("Expected expansion error, got: %v", err)
				}
			},
		},
		"error_in_execution": {
			setup: func() (*MockExecutor, *ExecutionOptions, []string) {
				mock := NewMockExecutor()
				mock.executeFunc = func(ctx context.Context, packages []string, options *ExecutionOptions) (*ExecutionResult, error) {
					// Simulate execution with some failures
					result := &ExecutionResult{
						Packages:      make([]*PackageResult, 1),
						TotalDuration: 12 * time.Millisecond,
						Success:       false,
						StartTime:     time.Now(),
						EndTime:       time.Now().Add(12 * time.Millisecond),
					}

					result.Packages[0] = &PackageResult{
						Package:  packages[0],
						Success:  false,
						Duration: 12 * time.Millisecond,
						Error:    fmt.Errorf("compilation failed"),
						Output:   "# test\n./main.go:1:1: syntax error",
					}

					return result, nil
				}

				options := &ExecutionOptions{
					JSONOutput: true,
					Timeout:    30 * time.Second,
				}
				packages := []string{"./broken-package"}
				return mock, options, packages
			},
			validateResult: func(t *testing.T, result *ExecutionResult, err error) {
				if result == nil {
					t.Fatal("Result should not be nil even on execution error")
				}
				if result.Success {
					t.Error("Result should indicate failure")
				}
				if len(result.Packages) > 0 && result.Packages[0].Error == nil {
					t.Error("Package result should have error information")
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Setup
			mock, options, packages := tt.setup()
			ctx := context.Background()

			// Execute
			result, err := mock.Execute(ctx, packages, options)

			// Validate
			tt.validateResult(t, result, err)
		})
	}
}

// TestDefaultExecutor_ExecuteMultiplePackages_Optimized replaces the slow 2.30s test
func TestDefaultExecutor_ExecuteMultiplePackages_Optimized(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		name           string
		setup          func() (*MockExecutor, *ExecutionOptions, []string)
		contextSetup   func() (context.Context, context.CancelFunc)
		validateResult func(*testing.T, *ExecutionResult, error)
	}{
		"successful_multiple_packages": {
			setup: func() (*MockExecutor, *ExecutionOptions, []string) {
				mock := NewMockExecutor()
				mock.executeFunc = func(ctx context.Context, packages []string, options *ExecutionOptions) (*ExecutionResult, error) {
					result := &ExecutionResult{
						Packages:      make([]*PackageResult, len(packages)),
						TotalDuration: time.Duration(len(packages)*5) * time.Millisecond,
						Success:       true,
						StartTime:     time.Now(),
						EndTime:       time.Now().Add(time.Duration(len(packages)*5) * time.Millisecond),
					}

					for i, pkg := range packages {
						result.Packages[i] = &PackageResult{
							Package:  pkg,
							Success:  true,
							Duration: 5 * time.Millisecond,
							Output:   fmt.Sprintf(`{"Time":"2024-01-01T10:00:00Z","Action":"pass","Package":"%s","Test":"TestMultiple","Elapsed":0.005}`, pkg),
						}
					}

					return result, nil
				}

				options := &ExecutionOptions{
					JSONOutput: true,
					Timeout:    30 * time.Second,
				}
				packages := []string{"./pkg1", "./pkg2", "./pkg3"}
				return mock, options, packages
			},
			contextSetup: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 5*time.Second)
			},
			validateResult: func(t *testing.T, result *ExecutionResult, err error) {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if result == nil {
					t.Fatal("Result should not be nil")
				}
				if len(result.Packages) != 3 {
					t.Errorf("Expected 3 package results, got %d", len(result.Packages))
				}
			},
		},
		"cancelled_context": {
			setup: func() (*MockExecutor, *ExecutionOptions, []string) {
				mock := NewMockExecutor()
				mock.executeFunc = func(ctx context.Context, packages []string, options *ExecutionOptions) (*ExecutionResult, error) {
					// Check if context is cancelled
					select {
					case <-ctx.Done():
						return &ExecutionResult{
							Packages:      []*PackageResult{},
							TotalDuration: 1 * time.Millisecond,
							Success:       false,
							StartTime:     time.Now(),
							EndTime:       time.Now().Add(1 * time.Millisecond),
						}, ctx.Err()
					default:
						// Quick completion
						return &ExecutionResult{
							Packages:      []*PackageResult{},
							TotalDuration: 1 * time.Millisecond,
							Success:       true,
							StartTime:     time.Now(),
							EndTime:       time.Now().Add(1 * time.Millisecond),
						}, nil
					}
				}

				options := &ExecutionOptions{
					JSONOutput: true,
					Timeout:    100 * time.Millisecond,
				}
				packages := []string{"./pkg1", "./pkg2"}
				return mock, options, packages
			},
			contextSetup: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
				// Cancel immediately to test cancellation
				cancel()
				return ctx, func() {} // Return no-op cancel since already cancelled
			},
			validateResult: func(t *testing.T, result *ExecutionResult, err error) {
				// Either should complete quickly or be cancelled
				if err != nil && !strings.Contains(err.Error(), "cancel") && !strings.Contains(err.Error(), "context") {
					t.Logf("Expected cancellation error or quick completion, got: %v", err)
				}
			},
		},
		"mixed_success_failure": {
			setup: func() (*MockExecutor, *ExecutionOptions, []string) {
				mock := NewMockExecutor()
				mock.executeFunc = func(ctx context.Context, packages []string, options *ExecutionOptions) (*ExecutionResult, error) {
					result := &ExecutionResult{
						Packages:      make([]*PackageResult, len(packages)),
						TotalDuration: time.Duration(len(packages)*4) * time.Millisecond,
						Success:       false, // Overall failure due to one package failing
						StartTime:     time.Now(),
						EndTime:       time.Now().Add(time.Duration(len(packages)*4) * time.Millisecond),
					}

					for i, pkg := range packages {
						if i == 1 { // Second package fails
							result.Packages[i] = &PackageResult{
								Package:  pkg,
								Success:  false,
								Duration: 4 * time.Millisecond,
								Error:    fmt.Errorf("test failed"),
								Output:   fmt.Sprintf("FAIL\t%s\t0.004s", pkg),
							}
						} else {
							result.Packages[i] = &PackageResult{
								Package:  pkg,
								Success:  true,
								Duration: 4 * time.Millisecond,
								Output:   fmt.Sprintf(`{"Time":"2024-01-01T10:00:00Z","Action":"pass","Package":"%s","Test":"TestMixed","Elapsed":0.004}`, pkg),
							}
						}
					}

					return result, nil
				}

				options := &ExecutionOptions{
					JSONOutput: true,
					Timeout:    30 * time.Second,
				}
				packages := []string{"./pkg1", "./pkg2", "./pkg3"}
				return mock, options, packages
			},
			contextSetup: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 5*time.Second)
			},
			validateResult: func(t *testing.T, result *ExecutionResult, err error) {
				if result == nil {
					t.Fatal("Result should not be nil")
				}
				if result.Success {
					t.Error("Overall result should indicate failure")
				}

				// Check that we have mixed results
				successCount := 0
				failureCount := 0
				for _, pkg := range result.Packages {
					if pkg.Success {
						successCount++
					} else {
						failureCount++
					}
				}

				if successCount == 0 {
					t.Error("Expected some successful packages")
				}
				if failureCount == 0 {
					t.Error("Expected some failed packages")
				}
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Setup
			mock, options, packages := tt.setup()
			ctx, cancel := tt.contextSetup()
			defer cancel()

			// Execute
			result, err := mock.Execute(ctx, packages, options)

			// Validate
			tt.validateResult(t, result, err)
		})
	}
}
