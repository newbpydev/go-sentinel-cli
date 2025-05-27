package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// DefaultAppDashboard implements AppDashboard interface
type DefaultAppDashboard struct {
	mu               sync.RWMutex
	metricsCollector AppMetricsCollector
	alertManager     *AppAlertManager
	trendAnalyzer    *AppTrendAnalyzer
	config           *AppDashboardConfig
	server           *http.Server
	websocketClients map[string]*AppWebSocketClient
	realTimeData     *AppRealTimeData
}

// DefaultAppDashboardFactory implements AppDashboardFactory interface
type DefaultAppDashboardFactory struct{}

// NewDefaultAppDashboardFactory creates a new dashboard factory
func NewDefaultAppDashboardFactory() AppDashboardFactory {
	return &DefaultAppDashboardFactory{}
}

// CreateDashboard creates a new dashboard
func (f *DefaultAppDashboardFactory) CreateDashboard(collector AppMetricsCollector, config *AppDashboardConfig) AppDashboard {
	if config == nil {
		config = DefaultAppDashboardConfig()
	}

	dashboard := &DefaultAppDashboard{
		metricsCollector: collector,
		alertManager:     NewAppAlertManager(),
		trendAnalyzer:    NewAppTrendAnalyzer(config.MaxDataPoints),
		config:           config,
		websocketClients: make(map[string]*AppWebSocketClient),
		realTimeData:     NewAppRealTimeData(),
	}

	// Set up default alert rules
	dashboard.setupDefaultAlertRules()

	return dashboard
}

// DefaultAppDashboardConfig returns a sensible default configuration
func DefaultAppDashboardConfig() *AppDashboardConfig {
	return &AppDashboardConfig{
		Port:                3000,
		RefreshInterval:     5 * time.Second,
		MaxDataPoints:       1000,
		EnableRealTime:      true,
		EnableAlerts:        true,
		ChartRetentionHours: 24,
		Theme:               "auto",
	}
}

// Start starts the dashboard
func (d *DefaultAppDashboard) Start(ctx context.Context) error {
	log.Printf("Starting dashboard on port %d", d.config.Port)

	// Start background processes
	go d.collectTrendData(ctx)
	go d.processAlerts(ctx)
	go d.updateRealTimeData(ctx)

	// Start HTTP server
	mux := http.NewServeMux()
	d.setupRoutes(mux)

	d.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", d.config.Port),
		Handler: mux,
	}

	go func() {
		if err := d.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Dashboard server error: %v", err)
		}
	}()

	return nil
}

// Stop stops the dashboard
func (d *DefaultAppDashboard) Stop(ctx context.Context) error {
	if d.server != nil {
		log.Println("Shutting down dashboard server...")
		return d.server.Shutdown(ctx)
	}
	return nil
}

// GetDashboardMetrics returns dashboard metrics
func (d *DefaultAppDashboard) GetDashboardMetrics() *AppDashboardMetrics {
	var baseMetrics *AppMetrics
	if d.metricsCollector != nil {
		baseMetrics = d.metricsCollector.GetMetrics()
	}

	if baseMetrics == nil {
		// Return empty metrics if collector is nil or fails
		baseMetrics = &AppMetrics{
			ErrorsByType:   make(map[string]int64),
			CustomCounters: make(map[string]int64),
			CustomGauges:   make(map[string]float64),
			CustomTimers:   make(map[string]time.Duration),
		}
	}

	return &AppDashboardMetrics{
		Overview:     d.buildOverviewMetrics(baseMetrics),
		Performance:  d.buildPerformanceMetrics(baseMetrics),
		TestMetrics:  d.buildTestMetrics(baseMetrics),
		SystemHealth: d.buildSystemHealthMetrics(baseMetrics),
		Trends:       d.buildTrendMetrics(),
		Alerts:       []*AppAlert{},
	}
}

