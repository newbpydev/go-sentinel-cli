// Package metrics provides functionality for collecting and exposing application metrics
package metrics

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"sync/atomic"
)

var (
	totalRequests uint64
	totalErrors   uint64
)

// Handler exposes Prometheus-style metrics
func Handler(w http.ResponseWriter, _ *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	req := atomic.LoadUint64(&totalRequests)
	err := atomic.LoadUint64(&totalErrors)

	w.Header().Set("Content-Type", "text/plain")

	// Write metrics with error checking
	if _, err := fmt.Fprintf(w, "# HELP go_mem_alloc_bytes Number of bytes allocated and still in use\n"); err != nil {
		log.Printf("Error writing metrics: %v", err)
		return
	}
	if _, err := fmt.Fprintf(w, "go_mem_alloc_bytes %d\n", memStats.Alloc); err != nil {
		log.Printf("Error writing metrics: %v", err)
		return
	}
	if _, err := fmt.Fprintf(w, "# HELP go_goroutines Number of goroutines\n"); err != nil {
		log.Printf("Error writing metrics: %v", err)
		return
	}
	if _, err := fmt.Fprintf(w, "go_goroutines %d\n", runtime.NumGoroutine()); err != nil {
		log.Printf("Error writing metrics: %v", err)
		return
	}
	if _, err := fmt.Fprintf(w, "# HELP api_total_requests Total HTTP requests\n"); err != nil {
		log.Printf("Error writing metrics: %v", err)
		return
	}
	if _, err := fmt.Fprintf(w, "api_total_requests %d\n", req); err != nil {
		log.Printf("Error writing metrics: %v", err)
		return
	}
	if _, err := fmt.Fprintf(w, "# HELP api_total_errors Total HTTP errors\n"); err != nil {
		log.Printf("Error writing metrics: %v", err)
		return
	}
	if _, err := fmt.Fprintf(w, "api_total_errors %d\n", err); err != nil {
		log.Printf("Error writing metrics: %v", err)
		return
	}
}

// Track increments request and error counters
func Track(status int) {
	atomic.AddUint64(&totalRequests, 1)
	if status >= 400 {
		atomic.AddUint64(&totalErrors, 1)
	}
}
