# Kainuguru API - Infrastructure Components Research

## 1. REDIS INTEGRATION

### Current Status: FULLY IMPLEMENTED

#### Redis Configuration (`internal/cache/redis.go` & `internal/config/config.go`)
- **Dependency**: `github.com/redis/go-redis/v9`
- **Configuration Source**: Environment variables with YAML fallback
  - REDIS_HOST (default: localhost)
  - REDIS_PORT (default: 6379)
  - REDIS_PASSWORD (optional)
  - REDIS_DB (default: 0)
  - REDIS_MAX_RETRIES (default: 3)
  - REDIS_POOL_SIZE (default: 10)

#### Features Implemented:
1. **Basic Operations**:
   - Set/Get (string values)
   - Del, Exists
   - Expire operations

2. **Data Structure Operations**:
   - Hash operations (HSet, HGet, HGetAll, HDel)
   - List operations (LPush, RPop, LLen)
   - Set operations (SAdd, SMembers, SRem)

3. **Advanced Features**:
   - JSON serialization/deserialization (SetJSON, GetJSON)
   - Distributed locking (AcquireLock, ReleaseLock, ExtendLock)
   - Rate limiting with sliding window (RateLimit using Sorted Sets)
   - Session management (SetSession, GetSession, DeleteSession)

4. **Health Checks**:
   - PING-based connection verification
   - Graceful connection closure

### Session Management Implementation

#### Database-based Sessions (Primary):
Location: `internal/repositories/session_repository.go` & `migrations/012_create_sessions.sql`

**Session Table Schema**:
- id (UUID, primary key)
- user_id (UUID, foreign key to users)
- token_hash (unique, indexed)
- is_active (boolean, default true)
- expires_at (timestamp with timezone)
- created_at, last_used_at (timestamps)
- ip_address, user_agent, device_type (session metadata)

**Key Operations**:
- GetActiveSessions: Filters by user_id, is_active=true, expires_at > NOW()
- InvalidateSession: Sets is_active=false
- InvalidateUserSessions: Logs out all user sessions
- CleanupExpiredSessions: Periodic cleanup of expired/inactive sessions (> 24h inactive)
- GetConcurrentSessions: Counts active sessions per user

**Statistics Available**:
- SessionRepositoryStats: Active, recent, total sessions per user
- GlobalSessionStats: System-wide active sessions, daily/weekly counts, avg duration

#### Session vs Redis:
- **Primary storage**: PostgreSQL (durable, with token_hash for fast lookup)
- **Optional Redis caching**: SetSession/GetSession methods exist but likely used for optional caching layer
- **No Redis dependency for core session management** - can function without Redis

---

## 2. DATABASE MIGRATION STRATEGY

### Tool: Custom Migrator (Bun-based)
Location: `internal/migrator/migrator.go`

#### Migration Format:
- **Tool compatibility**: Goose-style comments (`-- +goose Up/Down`)
- **Storage**: SQL files in `/migrations` directory
- **Tracking**: `schema_migrations` table (version, name, applied_at)

#### Migration Files (33+ migrations):
```
001-create_stores.sql               # Core store management
002-create_flyers.sql               # Flyer system
003-create_flyer_pages.sql          # Page extraction
004-create_products.sql             # Product catalog
005-partition_function.sql          # Partitioning setup
006-create_extraction_jobs.sql      # Job tracking
007-seed_stores.sql                 # Store seeding
010-trigram_indexes.sql             # Full-text search indexes
011-create_users.sql                # User management
012-create_sessions.sql             # Session tracking
013-create_shopping_lists.sql       # Shopping list core
014-create_shopping_list_items.sql  # Shopping list items (COMPLEX)
015-create_product_masters.sql      # Product master catalog
016-create_tags_categories.sql      # Tagging system
017-fix_tag_name_ambiguity.sql
021-fix_search_functions.sql
022-fix_similarity_types.sql
023-create_login_attempts.sql       # Security tracking
024-create_price_history.sql        # Price tracking
025-performance_indexes.sql
026-fix_hybrid_search_tsquery.sql
027-add_subcategory_to_products.sql
028-add_missing_product_columns.sql
029-add_tags_to_search_functions.sql
030-product_master_improvements.sql
031-create_product_master_matches.sql
032-add_special_discount_to_products.sql
033-update_flyer_page_image_url_to_relative_path.sql
```

