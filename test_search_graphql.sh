#!/bin/bash

# Test all search endpoints

echo "Testing GraphQL Search Endpoints"
echo "================================="
echo ""

# Test 1: Basic hybrid search
echo "1. Testing basic hybrid search for 'pienas'"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { searchProducts(input: { q: \"pienas\", first: 5 }) { products { product { id name price { current } } searchScore matchType } totalCount } }"
  }' | jq '.data.searchProducts' || echo "FAILED"
echo ""

# Test 2: Fuzzy search
echo "2. Testing fuzzy search for 'pienas'"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { searchProducts(input: { q: \"pienas\", preferFuzzy: true, first: 5 }) { products { product { id name } searchScore matchType } totalCount } }"
  }' | jq '.data.searchProducts' || echo "FAILED"
echo ""

# Test 3: On-sale only filter
echo "3. Testing on-sale only filter"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { searchProducts(input: { q: \"pienas\", onSaleOnly: true, first: 10 }) { products { product { id name isOnSale } searchScore } totalCount } }"
  }' | jq '.data.searchProducts' || echo "FAILED"
echo ""

# Test 4: Price range filter
echo "4. Testing price range filter (0.50 - 1.00)"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { searchProducts(input: { q: \"pienas\", minPrice: 0.50, maxPrice: 1.00, first: 5 }) { products { product { id name price { current } } } totalCount } }"
  }' | jq '.data.searchProducts' || echo "FAILED"
echo ""

# Test 5: Category filter
echo "5. Testing Lithuanian text search 'duona'"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { searchProducts(input: { q: \"duona\", first: 5 }) { products { product { id name } searchScore matchType } totalCount } }"
  }' | jq '.data.searchProducts' || echo "FAILED"
echo ""

# Test 6: Error handling - empty query
echo "6. Testing error handling with empty query"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { searchProducts(input: { q: \"\" }) { products { product { id } } totalCount } }"
  }' | jq '.errors // .data' || echo "FAILED"
echo ""

echo "================================="
echo "Search endpoint tests completed"
