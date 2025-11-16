# Code Patterns Analysis - Kainuguru API

**Generated:** 2025-11-14
**Purpose:** Comprehensive analysis of actual code patterns used throughout the Golang API project

---

## Table of Contents

1. [Internal Directory Patterns](#1-internal-directory-patterns)
2. [Pkg Directory Patterns](#2-pkg-directory-patterns)
3. [Testing Patterns](#3-testing-patterns)
4. [Code Quality Patterns](#4-code-quality-patterns)
5. [Database Patterns](#5-database-patterns)
6. [GraphQL Patterns](#6-graphql-patterns)
7. [Pattern Consistency Analysis](#7-pattern-consistency-analysis)
8. [Recommended Patterns](#8-recommended-patterns)

---

## 1. Internal Directory Patterns

### 1.1 Service Implementation Patterns

#### ✅ Good Pattern: Interface-Based Services

**Location:** `/internal/services/interfaces.go`

```go
// Service interfaces are defined centrally
type FlyerService interface {
    GetByID(ctx context.Context, id int) (*models.Flyer, error)
    GetByIDs(ctx context.Context, ids []int) ([]*models.Flyer, error)
    // ... more methods
}

// Type aliases for filter structures from domain packages
type FlyerFilters = flyer.Filters
```

**Benefits:**
- Clear contract definition
- Easy to mock for testing
- Type aliases prevent import cycles
- Centralized interface documentation

#### ✅ Good Pattern: Service Constructor with Repository Injection

**Location:** `/internal/services/flyer_service.go`

```go
type flyerService struct {
    repo flyer.Repository  // Interface, not concrete type
}

// Primary constructor
func NewFlyerService(db *bun.DB) FlyerService {
    return NewFlyerServiceWithRepository(newFlyerRepository(db))
}

// Test-friendly constructor
func NewFlyerServiceWithRepository(repo flyer.Repository) FlyerService {
    if repo == nil {
        panic("flyer repository cannot be nil")
    }
    return &flyerService{repo: repo}
}
```

**Benefits:**
- Dependency injection ready
- Testability via repository mocking
- Nil-check for safety
- Clear separation of concerns

#### ✅ Good Pattern: Service Method Delegation with Error Wrapping

**Location:** `/internal/services/flyer_service.go`

```go
func (fs *flyerService) GetByID(ctx context.Context, id int) (*models.Flyer, error) {
    flyer, err := fs.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, apperrors.NotFound(fmt.Sprintf("flyer not found with ID %d", id))
        }
        return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get flyer by ID %d", id)
    }
    return flyer, nil
}
```

**Pattern:**
1. Call repository method
2. Check for specific errors (e.g., `sql.ErrNoRows`)
3. Wrap errors with context using custom error package
4. Return typed errors (NotFound, Internal, etc.)

#### ✅ Good Pattern: Complex Service with Multiple Dependencies

**Location:** `/internal/services/auth/service.go`

```go
type authServiceImpl struct {
    db              *bun.DB
    config          *AuthConfig
    passwordService PasswordService
    jwtService      JWTService
    sessionService  SessionService
    emailService    EmailService
}

func NewAuthServiceImpl(
    db *bun.DB,
    config *AuthConfig,
    passwordService PasswordService,
    jwtService JWTService,
    emailService EmailService,
) AuthService {
    // Create sub-services as needed
    sessionService := NewSessionService(db, config)

    return &authServiceImpl{
        db:              db,
        config:          config,
        passwordService: passwordService,
        jwtService:      jwtService,
        sessionService:  sessionService,
        emailService:    emailService,
    }
}
```

**Benefits:**
- Clear dependency declaration
- Composed of smaller services
- Easy to test individual components

#### ⚠️ Pattern to Improve: Direct DB Access in Services

**Issue:** Some services use direct DB access alongside repository pattern

```go
func (a *authServiceImpl) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
    user := &models.User{}
    err := a.db.NewSelect().Model(user).Where("email = ?", email).Scan(ctx)
    // Direct DB query instead of repository delegation
}
```

**Recommendation:** Extract all DB queries to repositories for consistency.

---

### 1.2 Repository Patterns

#### ✅ Excellent Pattern: Generic Base Repository

**Location:** `/internal/repositories/base/repository.go`

```go
type Repository[T any] struct {
    db       *bun.DB
    pkColumn string
}

func NewRepository[T any](db *bun.DB, pkColumn string) *Repository[T] {
    if db == nil {
        panic("base repository requires a non-nil *bun.DB")
    }
    if pkColumn == "" {
        panic("base repository requires a primary key column name")
    }
    return &Repository[T]{db: db, pkColumn: pkColumn}
}

func (r *Repository[T]) GetByID(ctx context.Context, id interface{}, opts ...SelectOption[T]) (*T, error) {
    model := new(T)
    query := r.db.NewSelect().
        Model(model).
        Where(fmt.Sprintf("%s = ?", r.pkColumn), id)

    query = applySelectOptions(query, opts)

    if err := query.Scan(ctx); err != nil {
        return nil, err
    }
    return model, nil
}
```

**Benefits:**
- DRY principle (no CRUD repetition)
- Type-safe with generics
- Flexible via functional options pattern
- Panic-on-misconfiguration (fail fast)

#### ✅ Excellent Pattern: Functional Options for Query Customization

```go
type SelectOption[T any] func(*bun.SelectQuery) *bun.SelectQuery

func WithQuery[T any](fn func(*bun.SelectQuery) *bun.SelectQuery) SelectOption[T] {
    return func(q *bun.SelectQuery) *bun.SelectQuery {
        if fn == nil {
            return q
        }
        return fn(q)
    }
}

func WithLimit[T any](limit int) SelectOption[T] {
    return func(q *bun.SelectQuery) *bun.SelectQuery {
        if limit > 0 {
            q.Limit(limit)
        }
        return q
    }
}

// Usage:
products, err := repo.GetByIDs(ctx, ids,
    base.WithQuery[models.Product](func(q *bun.SelectQuery) *bun.SelectQuery {
        return q.Relation("Store").Relation("Flyer")
    }),
)
```

**Benefits:**
- Highly composable
- Type-safe
- Readable API
- No option explosion

#### ✅ Good Pattern: Specific Repository with Base Composition

**Location:** `/internal/repositories/product_repository.go`

```go
type productRepository struct {
    db   *bun.DB
    base *base.Repository[models.Product]
}

func NewProductRepository(db *bun.DB) product.Repository {
    return &productRepository{
        db:   db,
        base: base.NewRepository[models.Product](db, "p.id"),
    }
}

func (r *productRepository) GetByID(ctx context.Context, id int) (*models.Product, error) {
    return r.base.GetByID(ctx, id, base.WithQuery[models.Product](func(q *bun.SelectQuery) *bun.SelectQuery {
        return attachProductRelations(q)
    }))
}

func attachProductRelations(q *bun.SelectQuery) *bun.SelectQuery {
    return q.Relation("Store").Relation("Flyer").Relation("FlyerPage")
}
```

**Pattern:**
1. Compose base repository for common operations
2. Keep db reference for complex queries
3. Helper functions for common query modifications
4. Domain-specific methods use base + customization

#### ✅ Good Pattern: Filter Application

```go
func applyProductFilters(q *bun.SelectQuery, filters *product.Filters) *bun.SelectQuery {
    if filters == nil {
        return q
    }
    if len(filters.StoreIDs) > 0 {
        q.Where("p.store_id IN (?)", bun.In(filters.StoreIDs))
    }
    if len(filters.FlyerIDs) > 0 {
        q.Where("p.flyer_id IN (?)", bun.In(filters.FlyerIDs))
    }
    if filters.IsOnSale != nil {
        q.Where("p.is_on_sale = ?", *filters.IsOnSale)
    }
    if filters.MinPrice != nil {
        q.Where("p.current_price >= ?", *filters.MinPrice)
    }
    return q
}
```

**Benefits:**
- Nil-safe (checks filters != nil)
- Pointer fields allow distinguishing "not set" from "false/zero"
- Progressive query building
- Reusable across methods

#### ✅ Good Pattern: Empty Slice Short-Circuit

```go
func (r *productRepository) GetProductsByFlyerIDs(ctx context.Context, flyerIDs []int) ([]*models.Product, error) {
    if len(flyerIDs) == 0 {
        return []*models.Product{}, nil  // Return empty slice, not nil
    }
    // ... rest of query
}
```

**Benefits:**
- Prevents SQL errors on empty IN clauses
- Consistent return type (empty slice vs nil)
- Performance optimization

---

### 1.3 Model Patterns

#### ✅ Good Pattern: Bun Model with Annotations

**Location:** `/internal/models/user.go`

```go
type User struct {
    bun.BaseModel `bun:"table:users,alias:u"`

    ID                uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
    Email             string    `bun:"email,unique,notnull" json:"email"`
    EmailVerified     bool      `bun:"email_verified,default:false" json:"emailVerified"`
    PasswordHash      string    `bun:"password_hash,notnull" json:"-"`  // Never expose
    FullName          *string   `bun:"full_name" json:"fullName"`

    // JSONB metadata field
    MetadataJSON json.RawMessage `bun:"metadata,type:jsonb,default:'{}'" json:"-"`
    Metadata     UserMetadata    `bun:"-" json:"metadata"`

    // Timestamps
    CreatedAt   time.Time  `bun:"created_at,default:now()" json:"createdAt"`
    UpdatedAt   time.Time  `bun:"updated_at,default:now()" json:"updatedAt"`
    LastLoginAt *time.Time `bun:"last_login_at" json:"lastLoginAt"`

    // Relations
    Sessions []*UserSession `bun:"rel:has-many,join:id=user_id" json:"sessions,omitempty"`
}
```

**Patterns:**
- Table alias for query optimization
- JSON tag `-` to exclude sensitive fields
- Pointer fields for nullable columns
- Embedded `bun.BaseModel`
- Separate JSON and DB fields for JSONB
- Relations with join specifications

#### ✅ Excellent Pattern: Model Hooks for JSONB

```go
func (u *User) BeforeAppendModel(query bun.Query) error {
    switch query.(type) {
    case *bun.InsertQuery:
        u.CreatedAt = time.Now()
        u.UpdatedAt = time.Now()
    case *bun.UpdateQuery:
        u.UpdatedAt = time.Now()
    }

    // Marshal metadata if it has non-zero values
    if u.hasNonZeroMetadata() {
        data, err := json.Marshal(u.Metadata)
        if err != nil {
            return err
        }
        u.MetadataJSON = data
    }

    return nil
}

func (u *User) AfterSelectModel() error {
    // Unmarshal metadata
    if len(u.MetadataJSON) > 0 {
        if err := json.Unmarshal(u.MetadataJSON, &u.Metadata); err != nil {
            u.Metadata = DefaultUserMetadata()  // Fallback to defaults
        }
    } else {
        u.Metadata = DefaultUserMetadata()
    }
    return nil
}
```

**Benefits:**
- Automatic timestamp management
- JSONB marshaling/unmarshaling
- Type-safe metadata access
- Graceful degradation on unmarshal errors

#### ✅ Good Pattern: Business Logic Methods on Models

**Location:** `/internal/models/shopping_list.go`

```go
func (sl *ShoppingList) GetCompletionPercentage() float64 {
    if sl.ItemCount == 0 {
        return 0.0
    }
    return float64(sl.CompletedItemCount) / float64(sl.ItemCount) * 100.0
}

func (sl *ShoppingList) IsCompleted() bool {
    return sl.ItemCount > 0 && sl.CompletedItemCount == sl.ItemCount
}

func (sl *ShoppingList) Archive() {
    sl.IsArchived = true
    sl.UpdatedAt = time.Now()
}

func (sl *ShoppingList) Validate() error {
    if len(sl.Name) == 0 {
        return NewValidationError("name", "List name is required")
    }
    if len(sl.Name) > 100 {
        return NewValidationError("name", "List name must be 100 characters or less")
    }
    return nil
}
```

**Benefits:**
- Domain logic lives with domain model
- Self-documenting behavior
- Reusable across services
- Validation at model level

#### ⚠️ Inconsistency: Validation Location

Some models have validation in:
- Model methods (`shopping_list.go`)
- Service layer
- GraphQL resolvers

**Recommendation:** Standardize on model-level validation with service-level orchestration.

---

### 1.4 Handler/Controller Patterns

#### ✅ Good Pattern: Middleware Composition

**Location:** `/internal/middleware/auth.go`

```go
type AuthMiddlewareConfig struct {
    Required       bool
    JWTService     auth.JWTService
    SessionService auth.SessionService
}

func NewAuthMiddleware(cfg AuthMiddlewareConfig) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := extractToken(c)
        if token == "" {
            if cfg.Required {
                return unauthorizedResponse(c, "Missing or invalid authorization header", "")
            }
            return c.Next()  // Optional auth, continue without user
        }

        claims, err := cfg.JWTService.ValidateAccessToken(token)
        if err != nil {
            if cfg.Required {
                return unauthorizedResponse(c, "Invalid or expired token", err.Error())
            }
            return c.Next()
        }

        // Store auth data in context
        ctx := context.WithValue(c.Context(), UserContextKey, claims.UserID)
        ctx = context.WithValue(ctx, SessionContextKey, session.ID)
        ctx = context.WithValue(ctx, ClaimsContextKey, claims)
        c.SetUserContext(ctx)

        return c.Next()
    }
}

// Convenience constructors
func AuthMiddleware(jwtService auth.JWTService, sessionService auth.SessionService) fiber.Handler {
    return NewAuthMiddleware(AuthMiddlewareConfig{
        Required:       true,
        JWTService:     jwtService,
        SessionService: sessionService,
    })
}

func OptionalAuthMiddleware(jwtService auth.JWTService, sessionService auth.SessionService) fiber.Handler {
    return NewAuthMiddleware(AuthMiddlewareConfig{
        Required:       false,
        JWTService:     jwtService,
        SessionService: sessionService,
    })
}
```

**Benefits:**
- Config-driven behavior
- Required vs optional authentication
- Context propagation
- Helper extractors

#### ✅ Good Pattern: Context Helpers

```go
func GetUserFromContext(ctx context.Context) (uuid.UUID, bool) {
    userID, ok := ctx.Value(UserContextKey).(uuid.UUID)
    return userID, ok
}

func GetSessionFromContext(ctx context.Context) (uuid.UUID, bool) {
    sessionID, ok := ctx.Value(SessionContextKey).(uuid.UUID)
    return sessionID, ok
}

func GetClaimsFromContext(ctx context.Context) (*auth.TokenClaims, bool) {
    claims, ok := ctx.Value(ClaimsContextKey).(*auth.TokenClaims)
    return claims, ok
}
```

**Benefits:**
- Type-safe context extraction
- Consistent API across codebase
- Boolean return for presence check

---

### 1.5 Middleware Implementation Patterns

#### ✅ Good Pattern: Chainable Middleware with Different Purposes

```go
// Token extraction
func extractToken(c *fiber.Ctx) string {
    authHeader := c.Get("Authorization")
    if authHeader == "" {
        return ""
    }
    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
        return ""
    }
    return parts[1]
}

// Session security checks (runs after auth)
func SessionSecurityMiddleware(sessionService auth.SessionService) fiber.Handler {
    return func(c *fiber.Ctx) error {
        sessionID, ok := GetSessionFromContext(c.Context())
        if !ok {
            return c.Next() // Skip if not authenticated
        }

        session, err := sessionService.GetSession(c.Context(), sessionID)
        if err != nil {
            return unauthorizedResponse(c, "Invalid session")
        }

        // Security checks
        currentIP := c.IP()
        if session.IPAddress != nil && session.IPAddress.String() != currentIP {
            fmt.Printf("⚠️  IP address change detected for session %s\n", sessionID)
        }

        return c.Next()
    }
}
```

**Pattern:**
- Small, focused middleware
- Composable in chains
- Graceful degradation (log warnings, don't block)

---

## 2. Pkg Directory Patterns

### 2.1 Error Handling Patterns

#### ✅ Excellent Pattern: Structured Application Errors

**Location:** `/pkg/errors/errors.go`

```go
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
```

**Benefits:**
- Implements `error` interface
- Implements `Unwrap()` for error chains
- HTTP status code mapping
- Type classification for error handling
- JSON serializable (minus internal fields)

#### ✅ Excellent Pattern: Error Constructors

```go
// Generic constructors
func New(errorType ErrorType, message string) *AppError { ... }
func Newf(errorType ErrorType, format string, args ...interface{}) *AppError { ... }
func Wrap(err error, errorType ErrorType, message string) *AppError { ... }
func Wrapf(err error, errorType ErrorType, format string, args ...interface{}) *AppError { ... }

// Specific type constructors
func Validation(message string) *AppError {
    return New(ErrorTypeValidation, message)
}

func Authentication(message string) *AppError {
    return New(ErrorTypeAuthentication, message)
}

func NotFound(message string) *AppError {
    return New(ErrorTypeNotFound, message)
}
```

**Benefits:**
- Consistent error creation
- Type-specific shortcuts
- Format string support
- Error wrapping preserves cause

#### ✅ Good Pattern: Fluent Error Enhancement

```go
func (e *AppError) WithCode(code string) *AppError {
    e.Code = code
    return e
}

func (e *AppError) WithDetails(details string) *AppError {
    e.Details = details
    return e
}

func (e *AppError) WithStatusCode(code int) *AppError {
    e.StatusCode = code
    return e
}

// Usage:
return apperrors.Validation("Invalid input").
    WithCode("INVALID_EMAIL").
    WithDetails("Email must be a valid email address")
```

**Benefits:**
- Chainable API
- Optional metadata
- Readable error construction

#### ✅ Good Pattern: Error Type Checking

```go
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

func IsType(err error, errorType ErrorType) bool {
    var appErr *AppError
    if !As(err, &appErr) {
        return false
    }
    return appErr.Type == errorType
}

func GetStatusCode(err error) int {
    var appErr *AppError
    if As(err, &appErr) {
        return appErr.StatusCode
    }
    return http.StatusInternalServerError
}
```

**Benefits:**
- Works with wrapped errors
- Type-safe error inspection
- HTTP status code extraction

---

### 2.2 Logger Patterns

#### ✅ Good Pattern: Structured Logging with Zerolog

**Location:** `/pkg/logger/logger.go`

```go
type Config struct {
    Level  string
    Format string
    Output string
}

func Setup(cfg Config) error {
    level, err := zerolog.ParseLevel(strings.ToLower(cfg.Level))
    if err != nil {
        level = zerolog.InfoLevel
    }
    zerolog.SetGlobalLevel(level)

    var output io.Writer = os.Stdout
    if cfg.Format == "console" {
        output = zerolog.ConsoleWriter{
            Out:        output,
            TimeFormat: time.RFC3339,
        }
    }

    log.Logger = zerolog.New(output).With().
        Timestamp().
        Caller().
        Logger()

    return nil
}
```

**Benefits:**
- Centralized configuration
- Console vs JSON format
- Automatic timestamps and caller info
- Level-based filtering

#### ✅ Good Pattern: Context-Specific Loggers

```go
func RequestLogger(requestID, method, path string) zerolog.Logger {
    return log.With().
        Str("request_id", requestID).
        Str("method", method).
        Str("path", path).
        Logger()
}

func DatabaseLogger(operation string) zerolog.Logger {
    return log.With().
        Str("component", "database").
        Str("operation", operation).
        Logger()
}

func ScraperLogger(store, operation string) zerolog.Logger {
    return log.With().
        Str("component", "scraper").
        Str("store", store).
        Str("operation", operation).
        Logger()
}
```

**Benefits:**
- Consistent context fields
- Easy to filter logs by component
- Type-safe logger creation

#### ✅ Good Pattern: Business Event Logging

```go
func LogUserRegistration(userID, email string, source string) {
    log.Info().
        Str("component", "business_event").
        Str("event", "user_registered").
        Str("user_id", userID).
        Str("email", email).
        Str("source", source).
        Msg("New user registered")
}

func LogProductExtraction(flyerID int, pageNum, productsFound int, duration time.Duration) {
    log.Info().
        Str("component", "ai").
        Str("operation", "extraction").
        Str("event", "products_extracted").
        Int("flyer_id", flyerID).
        Int("page_number", pageNum).
        Int("products_found", productsFound).
        Dur("duration", duration).
        Msg("Products extracted from flyer")
}
```

**Benefits:**
- Structured event logging
- Queryable metrics
- Duration tracking
- Business intelligence data

---

## 3. Testing Patterns

### 3.1 Test Organization

#### ✅ Good Pattern: Table-Driven Tests

**Common throughout codebase:**

```go
func TestSomeFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {
            name: "valid input",
            input: InputType{...},
            want: OutputType{...},
            wantErr: false,
        },
        {
            name: "invalid input",
            input: InputType{...},
            want: nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := SomeFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("SomeFunction() error = %v, wantErr %v", err, tt.wantErr)
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("SomeFunction() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

**Benefits:**
- Easy to add new test cases
- Parallel execution ready
- Named subtests for clarity

### 3.2 Mock Patterns

#### ✅ Excellent Pattern: Fake Implementations with Function Fields

**Location:** `/internal/services/auth/service_test.go`

```go
type fakePasswordService struct {
    hashPasswordFn             func(password string) (string, error)
    verifyPasswordFn           func(password, hash string) error
    validatePasswordStrengthFn func(password string) error
}

func (f *fakePasswordService) HashPassword(password string) (string, error) {
    if f.hashPasswordFn != nil {
        return f.hashPasswordFn(password)
    }
    // Default implementation
    return "hashed_" + password, nil
}

func (f *fakePasswordService) VerifyPassword(password, hash string) error {
    if f.verifyPasswordFn != nil {
        return f.verifyPasswordFn(password, hash)
    }
    // Default implementation
    if hash == "hashed_"+password {
        return nil
    }
    return fmt.Errorf("invalid password")
}
```

**Benefits:**
- Flexible: can override any method
- Default behavior for simple tests
- No external mocking framework needed
- Explicit control over behavior

#### ✅ Good Pattern: Test-Specific Behavior

```go
func TestAuthService_Login_ChecksRateLimiting(t *testing.T) {
    email := "test@example.com"
    called := false

    jwtService := &fakeJWTService{}
    sessionService := &fakeSessionService{
        validateSessionFn: func(ctx context.Context, sid uuid.UUID) (*models.UserSession, error) {
            called = true
            return &models.UserSession{ID: sid, IsActive: true}, nil
        },
    }

    service := newTestAuthService(nil, &fakePasswordService{}, jwtService, sessionService, nil)

    _, err := service.Login(context.Background(), email, "password", nil)

    if !called {
        t.Fatal("expected session validation to be called")
    }
}
```

**Benefits:**
- Test verifies specific interactions
- Closure captures test variables
- Clear assertion of behavior

### 3.3 Test Data Setup

#### ✅ Good Pattern: Helper Functions for Test Data

**Location:** `/internal/repositories/session_repository_test.go`

```go
func setupSessionRepoTestDB(t *testing.T) (*bun.DB, SessionRepository, func()) {
    t.Helper()
    sqldb, err := sql.Open(sqliteshim.DriverName(), "file:session_repo_test?mode=memory&cache=shared")
    if err != nil {
        t.Fatalf("failed to open sqlite: %v", err)
    }
    db := bun.NewDB(sqldb, sqlitedialect.New())

    // Create schema
    schema := `CREATE TABLE user_sessions (...)`
    if _, err := db.ExecContext(context.Background(), schema); err != nil {
        t.Fatalf("failed to create schema: %v", err)
    }

    repo := NewSessionRepository(db)
    cleanup := func() {
        _ = db.Close()
    }
    return db, repo, cleanup
}

func insertTestSession(t *testing.T, db *bun.DB, session *models.UserSession) {
    t.Helper()
    // Set defaults
    if session.CreatedAt.IsZero() {
        session.CreatedAt = time.Now()
    }
    if session.TokenHash == "" {
        session.TokenHash = uuid.NewString()
    }

    // Insert
    if _, err := db.ExecContext(context.Background(), `INSERT INTO...`, ...); err != nil {
        t.Fatalf("failed to insert session: %v", err)
    }
}
```

**Benefits:**
- `t.Helper()` for cleaner stack traces
- Cleanup function returned
- In-memory SQLite for speed
- Sensible defaults for test data

### 3.4 Integration vs Unit Test Patterns

#### ✅ Unit Test Pattern: Dependency Injection

```go
func TestFlyerService_GetByID_NotFound(t *testing.T) {
    mockRepo := &mockFlyerRepository{
        GetByIDFn: func(ctx context.Context, id int) (*models.Flyer, error) {
            return nil, sql.ErrNoRows
        },
    }

    service := NewFlyerServiceWithRepository(mockRepo)

    _, err := service.GetByID(context.Background(), 999)

    if !apperrors.IsType(err, apperrors.ErrorTypeNotFound) {
        t.Errorf("expected NotFound error, got %v", err)
    }
}
```

#### ✅ Integration Test Pattern: Real DB

**Location:** Test files with `setupTestDB` helpers

```go
func TestSessionRepository_CleanupExpiredSessions(t *testing.T) {
    ctx := context.Background()
    db, repo, cleanup := setupSessionRepoTestDB(t)
    defer cleanup()

    // Insert expired session
    insertTestSession(t, db, &models.UserSession{
        ID:         uuid.New(),
        UserID:     uuid.New(),
        IsActive:   false,
        ExpiresAt:  time.Now().Add(-time.Hour),
    })

    // Insert valid session
    insertTestSession(t, db, &models.UserSession{
        ID:         uuid.New(),
        UserID:     uuid.New(),
        IsActive:   true,
        ExpiresAt:  time.Now().Add(time.Hour),
    })

    deleted, err := repo.CleanupExpiredSessions(ctx)

    if err != nil {
        t.Fatalf("CleanupExpiredSessions returned error: %v", err)
    }
    if deleted != 1 {
        t.Fatalf("expected to delete 1 expired session, deleted %d", deleted)
    }
}
```

**Pattern:**
- Setup real database (in-memory SQLite)
- Insert known test data
- Execute operation
- Assert results
- Cleanup via defer

---

## 4. Code Quality Patterns

### 4.1 Import Organization

#### ✅ Standard Pattern Observed

```go
package services

import (
    "context"          // Standard library
    "database/sql"
    "errors"
    "fmt"

    "github.com/google/uuid"                                    // Third-party
    "github.com/uptrace/bun"

    apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"  // Project packages (aliased)
    "github.com/kainuguru/kainuguru-api/internal/flyer"        // Project packages
    "github.com/kainuguru/kainuguru-api/internal/models"
)
```

**Pattern:**
1. Standard library imports
2. Blank line
3. Third-party imports
4. Blank line
5. Project imports (with aliases if needed)

### 4.2 Function Signatures

#### ✅ Consistent Pattern: Context First, Options Last

```go
// Standard CRUD
func GetByID(ctx context.Context, id int) (*Model, error)
func Create(ctx context.Context, model *Model) error
func Update(ctx context.Context, model *Model) error

// With filters
func GetAll(ctx context.Context, filters Filters) ([]*Model, error)

// With variadic options
func GetByID(ctx context.Context, id int, opts ...SelectOption[T]) (*T, error)
```

**Benefits:**
- Context always first (Go convention)
- Filters/options always last
- Consistent across codebase

### 4.3 Error Handling Consistency

#### ✅ Excellent Pattern: Consistent Error Wrapping

**Service Layer:**
```go
func (fs *flyerService) GetByID(ctx context.Context, id int) (*models.Flyer, error) {
    flyer, err := fs.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, apperrors.NotFound(fmt.Sprintf("flyer not found with ID %d", id))
        }
        return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get flyer by ID %d", id)
    }
    return flyer, nil
}
```

**Repository Layer:**
```go
func (r *Repository[T]) GetByID(ctx context.Context, id interface{}) (*T, error) {
    // ... query ...
    if err := query.Scan(ctx); err != nil {
        return nil, err  // Return raw error, let service wrap
    }
    return model, nil
}
```

**Pattern:**
- Repositories return raw errors
- Services wrap with context and type
- Specific error checks (sql.ErrNoRows)
- Include IDs/identifiers in error messages

### 4.4 Logging Patterns

#### ✅ Good Pattern: Structured Logging in Services

```go
func (a *authServiceImpl) Login(ctx context.Context, email, password string, metadata *SessionMetadata) (*AuthResult, error) {
    // Async logging, non-blocking
    defer func() {
        go a.RecordLoginAttempt(context.Background(), email, false, metadata)
    }()

    // ... authentication logic ...

    // Log on failure (non-blocking)
    if err != nil {
        fmt.Printf("Failed to update last login time: %v\n", err)
    }

    // Success logging
    go a.RecordLoginAttempt(context.Background(), email, true, metadata)

    return &AuthResult{...}, nil
}
```

**Pattern:**
- Async logging (goroutine) for non-critical logs
- Background context for fire-and-forget logs
- Don't fail operations on logging errors
- Structured log messages

#### ⚠️ Inconsistency: Printf vs Logger

Some code uses:
- `fmt.Printf()` for logging
- `log.Info()` structured logging
- No logging at all

**Recommendation:** Standardize on zerolog structured logging.

### 4.5 Context Usage

#### ✅ Good Pattern: Context Propagation

```go
// Context passed through all layers
func (s *Service) Method(ctx context.Context) error {
    return s.repo.Method(ctx)
}

// Context used for cancellation
func (r *Repository) GetByID(ctx context.Context, id int) (*Model, error) {
    err := r.db.NewSelect().Model(&model).Where("id = ?", id).Scan(ctx)
    return model, err
}
```

#### ✅ Good Pattern: Background Context for Async Operations

```go
// Email sending in background
go func() {
    if err := a.emailService.SendWelcomeEmail(context.Background(), user); err != nil {
        fmt.Printf("Failed to send welcome email: %v\n", err)
    }
}()

// Session activity update
go a.sessionService.UpdateSessionActivity(context.Background(), claims.SessionID)
```

**Benefits:**
- Request context for main operations
- Background context for fire-and-forget
- Proper cancellation propagation

### 4.6 Naming Conventions

#### ✅ Observed Conventions

**Interfaces:**
- Service interfaces: `FlyerService`, `ProductService`
- Repository interfaces: `Repository` (in domain package)
- Suffix: `-able` for capabilities: `Validatable`

**Implementations:**
- Unexported: `flyerService`, `productRepository`
- Suffix: `Impl` for alternate implementations: `authServiceImpl`

**Functions:**
- Constructors: `New<Type>`, `New<Type>With<Dependency>`
- Getters: `GetByID`, `GetByEmail`, `GetAll`
- Actions: `Create`, `Update`, `Delete`, `Archive`, `Process`
- Boolean: `IsActive`, `CanBeProcessed`, `HasPermission`

**Variables:**
- Receivers: Short abbreviations (`fs` for `flyerService`, `r` for `repository`)
- Loop vars: `i`, `idx`, or descriptive names
- Errors: Always `err`
- Context: Always `ctx`

### 4.7 Comment Patterns

#### ✅ Good Pattern: Godoc Comments

```go
// NewFlyerService creates a new flyer service instance backed by the shared repository.
func NewFlyerService(db *bun.DB) FlyerService {
    return NewFlyerServiceWithRepository(newFlyerRepository(db))
}

// GetByID retrieves a flyer by ID.
// Returns NotFound error if flyer doesn't exist.
func (fs *flyerService) GetByID(ctx context.Context, id int) (*models.Flyer, error) {
    // ...
}
```

#### ⚠️ Inconsistency: Comment Coverage

Some files have:
- Full godoc comments
- Inline explanations
- TODOs and explanatory comments

Others have:
- No comments at all
- Only complex logic explained

**Recommendation:** Add godoc comments to all exported types and functions.

---

## 5. Database Patterns

### 5.1 Query Builder Usage

#### ✅ Good Pattern: Bun Query Builder

```go
// Simple query
func (r *Repository[T]) GetByID(ctx context.Context, id interface{}) (*T, error) {
    model := new(T)
    err := r.db.NewSelect().
        Model(model).
        Where(fmt.Sprintf("%s = ?", r.pkColumn), id).
        Scan(ctx)
    return model, err
}

// Complex query with filters
query := r.db.NewSelect().
    Model(&products).
    Relation("Store").
    Relation("Flyer").
    Where("p.store_id IN (?)", bun.In(storeIDs)).
    Where("p.valid_to >= CURRENT_TIMESTAMP").
    Order("p.created_at DESC").
    Limit(limit).
    Offset(offset)
```

**Benefits:**
- Type-safe query building
- Automatic parameter escaping
- Relation loading
- Composable queries

### 5.2 Transaction Patterns

#### ⚠️ Pattern Not Consistently Used

**Current state:** Most operations don't use transactions explicitly.

**Example of good transaction pattern (recommended):**

```go
func (s *Service) ComplexOperation(ctx context.Context, data *Data) error {
    return s.db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
        // All operations in transaction
        if err := tx.NewInsert().Model(data).Exec(ctx); err != nil {
            return err
        }

        if err := tx.NewUpdate().Model(related).WherePK().Exec(ctx); err != nil {
            return err
        }

        return nil  // Commit
    })
}
```

**Recommendation:** Implement transaction support for:
- Multi-entity updates
- Cascading operations
- Financial/critical data changes

### 5.3 Connection Management

**Location:** `/internal/database/bun.go`

Pattern: Connection pooling via Bun's built-in mechanism

```go
db := bun.NewDB(sqldb, pgdialect.New())
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)
```

**Benefits:**
- Centralized configuration
- Connection reuse
- Prevents connection exhaustion

---

## 6. GraphQL Patterns

### 6.1 Resolver Organization

#### ✅ Good Pattern: Root Resolver with Service Composition

**Location:** `/internal/graphql/resolvers/resolver.go`

```go
type Resolver struct {
    storeService            services.StoreService
    flyerService            services.FlyerService
    flyerPageService        services.FlyerPageService
    productService          services.ProductService
    productMasterService    services.ProductMasterService
    extractionJobService    services.ExtractionJobService
    searchService           search.Service
    authService             auth.AuthService
    shoppingListService     services.ShoppingListService
    shoppingListItemService services.ShoppingListItemService
    priceHistoryService     services.PriceHistoryService
    db                      *bun.DB
}

func NewServiceResolver(
    storeService services.StoreService,
    flyerService services.FlyerService,
    // ... all services ...
    db *bun.DB,
) *Resolver {
    return &Resolver{
        storeService:            storeService,
        flyerService:            flyerService,
        // ...
        db:                      db,
    }
}
```

**Benefits:**
- All services injected at construction
- Testable via mock services
- Single source of truth for dependencies

### 6.2 DataLoader Patterns

**Expected pattern** (based on interfaces):

```go
// Service method for batch loading
func (s *Service) GetByIDs(ctx context.Context, ids []int) ([]*Model, error) {
    if len(ids) == 0 {
        return []*Model{}, nil
    }
    // Single query for all IDs
    return s.repo.GetByIDs(ctx, ids)
}

// DataLoader usage in resolver
func (r *Resolver) Products(ctx context.Context, flyer *models.Flyer) ([]*models.Product, error) {
    loader := getProductLoaderFromContext(ctx)
    return loader.Load(flyer.ID)
}
```

**Benefits:**
- N+1 query prevention
- Batched database queries
- Caching within request

### 6.3 Error Handling in Resolvers

**Expected pattern:**

```go
func (r *queryResolver) Flyer(ctx context.Context, id int) (*models.Flyer, error) {
    flyer, err := r.flyerService.GetByID(ctx, id)
    if err != nil {
        // Service already wrapped error with type
        return nil, err  // GraphQL error handling will process AppError
    }
    return flyer, nil
}
```

**Pattern:**
- Services return typed errors
- Resolvers pass through errors
- GraphQL layer converts AppError to GraphQL errors

---

## 7. Pattern Consistency Analysis

### 7.1 Highly Consistent Patterns

✅ **Error wrapping** - Consistently uses `pkg/errors` across services
✅ **Repository interfaces** - All domain packages define repository interfaces
✅ **Service constructors** - `NewXService(db)` and `NewXServiceWithRepository(repo)` pattern
✅ **Context propagation** - Context always first parameter
✅ **Filter application** - Nil-safe filter functions
✅ **Generic base repository** - Used across specific repositories

### 7.2 Moderately Consistent Patterns

⚠️ **Logging** - Mix of `fmt.Printf`, structured logging, and no logging
⚠️ **Validation** - Some in models, some in services, some in resolvers
⚠️ **Transaction usage** - Not consistently applied
⚠️ **Comment coverage** - Variable across files

### 7.3 Inconsistent Patterns

❌ **Direct DB access** - Some services use repositories, others query DB directly
❌ **Async operations** - Inconsistent use of goroutines for logging/emails
❌ **Test coverage** - Some packages well-tested, others missing tests

---

## 8. Recommended Patterns

### 8.1 For New Services

```go
package myservice

import (
    "context"
    "database/sql"
    "errors"

    apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
    "github.com/kainuguru/kainuguru-api/internal/models"
    "github.com/kainuguru/kainuguru-api/internal/myservice"
    "github.com/uptrace/bun"
)

// Interface defined in services/interfaces.go
type myServiceImpl struct {
    repo myservice.Repository
}

func NewMyService(db *bun.DB) services.MyService {
    return NewMyServiceWithRepository(newMyRepository(db))
}

func NewMyServiceWithRepository(repo myservice.Repository) services.MyService {
    if repo == nil {
        panic("myservice repository cannot be nil")
    }
    return &myServiceImpl{repo: repo}
}

func (s *myServiceImpl) GetByID(ctx context.Context, id int) (*models.MyModel, error) {
    model, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, apperrors.NotFound(fmt.Sprintf("mymodel not found with ID %d", id))
        }
        return nil, apperrors.Wrapf(err, apperrors.ErrorTypeInternal, "failed to get mymodel by ID %d", id)
    }
    return model, nil
}
```

### 8.2 For New Repositories

```go
package repositories

import (
    "context"

    "github.com/kainuguru/kainuguru-api/internal/models"
    "github.com/kainuguru/kainuguru-api/internal/mymodel"
    "github.com/kainuguru/kainuguru-api/internal/repositories/base"
    "github.com/uptrace/bun"
)

type myModelRepository struct {
    db   *bun.DB
    base *base.Repository[models.MyModel]
}

func NewMyModelRepository(db *bun.DB) mymodel.Repository {
    return &myModelRepository{
        db:   db,
        base: base.NewRepository[models.MyModel](db, "mm.id"),
    }
}

func (r *myModelRepository) GetByID(ctx context.Context, id int) (*models.MyModel, error) {
    return r.base.GetByID(ctx, id, base.WithQuery[models.MyModel](func(q *bun.SelectQuery) *bun.SelectQuery {
        return attachMyModelRelations(q)
    }))
}

func (r *myModelRepository) GetAll(ctx context.Context, filters *mymodel.Filters) ([]*models.MyModel, error) {
    return r.base.GetAll(ctx, base.WithQuery[models.MyModel](func(q *bun.SelectQuery) *bun.SelectQuery {
        q = attachMyModelRelations(q)
        q = applyMyModelFilters(q, filters)
        return applyMyModelPagination(q, filters)
    }))
}

func attachMyModelRelations(q *bun.SelectQuery) *bun.SelectQuery {
    return q.Relation("RelatedEntity")
}

func applyMyModelFilters(q *bun.SelectQuery, filters *mymodel.Filters) *bun.SelectQuery {
    if filters == nil {
        return q
    }
    if len(filters.IDs) > 0 {
        q.Where("mm.id IN (?)", bun.In(filters.IDs))
    }
    return q
}
```

### 8.3 For New Models

```go
package models

import (
    "time"

    "github.com/uptrace/bun"
)

type MyModel struct {
    bun.BaseModel `bun:"table:my_models,alias:mm"`

    ID          int       `bun:"id,pk,autoincrement" json:"id"`
    Name        string    `bun:"name,notnull" json:"name"`
    Description *string   `bun:"description" json:"description"`
    IsActive    bool      `bun:"is_active,default:true" json:"is_active"`

    // Timestamps
    CreatedAt time.Time `bun:"created_at,default:now()" json:"created_at"`
    UpdatedAt time.Time `bun:"updated_at,default:now()" json:"updated_at"`

    // Relations
    RelatedEntity *RelatedEntity `bun:"rel:belongs-to,join:related_id=id" json:"related_entity,omitempty"`
}

// Validate validates the model
func (m *MyModel) Validate() error {
    if m.Name == "" {
        return NewValidationError("name", "Name is required")
    }
    return nil
}

// Business logic methods
func (m *MyModel) Activate() {
    m.IsActive = true
    m.UpdatedAt = time.Now()
}

func (m *MyModel) Deactivate() {
    m.IsActive = false
    m.UpdatedAt = time.Now()
}
```

### 8.4 For Tests

```go
package myservice_test

import (
    "context"
    "testing"

    "github.com/kainuguru/kainuguru-api/internal/models"
    "github.com/kainuguru/kainuguru-api/internal/services"
    apperrors "github.com/kainuguru/kainuguru-api/pkg/errors"
)

type fakeMyRepository struct {
    getByIDFn func(ctx context.Context, id int) (*models.MyModel, error)
}

func (f *fakeMyRepository) GetByID(ctx context.Context, id int) (*models.MyModel, error) {
    if f.getByIDFn != nil {
        return f.getByIDFn(ctx, id)
    }
    return &models.MyModel{ID: id}, nil
}

func TestMyService_GetByID_Success(t *testing.T) {
    repo := &fakeMyRepository{
        getByIDFn: func(ctx context.Context, id int) (*models.MyModel, error) {
            return &models.MyModel{ID: id, Name: "Test"}, nil
        },
    }

    service := services.NewMyServiceWithRepository(repo)

    model, err := service.GetByID(context.Background(), 1)

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if model.ID != 1 {
        t.Errorf("expected ID 1, got %d", model.ID)
    }
}

func TestMyService_GetByID_NotFound(t *testing.T) {
    repo := &fakeMyRepository{
        getByIDFn: func(ctx context.Context, id int) (*models.MyModel, error) {
            return nil, sql.ErrNoRows
        },
    }

    service := services.NewMyServiceWithRepository(repo)

    _, err := service.GetByID(context.Background(), 999)

    if !apperrors.IsType(err, apperrors.ErrorTypeNotFound) {
        t.Errorf("expected NotFound error, got %v", err)
    }
}
```

---

## Summary

This codebase demonstrates **strong architectural patterns** with:

1. **Clear separation of concerns** (Repository → Service → Handler/Resolver)
2. **Excellent error handling** with typed, wrappable errors
3. **Strong testability** through dependency injection
4. **Modern Go patterns** (generics, functional options, interfaces)
5. **Good database abstractions** (base repository, query builders)

**Areas for improvement:**

1. **Logging consistency** - Standardize on structured logging
2. **Transaction usage** - Apply consistently for multi-entity operations
3. **Direct DB access** - Move all queries to repositories
4. **Test coverage** - Increase coverage in underrepresented areas
5. **Documentation** - Add godoc comments to all exported APIs

**Overall Assessment:** 8.5/10 - Well-structured, maintainable, and following Go best practices with room for consistency improvements.