// ExportDashboardData exports dashboard data
func (d *DefaultAppDashboard) ExportDashboardData(format string) ([]byte, error) {
	metrics := d.GetDashboardMetrics()

	switch format {
	case "json":
		return json.MarshalIndent(metrics, "", "  ")
	default:
		return json.MarshalIndent(metrics, "", "  ")
	}
}

// Private helper methods

func (d *DefaultAppDashboard) buildOverviewMetrics(base *AppMetrics) *AppOverviewMetrics {
	successRate := float64(0)
	if base.TestsExecuted > 0 {
		successRate = float64(base.TestsSucceeded) / float64(base.TestsExecuted) * 100
	}

	return &AppOverviewMetrics{
		SystemStatus:   "healthy",
		Uptime:         base.Uptime,
		TotalTests:     base.TestsExecuted,
		SuccessRate:    successRate,
		AverageTime:    base.AverageTestTime,
		ActiveAlerts:   0,
		CriticalAlerts: 0,
		CurrentVersion: "1.0.0",
	}
}

func (d *DefaultAppDashboard) buildPerformanceMetrics(base *AppMetrics) *AppPerformanceMetrics {
	cacheHitRate := float64(0)
	total := base.CacheHits + base.CacheMisses
	if total > 0 {
		cacheHitRate = float64(base.CacheHits) / float64(total) * 100
	}

	return &AppPerformanceMetrics{
		MemoryUsage:    base.MemoryUsage / 1024 / 1024, // Convert to MB
		CPUUsage:       base.CPUUsage,
		GoroutineCount: base.GoroutineCount,
		CacheHitRate:   cacheHitRate,
		ResponseTime:   base.ProcessingDuration,
		ErrorRate:      base.ErrorRate,
	}
}

func (d *DefaultAppDashboard) buildTestMetrics(base *AppMetrics) *AppTestSummaryMetrics {
	return &AppTestSummaryMetrics{
		TotalExecutions: base.TestsExecuted,
		PassedTests:     base.TestsSucceeded,
		FailedTests:     base.TestsFailed,
		SkippedTests:    base.TestsSkipped,
		Flakiness:       2.5, // Placeholder
		TopFailures:     []string{"test1", "test2"},
		SlowestTests:    []string{"slow_test1", "slow_test2"},
	}
}

func (d *DefaultAppDashboard) buildSystemHealthMetrics(base *AppMetrics) *AppSystemHealthMetrics {
	return &AppSystemHealthMetrics{
		NetworkStatus:    "CONNECTED",
		DependencyStatus: map[string]string{"database": "HEALTHY"},
		ServiceStatus:    map[string]string{"test_runner": "RUNNING"},
		HealthScore:      95.5,
		RecentIncidents:  []AppIncidentSummary{},
	}
}

func (d *DefaultAppDashboard) buildTrendMetrics() *AppTrendMetrics {
	return &AppTrendMetrics{
		PerformanceTrend: "STABLE",
		TestSuccessTrend: "IMPROVING",
		ErrorTrend:       "STABLE",
		UsageTrend:       "INCREASING",
		TrendCharts:      make(map[string][]AppTimeSeriesPoint),
		Predictions:      make(map[string]float64),
	}
}

// Constructor functions for dashboard components

func NewAppAlertManager() *AppAlertManager {
	return &AppAlertManager{
		ActiveAlerts:    make(map[string]*AppAlert),
		AlertRules:      []*AppAlertRule{},
		WebhookURLs:     []string{},
		SilencedAlerts:  make(map[string]time.Time),
		EscalationRules: make(map[string]*AppEscalationRule),
	}
}

func NewAppTrendAnalyzer(maxDataPoints int) *AppTrendAnalyzer {
	return &AppTrendAnalyzer{
		HistoricalData: make(map[string][]AppTimeSeriesPoint),
		MaxDataPoints:  maxDataPoints,
	}
}

func NewAppRealTimeData() *AppRealTimeData {
	return &AppRealTimeData{
		CurrentMetrics: make(map[string]interface{}),
		LastUpdate:     time.Now(),
		Subscribers:    make(map[string][]*AppWebSocketClient),
	}
}

