# Insomnia Testing Guide - Phase 1.2

## 🚀 Quick Start

**API Endpoint**: `http://localhost:8080/graphql`
**Content-Type**: `application/json`

---

## ✅ Working Queries (Corrected for Current Schema)

### 1. Query All Stores

```graphql
query GetStores {
  stores(first: 10) {
    totalCount
    pageInfo {
      hasNextPage
      endCursor
    }
    edges {
      cursor
      node {
        id
        code
        name
        logoURL
        websiteURL
        isActive
        locations {
          city
          address
        }
      }
    }
  }
}
```

**Insomnia JSON**:
```json
{
  "query": "query GetStores { stores(first: 10) { totalCount pageInfo { hasNextPage endCursor } edges { cursor node { id code name logoURL websiteURL isActive locations { city address } } } } }"
}
```

---

### 2. Get Store by Code

```graphql
query GetStoreByCode($code: String!) {
  storeByCode(code: $code) {
    id
    code
    name
    logoURL
    websiteURL
    isActive
    flyers(first: 5, filters: { isCurrent: true }) {
      totalCount
      edges {
        node {
          id
          title
          validFrom
          validTo
          isValid
          isCurrent
          daysRemaining
        }
      }
    }
  }
}
```

**Variables**:
```json
{
  "code": "iki"
}
```

**Insomnia JSON**:
```json
{
  "query": "query GetStoreByCode($code: String!) { storeByCode(code: $code) { id code name logoURL websiteURL isActive flyers(first: 5, filters: { isCurrent: true }) { totalCount edges { node { id title validFrom validTo isValid isCurrent daysRemaining } } } } }",
  "variables": {
    "code": "iki"
  }
}
```

---

### 3. Query Products (⚠️ CORRECTED - Uses nested `price`)

```graphql
query GetProducts($first: Int, $filters: ProductFilters) {
  products(first: $first, filters: $filters) {
    totalCount
    pageInfo {
      hasNextPage
      endCursor
    }
    edges {
      cursor
      node {
        id
        name
        brand
        category
        subcategory

        # ⚠️ IMPORTANT: price is nested!
        price {
          current
          original
          currency
          discount
          discountPercent
          discountAmount
          isDiscounted
        }

        unitSize
        unitType
        isOnSale
        isCurrentlyOnSale
        isAvailable

        store {
          id
          code
          name
        }

        flyer {
          id
          title
          validFrom
          validTo
        }
      }
    }
  }
}
```

**Variables**:
```json
{
  "first": 10,
  "filters": {
    "isAvailable": true
  }
}
```

**Insomnia JSON**:
```json
{
  "query": "query GetProducts($first: Int, $filters: ProductFilters) { products(first: $first, filters: $filters) { totalCount pageInfo { hasNextPage endCursor } edges { cursor node { id name brand category subcategory price { current original currency discount discountPercent discountAmount isDiscounted } unitSize unitType isOnSale isCurrentlyOnSale isAvailable store { id code name } flyer { id title validFrom validTo } } } } }",
  "variables": {
    "first": 10,
    "filters": {
      "isAvailable": true
    }
  }
}
```

---

### 4. Products On Sale (⚠️ CORRECTED)

```graphql
query GetProductsOnSale($storeIDs: [Int!], $filters: ProductFilters, $first: Int) {
  productsOnSale(storeIDs: $storeIDs, filters: $filters, first: $first) {
    totalCount
    pageInfo {
      hasNextPage
      endCursor
    }
    edges {
      cursor
      node {
        id
        name
        brand
        category

        # ⚠️ Price is nested!
        price {
          current
          original
          currency
          discountAmount
          discountPercent
          isDiscounted
        }

        unitSize
        unitPrice
        isOnSale
        isCurrentlyOnSale
        saleStartDate
        saleEndDate

        store {
          id
          code
          name
          logoURL
        }
      }
    }
  }
}
```

**Variables**:
```json
{
  "first": 20,
  "filters": {
    "minPrice": 0.5,
    "maxPrice": 50.0
  }
}
```

---

### 5. Search Products (⚠️ CORRECTED)

```graphql
query SearchProducts($input: SearchInput!) {
  searchProducts(input: $input) {
    queryString
    totalCount
    hasMore
    products {
      product {
        id
        name
        brand
        category

        # ⚠️ Price is nested!
        price {
          current
          original
          currency
        }

        isOnSale

        store {
          id
          code
          name
        }
      }
      searchScore
      matchType
    }
  }
}
```

**Variables**:
```json
{
  "input": {
    "q": "pienas",
    "first": 10
  }
}
```

**Insomnia JSON**:
```json
{
  "query": "query SearchProducts($input: SearchInput!) { searchProducts(input: $input) { queryString totalCount hasMore products { product { id name brand category price { current original currency } isOnSale store { id code name } } searchScore matchType } } }",
  "variables": {
    "input": {
      "q": "pienas",
      "first": 10
    }
  }
}
```

---

### 6. Get Product by ID (Full Details)

