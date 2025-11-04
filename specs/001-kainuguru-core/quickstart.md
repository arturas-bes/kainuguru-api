# Kainuguru MVP - Quick Start Guide

## Prerequisites

- Docker & Docker Compose installed
- Go 1.22+ (for local development)
- PostgreSQL client tools (optional)
- Make utility

## Environment Setup

### 1. Clone and Navigate

```bash
git clone <repository-url>
cd kainuguru-api
git checkout 001-kainuguru-core
```

### 2. Configure Environment Variables

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# Database
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=kainuguru
POSTGRES_USER=kainuguru
POSTGRES_PASSWORD=secure_password_here

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# API
API_PORT=8080
API_ENV=development

# ChatGPT
OPENAI_API_KEY=your_openai_api_key_here
OPENAI_MODEL=gpt-4o

# JWT
JWT_SECRET=your_jwt_secret_here_min_32_chars
JWT_EXPIRY=24h

# Rate Limiting
RATE_LIMIT_RPM=100
RATE_LIMIT_DAILY_QUOTA=50000
```

## Quick Start with Docker

### Start All Services

```bash
make up
```

This starts:
- PostgreSQL 15 (port 5432)
- Redis 7 (port 6379)
- Kainuguru API (port 8080)

### Initialize Database

```bash
make migrate
make seed
```

### Verify Installation

```bash
# Check services are running
docker-compose ps

# Test GraphQL endpoint
curl http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -d '{"query": "{ stores { code name } }"}'
```

## Development Setup

### Install Dependencies

```bash
make deps
```

### Run Locally (without Docker)

```bash
# Start dependencies only
docker-compose up -d postgres redis

# Run API locally
make run
```

### Run Tests

```bash
# Unit tests
make test

# BDD tests
make test-bdd

# All tests with coverage
make test-all
```

## Common Operations

### Scrape Flyers Manually

```bash
make scrape STORE=iki
```

### Access GraphQL Playground

Open browser to: http://localhost:8080/playground

### View Logs

```bash
# All services
docker-compose logs -f

# API only
docker-compose logs -f api

# Scraper logs
docker-compose logs -f api | grep scraper
```

### Database Access

```bash
# Connect to database
make db-shell

# Inside psql:
\dt                          # List tables
SELECT * FROM stores;        # View stores
SELECT * FROM flyers LIMIT 5; # View flyers
```

## Sample GraphQL Queries

### Get Current Flyers (Public)

```graphql
query GetCurrentFlyers {
  currentFlyers {
    id
    store {
      code
      name
    }
    validFrom
    validTo
    pageCount
    productsExtracted
  }
}
```

### Search Products (Public)

```graphql
query SearchProducts {
  searchProducts(input: {
    query: "pienas"
    onlyCurrentFlyers: true
    limit: 10
  }) {
    products {
      name
      brand
      priceCurrent
      store {
        name
      }
    }
    totalCount
  }
}
```

### Register User

```graphql
mutation Register {
  register(input: {
    email: "user@example.com"
    password: "SecurePassword123!"
    fullName: "Jonas Jonaitis"
  }) {
    token
    user {
      id
      email
    }
  }
}
```

### Create Shopping List (Authenticated)

```graphql
mutation CreateList {
  createShoppingList(input: {
    name: "Savaitės pirkiniai"
    isDefault: true
  }) {
    id
    name
    shareToken
  }
}
```

### Add Item to Shopping List

```graphql
mutation AddItem {
  addShoppingListItem(input: {
    shoppingListId: "list-uuid-here"
    textDescription: "Pienas"
    quantity: 2
    unit: "L"
  }) {
    id
    textDescription
    suggestedAlternatives {
      name
      priceCurrent
      store {
        name
      }
    }
  }
}
```

## Monitoring & Debugging

### Check Extraction Status

```bash
# View pending extraction jobs
make db-query "SELECT * FROM extraction_jobs WHERE status='pending';"

# View failed extractions
make db-query "SELECT * FROM flyer_pages WHERE needs_manual_review=true;"
```

### Monitor API Performance

```bash
# View API logs with timing
docker-compose logs api | grep "request_duration"

# Check Redis for rate limiting
docker-compose exec redis redis-cli
> KEYS rate_limit:*
> GET rate_limit:global
```

### Debug ChatGPT Integration

```bash
# View ChatGPT API calls
docker-compose logs api | grep "openai"

# Check cost tracking
make db-query "SELECT DATE(created_at), SUM(cost) FROM api_usage GROUP BY DATE(created_at);"
```

## Troubleshooting

### Service Won't Start

```bash
# Clean restart
make down
make clean
make up
make migrate
```

### Database Connection Issues

```bash
# Check PostgreSQL is running
docker-compose ps postgres

# Test connection
make db-ping
```

### High Memory Usage

```bash
# Restart specific service
docker-compose restart api

# Check resource usage
docker stats
```

### ChatGPT API Errors

1. Check API key in `.env`
2. Verify rate limits not exceeded
3. Check network connectivity
4. Review logs for specific error messages

## Production Deployment

### Build Production Image

```bash
make build-prod
```

### Deploy to DigitalOcean

```bash
# Build and push to registry
docker tag kainuguru-api:latest registry.digitalocean.com/your-registry/kainuguru-api:latest
docker push registry.digitalocean.com/your-registry/kainuguru-api:latest

# On production server
docker-compose -f docker-compose.prod.yml up -d
```

### Setup SSL with Caddy

```yaml
# docker-compose.prod.yml addition
caddy:
  image: caddy:2
  ports:
    - "80:80"
    - "443:443"
  volumes:
    - ./Caddyfile:/etc/caddy/Caddyfile
    - caddy_data:/data
```

## Useful Make Commands

```bash
make help        # Show all available commands
make lint        # Run linters
make fmt         # Format code
make migrate-up  # Apply migrations
make migrate-down # Rollback migrations
make logs        # View logs
make shell       # Enter API container shell
make clean       # Clean up everything
```

## Weekly Maintenance

```bash
# Create next week's partition
make partition-create

# Archive old flyers (remove images)
make archive-flyers WEEKS_OLD=4

# Clean up old sessions
make cleanup-sessions
```

## Support & Documentation

- API Documentation: http://localhost:8080/docs
- GraphQL Schema: See `contracts/schema.graphql`
- Data Model: See `specs/001-kainuguru-core/data-model.md`
- Logs: Check `docker-compose logs` for issues

## Quick Health Check

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "services": {
    "database": "connected",
    "redis": "connected",
    "openai": "configured"
  }
}
```