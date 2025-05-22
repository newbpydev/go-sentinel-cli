package demo

import (
	"fmt"
	"time"

	"github.com/newbpydev/go-sentinel/internal/cli"
)

// createSampleTestResult creates a sample test result for demonstration
func createSampleTestResult() *cli.TestResult {
	return &cli.TestResult{
		Name:     "TestSampleFunction",
		Status:   cli.StatusPassed,
		Duration: 50 * time.Millisecond,
		Package:  "github.com/newbpydev/go-sentinel/pkg/example",
		Test:     "TestSampleFunction",
		Output:   "PASS: TestSampleFunction (0.05s)",
		Subtests: []*cli.TestResult{
			{
				Name:     "TestSampleFunction/subtest_case_1",
				Status:   cli.StatusPassed,
				Duration: 20 * time.Millisecond,
				Package:  "github.com/newbpydev/go-sentinel/pkg/example",
				Test:     "TestSampleFunction/subtest_case_1",
				Parent:   "TestSampleFunction",
				Output:   "PASS: subtest_case_1 (0.02s)",
			},
			{
				Name:     "TestSampleFunction/subtest_case_2",
				Status:   cli.StatusFailed,
				Duration: 15 * time.Millisecond,
				Package:  "github.com/newbpydev/go-sentinel/pkg/example",
				Test:     "TestSampleFunction/subtest_case_2",
				Parent:   "TestSampleFunction",
				Output:   "FAIL: subtest_case_2 (0.015s)",
				Error: &cli.TestError{
					Message: "Expected 5, got 10",
					Type:    "AssertionError",
					Location: &cli.SourceLocation{
						File: "example_test.go",
						Line: 42,
					},
				},
			},
		},
	}
}

// createSampleTestSuite creates a sample test suite for demonstration
func createSampleTestSuite() *cli.TestSuite {
	suite := &cli.TestSuite{
		FilePath:     "github.com/newbpydev/go-sentinel/pkg/example/example_test.go",
		Duration:     100 * time.Millisecond,
		MemoryUsage:  10 * 1024 * 1024, // 10 MB
		TestCount:    3,
		PassedCount:  2,
		FailedCount:  1,
		SkippedCount: 0,
	}

	// Add the tests to the suite
	suite.Tests = append(suite.Tests, createSampleTestResult())

	return suite
}

