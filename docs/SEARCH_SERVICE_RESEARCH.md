# SearchService Implementation Research Report

## Executive Summary
The kainuguru-api SearchService is a comprehensive search solution built on PostgreSQL with support for both full-text search (FTS) and fuzzy matching via trigram similarity. The service is well-integrated with the GraphQL API and ProductMaster, offering multiple search strategies and faceted navigation.

---

## 1. FuzzySearchProducts Method Signature & API Contract

### Method Signature
```go
func (s *searchService) FuzzySearchProducts(
    ctx context.Context, 
    req *SearchRequest
) (*SearchResponse, error)
```

### SearchRequest Structure
```go
type SearchRequest struct {
    Query       string    `json:"query" validate:"required,min=1,max=255"`
    StoreIDs    []int     `json:"store_ids,omitempty"`
    MinPrice    *float64  `json:"min_price,omitempty" validate:"omitempty,gte=0"`
    MaxPrice    *float64  `json:"max_price,omitempty" validate:"omitempty,gte=0"`
    OnSaleOnly  bool      `json:"on_sale_only"`
    Category    string    `json:"category,omitempty"`
    Tags        []string  `json:"tags,omitempty"`
    Limit       int       `json:"limit" validate:"min=1,max=100"`
    Offset      int       `json:"offset" validate:"min=0"`
    PreferFuzzy bool      `json:"prefer_fuzzy"`
}
```

### SearchResponse Structure
```go
type SearchResponse struct {
    Products    []ProductSearchResult `json:"products"`
    TotalCount  int                   `json:"total_count"`
    QueryTime   time.Duration         `json:"query_time"`
    Suggestions []string              `json:"suggestions,omitempty"`
    HasMore     bool                  `json:"has_more"`
}
```

### ProductSearchResult Structure
```go
type ProductSearchResult struct {
    Product     *models.Product `json:"product"`
    SearchScore float64         `json:"search_score"`
    MatchType   string          `json:"match_type"`
    Similarity  float64         `json:"similarity,omitempty"`
    Highlights  []string        `json:"highlights,omitempty"`
}
```

### Validation Rules
- Query: 1-255 characters, UTF-8 valid, no SQL injection patterns
- Limit: 1-100 (default varies by resolver)
- Offset: 0-10000
- Price: 0-10000 EUR range
- Store IDs: max 50 stores
- Category: optional, max 100 characters
- Tags: array filter support

---

## 2. Similarity Scores & Usage

### Score Types in FuzzySearchProducts

The fuzzy search returns **three separate similarity scores** calculated in the database:

#### a) Name Similarity
```sql
similarity(p.name, search_query)::FLOAT as name_similarity
```
- Uses PostgreSQL's trigram similarity function (0.0 to 1.0)
- Matches product name against search query
- Threshold: 0.3 (30% minimum match)

#### b) Brand Similarity
```sql
COALESCE(similarity(p.brand, search_query), 0)::FLOAT as brand_similarity
```
- Matches product brand against search query
- Only used if brand exists (COALESCE defaults to 0)
- Same 0.3 threshold

#### c) Combined Similarity (Primary Score)
```sql
(
    similarity(p.name, search_query) * 0.7 +
    COALESCE(similarity(p.brand, search_query), 0) * 0.3 +
    similarity(p.normalized_name, normalized_query) * 0.2
)::FLOAT as combined_similarity
```

**Score Weighting:**
- Name match: 70% weight (0.7)
- Brand match: 30% weight (0.3)
- Normalized name match: 20% bonus weight

**Usage Pattern:**
1. The database function returns all three scores
2. Go service receives combined_similarity in scoresMap (line 97 of service.go)
3. Results are sorted by combined_similarity DESC (database level)
4. Returned as `SearchScore` field in ProductSearchResult

### Hybrid Search Scores
Unlike fuzzy search, hybrid search returns:
- `search_score`: FLOAT
- `match_type`: "fts" (full-text search) or "fuzzy" (trigram)

#### Score Calculation in Hybrid
```sql
CASE
    WHEN prefer_fuzzy THEN cr.score * (CASE WHEN cr.match_type = 'fuzzy' THEN 1.2 ELSE 1.0 END)
    ELSE cr.score
END as search_score
```

- FTS matches: `ts_rank_cd(p.search_vector, query_tsquery) * 2.0`
- Fuzzy matches: weighted combination (same as fuzzy search)
- PreferFuzzy boost: multiplies fuzzy matches by 1.2x

### Score Range
- **Minimum threshold**: 0.3 (30% similarity)
- **Maximum**: 1.0 (perfect match)
- **Practical range**: 0.3-0.95 (rarely 1.0 in real data)

