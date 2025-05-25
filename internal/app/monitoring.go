package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/newbpydev/go-sentinel/pkg/events"
	"github.com/newbpydev/go-sentinel/pkg/models"
)

// MetricsCollector handles application metrics and observability
type MetricsCollector struct {
	mu           sync.RWMutex
	metrics      *AppMetrics
	startTime    time.Time
	httpServer   *http.Server
	healthChecks map[string]HealthCheckFunc
	eventBus     events.EventBus
	config       *MonitoringConfig
}

// AppMetrics contains application performance and usage metrics
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

// MonitoringConfig configures the monitoring system
type MonitoringConfig struct {
	Enabled         bool            `json:"enabled"`
	MetricsPort     int             `json:"metrics_port"`
	HealthPort      int             `json:"health_port"`
	MetricsInterval time.Duration   `json:"metrics_interval"`
	EnableProfiling bool            `json:"enable_profiling"`
	EnableTracing   bool            `json:"enable_tracing"`
	ExportFormat    string          `json:"export_format"` // json, prometheus, opentelemetry
	RetentionPeriod time.Duration   `json:"retention_period"`
	AlertThresholds AlertThresholds `json:"alert_thresholds"`
}

// AlertThresholds defines when to trigger alerts
type AlertThresholds struct {
	ErrorRatePercent    float64       `json:"error_rate_percent"`
	MemoryUsageMB       int64         `json:"memory_usage_mb"`
	ResponseTimeMs      time.Duration `json:"response_time_ms"`
	CacheHitRatePercent float64       `json:"cache_hit_rate_percent"`
}

// HealthCheckFunc defines a health check function
type HealthCheckFunc func() error

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status      string                 `json:"status"` // healthy, degraded, unhealthy
	Checks      map[string]CheckResult `json:"checks"`
	LastCheck   time.Time              `json:"last_check"`
	Uptime      time.Duration          `json:"uptime"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
}

// CheckResult represents the result of a health check
type CheckResult struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Latency time.Duration `json:"latency"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(config *MonitoringConfig, eventBus events.EventBus) *MetricsCollector {
	if config == nil {
		config = DefaultMonitoringConfig()
	}

	collector := &MetricsCollector{
		metrics: &AppMetrics{
			ErrorsByType:   make(map[string]int64),
			CustomCounters: make(map[string]int64),
			CustomGauges:   make(map[string]float64),
			CustomTimers:   make(map[string]time.Duration),
		},
		startTime:    time.Now(),
		healthChecks: make(map[string]HealthCheckFunc),
		eventBus:     eventBus,
		config:       config,
	}

	// Set up default health checks
	collector.setupDefaultHealthChecks()

	// Subscribe to events for automatic metrics collection
	collector.subscribeToEvents()

	return collector
}

// DefaultMonitoringConfig returns a sensible default configuration
func DefaultMonitoringConfig() *MonitoringConfig {
	return &MonitoringConfig{
		Enabled:         true,
		MetricsPort:     8080,
		HealthPort:      8081,
		MetricsInterval: 30 * time.Second,
		EnableProfiling: false,
		EnableTracing:   false,
		ExportFormat:    "json",
		RetentionPeriod: 24 * time.Hour,
		AlertThresholds: AlertThresholds{
			ErrorRatePercent:    5.0,
			MemoryUsageMB:       500,
			ResponseTimeMs:      1000 * time.Millisecond,
			CacheHitRatePercent: 80.0,
		},
	}
}

// Start initializes the monitoring system
func (mc *MetricsCollector) Start(ctx context.Context) error {
	if !mc.config.Enabled {
		log.Println("Monitoring disabled, skipping start")
		return nil
	}

	log.Printf("Starting monitoring system on ports %d (metrics) and %d (health)",
		mc.config.MetricsPort, mc.config.HealthPort)

	// Start metrics collection
	go mc.collectMetricsPeriodically(ctx)

	// Start HTTP servers for metrics and health endpoints
	if err := mc.startHTTPServers(); err != nil {
		return fmt.Errorf("failed to start HTTP servers: %w", err)
	}

	log.Println("Monitoring system started successfully")
	return nil
}

