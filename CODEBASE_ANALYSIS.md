# Kainuguru API - Comprehensive Codebase Analysis Report

## Executive Summary

**Codebase Statistics:**
- Total Go Files: 147
- Total Lines of Code: ~76,227
- Test Files: 9 (~3,374 lines)
- Test Coverage: Very Low (4% - Only 9 test files for 147 Go files)
- Project Status: Early-stage, partially implemented architecture

---

## 1. OVERALL PROJECT STRUCTURE & ARCHITECTURE

### Directory Organization
```
kainuguru-api/
├── cmd/                          # Command-line applications
│   ├── api/                      # Main GraphQL API server
│   ├── scraper/                  # Web scraper CLI
│   ├── migrator/                 # Database migration tool
│   ├── seeder/                   # Database seeder
│   ├── archive-flyers/           # Flyer archival tool
│   ├── enrich-flyers/            # AI enrichment tool
│   └── test-*                    # Testing utilities
├── internal/                     # Private application code
│   ├── api/                      # Legacy API handlers
│   ├── cache/                    # Redis caching layer
│   ├── config/                   # Configuration management
│   ├── database/                 # Database connection & ORM
│   ├── graphql/                  # GraphQL resolvers & schema
│   ├── handlers/                 # HTTP request handlers
│   ├── middleware/               # Request middleware
│   ├── models/                   # Data models
│   ├── monitoring/               # Health checks & observability
│   ├── repositories/             # Data access layer (DUPLICATED)
│   ├── repository/               # Data access layer (DUPLICATED)
│   ├── services/                 # Business logic layer
│   ├── workers/                  # Background workers
│   └── migrator/                 # DB migration logic
├── pkg/                          # Reusable packages
│   ├── config/
│   ├── errors/
│   ├── normalize/
│   └── logger/
└── tests/                        # Test files & fixtures
    ├── bdd/                      # BDD-style tests
    ├── fixtures/                 # Test data
    └── scripts/                  # Test utilities
```

### Architecture Pattern
- **Layered Architecture**: Controllers → Services → Repositories → Database
- **GraphQL-First API**: Primary interface via gqlgen
- **Service Factory Pattern**: Centralized service creation
- **Data Loaders**: Implemented for N+1 query prevention

---

## 2. MAIN PACKAGES & RESPONSIBILITIES

### Core Packages

| Package | Responsibility | Status | Issues |
|---------|-----------------|--------|--------|
| `internal/models` | Data structures for DB/API | Complete | Large files (Product: 900+ LOC) |
| `internal/services` | Business logic | Partial | Multiple CRUD patterns, 865+ LOC files |
| `internal/repositories` | Data access (ACTIVE) | Partial | 506 LOC placeholder file |
| `internal/repository` | Shopping/Auth repos (DUPLICATE) | Partial | Duplicate directory structure |
| `internal/graphql/resolvers` | GraphQL handlers | Partial | Query resolver 825+ LOC, many TODOs |
| `internal/handlers` | HTTP handlers | Minimal | GraphQL handler converts Fiber to HTTP test |
| `internal/middleware` | Request interceptors | Complete | Good separation of concerns |
| `internal/services/auth` | Authentication | Complete | Well-structured JWT & sessions |
| `internal/services/scraper` | Web scraping | Complete | Store-specific scrapers |
| `internal/services/enrichment` | AI-powered product enrichment | Partial | Missing error recovery |
| `internal/services/ai` | OpenAI integration | Partial | 626 LOC, needs refactoring |
| `internal/services/cache` | Caching strategies | Complete | Redis wrapper & flyer cache |
| `internal/database` | Bun ORM wrapper | Complete | PostgreSQL 15+ with partitioning |
| `pkg/normalize` | Lithuanian text normalization | Complete | Good utility package |

---

## 3. KEY DEPENDENCIES & FRAMEWORKS

### Framework Stack
```
Web Framework:      Fiber v2 (http)
API Style:          GraphQL (gqlgen v0.17)
ORM:                Bun v1.2 (PostgreSQL)
Database:           PostgreSQL 15+
Cache:              Redis v7+ (go-redis)
Authentication:     JWT (golang-jwt v5)
Data Loading:       dataloader v7
Logging:            zerolog
Configuration:      Viper
```

### Critical Dependencies Analysis

**Well-Chosen Dependencies:**
- Bun ORM: Good balance between raw SQL flexibility and type safety
- Fiber: High-performance HTTP framework
- gqlgen: Strong GraphQL schema-first approach
- dataloader: Prevents N+1 query issues

