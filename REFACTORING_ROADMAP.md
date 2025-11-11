# Kainuguru API - Comprehensive Refactoring Roadmap

## Overview

This roadmap outlines a structured approach to refactor the Kainuguru API codebase, addressing critical issues and improving maintainability. The effort is estimated at **3-4 weeks** for comprehensive cleanup.

---

## Phase 1: CRITICAL FIXES (Week 1 - MUST DO)

### 1.1 Consolidate Repository Directories
**Issue:** Duplicate `repositories/` and `repository/` directories causing confusion
**Impact:** HIGH - Architectural confusion
**Effort:** 4 hours

**Tasks:**
- [x] Audit both directories for actual implementations (only session/shopping/user repos lived under `internal/repository`)
- [x] Determine which is the "source of truth" (`internal/repositories` now owns all repository code)
- [x] Merge implementations into single directory (moved session/shopping/user repos into `internal/repositories`, removed legacy folder)
- [x] Update all import statements (no remaining references to `internal/repository`)
- [ ] Run full test suite *(blocked by pre-existing failures in `internal/services/ai/prompt_builder.go:144` and duplicate `main` funcs under `scripts/` — see testing section)*

**Files to Touch:**
- `internal/repositories/*` (consolidate)
- `internal/repository/*` (consolidate)
- Update ~20+ files with corrected imports

---

### 1.2 Fix Context Abuse in GraphQL Handler
**Issue:** Using `context.Background()` ignores request context, breaks timeouts/cancellation
**Impact:** CRITICAL - Production bug
**Effort:** 2 hours

**Current Code (handlers/graphql.go:68):**
```go
baseCtx := c.Context()
ctx := context.Background()  // ❌ WRONG
ctx = context.WithValue(ctx, middleware.UserContextKey, claims.UserID)
```

**Fixed Code:**
```go
ctx := c.Context()  // ✅ Preserve request context
ctx = context.WithValue(ctx, middleware.UserContextKey, claims.UserID)
ctx = context.WithValue(ctx, middleware.SessionContextKey, claims.SessionID)
ctx = context.WithValue(ctx, middleware.ClaimsContextKey, claims)
```

**Tasks:**
- [x] Preserve request context propagation in `internal/handlers/graphql.go`
- [x] Add request-cancellation regression test (`internal/handlers/graphql_test.go`)

---

### 1.3 Remove Placeholder Repository File
**Issue:** `placeholder_repos.go` (506 LOC) contains unimplemented stubs
**Impact:** MEDIUM - Dead code and confusion
**Effort:** 1 hour

**Tasks:**
- [x] Verify no code depends on placeholder repos (git grep shows placeholder structs unused outside the deleted file)
- [x] Delete `repositories/placeholder_repos.go`
- [x] Update factory.go to point to real implementations (`internal/repositories/factory.go` now returns concrete store/session/shopping/user repos)

---

### 1.4 Resolve All TODOs or Create Issues
**Issue:** 21+ unresolved TODOs scattered throughout codebase
**Impact:** MEDIUM - Technical debt tracker
**Effort:** 2 hours

**Action Items:**
- [x] Create GitHub issues for each TODO (documented in TODO triage doc)
- [x] Link issues to this roadmap (reference ISSUE_TRACKER.md, section Phase 1 TODOs)
- [x] Add labels: `deferred`, `nice-to-have`, `critical`
- [x] Remove inline TODO comments once issues created (only active TODOs with open issues remain)

**TODO issues created/still open:**
- Price alerts placeholder (auth.go:128) [`issue #123`]
- Shopping list categories (shopping_list.go:303) [`issue #124`]
- Pagination implementation (multiple files) [`issue #125`]
- JSON serialization in Redis (redis.go:202) [`issue #126`]
- User data loader (shopping_list.go:290) [`issue #127`]

---

### 1.5 Clean Up Stale Configuration Files
**Issue:** Stale `.env.bak` and `.env.bak2` files in repository
**Impact:** LOW - Cleanup
**Effort:** 0.5 hours

**Tasks:**
- [x] Delete `.env.bak` (confirmed absent via git ls-files; no longer tracked)
- [x] Delete `.env.bak2`
- [x] Add .env backups to .gitignore

---

## Phase 2: ARCHITECTURAL REFACTORING (Week 2 - HIGH PRIORITY)

### 2.1 Create Generic CRUD Repository Pattern
**Issue:** 15 services with identical GetByID/GetAll/Create/Update/Delete patterns (~1,500 LOC duplication)
**Impact:** HIGH - Maintainability nightmare
**Effort:** 16 hours

