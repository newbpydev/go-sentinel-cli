package stress_tests

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"
)

// Test that hangs indefinitely (simulate deadlock)
func TestHangingTest(t *testing.T) {
	if os.Getenv("ENABLE_HANGING_TEST") == "" {
		t.Skip("Skipping hanging test - set ENABLE_HANGING_TEST to run")
	}

	// This would hang indefinitely
	select {} // Never completes
}

// Test with very long execution time
func TestVeryLongRunningTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long test in short mode")
	}

	// Simulate a test that takes a very long time
	time.Sleep(2 * time.Second)
	t.Log("Very long operation completed")
}

// Test with massive output
func TestMassiveOutput(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Logf("Output line %d: This is a very long log message that contains a lot of text and information about what is happening in this particular iteration of the loop", i)
	}

	// Generate a lot of output and then fail
	t.Error("Test failed after generating massive output")
}

// Test with nested subtests (deep nesting)
func TestDeeplyNestedSubtests(t *testing.T) {
	t.Run("level_1", func(t *testing.T) {
		t.Run("level_2", func(t *testing.T) {
			t.Run("level_3", func(t *testing.T) {
				t.Run("level_4_pass", func(t *testing.T) {
					t.Log("Deep nesting works")
				})
				t.Run("level_4_fail", func(t *testing.T) {
					t.Error("Deep nested failure")
				})
				t.Run("level_4_skip", func(t *testing.T) {
					t.Skip("Deep nested skip")
				})
			})
		})
	})
}

// Test with table-driven tests that have mixed results
func TestTableDrivenMixed(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
		shouldSkip  bool
	}{
		{"valid_input", "hello", "HELLO", false, false},
		{"empty_input", "", "", false, false},
		{"error_case", "error", "", true, false},
		{"skip_case", "skip", "", false, true},
		{"another_valid", "world", "WORLD", false, false},
		{"another_error", "fail", "", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldSkip {
				t.Skip("Skipping test case")
			}

			if tt.shouldError {
				t.Errorf("Simulated error for input: %s", tt.input)
				return
			}

			// Simulate string processing
			result := tt.input // In real test, this would be actual processing
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Test with goroutine leaks
func TestGoroutineLeaks(t *testing.T) {
	// Start multiple goroutines that don't properly clean up
	for i := 0; i < 10; i++ {
		go func(id int) {
			// Simulate work
			time.Sleep(10 * time.Millisecond)
			// Note: These goroutines don't have a way to be cancelled
		}(i)
	}

	t.Error("Test with potential goroutine leaks")
}

// Test with mutex deadlock scenario
func TestMutexDeadlock(t *testing.T) {
	if os.Getenv("ENABLE_DEADLOCK_TEST") == "" {
		t.Skip("Skipping deadlock test - set ENABLE_DEADLOCK_TEST to run")
	}

	var mu1, mu2 sync.Mutex
	done := make(chan bool, 2)

	// Goroutine 1: locks mu1 then mu2
	go func() {
		mu1.Lock()
		time.Sleep(10 * time.Millisecond)
		mu2.Lock()
		mu2.Unlock()
		mu1.Unlock()
		done <- true
	}()

	// Goroutine 2: locks mu2 then mu1 (potential deadlock)
	go func() {
		mu2.Lock()
		time.Sleep(10 * time.Millisecond)
		mu1.Lock()
		mu1.Unlock()
		mu2.Unlock()
		done <- true
	}()

	// Wait for completion with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	completed := 0
	for completed < 2 {
		select {
		case <-done:
			completed++
		case <-ctx.Done():
			t.Error("Deadlock detected - test timed out")
			return
		}
	}

	t.Log("No deadlock occurred")
}

// Test with multiple panics in subtests
func TestMultiplePanics(t *testing.T) {
	t.Run("panic_1", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Recovered from panic: %v", r)
			}
		}()

		panic("First panic")
	})

	t.Run("panic_2", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Recovered from panic: %v", r)
			}
		}()

		var slice []string
		_ = slice[100] // Out of bounds panic
	})

	t.Run("panic_3", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Recovered from panic: %v", r)
			}
		}()

		var m map[string]int
		m["key"] = 1 // Nil map panic
	})
}

// Test with file system operations that might fail
func TestFileSystemOperations(t *testing.T) {
	t.Run("create_temp_file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test_")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		t.Logf("Created temp file: %s", tmpFile.Name())
	})

	t.Run("read_nonexistent_file", func(t *testing.T) {
		_, err := os.ReadFile("/nonexistent/path/file.txt")
		if err == nil {
			t.Error("Expected error reading nonexistent file")
		} else {
			t.Logf("Got expected error: %v", err)
		}
	})

	t.Run("permission_denied", func(t *testing.T) {
		// Try to create file in root (should fail on most systems)
		_, err := os.Create("/root_file_test.txt")
		if err != nil {
			t.Logf("Got expected permission error: %v", err)
		} else {
			t.Log("Surprisingly, file creation succeeded")
		}
	})
}

// Test with environment variable dependencies
func TestEnvironmentDependencies(t *testing.T) {
	requiredEnv := "REQUIRED_TEST_ENV"

	if os.Getenv(requiredEnv) == "" {
		t.Errorf("Required environment variable %s is not set", requiredEnv)
	}

	// Test with different environment values
	testEnv := os.Getenv("TEST_MODE")
	switch testEnv {
	case "development":
		t.Log("Running in development mode")
	case "production":
		t.Log("Running in production mode")
	default:
		t.Errorf("Unknown TEST_MODE: %s", testEnv)
	}
}
