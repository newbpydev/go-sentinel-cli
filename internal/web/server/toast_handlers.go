package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/newbpydev/go-sentinel/internal/web/toast"
)

// handleToastTest is a handler for testing different types of toast notifications
func (s *Server) handleToastTest(w http.ResponseWriter, r *http.Request) {
	toastType := chi.URLParam(r, "type")

	switch toastType {
	case "success":
		t := toast.NewSuccess("Operation completed successfully!")
		if err := t.AddHeader(w); err != nil {
			http.Error(w, "Failed to add header", http.StatusInternalServerError)
			return
		}
	case "error":
		t := toast.NewError("An error occurred during the operation.")
		if err := t.AddHeader(w); err != nil {
			http.Error(w, "Failed to add header", http.StatusInternalServerError)
			return
		}
	case "warning":
		t := toast.NewWarning("Please be cautious with this operation.")
		if err := t.AddHeader(w); err != nil {
			http.Error(w, "Failed to add header", http.StatusInternalServerError)
			return
		}
	case "info":
		t := toast.NewInfo("This is an informational message.")
		if err := t.AddHeader(w); err != nil {
			http.Error(w, "Failed to add header", http.StatusInternalServerError)
			return
		}
	default:
		// If no valid type is provided, show all types sequentially
		t := toast.NewInfo("This is how toast notifications look.")
		if err := t.AddHeader(w); err != nil {
			http.Error(w, "Failed to add header", http.StatusInternalServerError)
			return
		}
	}

	// Return a simple success response
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write([]byte(`{"status": "success"}`)); err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
