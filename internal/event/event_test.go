package event

import (
	"errors"
	"testing"
)

func TestErrorEvent_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "<nil>",
		},
		{
			name:     "simple error",
			err:      errors.New("test error"),
			expected: "test error",
		},
		{
			name:     "empty error",
			err:      errors.New(""),
			expected: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ev := ErrorEvent{Err: tc.err}
			if got := ev.Error(); got != tc.expected {
				t.Errorf("ErrorEvent.Error() = %q, want %q", got, tc.expected)
			}
		})
	}
}

func TestBoolPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected bool
	}{
		{
			name:     "true value",
			input:    true,
			expected: true,
		},
		{
			name:     "false value",
			input:    false,
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ptr := BoolPtr(tc.input)
			if ptr == nil {
				t.Fatal("BoolPtr returned nil")
			}
			if *ptr != tc.expected {
				t.Errorf("BoolPtr(%v) = %v, want %v", tc.input, *ptr, tc.expected)
			}
		})
	}
}

func TestTreeNode_Hierarchy(t *testing.T) {
	// Create a simple tree structure
	root := &TreeNode{
		Title:    "root",
		Level:    0,
		Expanded: true,
	}

	child1 := &TreeNode{
		Title:    "child1",
		Level:    1,
		Parent:   root,
		Expanded: false,
		Coverage: 0.75,
		Passed:   BoolPtr(true),
		Duration: 1.5,
	}

	child2 := &TreeNode{
		Title:    "child2",
		Level:    1,
		Parent:   root,
		Expanded: true,
		Coverage: 0.5,
		Passed:   BoolPtr(false),
		Duration: 2.0,
		Error:    "test failed",
	}

	root.Children = []*TreeNode{child1, child2}

	// Test tree structure
	if len(root.Children) != 2 {
		t.Errorf("root should have 2 children, got %d", len(root.Children))
	}

	// Test parent-child relationships
	for _, child := range root.Children {
		if child.Parent != root {
			t.Errorf("child %s parent should be root", child.Title)
		}
		if child.Level != root.Level+1 {
			t.Errorf("child %s level should be parent level + 1", child.Title)
		}
	}

	// Test node properties
	if !root.Expanded {
		t.Error("root should be expanded")
	}
	if child1.Expanded {
		t.Error("child1 should not be expanded")
	}
	if !child2.Expanded {
		t.Error("child2 should be expanded")
	}

	// Test test result properties
	if child1.Coverage != 0.75 {
		t.Errorf("child1 coverage = %v, want 0.75", child1.Coverage)
	}
	if !*child1.Passed {
		t.Error("child1 should be passed")
	}
	if child1.Duration != 1.5 {
		t.Errorf("child1 duration = %v, want 1.5", child1.Duration)
	}

	if child2.Coverage != 0.5 {
		t.Errorf("child2 coverage = %v, want 0.5", child2.Coverage)
	}
	if *child2.Passed {
		t.Error("child2 should be failed")
	}
	if child2.Duration != 2.0 {
		t.Errorf("child2 duration = %v, want 2.0", child2.Duration)
	}
	if child2.Error != "test failed" {
		t.Errorf("child2 error = %q, want 'test failed'", child2.Error)
	}
}

func TestTestResult_Fields(t *testing.T) {
	result := TestResult{
		Package: "pkg/foo",
		Test:    "TestBar",
		Passed:  true,
		Elapsed: 1.5,
		Output:  "test output",
		Error:   "",
		File:    "foo_test.go",
		Line:    42,
	}

	// Test all fields
	if result.Package != "pkg/foo" {
		t.Errorf("Package = %q, want 'pkg/foo'", result.Package)
	}
	if result.Test != "TestBar" {
		t.Errorf("Test = %q, want 'TestBar'", result.Test)
	}
	if !result.Passed {
		t.Error("Passed = false, want true")
	}
	if result.Elapsed != 1.5 {
		t.Errorf("Elapsed = %v, want 1.5", result.Elapsed)
	}
	if result.Output != "test output" {
		t.Errorf("Output = %q, want 'test output'", result.Output)
	}
	if result.Error != "" {
		t.Errorf("Error = %q, want ''", result.Error)
	}
	if result.File != "foo_test.go" {
		t.Errorf("File = %q, want 'foo_test.go'", result.File)
	}
	if result.Line != 42 {
		t.Errorf("Line = %d, want 42", result.Line)
	}
}

