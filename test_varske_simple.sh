#!/bin/bash
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"varske\", preferFuzzy: true, first: 10 }) { totalCount products { product { name price { current } } searchScore } } }"}' | jq '.'