// createMockTestSuites creates mock test suites that match the Vitest screenshot
func createMockTestSuites() []*cli.TestSuite {
	var suites []*cli.TestSuite

	// Create settings.test.ts suite (all passed)
	settingsSuite := &cli.TestSuite{
		FilePath:     "test/settings.test.ts",
		TestCount:    12,
		PassedCount:  12,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     119 * time.Millisecond,
		MemoryUsage:  33 * 1024 * 1024, // 33 MB
	}

	// Create mock test results for settings tests
	for i := 1; i <= 12; i++ {
		test := &cli.TestResult{
			Name:     fmt.Sprintf("SettingsTest%d", i),
			Status:   cli.StatusPassed,
			Duration: 10 * time.Millisecond,
			Package:  "test",
			Test:     fmt.Sprintf("SettingsTest%d", i),
		}
		settingsSuite.Tests = append(settingsSuite.Tests, test)
	}

	// Create websocket.test.ts suite (all failed)
	websocketSuite := &cli.TestSuite{
		FilePath:     "test/websocket.test.ts",
		TestCount:    8,
		PassedCount:  0,
		FailedCount:  8,
		SkippedCount: 0,
		Duration:     21 * time.Millisecond,
		MemoryUsage:  32 * 1024 * 1024, // 32 MB
	}

	// Create mock test results for websocket tests (failing tests)
	failingTests := []struct {
		name    string
		message string
		time    time.Duration
	}{
		{"WebSocketClient - connect method - should create a WebSocket with the given URL", "wsClient.connect is not a function", 8 * time.Millisecond},
		{"WebSocketClient - event handlers - should register open event handlers", "wsClient.connect is not a function", 1 * time.Millisecond},
		{"WebSocketClient - event handlers - should register close event handlers", "wsClient.connect is not a function", 1 * time.Millisecond},
		{"WebSocketClient - event handlers - should register and handle message events", "wsClient.connect is not a function", 1 * time.Millisecond},
		{"WebSocketClient - event handlers - should register error event handlers", "wsClient.connect is not a function", 3 * time.Millisecond},
		{"WebSocketClient - send method - should send JSON-stringified data when socket is open", "wsClient.connect is not a function", 1 * time.Millisecond},
		{"WebSocketClient - send method - should not send data when socket is not open", "wsClient.connect is not a function", 1 * time.Millisecond},
		{"WebSocketClient - disconnect method - should close the WebSocket connection", "wsClient.connect is not a function", 1 * time.Millisecond},
	}

	for _, ft := range failingTests {
		test := &cli.TestResult{
			Name:     ft.name,
			Status:   cli.StatusFailed,
			Duration: ft.time,
			Package:  "test",
			Test:     ft.name,
			Error: &cli.TestError{
				Message: ft.message,
				Type:    "TypeError",
			},
		}
		websocketSuite.Tests = append(websocketSuite.Tests, test)
	}

	// Create toast.test.ts suite (all passed)
	toastSuite := &cli.TestSuite{
		FilePath:     "test/toast.test.ts",
		TestCount:    8,
		PassedCount:  8,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     93 * time.Millisecond,
		MemoryUsage:  34 * 1024 * 1024, // 34 MB
	}

	// Create mock test results for toast tests
	for i := 1; i <= 8; i++ {
		test := &cli.TestResult{
			Name:     fmt.Sprintf("ToastTest%d", i),
			Status:   cli.StatusPassed,
			Duration: 10 * time.Millisecond,
			Package:  "test",
			Test:     fmt.Sprintf("ToastTest%d", i),
		}
		toastSuite.Tests = append(toastSuite.Tests, test)
	}

	// Create main.test.ts suite (all passed)
	mainSuite := &cli.TestSuite{
		FilePath:     "test/main.test.ts",
		TestCount:    10,
		PassedCount:  10,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     382 * time.Millisecond,
		MemoryUsage:  36 * 1024 * 1024, // 36 MB
	}

	// Create mock test results for main tests
	for i := 1; i <= 10; i++ {
		test := &cli.TestResult{
			Name:     fmt.Sprintf("MainTest%d", i),
			Status:   cli.StatusPassed,
			Duration: 30 * time.Millisecond,
			Package:  "test",
			Test:     fmt.Sprintf("MainTest%d", i),
		}
		mainSuite.Tests = append(mainSuite.Tests, test)
	}

	// Create coverage.test.ts suite (all passed)
	coverageSuite := &cli.TestSuite{
		FilePath:     "test/coverage.test.ts",
		TestCount:    20,
		PassedCount:  20,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     313 * time.Millisecond,
		MemoryUsage:  35 * 1024 * 1024, // 35 MB
	}

	// Create mock test results for coverage tests
	for i := 1; i <= 20; i++ {
		test := &cli.TestResult{
			Name:     fmt.Sprintf("CoverageTest%d", i),
			Status:   cli.StatusPassed,
			Duration: 15 * time.Millisecond,
			Package:  "test",
			Test:     fmt.Sprintf("CoverageTest%d", i),
		}
		coverageSuite.Tests = append(coverageSuite.Tests, test)
	}

	// Create utils/websocket.test.ts suite (all passed)
	utilsWebsocketSuite := &cli.TestSuite{
		FilePath:     "test/utils/websocket.test.ts",
		TestCount:    10,
		PassedCount:  10,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     24 * time.Millisecond,
		MemoryUsage:  40 * 1024 * 1024, // 40 MB
	}

	// Create mock test results for utils/websocket tests
	for i := 1; i <= 10; i++ {
		test := &cli.TestResult{
			Name:     fmt.Sprintf("UtilsWebSocketTest%d", i),
			Status:   cli.StatusPassed,
			Duration: 2 * time.Millisecond,
			Package:  "test/utils",
			Test:     fmt.Sprintf("UtilsWebSocketTest%d", i),
		}
		utilsWebsocketSuite.Tests = append(utilsWebsocketSuite.Tests, test)
	}

	// Create example.test.ts suite (all passed)
	exampleSuite := &cli.TestSuite{
		FilePath:     "test/example.test.ts",
		TestCount:    2,
		PassedCount:  2,
		FailedCount:  0,
		SkippedCount: 0,
		Duration:     14 * time.Millisecond,
		MemoryUsage:  39 * 1024 * 1024, // 39 MB
	}

	// Create mock test results for example tests
	for i := 1; i <= 2; i++ {
		test := &cli.TestResult{
			Name:     fmt.Sprintf("ExampleTest%d", i),
			Status:   cli.StatusPassed,
			Duration: 7 * time.Millisecond,
			Package:  "test",
			Test:     fmt.Sprintf("ExampleTest%d", i),
		}
		exampleSuite.Tests = append(exampleSuite.Tests, test)
	}

	// Add all suites to the result
	suites = append(suites,
		settingsSuite,
		websocketSuite,
		toastSuite,
		mainSuite,
		coverageSuite,
		utilsWebsocketSuite,
		exampleSuite,
	)

	return suites
}

