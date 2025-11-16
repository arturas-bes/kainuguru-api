# Kainuguru API - Architecture Analysis Documentation Index

## Overview

This directory contains comprehensive documentation of the Kainuguru API project structure and architecture. Two primary documents were created as part of this analysis:

### Primary Documentation

#### 1. **PROJECT_ARCHITECTURE_ANALYSIS.md** (627 lines, ~18KB)
   **Comprehensive Technical Reference**
   
   Complete detailed analysis covering:
   - Directory structure and organization
   - Architectural patterns (Clean Architecture, DDD, SOA)
   - Module structure deep dive
   - Data access layer architecture
   - Database layer design
   - GraphQL implementation
   - HTTP and middleware layer
   - Service layer implementation
   - Testing architecture
   - Technology stack summary
   - Dependency injection patterns
   - Separation of concerns
   - File naming conventions
   - Error handling architecture
   - Configuration management
   - Development and deployment
   - Code quality standards
   - Key design decisions
   - Scalability considerations
   - Security features
   
   **Best for:** Understanding the complete architecture, writing design documents, architectural review, developer onboarding

#### 2. **ARCHITECTURE_QUICK_REFERENCE.md** (501 lines, ~14KB)
   **Visual Reference Guide**
   
   Quick reference with:
   - Project at-a-glance stats
   - Directory map with file counts
   - Architecture layer diagram
   - Core services list
   - Data model relationship diagram
   - Request flow visualization
   - Configuration hierarchy
   - Dependency injection pattern
   - Migration categorization
   - GraphQL resolver structure
   - Error handling system
   - Development commands
   - Testing strategy
   - Performance optimization features
   - Security architecture
   - File naming patterns
   - Technology decision rationale
   - Architectural strengths summary
   
   **Best for:** Quick lookups, team reference, architecture discussions, development guidelines

---

## Key Findings Summary

### Project Statistics
- **Language:** Go 1.24.0
- **Framework:** Fiber v2.52.9 + GraphQL (99designs/gqlgen)
- **Database:** PostgreSQL 15 with Bun ORM
- **Cache:** Redis 7
- **Resolver Code:** 3,225+ lines
- **Migrations:** 33 SQL migrations
- **Models:** 15 domain models
- **Repositories:** 22 implementations
- **Services:** 26 (14 core + 12 specialized)
- **CLI Commands:** 8

### Architecture Pattern
**Clean Architecture + Domain-Driven Design + Service-Oriented Architecture**

```
Presentation (GraphQL Handlers)
    ↓
Application (Services)
    ↓
Data Access (Repositories)
    ↓
Data Layer (Models)
    ↓
Persistence (PostgreSQL + Redis)
```

### Architectural Strengths
1. **Scalability** - Service factory, generics, connection pooling, workers
2. **Testability** - Interface-based design, dependency injection
3. **Maintainability** - Clear separation of concerns, domain organization
4. **Performance** - Indexing, caching, dataloader, pagination
5. **Security** - JWT, bcrypt, rate limiting, input validation
6. **Flexibility** - Service swapping, extensible design

---

## Directory Structure Overview

### cmd/ (8 CLI Commands)
```
cmd/
├── api/                    # Main GraphQL API server
├── scraper/               # Web scraping command
├── migrator/              # Database migration runner
├── seeder/                # Data seeding utility
├── archive-flyers/        # Flyer archival command
├── enrich-flyers/         # Product enrichment command
├── test-scraper/          # Scraper testing utility
└── test-full-pipeline/    # Full pipeline test
```

### internal/ (24 Directories)
```
internal/
├── bootstrap/             # Dependency registration
├── config/                # Configuration management
├── database/              # Bun ORM setup
├── cache/                 # Redis client
├── middleware/            # HTTP middleware (5 implementations)
├── models/                # 15 domain models
├── repositories/          # 22 repository implementations
├── services/              # Business logic (39 subdirs)
│   ├── auth/              # Authentication & JWT
│   ├── search/            # Full-text & hybrid search
│   ├── scraper/           # Web scraping
│   ├── matching/          # Product matching
│   ├── enrichment/        # AI-powered enrichment
│   ├── recommendation/    # Shopping optimization
│   ├── worker/            # Background jobs
│   ├── cache/             # Caching abstraction
│   ├── archive/           # Flyer archival
│   ├── storage/           # File storage
│   ├── email/             # Email sending
│   └── ai/                # OpenAI integration
├── graphql/               # GraphQL implementation
│   ├── schema/            # GraphQL SDL
│   ├── resolvers/         # Manual implementations (19+ files)
│   ├── generated/         # gqlgen-generated code
│   ├── dataloaders/       # N+1 prevention
│   ├── scalars/           # Custom scalars
│   └── model/             # Generated models
└── [domain modules]       # Store, Flyer, Product, etc.
```