**Design:**
```go
// internal/repositories/base.go
type BaseRepository[T any] struct {
    db    *bun.DB
    table string
}

func (r *BaseRepository[T]) GetByID(ctx context.Context, id interface{}) (*T, error) {
    // Generic implementation
}

func (r *BaseRepository[T]) GetAll(ctx context.Context, opts ...QueryOption) ([]*T, error) {
    // Generic implementation with filters
}

// ... Create, Update, Delete
```

**Tasks:**
- [ ] Design generic repository interface
- [ ] Implement BaseRepository with Go 1.18+ generics
- [ ] Create QueryOption builder pattern
- [ ] Migrate 15 services to use base repository
- [ ] Add unit tests for generic repository
- [ ] Verify all tests still pass

**Services to Migrate:**
1. `product_master_service.go`
2. `shopping_list_item_service.go`
3. `flyer_service.go`
4. `flyer_page_service.go`
5. `store_service.go`
6. `product_service.go`
7. `shopping_list_service.go`
8. `price_history_service.go`
9. `extraction_job_service.go`
10. And remaining CRUD services...

**Expected Outcome:**
- Reduce ~1,500 LOC to ~150 LOC
- Single source of truth for CRUD operations
- Easier to add new services

---

### 2.2 Extract Pagination Helper
**Issue:** Pagination logic duplicated in 8+ GraphQL resolvers (~320 LOC)
**Impact:** MEDIUM-HIGH - Maintainability
**Effort:** 8 hours

**Design:**
```go
// internal/graphql/pagination/helper.go
type PaginationParams struct {
    First  *int
    After  *string
    Last   *int
    Before *string
}

type PageInfo struct {
    HasNextPage     bool
    HasPreviousPage bool
    StartCursor     *string
    EndCursor       *string
}

type Edge[T any] struct {
    Node   *T
    Cursor string
}

type Connection[T any] struct {
    Edges      []*Edge[T]
    PageInfo   *PageInfo
    TotalCount int
}

func NewPaginationHelper[T any](items []*T, params PaginationParams) (*Connection[T], error) {
    // Generic pagination implementation
}
```

**Tasks:**
- [ ] Create pagination helper package
- [ ] Implement cursor encoding/decoding
- [ ] Implement offset-based pagination
- [ ] Update all 8+ resolvers to use helper
- [ ] Add tests for pagination edge cases

**Resolvers to Update:**
- `Stores()`
- `Flyers()`
- `Products()`
- `ShoppingLists()`
- `PriceHistory()`
- And others...

**Expected Outcome:**
- Reduce ~320 LOC pagination logic to ~50 LOC
- Consistent pagination across all resolvers
- Easier to change pagination strategy

---

### 2.3 Consolidate Authentication Middleware
**Issue:** AuthMiddleware and OptionalAuthMiddleware are 95% identical (100+ LOC duplication)
**Impact:** MEDIUM - Maintainability
**Effort:** 4 hours

**Current Pattern:**
```go
func AuthMiddleware(...) fiber.Handler { }        // ~50 LOC
func OptionalAuthMiddleware(...) fiber.Handler { } // ~50 LOC
// 95% identical code
```

**Refactored Pattern:**
```go
type AuthMiddlewareConfig struct {
    Required bool
    Services *AuthServices
}

func AuthMiddleware(config AuthMiddlewareConfig) fiber.Handler {
    return func(c *fiber.Ctx) error {
        token := extractToken(c)
        if token == "" && config.Required {
            return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
                "error": "Unauthorized",
            })
        }
        if token == "" && !config.Required {
            return c.Next()
        }
        
        // Rest of validation...
    }
}
```

**Tasks:**
- [ ] Create AuthMiddlewareConfig struct
- [ ] Merge AuthMiddleware and OptionalAuthMiddleware
- [ ] Update server setup to use configuration
- [ ] Add tests for both required/optional paths

---

### 2.4 Create Error Package with Domain-Specific Types
**Issue:** 971 instances of `fmt.Errorf("failed to X: %w", err)` - no error strategy
**Impact:** MEDIUM - Error handling consistency
**Effort:** 6 hours

**Design:**
```go
// pkg/errors/errors.go
var (
    ErrNotFound         = errors.New("resource not found")
    ErrInvalidInput     = errors.New("invalid input")
    ErrDatabase         = errors.New("database error")
    ErrUnauthorized     = errors.New("unauthorized")
    ErrForbidden        = errors.New("forbidden")
    ErrConflict         = errors.New("conflict")
    ErrInternal         = errors.New("internal server error")
    ErrNotImplemented   = errors.New("not implemented")
)

// With context
type ContextError struct {
    Op  string // Operation
    Err error  // Underlying error
}

func (e *ContextError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Op, e.Err)
    }
    return e.Op
}

// Usage:
if err != nil {
    return nil, &ContextError{Op: "GetProduct", Err: err}
}
```

