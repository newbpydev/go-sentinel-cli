package middleware

import (
	"net/http"

	"github.com/newbpydev/go-sentinel/internal/web/toast"
)

// ToastErrorHandler is middleware that converts errors to toast notifications
func ToastErrorHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a response wrapper to intercept writes
		ww := NewResponseWriter(w)
		
		// Call the next handler
		next.ServeHTTP(ww, r)
		
		// If there was an error (status >= 400), convert it to a toast notification
		// Only for HTMX requests to avoid affecting regular page loads
		if ww.Status() >= 400 && r.Header.Get("HX-Request") == "true" {
			// Get error message from response body
			errMsg := string(ww.Body())
			if errMsg == "" {
				errMsg = http.StatusText(ww.Status())
			}
			
			// Create toast notification
			t := toast.NewError(errMsg)
			t.AddHeader(w)
			
			// Set status to 200 to avoid HTMX error handling
			w.WriteHeader(http.StatusOK)
		}
	})
}

// ResponseWriter is a wrapper around http.ResponseWriter that captures status and body
type ResponseWriter struct {
	http.ResponseWriter
	status int
	body   []byte
}

// NewResponseWriter creates a new ResponseWriter
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

// WriteHeader captures the status code
func (rw *ResponseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

// Write captures the body
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return rw.ResponseWriter.Write(b)
}

// Status returns the captured status code
func (rw *ResponseWriter) Status() int {
	return rw.status
}

// Body returns the captured body
func (rw *ResponseWriter) Body() []byte {
	return rw.body
}
