# Shopping List Migration Enhancement - Analysis & Decision Document

**Date:** 2025-11-14
**Status:** Planning - Awaiting Decisions
**Stakeholder:** Product/Engineering Team

---

## Executive Summary

This document analyzes the proposed enhancement to the shopping list migration system: replacing automatic Product Master migration with user-driven interactive selection of top 3 similar products when flyer items expire.

**Current State:** System automatically migrates expired flyer products to Product Masters with varying confidence levels, no user interaction.

**Proposed State:** Present users with top 3 alternative products from current flyers and allow manual selection.

**Key Decision:** Is this the right trade-off between automation and user control?

---

## Current System Behavior

### How Shopping List Migration Works Today

**Architecture:**
```
Shopping List Item
  ├── LinkedProductID (int64, nullable)     → Points to specific flyer product (expires weekly)
  ├── ProductMasterID (int64, nullable)     → Points to canonical product (persists forever)
  ├── MatchingConfidence (float64)          → Score 0-1 of match quality
  └── AvailabilityStatus (string)           → 'available', 'migrated', 'unknown'
```

**Migration Flow** (see: `internal/services/shopping_list_migration_service.go`):

1. **Worker Runs Daily** (`internal/workers/shopping_list_migration_worker.go:41`)
   - Finds items with `product_master_id IS NULL AND linked_product_id IS NOT NULL`
   - These are items linked to expired flyer products

2. **Automatic Matching** (line 265-303)
   ```go
   // Try to match by linked product first
   if item.LinkedProductID != nil {
       master, err := findMasterByLinkedProduct(ctx, *item.LinkedProductID)
       if err == nil && master != nil {
           return master, 1.0, nil // Perfect match - confidence 1.0
       }
   }

   // Fall back to text search using normalized description
   masters, err := productMasterService.FindMatchingMasters(ctx, searchQuery, "", "")
   // Returns best match if confidence >= 0.5
   ```

3. **Confidence Scoring** (line 119-126)
   - **≥ 0.85:** Marked as "Successful Migration"
   - **< 0.85:** Marked as "Requires Review"
   - **< 0.50:** No match found, item unchanged
   - **No user action** required in any scenario

4. **Silent Update** (line 235-250)
   ```go
   // Item is automatically updated with new Product Master
   Set("product_master_id = ?", master.ID)
   Set("matching_confidence = ?", confidence)
   Set("availability_status = ?", "migrated")
   ```

### Matching Algorithm Quality

**Current Strategies** (`internal/services/matching/product_matcher.go`):

| Strategy | Weight | How It Works |
|----------|--------|--------------|
| **Exact Name Match** | 1.0 | Normalized name exact match or alternative name match → Score 0.95-1.0 |
| **Fuzzy Name Match** | 0.8 | Trigram similarity (threshold 0.7) → Score 0.7-1.0 |
| **Brand/Category** | 0.6 | Brand match (0.5) + Category match (0.3) + Word overlap (0.2) |
| **Barcode** | 1.0 | Not implemented (returns 0.0) |

**Composite Score Calculation:**
```
finalScore = (exactNameScore * 1.0 + fuzzyScore * 0.8 + brandCatScore * 0.6) / 2.4
```

**Weaknesses:**
- No price awareness
- No size/quantity matching
- No store preference
- No user preference learning
- Barcode matching not implemented

---

## Proposed Enhancement

### User-Driven Product Selection

**Core Idea:**
When a flyer product expires and the shopping list item needs migration:
1. Find top 3 similar Product Masters that have active flyer products
2. Present these options to the user
3. Let user select which one to use (or dismiss all)

**Key Phrase from Request:**
> "give user to choose from top 3 similar products which is **not available in new flyer list**"

**⚠️ CLARIFICATION NEEDED:** This phrasing is ambiguous. I believe you mean:
- Products that **ARE available** in current week's flyers
- NOT the expired product itself

---

## Critical Questions & Decisions Needed

### 🔴 DECISION 1: User Engagement Model

**Question:** How much user interaction is acceptable?

**Scenario Analysis:**

| User Has | Items Expiring/Week | Decisions Required | Time Investment |
|----------|---------------------|-------------------|-----------------|
| 1 list, 10 items | 7 items (70%) | 7 decisions | ~2 minutes |
| 1 list, 25 items | 18 items (72%) | 18 decisions | ~5 minutes |
| 3 lists, 30 items total | 22 items (73%) | 22 decisions | ~7 minutes |

**Trade-off:**
- ✅ **Pro:** User has full control, better product selection
- ❌ **Con:** Decision fatigue, abandoned lists, friction

**Options:**
- [ ] **A.** Require user selection for ALL migrations (no auto-migration)
- [ ] **B.** Auto-migrate high confidence (≥0.90), ask user for medium confidence (0.70-0.89)
- [ ] **C.** Auto-migrate ALL, but allow user to manually review/change after
- [ ] **D.** Something else: _______________________________

**Your Decision:** ________________

**Reasoning:** ________________

---

### 🔴 DECISION 2: "Available in New Flyers" Definition

**Question:** What constitutes an "available" product for suggestions?

**Options:**

**A. Any Active Flyer (All Stores)**
```
- Search: All Product Masters
- Filter: Must have at least one linked product in any active flyer
- Result: Maximum choice, but could suggest products from distant stores
```