### Shopping List Items Table Structure (Migration 014)

**Core Schema**:
```sql
CREATE TABLE shopping_list_items (
  id BIGSERIAL PRIMARY KEY
  shopping_list_id BIGINT NOT NULL (FK -> shopping_lists)
  user_id UUID NOT NULL (FK -> users) -- Attribution
  
  -- Item Details
  description VARCHAR(255) NOT NULL
  normalized_description VARCHAR(255) -- Lithuanian normalization
  notes TEXT
  
  -- Quantity & Units
  quantity DECIMAL(10,3) CHECK (> 0, <= 999)
  unit VARCHAR(50) -- '1L', '500g', 'vnt'
  unit_type VARCHAR(20) -- 'volume', 'weight', 'count'
  
  -- State Management
  is_checked BOOLEAN DEFAULT false
  checked_at TIMESTAMP WITH TIME ZONE
  checked_by_user_id UUID (FK -> users)
  sort_order INTEGER DEFAULT 0
  
  -- Product Linking
  product_master_id BIGINT (FK -> product_masters)
  linked_product_id BIGINT (FK -> products)
  store_id INTEGER (FK -> stores)
  flyer_id INTEGER (FK -> flyers)
  
  -- Pricing
  estimated_price DECIMAL(10,2)
  actual_price DECIMAL(10,2)
  price_source VARCHAR(50) -- 'flyer', 'user_estimate', 'historical'
  
  -- Organization
  category VARCHAR(100)
  tags TEXT[] -- Array of tags
  
  -- Smart Suggestions
  suggestion_source VARCHAR(50) -- 'manual', 'flyer', 'previous_items', 'popular'
  matching_confidence DECIMAL(3,2) -- 0.00-1.00
  
  -- Availability
  availability_status VARCHAR(20) DEFAULT 'unknown'
  availability_checked_at TIMESTAMP WITH TIME ZONE
  
  -- Full-text Search
  search_vector tsvector
  
  -- Timestamps
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
)
```

#### Foreign Key Constraints:
- `shopping_list_id` -> `shopping_lists(id)` ON DELETE CASCADE
- `user_id` -> `users(id)` ON DELETE CASCADE
- `checked_by_user_id` -> `users(id)` (nullable)
- `product_master_id` -> `product_masters(id)` (nullable)
- `linked_product_id` -> `products(id)` (nullable)
- `store_id` -> `stores(id)` (nullable)
- `flyer_id` -> `flyers(id)` (nullable)

#### Indexes:
- Primary key on `id`
- Composite index: `(shopping_list_id, sort_order)` - for ordering
- Composite index: `(shopping_list_id, is_checked)` - for checked status
- Filtered indexes on nullable FKs (product_master_id, store_id, category)
- TRGM index on normalized_description - for full-text search
- GIN index on search_vector - for PostgreSQL full-text search
- GIN index on tags array

#### Triggers (PL/pgSQL Functions):
1. **update_shopping_list_item_search()**: 
   - Normalizes Lithuanian text
   - Updates search_vector tsvector
   - Runs before INSERT/UPDATE

2. **set_default_sort_order()**:
   - Auto-assigns sort_order as MAX(sort_order) + 1
   - Runs before INSERT

3. **update_item_checked_status()**:
   - Sets checked_at timestamp when is_checked changes
   - Runs before UPDATE

4. **prevent_duplicate_items()**:
   - Prevents duplicate items by normalized description in same list
   - Raises UNIQUE_VIOLATION error
   - Can be disabled for bulk operations

