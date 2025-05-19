package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

// MetricsHandler handles requests for test metrics
type MetricsHandler struct {
	// This would be connected to the actual metrics collection in the real implementation
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

// GetMetrics returns the current test metrics
func (h *MetricsHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	// Get metrics data (mock for now)
	metrics := getMetricsData()

	// For HTMX requests, render HTML
	if r.Header.Get("HX-Request") == "true" {
		h.renderMetricsHTML(w, metrics)
		return
	}

	// For API requests, return JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// renderMetricsHTML renders HTML for metrics (for HTMX)
func (h *MetricsHandler) renderMetricsHTML(w http.ResponseWriter, metrics map[string]interface{}) {
	// In a real implementation, this would use a template engine
	// For now, we'll render simple HTML
	w.Header().Set("Content-Type", "text/html")

	html := `<div class="stats-cards">
		<!-- Total Tests Card -->
		<div class="stat-card" role="region" aria-label="Total Tests">
			<div class="stat-card-content">
				<div class="stat-title">Total Tests</div>
				<div class="stat-value">` + metrics["TotalTests"].(string) + `</div>
				<div class="stat-change">` + metrics["TotalChange"].(string) + `</div>
			</div>
		</div>
		
		<!-- Passing Tests Card -->
		<div class="stat-card success" role="region" aria-label="Passing Tests">
			<div class="stat-card-content">
				<div class="stat-title">Passing</div>
				<div class="stat-value">` + metrics["Passing"].(string) + `</div>
				<div class="stat-change">` + metrics["PassingRate"].(string) + `</div>
			</div>
		</div>
		
		<!-- Failing Tests Card -->
		<div class="stat-card error" role="region" aria-label="Failing Tests">
			<div class="stat-card-content">
				<div class="stat-title">Failing</div>
				<div class="stat-value">` + metrics["Failing"].(string) + `</div>
				<div class="stat-change">` + metrics["FailingChange"].(string) + `</div>
			</div>
		</div>
		
		<!-- Average Duration Card -->
		<div class="stat-card" role="region" aria-label="Average Test Duration">
			<div class="stat-card-content">
				<div class="stat-title">Avg. Duration</div>
				<div class="stat-value">` + metrics["Duration"].(string) + `</div>
				<div class="stat-change">` + metrics["DurationChange"].(string) + `</div>
			</div>
		</div>
	</div>`

	if _, err := w.Write([]byte(html)); err != nil {
		log.Printf("Error writing metrics HTML response: %v", err)
	}
}

// getMetricsData returns mock metrics data
func getMetricsData() map[string]interface{} {
	return map[string]interface{}{
		"TotalTests":     "128",
		"TotalChange":    "+3 since yesterday",
		"Passing":        "119",
		"PassingRate":    "93% success rate",
		"Failing":        "9",
		"FailingChange":  "-2 since yesterday",
		"Duration":       "1.2s",
		"DurationChange": "-0.3s from last run",
		"LastUpdated":    time.Now(),
	}
}
