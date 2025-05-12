package ui_test

import (
	"strings"
	"testing"

	"github.com/newbpydev/go-sentinel/internal/event"
	"github.com/newbpydev/go-sentinel/internal/ui"
)

// TestResultsMsgHandling tests the handling of TestResultsMsg
func TestResultsMsgHandling(t *testing.T) {
	// Create a simple tree for testing
	root := &ui.TreeNode{
		Title:    "root",
		Expanded: true,
		Children: []*ui.TreeNode{
			{Title: "pkg/test", Expanded: true, Children: []*ui.TreeNode{
				{Title: "TestA"},
			}},
		},
	}

	// Create initial model
	model := ui.NewTUITestExplorerModelWithNoExpansion(root)
	initialItemCount := len(model.Items)

	// Create test results and tree
	newRoot := &ui.TreeNode{
		Title:    "root",
		Expanded: true,
		Children: []*ui.TreeNode{
			{Title: "pkg/test", Expanded: true, Children: []*ui.TreeNode{
				{Title: "TestA"},
				{Title: "TestB"}, // Added a new test
			}},
		},
	}

	// Create test results message
	msg := ui.TestResultsMsg{
		Tree: newRoot,
		Results: []event.TestResult{
			{Package: "pkg/test", Test: "TestA"},
			{Package: "pkg/test", Test: "TestB"},
		},
	}

	// Apply the message
	updatedModel, _ := model.Update(msg)
	updatedTuiModel, ok := updatedModel.(*ui.TUITestExplorerModel)
	if !ok {
		t.Fatalf("Expected *TUITestExplorerModel, got %T", updatedModel)
	}

	// Verify model was updated with new results
	if len(updatedTuiModel.Items) <= initialItemCount {
		t.Errorf("Expected more items after update, got %d (was %d)",
			len(updatedTuiModel.Items), initialItemCount)
	}

	// Verify tree was updated
	foundNewTest := false
	for _, item := range updatedTuiModel.Items {
		if strings.Contains(item.FilterValue(), "TestB") {
			foundNewTest = true
			break
		}
	}
	if !foundNewTest {
		t.Error("New test 'TestB' not found in updated model")
	}
}

// TestTestsStartedMsgHandling tests the handling of TestsStartedMsg
func TestTestsStartedMsgHandling(t *testing.T) {
	// Create a simple tree for testing
	root := &ui.TreeNode{
		Title:    "root",
		Expanded: true,
		Children: []*ui.TreeNode{
			{Title: "pkg/test", Expanded: true, Children: []*ui.TreeNode{
				{Title: "TestA"},
			}},
		},
	}
	model := ui.NewTUITestExplorerModelWithNoExpansion(root)

	// Send tests started message
	msg := ui.TestsStartedMsg{
		Package: "pkg/test",
	}
	updatedModel, _ := model.Update(msg)
	updatedTuiModel, ok := updatedModel.(*ui.TUITestExplorerModel)
	if !ok {
		t.Fatalf("Expected *ui.TUITestExplorerModel, got %T", updatedModel)
	}

	// Verify model is in testing state
	if !updatedTuiModel.TestsRunning {
		t.Error("TestsRunning flag not set after TestsStartedMsg")
	}

	if updatedTuiModel.RunningPackage != "pkg/test" {
		t.Errorf("Expected RunningPackage to be 'pkg/test', got '%s'",
			updatedTuiModel.RunningPackage)
	}
}

// TestTestsCompletedMsgHandling tests the handling of TestsCompletedMsg
func TestTestsCompletedMsgHandling(t *testing.T) {
	// Create a simple model
	root := &ui.TreeNode{
		Title:    "root",
		Expanded: true,
		Children: []*ui.TreeNode{
			{Title: "pkg/test", Expanded: true, Children: []*ui.TreeNode{
				{Title: "TestA"},
			}},
		},
	}
	model := ui.NewTUITestExplorerModelWithNoExpansion(root)

	// Start a test run first
	model.TestsRunning = true
	model.RunningPackage = "pkg/test"

	// Then complete it
	msg := ui.TestsCompletedMsg{Success: true}
	updatedModel, _ := model.Update(msg)
	updatedTuiModel, ok := updatedModel.(*ui.TUITestExplorerModel)
	if !ok {
		t.Fatalf("Expected *TUITestExplorerModel, got %T", updatedModel)
	}

	// Verify model is no longer in testing state
	if updatedTuiModel.TestsRunning {
		t.Error("TestsRunning flag still set after TestsCompletedMsg")
	}

	// Verify success status is recorded
	if !updatedTuiModel.LastRunSuccess {
		t.Error("LastRunSuccess not set to true after successful test run")
	}
}

// TestRunTestsMsgHandling tests the handling of RunTestsMsg
func TestRunTestsMsgHandling(t *testing.T) {
	// Create a simple model
	root := &ui.TreeNode{
		Title:    "root",
		Expanded: true,
		Children: []*ui.TreeNode{
			{Title: "pkg/test", Expanded: true, Children: []*ui.TreeNode{
				{Title: "TestA"},
			}},
		},
	}
	model := ui.NewTUITestExplorerModelWithNoExpansion(root)

	// Send run tests message
	msg := ui.RunTestsMsg{Package: "pkg/test", Test: "TestA"}
	_, cmd := model.Update(msg)

	// Verify a command was returned (we can't easily test the command itself)
	if cmd == nil {
		t.Error("Expected command to be returned from RunTestsMsg")
	}
}

// TestFileChangedMsgHandling tests the handling of FileChangedMsg
func TestFileChangedMsgHandling(t *testing.T) {
	// Create a simple model
	root := &ui.TreeNode{
		Title:    "root",
		Expanded: true,
		Children: []*ui.TreeNode{
			{Title: "pkg/test", Expanded: true, Children: []*ui.TreeNode{
				{Title: "TestA"},
			}},
		},
	}
	model := ui.NewTUITestExplorerModelWithNoExpansion(root)

	// Send file changed message
	msg := ui.FileChangedMsg{Path: "pkg/test/test.go"}
	updatedModel, cmd := model.Update(msg)
	updatedTuiModel, ok := updatedModel.(*ui.TUITestExplorerModel)
	if !ok {
		t.Fatalf("Expected *TUITestExplorerModel, got %T", updatedModel)
	}

	// Verify the last changed file was recorded
	if updatedTuiModel.LastChangedFile != "pkg/test/test.go" {
		t.Errorf("Expected LastChangedFile to be 'pkg/test/test.go', got '%s'",
			updatedTuiModel.LastChangedFile)
	}

	// Should typically trigger a command to run tests
	if cmd == nil {
		t.Error("Expected command to be returned from FileChangedMsg")
	}
}
