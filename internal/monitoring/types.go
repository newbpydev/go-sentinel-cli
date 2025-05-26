package monitoring

import (
	"time"
)

// AppMetrics contains application performance and usage metrics for app package
type AppMetrics struct {
	// Test Execution Metrics
	TestsExecuted      int64         `json:"tests_executed"`
	TestsSucceeded     int64         `json:"tests_succeeded"`
	TestsFailed        int64         `json:"tests_failed"`
	TestsSkipped       int64         `json:"tests_skipped"`
	TotalExecutionTime time.Duration `json:"total_execution_time_ms"`
	AverageTestTime    time.Duration `json:"average_test_time_ms"`

	// File Watching Metrics
	FilesWatched        int64 `json:"files_watched"`
	FileChangesDetected int64 `json:"file_changes_detected"`
	WatchCycles         int64 `json:"watch_cycles"`

	// Performance Metrics
	CacheHits          int64         `json:"cache_hits"`
	CacheMisses        int64         `json:"cache_misses"`
	MemoryUsage        int64         `json:"memory_usage_bytes"`
	CPUUsage           float64       `json:"cpu_usage_percent"`
	GoroutineCount     int           `json:"goroutine_count"`
	ProcessingDuration time.Duration `json:"processing_duration_ms"`

	// Error Metrics
	ErrorsTotal      int64            `json:"errors_total"`
	ErrorsByType     map[string]int64 `json:"errors_by_type"`
	RecoveryAttempts int64            `json:"recovery_attempts"`
	ErrorRate        float64          `json:"error_rate_percent"`

	// System Metrics
	Uptime           time.Duration `json:"uptime_ms"`
	LastUpdate       time.Time     `json:"last_update"`
	ActiveOperations int64         `json:"active_operations"`

	// Custom Metrics
	CustomCounters map[string]int64         `json:"custom_counters"`
	CustomGauges   map[string]float64       `json:"custom_gauges"`
	CustomTimers   map[string]time.Duration `json:"custom_timers"`
}

// AppHealthStatus represents the health status for app package
type AppHealthStatus struct {
	Status      string                    `json:"status"` // healthy, degraded, unhealthy
	Checks      map[string]AppCheckResult `json:"checks"`
	LastCheck   time.Time                 `json:"last_check"`
	Uptime      time.Duration             `json:"uptime"`
	Version     string                    `json:"version"`
	Environment string                    `json:"environment"`
}

// AppCheckResult represents the result of a health check for app package
type AppCheckResult struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Latency time.Duration `json:"latency"`
}

// AppHealthCheckFunc defines a health check function for app package
type AppHealthCheckFunc func() error

// AppMonitoringConfig configures the monitoring system for app package
type AppMonitoringConfig struct {
	Enabled         bool               `json:"enabled"`
	MetricsPort     int                `json:"metrics_port"`
	HealthPort      int                `json:"health_port"`
	MetricsInterval time.Duration      `json:"metrics_interval"`
	EnableProfiling bool               `json:"enable_profiling"`
	EnableTracing   bool               `json:"enable_tracing"`
	ExportFormat    string             `json:"export_format"` // json, prometheus, opentelemetry
	RetentionPeriod time.Duration      `json:"retention_period"`
	AlertThresholds AppAlertThresholds `json:"alert_thresholds"`
}

// AppAlertThresholds defines when to trigger alerts for app package
type AppAlertThresholds struct {
	ErrorRatePercent    float64       `json:"error_rate_percent"`
	MemoryUsageMB       int64         `json:"memory_usage_mb"`
	ResponseTimeMs      time.Duration `json:"response_time_ms"`
	CacheHitRatePercent float64       `json:"cache_hit_rate_percent"`
}

// AppDashboardConfig configures the monitoring dashboard for app package
type AppDashboardConfig struct {
	Port                int           `json:"port"`
	RefreshInterval     time.Duration `json:"refresh_interval"`
	MaxDataPoints       int           `json:"max_data_points"`
	EnableRealTime      bool          `json:"enable_real_time"`
	EnableAlerts        bool          `json:"enable_alerts"`
	ChartRetentionHours int           `json:"chart_retention_hours"`
	Theme               string        `json:"theme"` // "light", "dark", "auto"
}

// AppDashboardMetrics provides dashboard-specific metrics aggregation for app package
type AppDashboardMetrics struct {
	Overview     *AppOverviewMetrics     `json:"overview"`
	Performance  *AppPerformanceMetrics  `json:"performance"`
	TestMetrics  *AppTestSummaryMetrics  `json:"test_metrics"`
	SystemHealth *AppSystemHealthMetrics `json:"system_health"`
	Trends       *AppTrendMetrics        `json:"trends"`
	Alerts       []*AppAlert             `json:"alerts"`
}

// AppOverviewMetrics provides high-level system overview for app package
type AppOverviewMetrics struct {
	SystemStatus   string        `json:"system_status"`
	Uptime         time.Duration `json:"uptime"`
	TotalTests     int64         `json:"total_tests"`
	SuccessRate    float64       `json:"success_rate"`
	AverageTime    time.Duration `json:"average_time"`
	ActiveAlerts   int           `json:"active_alerts"`
	CriticalAlerts int           `json:"critical_alerts"`
	LastDeployment *time.Time    `json:"last_deployment,omitempty"`
	CurrentVersion string        `json:"current_version"`
}

