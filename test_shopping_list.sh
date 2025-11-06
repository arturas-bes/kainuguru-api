#!/bin/bash

TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJrYWludWd1cnUtYXBpIiwiZW1haWwiOiIiLCJleHAiOjE3NjI1MTAxNjEsImlhdCI6MTc2MjQyMzc2MSwiaXNzIjoia2FpbnVndXJ1LWF1dGgiLCJqdGkiOiJiMDI5M2UzYi04ZjNjLTQ1MDAtODM1Yi0yZTI1YmZjODMyNjIiLCJzaWQiOiJmOTk3NzJlNy1iODY2LTQ1M2MtYTI4MS02ZjhkMzhlMWY3ZjQiLCJzdWIiOiJiY2RhY2FiNS1iMTk4LTQ4ODktODhkZC1kNDcxODRiY2JhZjkiLCJ0eXBlIjoiYWNjZXNzIn0.eC3YeNs-fX27zTS9PfVWqS6Qmv0jbk90X2cPkTx_8Sk"

echo "=== Test 1: Create Shopping List ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"mutation { createShoppingList(input: { name: \"Weekly Groceries\", description: \"My weekly shopping list\" }) { id name description isDefault createdAt user { email } } }"}' | jq

echo ""
echo "=== Test 2: Query My Shopping Lists ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"query { shoppingLists { edges { node { id name description isDefault itemCount } } } }"}' | jq

echo ""
echo "=== Test 3: Query My Default Shopping List ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"query { myDefaultShoppingList { id name description isDefault itemCount completionPercentage } }"}' | jq

echo ""
echo "=== Test 4: Update Shopping List ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"mutation { updateShoppingList(id: 1, input: { name: \"Updated Weekly Groceries\", description: \"Updated description\" }) { id name description } }"}' | jq

echo ""
echo "=== Test 5: Create Second Shopping List ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"mutation { createShoppingList(input: { name: \"Monthly Bulk Buy\", description: \"Monthly shopping\" }) { id name isDefault } }"}' | jq

echo ""
echo "=== Test 6: Set Second List as Default ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"mutation { setDefaultShoppingList(id: 2) { id name isDefault } }"}' | jq
