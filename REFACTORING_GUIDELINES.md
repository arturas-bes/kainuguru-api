# Go Refactoring Guidelines for Kainuguru API

## Table of Contents
1. [Core Refactoring Principles](#core-refactoring-principles)
2. [Go-Specific Guidelines](#go-specific-guidelines)
3. [Refactoring Process](#refactoring-process)
4. [Code Smell Detection](#code-smell-detection)
5. [Refactoring Patterns](#refactoring-patterns)
6. [Testing Strategy](#testing-strategy)
7. [Safety Checklist](#safety-checklist)
8. [Measurement Criteria](#measurement-criteria)

---

## Core Refactoring Principles

### 1. The Boy Scout Rule
> "Always leave the code better than you found it"

- Make small improvements whenever touching code
- Fix obvious issues immediately (typos, formatting, simple bugs)
- Don't ignore technical debt - document it if you can't fix it now

### 2. Small, Incremental Changes
- **One refactoring at a time** - Don't mix multiple refactoring patterns
- **Commit frequently** - Each logical change should be a separate commit
- **Keep PRs focused** - Maximum 400 lines changed per PR
- **Maintain working state** - Code must compile and pass tests after each change

### 3. Test-First Refactoring
```go
// WRONG: Refactor then test
func RefactorCode() {
    // Change code
    // Hope it works
    // Write tests maybe
}

// RIGHT: Test coverage first
func RefactorCodeSafely() {
    // 1. Write tests for existing behavior
    // 2. Verify tests pass
    // 3. Refactor code
    // 4. Verify tests still pass
}
```

### 4. Preserve Behavior
- **Refactoring !== Feature Addition**
- External behavior must remain identical
- API contracts must not break
- Performance should not degrade (measure before/after)

### 5. Clear Communication
- Document WHY, not WHAT in commits
- Update documentation immediately
- Add deprecation notices for replaced code
- Communicate breaking changes clearly

---

## Go-Specific Guidelines

### 1. Embrace Go Idioms

#### Use Named Returns Sparingly
```go
// AVOID: Named returns without clear benefit
func Calculate() (result int, err error) {
    result = 42
    return // Implicit return is confusing
}

// PREFER: Explicit returns
func Calculate() (int, error) {
    result := 42
    return result, nil
}

// EXCEPTION: Named returns for documentation
func ParseConfig() (config *Config, warnings []string, err error) {
    // Named returns clarify complex return types
    return config, warnings, nil
}
```

#### Error Handling Pattern
```go
// WRONG: Ignoring errors
value, _ := strconv.Atoi(input)

// WRONG: Nested error handling
if err := doFirst(); err == nil {
    if err := doSecond(); err == nil {
        // Deep nesting
    }
}

// RIGHT: Early returns
if err := doFirst(); err != nil {
    return fmt.Errorf("first operation failed: %w", err)
}

if err := doSecond(); err != nil {
    return fmt.Errorf("second operation failed: %w", err)
}
```

### 2. Interface Segregation

```go
// WRONG: Fat interface
type DataStore interface {
    Create(ctx context.Context, data interface{}) error
    Read(ctx context.Context, id string) (interface{}, error)
    Update(ctx context.Context, id string, data interface{}) error
    Delete(ctx context.Context, id string) error
    BeginTransaction() (*sql.Tx, error)
    Migrate() error
    Backup() error
    // Too many responsibilities
}

// RIGHT: Focused interfaces
type Reader interface {
    Read(ctx context.Context, id string) (interface{}, error)
}

type Writer interface {
    Create(ctx context.Context, data interface{}) error
    Update(ctx context.Context, id string, data interface{}) error
    Delete(ctx context.Context, id string) error
}

type Migrator interface {
    Migrate() error
}
```

### 3. Avoid Premature Optimization

```go
// WRONG: Premature optimization
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func ProcessSimpleString(s string) string {
    // Using sync.Pool for simple operations
    buf := bufferPool.Get().(*bytes.Buffer)
    defer bufferPool.Put(buf)
    buf.Reset()
    buf.WriteString(s)
    return buf.String()
}

// RIGHT: Simple and clear
func ProcessSimpleString(s string) string {
    return strings.ToUpper(s)
}
```

### 4. Proper Context Usage

```go
// WRONG: Creating new context
func (s *Service) Process(ctx context.Context, data string) error {
    // Never do this - breaks cancellation chain
    newCtx := context.Background()
    return s.repo.Save(newCtx, data)
}

// RIGHT: Propagate context
func (s *Service) Process(ctx context.Context, data string) error {
    // Add timeout if needed, but keep parent chain
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    return s.repo.Save(ctx, data)
}
```

---

## Refactoring Process

### Step 1: Identify Target
1. Use metrics to find problem areas:
   - Cyclomatic complexity > 10
   - Functions > 50 lines
   - Files > 500 lines
   - Test coverage < 70%
   - High coupling (> 5 dependencies)

### Step 2: Characterize Existing Behavior
```bash
# Create characterization tests
go test -v -run TestExistingBehavior
go test -bench=. -benchmem > before_bench.txt
```

### Step 3: Prepare Safety Net
```go
// Add integration test for critical paths
func TestCriticalPath(t *testing.T) {
    // Capture current behavior exactly
    input := getCurrentProductionInput()
    expected := getCurrentProductionOutput()

    actual := SystemUnderRefactoring(input)
    assert.Equal(t, expected, actual, "Behavior changed!")
}
```

### Step 4: Apply Refactoring Pattern
- Choose ONE pattern per iteration
- Apply systematically
- Run tests after each change

### Step 5: Verify and Measure
```bash
# Verify behavior unchanged
go test ./...

# Measure performance impact
go test -bench=. -benchmem > after_bench.txt
benchcmp before_bench.txt after_bench.txt

# Check test coverage improved
go test -cover ./...
```

---

## Code Smell Detection

### 1. Duplicate Code (DRY Violation)

**Detection:**
```go
// Look for patterns like this repeated across files
func (s *UserService) GetByID(ctx context.Context, id string) (*User, error) {
    var user User
    err := s.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
    if err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return &user, nil
}

func (s *ProductService) GetByID(ctx context.Context, id string) (*Product, error) {
    var product Product
    err := s.db.WithContext(ctx).Where("id = ?", id).First(&product).Error
    if err != nil {
        return nil, fmt.Errorf("failed to get product: %w", err)
    }
    return &product, nil
}
```

**Refactoring:**
```go
// Generic repository pattern
type Repository[T any] struct {
    db *gorm.DB
}

func (r *Repository[T]) GetByID(ctx context.Context, id string) (*T, error) {
    var entity T
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
    if err != nil {
        return nil, fmt.Errorf("failed to get entity: %w", err)
    }
    return &entity, nil
}
```

### 2. Long Functions

**Detection:** Functions > 50 lines
**Solution:** Extract Method pattern

```go
// BEFORE: 100+ line function
func ProcessOrder(order *Order) error {
    // Validate order
    if order.ID == "" {
        return errors.New("invalid ID")
    }
    if order.CustomerID == "" {
        return errors.New("invalid customer")
    }
    // ... 20 more validations

    // Calculate pricing
    subtotal := 0.0
    for _, item := range order.Items {
        subtotal += item.Price * float64(item.Quantity)
    }
    // ... 30 lines of pricing logic

    // Send notifications
    // ... 30 lines of notification code

    // Update inventory
    // ... 20 lines of inventory code

    return nil
}

// AFTER: Composed of small functions
func ProcessOrder(order *Order) error {
    if err := validateOrder(order); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    pricing, err := calculatePricing(order)
    if err != nil {
        return fmt.Errorf("pricing failed: %w", err)
    }

    if err := sendOrderNotifications(order, pricing); err != nil {
        return fmt.Errorf("notification failed: %w", err)
    }

    if err := updateInventory(order); err != nil {
        return fmt.Errorf("inventory update failed: %w", err)
    }

    return nil
}
```

### 3. Feature Envy

**Detection:** Method uses another object's data more than its own

```go
// WRONG: Service knows too much about repository internals
func (s *Service) ComplexQuery(filters map[string]interface{}) ([]Entity, error) {
    query := s.repo.db.Model(&Entity{})

    if val, ok := filters["name"]; ok {
        query = query.Where("name LIKE ?", "%"+val.(string)+"%")
    }
    if val, ok := filters["status"]; ok {
        query = query.Where("status = ?", val)
    }
    // Service is building repository's query

    var results []Entity
    return results, query.Find(&results).Error
}

// RIGHT: Move behavior to the right place
func (r *Repository) FindWithFilters(filters map[string]interface{}) ([]Entity, error) {
    query := r.buildQuery(filters)
    var results []Entity
    return results, query.Find(&results).Error
}
```

### 4. Primitive Obsession

```go
// WRONG: Using primitives for domain concepts
func TransferMoney(fromAccount string, toAccount string, amount float64, currency string) error {
    // Validation scattered everywhere
    if amount <= 0 {
        return errors.New("invalid amount")
    }
    if currency != "USD" && currency != "EUR" {
        return errors.New("unsupported currency")
    }
    // ...
}

// RIGHT: Domain types
type Money struct {
    Amount   decimal.Decimal
    Currency Currency
}

type AccountID string

func (a AccountID) Validate() error {
    if string(a) == "" {
        return errors.New("empty account ID")
    }
    return nil
}

func TransferMoney(from AccountID, to AccountID, money Money) error {
    if err := from.Validate(); err != nil {
        return err
    }
    if err := to.Validate(); err != nil {
        return err
    }
    if err := money.Validate(); err != nil {
        return err
    }
    // Domain logic with type safety
}
```

---

## Refactoring Patterns

### 1. Extract Function
**When:** Function > 50 lines or does multiple things
```go
// Extract cohesive blocks into named functions
// Each function should do ONE thing well
```

### 2. Extract Interface
**When:** Multiple implementations of similar behavior
```go
type PaymentProcessor interface {
    ProcessPayment(ctx context.Context, amount Money) error
}

// Now you can have StripeProcessor, PayPalProcessor, etc.
```

### 3. Replace Constructor with Factory
**When:** Complex object creation logic
```go
// BEFORE
service := &Service{
    repo:   repo,
    cache:  cache,
    logger: logger,
    // Complex initialization
}

// AFTER
service := NewServiceWithDefaults(repo)
// Factory handles complex setup
```

### 4. Replace Temp with Query
**When:** Temporary variable used only once
```go
// BEFORE
total := calculateTotal(items)
return total > 100

// AFTER
return calculateTotal(items) > 100
```

### 5. Introduce Parameter Object
**When:** Functions with > 3 parameters
```go
// BEFORE
func CreateUser(name, email, phone string, age int, address string) error

// AFTER
type UserRequest struct {
    Name    string
    Email   string
    Phone   string
    Age     int
    Address string
}

func CreateUser(req UserRequest) error
```

---

## Testing Strategy

### 1. Test Pyramid
```
        /\
       /  \     E2E Tests (5%)
      /----\
     /      \   Integration Tests (25%)
    /--------\
   /          \ Unit Tests (70%)
```

### 2. Test Naming Convention
```go
func TestServiceName_MethodName_StateUnderTest_ExpectedBehavior(t *testing.T) {
    // Example:
    // TestUserService_CreateUser_WithInvalidEmail_ReturnsError
}
```

### 3. Table-Driven Tests
```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    bool
        wantErr error
    }{
        {"valid email", "user@example.com", true, nil},
        {"invalid email", "not-an-email", false, ErrInvalidEmail},
        {"empty email", "", false, ErrRequired},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Validate(tt.input)
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
            if err != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 4. Test Helpers
```go
// Create test fixtures
func newTestService(t *testing.T) *Service {
    t.Helper() // Mark as helper

    db := newTestDB(t)
    t.Cleanup(func() {
        db.Close()
    })

    return &Service{db: db}
}
```

---

## Safety Checklist

### Before Refactoring
- [ ] Tests exist and pass for code being refactored
- [ ] Benchmark captured if performance-critical
- [ ] Git branch created for refactoring work
- [ ] Dependencies documented

### During Refactoring
- [ ] One pattern at a time
- [ ] Tests run after each change
- [ ] Commits are atomic and well-described
- [ ] No mixing of refactoring and feature work

### After Refactoring
- [ ] All tests pass
- [ ] Performance benchmarks comparable or better
- [ ] Documentation updated
- [ ] Code review requested
- [ ] No compiler warnings
- [ ] Linter warnings resolved

---

## Measurement Criteria

### Code Quality Metrics

#### Cyclomatic Complexity
```bash
# Use gocyclo
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
gocyclo -over 10 .
```
- **Target:** All functions ≤ 10
- **Acceptable:** ≤ 15 with justification
- **Refactor:** > 15 always

#### Test Coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```
- **Target:** ≥ 80% for critical paths
- **Minimum:** ≥ 70% overall
- **New code:** ≥ 90%

#### Code Duplication
```bash
# Use dupl
go install github.com/mibk/dupl@latest
dupl -t 50 .
```
- **Target:** < 5% duplication
- **Action:** Extract common patterns

#### Function Length
- **Target:** ≤ 50 lines
- **Maximum:** 100 lines with justification
- **Cognitive Load:** ≤ 7 ± 2 concepts

#### Dependencies
```bash
# Use godepgraph
go install github.com/kisielk/godepgraph@latest
godepgraph . | dot -Tpng -o deps.png
```
- **Target:** ≤ 5 direct dependencies per package
- **Coupling:** Low coupling, high cohesion

### Performance Metrics

```go
// Benchmark template
func BenchmarkFunction(b *testing.B) {
    // Setup
    data := prepareTestData()

    b.ResetTimer() // Don't count setup time

    for i := 0; i < b.N; i++ {
        // Code to benchmark
        Function(data)
    }
}
```

Track:
- Execution time (ns/op)
- Memory allocations (allocs/op)
- Memory usage (B/op)

---

## Specific Guidelines for This Codebase

### 1. CRUD Duplication
**Current Issue:** 15 services with identical CRUD patterns
**Solution:** Generic repository pattern with type parameters

### 2. Context Handling
**Current Issue:** Using context.Background() in GraphQL handlers
**Solution:** Always propagate request context

### 3. Error Handling
**Current Issue:** 971 instances of inconsistent error handling
**Solution:** Centralized error package with wrapping

### 4. Large Files
**Current Issue:** Files with 800+ LOC
**Solution:** Split by responsibility, max 300 LOC per file

### 5. Test Coverage
**Current Issue:** 4.4% coverage
**Solution:** Add tests before any refactoring

### Priority Order:
1. Add tests (safety net)
2. Fix critical bugs (context handling)
3. Remove duplication (biggest impact)
4. Improve structure (maintainability)
5. Optimize performance (if needed)

---

## Red Flags to Avoid

1. **Never refactor without tests**
2. **Never mix refactoring with features**
3. **Never change API contracts without versioning**
4. **Never optimize without measuring**
5. **Never ignore compiler warnings**
6. **Never leave code worse than you found it**
7. **Never commit commented-out code**
8. **Never use panic for error handling**
9. **Never ignore context cancellation**
10. **Never create global mutable state**

---

## Communication Template

### Commit Message Format
```
refactor(package): extract common repository pattern

- Eliminated 1,500 LOC of duplication across 15 services
- Created generic Repository[T] with common CRUD operations
- Maintained backward compatibility with existing interfaces
- Added comprehensive test coverage (from 0% to 85%)

Metrics:
- Before: 15 files, 1,500 LOC total
- After: 1 generic file, 150 LOC
- Test coverage: 85%
- Performance: No regression (benchmarks included)

Refs: #123
```

### Pull Request Template
```markdown
## What
Brief description of refactoring

## Why
- Problem being solved
- Metrics showing the issue

## How
- Refactoring pattern applied
- Step-by-step changes

## Testing
- [ ] Existing tests pass
- [ ] New tests added
- [ ] Benchmarks show no regression
- [ ] Manual testing completed

## Metrics
| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| LOC | 1,500 | 150 | -90% |
| Coverage | 0% | 85% | +85% |
| Complexity | 15 | 5 | -67% |

## Breaking Changes
- None / List any

## Rollback Plan
- How to revert if needed
```

---

## Resources

### Tools
- `gofmt` - Format code
- `golangci-lint` - Comprehensive linting
- `go-critic` - Advanced static analysis
- `gocyclo` - Cyclomatic complexity
- `dupl` - Duplicate detection
- `godepgraph` - Dependency visualization
- `go test -race` - Race condition detection

### Reading
- "Refactoring" by Martin Fowler
- "Clean Code" by Robert Martin
- "Effective Go" - golang.org
- "Go Code Review Comments" - GitHub wiki

### Principles
- SOLID (Single Responsibility, Open/Closed, Liskov, Interface Segregation, Dependency Inversion)
- DRY (Don't Repeat Yourself)
- KISS (Keep It Simple, Stupid)
- YAGNI (You Aren't Gonna Need It)
- Law of Demeter (Principle of Least Knowledge)

---

*Last Updated: 2024*
*Version: 1.0*
*Specific to: Kainuguru API Codebase*