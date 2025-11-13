# Go-Specific Refactoring Patterns

## Essential Go Refactoring Patterns with Examples

This document provides concrete, copy-paste-ready refactoring patterns specifically for Go codebases. Each pattern includes detection criteria, implementation steps, and real examples from this codebase.

---

## 1. Generic Repository Pattern (Eliminate CRUD Duplication)

### Detection
Look for repeated CRUD operations across services:
```bash
grep -r "func.*GetByID" --include="*.go" | wc -l  # If > 5, refactor needed
```

### Implementation

#### Step 1: Create Generic Repository
```go
// repository/generic_repository.go
package repository

import (
    "context"
    "fmt"
    "gorm.io/gorm"
)

// Repository provides generic CRUD operations for any entity type
type Repository[T any] struct {
    db        *gorm.DB
    tableName string
}

// NewRepository creates a new generic repository instance
func NewRepository[T any](db *gorm.DB, tableName string) *Repository[T] {
    return &Repository[T]{
        db:        db,
        tableName: tableName,
    }
}

// Create inserts a new entity
func (r *Repository[T]) Create(ctx context.Context, entity *T) error {
    return r.db.WithContext(ctx).Table(r.tableName).Create(entity).Error
}

// GetByID retrieves an entity by ID
func (r *Repository[T]) GetByID(ctx context.Context, id interface{}) (*T, error) {
    var entity T
    err := r.db.WithContext(ctx).Table(r.tableName).Where("id = ?", id).First(&entity).Error
    if err != nil {
        return nil, fmt.Errorf("failed to get %s by id %v: %w", r.tableName, id, err)
    }
    return &entity, nil
}

// GetAll retrieves all entities with optional preloading
func (r *Repository[T]) GetAll(ctx context.Context, preloads ...string) ([]T, error) {
    var entities []T
    query := r.db.WithContext(ctx).Table(r.tableName)

    for _, preload := range preloads {
        query = query.Preload(preload)
    }

    err := query.Find(&entities).Error
    return entities, err
}

// Update modifies an existing entity
func (r *Repository[T]) Update(ctx context.Context, id interface{}, updates map[string]interface{}) error {
    return r.db.WithContext(ctx).Table(r.tableName).Where("id = ?", id).Updates(updates).Error
}

// Delete removes an entity
func (r *Repository[T]) Delete(ctx context.Context, id interface{}) error {
    var entity T
    return r.db.WithContext(ctx).Table(r.tableName).Where("id = ?", id).Delete(&entity).Error
}

// Count returns the total number of entities
func (r *Repository[T]) Count(ctx context.Context) (int64, error) {
    var count int64
    err := r.db.WithContext(ctx).Table(r.tableName).Count(&count).Error
    return count, err
}

// Exists checks if an entity exists by ID
func (r *Repository[T]) Exists(ctx context.Context, id interface{}) (bool, error) {
    var count int64
    err := r.db.WithContext(ctx).Table(r.tableName).Where("id = ?", id).Count(&count).Error
    return count > 0, err
}
```

#### Step 2: Implement Specific Repositories
```go
// repository/user_repository.go
package repository

import (
    "context"
    "kainuguru/models"
    "gorm.io/gorm"
)

type UserRepository struct {
    *Repository[models.User]  // Embed generic repository
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{
        Repository: NewRepository[models.User](db, "users"),
        db:        db,
    }
}

// Add domain-specific methods
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
    var user models.User
    err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
    if err != nil {
        return nil, fmt.Errorf("failed to find user by email %s: %w", email, err)
    }
    return &user, nil
}

func (r *UserRepository) FindActiveUsers(ctx context.Context) ([]models.User, error) {
    var users []models.User
    err := r.db.WithContext(ctx).Where("is_active = ?", true).Find(&users).Error
    return users, err
}
```

---

## 2. Functional Options Pattern (Clean Constructors)

### Detection
Look for constructors with > 3 parameters or multiple constructor variants:
```bash
grep -r "func New.*(" --include="*.go" | grep -E "\(.*,.*,.*,.*\)"
```

### Implementation

