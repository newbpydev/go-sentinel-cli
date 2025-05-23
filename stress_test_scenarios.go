package main

import (
	"fmt"
	"testing"
	"time"
)

// Test 1: Simple passing tests
func TestBasicPass(t *testing.T) {
	if 1+1 == 2 {
		t.Log("Basic math works")
	}
}

// Test 2: Simple failing test
func TestBasicFail(t *testing.T) {
	if 1+1 == 3 {
		t.Log("This should not happen")
	} else {
		t.Errorf("Expected 1+1 to equal 3, but got %d", 1+1)
	}
}

// Test 3: Skipped test
func TestSkipped(t *testing.T) {
	t.Skip("Skipping this test for demonstration purposes")
	t.Log("This should never be reached")
}

// Test 4: Test with subtests - mixed results
func TestMixedSubtests(t *testing.T) {
	t.Run("passing_subtest", func(t *testing.T) {
		if 2*2 == 4 {
			t.Log("Multiplication works")
		}
	})

	t.Run("failing_subtest", func(t *testing.T) {
		t.Errorf("This subtest is designed to fail")
	})

	t.Run("skipped_subtest", func(t *testing.T) {
		t.Skip("Skipping subtest")
	})

	t.Run("another_passing", func(t *testing.T) {
		if len("hello") == 5 {
			t.Log("String length correct")
		}
	})
}

// Test 5: Test that panics
func TestPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Test panicked: %v", r)
		}
	}()

	// This will cause a panic
	var slice []int
	_ = slice[10] // Index out of bounds
}

// Test 6: Test with timeout (long-running)
func TestTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping timeout test in short mode")
	}

	// Simulate a long-running operation
	time.Sleep(100 * time.Millisecond)
	t.Log("Long operation completed")
}

// Test 7: Test with lots of output
func TestVerboseOutput(t *testing.T) {
	for i := 0; i < 5; i++ {
		t.Logf("Log message %d: Processing item %d", i, i*2)
	}

	if false {
		t.Error("This should cause the test to fail with verbose output")
	}
}

// Test 8: Test with assertion failures
func TestAssertionFailures(t *testing.T) {
	tests := []struct {
		name       string
		input      int
		expected   int
		shouldFail bool
	}{
		{"pass_case", 5, 5, false},
		{"fail_case", 5, 10, true},
		{"another_pass", 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldFail {
				t.Errorf("Expected %d, got %d", tt.expected, tt.input)
			} else {
				if tt.input != tt.expected {
					t.Errorf("Expected %d, got %d", tt.expected, tt.input)
				}
			}
		})
	}
}

// Test 9: Test with complex error messages
func TestComplexErrors(t *testing.T) {
	user := struct {
		Name string
		Age  int
	}{
		Name: "John",
		Age:  25,
	}

	if user.Age < 18 {
		t.Errorf("User validation failed:\n  User: %+v\n  Expected age >= 18, got %d\n  Additional context: This is a complex error with multiple lines", user, user.Age)
	}

	// This will fail with a complex error
	expected := map[string]int{"a": 1, "b": 2}
	actual := map[string]int{"a": 1, "b": 3, "c": 4}

	if len(expected) != len(actual) {
		t.Errorf("Map comparison failed:\n  Expected: %v\n  Actual: %v\n  Difference: Expected %d keys, got %d keys", expected, actual, len(expected), len(actual))
	}
}

// Test 10: Test that fails with nil pointer
func TestNilPointer(t *testing.T) {
	var ptr *string
	if ptr != nil {
		t.Logf("Pointer value: %s", *ptr) // This would panic if ptr was nil
	} else {
		t.Error("Pointer is nil, this is a controlled failure")
	}
}

// Test 11: Test with goroutine issues (race condition simulation)
func TestConcurrency(t *testing.T) {
	counter := 0
	done := make(chan bool, 2)

	// Start two goroutines that increment counter
	go func() {
		for i := 0; i < 100; i++ {
			counter++
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			counter++
		}
		done <- true
	}()

	// Wait for both to complete
	<-done
	<-done

	// This might fail due to race condition (without proper synchronization)
	if counter != 200 {
		t.Errorf("Race condition detected: expected 200, got %d", counter)
	}
}

// Test 12: Test with conditional skips
func TestConditionalSkip(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// Simulate environment-specific test
	if true { // Simulate condition
		t.Skip("Skipping because test environment not available")
	}

	t.Log("This should not execute")
}

// Test 13: Test with benchmark (if run with -bench)
func BenchmarkExample(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("iteration %d", i)
	}
}

// Test 14: Test with cleanup that might fail
func TestWithCleanup(t *testing.T) {
	t.Cleanup(func() {
		// Simulate cleanup that might log or fail
		t.Log("Cleanup executed")
	})

	t.Log("Test body executed")
	t.Error("Failing after cleanup setup")
}
