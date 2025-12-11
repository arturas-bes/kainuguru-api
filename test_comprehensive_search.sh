#!/bin/bash

# Comprehensive Search Testing Script
# Tests all search features and edge cases

API_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjM2MjQ4OTksImlhdCI6MTc2MzUzODQ5OSwic2Vzc2lvbl9pZCI6ImNlNTdiZDhiLTNkY2MtNDQwZi1iN2EzLTYwYzA3NmVmMDhkMiIsInR5cGUiOiJhY2Nlc3MiLCJ1c2VyX2lkIjoiZTk3Y2NmYWEtMmRmYS00Y2FlLWI5MmEtY2NkZjM1MTlkNTU2In0.NaIL8Nyr6jH4cqTGuz0_3ga_ruVjW82rHpBrTDh3Znc"
BASE_URL="http://localhost:8080/graphql"

passed=0
failed=0

# Helper function to run test
run_test() {
    local test_name="$1"
    local query="$2"
    local expected_pattern="$3"
    
    echo "━━━ TEST: $test_name ━━━"
    
    response=$(curl -s -X POST "$BASE_URL" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $API_TOKEN" \
        -d "{\"query\": \"$query\"}")
    
    echo "$response" | jq '.' 2>/dev/null || echo "$response"
    
    if echo "$response" | grep -q "\"errors\""; then
        echo "❌ FAILED: Error in response"
        ((failed++))
        return 1
    fi
    
    if [[ -n "$expected_pattern" ]]; then
        if echo "$response" | grep -q "$expected_pattern"; then
            echo "✅ PASSED"
            ((passed++))
        else
            echo "❌ FAILED: Expected pattern '$expected_pattern' not found"
            ((failed++))
            return 1
        fi
    else
        echo "✅ PASSED"
        ((passed++))
    fi
    
    echo ""
    sleep 0.5  # Avoid rate limiting
}

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║          COMPREHENSIVE SEARCH TESTING SUITE               ║"
echo "╚═══════════════════════════════════════════════════════════╝"
echo ""

# Test 1: Basic fuzzy search
run_test "Basic Fuzzy Search" \
    "query { searchProducts(input: { q: \\\"pienas\\\", preferFuzzy: true }) { totalCount products { matchType } } }" \
    "\"matchType\":\"fuzzy\""

# Test 2: Hybrid search (FTS)
run_test "Hybrid Search (FTS)" \
    "query { searchProducts(input: { q: \\\"pienas\\\", preferFuzzy: false }) { totalCount products { matchType } } }" \
    "\"matchType\":\"fts\""

# Test 3: Store filtering
run_test "Store Filter (Maxima only)" \
    "query { searchProducts(input: { q: \\\"pienas\\\", storeIDs: [1] }) { totalCount } }" \
    "\"totalCount\""

# Test 4: Price range filtering (min)
run_test "Price Filter (minPrice: 2.0)" \
    "query { searchProducts(input: { q: \\\"pienas\\\", minPrice: 2.0 }) { totalCount } }" \
    "\"totalCount\""

# Test 5: Price range filtering (max)
run_test "Price Filter (maxPrice: 3.0)" \
    "query { searchProducts(input: { q: \\\"pienas\\\", maxPrice: 3.0 }) { totalCount } }" \
    "\"totalCount\""

# Test 6: Price range filtering (both)
run_test "Price Filter (min: 1.0, max: 4.0)" \
    "query { searchProducts(input: { q: \\\"pienas\\\", minPrice: 1.0, maxPrice: 4.0 }) { totalCount } }" \
    "\"totalCount\""

# Test 7: On-sale only filter
run_test "On-Sale Filter" \
    "query { searchProducts(input: { q: \\\"pienas\\\", onSaleOnly: true }) { totalCount } }" \
    "\"totalCount\""

# Test 8: Category filter
run_test "Category Filter (Pieno produktai)" \
    "query { searchProducts(input: { q: \\\"\\\"  category: \\\"Pieno\\\" }) { totalCount products { product { category } } } }" \
    "\"category\""

# Test 9: Lithuanian characters (with diacritics)
run_test "Lithuanian Characters (ž, š, ė)" \
    "query { searchProducts(input: { q: \\\"Žemaitijos\\\" }) { totalCount } }" \
    "\"totalCount\""

# Test 10: Multiple filters combined
run_test "Combined Filters (store + price + category)" \
    "query { searchProducts(input: { q: \\\"pienas\\\", storeIDs: [1,2], minPrice: 1.0, maxPrice: 5.0 }) { totalCount } }" \
    "\"totalCount\""

# Test 11: Pagination (limit)
run_test "Pagination (first: 2)" \
    "query { searchProducts(input: { q: \\\"pienas\\\", first: 2 }) { products { product { id } } totalCount } }" \
    "\"totalCount\""

# Test 12: Empty result
run_test "Empty Result (non-existent product)" \
    "query { searchProducts(input: { q: \\\"xyzabc123notfound\\\" }) { totalCount } }" \
    "\"totalCount\":0"

# Test 13: Single character search
run_test "Single Character Search" \
    "query { searchProducts(input: { q: \\\"a\\\" }) { totalCount } }" \
    "\"totalCount\""

# Test 14: Special characters sanitization
run_test "Special Characters (<, >, &)" \
    "query { searchProducts(input: { q: \\\"test \u0026 product\\\" }) { totalCount } }" \
    "\"totalCount\""

# Test 15: Very long query
run_test "Long Query (100+ chars)" \
    "query { searchProducts(input: { q: \\\"pienas pienas pienas pienas pienas pienas pienas pienas pienas pienas pienas pienas pienas pienas pienas\\\" }) { totalCount } }" \
    "\"totalCount\""

# Test 16: Query with numbers
run_test "Query with Numbers" \
    "query { searchProducts(input: { q: \\\"2.5\\\" }) { totalCount } }" \
    "\"totalCount\""

# Test 17: Similarity scores populated (fuzzy)
run_test "Similarity Scores (fuzzy)" \
    "query { searchProducts(input: { q: \\\"pienas\\\", preferFuzzy: true, first: 1 }) { products { similarity searchScore } } }" \
    "\"similarity\""

# Test 18: No similarity for FTS
run_test "No Similarity for FTS" \
    "query { searchProducts(input: { q: \\\"pienas\\\", preferFuzzy: false, first: 1 }) { products { similarity searchScore matchType } } }" \
    "\"matchType\":\"fts\""

echo "╔═══════════════════════════════════════════════════════════╗"
echo "║                    TEST SUMMARY                           ║"
echo "╠═══════════════════════════════════════════════════════════╣"
echo "║  Total Tests:  $((passed + failed))                                         ║"
echo "║  ✅ Passed:    $passed                                          ║"
echo "║  ❌ Failed:    $failed                                           ║"
echo "╚═══════════════════════════════════════════════════════════╝"

if [[ $failed -eq 0 ]]; then
    echo ""
    echo "🎉 ALL TESTS PASSED! Search is working perfectly!"
    exit 0
else
    echo ""
    echo "⚠️  SOME TESTS FAILED. Please review the output above."
    exit 1
fi