**Tasks:**
- [ ] Create error package with standard errors
- [ ] Create context error wrapper
- [ ] Update 30 critical service methods to use new errors
- [ ] Document error handling strategy
- [ ] Add error middleware for GraphQL/REST

**Services to Update (Priority Order):**
1. `product_master_service.go`
2. `shopping_list_item_service.go`
3. `auth/service.go`
4. Rest of critical services

---

## Phase 3: QUALITY IMPROVEMENTS (Week 3 - MEDIUM PRIORITY)

### 3.1 Add Unit Tests for Services (Target 70% Coverage)
**Issue:** Only 4.4% test coverage, most services untested
**Impact:** HIGH - Code reliability
**Effort:** 20 hours

**Strategy:**
- Use table-driven tests
- Create test fixtures and factories
- Test happy path + error cases
- Use mock repositories

**Services to Test:**
1. `product_master_service.go` - 15 tests
2. `shopping_list_item_service.go` - 12 tests
3. `flyer_service.go` - 10 tests
4. `store_service.go` - 8 tests
5. `product_service.go` - 10 tests
6. And remaining services...

**Files to Create:**
```
internal/services/product_master_service_test.go
internal/services/shopping_list_item_service_test.go
internal/services/flyer_service_test.go
... (one test file per service)
```

**Example Test Structure:**
```go
func TestProductMasterService_GetByID(t *testing.T) {
    tests := []struct {
        name    string
        id      int64
        wantErr bool
        setup   func(*mockRepo)
    }{
        {
            name:    "returns product when found",
            id:      1,
            wantErr: false,
            setup: func(m *mockRepo) {
                m.On("GetByID", mock.Anything, 1).Return(&Product{}, nil)
            },
        },
        {
            name:    "returns error when not found",
            id:      999,
            wantErr: true,
            setup: func(m *mockRepo) {
                m.On("GetByID", mock.Anything, 999).Return(nil, ErrNotFound)
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

---

### 3.2 Split Large Service Files
**Issue:** Services with 865+ LOC are hard to maintain
**Impact:** MEDIUM - Maintainability
**Effort:** 12 hours

**Files to Split:**

**1. product_master_service.go (865 LOC) → 3 files:**
- `product_master_query_service.go` - GetByID, GetAll, GetByName (~250 LOC)
- `product_master_mutation_service.go` - Create, Update, Delete (~200 LOC)
- `product_master_matching_service.go` - Matching logic (~200 LOC)
- Common interface stays at `product_master_service.go`

**2. shopping_list_item_service.go (771 LOC) → 2 files:**
- `shopping_list_item_query_service.go` - Get operations (~300 LOC)
- `shopping_list_item_mutation_service.go` - Create, Update, Delete (~300 LOC)

**3. flyer_service.go → Refactor if needed**

---

### 3.3 Add GraphQL Request Validation
**Issue:** No input validation on mutations, vulnerable to invalid data
**Impact:** MEDIUM - Security/correctness
**Effort:** 8 hours

**Tasks:**
- [ ] Create validation middleware for GraphQL inputs
- [ ] Add constraints to GraphQL schema
- [ ] Validate required fields
- [ ] Add regex patterns for string formats
- [ ] Test validation on sample mutations

---

### 3.4 Document Service Dependencies
**Issue:** Unclear service dependencies and creation flow
**Impact:** MEDIUM - Onboarding
**Effort:** 4 hours

**Tasks:**
- [ ] Create service dependency diagram
- [ ] Document service factory initialization
- [ ] Add package-level documentation
- [ ] Document data flow for key operations

---

## Phase 4: TECHNICAL POLISH (Week 4 - LOWER PRIORITY)

### 4.1 Configuration Validation
**Issue:** No validation of required environment variables, inconsistent naming
**Impact:** LOW-MEDIUM - Runtime safety
**Effort:** 4 hours

**Tasks:**
- [ ] Add validation to config.Load()
- [ ] Create config validation functions
- [ ] Standardize configuration naming (`MaxRetries` vs `RetryAttempts`)
- [ ] Document all required environment variables

---

### 4.2 Performance Optimization
**Issue:** Potential N+1 queries, no query complexity limits
**Impact:** LOW - Performance
**Effort:** 6 hours

**Tasks:**
- [ ] Profile queries to find N+1 issues
- [ ] Add GraphQL query complexity limits
- [ ] Optimize expensive queries
- [ ] Add query execution metrics

---

### 4.3 Documentation Improvements
**Issue:** Limited documentation (30% of ideal)
**Impact:** LOW-MEDIUM - Maintainability
**Effort:** 8 hours

**Tasks:**
- [ ] Add package-level documentation comments
- [ ] Document GraphQL resolvers
- [ ] Create architecture decision records (ADRs)
- [ ] Add examples for complex operations

---

### 4.4 Code Style Consistency
**Issue:** Inconsistent naming, formatting, patterns
**Impact:** LOW - Code quality
**Effort:** 4 hours

**Tasks:**
- [ ] Run golangci-lint and fix issues
- [ ] Standardize error variable naming
- [ ] Standardize receiver names (s, r, svc)
- [ ] Add pre-commit hooks

---

## Priority Summary

### CRITICAL (Do Immediately)
- [ ] Fix context abuse in GraphQL handler (breaks timeouts)
- [ ] Consolidate repository directories (architectural confusion)
- [ ] Remove placeholder repository file (dead code)

### HIGH (Do in Phase 1-2)
- [ ] Generic CRUD repository pattern (eliminates 1,500 LOC duplication)
- [ ] Add unit tests (4.4% is dangerously low)
- [ ] Error handling package (971 unwritten patterns)

### MEDIUM (Do in Phase 2-3)
- [ ] Extract pagination helper (320 LOC duplication)
- [ ] Split large service files (hard to maintain)
- [ ] Consolidate middleware (95% duplication)
- [ ] Resolve TODOs (clear technical debt tracking)

### LOW (Phase 4, Nice to Have)
- [ ] Configuration validation
- [ ] Performance optimization
- [ ] Documentation
- [ ] Code style consistency

---

## Success Metrics

After completing all phases, these metrics should improve:

| Metric | Current | Target | Status |
|--------|---------|--------|--------|
| Test Coverage | 4.4% | >70% | CRITICAL |
| Largest File | 865 LOC | <300 LOC | HIGH |
| Code Duplication | ~15% | <5% | HIGH |
| Number of Files | 147 | ~130 | MEDIUM |
| Average Function Length | ~60 LOC | <30 LOC | MEDIUM |
| TODO Comments | 21+ | 0 | MEDIUM |
| Documentation | ~30% | >80% | LOW |

---

## Timeline

```
Week 1: Critical Fixes
├── Days 1-2: Repository consolidation + Context fix
├── Days 3-4: Placeholder cleanup + TODO resolution
└── Day 5: Testing & verification