// getMockFailedTests extracts all failed tests from mock suites
func getMockFailedTests(suites []*cli.TestSuite) []*cli.TestResult {
	var failedTests []*cli.TestResult

	for _, suite := range suites {
		for _, test := range suite.Tests {
			if test.Status == cli.StatusFailed {
				// Set the filepath for better display
				if test.Error != nil && test.Error.Location == nil {
					test.Error.Location = &cli.SourceLocation{
						File: suite.FilePath,
						Line: 42, // Mock line number
					}
				}

				failedTests = append(failedTests, test)
			}
		}
	}

	return failedTests
}

// createMockFailedTestsWithSourceContext creates sample failed tests with source context
func createMockFailedTestsWithSourceContext() []*cli.TestResult {
	// Create failed tests similar to the Vitest screenshot
	return []*cli.TestResult{
		{
			Name:   "WebSocketClient - connect method - should create a WebSocket with the given URL",
			Status: cli.StatusFailed,
			Error: &cli.TestError{
				Type:    "TypeError",
				Message: "wsClient.connect is not a function",
				Location: &cli.SourceLocation{
					File:   "test/websocket.test.ts",
					Line:   61,
					Column: 16,
				},
				SourceContext: []string{
					"it('should create a WebSocket with the given URL', () => {",
					"  // When",
					"  wsClient.connect(testUrl);",
					"",
					"  // Then",
				},
				HighlightedLine: 2, // 0-based index to the wsClient.connect line
			},
		},
		{
			Name:   "WebSocketClient - event handlers - should register open event handlers",
			Status: cli.StatusFailed,
			Error: &cli.TestError{
				Type:    "TypeError",
				Message: "wsClient.connect is not a function",
				Location: &cli.SourceLocation{
					File:   "test/websocket.test.ts",
					Line:   72,
					Column: 16,
				},
				SourceContext: []string{
					"  // Connect to ensure the socket is initialized, but don't await it",
					"  // since we're mocking the implementation",
					"  wsClient.connect(testUrl);",
					"",
					"  // Resolve the promise by simulating open event",
				},
				HighlightedLine: 2, // 0-based index to the wsClient.connect line
			},
		},
		{
			Name:   "WebSocketClient - event handlers - should register close event handlers",
			Status: cli.StatusFailed,
			Error: &cli.TestError{
				Type:    "TypeError",
				Message: "wsClient.connect is not a function",
				Location: &cli.SourceLocation{
					File:   "test/websocket.test.ts",
					Line:   72,
					Column: 16,
				},
			},
		},
		{
			Name:   "WebSocketClient - event handlers - should register and handle message events",
			Status: cli.StatusFailed,
			Error: &cli.TestError{
				Type:    "TypeError",
				Message: "wsClient.connect is not a function",
				Location: &cli.SourceLocation{
					File:   "test/websocket.test.ts",
					Line:   72,
					Column: 16,
				},
			},
		},
	}
}
