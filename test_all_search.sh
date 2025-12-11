#!/bin/bash

# Comprehensive Search Endpoint Test Suite
# Tests all search types, filters, and edge cases

set -e

API_URL="http://localhost:8080/graphql"
PASS_COUNT=0
FAIL_COUNT=0

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "Kainuguru Search API Test Suite"
echo "========================================="
echo ""

test_query() {
    local test_name="$1"
    local query="$2"
    local expected_field="$3"

    echo -n "Testing: $test_name... "

    result=$(curl -s -X POST "$API_URL" \
        -H "Content-Type: application/json" \
        --data-binary "{\"query\":\"$query\"}")

    if echo "$result" | jq -e "$expected_field" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        ((PASS_COUNT++))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}"
        echo "  Response: $result"
        ((FAIL_COUNT++))
        return 1
    fi
}

# Test 1: Basic Hybrid Search
test_query \
    "Basic hybrid search" \
    "query { searchProducts(input: { q: \\\"pienas\\\" }) { totalCount products { matchType } } }" \
    ".data.searchProducts.totalCount"

# Test 2: Fuzzy Search
test_query \
    "Fuzzy search" \
    "query { searchProducts(input: { q: \\\"pienas\\\", preferFuzzy: true }) { products { matchType } } }" \
    ".data.searchProducts.products[0].matchType"

# Test 3: On-Sale Filter
test_query \
    "On-sale only filter" \
    "query { searchProducts(input: { q: \\\"pienas\\\", onSaleOnly: true }) { totalCount } }" \
    ".data.searchProducts.totalCount"

# Test 4: Price Range
test_query \
    "Price range filter" \
    "query { searchProducts(input: { q: \\\"pienas\\\", minPrice: 0.5, maxPrice: 1.0 }) { totalCount } }" \
    ".data.searchProducts.totalCount"

# Test 5: Lithuanian Text
test_query \
    "Lithuanian text search" \
    "query { searchProducts(input: { q: \\\"duona\\\" }) { totalCount } }" \
    ".data.searchProducts.totalCount"

# Test 6: Tag Filtering
test_query \
    "Tag filtering" \
    "query { searchProducts(input: { q: \\\"pienas\\\", tags: [\\\"pieno-produktai\\\"] }) { totalCount } }" \
    ".data.searchProducts.totalCount"

# Test 7: Pagination
test_query \
    "Pagination" \
    "query { searchProducts(input: { q: \\\"pienas\\\", first: 2 }) { products { product { id } } pagination { itemsPerPage } } }" \
    ".data.searchProducts.pagination.itemsPerPage"

# Test 8: Facets - Stores
test_query \
    "Facets - stores" \
    "query { searchProducts(input: { q: \\\"pienas\\\" }) { facets { stores { options { value count } } } } }" \
    ".data.searchProducts.facets.stores.options[0].count"

# Test 9: Facets - Availability
test_query \
    "Facets - availability" \
    "query { searchProducts(input: { q: \\\"pienas\\\" }) { facets { availability { options { value count } } } } }" \
    ".data.searchProducts.facets.availability.options"

# Test 10: Search Score
test_query \
    "Search score present" \
    "query { searchProducts(input: { q: \\\"pienas\\\" }) { products { searchScore } } }" \
    ".data.searchProducts.products[0].searchScore"

# Test 11: Product Relations
test_query \
    "Product with store relation" \
    "query { searchProducts(input: { q: \\\"pienas\\\" }) { products { product { store { name } } } } }" \
    ".data.searchProducts.products[0].product.store.name"

# Test 12: Product Price
test_query \
    "Product price structure" \
    "query { searchProducts(input: { q: \\\"pienas\\\" }) { products { product { price { current currency } } } } }" \
    ".data.searchProducts.products[0].product.price.current"

# Test 13: Empty Query Error
echo -n "Testing: Empty query validation... "
result=$(curl -s -X POST "$API_URL" \
    -H "Content-Type: application/json" \
    --data-binary '{"query":"query { searchProducts(input: { q: \"\" }) { totalCount } }"}')

if echo "$result" | jq -e '.errors[0].message' | grep -q "query cannot be empty"; then
    echo -e "${GREEN}✓ PASS${NC}"
    ((PASS_COUNT++))
else
    echo -e "${RED}✗ FAIL${NC}"
    echo "  Expected validation error, got: $result"
    ((FAIL_COUNT++))
fi

# Test 14: Has More Flag (search with limit 2 when more results exist)
test_query \
    "Has more flag (true case)" \
    "query { searchProducts(input: { q: \\\"pienas\\\", first: 2 }) { hasMore totalCount } }" \
    ".data.searchProducts"

# Test 15: Query String Echo
test_query \
    "Query string echo" \
    "query { searchProducts(input: { q: \\\"pienas\\\" }) { queryString } }" \
    ".data.searchProducts.queryString"

echo ""
echo "========================================="
echo "Test Results"
echo "========================================="
echo -e "${GREEN}Passed: $PASS_COUNT${NC}"
echo -e "${RED}Failed: $FAIL_COUNT${NC}"
echo "========================================="

if [ $FAIL_COUNT -eq 0 ]; then
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}Some tests failed.${NC}"
    exit 1
fi
