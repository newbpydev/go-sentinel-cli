package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"bytes"
)

// ValidateJSON returns a middleware that rejects requests with malformed JSON bodies
func ValidateJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			if r.Body != nil {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "invalid body", http.StatusBadRequest)
					return
				}
				var js json.RawMessage
				if err := json.Unmarshal(body, &js); err != nil {
					http.Error(w, "malformed JSON", http.StatusBadRequest)
					return
				}
				// Replace body for downstream handlers
				r.Body = io.NopCloser(bytes.NewReader(body))
			}
		}
		next.ServeHTTP(w, r)
	})
}
