# Kainuguru MVP - Final Production Specification
## 🔴 SOURCE OF TRUTH - Single Implementation Guide

**Version**: 1.0 Final
**Date**: November 2024
**Status**: ACTIVE - Use this document for all development
**Focus**: Simple MVP without over-engineering

---

## Executive Summary

**🎯 Core Goal**: Build a functional grocery flyer aggregation system that:
1. Scrapes flyers from Lithuanian stores (IKI, Maxima, Rimi)
2. Extracts product information using ChatGPT/GPT-4 Vision
3. Provides searchable database using PostgreSQL FTS
4. Allows users to create smart shopping lists that persist across weeks
5. Uses GraphQL API (gqlgen) from the start

This specification addresses the real-world challenges:
- Product data is **inconsistent** and extracted from images via AI
- Same products appear with **different names** each week
- **No consistent product IDs** exist across flyers
- Lithuanian language requires **special character handling**
- Shopping lists need to work despite **weekly data changes**
- **NEW**: Race condition handling for concurrent scrapers
- **NEW**: Simple error handling when extraction fails

---

## 1. Critical Challenges & Solutions

### 1.0 Data Consistency & Race Conditions (NEW)
**Problem**: Multiple scrapers updating same flyer simultaneously

**Solution**: Distributed locking with Redis
```go
type DistributedLock struct {
    redis *redis.Client
    key   string
    ttl   time.Duration
}

func (s *ScraperService) ProcessFlyer(flyerID int64) error {
    lock := NewDistributedLock(s.redis, fmt.Sprintf("flyer:%d", flyerID), 5*time.Minute)
    if !lock.Acquire() {
        return ErrFlyerBeingProcessed
    }
    defer lock.Release()

    // Process flyer safely...
    return nil
}
```

### 1.1 Product Identity Crisis
**Problem**: No consistent product identification across weeks/stores
```
Week 1: "Pienas ŽEMAITIJOS 2.5% 1L"
Week 2: "ŽEMAITIJOS pienas 2,5% 1l"
Week 3: "Pienas Žemaitijos 2.5"
```

**Solution**: Multi-layer identification system
```sql
-- Product Master Table (NEW)
CREATE TABLE product_masters (
    id BIGSERIAL PRIMARY KEY,
    canonical_name VARCHAR(255), -- "Žemaitijos pienas 2.5%"
    brand_id BIGINT REFERENCES brands(id),
    product_type VARCHAR(100), -- "milk"
    tags TEXT[], -- Array of tags
    attributes JSONB, -- {"fat_content": "2.5", "volume": "1L"}
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Product Instances (weekly occurrences)
CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    master_id BIGINT REFERENCES product_masters(id), -- Nullable initially
    flyer_id BIGINT NOT NULL REFERENCES flyers(id),
    raw_name VARCHAR(255) NOT NULL, -- Exactly as extracted
    normalized_name VARCHAR(255), -- For search
    -- ... rest of fields
);

-- Fuzzy Matching Index
CREATE INDEX idx_products_normalized_trgm
ON products USING gin(normalized_name gin_trgm_ops);
```

### 1.2 Smart Shopping List Persistence
**Problem**: Shopping list items become orphaned when flyers expire
```yaml
User adds: "Milk" → Links to Product ID 123 (Week 1)
Week 2: Product ID 123 no longer exists
Result: Broken shopping list
```

**Solution**: Tag-based shopping lists with automatic migration
```go
type ShoppingListItem struct {
    ID           int64
    ListID       int64

    // Three-tier identification
    ProductID    *int64  // Current specific product
    MasterID     *int64  // Product master reference
    TagID        *int64  // Generic tag (milk, bread)
    CustomName   string  // User's original input

    // Smart matching
    LastMatchedAt *time.Time
    MatchHistory  []int64 // Previous product IDs
}

func (s *ShoppingListService) GetCurrentMatches(item *ShoppingListItem) []Product {
    // Priority order:
    // 1. Try exact product (if still active)
    // 2. Try product master matches
    // 3. Try tag matches
    // 4. Fuzzy search custom name
    return s.findBestMatches(item)
}

// Automatic migration for expired items
func (s *ShoppingListService) MigrateExpiredItems() error {
    items := s.GetItemsWithExpiredProducts()

    for _, item := range items {
        matches := s.FindMatches(item)
        bestMatch := s.ScoreBestMatch(matches)

        if bestMatch.Score > 0.8 {
            // High confidence - auto-migrate
            item.ProductID = &bestMatch.ProductID
            item.LastMatchedAt = time.Now()
            s.db.Save(item)
        } else if bestMatch.Score > 0.5 {
            // Medium confidence - notify user
            s.NotifyUserForConfirmation(item, matches)
        }
        // Low confidence - keep as-is with tag fallback
    }
    return nil
}
```

### 1.3 AI Extraction Inconsistency
**Problem**: ChatGPT/GPT-4 Vision returns different formats each time

**Solution**: Simple extraction with validation
```go
type ProductExtractor struct {
    ai        *ChatGPTExtractor  // Use ChatGPT for MVP simplicity
    validator *DataValidator
    // TODO: Add cost tracking in production
    // TODO: Add fallback extractors when costs are known
}

func (e *ProductExtractor) Extract(image []byte) ([]Product, error) {
    // Simple ChatGPT extraction for MVP
    // TODO: Add cost monitoring after MVP launch
    aiProducts, err := e.ai.Extract(image)
    if err != nil {
        log.Error("ChatGPT extraction failed", err)
        // TODO: Implement fallback strategies in v2
        return []Product{}, err
    }

    // Validate and clean
    validated := []Product{}
    for _, p := range aiProducts {
        // Minimum viable product
        if p.Name == "" {
            continue
        }

        // Clean and normalize
        p.Name = cleanLithuanian(p.Name)
        p.NormalizedName = normalize(p.Name)

        // Extract brand if not provided
        if p.BrandID == nil {
            p.BrandID = e.extractBrand(p.Name)
        }

        // Fix price inconsistencies
        if p.DiscountPrice != nil && p.RegularPrice != nil {
            if *p.DiscountPrice > *p.RegularPrice {
                p.DiscountPrice, p.RegularPrice = p.RegularPrice, p.DiscountPrice
            }
        }

        validated = append(validated, p)
    }

    return validated, nil
}
```

### 1.4 Lithuanian Language Challenges
**Problem**: Search breaks with Lithuanian characters

