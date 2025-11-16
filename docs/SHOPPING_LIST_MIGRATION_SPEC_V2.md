/s# Shopping List Migration Specification V2.0
## Integration-Ready Implementation

**Version:** 2.0
**Date:** November 14, 2025
**Status:** Final - Integration Ready
**Focus:** Lean implementation with existing system integration

---

## Change Summary

### Key Changes from V1
1. **Simplified free list** - No enrichment initially, just CRUD
2. **Two-pass search strategy** - Brand-aware substitution using existing search endpoint
3. **Store cap enforcement** - Maximum 2 stores per plan
4. **Deterministic scoring** - Fixed weights, no ML initially
5. **Lineage tracking** - Added fields to track item origin (flyer vs free text)
6. **Offer snapshots** - Immutable history of applied decisions

---

## 1. Data Model Updates

### 1.1 ShoppingListItem Extensions

**Add these fields to existing model:**

```sql
ALTER TABLE shopping_list_items ADD COLUMN IF NOT EXISTS
    origin VARCHAR(20) DEFAULT 'free_text' CHECK (origin IN ('flyer', 'free_text', 'manual')),
    flyer_product_id BIGINT REFERENCES products(id) ON DELETE SET NULL,
    store_id INT REFERENCES stores(id) ON DELETE SET NULL,
    -- product_master_id already exists in current system
    last_migration_at TIMESTAMP,
    migration_count INT DEFAULT 0;

-- Indexes for migration queries
CREATE INDEX idx_shopping_items_origin ON shopping_list_items(origin, shopping_list_id);
CREATE INDEX idx_shopping_items_flyer_expired ON shopping_list_items(flyer_product_id)
    WHERE origin = 'flyer' AND product_master_id IS NOT NULL;
```

**Go Model Updates:**
```go
type ShoppingListItem struct {
    // ... existing fields ...

    // New migration tracking fields
    Origin           ItemOrigin  `bun:"origin,notnull,default:'free_text'"`
    FlyerProductID   *int64      `bun:"flyer_product_id"`
    StoreID          *int        `bun:"store_id"`
    LastMigrationAt  *time.Time  `bun:"last_migration_at"`
    MigrationCount   int         `bun:"migration_count,default:0"`
}

type ItemOrigin string
const (
    ItemOriginFlyer    ItemOrigin = "flyer"
    ItemOriginFreeText ItemOrigin = "free_text"
    ItemOriginManual   ItemOrigin = "manual"
)
```

### 1.2 Offer Snapshot Table (New)

**Purpose:** Immutable record of what was offered/selected during migration

```sql
CREATE TABLE offer_snapshots (
    id BIGSERIAL PRIMARY KEY,
    shopping_list_item_id BIGINT NOT NULL REFERENCES shopping_list_items(id),
    wizard_session_id UUID REFERENCES wizard_sessions(id),

    -- What was offered
    product_id BIGINT REFERENCES products(id),
    product_master_id BIGINT REFERENCES product_masters(id),
    store_id INT NOT NULL REFERENCES stores(id),

    -- Price at time of offer
    price DECIMAL(10,2) NOT NULL,
    price_unit VARCHAR(20),
    is_estimated BOOLEAN DEFAULT false,

    -- Validity
    valid_from TIMESTAMP NOT NULL,
    valid_to TIMESTAMP,

    -- Decision
    was_selected BOOLEAN DEFAULT false,
    decision_action VARCHAR(20), -- 'replace', 'keep', 'remove'

    -- Metadata
    confidence_score FLOAT,
    rank INT,
    explanation TEXT,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Prevent duplicates
    UNIQUE(shopping_list_item_id, wizard_session_id, product_id, store_id)
);

CREATE INDEX idx_offer_snapshots_item ON offer_snapshots(shopping_list_item_id);
CREATE INDEX idx_offer_snapshots_session ON offer_snapshots(wizard_session_id);
```

---

## 2. Integration with Existing Search Service

### 2.1 Two-Pass Search Strategy

**Use existing `SearchService.FuzzySearchProducts` with brand awareness:**

