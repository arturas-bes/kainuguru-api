# Error Handling Package

## Overview

This package provides structured error handling for the Kainuguru API with domain-specific error types, HTTP status code mapping, and error wrapping capabilities.

## Design Principles

1. **Type Safety**: Use typed errors for different error categories
2. **Error Wrapping**: Preserve error chains with `%w` semantics
3. **HTTP Mapping**: Automatic HTTP status code assignment
4. **Context Preservation**: Maintain error context through the call stack
5. **GraphQL Compatible**: Works seamlessly with GraphQL error handling

## Error Types

| Type | HTTP Status | Use Case |
|------|-------------|----------|
| `ErrorTypeValidation` | 400 Bad Request | Invalid input, validation failures |
| `ErrorTypeAuthentication` | 401 Unauthorized | Missing or invalid credentials |
| `ErrorTypeAuthorization` | 403 Forbidden | Insufficient permissions |
| `ErrorTypeNotFound` | 404 Not Found | Resource doesn't exist |
| `ErrorTypeConflict` | 409 Conflict | Duplicate resources, constraint violations |
| `ErrorTypeRateLimit` | 429 Too Many Requests | Rate limiting |
| `ErrorTypeExternal` | 502 Bad Gateway | External service failures |
| `ErrorTypeInternal` | 500 Internal Server Error | Unexpected errors |

## Usage Patterns

### Creating New Errors

```go
import apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"

// Simple error
err := apperrors.NotFound("product not found")

// Formatted error
err := apperrors.ValidationF("invalid email format: %s", email)

// With additional context
err := apperrors.Validation("invalid input").
    WithCode("VAL001").
    WithDetails("field 'email' is required")
```

### Wrapping Errors

```go
// Wrap database errors
product, err := s.repo.GetByID(ctx, id)
if err != nil {
    return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get product")
}

// Wrap with formatted message
err := s.externalAPI.Fetch(url)
if err != nil {
    return apperrors.Wrapf(err, apperrors.ErrorTypeExternal, "failed to fetch from %s", url)
}
```

### Service Layer Pattern

```go
func (s *productService) GetByID(ctx context.Context, id int64) (*models.Product, error) {
    // Validate input
    if id <= 0 {
        return nil, apperrors.Validation("product ID must be positive")
    }

    // Call repository
    product, err := s.repo.GetByID(ctx, id)
    if err != nil {
        // Check for specific errors
        if errors.Is(err, sql.ErrNoRows) {
            return nil, apperrors.NotFound("product not found")
        }
        // Wrap unexpected errors
        return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get product")
    }

    return product, nil
}
```

### GraphQL Resolver Pattern

```go
func (r *queryResolver) Product(ctx context.Context, id int64) (*models.Product, error) {
    product, err := r.services.Product.GetByID(ctx, id)
    if err != nil {
        // Error already wrapped by service layer
        // GraphQL will extract message and status code
        return nil, err
    }
    return product, nil
}
```

### Checking Error Types

```go
err := doSomething()

// Check if error is specific type
if apperrors.IsType(err, apperrors.ErrorTypeNotFound) {
    // Handle not found
}

// Extract AppError details
var appErr *apperrors.AppError
if apperrors.As(err, &appErr) {
    log.Error("error", "type", appErr.Type, "code", appErr.Code)
}

// Get HTTP status code
statusCode := apperrors.GetStatusCode(err)
```

## Migration Strategy

### Phase 1: Service Layer (Priority Order)

1. **Critical Services** (Week 1)
   - `product_master_service.go`
   - `shopping_list_item_service.go`
   - `auth/service.go`

2. **High-Volume Services** (Week 2)
   - `product_service.go`
   - `flyer_service.go`
   - `store_service.go`

3. **Remaining Services** (Week 3)
   - All other service files

### Migration Rules

1. **Preserve Error Messages**: Keep exact wording for existing errors
2. **Add Type Information**: Map errors to appropriate types
3. **Maintain Tests**: Ensure all tests pass after migration
4. **One Service at a Time**: Migrate incrementally

### Before Migration

```go
func (s *service) GetByID(ctx context.Context, id int64) (*Model, error) {
    model, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("failed to get model: %w", err)
    }
    return model, nil
}
```

