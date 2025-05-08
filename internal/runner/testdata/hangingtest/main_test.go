package hangingtest

import (
	"sync"
	"testing"
)

func TestHanging(t *testing.T) {
	// This test will hang indefinitely until timeout
	var wg sync.WaitGroup
	wg.Add(1)
	
	// Create a deadlock - no one will ever call wg.Done()
	wg.Wait()
	
	t.Log("This should never be reached")
}
