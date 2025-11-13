# AI Agent Refactoring Instructions for Kainuguru API

## Quick Reference for AI Agents

When refactoring this Go codebase, follow these strict instructions to ensure consistent, safe, and effective improvements.

---

## CRITICAL RULES (NEVER VIOLATE)

1. **NEVER refactor without tests** - If tests don't exist, write them FIRST
2. **NEVER mix refactoring with features** - One PR = One concern
3. **NEVER use context.Background()** in handlers - Always propagate request context
4. **NEVER commit without running tests** - `go test ./...` must pass
5. **NEVER leave code worse** - Every touch must improve the code

---

## AGENT WORKFLOW

### Step 1: Assessment Phase
```bash
# Run these commands first
go test ./... 2>&1 | grep -c "FAIL"  # Check failing tests
go test -cover ./... 2>&1 | grep "coverage"  # Check coverage
gocyclo -over 10 . 2>/dev/null | wc -l  # Count complex functions
find . -name "*.go" -exec wc -l {} + | awk '$1 > 500'  # Find large files
grep -r "TODO" --include="*.go" | wc -l  # Count TODOs
```

### Step 2: Prioritization Matrix

| Priority | Issue | Action | Time |
|----------|-------|--------|------|
| P0 | Production bugs | Fix immediately | < 2h |
| P1 | No tests | Write tests before ANY changes | 4-8h |
| P2 | Duplication > 100 LOC | Extract common pattern | 2-4h |
| P3 | File > 500 LOC | Split into multiple files | 2-3h |
| P4 | Function > 50 LOC | Extract smaller functions | 1h |

### Step 3: Refactoring Execution

#### For EVERY refactoring:
1. **Create branch**: `git checkout -b refactor/<specific-improvement>`
2. **Write characterization test**: Capture current behavior
3. **Make ONE change**: Single responsibility per commit
4. **Verify tests**: `go test ./...` must pass
5. **Check performance**: No degradation allowed
6. **Commit with metrics**: Include LOC reduced, coverage increased

---

## PATTERN RECOGNITION & FIXES

### Pattern 1: CRUD Duplication
**Detect:**
```go
func (s *XService) GetByID(ctx context.Context, id string) (*X, error) {
    // Nearly identical across services
}
```

**Fix:**
```go
// Create generic repository
type Repository[T any] struct {
    db *gorm.DB
}

func NewRepository[T any](db *gorm.DB) *Repository[T] {
    return &Repository[T]{db: db}
}

func (r *Repository[T]) GetByID(ctx context.Context, id string) (*T, error) {
    var entity T
    err := r.db.WithContext(ctx).Where("id = ?", id).First(&entity).Error
    return &entity, err
}
```

### Pattern 2: Context Abuse
**Detect:**
```go
ctx := context.Background() // WRONG in request handlers
```

**Fix:**
```go
// Always propagate from request
func (h *Handler) Handle(ctx context.Context) error {
    return h.service.Process(ctx) // Pass context down
}
```

### Pattern 3: Error Handling Chaos
**Detect:**
```go
if err != nil {
    return nil, err  // No context
}
```

**Fix:**
```go
if err != nil {
    return nil, fmt.Errorf("failed to fetch user id=%s: %w", id, err)
}
```

### Pattern 4: Large Functions
**Detect:** Function > 50 lines

**Fix:** Extract Method
```go
// BEFORE: 100 line function
func ProcessOrder(order *Order) error {
    // validation logic (30 lines)
    // pricing logic (30 lines)
    // notification logic (20 lines)
    // inventory logic (20 lines)
}

// AFTER: Composed functions
func ProcessOrder(order *Order) error {
    if err := validateOrder(order); err != nil {
        return fmt.Errorf("validation: %w", err)
    }

    pricing, err := calculatePricing(order)
    if err != nil {
        return fmt.Errorf("pricing: %w", err)
    }

    if err := notifyOrder(order, pricing); err != nil {
        return fmt.Errorf("notification: %w", err)
    }

    return updateInventory(order)
}
```

### Pattern 5: Missing Abstractions
**Detect:** Repeated inline logic

**Fix:** Extract to domain type
```go
// BEFORE: Validation everywhere
if email == "" || !strings.Contains(email, "@") {
    return errors.New("invalid email")
}

// AFTER: Domain type
type Email string

func (e Email) Validate() error {
    if string(e) == "" || !strings.Contains(string(e), "@") {
        return errors.New("invalid email")
    }
    return nil
}
```

---

## CODEBASE-SPECIFIC ISSUES

### Issue 1: Duplicate Repository Directories
- **Location:** `/repositories/` and `/repository/`
- **Fix:** Consolidate to `/repository/`, update all imports

### Issue 2: Placeholder Service
- **Location:** `services/placeholder_service.go` (506 LOC)
- **Fix:** DELETE entirely - it's all "not implemented"

### Issue 3: Large Service Files
- **Files:**
  - `product_service.go` (865 LOC)
  - `enrichment_service.go` (720 LOC)
- **Fix:** Split by domain:
  - `product_crud_service.go`
  - `product_enrichment_service.go`
  - `product_validation_service.go`

### Issue 4: GraphQL Context Bug
- **Location:** `graph/resolver.go:32`
- **Current:** `ctx := context.Background()`
- **Fix:** Use request context from resolver

### Issue 5: No Error Package
- **Issue:** 971 instances of ad-hoc error handling
- **Fix:** Create `pkg/errors/errors.go` with standard types

