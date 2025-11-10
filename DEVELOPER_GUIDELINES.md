# Kainuguru API - Developer Guidelines

## Table of Contents
1. [Project Overview](#project-overview)
2. [Getting Started](#getting-started)
3. [Architecture & Design Principles](#architecture--design-principles)
4. [Development Workflow](#development-workflow)
5. [Code Organization & Standards](#code-organization--standards)
6. [Core Components Guide](#core-components-guide)
7. [GraphQL Development](#graphql-development)
8. [Database Guidelines](#database-guidelines)
9. [Authentication & Security](#authentication--security)
10. [Testing Strategy](#testing-strategy)
11. [Performance Considerations](#performance-considerations)
12. [Common Tasks & How-To](#common-tasks--how-to)
13. [Troubleshooting](#troubleshooting)
14. [API Documentation](#api-documentation)

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

*Last Updated: November 2024*
*Version: 1.0.0*