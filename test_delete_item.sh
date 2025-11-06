#!/bin/bash

TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJrYWludWd1cnUtYXBpIiwiZW1haWwiOiIiLCJleHAiOjE3NjI1MTAxNjEsImlhdCI6MTc2MjQyMzc2MSwiaXNzIjoia2FpbnVndXJ1LWF1dGgiLCJqdGkiOiJiMDI5M2UzYi04ZjNjLTQ1MDAtODM1Yi0yZTI1YmZjODMyNjIiLCJzaWQiOiJmOTk3NzJlNy1iODY2LTQ1M2MtYTI4MS02ZjhkMzhlMWY3ZjQiLCJzdWIiOiJiY2RhY2FiNS1iMTk4LTQ4ODktODhkZC1kNDcxODRiY2JhZjkiLCJ0eXBlIjoiYWNjZXNzIn0.eC3YeNs-fX27zTS9PfVWqS6Qmv0jbk90X2cPkTx_8Sk"

echo "=== Test Delete Item 4 (Milk) ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"mutation { deleteShoppingListItem(id: 4) }"}' | jq

echo ""
echo "=== Verify Item 4 Deleted - Query List ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"query { shoppingList(id: 1) { id itemCount items { edges { node { id description } } } } }"}' | jq
