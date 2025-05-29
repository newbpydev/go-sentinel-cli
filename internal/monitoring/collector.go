package monitoring

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

// DefaultAppMetricsCollector implements AppMetricsCollector interface
type DefaultAppMetricsCollector struct {
	mu           sync.RWMutex
	metrics      *AppMetrics
	startTime    time.Time
	httpServer   *http.Server
	healthChecks map[string]AppHealthCheckFunc
	eventBus     events.EventBus
	config       *AppMonitoringConfig
}

// DefaultAppMetricsCollectorFactory implements AppMetricsCollectorFactory interface
type DefaultAppMetricsCollectorFactory struct{}

// NewDefaultAppMetricsCollectorFactory creates a new metrics collector factory
func NewDefaultAppMetricsCollectorFactory() AppMetricsCollectorFactory {
	return &DefaultAppMetricsCollectorFactory{}
}

// CreateMetricsCollector creates a new metrics collector
func (f *DefaultAppMetricsCollectorFactory) CreateMetricsCollector(config *AppMonitoringConfig, eventBus events.EventBus) AppMetricsCollector {
	if config == nil {
		config = DefaultAppMonitoringConfig()
	}

	collector := &DefaultAppMetricsCollector{
		metrics: &AppMetrics{
			ErrorsByType:   make(map[string]int64),
			CustomCounters: make(map[string]int64),
			CustomGauges:   make(map[string]float64),
			CustomTimers:   make(map[string]time.Duration),
		},
		startTime:    time.Now(),
		healthChecks: make(map[string]AppHealthCheckFunc),
		eventBus:     eventBus,
		config:       config,
	}

	// Set up default health checks
	collector.setupDefaultHealthChecks()

	// Subscribe to events for automatic metrics collection
	collector.subscribeToEvents()

	return collector
}

// DefaultAppMonitoringConfig returns a sensible default configuration
func DefaultAppMonitoringConfig() *AppMonitoringConfig {
	return &AppMonitoringConfig{
		Enabled:         true,
		MetricsPort:     0,
		HealthPort:      0,
		MetricsInterval: 30 * time.Second,
		EnableProfiling: false,
		EnableTracing:   false,
		ExportFormat:    "json",
		RetentionPeriod: 24 * time.Hour,
		AlertThresholds: AppAlertThresholds{
			ErrorRatePercent:    5.0,
			MemoryUsageMB:       500,
			ResponseTimeMs:      1000 * time.Millisecond,
			CacheHitRatePercent: 80.0,
		},
	}
}

// Start starts the metrics collector
func (c *DefaultAppMetricsCollector) Start(ctx context.Context) error {
	if !c.config.Enabled {
		log.Println("Monitoring disabled, skipping start")
		return nil
	}

	log.Printf("Starting monitoring system on ports %d (metrics) and %d (health)",
		c.config.MetricsPort, c.config.HealthPort)

	// Start metrics collection
	go c.collectMetricsPeriodically(ctx)

	// Start HTTP servers for metrics and health endpoints
	if err := c.startHTTPServers(); err != nil {
		return fmt.Errorf("failed to start HTTP servers: %w", err)
	}

	log.Println("Monitoring system started successfully")
	return nil
}

// Stop gracefully shuts down the monitoring system
func (c *DefaultAppMetricsCollector) Stop(ctx context.Context) error {
	if c.httpServer != nil {
		log.Println("Shutting down monitoring HTTP server...")
		return c.httpServer.Shutdown(ctx)
	}
	return nil
}

// RecordTestExecution records metrics for a test execution
func (c *DefaultAppMetricsCollector) RecordTestExecution(result *models.TestResult, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics.TestsExecuted++
	c.metrics.TotalExecutionTime += duration

	// Handle nil result gracefully
	if result != nil {
		switch result.Status {
		case models.TestStatusPassed:
			c.metrics.TestsSucceeded++
		case models.TestStatusFailed:
			c.metrics.TestsFailed++
		case models.TestStatusSkipped:
			c.metrics.TestsSkipped++
		}
	}

	// Calculate average test time
	if c.metrics.TestsExecuted > 0 {
		c.metrics.AverageTestTime = c.metrics.TotalExecutionTime / time.Duration(c.metrics.TestsExecuted)
	}

	c.metrics.LastUpdate = time.Now()
}