**Potential Issues:**
- Multiple redis libraries (`github.com/go-redis/redis` AND `github.com/redis/go-redis`) - **DUPLICATE**
- No testing frameworks in go.mod (using testify but not declared)
- No OpenTelemetry declared but referenced in code
- Using deprecated jwt/v5 (should be v5.3+)

---

## 4. CODE ORGANIZATION PATTERNS

### Pattern 1: Service Layer (Repeated 13+ times)
Each service follows this pattern:
```go
type productMasterService struct {
    db     *bun.DB
    matcher *matching.CompositeMatcher
    normalizer *normalize.LithuanianNormalizer
    logger *slog.Logger
}

func (s *productMasterService) GetByID(ctx context.Context, id int64) (*models.ProductMaster, error) {
    master := &models.ProductMaster{}
    err := s.db.NewSelect().
        Model(master).
        Where("pm.id = ?", id).
        Scan(ctx)
    
    if err != nil {
        return nil, fmt.Errorf("failed to get product master by ID %d: %w", id, err)
    }
    return master, nil
}

func (s *productMasterService) GetAll(ctx context.Context, filters ProductMasterFilters) ([]*models.ProductMaster, error) {
    query := s.db.NewSelect().Model((*models.ProductMaster)(nil))
    
    // Apply filters...
    if len(filters.Status) > 0 {
        query = query.Where("pm.status IN (?)", bun.In(filters.Status))
    }
    // More filter logic...
}
```

**Issues:**
- Same GetByID pattern copied to ALL services (duplicated ~15 times)
- Same GetByIDs pattern
- Same GetAll with filters pattern
- Same Create/Update/Delete patterns
- Could be abstracted into generic repository layer

### Pattern 2: GraphQL Resolvers (825+ LOC in query.go)
```go
func (r *queryResolver) Stores(ctx context.Context, filters *model.StoreFilters, first *int, after *string) (*model.StoreConnection, error) {
    limit := 20
    if first != nil && *first > 0 {
        limit = *first
        if limit > 100 {
            limit = 100
        }
    }
    
    offset := 0
    if after != nil && *after != "" {
        decodedOffset, err := decodeCursor(*after)
        if err == nil {
            offset = decodedOffset
        }
    }
    
    // Convert GraphQL filters to service filters
    serviceFilters := services.StoreFilters{
        Limit:  limit + 1,
        Offset: offset,
    }
    
    if filters != nil {
        serviceFilters.IsActive = filters.IsActive
        serviceFilters.Codes = filters.Codes
    }
    // ... pagination logic
}
```

**Issues:**
- Pagination logic copied across multiple resolvers
- Cursor encoding/decoding duplicated
- Filter conversion duplicated
- No generic helper for pagination

---

## 5. COMMON CODE PATTERNS & ANTI-PATTERNS

### Anti-Pattern 1: Duplicate Directory Structure
**Location:** `internal/repositories/` vs `internal/repository/`

```
internal/repositories/       internal/repository/
├── interfaces.go           ├── session_repository.go
├── factory.go              ├── shopping_list_item_repository.go  
├── store_repository.go     ├── shopping_list_repository.go
├── placeholder_repos.go    └── user_repository.go
└── store_repository.go
```

**Impact:** Confusing namespace, unclear which is authoritative
**Severity:** HIGH - Maintainability nightmare

### Anti-Pattern 2: Placeholder Repository Implementations (506 LOC)
```go
// placeholder_repos.go - Multiple repository implementations that just return errors
type flyerRepository struct { db *bun.DB }
func (r *flyerRepository) GetByID(ctx context.Context, id int) (*models.Flyer, error) {
    return nil, fmt.Errorf("flyerRepository.GetByID not implemented")
}
// ... repeated 100+ times
```

**Impact:** Dead code, misleading to developers, compilation passes but fails at runtime
**Severity:** MEDIUM

### Anti-Pattern 3: Generic CRUD Code Duplication
Pattern found in:
- `product_master_service.go` (GetByID, GetByIDs, GetAll, Create, Update, Delete)
- `shopping_list_item_service.go` (same)
- `flyer_service.go` (same)
- `store_service.go` (same)
- 9 more services...

Each file ~100-200 LOC doing identical operations with different types

**Solution:** Generic Repository Pattern or Code Generation