```go
// internal/services/wizard/matcher.go

type ExpiredItemMatcher struct {
    searchService search.Service
    logger       *slog.Logger
}

func (m *ExpiredItemMatcher) FindCandidates(
    ctx context.Context,
    item *models.ShoppingListItem,
    filters *MigrationFilters,
) ([]*ProductCandidate, error) {

    // Must have product master for expired items
    if item.ProductMasterID == nil {
        return nil, ErrNoProductMaster
    }

    // Load product master for canonical name and brand
    master, err := m.loadProductMaster(ctx, *item.ProductMasterID)
    if err != nil {
        return nil, err
    }

    candidates := make([]*ProductCandidate, 0)

    // PASS 1: Strong search (same brand + name)
    if master.Brand != nil && *master.Brand != "" {
        strongReq := &search.SearchRequest{
            Query:       fmt.Sprintf("%s %s", *master.Brand, master.CanonicalName),
            Limit:       10,
            StoreIDs:    filters.StoreIDs,
            Category:    master.Category,
            PreferFuzzy: true,
        }

        if item.StoreID != nil {
            // Prioritize same store
            strongReq.StoreIDs = []int{*item.StoreID}
        }

        strongResults, err := m.searchService.FuzzySearchProducts(ctx, strongReq)
        if err == nil && len(strongResults.Products) > 0 {
            for _, p := range strongResults.Products {
                candidates = append(candidates, &ProductCandidate{
                    Product:     p,
                    IsSameBrand: p.Brand != nil && master.Brand != nil &&
                                *p.Brand == *master.Brand,
                    SearchPass:  1,
                    BaseScore:   strongResults.Scores[p.ID],
                })
            }
        }
    }

    // PASS 2: Loose search (name only, cross-brand)
    // Only if we need more candidates
    if len(candidates) < 5 {
        looseReq := &search.SearchRequest{
            Query:       master.CanonicalName,
            Limit:       10 - len(candidates),
            StoreIDs:    filters.StoreIDs,
            Category:    master.Category,
            PreferFuzzy: true,
        }

        looseResults, err := m.searchService.FuzzySearchProducts(ctx, looseReq)
        if err == nil {
            for _, p := range looseResults.Products {
                // Skip if already in candidates
                if !containsProduct(candidates, p.ID) {
                    candidates = append(candidates, &ProductCandidate{
                        Product:     p,
                        IsSameBrand: false,
                        SearchPass:  2,
                        BaseScore:   looseResults.Scores[p.ID] * 0.8, // Penalty for loose match
                    })
                }
            }
        }
    }

    return candidates, nil
}
```

---

## 3. Deterministic Scoring (No ML)

### 3.1 Fixed Weight Scoring

```go
// internal/services/wizard/scoring.go

type ScoringWeights struct {
    SameBrand       float64 // 3.0
    OriginalStore   float64 // 2.0
    PreferredStore  float64 // 2.0
    SizeMatch       float64 // 1.0 (within ±20%)
    PriceBetter     float64 // 1.0 (cheaper than original)
}

var DefaultWeights = ScoringWeights{
    SameBrand:      3.0,
    OriginalStore:  2.0,
    PreferredStore: 2.0,
    SizeMatch:      1.0,
    PriceBetter:    1.0,
}

func ScoreCandidate(
    original *models.ShoppingListItem,
    candidate *ProductCandidate,
    context *ScoringContext,
) float64 {
    score := candidate.BaseScore // From search

    // Brand match
    if candidate.IsSameBrand {
        score += DefaultWeights.SameBrand
    }

    // Store match
    if original.StoreID != nil && candidate.Product.StoreID == *original.StoreID {
        score += DefaultWeights.OriginalStore
    }

    // Preferred store
    for _, storeID := range context.PreferredStores {
        if candidate.Product.StoreID == storeID {
            score += DefaultWeights.PreferredStore
            break
        }
    }

    // Size within tolerance (±20%)
    if original.Quantity > 0 && candidate.Product.PackageSize != nil {
        sizeDelta := math.Abs(candidate.Product.PackageSize - original.Quantity)
        if sizeDelta / original.Quantity <= 0.20 {
            score += DefaultWeights.SizeMatch
        }
    }

    // Price improvement
    if original.EstimatedPrice != nil && candidate.Product.CurrentPrice < *original.EstimatedPrice {
        score += DefaultWeights.PriceBetter
    }

    return score
}

// Rank candidates deterministically
func RankCandidates(candidates []*ProductCandidate) {
    sort.Slice(candidates, func(i, j int) bool {
        if candidates[i].FinalScore != candidates[j].FinalScore {
            return candidates[i].FinalScore > candidates[j].FinalScore
        }
        // Tie-break by price (ascending)
        return candidates[i].Product.CurrentPrice < candidates[j].Product.CurrentPrice
    })
}
```