**Solution**: Dual storage with normalization
```go
func normalize(text string) string {
    replacements := map[string]string{
        "ą": "a", "č": "c", "ę": "e", "ė": "e",
        "į": "i", "š": "s", "ų": "u", "ū": "u",
        "ž": "z", "Ą": "A", "Č": "C", "Ę": "E",
        "Ė": "E", "Į": "I", "Š": "S", "Ų": "U",
        "Ū": "U", "Ž": "Z",
    }

    result := text
    for lt, en := range replacements {
        result = strings.ReplaceAll(result, lt, en)
    }
    return strings.ToLower(result)
}

// Store both versions
product.RawName = "Pienas ŽEMAITIJOS"
product.NormalizedName = "pienas zemaitijos"
product.SearchVector = to_tsvector('lithuanian', product.RawName)
```

### 1.5 Scale & Performance Issues
**Problem**: 4,500+ products weekly, growing database

**Solution**: Aggressive archival strategy
```sql
-- Partitioned tables by month
CREATE TABLE products_2024_11 PARTITION OF products
FOR VALUES FROM ('2024-11-01') TO ('2024-12-01');

-- Auto-archive old flyers
CREATE OR REPLACE FUNCTION archive_old_flyers()
RETURNS void AS $$
BEGIN
    -- Move to cold storage after 2 weeks
    INSERT INTO archived_products
    SELECT * FROM products
    WHERE flyer_id IN (
        SELECT id FROM flyers
        WHERE valid_to < NOW() - INTERVAL '14 days'
    );

    -- Delete from hot table
    DELETE FROM products
    WHERE flyer_id IN (
        SELECT id FROM flyers
        WHERE valid_to < NOW() - INTERVAL '14 days'
    );
END;
$$ LANGUAGE plpgsql;

-- Daily cron
SELECT cron.schedule('archive-flyers', '0 2 * * *',
    'SELECT archive_old_flyers()');
```

---

## 2. Core Data Model (Production-Ready)

### 2.1 Essential Tables

```sql
-- Users (OAuth-ready as per Research.txt)
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255), -- Nullable for OAuth users

    -- OAuth preparation fields
    auth_provider VARCHAR(20) DEFAULT 'email', -- 'email', 'google', 'facebook'
    oauth_id VARCHAR(255), -- External OAuth user ID
    oauth_token TEXT, -- Encrypted OAuth token if needed

    -- Profile
    username VARCHAR(100) UNIQUE,
    full_name VARCHAR(255),
    phone VARCHAR(20),

    -- Metadata
    is_active BOOLEAN DEFAULT true,
    email_verified BOOLEAN DEFAULT false,
    role VARCHAR(20) DEFAULT 'user', -- 'user', 'admin', 'moderator'
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_users_email (email),
    INDEX idx_users_oauth (auth_provider, oauth_id),
    CHECK (
        (auth_provider = 'email' AND password_hash IS NOT NULL) OR
        (auth_provider != 'email' AND oauth_id IS NOT NULL)
    )
);

-- Shops (stable)
CREATE TABLE shops (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(20) UNIQUE NOT NULL, -- 'iki', 'maxima', 'rimi'
    logo_url TEXT,
    -- Location data for future route planning (Research.txt recommendation)
    address TEXT,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    scraper_config JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Flyers (weekly rotation)
CREATE TABLE flyers (
    id BIGSERIAL PRIMARY KEY,
    shop_id BIGINT NOT NULL REFERENCES shops(id),
    title VARCHAR(255) NOT NULL,
    valid_from DATE,
    valid_to DATE,
    status VARCHAR(20) DEFAULT 'active',
    page_count INTEGER DEFAULT 0,
    pdf_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_flyers_active (shop_id, status) WHERE status = 'active',
    INDEX idx_flyers_dates (valid_from, valid_to)
);

-- Products (high churn)
CREATE TABLE products (
    id BIGSERIAL PRIMARY KEY,
    flyer_id BIGINT NOT NULL REFERENCES flyers(id) ON DELETE CASCADE,

    -- Required fields (always present)
    raw_name VARCHAR(255) NOT NULL,
    status VARCHAR(20) DEFAULT 'active',

    -- Optional fields (often missing)
    normalized_name VARCHAR(255),
    brand_id BIGINT REFERENCES brands(id),
    brand_name VARCHAR(100), -- Denormalized
    description TEXT, -- Contains "1+1", "SUPER KAINA", etc.
    regular_price DECIMAL(10,2),
    discount_price DECIMAL(10,2),
    page_number INTEGER,
    image_url TEXT,

    -- Metadata
    extraction_confidence DECIMAL(3,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,

    -- Search
    search_vector tsvector GENERATED ALWAYS AS (
        setweight(to_tsvector('simple', COALESCE(raw_name, '')), 'A') ||
        setweight(to_tsvector('simple', COALESCE(brand_name, '')), 'B')
    ) STORED,

    INDEX idx_products_search USING GIN (search_vector),
    INDEX idx_products_flyer (flyer_id),
    INDEX idx_products_normalized USING gin(normalized_name gin_trgm_ops)
) PARTITION BY RANGE (created_at);

-- Product Tags (stable)
CREATE TABLE tags (
    id BIGSERIAL PRIMARY KEY,
    tag VARCHAR(50) UNIQUE NOT NULL,
    tag_lt VARCHAR(50), -- Lithuanian
    tag_en VARCHAR(50), -- English
    parent_tag_id BIGINT REFERENCES tags(id),
    usage_count INTEGER DEFAULT 0
);

-- Product-Tag mapping
CREATE TABLE product_tags (
    product_id BIGINT NOT NULL,
    tag_id BIGINT NOT NULL REFERENCES tags(id),
    confidence DECIMAL(3,2) DEFAULT 1.0,
    method VARCHAR(20), -- 'ai', 'rule', 'manual'
    PRIMARY KEY (product_id, tag_id)
) PARTITION BY RANGE (product_id);

-- User interaction tracking (Research.txt recommendation for future recommendations)
CREATE TABLE user_product_interactions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    product_id BIGINT,
    interaction_type VARCHAR(20) NOT NULL, -- 'view', 'search', 'add_to_list', 'click'
    search_query TEXT, -- Original search if applicable
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_interactions_user (user_id),
    INDEX idx_interactions_product (product_id),
    INDEX idx_interactions_date (created_at)
) PARTITION BY RANGE (created_at);

-- Smart Shopping Lists
CREATE TABLE shopping_lists (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    name VARCHAR(100) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    share_token VARCHAR(32),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE shopping_list_items (
    id BIGSERIAL PRIMARY KEY,
    list_id BIGINT NOT NULL REFERENCES shopping_lists(id) ON DELETE CASCADE,

    -- Flexible item identification
    specific_product_id BIGINT REFERENCES products(id) ON DELETE SET NULL,
    tag_id BIGINT REFERENCES tags(id),
    custom_text VARCHAR(255),

    -- Item state
    quantity INTEGER DEFAULT 1,
    is_purchased BOOLEAN DEFAULT false,
    last_matched_at TIMESTAMP,

    -- At least one identification method required
    CHECK (
        specific_product_id IS NOT NULL OR
        tag_id IS NOT NULL OR
        custom_text IS NOT NULL
    )
);
```