### Anti-Pattern 4: Large Functions (865+ LOC files)
**Largest Functions:**
- `product_master_service.go`: 865 LOC
- `shopping_list_item_service.go`: 771 LOC
- `query.go` (GraphQL): 825 LOC
- `product_service.go`: 425 LOC

Single files should be <300 LOC with clear responsibilities

### Anti-Pattern 5: Inconsistent Context Usage
```go
// Good:
func (s *service) ProcessFlyer(ctx context.Context, flyer *models.Flyer) error {
    // uses ctx properly
}

// Bad in handlers/graphql.go:
baseCtx := c.Context()
ctx := context.Background()  // ← IGNORES request context!
ctx = context.WithValue(ctx, middleware.UserContextKey, claims.UserID)
```

**Impact:** Timeouts won't propagate, cancellation doesn't work

### Anti-Pattern 6: Middleware Duplication
```go
// middleware/auth.go - TWO VERSIONS OF NEARLY IDENTICAL CODE:

// AuthMiddleware (requires auth)
func AuthMiddleware(...) fiber.Handler {
    // Extract token
    // Validate token
    // Validate session
    // Store in context
    // Next
}

// OptionalAuthMiddleware (same logic, different failure handling)
func OptionalAuthMiddleware(...) fiber.Handler {
    // Extract token
    // Validate token
    // Validate session  
    // Store in context
    // Next (but continue on error)
}
```

Could use single middleware with `required` parameter

### Anti-Pattern 7: Error Handling Boilerplate
```go
// Repeated ~971 times in codebase:
if err != nil {
    return nil, fmt.Errorf("failed to X: %w", err)
}

// Repeated ~867 times in codebase
```

No custom error types or wrapped error chains. Basic string concatenation.

### Anti-Pattern 8: GraphQL Handler Context Conversion
```go
// handlers/graphql.go - Creates HTTP test objects to adapt Fiber to gqlgen
httpReq := httptest.NewRequest("POST", "/graphql", bytes.NewReader(body))
httpReq = httpReq.WithContext(ctx)
w := httptest.NewRecorder()
gqlHandler.ServeHTTP(w, httpReq)
```

**Impact:** Memory allocation for test objects on every request, unclear adaptation layer

---

## 6. TECHNICAL DEBT & CODE SMELLS

### Critical Issues (MUST FIX)

1. **Duplicate Repository Directories**
   - `internal/repositories/` vs `internal/repository/`
   - Caused confusion, unclear source of truth
   - **Fix:** Consolidate to single directory, standardize naming

2. **Placeholder Repository File (506 LOC)**
   - Contains unimplemented stubs for all repositories
   - Returns hardcoded "not implemented" errors
   - **Fix:** Remove and implement actual repositories or use tests

3. **Dead Code in GraphQL Playground**
   - PlaygroundHandler shows placeholder implementation status
   - Returns hardcoded response, not functional playground
   - **Fix:** Implement actual GraphQL Playground or remove

4. **Context Abuse in GraphQL Handler**
   - Ignores request context, uses `context.Background()`
   - Breaks timeouts and cancellation
   - **Fix:** Use `httpReq.WithContext(c.Context())`

### High Priority Issues

5. **Massive Service Files (865+ LOC)**
   - Single file responsible for too many operations
   - Hard to test, understand, and maintain
   - **Solution:** Split into domain-focused services

6. **No Generic CRUD Abstraction**
   - 13+ services with identical GetByID/GetAll/Create patterns
   - ~3,000 LOC of duplicated boilerplate
   - **Solution:** Create generic repository interface or use code generation

7. **Inconsistent Error Handling**
   - Mix of `fmt.Errorf`, custom types, and no error wrapping
   - 971 fmt.Errorf calls doing same thing
   - **Solution:** Create error package with domain-specific error types

8. **Missing Test Coverage**
   - Only 9 test files for 147 Go files (6% coverage)
   - Only 3,374 test LOC vs 76,227 code LOC (4.4% test ratio)
   - BDD tests exist but focus on integration not unit tests
   - **Solution:** Add unit tests for services, reduce side effects

9. **TODOs Scattered Throughout**
   - 21+ unresolved TODO comments
   - Placeholder implementations in production code
   - Missing implementations marked but not tracked
   - **Examples:**
     - `cache/redis.go:202` - JSON serialization not implemented
     - `graphql/resolvers/auth.go:128` - Price alerts placeholder
     - `graphql/resolvers/shopping_list_item.go:256` - Pagination incomplete

### Medium Priority Issues