// RecordFileChange records metrics for file changes
func (c *DefaultAppMetricsCollector) RecordFileChange(changeType string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics.FileChangesDetected++
	c.metrics.LastUpdate = time.Now()
}

// RecordCacheOperation records cache hit/miss metrics
func (c *DefaultAppMetricsCollector) RecordCacheOperation(hit bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if hit {
		c.metrics.CacheHits++
	} else {
		c.metrics.CacheMisses++
	}
	c.metrics.LastUpdate = time.Now()
}

// RecordError records error metrics
func (c *DefaultAppMetricsCollector) RecordError(errorType string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics.ErrorsTotal++
	if c.metrics.ErrorsByType[errorType] == 0 {
		c.metrics.ErrorsByType[errorType] = 0
	}
	c.metrics.ErrorsByType[errorType]++

	// Calculate error rate
	if c.metrics.TestsExecuted > 0 {
		c.metrics.ErrorRate = float64(c.metrics.ErrorsTotal) / float64(c.metrics.TestsExecuted) * 100
	}

	c.metrics.LastUpdate = time.Now()
}

// IncrementCustomCounter increments a custom counter metric
func (c *DefaultAppMetricsCollector) IncrementCustomCounter(name string, value int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics.CustomCounters[name] += value
	c.metrics.LastUpdate = time.Now()
}

// SetCustomGauge sets a custom gauge metric
func (c *DefaultAppMetricsCollector) SetCustomGauge(name string, value float64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics.CustomGauges[name] = value
	c.metrics.LastUpdate = time.Now()
}

// RecordCustomTimer records a custom timer metric
func (c *DefaultAppMetricsCollector) RecordCustomTimer(name string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.metrics.CustomTimers[name] = duration
	c.metrics.LastUpdate = time.Now()
}

// GetMetrics returns a copy of current metrics
func (c *DefaultAppMetricsCollector) GetMetrics() *AppMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Update runtime metrics
	c.updateRuntimeMetrics()

	// Create a deep copy
	metricsCopy := *c.metrics
	metricsCopy.ErrorsByType = make(map[string]int64)
	for k, v := range c.metrics.ErrorsByType {
		metricsCopy.ErrorsByType[k] = v
	}

	metricsCopy.CustomCounters = make(map[string]int64)
	for k, v := range c.metrics.CustomCounters {
		metricsCopy.CustomCounters[k] = v
	}

	metricsCopy.CustomGauges = make(map[string]float64)
	for k, v := range c.metrics.CustomGauges {
		metricsCopy.CustomGauges[k] = v
	}

	metricsCopy.CustomTimers = make(map[string]time.Duration)
	for k, v := range c.metrics.CustomTimers {
		metricsCopy.CustomTimers[k] = v
	}

	return &metricsCopy
}

// ExportMetrics exports metrics in the specified format
func (c *DefaultAppMetricsCollector) ExportMetrics(format string) ([]byte, error) {
	metrics := c.GetMetrics()

	switch format {
	case "json":
		return json.MarshalIndent(metrics, "", "  ")
	default:
		return json.MarshalIndent(metrics, "", "  ")
	}
}

// GetHealthStatus returns the current health status
func (c *DefaultAppMetricsCollector) GetHealthStatus() *AppHealthStatus {
	status := &AppHealthStatus{
		Status:      "healthy",
		Checks:      make(map[string]AppCheckResult),
		LastCheck:   time.Now(),
		Uptime:      time.Since(c.startTime),
		Version:     "1.0.0",
		Environment: "development",
	}

	// Run all health checks
	for name, check := range c.healthChecks {
		start := time.Now()

		// Handle nil check gracefully
		if check == nil {
			status.Checks[name] = AppCheckResult{
				Status:  "unknown",
				Message: "Health check function is nil",
				Latency: time.Since(start),
			}
			continue
		}

		err := check()
		latency := time.Since(start)

		if err != nil {
			status.Checks[name] = AppCheckResult{
				Status:  "unhealthy",
				Message: err.Error(),
				Latency: latency,
			}
			status.Status = "unhealthy"
		} else {
			status.Checks[name] = AppCheckResult{
				Status:  "healthy",
				Message: "OK",
				Latency: latency,
			}
		}
	}

	return status
}

