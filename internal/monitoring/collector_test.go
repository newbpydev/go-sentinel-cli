package monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/events"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

func TestDefaultAppMetricsCollector_Implementation(t *testing.T) {
	// Create a default metrics collector
	eventBus := &MockEventBus{}
	config := &AppMonitoringConfig{
		Enabled:         true,
		MetricsPort:     8080,
		MetricsInterval: 1 * time.Second,
	}

	factory := NewDefaultAppMetricsCollectorFactory()
	if factory == nil {
		t.Fatal("Expected factory to exist, got nil")
	}

	collector := factory.CreateMetricsCollector(config, eventBus)
	if collector == nil {
		t.Fatal("Expected collector to be created, got nil")
	}

	// Test lifecycle
	ctx := context.Background()
	if err := collector.Start(ctx); err != nil {
		t.Fatalf("Expected no error starting collector, got: %v", err)
	}

	// Test metrics recording
	testResult := &models.TestResult{
		Name:     "test_example",
		Status:   models.TestStatusPassed,
		Duration: 100 * time.Millisecond,
	}
	collector.RecordTestExecution(testResult, 100*time.Millisecond)

	// Test metrics retrieval
	metrics := collector.GetMetrics()
	if metrics == nil {
		t.Fatal("Expected metrics to be returned, got nil")
	}
	if metrics.TestsExecuted != 1 {
		t.Errorf("Expected TestsExecuted to be 1, got %d", metrics.TestsExecuted)
	}
	if metrics.TestsSucceeded != 1 {
		t.Errorf("Expected TestsSucceeded to be 1, got %d", metrics.TestsSucceeded)
	}

	// Test shutdown
	if err := collector.Stop(ctx); err != nil {
		t.Fatalf("Expected no error stopping collector, got: %v", err)
	}
}

// Mock implementations for testing

type MockEventBus struct{}

func (m *MockEventBus) Publish(ctx context.Context, event events.Event) error {
	return nil
}

func (m *MockEventBus) PublishAsync(ctx context.Context, event events.Event) error {
	return nil
}

func (m *MockEventBus) Subscribe(eventType string, handler events.EventHandler) (events.Subscription, error) {
	return &MockSubscription{}, nil
}

func (m *MockEventBus) SubscribeWithFilter(filter events.EventFilter, handler events.EventHandler) (events.Subscription, error) {
	return &MockSubscription{}, nil
}

func (m *MockEventBus) Unsubscribe(subscription events.Subscription) error {
	return nil
}

func (m *MockEventBus) Close() error {
	return nil
}

func (m *MockEventBus) GetMetrics() *events.EventBusMetrics {
	return &events.EventBusMetrics{}
}

type MockSubscription struct{}

func (m *MockSubscription) ID() string        { return "mock-subscription" }
func (m *MockSubscription) EventType() string { return "mock-event" }
func (m *MockSubscription) IsActive() bool    { return true }
func (m *MockSubscription) Cancel() error     { return nil }
func (m *MockSubscription) GetStats() *events.SubscriptionStats {
	return &events.SubscriptionStats{}
}

// Real implementations are now in collector.go and dashboard.go