5. **update_shopping_list_stats()**:
   - Updates parent shopping_list statistics
   - Runs after INSERT/UPDATE/DELETE

#### Special Features:
- **Lithuanian Text Normalization**: Custom function translates ąčęėįšųūž to ascii equivalents
- **Full-text Search**: Lithuanian tsvector support for product searching
- **Constraints**:
  - Description length: 1-255 chars
  - Quantity: > 0 and <= 999
  - Matching confidence: 0.00-1.00

---

## 3. WORKER INFRASTRUCTURE

### Framework: Custom Implementation with Cron + Redis

#### Components:

#### 1. Job Queue (`internal/services/worker/queue.go`)
- **Type**: Redis-backed queue (likely using Streams or Lists)
- **Job Structure**:
  ```go
  type Job struct {
    ID           string
    Type         JobType
    Priority     int
    Payload      map[string]interface{}
    MaxAttempts  int
    RetryDelay   time.Duration
    CreatedAt    time.Time
    StartedAt    *time.Time
    CompletedAt  *time.Time
    Status       string
  }
  ```
- **Operations**: Enqueue, Dequeue, Update status, Retry handling

#### 2. Job Scheduler (`internal/services/worker/scheduler.go`)
- **Library**: `robfig/cron/v3` (with seconds precision)
- **Features**:
  - Distributed locking via Redis (prevents duplicate execution)
  - Cron expressions (standard format with seconds)
  - Job status tracking (LastRun, NextRun)
  
**Default Scheduled Jobs**:
```
"0 0 6 * * MON"     -> Weekly Flyer Scraping (Monday 6 AM)
"0 0 8 * * *"       -> Daily Price Update Check (every day 8 AM)
"0 0 2 * * SUN"     -> Weekly Data Archival (Sunday 2 AM, archive > 90 days)
"0 0 3 1 * *"       -> Monthly Cleanup (1st of month 3 AM, cleanup > 180 days)
"0 0 * * * *"       -> Hourly Product Extraction Queue (every hour)
```

#### 3. Job Processor (`internal/services/worker/processor.go`)
- Executes jobs from queue
- Handles retries with exponential backoff
- Manages job lifecycle (started → completed/failed)
- Error tracking and logging

#### 4. Lock Manager (`internal/services/worker/lock.go`)
- Distributed locking via Redis
- Prevents duplicate job execution across instances
- Automatic lock expiration
- Lock renewal support

### Shopping List Migration Worker
Location: `internal/workers/shopping_list_migration_worker.go`

**Purpose**: Automatic migration of expired shopping list items

**Architecture**:
```go
type ShoppingListMigrationWorker struct {
  migrationService ShoppingListMigrationService
  logger           *slog.Logger
  stopChan         chan struct{}
  interval         time.Duration  // Default: 24 hours
}
```

**Key Methods**:
- **Start(ctx)**: Begins scheduled execution
  - Runs immediately on start
  - Then runs on configured interval
- **Stop()**: Gracefully stops worker
- **RunOnce(ctx)**: Manual single execution

**Execution Flow**:
1. Get migration stats BEFORE
2. Call `MigrateExpiredItems(ctx)` -> returns MigrationResult
3. Get migration stats AFTER
4. Log comprehensive metrics
5. Alert if error rate > 10%

**MigrationResult Tracked**:
- total_processed
- successful_migration
- requires_review
- no_match_found
- already_migrated
- errors

**Logging**: Uses slog with structured context (worker: shopping_list_migration)

### Product Master Worker
Location: `internal/workers/product_master_worker.go`

**Purpose**: Match unmatched products to master catalog

**Key Methods**:
1. **ProcessUnmatchedProducts(ctx)**:
   - Batches: 100 products at a time
   - Matching logic by score:
     - >= 0.85: Auto-match
     - >= 0.65: Mark for review
     - < 0.65: Create new master from product
   - Stats: matched, created, review_needed, failed

