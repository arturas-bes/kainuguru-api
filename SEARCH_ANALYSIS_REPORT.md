# Search Endpoints Deep Analysis Report

**Date**: 2025-11-19
**Status**: ✅ All search types verified and working correctly

## Executive Summary

Comprehensive analysis of all search endpoints in the Kainuguru API revealed that **all search types are functioning correctly** after fixing the `is_on_sale` column issue. The system implements a sophisticated hybrid search strategy combining Full-Text Search (FTS) and fuzzy matching for Lithuanian text.

---

## Search Architecture

### 1. Search Types

The API supports three primary search strategies:

#### **A. Hybrid Search** (Default)
- **Entry Point**: `SearchProducts(ctx, req)`
- **Database Function**: `hybrid_search_products()`
- **Strategy**: Combines FTS with trigram-based fuzzy matching
- **Parameters**: 10 parameters including query, filters, pagination
- **Returns**: Products with `search_score` and `match_type` (fts/fuzzy)

#### **B. Fuzzy Search** (Optional)
- **Entry Point**: `FuzzySearchProducts(ctx, req)`
- **Database Function**: `fuzzy_search_products()`
- **Strategy**: Pure trigram similarity matching
- **Parameters**: 10 parameters including similarity_threshold (default: 0.3)
- **Returns**: Products with similarity scores (name, brand, combined)

#### **C. Full-Text Search** (via Hybrid)
- **Embedded in**: Hybrid search (FTS results prioritized)
- **Language**: Lithuanian text search configuration
- **Features**: Stemming, stop words, normalized Lithuanian characters

### 2. GraphQL Schema

```graphql
type Query {
  searchProducts(input: SearchInput!): SearchResult!
}

input SearchInput {
  q: String!                    # Search query (required)
  storeIDs: [Int!]             # Filter by stores
  minPrice: Float              # Price range minimum
  maxPrice: Float              # Price range maximum
  onSaleOnly: Boolean = false  # Only on-sale products
  category: String             # Category filter
  tags: [String!]              # Tag filters (array overlap)
  first: Int = 50              # Pagination limit
  after: String                # Cursor-based pagination
  preferFuzzy: Boolean = false # Force fuzzy search
}

type SearchResult {
  products: [ProductSearchResult!]!
  totalCount: Int!
  queryString: String!
  suggestions: [String!]
  hasMore: Boolean!
  facets: SearchFacets!
  pagination: Pagination!
}
```

---

## Database Functions Analysis

### 1. hybrid_search_products()

**Signature**:
```sql
hybrid_search_products(
    search_query TEXT,
    limit_count INTEGER DEFAULT 50,
    offset_count INTEGER DEFAULT 0,
    store_ids INTEGER[] DEFAULT NULL,
    min_price NUMERIC DEFAULT NULL,
    max_price NUMERIC DEFAULT NULL,
    prefer_fuzzy BOOLEAN DEFAULT false,
    category_filter TEXT DEFAULT NULL,
    on_sale_only BOOLEAN DEFAULT false,
    tag_filters TEXT[] DEFAULT NULL
)
RETURNS TABLE(
    product_id BIGINT,
    name TEXT,
    brand TEXT,
    current_price NUMERIC,
    store_id INTEGER,
    flyer_id INTEGER,
    search_score FLOAT,
    match_type TEXT
)
```

**Implementation Details**:
- Uses `plainto_tsquery('lithuanian', search_query)` for FTS
- Applies `normalize_lithuanian_text()` for consistent matching
- Combines FTS (ts_rank_cd × 2.0) with trigram similarity
- Filters: valid flyers, available products, store/price/category/tags
- ✅ **Correctly uses `p.is_on_sale` filter** (line 77, 90)

### 2. fuzzy_search_products()

