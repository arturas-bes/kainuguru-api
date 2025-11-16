# SearchService Architecture Diagram & Data Flow

## System Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     GraphQL API Layer                            │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  SearchProducts Query (query.go:265)                     │   │
│  │  Input: SearchInput (q, storeIDs, minPrice, etc.)        │   │
│  └──────────────────────────────────────────────────────────┘   │
└────────────────────────────┬─────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              Resolver Layer (query.go:265-376)                   │
│                                                                  │
│  1. Convert GraphQL SearchInput → search.SearchRequest           │
│  2. Call searchService.SearchProducts()                          │
│  3. Compute facets via computeSearchFacets()                     │
│  4. Convert response to SearchResult                             │
└────────────────────────────┬─────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│         Service Layer (search/service.go:27-40)                  │
│                                                                  │
│  SearchProducts()                                                │
│    ├─ Validate request                                           │
│    ├─ Sanitize query                                             │
│    └─ Route based on PreferFuzzy flag:                           │
│        ├─ TRUE  → FuzzySearchProducts()                          │
│        └─ FALSE → HybridSearchProducts()                         │
└────────────────────────────┬─────────────────────────────────────┘
                              │
                 ┌────────────┴────────────┐
                 │                         │
                 ▼                         ▼
    ┌────────────────────────┐  ┌────────────────────────┐
    │  FuzzySearchProducts   │  │ HybridSearchProducts   │
    │   (service.go:42)      │  │  (service.go:171)      │
    └────────────────────────┘  └────────────────────────┘
                 │                         │
                 └────────────┬────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│              Database Layer - PostgreSQL                         │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │ fuzzy_search_products()  OR  hybrid_search_products()    │   │
│  │ (Both: migrations/029_add_tags_to_search_functions.sql) │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│         ┌────────────────────┼────────────────────┐              │
│         │                    │                    │              │
│         ▼                    ▼                    ▼              │
│  ┌─────────────┐    ┌──────────────┐    ┌──────────────┐       │
│  │   Products  │    │    Flyers    │    │ Trigram Index│       │
│  │   (search   │    │ (availability)    │ (similarity) │       │
│  │  _vector)   │    │                  │              │       │
│  └─────────────┘    └──────────────┘    └──────────────┘       │
└─────────────────────────────────────────────────────────────────┘
```

---

## FuzzySearchProducts Data Flow

```
GraphQL Request
  │
  │ SearchInput {
  │   q: "pienas",
  │   storeIDs: [1, 2],
  │   minPrice: 0.5,
  │   maxPrice: 10.0,
  │   onSaleOnly: false,
  │   preferFuzzy: true
  │ }
  │
  ▼
Request Validation
  ├─ Query length: 1-255 ✓
  ├─ Price range valid ✓
  ├─ Store IDs valid ✓
  └─ No SQL injection ✓
  │
  ▼
Query Sanitization (SanitizeQuery)
  ├─ Trim whitespace
  ├─ Remove control characters
  └─ Normalize multiple spaces
  │
  ▼
FuzzySearchProducts()
  │
  ├─ Call fuzzy_search_products() SQL function with:
  │   ├─ query: "pienas"
  │   ├─ similarity_threshold: 0.3
  │   ├─ store_ids: {1, 2}
  │   ├─ min_price: 0.5
  │   ├─ max_price: 10.0
  │   ├─ on_sale_only: false
  │   └─ tag_filters: null
  │
  └─ Database processes:
     │
     ├─ Normalize query: "pienas" → normalized form
     │
     ├─ For each product in database:
     │   │
     │   ├─ name_similarity = similarity(name, "pienas")
     │   │   Example: "Pastatomi pieno produktai" → 0.75
     │   │
     │   ├─ brand_similarity = similarity(brand, "pienas")
     │   │   Example: "PIENO KARALIUS" → 0.82
     │   │
     │   ├─ combined_similarity =
     │   │   (0.75 * 0.7) +           // name 70%
     │   │   (0.82 * 0.3) +           // brand 30%
     │   │   (normalized * 0.2)       // diacritic 20% bonus
     │   │   = 0.525 + 0.246 + 0.15
     │   │   = 0.921
     │   │
     │   └─ Apply filters:
     │       ├─ Flyer is valid (not archived, within dates) ✓
     │       ├─ Product is available ✓
     │       ├─ Store in {1, 2} ✓
     │       ├─ Price between 0.5-10.0 ✓
     │       ├─ Not on sale (onSaleOnly=false) ✓
     │       └─ Meets threshold (>= 0.3) ✓
     │
     ├─ Keep only matches above threshold
     │
     └─ Sort by: combined_similarity DESC, price ASC, name ASC
        Result set returned
        │
        ▼
