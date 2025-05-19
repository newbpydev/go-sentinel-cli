package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strconv"
)

// SettingsHandler handles requests related to application settings
type SettingsHandler struct {
	templates *template.Template
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(tmpl *template.Template) *SettingsHandler {
	return &SettingsHandler{
		templates: tmpl,
	}
}

// Settings represents the application settings
type Settings struct {
	// Test Runner settings
	TestTimeout   int  `json:"testTimeout"`
	ParallelTests int  `json:"parallelTests"`
	VerboseOutput bool `json:"verboseOutput"`
	AutoRun       bool `json:"autoRun"`

	// Notification settings
	NotifySuccess        bool   `json:"notifySuccess"`
	NotifyFailure        bool   `json:"notifyFailure"`
	NotificationDuration int    `json:"notificationDuration"`
	NotificationPosition string `json:"notificationPosition"`

	// Coverage settings
	CollectCoverage   bool    `json:"collectCoverage"`
	CoverageThreshold float64 `json:"coverageThreshold"`
	CoverageExclude   string  `json:"coverageExclude"`

	// Appearance settings
	Theme             string `json:"theme"`
	FontSize          string `json:"fontSize"`
	AnimationsEnabled bool   `json:"animationsEnabled"`

	// Advanced settings
	LogLevel      string `json:"logLevel"`
	CacheDuration int    `json:"cacheDuration"`
	DataDirectory string `json:"dataDirectory"`
	DebugMode     bool   `json:"debugMode"`
}

// ValidationResponse represents the response for settings validation
type ValidationResponse struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors,omitempty"`
	Message string   `json:"message,omitempty"`
}

// GetDefaultSettings returns the default application settings
func GetDefaultSettings() Settings {
	return Settings{
		// Test Runner defaults
		TestTimeout:   30,
		ParallelTests: 4,
		VerboseOutput: true,
		AutoRun:       true,

		// Notification defaults
		NotifySuccess:        true,
		NotifyFailure:        true,
		NotificationDuration: 5,
		NotificationPosition: "top-right",

		// Coverage defaults
		CollectCoverage:   true,
		CoverageThreshold: 80.0,
		CoverageExclude:   "vendor/\n.git/\n*_test.go",

		// Appearance defaults
		Theme:             "dark",
		FontSize:          "medium",
		AnimationsEnabled: true,

		// Advanced defaults
		LogLevel:      "info",
		CacheDuration: 30,
		DataDirectory: "./data",
		DebugMode:     false,
	}
}

// GetSettings handles requests for the current settings
func (h *SettingsHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would fetch settings from storage
	// For now, we'll use default settings
	settings := GetDefaultSettings()

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(settings); err != nil {
		log.Printf("failed to encode settings: %v", err)
	}
}

// ValidateSettings handles validation of settings before saving
func (h *SettingsHandler) ValidateSettings(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		sendValidationError(w, "Failed to parse form data")
		return
	}

	// Extract and validate settings
	testTimeout := getFormInt(r, "testTimeout", 30)
	if testTimeout < 1 || testTimeout > 300 {
		sendValidationError(w, "Test timeout must be between 1 and 300 seconds")
		return
	}

	parallelTests := getFormInt(r, "parallelTests", 4)
	if parallelTests < 1 || parallelTests > 32 {
		sendValidationError(w, "Parallel tests must be between 1 and 32")
		return
	}

	coverageThreshold := getFormFloat(r, "coverageThreshold", 80.0)
	if coverageThreshold < 0 || coverageThreshold > 100 {
		sendValidationError(w, "Coverage threshold must be between 0 and 100 percent")
		return
	}

	// If validation passes, send success response
	response := ValidationResponse{
		Valid:   true,
		Message: "Settings are valid",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

// SaveSettings handles saving of settings
func (h *SettingsHandler) SaveSettings(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	if err := r.ParseForm(); err != nil {
		sendValidationError(w, "Failed to parse form data")
		return
	}

	// Validate settings first
	testTimeout := getFormInt(r, "testTimeout", 30)
	if testTimeout < 1 || testTimeout > 300 {
		sendValidationError(w, "Test timeout must be between 1 and 300 seconds")
		return
	}

	// In a real implementation, this would save settings to storage
	// For now, we'll just return success

	// Create feedback message
	feedbackHTML := `<div class="settings-feedback success">Settings saved successfully</div>`
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(feedbackHTML)); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// ResetSettings handles resetting settings to defaults
func (h *SettingsHandler) ResetSettings(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would reset settings in storage to defaults
	// For now, we'll just return success

	// Redirect to settings page with default values
	http.Redirect(w, r, "/settings", http.StatusSeeOther)
}

// ClearCache handles clearing the application cache
func (h *SettingsHandler) ClearCache(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would clear the cache
	// For now, we'll just return success

	// Create feedback message
	feedbackHTML := `<div class="settings-feedback success">Cache cleared successfully</div>`
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(feedbackHTML)); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// ClearHistory handles clearing the test history
func (h *SettingsHandler) ClearHistory(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would clear the test history
	// For now, we'll just return success

	// Create feedback message
	feedbackHTML := `<div class="settings-feedback success">Test history cleared successfully</div>`
	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(feedbackHTML)); err != nil {
		log.Printf("Error writing response: %v", err)
	}
}

// Helper function to send validation error response
func sendValidationError(w http.ResponseWriter, message string) {
	response := ValidationResponse{
		Valid:   false,
		Errors:  []string{message},
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

// Helper function to get integer from form
func getFormInt(r *http.Request, key string, defaultValue int) int {
	valueStr := r.FormValue(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// Helper function to get float from form
func getFormFloat(r *http.Request, key string, defaultValue float64) float64 {
	valueStr := r.FormValue(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		return defaultValue
	}

	return value
}