---

## 4. Store Selection & Capping

### 4.1 Store Strategy with Hard Cap

```go
type StoreSelectionStrategy struct {
    MaxStores int // Default: 1, Max: 2
}

func (s *StoreSelectionStrategy) SelectStores(
    items []*MigrationCandidate,
    allCandidates map[int64][]*ProductCandidate,
) ([]int, error) {

    if s.MaxStores <= 0 {
        s.MaxStores = 1
    }
    if s.MaxStores > 2 {
        s.MaxStores = 2 // Hard cap
    }

    // Count coverage by store
    storeCoverage := make(map[int]int)
    storePrices := make(map[int]float64)

    for itemID, candidates := range allCandidates {
        for _, c := range candidates {
            storeCoverage[c.Product.StoreID]++
            storePrices[c.Product.StoreID] += c.Product.CurrentPrice
        }
    }

    // Greedy selection: highest coverage, lowest price
    type storeScore struct {
        StoreID  int
        Coverage int
        TotalPrice float64
    }

    scores := make([]storeScore, 0, len(storeCoverage))
    for storeID, coverage := range storeCoverage {
        scores = append(scores, storeScore{
            StoreID:    storeID,
            Coverage:   coverage,
            TotalPrice: storePrices[storeID],
        })
    }

    // Sort by coverage desc, then price asc
    sort.Slice(scores, func(i, j int) bool {
        if scores[i].Coverage != scores[j].Coverage {
            return scores[i].Coverage > scores[j].Coverage
        }
        return scores[i].TotalPrice < scores[j].TotalPrice
    })

    // Select up to MaxStores
    selected := make([]int, 0, s.MaxStores)
    for i := 0; i < len(scores) && i < s.MaxStores; i++ {
        selected = append(selected, scores[i].StoreID)

        // Only add second store if it adds significant value
        if i == 1 {
            additionalCoverage := scores[i].Coverage
            priceSavings := scores[0].TotalPrice - scores[i].TotalPrice

            if additionalCoverage < 2 && priceSavings < 5.0 {
                break // Not worth the extra store trip
            }
        }
    }

    return selected, nil
}
```

---

## 5. GraphQL Schema Updates

### 5.1 Fixed Enums and Types

```graphql
# Fixed PriceStrategy enum (removed invalid BALANCED)
enum PriceStrategy {
  CHEAPEST
  SIMILAR
  PREMIUM
  BEST_VALUE
}

# Fixed AutoUpdateStrategy with documented thresholds
enum AutoUpdateStrategy {
  CONSERVATIVE  # ≥0.90 confidence
  BALANCED      # ≥0.75 confidence
  AGGRESSIVE    # ≥0.55 confidence
  SMART         # Uses ML (future)
}

# Updated MigrationFiltersInput
input MigrationFiltersInput {
  storeIds: [ID!]
  storeStrategy: StoreStrategy
  priceStrategy: PriceStrategy
  maxPriceDeltaPercent: Float
  brandStrategy: BrandStrategy
  sizeStrategy: SizeStrategy
  qualityTier: QualityTier

  # NEW: Store cap
  maxStores: Int = 1  # Default 1, max 2

  requireOrganic: Boolean
  requireLocal: Boolean
  excludeAllergens: [String!]
}

# Enhanced suggestion type
type RankedSuggestion {
  rank: Int!
  productMaster: ProductMaster!
  activeProduct: Product!
  store: Store!

  # Scoring
  scores: SuggestionScores!

  # NEW: Clearer fields
  isSameBrand: Boolean!      # True if brand matches original
  isEstimated: Boolean!      # True if price is estimated
  explanation: String!       # e.g., "Same brand, €0.40 cheaper at Maxima"

  # Comparison
  priceDelta: Float!
  priceDeltaPercent: Float!
  sizeDelta: String

  # User guidance
  pros: [String!]!
  cons: [String!]!
  badges: [String!]!
  confidence: Float!
}
```

