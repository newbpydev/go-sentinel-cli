package toast

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Toast levels
const (
	Info    = "info"
	Success = "success"
	Warning = "warning"
	Error   = "error"
)

// Toast represents a toast notification
type Toast struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Timeout int    `json:"timeout"` // milliseconds
}

// New creates a new toast notification
func New(level string, message string, timeout int) Toast {
	return Toast{
		Level:   level,
		Message: message,
		Timeout: timeout,
	}
}

// NewInfo creates an info toast
func NewInfo(message string) Toast {
	return New(Info, message, 3000)
}

// NewSuccess creates a success toast
func NewSuccess(message string) Toast {
	return New(Success, message, 3000)
}

// NewWarning creates a warning toast
func NewWarning(message string) Toast {
	return New(Warning, message, 4000)
}

// NewError creates an error toast
func NewError(message string) Toast {
	return New(Error, message, 5000)
}

// ToJSON converts the toast to a JSON string
func (t Toast) ToJSON() (string, error) {
	eventData := map[string]Toast{
		"showToast": t,
	}
	
	data, err := json.Marshal(eventData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal toast: %w", err)
	}
	
	return string(data), nil
}

// AddHeader adds the toast as an HX-Trigger header to the response
func (t Toast) AddHeader(w http.ResponseWriter) error {
	jsonData, err := t.ToJSON()
	if err != nil {
		return err
	}
	
	// Set HX-Trigger header to trigger the toast notification
	w.Header().Set("HX-Trigger", jsonData)
	
	// If this is an error toast, also prevent HTML swap
	if t.Level == Error {
		w.Header().Set("HX-Reswap", "none")
	}
	
	return nil
}

// Error implements the error interface, allowing Toast to be returned as an error
func (t Toast) Error() string {
	return fmt.Sprintf("%s: %s", t.Level, t.Message)
}