**B. Same Store Priority**
```
- Search: All Product Masters
- Rank: Products from original store ranked higher
- Result: Balanced - prefers convenience but shows alternatives
```

**C. Same Store Only**
```
- Search: Only Product Masters with active products in original store
- Result: Limited choice, but optimized for single-store shopping
```

**D. Smart Store Selection (Based on User's Total List)**
```
- Analysis: Calculate optimal store combination for user's full list
- Filter: Only suggest products from stores in optimal route
- Result: Best total price/route, but complex
```

**Your Decision:** ________________

**Additional Constraints:**
- [ ] Only products published THIS week
- [ ] Include products from last week if still marked "active"
- [ ] Time-based: Only products valid for next 3+ days
- [ ] Other: _______________________________

---

### 🔴 DECISION 3: Ranking Criteria for Top 3

**Question:** How should we rank the suggested products?

**Current System Uses:**
- Text similarity (trigram matching)
- Brand/category match
- (No price, size, or store consideration)

**Proposed Ranking Factors (assign weights 0-10):**

| Factor | Current Weight | Your Weight | Notes |
|--------|----------------|-------------|-------|
| Text Similarity | 10 | ____ | How similar product names are |
| Brand Match | 6 | ____ | Same brand as original |
| Category Match | 3 | ____ | Same category (dairy, snacks, etc.) |
| **Price Similarity** | 0 | ____ | Within ±20% of original price |
| **Size/Quantity Match** | 0 | ____ | Same or similar package size |
| **Store Match** | 0 | ____ | Same store as original product |
| **Price (Absolute)** | 0 | ____ | Lowest price regardless of match |
| **Popularity** | 0 | ____ | Most commonly purchased product |
| **User History** | 0 | ____ | User has bought this before |

**Example Scenarios - Which should rank #1?**

**Scenario A: Milk Migration**
- Original: "Pienas Žemaitijos 2.5% 1L" @ Maxima, €1.50 (expired)
- Option 1: "Pienas Žemaitijos 2.5% 1L" @ Maxima, €1.60 (same brand, same store, +7% price)
- Option 2: "Pienas Žemaitijos 3.2% 1L" @ Maxima, €1.55 (same brand, different fat %, +3% price)
- Option 3: "Pienas Rokiškio 2.5% 1L" @ Rimi, €1.40 (different brand, different store, -7% price)

**Your Ranking:** 1st: ____ | 2nd: ____ | 3rd: ____

**Scenario B: Sour Cream Migration**
- Original: "Grietinė rūgštus 20% 400g Dvaro" @ Maxima, €2.20 (expired)
- Option 1: "Grietinė rūgštus 15% 330g Žemaitijos" @ Maxima, €1.80 (different brand/fat %/size, same store, -18% price)
- Option 2: "Grietinė rūgštus 20% 500g Pieno žvaigždės" @ Maxima, €2.60 (different brand, bigger size, +18% price)
- Option 3: "Varškės sūris 20% 400g Dvaro" @ Rimi, €2.10 (same brand, DIFFERENT product type, -5% price)

**Your Ranking:** 1st: ____ | 2nd: ____ | 3rd: ____

---

### 🔴 DECISION 4: No Good Match Found

**Question:** What if no products meet minimum criteria?

**Scenarios:**
- No Product Masters have active flyer products
- All matches have <0.50 confidence
- All prices are >50% more expensive than original
- Original was specialty/seasonal product

**Options:**
- [ ] **A.** Keep item as unlinked text (preserve user's intent, no product linking)
- [ ] **B.** Remove item from list automatically with notification
- [ ] **C.** Mark item as "needs attention" - user must manually search/replace
- [ ] **D.** Show top 3 anyway even with low confidence
- [ ] **E.** Suggest similar Product Masters even if NOT in active flyers
- [ ] **F.** Other: _______________________________

**Your Decision:** ________________

---

### 🔴 DECISION 5: User Non-Response Handling

**Question:** What happens if user doesn't review suggestions?

**Auto-Accept Deadline Options:**
- [ ] **A.** After 24 hours - auto-accept #1 ranked suggestion
- [ ] **B.** After 48 hours - auto-accept #1 ranked suggestion
- [ ] **C.** After 7 days - auto-accept #1 ranked suggestion
- [ ] **D.** Never auto-accept - keep item in "needs review" state indefinitely
- [ ] **E.** After deadline - remove item from list (assume user doesn't want it)
- [x ] **F.** Other: Leave it as is untill user changes this

**Your Decision:** ________________

**Notification Strategy:**
- [ ] Push notification when suggestions ready
- [ ] Push notification 24h before auto-accept
- [ ] Email digest of pending migrations weekly
- [x ] In-app badge only
- [ ] Other: _______________________________

---

### 🔴 DECISION 6: Multi-Store Shopping Strategy

**Context:** From `docs/checklist:16` - System already recommends where to buy items cheapest.

**Conflict:**
User added "Milk" from Store A @ €2.50 (now expired).

**New Options:**
1. Store A: Different brand milk @ €2.80
2. Store B: Same brand milk @ €2.30 (saves €0.20)
3. Store C: Generic milk @ €1.90 (saves €0.60)

**Question:** Should migration suggestions prioritize:

- [ ] **A.** Same store (minimize trips, route optimization)
- [ ] **B.** Cheapest price per item (maximize savings)
- [ ] **C.** Cheapest total basket across optimal store combination
- [ ] **D.** User's most frequently shopped stores
- [ ] **E.** Let user configure preference in settings
- [ ] **F.** Other: We should have user to decide globally what are their preferences

**Your Decision:** ________________

**Follow-up:** Should we recalculate optimal store recommendations AFTER user selects migrations?
- [x ] Yes - show "You can save €3.50 by switching 4 items to Rimi"
- [ ] No - respect user's selections
- [ ] Other: _______________________________

---

### 🔴 DECISION 7: Confidence Tiers & Automation

**Question:** Should some migrations happen automatically without user input?

**Proposed Tiers:**

| Tier | Confidence | Behavior | Example |
|------|-----------|----------|---------|
| **High** | ≥ 0.90 | Auto-migrate silently | Same product master has exact product in new flyer |
| **Medium** | 0.70-0.89 | Ask user to select from top 3 | Similar products, same category/brand |
| **Low** | 0.50-0.69 | Mark "needs attention" | Weak matches, different category |
| **None** | < 0.50 | No suggestion | No reasonable match |

**Do you agree with this approach?**
- [ ] Yes - use tiered system
- [ ] No - always ask user (even for 0.95 confidence)
- [ ] No - always auto-migrate (never ask user)
- [ ] Modify tiers: Ask user when confidence is medium on high case auto migrate, this will make less work for user each week

**Your Decision:** ________________

---

### 🔴 DECISION 8: User Preference Learning

**Question:** Should the system learn from user's selections over time?

**Example:**
User consistently selects cheapest option regardless of brand → System adjusts ranking to prioritize price.

**Trackable Patterns:**
- Brand loyalty (always picks same brand)
- Price sensitivity (always picks cheapest)
- Size preference (prefers larger/smaller packages)
- Store loyalty (prefers specific stores)
- Quality preference (picks premium brands)

**Options:**
- [ ] **A.** Implement ML-based preference learning (Phase 2)
- [ ] **B.** Simple rules: "User picked cheap option 5 times → boost price weight"
- [ ] **C.** Don't implement - keep ranking static
- [ x] **D.** Let users explicitly set preferences in settings

**Your Decision:** ________________

**If yes, when:**
- [ x] MVP (Phase 1)
- [ ] Later iteration (Phase 2-3)
- [ ] Future consideration

---

### 🔴 DECISION 9: Batch Review vs. Item-by-Item

**Question:** How should users review migration suggestions?

**Option A: Batch Review Screen**
```
┌─────────────────────────────────────┐
│ Review 8 Product Changes            │
├─────────────────────────────────────┤
│ 1. Milk (3 options) [View >]        │
│ 2. Bread (3 options) [View >]       │
│ 3. Butter (Auto-migrated ✓)         │
│ ...                                  │
│                                      │
│ [Accept All] [Review Each]          │
└─────────────────────────────────────┘
```
- User sees all pending migrations at once
- Can bulk accept trusted items
- Drill into specific items to choose

**Option B: Item-by-Item Notification**
```
Notification: "Milk needs update"
→ User opens app → Sees 3 options for Milk only
→ Selects one
Next notification: "Bread needs update"
```
- Less overwhelming
- More friction (multiple interactions)

**Option C: Inline in Shopping List**
```
Shopping List:
☐ Eggs
☐ Milk ⚠️ [Product expired - Choose replacement >]
☐ Bread
```
- Context-aware (see full list while choosing)
- Clutters main shopping list view

**Your Preference:** ________________

---

### 🔴 DECISION 10: Mobile vs. Web Experience

**Question:** Where will users interact with this feature?

**Platform Priority:**
- [X ] Mobile app (iOS/Android) - Primary
- [ X] Web app - Primary
- [ X] Both equally
- [ ] Mobile only (no web implementation)

**Mobile Considerations:**
- Swipeable cards for suggestions?
- Quick actions (swipe right to accept #1)?
- Offline support (cache suggestions)?

**Are frontend teams aligned on this feature?**
- [ ] Yes - already discussed
- [ ] No - needs frontend planning
- [ ] Unknown

**Your Answer:** Front end is not in this scope

---

## Alternative Approach: Hybrid Smart Migration

Instead of forcing user decisions on every migration, consider a **hybrid approach**:

### Tier 1: High Confidence Auto-Migration (≥0.90)
**No user interaction needed**

**Criteria:**
- Same Product Master ID
- Product Master has active flyer product(s)
- Same brand, same category
- Size within ±10%
- Price within ±15%

**Example:**
- Old: "Pienas Žemaitijos 2.5% 1L" @ €1.50
- New: "Pienas Žemaitijos 2.5% 1L" @ €1.55
- **Action:** Auto-migrate silently ✓

**Notification:**
"We updated Milk to current flyer (€1.55)" - dismissible banner

---

### Tier 2: Medium Confidence - Batch Review (0.70-0.89)
**User reviews once per week**

**Criteria:**
- Similar Product Master (different brand/size)
- Active flyer product exists
- Price within ±30%
- Same category

**User Experience:**
- Single "Review Changes" screen with all medium-confidence items
- Show old → new comparison
- Allow bulk actions: "Accept all from Store X"
- Default: Auto-accept #1 suggestion after 48 hours if no response

**Example:**
- 8 items need review
- User reviews in 2 minutes
- Rest auto-accept after deadline

---

### Tier 3: Low Confidence - Mark for Attention (< 0.70)
**Don't force bad matches**

**Criteria:**
- Weak text similarity
- Different category or brand
- No active flyer products
- Price difference >30%

**User Experience:**
- Item marked with ⚠️ icon in shopping list
- Notification: "3 items need updating"
- User can:
  - Search manually for replacement
  - Remove from list
  - Keep as text-only item (no product linking)

**Example:**
- Old: "Organic almond butter 500g"
- No good matches found
- **Action:** Don't auto-migrate to "regular peanut butter"

---

**Do you prefer this hybrid approach?**
- [ ] Yes - implement tiered system
- [ ] No - stick with original proposal (always show top 3)
- [ ] Maybe - need to see mockups first

**Your Decision:** ________________

---

## Technical Implementation Roadmap

### Phase 1: Foundation (Weeks 1-2)

#### 1.1 Enhanced Product Search Service

**Create:** `internal/services/migration/suggestion_service.go`

```go
type SuggestionService interface {
    // Find top N product suggestions for a shopping list item
    FindSuggestions(ctx context.Context, item *ShoppingListItem, limit int) ([]*ProductSuggestion, error)

    // Create migration suggestions for expired items
    CreateSuggestionsForExpiredItems(ctx context.Context, listID int64) ([]*MigrationSuggestion, error)

    // Apply user's selection
    ApplyUserSelection(ctx context.Context, suggestionID int64, selectedProductMasterID int64) error
}
```

**Responsibilities:**
- Search Product Masters by similarity
- Filter to only masters with active flyer products
- Rank by configured criteria (see Decision 3)
- Generate explanations for suggestions

---

#### 1.2 New Database Models

**Table: `migration_suggestions`**
```sql
CREATE TABLE migration_suggestions (
    id BIGSERIAL PRIMARY KEY,
    shopping_list_item_id BIGINT NOT NULL REFERENCES shopping_list_items(id) ON DELETE CASCADE,
    original_product_id BIGINT REFERENCES products(id) ON DELETE SET NULL,
    original_product_master_id BIGINT REFERENCES product_masters(id) ON DELETE SET NULL,

    -- Suggestions (JSON array of suggested product master IDs with metadata)
    suggestions JSONB NOT NULL, -- Array of ProductSuggestion objects

    -- State
    state VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending, user_selected, auto_accepted, dismissed

    -- Selection
    selected_product_master_id BIGINT REFERENCES product_masters(id) ON DELETE SET NULL,
    selected_at TIMESTAMP,
    selection_method VARCHAR(50), -- 'user', 'auto_accept', 'fallback'

    -- Metadata
    migration_reason VARCHAR(100) NOT NULL, -- 'product_expired', 'low_confidence', 'price_improvement'
    requires_user_action BOOLEAN NOT NULL DEFAULT false,
    confidence_tier VARCHAR(20), -- 'high', 'medium', 'low'

    -- Auto-accept logic
    auto_accept_at TIMESTAMP,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP, -- After this, suggestion is no longer valid

    -- Indexes
    INDEX idx_migration_suggestions_item_id (shopping_list_item_id),
    INDEX idx_migration_suggestions_state (state),
    INDEX idx_migration_suggestions_auto_accept (auto_accept_at) WHERE state = 'pending'
);
```

**Go Models:**
```go
// internal/models/migration_suggestion.go
type MigrationSuggestion struct {
    ID                       int64                `bun:"id,pk,autoincrement"`
    ShoppingListItemID       int64                `bun:"shopping_list_item_id,notnull"`
    OriginalProductID        *int64               `bun:"original_product_id"`
    OriginalProductMasterID  *int64               `bun:"original_product_master_id"`

    Suggestions              []*ProductSuggestion `bun:"suggestions,type:jsonb"`

    State                    string               `bun:"state,notnull,default:'pending'"`
    SelectedProductMasterID  *int64               `bun:"selected_product_master_id"`
    SelectedAt               *time.Time           `bun:"selected_at"`
    SelectionMethod          *string              `bun:"selection_method"`

    MigrationReason          string               `bun:"migration_reason,notnull"`
    RequiresUserAction       bool                 `bun:"requires_user_action,notnull,default:false"`
    ConfidenceTier           *string              `bun:"confidence_tier"`

    AutoAcceptAt             *time.Time           `bun:"auto_accept_at"`

    CreatedAt                time.Time            `bun:"created_at,notnull,default:now()"`
    UpdatedAt                time.Time            `bun:"updated_at,notnull,default:now()"`
    ExpiresAt                *time.Time           `bun:"expires_at"`

    // Relations
    ShoppingListItem         *ShoppingListItem    `bun:"rel:belongs-to,join:shopping_list_item_id=id"`
    OriginalProduct          *Product             `bun:"rel:belongs-to,join:original_product_id=id"`
    OriginalProductMaster    *ProductMaster       `bun:"rel:belongs-to,join:original_product_master_id=id"`
    SelectedProductMaster    *ProductMaster       `bun:"rel:belongs-to,join:selected_product_master_id=id"`
}

type ProductSuggestion struct {
    ProductMasterID      int64                `json:"product_master_id"`
    Ranking              int                  `json:"ranking"` // 1, 2, or 3

    // Active flyer product details
    ActiveFlyerProduct   *FlyerProductInfo    `json:"active_flyer_product"`

    // Scoring
    SimilarityScore      float64              `json:"similarity_score"`
    CompositeScore       float64              `json:"composite_score"`

    // Comparison to original
    PriceDelta           *float64             `json:"price_delta"` // Percentage change
    PriceDeltaAbsolute   *float64             `json:"price_delta_absolute"` // Euro change
    SizeDelta            *string              `json:"size_delta"` // "+100g", "-50ml"

    // Explanation
    Explanation          string               `json:"explanation"` // "Same brand, larger size, 5% cheaper"
    MatchReasons         []string             `json:"match_reasons"` // ["brand_match", "category_match", "size_similar"]
}

type FlyerProductInfo struct {
    ProductID            int64                `json:"product_id"`
    FlyerID              int                  `json:"flyer_id"`
    StoreID              int                  `json:"store_id"`
    StoreName            string               `json:"store_name"`
    Price                float64              `json:"price"`
    PriceUnit            *string              `json:"price_unit"`
    ValidUntil           *time.Time           `json:"valid_until"`
}

// State constants
const (
    MigrationStatePending      = "pending"
    MigrationStateUserSelected = "user_selected"
    MigrationStateAutoAccepted = "auto_accepted"
    MigrationStateDismissed    = "dismissed"
)

// Reason constants
const (
    MigrationReasonProductExpired    = "product_expired"
    MigrationReasonLowConfidence     = "low_confidence"
    MigrationReasonPriceImprovement  = "price_improvement"
    MigrationReasonBetterMatch       = "better_match"
)

// Tier constants
const (
    ConfidenceTierHigh   = "high"
    ConfidenceTierMedium = "medium"
    ConfidenceTierLow    = "low"
)
```

---

#### 1.3 Migration Worker Updates

**Modify:** `internal/workers/shopping_list_migration_worker.go`

**Current Behavior:**
- Finds expired items
- Auto-migrates to Product Masters

**New Behavior:**
- Finds expired items
- Creates `MigrationSuggestion` records based on confidence tier
- High confidence: Auto-migrates immediately (if Decision 7 = tiered)
- Medium/Low confidence: Creates pending suggestions

**New Worker:** `internal/workers/migration_auto_accept_worker.go`
- Runs every 6 hours
- Finds suggestions where `auto_accept_at < NOW() AND state = 'pending'`
- Auto-applies #1 ranked suggestion
- Updates state to `auto_accepted`

---

### Phase 2: Ranking & Intelligence (Weeks 3-4)

#### 2.1 Configurable Ranking Engine

**Create:** `internal/services/migration/ranking_engine.go`

```go
type RankingConfig struct {
    Weights RankingWeights
    Filters RankingFilters
}

type RankingWeights struct {
    TextSimilarity   float64 // Default: 10
    BrandMatch       float64 // Default: 6
    CategoryMatch    float64 // Default: 3
    PriceSimilarity  float64 // Default: 5 (see Decision 3)
    SizeMatch        float64 // Default: 4
    StoreMatch       float64 // Default: 3
    PriceAbsolute    float64 // Default: 2
    Popularity       float64 // Default: 1
    UserHistory      float64 // Default: 0 (Phase 2)
}

type RankingFilters struct {
    MaxPriceDeltaPercent float64  // Don't suggest if price > +30%
    SameStoreOnly        bool     // Only suggest from original store
    MinConfidence        float64  // Minimum similarity score
    ActiveFlyersOnly     bool     // Only products in active flyers
    ExcludeStores        []int    // User-blocked stores
}

type RankingEngine struct {
    config RankingConfig
}

func (e *RankingEngine) RankSuggestions(
    ctx context.Context,
    originalItem *ShoppingListItem,
    candidates []*ProductMaster,
) ([]*ProductSuggestion, error) {
    // Score each candidate
    // Apply filters
    // Sort by composite score
    // Return top N
}
```

**Configuration Source (see Decision 3):**
- [ ] Hard-coded defaults
- [ ] Environment variables
- [ ] Database settings table
- [ ] User preferences (per-user customization)

---

#### 2.2 Explanation Generator

**Create:** `internal/services/migration/explanation_generator.go`

```go
func GenerateExplanation(original *ShoppingListItem, suggestion *ProductSuggestion) string {
    reasons := []string{}

    if suggestion.BrandMatch {
        reasons = append(reasons, "same brand")
    }
    if suggestion.SizeDelta == "" {
        reasons = append(reasons, "same size")
    } else {
        reasons = append(reasons, suggestion.SizeDelta)
    }
    if suggestion.PriceDelta < 0 {
        reasons = append(reasons, fmt.Sprintf("%.0f%% cheaper", abs(suggestion.PriceDelta)))
    }
    if suggestion.StoreMatch {
        reasons = append(reasons, "same store")
    }

    return strings.Join(reasons, ", ")
}
```

**Example Outputs:**
- "Same brand, same size, 5% cheaper"
- "Different brand, larger size (+100g), same price"
- "Same brand, same store, 10% more expensive"

---

### Phase 3: GraphQL API (Weeks 5-6)

#### 3.1 Schema Extensions

**File:** `internal/graphql/schema/migration.graphql`

```graphql
# ============================================
# Types
# ============================================

"""
Represents a migration suggestion for a shopping list item
when its linked product expires or needs updating
"""
type MigrationSuggestion {
  id: ID!

  """The shopping list item that needs migration"""
  shoppingListItem: ShoppingListItem!

  """The original flyer product (may be expired)"""
  originalProduct: Product

  """The original product master"""
  originalProductMaster: ProductMaster

  """Top ranked product suggestions"""
  suggestions: [ProductSuggestion!]!

  """Current state of the suggestion"""
  state: MigrationState!

  """Selected product master (if user has chosen)"""
  selectedProductMaster: ProductMaster

  """When the selection was made"""
  selectedAt: DateTime

  """How the selection was made"""
  selectionMethod: SelectionMethod

  """Why this migration was triggered"""
  migrationReason: MigrationReason!

  """Whether user action is required"""
  requiresUserAction: Boolean!

  """Confidence tier of auto-migration"""
  confidenceTier: ConfidenceTier

  """When this will auto-accept if no user action"""
  autoAcceptAt: DateTime

  """When this suggestion expires"""
  expiresAt: DateTime

  createdAt: DateTime!
  updatedAt: DateTime!
}

"""
A suggested product as replacement for an expired item
"""
type ProductSuggestion {
  """The suggested product master"""
  productMaster: ProductMaster!

  """Ranking position (1 = best match)"""
  ranking: Int!

  """Current active flyer product for this master"""
  activeFlyerProduct: FlyerProductInfo

  """How similar to original (0-1)"""
  similarityScore: Float!

  """Overall composite score including price, brand, etc."""
  compositeScore: Float!

  """Price difference as percentage (negative = cheaper)"""
  priceDelta: Float

  """Price difference in currency"""
  priceDeltaAbsolute: Float

  """Size difference description"""
  sizeDelta: String

  """Human-readable explanation"""
  explanation: String!

  """Reasons for the match"""
  matchReasons: [String!]!
}

"""
Active flyer product information
"""
type FlyerProductInfo {
  product: Product!
  flyer: Flyer!
  store: Store!
  price: Float!
  priceUnit: String
  validUntil: DateTime
}

# ============================================
# Enums
# ============================================

enum MigrationState {
  PENDING
  USER_SELECTED
  AUTO_ACCEPTED
  DISMISSED
}

enum SelectionMethod {
  USER
  AUTO_ACCEPT
  FALLBACK
}

enum MigrationReason {
  PRODUCT_EXPIRED
  LOW_CONFIDENCE
  PRICE_IMPROVEMENT
  BETTER_MATCH
}

enum ConfidenceTier {
  HIGH
  MEDIUM
  LOW
}

# ============================================
# Inputs
# ============================================

input MigrationAcceptanceFilters {
  """Only accept suggestions from specific stores"""
  storeIds: [ID!]

  """Only accept if price delta is within range"""
  maxPriceDeltaPercent: Float

  """Only accept specific confidence tiers"""
  confidenceTiers: [ConfidenceTier!]

  """Auto-accept highest ranked suggestions"""
  autoAcceptRank1: Boolean
}

# ============================================
# Queries
# ============================================

extend type Query {
  """
  Get pending migration suggestions for a shopping list
  """
  pendingMigrations(listId: ID!): [MigrationSuggestion!]!

  """
  Get migration history (completed/dismissed) for a list
  """
  migrationHistory(
    listId: ID!
    limit: Int = 20
    offset: Int = 0
  ): MigrationHistoryConnection!

  """
  Get a specific migration suggestion by ID
  """
  migrationSuggestion(id: ID!): MigrationSuggestion

  """
  Get summary stats for migrations
  """
  migrationStats(listId: ID!): MigrationStats!
}

type MigrationHistoryConnection {
  nodes: [MigrationSuggestion!]!
  totalCount: Int!
  hasMore: Boolean!
}

type MigrationStats {
  totalPending: Int!
  totalAutoAccepted: Int!
  totalUserSelected: Int!
  totalDismissed: Int!
  avgConfidence: Float!
}

# ============================================
# Mutations
# ============================================

extend type Mutation {
  """
  User selects a specific suggestion for a migration
  """
  selectMigrationSuggestion(
    suggestionId: ID!
    selectedProductMasterId: ID!
  ): ShoppingListItem!

  """
  User dismisses a migration suggestion (keep item as-is)
  """
  dismissMigrationSuggestion(
    suggestionId: ID!
  ): Boolean!

  """
  Accept all pending migrations for a list with optional filters
  """
  acceptAllMigrations(
    listId: ID!
    filters: MigrationAcceptanceFilters
  ): AcceptAllMigrationsResult!

  """
  Manually trigger migration suggestion generation for a single item
  """
  createMigrationSuggestion(
    itemId: ID!
  ): MigrationSuggestion

  """
  Manually trigger migration for entire list
  """
  triggerListMigration(
    listId: ID!
  ): TriggerListMigrationResult!
}

type AcceptAllMigrationsResult {
  acceptedCount: Int!
  dismissedCount: Int!
  updatedItems: [ShoppingListItem!]!
}

type TriggerListMigrationResult {
  suggestionsCreated: Int!
  autoMigratedCount: Int!
  requiresReviewCount: Int!
}
```

---

#### 3.2 Resolver Implementation

**Create:** `internal/graphql/resolvers/migration_suggestion.go`

```go
// Query: pendingMigrations
func (r *queryResolver) PendingMigrations(ctx context.Context, listID string) ([]*models.MigrationSuggestion, error) {
    listIDInt, err := parseID(listID)
    if err != nil {
        return nil, err
    }

    // Check authorization (user must own the list)
    userID := auth.GetUserIDFromContext(ctx)
    list, err := r.ShoppingListService.GetByID(ctx, listIDInt)
    if err != nil {
        return nil, err
    }
    if list.UserID != userID {
        return nil, apperrors.Unauthorized("you don't have access to this list")
    }

    // Get pending suggestions
    suggestions, err := r.MigrationSuggestionService.GetPendingByListID(ctx, listIDInt)
    if err != nil {
        return nil, err
    }

    return suggestions, nil
}

// Mutation: selectMigrationSuggestion
func (r *mutationResolver) SelectMigrationSuggestion(
    ctx context.Context,
    suggestionID string,
    selectedProductMasterID string,
) (*models.ShoppingListItem, error) {
    suggestionIDInt, err := parseID(suggestionID)
    if err != nil {
        return nil, err
    }

    productMasterIDInt, err := parseID(selectedProductMasterID)
    if err != nil {
        return nil, err
    }

    // Apply selection
    item, err := r.MigrationSuggestionService.ApplyUserSelection(
        ctx,
        suggestionIDInt,
        productMasterIDInt,
    )
    if err != nil {
        return nil, err
    }

    return item, nil
}

// Mutation: acceptAllMigrations
func (r *mutationResolver) AcceptAllMigrations(
    ctx context.Context,
    listID string,
    filters *MigrationAcceptanceFilters,
) (*AcceptAllMigrationsResult, error) {
    listIDInt, err := parseID(listID)
    if err != nil {
        return nil, err
    }

    result, err := r.MigrationSuggestionService.AcceptAllPending(
        ctx,
        listIDInt,
        filters,
    )
    if err != nil {
        return nil, err
    }

    return result, nil
}
```

---

### Phase 4: Testing & Optimization (Weeks 7-8)

#### 4.1 Unit Tests

**Test Files to Create:**
- `internal/services/migration/suggestion_service_test.go`
- `internal/services/migration/ranking_engine_test.go`
- `internal/repositories/migration_suggestion_repository_test.go`

**Test Scenarios:**
```go
func TestSuggestionService_FindSuggestions(t *testing.T) {
    tests := []struct {
        name           string
        item           *ShoppingListItem
        expectedCount  int
        expectedRank1  string
    }{
        {
            name: "exact match available in new flyer",
            item: &ShoppingListItem{
                Description: "Pienas Žemaitijos 2.5% 1L",
                LinkedProductID: ptr(123),
            },
            expectedCount: 3,
            expectedRank1: "Pienas Žemaitijos 2.5% 1L", // Same product in new flyer
        },
        {
            name: "similar products different brands",
            item: &ShoppingListItem{
                Description: "Grietinė 20% 400g",
            },
            expectedCount: 3,
            expectedRank1: "", // Test assertion varies based on ranking config
        },
        {
            name: "no active flyer products",
            item: &ShoppingListItem{
                Description: "Organic quinoa 500g",
            },
            expectedCount: 0,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

---

#### 4.2 Integration Tests

**Create:** `tests/integration/migration_flow_test.go`

```go
func TestMigrationFlow_FullCycle(t *testing.T) {
    // Setup: Create user, list, add items linked to flyer products
    // Expire flyer (mark as inactive)
    // Trigger migration worker
    // Assert: MigrationSuggestions created with correct state
    // Simulate user selection via GraphQL mutation
    // Assert: ShoppingListItem updated with new ProductMasterID
    // Assert: MigrationSuggestion state = 'user_selected'
}

func TestMigrationFlow_AutoAccept(t *testing.T) {
    // Setup: Create pending suggestion with autoAcceptAt in past
    // Trigger auto-accept worker
    // Assert: Item migrated to rank 1 suggestion
    // Assert: MigrationSuggestion state = 'auto_accepted'
}
```

---

#### 4.3 Performance Optimization

**Database Indexes:**
```sql
-- Fast lookup of pending suggestions by list
CREATE INDEX idx_migration_suggestions_list_state
ON migration_suggestions(shopping_list_item_id, state)
WHERE state = 'pending';

-- Auto-accept worker query optimization
CREATE INDEX idx_migration_suggestions_auto_accept
ON migration_suggestions(auto_accept_at)
WHERE state = 'pending' AND auto_accept_at IS NOT NULL;

-- Fast product master search with active flyer products
CREATE INDEX idx_products_master_flyer_active
ON products(product_master_id, flyer_id)
WHERE product_master_id IS NOT NULL;
```

**Caching Strategy:**
```go
// Cache active flyer product IDs for 1 hour
type ActiveFlyerProductCache struct {
    redis *redis.Client
    ttl   time.Duration
}

func (c *ActiveFlyerProductCache) GetActiveProductMasterIDs(ctx context.Context) ([]int64, error) {
    // Check Redis cache
    // If miss, query DB and cache result
}
```

**Query Optimization:**
```go
// Instead of N+1 queries for each suggestion's flyer product
// Use dataloader pattern

type FlyerProductLoader struct {
    db *bun.DB
}

func (l *FlyerProductLoader) LoadBatch(ctx context.Context, productMasterIDs []int64) ([]*Product, error) {
    // Single query to get all active flyer products for all masters
    var products []*Product
    err := l.db.NewSelect().
        Model(&products).
        Where("product_master_id IN (?)", bun.In(productMasterIDs)).
        Join("JOIN flyers f ON f.id = product.flyer_id").
        Where("f.status = ?", "active").
        Order("product.price ASC"). // Cheapest first
        Scan(ctx)

    return products, err
}
```

---

## Implementation Phases Summary

| Phase | Duration | Deliverables | Dependencies |
|-------|----------|--------------|--------------|
| **Phase 1: Foundation** | 2 weeks | Database schema, models, basic suggestion service | Decisions 1-7 finalized |
| **Phase 2: Intelligence** | 2 weeks | Ranking engine, explanation generator, confidence tiers | Phase 1, Decision 3 (weights) |
| **Phase 3: API** | 2 weeks | GraphQL schema, resolvers, mutations | Phase 2, Decision 9 (UX flow) |
| **Phase 4: Testing** | 2 weeks | Unit tests, integration tests, performance optimization | Phase 3 |
| **Phase 5: Frontend** | 4 weeks | Mobile/web UI (not covered in backend roadmap) | Phase 3 API complete |

**Total Backend Effort:** ~8 weeks (2 engineers) or ~4 weeks (4 engineers)

---

## Risks & Mitigation

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| **Users ignore suggestions** | High | Medium | Implement auto-accept with notifications |
| **Suggestion quality poor** | High | Medium | Start with high confidence tier only, iterate |
| **Performance degradation** | Medium | Low | Implement caching, optimize queries, use dataloaders |
| **Frontend not aligned** | High | Medium | Get frontend buy-in BEFORE starting backend work |
| **Scope creep (ML, preferences)** | Medium | High | Stick to MVP, defer ML to Phase 2+ |
| **User confusion (too many options)** | Medium | Medium | Start with batch review, A/B test UX approaches |

---

## Success Metrics

**After 1 Month of Deployment:**

| Metric | Target | How to Measure |
|--------|--------|----------------|
| **Suggestion Acceptance Rate** | >60% | `(user_selected + auto_accepted) / total_suggestions` |
| **User Engagement Rate** | >40% | `users_who_reviewed / users_with_suggestions` |
| **High Confidence Auto-Migration Accuracy** | >95% | Sample user feedback, complaints |
| **Avg Time to Review** | <3 minutes | `selected_at - created_at` for user_selected |
| **List Abandonment Rate** | <5% increase | Compare before/after feature launch |

**Red Flags (Re-evaluate Feature):**
- Acceptance rate <30%
- Engagement rate <15%
- List abandonment >20% increase

---

## Open Questions for Engineering

1. **Should we version the suggestion algorithm?**
   - Store algorithm version in `migration_suggestions.algorithm_version`
   - Allows A/B testing different ranking strategies

2. **How long to retain suggestion history?**
   - [ ] 30 days
   - [ ] 90 days
   - [ ] Forever (for ML training)

3. **Should dismissed suggestions prevent future similar suggestions?**
   - Example: User dismisses "Brand X milk" → Never suggest Brand X again for milk?

4. **Notification delivery mechanism?**
   - [ ] Push notifications (requires mobile app integration)
   - [ ] Email digest
   - [ ] SMS
   - [ ] In-app only

---

## Your Decisions Checklist

Before implementation begins, please provide answers to:

- [ ] Decision 1: User Engagement Model (Option A/B/C/D)
- [ ] Decision 2: "Available in New Flyers" Definition (Option A/B/C/D)
- [ ] Decision 3: Ranking Criteria Weights (Fill in table)
- [ ] Decision 4: No Good Match Found (Option A/B/C/D/E/F)
- [ ] Decision 5: User Non-Response Handling (Option A/B/C/D/E/F)
- [ ] Decision 6: Multi-Store Shopping Strategy (Option A/B/C/D/E/F)
- [ ] Decision 7: Confidence Tiers & Automation (Yes/No/Modify)
- [ ] Decision 8: User Preference Learning (Option A/B/C/D)
- [ ] Decision 9: Batch Review vs. Item-by-Item (Option A/B/C)
- [ ] Decision 10: Mobile vs. Web Experience (Fill in)

**Hybrid Approach Preference:** Yes / No / Maybe

---

## Recommended Next Steps

1. **Review this document** and fill in all decision sections
2. **Schedule alignment meeting** with product, engineering, and design teams
3. **Create UI mockups** based on Decision 9 (get frontend input)
4. **Validate with users** - show mockups to 5-10 users, get feedback
5. **Finalize technical approach** (original proposal vs. hybrid)
6. **Create detailed technical specification** based on decisions
7. **Begin Phase 1 implementation**

---

## Contact & Discussion

For questions or clarifications on this analysis:
- Technical Architecture: [System Architect]
- Product Decisions: [Product Manager]
- User Research: [UX Team]

---

**Document Version:** 1.0
**Last Updated:** 2025-11-14
**Status:** Awaiting Stakeholder Input

IMPORTANT All decisions are made, we focus only on backend side

Vision of frotnend it must satisfy
ALL these rules has to be setisfied from Backend side:

User has items in shopping list
Items are expired
System suggest to repopulate items from current weeks flyers
System prompt user to a wizard like form
User can select filters by which items are being suggested shop, brand, cheapest, premium
User can see 3-5 simmilar options, selects one
Items are being repopulted as active in shopping list