### Using Scores in Application
```go
// From line 127-132 of service.go
results = append(results, ProductSearchResult{
    Product:     product,
    SearchScore: score,        // Combined similarity score
    MatchType:   "fuzzy",
    Similarity:  score,        // Also stored here for GraphQL
})
```

---

## 3. Brand Filtering & Brand-Aware Searching

### Current Brand Support

#### a) **Database-Level Brand Matching**
In `fuzzy_search_products()`:
```sql
(p.brand IS NOT NULL AND similarity(p.brand, search_query) >= similarity_threshold)
```
- Brand field IS used in similarity calculations
- Contributes 30% to combined score
- Part of matching condition but secondary to name

#### b) **Combined Name+Brand Matching**
```sql
similarity(p.name || ' ' || COALESCE(p.brand, ''), search_query) >= similarity_threshold
```
- Concatenates name and brand with space
- Fourth matching condition for fuzzy search
- Helps find products when brand is closely associated with name

### Brand Faceting

Yes, brand faceting IS implemented (lines 567-593 in query.go):
```go
type brandFacetRow struct {
    Brand string `bun:"brand"`
    Count int    `bun:"count"`
}
var brandRows []brandFacetRow
brandQuery := buildBaseQuery(r.db.NewSelect()).
    ColumnExpr("p.brand").
    ColumnExpr("COUNT(*) AS count").
    Where("p.brand IS NOT NULL AND p.brand != ''").
    Group("p.brand").
    Order("count DESC").
    Limit(20)  // Top 20 brands
```

**Facet Returns:**
- List of available brands with product counts
- Limited to top 20 brands for performance
- Included in SearchFacets response

### Brand Filtering Limitation

**IMPORTANT**: Brand filtering is NOT implemented in SearchRequest/SearchInput:
- SearchRequest has no `Brands` field
- GraphQL SearchInput has no brand filter
- Comment at line 704: `// Brand filtering not implemented in SearchRequest yet`
- Brands are returned as facets but cannot be used to filter search results

### Recommendation for Brand-Aware Search

To enhance brand awareness:
1. **Add brand filter to SearchRequest** (currently missing)
2. **Database function already supports it** - just need to pass brands parameter
3. **GraphQL schema** needs `brands: [String!]` in SearchInput
4. **Resolver** needs to convert and pass brand filters

---

## 4. Integration with ProductMaster

### Current Integration Status

#### a) **Data Model Integration**
In GraphQL schema (line 66 of schema.graphql):
```graphql
type Product {
    ...
    productMaster: ProductMaster
}
```

Product model has relation to ProductMaster, but search service does NOT currently join it.

#### b) **Search Service Integration**
In `FuzzySearchProducts()` (lines 104-110):
```go
// Load full products with relations
var products []*models.Product
if len(productIDs) > 0 {
    s.logger.Info("loading products with relations", "product_ids", productIDs)
    err = s.db.NewSelect().
        Model(&products).
        Where("p.id IN (?)", bun.In(productIDs)).
        Relation("Store").
        Relation("Flyer").
        Relation("FlyerPage").
        Scan(ctx)
```

**Relations Loaded:**
- Store (always)
- Flyer (always)
- FlyerPage (always)
- **ProductMaster: NOT loaded** (absent from relations)

#### c) **ProductMaster Purpose**
ProductMaster is a master catalog that deduplicates and normalizes product data:
- Groups identical products from different stores
- Maintains canonical product information
- Used for shopping list item linking
- Currently available via separate `productMaster` query, not search results

#### d) **How ProductMaster is Queried**
```go
func (r *queryResolver) ProductMaster(ctx context.Context, id int) (*models.ProductMaster, error) {
    return r.productMasterService.GetByID(ctx, int64(id))
}
```

Separate from search service - accessed independently.

### Recommendation for ProductMaster Integration

1. **Add ProductMaster relation** to search product loading
2. **Consider product deduplication** - search currently returns individual product instances (store-specific)
3. **Could use master matching** for better relevance in multi-store environments

---

## 5. Complete GraphQL API Integration

### Query Entry Point
```graphql
query SearchProducts($input: SearchInput!) {
    searchProducts(input: $input): SearchResult!
}
```

### GraphQL SearchInput Type
```graphql
input SearchInput {
    q: String!                           # Required: search query
    storeIDs: [Int!]                    # Optional: filter by stores
    minPrice: Float                     # Optional: minimum price
    maxPrice: Float                     # Optional: maximum price
    onSaleOnly: Boolean = false         # Optional: only sale items
    category: String                    # Optional: category filter
    tags: [String!]                     # Optional: tag filters
    first: Int = 50                     # Optional: limit
    after: String                       # Optional: cursor (not used)
    preferFuzzy: Boolean = false        # Optional: boost fuzzy matches
}
```

