package monitoring

import (
	"context"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/events"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// AppMetricsCollector defines what the app package needs from monitoring
type AppMetricsCollector interface {
	// Lifecycle management
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Metrics recording
	RecordTestExecution(result *models.TestResult, duration time.Duration)
	RecordFileChange(changeType string)
	RecordCacheOperation(hit bool)
	RecordError(errorType string, err error)

	// Custom metrics
	IncrementCustomCounter(name string, value int64)
	SetCustomGauge(name string, value float64)
	RecordCustomTimer(name string, duration time.Duration)

	// Data access
	GetMetrics() *AppMetrics
	ExportMetrics(format string) ([]byte, error)

	// Health monitoring
	GetHealthStatus() *AppHealthStatus
	AddHealthCheck(name string, check AppHealthCheckFunc)
}

// AppDashboard defines what the app package needs from the monitoring dashboard
type AppDashboard interface {
	// Lifecycle management
	Start(ctx context.Context) error
	Stop(ctx context.Context) error

	// Data access
	GetDashboardMetrics() *AppDashboardMetrics
	ExportDashboardData(format string) ([]byte, error)
}

// AppMetricsCollectorFactory creates metrics collectors for the app package
type AppMetricsCollectorFactory interface {
	CreateMetricsCollector(config *AppMonitoringConfig, eventBus events.EventBus) AppMetricsCollector
}

// AppDashboardFactory creates dashboards for the app package
type AppDashboardFactory interface {
	CreateDashboard(collector AppMetricsCollector, config *AppDashboardConfig) AppDashboard
}

// AppMetricsCollectorDependencies contains dependencies for metrics collector creation
type AppMetricsCollectorDependencies struct {
	EventBus events.EventBus
	Config   *AppMonitoringConfig
}

// AppDashboardDependencies contains dependencies for dashboard creation
type AppDashboardDependencies struct {
	MetricsCollector AppMetricsCollector
	Config           *AppDashboardConfig
}