2. **UpdateMasterConfidence(ctx)**:
   - Updates confidence based on product count:
     - >= 10 products: 0.9
     - >= 5 products: 0.7
     - >= 2 products: 0.6
     - else: 0.5

3. **Run(ctx, interval)**:
   - Scheduled execution with configurable interval
   - Runs both ProcessUnmatched and UpdateConfidence

---

## 4. GRAPHQL SCHEMA

### Schema Organization
Location: `internal/graphql/schema/schema.graphql`

#### Core Type Groups:

#### 1. **Product Types** (Hyena-style)
```graphql
type Product {
  # Identity: id, sku, slug, name, normalizedName, brand
  # Content: description, category, subcategory, tags
  # Pricing: price (ProductPrice), isOnSale, specialDiscount
  # Physical: unitSize, unitType, packageSize, weight, volume
  # Visuals: imageURL, boundingBox, pagePosition
  # Store Context: store, flyer, flyerPage
  # Business: isAvailable, stockLevel, extractionConfidence, requiresReview
  # Temporal: validFrom, validTo, saleStartDate, saleEndDate
  # Computed: isCurrentlyOnSale, isValid, isExpired, validityPeriod
  # Relations: productMaster, priceHistory
}

type ProductPrice {
  current, original, currency
  discount, discountPercent, specialDiscount
  discountAmount, isDiscounted (computed)
}
```

#### 2. **Store & Flyer Types**
```graphql
type Store {
  id, code, name, logoURL, websiteURL
  scraperConfig, scrapeSchedule, lastScrapedAt
  isActive, locations
  Relations: flyers(pagination), products(pagination)
}

type Flyer {
  id, storeID, title
  validFrom, validTo, pageCount
  status (PENDING|PROCESSING|COMPLETED|FAILED)
  isArchived, extractionProgress metrics
  Computed: isValid, isCurrent, daysRemaining, validityPeriod
  Relations: store, flyerPages(pagination), products(pagination)
}

type FlyerPage {
  id, flyerID, pageNumber, imageURL
  status, extraction metrics
  Computed: hasImage, imageDimensions, processingDuration, efficiency
}
```

#### 3. **User & Authentication**
```graphql
type User {
  id (ID!), email, emailVerified
  fullName, preferredLanguage, isActive
  lastLoginAt, createdAt, updatedAt
  Relations: shoppingLists, priceAlerts
}

type AuthPayload {
  user: User!
  accessToken, refreshToken, expiresAt
  tokenType
}
```

#### 4. **Shopping List Types**
```graphql
type ShoppingList {
  id, userID, name, description
  isDefault, isArchived, isPublic, shareCode
  Statistics: itemCount, completedItemCount, estimatedTotalPrice
  Computed: completionPercentage, isCompleted, canBeShared
  Relations: user, items(pagination), categories
}

type ShoppingListItem {
  # Core: id, shoppingListID, userID, description, normalizedDescription
  # Quantity: quantity, unit, unitType
  # State: isChecked, checkedAt, checkedByUserID, sortOrder
  # Linking: productMasterID, linkedProductID, storeID, flyerID
  # Pricing: estimatedPrice, actualPrice, priceSource
  # Organization: category, tags, suggestionSource, matchingConfidence
  # Availability: availabilityStatus, availabilityCheckedAt
  # Computed: totalEstimatedPrice, totalActualPrice, isLinked
  # Relations: shoppingList, user, checkedByUser, productMaster, linkedProduct, store, flyer
}

type ShoppingListCategory {
  id, shoppingListID, userID, name
  colorHex, iconName, sortOrder, itemCount
}
```

#### 5. **Product Master Types**
```graphql
type ProductMaster {
  # Identity: id, canonicalName, normalizedName, brand, category, subcategory
  # Standardization: standardUnitSize, standardUnitType, standardPackageSize, etc.
  # Matching: matchingKeywords, alternativeNames, exclusionKeywords, confidenceScore
  # Statistics: matchedProducts, successfulMatches, failedMatches
  # Status: status (ACTIVE|INACTIVE|PENDING|DUPLICATE|DEPRECATED), isVerified
  # Temporal: lastMatchedAt, verifiedAt, verifiedBy, createdAt, updatedAt
  # Computed: matchSuccessRate
  # Relations: products(pagination)
}

enum ProductMasterStatus {
  ACTIVE, INACTIVE, PENDING, DUPLICATE, DEPRECATED
}
```