### pkg/ (8 Utility Packages)
```
pkg/
├── errors/                # Structured error handling
├── logger/                # Zerolog logging
├── config/                # Configuration utilities
├── normalize/             # Text normalization (Lithuanian)
├── openai/                # OpenAI client wrapper
├── pdf/                   # PDF processing
└── image/                 # Image optimization
```

### Additional Directories
```
migrations/               # 33 SQL migrations
tests/                   # Unit + BDD integration tests
configs/                 # Configuration files
scripts/                 # Helper scripts
docs/                    # Documentation
nginx/                   # Nginx configuration
```

---

## Core Components

### Data Models (15 Total)
```
Store → Flyer → FlyerPage → Product → ProductMaster
User → UserSession → ShoppingList → ShoppingListItem
ExtractionJob, PriceHistory, Category, LoginAttempt
```

### Services (26 Total)

**Core Business Services:**
- StoreService - Store operations
- FlyerService - Flyer management
- FlyerPageService - Page handling
- ProductService - Product operations
- ProductMasterService - Master product data
- ShoppingListService - Shopping lists
- ShoppingListItemService - List items
- ExtractionJobService - Job tracking
- PriceHistoryService - Price tracking

**Specialized Services:**
- AuthService - Authentication & JWT
- SearchService - Full-text search
- ScraperService - Web scraping
- MatchingService - Product matching
- EnrichmentService - AI enrichment
- RecommendationService - Shopping optimization
- WorkerService - Background jobs
- CacheService - Caching layer
- ArchiveService - Flyer archival
- StorageService - File storage
- EmailService - Email sending
- AIService - OpenAI integration

### Repositories (22 Total)
- Base Repository (generic CRUD with Go generics)
- 22 entity-specific implementations with specialized queries

### GraphQL Implementation
- Single endpoint: `/graphql`
- 19+ resolver files with 3,225 lines of code
- Cursor-based pagination
- Dataloader for batch loading
- Custom scalars (DateTime, JSON)
- Snapshot testing support

---

## Database Design

### Schema Organization
**Schema Creation:** Stores, Flyers, Flyer Pages, Products
**Infrastructure:** Partition functions, Extraction jobs
**Users & Security:** Users, Sessions, Login attempts
**Shopping:** Shopping lists, Shopping list items
**Product Management:** Product masters, Tags, Categories, Matches
**Search:** Trigram indexes, FTS functions, Hybrid search
**Performance:** Price history, Performance indexes, Optimizations

### Key Features
- 33 total migrations
- 25+ performance indexes
- Full-text search with trigram support
- Table partitioning by date
- Connection pooling configuration

---

## Technology Stack

### Core
- Go 1.24.0
- Fiber v2.52.9 (Web Framework)
- 99designs/gqlgen 0.17.81 (GraphQL)

### Data Access
- Uptrace Bun v1.2.15 (ORM)
- jackc/pgx v5 (PostgreSQL driver)
- PostgreSQL 15
- Redis 7

### Authentication & Security
- golang-jwt/jwt v5.3.0 (JWT)
- golang.org/x/crypto (Password hashing)
- CORS middleware

### Utilities
- rs/zerolog (Logging)
- spf13/viper (Configuration)
- google/uuid (UUID generation)
- PuerkitoBio/goquery (HTML parsing)
- jordan-wright/email (Email)
- robfig/cron/v3 (Scheduled jobs)

### Testing
- stretchr/testify (Testing utilities)
- Custom snapshot testing

---

## Configuration Management

### Environment Structure
```
Config
├── Server (host, port, timeouts)
├── Database (PostgreSQL connection)
├── Redis (cache configuration)
├── Logging (zerolog setup)
├── OpenAI (AI configuration)
├── Auth (JWT, sessions)
├── CORS (cross-origin policy)
├── Scraper (rate limits, timeouts)
├── Worker (job processing)
├── Email (SMTP configuration)
└── Storage (file storage setup)
```

### Loading Priority
1. `.env` file (via godotenv)
2. Environment variables
3. Default values in code

---

## Key Design Decisions

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Web Framework | Fiber v2 | Lightweight, async-first, excellent middleware |
| ORM | Bun | PostgreSQL-optimized, modern Go patterns |
| API | GraphQL | Type-safe, single endpoint, mobile-friendly |
| DI Pattern | Service Factory | Centralized management, easy testing |
| Repositories | Generic Base | Code reuse, type safety with generics |
| Search | PostgreSQL FTS | Native, trigram support, excellent performance |
| Cache | Redis | Fast in-memory, sessions, rate limiting |
| Batching | Dataloader | Prevents N+1 queries in GraphQL |

---

## Testing Architecture

### Unit Tests
- Located alongside implementations (`*_test.go`)
- Coverage: Services, repositories, utilities
- Mocking support via interfaces