// Stop gracefully shuts down the monitoring system
func (mc *MetricsCollector) Stop(ctx context.Context) error {
	if mc.httpServer != nil {
		log.Println("Shutting down monitoring HTTP server...")
		return mc.httpServer.Shutdown(ctx)
	}
	return nil
}

// RecordTestExecution records metrics for a test execution
func (mc *MetricsCollector) RecordTestExecution(result *models.TestResult, duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics.TestsExecuted++
	mc.metrics.TotalExecutionTime += duration

	switch result.Status {
	case models.TestStatusPassed:
		mc.metrics.TestsSucceeded++
	case models.TestStatusFailed:
		mc.metrics.TestsFailed++
	case models.TestStatusSkipped:
		mc.metrics.TestsSkipped++
	}

	// Calculate average test time
	if mc.metrics.TestsExecuted > 0 {
		mc.metrics.AverageTestTime = mc.metrics.TotalExecutionTime / time.Duration(mc.metrics.TestsExecuted)
	}

	mc.metrics.LastUpdate = time.Now()
}

// RecordFileChange records metrics for file changes
func (mc *MetricsCollector) RecordFileChange(changeType string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics.FileChangesDetected++
	mc.metrics.LastUpdate = time.Now()
}

// RecordCacheOperation records cache hit/miss metrics
func (mc *MetricsCollector) RecordCacheOperation(hit bool) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if hit {
		mc.metrics.CacheHits++
	} else {
		mc.metrics.CacheMisses++
	}
	mc.metrics.LastUpdate = time.Now()
}

// RecordError records error metrics
func (mc *MetricsCollector) RecordError(errorType string, err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics.ErrorsTotal++
	if mc.metrics.ErrorsByType[errorType] == 0 {
		mc.metrics.ErrorsByType[errorType] = 0
	}
	mc.metrics.ErrorsByType[errorType]++

	// Calculate error rate
	if mc.metrics.TestsExecuted > 0 {
		mc.metrics.ErrorRate = float64(mc.metrics.ErrorsTotal) / float64(mc.metrics.TestsExecuted) * 100
	}

	mc.metrics.LastUpdate = time.Now()

	log.Printf("Error recorded: %s - %v", errorType, err)
}

// IncrementCustomCounter increments a custom counter metric
func (mc *MetricsCollector) IncrementCustomCounter(name string, value int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics.CustomCounters[name] += value
	mc.metrics.LastUpdate = time.Now()
}

// SetCustomGauge sets a custom gauge metric
func (mc *MetricsCollector) SetCustomGauge(name string, value float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics.CustomGauges[name] = value
	mc.metrics.LastUpdate = time.Now()
}

// RecordCustomTimer records a custom timer metric
func (mc *MetricsCollector) RecordCustomTimer(name string, duration time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics.CustomTimers[name] = duration
	mc.metrics.LastUpdate = time.Now()
}

// GetMetrics returns a copy of current metrics
func (mc *MetricsCollector) GetMetrics() *AppMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Update runtime metrics
	mc.updateRuntimeMetrics()

	// Create a deep copy
	metricsCopy := *mc.metrics
	metricsCopy.ErrorsByType = make(map[string]int64)
	for k, v := range mc.metrics.ErrorsByType {
		metricsCopy.ErrorsByType[k] = v
	}

	metricsCopy.CustomCounters = make(map[string]int64)
	for k, v := range mc.metrics.CustomCounters {
		metricsCopy.CustomCounters[k] = v
	}

	metricsCopy.CustomGauges = make(map[string]float64)
	for k, v := range mc.metrics.CustomGauges {
		metricsCopy.CustomGauges[k] = v
	}

	metricsCopy.CustomTimers = make(map[string]time.Duration)
	for k, v := range mc.metrics.CustomTimers {
		metricsCopy.CustomTimers[k] = v
	}

	return &metricsCopy
}

