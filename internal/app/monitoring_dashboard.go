package app

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// MonitoringDashboard provides advanced monitoring and visualization capabilities
type MonitoringDashboard struct {
	mu               sync.RWMutex
	metricsCollector *MetricsCollector
	alertManager     *AlertManager
	trendAnalyzer    *TrendAnalyzer
	config           *DashboardConfig
	server           *http.Server
	websocketClients map[string]*WebSocketClient
	realTimeData     *RealTimeData
}

// DashboardConfig configures the monitoring dashboard
type DashboardConfig struct {
	Port                int           `json:"port"`
	RefreshInterval     time.Duration `json:"refresh_interval"`
	MaxDataPoints       int           `json:"max_data_points"`
	EnableRealTime      bool          `json:"enable_real_time"`
	EnableAlerts        bool          `json:"enable_alerts"`
	ChartRetentionHours int           `json:"chart_retention_hours"`
	Theme               string        `json:"theme"` // "light", "dark", "auto"
}

// AlertManager handles monitoring alerts and notifications
type AlertManager struct {
	mu              sync.RWMutex
	activeAlerts    map[string]*Alert
	alertRules      []*AlertRule
	webhookURLs     []string
	silencedAlerts  map[string]time.Time
	escalationRules map[string]*EscalationRule
}

// Alert represents an active monitoring alert
type Alert struct {
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

// AlertRule defines conditions for triggering alerts
type AlertRule struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Metric      string            `json:"metric"`
	Operator    string            `json:"operator"` // "gt", "lt", "eq", "ne", "gte", "lte"
	Threshold   float64           `json:"threshold"`
	Duration    time.Duration     `json:"duration"`
	Severity    string            `json:"severity"`
	Labels      map[string]string `json:"labels"`
	Actions     []AlertAction     `json:"actions"`
}

// AlertAction defines what to do when an alert triggers
type AlertAction struct {
	Type   string                 `json:"type"` // "webhook", "email", "slack", "log"
	Config map[string]interface{} `json:"config"`
}

// EscalationRule defines alert escalation policies
type EscalationRule struct {
	AlertName     string        `json:"alert_name"`
	EscalateAfter time.Duration `json:"escalate_after"`
	Actions       []AlertAction `json:"actions"`
}

// TrendAnalyzer analyzes metric trends and patterns
type TrendAnalyzer struct {
	mu             sync.RWMutex
	historicalData map[string][]TimeSeriesPoint
	maxDataPoints  int
}

