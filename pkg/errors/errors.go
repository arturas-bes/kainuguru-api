package errors

import (
	"fmt"
	"net/http"
)

// ErrorType represents the type of error
type ErrorType string

const (
	ErrorTypeValidation     ErrorType = "validation"
	ErrorTypeAuthentication ErrorType = "authentication"
	ErrorTypeAuthorization  ErrorType = "authorization"
	ErrorTypeNotFound       ErrorType = "not_found"
	ErrorTypeConflict       ErrorType = "conflict"
	ErrorTypeInternal       ErrorType = "internal"
	ErrorTypeExternal       ErrorType = "external"
	ErrorTypeRateLimit      ErrorType = "rate_limit"
)

// AppError represents a structured application error
type AppError struct {
	Type       ErrorType `json:"type"`
	Message    string    `json:"message"`
	Code       string    `json:"code,omitempty"`
	Details    string    `json:"details,omitempty"`
	StatusCode int       `json:"-"`
	Cause      error     `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates a new AppError
func New(errorType ErrorType, message string) *AppError {
	return &AppError{
		Type:       errorType,
		Message:    message,
		StatusCode: getDefaultStatusCode(errorType),
	}
}

// Newf creates a new AppError with formatted message
func Newf(errorType ErrorType, format string, args ...interface{}) *AppError {
	return &AppError{
		Type:       errorType,
		Message:    fmt.Sprintf(format, args...),
		StatusCode: getDefaultStatusCode(errorType),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errorType ErrorType, message string) *AppError {
	return &AppError{
		Type:       errorType,
		Message:    message,
		StatusCode: getDefaultStatusCode(errorType),
		Cause:      err,
	}
}

// Wrapf wraps an existing error with formatted message
func Wrapf(err error, errorType ErrorType, format string, args ...interface{}) *AppError {
	return &AppError{
		Type:       errorType,
		Message:    fmt.Sprintf(format, args...),
		StatusCode: getDefaultStatusCode(errorType),
		Cause:      err,
	}
}

// WithCode adds a specific error code
func (e *AppError) WithCode(code string) *AppError {
	e.Code = code
	return e
}

// WithDetails adds additional details
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithStatusCode overrides the default status code
func (e *AppError) WithStatusCode(code int) *AppError {
	e.StatusCode = code
	return e
}

// getDefaultStatusCode returns the default HTTP status code for an error type
func getDefaultStatusCode(errorType ErrorType) int {
	switch errorType {
	case ErrorTypeValidation:
		return http.StatusBadRequest
	case ErrorTypeAuthentication:
		return http.StatusUnauthorized
	case ErrorTypeAuthorization:
		return http.StatusForbidden
	case ErrorTypeNotFound:
		return http.StatusNotFound
	case ErrorTypeConflict:
		return http.StatusConflict
	case ErrorTypeRateLimit:
		return http.StatusTooManyRequests
	case ErrorTypeExternal:
		return http.StatusBadGateway
	case ErrorTypeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// Common error constructors
func Validation(message string) *AppError {
	return New(ErrorTypeValidation, message)
}

func ValidationF(format string, args ...interface{}) *AppError {
	return Newf(ErrorTypeValidation, format, args...)
}

func Authentication(message string) *AppError {
	return New(ErrorTypeAuthentication, message)
}

func Authorization(message string) *AppError {
	return New(ErrorTypeAuthorization, message)
}

func NotFound(message string) *AppError {
	return New(ErrorTypeNotFound, message)
}

func Conflict(message string) *AppError {
	return New(ErrorTypeConflict, message)
}

func Internal(message string) *AppError {
	return New(ErrorTypeInternal, message)
}

func InternalF(format string, args ...interface{}) *AppError {
	return Newf(ErrorTypeInternal, format, args...)
}

func External(message string) *AppError {
	return New(ErrorTypeExternal, message)
}

func RateLimit(message string) *AppError {
	return New(ErrorTypeRateLimit, message)
}

// IsType checks if an error is of a specific type
func IsType(err error, errorType ErrorType) bool {
	var appErr *AppError
	if !As(err, &appErr) {
		return false
	}
	return appErr.Type == errorType
}

// As is a wrapper around errors.As for AppError
func As(err error, target **AppError) bool {
	for err != nil {
		if appErr, ok := err.(*AppError); ok {
			*target = appErr
			return true
		}
		// Handle wrapped errors
		if unwrappable, ok := err.(interface{ Unwrap() error }); ok {
			err = unwrappable.Unwrap()
		} else {
			break
		}
	}
	return false
}

// GetStatusCode extracts HTTP status code from any error
func GetStatusCode(err error) int {
	var appErr *AppError
	if As(err, &appErr) {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}