```go
// service/options.go
package service

import (
    "time"
    "go.uber.org/zap"
)

// ServiceConfig holds all configuration for a service
type ServiceConfig struct {
    Logger      *zap.Logger
    Timeout     time.Duration
    MaxRetries  int
    CacheEnabled bool
    BatchSize   int
}

// Option is a function that modifies ServiceConfig
type Option func(*ServiceConfig)

// DefaultConfig returns default configuration
func DefaultConfig() *ServiceConfig {
    return &ServiceConfig{
        Timeout:    30 * time.Second,
        MaxRetries: 3,
        BatchSize:  100,
    }
}

// WithLogger sets the logger
func WithLogger(logger *zap.Logger) Option {
    return func(cfg *ServiceConfig) {
        cfg.Logger = logger
    }
}

// WithTimeout sets the timeout
func WithTimeout(timeout time.Duration) Option {
    return func(cfg *ServiceConfig) {
        cfg.Timeout = timeout
    }
}

// WithCache enables caching
func WithCache(enabled bool) Option {
    return func(cfg *ServiceConfig) {
        cfg.CacheEnabled = enabled
    }
}

// WithBatchSize sets the batch processing size
func WithBatchSize(size int) Option {
    return func(cfg *ServiceConfig) {
        if size > 0 {
            cfg.BatchSize = size
        }
    }
}

// Usage example:
type ProductService struct {
    config *ServiceConfig
    repo   *ProductRepository
}

func NewProductService(repo *ProductRepository, opts ...Option) *ProductService {
    config := DefaultConfig()

    // Apply all options
    for _, opt := range opts {
        opt(config)
    }

    return &ProductService{
        config: config,
        repo:   repo,
    }
}

// Clean usage:
// service := NewProductService(repo,
//     WithLogger(logger),
//     WithTimeout(5*time.Second),
//     WithCache(true),
// )
```

---

## 3. Error Wrapping Pattern (Consistent Error Handling)

### Detection
```bash
grep -r "return nil, err" --include="*.go" | wc -l  # If > 50, needs standardization
```

### Implementation

```go
// pkg/errors/errors.go
package errors

import (
    "errors"
    "fmt"
)

// Domain errors
var (
    ErrNotFound      = errors.New("entity not found")
    ErrAlreadyExists = errors.New("entity already exists")
    ErrInvalidInput  = errors.New("invalid input")
    ErrUnauthorized  = errors.New("unauthorized")
    ErrForbidden     = errors.New("forbidden")
    ErrInternal      = errors.New("internal error")
)

// Error types for categorization
type ErrorType int

const (
    ErrorTypeValidation ErrorType = iota
    ErrorTypeNotFound
    ErrorTypeDuplicate
    ErrorTypeAuthorization
    ErrorTypeInternal
)

// AppError represents an application error with context
type AppError struct {
    Type    ErrorType
    Message string
    Err     error
    Context map[string]interface{}
}

func (e *AppError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Err)
    }
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Err
}

// Constructors for common error types
func NewValidationError(message string, err error) *AppError {
    return &AppError{
        Type:    ErrorTypeValidation,
        Message: message,
        Err:     err,
    }
}

func NewNotFoundError(entity string, id interface{}) *AppError {
    return &AppError{
        Type:    ErrorTypeNotFound,
        Message: fmt.Sprintf("%s not found", entity),
        Context: map[string]interface{}{
            "entity": entity,
            "id":     id,
        },
    }
}

func NewDuplicateError(entity string, field string, value interface{}) *AppError {
    return &AppError{
        Type:    ErrorTypeDuplicate,
        Message: fmt.Sprintf("%s with %s '%v' already exists", entity, field, value),
        Context: map[string]interface{}{
            "entity": entity,
            "field":  field,
            "value":  value,
        },
    }
}

// Error wrapping utilities
func Wrap(err error, message string) error {
    if err == nil {
        return nil
    }
    return fmt.Errorf("%s: %w", message, err)
}

func WrapWithContext(err error, message string, ctx map[string]interface{}) *AppError {
    return &AppError{
        Type:    ErrorTypeInternal,
        Message: message,
        Err:     err,
        Context: ctx,
    }
}

// Usage in service:
func (s *UserService) GetUser(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, NewNotFoundError("user", id)
        }
        return nil, Wrap(err, "failed to get user")
    }
    return user, nil
}
```