#### 6. **Search Types** (Advanced)
```graphql
type SearchResult {
  products: [ProductSearchResult!]!
  totalCount, queryString, suggestions
  hasMore, facets (SearchFacets), pagination
}

type SearchFacets {
  stores, categories, brands, priceRanges, availability
  Each facet has: name, options[FacetOption], activeValue[]
}
```

#### 7. **Price & Alert Types**
```graphql
type PriceHistory {
  id, productMasterID, storeID, flyerID
  price, originalPrice, currency
  isOnSale, recordedAt, validFrom, validTo
  saleStartDate, saleEndDate
  source, extractionMethod, confidence
  isAvailable, stockLevel, notes
  Computed: isCurrentlyValid, isCurrentlySale, discountAmount, etc.
}

type PriceAlert {
  id, userID, productMasterID, storeID
  alertType (PRICE_DROP|TARGET_PRICE|PERCENTAGE_DROP)
  targetPrice, dropPercent, isActive
  notifyEmail, notifyPush, lastTriggered
  triggerCount, lastPrice, expiresAt
  Computed: isActiveAlert
}

enum AlertType {
  PRICE_DROP, TARGET_PRICE, PERCENTAGE_DROP
}
```

#### 8. **Connection Types** (Cursor-based Pagination)
```graphql
# Pattern for all paginated connections:
type [T]Connection {
  edges: [[T]Edge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type [T]Edge {
  node: [T]!
  cursor: String!
}

type PageInfo {
  hasNextPage, hasPreviousPage
  startCursor, endCursor
}
```

### Query Root (Main Operations)

#### Store Queries:
```graphql
store(id: Int!): Store
storeByCode(code: String!): Store
stores(filters: StoreFilters, first: Int, after: String): StoreConnection!
```

#### Flyer Queries:
```graphql
flyer(id: Int!): Flyer
flyers(filters: FlyerFilters, ...): FlyerConnection!
currentFlyers(storeIDs: [Int!], ...): FlyerConnection!
validFlyers(storeIDs: [Int!], ...): FlyerConnection!
```

#### Product Queries:
```graphql
product(id: Int!): Product
products(filters: ProductFilters, ...): ProductConnection!
productsOnSale(storeIDs: [Int!], ...): ProductConnection!
searchProducts(input: SearchInput!): SearchResult!
```

#### Product Master Queries:
```graphql
productMaster(id: Int!): ProductMaster
productMasters(filters: ProductMasterFilters, ...): ProductMasterConnection!
```

#### User/Auth Queries:
```graphql
me: User  # Requires auth
```

#### Shopping List Queries:
```graphql
shoppingList(id: Int!): ShoppingList
shoppingLists(filters: ShoppingListFilters, ...): ShoppingListConnection!
myDefaultShoppingList: ShoppingList
sharedShoppingList(shareCode: String!): ShoppingList
```

#### Price Queries:
```graphql
priceHistory(productMasterID: Int!, ...): PriceHistoryConnection!
currentPrice(productMasterID: Int!, storeID: Int): PriceHistory
priceAlert(id: ID!): PriceAlert
priceAlerts(filters: PriceAlertFilters, ...): PriceAlertConnection!
myPriceAlerts: [PriceAlert!]!
```

### Mutation Root (State Changes)

#### Authentication:
```graphql
register(input: RegisterInput!): AuthPayload!
login(input: LoginInput!): AuthPayload!
logout: Boolean!
refreshToken: AuthPayload!
```