### 2.2 Common Lithuanian Product Tags

```sql
-- Essential tags for MVP
INSERT INTO tags (tag, tag_lt, tag_en) VALUES
-- Dairy
('pienas', 'Pienas', 'milk'),
('kefyras', 'Kefyras', 'kefir'),
('jogurtas', 'Jogurtas', 'yogurt'),
('suris', 'Sūris', 'cheese'),
('sviestas', 'Sviestas', 'butter'),
('grietine', 'Grietinė', 'sour cream'),

-- Bread
('duona', 'Duona', 'bread'),
('batonas', 'Batonas', 'white bread'),

-- Meat
('mesa', 'Mėsa', 'meat'),
('vistiena', 'Vištiena', 'chicken'),
('kiauliena', 'Kiauliena', 'pork'),
('jautiena', 'Jautiena', 'beef'),
('desra', 'Dešra', 'sausage'),

-- Vegetables
('bulves', 'Bulvės', 'potatoes'),
('morkos', 'Morkos', 'carrots'),
('svogunai', 'Svogūnai', 'onions'),
('pomidorai', 'Pomidorai', 'tomatoes'),

-- Fruits
('obuoliai', 'Obuoliai', 'apples'),
('bananai', 'Bananai', 'bananas'),

-- Beverages
('alus', 'Alus', 'beer'),
('vanduo', 'Vanduo', 'water'),
('sultys', 'Sultys', 'juice'),

-- Common brands as tags (for brand-loyal shopping)
('zemaitijos', 'Žemaitijos', 'Zemaitijos'),
('pieno-zvaigzdes', 'Pieno žvaigždės', 'Pieno zvaigzdes'),
('dvaro', 'Dvaro', 'Dvaro');
```

---

## 3. AI Extraction Strategy

### 3.1 Improved Prompt Engineering

```python
EXTRACTION_PROMPT = """
Analyze this Lithuanian grocery flyer page. Extract products with these rules:

REQUIRED (skip product if missing):
- Product name in Lithuanian

OPTIONAL (use null if not visible):
- Brand name
- Current/sale price
- Original price (crossed out)
- Special offer text (1+1, -30%, SUPER KAINA, etc.)

IMPORTANT:
- Lithuanian prices use comma for decimals (2,99 €)
- Common words: už (for), vnt (pieces), kg, l (liter)
- If you see "1+1" or "2+1", put it in description field
- Brand might be in product name (extract it)

Output clean JSON only:
[{
  "name": "exact text from image",
  "brand": "if visible",
  "price_current": 2.99,
  "price_original": 3.99,
  "description": "special offer text"
}]
"""
```

### 3.2 Fallback Strategies

```go
func processFlyer(flyer Flyer) error {
    for _, page := range flyer.Pages {
        products, err := extractWithAI(page)

        if err != nil || len(products) == 0 {
            // Fallback 1: Try different AI model
            products, err = extractWithGPT35(page)
        }

        if err != nil || len(products) < 5 {
            // Fallback 2: Manual review queue
            queueForManualReview(page)

            // Fallback 3: Use last week's similar page
            products = findSimilarProducts(flyer.ShopID, page.Number)
        }

        // Always validate
        products = validateAndClean(products)

        saveProducts(products)
    }
    return nil
}
```

---

## 4. Search & Discovery

### 4.1 Multi-Strategy Search

```go
type SearchService struct {
    db *bun.DB
}

func (s *SearchService) Search(query string) ([]Product, error) {
    normalized := normalize(query)

    // Strategy 1: Exact match on normalized name
    exact := s.exactMatch(normalized)
    if len(exact) > 0 {
        return exact, nil
    }

    // Strategy 2: PostgreSQL full-text search
    fts := s.fullTextSearch(query)
    if len(fts) > 0 {
        return fts, nil
    }

    // Strategy 3: Trigram similarity
    fuzzy := s.fuzzySearch(normalized, 0.3) // 30% similarity threshold
    if len(fuzzy) > 0 {
        return fuzzy, nil
    }

    // Strategy 4: Tag-based search
    return s.tagSearch(normalized)
}

func (s *SearchService) fuzzySearch(query string, threshold float64) []Product {
    sql := `
        SELECT * FROM products
        WHERE status = 'active'
        AND similarity(normalized_name, ?) > ?
        ORDER BY similarity(normalized_name, ?) DESC
        LIMIT 50
    `

    var products []Product
    s.db.NewRaw(sql, query, threshold, query).Scan(&products)
    return products
}
```

### 4.2 Shopping List Matching

```go
func (s *ShoppingListService) MatchItemsToCurrentFlyers(items []ShoppingListItem) {
    for _, item := range items {
        var matches []Product

        if item.TagID != nil {
            // Get all products with this tag from active flyers
            matches = s.getProductsByTag(*item.TagID)
        } else if item.CustomText != "" {
            // Fuzzy search
            matches = s.searchProducts(item.CustomText)
        }

        // Sort by best deal
        sort.Slice(matches, func(i, j int) bool {
            // Prioritize: bigger discount % > lower price
            discountI := calculateDiscount(matches[i])
            discountJ := calculateDiscount(matches[j])
            return discountI > discountJ
        })

        // Update item with best matches
        item.SuggestedProducts = matches[:min(5, len(matches))]
        item.LastMatchedAt = time.Now()
    }
}
```

---

## 5. API Design (GraphQL)

### 5.1 Essential Queries

