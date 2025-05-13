package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logger is a middleware that logs request details
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Process the request
		next.ServeHTTP(w, r)
		
		// Log details after the request is complete
		log.Printf(
			"%s %s %s %s",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			time.Since(start),
		)
	})
}
