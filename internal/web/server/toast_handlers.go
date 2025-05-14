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
		t.AddHeader(w)
	case "error":
		t := toast.NewError("An error occurred during the operation.")
		t.AddHeader(w)
	case "warning":
		t := toast.NewWarning("Please be cautious with this operation.")
		t.AddHeader(w)
	case "info":
		t := toast.NewInfo("This is an informational message.")
		t.AddHeader(w)
	default:
		// If no valid type is provided, show all types sequentially
		t := toast.NewInfo("This is how toast notifications look.")
		t.AddHeader(w)
	}

	// Return a simple success response
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success"}`))
}
