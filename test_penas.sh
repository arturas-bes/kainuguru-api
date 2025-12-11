#!/bin/bash
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  --data-binary '{"query":"query { searchProducts(input: { q: \"penas\", preferFuzzy: true }) { totalCount products { product { name } searchScore } } }"}' | jq '.'