Week 2: Architecture Refactoring
├── Days 1-3: Generic CRUD repository
├── Days 4: Pagination helper
└── Day 5: Middleware consolidation

Week 3: Quality Improvements
├── Days 1-3: Unit tests (continuous)
├── Day 4: Split large files
└── Day 5: Request validation

Week 4: Polish
├── Days 1-2: Configuration validation
├── Day 3: Performance optimization
├── Day 4: Documentation
└── Day 5: Code style & QA

Total: 20 working days (~4 weeks)
```

---

## Risk Mitigation

**Risk 1: Breaking existing functionality**
- Mitigation: Run full test suite after each phase
- Mitigation: Use feature branches and PR reviews
- Mitigation: Have staging environment for testing

**Risk 2: Incomplete migration of patterns**
- Mitigation: Use code generation or scripting to verify all instances updated
- Mitigation: Run linters to catch inconsistencies
- Mitigation: Create pre-commit hooks to prevent regressions

**Risk 3: Team knowledge gaps**
- Mitigation: Document changes in ADRs
- Mitigation: Pair programming for complex refactoring
- Mitigation: Code review with junior developers

---

## Implementation Checklist

- [ ] Week 1: All critical fixes completed
- [ ] Week 1: All TODOs resolved or tracked
- [ ] Week 2: Generic CRUD repository implemented
- [ ] Week 2: Pagination helper complete
- [ ] Week 2: Middleware consolidated
- [ ] Week 3: Unit tests for 70% coverage
- [ ] Week 3: Large files split
- [ ] Week 4: Documentation complete
- [ ] Final: All metrics meet targets
- [ ] Final: All tests pass
- [ ] Final: Code review approved
- [ ] Final: Merge to main branch

---

## Next Steps

1. **Immediately (This Week):**
   - [ ] Review this roadmap with team
   - [ ] Create GitHub issues for all tasks
   - [ ] Assign team members to work items
   - [ ] Set up feature branches

2. **This Sprint:**
   - [ ] Complete all CRITICAL fixes
   - [ ] Begin HIGH priority work
   - [ ] Measure baseline metrics

3. **Next Sprint:**
   - [ ] Complete Phase 1 & 2
   - [ ] Measure improvement
   - [ ] Plan Phase 3

---

## Questions & Discussions

**For Discussion:**
1. Should repository consolidation merge into `repositories/` or create new `repository_impl/`?
2. For generic CRUD, should we use generics or code generation?
3. Should test coverage target be 70% or higher?
4. Timeline feasibility - is 4 weeks realistic with current team size?

---

**Last Updated:** 2024-11-10
**Status:** Draft - Awaiting Team Review
**Owner:** Engineering Lead
**Reviewers:** Team Leads, Architects
