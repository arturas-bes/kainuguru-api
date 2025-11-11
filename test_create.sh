#!/bin/bash

if [ -z "$API_TOKEN" ]; then
  echo "Set API_TOKEN environment variable to a valid JWT before running this script."
  exit 1
fi

AUTH_HEADER="Authorization: Bearer $API_TOKEN"

echo "=== Test: Create Shopping List with User Relation ==="
curl -s -X POST http://localhost:8080/graphql \
  -H 'Content-Type: application/json' \
  -H "$AUTH_HEADER" \
  -d '{"query":"mutation { createShoppingList(input: { name: \"Test List\", description: \"Testing user relation\" }) { id name description user { email fullName } } }"}' | jq