---

## 4. Middleware Chain Pattern (Composable Middleware)

### Detection
```bash
find . -name "*middleware*.go" -exec wc -l {} + | awk '$1 > 100'
```

### Implementation

```go
// middleware/chain.go
package middleware

import (
    "github.com/gofiber/fiber/v2"
    "time"
)

// Middleware represents a Fiber middleware function
type Middleware func(*fiber.Ctx) error

// Chain creates a middleware chain from multiple middlewares
type Chain struct {
    middlewares []Middleware
}

// NewChain creates a new middleware chain
func NewChain(middlewares ...Middleware) *Chain {
    return &Chain{middlewares: middlewares}
}

// Append adds middlewares to the chain
func (c *Chain) Append(middlewares ...Middleware) *Chain {
    c.middlewares = append(c.middlewares, middlewares...)
    return c
}

// Handler returns the final handler with all middlewares applied
func (c *Chain) Handler(handler fiber.Handler) fiber.Handler {
    for i := len(c.middlewares) - 1; i >= 0; i-- {
        final := handler
        mw := c.middlewares[i]
        handler = func(ctx *fiber.Ctx) error {
            return mw(ctx)
        }
    }
    return handler
}

// Common middleware implementations
func Logger(logger *zap.Logger) Middleware {
    return func(c *fiber.Ctx) error {
        start := time.Now()

        err := c.Next()

        logger.Info("request",
            zap.String("method", c.Method()),
            zap.String("path", c.Path()),
            zap.Int("status", c.Response().StatusCode()),
            zap.Duration("latency", time.Since(start)),
        )

        return err
    }
}

func RateLimiter(limit int, window time.Duration) Middleware {
    limiter := rate.NewLimiter(rate.Every(window/time.Duration(limit)), limit)

    return func(c *fiber.Ctx) error {
        if !limiter.Allow() {
            return c.Status(429).JSON(fiber.Map{
                "error": "rate limit exceeded",
            })
        }
        return c.Next()
    }
}

func RequireAuth(authService *AuthService) Middleware {
    return func(c *fiber.Ctx) error {
        token := c.Get("Authorization")
        if token == "" {
            return c.Status(401).JSON(fiber.Map{
                "error": "missing authorization token",
            })
        }

        user, err := authService.ValidateToken(c.Context(), token)
        if err != nil {
            return c.Status(401).JSON(fiber.Map{
                "error": "invalid token",
            })
        }

        c.Locals("user", user)
        return c.Next()
    }
}

// Usage:
func SetupRoutes(app *fiber.App, services *Services) {
    // Create reusable chains
    publicChain := NewChain(
        Logger(services.Logger),
        RateLimiter(100, time.Minute),
    )

    protectedChain := NewChain(
        Logger(services.Logger),
        RateLimiter(100, time.Minute),
        RequireAuth(services.Auth),
    )

    // Apply to routes
    app.Get("/public", publicChain.Handler(publicHandler))
    app.Get("/protected", protectedChain.Handler(protectedHandler))
}
```

---

## 5. Builder Pattern for Complex Queries

### Detection
```bash
grep -r "db\." --include="*.go" | grep -E "Where.*Where.*Where" | wc -l
```

### Implementation