```graphql
type Query {
  # Shops
  shops: [Shop!]!
  shop(id: ID!): Shop
  shopByCode(code: String!): Shop

  # Flyers with filtering
  flyers(
    shopId: ID
    shopIds: [ID!]
    status: FlyerStatus
    validOn: Date
    limit: Int = 20
    offset: Int = 0
  ): FlyerConnection!

  activeFlyers(shopId: ID): [Flyer!]!
  flyer(id: ID!): Flyer

  # Flyer details with products
  flyerWithProducts(id: ID!): FlyerDetail!

  # Search with multiple strategies
  searchProducts(
    query: String!
    shopIds: [ID!]
    flyerIds: [ID!]
    priceMax: Float
    priceMin: Float
    onlyDiscounted: Boolean
    tags: [String!]
    limit: Int = 50
    offset: Int = 0
  ): ProductSearchResult!

  # Smart shopping list
  myShoppingLists: [ShoppingList!]! @auth
  shoppingList(id: ID!): ShoppingList @auth
  suggestionsForList(listId: ID!): [ProductSuggestion!]! @auth

  # Shopping list item matching
  matchItemToProducts(
    text: String!
    shopIds: [ID!]
    maxResults: Int = 10
  ): [ProductMatch!]! @auth

  # Tags for autocomplete
  searchTags(query: String!, limit: Int = 20): [Tag!]!
  popularTags(limit: Int = 30): [Tag!]!
}

type Mutation {
  # Auth
  login(email: String!, password: String!): AuthPayload!
  register(email: String!, password: String!, name: String): AuthPayload!
  refreshToken(refreshToken: String!): AuthPayload!
  logout: Boolean! @auth

  # Shopping Lists with smart features
  createShoppingList(name: String!): ShoppingList! @auth
  updateShoppingList(id: ID!, name: String!): ShoppingList! @auth
  deleteShoppingList(id: ID!): Boolean! @auth
  shareShoppingList(id: ID!): String! @auth # Returns share token

  # Smart item addition with multiple strategies
  addToShoppingListSmart(
    listId: ID!
    text: String! # "pienas" - will auto-tag
    quantity: Int = 1
  ): ShoppingListItem! @auth

  # Add by tag (for category shopping)
  addTagToList(
    listId: ID!
    tagId: ID!
    customText: String
    quantity: Int = 1
  ): ShoppingListItem! @auth

  # Manual product selection
  addProductToList(
    listId: ID!
    productId: ID!
    quantity: Int = 1
  ): ShoppingListItem! @auth

  # Batch add products
  addProductsToList(
    listId: ID!
    products: [ProductInput!]!
  ): [ShoppingListItem!]! @auth

  # Item management
  updateShoppingListItem(
    itemId: ID!
    quantity: Int
    customText: String
  ): ShoppingListItem! @auth

  removeFromList(itemId: ID!): Boolean! @auth
  markAsPurchased(itemId: ID!, purchased: Boolean = true): ShoppingListItem! @auth
  clearPurchasedItems(listId: ID!): Int! @auth # Returns count of cleared items

  # Track user interactions for recommendations
  trackProductView(productId: ID!): Boolean! @auth
  trackProductClick(productId: ID!, source: String!): Boolean! @auth
}

# Types
type Shop {
  id: ID!
  name: String!
  code: String! # 'iki', 'maxima', 'rimi'
  logoUrl: String
  address: String
  latitude: Float
  longitude: Float
  isActive: Boolean!
  currentFlyer: Flyer
  flyerCount: Int!
}

type Flyer {
  id: ID!
  shop: Shop!
  title: String!
  validFrom: Date
  validTo: Date
  status: FlyerStatus!
  pageCount: Int!
  pdfUrl: String
  productCount: Int!
  createdAt: DateTime!
}

type FlyerDetail {
  id: ID!
  shop: Shop!
  title: String!
  validFrom: Date
  validTo: Date
  status: FlyerStatus!
  pageCount: Int!
  pdfUrl: String
  pages: [FlyerPage!]!
  products(page: Int, search: String): [Product!]!
  productStats: ProductStats!
}

type FlyerPage {
  pageNumber: Int!
  imageUrl: String!
  productCount: Int!
}

type Product {
  id: ID!
  masterId: ID # Link to canonical product
  rawName: String! # Always present
  normalizedName: String
  brand: Brand
  description: String # "1+1", "SUPER KAINA", etc.
  regularPrice: Float
  discountPrice: Float
  discountPercent: Float
  savings: Float
  flyer: Flyer!
  pageNumber: Int
  tags: [Tag!]!
  imageUrl: String
  extractionConfidence: Float
  # For shopping list matching
  matchScore: Float
}

type ProductStats {
  total: Int!
  withDiscount: Int!
  averageDiscount: Float!
  priceRange: PriceRange!
}

type PriceRange {
  min: Float!
  max: Float!
}

type Tag {
  id: ID!
  tag: String!
  tagLt: String # Lithuanian
  tagEn: String # English
  parentTag: Tag
  usageCount: Int!
  relatedProducts: Int!
}

type ShoppingList {
  id: ID!
  name: String!
  itemCount: Int!
  purchasedCount: Int!
  estimatedTotal: Float
  items: [ShoppingListItem!]!
  shareToken: String
  isActive: Boolean!
  createdAt: DateTime!
  updatedAt: DateTime!
}

type ShoppingListItem {
  id: ID!
  list: ShoppingList!

  # Flexible identification
  specificProduct: Product # Current exact match
  tag: Tag # Category/tag match
  customText: String! # User's original input

  quantity: Int!
  isPurchased: Boolean!

  # Smart matching results
  currentMatch: Product # Best current match
  suggestedProducts: [Product!]! # Top alternatives
  matchMethod: MatchMethod! # How we matched it
  lastMatchedAt: DateTime
}

type ProductMatch {
  product: Product!
  matchScore: Float!
  matchReason: String!
  tag: Tag
}

type ProductSuggestion {
  item: ShoppingListItem!
  bestDeal: Product
  alternatives: [Product!]!
  savingsAmount: Float
  availableShops: [Shop!]!
}

type ProductSearchResult {
  products: [Product!]!
  totalCount: Int!
  facets: SearchFacets!
}

type SearchFacets {
  shops: [ShopFacet!]!
  tags: [TagFacet!]!
  priceRanges: [PriceRangeFacet!]!
}

type ShopFacet {
  shop: Shop!
  count: Int!
}

type TagFacet {
  tag: Tag!
  count: Int!
}

type PriceRangeFacet {
  range: String!
  min: Float!
  max: Float!
  count: Int!
}

type FlyerConnection {
  edges: [FlyerEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type FlyerEdge {
  node: Flyer!
  cursor: String!
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}

type AuthPayload {
  token: String!
  refreshToken: String!
  user: User!
}

type User {
  id: ID!
  email: String!
  name: String
  role: UserRole!
  emailVerified: Boolean!
  createdAt: DateTime!
}

# Enums
enum FlyerStatus {
  DRAFT
  PROCESSING
  ACTIVE
  EXPIRED
  ARCHIVED
}

enum MatchMethod {
  EXACT_PRODUCT
  MASTER_PRODUCT
  TAG_MATCH
  FUZZY_SEARCH
  MANUAL
}

enum UserRole {
  USER
  ADMIN
  MODERATOR
}

# Input types
input ProductInput {
  productId: ID!
  quantity: Int = 1
}
```

---

## 6. Testing Strategy for Real Data

### 6.1 Data Quality Tests