### 5.2 Session Integrity

```graphql
type WizardSession {
  id: ID!
  state: WizardState!

  # NEW: Data version tracking
  datasetVersion: String!    # Hash of last flyer ingestion

  # ... existing fields ...
}

# Completion with validation
type WizardCompletionResult {
  success: Boolean!

  # NEW: Validation status
  validationStatus: ValidationStatus!
  staleItems: [StaleItemWarning!]

  # ... existing fields ...
}

enum ValidationStatus {
  VALID
  STALE_DATA      # Dataset changed
  EXPIRED_ITEMS   # Some candidates expired
  PRICE_CHANGES   # Significant price changes detected
}

type StaleItemWarning {
  item: ShoppingListItem!
  reason: String!
  suggestedAction: String!
}

# Idempotency support
input CompleteWizardInput {
  sessionId: ID!
  idempotencyKey: String  # Optional, prevents double-apply
}
```

---

## 6. Migration Triggers & Workflow

### 6.1 Expired Item Detection

```go
// internal/workers/flyer_expiration_worker.go

func (w *FlyerExpirationWorker) ProcessExpiredFlyers(ctx context.Context) error {
    // Mark expired flyers
    expiredFlyerIDs, err := w.markExpiredFlyers(ctx)
    if err != nil {
        return err
    }

    if len(expiredFlyerIDs) == 0 {
        return nil
    }

    // Find affected shopping list items
    affected, err := w.findAffectedItems(ctx, expiredFlyerIDs)
    if err != nil {
        return err
    }

    // Group by shopping list
    byList := groupByList(affected)

    // Create wizard sessions for affected lists
    for listID, items := range byList {
        session := &models.WizardSession{
            ShoppingListID: listID,
            State:         models.WizardStateInitialized,
            TriggerReason: "flyer_expiration",
            PendingItems:  items,
            DatasetVersion: w.getCurrentDatasetVersion(),
        }

        if err := w.wizardService.CreateSession(ctx, session); err != nil {
            w.logger.Error("failed to create wizard session",
                "list_id", listID, "error", err)
        }

        // Notify user (if preferences allow)
        w.notifyUser(ctx, listID, len(items))
    }

    return nil
}
```

### 6.2 Free List (Simple CRUD - No Enrichment Initially)

```go
// Keep shopping lists simple for now
type ShoppingListService interface {
    // Basic CRUD - no automatic enrichment
    CreateItem(ctx context.Context, item *models.ShoppingListItem) error
    UpdateItem(ctx context.Context, item *models.ShoppingListItem) error
    DeleteItem(ctx context.Context, itemID int64) error

    // Manual trigger for enrichment (Phase 2)
    // FindOffersForList(ctx context.Context, listID int64) (*EnrichmentResult, error)
}
```

---

## 7. Metrics & Observability

### 7.1 Essential Metrics

