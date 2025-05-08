package timeouttest

import (
	"testing"
	"time"
)

func TestLongRunning(t *testing.T) {
	// This test should run longer than our test timeout
	time.Sleep(10 * time.Second)
	t.Log("This should not be reached due to timeout")
}
