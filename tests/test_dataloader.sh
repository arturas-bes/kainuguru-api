#!/bin/bash

# DataLoader Performance Test Script
# This script runs GraphQL queries and helps verify DataLoader is working

set -e

# Configuration
API_URL="${API_URL:-http://localhost:8080/graphql}"
AUTH_TOKEN="${AUTH_TOKEN:-}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper function to print colored output
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Helper function to run GraphQL query
run_query() {
    local query_name=$1
    local query=$2

    print_header "Running: $query_name"

    # Prepare headers
    local headers="Content-Type: application/json"
    if [ -n "$AUTH_TOKEN" ]; then
        headers="$headers|Authorization: Bearer $AUTH_TOKEN"
    fi

    # Run query and measure time
    local start_time=$(date +%s%N)

    local response=$(curl -s -X POST "$API_URL" \
        -H "Content-Type: application/json" \
        $([ -n "$AUTH_TOKEN" ] && echo "-H \"Authorization: Bearer $AUTH_TOKEN\"") \
        -d "{\"query\": $(echo "$query" | jq -Rs .)}")

    local end_time=$(date +%s%N)
    local duration=$(( (end_time - start_time) / 1000000 )) # Convert to ms

    # Check for errors
    if echo "$response" | jq -e '.errors' > /dev/null 2>&1; then
        print_error "Query failed with errors:"
        echo "$response" | jq '.errors'
        return 1
    fi

    # Print results
    print_success "Response time: ${duration}ms"

    # Try to extract data count
    local data_count=$(echo "$response" | jq -r '.data | .. | objects | select(.edges) | .edges | length' 2>/dev/null | head -1)
    if [ -n "$data_count" ]; then
        echo "  Items returned: $data_count"
    fi

    # Check response time
    if [ "$duration" -lt 100 ]; then
        print_success "Excellent performance! (< 100ms)"
    elif [ "$duration" -lt 300 ]; then
        print_warning "Good performance (100-300ms)"
    else
        print_warning "Slow response (> 300ms) - check query optimization"
    fi

    echo ""
    return 0
}

# Check if API is running
print_header "Checking API Availability"
if curl -s -f "$API_URL" > /dev/null 2>&1 || curl -s -f "${API_URL/graphql/health}" > /dev/null 2>&1; then
    print_success "API is running at $API_URL"
else
    print_error "API is not accessible at $API_URL"
    echo "Please start the API server first:"
    echo "  make run"
    echo "  or"
    echo "  go run cmd/api/main.go"
    exit 1
fi

# Test 1: Simple Products Query (Baseline)
print_header "TEST 1: Baseline - Products Only"
echo "Expected: 1 query, ~20-50ms"
run_query "Baseline_ProductsOnly" '
query Baseline_ProductsOnly {
  products(first: 20) {
    edges {
      node {
        id
        name
        currentPrice
      }
    }
  }
}'

# Test 2: Products with Nested Fields
print_header "TEST 2: Products with Store, Flyer, ProductMaster"
echo "Expected: 4-6 queries, ~50-100ms"
echo "WITHOUT DataLoader: 60-100 queries, 500-800ms"
run_query "Test1_ProductsWithNesting" '
query Test1_ProductsWithNesting {
  products(first: 20) {
    edges {
      node {
        id
        name
        currentPrice
        store {
          id
          name
          code
        }
        flyer {
          id
          title
        }
        productMaster {
          id
          canonicalName
        }
      }
    }
  }
}'

# Test 3: Flyers with Store
print_header "TEST 3: Flyers with Store"
echo "Expected: 2 queries, ~30-70ms"
run_query "Test3_FlyersWithStore" '
query Test3_FlyersWithStore {
  flyers(first: 20) {
    edges {
      node {
        id
        title
        store {
          id
          name
          code
        }
      }
    }
  }
}'

# Test 4: Complex Nested Query (Stress Test)
print_header "TEST 4: Complex Nested - Stores > Flyers > Products"
echo "Expected: 5-10 queries, ~100-200ms"
echo "WITHOUT DataLoader: 200+ queries, 2000+ ms"
run_query "Test4_ComplexNested" '
query Test4_ComplexNested {
  stores(first: 5) {
    edges {
      node {
        id
        name
        flyers(first: 3) {
          edges {
            node {
              id
              title
              products(first: 10) {
                edges {
                  node {
                    id
                    name
                    currentPrice
                    productMaster {
                      id
                      canonicalName
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}'

# Summary
print_header "Test Summary"
echo "All tests completed!"
echo ""
echo "To verify DataLoader is working:"
echo "1. Check server logs for SQL queries"
echo "2. Look for 'WHERE id IN (...)' patterns (batched queries)"
echo "3. Confirm query counts are low (< 10 per test)"
echo ""
echo "Enable query logging with:"
echo "  export BUN_DEBUG=1"
echo "  go run cmd/api/main.go"
echo ""
print_success "If all tests completed with good response times, DataLoader is working!"