```go
func TestProductExtraction(t *testing.T) {
    // Test with real flyer images
    testCases := []struct {
        name     string
        image    string
        minItems int
        maxItems int
    }{
        {"IKI Page 1", "testdata/iki_page1.jpg", 10, 20},
        {"Maxima Page 1", "testdata/maxima_page1.jpg", 8, 25},
        {"Rimi Sale Page", "testdata/rimi_sale.jpg", 15, 30},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            products, err := ExtractProducts(tc.image)
            assert.NoError(t, err)
            assert.True(t, len(products) >= tc.minItems)
            assert.True(t, len(products) <= tc.maxItems)

            // Validate each product
            for _, p := range products {
                assert.NotEmpty(t, p.Name, "Product must have name")

                // Price sanity checks
                if p.DiscountPrice != nil && p.RegularPrice != nil {
                    assert.True(t, *p.DiscountPrice <= *p.RegularPrice)
                }

                if p.DiscountPrice != nil {
                    assert.True(t, *p.DiscountPrice > 0 && *p.DiscountPrice < 1000)
                }
            }
        })
    }
}
```

### 6.2 Shopping List Resilience Tests

```go
func TestShoppingListPersistence(t *testing.T) {
    // Week 1: Add items
    list := CreateShoppingList("Weekly")
    item1 := AddItemSmart(list, "pienas") // Will tag as 'milk'
    item2 := AddItemSmart(list, "Žemaitijos pienas 2.5%")

    // Verify matches
    suggestions := GetSuggestions(list)
    assert.True(t, len(suggestions) > 0)

    // Simulate week change - archive old flyers
    ArchiveOldFlyers()
    CreateNewFlyers() // Week 2 data

    // Items should still work
    suggestions2 := GetSuggestions(list)
    assert.True(t, len(suggestions2) > 0)

    // Should suggest different products but same tags
    assert.NotEqual(t, suggestions[0].ProductID, suggestions2[0].ProductID)
    assert.Equal(t, suggestions[0].TagID, suggestions2[0].TagID)
}
```

---

## 7. Deployment Configuration

### 7.1 Environment Variables

```bash
# Core
DATABASE_URL=postgres://user:pass@localhost/kainuguru
REDIS_URL=redis://localhost:6379

# AI
OPENAI_API_KEY=sk-...
OPENAI_MODEL=gpt-4-vision-preview
AI_TIMEOUT=30s
AI_MAX_RETRIES=3
FALLBACK_TO_GPT35=true

# Scraping
SCRAPER_USER_AGENT="Mozilla/5.0..."
PDF_DPI=150
MAX_PAGES_PER_FLYER=60
PARALLEL_PAGES=3

# Features
ENABLE_FUZZY_SEARCH=true
FUZZY_THRESHOLD=0.3
ENABLE_AUTO_TAGGING=true
ARCHIVE_AFTER_DAYS=14
```

### 7.2 Database Indexes for Performance

```sql
-- Critical indexes for production load
CREATE INDEX CONCURRENTLY idx_products_active_search
ON products(flyer_id, status)
WHERE status = 'active';

CREATE INDEX CONCURRENTLY idx_products_normalized_gin
ON products USING gin(normalized_name gin_trgm_ops);

CREATE INDEX CONCURRENTLY idx_product_tags_lookup
ON product_tags(tag_id, product_id);

CREATE INDEX CONCURRENTLY idx_shopping_items_list
ON shopping_list_items(list_id)
WHERE is_purchased = false;

-- Materialized view for popular products
CREATE MATERIALIZED VIEW popular_products AS
SELECT
    p.*,
    COUNT(sli.id) as list_count
FROM products p
JOIN shopping_list_items sli ON sli.specific_product_id = p.id
WHERE p.created_at > NOW() - INTERVAL '7 days'
GROUP BY p.id
ORDER BY list_count DESC;

CREATE INDEX ON popular_products(list_count);
```

---

## 8. Critical Production Safeguards

### 8.1 Rate Limiting & DDoS Protection

```go
type RateLimiter struct {
    redis *redis.Client
    limits map[string]int // endpoint -> requests per minute
}

func (r *RateLimiter) Middleware() fiber.Handler {
    return func(c *fiber.Ctx) error {
        key := fmt.Sprintf("rate:%s:%s", c.IP(), c.Path())

        count, _ := r.redis.Incr(ctx, key).Result()
        if count == 1 {
            r.redis.Expire(ctx, key, time.Minute)
        }

        limit := r.limits[c.Path()]
        if limit == 0 {
            limit = 100 // default
        }

        if count > limit {
            return c.Status(429).JSON(fiber.Map{
                "error": "Too many requests",
                "retry_after": 60,
            })
        }

        return c.Next()
    }
}
```

### 8.2 Bulk Operations for Performance

```go
func (s *ProductService) BulkUpsert(products []Product) error {
    // Process in chunks to avoid memory issues
    const chunkSize = 1000

    for i := 0; i < len(products); i += chunkSize {
        end := min(i+chunkSize, len(products))
        batch := products[i:end]

        // Use PostgreSQL's ON CONFLICT for efficient upsert
        _, err := s.db.NewInsert().
            Model(&batch).
            On("CONFLICT (flyer_id, normalized_name) DO UPDATE").
            Set("discount_price = EXCLUDED.discount_price").
            Set("updated_at = NOW()").
            Exec(ctx)

        if err != nil {
            log.Error("Bulk upsert failed for batch", err)
            // Continue with next batch instead of failing completely
        }
    }
    return nil
}
```

### 8.3 Database Connection Pool Configuration

```go
// Critical for production stability
pgconfig := pgdriver.NewConnector(
    pgdriver.WithDSN(dsn),
    pgdriver.WithTimeout(5*time.Second),
    pgdriver.WithDialTimeout(10*time.Second),
    pgdriver.WithReadTimeout(30*time.Second),
    pgdriver.WithWriteTimeout(30*time.Second),
)

sqldb := sql.OpenDB(pgconfig)
sqldb.SetMaxOpenConns(25)      // Total connections
sqldb.SetMaxIdleConns(5)       // Keep alive minimum
sqldb.SetConnMaxLifetime(1*time.Hour)
sqldb.SetConnMaxIdleTime(15*time.Minute)

// TODO: Add monitoring in production
// For MVP, just use basic logging
```

### 8.4 Simple Error Handling (MVP)

```go
type ExtractionService struct {
    chatgpt *ChatGPTClient
}

func (e *ExtractionService) ExtractProducts(image []byte) ([]Product, error) {
    // Simple ChatGPT extraction
    products, err := e.chatgpt.Extract(image)
    if err != nil {
        // Log error and continue
        log.Error("Extraction failed", err)
        // TODO: Queue for manual review in v2
        // TODO: Add fallback strategies when needed
        return []Product{}, nil // Return empty to continue processing
    }

    // Basic validation
    validProducts := []Product{}
    for _, p := range products {
        if p.Name != "" {
            validProducts = append(validProducts, p)
        }
    }

    return validProducts, nil
}
```