Go Service Processing:
  │
  ├─ Extract productIDs from results
  │
  ├─ Build scoresMap: map[int]float64
  │   {
  │     123: 0.921,
  │     456: 0.875,
  │     789: 0.650
  │   }
  │
  ├─ Load full products with relations:
  │   SELECT p.* FROM products p
  │   WHERE p.id IN (123, 456, 789)
  │
  │   Load relations:
  │   ├─ Store (required for price context)
  │   ├─ Flyer (validity info)
  │   └─ FlyerPage (image reference)
  │
  ├─ Set currency to "EUR" for all products
  │
  ├─ Build results array:
  │   {
  │     {
  │       Product: { id: 123, name: "Pastatomi pieno...", ... },
  │       SearchScore: 0.921,
  │       MatchType: "fuzzy",
  │       Similarity: 0.921
  │     },
  │     {
  │       Product: { id: 456, name: "Pieno gaminys...", ... },
  │       SearchScore: 0.875,
  │       MatchType: "fuzzy",
  │       Similarity: 0.875
  │     }
  │   }
  │
  ├─ Get total count via second SQL query
  │   SELECT COUNT(*) FROM fuzzy_search_products(...)
  │
  └─ Log analytics asynchronously
  │
  ▼
Return SearchResponse
  {
    Products: [...],
    TotalCount: 1234,
    QueryTime: 145ms,
    HasMore: true
  }
  │
  ▼
Resolver: computeSearchFacets()
  │
  ├─ Query 1: Stores facet (group by store_id, count)
  │   Result: [{ id: 1, name: "Lidl", count: 45 }, ...]
  │
  ├─ Query 2: Categories facet (group by category, count)
  │   Result: [{ name: "Pieno produktai", count: 38 }, ...]
  │
  ├─ Query 3: Brands facet (group by brand, count, limit 20)
  │   Result: [{ name: "PIENO KARALIUS", count: 12 }, ...]
  │
  ├─ Query 4: Price ranges (predefined buckets)
  │   Result: [{ range: "0-1", count: 5 }, { range: "1-3", count: 28 }, ...]
  │
  └─ Query 5: Availability (on-sale vs regular)
     Result: [{ label: "On Sale", count: 12 }, { label: "Regular", count: 78 }]
  │
  ▼
GraphQL Response
{
  "data": {
    "searchProducts": {
      "queryString": "pienas",
      "totalCount": 1234,
      "products": [
        {
          "product": {
            "id": 123,
            "name": "Pastatomi pieno produktai",
            "brand": "PIENO KARALIUS",
            "price": { "current": 2.49, "currency": "EUR" },
            ...
          },
          "searchScore": 0.921,
          "matchType": "fuzzy",
          "similarity": 0.921,
          "highlights": []
        },
        ...
      ],
      "facets": {
        "stores": { "options": [...], "activeValue": ["1", "2"] },
        "categories": { "options": [...], "activeValue": [] },
        "brands": { "options": [...], "activeValue": [] },
        "priceRanges": { "options": [...], "activeValue": [] },
        "availability": { "options": [...], "activeValue": [] }
      },
      "hasMore": true
    }
  }
}
```

---

## Score Calculation Deep Dive

### FuzzySearch Score Composition

```
For Product: "Pastatomi pieno produktai" by brand "PIENO KARALIUS"
Query: "pienas"

Step 1: Calculate Individual Similarities
  ├─ name_similarity = similarity("Pastatomi pieno produktai", "pienas")
  │  Uses PostgreSQL trigram algorithm
  │  Result: 0.75 (75% match - "pieno" is in the name)
  │
  ├─ brand_similarity = similarity("PIENO KARALIUS", "pienas")
  │  Result: 0.82 (82% match - "PIENO" matches well)
  │
  └─ normalized_similarity = similarity(
       "pastatomi_pieno_produktai",  // Lithuanian normalized
       "pienas"
     )
     Result: 0.80 (80% match with diacritics removed)

Step 2: Calculate Combined Score
  combined_similarity = 
    (name_similarity * 0.7) +
    (brand_similarity * 0.3) +
    (normalized_similarity * 0.2)
    
  = (0.75 * 0.7) + (0.82 * 0.3) + (0.80 * 0.2)
  = 0.525 + 0.246 + 0.160
  = 0.931

