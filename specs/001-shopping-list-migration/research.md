# Research Report: Shopping List Migration Wizard

**Date**: 2025-11-15
**Feature**: Shopping List Migration Wizard
**Status**: All clarifications resolved ✅

## Executive Summary

All technical integration points have been researched and verified. The existing infrastructure provides strong support for the Shopping List Migration Wizard feature:

- **SearchService**: Fully supports two-pass search with similarity scoring
- **ProductMaster**: 27+ brands normalized, confidence scoring implemented
- **Redis**: Production-ready with session management capabilities
- **Database**: Sophisticated migration system with 33+ migrations
- **Workers**: Existing framework with scheduled job support
- **GraphQL**: Professional Hyena-style schema with error handling

## Research Findings

### 1. SearchService Integration ✅

**Decision**: Use existing FuzzySearchProducts method with two-pass strategy
**Rationale**: Service already returns three similarity scores (name, brand, combined) that perfectly align with our needs
**Alternatives considered**: Building custom search - rejected due to Constitution principle "Simplicity First"

**Key Findings**:
- Method signature: `FuzzySearchProducts(ctx context.Context, req *SearchRequest) (*SearchResponse, error)`
- Returns similarity scores: name_similarity (70% weight), brand_similarity (30% weight), combined_similarity
- PostgreSQL trigram matching with Lithuanian text normalization
- Brand filtering NOT yet implemented but database functions support it (easy to add)
- Minimum similarity threshold: 0.3
- Results automatically sorted by combined_similarity DESC

**Implementation Notes**:
- Pass 1: Search with `query = "brand_name product_name"`
- Pass 2: Search with `query = "product_name"` only
- Use combined_similarity score directly for ranking
- Can add brand filter when needed (database ready)

### 2. ProductMaster Data Quality ✅

**Decision**: Current implementation sufficient, meets 80% coverage target
**Rationale**: 27 brands normalized, confidence scoring system in place, Lithuanian language support
**Alternatives considered**: Expanding brand list immediately - deferred as TODO for post-MVP

**Key Findings**:
- **Brand Coverage**: 27 brands (15 Lithuanian, 12+ International) with alias mapping
- **Confidence Scoring**: 0.0-1.0 scale, threshold 0.8 for verified status
- **Normalization**: Two-level system (display name + normalized search key)
- **Lithuanian Support**: Special character handling (ą→a, č→c, etc.)
- **Quality Metrics**:
  - MatchCount tracks successful matches
  - ConfidenceScore based on product count
  - Auto-confidence updates: 10+ products = 0.9, 5+ = 0.7, 2+ = 0.6

**Data Model Fields Available**:
- Brand, Name, NormalizedName
- Category, Subcategory, Tags
- StandardUnit, StandardSize
- AvgPrice, MinPrice, MaxPrice
- AvailabilityScore, PopularityScore
- ConfidenceScore, MatchCount

### 3. Redis Session Management ✅

**Decision**: Use Redis for wizard sessions with 30-minute TTL
**Rationale**: Redis already configured and production-ready, perfect for ephemeral session data
**Alternatives considered**: PostgreSQL sessions - rejected due to performance overhead for temporary data

**Key Findings**:
- **Client**: go-redis/v9 fully configured
- **Features**: Hashes, lists, sets, distributed locking
- **Config**: Environment variables with graceful shutdown
- **Session Pattern**: Already used for rate limiting and caching
- **Serialization**: JSON marshaling for Go structs

**Implementation Approach**:
```go
// Session key pattern
key := fmt.Sprintf("wizard:session:%s", sessionID)
// Use HSET for session fields
// Set TTL to 30 minutes
// Track dataset version for staleness detection
```

### 4. Database Migration Strategy ✅

**Decision**: Extend shopping_list_items table with backward-compatible additions
**Rationale**: Table already has complex structure with triggers, safer to extend than recreate
**Alternatives considered**: New table - rejected due to foreign key complexity

**Key Findings**:
- **Migration Tool**: Bun-based migrator (Goose-compatible)
- **Current State**: 33+ migrations tracked
- **Shopping List Items**: 40+ columns, 5 PL/pgSQL triggers
- **Foreign Keys**: All use ON DELETE CASCADE (safe for extensions)
- **Rollback**: Down migrations supported

**Migration Plan**:
```sql
-- Add to shopping_list_items (backward compatible)
ALTER TABLE shopping_list_items ADD COLUMN IF NOT EXISTS
  migration_status VARCHAR(20) DEFAULT NULL,
  original_product_id INTEGER REFERENCES products(id),
  original_price DECIMAL(10,2),
  migration_session_id UUID,
  migrated_at TIMESTAMP;

-- New offer_snapshots table
CREATE TABLE offer_snapshots (
  id SERIAL PRIMARY KEY,
  session_id UUID NOT NULL,
  original_item JSONB NOT NULL,
  suggestions JSONB NOT NULL,
  selected_suggestion JSONB,
  created_at TIMESTAMP DEFAULT NOW()
);
```

### 5. Worker Infrastructure ✅