---

## 9. Architecture Patterns (Enhanced)

### 9.1 Service Interface Abstraction

```go
// Define interfaces for future microservice extraction
type SearchService interface {
    Search(ctx context.Context, query string, opts SearchOptions) ([]Product, error)
    SearchByTag(ctx context.Context, tagID int64) ([]Product, error)
    SearchFuzzy(ctx context.Context, text string, threshold float64) ([]Product, error)
}

type ProductService interface {
    GetByID(ctx context.Context, id int64) (*Product, error)
    GetByMasterID(ctx context.Context, masterID int64) ([]Product, error)
    CreateMasterProduct(ctx context.Context, product *Product) (*ProductMaster, error)
    LinkToMaster(ctx context.Context, productID, masterID int64) error
}

type ScraperService interface {
    ScrapeFlyer(ctx context.Context, flyerID int64) error
    ProcessImage(ctx context.Context, imageURL string) ([]Product, error)
}

// Initial monolith implementations
type postgresSearchService struct {
    db *sql.DB
}

type inMemoryProductService struct {
    db *sql.DB
    cache *redis.Client
}
```

### 9.2 Simple Queue with PostgreSQL (MVP-friendly)

```go
// Use PostgreSQL as queue for MVP - no extra infrastructure
type SimpleQueue struct {
    db *bun.DB
}

func (q *SimpleQueue) Enqueue(job Job) error {
    job.Status = "pending"
    job.CreatedAt = time.Now()
    return q.db.NewInsert().Model(&job).Exec(ctx)
}

func (q *SimpleQueue) Dequeue() (*Job, error) {
    var job Job
    // SKIP LOCKED prevents multiple workers from getting same job
    err := q.db.NewSelect().
        Model(&job).
        Where("status = ?", "pending").
        Where("retry_count < ?", 3).
        For("UPDATE SKIP LOCKED").
        Limit(1).
        Scan(ctx)

    if err == nil {
        job.Status = "processing"
        q.db.NewUpdate().Model(&job).Where("id = ?", job.ID).Exec(ctx)
    }
    return &job, err
}

// Migrate to NATS only when you need:
// - 10k+ messages/second
// - Complex routing
// - Multi-consumer groups
```

### 9.3 Event-Driven Pattern

```go
// Event system for future NATS migration
type EventType string

const (
    EventFlyerScraped   EventType = "flyer.scraped"
    EventProductCreated EventType = "product.created"
    EventListUpdated    EventType = "list.updated"
)

type Event struct {
    Type      EventType
    Payload   interface{}
    Timestamp time.Time
}

// Start with Go channels, migrate to NATS later
type EventBus interface {
    Publish(ctx context.Context, event Event) error
    Subscribe(eventType EventType, handler func(Event)) error
}

// Channel-based implementation for MVP
type channelEventBus struct {
    subscribers map[EventType][]chan Event
    mu          sync.RWMutex
}

func (bus *channelEventBus) Publish(ctx context.Context, event Event) error {
    bus.mu.RLock()
    defer bus.mu.RUnlock()

    for _, ch := range bus.subscribers[event.Type] {
        select {
        case ch <- event:
        case <-ctx.Done():
            return ctx.Err()
        default:
            // Non-blocking, log dropped event
        }
    }
    return nil
}

// Job queue for background tasks
type JobQueue interface {
    Enqueue(job Job) error
    Process(ctx context.Context, workers int) error
}

type Job struct {
    ID       string
    Type     string
    Payload  json.RawMessage
    Retries  int
    MaxRetries int
    CreatedAt time.Time
}

// Database-backed queue for MVP
CREATE TABLE job_queue (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    retries INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    INDEX idx_jobs_pending (status, created_at) WHERE status = 'pending'
);
```

### 8.3 Simple Logging (MVP)

```go
// Use zerolog for structured logging - no fancy monitoring
logger := zerolog.New(os.Stdout).With().
    Timestamp().
    Str("service", "kainuguru").
    Logger()

// Basic request logging
func LogRequest(c *fiber.Ctx) error {
    start := time.Now()

    err := c.Next()

    logger.Info().
        Str("method", c.Method()).
        Str("path", c.Path()).
        Int("status", c.Response().StatusCode()).
        Dur("latency", time.Since(start)).
        Msg("request")

    return err
}

// TODO: Add proper monitoring post-MVP:
// - Prometheus metrics
// - Grafana dashboards
// - Error tracking (Sentry)
// - APM tools
```

---

## 9. Additional Checklist Features

### 9.1 Flyer Browsing Enhancements

```graphql
# Add to Query type
type Query {
  # Show expired flyers with visual distinction
  allFlyers(
    includeExpired: Boolean = false
    shopId: ID
  ): [Flyer!]!

  # Search within specific flyer
  searchProductsInFlyer(
    flyerId: ID!
    query: String!
  ): [Product!]!

  # Search flyers by title
  searchFlyers(
    query: String! # "IKI Weekly Deals"
  ): [Flyer!]!
}

# Update Flyer type
type Flyer {
  # ... existing fields ...
  isExpired: Boolean!
  expirationLabel: String # "Expired 2 days ago"
  displayStatus: DisplayStatus! # ACTIVE, EXPIRING_SOON, EXPIRED
}

enum DisplayStatus {
  ACTIVE
  EXPIRING_SOON # Last 24h
  EXPIRED
}
```

### 9.2 Notifications & Reminders

```go
// Notification service for new flyers
type NotificationService interface {
    NotifyNewFlyers(ctx context.Context, userID int64, flyers []Flyer) error
    ScheduleReminder(ctx context.Context, userID int64, reminder Reminder) error
}

type Reminder struct {
    Type      string // "new_flyers", "expiring_deals"
    UserID    int64
    Frequency string // "weekly", "daily"
    Settings  map[string]interface{}
}

// Add to user preferences
CREATE TABLE user_preferences (
    user_id BIGINT PRIMARY KEY REFERENCES users(id),
    notify_new_flyers BOOLEAN DEFAULT true,
    notify_expiring_deals BOOLEAN DEFAULT false,
    notification_email BOOLEAN DEFAULT true,
    notification_push BOOLEAN DEFAULT false,
    preferred_shops JSONB DEFAULT '[]', -- ["iki", "maxima"]
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 9.3 Smart Recommendations

```go
// Recommendation engine interface
type RecommendationEngine interface {
    // Learn from user behavior
    TrackAddedProduct(ctx context.Context, userID, productID int64)
    TrackDismissedProduct(ctx context.Context, userID, productID int64)
    TrackViewedFlyer(ctx context.Context, userID, flyerID int64)

    // Generate recommendations
    GetPersonalizedDeals(ctx context.Context, userID int64) ([]Product, error)
    GetSimilarProducts(ctx context.Context, productID int64) ([]Product, error)
    GetCheapestOptions(ctx context.Context, items []ShoppingListItem) (map[int64]Product, error)
}

