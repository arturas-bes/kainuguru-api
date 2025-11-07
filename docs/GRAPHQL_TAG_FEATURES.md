# GraphQL Tag Features - Product Tagging System

**Last Updated:** November 7, 2025

## Overview

The Kainuguru API now includes a fully operational product tagging system that enables powerful filtering and categorization of products across all GraphQL queries.

---

## New Features

### 1. **Product Tags Field**

All products now include a `tags` field containing an array of tag strings.

```graphql
type Product {
  # ... other fields
  tags: [String!]
}
```

### 2. **Tag-Based Search Filtering**

The `SearchInput` type now supports tag filtering:

```graphql
input SearchInput {
  q: String!
  tags: [String!]          # NEW: Filter by tags
  storeIDs: [Int!]
  minPrice: Float
  maxPrice: Float
  onSaleOnly: Boolean = false
  category: String
  first: Int = 50
  after: String
  preferFuzzy: Boolean = false
}
```

---

## Available Product Tags

### Dairy Products (`pieno-produktai`)
- `pieno-produktai` - All dairy products
- `skystis` - Liquid products (milk, kefir)
- `fermentuoti` - Fermented products (yogurt, cheese)
- `riebalai` - Fats (butter, cream)
- `sūriai` - Cheese products

### Baked Goods
- `duonos-gaminiai` - Bread products
- `grūdai` - Grain-based products

### Usage Categories
- `kasdieniai` - Daily essentials
- `kepimui` - Baking ingredients

---

## GraphQL Query Examples

### 1. Basic Tag Search

Find all dairy products:

```graphql
query SearchDairyProducts {
  searchProducts(input: {
    q: "pienas",
    tags: ["pieno-produktai"],
    first: 20
  }) {
    totalCount
    products {
      product {
        id
        name
        tags
        price {
          current
        }
      }
    }
  }
}
```

### 2. Multiple Tag Filtering

Find fermented dairy products (AND logic):

```graphql
query SearchFermentedDairy {
  searchProducts(input: {
    q: "jogurtas",
    tags: ["pieno-produktai", "fermentuoti"],
    first: 10
  }) {
    totalCount
    products {
      product {
        id
        name
        tags
        price {
          current
        }
        store {
          name
        }
      }
      searchScore
      matchType
    }
  }
}
```

### 3. Comprehensive Multi-Filter Search

Combine tags with price, store, and other filters:

```graphql
query ComprehensiveSearch {
  searchProducts(input: {
    q: "pienas",
    tags: ["pieno-produktai", "kasdieniai"],
    minPrice: 0.5,
    maxPrice: 2.0,
    storeIDs: [1, 2],  # IKI and Maxima
    onSaleOnly: false,
    first: 20,
    preferFuzzy: true
  }) {
    queryString
    totalCount
    hasMore
    products {
      product {
        id
        name
        brand
        category
        subcategory
        tags
        price {
          current
          original
          discountPercent
          isDiscounted
        }
        unitSize
        unitPrice
        isOnSale
        isAvailable
        store {
          code
          name
        }
        flyer {
          title
          validFrom
          validTo
        }
      }
      searchScore
      matchType
    }
  }
}
```

### 4. Get Product with Tags

Retrieve complete product details including tags:

```graphql
query GetProductDetails($id: Int!) {
  product(id: $id) {
    id
    name
    brand
    category
    subcategory
    tags               # Array of tag strings
    price {
      current
      original
      currency
    }
    store {
      name
      code
    }
  }
}
```

---

## Search Behavior

### Tag Filtering Logic

- **Multiple tags = AND logic**: Product must have ALL specified tags
- **Tags + text query**: Both must match (tag filter AND text search)
- **Empty tags array**: No tag filtering applied
- **Null/missing tags param**: No tag filtering applied

### Examples

| Query | Tags Filter | Result |
|-------|-------------|--------|
| `q: "pienas"` | `["pieno-produktai"]` | Milk products with tag |
| `q: "jogurtas"` | `["fermentuoti", "kasdieniai"]` | Daily fermented products (yogurt) |
| `q: "duona"` | `["kasdieniai"]` | Daily bread products |
| `q: "produktas"` | `["pieno-produktai"]` | All dairy (broad query) |

---

## Lithuanian Language Support

The search system fully supports Lithuanian characters and diacritics:

### Supported Characters
- **ą, č, ę, ė, į, š, ų, ū, ž**

### Example Query

```graphql
query LithuanianSearch {
  searchProducts(input: {
    q: "sūris",           # Lithuanian word for cheese with ū
    first: 10,
    preferFuzzy: true
  }) {
    products {
      product {
        id
        name               # "Sūris lietuviškas"
        normalizedName     # "suris lietuviškas"
        tags
      }
      searchScore
    }
  }
}
```

### Text Normalization

- **Original**: `Sūris lietuviškas`
- **Normalized**: `suris lietuviškas`
- Both forms are searchable with fuzzy matching

---

## Search Modes

### 1. **Hybrid Search (Default)**

Combines full-text search with fuzzy matching:

```graphql
{
  searchProducts(input: {
    q: "pienas",
    preferFuzzy: false  # Default
  }) { ... }
}
```