### GraphQL SearchResult Type
```graphql
type SearchResult {
    products: [ProductSearchResult!]!
    totalCount: Int!
    queryString: String!
    suggestions: [String!]
    hasMore: Boolean!
    facets: SearchFacets!
    pagination: Pagination!
}

type ProductSearchResult {
    product: Product!
    searchScore: Float!        # Combined similarity (0.0-1.0)
    matchType: String!         # "exact" in resolver (always)
    similarity: Float          # Same as searchScore
    highlights: [String!]      # Empty (not populated)
}
```

### Resolver Implementation (lines 265-376 in query.go)
```go
func (r *queryResolver) SearchProducts(
    ctx context.Context, 
    input model.SearchInput
) (*model.SearchResult, error)
```

**Flow:**
1. Convert GraphQL SearchInput → search.SearchRequest
2. Call `r.searchService.SearchProducts(ctx, searchReq)`
3. Route to FuzzySearchProducts or HybridSearchProducts based on preferFuzzy flag
4. Load additional relations (Store, Flyer, FlyerPage)
5. Compute facets via `computeSearchFacets()`
6. Return SearchResult with pagination

### Facets Computation
Eight separate database queries compute:
- **Stores**: Available stores with product counts (lines 500-533)
- **Categories**: Available categories with counts (lines 535-565)
- **Brands**: Top 20 brands with counts (lines 567-593)
- **Price Ranges**: 5 predefined ranges (lines 595-648)
- **Availability**: On-sale vs regular price (lines 653-688)

---

## 6. SQL Functions (Database Level)

### fuzzy_search_products()
**File:** migration 029_add_tags_to_search_functions.sql (lines 122-192)

**Parameters:**
```sql
fuzzy_search_products(
    search_query TEXT,
    similarity_threshold FLOAT DEFAULT 0.3,
    limit_count INTEGER DEFAULT 50,
    offset_count INTEGER DEFAULT 0,
    store_ids INTEGER[] DEFAULT NULL,
    category_filter TEXT DEFAULT NULL,
    min_price DECIMAL DEFAULT NULL,
    max_price DECIMAL DEFAULT NULL,
    on_sale_only BOOLEAN DEFAULT FALSE,
    tag_filters TEXT[] DEFAULT NULL
)
```

**Returns:**
```sql
product_id BIGINT,
name TEXT,
brand TEXT,
category TEXT,
current_price DECIMAL,
store_id INTEGER,
flyer_id INTEGER,
name_similarity FLOAT,
brand_similarity FLOAT,
combined_similarity FLOAT
```

**Key Operations:**
1. Normalizes query using `normalize_lithuanian_text()`
2. Checks four fuzzy conditions (name, normalized_name, brand, combined)
3. Applies all filters (store, price, category, tags, sale)
4. Orders by combined_similarity DESC, price ASC, name ASC
5. Supports pagination with LIMIT/OFFSET

### hybrid_search_products()
**File:** migration 029_add_tags_to_search_functions.sql (lines 7-117)

**Parameters:**
```sql
hybrid_search_products(
    search_query TEXT,
    limit_count INTEGER DEFAULT 50,
    offset_count INTEGER DEFAULT 0,
    store_ids INTEGER[] DEFAULT NULL,
    min_price DECIMAL DEFAULT NULL,
    max_price DECIMAL DEFAULT NULL,
    prefer_fuzzy BOOLEAN DEFAULT FALSE,
    category_filter TEXT DEFAULT NULL,
    on_sale_only BOOLEAN DEFAULT FALSE,
    tag_filters TEXT[] DEFAULT NULL
)
```

**Returns:**
```sql
product_id BIGINT,
name TEXT,
brand TEXT,
current_price DECIMAL,
store_id INTEGER,
flyer_id INTEGER,
search_score FLOAT,
match_type TEXT
```

**Algorithm:**
1. **FTS Results**: Full-text search using `plainto_tsquery()` + search_vector
   - Scored: `ts_rank_cd(search_vector, query_tsquery) * 2.0`
   - Match type: "fts"
   
2. **Trigram Results**: Fuzzy matching (same as fuzzy_search_products)
   - Scored: weighted combination
   - Match type: "fuzzy"
   - Excludes FTS matches (EXCEPT)

3. **Combined Ranking**: Union of both results
   - Apply preferFuzzy boost (1.2x multiplier for fuzzy matches)
   - Order by search_score DESC, price ASC, name ASC

---

## 7. Service Configuration & Defaults

### Hardcoded Values in Service