// GetHealthStatus returns the current health status
func (mc *MetricsCollector) GetHealthStatus() *HealthStatus {
	status := &HealthStatus{
		Status:      "healthy",
		Checks:      make(map[string]CheckResult),
		LastCheck:   time.Now(),
		Uptime:      time.Since(mc.startTime),
		Version:     os.Getenv("VERSION"),
		Environment: os.Getenv("ENVIRONMENT"),
	}

	// Run all health checks
	for name, check := range mc.healthChecks {
		start := time.Now()
		err := check()
		latency := time.Since(start)

		if err != nil {
			status.Checks[name] = CheckResult{
				Status:  "unhealthy",
				Message: err.Error(),
				Latency: latency,
			}
			status.Status = "unhealthy"
		} else {
			status.Checks[name] = CheckResult{
				Status:  "healthy",
				Message: "OK",
				Latency: latency,
			}
		}
	}

	return status
}

// AddHealthCheck adds a custom health check
func (mc *MetricsCollector) AddHealthCheck(name string, check HealthCheckFunc) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.healthChecks[name] = check
}

// ExportMetrics exports metrics in the specified format
func (mc *MetricsCollector) ExportMetrics(format string) ([]byte, error) {
	metrics := mc.GetMetrics()

	switch format {
	case "json":
		return json.MarshalIndent(metrics, "", "  ")
	case "prometheus":
		return mc.exportPrometheusFormat(metrics)
	default:
		return json.MarshalIndent(metrics, "", "  ")
	}
}

// Private methods

func (mc *MetricsCollector) collectMetricsPeriodically(ctx context.Context) {
	ticker := time.NewTicker(mc.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mc.updateRuntimeMetrics()
		}
	}
}

func (mc *MetricsCollector) updateRuntimeMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics.MemoryUsage = int64(m.Alloc)
	mc.metrics.GoroutineCount = runtime.NumGoroutine()
	mc.metrics.Uptime = time.Since(mc.startTime)
}

func (mc *MetricsCollector) setupDefaultHealthChecks() {
	// Memory usage health check
	mc.AddHealthCheck("memory", func() error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		usageMB := int64(m.Alloc / 1024 / 1024)

		if usageMB > mc.config.AlertThresholds.MemoryUsageMB {
			return fmt.Errorf("memory usage %d MB exceeds threshold %d MB",
				usageMB, mc.config.AlertThresholds.MemoryUsageMB)
		}
		return nil
	})

	// Error rate health check
	mc.AddHealthCheck("error_rate", func() error {
		if mc.metrics.ErrorRate > mc.config.AlertThresholds.ErrorRatePercent {
			return fmt.Errorf("error rate %.2f%% exceeds threshold %.2f%%",
				mc.metrics.ErrorRate, mc.config.AlertThresholds.ErrorRatePercent)
		}
		return nil
	})

	// Cache performance health check
	mc.AddHealthCheck("cache_performance", func() error {
		total := mc.metrics.CacheHits + mc.metrics.CacheMisses
		if total > 0 {
			hitRate := float64(mc.metrics.CacheHits) / float64(total) * 100
			if hitRate < mc.config.AlertThresholds.CacheHitRatePercent {
				return fmt.Errorf("cache hit rate %.2f%% is below threshold %.2f%%",
					hitRate, mc.config.AlertThresholds.CacheHitRatePercent)
			}
		}
		return nil
	})
}

func (mc *MetricsCollector) subscribeToEvents() {
	if mc.eventBus == nil {
		return
	}

	// Subscribe to test events
	testHandler := &simpleEventHandler{
		handlerFunc: func(ctx context.Context, event events.Event) error {
			if testEvent, ok := event.(*events.TestCompletedEvent); ok {
				// Create a mock test result based on the event
				result := &models.TestResult{
					Name:   testEvent.TestName,
					Status: models.TestStatusPassed,
				}
				if !testEvent.Success {
					result.Status = models.TestStatusFailed
				}
				mc.RecordTestExecution(result, testEvent.Duration)
			}
			return nil
		},
	}
	mc.eventBus.Subscribe("test.completed", testHandler)

	// Subscribe to file change events
	fileHandler := &simpleEventHandler{
		handlerFunc: func(ctx context.Context, event events.Event) error {
			if fileEvent, ok := event.(*events.FileChangedEvent); ok {
				mc.RecordFileChange(fileEvent.ChangeType)
			}
			return nil
		},
	}
	mc.eventBus.Subscribe("file.changed", fileHandler)
}