### 2. **Fuzzy Search Mode**

Prioritizes fuzzy/similarity matching:

```graphql
{
  searchProducts(input: {
    q: "piens",         # Typo in "pienas"
    preferFuzzy: true   # Enable fuzzy mode
  }) { ... }
}
```

---

## Response Structure

### SearchResult Type

```graphql
type SearchResult {
  queryString: String!
  totalCount: Int!
  hasMore: Boolean!
  products: [ProductSearchResult!]!
  suggestions: [String!]
  facets: SearchFacets!
  pagination: Pagination!
}
```

### ProductSearchResult

```graphql
type ProductSearchResult {
  product: Product!
  searchScore: Float!     # Relevance score (0-1)
  matchType: String!      # "fts" or "fuzzy"
  similarity: Float
  highlights: [String!]
}
```

---

## Performance

| Operation | Avg Response Time | Status |
|-----------|-------------------|--------|
| Basic search | < 10ms | ✅ Excellent |
| Fuzzy search | < 20ms | ✅ Very Good |
| Tag filtering | < 10ms | ✅ Excellent |
| Multi-filter search | < 130ms | ✅ Good |

---

## Insomnia Collection

### Import Instructions

1. **Standalone Tag Collection**: Import `kainuguru-graphql-insomnia-collection.json`
2. **Full Collection**: Use `kainuguru-insomnia.json` (includes all endpoints)

### New Requests in Collection

1. **Search with Tags Filter** - Basic tag search example
2. **Search with Multiple Tags** - AND logic demonstration
3. **Comprehensive Search (All Filters)** - Advanced multi-filter example
4. **Lithuanian Characters Search** - Diacritic handling test
5. **Get Product with Tags** - Complete product retrieval

### Environment Variables

Set in Insomnia:

```json
{
  "base_url": "http://localhost:8080",
  "graphql_endpoint": "http://localhost:8080/graphql"
}
```

---

## Tag Management

### Current Tag Schema

Tags are stored as PostgreSQL `TEXT[]` arrays:

```sql
CREATE TABLE products (
    -- ...
    tags TEXT[],
    -- ...
);
```

### Adding Tags to Products

Tags can be set during product creation or updated:

```sql
UPDATE products
SET tags = ARRAY['pieno-produktai', 'skystis', 'kasdieniai']
WHERE name ILIKE '%pienas%';
```

### Tag Indexing

For performance optimization, a GIN index can be added:

```sql
CREATE INDEX idx_products_tags_gin
ON products USING gin(tags);
```

---

## Migration History

| Migration | Description | Status |
|-----------|-------------|--------|
| 027 | Added `subcategory` column | ✅ Applied |
| 028 | Added missing product columns | ✅ Applied |
| 029 | Added tag filtering to search functions | ✅ Applied |

---

## Error Handling

### Common Issues

**1. Empty Results with Tags**
- Ensure tag names match exactly (case-sensitive)
- Check if products have the specified tags

**2. Tag Array Format**
```graphql
# ✅ Correct
tags: ["pieno-produktai", "kasdieniai"]

# ❌ Wrong
tags: "pieno-produktai, kasdieniai"
```

**3. No Tag Filtering**
```graphql
# Both are valid for no filtering:
tags: []
tags: null
# Omit tags parameter entirely
```

---

## API Endpoints

### GraphQL Endpoint
```
POST http://localhost:8080/graphql
```

### Health Check
```
GET http://localhost:8080/health
```

### GraphQL Playground
```
GET http://localhost:8080/playground
```

---

## Sample Tag Categories

### Recommended Tag Structure

```
Product Categories:
├── pieno-produktai (Dairy)
│   ├── skystis (Liquid)
│   ├── fermentuoti (Fermented)
│   ├── riebalai (Fats)
│   └── sūriai (Cheese)
├── duonos-gaminiai (Baked Goods)
│   └── grūdai (Grains)
└── Usage Tags
    ├── kasdieniai (Daily Essentials)
    └── kepimui (For Baking)
```

---

## Testing

### cURL Examples

**Basic tag search:**
```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { searchProducts(input: {q: \"pienas\", tags: [\"pieno-produktai\"]}) { totalCount products { product { id name tags } } } }"
  }'
```

**Multi-filter search:**
```bash
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { searchProducts(input: {q: \"pienas\", tags: [\"pieno-produktai\", \"kasdieniai\"], minPrice: 0.5, maxPrice: 2.0, storeIDs: [1]}) { totalCount } }"
  }'
```

---

## Future Enhancements

### Planned Features

1. **Tag Facets** - Show available tags in search results
2. **Tag Suggestions** - Auto-suggest tags based on product name
3. **Tag Hierarchy** - Parent-child tag relationships
4. **Tag Analytics** - Most popular tags, trending tags
5. **Normalized Tag System** - Junction table for referential integrity

---

## Support

For issues or questions:
- Check system health: `GET /health`
- View GraphQL schema: `GET /playground`
- Review logs: `docker-compose logs api`

---

**Document Version:** 1.0
**Last Tested:** November 7, 2025
**API Version:** Latest (29 migrations applied)
