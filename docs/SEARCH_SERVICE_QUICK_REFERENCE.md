# SearchService - Quick Reference Guide

## Method Signatures

### Main Entry Point
```go
func (s *searchService) SearchProducts(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
```

### Core Search Methods
```go
func (s *searchService) FuzzySearchProducts(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
func (s *searchService) HybridSearchProducts(ctx context.Context, req *SearchRequest) (*SearchResponse, error)
```

### Other Methods
```go
func (s *searchService) GetSearchSuggestions(ctx context.Context, req *SuggestionRequest) (*SuggestionResponse, error)
func (s *searchService) FindSimilarProducts(ctx context.Context, req *SimilarProductsRequest) (*SimilarProductsResponse, error)
func (s *searchService) SuggestQueryCorrections(ctx context.Context, req *CorrectionRequest) (*CorrectionResponse, error)
```

---

## SearchRequest Structure

| Field | Type | Required | Constraints | Notes |
|-------|------|----------|-------------|-------|
| Query | string | Yes | 1-255 chars, no SQL injection | The search query |
| StoreIDs | []int | No | Max 50 stores | Filter by specific stores |
| MinPrice | *float64 | No | 0-10000 EUR | Price filter minimum |
| MaxPrice | *float64 | No | 0-10000 EUR | Price filter maximum |
| OnSaleOnly | bool | No | Default: false | Only sale items |
| Category | string | No | Max 100 chars | Category filter (ILIKE) |
| Tags | []string | No | Array | Tag filter (array overlap) |
| Limit | int | No | 1-100 | Results per page |
| Offset | int | No | 0-10000 | Pagination offset |
| PreferFuzzy | bool | No | Default: false | Boost fuzzy matches in hybrid |

---

## Similarity Scores Explained

### Three-Score Model (FuzzySearch)
The database returns three separate similarity scores (0.0 to 1.0):

1. **name_similarity**: How well product name matches query
   - Weight in final score: 70%
   - Example: "pienas" query vs "Pastatomi pieno produktai" = 0.75

2. **brand_similarity**: How well brand matches query
   - Weight in final score: 30%
   - Example: "pienas" query vs "PIENO KARALIUS" = 0.82
   - Can be 0 if brand is NULL

3. **combined_similarity**: Final ranking score
   - Formula: `(name * 0.7) + (brand * 0.3) + (normalized * 0.2)`
   - Range: 0.0 to 1.0+ (can exceed 1.0 due to normalized bonus)
   - Threshold: 0.3 minimum
   - Used for result ordering

### Two-Score Model (HybridSearch)
Returns one score with match type:

1. **search_score**: Relevance score
   - FTS matches: `ts_rank_cd() * 2.0` (keyword matching)
   - Fuzzy matches: weighted combination (trigram similarity)
   - PreferFuzzy boost: multiplies fuzzy matches by 1.2x
   - Range: 0.0 to 1.0+

2. **match_type**: How the match was found
   - "fts": Full-text search (keyword stemming)
   - "fuzzy": Trigram similarity (typo tolerance)

---

## GraphQL Query Examples

### Basic Search
```graphql
query {
  searchProducts(input: { q: "pienas", first: 10 }) {
    queryString
    totalCount
    products {
      product { id name brand }
      searchScore
      matchType
    }
  }
}
```

### Advanced Search with Filters
```graphql
query {
  searchProducts(input: {
    q: "pienas"
    storeIDs: [1, 2]
    minPrice: 0.5
    maxPrice: 5.0
    onSaleOnly: false
    category: "pieno produktai"
    tags: ["vegan"]
    first: 50
  }) {
    totalCount
    products {
      product {
        id
        name
        brand
        price { current currency }
        store { id name }
      }
      searchScore
      matchType
      similarity
    }
    facets {
      stores { name options { value name count } activeValue }
      brands { options { value name count } }
      priceRanges { options { value name count } }
    }
    pagination { totalItems currentPage totalPages }
  }
}
```

### With PreferFuzzy (Typo Tolerance)
```graphql
query {
  searchProducts(input: {
    q: "pienos"  # Typo: should be "pienas"
    preferFuzzy: true
  }) {
    products {
      searchScore
      matchType  # Will show "fuzzy" matches higher
    }
  }
}
```

