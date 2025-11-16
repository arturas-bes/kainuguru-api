# Kainuguru API - Complete Project Structure & Architecture Analysis

## Executive Summary

The Kainuguru API is a sophisticated Golang backend GraphQL API for grocery flyer price comparison and shopping optimization. The project follows a **Clean Architecture** pattern with strong separation of concerns, organized around domain-driven design principles with a service-oriented architecture.

**Key Statistics:**
- Go 1.24.0 with extensive module dependencies
- PostgreSQL database with Bun ORM
- GraphQL API with 99designs/gqlgen
- Redis caching layer
- Fiber web framework
- 33 database migrations
- 3,225 lines of resolver code
- Multiple CLI commands for data processing

---

## 1. Directory Structure & Organization

### Root Level Organization

```
kainuguru-api/
├── cmd/                      # Command-line entry points
├── internal/                 # Private application code
├── pkg/                      # Reusable utility packages
├── migrations/               # Database migrations
├── tests/                    # Integration and BDD tests
├── configs/                  # Configuration files
├── scripts/                  # Helper scripts
├── docs/                     # Documentation
├── nginx/                    # Nginx configuration
└── docker-compose.yml        # Docker orchestration
```

### Clear Separation Pattern:
- **cmd/** = Entry points (main functions)
- **internal/** = Application logic (not importable by external packages)
- **pkg/** = Shared utilities (can be imported by others)
- **tests/** = Test suites and fixtures

---

## 2. Architectural Pattern Analysis

### 2.1 Clean Architecture Implementation

The project follows **Clean Architecture** principles with clear layer separation:

```
Presentation Layer (Handlers) → Business Logic (Services) → Data Access (Repositories) → Data Layer (Models)
```

**Layer Dependencies (Direction of Dependency Injection):**
```
GraphQL Handlers
    ↓
Services (Business Logic)
    ↓
Repositories (Data Access)
    ↓
Database Models
```

### 2.2 Domain-Driven Design Elements

The project uses **bounded contexts** represented as domain-specific modules:
- Store Domain (`/internal/store/`)
- Flyer Domain (`/internal/flyer/`)
- FlyerPage Domain (`/internal/flyerpage/`)
- Product Domain (`/internal/product/`)
- ProductMaster Domain (`/internal/productmaster/`)
- ShoppingList Domain (`/internal/shoppinglist/`)
- ShoppingListItem Domain (`/internal/shoppinglistitem/`)
- ExtractionJob Domain (`/internal/extractionjob/`)
- PriceHistory Domain (`/internal/pricehistory/`)

Each domain module contains:
- `filters.go` - Query filters for domain entities
- `repository.go` - Repository interface definition

### 2.3 Service-Oriented Architecture

Central to the design is the Service Factory pattern with interface-based dependency injection:

**Service Factory (`internal/services/factory.go`):**
```go
type ServiceFactory struct {
    db     *bun.DB
    config *config.Config
}
```

Creates and manages service instances with lazy initialization and singleton pattern.

### 2.4 Repository Pattern Implementation

**Base Repository** (`internal/repositories/base/repository.go`):
- Generic CRUD operations using Go generics
- Supports custom select, insert, update, delete options
- Type-safe database operations

**Entity-Specific Repositories:**
- 22 repository implementations
- Each follows interface contract
- Batch operation support
- Advanced filtering capabilities

---

## 3. Module Structure Deep Dive

### 3.1 cmd/ - Entry Points (8 CLI Commands)

```
cmd/
├── api/
│   ├── main.go                 # Main API server entry point
│   └── server/
│       └── server.go           # Server setup and configuration
├── scraper/main.go             # Web scraper for flyers
├── migrator/main.go            # Database migration runner
├── seeder/main.go              # Database seeding utility
├── archive-flyers/main.go      # Flyer archival command
├── enrich-flyers/main.go       # Product enrichment command
├── test-scraper/main.go        # Scraper testing utility
└── test-full-pipeline/main.go  # Full pipeline test harness
```

**Pattern:** Each command is independently executable with its own configuration and initialization.

### 3.2 internal/ - Core Application Logic (24 Directories)

#### Core Infrastructure
- **bootstrap/** - Dependency initialization (`init()` function)
- **config/** - Configuration management (Viper-based)
- **database/** - ORM setup (Bun with PostgreSQL)
- **cache/** - Redis client wrapper
- **middleware/** - HTTP middleware stack
- **monitoring/** - Health checks and metrics
- **migrator/** - Migration execution

#### Domain Layer
- **models/** - 15 data models (Bun-annotated structs)
- **repositories/** - 22 repository implementations
- **services/** - Business logic layer (25+ service files)

#### API Layer
- **handlers/** - HTTP/GraphQL request handlers
- **graphql/** - GraphQL schema and resolvers
  - `schema/` - GraphQL SDL definitions
  - `resolvers/` - Query/Mutation/Subscription implementations
  - `generated/` - gqlgen-generated code
  - `dataloaders/` - N+1 query prevention
  - `scalars/` - Custom scalar types
  - `model/` - GraphQL model generation

#### Domain-Specific Modules
```
internal/
├── store/              {filters.go, repository.go}
├── flyer/              {filters.go, repository.go}
├── flyerpage/          {filters.go, repository.go}
├── product/            {filters.go, repository.go}
├── productmaster/      {filters.go, repository.go}
├── shoppinglist/       {filters.go, repository.go}
├── shoppinglistitem/   {filters.go, repository.go}
├── extractionjob/      {filters.go, repository.go}
├── pricehistory/       {filters.go, repository.go}
└── workers/            {Internal background job coordination}
```

### 3.3 services/ - Business Logic Layer (39 Directories)

#### Core Business Services (Main Layer)
```
internal/services/
├── factory.go                          # Service factory and DI container
├── interfaces.go                       # Service interface definitions
├── repository_registry.go              # Repository factory registration
├── store_service.go                    # Store operations
├── flyer_service.go                    # Flyer operations
├── flyer_page_service.go              # Flyer page operations
├── product_service.go                 # Product operations
├── product_master_service.go          # Product master operations
├── extraction_job_service.go          # Job processing
├── shopping_list_service.go           # Shopping list operations
├── shopping_list_item_service.go      # Shopping list items
├── shopping_list_migration_service.go # Data migration for lists
├── price_history_service.go           # Price tracking
└── product_utils.go                   # Product utilities
```

#### Specialized Service Modules
```
auth/               # Authentication and JWT
search/             # Full-text and hybrid search
scraper/            # Web scraping functionality
matching/           # Product matching
enrichment/         # Product enrichment (AI-powered)
recommendation/     # Shopping recommendations
worker/             # Background job processing
cache/              # Caching layer
archive/            # Flyer archival
storage/            # File storage
email/              # Email sending
ai/                 # AI integration (OpenAI)
```

### 3.4 pkg/ - Shared Utilities (8 Packages)

```
pkg/
├── errors/         # Structured error handling
├── logger/         # Logging (zerolog)
├── config/         # Configuration utilities
├── normalize/      # Text normalization (Lithuanian-specific)
├── openai/         # OpenAI client wrapper
├── pdf/            # PDF processing
└── image/          # Image optimization
```

---

## 4. Data Access Layer Architecture

### 4.1 Repository Pattern Implementation

**Interface Segregation:**
Each repository defines a focused interface in `internal/services/interfaces.go`:

```go
type FlyerRepository interface {
    // Basic CRUD
    GetByID(ctx context.Context, id int) (*models.Flyer, error)
    GetAll(ctx context.Context, filters services.FlyerFilters) ([]*models.Flyer, error)
    Create(ctx context.Context, flyer *models.Flyer) error
    Update(ctx context.Context, flyer *models.Flyer) error
    Delete(ctx context.Context, id int) error
    
    // Specialized queries
    GetCurrentFlyers(ctx context.Context, storeIDs []int) ([]*models.Flyer, error)
    GetProcessableFlyers(ctx context.Context, limit int) ([]*models.Flyer, error)
    
    // Relations
    GetWithPages(ctx context.Context, flyerID int) (*models.Flyer, error)
    
    // Batch operations
    CreateBatch(ctx context.Context, flyers []*models.Flyer) error
    UpdateBatch(ctx context.Context, flyers []*models.Flyer) error
}
```

### 4.2 Repository Registry Pattern

**bootstrap/bootstrap.go** uses `init()` for dependency registration:
```go
func init() {
    services.RegisterStoreRepositoryFactory(repositories.NewStoreRepository)
    services.RegisterFlyerRepositoryFactory(repositories.NewFlyerRepository)
    // ... more registrations
}
```

This enables loose coupling and testable dependencies.

### 4.3 Generic Base Repository

Uses Go 1.18+ generics for reusable CRUD:
```go
type Repository[T any] struct {
    db       *bun.DB
    pkColumn string
}

func (r *Repository[T]) GetByID(ctx context.Context, id interface{}, 
    opts ...SelectOption[T]) (*T, error)
```

---

## 5. Configuration Management

### 5.1 Configuration Structure

**Config Hierarchy** (`internal/config/config.go`):
```go
type Config struct {
    Server   ServerConfig     // Port, timeouts, graceful shutdown
    Database database.Config  // PostgreSQL connection
    Redis    RedisConfig      // Cache configuration
    Logging  LoggingConfig    // Zerolog setup
    OpenAI   OpenAIConfig     // AI/enrichment config
    Scraper  ScraperConfig    // Web scraper settings
    Worker   WorkerConfig     // Background job settings
    CORS     CORSConfig       // CORS policy
    Auth     AuthConfig       // JWT and session config
    App      AppConfig        // Application metadata
    Email    EmailConfig      // Email service config
    Storage  StorageConfig    // File storage config
}
```

### 5.2 Environment Configuration

**Loading Priority:**
1. `.env` file (via godotenv)
2. Environment variables
3. Default values from code

**Environment File** (`.env.dist`):
- Server configuration
- Database credentials
- Redis settings
- CORS policy
- API keys (OpenAI)
- JWT secrets
- Email configuration
- Storage paths

---

## 6. Database Layer Architecture

### 6.1 ORM: Bun with PostgreSQL

**BunDB Wrapper** (`internal/database/bun.go`):
```go
type BunDB struct {
    *bun.DB
    config Config
}
```

**Features:**
- Connection pooling (configurable max connections)
- Query logging in development
- SQL dialect abstraction (PostgreSQL)
- Connection health monitoring

### 6.2 Database Migrations (33 Total)

Strategic migration organization covering:
- Core schema creation (stores, flyers, products)
- Table partitioning for scalability
- Full-text search setup
- Security tables (login attempts)
- Price history tracking
- Performance optimizations

---

## 7. GraphQL Implementation

### 7.1 Schema Organization

**Schema File:**
```
internal/graphql/schema/schema.graphql
```

**Key GraphQL Types:**
- Queries: Store, Flyer, Product, ShoppingList, Search
- Mutations: User auth, Shopping list operations, Product matching
- Subscriptions: Real-time updates
- Custom Scalars: DateTime, JSON
- Connection Types: Cursor-based pagination

### 7.2 Generated Code Structure

```
internal/graphql/
├── schema/                    # GraphQL SDL definitions
├── generated/                 # Auto-generated by gqlgen
├── model/                     # GraphQL model generation
├── resolvers/                 # Manual resolver implementations
├── dataloaders/               # N+1 query prevention
└── scalars/                   # Custom scalar implementations
```

---

## 8. HTTP & Middleware Layer

### 8.1 Web Framework: Fiber v2

**Server Setup** (`cmd/api/server/server.go`):
- HTTP server with configurable timeouts
- Graceful shutdown support
- Middleware pipeline setup
- Route registration

### 8.2 Middleware Stack

**Order of Execution:**
```
1. Recovery      - Panic handling
2. RequestID     - Request tracing
3. CORS          - Cross-origin handling
4. RateLimit     - Rate limiting (Redis-backed)
5. Logger        - Request/response logging
6. GraphQL       - GraphQL handler
```

---

## 9. Service Layer Deep Dive

### 9.1 Service Interface Pattern

All services follow interface-based design in `internal/services/interfaces.go`:

```go
type StoreService interface {
    GetByID(ctx context.Context, id int) (*models.Store, error)
    GetAll(ctx context.Context, filters StoreFilters) ([]*models.Store, error)
    Create(ctx context.Context, store *models.Store) error
    Update(ctx context.Context, store *models.Store) error
    Delete(ctx context.Context, id int) error
    // ... domain-specific operations
}
```

### 9.2 Specialized Services

**Search Service:**
- Full-text search on products
- Hybrid search (keyword + semantic)
- Lithuanian language support

**Auth Service:**
- User registration and login
- JWT token generation
- Password hashing (bcrypt)
- Session management

**Product Enrichment Service:**
- AI-powered product classification
- Brand extraction
- Category assignment

**Worker Service:**
- Background job processing
- Job queue management
- Retry logic

---

## 10. Testing Architecture

### 10.1 Test Structure

```
tests/
├── fixtures/         # Test data setup
├── bdd/              # Behavior-driven tests
└── scripts/          # Data loading scripts
```

### 10.2 Testing Patterns

- **Unit Tests:** Alongside implementation (`*_test.go`)
- **Integration Tests:** BDD-style tests in `/tests/bdd/`
- **Fixtures:** Database setup and test data

---

## 11. Technology Stack Summary

### Core Framework
- Go 1.24.0
- Fiber v2.52.9 (Web Framework)
- 99designs/gqlgen 0.17.81 (GraphQL)

### Database & Caching
- PostgreSQL 15
- Uptrace Bun v1.2.15 (ORM)
- Redis 7

### Key Dependencies
- golang-jwt/jwt v5.3.0 (Authentication)
- golang.org/x/crypto (Password hashing)
- rs/zerolog (Logging)
- spf13/viper (Configuration)
- PuerkitoBio/goquery (HTML parsing)

---

## 12. Separation of Concerns

### Layer Separation

```
GraphQL Handlers (Presentation)
    ↓
Services (Business Logic)
    ↓
Repositories (Data Access)
    ↓
Models + ORM (Data Layer)
    ↓
PostgreSQL + Redis (Persistence)
```

### Domain Separation

Each domain has:
- **Model** - In `internal/models/`
- **Repository** - In `internal/repositories/`
- **Service** - In `internal/services/`
- **GraphQL Resolver** - In `internal/graphql/resolvers/`
- **Filters** - In domain-specific module

---

## 13. File Naming Conventions

**Consistent patterns:**
- `{entity}.go` - Single model per file
- `{entity}_repository.go` - Repository implementation
- `{entity}_service.go` - Service implementation
- `{entity}_test.go` - Tests
- `*_test.go` - Unit tests alongside code

---

## 14. Error Handling Architecture

### Structured Errors (pkg/errors)

**AppError Type:**
```go
type AppError struct {
    Type       ErrorType  // Error classification
    Message    string     // User-friendly message
    Code       string     // Machine-readable code
    Details    string     // Additional context
    StatusCode int        // HTTP status code
    Cause      error      // Root cause
}
```

**Error Types:**
- validation (400)
- authentication (401)
- authorization (403)
- not_found (404)
- conflict (409)
- internal (500)
- external (502)
- rate_limit (429)

---

## 15. Key Design Decisions

| Decision | Reason |
|----------|--------|
| Fiber over Gin | Lightweight, async-first, easier middleware |
| Bun ORM | PostgreSQL-optimized, modern patterns |
| GraphQL | Type-safe API, single endpoint |
| Service Factory | Centralized DI, easy testing |
| Repository Pattern | Data access abstraction |
| Generic Base Repo | Code reuse, type safety |
| Bootstrap init() | Compile-time DI registration |
| Dataloader | Prevents N+1 queries |
| Redis Cache | Fast caching, session storage |

---

## 16. Scalability Considerations

### Database Design
- Partitioning on products table
- Strategic indexes for search
- PostgreSQL FTS with trigram support
- Connection pooling

### Caching Strategy
- Redis for sessions and caching
- Dataloader for batch query optimization
- Service-level caching

### Background Processing
- Configurable worker pool
- Job queues for long operations
- Status tracking via ExtractionJob

---

## 17. Security Features

### Authentication
- JWT-based stateless auth
- Bcrypt password hashing
- Session token management
- Login attempt tracking

### Authorization
- User context injection
- GraphQL field-level authorization
- Role-based access control ready

### API Security
- CORS policy enforcement
- Rate limiting
- Request validation

---

## Conclusion

The Kainuguru API demonstrates a **well-architected, enterprise-grade Golang backend** with:

1. **Clear separation of concerns** - Distinct layers for presentation, business logic, and data access
2. **Scalable design** - Service factory, generics, and interface-based architecture
3. **Modern Go patterns** - Context usage, error handling, generics, and interfaces
4. **Production-ready features** - Logging, monitoring, rate limiting, caching
5. **Comprehensive testing** - Unit tests, integration tests, and BDD scenarios
6. **Domain-driven organization** - Each entity has clear boundaries and responsibilities
7. **GraphQL-first API** - Type-safe, efficient API layer with proper pagination and batching
8. **Database optimization** - Partitioning, indexing, full-text search support
9. **Developer experience** - Makefile commands, Docker setup, hot reload
10. **Error handling** - Structured errors with HTTP status code mapping

This architecture supports growth from current state to significant scale while maintaining code quality and testability.
