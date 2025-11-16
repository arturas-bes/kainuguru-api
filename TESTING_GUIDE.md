# Testing Guide - Kainuguru API

## Overview

This guide provides comprehensive instructions for testing the Kainuguru API, with a focus on the new Shopping List Migration Wizard feature.

## Test Infrastructure

### Available Test Scripts

| Script | Purpose | Executable | Lines |
|--------|---------|-----------|-------|
| `run_all_tests.sh` | Master test runner - executes all tests in order | ✅ | 242 |
| `test_wizard_integration.sh` | **NEW** - Complete wizard integration test | ✅ | 348 |
| `test_shopping_list.sh` | Shopping list CRUD operations | ✅ | ~100 |
| `test_shopping_items.sh` | Shopping list item management | ✅ | ~120 |
| `test_create.sh` | Create shopping list with user relation | ✅ | ~50 |
| `test_delete_item.sh` | Delete shopping list item | ✅ | ~40 |
| `test_product_master.sh` | Product master operations | ✅ | ~60 |
| `test_search_verification.sh` | Search functionality | ✅ | ~80 |
| `test_enrichment.sh` | Product enrichment workflow | ✅ | ~200 |
| `test_enrichment_cycle.sh` | Full enrichment cycle | ✅ | ~120 |
| `test-docker-pdf.sh` | Docker PDF processing | ✅ | ~40 |
| `test_fixtures.sql` | Sample data for testing | N/A | ~200 |

### Test Categories

**Phase 1: Basic CRUD** (4 tests)
- Shopping list create/read/update/delete
- Shopping list item operations
- User relationship management

**Phase 2: Core Features** (2 tests)
- Product master CRUD
- Search functionality (name, brand, category)

**Phase 3: Enrichment Pipeline** (2 tests)
- Flyer product enrichment
- Full enrichment cycle

**Phase 4: Wizard Integration** (1 test - NEW)
- Complete wizard lifecycle
- 11 test scenarios covering all features

**Phase 5: Infrastructure** (1 test)
- Docker PDF processing

## Prerequisites

### 1. Running Services

Ensure all services are running via Docker Compose:

```bash
# Start all services
docker-compose up -d

# Check service status
docker-compose ps

# Expected output:
# api     - Up (port 8080)
# db      - Up (healthy, port 5439)
# redis   - Up (healthy, port 6379)
# scraper - Up
```

### 2. Database Migrations

Apply all migrations:

```bash
# Run migrations
go run cmd/migrator/main.go up

# Verify wizard tables exist
docker-compose exec db psql -U kainuguru -d kainuguru_db -c "
  SELECT table_name FROM information_schema.tables 
  WHERE table_schema = 'public' 
  AND table_name IN ('shopping_lists', 'shopping_list_items', 'offer_snapshots')
  ORDER BY table_name;
"
```

### 3. Authentication Token

Obtain a valid JWT token:

```bash
# Option 1: Create a test user and get token via API
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { register(input: { email: \"test@example.com\", password: \"Test1234!\" }) { token user { id email } } }"
  }'

# Option 2: Use existing user
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { login(email: \"test@example.com\", password: \"Test1234!\") { token user { id email } } }"
  }'

# Extract token and export
export API_TOKEN="your_jwt_token_here"
```

### 4. Verify Prerequisites

```bash
# Quick verification script
echo "=== Checking Prerequisites ==="

# Check Docker services
docker-compose ps | grep -E "(api|db|redis)" && echo "✅ Docker services running"

# Check Redis
redis-cli ping | grep PONG && echo "✅ Redis available"

# Check API health
curl -s http://localhost:8080/health | grep healthy && echo "✅ API healthy"

# Check API token
[ -n "$API_TOKEN" ] && echo "✅ API_TOKEN set" || echo "❌ API_TOKEN not set"

# Check jq
command -v jq && echo "✅ jq available"
```

## Running Tests

### Quick Start - Run All Tests

```bash
# Set your JWT token
export API_TOKEN="your_jwt_token_here"

# Optionally load test fixtures
export LOAD_FIXTURES=1

# Run all tests
./run_all_tests.sh
```