// AppPerformanceMetrics provides detailed performance data for app package
type AppPerformanceMetrics struct {
	MemoryUsage    int64         `json:"memory_usage_mb"`
	CPUUsage       float64       `json:"cpu_usage_percent"`
	GoroutineCount int           `json:"goroutine_count"`
	CacheHitRate   float64       `json:"cache_hit_rate"`
	ResponseTime   time.Duration `json:"response_time_ms"`
	ThroughputRPS  float64       `json:"throughput_rps"`
	ErrorRate      float64       `json:"error_rate"`
	NetworkLatency time.Duration `json:"network_latency_ms"`
}

// AppTestSummaryMetrics provides test execution summaries for app package
type AppTestSummaryMetrics struct {
	TotalExecutions int64    `json:"total_executions"`
	PassedTests     int64    `json:"passed_tests"`
	FailedTests     int64    `json:"failed_tests"`
	SkippedTests    int64    `json:"skipped_tests"`
	Coverage        float64  `json:"coverage_percent"`
	Flakiness       float64  `json:"flakiness_percent"`
	TopFailures     []string `json:"top_failures"`
	SlowestTests    []string `json:"slowest_tests"`
}

// AppSystemHealthMetrics provides system health indicators for app package
type AppSystemHealthMetrics struct {
	DiskUsage        float64              `json:"disk_usage_percent"`
	NetworkStatus    string               `json:"network_status"`
	DependencyStatus map[string]string    `json:"dependency_status"`
	ServiceStatus    map[string]string    `json:"service_status"`
	HealthScore      float64              `json:"health_score"`
	RecentIncidents  []AppIncidentSummary `json:"recent_incidents"`
}

// AppTrendMetrics provides trend analysis data for app package
type AppTrendMetrics struct {
	PerformanceTrend string                          `json:"performance_trend"` // "IMPROVING", "STABLE", "DEGRADING"
	TestSuccessTrend string                          `json:"test_success_trend"`
	ErrorTrend       string                          `json:"error_trend"`
	UsageTrend       string                          `json:"usage_trend"`
	TrendCharts      map[string][]AppTimeSeriesPoint `json:"trend_charts"`
	Predictions      map[string]float64              `json:"predictions"`
}

// AppAlert represents an active monitoring alert for app package
type AppAlert struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"` // "LOW", "MEDIUM", "HIGH", "CRITICAL"
	Status      string                 `json:"status"`   // "FIRING", "RESOLVED", "SILENCED"
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Value       float64                `json:"value"`
	Threshold   float64                `json:"threshold"`
	Labels      map[string]string      `json:"labels"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// AppIncidentSummary provides summary of recent incidents for app package
type AppIncidentSummary struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Severity  string        `json:"severity"`
	Status    string        `json:"status"`
	StartTime time.Time     `json:"start_time"`
	Duration  time.Duration `json:"duration"`
	Impact    string        `json:"impact"`
}

// AppTimeSeriesPoint represents a single data point in time series for app package
type AppTimeSeriesPoint struct {
	Timestamp time.Time         `json:"timestamp"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// AppAlertManager handles monitoring alerts and notifications for app package
type AppAlertManager struct {
	ActiveAlerts    map[string]*AppAlert          `json:"active_alerts"`
	AlertRules      []*AppAlertRule               `json:"alert_rules"`
	WebhookURLs     []string                      `json:"webhook_urls"`
	SilencedAlerts  map[string]time.Time          `json:"silenced_alerts"`
	EscalationRules map[string]*AppEscalationRule `json:"escalation_rules"`
}

// AppAlertRule defines conditions for triggering alerts for app package
type AppAlertRule struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Metric      string            `json:"metric"`
	Operator    string            `json:"operator"` // "gt", "lt", "eq", "ne", "gte", "lte"
	Threshold   float64           `json:"threshold"`
	Duration    time.Duration     `json:"duration"`
	Severity    string            `json:"severity"`
	Labels      map[string]string `json:"labels"`
	Actions     []AppAlertAction  `json:"actions"`
}

// AppAlertAction defines what to do when an alert triggers for app package
type AppAlertAction struct {
	Type   string                 `json:"type"` // "webhook", "email", "slack", "log"
	Config map[string]interface{} `json:"config"`
}

// AppEscalationRule defines alert escalation policies for app package
type AppEscalationRule struct {
	AlertName     string           `json:"alert_name"`
	EscalateAfter time.Duration    `json:"escalate_after"`
	Actions       []AppAlertAction `json:"actions"`
}

// AppTrendAnalyzer analyzes metric trends and patterns for app package
type AppTrendAnalyzer struct {
	HistoricalData map[string][]AppTimeSeriesPoint `json:"historical_data"`
	MaxDataPoints  int                             `json:"max_data_points"`
}

// AppWebSocketClient represents a connected WebSocket client for app package
type AppWebSocketClient struct {
	ID         string      `json:"id"`
	Connection interface{} `json:"connection"` // WebSocket connection placeholder
	LastPing   time.Time   `json:"last_ping"`
	Subscribed []string    `json:"subscribed"` // Subscribed metric names
}

// AppRealTimeData holds real-time streaming data for app package
type AppRealTimeData struct {
	CurrentMetrics map[string]interface{}           `json:"current_metrics"`
	LastUpdate     time.Time                        `json:"last_update"`
	Subscribers    map[string][]*AppWebSocketClient `json:"subscribers"`
}
