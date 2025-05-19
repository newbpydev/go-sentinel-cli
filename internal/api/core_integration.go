package api

import (
	"sync"
)

// CoreEngine defines the interface for the core test execution engine.
// It provides methods for running tests and managing test execution.
type CoreEngine interface {
	SendUpdate(update string)
	ReceiveCommand(cmd string) error
}

// CoreAdapter provides an adapter layer between the API and the core engine.
// It handles communication and data transformation between the two layers.
type CoreAdapter struct {
	core       CoreEngine
	updatesCh  chan string
	commandsCh chan string
	errCh      chan error
	wg         sync.WaitGroup
}

// NewCoreAdapter creates a new CoreAdapter with the given core engine.
// It initializes all necessary channels and goroutines for communication.
func NewCoreAdapter(core CoreEngine) *CoreAdapter {
	a := &CoreAdapter{
		core:       core,
		updatesCh:  make(chan string, 8),
		commandsCh: make(chan string, 8),
		errCh:      make(chan error, 4),
	}
	a.start()
	return a
}

func (a *CoreAdapter) start() {
	a.wg.Add(2)
	go func() {
		defer a.wg.Done()
		for update := range a.updatesCh {
			a.core.SendUpdate(update)
		}
	}()
	go func() {
		defer a.wg.Done()
		for cmd := range a.commandsCh {
			err := a.core.ReceiveCommand(cmd)
			if err != nil {
				a.errCh <- err
			}
		}
	}()
}

// SendUpdate sends an update message to the core engine.
// This is used to notify the core engine about changes in test status or configuration.
func (a *CoreAdapter) SendUpdate(update string) {
	a.updatesCh <- update
}

// SendCommand sends a command to the core engine and waits for a response.
// Returns an error if the command fails or times out.
func (a *CoreAdapter) SendCommand(cmd string) error {
	a.commandsCh <- cmd
	select {
	case err := <-a.errCh:
		return err
	default:
		return nil
	}
}

// Close shuts down the adapter and releases all resources.
// This should be called when the adapter is no longer needed.
func (a *CoreAdapter) Close() {
	close(a.updatesCh)
	close(a.commandsCh)
	a.wg.Wait()
}
