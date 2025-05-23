package main

import "testing"

func TestSimple(t *testing.T) {
	t.Log("This is a simple test")
	// Simple passing test
	if 1+1 != 2 {
		t.Error("Math doesn't work")
	}
}

// Modified