// Track dismissals and preferences
CREATE TABLE user_dismissals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    product_id BIGINT,
    tag_id BIGINT REFERENCES tags(id),
    reason VARCHAR(50), -- 'not_interested', 'too_expensive', 'wrong_brand'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_dismissals_user (user_id, created_at)
);
```

### 9.4 Product Highlighting in Flyer View

```go
// Product location on flyer page
type ProductLocation struct {
    ProductID   int64
    PageNumber  int
    BoundingBox BBox
}

type BBox struct {
    X      float64 `json:"x"`
    Y      float64 `json:"y"`
    Width  float64 `json:"width"`
    Height float64 `json:"height"`
}

// Store product locations from AI extraction
ALTER TABLE products ADD COLUMN bbox JSONB;

// GraphQL addition
type Product {
  # ... existing fields ...
  boundingBox: BoundingBox # For flyer highlighting
}

type BoundingBox {
  x: Float!
  y: Float!
  width: Float!
  height: Float!
}
```

### 9.5 Shopping List Enhancements

```graphql
# Add to Mutation type
type Mutation {
  # Reorder items in list
  reorderListItems(
    listId: ID!
    itemIds: [ID!]! # New order
  ): ShoppingList! @auth

  # Clear entire list
  clearShoppingList(listId: ID!): ShoppingList! @auth

  # Clone a list
  cloneShoppingList(
    listId: ID!
    newName: String!
  ): ShoppingList! @auth
}

# Shopping optimization
type Query {
  # Find cheapest shop for entire list
  optimizeShoppingRoute(
    listId: ID!
    maxStores: Int = 2
  ): ShoppingOptimization! @auth
}

type ShoppingOptimization {
  totalSavings: Float!
  stores: [StoreOptimization!]!
  missingItems: [ShoppingListItem!]!
}

type StoreOptimization {
  shop: Shop!
  items: [OptimizedItem!]!
  subtotal: Float!
  savings: Float!
}

type OptimizedItem {
  listItem: ShoppingListItem!
  product: Product!
  price: Float!
  savings: Float
}
```

### 9.6 Duplicate Detection & Scheduled Scraping

```go
// Duplicate detection
func (s *ScraperService) IsDuplicateFlyer(flyer *Flyer) bool {
    // Check by shop + date range + title hash
    exists, _ := s.db.Exists(`
        SELECT 1 FROM flyers
        WHERE shop_id = $1
        AND valid_from = $2
        AND valid_to = $3
        AND title = $4
    `, flyer.ShopID, flyer.ValidFrom, flyer.ValidTo, flyer.Title)
    return exists
}

// Scheduled scraping with cron
CREATE TABLE scraper_schedule (
    id SERIAL PRIMARY KEY,
    shop_code VARCHAR(20) NOT NULL,
    cron_expression VARCHAR(100) NOT NULL, -- "0 6 * * MON"
    is_active BOOLEAN DEFAULT true,
    last_run TIMESTAMP,
    next_run TIMESTAMP,
    consecutive_failures INT DEFAULT 0,
    UNIQUE(shop_code)
);

-- Default schedules
INSERT INTO scraper_schedule (shop_code, cron_expression) VALUES
('iki', '0 6 * * MON'),      -- Monday 6 AM
('maxima', '0 6 * * TUE'),   -- Tuesday 6 AM
('rimi', '0 6 * * WED'),     -- Wednesday 6 AM
('lidl', '0 6 * * THU'),     -- Thursday 6 AM
('norfa', '0 6 * * FRI');    -- Friday 6 AM
```

### 9.7 Manual Review Queue

```sql
-- Failed extraction tracking
CREATE TABLE extraction_failures (
    id BIGSERIAL PRIMARY KEY,
    flyer_id BIGINT REFERENCES flyers(id),
    page_number INT,
    error_type VARCHAR(50), -- 'ai_failed', 'no_products', 'validation_failed'
    error_message TEXT,
    retry_count INT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'reviewing', 'resolved'
    resolved_by BIGINT REFERENCES users(id),
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_failures_status (status, created_at)
);
```

### 9.8 Migration Support

```go
// Database migrations with Goose
type MigratorService struct {
    db *sql.DB
}

func (m *MigratorService) RunMigrations() error {
    return goose.Up(m.db, "./migrations")
}

// Migration files structure
// migrations/
//   001_initial_schema.sql
//   002_add_product_masters.sql
//   003_add_user_interactions.sql
//   004_add_notifications.sql
```

---

## 10. MVP Success Metrics

### 10.1 BDD Testing Strategy

```go
// Using Ginkgo & Gomega for BDD
var _ = Describe("Shopping List Feature", func() {
    var (
        app *fiber.App
        db  *bun.DB
        token string
    )

    BeforeEach(func() {
        // Setup test database and auth token
    })

    Describe("Creating a shopping list", func() {
        Context("when user is authenticated", func() {
            It("should create a new shopping list", func() {
                req := httptest.NewRequest(
                    "POST",
                    "/graphql",
                    strings.NewReader(`{
                        "query": "mutation { createShoppingList(name: \"Weekly\") { id name } }"
                    }`),
                )
                req.Header.Set("Authorization", "Bearer " + token)

                resp, err := app.Test(req)
                Expect(err).NotTo(HaveOccurred())
                Expect(resp.StatusCode).To(Equal(200))
            })
        })
    })

    Describe("Smart item matching", func() {
        It("should match items across flyer rotations", func() {
            // Add item "pienas"
            // Archive old flyer
            // Load new flyer
            // Verify item still matches via tag
        })
    })
})
```

### 10.2 Implementation Timeline

#### Phase 1: Foundation (Week 1)
- **Day 1-2**: Project setup, Docker, PostgreSQL + Redis
- **Day 3-4**: Database schemas, Bun models, migrations
- **Day 5-7**: Core services (user, auth, flyer, product)

#### Phase 2: API Layer (Week 2)
- **Day 8-10**: GraphQL setup with gqlgen, resolvers, auth middleware
- **Day 11-12**: Search implementation with PostgreSQL FTS
- **Day 13-14**: Unit tests and BDD framework setup

#### Phase 3: Scraping & AI (Week 3)
- **Day 15-17**: Store scrapers (IKI, Maxima, Rimi), PDF processing
- **Day 18-19**: OpenAI GPT-4 Vision integration, fallback extractors
- **Day 20-21**: Processing pipeline, queue system, logging

#### Phase 4: Polish & Deploy (Week 4)
- **Day 22-23**: Complete test coverage
- **Day 24-25**: Performance optimization, caching
- **Day 26-27**: Documentation, deployment guides
- **Day 28**: Production deployment

### 10.3 Simple Success Criteria (MVP)

**Basic Functionality**
- ✅ Flyers are scraped successfully
- ✅ Products are extracted (any success rate)
- ✅ Search returns results
- ✅ Shopping lists work
- ✅ Users can register and login

**Performance**
- API responds in reasonable time (<1s)
- Site doesn't crash under normal load
- Database queries complete

**No Complex Metrics for MVP**
- TODO: Add analytics after launch
- TODO: Track extraction success rates post-MVP
- TODO: Monitor user behavior after getting users
- TODO: Implement cost tracking when needed

---

## 11. Simplified Tech Stack for MVP

### 11.1 Essential Stack (MVP)

```yaml
Core:
  Language: Go 1.22+
  Framework: Fiber v2
  API: GraphQL with gqlgen (from start)
  Database: PostgreSQL 15+
  Cache: Redis (sessions only)
  Container: Docker

