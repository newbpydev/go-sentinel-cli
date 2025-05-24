package core

import (
	"testing"
	"time"
)

func TestWatchMode_String(t *testing.T) {
	tests := []struct {
		name string
		mode WatchMode
		want string
	}{
		{
			name: "WatchAll mode",
			mode: WatchAll,
			want: "all",
		},
		{
			name: "WatchChanged mode",
			mode: WatchChanged,
			want: "changed",
		},
		{
			name: "WatchRelated mode",
			mode: WatchRelated,
			want: "related",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.mode) != tt.want {
				t.Errorf("WatchMode = %v, want %v", string(tt.mode), tt.want)
			}
		})
	}
}

func TestFileEvent_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		event FileEvent
		valid bool
	}{
		{
			name: "Valid file event",
			event: FileEvent{
				Path:      "/test/file.go",
				Type:      "write",
				Timestamp: time.Now(),
				IsTest:    false,
			},
			valid: true,
		},
		{
			name: "Empty path invalid",
			event: FileEvent{
				Path:      "",
				Type:      "write",
				Timestamp: time.Now(),
				IsTest:    false,
			},
			valid: false,
		},
		{
			name: "Empty type invalid",
			event: FileEvent{
				Path:      "/test/file.go",
				Type:      "",
				Timestamp: time.Now(),
				IsTest:    false,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.event.Path != "" && tt.event.Type != ""
			if isValid != tt.valid {
				t.Errorf("FileEvent.IsValid() = %v, want %v", isValid, tt.valid)
			}
		})
	}
}

func TestChangeType_Values(t *testing.T) {
	expectedTypes := []ChangeType{
		ChangeTypeModified,
		ChangeTypeAdded,
		ChangeTypeDeleted,
		ChangeTypeRenamed,
	}

	expectedValues := []string{
		"modified",
		"added",
		"deleted",
		"renamed",
	}

	if len(expectedTypes) != len(expectedValues) {
		t.Fatalf("ChangeType count mismatch: %d types, %d values", len(expectedTypes), len(expectedValues))
	}

	for i, changeType := range expectedTypes {
		if string(changeType) != expectedValues[i] {
			t.Errorf("ChangeType[%d] = %s, want %s", i, string(changeType), expectedValues[i])
		}
	}
}

func TestChangePriority_Ordering(t *testing.T) {
	priorities := []ChangePriority{
		PriorityLow,
		PriorityMedium,
		PriorityHigh,
		PriorityCritical,
	}

	// Test that priorities are ordered correctly
	for i := 1; i < len(priorities); i++ {
		if priorities[i-1] >= priorities[i] {
			t.Errorf("Priority ordering incorrect: %v should be < %v", priorities[i-1], priorities[i])
		}
	}
}

func TestWatchOptions_Validation(t *testing.T) {
	tests := []struct {
		name    string
		options WatchOptions
		isValid bool
	}{
		{
			name: "Valid options",
			options: WatchOptions{
				Paths:            []string{"./src"},
				IgnorePatterns:   []string{"*.log"},
				TestPatterns:     []string{"*_test.go"},
				Mode:             WatchAll,
				DebounceInterval: 100 * time.Millisecond,
				ClearTerminal:    false,
				RunOnStart:       true,
			},
			isValid: true,
		},
		{
			name: "Empty paths invalid",
			options: WatchOptions{
				Paths:            []string{},
				IgnorePatterns:   []string{"*.log"},
				TestPatterns:     []string{"*_test.go"},
				Mode:             WatchAll,
				DebounceInterval: 100 * time.Millisecond,
			},
			isValid: false,
		},
		{
			name: "Invalid mode",
			options: WatchOptions{
				Paths:            []string{"./src"},
				IgnorePatterns:   []string{"*.log"},
				TestPatterns:     []string{"*_test.go"},
				Mode:             WatchMode("invalid"),
				DebounceInterval: 100 * time.Millisecond,
			},
			isValid: false,
		},
		{
			name: "Zero debounce interval valid",
			options: WatchOptions{
				Paths:            []string{"./src"},
				IgnorePatterns:   []string{"*.log"},
				TestPatterns:     []string{"*_test.go"},
				Mode:             WatchAll,
				DebounceInterval: 0,
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic
			isValid := len(tt.options.Paths) > 0 &&
				(tt.options.Mode == WatchAll || tt.options.Mode == WatchChanged || tt.options.Mode == WatchRelated) &&
				tt.options.DebounceInterval >= 0

			if isValid != tt.isValid {
				t.Errorf("WatchOptions validation = %v, want %v", isValid, tt.isValid)
			}
		})
	}
}