#### Shopping List Management:
```graphql
createShoppingList(input: CreateShoppingListInput!): ShoppingList!
updateShoppingList(id: Int!, input: UpdateShoppingListInput!): ShoppingList!
deleteShoppingList(id: Int!): Boolean!
setDefaultShoppingList(id: Int!): ShoppingList!
```

#### Shopping List Item Management:
```graphql
createShoppingListItem(input: CreateShoppingListItemInput!): ShoppingListItem!
updateShoppingListItem(id: Int!, input: UpdateShoppingListItemInput!): ShoppingListItem!
deleteShoppingListItem(id: Int!): Boolean!
checkShoppingListItem(id: Int!): ShoppingListItem!
uncheckShoppingListItem(id: Int!): ShoppingListItem!
```

#### Price Alert Management:
```graphql
createPriceAlert(input: CreatePriceAlertInput!): PriceAlert!
updatePriceAlert(id: ID!, input: UpdatePriceAlertInput!): PriceAlert!
deletePriceAlert(id: ID!): Boolean!
activatePriceAlert(id: ID!): PriceAlert!
deactivatePriceAlert(id: ID!): PriceAlert!
```

### Input Types

#### Authentication:
```graphql
input RegisterInput {
  email: String!, password: String!
  fullName: String, preferredLanguage: String = "lt"
}

input LoginInput {
  email: String!, password: String!
}
```

#### Filters:
- StoreFilters: isActive, hasFlyers, codes
- FlyerFilters: storeIDs, storeCodes, status, isArchived, date ranges, isCurrent, isValid
- ProductFilters: storeIDs, flyerIDs, flyerPageIDs, productMasterIDs, categories, brands, pricing, dates
- ProductMasterFilters: status, categories, brands, minMatches, minConfidence
- ShoppingListFilters: isDefault, isArchived, isPublic, hasItems, date ranges
- ShoppingListItemFilters: isChecked, categories, tags, hasPrice, isLinked, storeIDs, dates
- PriceHistoryFilters: productMasterID, storeID, dates, isOnSale, pricing, source
- PriceAlertFilters: productMasterID, storeID, alertType, isActive

#### Creation/Update:
```graphql
input CreateShoppingListInput {
  name: String!, description: String
  isDefault: Boolean = false
}

input CreateShoppingListItemInput {
  shoppingListID: Int!, description: String!
  notes: String, quantity: Float = 1
  unit: String, unitType: String
  category: String, tags: [String!] = []
  estimatedPrice: Float
  productMasterID: Int, linkedProductID: Int, storeID: Int
}

input UpdateShoppingListItemInput {
  description: String, notes: String, quantity: Float
  unit: String, unitType: String, category: String
  tags: [String!], estimatedPrice: Float, actualPrice: Float
}

input CreatePriceAlertInput {
  productID: Int!, storeID: Int
  alertType: AlertType!, targetPrice: Float!
  dropPercent: Float, notifyEmail: Boolean = true
  notifyPush: Boolean = false
  notes: String, expiresAt: String
}
```

### Resolver Organization
Location: `internal/graphql/resolvers/`

**Files**:
- `resolver.go`: Root resolver setup
- `query.go`: Query implementations
- `auth.go`: Authentication mutations (Register, Login, Logout, RefreshToken)
- `store.go`: Store-related resolvers
- `flyer.go`: Flyer-related resolvers
- `product.go`: Product-related resolvers
- `shopping_list_item.go`: Shopping list item resolvers
- `price_history.go`: Price history resolvers
- `generated.go`: Code-generated types

#### Error Handling Pattern

Location: `pkg/errors/errors.go`

```go
type ErrorType string

const (
  ErrorTypeValidation = "VALIDATION_ERROR"
  ErrorTypeNotFound = "NOT_FOUND"
  ErrorTypeConflict = "CONFLICT"
  ErrorTypeInternal = "INTERNAL_ERROR"
  // ... more types
)

type AppError struct {
  Type    ErrorType
  Code    string
  Message string
  Details map[string]interface{}
  Err     error
}
```