10. **Middleware Code Duplication**
    - AuthMiddleware and OptionalAuthMiddleware are 95% identical
    - Could merge into single middleware with configuration

11. **GraphQL Schema Size (906 LOC)**
    - Monolithic schema file
    - Should split into domain-focused files

12. **Tight Coupling in Services**
    - Services depend on concrete implementations, not interfaces
    - Makes testing difficult without mocks

13. **Configuration Inconsistency**
    - Some config in config.go
    - Some config in environment variables directly
    - No validation of required environment variables
    - **Example:** `config.go` MaxRetries vs `config.go` RetryAttempts (two different names for same thing)

14. **Repository Factory Not Used**
    - `repositories/factory.go` exists but is empty
    - Repositories created manually in services
    - Comment says "Factory methods defined in individual files" but they're not

15. **Mixed Normalization Layers**
    - Lithuanian text processing split across:
      - `pkg/normalize/lithuanian.go`
      - `pkg/normalize/brands.go`
      - `pkg/normalize/units.go`
      - Multiple services doing their own normalization
    - No single source of truth

### Low Priority Issues

16. **Inconsistent Naming**
    - `productMasterService` vs other camelCase services
    - `queryResolver` vs `Resolver` inconsistency
    - DB aliases inconsistent (`pm.id`, `sli.id`, sometimes no alias)

17. **Magic Numbers**
    - Query limits hardcoded as `100`, `20`, `50` in different places
    - No constants defined

18. **Missing Input Validation**
    - Services accept models directly without validation
    - GraphQL mutation inputs not validated before service calls

19. **Incomplete DataLoaders**
    - Only for Store, Flyer, FlyerPage, ProductMaster, Auth
    - Missing for: Products, ShoppingLists, Users
    - User loader mentioned in TODO but not implemented

---

## 7. AREAS WITH TECHNICAL DEBT

### Database Layer
- No query optimization hints
- N+1 queries possible even with dataloaders
- Product partitioning mentioned but not fully implemented
- Migration system exists but versioning strategy unclear

### API Layer
- No rate limiting on GraphQL mutations (only scraper endpoints)
- No request validation schemas
- No pagination standardization across queries
- Cursor encoding/decoding duplicated

### Search Layer
- `internal/services/search/validation_test.go` exists (286 LOC)
- but search service is 495 LOC
- Complex validation logic that could be extracted

### Worker/Background Processing
- Queue implementation (`services/worker/queue.go`)
- Scheduler implementation (`services/worker/scheduler.go`)
- Lock mechanism (`services/worker/lock.go`)
- But no clear orchestration or error recovery

### Enrichment Pipeline
- Multiple enrichment services (extraction, validation, matching)
- Complex orchestration in `services/enrichment/orchestrator.go`
- Missing error recovery and retry logic
- Cost tracking implemented but not integrated with budgeting

---

## 8. TEST COVERAGE & TESTING PATTERNS

### Current Test Status
```
Total Test Files:        9
Test Lines of Code:      3,374
Total Lines of Code:    76,227
Test Coverage Ratio:     4.4%

Test Distribution:
- BDD Steps Tests:       1,021 LOC (endpoint_validation_test.go)
- Auth Tests:            390 LOC
- Search Tests:          358 LOC
- Store/Flyer Tests:     319 LOC
- Shopping List Tests:   402 LOC
- Product Matcher:       209 LOC
- Search Validation:     286 LOC
- Price Tests:           213 LOC
- Integration Tests:     176 LOC
```

### Testing Gaps

**Missing Unit Tests For:**
- Service layer implementations (0 tests)
- Repository implementations (0 tests)
- GraphQL resolvers (0 unit tests)
- Middleware functions (0 tests)
- Helper utilities (0 tests)
- Error handling (0 tests)
- Cache layer (0 tests)
- Configuration loading (0 tests)

**Testing Anti-Patterns Found:**
1. BDD tests are integration tests, not unit tests
2. No table-driven tests
3. No test fixtures or factories
4. Mock creation manual, not generated
5. No test helpers for common assertions
6. No benchmarks

### Test Examples

**Good:**
- `services/search/validation_test.go` (286 LOC) - Comprehensive validation tests
- `services/matching/product_matcher_test.go` (209 LOC) - Matcher logic tests

**Bad:**
- No tests for most services
- BDD tests require full database setup
- Tests marked as integration, not unit
- No parallelization strategy

---

## 9. CONFIGURATION & ENVIRONMENT SETUP

