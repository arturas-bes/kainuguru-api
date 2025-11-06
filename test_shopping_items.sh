#!/bin/bash

# Get access token first
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJrYWludWd1cnUtYXBpIiwiZW1haWwiOiIiLCJleHAiOjE3NjI1MTAxNjEsImlhdCI6MTc2MjQyMzc2MSwiaXNzIjoia2FpbnVndXJ1LWF1dGgiLCJqdGkiOiJiMDI5M2UzYi04ZjNjLTQ1MDAtODM1Yi0yZTI1YmZjODMyNjIiLCJzaWQiOiJmOTk3NzJlNy1iODY2LTQ1M2MtYTI4MS02ZjhkMzhlMWY3ZjQiLCJzdWIiOiJiY2RhY2FiNS1iMTk4LTQ4ODktODhkZC1kNDcxODRiY2JhZjkiLCJ0eXBlIjoiYWNjZXNzIn0.eC3YeNs-fX27zTS9PfVWqS6Qmv0jbk90X2cPkTx_8Sk"

echo "=== Test 1: Create Shopping List Items ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"mutation { item1: createShoppingListItem(input: { shoppingListID: 1, description: \"Milk\", quantity: 2, unit: \"L\", category: \"Dairy\" }) { id description quantity unit category } }"}' | jq

echo ""
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"mutation { item2: createShoppingListItem(input: { shoppingListID: 1, description: \"Bread\", quantity: 1, category: \"Bakery\", tags: [\"whole-wheat\"] }) { id description category tags } }"}' | jq

echo ""
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"mutation { item3: createShoppingListItem(input: { shoppingListID: 1, description: \"Apples\", quantity: 6, unit: \"pcs\", category: \"Produce\", estimatedPrice: 2.99 }) { id description quantity estimatedPrice } }"}' | jq

echo ""
echo "=== Test 2: Query Shopping List with Items ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"query { shoppingList(id: 1) { id name itemCount items { edges { node { id description quantity unit category isChecked } } } } }"}' | jq

echo ""
echo "=== Test 3: Update Shopping List Item ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"mutation { updateShoppingListItem(id: 1, input: { description: \"Organic Milk\", quantity: 3 }) { id description quantity } }"}' | jq

echo ""
echo "=== Test 4: Check Shopping List Item ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"mutation { checkShoppingListItem(id: 1) { id description isChecked checkedAt } }"}' | jq

echo ""
echo "=== Test 5: Uncheck Shopping List Item ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"mutation { uncheckShoppingListItem(id: 1) { id description isChecked } }"}' | jq

echo ""
echo "=== Test 6: Query Items by Status (Unchecked Only) ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"query { shoppingList(id: 1) { items(filters: { isChecked: false }) { edges { node { id description isChecked } } } } }"}' | jq

echo ""
echo "=== Test 7: Delete Shopping List Item ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"mutation { deleteShoppingListItem(id: 2) }"}' | jq

echo ""
echo "=== Test 8: Verify List Statistics Updated ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query":"query { shoppingList(id: 1) { id itemCount completedItemCount completionPercentage } }"}' | jq

echo ""
echo "=== All Tests Complete ==="