**Expected Output:**
```
═══════════════════════════════════════════════════════════════
  Kainuguru API - Integration Test Suite
═══════════════════════════════════════════════════════════════

Test Configuration:
  API URL:    http://localhost:8080/graphql
  Health URL: http://localhost:8080/health

═══════════════════════════════════════════════════════════════
  Checking Prerequisites
═══════════════════════════════════════════════════════════════

✅ API server is healthy
✅ Redis is available
✅ API_TOKEN is set
✅ jq is available

All prerequisites met! Ready to run tests.

... [test execution] ...

═══════════════════════════════════════════════════════════════
  Test Execution Summary
═══════════════════════════════════════════════════════════════

✅ PASS: Shopping List CRUD
✅ PASS: Shopping List Items CRUD
✅ PASS: Create Shopping List
✅ PASS: Delete Shopping List Item
✅ PASS: Product Master Operations
✅ PASS: Search Functionality
✅ PASS: Product Enrichment
✅ PASS: Full Enrichment Cycle
✅ PASS: Wizard Full Integration
✅ PASS: Docker PDF Processing

─────────────────────────────────────────────────────────────
Total Tests:   10
Passed Tests:  10
Failed Tests:  0
Skipped Tests: 0

🎉 Pass Rate: 100% - ALL TESTS PASSED!
```

### Run Individual Tests

#### Test Wizard Integration (Comprehensive)

```bash
export API_TOKEN="your_jwt_token_here"
./test_wizard_integration.sh
```

**What This Tests:**
1. ✅ Query shopping lists for expired items
2. ✅ Start wizard session (validates lock)
3. ✅ Query active wizard session by ID
4. ✅ Record REPLACE decision with suggestion
5. ✅ Record SKIP decision (keep expired item)
6. ✅ Record REMOVE decision (delete item)
7. ✅ Idempotency test (duplicate decision)
8. ✅ Check session progress
9. ✅ Interactive: Cancel wizard OR Confirm wizard
10. ✅ Verify list unlocked after completion
11. ✅ Rate limiting (6 attempts, limit 5/hour)

**Interactive Prompts:**
- When asked "Do you want to confirm the wizard? (y/n)", choose:
  - `y` → Tests wizard confirmation and atomic changes
  - `n` → Tests wizard cancellation and cleanup

#### Test Shopping List CRUD

```bash
export API_TOKEN="your_jwt_token_here"
./test_shopping_list.sh
```

**What This Tests:**
- Create shopping list
- Query shopping lists
- Update shopping list
- Delete shopping list

#### Test Search Functionality

```bash
export API_TOKEN="your_jwt_token_here"
./test_search_verification.sh
```

**What This Tests:**
- Search by name
- Search by brand
- Search by category
- Pagination

### Custom Test Runs

**Run specific test phases:**

```bash
# Phase 1: CRUD only
./test_shopping_list.sh && ./test_shopping_items.sh

# Phase 2: Core features only
./test_product_master.sh && ./test_search_verification.sh

# Phase 4: Wizard only
./test_wizard_integration.sh
```

**Custom API endpoint:**

```bash
export API_URL="https://staging.kainuguru.com/graphql"
export API_TOKEN="your_staging_token"
./test_wizard_integration.sh
```

## Test Data Setup

### Loading Test Fixtures

```bash
# Option 1: Via Docker (recommended)
docker-compose exec -T db psql -U kainuguru -d kainuguru_db < test_fixtures.sql

# Option 2: Direct connection
psql -U kainuguru -d kainuguru_db -f test_fixtures.sql

# Option 3: Via run_all_tests.sh
LOAD_FIXTURES=1 ./run_all_tests.sh
```

**Fixtures Include:**
- Users with valid credentials
- Stores (Maxima, Rimi, Lidl)
- Active flyers with products
- **Expired flyers** (for wizard testing)
- Shopping lists
- Shopping list items (linked to expired products)
- Product master data
- Lithuanian grocery items (with normalized names)

### Creating Test Data for Wizard

If test fixtures aren't available, you can create test data:

```bash
# 1. Create a shopping list
LIST_ID=$(curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_TOKEN" \
  -d '{
    "query": "mutation { createShoppingList(input: { name: \"Test List\", description: \"For wizard testing\" }) { id } }"
  }' | jq -r '.data.createShoppingList.id')

# 2. Add items linked to expired flyer products
# (Requires expired flyers in database - see test_fixtures.sql)

curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_TOKEN" \
  -d "{
    \"query\": \"mutation { addItemToList(listId: \\\"$LIST_ID\\\", input: { productId: \\\"EXPIRED_PRODUCT_ID\\\", quantity: 1 }) { id } }\"
  }"
```

## Troubleshooting

### Common Issues

#### ❌ "API server is not available"

**Cause:** API server not running or not listening on port 8080

**Solution:**
```bash
# Check if API is running
docker-compose ps | grep api

# Restart API service
docker-compose restart api

# Check logs
docker-compose logs api --tail=50
```

#### ❌ "Redis is not available"

**Cause:** Redis server not running or not accessible

**Solution:**
```bash
# Check if Redis is running
docker-compose ps | grep redis

# Test connectivity
redis-cli ping

# Restart Redis
docker-compose restart redis
```

#### ❌ "API_TOKEN environment variable is not set"

**Cause:** Missing JWT authentication token

**Solution:**
```bash
# Option 1: Use existing token
export API_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Option 2: Get new token via login
TOKEN=$(curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { login(email: \"test@example.com\", password: \"Test1234!\") { token } }"
  }' | jq -r '.data.login.token')

export API_TOKEN="$TOKEN"
```

#### ❌ "jq is not installed"

**Cause:** Missing `jq` JSON parser

**Solution:**
```bash
# macOS
brew install jq

# Ubuntu/Debian
sudo apt-get install jq

# Arch Linux
sudo pacman -S jq
```

#### ❌ "rate limit exceeded: maximum 5 wizard sessions per hour"

**Cause:** Rate limiter triggered (expected behavior in test 10)

**Solution:**
```bash
# Clear rate limiter for your user
redis-cli DEL "rate_limit:wizard:YOUR_USER_ID"

# Or wait 1 hour for sliding window to reset
```

#### ❌ "shopping list is already being migrated"

**Cause:** List is locked by an active wizard session

**Solution:**
```bash
# Option 1: Cancel the active wizard
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_TOKEN" \
  -d "{
    \"query\": \"mutation { cancelWizard(sessionId: \\\"SESSION_ID\\\") { status } }\"
  }"

# Option 2: Manually unlock (use with caution)
docker-compose exec db psql -U kainuguru -d kainuguru_db -c "
  UPDATE shopping_lists SET is_locked = false WHERE id = 'YOUR_LIST_ID';
"

# Option 3: Clear all wizard sessions from Redis
redis-cli KEYS "wizard:*" | xargs redis-cli DEL
```

#### ❌ "no expired items found"

**Cause:** No expired flyer products linked to shopping list items

**Solution:**
```bash
# Load test fixtures which include expired flyers
docker-compose exec -T db psql -U kainuguru -d kainuguru_db < test_fixtures.sql

# Or manually expire a flyer
docker-compose exec db psql -U kainuguru -d kainuguru_db -c "
  UPDATE flyers 
  SET valid_to = NOW() - INTERVAL '1 day' 
  WHERE id = 'SOME_FLYER_ID';
"
```

### Debug Mode

Enable verbose output for debugging:

```bash
# Add debug flag to test scripts
DEBUG=1 ./test_wizard_integration.sh

# Or manually inspect GraphQL responses
curl -v -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_TOKEN" \
  -d '{
    "query": "query { shoppingLists { id name expiredItemCount } }"
  }' | jq '.'
```

## Test Coverage

### Wizard Test Scenarios (test_wizard_integration.sh)