Step 3: Score Range & Sorting
  ├─ Minimum threshold: 0.3 (products scoring < 0.3 excluded)
  ├─ This product's score: 0.931 (well above threshold)
  └─ Sorted position: Ranked by score DESC
     Higher scores = better matches = shown first

Scoring Rationale:
  • Name is most important (70%) - what the product is called
  • Brand secondary (30%) - manufacturer/brand signal
  • Normalized bonus (20%) - helps with Lithuanian diacritics
  • Total can exceed 1.0 due to bonus weighting
```

### HybridSearch Score Composition

```
Same product, but using HybridSearch (preferFuzzy: false)

Step 1: FTS (Full-Text Search) Check
  ├─ Query: plainto_tsquery('lithuanian', "pienas")
  │  Converts to: 'pien':* (stemmed)
  │
  ├─ Check: product.search_vector @@ 'pien':*
  │  If product has "pieno", "pienas", "pienai" → MATCH
  │
  └─ FTS Score:
     ts_rank_cd(search_vector, query_tsquery) * 2.0
     = (0.4 * 2.0)  // ts_rank_cd returns 0-1
     = 0.8
     Match type: "fts"

Step 2: Trigram Fallback
  If not found by FTS, apply fuzzy:
  ├─ similarity(name, "pienas") = 0.75
  ├─ similarity(brand, "pienas") = 0.82
  └─ Combined = 0.931
     Match type: "fuzzy"

Step 3: Preference Boost
  If preferFuzzy = true:
  ├─ FTS Score: 0.8 * 1.0 = 0.8
  ├─ Fuzzy Score: 0.931 * 1.2 = 1.117 (boosted!)
  └─ Final ranking prioritizes fuzzy matches

Step 4: Final Ranking
  ORDER BY search_score DESC, price ASC, name ASC
  
  If two products:
    Product A: FTS match, score 0.8
    Product B: Fuzzy match, score 0.931 * 1.2 = 1.117
    
  With preferFuzzy=true → Product B wins
  With preferFuzzy=false → Still Product B (0.931 > 0.8)
```

---

## ProductMaster Relationship Diagram

```
Current State (NOT INTEGRATED)
┌──────────────┐
│   Products   │  (Search returns these - store-specific instances)
│              │
│ - id: 123    │
│ - name: ...  │
│ - brand: ...  │
│ - store_id: 1│
└──────┬───────┘
       │
       │ (Optional relation - NOT loaded in search)
       │
       ▼
┌──────────────────┐
│ ProductMaster    │  (Master record - available via separate query)
│                  │
│ - id: 999        │
│ - canonical_name │
│ - category       │
│ - attributes     │
└──────────────────┘

Desired State (IF IMPLEMENTED)
Search Results would include:
{
  "product": {
    "id": 123,           ← Individual store product
    "name": "Pienas",
    "store": { "id": 1 }
  },
  "productMaster": {      ← Master record
    "id": 999,
    "canonicalName": "Pienas",
    "category": "Dairy"
  },
  "searchScore": 0.921
}

Benefits:
• Deduplication - show "one" product per master
• Canonical data - consistent naming
• Shopping list linking - reference master, not store-specific
• Price comparison - easier to show all store prices for same master product
```

---

## Filtering Architecture

```
SearchRequest Input
    │
    ├─ Query: "pienas"
    │
    ├─ Filters:
    │   ├─ StoreIDs: [1, 2, 3]
    │   │   └─ Applied: WHERE store_id = ANY($1)
    │   │
    │   ├─ Category: "pieno produktai"
    │   │   └─ Applied: WHERE category ILIKE ('%' || $2 || '%')
    │   │
    │   ├─ MinPrice: 0.5, MaxPrice: 10.0
    │   │   └─ Applied: WHERE price BETWEEN $3 AND $4
    │   │
    │   ├─ OnSaleOnly: false
    │   │   └─ Applied: WHERE (NOT $5 OR is_on_sale = TRUE)
    │   │
    │   ├─ Tags: ["vegan", "organic"]
    │   │   └─ Applied: WHERE tags && $6  (array overlap)
    │   │
    │   ├─ Limit: 50, Offset: 0
    │   │   └─ Applied: LIMIT 50 OFFSET 0
    │   │
    │   └─ PreferFuzzy: true (only in HybridSearch)
    │       └─ Applied: CASE WHEN match_type='fuzzy' THEN score*1.2 ELSE score END
    │
    └─ Always Applied Filters:
        ├─ Flyer.is_archived = FALSE
        ├─ Flyer.valid_from <= NOW()
        ├─ Flyer.valid_to >= NOW()
        └─ Product.is_available = TRUE