**Usage in Mutations**:
```go
// From auth.go
func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthPayload, error) {
  result, err := r.authService.Login(ctx, input.Email, input.Password, nil)
  if err != nil {
    return nil, fmt.Errorf("login failed: %w", err)
  }
  // ...
}
```

**Authentication Flow in GraphQL**:
1. User calls `login(input: LoginInput)` mutation
2. Resolver extracts email/password from input
3. Calls `authService.Login(ctx, email, password, metadata)`
4. Service validates credentials and creates session
5. Returns AuthPayload with tokens
6. Frontend stores tokens (usually in HTTP headers via Authorization)
7. Subsequent requests use middleware to extract user from context

**Middleware Integration** (`internal/middleware/auth.go`):
```go
// Extracts user from context
userID, ok := middleware.GetUserFromContext(ctx)
sessionID, ok := middleware.GetSessionFromContext(ctx)
```

---

## SUMMARY TABLE

| Component | Implementation | Storage | Tool/Library |
|-----------|----------------|---------|--------------|
| **Redis** | Full suite with locking & rate limiting | In-memory | go-redis/v9 |
| **Sessions** | Database + optional Redis caching | PostgreSQL (primary) | Bun ORM |
| **Migrations** | Custom Goose-compatible | SQL files | Bun migrator |
| **Shopping List Items** | Complex schema with triggers | PostgreSQL | Bun ORM |
| **Workers** | Scheduled + queue-based | Redis queue + PostgreSQL | robfig/cron + custom |
| **Shopping List Migration** | Interval-based (default 24h) | PostgreSQL | Custom goroutine |
| **Product Master Matching** | Batch processing with thresholds | PostgreSQL | Custom worker |
| **GraphQL** | Full schema with mutations | Generated resolvers | gqlgen |
| **Error Handling** | Structured AppError type | In-memory | Custom pkg/errors |


---

## APPENDIX: QUICK REFERENCE

### Configuration Environment Variables
```
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_DB=0
REDIS_MAX_RETRIES=3
REDIS_POOL_SIZE=10

DB_HOST=localhost
DB_PORT=5432
DB_NAME=kainuguru
DB_USER=postgres
DB_SSLMODE=disable

SERVER_PORT=8080
SERVER_HOST=localhost
```

### Migration Commands (via Bun migrator)
```bash
# Run pending migrations
go run cmd/migrator/main.go up

# Rollback migrations
go run cmd/migrator/main.go down

# Check migration status
go run cmd/migrator/main.go status

# Reset database
go run cmd/migrator/main.go reset
```

### Shopping List Item Expiry Detection
The shopping list item system includes:
- `expires_at` tracking at database level (if implemented)
- Automatic migration worker that runs daily
- Full audit trail with before/after statistics
- Manual `RunOnce()` capability for testing

### Session Cleanup
Automatic cleanup removes sessions matching:
- `expires_at < NOW()` - expired sessions
- `is_active = false AND last_used_at < (NOW() - 24h)` - inactive sessions

### Worker Job Types
Defined in `internal/services/worker/queue.go`:
```go
JobTypeScrapeFlyer      // Weekly flyer scraping
JobTypeUpdatePrices     // Daily price updates
JobTypeArchiveData      // Weekly archival (> 90 days)
JobTypeCleanupData      // Monthly cleanup (> 180 days)
JobTypeExtractProducts  // Hourly product extraction
```

### GraphQL Error Responses
All mutations return structured errors:
```json
{
  "errors": [
    {
      "message": "error description",
      "extensions": {
        "code": "ERROR_TYPE",
        "details": {}
      }
    }
  ]
}
```

### Testing Utilities
Session statistics available:
- `SessionRepository.GetSessionStats(userID)` - Per-user stats
- `SessionRepository.GetGlobalSessionStats()` - System-wide stats
- `SessionRepository.GetSuspiciousSessions()` - Anomaly detection

---

End of Infrastructure Research Document
Generated: 2024-11-15
Project: Kainuguru API