// AddHealthCheck adds a custom health check
func (c *DefaultAppMetricsCollector) AddHealthCheck(name string, check AppHealthCheckFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.healthChecks[name] = check
}

// Private methods

func (c *DefaultAppMetricsCollector) updateRuntimeMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	c.metrics.MemoryUsage = int64(m.Alloc)
	c.metrics.GoroutineCount = runtime.NumGoroutine()
	c.metrics.Uptime = time.Since(c.startTime)
}

func (c *DefaultAppMetricsCollector) collectMetricsPeriodically(ctx context.Context) {
	// Ensure minimum interval to prevent panic
	interval := c.config.MetricsInterval
	if interval <= 0 {
		interval = 1 * time.Second // Default minimum interval
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.updateRuntimeMetrics()
		}
	}
}

func (c *DefaultAppMetricsCollector) setupDefaultHealthChecks() {
	// Add basic system health checks
	c.healthChecks["memory"] = func() error {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		if m.Alloc > 1024*1024*1024 { // 1GB threshold
			return fmt.Errorf("memory usage too high: %d bytes", m.Alloc)
		}
		return nil
	}

	c.healthChecks["goroutines"] = func() error {
		count := runtime.NumGoroutine()
		if count > 1000 { // 1000 goroutines threshold
			return fmt.Errorf("too many goroutines: %d", count)
		}
		return nil
	}

	c.healthChecks["disk"] = func() error {
		// Basic disk space check - can be enhanced
		if _, err := os.Stat("."); err != nil {
			return fmt.Errorf("cannot access current directory: %w", err)
		}
		return nil
	}
}

func (c *DefaultAppMetricsCollector) subscribeToEvents() {
	if c.eventBus == nil {
		return
	}

	// Subscribe to test events
	testHandler := &simpleEventHandler{
		handlerFunc: func(ctx context.Context, event events.Event) error {
			if event.Type() == "test.completed" {
				// Extract test result from event data
				if data, ok := event.Data().(map[string]interface{}); ok {
					if result, ok := data["result"].(*models.TestResult); ok {
						if duration, ok := data["duration"].(time.Duration); ok {
							c.RecordTestExecution(result, duration)
						}
					}
				}
			}
			return nil
		},
	}

	c.eventBus.Subscribe("test.completed", testHandler)

	// Subscribe to file change events
	fileHandler := &simpleEventHandler{
		handlerFunc: func(ctx context.Context, event events.Event) error {
			if event.Type() == "file.changed" {
				c.RecordFileChange("file_change")
			}
			return nil
		},
	}

	c.eventBus.Subscribe("file.changed", fileHandler)
}

type simpleEventHandler struct {
	handlerFunc func(context.Context, events.Event) error
}

func (h *simpleEventHandler) Handle(ctx context.Context, event events.Event) error {
	return h.handlerFunc(ctx, event)
}

func (h *simpleEventHandler) CanHandle(event events.Event) bool {
	return true
}

func (h *simpleEventHandler) Priority() int {
	return 0
}

func (c *DefaultAppMetricsCollector) startHTTPServers() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", c.handleMetrics)
	mux.HandleFunc("/health", c.handleHealth)
	mux.HandleFunc("/health/ready", c.handleReadiness)
	mux.HandleFunc("/health/live", c.handleLiveness)

	c.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", c.config.MetricsPort),
		Handler: mux,
	}

	// Check if port is valid before starting server
	if c.config.MetricsPort < 0 || c.config.MetricsPort > 65535 {
		return fmt.Errorf("invalid port number: %d", c.config.MetricsPort)
	}

	go func() {
		if err := c.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

func (c *DefaultAppMetricsCollector) handleMetrics(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = c.config.ExportFormat
	}

	data, err := c.ExportMetrics(format)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error exporting metrics: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (c *DefaultAppMetricsCollector) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := c.GetHealthStatus()
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		http.Error(w, fmt.Sprintf("Error marshaling health status: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if status.Status != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	w.Write(data)
}

func (c *DefaultAppMetricsCollector) handleReadiness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (c *DefaultAppMetricsCollector) handleLiveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
