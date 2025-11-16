# Kainuguru API - Architecture Quick Reference

## Project At A Glance

```
Language:   Go 1.24.0
Framework:  Fiber v2 + GraphQL (99designs/gqlgen)
Database:   PostgreSQL 15 + Redis 7
ORM:        Uptrace Bun
Tests:      Unit + BDD Integration
Lines:      3,225 resolver code lines
```

---

## Directory Map

```
kainuguru-api/
├── cmd/                    Entry points (8 CLI commands)
├── internal/               Core application (24 dirs)
│   ├── bootstrap/          Dependency registration
│   ├── config/             Configuration management
│   ├── database/           Bun ORM setup
│   ├── cache/              Redis client
│   ├── middleware/         HTTP middleware
│   ├── models/             15 data models
│   ├── repositories/       22 repository implementations
│   ├── services/           Business logic (39 subdirs)
│   ├── graphql/            GraphQL implementation
│   └── [domain modules]    Store, Flyer, Product, etc.
├── pkg/                    Shared utilities (8 packages)
│   ├── errors/             Structured error handling
│   ├── logger/             Zerolog logging
│   ├── normalize/          Text normalization
│   └── [others]
├── migrations/             33 SQL migrations
├── tests/                  Integration & BDD tests
└── configs/                Configuration files
```

---

## Architecture Layers

```
┌────────────────────────────────────────────┐
│ Presentation Layer                         │
│ - GraphQL Handlers (@internal/handlers/)   │
│ - Middleware Stack                         │
└──────────────┬─────────────────────────────┘
               │
┌──────────────▼──────────────────────────────┐
│ Application Layer                          │
│ - Services (@internal/services/)           │
│ - 14 Core services                         │
│ - 12 Specialized service modules           │
└──────────────┬──────────────────────────────┘
               │
┌──────────────▼──────────────────────────────┐
│ Data Access Layer                          │
│ - Repositories (@internal/repositories/)   │
│ - 22 Repository implementations            │
│ - Generic Base Repository                  │
└──────────────┬──────────────────────────────┘
               │
┌──────────────▼──────────────────────────────┐
│ Data Layer                                 │
│ - Models (@internal/models/)               │
│ - 15 Domain models                         │
│ - Bun ORM annotations                      │
└──────────────┬──────────────────────────────┘
               │
┌──────────────▼──────────────────────────────┐
│ Persistence Layer                          │
│ - PostgreSQL 15                            │
│ - Redis 7                                  │
└────────────────────────────────────────────┘
```

---

## Core Services

### Business Services (interfaces.go)
```
StoreService              Commands                 Tests
FlyerService             FlyerPageService         ProductService
ProductMasterService     ShoppingListService      ShoppingListItemService
ExtractionJobService     PriceHistoryService      
```

### Specialized Services (/services/*)
```
auth/              - User authentication & JWT
search/            - Full-text & hybrid search
scraper/           - Web scraping
matching/          - Product matching
enrichment/        - AI-powered enrichment
recommendation/    - Price optimization
worker/            - Background job processing
cache/             - Caching abstraction
archive/           - Flyer archival
storage/           - File storage
email/             - Email sending
ai/                - OpenAI integration
```

---

## Data Model Diagram

```
Store (1) ──┬─→ (Many) Flyer
            │
            └─→ (Many) ScraperConfig

Flyer (1) ──┬─→ (Many) FlyerPage
            └─→ (Many) Product

FlyerPage (1) ──→ (Many) Product

Product (Many) ──→ (1) ProductMaster
Product (Many) ──→ (1) Store

ProductMaster (1) ──→ (Many) PriceHistory

User (1) ──→ (Many) UserSession
User (1) ──→ (Many) ShoppingList

ShoppingList (1) ──→ (Many) ShoppingListItem
ShoppingListItem ──→ Product / ProductMaster
```

---

## Request Flow