| Parameter | Value | Location | Note |
|-----------|-------|----------|------|
| Similarity Threshold | 0.3 | service.go:61 | 30% minimum match |
| Default Limit | 50 | query.go:288 | GraphQL resolver default |
| FTS Boost | 2.0x | migration 021:45 | Full-text search multiplier |
| Fuzzy Boost | 1.2x | migration 021:109 | When preferFuzzy enabled |
| Name Weight | 70% | migration 021:163 | In combined score |
| Brand Weight | 30% | migration 021:164 | In combined score |
| Normalized Weight | 20% | migration 021:165 | Bonus for diacritic matching |
| Top Brands | 20 | query.go:579 | Facet limit |
| Max Stores | 50 | validation.go:195 | Filter validation |

### Service Initialization
```go
// From internal/services/factory.go
return search.NewSearchService(f.db, logger)

// Constructor from search/service.go:20-24
func NewSearchService(db *bun.DB, logger *slog.Logger) Service {
    return &searchService{
        db:     db,
        logger: logger,
    }
}
```

---

## 8. Key Limitations & Gaps

### Current Limitations

1. **Brand Filtering Missing**
   - Facets show brands but cannot filter by them
   - Need to add `brands` field to SearchInput and SearchRequest

2. **ProductMaster Not Loaded**
   - Search results don't include ProductMaster relation
   - Would need: `.Relation("ProductMaster")` in line 107

3. **Highlights Not Populated**
   - GraphQL schema includes highlights field
   - Always returns empty array
   - Would need highlighting logic at database or application level

4. **Match Type Always "exact"**
   - Resolver hardcodes "exact" (line 306)
   - Should use actual match_type from database ("fts" or "fuzzy")

5. **No Query Correction Support in Fuzzy**
   - HybridSearchProducts includes FTS advantages
   - FuzzySearchProducts is pure trigram (no spell-check)

6. **Limited Facet Active Values**
   - Stores and categories track active filters
   - Brands/prices/availability don't track active state

---

## 9. Performance Characteristics

### Database-Level Optimizations

**Indexes Supporting Search:**
- Trigram indexes on product name (migration 010)
- Full-text search vector index
- Store ID index
- Price range indexes
- Tag array overlap (@@ operator)

**Optimization Strategies:**
1. Database performs similarity calculations (not application)
2. Tag filtering uses array overlap (fast)
3. FTS excludes trigram results (avoid duplication)
4. Price filtering at database level
5. LIMIT/OFFSET at database (don't fetch all then filter)

### Query Performance

From logging in service.go:
- Execution time tracked and logged
- Returned in response as QueryTime (time.Duration)
- Analytics logged separately

---

## 10. Testing Surface

### GraphQL Integration Test
File: tests/bdd/steps/search_test.go

**Query Example:**
```graphql
query SearchProducts($input: SearchInput!) {
    searchProducts(input: $input) {
        queryString
        totalCount
        products {
            product {
                id
                name
                price { current currency }
            }
            searchScore
            matchType
        }
        suggestions
        hasMore
    }
}
```

**Tested with Variables:**
```json
{
    "input": {
        "q": "pienas",
        "first": 10
    }
}
```

---

## Summary Table

| Aspect | Status | Details |
|--------|--------|---------|
| **FuzzySearchProducts Method** | IMPLEMENTED | Returns combined similarity scores (name, brand, combined) |
| **Similarity Scores** | FULL | 3 scores returned; combined score used for ranking |
| **Brand Similarity** | PARTIAL | Calculated in scoring (30% weight) but cannot filter by brand |
| **Brand Faceting** | IMPLEMENTED | Shows top 20 brands with counts |
| **ProductMaster Integration** | NOT INTEGRATED | Available via separate query, not loaded in search results |
| **GraphQL API** | COMPLETE | Full SearchInput/SearchResult types with facets |
| **Fuzzy vs Hybrid** | BOTH | Service supports both; resolver can prefer either |
| **Tag Filtering** | IMPLEMENTED | Database-level array overlap filtering |
| **Price/Category/Store Filtering** | IMPLEMENTED | All working with validation |
| **Pagination** | IMPLEMENTED | Limit/offset with hasMore flag |

---

## Recommended Next Steps for Enhancement

1. **Add Brand Filtering**
   - Update SearchInput GraphQL type
   - Add brands field to SearchRequest
   - Pass to SQL function
   - Update resolver

2. **Load ProductMaster**
   - Add relation load in FuzzySearchProducts
   - Decide on deduplication strategy

3. **Fix Match Type**
   - Remove hardcoded "exact" in resolver
   - Use actual match_type from database

4. **Populate Highlights**
   - Implement context snippet extraction
   - Return matched text fragments

5. **Improve Brand-Aware Ranking**
   - Consider exact brand matches higher
   - Add brand-specific boosting