**Signature**:
```sql
fuzzy_search_products(
    search_query TEXT,
    similarity_threshold FLOAT DEFAULT 0.3,
    limit_count INTEGER DEFAULT 50,
    offset_count INTEGER DEFAULT 0,
    store_ids INTEGER[] DEFAULT NULL,
    category_filter TEXT DEFAULT NULL,
    min_price NUMERIC DEFAULT NULL,
    max_price NUMERIC DEFAULT NULL,
    on_sale_only BOOLEAN DEFAULT false,
    tag_filters TEXT[] DEFAULT NULL
)
RETURNS TABLE(
    product_id BIGINT,
    name TEXT,
    brand TEXT,
    category TEXT,
    current_price NUMERIC,
    store_id INTEGER,
    flyer_id INTEGER,
    name_similarity FLOAT,
    brand_similarity FLOAT,
    combined_similarity FLOAT
)
```

**Implementation Details**:
- Uses PostgreSQL `similarity()` function (pg_trgm extension)
- Weighted scoring: name (0.7) + brand (0.3) + normalized (0.2)
- Matches name, brand, or combined name+brand
- Same filters as hybrid search
- ✅ **Correctly uses `p.is_on_sale` filter**

---

## Test Results

### Basic Functionality Tests

#### ✅ Test 1: Hybrid Search (Default)
```bash
Query: "pienas"
Results: 4 products
Match Type: "fts"
Search Score: 0.4000000059604645
```

#### ✅ Test 2: Fuzzy Search
```bash
Query: "pienas" with preferFuzzy: true
Results: 4 products
Match Type: "fuzzy"
Combined Similarity: 0.3315789431333542
```

#### ✅ Test 3: On-Sale Filter
```bash
Query: "pienas" with onSaleOnly: true
Results: 1 product (id: 2)
Verification: Only product marked as is_on_sale=TRUE returned
```

#### ✅ Test 4: Price Range Filter
```bash
Query: "pienas" with minPrice: 0.50, maxPrice: 1.00
Results: 4 products (all with price €0.89)
```

#### ✅ Test 5: Lithuanian Text
```bash
Query: "duona" (Lithuanian for "bread")
Results: 4 products
Match Type: "fts"
Verification: Properly handles Lithuanian characters
```

#### ✅ Test 6: Tag Filtering
```bash
Query: "pienas" with tags: ["pieno-produktai"]
Results: 2 products (ids: 2, 17)
Verification: Array overlap operator (&&) works correctly
```

#### ✅ Test 7: Error Handling
```bash
Query: "" (empty string)
Result: Validation error
Message: "query cannot be empty"
```

### Facets Tests

#### ✅ Test 8: Search Facets
```json
{
  "stores": {
    "options": [{"value": "1", "name": "IKI", "count": 4}]
  },
  "brands": {
    "options": []
  },
  "availability": {
    "options": [
      {"value": "on_sale", "count": 1},
      {"value": "regular", "count": 3}
    ]
  }
}
```

---

## Service Layer Implementation

### Search Service (`internal/services/search/service.go`)

**Key Methods**:

1. **SearchProducts** (line 27-40)
   - Entry point that routes to Fuzzy or Hybrid based on `PreferFuzzy` flag
   - Validates and sanitizes input
   - Returns `SearchResponse` with products, count, time, suggestions

2. **FuzzySearchProducts** (line 42-131)
   - Calls `fuzzy_search_products()` DB function
   - Loads full product relations (Store, Flyer, FlyerPage)
   - Sets EUR currency for all products
   - Computes total count
   - Determines hasMore flag

3. **HybridSearchProducts** (line 133-295)
   - Calls `hybrid_search_products()` DB function
   - Similar flow to Fuzzy search
   - Returns products with match_type differentiation

### GraphQL Resolver (`internal/graphql/resolvers/query.go`)

**SearchProducts Resolver** (line 269-385):
- Converts GraphQL input to service SearchRequest
- Handles optional parameters (onSaleOnly, category, preferFuzzy)
- Computes search facets (stores, categories, brands, price ranges, availability)
- Builds pagination info
- Returns SearchResult with all metadata

**computeSearchFacets** (line 475-649):
- ✅ **Correctly applies `p.is_on_sale` filter** (line 507)
- Groups results by store, category, brand
- Computes price range buckets
- Calculates availability counts (on_sale vs regular)

---

## Key Fixes Applied

### Issue: Missing `is_on_sale` Column