| Test | Scenario | Constitution Requirement | Status |
|------|----------|-------------------------|--------|
| 1 | Query lists with expiredItemCount | FR-001 | ✅ |
| 2 | Start wizard (lock validation) | FR-016 | ✅ |
| 3 | Query active wizard session | FR-002 | ✅ |
| 4 | Record REPLACE decision | FR-006, FR-020 | ✅ |
| 5 | Record SKIP decision | FR-007 | ✅ |
| 6 | Record REMOVE decision | FR-008 | ✅ |
| 7 | Idempotency test | FR-014 | ✅ |
| 8 | Check session progress | FR-017 | ✅ |
| 9a | Cancel wizard | FR-011 | ✅ |
| 9b | Confirm wizard (atomic) | FR-009, FR-010 | ✅ |
| 10 | Rate limiting (5/hour) | FR-018 | ✅ |
| 11 | Expired session (30min TTL) | FR-015 | Manual |

### Missing Test Coverage (Planned)

**Unit Tests** (Phase 14: T080-T084):
- Scoring algorithm edge cases
- Store selection with various distributions
- Explanation generation
- Two-pass search logic
- Revalidation scenarios

**Load Tests** (Optional):
- Concurrent wizard sessions (100 users)
- Redis memory usage under load
- Transaction throughput
- Search performance with large catalogs

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_USER: kainuguru
          POSTGRES_PASSWORD: kainuguru
          POSTGRES_DB: kainuguru_db
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5439:5432
      
      redis:
        image: redis:7-alpine
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379
    
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Install dependencies
        run: |
          go mod download
          sudo apt-get update
          sudo apt-get install -y jq postgresql-client
      
      - name: Run migrations
        run: go run cmd/migrator/main.go up
        env:
          DATABASE_URL: postgres://kainuguru:kainuguru@localhost:5439/kainuguru_db?sslmode=disable
          REDIS_URL: redis://localhost:6379
      
      - name: Load fixtures
        run: psql -U kainuguru -h localhost -p 5439 -d kainuguru_db -f test_fixtures.sql
        env:
          PGPASSWORD: kainuguru
      
      - name: Start API server
        run: go run cmd/api/main.go &
        env:
          DATABASE_URL: postgres://kainuguru:kainuguru@localhost:5439/kainuguru_db?sslmode=disable
          REDIS_URL: redis://localhost:6379
          JWT_SECRET: test-secret-key
      
      - name: Wait for API
        run: |
          timeout 30 bash -c 'until curl -sf http://localhost:8080/health; do sleep 1; done'
      
      - name: Get test token
        id: token
        run: |
          TOKEN=$(curl -s -X POST http://localhost:8080/graphql \
            -H "Content-Type: application/json" \
            -d '{"query":"mutation{register(input:{email:\"ci@test.com\",password:\"Test1234!\"}){token}}"}' \
            | jq -r '.data.register.token')
          echo "::set-output name=api_token::$TOKEN"
      
      - name: Run integration tests
        run: ./run_all_tests.sh
        env:
          API_TOKEN: ${{ steps.token.outputs.api_token }}
          API_URL: http://localhost:8080/graphql
```

## Metrics Monitoring During Tests

### Key Metrics to Watch

```bash
# After running tests, check metrics endpoint
curl -s http://localhost:8080/metrics | grep wizard

# Expected wizard metrics:
# wizard_sessions_total{status="ACTIVE"} 0
# wizard_sessions_total{status="COMPLETED"} N
# wizard_sessions_total{status="CANCELLED"} M
# wizard_suggestions_returned{has_same_brand="true"} X
# wizard_acceptance_rate_total{decision="REPLACE"} Y
# wizard_latency_ms_bucket{operation="confirm",le="1000"} Z
```

## Conclusion

This comprehensive testing guide ensures:
- ✅ All services properly configured
- ✅ Authentication and authorization working
- ✅ Core CRUD operations functioning
- ✅ Wizard integration fully tested (11 scenarios)
- ✅ Rate limiting enforced
- ✅ List locking working correctly
- ✅ Idempotency guaranteed
- ✅ Ready for production deployment

For issues or questions, refer to:
- `WIZARD_TESTING_ANALYSIS.md` - Deep technical analysis
- `specs/001-shopping-list-migration/quickstart.md` - Feature quickstart
- `WIZARD_MVP_COMPLETE.md` - Implementation summary
