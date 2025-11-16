#!/bin/bash

# Wizard Integration Test Script
# Tests the complete wizard flow: start → decide → confirm/cancel

set -e

if [ -z "$API_TOKEN" ]; then
  echo "❌ Set API_TOKEN environment variable to a valid JWT before running this script."
  exit 1
fi

AUTH_HEADER="Authorization: Bearer $API_TOKEN"
API_URL="${API_URL:-http://localhost:8080/graphql}"
TIMESTAMP=$(date +%s)

echo "=============================================="
echo "🧙 Wizard Integration Test Suite"
echo "=============================================="
echo "API URL: $API_URL"
echo "Timestamp: $TIMESTAMP"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to make GraphQL requests
graphql_query() {
  local query="$1"
  local description="$2"
  
  echo -e "${YELLOW}Testing: $description${NC}"
  
  RESPONSE=$(curl -s -X POST "$API_URL" \
    -H 'Content-Type: application/json' \
    -H "$AUTH_HEADER" \
    -d "{\"query\":$(echo "$query" | jq -Rs .)}")
  
  echo "$RESPONSE" | jq '.'
  
  # Check for errors
  if echo "$RESPONSE" | jq -e '.errors' > /dev/null 2>&1; then
    echo -e "${RED}❌ FAILED: $description${NC}"
    echo "$RESPONSE" | jq '.errors'
    TESTS_FAILED=$((TESTS_FAILED + 1))
    return 1
  else
    echo -e "${GREEN}✅ PASSED: $description${NC}"
    TESTS_PASSED=$((TESTS_PASSED + 1))
    echo "$RESPONSE"
    return 0
  fi
}

# Test 1: Check for expired items
echo ""
echo "=== Test 1: Check Shopping Lists for Expired Items ==="
QUERY_EXPIRED='
query {
  myShoppingLists {
    id
    name
    expiredItemCount
    hasActiveWizardSession
    isLocked
    itemCount
  }
}'

if graphql_query "$QUERY_EXPIRED" "Query shopping lists with expired item counts"; then
  SHOPPING_LISTS=$(echo "$RESPONSE" | jq -r '.data.myShoppingLists')
  
  # Find first list with expired items
  LIST_WITH_EXPIRED=$(echo "$SHOPPING_LISTS" | jq -r '.[] | select(.expiredItemCount > 0) | .id' | head -1)
  
  if [ -z "$LIST_WITH_EXPIRED" ] || [ "$LIST_WITH_EXPIRED" = "null" ]; then
    echo ""
    echo -e "${YELLOW}⚠️  No lists with expired items found. Creating test data...${NC}"
    
    # Create a test shopping list with expired items
    CREATE_LIST_QUERY='
    mutation {
      createShoppingList(input: {
        name: "Test Wizard List '$TIMESTAMP'"
        description: "List for wizard testing"
      }) {
        id
        name
      }
    }'
    
    if graphql_query "$CREATE_LIST_QUERY" "Create test shopping list"; then
      TEST_LIST_ID=$(echo "$RESPONSE" | jq -r '.data.createShoppingList.id')
      echo "Created test list ID: $TEST_LIST_ID"
      
      # Note: In real scenario, you would add items linked to expired flyer products
      echo -e "${YELLOW}⚠️  Please add items with expired flyer products to list $TEST_LIST_ID${NC}"
      echo "Test suite requires shopping list with expired items to continue."
      exit 0
    fi
  else
    TEST_LIST_ID="$LIST_WITH_EXPIRED"
    echo "Found list with expired items: $TEST_LIST_ID"
  fi
fi

# Test 2: Start Wizard Session
echo ""
echo "=== Test 2: Start Wizard Session ==="
START_WIZARD_QUERY='
mutation {
  startWizard(input: {
    shoppingListID: "'$TEST_LIST_ID'"
  }) {
    id
    status
    expiresAt
    selectedStores {
      store {
        id
        name
      }
      itemCount
    }
    expiredItems {
      id
      itemID
      productName
      brand
      originalPrice
      suggestions {
        id
        confidence
        explanation
        score
      }
    }
    progress {
      totalItems
      currentItem
      percentComplete
    }
  }
}'

if graphql_query "$START_WIZARD_QUERY" "Start wizard session"; then
  WIZARD_SESSION_ID=$(echo "$RESPONSE" | jq -r '.data.startWizard.id')
  WIZARD_STATUS=$(echo "$RESPONSE" | jq -r '.data.startWizard.status')
  EXPIRED_ITEMS=$(echo "$RESPONSE" | jq -r '.data.startWizard.expiredItems')
  
  echo ""
  echo "Wizard Session ID: $WIZARD_SESSION_ID"
  echo "Status: $WIZARD_STATUS"
  echo "Expired Items Count: $(echo "$EXPIRED_ITEMS" | jq 'length')"
  
  # Save first expired item for decision testing
  FIRST_ITEM_ID=$(echo "$EXPIRED_ITEMS" | jq -r '.[0].itemID')
  FIRST_SUGGESTION_ID=$(echo "$EXPIRED_ITEMS" | jq -r '.[0].suggestions[0].id')
  
  echo "First Item ID: $FIRST_ITEM_ID"
  echo "First Suggestion ID: $FIRST_SUGGESTION_ID"