```go
// internal/metrics/wizard.go

var (
    // Items flagged for migration
    ItemsFlagged = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "wizard_items_flagged_total",
            Help: "Total items flagged for migration",
        },
        []string{"reason"}, // "expired", "unavailable", "better_price"
    )

    // Suggestions returned
    SuggestionsReturned = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "wizard_suggestions_returned",
            Help: "Number of suggestions per item",
            Buckets: []float64{0, 1, 2, 3, 4, 5, 10},
        },
        []string{"has_same_brand"},
    )

    // Acceptance rate
    AcceptanceRate = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "wizard_decision_total",
            Help: "User decisions by action",
        },
        []string{"action"}, // "replace", "keep", "remove", "skip"
    )

    // Store count in plans
    SelectedStoreCount = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name: "wizard_selected_store_count",
            Help: "Number of stores in migration plans",
            Buckets: []float64{1, 2, 3},
        },
    )

    // Latency
    WizardLatency = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "wizard_latency_ms",
            Help: "End-to-end wizard completion time",
            Buckets: prometheus.DefBuckets,
        },
        []string{"outcome"}, // "completed", "cancelled", "timeout"
    )
)
```

---

## 8. Testing Requirements

### 8.1 Core Test Scenarios

```go
// tests/integration/wizard_core_test.go

func TestWizard_ExpiredItemMigration(t *testing.T) {
    // Given: Shopping list items linked to expired flyer products
    list := createTestList(t)
    items := linkItemsToExpiringProducts(t, list)

    // When: Flyers expire
    expireFlyers(t)

    // Then: Wizard session created with candidates
    session := getWizardSession(t, list.ID)
    assert.NotNil(t, session)
    assert.Equal(t, len(items), len(session.PendingItems))

    // And: Candidates include same-brand options when available
    for _, item := range session.PendingItems {
        suggestions := getSuggestions(t, session.ID, item.ItemID)

        if hasSameBrandInMarket(item) {
            assert.True(t, hasSameBrandSuggestion(suggestions),
                "Must include same-brand option when available")
        }
    }
}

func TestWizard_StoreCapEnforcement(t *testing.T) {
    // Given: Wizard with maxStores=1
    filters := &MigrationFilters{MaxStores: 1}

    // When: Generating plan
    plan := generatePlan(t, items, filters)

    // Then: Never uses more than 1 store
    stores := extractStores(plan)
    assert.LessOrEqual(t, len(stores), 1,
        "Plan must respect maxStores cap")
}

func TestWizard_DeterministicRanking(t *testing.T) {
    // Given: Fixed set of candidates
    candidates := getTestCandidates()

    // When: Ranking multiple times
    rankings := make([][]int, 10)
    for i := 0; i < 10; i++ {
        ranked := rankCandidates(candidates)
        rankings[i] = extractIDs(ranked)
    }

    // Then: Ranking is deterministic
    for i := 1; i < 10; i++ {
        assert.Equal(t, rankings[0], rankings[i],
            "Ranking must be deterministic")
    }
}

func TestWizard_SessionRevalidation(t *testing.T) {
    // Given: Active wizard session
    session := createTestSession(t)
    selectSuggestions(t, session)

    // When: Product expires mid-session
    expireProduct(t, session.SelectedProducts[0])

    // Then: Complete returns validation warning
    result := completeWizard(t, session.ID)
    assert.Equal(t, ValidationStatusStaleData, result.ValidationStatus)
    assert.NotEmpty(t, result.StaleItems)
}

func TestWizard_SameBrandAlwaysShown(t *testing.T) {
    // Given: Item with brand "Heinz"
    item := createTestItem(t, "Heinz Ketchup")

    // When: Getting suggestions (even if Heinz is in different store)
    suggestions := getSuggestions(t, item)

    // Then: At least one Heinz option is shown
    heinzFound := false
    for _, s := range suggestions {
        if s.IsSameBrand {
            heinzFound = true
            break
        }
    }
    assert.True(t, heinzFound,
        "Must show same-brand option when available anywhere")
}
```

---

## 9. Rollout Plan

### 9.1 Milestone 1: Expired Items Wizard (Week 1-2)

**Deliverables:**
- [ ] Database migrations for new fields
- [ ] Offer snapshot table
- [ ] Two-pass search implementation
- [ ] Deterministic scoring
- [ ] Store cap enforcement
- [ ] Session validation
- [ ] 5 core integration tests
- [ ] Metrics dashboard

**Acceptance:**
- Expired items trigger wizard
- Same-brand products prioritized
- MaxStores respected
- Deterministic results

