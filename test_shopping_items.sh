#!/bin/bash

if [ -z "$API_TOKEN" ]; then
  echo "Set API_TOKEN environment variable to a valid JWT before running this script."
  exit 1
fi

AUTH_HEADER="Authorization: Bearer $API_TOKEN"
LIST_ID=${SHOPPING_LIST_ID:-1}

graphql() {
  curl -s -X POST http://localhost:8080/graphql \
    -H 'Content-Type: application/json' \
    -H "$AUTH_HEADER" \
    -d "$1"
}

echo "=== Test 1: Create Shopping List Items ==="
RESP=$(graphql "{\"query\":\"mutation { createShoppingListItem(input: { shoppingListID: $LIST_ID, description: \\\"Milk\\\", quantity: 2, unit: \\\"L\\\", category: \\\"Dairy\\\" }) { id description quantity unit category } }\"}")
echo "$RESP" | jq
ITEM1_ID=$(echo "$RESP" | jq -r '.data.createShoppingListItem.id')

RESP=$(graphql "{\"query\":\"mutation { createShoppingListItem(input: { shoppingListID: $LIST_ID, description: \\\"Bread\\\", quantity: 1, category: \\\"Bakery\\\", tags: [\\\"whole-wheat\\\"] }) { id description category tags } }\"}")
echo "$RESP" | jq
ITEM2_ID=$(echo "$RESP" | jq -r '.data.createShoppingListItem.id')

RESP=$(graphql "{\"query\":\"mutation { createShoppingListItem(input: { shoppingListID: $LIST_ID, description: \\\"Apples\\\", quantity: 6, unit: \\\"pcs\\\", category: \\\"Produce\\\", estimatedPrice: 2.99 }) { id description quantity estimatedPrice } }\"}")
echo "$RESP" | jq

echo ""
echo "=== Test 2: Query Shopping List with Items ==="
graphql "{\"query\":\"query { shoppingList(id: $LIST_ID) { id name itemCount items { edges { node { id description quantity unit category isChecked } } } } }\"}" | jq

echo ""
echo "=== Test 3: Update Shopping List Item ==="
if [ -n "$ITEM1_ID" ] && [ "$ITEM1_ID" != "null" ]; then
  graphql "{\"query\":\"mutation { updateShoppingListItem(id: $ITEM1_ID, input: { description: \\\"Organic Milk\\\", quantity: 3 }) { id description quantity } }\"}" | jq
else
  echo "Skipping update; first item ID missing."
fi

echo ""
echo "=== Test 4: Check Shopping List Item ==="
if [ -n "$ITEM1_ID" ] && [ "$ITEM1_ID" != "null" ]; then
  graphql "{\"query\":\"mutation { checkShoppingListItem(id: $ITEM1_ID) { id description isChecked checkedAt } }\"}" | jq
else
  echo "Skipping check; first item ID missing."
fi

echo ""
echo "=== Test 5: Uncheck Shopping List Item ==="
if [ -n "$ITEM1_ID" ] && [ "$ITEM1_ID" != "null" ]; then
  graphql "{\"query\":\"mutation { uncheckShoppingListItem(id: $ITEM1_ID) { id description isChecked } }\"}" | jq
else
  echo "Skipping uncheck; first item ID missing."
fi

echo ""
echo "=== Test 6: Query Items by Status (Unchecked Only) ==="
graphql "{\"query\":\"query { shoppingList(id: $LIST_ID) { items(filters: { isChecked: false }) { edges { node { id description isChecked } } } } }\"}" | jq

echo ""
echo "=== Test 7: Delete Shopping List Item ==="
if [ -n "$ITEM2_ID" ] && [ "$ITEM2_ID" != "null" ]; then
  graphql "{\"query\":\"mutation { deleteShoppingListItem(id: $ITEM2_ID) }\"}" | jq
else
  echo "Skipping delete; second item ID missing."
fi

echo ""
echo "=== Test 8: Verify List Statistics Updated ==="
graphql "{\"query\":\"query { shoppingList(id: $LIST_ID) { id itemCount completedItemCount completionPercentage } }\"}" | jq

echo ""
echo "=== All Tests Complete ==="
