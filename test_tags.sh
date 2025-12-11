#!/bin/bash

echo "Testing tag filtering"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"pienas\", tags: [\"pieno-produktai\"] }) { totalCount products { product { id name tags } } } }"}' | jq '.data.searchProducts'
