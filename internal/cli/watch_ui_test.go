package cli

import (
	"bytes"
	"testing"
)

func TestWatchModeUIComponents(t *testing.T) {
	testCases := []struct {
		name     string
		function func(*TestWatcher, *bytes.Buffer)
		expected []string
	}{
		{
			name: "status display",
			function: func(w *TestWatcher, b *bytes.Buffer) {
				w.printStatus("Watching for changes...")
			},
			expected: []string{"Watching for changes..."},
		},
		{
			name: "watch info display",
			function: func(w *TestWatcher, b *bytes.Buffer) {
				w.printWatchInfo()
			},
			expected: []string{
				"Watch mode:",
				"Press 'a' to run all tests",
				"Press 'c' to run only changed tests",
				"Press 'r' to run related tests",
				"Press 'q' to quit",
			},
		},
		{
			name: "clear terminal",
			function: func(w *TestWatcher, b *bytes.Buffer) {
				w.clearTerminal()
			},
			expected: []string{"\033[2J\033[H"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buffer := &bytes.Buffer{}

			// Create a test watcher
			watcher := &TestWatcher{
				options: WatchOptions{
					Paths:  []string{"."},
					Mode:   WatchAll,
					Writer: buffer,
				},
				formatter: NewColorFormatter(false),
			}

			// Run the test function
			tc.function(watcher, buffer)

			// Check output
			output := buffer.String()
			for _, expected := range tc.expected {
				if !bytes.Contains(buffer.Bytes(), []byte(expected)) {
					t.Errorf("expected output to contain '%s', got '%s'", expected, output)
				}
			}
		})
	}
}

func TestWatchModeWithDifferentModes(t *testing.T) {
	testCases := []struct {
		name  string
		mode  WatchMode
		setup func(*testing.T) (*TestWatcher, *bytes.Buffer)
	}{
		{
			name: "all mode",
			mode: WatchAll,
			setup: func(t *testing.T) (*TestWatcher, *bytes.Buffer) {
				buffer := &bytes.Buffer{}
				watcher := &TestWatcher{
					options: WatchOptions{
						Paths:  []string{"."},
						Mode:   WatchAll,
						Writer: buffer,
					},
					formatter: NewColorFormatter(false),
				}
				return watcher, buffer
			},
		},
		{
			name: "changed mode",
			mode: WatchChanged,
			setup: func(t *testing.T) (*TestWatcher, *bytes.Buffer) {
				buffer := &bytes.Buffer{}
				watcher := &TestWatcher{
					options: WatchOptions{
						Paths:  []string{"."},
						Mode:   WatchChanged,
						Writer: buffer,
					},
					formatter: NewColorFormatter(false),
				}
				return watcher, buffer
			},
		},
		{
			name: "related mode",
			mode: WatchRelated,
			setup: func(t *testing.T) (*TestWatcher, *bytes.Buffer) {
				buffer := &bytes.Buffer{}
				watcher := &TestWatcher{
					options: WatchOptions{
						Paths:  []string{"."},
						Mode:   WatchRelated,
						Writer: buffer,
					},
					formatter: NewColorFormatter(false),
				}
				return watcher, buffer
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			watcher, buffer := tc.setup(t)

			// Check that the mode is set correctly
			if watcher.options.Mode != tc.mode {
				t.Errorf("expected mode %s, got %s", tc.mode, watcher.options.Mode)
			}

			// Print watch info
			watcher.printWatchInfo()

			// Check output contains the mode
			output := buffer.String()
			expectedModeText := "Watch mode: " + string(tc.mode)
			if !bytes.Contains(buffer.Bytes(), []byte(expectedModeText)) {
				t.Errorf("expected output to contain '%s', got '%s'", expectedModeText, output)
			}
		})
	}
}