**Decision**: Create new FlyerExpirationWorker using existing framework
**Rationale**: Framework supports scheduled jobs with Redis locking for distributed execution
**Alternatives considered**: External cron - rejected as infrastructure already exists

**Key Findings**:
- **Scheduler**: robfig/cron/v3 with Redis distributed locking
- **Queue**: Redis-backed with retry logic
- **Existing Pattern**: ProductMasterWorker processes in batches of 100
- **Error Handling**: Alerts if error rate > 10%
- **Schedule Format**: Cron expressions

**Implementation Approach**:
```go
// Register in worker package
scheduler.AddFunc("0 0 * * *", // Daily at midnight
    WithRedisLock("flyer-expiration",
        worker.DetectExpiredItems))

// Batch processing pattern from ProductMasterWorker
batchSize := 100
offset := 0
for {
    items := getExpiredItems(batchSize, offset)
    if len(items) == 0 { break }
    processItems(items)
    offset += batchSize
}
```

### 6. GraphQL API Structure ✅

**Decision**: Add wizard mutations following Hyena-style patterns
**Rationale**: Consistent with existing 50+ types and mutations
**Alternatives considered**: REST endpoints - rejected per Constitution "GraphQL from start"

**Key Patterns**:
- **Naming**: camelCase for fields, Input suffix for inputs
- **Pagination**: Connection pattern with cursors
- **Errors**: AppError type with ErrorType enum
- **Organization**: Separate schema files by domain

**Schema Additions**:
```graphql
# wizard.graphqls
type WizardSession {
  id: ID!
  status: WizardStatus!
  expiredItems: [ExpiredItem!]!
  currentItem: ExpiredItem
  progress: WizardProgress!
  expiresAt: DateTime!
}

type Suggestion {
  product: Product!
  score: Float!
  confidence: Float!
  explanation: String!
  matchedFields: [String!]!
}

input StartWizardInput {
  shoppingListId: ID!
  autoMode: Boolean
}

extend type Mutation {
  startWizard(input: StartWizardInput!): WizardSession!
  getItemSuggestions(sessionId: ID!, itemId: ID!): [Suggestion!]!
  recordDecision(sessionId: ID!, itemId: ID!, suggestionId: ID): Boolean!
  completeWizard(sessionId: ID!): WizardResult!
}
```

## Integration Architecture

### Data Flow
```
1. FlyerExpirationWorker (daily)
   → Detects expired products
   → Flags shopping_list_items
   → Sends notifications

2. User starts wizard
   → Creates Redis session
   → Loads expired items
   → Begins item-by-item flow

3. For each item:
   → Pass 1: SearchService with brand+name
   → Pass 2: SearchService with name only (if needed)
   → Apply deterministic scoring
   → Store selection limit (2 stores max)
   → Record decision in offer_snapshots

4. Complete wizard:
   → Validate dataset freshness
   → Apply changes atomically
   → Clear Redis session
   → Update shopping_list_items
```

### Component Responsibilities

**WizardService** (NEW):
- Session management (Redis)
- Orchestrate two-pass search
- Apply scoring algorithm
- Enforce store limits

**SuggestionEngine** (NEW):
- Calculate scores: brand(3.0), store(2.0), size(1.0), price(1.0)
- Rank suggestions deterministically
- Generate explanations

**StoreSelector** (NEW):
- Greedy algorithm for 2-store optimization
- Calculate coverage and savings
- Handle store preference logic

**Existing Services** (REUSE):
- SearchService: Product searching
- ProductMasterService: Brand/canonical data
- ShoppingListService: List management

## Risk Assessment

### Identified Risks & Mitigations

1. **SearchService Performance**
   - Risk: Two-pass search might be slow
   - Mitigation: Parallel execution, caching frequent searches

2. **Brand Coverage Gaps**
   - Risk: Only 27 brands normalized
   - Mitigation: Fallback to name-only matching, TODO for expansion

3. **Session Data Loss**
   - Risk: Redis restart loses sessions
   - Mitigation: Accept 30-minute window, quick wizard completion

4. **Database Migration Complexity**
   - Risk: Shopping list triggers might conflict
   - Mitigation: Careful testing, rollback plan ready

## Recommendations

### Immediate Actions
1. ✅ Proceed with implementation using existing infrastructure
2. ✅ Use Redis for sessions (already configured)
3. ✅ Extend shopping_list_items table (backward compatible)
4. ✅ Create worker using existing framework

### Post-MVP TODOs
1. Add brand filtering to SearchService (database ready)
2. Expand brand list beyond 27 current brands
3. Implement barcode matching (currently stubbed)
4. Add ML scoring after collecting user feedback
5. Consider PostgreSQL session backup for recovery

## Conclusion

All technical clarifications have been resolved. The existing infrastructure provides excellent support for the Shopping List Migration Wizard. No blocking issues identified. The implementation can proceed with Phase 1 (Design & Contracts) using the decisions documented above.

**Next Steps**:
1. Generate data-model.md with entity schemas
2. Create GraphQL contract files
3. Write quickstart.md for testing
4. Update agent context with new patterns