package api

import (
	"testing"
	"time"
)

type mockCoreEngine struct {
	updatesSent   []string
	commandsRcvd  []string
	failOnCommand bool
}

func (m *mockCoreEngine) SendUpdate(update string) {
	m.updatesSent = append(m.updatesSent, update)
}
func (m *mockCoreEngine) ReceiveCommand(cmd string) error {
	if m.failOnCommand {
		return &CoreError{"command failed"}
	}
	m.commandsRcvd = append(m.commandsRcvd, cmd)
	return nil
}

type CoreError struct{ msg string }

func (e *CoreError) Error() string { return e.msg }

func TestAPIReceivesUpdatesFromCore(t *testing.T) {
	core := &mockCoreEngine{}
	adapter := NewCoreAdapter(core)
	adapter.SendUpdate("test-update")
	deadline := time.Now().Add(100 * time.Millisecond)
	for time.Now().Before(deadline) {
		if len(core.updatesSent) == 1 && core.updatesSent[0] == "test-update" {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Errorf("API did not receive update from core: %+v", core.updatesSent)
}

func TestCommandsFromAPIPropagateToCore(t *testing.T) {
	core := &mockCoreEngine{}
	adapter := NewCoreAdapter(core)
	err := adapter.SendCommand("run-test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	deadline := time.Now().Add(100 * time.Millisecond)
	for time.Now().Before(deadline) {
		if len(core.commandsRcvd) == 1 && core.commandsRcvd[0] == "run-test" {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Errorf("Command not propagated: %+v", core.commandsRcvd)
}

func TestProperErrorHandlingBetweenComponents(t *testing.T) {
	core := &mockCoreEngine{failOnCommand: true}
	adapter := NewCoreAdapter(core)
	deadline := time.Now().Add(100 * time.Millisecond)
	var err error
	for time.Now().Before(deadline) {
		err = adapter.SendCommand("fail-cmd")
		if err != nil && err.Error() == "command failed" {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	if err == nil || err.Error() != "command failed" {
		t.Errorf("Expected error, got %v", err)
	}
}
