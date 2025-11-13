#!/bin/bash

if [ -z "$API_TOKEN" ]; then
  echo "Set API_TOKEN environment variable to a valid JWT before running this script."
  exit 1
fi

AUTH_HEADER="Authorization: Bearer $API_TOKEN"
SUFFIX=$(date +%s)
LIST_NAME_ONE="Weekly Groceries $SUFFIX"
LIST_NAME_TWO="Monthly Bulk Buy $SUFFIX"

echo "=== Test 1: Create Shopping List ==="
RESPONSE=$(curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "$AUTH_HEADER" \
  -d "{\"query\":\"mutation { createShoppingList(input: { name: \\\"$LIST_NAME_ONE\\\", description: \\\"My weekly shopping list\\\" }) { id name description isDefault createdAt user { email } } }\"}")
echo "$RESPONSE" | jq
LIST_ONE_ID=$(echo "$RESPONSE" | jq -r '.data.createShoppingList.id')

echo ""
echo "=== Test 2: Query My Shopping Lists ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "$AUTH_HEADER" \
  -d '{"query":"query { shoppingLists { edges { node { id name description isDefault itemCount } } } }"}' | jq

echo ""
echo "=== Test 3: Query My Default Shopping List ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "$AUTH_HEADER" \
  -d '{"query":"query { myDefaultShoppingList { id name description isDefault itemCount completionPercentage } }"}' | jq

echo ""
echo "=== Test 4: Update Shopping List ==="
if [ -n "$LIST_ONE_ID" ] && [ "$LIST_ONE_ID" != "null" ]; then
  curl -s -X POST http://localhost:8080/graphql \
    -H 'Content-Type: application/json' \
    -H "$AUTH_HEADER" \
    -d "{\"query\":\"mutation { updateShoppingList(id: $LIST_ONE_ID, input: { name: \\\"Updated Weekly Groceries\\\", description: \\\"Updated description\\\" }) { id name description } }\"}" | jq
else
  echo "Skipping update; could not capture first list ID."
fi

echo ""
echo "=== Test 5: Create Second Shopping List ==="
RESPONSE=$(curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "$AUTH_HEADER" \
  -d "{\"query\":\"mutation { createShoppingList(input: { name: \\\"$LIST_NAME_TWO\\\", description: \\\"Monthly shopping\\\" }) { id name isDefault } }\"}")
echo "$RESPONSE" | jq
LIST_TWO_ID=$(echo "$RESPONSE" | jq -r '.data.createShoppingList.id')

echo ""
echo "=== Test 6: Set Second List as Default ==="
if [ -n "$LIST_TWO_ID" ] && [ "$LIST_TWO_ID" != "null" ]; then
  curl -s -X POST http://localhost:8080/graphql \
    -H 'Content-Type: application/json' \
    -H "$AUTH_HEADER" \
    -d "{\"query\":\"mutation { setDefaultShoppingList(id: $LIST_TWO_ID) { id name isDefault } }\"}" | jq
else
  echo "Skipping default assignment; second list not created."
fi