### Integration Tests
- Located in `/tests/bdd/`
- Behavior-driven test style
- Database transaction testing
- End-to-end workflow validation

### Fixtures
- Test data setup and teardown
- Database fixtures in `/tests/fixtures/`
- GraphQL snapshot testing

---

## Security Features

### Authentication
- JWT-based stateless authentication
- Bcrypt password hashing
- Session token management
- Login attempt tracking (brute-force protection)

### Authorization
- User context injection via middleware
- GraphQL field-level authorization ready
- Role-based access control foundation

### API Security
- CORS policy enforcement
- Redis-backed rate limiting
- Input validation
- Structured error handling (no data leaks)

---

## Performance Features

### GraphQL Level
- Dataloader for batch loading
- Cursor-based pagination
- Field selection optimization

### Database Level
- Connection pooling
- 25+ performance indexes
- Table partitioning by date
- Full-text search with trigrams
- Prepared statements via Bun

### Application Level
- Redis caching
- Rate limiting
- Worker pool for background jobs
- Async job processing

---

## File Naming Conventions

```
Models:                {entity}.go
Repositories:          {entity}_repository.go, {entity}_repository_test.go
Services:              {entity}_service.go, {entity}_service_test.go
Domain Modules:        filters.go, repository.go
Tests:                 *_test.go (alongside code)
Test Fixtures:         testdata/
Configuration:         config.go, connections.go
Utilities:             Named by function (logger.go, errors.go)
```

---

## Development Commands

### Setup & Data
```bash
make install              # Spin up Docker environment
make seed-data           # Load test fixtures
make db-reset            # Reset database
```

### Development
```bash
make format              # Code formatting
make run                 # Run locally
make test                # Run all tests
make test-snapshots      # GraphQL snapshots
make clean               # Stop containers
```

### Validation
```bash
make validate-all        # Complete validation
make validate-graphql    # GraphQL endpoints
make validate-auth       # Authentication flows
make validate-search     # Search functionality
```

---

## Architecture Visualization

### Request Flow
```
Client GraphQL Query
    ↓
Middleware Stack (Recovery, RequestID, CORS, RateLimit, Logger)
    ↓
GraphQL Handler
    ↓
Query Resolver
    ↓
Service Layer
    ↓
Repository Layer
    ↓
Bun ORM
    ↓
PostgreSQL / Redis
    ↓
Response (JSON)
```

### Dependency Injection Pattern
```
bootstrap/bootstrap.go
    ↓ (init())
    ├─ Register all repository factories
    └─ Initialize dependencies at compile-time

cmd/api/main.go
    ↓
config.Load()
    ↓
server.New()
    ├─ database.NewBun()
    ├─ cache.NewRedis()
    ↓
ServiceFactory
    ├─ AuthService (singleton)
    ├─ StoreService (lazy)
    ├─ FlyerService (lazy)
    └─ ... other services
```

---

## Scalability Considerations

### Database
- Connection pooling (configurable limits)
- Index optimization (25+ indexes)
- Table partitioning by date
- Read replica ready

### Caching
- Redis for session and cache storage
- Dataloader for batch optimization
- Service-level caching ready

### Background Processing
- Configurable worker pool
- Job queue management
- Retry logic built-in
- Status tracking via ExtractionJob

---

## References

### Go Best Practices
- Interface-based design throughout
- Proper error handling with context
- Context usage for async operations
- Modern Go patterns (generics, error wrapping)

### Architecture Patterns
- Clean Architecture (5 layers)
- Domain-Driven Design (9 bounded contexts)
- Service-Oriented Architecture
- Repository Pattern with generics
- Service Factory pattern

---

## Next Steps

### For Developers
1. Review PROJECT_ARCHITECTURE_ANALYSIS.md for comprehensive understanding
2. Use ARCHITECTURE_QUICK_REFERENCE.md as daily reference
3. Follow existing patterns for new features
4. Maintain separation of concerns

### For Team Leadership
1. Use these docs for architecture discussions
2. Reference for design reviews
3. Basis for team standards
4. Onboarding documentation

### For DevOps/Infrastructure
1. Review Docker Compose setup
2. Understand configuration hierarchy
3. Plan for scaling (connection pools, worker counts)
4. Security checklist (JWT secrets, rate limiting)

---

## Document Structure

- **PROJECT_ARCHITECTURE_ANALYSIS.md** - Comprehensive technical reference
- **ARCHITECTURE_QUICK_REFERENCE.md** - Quick visual reference guide
- **ARCHITECTURE_ANALYSIS_INDEX.md** - This file (navigation and summary)

---

**Last Updated:** November 14, 2025
**Analysis Scope:** Complete project structure and architectural patterns
**Coverage:** Directory structure, services, repositories, models, GraphQL, middleware, testing, security, scalability
