package errors_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
)

// Example demonstrates basic error creation
func Example_basicUsage() {
	// Creating errors
	err1 := apperrors.NotFound("product not found")
	fmt.Println(err1.Error())

	err2 := apperrors.ValidationF("invalid ID: %d", -1)
	fmt.Println(err2.Error())

	// Output:
	// not_found: product not found
	// validation: invalid ID: -1
}

// Example demonstrates error wrapping
func Example_wrapping() {
	dbErr := sql.ErrNoRows
	wrapped := apperrors.Wrap(dbErr, apperrors.ErrorTypeNotFound, "product not found")
	fmt.Println(wrapped.Error())
	fmt.Println(errors.Is(wrapped, sql.ErrNoRows))

	// Output:
	// not_found: product not found (caused by: sql: no rows in result set)
	// true
}

// Example demonstrates checking error types
func Example_typeChecking() {
	err := apperrors.NotFound("resource not found")

	if apperrors.IsType(err, apperrors.ErrorTypeNotFound) {
		fmt.Println("Error is NotFound type")
	}

	statusCode := apperrors.GetStatusCode(err)
	fmt.Printf("HTTP Status: %d\n", statusCode)

	// Output:
	// Error is NotFound type
	// HTTP Status: 404
}

// Example demonstrates service layer pattern
func Example_serviceLayer() {
	type Product struct {
		ID   int64
		Name string
	}

	type productRepository interface {
		GetByID(ctx context.Context, id int64) (*Product, error)
	}

	type productService struct {
		repo productRepository
	}

	// Service method showing proper error handling
	getProduct := func(s *productService, ctx context.Context, id int64) (*Product, error) {
		// Validation
		if id <= 0 {
			return nil, apperrors.Validation("product ID must be positive")
		}

		// Repository call
		product, err := s.repo.GetByID(ctx, id)
		if err != nil {
			// Map known errors
			if errors.Is(err, sql.ErrNoRows) {
				return nil, apperrors.NotFound("product not found")
			}
			// Wrap unexpected errors
			return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get product")
		}

		return product, nil
	}

	_ = getProduct // Avoid unused warning
}

// Example demonstrates migration pattern
func Example_migration() {
	// BEFORE: Generic error wrapping
	oldPattern := func() error {
		dbErr := sql.ErrNoRows
		return fmt.Errorf("failed to get product: %w", dbErr)
	}

	// AFTER: Typed error with context
	newPattern := func() error {
		dbErr := sql.ErrNoRows
		if errors.Is(dbErr, sql.ErrNoRows) {
			return apperrors.NotFound("product not found")
		}
		return apperrors.Wrap(dbErr, apperrors.ErrorTypeInternal, "failed to get product")
	}

	fmt.Println("Old:", oldPattern().Error())
	fmt.Println("New:", newPattern().Error())

	// Output:
	// Old: failed to get product: sql: no rows in result set
	// New: not_found: product not found
}

// Example demonstrates adding error details
func Example_errorDetails() {
	err := apperrors.Validation("invalid input").
		WithCode("VAL001").
		WithDetails("field 'email' is required and must be valid")

	fmt.Printf("Type: %s\n", err.Type)
	fmt.Printf("Code: %s\n", err.Code)
	fmt.Printf("Message: %s\n", err.Message)

	// Output:
	// Type: validation
	// Code: VAL001
	// Message: invalid input
}

// Example demonstrates common database error mapping
func Example_databaseErrors() {
	mapDBError := func(err error, operation string) error {
		if err == nil {
			return nil
		}

		// No rows found
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NotFound("resource not found")
		}

		// Connection errors
		if errors.Is(err, sql.ErrConnDone) {
			return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "database connection closed")
		}

		// Constraint violations (would need to check error string/code in real code)
		// This is a simplified example
		errStr := err.Error()
		if len(errStr) > 0 && errStr[0] == 'c' { // Simulated constraint check
			return apperrors.Wrap(err, apperrors.ErrorTypeConflict, "constraint violation")
		}

		// Default: wrap as internal error
		return apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "database operation failed: %s", operation)
	}

	err1 := mapDBError(sql.ErrNoRows, "GetByID")
	fmt.Println(err1.Error())

	err2 := mapDBError(sql.ErrConnDone, "Create")
	fmt.Println(errors.Is(err2, sql.ErrConnDone))

	// Output:
	// not_found: resource not found
	// true
}
