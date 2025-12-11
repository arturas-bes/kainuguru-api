#!/bin/bash

echo "Testing totalCount fix"
echo "====================="
echo ""

echo "1. Test 'penas' (typo)"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"penas\", preferFuzzy: true }) { totalCount products { product { name } } } }"}' | jq '.data.searchProducts | {totalCount, productCount: (.products | length)}'
echo ""

echo "2. Test 'pien' (partial)"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"pien\", preferFuzzy: true }) { totalCount products { product { name } } } }"}' | jq '.data.searchProducts | {totalCount, productCount: (.products | length)}'
echo ""

echo "3. Test 'pieno' (variation)"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"pieno\", preferFuzzy: true }) { totalCount products { product { name } } } }"}' | jq '.data.searchProducts | {totalCount, productCount: (.products | length)}'
echo ""

echo "4. Test 'dona' (Lithuanian typo)"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"dona\", preferFuzzy: true }) { totalCount products { product { name } } } }"}' | jq '.data.searchProducts | {totalCount, productCount: (.products | length)}'