---

## Response Structure

### SearchResponse
```go
type SearchResponse struct {
    Products    []ProductSearchResult  // Array of results
    TotalCount  int                    // Total matches (before pagination)
    QueryTime   time.Duration          // Query execution time
    Suggestions []string               // Search suggestions (empty for fuzzy)
    HasMore     bool                   // More results available?
}
```

### ProductSearchResult
```go
type ProductSearchResult struct {
    Product     *models.Product  // Full product object with relations
    SearchScore float64          // Similarity score (0.0-1.0)
    MatchType   string           // "fuzzy", "fts", or "exact"
    Similarity  float64          // Same as SearchScore
    Highlights  []string         // Empty (not yet implemented)
}
```

### GraphQL SearchResult
Includes additional computed fields:
- Facets (brands, categories, stores, prices, availability)
- Pagination info
- Suggestions

---

## Performance Tips

### Query Optimization
1. Use `first` parameter to limit results (default: 50)
2. Combine filters to reduce result set
3. Use store filters when possible (indexed)
4. Price ranges are cheaper than category filters

### When to Use Which Method
| Use Case | Method | PreferFuzzy |
|----------|--------|------------|
| Exact keywords | Hybrid (FTS) | false |
| Typo tolerance | Fuzzy | N/A |
| Mixed approach | Hybrid | true |
| Lithuanian matches | Fuzzy | N/A |
| Product name prefix | Suggestions | N/A |

### Typical Response Times
- Basic search (10 results): 150-250ms
- Complex filters (50 results): 250-400ms
- With facets: +100-200ms
- Large result sets (1000+): 400-600ms

---

## Key Features

### Supported
✓ Full-text search with stemming (Lithuanian)
✓ Fuzzy matching with trigram similarity
✓ Price range filtering
✓ Store filtering (up to 50 stores)
✓ Category filtering (ILIKE)
✓ Tag filtering (array overlap)
✓ On-sale item filtering
✓ Combined similarity scoring
✓ Brand matching (30% weight in score)
✓ Faceted navigation
✓ Search suggestions
✓ Query normalization (Lithuanian diacritics)
✓ Pagination with hasMore flag
✓ Execution time tracking

### Not Supported (Limitations)
✗ Brand filtering (facets only, no filter)
✗ ProductMaster relations in search results
✗ Highlight/context snippets
✗ Query spell correction (in fuzzy)
✗ Brand active value in facets
✗ Cursor-based pagination
✗ Custom sorting options

---

## Database Functions

### fuzzy_search_products()
**When to use**: You want typo tolerance and brand matching

**Parameters**: 
- search_query (TEXT)
- similarity_threshold (FLOAT) = 0.3
- limit_count (INT) = 50
- offset_count (INT) = 0
- store_ids (INT[]) = NULL
- category_filter (TEXT) = NULL
- min_price (DECIMAL) = NULL
- max_price (DECIMAL) = NULL
- on_sale_only (BOOLEAN) = FALSE
- tag_filters (TEXT[]) = NULL

**Returns**: 10 columns including name_similarity, brand_similarity, combined_similarity

```sql
-- Example call
SELECT * FROM fuzzy_search_products(
    'pienas',      -- query
    0.3,           -- threshold
    50,            -- limit
    0,             -- offset
    ARRAY[1,2],    -- stores
    'pieno produktai'  -- category
);
```

### hybrid_search_products()
**When to use**: You want both FTS and fuzzy, with flexible weighting

**Parameters**: Same as fuzzy + prefer_fuzzy (BOOLEAN)

**Returns**: product_id, name, brand, current_price, store_id, flyer_id, search_score, match_type

```sql
-- Example call
SELECT * FROM hybrid_search_products(
    'pienas',
    50,           -- limit
    0,            -- offset
    ARRAY[1,2],   -- stores
    0.5,          -- min_price
    10.0,         -- max_price
    false         -- prefer_fuzzy
);
```

---

## Scoring Formulas

### FuzzySearch Combined Score
```
combined_similarity = 
    (name_similarity * 0.70) +
    (brand_similarity * 0.30) +
    (normalized_name_similarity * 0.20)

Example:
= (0.75 * 0.70) + (0.82 * 0.30) + (0.80 * 0.20)
= 0.525 + 0.246 + 0.160
= 0.931
```