func TestFileEvent_Fields(t *testing.T) {
	event := FileEvent{
		Path: "path/to/file.go",
		Op:   "write",
	}

	if event.Path != "path/to/file.go" {
		t.Errorf("Path = %q, want 'path/to/file.go'", event.Path)
	}
	if event.Op != "write" {
		t.Errorf("Op = %q, want 'write'", event.Op)
	}
}

func TestRunnerEvent_Fields(t *testing.T) {
	results := []TestResult{
		{
			Package: "pkg/foo",
			Test:    "TestA",
			Passed:  true,
		},
		{
			Package: "pkg/foo",
			Test:    "TestB",
			Passed:  false,
		},
	}

	event := RunnerEvent{
		Package: "pkg/foo",
		Test:    "TestA",
		Status:  "completed",
		Results: results,
	}

	if event.Package != "pkg/foo" {
		t.Errorf("Package = %q, want 'pkg/foo'", event.Package)
	}
	if event.Test != "TestA" {
		t.Errorf("Test = %q, want 'TestA'", event.Test)
	}
	if event.Status != "completed" {
		t.Errorf("Status = %q, want 'completed'", event.Status)
	}
	if len(event.Results) != 2 {
		t.Errorf("len(Results) = %d, want 2", len(event.Results))
	}
}

