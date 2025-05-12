package metrics

import (
	"net/http"
	"runtime"
	"sync/atomic"
	"fmt"
)

var (
	totalRequests uint64
	totalErrors   uint64
)

// Handler exposes Prometheus-style metrics
func Handler(w http.ResponseWriter, r *http.Request) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	req := atomic.LoadUint64(&totalRequests)
	err := atomic.LoadUint64(&totalErrors)

	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "# HELP go_mem_alloc_bytes Number of bytes allocated and still in use\n")
	fmt.Fprintf(w, "go_mem_alloc_bytes %d\n", memStats.Alloc)
	fmt.Fprintf(w, "# HELP go_goroutines Number of goroutines\n")
	fmt.Fprintf(w, "go_goroutines %d\n", runtime.NumGoroutine())
	fmt.Fprintf(w, "# HELP api_total_requests Total HTTP requests\n")
	fmt.Fprintf(w, "api_total_requests %d\n", req)
	fmt.Fprintf(w, "# HELP api_total_errors Total HTTP errors\n")
	fmt.Fprintf(w, "api_total_errors %d\n", err)
}

// Track increments request and error counters
func Track(status int) {
	atomic.AddUint64(&totalRequests, 1)
	if status >= 400 {
		atomic.AddUint64(&totalErrors, 1)
	}
}