// Dashboard background processes

func (d *DefaultAppDashboard) collectTrendData(ctx context.Context) {
	// Ensure minimum interval to prevent panic
	interval := d.config.RefreshInterval
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
			// Collect current metrics and add to trend data
			if d.metricsCollector != nil {
				metrics := d.metricsCollector.GetMetrics()
				if metrics != nil {
					now := time.Now()

					// Add data points for key metrics
					if metrics.TestsExecuted > 0 {
						d.trendAnalyzer.addDataPoint("test_success_rate", float64(metrics.TestsSucceeded)/float64(metrics.TestsExecuted)*100, now)
					}
					d.trendAnalyzer.addDataPoint("error_rate", metrics.ErrorRate, now)
					d.trendAnalyzer.addDataPoint("memory_usage", float64(metrics.MemoryUsage), now)
					d.trendAnalyzer.addDataPoint("cpu_usage", metrics.CPUUsage, now)
				}
			}
		}
	}
}

func (d *DefaultAppDashboard) processAlerts(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Check alerts every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if d.config.EnableAlerts && d.metricsCollector != nil {
				metrics := d.metricsCollector.GetMetrics()
				if metrics != nil {
					d.alertManager.evaluateAlerts(metrics)
				}
			}
		}
	}
}

func (d *DefaultAppDashboard) updateRealTimeData(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second) // Update real-time data every second
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if d.config.EnableRealTime && d.metricsCollector != nil {
				metrics := d.metricsCollector.GetMetrics()
				if metrics != nil {
					metricsMap := map[string]interface{}{
						"tests_executed":  metrics.TestsExecuted,
						"tests_succeeded": metrics.TestsSucceeded,
						"tests_failed":    metrics.TestsFailed,
						"memory_usage":    metrics.MemoryUsage,
						"cpu_usage":       metrics.CPUUsage,
						"error_rate":      metrics.ErrorRate,
					}
					d.realTimeData.update(metricsMap)
				}
			}
		}
	}
}

func (d *DefaultAppDashboard) setupRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", d.handleDashboard)
	mux.HandleFunc("/api/metrics", d.handleAPIMetrics)
	mux.HandleFunc("/api/alerts", d.handleAPIAlerts)
	mux.HandleFunc("/api/trends", d.handleAPITrends)
	mux.HandleFunc("/api/export", d.handleAPIExport)
	mux.HandleFunc("/ws", d.handleWebSocket)
	mux.HandleFunc("/static/", d.handleStatic)
}

func (d *DefaultAppDashboard) setupDefaultAlertRules() {
	// Error rate alert
	d.alertManager.AlertRules = append(d.alertManager.AlertRules, &AppAlertRule{
		Name:        "high_error_rate",
		Description: "Error rate is above threshold",
		Metric:      "error_rate",
		Operator:    "gt",
		Threshold:   5.0,
		Duration:    5 * time.Minute,
		Severity:    "HIGH",
		Labels:      map[string]string{"component": "test_runner"},
		Actions: []AppAlertAction{
			{Type: "log", Config: map[string]interface{}{"level": "warn"}},
		},
	})

	// Memory usage alert
	d.alertManager.AlertRules = append(d.alertManager.AlertRules, &AppAlertRule{
		Name:        "high_memory_usage",
		Description: "Memory usage is above threshold",
		Metric:      "memory_usage",
		Operator:    "gt",
		Threshold:   500 * 1024 * 1024, // 500MB
		Duration:    2 * time.Minute,
		Severity:    "MEDIUM",
		Labels:      map[string]string{"component": "system"},
		Actions: []AppAlertAction{
			{Type: "log", Config: map[string]interface{}{"level": "warn"}},
		},
	})
}

// HTTP handlers