```graphql
query GetProduct($id: Int!) {
  product(id: $id) {
    id
    sku
    slug
    name
    normalizedName
    brand
    category
    subcategory
    description

    # ⚠️ Price nested structure
    price {
      current
      original
      currency
      discount
      discountPercent
      discountAmount
      isDiscounted
    }

    # Physical properties
    unitSize
    unitType
    unitPrice
    packageSize
    weight
    volume

    # Visual
    imageURL

    # Availability
    isOnSale
    isCurrentlyOnSale
    isAvailable
    stockLevel

    # Temporal
    validFrom
    validTo
    saleStartDate
    saleEndDate
    isValid
    isExpired
    validityPeriod

    # Metadata
    extractionConfidence
    extractionMethod
    requiresReview

    # Relations
    store {
      id
      code
      name
      logoURL
    }

    flyer {
      id
      title
      validFrom
      validTo
      isCurrent
    }

    flyerPage {
      id
      pageNumber
      imageURL
    }
  }
}
```

**Variables**:
```json
{
  "id": 1
}
```

---

### 7. Get Current Flyers

```graphql
query GetCurrentFlyers($storeIDs: [Int!], $first: Int) {
  currentFlyers(storeIDs: $storeIDs, first: $first) {
    totalCount
    pageInfo {
      hasNextPage
      endCursor
    }
    edges {
      cursor
      node {
        id
        title
        validFrom
        validTo
        isValid
        isCurrent
        daysRemaining
        validityPeriod
        pageCount
        productsExtracted
        status

        store {
          id
          code
          name
          logoURL
        }

        products(first: 5) {
          totalCount
          edges {
            node {
              id
              name
              price {
                current
              }
            }
          }
        }
      }
    }
  }
}
```

**Variables**:
```json
{
  "first": 20
}
```

---

## 🔑 Key Differences from Old Schema

### ❌ OLD (Broken):
```graphql
{
  products {
    edges {
      node {
        currentPrice      # ❌ Doesn't exist!
        originalPrice     # ❌ Doesn't exist!
        discountPercent   # ❌ Doesn't exist!
      }
    }
  }
}
```

### ✅ NEW (Correct):
```graphql
{
  products {
    edges {
      node {
        price {              # ✅ Nested structure
          current            # ✅ Use these fields
          original
          discountPercent
          discountAmount
          isDiscounted
        }
      }
    }
  }
}
```

---

## 📊 Available Test Data

### Stores (7 total)
- IKI (`code: "iki"`, id: 1)
- Maxima (`code: "maxima"`, id: 2)
- Rimi (`code: "rimi"`, id: 3)
- Lidl, Norfa, Elta, Barbora

### Products (10 total)
- **IKI**: 5 products (Pienas, Jogurtas, Duona, Pomidorai, Vištiena)
- **Maxima**: 2 products (Sviestas, Obuoliai)
- **Rimi**: 3 products (Kava, Alus, Šokoladas)

### Price Range
- Min: €0.79 (Alus)
- Max: €5.99 (Vištiena)
- **7 products on sale** with discounts

### Categories
- Pieno produktai (Dairy)
- Kepiniai (Bakery)
- Daržovės (Vegetables)
- Mėsa (Meat)
- Vaisiai (Fruits)
- Gėrimai (Beverages)
- Saldumynai (Sweets)

---

## 🧪 Quick Test Commands

### Using curl:

```bash
# Test 1: Get all products
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ products(first: 5) { edges { node { id name price { current } } } } }"}'

# Test 2: Search for milk (pienas)
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ searchProducts(input: {q: \"pienas\", first: 3}) { totalCount products { product { id name } } } }"}'

# Test 3: Products on sale
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ productsOnSale(first: 10) { totalCount edges { node { id name price { current original } } } } }"}'
```

---

## ✨ Expected Results

All queries should return:
- ✅ HTTP 200 OK
- ✅ Valid JSON response
- ✅ No GraphQL errors
- ✅ Data matching the schema

---

## 🐛 Troubleshooting

### Error: "Cannot query field currentPrice"
**Solution**: Update query to use `price { current }` instead

### Error: "Cannot query field originalPrice"
**Solution**: Update query to use `price { original }` instead

### Empty results
**Check**: Database has test data (10 products should exist)
**Verify**:
```sql
SELECT COUNT(*) FROM products;
-- Should return 10
```

### Search returns no results
**Issue**: Search query might be too specific
**Solution**: Try simpler queries like "pienas", "alus", "kava"

---

## 📝 Notes

1. **Price is always nested** under `price { ... }`
2. **All prices are in EUR** (€)
3. **Product names are in Lithuanian**
4. **Search supports fuzzy matching** for Lithuanian characters (ąčęėįšųūž)
5. **All timestamps are ISO 8601** format

---

## 🔗 Related Files

- Full Analysis: `/PHASE_1.2_ANALYSIS.md`
- Test Script: `/tmp/test_phase1_comprehensive.sh`
- GraphQL Schema: `/schema.graphql`
