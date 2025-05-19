package server

import (
	"net/http"

	"github.com/newbpydev/go-sentinel/internal/web/toast"
)

// ShowSuccessToast adds a success toast notification to the response
func (s *Server) ShowSuccessToast(w http.ResponseWriter, message string) {
	t := toast.NewSuccess(message)
	if err := t.AddHeader(w); err != nil {
		http.Error(w, "Failed to add toast header", http.StatusInternalServerError)
		return
	}
}

// ShowInfoToast adds an info toast notification to the response
func (s *Server) ShowInfoToast(w http.ResponseWriter, message string) {
	t := toast.NewInfo(message)
	if err := t.AddHeader(w); err != nil {
		http.Error(w, "Failed to add toast header", http.StatusInternalServerError)
		return
	}
}

// ShowWarningToast adds a warning toast notification to the response
func (s *Server) ShowWarningToast(w http.ResponseWriter, message string) {
	t := toast.NewWarning(message)
	if err := t.AddHeader(w); err != nil {
		http.Error(w, "Failed to add toast header", http.StatusInternalServerError)
		return
	}
}

// ShowErrorToast adds an error toast notification to the response
func (s *Server) ShowErrorToast(w http.ResponseWriter, message string) {
	t := toast.NewError(message)
	if err := t.AddHeader(w); err != nil {
		http.Error(w, "Failed to add toast header", http.StatusInternalServerError)
		return
	}
}
