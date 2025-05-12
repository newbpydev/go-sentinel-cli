package models

import "errors"

// Validate checks that the APITestEvent has the required fields for API usage.
func (e APITestEvent) Validate() error {
	if e.Action == "" {
		return errors.New("action is required")
	}
	if e.Package == "" {
		return errors.New("package is required")
	}
	// Elapsed, Output, Test, and Time can be optional
	return nil
}