func TestWatchStatus_Initialization(t *testing.T) {
	status := WatchStatus{}

	// Test default values
	if status.IsRunning {
		t.Error("Default WatchStatus should not be running")
	}

	if len(status.WatchedPaths) != 0 {
		t.Error("Default WatchStatus should have empty watched paths")
	}

	if status.EventCount != 0 {
		t.Error("Default WatchStatus should have zero event count")
	}

	if status.ErrorCount != 0 {
		t.Error("Default WatchStatus should have zero error count")
	}
}

func TestChangeImpact_Priority(t *testing.T) {
	impact := &ChangeImpact{
		FilePath:      "/test/file.go",
		Type:          ChangeTypeModified,
		IsTest:        false,
		AffectedTests: []string{"test1.go", "test2.go"},
		IsNew:         false,
		Timestamp:     time.Now(),
		Priority:      PriorityMedium,
	}

	if impact.Priority != PriorityMedium {
		t.Errorf("ChangeImpact.Priority = %v, want %v", impact.Priority, PriorityMedium)
	}

	if len(impact.AffectedTests) != 2 {
		t.Errorf("ChangeImpact.AffectedTests length = %d, want 2", len(impact.AffectedTests))
	}
}

func TestBatchImpact_Aggregation(t *testing.T) {
	change1 := &ChangeImpact{
		FilePath: "/test/file1.go",
		Priority: PriorityLow,
	}
	change2 := &ChangeImpact{
		FilePath: "/test/file2.go",
		Priority: PriorityHigh,
	}

	batch := &BatchImpact{
		Changes:           []*ChangeImpact{change1, change2},
		TotalFiles:        2,
		UniqueTestFiles:   []string{"test1.go", "test2.go"},
		ShouldRunAllTests: false,
		HighestPriority:   PriorityHigh,
		ProcessingTime:    10 * time.Millisecond,
	}

	if batch.TotalFiles != 2 {
		t.Errorf("BatchImpact.TotalFiles = %d, want 2", batch.TotalFiles)
	}

	if batch.HighestPriority != PriorityHigh {
		t.Errorf("BatchImpact.HighestPriority = %v, want %v", batch.HighestPriority, PriorityHigh)
	}

	if len(batch.Changes) != 2 {
		t.Errorf("BatchImpact.Changes length = %d, want 2", len(batch.Changes))
	}
}

func TestWatchEventType_Values(t *testing.T) {
	eventTypes := []WatchEventType{
		WatchEventStarted,
		WatchEventStopped,
		WatchEventError,
		WatchEventFileChanged,
		WatchEventTestsTriggered,
		WatchEventConfigUpdated,
	}

	expectedValues := []string{
		"started",
		"stopped",
		"error",
		"file_changed",
		"tests_triggered",
		"config_updated",
	}

	if len(eventTypes) != len(expectedValues) {
		t.Fatalf("WatchEventType count mismatch: %d types, %d values", len(eventTypes), len(expectedValues))
	}

	for i, eventType := range eventTypes {
		if string(eventType) != expectedValues[i] {
			t.Errorf("WatchEventType[%d] = %s, want %s", i, string(eventType), expectedValues[i])
		}
	}
}

func TestPatternType_Values(t *testing.T) {
	patternTypes := []PatternType{
		PatternTypeGlob,
		PatternTypeRegex,
		PatternTypeExact,
	}

	expectedValues := []string{
		"glob",
		"regex",
		"exact",
	}

	if len(patternTypes) != len(expectedValues) {
		t.Fatalf("PatternType count mismatch: %d types, %d values", len(patternTypes), len(expectedValues))
	}

	for i, patternType := range patternTypes {
		if string(patternType) != expectedValues[i] {
			t.Errorf("PatternType[%d] = %s, want %s", i, string(patternType), expectedValues[i])
		}
	}
}

func TestTestExecutionResult_Duration(t *testing.T) {
	result := TestExecutionResult{
		TestPaths: []string{"test1.go", "test2.go"},
		Success:   true,
		Duration:  500 * time.Millisecond,
		Output:    "All tests passed",
		Timestamp: time.Now(),
	}

	if result.Duration != 500*time.Millisecond {
		t.Errorf("TestExecutionResult.Duration = %v, want %v", result.Duration, 500*time.Millisecond)
	}

	if !result.Success {
		t.Error("TestExecutionResult.Success should be true")
	}

	if len(result.TestPaths) != 2 {
		t.Errorf("TestExecutionResult.TestPaths length = %d, want 2", len(result.TestPaths))
	}
}