### 9.2 Milestone 2: Free List Enrichment (Week 3-4)

**Deliverables:**
- [ ] Manual "Find offers" button
- [ ] Reuse matching engine
- [ ] User preference learning
- [ ] A/B testing framework

**Acceptance:**
- Feature flag controlled
- Same scoring as expired items
- Metrics show adoption rate

---

## 10. Integration Checklist

### 10.1 No Duplicate Logic

- [x] **Use existing SearchService** - Don't create new search logic
- [x] **Use existing ProductMaster** - Already has canonical names
- [x] **Use existing scoring from search** - Base scores from fuzzy search
- [x] **Use existing flyer status** - Don't duplicate expiration logic

### 10.2 System Compatibility

- [x] **Shopping list items backward compatible** - New fields nullable
- [x] **GraphQL schema additive only** - No breaking changes
- [x] **Existing migrations work** - shopping_list_migration_service unchanged
- [x] **Worker integration clean** - Reuses existing worker patterns

### 10.3 Data Integrity

- [x] **Foreign keys maintained** - All references valid
- [x] **Indexes for performance** - Cover new query patterns
- [x] **Constraints enforced** - origin enum, maxStores cap
- [x] **Snapshots immutable** - Historical accuracy preserved

---

## 11. Configuration

```yaml
wizard:
  matching:
    two_pass_enabled: true
    similarity_threshold: 0.3

  scoring:
    weights:
      same_brand: 3.0
      original_store: 2.0
      preferred_store: 2.0
      size_match: 1.0
      price_better: 1.0

  stores:
    max_stores_default: 1
    max_stores_limit: 2
    min_additional_coverage: 2  # Items to justify 2nd store
    min_price_savings: 5.0      # EUR to justify 2nd store

  auto_update:
    conservative_threshold: 0.90
    balanced_threshold: 0.75
    aggressive_threshold: 0.55

  session:
    timeout_minutes: 30
    revalidation_enabled: true
```

---

## 12. Risk Mitigation

### Hard Rules (Never Violate)

1. **Never fabricate prices** - Only use actual flyer prices
2. **Never exceed maxStores** - Hard cap at 2
3. **Always show same-brand** - If it exists anywhere
4. **Keep previous option available** - User can always "keep"
5. **Validate before applying** - Check for stale data

### Trust Safeguards

- Explanations for every suggestion
- Confidence scores visible
- Rollback capability within 24 hours
- Audit log of all decisions

---

## Appendix A: API Examples (Corrected)

### Initialize Wizard
```graphql
mutation {
  startWizard(input: {
    shoppingListId: "123"
    filters: {
      storeStrategy: SAME_STORE
      priceStrategy: BEST_VALUE  # Fixed: was BALANCED
      brandStrategy: SAME_BRAND
      maxStores: 1               # NEW: Store cap
    }
  }) {
    id
    datasetVersion              # NEW: For validation
    totalItems
    pendingItems {
      item {
        id
        description
        origin                  # NEW: flyer/free_text
      }
      reason
    }
  }
}
```

### Get Suggestions with Explanation
```graphql
query {
  getItemSuggestions(
    sessionId: "wizard-123"
    itemId: "456"
  ) {
    suggestions {
      rank
      productMaster {
        name
        brand
      }
      activeProduct {
        price
        store {
          name
        }
      }
      isSameBrand              # NEW: Clear boolean
      isEstimated              # NEW: Price confidence
      explanation              # NEW: Why this suggestion
      confidence
    }
  }
}
```

### Complete with Idempotency
```graphql
mutation {
  completeWizard(input: {
    sessionId: "wizard-123"
    idempotencyKey: "complete-123-v1"  # NEW: Prevent double-apply
  }) {
    success
    validationStatus           # NEW: Check for stale data
    staleItems {
      item {
        description
      }
      reason
    }
    updatedItems {
      id
      productMaster {
        name
      }
    }
  }
}
```

---

**Document Status:** Ready for Implementation
**Next Step:** Create database migrations for M1