func TestTestResult_Validation(t *testing.T) {
	tests := []struct {
		name    string
		result  TestResult
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid result",
			result: TestResult{
				Package: "pkg/foo",
				Test:    "TestBar",
				Passed:  true,
				Elapsed: 1.5,
			},
			wantErr: false,
		},
		{
			name: "missing package",
			result: TestResult{
				Test:    "TestBar",
				Passed:  true,
				Elapsed: 1.5,
			},
			wantErr: true,
			errMsg:  "package is required",
		},
		{
			name: "zero elapsed time",
			result: TestResult{
				Package: "pkg/foo",
				Test:    "TestBar",
				Passed:  true,
				Elapsed: 0,
			},
			wantErr: false,
		},
		{
			name: "negative elapsed time",
			result: TestResult{
				Package: "pkg/foo",
				Test:    "TestBar",
				Passed:  true,
				Elapsed: -1.5,
			},
			wantErr: true,
			errMsg:  "elapsed time cannot be negative",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.result.Validate()
			if tc.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if err.Error() != tc.errMsg {
					t.Errorf("expected error message %q, got %q", tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestFileEvent_Validation(t *testing.T) {
	tests := []struct {
		name    string
		event   FileEvent
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid event",
			event: FileEvent{
				Path: "path/to/file.go",
				Op:   "write",
			},
			wantErr: false,
		},
		{
			name: "missing path",
			event: FileEvent{
				Op: "write",
			},
			wantErr: true,
			errMsg:  "path is required",
		},
		{
			name: "missing operation",
			event: FileEvent{
				Path: "path/to/file.go",
			},
			wantErr: true,
			errMsg:  "operation is required",
		},
		{
			name: "invalid operation",
			event: FileEvent{
				Path: "path/to/file.go",
				Op:   "invalid",
			},
			wantErr: true,
			errMsg:  "invalid operation: invalid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.event.Validate()
			if tc.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if err.Error() != tc.errMsg {
					t.Errorf("expected error message %q, got %q", tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestRunnerEvent_Validation(t *testing.T) {
	tests := []struct {
		name    string
		event   RunnerEvent
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid event",
			event: RunnerEvent{
				Package: "pkg/foo",
				Test:    "TestBar",
				Status:  "completed",
				Results: []TestResult{
					{Package: "pkg/foo", Test: "TestBar", Passed: true},
				},
			},
			wantErr: false,
		},
		{
			name: "missing package",
			event: RunnerEvent{
				Test:   "TestBar",
				Status: "completed",
			},
			wantErr: true,
			errMsg:  "package is required",
		},
		{
			name: "missing status",
			event: RunnerEvent{
				Package: "pkg/foo",
				Test:    "TestBar",
			},
			wantErr: true,
			errMsg:  "status is required",
		},
		{
			name: "invalid status",
			event: RunnerEvent{
				Package: "pkg/foo",
				Test:    "TestBar",
				Status:  "invalid",
			},
			wantErr: true,
			errMsg:  "invalid status: invalid",
		},
		{
			name: "invalid test result",
			event: RunnerEvent{
				Package: "pkg/foo",
				Test:    "TestBar",
				Status:  "completed",
				Results: []TestResult{
					{Test: "TestBar", Passed: true}, // Missing package
				},
			},
			wantErr: true,
			errMsg:  "package is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.event.Validate()
			if tc.wantErr {
				if err == nil {
					t.Error("expected error but got nil")
				} else if err.Error() != tc.errMsg {
					t.Errorf("expected error message %q, got %q", tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestTreeNode_AddChild(t *testing.T) {
	root := &TreeNode{
		Title:    "root",
		Level:    0,
		Expanded: true,
	}

	child := &TreeNode{
		Title:    "child",
		Level:    1,
		Expanded: false,
	}

	// Add child to root
	root.Children = append(root.Children, child)
	child.Parent = root

	// Verify parent-child relationship
	if child.Parent != root {
		t.Error("child's parent should be root")
	}
	if len(root.Children) != 1 || root.Children[0] != child {
		t.Error("root should have child in its children")
	}
	if child.Level != root.Level+1 {
		t.Errorf("child level should be parent level + 1, got %d, want %d", child.Level, root.Level+1)
	}
}

func TestTreeNode_DeepNesting(t *testing.T) {
	root := &TreeNode{
		Title:    "root",
		Level:    0,
		Expanded: true,
	}

	// Create a chain of nodes: root -> a -> b -> c
	a := &TreeNode{Title: "a", Level: 1, Expanded: true}
	b := &TreeNode{Title: "b", Level: 2, Expanded: true}
	c := &TreeNode{Title: "c", Level: 3, Expanded: false}

	root.Children = append(root.Children, a)
	a.Parent = root
	a.Children = append(a.Children, b)
	b.Parent = a
	b.Children = append(b.Children, c)
	c.Parent = b

	// Verify the chain
	if len(root.Children) != 1 || root.Children[0] != a {
		t.Error("root should have 'a' as child")
	}
	if len(a.Children) != 1 || a.Children[0] != b {
		t.Error("'a' should have 'b' as child")
	}
	if len(b.Children) != 1 || b.Children[0] != c {
		t.Error("'b' should have 'c' as child")
	}
	if c.Parent != b || b.Parent != a || a.Parent != root {
		t.Error("parent chain should be intact")
	}
}

func TestTreeNode_TestResults(t *testing.T) {
	tests := []struct {
		name     string
		node     *TreeNode
		wantPass bool
	}{
		{
			name: "passing test",
			node: &TreeNode{
				Title:    "TestPass",
				Level:    1,
				Coverage: 0.85,
				Passed:   BoolPtr(true),
				Duration: 0.123,
			},
			wantPass: true,
		},
		{
			name: "failing test",
			node: &TreeNode{
				Title:    "TestFail",
				Level:    1,
				Coverage: 0.75,
				Passed:   BoolPtr(false),
				Duration: 0.456,
				Error:    "assertion failed",
			},
			wantPass: false,
		},
		{
			name: "skipped test",
			node: &TreeNode{
				Title:    "TestSkip",
				Level:    1,
				Passed:   nil,
				Duration: 0,
			},
			wantPass: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.node.Passed != nil && *tc.node.Passed != tc.wantPass {
				t.Errorf("node %s: got passed=%v, want %v", tc.node.Title, *tc.node.Passed, tc.wantPass)
			}
			if tc.node.Error != "" && *tc.node.Passed {
				t.Errorf("node %s: has error but marked as passed", tc.node.Title)
			}
		})
	}
}

func TestTreeNode_Coverage(t *testing.T) {
	tests := []struct {
		name         string
		node         *TreeNode
		wantCoverage float64
	}{
		{
			name: "full coverage",
			node: &TreeNode{
				Title:    "TestFull",
				Coverage: 1.0,
			},
			wantCoverage: 1.0,
		},
		{
			name: "partial coverage",
			node: &TreeNode{
				Title:    "TestPartial",
				Coverage: 0.75,
			},
			wantCoverage: 0.75,
		},
		{
			name: "no coverage",
			node: &TreeNode{
				Title:    "TestNone",
				Coverage: 0.0,
			},
			wantCoverage: 0.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.node.Coverage != tc.wantCoverage {
				t.Errorf("node %s: got coverage=%v, want %v", tc.node.Title, tc.node.Coverage, tc.wantCoverage)
			}
		})
	}
}
