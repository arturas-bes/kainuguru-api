package models

import "fmt"

// ValidationError represents a validation error for model fields
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation error for field '%s': %s", ve.Field, ve.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

// Error implements the error interface for multiple validation errors
func (ves ValidationErrors) Error() string {
	if len(ves) == 0 {
		return "no validation errors"
	}
	if len(ves) == 1 {
		return ves[0].Error()
	}
	return fmt.Sprintf("multiple validation errors: %d errors found", len(ves))
}

// Add adds a validation error to the collection
func (ves *ValidationErrors) Add(field, message string) {
	*ves = append(*ves, NewValidationError(field, message))
}

// HasErrors returns true if there are any validation errors
func (ves ValidationErrors) HasErrors() bool {
	return len(ves) > 0
}

// GetFieldError returns the error for a specific field, if any
func (ves ValidationErrors) GetFieldError(field string) *ValidationError {
	for _, ve := range ves {
		if ve.Field == field {
			return &ve
		}
	}
	return nil
}

// ToMap converts validation errors to a map for easier JSON serialization
func (ves ValidationErrors) ToMap() map[string]string {
	result := make(map[string]string)
	for _, ve := range ves {
		result[ve.Field] = ve.Message
	}
	return result
}