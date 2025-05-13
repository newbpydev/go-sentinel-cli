package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(NotificationResponse{
		Status:  "ok",
		Type:    typeParam,
		Message: "Notification type '" + typeParam + "' triggered successfully.",
	})
}
