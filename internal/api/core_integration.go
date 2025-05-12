package api

import (
	"sync"
)

type CoreEngine interface {
	SendUpdate(update string)
	ReceiveCommand(cmd string) error
}

type CoreAdapter struct {
	core      CoreEngine
	updatesCh chan string
	commandsCh chan string
	errCh     chan error
	wg        sync.WaitGroup
}

func NewCoreAdapter(core CoreEngine) *CoreAdapter {
	a := &CoreAdapter{
		core:      core,
		updatesCh: make(chan string, 8),
		commandsCh: make(chan string, 8),
		errCh:     make(chan error, 4),
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

func (a *CoreAdapter) SendUpdate(update string) {
	a.updatesCh <- update
}

func (a *CoreAdapter) SendCommand(cmd string) error {
	a.commandsCh <- cmd
	select {
	case err := <-a.errCh:
		return err
	default:
		return nil
	}
}

func (a *CoreAdapter) Close() {
	close(a.updatesCh)
	close(a.commandsCh)
	a.wg.Wait()
}
