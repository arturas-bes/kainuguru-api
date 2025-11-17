package errors

import (
	"fmt"

	"github.com/vektah/gqlparser/v2/gqlerror"
)

// GraphQL error codes for wizard operations
const (
	CodeValidationError    = "VALIDATION_ERROR"
	CodeNotFound           = "NOT_FOUND"
	CodeSessionExpired     = "SESSION_EXPIRED"
	CodeStaleData          = "STALE_DATA"
	CodeListLocked         = "LIST_LOCKED"
	CodeRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	CodeNoExpiredItems     = "NO_EXPIRED_ITEMS"
	CodeInvalidDecision    = "INVALID_DECISION"
	CodeRevalidationFailed = "REVALIDATION_FAILED"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeInternalError      = "INTERNAL_ERROR"
)

// WizardError is the base interface for all wizard-specific errors
type WizardError interface {
	error
	Code() string
	ToGraphQLError() *gqlerror.Error
}

// ValidationError represents input validation failures
type ValidationError struct {
	Message string
	Field   string
	Details []string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

func (e *ValidationError) Code() string {
	return CodeValidationError
}

func (e *ValidationError) ToGraphQLError() *gqlerror.Error {
	return &gqlerror.Error{
		Message: e.Error(),
		Extensions: map[string]interface{}{
			"code":    e.Code(),
			"field":   e.Field,
			"details": e.Details,
		},
	}
}

// NewValidationError creates a new validation error
func NewValidationError(message, field string, details ...string) *ValidationError {
	return &ValidationError{
		Message: message,
		Field:   field,
		Details: details,
	}
}

// NotFoundError represents resource not found errors
type NotFoundError struct {
	Message    string
	ResourceID string
}

func (e *NotFoundError) Error() string {
	if e.ResourceID != "" {
		return fmt.Sprintf("resource not found: %s (id: %s)", e.Message, e.ResourceID)
	}
	return fmt.Sprintf("resource not found: %s", e.Message)
}

func (e *NotFoundError) Code() string {
	return CodeNotFound
}

func (e *NotFoundError) ToGraphQLError() *gqlerror.Error {
	return &gqlerror.Error{
		Message: e.Error(),
		Extensions: map[string]interface{}{
			"code":        e.Code(),
			"resource_id": e.ResourceID,
		},
	}
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message, resourceID string) *NotFoundError {
	return &NotFoundError{
		Message:    message,
		ResourceID: resourceID,
	}
}

// ListLockedError represents shopping list locked by another wizard session
type ListLockedError struct {
	Message  string
	ListID   int64
	LockedBy string
}

func (e *ListLockedError) Error() string {
	return fmt.Sprintf("shopping list is locked: %s (list_id: %d)", e.Message, e.ListID)
}

func (e *ListLockedError) Code() string {
	return CodeListLocked
}

func (e *ListLockedError) ToGraphQLError() *gqlerror.Error {
	return &gqlerror.Error{
		Message: e.Error(),
		Extensions: map[string]interface{}{
			"code":      e.Code(),
			"list_id":   e.ListID,
			"locked_by": e.LockedBy,
		},
	}
}

// NewListLockedError creates a new list locked error
func NewListLockedError(listID int64, lockedBy string) *ListLockedError {
	return &ListLockedError{
		Message:  "shopping list is already being migrated by another active wizard session",
		ListID:   listID,
		LockedBy: lockedBy,
	}
}

// RateLimitExceededError represents rate limit exceeded errors
type RateLimitExceededError struct {
	Message       string
	Limit         int
	WindowSeconds int
	RetryAfter    int
}

func (e *RateLimitExceededError) Error() string {
	return fmt.Sprintf("rate limit exceeded: %s (limit: %d per %d seconds, retry after: %d seconds)",
		e.Message, e.Limit, e.WindowSeconds, e.RetryAfter)
}

func (e *RateLimitExceededError) Code() string {
	return CodeRateLimitExceeded
}

func (e *RateLimitExceededError) ToGraphQLError() *gqlerror.Error {
	return &gqlerror.Error{
		Message: e.Error(),
		Extensions: map[string]interface{}{
			"code":           e.Code(),
			"limit":          e.Limit,
			"window_seconds": e.WindowSeconds,
			"retry_after":    e.RetryAfter,
		},
	}
}

// NewRateLimitExceededError creates a new rate limit exceeded error
func NewRateLimitExceededError(limit, windowSeconds, retryAfter int) *RateLimitExceededError {
	return &RateLimitExceededError{
		Message:       fmt.Sprintf("maximum %d wizard sessions per hour", limit),
		Limit:         limit,
		WindowSeconds: windowSeconds,
		RetryAfter:    retryAfter,
	}
}

// ToGraphQLError converts any error to a GraphQL error
// If the error implements WizardError, uses its custom mapping
// Otherwise, wraps it as a generic error
func ToGraphQLError(err error) *gqlerror.Error {
	if err == nil {
		return nil
	}

	// Check if error already implements WizardError
	if wizErr, ok := err.(WizardError); ok {
		return wizErr.ToGraphQLError()
	}

	// Generic error wrapping
	return &gqlerror.Error{
		Message: err.Error(),
		Extensions: map[string]interface{}{
			"code": CodeInternalError,
		},
	}
}