else
  echo -e "${RED}❌ Failed to start wizard. Exiting.${NC}"
  exit 1
fi

# Test 3: Query Wizard Session
echo ""
echo "=== Test 3: Query Active Wizard Session ==="
QUERY_SESSION='
query {
  wizardSession(id: "'$WIZARD_SESSION_ID'") {
    id
    status
    expiresAt
    expiredItems {
      id
      productName
      suggestions {
        id
        confidence
        explanation
      }
    }
    progress {
      totalItems
      itemsMigrated
      itemsSkipped
      itemsRemoved
      percentComplete
    }
  }
}'

graphql_query "$QUERY_SESSION" "Query wizard session by ID"

# Test 4: Record Decision - REPLACE
echo ""
echo "=== Test 4: Record Decision (REPLACE with suggestion) ==="
if [ -n "$FIRST_ITEM_ID" ] && [ "$FIRST_ITEM_ID" != "null" ] && [ -n "$FIRST_SUGGESTION_ID" ] && [ "$FIRST_SUGGESTION_ID" != "null" ]; then
  RECORD_REPLACE_QUERY='
  mutation {
    recordDecision(input: {
      sessionID: "'$WIZARD_SESSION_ID'"
      itemID: "'$FIRST_ITEM_ID'"
      decision: REPLACE
      suggestionID: "'$FIRST_SUGGESTION_ID'"
    }) {
      id
      status
      progress {
        itemsMigrated
        percentComplete
      }
    }
  }'
  
  graphql_query "$RECORD_REPLACE_QUERY" "Record REPLACE decision"
else
  echo -e "${YELLOW}⚠️  Skipping REPLACE test - no items/suggestions available${NC}"
fi

# Test 5: Record Decision - SKIP
echo ""
echo "=== Test 5: Record Decision (SKIP item) ==="
SECOND_ITEM_ID=$(echo "$EXPIRED_ITEMS" | jq -r '.[1].itemID')

if [ -n "$SECOND_ITEM_ID" ] && [ "$SECOND_ITEM_ID" != "null" ]; then
  RECORD_SKIP_QUERY='
  mutation {
    recordDecision(input: {
      sessionID: "'$WIZARD_SESSION_ID'"
      itemID: "'$SECOND_ITEM_ID'"
      decision: SKIP
    }) {
      id
      status
      progress {
        itemsSkipped
        percentComplete
      }
    }
  }'
  
  graphql_query "$RECORD_SKIP_QUERY" "Record SKIP decision"
else
  echo -e "${YELLOW}⚠️  Skipping SKIP test - no second item available${NC}"
fi

# Test 6: Record Decision - REMOVE
echo ""
echo "=== Test 6: Record Decision (REMOVE item) ==="
THIRD_ITEM_ID=$(echo "$EXPIRED_ITEMS" | jq -r '.[2].itemID')

if [ -n "$THIRD_ITEM_ID" ] && [ "$THIRD_ITEM_ID" != "null" ]; then
  RECORD_REMOVE_QUERY='
  mutation {
    recordDecision(input: {
      sessionID: "'$WIZARD_SESSION_ID'"
      itemID: "'$THIRD_ITEM_ID'"
      decision: REMOVE
    }) {
      id
      status
      progress {
        itemsRemoved
        percentComplete
      }
    }
  }'
  
  graphql_query "$RECORD_REMOVE_QUERY" "Record REMOVE decision"
else
  echo -e "${YELLOW}⚠️  Skipping REMOVE test - no third item available${NC}"
fi

# Test 7: Idempotency Check
echo ""
echo "=== Test 7: Idempotency Test (Record same decision twice) ==="
if [ -n "$FIRST_ITEM_ID" ] && [ "$FIRST_ITEM_ID" != "null" ]; then
  # Call same mutation again
  graphql_query "$RECORD_REPLACE_QUERY" "Idempotency check - duplicate decision"
fi

# Test 8: Query Session Progress
echo ""
echo "=== Test 8: Check Session Progress ==="
graphql_query "$QUERY_SESSION" "Query session progress after decisions"

# Ask user what to do next
echo ""
echo "=== Decision Point ==="
echo "Wizard session created with ID: $WIZARD_SESSION_ID"
echo ""
read -p "Do you want to (C)onfirm wizard and apply changes, or (X)cancel wizard? [C/x]: " -n 1 -r
echo ""