```go
// repository/query_builder.go
package repository

import (
    "fmt"
    "gorm.io/gorm"
)

// QueryBuilder helps construct complex database queries
type QueryBuilder struct {
    db         *gorm.DB
    conditions []condition
    preloads   []string
    order      string
    limit      int
    offset     int
}

type condition struct {
    query string
    args  []interface{}
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder(db *gorm.DB) *QueryBuilder {
    return &QueryBuilder{
        db:         db,
        conditions: make([]condition, 0),
        preloads:   make([]string, 0),
    }
}

// Where adds a WHERE condition
func (qb *QueryBuilder) Where(query string, args ...interface{}) *QueryBuilder {
    qb.conditions = append(qb.conditions, condition{
        query: query,
        args:  args,
    })
    return qb
}

// WhereIf adds a WHERE condition only if the condition is true
func (qb *QueryBuilder) WhereIf(cond bool, query string, args ...interface{}) *QueryBuilder {
    if cond {
        return qb.Where(query, args...)
    }
    return qb
}

// WhereNotEmpty adds a WHERE condition only if value is not empty
func (qb *QueryBuilder) WhereNotEmpty(field string, value string) *QueryBuilder {
    if value != "" {
        return qb.Where(fmt.Sprintf("%s = ?", field), value)
    }
    return qb
}

// Preload adds a preload association
func (qb *QueryBuilder) Preload(association string) *QueryBuilder {
    qb.preloads = append(qb.preloads, association)
    return qb
}

// OrderBy sets the ORDER BY clause
func (qb *QueryBuilder) OrderBy(order string) *QueryBuilder {
    qb.order = order
    return qb
}

// Limit sets the LIMIT
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
    qb.limit = limit
    return qb
}

// Offset sets the OFFSET
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
    qb.offset = offset
    return qb
}

// Paginate sets limit and offset for pagination
func (qb *QueryBuilder) Paginate(page, pageSize int) *QueryBuilder {
    if page < 1 {
        page = 1
    }
    if pageSize < 1 {
        pageSize = 10
    }

    qb.offset = (page - 1) * pageSize
    qb.limit = pageSize
    return qb
}

// Build constructs the final query
func (qb *QueryBuilder) Build() *gorm.DB {
    query := qb.db

    // Apply conditions
    for _, cond := range qb.conditions {
        query = query.Where(cond.query, cond.args...)
    }

    // Apply preloads
    for _, preload := range qb.preloads {
        query = query.Preload(preload)
    }

    // Apply order
    if qb.order != "" {
        query = query.Order(qb.order)
    }

    // Apply limit and offset
    if qb.limit > 0 {
        query = query.Limit(qb.limit)
    }
    if qb.offset > 0 {
        query = query.Offset(qb.offset)
    }

    return query
}

// FindAll executes the query and returns results
func (qb *QueryBuilder) FindAll(dest interface{}) error {
    return qb.Build().Find(dest).Error
}

// First executes the query and returns first result
func (qb *QueryBuilder) First(dest interface{}) error {
    return qb.Build().First(dest).Error
}

// Count returns the count of matching records
func (qb *QueryBuilder) Count() (int64, error) {
    var count int64
    err := qb.Build().Count(&count).Error
    return count, err
}

// Usage example:
func (r *ProductRepository) SearchProducts(filters ProductFilters) ([]Product, error) {
    var products []Product

    err := NewQueryBuilder(r.db).
        WhereNotEmpty("category", filters.Category).
        WhereIf(filters.MinPrice > 0, "price >= ?", filters.MinPrice).
        WhereIf(filters.MaxPrice > 0, "price <= ?", filters.MaxPrice).
        Where("is_active = ?", true).
        Preload("Category").
        Preload("Images").
        OrderBy("created_at DESC").
        Paginate(filters.Page, filters.PageSize).
        FindAll(&products)

    return products, err
}
```

---

## 6. Result Type Pattern (Better Error Handling)

### Detection
Multiple return values with error:
```bash
grep -r "func.*) (.*,.*error)" --include="*.go" | wc -l
```

### Implementation