### After Migration

```go
import apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"

func (s *service) GetByID(ctx context.Context, id int64) (*Model, error) {
    model, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, apperrors.NotFound("model not found")
        }
        return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to get model")
    }
    return model, nil
}
```

## Error Handling Guidelines

### DO

✅ Use typed errors for domain logic
✅ Wrap errors to preserve context
✅ Map database errors to appropriate types
✅ Include operation context in messages
✅ Use error codes for API consumers
✅ Log detailed errors, return user-friendly messages

### DON'T

❌ Lose error chains (always use `Wrap` not `New`)
❌ Return generic "internal error" for known issues
❌ Expose internal details in error messages
❌ Swallow errors without logging
❌ Create new error types unnecessarily
❌ Change error behavior during refactoring

## Testing

```go
func TestService_GetByID(t *testing.T) {
    t.Run("returns not found error", func(t *testing.T) {
        // Setup
        repo := &mockRepo{}
        repo.On("GetByID", mock.Anything, int64(999)).
            Return(nil, sql.ErrNoRows)

        svc := NewService(repo)

        // Execute
        _, err := svc.GetByID(context.Background(), 999)

        // Assert error type
        if !apperrors.IsType(err, apperrors.ErrorTypeNotFound) {
            t.Errorf("expected NotFound error, got %v", err)
        }

        // Assert status code
        if apperrors.GetStatusCode(err) != http.StatusNotFound {
            t.Error("expected 404 status code")
        }
    })
}
```

## HTTP Integration

```go
// Fiber middleware
func ErrorHandler(c *fiber.Ctx, err error) error {
    statusCode := apperrors.GetStatusCode(err)

    var appErr *apperrors.AppError
    if apperrors.As(err, &appErr) {
        return c.Status(statusCode).JSON(fiber.Map{
            "error": appErr.Message,
            "code":  appErr.Code,
            "type":  appErr.Type,
        })
    }

    return c.Status(statusCode).JSON(fiber.Map{
        "error": "internal server error",
    })
}
```

## GraphQL Integration

The error package integrates seamlessly with GraphQL. GraphQL will automatically:
- Extract error messages
- Map to appropriate HTTP status codes
- Preserve error extensions

```go
// GraphQL error formatter (if needed)
func FormatError(err error) *gqlerror.Error {
    var appErr *apperrors.AppError
    if apperrors.As(err, &appErr) {
        return &gqlerror.Error{
            Message: appErr.Message,
            Extensions: map[string]interface{}{
                "code": appErr.Code,
                "type": appErr.Type,
            },
        }
    }
    return &gqlerror.Error{Message: err.Error()}
}
```

## Best Practices

1. **Repository Layer**: Return standard Go errors (sql.ErrNoRows, etc.)
2. **Service Layer**: Wrap with AppError, add business context
3. **Handler/Resolver Layer**: Pass through, don't re-wrap
4. **Middleware**: Extract status codes, format responses
5. **Logging**: Log full error chain with context
6. **Testing**: Assert error types, not string messages

## Common Patterns

### Database Not Found

```go
if errors.Is(err, sql.ErrNoRows) {
    return nil, apperrors.NotFound("resource not found")
}
```

### Validation

```go
if len(email) == 0 {
    return apperrors.Validation("email is required")
}

if !isValidEmail(email) {
    return apperrors.ValidationF("invalid email format: %s", email)
}
```

### External API Failure

```go
resp, err := http.Get(url)
if err != nil {
    return apperrors.Wrap(err, apperrors.ErrorTypeExternal, "failed to fetch from API")
}
```

### Authorization

```go
if user.Role != "admin" {
    return apperrors.Authorization("admin role required")
}
```

## Performance Considerations

- Error creation is cheap (no stack traces by default)
- Error wrapping preserves original error (no copying)
- Type checking is O(1) for AppError
- Status code lookup is O(1) switch statement

## Future Enhancements

Potential additions (out of scope for initial migration):

- Stack trace capture (opt-in)
- Structured error fields
- Error metrics/monitoring integration
- i18n error messages
- Error recovery strategies
