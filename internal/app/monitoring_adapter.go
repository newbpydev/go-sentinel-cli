package app

import (
	"context"
	"time"

	"github.com/newbpydev/go-sentinel/internal/monitoring"
	"github.com/newbpydev/go-sentinel/pkg/events"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// MonitoringAdapter provides backward compatibility for app package
// Wraps the new monitoring package with the old interface expected by app code
type MonitoringAdapter struct {
	collector monitoring.AppMetricsCollector
	dashboard monitoring.AppDashboard
}

// NewMonitoringAdapter creates a new monitoring adapter using the monitoring package
func NewMonitoringAdapter(eventBus events.EventBus) *MonitoringAdapter {
	// Create monitoring config
	config := monitoring.DefaultAppMonitoringConfig()

	// Create metrics collector using factory
	collectorFactory := monitoring.NewDefaultAppMetricsCollectorFactory()
	collector := collectorFactory.CreateMetricsCollector(config, eventBus)

	// Create dashboard using factory
	dashboardFactory := monitoring.NewDefaultAppDashboardFactory()
	dashboardConfig := monitoring.DefaultAppDashboardConfig()
	dashboard := dashboardFactory.CreateDashboard(collector, dashboardConfig)

	return &MonitoringAdapter{
		collector: collector,
		dashboard: dashboard,
	}
}

// Start initializes the monitoring system
func (ma *MonitoringAdapter) Start(ctx context.Context) error {
	// Start metrics collector first
	if err := ma.collector.Start(ctx); err != nil {
		return err
	}

	// Then start dashboard
	return ma.dashboard.Start(ctx)
}

// Stop gracefully shuts down the monitoring system
func (ma *MonitoringAdapter) Stop(ctx context.Context) error {
	// Stop dashboard first
	if err := ma.dashboard.Stop(ctx); err != nil {
		return err
	}

	// Then stop collector
	return ma.collector.Stop(ctx)
}

// RecordTestExecution records metrics for a test execution
func (ma *MonitoringAdapter) RecordTestExecution(result *models.TestResult, duration time.Duration) {
	ma.collector.RecordTestExecution(result, duration)
}

// RecordFileChange records metrics for file changes
func (ma *MonitoringAdapter) RecordFileChange(changeType string) {
	ma.collector.RecordFileChange(changeType)
}

// RecordCacheOperation records cache hit/miss metrics
func (ma *MonitoringAdapter) RecordCacheOperation(hit bool) {
	ma.collector.RecordCacheOperation(hit)
}

// RecordError records error metrics
func (ma *MonitoringAdapter) RecordError(errorType string, err error) {
	ma.collector.RecordError(errorType, err)
}

// IncrementCustomCounter increments a custom counter metric
func (ma *MonitoringAdapter) IncrementCustomCounter(name string, value int64) {
	ma.collector.IncrementCustomCounter(name, value)
}

// SetCustomGauge sets a custom gauge metric
func (ma *MonitoringAdapter) SetCustomGauge(name string, value float64) {
	ma.collector.SetCustomGauge(name, value)
}

// RecordCustomTimer records a custom timer metric
func (ma *MonitoringAdapter) RecordCustomTimer(name string, duration time.Duration) {
	ma.collector.RecordCustomTimer(name, duration)
}

// ExportMetrics exports metrics in the specified format
func (ma *MonitoringAdapter) ExportMetrics(format string) ([]byte, error) {
	return ma.collector.ExportMetrics(format)
}