```go
// pkg/result/result.go
package result

// Result represents either a success value or an error
type Result[T any] struct {
    value *T
    err   error
}

// Ok creates a successful result
func Ok[T any](value T) Result[T] {
    return Result[T]{value: &value}
}

// Err creates an error result
func Err[T any](err error) Result[T] {
    return Result[T]{err: err}
}

// IsOk returns true if the result is successful
func (r Result[T]) IsOk() bool {
    return r.err == nil
}

// IsErr returns true if the result is an error
func (r Result[T]) IsErr() bool {
    return r.err != nil
}

// Unwrap returns the value, panics if error
func (r Result[T]) Unwrap() T {
    if r.err != nil {
        panic(r.err)
    }
    return *r.value
}

// UnwrapOr returns the value or a default if error
func (r Result[T]) UnwrapOr(defaultValue T) T {
    if r.err != nil {
        return defaultValue
    }
    return *r.value
}

// UnwrapErr returns the error
func (r Result[T]) UnwrapErr() error {
    return r.err
}

// Map transforms the value if successful
func (r Result[T]) Map(fn func(T) T) Result[T] {
    if r.err != nil {
        return r
    }
    return Ok(fn(*r.value))
}

// MapErr transforms the error if present
func (r Result[T]) MapErr(fn func(error) error) Result[T] {
    if r.err == nil {
        return r
    }
    return Err[T](fn(r.err))
}

// AndThen chains operations that return Results
func (r Result[T]) AndThen(fn func(T) Result[T]) Result[T] {
    if r.err != nil {
        return r
    }
    return fn(*r.value)
}

// Usage:
func GetUserByEmail(email string) Result[*User] {
    if email == "" {
        return Err[*User](errors.New("email is required"))
    }

    user, err := db.FindUserByEmail(email)
    if err != nil {
        return Err[*User](err)
    }

    return Ok(user)
}

func ProcessUser(email string) Result[string] {
    return GetUserByEmail(email).
        Map(func(u *User) *User {
            u.LastActive = time.Now()
            return u
        }).
        AndThen(func(u *User) Result[string] {
            if err := db.Save(u); err != nil {
                return Err[string](err)
            }
            return Ok(u.Name)
        })
}
```

---

## 7. Context Value Pattern (Type-Safe Context)

### Detection
```bash
grep -r "context.Value" --include="*.go" | wc -l
```

### Implementation

```go
// pkg/ctxutil/context.go
package ctxutil

import (
    "context"
    "kainuguru/models"
)

// Define typed keys for context values
type contextKey int

const (
    userKey contextKey = iota
    requestIDKey
    tenantKey
    traceIDKey
)

// WithUser adds user to context
func WithUser(ctx context.Context, user *models.User) context.Context {
    return context.WithValue(ctx, userKey, user)
}

// GetUser retrieves user from context
func GetUser(ctx context.Context) (*models.User, bool) {
    user, ok := ctx.Value(userKey).(*models.User)
    return user, ok
}

// MustGetUser retrieves user from context or panics
func MustGetUser(ctx context.Context) *models.User {
    user, ok := GetUser(ctx)
    if !ok {
        panic("user not found in context")
    }
    return user
}

// WithRequestID adds request ID to context
func WithRequestID(ctx context.Context, requestID string) context.Context {
    return context.WithValue(ctx, requestIDKey, requestID)
}

// GetRequestID retrieves request ID from context
func GetRequestID(ctx context.Context) (string, bool) {
    id, ok := ctx.Value(requestIDKey).(string)
    return id, ok
}

// WithTenant adds tenant to context
func WithTenant(ctx context.Context, tenantID string) context.Context {
    return context.WithValue(ctx, tenantKey, tenantID)
}

// GetTenant retrieves tenant from context
func GetTenant(ctx context.Context) (string, bool) {
    tenant, ok := ctx.Value(tenantKey).(string)
    return tenant, ok
}

// Usage in middleware:
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        user, err := validateToken(r.Header.Get("Authorization"))
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        ctx := WithUser(r.Context(), user)
        ctx = WithRequestID(ctx, generateRequestID())

        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// Usage in handler:
func GetProfile(w http.ResponseWriter, r *http.Request) {
    user, ok := GetUser(r.Context())
    if !ok {
        http.Error(w, "User not found", http.StatusInternalServerError)
        return
    }

    // Use user safely
    json.NewEncoder(w).Encode(user)
}
```

---

## 8. Validation Pattern (Centralized Validation)

### Detection
```bash
grep -r "if.*==\"\"" --include="*.go" | wc -l  # Basic validation scattered
```

### Implementation