### HybridSearch with PreferFuzzy
```
If match_type = "fts":
    final_score = ts_rank_cd(...) * 2.0 * 1.0

If match_type = "fuzzy":
    final_score = combined_sim * 1.2

Order by: final_score DESC, price ASC, name ASC
```

---

## Common Patterns

### Search with All Filters
```go
req := &search.SearchRequest{
    Query:       "pienas",
    StoreIDs:    []int{1, 2, 3},
    MinPrice:    floatPtr(0.5),
    MaxPrice:    floatPtr(10.0),
    OnSaleOnly:  false,
    Category:    "pieno produktai",
    Tags:        []string{"vegan", "bio"},
    Limit:       50,
    Offset:      0,
    PreferFuzzy: true,
}

resp, err := searchService.SearchProducts(ctx, req)
```

### Pagination
```go
// First page
req.Limit = 10
req.Offset = 0

// Next page
req.Offset = 10  // Increment by limit

// Has more pages?
if resp.HasMore {
    // Load next page
}
```

### Brand-Aware Search (Current - Limited)
```go
// Get brand facets
resp, _ := searchService.SearchProducts(ctx, req)

// Extract brands from facets
for _, option := range resp.Facets.Brands.Options {
    // Show brand with count
    // But cannot filter by brand yet!
}

// Workaround: filter in application after getting results
```

---

## Error Handling

### Common Errors
| Error | Cause | Solution |
|-------|-------|----------|
| "invalid search request" | Query validation failed | Check query length, content |
| "failed to scan fuzzy search result" | Database inconsistency | Retry, check DB integrity |
| "failed to load products with relations" | Product not found | Product may have been deleted |
| "error iterating fuzzy search results" | Database error | Check PostgreSQL logs |
| "failed to compute search facets" | Facet query failed | Returns empty facets, still works |

---

## Testing

### Unit Test Example
```go
func TestFuzzySearchProducts(t *testing.T) {
    svc := search.NewSearchService(db, logger)
    
    req := &search.SearchRequest{
        Query:   "test product",
        Limit:   10,
        Offset:  0,
    }
    
    resp, err := svc.FuzzySearchProducts(context.Background(), req)
    
    if err != nil {
        t.Fatalf("search failed: %v", err)
    }
    
    if len(resp.Products) == 0 {
        t.Fatal("expected products")
    }
    
    // Check scoring
    for i := 0; i < len(resp.Products)-1; i++ {
        if resp.Products[i].SearchScore < resp.Products[i+1].SearchScore {
            t.Fatal("results not sorted by score DESC")
        }
    }
}
```

### GraphQL Integration Test Example
```go
query := `
    query SearchProducts($input: SearchInput!) {
        searchProducts(input: $input) {
            totalCount
            products { product { id } searchScore }
        }
    }
`

variables := map[string]interface{}{
    "input": map[string]interface{}{
        "q":     "test",
        "first": 10,
    },
}

resp, err := client.Query(query, variables)
```

---

## File Locations

| File | Purpose |
|------|---------|
| `internal/services/search/service.go` | Main service implementation |
| `internal/services/search/search.go` | Type definitions |
| `internal/services/search/validation.go` | Input validation & sanitization |
| `internal/graphql/resolvers/query.go:265-376` | GraphQL resolver |
| `internal/graphql/schema/schema.graphql` | GraphQL type definitions |
| `migrations/029_add_tags_to_search_functions.sql` | Database functions |
| `tests/bdd/steps/search_test.go` | Integration tests |
| `tests/bdd/features/search_products.feature` | BDD scenarios |

---

## Related Services

- **ProductService**: Loads full product data
- **StoreService**: Store context for results
- **FlyerService**: Flyer validity checking
- **ProductMasterService**: Master catalog (separate queries)

---

## Future Enhancements

1. Add brand filtering to SearchRequest & SearchInput
2. Load ProductMaster relations in search results
3. Implement highlight/snippet generation
4. Fix matchType to use actual database value (not hardcoded "exact")
5. Add brand to facet active values
6. Implement cursor-based pagination
7. Add query spell correction to fuzzy search

