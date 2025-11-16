#!/bin/bash

if [ -z "$API_TOKEN" ]; then
  echo "Set API_TOKEN environment variable to a valid JWT before running this script."
  exit 1
fi

AUTH_HEADER="Authorization: Bearer $API_TOKEN"

echo "=== Test Delete Item 4 (Milk) ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "$AUTH_HEADER" \
  -d '{"query":"mutation { deleteShoppingListItem(id: 4) }"}' | jq

echo ""
echo "=== Verify Item 4 Deleted - Query List ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "$AUTH_HEADER" \
  -d '{"query":"query { shoppingList(id: 1) { id itemCount items { edges { node { id description } } } } }"}' | jq
