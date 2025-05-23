package main

import "testing"

func TestPassing(t *testing.T) {
	if 2+2 == 4 {
		t.Log("Math works correctly")
	}
}

func TestFailing(t *testing.T) {
	if 2+2 == 5 {
		t.Log("This should not pass")
	} else {
		t.Error("Math is broken - 2+2 should equal 4, not 5")
	}
}

func TestAnother(t *testing.T) {
	result := 3 * 3
	if result != 9 {
		t.Errorf("Expected 3*3=9, got %d", result)
	}
}
