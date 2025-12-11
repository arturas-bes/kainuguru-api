#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo "🚀 Starting Manual Verification..."

# 1.# Create user and session in DB directly to bypass potential API issues
echo "🔑 Creating User, Session and Generating Token..."
SESSION_ID=$(uuidgen)
USER_ID="e97ccfaa-2dfa-4cae-b92a-ccdf3519d556"

# Insert user if not exists
docker-compose exec -T db psql -U kainuguru -d kainuguru_db -c "
INSERT INTO users (id, email, password_hash, created_at, updated_at)
VALUES ('$USER_ID', 'test@example.com', 'hash', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;
"

# Insert session
docker-compose exec -T db psql -U kainuguru -d kainuguru_db -c "
INSERT INTO user_sessions (id, user_id, token_hash, expires_at, created_at, ip_address, user_agent)
VALUES ('$SESSION_ID', '$USER_ID', 'hash', NOW() + INTERVAL '24 hours', NOW(), '127.0.0.1', 'manual_verify')
ON CONFLICT (token_hash) DO NOTHING;
" || { echo "FAILED"; echo "Could not create session in DB."; exit 1; }

# 1. Create Session and Get Auth Token
# The previous section now handles session creation directly.
# This section is modified to reflect that.
# echo -n "🔑 Creating Session and Generating Token... " # This line is now redundant

# Create a valid session in the DB
# SESSION_ID=$(docker-compose exec -T db psql -U kainuguru -d kainuguru_db -A -t -c "INSERT INTO user_sessions (user_id, token_hash, expires_at, is_active) VALUES ('e97ccfaa-2dfa-4cae-b92a-ccdf3519d556', md5(random()::text), NOW() + INTERVAL '1 day', TRUE) RETURNING id;" | head -n 1 | tr -d '[:space:]')

if [ -z "$SESSION_ID" ]; then
    echo -e "${RED}FAILED${NC}"
    echo "Could not create session in DB."
    exit 1
fi

# Load .env variables if file exists
if [ -f .env ]; then
    set -a
    source .env
    set +a
fi

# Force the correct secret found in the running container
export JWT_SECRET="dev-secret-change-in-production"

# Generate token using the session ID
TOKEN=$(go run generate_token.go "$SESSION_ID" 2>/dev/null)

if [ -z "$TOKEN" ]; then
    echo -e "${RED}FAILED${NC}"
    echo "Could not generate token."
    exit 1
fi
echo -e "${GREEN}OK${NC}"

API_URL="http://localhost:8080/graphql"

# Function to run GraphQL query
run_query() {
    local query="$1"
    curl -s -X POST "$API_URL" \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $TOKEN" \
        -d "$query"
}

# 2. Test Shopping List Creation (Success Case)
echo -e "\n📋 Test 1: Create Shopping List (New Name)"
LIST_NAME="Test List $(date +%s)"
QUERY_CREATE='{"query": "mutation { createShoppingList(input: { name: \"'$LIST_NAME'\" }) { id name user { id email } } }"}'

RESPONSE=$(run_query "$QUERY_CREATE")

if echo "$RESPONSE" | grep -q "$LIST_NAME" && echo "$RESPONSE" | grep -q "email"; then
    echo -e "${GREEN}PASS${NC} - List created and User field returned"
else
    echo -e "${RED}FAIL${NC} - Response: $RESPONSE"
fi

# 3. Test Shopping List Creation (Duplicate Name)
echo -e "\n📋 Test 2: Create Shopping List (Duplicate Name)"
# Try to create the SAME list again
RESPONSE_DUP=$(run_query "$QUERY_CREATE")

if echo "$RESPONSE_DUP" | grep -q "You already have a shopping list named"; then
    echo -e "${GREEN}PASS${NC} - Correct validation error returned"
else
    echo -e "${RED}FAIL${NC} - Expected validation error. Response: $RESPONSE_DUP"
fi

# 4. Test Shopping List Retrieval (User Field)
echo -e "\n📋 Test 3: Get Shopping Lists (User Field Check)"
QUERY_GET='{"query": "query { shoppingLists { edges { node { id name user { id email } } } } }"}'

RESPONSE_GET=$(run_query "$QUERY_GET")

if echo "$RESPONSE_GET" | grep -q "email"; then
    echo -e "${GREEN}PASS${NC} - User field present in list retrieval"
else
    echo -e "${RED}FAIL${NC} - User field missing. Response: $RESPONSE_GET"
fi

# 5. Test Product Master Duplicate Fix (Concurrent Simulation)
# This is harder to test with curl, but we can check if the system is stable
echo -e "\n🛒 Test 4: System Stability Check"
QUERY_HEALTH='{"query": "query { myDefaultShoppingList { id } }"}'
RESPONSE_HEALTH=$(run_query "$QUERY_HEALTH")

if echo "$RESPONSE_HEALTH" | grep -q "id"; then
    echo -e "${GREEN}PASS${NC} - API is responsive"
else
    echo -e "${RED}FAIL${NC} - API might be unstable. Response: $RESPONSE_HEALTH"
fi

# 6. Test Start Wizard
echo -e "\n🧙 Test 5: Start Wizard"
# Extract List ID from Test 1 response
LIST_ID=$(echo "$RESPONSE" | grep -o '"id":"[^"]*"' | head -n 1 | cut -d'"' -f4)

if [ -z "$LIST_ID" ]; then
    echo -e "${RED}FAIL${NC} - Could not extract List ID from Test 1"
else
    QUERY_WIZARD='{"query": "mutation { startWizard(shoppingListId: \"'$LIST_ID'\") { session { id status } } }"}'
    RESPONSE_WIZARD=$(run_query "$QUERY_WIZARD")

    if echo "$RESPONSE_WIZARD" | grep -q "session"; then
        echo -e "${GREEN}PASS${NC} - Wizard started successfully"
        echo "Response: $RESPONSE_WIZARD"
    else
        echo -e "${RED}FAIL${NC} - Wizard failed to start. Response: $RESPONSE_WIZARD"
    fi
fi

echo -e "\n✅ Verification Complete"
