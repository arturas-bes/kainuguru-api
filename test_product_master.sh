#!/bin/bash

echo "Testing ProductMaster endpoint..."
echo ""

# Test 1: Get single product master
echo "Test 1: Get product master by ID"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "query { productMaster(id: 1) { id canonicalName } }"}' | jq .

echo ""
echo "---"
echo ""

# Test 2: Get all product masters
echo "Test 2: Get all product masters"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "query { productMasters(first: 5) { totalCount edges { node { id canonicalName } } } }"}' | jq .

echo ""
echo "---"
echo ""

# Test 3: With filters
echo "Test 3: Get product masters with filters"
curl -s -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{"query": "query { productMasters(first: 5, filters: { status: [ACTIVE], minConfidence: 0.5 }) { totalCount edges { node { id canonicalName status confidenceScore } } } }"}' | jq .