```go
// pkg/validation/validator.go
package validation

import (
    "fmt"
    "regexp"
    "strings"
)

// ValidationError represents a validation error
type ValidationError struct {
    Field   string
    Message string
    Code    string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
    var messages []string
    for _, err := range e {
        messages = append(messages, err.Error())
    }
    return strings.Join(messages, "; ")
}

// Validator provides validation methods
type Validator struct {
    errors ValidationErrors
}

// New creates a new validator
func New() *Validator {
    return &Validator{
        errors: make(ValidationErrors, 0),
    }
}

// Required checks if a field is not empty
func (v *Validator) Required(field, value string) *Validator {
    if strings.TrimSpace(value) == "" {
        v.errors = append(v.errors, ValidationError{
            Field:   field,
            Message: "is required",
            Code:    "REQUIRED",
        })
    }
    return v
}

// MinLength checks minimum length
func (v *Validator) MinLength(field, value string, min int) *Validator {
    if len(value) < min {
        v.errors = append(v.errors, ValidationError{
            Field:   field,
            Message: fmt.Sprintf("must be at least %d characters", min),
            Code:    "MIN_LENGTH",
        })
    }
    return v
}

// MaxLength checks maximum length
func (v *Validator) MaxLength(field, value string, max int) *Validator {
    if len(value) > max {
        v.errors = append(v.errors, ValidationError{
            Field:   field,
            Message: fmt.Sprintf("must not exceed %d characters", max),
            Code:    "MAX_LENGTH",
        })
    }
    return v
}

// Email validates email format
func (v *Validator) Email(field, value string) *Validator {
    emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
    if !emailRegex.MatchString(value) {
        v.errors = append(v.errors, ValidationError{
            Field:   field,
            Message: "must be a valid email address",
            Code:    "INVALID_EMAIL",
        })
    }
    return v
}

// Range checks if value is within range
func (v *Validator) Range(field string, value, min, max float64) *Validator {
    if value < min || value > max {
        v.errors = append(v.errors, ValidationError{
            Field:   field,
            Message: fmt.Sprintf("must be between %v and %v", min, max),
            Code:    "OUT_OF_RANGE",
        })
    }
    return v
}

// Custom adds a custom validation
func (v *Validator) Custom(field string, valid bool, message string) *Validator {
    if !valid {
        v.errors = append(v.errors, ValidationError{
            Field:   field,
            Message: message,
            Code:    "CUSTOM",
        })
    }
    return v
}

// Valid returns true if no errors
func (v *Validator) Valid() bool {
    return len(v.errors) == 0
}

// Errors returns all validation errors
func (v *Validator) Errors() ValidationErrors {
    return v.errors
}

// Validate returns error if validation fails
func (v *Validator) Validate() error {
    if !v.Valid() {
        return v.errors
    }
    return nil
}

// Domain validation
type UserValidator struct{}

func (uv *UserValidator) ValidateCreate(user *User) error {
    v := New()

    v.Required("email", user.Email).
        Email("email", user.Email).
        Required("name", user.Name).
        MinLength("name", user.Name, 2).
        MaxLength("name", user.Name, 100).
        MinLength("password", user.Password, 8).
        Custom("age", user.Age >= 18, "must be at least 18 years old")

    return v.Validate()
}

// Usage in handler:
func CreateUser(req CreateUserRequest) error {
    validator := &UserValidator{}
    if err := validator.ValidateCreate(&req); err != nil {
        return err // Returns detailed validation errors
    }

    // Proceed with creation
    return userService.Create(req)
}
```

---

## 9. Dependency Injection Container

### Detection
```bash
grep -r "New.*Service" --include="*.go" | grep -E "New.*\(.*,.*,.*,.*\)" | wc -l
```

### Implementation

