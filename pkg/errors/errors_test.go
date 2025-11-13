package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		errorType  ErrorType
		message    string
		wantStatus int
	}{
		{
			name:       "validation error",
			errorType:  ErrorTypeValidation,
			message:    "invalid input",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found error",
			errorType:  ErrorTypeNotFound,
			message:    "resource not found",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "internal error",
			errorType:  ErrorTypeInternal,
			message:    "something went wrong",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.errorType, tt.message)

			if err.Type != tt.errorType {
				t.Errorf("New() type = %v, want %v", err.Type, tt.errorType)
			}
			if err.Message != tt.message {
				t.Errorf("New() message = %v, want %v", err.Message, tt.message)
			}
			if err.StatusCode != tt.wantStatus {
				t.Errorf("New() status = %v, want %v", err.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestNewf(t *testing.T) {
	err := Newf(ErrorTypeValidation, "invalid field: %s", "email")

	want := "invalid field: email"
	if err.Message != want {
		t.Errorf("Newf() message = %v, want %v", err.Message, want)
	}
}

func TestWrap(t *testing.T) {
	original := errors.New("database connection failed")
	wrapped := Wrap(original, ErrorTypeInternal, "failed to query database")

	if wrapped.Cause != original {
		t.Error("Wrap() did not preserve original error")
	}

	if !errors.Is(wrapped, original) {
		t.Error("Wrap() error chain broken")
	}
}

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *AppError
		want string
	}{
		{
			name: "error without cause",
			err:  New(ErrorTypeValidation, "invalid input"),
			want: "validation: invalid input",
		},
		{
			name: "error with cause",
			err:  Wrap(errors.New("db error"), ErrorTypeInternal, "query failed"),
			want: "internal: query failed (caused by: db error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap(original, ErrorTypeInternal, "wrapped")

	if wrapped.Unwrap() != original {
		t.Error("Unwrap() did not return original error")
	}
}

func TestAppError_WithCode(t *testing.T) {
	err := New(ErrorTypeValidation, "test").WithCode("VAL001")

	if err.Code != "VAL001" {
		t.Errorf("WithCode() = %v, want VAL001", err.Code)
	}
}

func TestAppError_WithDetails(t *testing.T) {
	err := New(ErrorTypeValidation, "test").WithDetails("field email is required")

	if err.Details != "field email is required" {
		t.Errorf("WithDetails() = %v, want 'field email is required'", err.Details)
	}
}

func TestAppError_WithStatusCode(t *testing.T) {
	err := New(ErrorTypeValidation, "test").WithStatusCode(http.StatusTeapot)

	if err.StatusCode != http.StatusTeapot {
		t.Errorf("WithStatusCode() = %v, want %v", err.StatusCode, http.StatusTeapot)
	}
}

func TestCommonConstructors(t *testing.T) {
	tests := []struct {
		name       string
		constructor func() *AppError
		wantType    ErrorType
		wantStatus  int
	}{
		{
			name:        "Validation",
			constructor: func() *AppError { return Validation("test") },
			wantType:    ErrorTypeValidation,
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "Authentication",
			constructor: func() *AppError { return Authentication("test") },
			wantType:    ErrorTypeAuthentication,
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name:        "Authorization",
			constructor: func() *AppError { return Authorization("test") },
			wantType:    ErrorTypeAuthorization,
			wantStatus:  http.StatusForbidden,
		},
		{
			name:        "NotFound",
			constructor: func() *AppError { return NotFound("test") },
			wantType:    ErrorTypeNotFound,
			wantStatus:  http.StatusNotFound,
		},
		{
			name:        "Conflict",
			constructor: func() *AppError { return Conflict("test") },
			wantType:    ErrorTypeConflict,
			wantStatus:  http.StatusConflict,
		},
		{
			name:        "Internal",
			constructor: func() *AppError { return Internal("test") },
			wantType:    ErrorTypeInternal,
			wantStatus:  http.StatusInternalServerError,
		},
		{
			name:        "External",
			constructor: func() *AppError { return External("test") },
			wantType:    ErrorTypeExternal,
			wantStatus:  http.StatusBadGateway,
		},
		{
			name:        "RateLimit",
			constructor: func() *AppError { return RateLimit("test") },
			wantType:    ErrorTypeRateLimit,
			wantStatus:  http.StatusTooManyRequests,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.constructor()

			if err.Type != tt.wantType {
				t.Errorf("constructor type = %v, want %v", err.Type, tt.wantType)
			}
			if err.StatusCode != tt.wantStatus {
				t.Errorf("constructor status = %v, want %v", err.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestIsType(t *testing.T) {
	err := Validation("test")

	if !IsType(err, ErrorTypeValidation) {
		t.Error("IsType() should return true for matching type")
	}

	if IsType(err, ErrorTypeInternal) {
		t.Error("IsType() should return false for non-matching type")
	}

	// Test with regular error
	regularErr := errors.New("regular error")
	if IsType(regularErr, ErrorTypeValidation) {
		t.Error("IsType() should return false for non-AppError")
	}
}

func TestAs(t *testing.T) {
	appErr := Validation("test")

	var target *AppError
	if !As(appErr, &target) {
		t.Error("As() should return true for AppError")
	}
	if target.Type != ErrorTypeValidation {
		t.Errorf("As() target type = %v, want %v", target.Type, ErrorTypeValidation)
	}

	// Test with regular error
	regularErr := errors.New("regular error")
	var target2 *AppError
	if As(regularErr, &target2) {
		t.Error("As() should return false for non-AppError")
	}
}

func TestGetStatusCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "AppError with validation",
			err:  Validation("test"),
			want: http.StatusBadRequest,
		},
		{
			name: "AppError with custom status",
			err:  Internal("test").WithStatusCode(http.StatusServiceUnavailable),
			want: http.StatusServiceUnavailable,
		},
		{
			name: "regular error defaults to 500",
			err:  errors.New("regular error"),
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStatusCode(tt.err)
			if got != tt.want {
				t.Errorf("GetStatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDefaultStatusCode(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		want      int
	}{
		{ErrorTypeValidation, http.StatusBadRequest},
		{ErrorTypeAuthentication, http.StatusUnauthorized},
		{ErrorTypeAuthorization, http.StatusForbidden},
		{ErrorTypeNotFound, http.StatusNotFound},
		{ErrorTypeConflict, http.StatusConflict},
		{ErrorTypeRateLimit, http.StatusTooManyRequests},
		{ErrorTypeExternal, http.StatusBadGateway},
		{ErrorTypeInternal, http.StatusInternalServerError},
		{ErrorType("unknown"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(string(tt.errorType), func(t *testing.T) {
			got := getDefaultStatusCode(tt.errorType)
			if got != tt.want {
				t.Errorf("getDefaultStatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWrappedErrorChain(t *testing.T) {
	// Create error chain: original -> internal wrap -> validation wrap
	original := errors.New("database connection failed")
	internal := Wrap(original, ErrorTypeInternal, "query failed")
	validation := Wrap(internal, ErrorTypeValidation, "invalid data")

	// Should be able to unwrap to original
	if !errors.Is(validation, original) {
		t.Error("error chain should preserve Is() relationship")
	}

	// Should extract as AppError
	var target *AppError
	if !As(validation, &target) {
		t.Error("should be able to extract AppError from chain")
	}

	// Should get outermost type
	if target.Type != ErrorTypeValidation {
		t.Errorf("As() should extract outermost error, got type %v", target.Type)
	}
}