// simpleEventHandler is a simple implementation of EventHandler
type simpleEventHandler struct {
	handlerFunc func(context.Context, events.Event) error
}

// Handle implements the EventHandler interface
func (h *simpleEventHandler) Handle(ctx context.Context, event events.Event) error {
	return h.handlerFunc(ctx, event)
}

// CanHandle implements the EventHandler interface
func (h *simpleEventHandler) CanHandle(event events.Event) bool {
	return true
}

// Priority implements the EventHandler interface
func (h *simpleEventHandler) Priority() int {
	return 0
}

func (mc *MetricsCollector) startHTTPServers() error {
	mux := http.NewServeMux()

	// Metrics endpoint
	mux.HandleFunc("/metrics", mc.handleMetrics)
	mux.HandleFunc("/health", mc.handleHealth)
	mux.HandleFunc("/health/ready", mc.handleReadiness)
	mux.HandleFunc("/health/live", mc.handleLiveness)

	// Profiling endpoints (if enabled)
	if mc.config.EnableProfiling {
		mux.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP)
	}

	mc.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", mc.config.MetricsPort),
		Handler: mux,
	}

	go func() {
		if err := mc.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Monitoring HTTP server error: %v", err)
		}
	}()

	return nil
}

func (mc *MetricsCollector) handleMetrics(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = mc.config.ExportFormat
	}

	data, err := mc.ExportMetrics(format)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to export metrics: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if format == "prometheus" {
		w.Header().Set("Content-Type", "text/plain")
	}

	w.Write(data)
}

func (mc *MetricsCollector) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := mc.GetHealthStatus()

	statusCode := http.StatusOK
	if health.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(health)
}

func (mc *MetricsCollector) handleReadiness(w http.ResponseWriter, r *http.Request) {
	// Basic readiness check - system is ready to receive requests
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

func (mc *MetricsCollector) handleLiveness(w http.ResponseWriter, r *http.Request) {
	// Basic liveness check - system is alive
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "alive"})
}

func (mc *MetricsCollector) exportPrometheusFormat(metrics *AppMetrics) ([]byte, error) {
	var output string

	// Test metrics
	output += fmt.Sprintf("go_sentinel_tests_executed_total %d\n", metrics.TestsExecuted)
	output += fmt.Sprintf("go_sentinel_tests_succeeded_total %d\n", metrics.TestsSucceeded)
	output += fmt.Sprintf("go_sentinel_tests_failed_total %d\n", metrics.TestsFailed)
	output += fmt.Sprintf("go_sentinel_tests_skipped_total %d\n", metrics.TestsSkipped)

	// Performance metrics
	output += fmt.Sprintf("go_sentinel_memory_usage_bytes %d\n", metrics.MemoryUsage)
	output += fmt.Sprintf("go_sentinel_goroutines %d\n", metrics.GoroutineCount)
	output += fmt.Sprintf("go_sentinel_cache_hits_total %d\n", metrics.CacheHits)
	output += fmt.Sprintf("go_sentinel_cache_misses_total %d\n", metrics.CacheMisses)

	// Error metrics
	output += fmt.Sprintf("go_sentinel_errors_total %d\n", metrics.ErrorsTotal)
	output += fmt.Sprintf("go_sentinel_error_rate_percent %f\n", metrics.ErrorRate)

	// Custom metrics
	for name, value := range metrics.CustomCounters {
		output += fmt.Sprintf("go_sentinel_custom_counter{name=\"%s\"} %d\n", name, value)
	}

	for name, value := range metrics.CustomGauges {
		output += fmt.Sprintf("go_sentinel_custom_gauge{name=\"%s\"} %f\n", name, value)
	}

	return []byte(output), nil
}