```go
// pkg/container/container.go
package container

import (
    "fmt"
    "reflect"
    "sync"
)

// Container manages dependencies
type Container struct {
    services map[reflect.Type]interface{}
    mu       sync.RWMutex
}

// New creates a new container
func New() *Container {
    return &Container{
        services: make(map[reflect.Type]interface{}),
    }
}

// Register registers a service
func (c *Container) Register(service interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()

    t := reflect.TypeOf(service)
    c.services[t] = service
}

// RegisterAs registers a service with a specific interface
func (c *Container) RegisterAs(service interface{}, iface interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()

    t := reflect.TypeOf(iface).Elem()
    c.services[t] = service
}

// Get retrieves a service
func (c *Container) Get(iface interface{}) (interface{}, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    t := reflect.TypeOf(iface).Elem()
    service, ok := c.services[t]
    if !ok {
        return nil, fmt.Errorf("service %s not found", t.String())
    }

    return service, nil
}

// MustGet retrieves a service or panics
func (c *Container) MustGet(iface interface{}) interface{} {
    service, err := c.Get(iface)
    if err != nil {
        panic(err)
    }
    return service
}

// Inject injects dependencies into a struct
func (c *Container) Inject(target interface{}) error {
    v := reflect.ValueOf(target).Elem()
    t := v.Type()

    for i := 0; i < t.NumField(); i++ {
        field := t.Field(i)
        if tag := field.Tag.Get("inject"); tag != "" {
            fieldType := field.Type
            service, err := c.Get(reflect.New(fieldType).Interface())
            if err != nil {
                return fmt.Errorf("failed to inject %s: %w", field.Name, err)
            }
            v.Field(i).Set(reflect.ValueOf(service))
        }
    }

    return nil
}

// Application setup:
func SetupContainer(config *Config) *Container {
    c := New()

    // Register services
    db := setupDatabase(config.Database)
    c.Register(db)

    logger := setupLogger(config.Logger)
    c.Register(logger)

    // Register repositories
    userRepo := repository.NewUserRepository(db)
    c.RegisterAs(userRepo, (*repository.UserRepository)(nil))

    productRepo := repository.NewProductRepository(db)
    c.RegisterAs(productRepo, (*repository.ProductRepository)(nil))

    // Register services with dependencies
    userService := service.NewUserService(userRepo, logger)
    c.RegisterAs(userService, (*service.UserService)(nil))

    return c
}

// Usage with injection:
type Handler struct {
    UserService    *service.UserService    `inject:"true"`
    ProductService *service.ProductService `inject:"true"`
    Logger         *zap.Logger            `inject:"true"`
}

func NewHandler(c *Container) *Handler {
    h := &Handler{}
    if err := c.Inject(h); err != nil {
        panic(err)
    }
    return h
}
```

---

## Quick Reference Commands

### Find Refactoring Candidates
```bash
# Large files
find . -name "*.go" -exec wc -l {} + | sort -rn | head -20

# Complex functions
gocyclo -over 10 . | head -20

# Duplication
dupl -t 50 . | head -20

# Long functions
grep -n "^func" *.go | awk -F: '{print $1":"$2}' | while read func; do
    file=$(echo $func | cut -d: -f1)
    start=$(echo $func | cut -d: -f2)
    end=$(awk "/^func/ && NR>$start {print NR; exit}" $file)
    if [ ! -z "$end" ]; then
        lines=$((end - start))
        if [ $lines -gt 50 ]; then
            echo "$file:$start $lines lines"
        fi
    fi
done | sort -k2 -rn

# Missing tests
for file in $(find . -name "*.go" ! -name "*_test.go"); do
    testfile="${file%.go}_test.go"
    if [ ! -f "$testfile" ]; then
        echo "Missing test: $file"
    fi
done
```

---

## Testing Each Pattern

Every refactoring pattern should include tests:

```go
// repository/generic_repository_test.go
func TestGenericRepository_CRUD(t *testing.T) {
    db := setupTestDB(t)
    repo := NewRepository[TestEntity](db, "test_entities")

    t.Run("Create", func(t *testing.T) {
        entity := &TestEntity{Name: "test"}
        err := repo.Create(context.Background(), entity)
        assert.NoError(t, err)
        assert.NotEmpty(t, entity.ID)
    })

    t.Run("GetByID", func(t *testing.T) {
        // ... test implementation
    })
}
```

---

This document provides concrete, tested patterns ready for immediate use in the Kainuguru API refactoring effort.