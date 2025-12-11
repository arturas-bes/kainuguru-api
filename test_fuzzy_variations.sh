#!/bin/bash

echo "Testing Fuzzy Search with Typos and Variations"
echo "==============================================="
echo ""

# Test exact match
echo "1. Exact match: 'pienas'"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"pienas\", preferFuzzy: true }) { totalCount products { product { name } searchScore } } }"}' | jq '.data.searchProducts'
echo ""

# Test with typo
echo "2. Typo: 'pienas' -> 'penas' (missing 'i')"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"penas\", preferFuzzy: true }) { totalCount products { product { name } searchScore } } }"}' | jq '.data.searchProducts'
echo ""

# Test with typo
echo "3. Typo: 'pienas' -> 'pienas' -> 'pianes' (transposition)"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"pianes\", preferFuzzy: true }) { totalCount products { product { name } searchScore } } }"}' | jq '.data.searchProducts'
echo ""

# Test partial match
echo "4. Partial: 'pien'"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"pien\", preferFuzzy: true }) { totalCount products { product { name } searchScore } } }"}' | jq '.data.searchProducts'
echo ""

# Test Lithuanian char variations
echo "5. Lithuanian: 'duona' -> 'duona' (exact)"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"duona\", preferFuzzy: true }) { totalCount products { product { name } searchScore } } }"}' | jq '.data.searchProducts'
echo ""

# Test Lithuanian char variations
echo "6. Lithuanian typo: 'duona' -> 'dona' (missing 'u')"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"dona\", preferFuzzy: true }) { totalCount products { product { name } searchScore } } }"}' | jq '.data.searchProducts'
echo ""

# Test different case
echo "7. Case: 'PIENAS' (uppercase)"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"PIENAS\", preferFuzzy: true }) { totalCount products { product { name } searchScore } } }"}' | jq '.data.searchProducts'
echo ""

# Direct database query to check similarity
echo "8. Direct DB: Testing similarity values for 'penas'"
docker-compose exec -T db psql -U kainuguru -d kainuguru_db -c "
SELECT
    name,
    similarity('penas', name) as name_sim,
    similarity('penas', normalized_name) as norm_sim
FROM products
WHERE name ILIKE '%pienas%'
LIMIT 3;
" 2>&1
echo ""

# Test with very similar word
echo "9. Similar word: 'pieno' (genitive case)"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"pieno\", preferFuzzy: true }) { totalCount products { product { name } searchScore } } }"}' | jq '.data.searchProducts'
echo ""
