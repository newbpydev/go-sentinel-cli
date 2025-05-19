package middleware

import (
	"net/http"
	"sync"
	"time"
)

type rateLimiter struct {
	mu        sync.Mutex
	requests  map[string][]time.Time
	limit     int
	window    time.Duration
}

func newRateLimiter(limit int, window time.Duration) *rateLimiter {
	return &rateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	windowStart := now.Add(-rl.window)
	reqs := rl.requests[ip]
	// Remove old requests
	var recent []time.Time
	for _, t := range reqs {
		if t.After(windowStart) {
			recent = append(recent, t)
		}
	}
	if len(recent) >= rl.limit {
		rl.requests[ip] = recent
		return false
	}
	recent = append(recent, now)
	rl.requests[ip] = recent
	return true
}

// RateLimit returns a middleware that limits requests per-IP
func RateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	rl := newRateLimiter(limit, window)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if !rl.allow(ip) {
				w.WriteHeader(http.StatusTooManyRequests)
				if _, err := w.Write([]byte("rate limit exceeded")); err != nil {
					// Log the error if we can't write the response
					http.Error(w, "internal server error", http.StatusInternalServerError)
				}
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