if [[ $REPLY =~ ^[Xx]$ ]]; then
  # Test 9a: Cancel Wizard
  echo ""
  echo "=== Test 9a: Cancel Wizard Session ==="
  CANCEL_WIZARD_QUERY='
  mutation {
    cancelWizard(sessionID: "'$WIZARD_SESSION_ID'") {
      success
    }
  }'
  
  if graphql_query "$CANCEL_WIZARD_QUERY" "Cancel wizard session"; then
    echo -e "${GREEN}✅ Wizard cancelled successfully${NC}"
  fi
  
  # Verify list is unlocked
  echo ""
  echo "=== Verify List Unlocked After Cancel ==="
  QUERY_LIST_LOCK='
  query {
    shoppingList(id: '$TEST_LIST_ID') {
      id
      name
      isLocked
      hasActiveWizardSession
    }
  }'
  
  graphql_query "$QUERY_LIST_LOCK" "Check list lock status after cancel"
  
else
  # Test 9b: Confirm Wizard
  echo ""
  echo "=== Test 9b: Confirm Wizard and Apply Changes ==="
  CONFIRM_WIZARD_QUERY='
  mutation {
    completeWizard(input: {
      sessionID: "'$WIZARD_SESSION_ID'"
    }) {
      success
      result {
        itemsUpdated
        itemsDeleted
        storeCount
        totalEstimatedPrice
      }
      session {
        id
        status
      }
    }
  }'
  
  if graphql_query "$CONFIRM_WIZARD_QUERY" "Confirm wizard and apply changes"; then
    echo -e "${GREEN}✅ Wizard confirmed successfully${NC}"
    
    ITEMS_UPDATED=$(echo "$RESPONSE" | jq -r '.data.completeWizard.result.itemsUpdated')
    ITEMS_DELETED=$(echo "$RESPONSE" | jq -r '.data.completeWizard.result.itemsDeleted')
    TOTAL_PRICE=$(echo "$RESPONSE" | jq -r '.data.completeWizard.result.totalEstimatedPrice')
    
    echo ""
    echo "Migration Results:"
    echo "  - Items Updated: $ITEMS_UPDATED"
    echo "  - Items Deleted: $ITEMS_DELETED"
    echo "  - Total Estimated Price: €$TOTAL_PRICE"
  fi
  
  # Verify list is unlocked
  echo ""
  echo "=== Verify List Unlocked After Confirm ==="
  QUERY_LIST_LOCK='
  query {
    shoppingList(id: '$TEST_LIST_ID') {
      id
      name
      isLocked
      hasActiveWizardSession
      expiredItemCount
    }
  }'
  
  graphql_query "$QUERY_LIST_LOCK" "Check list lock status after confirm"
fi

# Test 10: Rate Limiting Test
echo ""
echo "=== Test 10: Rate Limiting Test ==="
echo "Attempting to start 6 wizard sessions rapidly (limit is 5/hour)..."

for i in {1..6}; do
  echo ""
  echo "Attempt $i/6..."
  
  RATE_LIMIT_QUERY='
  mutation {
    startWizard(input: {
      shoppingListID: "'$TEST_LIST_ID'"
    }) {
      id
      status
    }
  }'
  
  RESPONSE=$(curl -s -X POST "$API_URL" \
    -H 'Content-Type: application/json' \
    -H "$AUTH_HEADER" \
    -d "{\"query\":$(echo "$RATE_LIMIT_QUERY" | jq -Rs .)}")
  
  if echo "$RESPONSE" | jq -e '.errors[] | select(.message | contains("rate limit"))' > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Rate limit enforced on attempt $i${NC}"
    echo "$RESPONSE" | jq '.errors[0].message'
    TESTS_PASSED=$((TESTS_PASSED + 1))
    break
  elif [ $i -eq 6 ]; then
    echo -e "${YELLOW}⚠️  Rate limit not triggered after 6 attempts${NC}"
  else
    # Cancel the wizard immediately
    NEW_SESSION_ID=$(echo "$RESPONSE" | jq -r '.data.startWizard.id')
    if [ -n "$NEW_SESSION_ID" ] && [ "$NEW_SESSION_ID" != "null" ]; then
      curl -s -X POST "$API_URL" \
        -H 'Content-Type: application/json' \
        -H "$AUTH_HEADER" \
        -d '{"query":"mutation { cancelWizard(sessionID: \"'$NEW_SESSION_ID'\") { success } }"}' > /dev/null
    fi
  fi
done

# Test 11: Expired Session Test
echo ""
echo "=== Test 11: Expired Session Handling ==="
echo "Note: Sessions expire after 30 minutes. This test requires an expired session."
echo "Skipping automated test. Manually test by querying a session after 30+ minutes."

# Summary
echo ""
echo "=============================================="
echo "📊 Test Summary"
echo "=============================================="
echo -e "${GREEN}Tests Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Tests Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
  echo -e "${GREEN}✅ All tests passed!${NC}"
  exit 0
else
  echo -e "${RED}❌ Some tests failed. Review output above.${NC}"
  exit 1
fi
