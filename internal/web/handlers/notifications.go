package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	
	"github.com/newbpydev/go-sentinel/internal/web/toast"
)

// NotificationResponse is a standard API response for notification test endpoint
// This follows best practices for API responses (JSON, status, message)
type NotificationResponse struct {
	Status  string `json:"status"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

// HandleTestNotification handles POST requests to /api/notifications/test securely
func HandleTestNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(NotificationResponse{
			Status:  "error",
			Type:    "",
			Message: "Method not allowed",
		})
		return
	}

	typeParam := r.URL.Query().Get("type")
	typeParam = strings.ToLower(typeParam)
	if typeParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(NotificationResponse{
			Status:  "error",
			Type:    "",
			Message: "Missing notification type",
		})
		return
	}

	// Allow only specific types
	allowed := map[string]bool{"success": true, "error": true, "warning": true, "info": true}
	if !allowed[typeParam] {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(NotificationResponse{
			Status:  "error",
			Type:    typeParam,
			Message: "Invalid notification type",
		})
		return
	}
	
	// Create toast notification based on type
	var t toast.Toast
	switch typeParam {
	case "success":
		t = toast.NewSuccess("Success! The operation was completed successfully.")
	case "error":
		t = toast.NewError("Error: Something went wrong with the operation.")
	case "warning":
		t = toast.NewWarning("Warning: This action may have consequences.")
	case "info":
		t = toast.NewInfo("Info: This is an informational notification.")
	}
	
	// Add the toast notification to the response headers
	t.AddHeader(w)
	
	// Return a JSON response as well (for API compatibility)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NotificationResponse{
		Status:  "ok",
		Type:    typeParam,
		Message: "Notification type '" + typeParam + "' triggered successfully.",
	})
}