All filters combined in WHERE clause at database level
→ Minimal result set returned to application
→ Performance optimized
```

---

## Key Data Structures

```go
// Input from GraphQL
SearchInput {
    Q:           "pienas"
    StoreIDs:    [1, 2]
    MinPrice:    0.5
    MaxPrice:    10.0
    OnSaleOnly:  false
    Category:    "pieno produktai"
    Tags:        ["vegan"]
    First:       50
    PreferFuzzy: true
}

    │ Converter (resolver)
    ▼

// Service request
SearchRequest {
    Query:       "pienas"
    StoreIDs:    [1, 2]
    MinPrice:    &0.5
    MaxPrice:    &10.0
    OnSaleOnly:  false
    Category:    "pieno produktai"
    Tags:        ["vegan"]
    Limit:       50
    Offset:      0
    PreferFuzzy: true
}

    │ Service processing
    ▼

// Database returns
fuzzy_search_products returns:
[
    {
        product_id: 123,
        name: "Pastatomi pieno produktai",
        brand: "PIENO KARALIUS",
        category: "Dairy",
        current_price: 2.49,
        store_id: 1,
        flyer_id: 456,
        name_similarity: 0.75,
        brand_similarity: 0.82,
        combined_similarity: 0.931
    },
    ...
]

    │ Service builds results
    ▼

// Application response
SearchResponse {
    Products: [
        ProductSearchResult {
            Product: &Product{...},      // Full object with relations
            SearchScore: 0.931,
            MatchType: "fuzzy",
            Similarity: 0.931,
            Highlights: []
        },
        ...
    ],
    TotalCount: 1234,
    QueryTime: 145ms,
    HasMore: true
}

    │ Resolver computes facets
    ▼

// GraphQL response
SearchResult {
    QueryString: "pienas",
    TotalCount: 1234,
    Products: [...],
    Suggestions: ["pieno produktai", "pieno gaminys"],
    HasMore: true,
    Facets: {
        Stores: {...},
        Categories: {...},
        Brands: {...},
        PriceRanges: {...},
        Availability: {...}
    },
    Pagination: {...}
}
```

---

## Performance Characteristics

```
Query Speed Factors:

1. Database Operations: ~100-200ms
   ├─ fuzzy_search_products() function: 80-150ms
   │  (Similarity calculations on all rows, filtering, sorting)
   │
   ├─ Load products relations: 20-50ms
   │  (SELECT ... WHERE id IN (...))
   │
   └─ Count query: 5-10ms
      (Second query for total count)

2. Facet Computation: ~100-300ms (parallel)
   ├─ Stores query: 20-50ms
   ├─ Categories query: 15-30ms
   ├─ Brands query: 15-30ms
   ├─ Price ranges query: 15-30ms
   └─ Availability query: 10-20ms

3. Application Processing: ~10-50ms
   ├─ Score mapping: 5ms
   ├─ Result building: 5-20ms
   ├─ Relation assignment: 5-20ms
   └─ Analytics logging: < 1ms

Total Request Time: 200-550ms (depending on:
  - Result set size
  - Number of relations
  - Facet complexity
  - Database load)

Optimization Applied:
✓ Similarity calculations in database (not application)
✓ Filtering at database level
✓ LIMIT/OFFSET at database
✓ Facet queries can run in parallel (in production)
✓ Trigram indexes for fast similarity matching
✓ FTS vector indexes for keyword matching
✓ Tag filtering via array overlap (indexed)
```

---

## Testing Coverage

```
Unit Tests:
├─ SearchRequest validation (search/validation_test.go)
│  ├─ Query validation (length, chars, injection)
│  ├─ Price range validation
│  ├─ Pagination validation
│  └─ Store IDs validation
│
├─ Query sanitization
│  ├─ Whitespace trimming
│  ├─ Control character removal
│  └─ Multiple space normalization
│
└─ Search score calculation
   ├─ Combined similarity formula
   ├─ Weighting ratios
   └─ Threshold application

Integration Tests: (tests/bdd/steps/search_test.go)
├─ Real GraphQL API search
│  ├─ Basic search: query + limit
│  ├─ Filtered search: price, store, category
│  ├─ Response structure verification
│  └─ Result ordering by score
│
└─ Performance testing
   ├─ Response time measurement
   ├─ Large result set handling
   └─ Concurrent query handling

BDD Tests: (tests/bdd/features/search_products.feature)
├─ Scenario: Search for products
├─ Scenario: Search with filters
├─ Scenario: Search suggestions
└─ Scenario: Lithuanian character handling
```