// TimeSeriesPoint represents a single data point in time series
type TimeSeriesPoint struct {
	Timestamp time.Time         `json:"timestamp"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
}

// WebSocketClient represents a connected WebSocket client
type WebSocketClient struct {
	ID         string
	Connection interface{} // WebSocket connection placeholder
	LastPing   time.Time
	Subscribed []string // Subscribed metric names
}

// RealTimeData holds real-time streaming data
type RealTimeData struct {
	mu             sync.RWMutex
	currentMetrics map[string]interface{}
	lastUpdate     time.Time
	subscribers    map[string][]*WebSocketClient
}

// DashboardMetrics provides dashboard-specific metrics aggregation
type DashboardMetrics struct {
	Overview     *OverviewMetrics     `json:"overview"`
	Performance  *PerformanceMetrics  `json:"performance"`
	TestMetrics  *TestSummaryMetrics  `json:"test_metrics"`
	SystemHealth *SystemHealthMetrics `json:"system_health"`
	Trends       *TrendMetrics        `json:"trends"`
	Alerts       []*Alert             `json:"alerts"`
}

// OverviewMetrics provides high-level system overview
type OverviewMetrics struct {
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

// PerformanceMetrics provides detailed performance data
type PerformanceMetrics struct {
	MemoryUsage    int64         `json:"memory_usage_mb"`
	CPUUsage       float64       `json:"cpu_usage_percent"`
	GoroutineCount int           `json:"goroutine_count"`
	CacheHitRate   float64       `json:"cache_hit_rate"`
	ResponseTime   time.Duration `json:"response_time_ms"`
	ThroughputRPS  float64       `json:"throughput_rps"`
	ErrorRate      float64       `json:"error_rate"`
	NetworkLatency time.Duration `json:"network_latency_ms"`
}

// TestSummaryMetrics provides test execution summaries
type TestSummaryMetrics struct {
	TotalExecutions int64    `json:"total_executions"`
	PassedTests     int64    `json:"passed_tests"`
	FailedTests     int64    `json:"failed_tests"`
	SkippedTests    int64    `json:"skipped_tests"`
	Coverage        float64  `json:"coverage_percent"`
	Flakiness       float64  `json:"flakiness_percent"`
	TopFailures     []string `json:"top_failures"`
	SlowestTests    []string `json:"slowest_tests"`
}

// SystemHealthMetrics provides system health indicators
type SystemHealthMetrics struct {
	DiskUsage        float64           `json:"disk_usage_percent"`
	NetworkStatus    string            `json:"network_status"`
	DependencyStatus map[string]string `json:"dependency_status"`
	ServiceStatus    map[string]string `json:"service_status"`
	HealthScore      float64           `json:"health_score"`
	RecentIncidents  []IncidentSummary `json:"recent_incidents"`
}

// TrendMetrics provides trend analysis data
type TrendMetrics struct {
	PerformanceTrend string                       `json:"performance_trend"` // "IMPROVING", "STABLE", "DEGRADING"
	TestSuccessTrend string                       `json:"test_success_trend"`
	ErrorTrend       string                       `json:"error_trend"`
	UsageTrend       string                       `json:"usage_trend"`
	TrendCharts      map[string][]TimeSeriesPoint `json:"trend_charts"`
	Predictions      map[string]float64           `json:"predictions"`
}

// IncidentSummary provides summary of recent incidents
type IncidentSummary struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Severity  string        `json:"severity"`
	Status    string        `json:"status"`
	StartTime time.Time     `json:"start_time"`
	Duration  time.Duration `json:"duration"`
	Impact    string        `json:"impact"`
}

// NewMonitoringDashboard creates a new monitoring dashboard
func NewMonitoringDashboard(collector *MetricsCollector, config *DashboardConfig) *MonitoringDashboard {
	if config == nil {
		config = DefaultDashboardConfig()
	}

	dashboard := &MonitoringDashboard{
		metricsCollector: collector,
		alertManager:     NewAlertManager(),
		trendAnalyzer:    NewTrendAnalyzer(config.MaxDataPoints),
		config:           config,
		websocketClients: make(map[string]*WebSocketClient),
		realTimeData:     NewRealTimeData(),
	}

	// Set up default alert rules
	dashboard.setupDefaultAlertRules()

	return dashboard
}

// DefaultDashboardConfig returns default dashboard configuration
func DefaultDashboardConfig() *DashboardConfig {
	return &DashboardConfig{
		Port:                3000,
		RefreshInterval:     5 * time.Second,
		MaxDataPoints:       1000,
		EnableRealTime:      true,
		EnableAlerts:        true,
		ChartRetentionHours: 24,
		Theme:               "dark",
	}
}

// Start starts the monitoring dashboard server
func (md *MonitoringDashboard) Start(ctx context.Context) error {
	log.Printf("Starting monitoring dashboard on port %d", md.config.Port)

	mux := http.NewServeMux()
	md.setupRoutes(mux)

	md.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", md.config.Port),
		Handler: mux,
	}

	// Start background tasks
	go md.collectTrendData(ctx)
	go md.processAlerts(ctx)
	go md.updateRealTimeData(ctx)

	// Start HTTP server
	go func() {
		if err := md.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Dashboard server error: %v", err)
		}
	}()

	log.Printf("Monitoring dashboard started at http://localhost:%d", md.config.Port)
	return nil
}

// Stop stops the monitoring dashboard
func (md *MonitoringDashboard) Stop(ctx context.Context) error {
	if md.server != nil {
		return md.server.Shutdown(ctx)
	}
	return nil
}

// GetDashboardMetrics returns comprehensive dashboard metrics
func (md *MonitoringDashboard) GetDashboardMetrics() *DashboardMetrics {
	baseMetrics := md.metricsCollector.GetMetrics()

	return &DashboardMetrics{
		Overview:     md.buildOverviewMetrics(baseMetrics),
		Performance:  md.buildPerformanceMetrics(baseMetrics),
		TestMetrics:  md.buildTestMetrics(baseMetrics),
		SystemHealth: md.buildSystemHealthMetrics(baseMetrics),
		Trends:       md.buildTrendMetrics(),
		Alerts:       md.alertManager.GetActiveAlerts(),
	}
}

// ExportDashboardData exports dashboard data in various formats
func (md *MonitoringDashboard) ExportDashboardData(format string) ([]byte, error) {
	metrics := md.GetDashboardMetrics()

	switch format {
	case "json":
		return json.MarshalIndent(metrics, "", "  ")
	case "csv":
		return md.exportCSVFormat(metrics)
	case "prometheus":
		return md.exportPrometheusFormat(metrics)
	default:
		return json.MarshalIndent(metrics, "", "  ")
	}
}

// Private methods for building metric components
func (md *MonitoringDashboard) buildOverviewMetrics(base *AppMetrics) *OverviewMetrics {
	successRate := float64(0)
	if base.TestsExecuted > 0 {
		successRate = float64(base.TestsSucceeded) / float64(base.TestsExecuted) * 100
	}

	activeAlerts := len(md.alertManager.GetActiveAlerts())
	criticalAlerts := len(md.alertManager.GetCriticalAlerts())

	return &OverviewMetrics{
		SystemStatus:   md.getSystemStatus(),
		Uptime:         base.Uptime,
		TotalTests:     base.TestsExecuted,
		SuccessRate:    successRate,
		AverageTime:    base.AverageTestTime,
		ActiveAlerts:   activeAlerts,
		CriticalAlerts: criticalAlerts,
		CurrentVersion: md.getCurrentVersion(),
	}
}

func (md *MonitoringDashboard) buildPerformanceMetrics(base *AppMetrics) *PerformanceMetrics {
	cacheHitRate := float64(0)
	total := base.CacheHits + base.CacheMisses
	if total > 0 {
		cacheHitRate = float64(base.CacheHits) / float64(total) * 100
	}

	return &PerformanceMetrics{
		MemoryUsage:    base.MemoryUsage / 1024 / 1024, // Convert to MB
		CPUUsage:       base.CPUUsage,
		GoroutineCount: base.GoroutineCount,
		CacheHitRate:   cacheHitRate,
		ResponseTime:   base.ProcessingDuration,
		ErrorRate:      base.ErrorRate,
	}
}

func (md *MonitoringDashboard) buildTestMetrics(base *AppMetrics) *TestSummaryMetrics {
	flakiness := md.calculateFlakiness()

	return &TestSummaryMetrics{
		TotalExecutions: base.TestsExecuted,
		PassedTests:     base.TestsSucceeded,
		FailedTests:     base.TestsFailed,
		SkippedTests:    base.TestsSkipped,
		Flakiness:       flakiness,
		TopFailures:     md.getTopFailures(),
		SlowestTests:    md.getSlowestTests(),
	}
}

func (md *MonitoringDashboard) buildSystemHealthMetrics(base *AppMetrics) *SystemHealthMetrics {
	return &SystemHealthMetrics{
		NetworkStatus:    "CONNECTED",
		DependencyStatus: md.checkDependencies(),
		ServiceStatus:    md.checkServices(),
		HealthScore:      md.calculateHealthScore(),
		RecentIncidents:  md.getRecentIncidents(),
	}
}

func (md *MonitoringDashboard) buildTrendMetrics() *TrendMetrics {
	return md.trendAnalyzer.GetTrendMetrics()
}

// Setup methods and helpers
func (md *MonitoringDashboard) setupRoutes(mux *http.ServeMux) {
	// Dashboard UI routes
	mux.HandleFunc("/", md.handleDashboard)
	mux.HandleFunc("/api/metrics", md.handleAPIMetrics)
	mux.HandleFunc("/api/alerts", md.handleAPIAlerts)
	mux.HandleFunc("/api/trends", md.handleAPITrends)
	mux.HandleFunc("/api/export", md.handleAPIExport)

	// Real-time WebSocket endpoint (placeholder)
	mux.HandleFunc("/ws", md.handleWebSocket)

	// Static assets
	mux.HandleFunc("/static/", md.handleStatic)
}

func (md *MonitoringDashboard) setupDefaultAlertRules() {
	rules := []*AlertRule{
		{
			Name:        "High Memory Usage",
			Description: "Memory usage exceeds 80%",
			Metric:      "memory_usage_percent",
			Operator:    "gt",
			Threshold:   80.0,
			Duration:    5 * time.Minute,
			Severity:    "HIGH",
			Actions: []AlertAction{
				{Type: "log", Config: map[string]interface{}{"level": "warn"}},
			},
		},
		{
			Name:        "High Error Rate",
			Description: "Error rate exceeds 5%",
			Metric:      "error_rate",
			Operator:    "gt",
			Threshold:   5.0,
			Duration:    2 * time.Minute,
			Severity:    "CRITICAL",
			Actions: []AlertAction{
				{Type: "log", Config: map[string]interface{}{"level": "error"}},
			},
		},
		{
			Name:        "Low Cache Hit Rate",
			Description: "Cache hit rate below 70%",
			Metric:      "cache_hit_rate",
			Operator:    "lt",
			Threshold:   70.0,
			Duration:    10 * time.Minute,
			Severity:    "MEDIUM",
			Actions: []AlertAction{
				{Type: "log", Config: map[string]interface{}{"level": "warn"}},
			},
		},
	}

	for _, rule := range rules {
		md.alertManager.AddAlertRule(rule)
	}
}

// HTTP handlers
func (md *MonitoringDashboard) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Serve the main dashboard HTML page
	dashboardHTML := md.generateDashboardHTML()
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashboardHTML))
}

func (md *MonitoringDashboard) handleAPIMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := md.GetDashboardMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

func (md *MonitoringDashboard) handleAPIAlerts(w http.ResponseWriter, r *http.Request) {
	alerts := md.alertManager.GetActiveAlerts()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(alerts)
}

func (md *MonitoringDashboard) handleAPITrends(w http.ResponseWriter, r *http.Request) {
	trends := md.trendAnalyzer.GetTrendMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trends)
}

func (md *MonitoringDashboard) handleAPIExport(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	data, err := md.ExportDashboardData(format)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var contentType string
	switch format {
	case "csv":
		contentType = "text/csv"
	case "prometheus":
		contentType = "text/plain"
	default:
		contentType = "application/json"
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}

func (md *MonitoringDashboard) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// WebSocket handler placeholder - would implement real-time streaming
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("WebSocket endpoint - implementation pending"))
}

func (md *MonitoringDashboard) handleStatic(w http.ResponseWriter, r *http.Request) {
	// Static file handler placeholder
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("Static files not implemented"))
}

// Background processing methods
func (md *MonitoringDashboard) collectTrendData(ctx context.Context) {
	ticker := time.NewTicker(md.config.RefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metrics := md.metricsCollector.GetMetrics()
			md.trendAnalyzer.AddDataPoint("memory_usage", float64(metrics.MemoryUsage))
			md.trendAnalyzer.AddDataPoint("error_rate", metrics.ErrorRate)
			md.trendAnalyzer.AddDataPoint("test_success_rate", md.calculateSuccessRate(metrics))
		}
	}
}

func (md *MonitoringDashboard) processAlerts(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			md.alertManager.EvaluateAlerts(md.metricsCollector.GetMetrics())
		}
	}
}

func (md *MonitoringDashboard) updateRealTimeData(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if md.config.EnableRealTime {
				metrics := md.metricsCollector.GetMetrics()
				md.realTimeData.Update(map[string]interface{}{
					"timestamp":    time.Now(),
					"memory_usage": metrics.MemoryUsage,
					"cpu_usage":    metrics.CPUUsage,
					"error_rate":   metrics.ErrorRate,
					"test_count":   metrics.TestsExecuted,
				})
			}
		}
	}
}

// Helper methods
func (md *MonitoringDashboard) generateDashboardHTML() string {
	// Enhanced dashboard HTML template with better styling and features
	return `<!DOCTYPE html>
<html>
<head>
    <title>Go Sentinel CLI - Advanced Monitoring Dashboard</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            background: linear-gradient(135deg, #1a1a1a 0%, #2d2d2d 100%);
            color: #fff;
            min-height: 100vh;
        }
        .container { max-width: 1400px; margin: 0 auto; padding: 20px; }
        .header {
            text-align: center;
            margin-bottom: 30px;
            background: rgba(255,255,255,0.05);
            padding: 20px;
            border-radius: 12px;
            backdrop-filter: blur(10px);
        }
        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
            background: linear-gradient(45deg, #4CAF50, #45a049);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        .status-bar {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            padding: 15px;
            background: rgba(255,255,255,0.1);
            border-radius: 8px;
            backdrop-filter: blur(5px);
        }
        .status-indicator {
            display: flex;
            align-items: center;
            gap: 8px;
        }
        .status-dot {
            width: 12px;
            height: 12px;
            border-radius: 50%;
            animation: pulse 2s infinite;
        }
        .status-healthy { background: #4CAF50; }
        .status-warning { background: #FF9800; }
        .status-critical { background: #f44336; }
        @keyframes pulse {
            0% { opacity: 1; }
            50% { opacity: 0.5; }
            100% { opacity: 1; }
        }
        .metrics-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
            gap: 20px;
            margin-bottom: 30px;
        }
        .metric-card {
            background: linear-gradient(145deg, #2d2d2d, #3a3a3a);
            padding: 25px;
            border-radius: 12px;
            border: 1px solid rgba(255,255,255,0.1);
            box-shadow: 0 8px 32px rgba(0,0,0,0.3);
            transition: transform 0.3s ease, box-shadow 0.3s ease;
            position: relative;
            overflow: hidden;
        }
        .metric-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 12px 40px rgba(0,0,0,0.4);
        }
        .metric-card::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: 3px;
            background: linear-gradient(90deg, #4CAF50, #45a049);
        }
        .metric-title {
            font-size: 16px;
            font-weight: 600;
            margin-bottom: 15px;
            color: #4CAF50;
            display: flex;
            align-items: center;
            gap: 8px;
        }
        .metric-value {
            font-size: 28px;
            font-weight: bold;
            margin-bottom: 8px;
            color: #fff;
        }
        .metric-subtitle {
            font-size: 12px;
            color: #aaa;
            margin-bottom: 10px;
        }
        .metric-trend {
            font-size: 14px;
            display: flex;
            align-items: center;
            gap: 5px;
        }
        .trend-up { color: #4CAF50; }
        .trend-down { color: #f44336; }
        .trend-stable { color: #FF9800; }
        .controls {
            display: flex;
            gap: 15px;
            margin-bottom: 20px;
            flex-wrap: wrap;
        }
        .btn {
            background: linear-gradient(145deg, #4CAF50, #45a049);
            color: white;
            border: none;
            padding: 12px 24px;
            border-radius: 8px;
            cursor: pointer;
            font-weight: 600;
            transition: all 0.3s ease;
            text-decoration: none;
            display: inline-block;
        }
        .btn:hover {
            background: linear-gradient(145deg, #45a049, #4CAF50);
            transform: translateY(-2px);
            box-shadow: 0 4px 12px rgba(76, 175, 80, 0.3);
        }
        .btn-secondary {
            background: linear-gradient(145deg, #666, #777);
        }
        .btn-secondary:hover {
            background: linear-gradient(145deg, #777, #666);
            box-shadow: 0 4px 12px rgba(119, 119, 119, 0.3);
        }
        .alerts-section {
            margin-top: 30px;
            padding: 20px;
            background: rgba(255,255,255,0.05);
            border-radius: 12px;
            backdrop-filter: blur(10px);
        }
        .alert-item {
            padding: 15px;
            margin: 10px 0;
            border-radius: 8px;
            border-left: 4px solid;
        }
        .alert-critical {
            background: rgba(244, 67, 54, 0.1);
            border-left-color: #f44336;
        }
        .alert-warning {
            background: rgba(255, 152, 0, 0.1);
            border-left-color: #FF9800;
        }
        .alert-info {
            background: rgba(76, 175, 80, 0.1);
            border-left-color: #4CAF50;
        }
        .footer {
            text-align: center;
            margin-top: 40px;
            padding: 20px;
            color: #aaa;
            border-top: 1px solid rgba(255,255,255,0.1);
        }
        .loading {
            display: inline-block;
            width: 20px;
            height: 20px;
            border: 3px solid rgba(255,255,255,0.3);
            border-radius: 50%;
            border-top-color: #4CAF50;
            animation: spin 1s ease-in-out infinite;
        }
        @keyframes spin {
            to { transform: rotate(360deg); }
        }
        .chart-container {
            height: 200px;
            background: rgba(255,255,255,0.05);
            border-radius: 8px;
            margin-top: 15px;
            display: flex;
            align-items: center;
            justify-content: center;
            color: #aaa;
        }
    </style>
    <script>
        let refreshInterval;
        let isAutoRefresh = true;

        function refreshData() {
            const loadingElements = document.querySelectorAll('.loading-indicator');
            loadingElements.forEach(el => el.style.display = 'inline-block');

            fetch('/api/metrics')
                .then(response => response.json())
                .then(data => {
                    updateMetrics(data);
                    updateStatus(data);
                    loadingElements.forEach(el => el.style.display = 'none');
                })
                .catch(error => {
                    console.error('Error fetching metrics:', error);
                    updateStatus({ overview: { system_status: 'ERROR' } });
                    loadingElements.forEach(el => el.style.display = 'none');
                });

            fetch('/api/alerts')
                .then(response => response.json())
                .then(alerts => updateAlerts(alerts))
                .catch(error => console.error('Error fetching alerts:', error));
        }

        function updateMetrics(data) {
            const elements = {
                'uptime': formatDuration(data.overview?.uptime || 0),
                'total-tests': (data.overview?.total_tests || 0).toLocaleString(),
                'success-rate': (data.overview?.success_rate || 0).toFixed(2) + '%',
                'memory-usage': (data.performance?.memory_usage_mb || 0) + ' MB',
                'error-rate': (data.performance?.error_rate || 0).toFixed(2) + '%',
                'active-alerts': data.overview?.active_alerts || 0,
                'cpu-usage': (data.performance?.cpu_usage_percent || 0).toFixed(1) + '%',
                'cache-hit-rate': (data.performance?.cache_hit_rate || 0).toFixed(1) + '%',
                'health-score': (data.system_health?.health_score || 0).toFixed(1)
            };

            Object.entries(elements).forEach(([id, value]) => {
                const element = document.getElementById(id);
                if (element) element.textContent = value;
            });
        }

        function updateStatus(data) {
            const statusElement = document.getElementById('system-status');
            const statusDot = document.getElementById('status-dot');
            const status = data.overview?.system_status || 'UNKNOWN';

            statusElement.textContent = status;
            statusDot.className = 'status-dot ' + getStatusClass(status);
        }

        function updateAlerts(alerts) {
            const container = document.getElementById('alerts-container');
            if (!container) return;

            if (!alerts || alerts.length === 0) {
                container.innerHTML = '<div class="alert-item alert-info">✅ No active alerts</div>';
                return;
            }

            container.innerHTML = alerts.map(alert =>
                '<div class="alert-item alert-' + getSeverityClass(alert.severity) + '">' +
                '<strong>' + alert.name + '</strong><br>' +
                alert.description + '<br>' +
                '<small>Started: ' + new Date(alert.start_time).toLocaleString() + '</small>' +
                '</div>'
            ).join('');
        }

        function getStatusClass(status) {
            switch(status.toLowerCase()) {
                case 'healthy': return 'status-healthy';
                case 'warning': case 'degraded': return 'status-warning';
                case 'critical': case 'unhealthy': case 'error': return 'status-critical';
                default: return 'status-warning';
            }
        }

        function getSeverityClass(severity) {
            switch(severity.toLowerCase()) {
                case 'critical': case 'high': return 'critical';
                case 'medium': case 'warning': return 'warning';
                default: return 'info';
            }
        }

        function formatDuration(seconds) {
            const days = Math.floor(seconds / 86400);
            const hours = Math.floor((seconds % 86400) / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);

            if (days > 0) return days + 'd ' + hours + 'h';
            if (hours > 0) return hours + 'h ' + minutes + 'm';
            return minutes + 'm';
        }

        function toggleAutoRefresh() {
            isAutoRefresh = !isAutoRefresh;
            const button = document.getElementById('auto-refresh-btn');

            if (isAutoRefresh) {
                refreshInterval = setInterval(refreshData, 5000);
                button.textContent = '⏸️ Pause Auto-Refresh';
                button.className = 'btn btn-secondary';
            } else {
                clearInterval(refreshInterval);
                button.textContent = '▶️ Resume Auto-Refresh';
                button.className = 'btn';
            }
        }

        function exportData(format) {
            window.open('/api/export?format=' + format, '_blank');
        }

        window.onload = function() {
            refreshData();
            refreshInterval = setInterval(refreshData, 5000);

            // Update timestamp every second
            setInterval(() => {
                document.getElementById('last-update').textContent =
                    'Last updated: ' + new Date().toLocaleTimeString();
            }, 1000);
        };

        window.onbeforeunload = function() {
            if (refreshInterval) clearInterval(refreshInterval);
        };
    </script>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 Go Sentinel CLI - Advanced Monitoring Dashboard</h1>
            <div class="status-bar">
                <div class="status-indicator">
                    <div id="status-dot" class="status-dot status-warning"></div>
                    <span>System Status: <strong id="system-status">Loading...</strong></span>
                </div>
                <div id="last-update">Last updated: --:--:--</div>
            </div>
        </div>

        <div class="controls">
            <button class="btn" onclick="refreshData()">🔄 Refresh Now</button>
            <button id="auto-refresh-btn" class="btn btn-secondary" onclick="toggleAutoRefresh()">⏸️ Pause Auto-Refresh</button>
            <button class="btn" onclick="exportData('json')">📊 Export JSON</button>
            <button class="btn" onclick="exportData('csv')">📈 Export CSV</button>
            <a href="/api/trends" class="btn" target="_blank">📉 View Trends</a>
        </div>

        <div class="metrics-grid">
            <div class="metric-card">
                <div class="metric-title">⏱️ System Uptime</div>
                <div class="metric-value" id="uptime">Loading... <span class="loading loading-indicator" style="display:none;"></span></div>
                <div class="metric-subtitle">Time since last restart</div>
            </div>
            <div class="metric-card">
                <div class="metric-title">🧪 Total Tests</div>
                <div class="metric-value" id="total-tests">Loading... <span class="loading loading-indicator" style="display:none;"></span></div>
                <div class="metric-subtitle">Tests executed since startup</div>
            </div>
            <div class="metric-card">
                <div class="metric-title">✅ Success Rate</div>
                <div class="metric-value" id="success-rate">Loading... <span class="loading loading-indicator" style="display:none;"></span></div>
                <div class="metric-subtitle">Percentage of passing tests</div>
            </div>
            <div class="metric-card">
                <div class="metric-title">💾 Memory Usage</div>
                <div class="metric-value" id="memory-usage">Loading... <span class="loading loading-indicator" style="display:none;"></span></div>
                <div class="metric-subtitle">Current memory consumption</div>
            </div>
            <div class="metric-card">
                <div class="metric-title">🔥 CPU Usage</div>
                <div class="metric-value" id="cpu-usage">Loading... <span class="loading loading-indicator" style="display:none;"></span></div>
                <div class="metric-subtitle">Current CPU utilization</div>
            </div>
            <div class="metric-card">
                <div class="metric-title">❌ Error Rate</div>
                <div class="metric-value" id="error-rate">Loading... <span class="loading loading-indicator" style="display:none;"></span></div>
                <div class="metric-subtitle">Percentage of failed operations</div>
            </div>
            <div class="metric-card">
                <div class="metric-title">⚡ Cache Hit Rate</div>
                <div class="metric-value" id="cache-hit-rate">Loading... <span class="loading loading-indicator" style="display:none;"></span></div>
                <div class="metric-subtitle">Cache efficiency percentage</div>
            </div>
            <div class="metric-card">
                <div class="metric-title">🚨 Active Alerts</div>
                <div class="metric-value" id="active-alerts">Loading... <span class="loading loading-indicator" style="display:none;"></span></div>
                <div class="metric-subtitle">Current system alerts</div>
            </div>
            <div class="metric-card">
                <div class="metric-title">💚 Health Score</div>
                <div class="metric-value" id="health-score">Loading... <span class="loading loading-indicator" style="display:none;"></span></div>
                <div class="metric-subtitle">Overall system health (0-100)</div>
            </div>
        </div>

        <div class="alerts-section">
            <h2>🚨 Active Alerts</h2>
            <div id="alerts-container">
                <div class="alert-item alert-info">Loading alerts...</div>
            </div>
        </div>

        <div class="footer">
            <p>Go Sentinel CLI Advanced Monitoring Dashboard v2.0</p>
            <p>Real-time monitoring with auto-refresh every 5 seconds</p>
        </div>
    </div>
</body>
</html>`
}

// Placeholder helper methods (would be implemented with real logic)
func (md *MonitoringDashboard) getSystemStatus() string {
	return "HEALTHY"
}

func (md *MonitoringDashboard) getCurrentVersion() string {
	return "v2.0.0"
}

func (md *MonitoringDashboard) calculateFlakiness() float64 {
	return 2.5 // Placeholder
}

func (md *MonitoringDashboard) getTopFailures() []string {
	return []string{"auth_test", "network_test", "db_test"}
}

func (md *MonitoringDashboard) getSlowestTests() []string {
	return []string{"integration_test", "e2e_test", "load_test"}
}

func (md *MonitoringDashboard) checkDependencies() map[string]string {
	return map[string]string{
		"database": "HEALTHY",
		"redis":    "HEALTHY",
		"api":      "HEALTHY",
	}
}

func (md *MonitoringDashboard) checkServices() map[string]string {
	return map[string]string{
		"test_runner":  "RUNNING",
		"file_watcher": "RUNNING",
		"cache":        "RUNNING",
	}
}

func (md *MonitoringDashboard) calculateHealthScore() float64 {
	return 95.5
}

func (md *MonitoringDashboard) getRecentIncidents() []IncidentSummary {
	return []IncidentSummary{
		{
			ID:        "INC-001",
			Title:     "High Memory Usage",
			Severity:  "MEDIUM",
			Status:    "RESOLVED",
			StartTime: time.Now().Add(-2 * time.Hour),
			Duration:  15 * time.Minute,
			Impact:    "Performance degradation",
		},
	}
}

func (md *MonitoringDashboard) calculateSuccessRate(metrics *AppMetrics) float64 {
	if metrics.TestsExecuted == 0 {
		return 0
	}
	return float64(metrics.TestsSucceeded) / float64(metrics.TestsExecuted) * 100
}

func (md *MonitoringDashboard) exportCSVFormat(metrics *DashboardMetrics) ([]byte, error) {
	// Placeholder CSV export
	csv := "metric,value\n"
	csv += fmt.Sprintf("uptime,%d\n", int64(metrics.Overview.Uptime.Seconds()))
	csv += fmt.Sprintf("total_tests,%d\n", metrics.Overview.TotalTests)
	csv += fmt.Sprintf("success_rate,%.2f\n", metrics.Overview.SuccessRate)
	csv += fmt.Sprintf("memory_usage_mb,%d\n", metrics.Performance.MemoryUsage)
	csv += fmt.Sprintf("error_rate,%.2f\n", metrics.Performance.ErrorRate)
	return []byte(csv), nil
}

func (md *MonitoringDashboard) exportPrometheusFormat(metrics *DashboardMetrics) ([]byte, error) {
	// Placeholder Prometheus export
	prom := fmt.Sprintf("go_sentinel_uptime_seconds %d\n", int64(metrics.Overview.Uptime.Seconds()))
	prom += fmt.Sprintf("go_sentinel_total_tests %d\n", metrics.Overview.TotalTests)
	prom += fmt.Sprintf("go_sentinel_success_rate %.2f\n", metrics.Overview.SuccessRate)
	prom += fmt.Sprintf("go_sentinel_memory_usage_mb %d\n", metrics.Performance.MemoryUsage)
	prom += fmt.Sprintf("go_sentinel_error_rate %.2f\n", metrics.Performance.ErrorRate)
	return []byte(prom), nil
}

// Supporting types and methods for AlertManager and TrendAnalyzer
func NewAlertManager() *AlertManager {
	return &AlertManager{
		activeAlerts:    make(map[string]*Alert),
		alertRules:      make([]*AlertRule, 0),
		silencedAlerts:  make(map[string]time.Time),
		escalationRules: make(map[string]*EscalationRule),
	}
}

func (am *AlertManager) AddAlertRule(rule *AlertRule) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.alertRules = append(am.alertRules, rule)
}

func (am *AlertManager) GetActiveAlerts() []*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make([]*Alert, 0, len(am.activeAlerts))
	for _, alert := range am.activeAlerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

func (am *AlertManager) GetCriticalAlerts() []*Alert {
	alerts := am.GetActiveAlerts()
	critical := make([]*Alert, 0)
	for _, alert := range alerts {
		if alert.Severity == "CRITICAL" {
			critical = append(critical, alert)
		}
	}
	return critical
}

func (am *AlertManager) EvaluateAlerts(metrics *AppMetrics) {
	// Placeholder alert evaluation logic
	// Would implement real alert evaluation against metrics
}

func NewTrendAnalyzer(maxDataPoints int) *TrendAnalyzer {
	return &TrendAnalyzer{
		historicalData: make(map[string][]TimeSeriesPoint),
		maxDataPoints:  maxDataPoints,
	}
}

func (ta *TrendAnalyzer) AddDataPoint(metric string, value float64) {
	ta.mu.Lock()
	defer ta.mu.Unlock()

	point := TimeSeriesPoint{
		Timestamp: time.Now(),
		Value:     value,
	}

	ta.historicalData[metric] = append(ta.historicalData[metric], point)

	// Keep only max data points
	if len(ta.historicalData[metric]) > ta.maxDataPoints {
		ta.historicalData[metric] = ta.historicalData[metric][1:]
	}
}

func (ta *TrendAnalyzer) GetTrendMetrics() *TrendMetrics {
	ta.mu.RLock()
	defer ta.mu.RUnlock()

	return &TrendMetrics{
		PerformanceTrend: "STABLE",
		TestSuccessTrend: "IMPROVING",
		ErrorTrend:       "STABLE",
		UsageTrend:       "INCREASING",
		TrendCharts:      ta.historicalData,
		Predictions:      make(map[string]float64),
	}
}

func NewRealTimeData() *RealTimeData {
	return &RealTimeData{
		currentMetrics: make(map[string]interface{}),
		subscribers:    make(map[string][]*WebSocketClient),
	}
}

func (rtd *RealTimeData) Update(metrics map[string]interface{}) {
	rtd.mu.Lock()
	defer rtd.mu.Unlock()

	rtd.currentMetrics = metrics
	rtd.lastUpdate = time.Now()

	// Would notify WebSocket subscribers here
}