---

## TESTING REQUIREMENTS

### Minimum Coverage Targets
- **New code:** 90%
- **Modified code:** 80%
- **Critical paths:** 95%
- **Overall goal:** 70%

### Test Template
```go
func TestServiceName_MethodName_Scenario_Expected(t *testing.T) {
    // Arrange
    service := setupTestService(t)
    input := createTestInput()

    // Act
    result, err := service.Method(context.Background(), input)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### Table-Driven Tests (Preferred)
```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name    string
        input   interface{}
        wantErr bool
    }{
        {"valid input", validData, false},
        {"nil input", nil, true},
        {"empty input", emptyData, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := Validate(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

---

## METRICS TO TRACK

### Before Starting
```bash
# Save baseline metrics
echo "=== Baseline Metrics ===" > metrics.txt
echo "Date: $(date)" >> metrics.txt
echo "LOC: $(find . -name '*.go' | xargs wc -l | tail -1)" >> metrics.txt
echo "Files: $(find . -name '*.go' | wc -l)" >> metrics.txt
echo "Coverage: $(go test -cover ./... 2>&1 | grep coverage)" >> metrics.txt
echo "TODOs: $(grep -r 'TODO' --include='*.go' | wc -l)" >> metrics.txt
```

### After Each Refactoring
Report these metrics in commit message:
- **LOC changed:** Before/After
- **Test coverage:** Before/After %
- **Complexity:** Functions over 10 reduced by X
- **Performance:** No regression (show benchmark)
- **Duplication:** Removed X lines

---

## COMMIT MESSAGE FORMAT

```
refactor(package): [action] [what]

Problem:
- Describe issue with metrics

Solution:
- What pattern/technique applied
- Specific improvements made

Impact:
- LOC: 500 -> 200 (-60%)
- Coverage: 0% -> 85%
- Complexity: 15 -> 5
- Performance: No regression

Tests:
- Added X unit tests
- All existing tests pass
```

### Example:
```
refactor(repository): extract generic CRUD operations

Problem:
- 15 services duplicating same CRUD pattern
- 1,500 LOC of duplication
- 0% test coverage on repositories

Solution:
- Created generic Repository[T] with type parameters
- Extracted common CRUD operations
- Services now compose generic repository

Impact:
- LOC: 1,500 -> 150 (-90%)
- Coverage: 0% -> 85%
- Duplication: Eliminated
- Performance: Identical (benchmarks included)

Tests:
- Added 25 unit tests for generic repository
- All 147 existing tests still pass
```

---

## SAFETY CHECKLIST

Before ANY refactoring:
- [ ] Current tests pass (`go test ./...`)
- [ ] Benchmark captured if performance-critical
- [ ] Branch created (`refactor/specific-improvement`)

During refactoring:
- [ ] ONE pattern at a time
- [ ] Tests pass after EACH change
- [ ] No feature additions mixed in
- [ ] Commits are atomic

After refactoring:
- [ ] All tests pass
- [ ] Coverage increased or maintained
- [ ] No performance regression
- [ ] No new TODOs added
- [ ] Documentation updated if needed

---

## COMMON MISTAKES TO AVOID

1. **Refactoring without tests** → Write tests FIRST
2. **Big bang refactoring** → Small incremental changes
3. **Mixing concerns** → One refactoring type per PR
4. **Ignoring context** → Always propagate context
5. **Premature abstraction** → Wait for 3+ duplications
6. **Over-engineering** → YAGNI principle
7. **Breaking APIs** → Maintain compatibility
8. **Ignoring errors** → Always handle or propagate
9. **Global state** → Inject dependencies
10. **Forgetting cleanup** → Use defer for resources

---

## QUICK WINS FOR THIS CODEBASE

### Immediate (< 1 hour each):
1. Delete `services/placeholder_service.go` (506 LOC of dead code)
2. Fix context.Background() in `graph/resolver.go:32`
3. Consolidate repository directories

### High Impact (2-4 hours):
1. Generic CRUD repository (saves 1,500 LOC)
2. Pagination helper (saves 320 LOC)
3. Error handling package (standardizes 971 instances)

### Test Coverage Boost (4-8 hours):
1. Repository package tests (currently 0%)
2. Service layer tests (currently < 5%)
3. GraphQL resolver tests (currently minimal)

---

## TOOLS TO USE

```bash
# Install essential tools
go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
go install github.com/mibk/dupl@latest
go install github.com/kisielk/godepgraph@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run before refactoring
gocyclo -over 10 .          # Find complex functions
dupl -t 50 .                 # Find duplication
golangci-lint run            # Find issues
go test -race ./...          # Check for races
```

---

## ESCALATION TRIGGERS

Stop and ask for human review if:
1. Change affects > 500 LOC
2. API contract needs to change
3. Database schema modification required
4. Performance degrades > 10%
5. Security implications unclear
6. Architectural pattern change needed

---

## SUCCESS CRITERIA

A refactoring is successful when:
- ✅ All tests pass
- ✅ Coverage increased or maintained
- ✅ Code is simpler/clearer
- ✅ No performance regression
- ✅ Metrics show improvement
- ✅ No new technical debt introduced
- ✅ Documentation is current
- ✅ Team can understand changes

---

*Agent Version: 1.0*
*Codebase: Kainuguru API*
*Last Updated: 2024*

**Remember: Small, safe, incremental improvements. Always.**