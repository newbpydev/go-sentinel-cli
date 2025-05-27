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
			if tt.mode.String() != tt.want {
				t.Errorf("WatchMode.String() = %v, want %v", tt.mode.String(), tt.want)
			}
		})
	}
}

func TestWatchMode_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		mode  WatchMode
		valid bool
	}{
		{"WatchAll valid", WatchAll, true},
		{"WatchChanged valid", WatchChanged, true},
		{"WatchRelated valid", WatchRelated, true},
		{"Invalid mode", WatchMode("invalid"), false},
		{"Empty mode", WatchMode(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.mode.IsValid(); got != tt.valid {
				t.Errorf("WatchMode.IsValid() = %v, want %v", got, tt.valid)
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
		{
			name: "Both empty invalid",
			event: FileEvent{
				Path:      "",
				Type:      "",
				Timestamp: time.Now(),
				IsTest:    false,
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.event.IsValid(); got != tt.valid {
				t.Errorf("FileEvent.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestChangeType_String(t *testing.T) {
	tests := []struct {
		name       string
		changeType ChangeType
		want       string
	}{
		{"Modified", ChangeTypeModified, "modified"},
		{"Added", ChangeTypeAdded, "added"},
		{"Deleted", ChangeTypeDeleted, "deleted"},
		{"Renamed", ChangeTypeRenamed, "renamed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.changeType.String(); got != tt.want {
				t.Errorf("ChangeType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeType_IsValid(t *testing.T) {
	tests := []struct {
		name       string
		changeType ChangeType
		valid      bool
	}{
		{"Modified valid", ChangeTypeModified, true},
		{"Added valid", ChangeTypeAdded, true},
		{"Deleted valid", ChangeTypeDeleted, true},
		{"Renamed valid", ChangeTypeRenamed, true},
		{"Invalid type", ChangeType("invalid"), false},
		{"Empty type", ChangeType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.changeType.IsValid(); got != tt.valid {
				t.Errorf("ChangeType.IsValid() = %v, want %v", got, tt.valid)
			}
		})
	}
}

func TestChangePriority_GetPriorityLevel(t *testing.T) {
	tests := []struct {
		name     string
		priority ChangePriority
		want     int
	}{
		{"Low priority", PriorityLow, 0},
		{"Medium priority", PriorityMedium, 1},
		{"High priority", PriorityHigh, 2},
		{"Critical priority", PriorityCritical, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.priority.GetPriorityLevel(); got != tt.want {
				t.Errorf("ChangePriority.GetPriorityLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangePriority_String(t *testing.T) {
	tests := []struct {
		name     string
		priority ChangePriority
		want     string
	}{
		{"Low priority", PriorityLow, "low"},
		{"Medium priority", PriorityMedium, "medium"},
		{"High priority", PriorityHigh, "high"},
		{"Critical priority", PriorityCritical, "critical"},
		{"Unknown priority", ChangePriority(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.priority.String(); got != tt.want {
				t.Errorf("ChangePriority.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangePriority_IsValidPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority ChangePriority
		valid    bool
	}{
		{"Low valid", PriorityLow, true},
		{"Medium valid", PriorityMedium, true},
		{"High valid", PriorityHigh, true},
		{"Critical valid", PriorityCritical, true},
		{"Below range invalid", ChangePriority(-1), false},
		{"Above range invalid", ChangePriority(99), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.priority.IsValidPriority(); got != tt.valid {
				t.Errorf("ChangePriority.IsValidPriority() = %v, want %v", got, tt.valid)
			}
		})
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

func TestWatchOptions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		options WatchOptions
		wantErr bool
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
			wantErr: false,
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
			wantErr: true,
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
			wantErr: true,
		},
		{
			name: "Negative debounce interval invalid",
			options: WatchOptions{
				Paths:            []string{"./src"},
				IgnorePatterns:   []string{"*.log"},
				TestPatterns:     []string{"*_test.go"},
				Mode:             WatchAll,
				DebounceInterval: -100 * time.Millisecond,
			},
			wantErr: true,
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
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WatchOptions.Validate() error = %v, wantErr %v", err, tt.wantErr)
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

func TestChangeImpact_HasHighPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority ChangePriority
		want     bool
	}{
		{"Low priority not high", PriorityLow, false},
		{"Medium priority not high", PriorityMedium, false},
		{"High priority is high", PriorityHigh, true},
		{"Critical priority is high", PriorityCritical, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impact := &ChangeImpact{Priority: tt.priority}
			if got := impact.HasHighPriority(); got != tt.want {
				t.Errorf("ChangeImpact.HasHighPriority() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeImpact_GetTestCount(t *testing.T) {
	tests := []struct {
		name          string
		affectedTests []string
		want          int
	}{
		{"No tests", []string{}, 0},
		{"One test", []string{"test1.go"}, 1},
		{"Multiple tests", []string{"test1.go", "test2.go", "test3.go"}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impact := &ChangeImpact{AffectedTests: tt.affectedTests}
			if got := impact.GetTestCount(); got != tt.want {
				t.Errorf("ChangeImpact.GetTestCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChangeImpact_IsTestChange(t *testing.T) {
	tests := []struct {
		name   string
		isTest bool
		want   bool
	}{
		{"Test file change", true, true},
		{"Implementation file change", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			impact := &ChangeImpact{IsTest: tt.isTest}
			if got := impact.IsTestChange(); got != tt.want {
				t.Errorf("ChangeImpact.IsTestChange() = %v, want %v", got, tt.want)
			}
		})
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

func TestBatchImpact_CalculateHighestPriority(t *testing.T) {
	tests := []struct {
		name    string
		changes []*ChangeImpact
		want    ChangePriority
	}{
		{
			name: "Single low priority",
			changes: []*ChangeImpact{
				{Priority: PriorityLow},
			},
			want: PriorityLow,
		},
		{
			name: "Mixed priorities",
			changes: []*ChangeImpact{
				{Priority: PriorityLow},
				{Priority: PriorityHigh},
				{Priority: PriorityMedium},
			},
			want: PriorityHigh,
		},
		{
			name: "Critical priority highest",
			changes: []*ChangeImpact{
				{Priority: PriorityHigh},
				{Priority: PriorityCritical},
				{Priority: PriorityMedium},
			},
			want: PriorityCritical,
		},
		{
			name:    "Empty changes",
			changes: []*ChangeImpact{},
			want:    PriorityLow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batch := &BatchImpact{Changes: tt.changes}
			if got := batch.CalculateHighestPriority(); got != tt.want {
				t.Errorf("BatchImpact.CalculateHighestPriority() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatchImpact_GetUniqueTestCount(t *testing.T) {
	tests := []struct {
		name            string
		uniqueTestFiles []string
		want            int
	}{
		{"No tests", []string{}, 0},
		{"One test", []string{"test1.go"}, 1},
		{"Multiple tests", []string{"test1.go", "test2.go", "test3.go"}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batch := &BatchImpact{UniqueTestFiles: tt.uniqueTestFiles}
			if got := batch.GetUniqueTestCount(); got != tt.want {
				t.Errorf("BatchImpact.GetUniqueTestCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBatchImpact_HasCriticalChanges(t *testing.T) {
	tests := []struct {
		name    string
		changes []*ChangeImpact
		want    bool
	}{
		{
			name: "No critical changes",
			changes: []*ChangeImpact{
				{Priority: PriorityLow},
				{Priority: PriorityMedium},
				{Priority: PriorityHigh},
			},
			want: false,
		},
		{
			name: "Has critical changes",
			changes: []*ChangeImpact{
				{Priority: PriorityLow},
				{Priority: PriorityCritical},
				{Priority: PriorityMedium},
			},
			want: true,
		},
		{
			name:    "Empty changes",
			changes: []*ChangeImpact{},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batch := &BatchImpact{Changes: tt.changes}
			if got := batch.HasCriticalChanges(); got != tt.want {
				t.Errorf("BatchImpact.HasCriticalChanges() = %v, want %v", got, tt.want)
			}
		})
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

func TestWatchEventType_String(t *testing.T) {
	tests := []struct {
		name      string
		eventType WatchEventType
		want      string
	}{
		{"Started", WatchEventStarted, "started"},
		{"Stopped", WatchEventStopped, "stopped"},
		{"Error", WatchEventError, "error"},
		{"File changed", WatchEventFileChanged, "file_changed"},
		{"Tests triggered", WatchEventTestsTriggered, "tests_triggered"},
		{"Config updated", WatchEventConfigUpdated, "config_updated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.eventType.String(); got != tt.want {
				t.Errorf("WatchEventType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWatchEventType_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		eventType WatchEventType
		valid     bool
	}{
		{"Started valid", WatchEventStarted, true},
		{"Stopped valid", WatchEventStopped, true},
		{"Error valid", WatchEventError, true},
		{"File changed valid", WatchEventFileChanged, true},
		{"Tests triggered valid", WatchEventTestsTriggered, true},
		{"Config updated valid", WatchEventConfigUpdated, true},
		{"Invalid type", WatchEventType("invalid"), false},
		{"Empty type", WatchEventType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.eventType.IsValid(); got != tt.valid {
				t.Errorf("WatchEventType.IsValid() = %v, want %v", got, tt.valid)
			}
		})
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

func TestPatternType_String(t *testing.T) {
	tests := []struct {
		name        string
		patternType PatternType
		want        string
	}{
		{"Glob", PatternTypeGlob, "glob"},
		{"Regex", PatternTypeRegex, "regex"},
		{"Exact", PatternTypeExact, "exact"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.patternType.String(); got != tt.want {
				t.Errorf("PatternType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPatternType_IsValid(t *testing.T) {
	tests := []struct {
		name        string
		patternType PatternType
		valid       bool
	}{
		{"Glob valid", PatternTypeGlob, true},
		{"Regex valid", PatternTypeRegex, true},
		{"Exact valid", PatternTypeExact, true},
		{"Invalid type", PatternType("invalid"), false},
		{"Empty type", PatternType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.patternType.IsValid(); got != tt.valid {
				t.Errorf("PatternType.IsValid() = %v, want %v", got, tt.valid)
			}
		})
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

func TestTestExecutionResult_IsSuccessful(t *testing.T) {
	tests := []struct {
		name   string
		result TestExecutionResult
		want   bool
	}{
		{
			name: "Successful with no error",
			result: TestExecutionResult{
				Success:      true,
				ErrorMessage: "",
			},
			want: true,
		},
		{
			name: "Not successful",
			result: TestExecutionResult{
				Success:      false,
				ErrorMessage: "",
			},
			want: false,
		},
		{
			name: "Success but has error message",
			result: TestExecutionResult{
				Success:      true,
				ErrorMessage: "warning message",
			},
			want: false,
		},
		{
			name: "Not successful with error",
			result: TestExecutionResult{
				Success:      false,
				ErrorMessage: "test failed",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.IsSuccessful(); got != tt.want {
				t.Errorf("TestExecutionResult.IsSuccessful() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTestExecutionResult_GetTestCount(t *testing.T) {
	tests := []struct {
		name      string
		testPaths []string
		want      int
	}{
		{"No tests", []string{}, 0},
		{"One test", []string{"test1.go"}, 1},
		{"Multiple tests", []string{"test1.go", "test2.go", "test3.go"}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TestExecutionResult{TestPaths: tt.testPaths}
			if got := result.GetTestCount(); got != tt.want {
				t.Errorf("TestExecutionResult.GetTestCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTestExecutionResult_HasOutput(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{"No output", "", false},
		{"Has output", "test output", true},
		{"Whitespace output", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TestExecutionResult{Output: tt.output}
			if got := result.HasOutput(); got != tt.want {
				t.Errorf("TestExecutionResult.HasOutput() = %v, want %v", got, tt.want)
			}
		})
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

func TestFilePattern_Matches(t *testing.T) {
	tests := []struct {
		name    string
		pattern FilePattern
		path    string
		want    bool
	}{
		{
			name: "Exact match",
			pattern: FilePattern{
				Pattern: "/test/file.go",
				Type:    PatternTypeExact,
			},
			path: "/test/file.go",
			want: true,
		},
		{
			name: "Exact no match",
			pattern: FilePattern{
				Pattern: "/test/file.go",
				Type:    PatternTypeExact,
			},
			path: "/test/other.go",
			want: false,
		},
		{
			name: "Glob match",
			pattern: FilePattern{
				Pattern: "*test*",
				Type:    PatternTypeGlob,
			},
			path: "/some/test/file.go",
			want: true,
		},
		{
			name: "Glob no match",
			pattern: FilePattern{
				Pattern: "*test*",
				Type:    PatternTypeGlob,
			},
			path: "/some/other/file.go",
			want: false,
		},
		{
			name: "Regex match",
			pattern: FilePattern{
				Pattern: "test",
				Type:    PatternTypeRegex,
			},
			path: "/some/test/file.go",
			want: true,
		},
		{
			name: "Regex no match",
			pattern: FilePattern{
				Pattern: "test",
				Type:    PatternTypeRegex,
			},
			path: "/some/other/file.go",
			want: false,
		},
		{
			name: "Invalid pattern type",
			pattern: FilePattern{
				Pattern: "test",
				Type:    PatternType("invalid"),
			},
			path: "/some/test/file.go",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pattern.Matches(tt.path); got != tt.want {
				t.Errorf("FilePattern.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilePattern_IsRecursive(t *testing.T) {
	tests := []struct {
		name      string
		recursive bool
		want      bool
	}{
		{"Recursive pattern", true, true},
		{"Non-recursive pattern", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := FilePattern{Recursive: tt.recursive}
			if got := pattern.IsRecursive(); got != tt.want {
				t.Errorf("FilePattern.IsRecursive() = %v, want %v", got, tt.want)
			}
		})
	}
}