```
Client GraphQL Query
    │
    ▼
Middleware Stack (Recovery, RequestID, CORS, RateLimit, Logger)
    │
    ▼
GraphQL Handler (/internal/handlers/graphql.go)
    │
    ▼
Query Resolver (/internal/graphql/resolvers/query.go)
    │
    ▼
Service Layer (e.g., StoreService.GetByID)
    │
    ▼
Repository Layer (e.g., StoreRepository.GetByID)
    │
    ▼
Bun ORM Query
    │
    ▼
PostgreSQL Database
    │
    ▼
Response (JSON) ──→ Client
```

---

## Configuration Hierarchy

```
Config (root)
├── Server
│   ├── Host, Port
│   ├── Read/Write/Idle Timeouts
│   └── Graceful Shutdown Timeout
├── Database
│   ├── Host, Port, User, Password
│   ├── Max Connections
│   └── SSL Mode
├── Redis
│   ├── Host, Port
│   ├── Max Retries, Pool Size
│   └── Password
├── Logging
│   ├── Level, Format
│   └── Output
├── OpenAI
│   ├── API Key, Base URL, Model
│   ├── Max Tokens, Temperature
│   └── Timeout, Max Retries
├── Auth
│   ├── JWT Secret, Expiry
│   ├── Bcrypt Cost
│   └── Session Timeout
├── CORS
│   ├── Allowed Origins/Methods/Headers
│   ├── Allow Credentials
│   └── Max Age
├── Scraper
│   ├── User Agent
│   ├── Rate Limits, Timeouts
│   └── Retry Settings
├── Worker
│   ├── Number of Workers
│   ├── Queue Check Interval
│   └── Job Timeout
├── Email
│   └── SMTP Configuration
└── Storage
    ├── Type (filesystem, S3, etc.)
    └── Path/URL Configuration
```

---

## Dependency Injection Pattern

```
bootstrap/bootstrap.go
    │
    ├─→ init() [compile-time execution]
    │
    ├─→ RegisterStoreRepositoryFactory(repositories.NewStoreRepository)
    ├─→ RegisterFlyerRepositoryFactory(repositories.NewFlyerRepository)
    └─→ [... more registrations]

cmd/api/main.go
    │
    ├─→ config.Load(env)
    ├─→ server.New(cfg)
    │
    └─→ server/server.go
        │
        ├─→ database.NewBun(cfg)
        ├─→ cache.NewRedis(cfg)
        │
        └─→ ServiceFactory (lazy initialization)
            │
            ├─→ AuthService() [singleton]
            ├─→ StoreService()
            ├─→ FlyerService()
            └─→ [... other services]
```

---

## Database Migrations

**Category Breakdown:**
```
Schema Creation (001-004)
├── 001: Stores
├── 002: Flyers
├── 003: Flyer Pages
└── 004: Products

Infrastructure (005-006)
├── 005: Partition Functions
└── 006: Extraction Jobs

Users & Security (007, 011-012, 023)
├── 007: Store Seeds
├── 011: Users
├── 012: Sessions
└── 023: Login Attempts

Shopping (013-014)
├── 013: Shopping Lists
└── 014: Shopping List Items

Product Management (015-016, 031)
├── 015: Product Masters
├── 016: Tags/Categories
└── 031: Product Master Matches

Search Optimization (010, 021, 026, 029)
├── 010: Trigram Indexes
├── 021: Search Functions
├── 026: Hybrid Search
└── 029: Tags in Search

Performance (024, 025, 027-028, 030, 032-033)
├── 024: Price History
├── 025: Performance Indexes
├── 027: Subcategories
├── 028: Missing Columns
├── 030: ProductMaster Improvements
├── 032: Special Discounts
└── 033: Image URL Updates
```

---

## GraphQL Resolver Structure

```
/internal/graphql/resolvers/
├── query.go              Root Query operations
├── mutation_*.go         Mutation operations
├── store_resolver.go     Store type resolvers
├── flyer_resolver.go     Flyer type resolvers
├── product_resolver.go   Product type resolvers
├── [...19 total resolver files]
│
├── testdata/
│   ├── store_connection.json
│   ├── flyer_connection.json
│   └── [...snapshot fixtures]
│
└── [3,225 total lines of code]
```

---