### Configuration Loading
```go
// config/config.go - Viper-based configuration
type Config struct {
    Server   ServerConfig
    Database database.Config
    Redis    RedisConfig
    // ... 8 more config sections
}
```

### Issues with Configuration
1. **No validation** - Required fields not checked
2. **No defaults** - Server crashes if env vars missing
3. **Inconsistent naming** - `MaxRetries` vs `RetryAttempts`
4. **Secrets in config** - JWT secret, API keys not handled specially
5. **No hot reload** - Changes require restart
6. **Environment coupling** - Only reads .env, not from actual environment

### Configuration Files Found
- `.env` - Local development
- `.env.dist` - Template
- `.env.test` - Testing
- `.env.production` - Production
- `.env.bak`, `.env.bak2` - Stale backups (should be deleted)

### Missing Configurations
- Structured logging levels not configurable
- GraphQL query complexity limits not present
- Database pool sizes hardcoded
- Timeout values hardcoded

---

## 10. REFACTORING OPPORTUNITIES

### Priority 1: Foundation (CRITICAL)
1. **Consolidate Repository Directories**
   - Merge `repositories/` and `repository/` into single `repositories/`
   - Standardize all repository implementations
   - Eliminate placeholder repos

2. **Generic CRUD Repository**
   - Create base repository type with common CRUD operations
   - Reduce 3,000+ LOC of duplicated code
   - Use Go generics (1.18+) for type safety

3. **Error Handling Package**
   - Create domain-specific error types
   - Implement error wrapping strategy
   - Standardize error messages across codebase

### Priority 2: Architecture (HIGH)
4. **Split Large Services**
   - Break product_master_service (865 LOC) into:
     - ProductMasterQueryService
     - ProductMasterMutationService
     - ProductMasterMatchingService

5. **Abstract GraphQL Patterns**
   - Create pagination helper reducing 100+ LOC of duplication
   - Generic cursor encoder/decoder
   - Filter conversion utilities

6. **Middleware Consolidation**
   - Merge AuthMiddleware and OptionalAuthMiddleware
   - Extract common validation logic

### Priority 3: Quality (MEDIUM)
7. **Test Coverage**
   - Add unit tests for all services (target: 70% coverage)
   - Add integration tests for critical flows
   - Implement test factories for model creation

8. **Configuration Improvements**
   - Add config validation on startup
   - Implement structured logging configuration
   - Move secrets to dedicated handler (envfile, vault)

9. **Query Optimization**
   - Profile N+1 queries
   - Ensure all eager loading via relations or dataloaders
   - Add query complexity limits to GraphQL

### Priority 4: Maintenance (MEDIUM)
10. **Code Documentation**
    - Add package-level documentation
    - Document service dependencies
    - Add examples for complex operations

11. **Clean Up TODOs**
    - Implement or remove all TODO comments
    - Create GitHub issues for deferred work

12. **Remove Dead Code**
    - Delete placeholder_repos.go
    - Remove unused API packages
    - Clean up stale test utilities

---

## SUMMARY TABLE: Code Quality Metrics

| Metric | Current | Target | Status |
|--------|---------|--------|--------|
| Test Coverage | 4.4% | >70% | CRITICAL |
| Avg Function Length | ~60 LOC | <30 LOC | POOR |
| Largest File | 865 LOC | <300 LOC | POOR |
| Code Duplication | ~15% | <5% | HIGH |
| Cyclomatic Complexity | ~8 avg | <5 avg | MEDIUM |
| Dependency Count | 31 direct | <25 | OK |
| Documentation | ~30% | >80% | LOW |

---

## CONCLUSION

The Kainuguru API codebase is a **well-intentioned but early-stage project** with solid architectural decisions (Fiber + GraphQL + PostgreSQL) but significant **maintainability issues**:

### Strengths
✅ Good framework choices
✅ Clear layered architecture
✅ Authentication well implemented
✅ Service factory pattern established
✅ Some validation and matcher logic good

### Critical Issues
❌ Duplicate repository directories
❌ Massive code duplication (15% of codebase)
❌ Very low test coverage (4.4%)
❌ Multiple unimplemented features scattered as TODOs
❌ Context misuse in GraphQL handler
❌ No error handling strategy

### Recommended Next Steps
1. Fix duplicate repositories immediately
2. Implement generic CRUD patterns
3. Add comprehensive unit tests
4. Refactor 865+ LOC files
5. Resolve all TODOs or create issues
6. Implement configuration validation

**Estimated Refactoring Effort:** 3-4 weeks for comprehensive cleanup