**Problem**:
- Database functions referenced `p.is_on_sale` column
- Products table didn't have this column
- GraphQL queries failed with "column p.is_on_sale does not exist"

**Solution** (Migration 036):
1. Added `is_on_sale BOOLEAN DEFAULT FALSE` column
2. Populated existing products based on discount data
3. Created index `idx_products_on_sale` for performance
4. Recreated `hybrid_search_products()` function with 10 parameters (added `tag_filters`)
5. Fixed Go model to use `discount_percentage` (DB column name)

**Files Modified**:
- `/migrations/036_add_is_on_sale_to_products.sql` (created)
- `/internal/models/product.go:34` (fixed bun tag)

---

## Performance Considerations

### Indexes Used

1. **FTS Indexes**:
   - `idx_products_search_vector` (GIN index on search_vector)
   - Enables fast full-text search on Lithuanian text

2. **Trigram Indexes**:
   - `idx_products_name_trigram` (GIN index on name)
   - `idx_products_normalized_trigram` (GIN index on normalized_name)
   - `idx_products_brand_trigram` (GIN index on brand)
   - Enable fuzzy matching with similarity threshold

3. **Filter Indexes**:
   - `idx_products_on_sale` (B-tree on is_on_sale, valid_from, valid_to)
   - `idx_products_store_valid` (B-tree on store_id, valid_from, valid_to)
   - `idx_products_price_range` (B-tree on current_price)
   - Enable efficient filtering

### Query Optimization

- **Hybrid search** uses CTE (Common Table Expressions) for clean separation
- **NOT EXISTS** clause prevents duplicate results from FTS and fuzzy
- **ORDER BY** prioritizes search_score, then price, then name
- **LIMIT/OFFSET** for efficient pagination

---

## Recommendations

### ✅ Already Implemented
- Hybrid search with FTS + fuzzy matching
- Lithuanian text normalization
- Tag filtering with array operators
- Comprehensive facets
- Proper error handling and validation

### 🎯 Potential Enhancements

1. **Query Suggestions**:
   - Implement `GetSearchSuggestions()` for autocomplete
   - Use trigram similarity on popular queries

2. **Search Analytics**:
   - Implement `LogSearchAnalytics()` to track:
     - Popular queries
     - Zero-result queries
     - Query performance metrics

3. **Query Corrections**:
   - Implement `SuggestQueryCorrections()` for typo handling
   - Use Levenshtein distance for "Did you mean?"

4. **Caching**:
   - Cache popular search queries in Redis
   - Invalidate on product/flyer updates

5. **Relevance Tuning**:
   - A/B test different scoring weights
   - Track click-through rates
   - Adjust FTS vs fuzzy balance

---

## Verification Commands

### Database Function Tests
```bash
# Test hybrid search
psql -c "SELECT product_id, name, search_score, match_type FROM hybrid_search_products('pienas', 5, 0, NULL, NULL, NULL, FALSE, NULL, FALSE, NULL);"

# Test fuzzy search
psql -c "SELECT product_id, name, combined_similarity FROM fuzzy_search_products('pienas', 0.3, 5, 0, NULL, NULL, NULL, NULL, FALSE, NULL);"

# Test on-sale filter
psql -c "SELECT product_id, name FROM hybrid_search_products('pienas', 10, 0, NULL, NULL, NULL, FALSE, NULL, TRUE, NULL);"
```

### GraphQL Tests
```bash
# Run comprehensive tests
./test_search_graphql.sh

# Test facets
./test_facets.sh

# Test tags
./test_tags.sh
```

---

## Conclusion

**Status**: ✅ **All search functionality verified and working correctly**

The Kainuguru API implements a robust, production-ready search system with:
- Multiple search strategies (hybrid, fuzzy, FTS)
- Comprehensive filtering (stores, price, categories, tags, on-sale)
- Rich faceting for UI refinement
- Proper Lithuanian text handling
- Efficient indexing and query optimization
- GraphQL integration with pagination and metadata

The recent fix for the `is_on_sale` column resolved the final blocker, and all search types now function as designed.

**Next Steps**: Consider implementing suggested enhancements for search analytics, query suggestions, and relevance tuning based on user behavior.
