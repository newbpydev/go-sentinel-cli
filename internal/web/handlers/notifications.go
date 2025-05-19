package handlers

import (
	"encoding/json"
	"log"
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
		if err := json.NewEncoder(w).Encode(NotificationResponse{
			Status:  "error",
			Type:    "",
			Message: "Method not allowed",
		}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
		return
	}

	typeParam := r.URL.Query().Get("type")
	typeParam = strings.ToLower(typeParam)
	if typeParam == "" {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NotificationResponse{
			Status:  "error",
			Type:    "",
			Message: "Missing notification type",
		}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
		return
	}

	// Allow only specific types
	allowed := map[string]bool{"success": true, "error": true, "warning": true, "info": true}
	if !allowed[typeParam] {
		w.WriteHeader(http.StatusBadRequest)
		if err := json.NewEncoder(w).Encode(NotificationResponse{
			Status:  "error",
			Type:    typeParam,
			Message: "Invalid notification type",
		}); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
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
	if err := t.AddHeader(w); err != nil {
		log.Printf("Error adding toast header: %v", err)
	}
	
	// Return a JSON response as well (for API compatibility)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(NotificationResponse{
		Status:  "ok",
		Type:    typeParam,
		Message: "Notification type '" + typeParam + "' triggered successfully.",
	}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