MVP Choices:
  ORM: Bun (good enough for MVP)
  Search: PostgreSQL FTS only (simple and effective)
  Queue: PostgreSQL table (no extra infrastructure)

AI Strategy:
  Primary: ChatGPT/GPT-4 Vision (simple to start)
  # TODO: Add cost monitoring after launch
  # TODO: Implement OCR fallback when costs known

Defer to Post-MVP:
  - Elasticsearch (when PG search isn't enough)
  - NATS (when PostgreSQL queue isn't enough)
  - Kubernetes (when single server isn't enough)
  - Cost optimization (after understanding usage patterns)
```

### 11.2 Infrastructure Budget

```yaml
Monthly Costs (MVP):
  DigitalOcean Droplet: €50 (4GB RAM)
  PostgreSQL Managed: €15 (1GB, 10GB storage)
  Redis: €10 (512MB)
  Object Storage: €5 (100GB images)
  ChatGPT API: TBD (monitor first month)
  Total: €80/month + API costs

Development Timeline:
  Week 1-2: Core backend, database, auth
  Week 3: OCR extraction, basic AI
  Week 4: Search, shopping lists
  Week 5-6: Testing, optimization, deployment

Team (Minimum):
  1 Senior Backend Dev (6 weeks)
  1 Junior Backend Dev (4 weeks)
  DevOps consultation (1 week)
```

---

## 12. Directory Structure

```
kainuguru-go/
├── cmd/
│   ├── api/          # API server entry point
│   ├── scraper/      # Scraper worker
│   └── migrator/     # Database migrator
├── internal/
│   ├── config/       # Configuration
│   ├── models/       # Database models
│   ├── repository/   # Data access layer
│   ├── services/     # Business logic
│   │   ├── auth/
│   │   ├── product/
│   │   ├── scraper/
│   │   └── ai/
│   ├── handlers/     # HTTP/GraphQL handlers
│   ├── middleware/   # Auth, logging, etc.
│   └── utils/        # Helpers
├── pkg/
│   ├── openai/       # OpenAI client
│   ├── scraper/      # Scraping utilities
│   └── search/       # Search implementations
├── migrations/       # Database migrations
├── configs/          # Config files
├── scripts/          # Utility scripts
├── tests/           # Test files
│   ├── unit/
│   ├── integration/
│   └── bdd/
└── docker/          # Docker configurations
```

---

## Conclusion

This MVP specification provides a complete foundation for the Kainuguru platform that:
- **Maintains original goal**: Grocery flyer aggregation with smart shopping lists
- **Uses GraphQL from start**: Modern API with gqlgen for better frontend experience
- **Simple AI approach**: ChatGPT/GPT-4 Vision without complex cost controls initially
- **PostgreSQL for everything**: Search, queue, and data storage in one place
- **Production safeguards**: Rate limiting, distributed locking, bulk operations
- **Focus on shipping**: TODOs for optimization rather than over-engineering
- **Ship first, optimize later**: Get real users before complex monitoring

**📌 This document is the SINGLE SOURCE OF TRUTH for Kainuguru MVP implementation.**

All development decisions should reference this specification. Other documents (Research.txt, analysis.md, etc.) are superseded by this final specification.

**Key Principles:**
- Ship MVP in 6 weeks
- Use simple, proven technologies
- No fancy monitoring or complex metrics
- PostgreSQL for everything (search, queue, storage)
- GraphQL from day one
- ChatGPT without cost controls initially
- Basic logging only
- TODOs for future optimizations
- Search response time: <200ms
- Flyer processing time: <2min per flyer
- Shopping list load time: <500ms
- Basic uptime (manual checking is fine for MVP)

### 8.3 Post-MVP TODOs
- Add real monitoring tools
- Track user metrics
- Implement cost controls
- Weekly active users: Growth >10% month-over-month
- Shopping list retention: >40% use weekly

---

## 9. Common Pitfalls to Avoid

1. **Don't trust AI blindly** - Always validate and have fallbacks
2. **Don't assume consistency** - Product data WILL be messy
3. **Don't over-normalize** - Keep raw data for debugging
4. **Don't forget Lithuanian specifics** - Prices use commas, special characters matter
5. **Don't block on perfect data** - 80% accuracy is better than no data
6. **Don't ignore user feedback** - They'll help identify matching issues
7. **Don't keep everything forever** - Aggressive archival is essential

---

## 10. Future Considerations (Post-MVP)

1. **Machine Learning Pipeline**
   - Train custom product extraction model
   - Learn user preferences for better matching
   - Predict price drops

2. **Community Features**
   - User-submitted price corrections
   - Crowd-sourced product tagging
   - Shopping list templates sharing

3. **Advanced Analytics**
   - Price trend analysis
   - Inflation tracking
   - Personal spending insights

4. **Integration Opportunities**
   - Store loyalty cards
   - Payment providers
   - Recipe platforms

---

## Conclusion

This MVP specification is designed for the **real world** where:
- Data is messy and inconsistent
- AI extraction is imperfect
- Products change weekly
- Users expect things to "just work"

By focusing on resilient data structures, smart fallbacks, and tag-based matching, the system can deliver value even with imperfect data. The key is accepting that perfection is impossible and building systems that degrade gracefully while still providing utility to users.