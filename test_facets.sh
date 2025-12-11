#!/bin/bash

echo "Testing search facets"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"query { searchProducts(input: { q: \"pienas\" }) { totalCount facets { stores { name options { value name count } } brands { name options { value count } } availability { name options { value count } } } } }"}' | jq '.'
