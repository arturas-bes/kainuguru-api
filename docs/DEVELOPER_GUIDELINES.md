# Kainuguru API - Developer Guidelines

## Table of Contents
1. [Project Overview](#project-overview)
2. [Getting Started](#getting-started)
3. [Architecture & Design Principles](#architecture--design-principles)
4. [Development Workflow](#development-workflow)
5. [Code Organization & Standards](#code-organization--standards)
6. [Core Components Guide](#core-components-guide)
7. [Error Handling Standards](#error-handling-standards)
8. [Testing Standards](#testing-standards)
9. [GraphQL Development](#graphql-development)
10. [Database Guidelines](#database-guidelines)
11. [Authentication & Security](#authentication--security)
12. [Performance Considerations](#performance-considerations)
13. [Common Tasks & How-To](#common-tasks--how-to)
14. [Code Patterns Reference](#code-patterns-reference)
15. [Troubleshooting](#troubleshooting)
16. [API Documentation](#api-documentation)

---

## Project Overview

### What is Kainuguru API?

Kainuguru API is a sophisticated GraphQL-based backend service for Lithuanian grocery price comparison and smart shopping. It aggregates weekly flyers from major Lithuanian retailers (IKI, Maxima, Rimi, Lidl, Norfa), extracts product information, and provides intelligent features like price tracking, shopping list optimization, and deal discovery.

### Technology Stack

- **Language**: Go 1.24+
- **Web Framework**: Fiber v2 (high-performance HTTP)
- **API**: GraphQL (via gqlgen)
- **Database**: PostgreSQL 15+ with Bun ORM
- **Cache**: Redis 7+
- **Authentication**: JWT tokens
- **Container**: Docker & Docker Compose
- **Monitoring**: Prometheus + Grafana

### Key Features

1. **Multi-Store Support**: Scrapes and normalizes data from 5+ Lithuanian grocery chains
2. **Smart Search**: Lithuanian-optimized fuzzy search with trigram support
3. **Price Intelligence**: Historical tracking, trend analysis, price alerts
4. **Shopping Lists**: Cross-store optimization with fuzzy product matching
5. **Product Master Catalog**: Unified product database across all stores

---

## Getting Started

### Prerequisites

- Go 1.24+ installed
- Docker and Docker Compose
- PostgreSQL 15+ (via Docker)
- Redis 7+ (via Docker)
- Make utility

### Initial Setup

```bash
# Clone the repository
git clone [repository-url]
cd kainuguru-api

# Install dependencies and start services
make install

# Run database migrations
make db-migrate

# Seed test data (development only)
make seed-data

# Start the API server
make run
```

### Environment Configuration

1. Copy `.env.dist` to `.env`
2. Configure required variables:
```env
# Database
DATABASE_HOST=localhost
DATABASE_NAME=kainuguru
DATABASE_USER=kainuguru
DATABASE_PASSWORD=secret

# Redis
REDIS_HOST=localhost:6379

# JWT Secret (generate a secure random string)
JWT_SECRET=your-secure-secret-here

# API Keys
OPENAI_API_KEY=your-openai-key  # For product extraction
```

### Health Check

Verify installation:
```bash
# Check API health
curl http://localhost:8080/health

# Access GraphQL Playground
open http://localhost:8080/playground
```

---

## Architecture & Design Principles

### Clean Architecture Layers

```
┌─────────────────────────────────────────┐
│   Presentation Layer                     │
│   (HTTP/GraphQL Handlers)                │
├─────────────────────────────────────────┤
│   Application Layer                      │
│   (Services & Business Logic)            │
├─────────────────────────────────────────┤
│   Domain Layer                           │
│   (Models & Business Rules)              │
├─────────────────────────────────────────┤
│   Infrastructure Layer                   │
│   (Database, Cache, External APIs)       │
└─────────────────────────────────────────┘
```

### Core Design Patterns

1. **Factory Pattern**: Service and repository creation
2. **Dependency Injection**: Constructor-based DI throughout
3. **Repository Pattern**: Data access abstraction
4. **Service Layer Pattern**: Business logic encapsulation
5. **DataLoader Pattern**: N+1 query prevention in GraphQL

### Key Principles

1. **Type Safety**: Leverage Go's type system fully
2. **Error Handling**: Wrap errors with context, never swallow
3. **Immutability**: Prefer immutable data structures
4. **Testability**: Design for testing from the start
5. **Performance**: Optimize hot paths, use caching wisely

---

## Development Workflow

### Branch Strategy

```
main
 ├── feature/feature-name
 ├── bugfix/issue-description
 ├── hotfix/critical-fix
 └── release/version
```

### Commit Convention

Follow conventional commits:
```
type(scope): description

[optional body]
[optional footer]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

### Pull Request Process

1. Create feature branch from `main`
2. Implement changes following guidelines
3. Write/update tests
4. Update documentation
5. Run validation: `make validate-all`
6. Create PR with description template
7. Address review feedback
8. Squash and merge

---

## Code Organization & Standards

### Project Structure

```
kainuguru-api/
├── cmd/                    # Application entry points
│   └── api/               # Main API server
├── internal/              # Private application code
│   ├── config/           # Configuration management
│   ├── models/           # Domain models
│   ├── services/         # Business logic
│   ├── repositories/     # Data access
│   ├── handlers/         # HTTP/GraphQL handlers
│   ├── middleware/       # HTTP middleware
│   └── graphql/          # GraphQL schema & resolvers
├── pkg/                   # Reusable packages
├── migrations/            # Database migrations
├── tests/                # Test suites
└── docs/                 # Documentation
```

### Naming Conventions

**Files**:
- Use lowercase with underscores: `product_service.go`
- Test files: `product_service_test.go`
- Interfaces in separate files: `interfaces.go`

**Code**:
- Exported types/functions: PascalCase
- Private types/functions: camelCase
- Constants: UPPER_SNAKE_CASE
- Interfaces: End with "-er" suffix when possible

**Database**:
- Tables: plural, snake_case: `products`, `shopping_lists`
- Columns: snake_case: `created_at`, `product_master_id`
- Indexes: `idx_tablename_columns`
- Foreign keys: `fk_tablename_reference`

### Code Standards

```go
// Good: Clear, explicit error handling
product, err := s.productService.GetByID(ctx, id)
if err != nil {
    return nil, fmt.Errorf("failed to get product %d: %w", id, err)
}

// Bad: Swallowing errors
product, _ := s.productService.GetByID(ctx, id)
```

### Import Organization

```go
import (
    // Standard library
    "context"
    "fmt"

    // Third-party
    "github.com/uptrace/bun"

    // Internal packages
    "github.com/kainuguru/kainuguru-api/internal/models"
    "github.com/kainuguru/kainuguru-api/pkg/errors"
)
```

---

## Core Components Guide

### 1. Models (`internal/models/`)

Domain entities with business logic:

```go
// Example: Product Master model
type ProductMaster struct {
    bun.BaseModel `bun:"table:product_masters"`

    ID             int64     `bun:"id,pk,autoincrement"`
    Name           string    `bun:"name,notnull"`
    NormalizedName string    `bun:"normalized_name,notnull"`
    // ... other fields
}

// Business logic methods
func (pm *ProductMaster) IsActive() bool
func (pm *ProductMaster) CanBeMatched() bool
```

**Best Practices**:
- Keep models focused on domain logic
- Use value objects for complex types
- Implement validation methods
- Add helper methods for common operations

### 2. Services (`internal/services/`)

Business logic layer:

```go
// Service interface
type ProductService interface {
    GetByID(ctx context.Context, id int64) (*models.Product, error)
    Search(ctx context.Context, query string) ([]*models.Product, error)
}

// Implementation
type productService struct {
    db    *bun.DB
    cache *redis.Client
}
```

**Best Practices**:
- One service per domain concept
- Inject dependencies via constructor
- Return domain models, not DTOs
- Handle transactions at service level
- Implement caching where beneficial

### 3. GraphQL Resolvers (`internal/graphql/resolvers/`)

GraphQL implementation:

```go
// Resolver implementation
func (r *queryResolver) Product(ctx context.Context, id int) (*model.Product, error) {
    // Extract user from context
    user := middleware.GetUserFromContext(ctx)

    // Call service layer
    product, err := r.services.Product.GetByID(ctx, int64(id))
    if err != nil {
        return nil, gqlerror.Errorf("Product not found")
    }

    // Convert to GraphQL model if needed
    return toGraphQLProduct(product), nil
}
```

**Best Practices**:
- Keep resolvers thin, delegate to services
- Use DataLoaders for N+1 prevention
- Handle authorization in resolvers
- Convert errors to user-friendly messages
- Use field resolvers for expensive operations

### 4. Middleware (`internal/middleware/`)

Cross-cutting concerns:

```go
// Example: Auth middleware
func AuthMiddleware(config *config.Config) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := extractToken(c)
        if token == "" {
            return c.Next() // Continue without auth
        }

        // Validate token
        claims, err := validateJWT(token, config.JWTSecret)
        if err != nil {
            return fiber.ErrUnauthorized
        }

        // Add to context
        c.Locals("user", claims.UserID)
        return c.Next()
    }
}
```

**Middleware Stack** (in order):
1. Recovery - Panic recovery
2. CORS - Cross-origin handling
3. RateLimit - Request throttling
4. Logger - Request logging
5. Auth - Authentication

---

## Error Handling Standards

### Mandatory Error Package Usage

**ALL new code and refactored code MUST use `pkg/errors`:**

```go
import (
    apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
)
```

### Error Types and When to Use Them

| Error Type | HTTP Status | Usage |
|------------|-------------|-------|
| `ErrorTypeValidation` | 400 | Invalid input, missing required fields, format errors |
| `ErrorTypeAuthentication` | 401 | Missing or invalid credentials, expired tokens |
| `ErrorTypeAuthorization` | 403 | Insufficient permissions, access denied |
| `ErrorTypeNotFound` | 404 | Resource doesn't exist |
| `ErrorTypeConflict` | 409 | Duplicate resources, state conflicts |
| `ErrorTypeInternal` | 500 | Server errors, unexpected conditions |
| `ErrorTypeExternal` | 502 | Third-party service failures |
| `ErrorTypeRateLimit` | 429 | Rate limit exceeded |

### Layer-Specific Error Handling

#### Repository Layer
```go
// GOOD: Return raw database errors or wrap with context
func (r *productRepository) GetByID(ctx context.Context, id int64) (*models.Product, error) {
    product := new(models.Product)
    err := r.db.NewSelect().Model(product).Where("id = ?", id).Scan(ctx)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil // Let service layer decide error type
        }
        return nil, err // Pass through other DB errors
    }
    return product, nil
}

// BAD: Don't create AppErrors in repository layer
func (r *productRepository) GetByID(ctx context.Context, id int64) (*models.Product, error) {
    // ❌ DON'T DO THIS
    return nil, apperrors.NotFound("product not found")
}
```

#### Service Layer
```go
// GOOD: Map errors to appropriate types
func (s *productService) GetByID(ctx context.Context, id int64) (*models.Product, error) {
    product, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to fetch product")
    }
    if product == nil {
        return nil, apperrors.NotFoundF("product with ID %d not found", id)
    }
    return product, nil
}

// GOOD: Input validation with specific error messages
func (s *authService) Register(ctx context.Context, input *models.UserInput) (*AuthResult, error) {
    // Validation errors
    if input.Email == "" {
        return nil, apperrors.Validation("email is required")
    }
    if !isValidEmail(input.Email) {
        return nil, apperrors.ValidationF("invalid email format: %s", input.Email)
    }

    // Check for conflicts
    existing, _ := s.repo.GetByEmail(ctx, input.Email)
    if existing != nil {
        return nil, apperrors.ConflictF("user with email %s already exists", input.Email)
    }

    // Wrap internal errors
    hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to hash password")
    }

    // Continue...
}
```

#### GraphQL Resolver Layer
```go
// GOOD: Use error helper for consistent GraphQL errors
import "github.com/kainuguru/kainuguru-api/internal/graphql/errorhelper"

func (r *mutationResolver) CreateProduct(ctx context.Context, input model.ProductInput) (*models.Product, error) {
    userID, ok := middleware.GetUserFromContext(ctx)
    if !ok {
        return nil, errorhelper.Unauthenticated("authentication required")
    }

    product, err := r.productService.Create(ctx, input)
    if err != nil {
        return nil, errorhelper.HandleServiceError(err)
    }

    return product, nil
}

// BAD: Don't use fmt.Errorf in resolvers
func (r *mutationResolver) CreateProduct(ctx context.Context, input model.ProductInput) (*models.Product, error) {
    // ❌ DON'T DO THIS
    return nil, fmt.Errorf("failed to create product: %w", err)
}
```

### Error Wrapping Guidelines

```go
// GOOD: Add context at each layer
func processOrder(orderID int64) error {
    order, err := getOrder(orderID)
    if err != nil {
        return apperrors.Wrapf(err, apperrors.ErrorTypeInternal,
            "failed to process order %d", orderID)
    }
    // ...
}

// GOOD: Chain multiple contexts
func syncProducts() error {
    products, err := fetchProducts()
    if err != nil {
        return apperrors.Wrap(err, apperrors.ErrorTypeExternal,
            "failed to fetch products from external API")
    }

    for _, p := range products {
        if err := saveProduct(p); err != nil {
            return apperrors.Wrapf(err, apperrors.ErrorTypeInternal,
                "failed to save product %s", p.Name)
        }
    }
    return nil
}
```

### Error Logging Pattern

```go
// GOOD: Log with structured fields
func (s *service) ProcessItem(ctx context.Context, id int64) error {
    err := s.doProcessing(ctx, id)
    if err != nil {
        // Log internal errors with full context
        if apperrors.IsType(err, apperrors.ErrorTypeInternal) {
            s.logger.Error("processing failed",
                slog.Int64("item_id", id),
                slog.String("error", err.Error()),
                slog.Any("error_details", apperrors.GetDetails(err)),
            )
        }
        return err
    }

    // Log successful operations
    s.logger.Info("item processed successfully",
        slog.Int64("item_id", id),
    )
    return nil
}
```

### GraphQL Error Helper (CREATE THIS FILE)

```go
// internal/graphql/errorhelper/errors.go
package errorhelper

import (
    "github.com/99designs/gqlgen/graphql"
    "github.com/vektah/gqlparser/v2/gqlerror"
    apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
)

// HandleServiceError converts service errors to GraphQL errors with proper extensions
func HandleServiceError(err error) error {
    var appErr *apperrors.AppError
    if !apperrors.As(err, &appErr) {
        // Not an AppError, return generic internal error
        return &gqlerror.Error{
            Message: "An internal error occurred",
            Extensions: map[string]interface{}{
                "code": "INTERNAL_ERROR",
            },
        }
    }

    // Map to GraphQL error with proper code
    return &gqlerror.Error{
        Message: appErr.Message,
        Extensions: map[string]interface{}{
            "code": getGraphQLCode(appErr.Type),
            "details": appErr.Details,
        },
    }
}

func getGraphQLCode(errType apperrors.ErrorType) string {
    switch errType {
    case apperrors.ErrorTypeValidation:
        return "VALIDATION_ERROR"
    case apperrors.ErrorTypeAuthentication:
        return "UNAUTHENTICATED"
    case apperrors.ErrorTypeAuthorization:
        return "FORBIDDEN"
    case apperrors.ErrorTypeNotFound:
        return "NOT_FOUND"
    case apperrors.ErrorTypeConflict:
        return "CONFLICT"
    case apperrors.ErrorTypeRateLimit:
        return "RATE_LIMITED"
    default:
        return "INTERNAL_ERROR"
    }
}

// Convenience functions
func Unauthenticated(message string) error {
    return HandleServiceError(apperrors.Authentication(message))
}

func Forbidden(message string) error {
    return HandleServiceError(apperrors.Authorization(message))
}

func NotFound(message string) error {
    return HandleServiceError(apperrors.NotFound(message))
}

func ValidationError(message string) error {
    return HandleServiceError(apperrors.Validation(message))
}
```

### Migration Checklist for Existing Code

When updating existing services to use `pkg/errors`:

- [ ] Replace all `fmt.Errorf` with appropriate `apperrors` functions
- [ ] Replace all `errors.New` with `apperrors.New`
- [ ] Add error type classification (validation, not found, etc.)
- [ ] Ensure error wrapping preserves context
- [ ] Update tests to check error types, not just messages
- [ ] Add structured logging for internal errors

---

## Testing Standards

### Test Organization

**File Structure:**
```
service.go          # Implementation
service_test.go     # Tests in same directory
```

### Repository Testing Pattern

**MANDATORY: Use this setup pattern for all repository tests:**

```go
package repositories

import (
    "context"
    "database/sql"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
    "github.com/uptrace/bun"
    "github.com/uptrace/bun/dialect/sqlitedialect"
    "github.com/uptrace/bun/driver/sqliteshim"
)

// Standard test database setup
func setupTestDB(t *testing.T) (*bun.DB, func()) {
    t.Helper()

    // In-memory SQLite for fast, isolated tests
    sqldb, err := sql.Open(sqliteshim.DriverName(), "file:test?mode=memory&cache=shared")
    require.NoError(t, err)

    db := bun.NewDB(sqldb, sqlitedialect.New())

    // Create schema
    ctx := context.Background()
    _, err = db.ExecContext(ctx, createTableSQL)
    require.NoError(t, err)

    // Cleanup function
    cleanup := func() {
        _ = db.Close()
        _ = sqldb.Close()
    }

    return db, cleanup
}

// Test example
func TestRepository_Create(t *testing.T) {
    t.Parallel() // Enable parallel execution

    ctx := context.Background()
    db, cleanup := setupTestDB(t)
    defer cleanup()

    repo := NewRepository(db)

    // Test implementation
    entity := &models.Entity{
        Name: "Test Entity",
    }

    err := repo.Create(ctx, entity)
    require.NoError(t, err)
    require.NotZero(t, entity.ID)
}
```

### Service Testing Pattern

**MANDATORY: Use fake implementations with function fields:**

```go
// Fake implementation for testing
type fakeRepository struct {
    createFunc  func(ctx context.Context, entity *models.Entity) error
    getByIDFunc func(ctx context.Context, id int64) (*models.Entity, error)
    updateFunc  func(ctx context.Context, entity *models.Entity) error
    deleteFunc  func(ctx context.Context, id int64) error
}

func (f *fakeRepository) Create(ctx context.Context, entity *models.Entity) error {
    if f.createFunc != nil {
        return f.createFunc(ctx, entity)
    }
    return nil
}

func (f *fakeRepository) GetByID(ctx context.Context, id int64) (*models.Entity, error) {
    if f.getByIDFunc != nil {
        return f.getByIDFunc(ctx, id)
    }
    return nil, nil
}

// Service test
func TestService_ProcessEntity(t *testing.T) {
    t.Parallel()

    ctx := context.Background()

    // Setup fake with specific behavior
    repo := &fakeRepository{
        getByIDFunc: func(ctx context.Context, id int64) (*models.Entity, error) {
            if id == 1 {
                return &models.Entity{ID: 1, Name: "Test"}, nil
            }
            return nil, apperrors.NotFound("entity not found")
        },
    }

    // Create service with fake
    svc := &service{
        repo:   repo,
        logger: testLogger(),
    }

    // Test success case
    result, err := svc.ProcessEntity(ctx, 1)
    require.NoError(t, err)
    require.NotNil(t, result)

    // Test error case
    _, err = svc.ProcessEntity(ctx, 999)
    require.Error(t, err)
    require.True(t, apperrors.IsType(err, apperrors.ErrorTypeNotFound))
}

// Test logger that discards output
func testLogger() *slog.Logger {
    return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}
```

### Test Helper Functions

**MANDATORY: Always use t.Helper() for test utilities:**

```go
// Helper to create test data
func createTestUser(t *testing.T, db *bun.DB, email string) *models.User {
    t.Helper() // REQUIRED: Better stack traces

    user := &models.User{
        Email:     email,
        CreatedAt: time.Now(),
    }

    _, err := db.NewInsert().Model(user).Exec(context.Background())
    require.NoError(t, err)

    return user
}

// Helper to assert error types
func requireErrorType(t *testing.T, err error, expectedType apperrors.ErrorType) {
    t.Helper()

    require.Error(t, err)
    require.True(t, apperrors.IsType(err, expectedType),
        "expected error type %v, got %v", expectedType, err)
}
```

### Table-Driven Tests

**Prefer table-driven tests for multiple scenarios:**

```go
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name      string
        email     string
        wantError bool
        errorType apperrors.ErrorType
    }{
        {
            name:      "valid email",
            email:     "user@example.com",
            wantError: false,
        },
        {
            name:      "empty email",
            email:     "",
            wantError: true,
            errorType: apperrors.ErrorTypeValidation,
        },
        {
            name:      "invalid format",
            email:     "not-an-email",
            wantError: true,
            errorType: apperrors.ErrorTypeValidation,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateEmail(tt.email)

            if tt.wantError {
                require.Error(t, err)
                require.True(t, apperrors.IsType(err, tt.errorType))
            } else {
                require.NoError(t, err)
            }
        })
    }
}
```

### Integration Testing

**For tests requiring real database:**

```go
// +build integration

func TestIntegration_FullWorkflow(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test in short mode")
    }

    // Use real PostgreSQL connection
    db := setupPostgresDB(t)
    defer cleanupDB(t, db)

    // Test complete workflow
    ctx := context.Background()

    // Create services with real dependencies
    userRepo := repositories.NewUserRepository(db)
    userService := services.NewUserService(userRepo)

    // Test workflow...
}
```

### Test Coverage Requirements

- **Minimum coverage**: 70% for new code
- **Service layer**: 70% coverage expected
- **Repository layer**: Focus on complex queries
- **Utility packages**: 90% coverage expected

### Running Tests

```bash
# Unit tests only
go test ./...

# With coverage
go test -cover ./...

# Integration tests
go test -tags=integration ./...

# Specific package with verbose output
go test -v ./internal/services/...

# Run tests in parallel
go test -parallel 4 ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Naming Conventions

```go
// Function under test: GetUserByID
// Test name format: Test{Type}_{Function}_{Scenario}

func TestRepository_GetUserByID_Success(t *testing.T)
func TestRepository_GetUserByID_NotFound(t *testing.T)
func TestService_CreateUser_ValidationError(t *testing.T)
func TestService_CreateUser_DuplicateEmail(t *testing.T)
```

### Mocking External Services

```go
// Mock HTTP client for external APIs
type mockHTTPClient struct {
    doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
    if m.doFunc != nil {
        return m.doFunc(req)
    }
    return &http.Response{StatusCode: 200}, nil
}

// Usage in tests
func TestExternalAPI_FetchData(t *testing.T) {
    client := &mockHTTPClient{
        doFunc: func(req *http.Request) (*http.Response, error) {
            // Return mock response
            body := `{"data": "test"}`
            return &http.Response{
                StatusCode: 200,
                Body:       io.NopCloser(strings.NewReader(body)),
            }, nil
        },
    }

    api := NewExternalAPI(client)
    data, err := api.FetchData(context.Background())
    require.NoError(t, err)
    require.Equal(t, "test", data)
}
```

### Testing Checklist

When writing tests:

- [ ] Use `t.Parallel()` where possible
- [ ] Use `t.Helper()` in all helper functions
- [ ] Test both success and error cases
- [ ] Use table-driven tests for multiple scenarios
- [ ] Mock external dependencies
- [ ] Use `require` for fatal assertions
- [ ] Use `assert` for non-fatal checks
- [ ] Clean up resources with `defer`
- [ ] Test edge cases and boundaries
- [ ] Verify error types, not just error presence

---

## GraphQL Development

### Schema-First Approach

1. Define schema in `internal/graphql/schema/schema.graphql`
2. Generate code: `make generate-graphql`
3. Implement resolvers in `internal/graphql/resolvers/`

### Schema Design Guidelines

```graphql
# Good: Clear, nested structure
type Product {
  id: Int!
  name: String!
  price: ProductPrice!
  store: Store!
  priceHistory: [PriceHistory!]!
}

type ProductPrice {
  current: Float!
  original: Float
  discount: Float
  currency: String!
}
```

### DataLoader Usage

Prevent N+1 queries:

```go
// In resolver constructor
func NewResolver(services *services.ServiceFactory) *Resolver {
    return &Resolver{
        services: services,
        loaders:  dataloaders.NewLoaders(services),
    }
}

// In field resolver
func (r *productResolver) Store(ctx context.Context, obj *model.Product) (*model.Store, error) {
    // Use dataloader instead of direct service call
    return r.loaders.StoreLoader.Load(ctx, obj.StoreID)
}
```

### Error Handling

```go
// User-facing error
return nil, gqlerror.Errorf("Product not found with ID: %d", id)

// Internal error with context
if err != nil {
    log.Error().Err(err).Int("productId", id).Msg("Failed to fetch product")
    return nil, gqlerror.Errorf("An error occurred while fetching the product")
}
```

---

## Database Guidelines

### Migration Management

```bash
# Create new migration
make migration name=add_user_preferences

# Run migrations
make db-migrate

# Rollback last migration
make db-rollback
```

### Migration Best Practices

```sql
-- Good: Transactional, reversible
-- +goose Up
-- +goose StatementBegin
ALTER TABLE products ADD COLUMN IF NOT EXISTS tags TEXT[];
CREATE INDEX CONCURRENTLY idx_products_tags ON products USING gin(tags);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_products_tags;
ALTER TABLE products DROP COLUMN IF EXISTS tags;
-- +goose StatementEnd
```

### Query Optimization

```go
// Good: Selective loading with relations
products, err := s.db.NewSelect().
    Model(&products).
    Where("store_id = ?", storeID).
    Where("valid_from <= ? AND valid_to >= ?", now, now).
    Relation("Store").
    Relation("ProductMaster").
    Limit(100).
    Scan(ctx)

// Bad: Loading everything
products, err := s.db.NewSelect().Model(&products).Scan(ctx)
```

### Indexing Strategy

Key indexes to maintain:
- Foreign keys: automatic in PostgreSQL
- Search fields: trigram indexes for fuzzy search
- Date ranges: B-tree for validity periods
- JSONB fields: GIN indexes for metadata

---

## Authentication & Security

### JWT Token Flow

```
1. User login → Generate token pair
2. Access token (15 min) → API requests
3. Refresh token (7 days) → Token renewal
4. Session tracking → Redis storage
```

### Security Checklist

- [ ] Input validation on all endpoints
- [ ] SQL injection prevention (parameterized queries)
- [ ] XSS prevention (output encoding)
- [ ] CSRF protection (token validation)
- [ ] Rate limiting configured
- [ ] Sensitive data encryption
- [ ] Audit logging enabled
- [ ] CORS properly configured

### Authorization Pattern

```go
// In resolver
func (r *mutationResolver) UpdateProduct(ctx context.Context, input model.UpdateProductInput) (*model.Product, error) {
    // Extract user from context
    user := middleware.GetUserFromContext(ctx)
    if user == nil {
        return nil, gqlerror.Errorf("Authentication required")
    }

    // Check permissions
    if !user.HasRole("admin") {
        return nil, gqlerror.Errorf("Insufficient permissions")
    }

    // Proceed with operation
    return r.services.Product.Update(ctx, input)
}
```

---

## Testing Strategy

### Test Levels

1. **Unit Tests**: Individual functions/methods
2. **Integration Tests**: Service layer with real DB
3. **E2E Tests**: Full GraphQL queries
4. **BDD Tests**: Cucumber scenarios

### Running Tests

```bash
# All tests
make test

# Unit tests only
go test ./internal/...

# Integration tests
make test-integration

# BDD tests
make test-bdd

# Validation suite
make validate-all
```

### Writing Tests

```go
// Unit test example
func TestProductMaster_CanBeMatched(t *testing.T) {
    tests := []struct {
        name     string
        product  *models.ProductMaster
        expected bool
    }{
        {
            name: "active product with good confidence",
            product: &models.ProductMaster{
                Status:          "active",
                ConfidenceScore: 0.8,
            },
            expected: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := tt.product.CanBeMatched()
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### BDD Scenarios

```gherkin
Feature: Product Search
  Scenario: Search products by name
    Given the following products exist:
      | name     | store |
      | Pienas   | IKI   |
      | Duona    | Maxima|
    When I search for "pienas"
    Then I should see 1 product
    And the product should be from "IKI"
```

---

## Performance Considerations

### Caching Strategy

1. **Redis Cache Layers**:
   - Session data: 24 hours
   - Product searches: 5 minutes
   - Store data: 1 hour
   - Flyer pages: Until expiry

2. **Implementation**:
```go
// Cache wrapper
func (s *productService) GetByID(ctx context.Context, id int64) (*models.Product, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("product:%d", id)
    cached, err := s.cache.Get(ctx, cacheKey).Result()
    if err == nil {
        return unmarshalProduct(cached), nil
    }

    // Load from database
    product, err := s.loadFromDB(ctx, id)
    if err != nil {
        return nil, err
    }

    // Cache for next time
    s.cache.Set(ctx, cacheKey, marshalProduct(product), 5*time.Minute)
    return product, nil
}
```

### Database Optimization

1. **Connection Pooling**:
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(5 * time.Minute)
db.SetConnMaxIdleTime(2 * time.Minute)
```

2. **Query Optimization**:
- Use EXPLAIN ANALYZE for slow queries
- Add appropriate indexes
- Use partitioning for time-series data
- Batch operations where possible

3. **N+1 Prevention**:
- Always use DataLoaders in GraphQL
- Eager load relations when needed
- Use JOIN queries appropriately

### Monitoring

Key metrics to track:
- Request latency (p50, p95, p99)
- Database query time
- Cache hit ratio
- Error rates
- Memory usage
- Goroutine count

---

## Common Tasks & How-To

### Add a New GraphQL Query

1. Update schema:
```graphql
# internal/graphql/schema/schema.graphql
extend type Query {
  productsByCategory(category: String!): [Product!]!
}
```

2. Generate code:
```bash
make generate-graphql
```

3. Implement resolver:
```go
// internal/graphql/resolvers/query.go
func (r *queryResolver) ProductsByCategory(ctx context.Context, category string) ([]*model.Product, error) {
    return r.services.Product.GetByCategory(ctx, category)
}
```

### Add a New Service

1. Define interface:
```go
// internal/services/interfaces.go
type YourService interface {
    YourMethod(ctx context.Context) error
}
```

2. Implement service:
```go
// internal/services/your_service.go
type yourService struct {
    db *bun.DB
}

func NewYourService(db *bun.DB) YourService {
    return &yourService{db: db}
}
```

3. Register in factory:
```go
// internal/services/factory.go
factory.YourService = NewYourService(db)
```

### Add Database Migration

1. Create migration file:
```bash
make migration name=add_new_table
```

2. Write migration:
```sql
-- +goose Up
CREATE TABLE new_table (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS new_table;
```

3. Run migration:
```bash
make db-migrate
```

### Scrape New Store

1. Implement scraper:
```go
// internal/services/scraper/newstore_scraper.go
type NewStoreScraper struct {
    baseURL string
}

func (s *NewStoreScraper) ScrapeFlyers() ([]*models.Flyer, error) {
    // Implementation
}
```

2. Register in factory:
```go
// internal/services/scraper/factory.go
case "newstore":
    return &NewStoreScraper{baseURL: config.URL}
```

---

## Code Patterns Reference

### Service Pattern with Dependency Injection

**MANDATORY: All services must follow this pattern:**

```go
// Interface definition
type ProductService interface {
    GetByID(ctx context.Context, id int64) (*models.Product, error)
    Create(ctx context.Context, input *models.ProductInput) (*models.Product, error)
    Update(ctx context.Context, id int64, input *models.ProductInput) (*models.Product, error)
    Delete(ctx context.Context, id int64) error
}

// Implementation
type productService struct {
    db     *bun.DB
    repo   repositories.ProductRepository
    cache  *redis.Client
    logger *slog.Logger
}

// Primary constructor (for production)
func NewProductService(db *bun.DB, cache *redis.Client) ProductService {
    return &productService{
        db:     db,
        repo:   repositories.NewProductRepository(db),
        cache:  cache,
        logger: logger.ServiceLogger("product"),
    }
}

// Test constructor (for testing with mocks)
func NewProductServiceWithDeps(
    repo repositories.ProductRepository,
    cache *redis.Client,
    logger *slog.Logger,
) ProductService {
    if repo == nil {
        panic("repository is required")
    }
    return &productService{
        repo:   repo,
        cache:  cache,
        logger: logger,
    }
}
```

### Generic Repository Pattern

**Use the base repository for standard CRUD operations:**

```go
// Base repository with generics
package base

type Repository[T any] struct {
    db    *bun.DB
    model T
}

func NewRepository[T any](db *bun.DB) *Repository[T] {
    return &Repository[T]{db: db}
}

// Functional options for queries
type QueryOption[T any] func(*bun.SelectQuery) *bun.SelectQuery

func WithWhere[T any](condition string, args ...interface{}) QueryOption[T] {
    return func(q *bun.SelectQuery) *bun.SelectQuery {
        return q.Where(condition, args...)
    }
}

func WithLimit[T any](limit int) QueryOption[T] {
    return func(q *bun.SelectQuery) *bun.SelectQuery {
        return q.Limit(limit)
    }
}

// Generic CRUD methods
func (r *Repository[T]) GetByID(ctx context.Context, id int64, opts ...QueryOption[T]) (*T, error) {
    var item T
    q := r.db.NewSelect().Model(&item).Where("id = ?", id)

    for _, opt := range opts {
        q = opt(q)
    }

    err := q.Scan(ctx)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
    return &item, nil
}

// Usage in specific repository
type productRepository struct {
    base *base.Repository[models.Product]
}

func (r *productRepository) GetActiveProducts(ctx context.Context) ([]*models.Product, error) {
    return r.base.GetAll(ctx,
        base.WithWhere[models.Product]("is_active = ?", true),
        base.WithOrderBy[models.Product]("created_at DESC"),
        base.WithLimit[models.Product](100),
    )
}
```

### Transaction Pattern

**Use RunInTx for atomic operations:**

```go
func (s *orderService) CreateOrder(ctx context.Context, input *OrderInput) (*models.Order, error) {
    var order *models.Order

    err := s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
        // 1. Create order
        order = &models.Order{
            UserID: input.UserID,
            Status: "pending",
        }
        if err := tx.NewInsert().Model(order).Scan(ctx); err != nil {
            return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to create order")
        }

        // 2. Create order items
        for _, item := range input.Items {
            orderItem := &models.OrderItem{
                OrderID:   order.ID,
                ProductID: item.ProductID,
                Quantity:  item.Quantity,
            }
            if _, err := tx.NewInsert().Model(orderItem).Exec(ctx); err != nil {
                return apperrors.Wrap(err, apperrors.ErrorTypeInternal, "failed to create order item")
            }
        }

        // 3. Update inventory
        for _, item := range input.Items {
            _, err := tx.NewUpdate().
                Model((*models.Product)(nil)).
                Set("stock = stock - ?", item.Quantity).
                Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
                Exec(ctx)
            if err != nil {
                return apperrors.Validation("insufficient stock")
            }
        }

        return nil
    })

    if err != nil {
        return nil, err
    }
    return order, nil
}
```

### DataLoader Pattern for GraphQL

**Prevent N+1 queries in GraphQL resolvers:**

```go
// Loader implementation
func batchUserLoader(service UserService) dataloader.BatchFunc[string, *models.User] {
    return func(ctx context.Context, keys []string) []*dataloader.Result[*models.User] {
        // Batch fetch
        users, err := service.GetByIDs(ctx, keys)
        if err != nil {
            // Return error for all keys
            results := make([]*dataloader.Result[*models.User], len(keys))
            for i := range results {
                results[i] = &dataloader.Result[*models.User]{Error: err}
            }
            return results
        }

        // Map results to original key order
        userMap := make(map[string]*models.User)
        for _, user := range users {
            userMap[user.ID] = user
        }

        results := make([]*dataloader.Result[*models.User], len(keys))
        for i, key := range keys {
            if user, ok := userMap[key]; ok {
                results[i] = &dataloader.Result[*models.User]{Data: user}
            } else {
                results[i] = &dataloader.Result[*models.User]{
                    Error: apperrors.NotFoundF("user %s not found", key),
                }
            }
        }
        return results
    }
}

// Usage in resolver
func (r *orderResolver) User(ctx context.Context, obj *models.Order) (*models.User, error) {
    loaders := dataloaders.FromContext(ctx)
    return loaders.UserLoader.Load(ctx, obj.UserID)()
}
```

### Context Propagation Pattern

**Always pass context through the call stack:**

```go
// GOOD: Context flows through all layers
func (r *resolver) GetProduct(ctx context.Context, id int) (*models.Product, error) {
    // Extract user from context
    userID, _ := middleware.GetUserFromContext(ctx)

    // Pass context to service
    product, err := r.productService.GetByID(ctx, int64(id))
    if err != nil {
        return nil, err
    }

    // Log with context
    logger.FromContext(ctx).Info("product fetched",
        slog.Int("product_id", id),
        slog.String("user_id", userID),
    )

    return product, nil
}

// Service layer
func (s *productService) GetByID(ctx context.Context, id int64) (*models.Product, error) {
    // Check cache with context
    if cached, err := s.cache.Get(ctx, fmt.Sprintf("product:%d", id)).Result(); err == nil {
        return unmarshalProduct(cached), nil
    }

    // Database query with context
    return s.repo.GetByID(ctx, id)
}

// Repository layer
func (r *productRepository) GetByID(ctx context.Context, id int64) (*models.Product, error) {
    var product models.Product
    err := r.db.NewSelect().
        Model(&product).
        Where("id = ?", id).
        Scan(ctx) // Context passed to database
    return &product, err
}
```

### Logging Pattern

**Use structured logging with domain context:**

```go
// Service initialization
func NewProductService(db *bun.DB) ProductService {
    return &productService{
        db:     db,
        logger: logger.ServiceLogger("product"),
    }
}

// Logging in methods
func (s *productService) CreateProduct(ctx context.Context, input *ProductInput) (*models.Product, error) {
    // Start operation log
    s.logger.Debug("creating product",
        slog.String("name", input.Name),
        slog.String("category", input.Category),
    )

    product, err := s.repo.Create(ctx, input)
    if err != nil {
        // Error logging
        s.logger.Error("failed to create product",
            slog.String("error", err.Error()),
            slog.String("name", input.Name),
        )
        return nil, err
    }

    // Success logging
    s.logger.Info("product created",
        slog.Int64("product_id", product.ID),
        slog.String("name", product.Name),
    )

    // Business event
    logger.LogProductCreated(product.ID, product.Name, input.Category)

    return product, nil
}
```

### Validation Pattern

**Model-level validation with clear errors:**

```go
// Model validation
func (p *Product) Validate() error {
    if p.Name == "" {
        return apperrors.Validation("product name is required")
    }
    if len(p.Name) < 3 {
        return apperrors.ValidationF("product name too short: %d characters", len(p.Name))
    }
    if p.Price < 0 {
        return apperrors.Validation("price cannot be negative")
    }
    if p.Category == "" {
        return apperrors.Validation("category is required")
    }
    return nil
}

// Service layer usage
func (s *productService) Create(ctx context.Context, input *ProductInput) (*models.Product, error) {
    product := &models.Product{
        Name:     input.Name,
        Price:    input.Price,
        Category: input.Category,
    }

    // Validate before saving
    if err := product.Validate(); err != nil {
        return nil, err
    }

    return s.repo.Create(ctx, product)
}
```

### Migration Patterns

**Database migration best practices:**

```sql
-- +goose Up
-- +goose StatementBegin
-- 1. Create table with constraints
CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    normalized_name VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    price DECIMAL(10,2) CHECK (price >= 0),
    stock INTEGER DEFAULT 0 CHECK (stock >= 0),
    is_active BOOLEAN DEFAULT true,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Create indexes for common queries
CREATE INDEX idx_products_name_trgm ON products USING gin(name gin_trgm_ops);
CREATE INDEX idx_products_normalized_name ON products(normalized_name);
CREATE INDEX idx_products_category ON products(category) WHERE category IS NOT NULL;
CREATE INDEX idx_products_active ON products(is_active) WHERE is_active = true;

-- 3. Add trigger for updated_at
CREATE TRIGGER products_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 4. Add comments for documentation
COMMENT ON TABLE products IS 'Product catalog with pricing and stock information';
COMMENT ON COLUMN products.normalized_name IS 'Lowercase, accent-removed version for searching';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products CASCADE;
-- +goose StatementEnd
```

### Common Anti-Patterns to Avoid

```go
// ❌ BAD: Direct database access in handlers
func Handler(c *fiber.Ctx) error {
    db := c.Locals("db").(*bun.DB)
    var product models.Product
    err := db.NewSelect().Model(&product).Scan(c.Context())
    // DON'T DO THIS
}

// ❌ BAD: Ignoring errors
result, _ := service.GetProduct(ctx, id)
// DON'T DO THIS

// ❌ BAD: Using panic for error handling
if err != nil {
    panic(err) // DON'T DO THIS
}

// ❌ BAD: Global variables
var globalDB *bun.DB // DON'T DO THIS

// ❌ BAD: Magic numbers/strings
if user.Role == 1 { // DON'T DO THIS
    // Use constants instead
}

// ❌ BAD: Mixing concerns
func (s *productService) RenderHTML() string {
    // Services shouldn't handle presentation
}

// ❌ BAD: Not using context
func GetProduct(id int64) (*Product, error) {
    // Always pass context
}

// ❌ BAD: SQL injection vulnerable
query := fmt.Sprintf("SELECT * FROM products WHERE name = '%s'", userInput)
// DON'T DO THIS - use parameterized queries
```

---

## Troubleshooting

### Common Issues

#### Database Connection Errors
```bash
# Check PostgreSQL is running
docker-compose ps db

# Check connection string
echo $DATABASE_URL

# Test connection
psql -h localhost -U kainuguru -d kainuguru
```

#### Redis Connection Issues
```bash
# Check Redis is running
docker-compose ps redis

# Test connection
redis-cli ping
```

#### Migration Failures
```bash
# Check migration status
./bin/migrator status

# Manually rollback if needed
./bin/migrator down
```

#### GraphQL Generation Errors
```bash
# Clean generated files
rm -rf internal/graphql/generated

# Regenerate
make generate-graphql
```

### Debug Mode

Enable debug logging:
```env
LOG_LEVEL=debug
APP_DEBUG=true
```

### Performance Issues

1. Enable query logging:
```go
db.AddQueryHook(bundebug.NewQueryHook(
    bundebug.WithVerbose(true),
))
```

2. Check slow queries:
```sql
SELECT query, calls, mean_exec_time
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;
```

3. Monitor Redis:
```bash
redis-cli monitor
```

---

## API Documentation

### GraphQL Playground

Access at: `http://localhost:8080/playground`

### Example Queries

**Search Products**:
```graphql
query SearchProducts($query: String!) {
  searchProducts(query: $query) {
    exact {
      id
      name
      price {
        current
        discount
      }
      store {
        name
      }
    }
    fuzzy {
      id
      name
      similarityScore
    }
  }
}
```

**Get Shopping List**:
```graphql
query GetShoppingList($id: ID!) {
  shoppingList(id: $id) {
    id
    name
    items {
      id
      productName
      quantity
      matchedProduct {
        name
        price {
          current
        }
        store {
          name
        }
      }
    }
    totalEstimate
  }
}
```

**Create Shopping List**:
```graphql
mutation CreateShoppingList($input: CreateShoppingListInput!) {
  createShoppingList(input: $input) {
    id
    name
    createdAt
  }
}
```

### Error Codes

| Code | Description |
|------|-------------|
| UNAUTHENTICATED | Missing or invalid authentication |
| UNAUTHORIZED | Insufficient permissions |
| NOT_FOUND | Resource not found |
| VALIDATION_ERROR | Input validation failed |
| INTERNAL_ERROR | Server error occurred |

---

## Best Practices Summary

### DO:
- ✅ Write tests for new features
- ✅ Handle errors explicitly
- ✅ Use context for cancellation
- ✅ Add appropriate logging
- ✅ Document complex logic
- ✅ Use transactions for data consistency
- ✅ Validate input data
- ✅ Follow naming conventions
- ✅ Use DataLoaders in GraphQL
- ✅ Cache expensive operations

### DON'T:
- ❌ Ignore errors
- ❌ Use global variables
- ❌ Hardcode configuration
- ❌ Skip tests
- ❌ Mix business logic in handlers
- ❌ Use raw SQL without parameters
- ❌ Expose internal errors to users
- ❌ Create N+1 queries
- ❌ Store sensitive data in logs
- ❌ Skip code reviews

---

## Contributing

1. Read this guide thoroughly
2. Follow the development workflow
3. Write comprehensive tests
4. Update documentation
5. Submit PR with clear description
6. Address review feedback promptly

For questions or issues, please refer to the project's issue tracker or contact the maintainers.

---

## Appendix

### Useful Commands

```bash
# Development
make run              # Start API server
make watch           # Auto-reload on changes
make logs            # View application logs

# Database
make db-console      # PostgreSQL console
make db-reset        # Reset database
make db-backup       # Create backup
make db-restore      # Restore from backup

# Testing
make test            # Run all tests
make test-coverage   # Generate coverage report
make validate-all    # Complete validation

# Docker
make docker-build    # Build Docker image
make docker-up       # Start all services
make docker-down     # Stop all services
make docker-logs     # View container logs

# Utilities
make fmt             # Format code
make lint            # Run linter
make generate        # Generate all code
make clean           # Clean build artifacts
```

### Environment Variables Reference

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| DATABASE_HOST | PostgreSQL host | Yes | localhost |
| DATABASE_PORT | PostgreSQL port | No | 5432 |
| DATABASE_NAME | Database name | Yes | kainuguru |
| DATABASE_USER | Database user | Yes | kainuguru |
| DATABASE_PASSWORD | Database password | Yes | - |
| REDIS_HOST | Redis host:port | Yes | localhost:6379 |
| JWT_SECRET | JWT signing secret | Yes | - |
| OPENAI_API_KEY | OpenAI API key | No | - |
| LOG_LEVEL | Logging level | No | info |
| APP_ENV | Environment | No | development |
| SERVER_PORT | API port | No | 8080 |

### Project Contacts

- **Technical Lead**: [Contact Information]
- **DevOps**: [Contact Information]
- **Product Owner**: [Contact Information]

---

*Last Updated: December 2024*
*Version: 2.0.0*

## Changelog

### Version 2.0.0 (December 2024)
- Added comprehensive Error Handling Standards section with `pkg/errors` usage
- Added detailed Testing Standards with mandatory patterns
- Added Code Patterns Reference section with real examples from codebase
- Enhanced GraphQL development section with DataLoader patterns
- Added migration patterns and best practices
- Documented anti-patterns to avoid
- Updated based on deep codebase analysis of actual implementation patterns