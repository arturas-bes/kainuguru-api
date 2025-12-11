#!/bin/bash

echo "Testing Fuzzy Search with 'varske'"
echo "==================================="
echo ""

echo "1. Exact match: 'Varškė' (with Lithuanian char)"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"Varškė\", preferFuzzy: true }) { totalCount products { product { name } searchScore } } }"}' | jq '.data.searchProducts'
echo ""

echo "2. Common spelling: 'varske' (without special char)"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"varske\", preferFuzzy: true }) { totalCount products { product { name } searchScore } } }"}' | jq '.data.searchProducts'
echo ""

echo "3. Typo: 'varke' (missing 's')"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"varke\", preferFuzzy: true }) { totalCount products { product { name } searchScore } } }"}' | jq '.data.searchProducts'
echo ""

echo "4. Partial: 'varsk'"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"varsk\", preferFuzzy: true }) { totalCount products { product { name } searchScore } } }"}' | jq '.data.searchProducts'
echo ""

echo "5. Hybrid search (default): 'varske'"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"varske\" }) { totalCount products { product { name } matchType searchScore } } }"}' | jq '.data.searchProducts'
echo ""

echo "6. Direct DB test: similarity scores"
docker-compose exec -T db psql -U kainuguru -d kainuguru_db -c "
SELECT
    name,
    similarity('varske', name) as name_sim,
    similarity('varske', normalized_name) as norm_sim,
    (
        similarity('varske', name) * 0.7 +
        similarity('varske', normalized_name) * 0.2
    ) as combined_sim
FROM products
WHERE name LIKE '%arsk%'
ORDER BY combined_sim DESC;
" 2>&1