func (d *DefaultAppDashboard) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<!DOCTYPE html><html><head><title>Go Sentinel Dashboard</title></head><body><h1>Go Sentinel Monitoring Dashboard</h1><p>Dashboard is running. Use /api/metrics for metrics data.</p></body></html>`))
}

func (d *DefaultAppDashboard) handleAPIMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := d.GetDashboardMetrics()
	data, _ := json.MarshalIndent(metrics, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (d *DefaultAppDashboard) handleAPIAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := d.alertManager.getActiveAlerts()
	data, _ := json.MarshalIndent(alerts, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (d *DefaultAppDashboard) handleAPITrends(w http.ResponseWriter, r *http.Request) {
	trends := d.buildTrendMetrics()
	data, _ := json.MarshalIndent(trends, "", "  ")
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (d *DefaultAppDashboard) handleAPIExport(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	data, err := d.ExportDashboardData(format)
	if err != nil {
		http.Error(w, fmt.Sprintf("Export error: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (d *DefaultAppDashboard) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// WebSocket placeholder - would implement with gorilla/websocket or similar
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"websocket_not_implemented"}`))
}

func (d *DefaultAppDashboard) handleStatic(w http.ResponseWriter, r *http.Request) {
	// Static file serving placeholder
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("Static files not implemented"))
}

// Helper methods for internal components

func (ta *AppTrendAnalyzer) addDataPoint(metric string, value float64, timestamp time.Time) {
	// Don't store data if MaxDataPoints is 0 or negative
	if ta.MaxDataPoints <= 0 {
		return
	}

	if ta.HistoricalData[metric] == nil {
		ta.HistoricalData[metric] = []AppTimeSeriesPoint{}
	}

	// Add new data point
	ta.HistoricalData[metric] = append(ta.HistoricalData[metric], AppTimeSeriesPoint{
		Timestamp: timestamp,
		Value:     value,
	})

	// Trim to max data points
	if len(ta.HistoricalData[metric]) > ta.MaxDataPoints {
		ta.HistoricalData[metric] = ta.HistoricalData[metric][1:]
	}
}

func (am *AppAlertManager) evaluateAlerts(metrics *AppMetrics) {
	// Clear previous alerts to avoid accumulation
	am.ActiveAlerts = make(map[string]*AppAlert)

	for _, rule := range am.AlertRules {
		var value float64

		switch rule.Metric {
		case "error_rate":
			value = metrics.ErrorRate
		case "memory_usage":
			value = float64(metrics.MemoryUsage)
		case "cpu_usage":
			value = metrics.CPUUsage
		case "tests_executed":
			value = float64(metrics.TestsExecuted)
		case "tests_succeeded":
			value = float64(metrics.TestsSucceeded)
		case "tests_failed":
			value = float64(metrics.TestsFailed)
		case "goroutine_count":
			value = float64(metrics.GoroutineCount)
		default:
			continue
		}

		// Simple threshold evaluation
		triggered := false
		switch rule.Operator {
		case "gt":
			triggered = value > rule.Threshold
		case "lt":
			triggered = value < rule.Threshold
		case "gte":
			triggered = value >= rule.Threshold
		case "lte":
			triggered = value <= rule.Threshold
		case "eq":
			triggered = value == rule.Threshold
		}

		if triggered {
			alertID := fmt.Sprintf("%s-%d", rule.Name, time.Now().Unix())
			am.ActiveAlerts[alertID] = &AppAlert{
				ID:          alertID,
				Name:        rule.Name,
				Description: rule.Description,
				Severity:    rule.Severity,
				Status:      "FIRING",
				StartTime:   time.Now(),
				Value:       value,
				Threshold:   rule.Threshold,
				Labels:      rule.Labels,
			}
		}
	}
}

func (am *AppAlertManager) getActiveAlerts() []*AppAlert {
	alerts := []*AppAlert{}
	for _, alert := range am.ActiveAlerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

func (rtd *AppRealTimeData) update(metrics map[string]interface{}) {
	rtd.CurrentMetrics = metrics
	rtd.LastUpdate = time.Now()
}