## Error Handling System

```
Layer          Error Type            Status Code
─────────────────────────────────────────────────
Request        validation            400
               rate_limit            429

Auth           authentication        401
               authorization         403

Business       not_found             404
               conflict              409

System         internal              500
               external              502
```

**Error Structure:**
```go
AppError {
    Type:       ErrorType  // enum: validation, auth, etc.
    Message:    string     // User-friendly
    Code:       string     // Machine-readable
    Details:    string     // Extra context
    StatusCode: int        // HTTP status
    Cause:      error      // Root cause
}
```

---

## Development Commands

```makefile
# Setup & Data
make install              # Docker environment
make seed-data           # Load test fixtures
make db-reset            # Reset database

# Development
make format              # Code formatting
make run                 # Run locally
make test                # Run all tests
make test-snapshots      # GraphQL snapshots
make clean               # Stop containers

# Validation
make validate-all        # Complete validation
make validate-graphql    # GraphQL endpoints
make validate-auth       # Authentication flows
make validate-search     # Search functionality
```

---

## Testing Strategy

```
Unit Tests
├── Located: *_test.go (alongside code)
├── Scope: Individual functions/methods
└── Coverage: Services, repositories, utilities

Integration Tests
├── Located: /tests/bdd/
├── Scope: Feature workflows
└── Style: Behavior-Driven (BDD)

Fixtures
├── Located: /tests/fixtures/
├── Purpose: Test data setup
└── Type: Database fixtures, snapshot data

Test Patterns
├── GraphQL: Snapshot testing
├── Services: Mock repositories
└── Repositories: In-memory or test DB
```

---

## Performance Optimization Features

```
GraphQL Level
├── Dataloader (batch N+1 prevention)
├── Cursor pagination
└── Field selection optimization

Database Level
├── Connection pooling (configurable)
├── Query indexes (25+ performance indexes)
├── Table partitioning (by date)
├── Full-text search (with trigrams)
└── Prepared statements (via Bun)

Application Level
├── Redis caching
├── Rate limiting
├── Worker pool
└── Async job processing
```

---

## Security Architecture

```
Transport
└── TLS/HTTPS (configurable)

Authentication
├── JWT tokens (golang-jwt)
├── Bcrypt hashing (golang.org/x/crypto)
├── Session tokens
└── Login attempt tracking

Authorization
├── User context injection
├── GraphQL field-level auth
└── Role-based access ready

API Security
├── CORS policy
├── Rate limiting (Redis)
├── Request validation
└── Error response sanitization
```

---

## File Naming Patterns

```
Models
  {entity}.go                    (e.g., product.go)

Repositories
  {entity}_repository.go         (implementation)
  {entity}_repository_test.go    (tests)

Services
  {entity}_service.go            (implementation)
  {entity}_service_test.go       (tests)

Domain Modules
  filters.go                     (query filters)
  repository.go                  (interface definition)

Tests
  {feature}_test.go              (unit tests)
  testdata/                      (test fixtures)

Configuration
  config.go                      (structures)
  connections.go                 (setup)
```

---

## Technology Decisions

| Decision | Benefit |
|----------|---------|
| Fiber over Gin | Lighter, async-first, better middleware |
| Bun ORM | PostgreSQL-optimized, modern Go |
| GraphQL | Type-safe, single endpoint, mobile-friendly |
| Service Factory | Centralized DI, easy testing |
| Generic Repositories | Code reuse, type safety |
| PostgreSQL FTS | Native full-text search |
| Redis | Fast caching, rate limiting |
| Dataloader | Eliminates N+1 queries |

---

## Summary: Why This Architecture

1. **Scalable** - Service factory, generic repositories, connection pooling
2. **Testable** - Interface-based design, dependency injection, mocks ready
3. **Maintainable** - Clear separation of concerns, domain organization
4. **Performant** - Indexing, caching, dataloader, pagination
5. **Secure** - JWT, bcrypt, rate limiting, input validation
6. **Modern** - Go 1.24, generics, context, clean code
7. **Flexible** - Service swapping, multiple DB support, extensible

