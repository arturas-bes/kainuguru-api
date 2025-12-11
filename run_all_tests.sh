#!/bin/bash

# run_all_tests.sh - Master test runner for Kainuguru API
# Executes all integration tests in a structured order

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
API_URL="${API_URL:-http://localhost:8080/graphql}"
HEALTH_URL="${HEALTH_URL:-http://localhost:8080/health}"

# Test counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0
SKIPPED_TESTS=0

# Results array
declare -a TEST_RESULTS

# Function to print colored message
print_message() {
  local color=$1
  local message=$2
  echo -e "${color}${message}${NC}"
}

# Function to print section header
print_header() {
  local title=$1
  echo ""
  print_message "$BLUE" "═══════════════════════════════════════════════════════════════"
  print_message "$BLUE" "  $title"
  print_message "$BLUE" "═══════════════════════════════════════════════════════════════"
  echo ""
}

# Function to check prerequisites
check_prerequisites() {
  print_header "Checking Prerequisites"
  
  # Check API health
  print_message "$YELLOW" "Checking API server health..."
  if curl -sf "$HEALTH_URL" > /dev/null 2>&1; then
    print_message "$GREEN" "✅ API server is healthy"
  else
    print_message "$RED" "❌ API server is not available at $HEALTH_URL"
    exit 1
  fi
  
  # Check Redis
  print_message "$YELLOW" "Checking Redis connectivity..."
  if redis-cli ping > /dev/null 2>&1; then
    print_message "$GREEN" "✅ Redis is available"
  else
    print_message "$RED" "❌ Redis is not available"
    exit 1
  fi
  
  # Check API token
  if [ -z "$API_TOKEN" ]; then
    print_message "$RED" "❌ API_TOKEN environment variable is not set"
    print_message "$YELLOW" "ℹ️  Export your JWT token: export API_TOKEN='your-token-here'"
    exit 1
  else
    print_message "$GREEN" "✅ API_TOKEN is set"
  fi
  
  # Check jq
  if ! command -v jq &> /dev/null; then
    print_message "$RED" "❌ jq is not installed (required for JSON parsing)"
    exit 1
  else
    print_message "$GREEN" "✅ jq is available"
  fi
  
  echo ""
  print_message "$GREEN" "All prerequisites met! Ready to run tests."
}

# Function to run a test script
run_test() {
  local script=$1
  local description=$2
  
  TOTAL_TESTS=$((TOTAL_TESTS + 1))
  
  echo ""
  print_message "$BLUE" "▶ Running: $description ($script)"
  echo "─────────────────────────────────────────────────────────────"
  
  # Run the test script
  if ./"$script"; then
    print_message "$GREEN" "✅ PASSED: $description"
    PASSED_TESTS=$((PASSED_TESTS + 1))
    TEST_RESULTS+=("✅ PASS: $description")
  else
    print_message "$RED" "❌ FAILED: $description"
    FAILED_TESTS=$((FAILED_TESTS + 1))
    TEST_RESULTS+=("❌ FAIL: $description")
  fi
}

# Function to run SQL fixtures
run_fixtures() {
  print_header "Loading Test Fixtures"
  
  print_message "$YELLOW" "Loading test_fixtures.sql..."
  
  # Try to load fixtures via Docker
  if docker-compose exec -T db psql -U kainuguru -d kainuguru_db < test_fixtures.sql 2>&1; then
    print_message "$GREEN" "✅ Test fixtures loaded successfully"
  else
    print_message "$YELLOW" "⚠️  Could not load fixtures via Docker, skipping..."
    print_message "$YELLOW" "ℹ️  You may need to run manually: psql -U kainuguru -d kainuguru_db -f test_fixtures.sql"
  fi
}

# Function to print summary
print_summary() {
  print_header "Test Execution Summary"
  
  echo ""
  print_message "$BLUE" "Test Results:"
  echo ""
  
  for result in "${TEST_RESULTS[@]}"; do
    if [[ $result == ✅* ]]; then
      print_message "$GREEN" "$result"
    elif [[ $result == ❌* ]]; then
      print_message "$RED" "$result"
    else
      print_message "$YELLOW" "$result"
    fi
  done
  
  echo ""
  print_message "$BLUE" "─────────────────────────────────────────────────────────────"
  echo ""
  print_message "$BLUE" "Total Tests:   $TOTAL_TESTS"
  print_message "$GREEN" "Passed Tests:  $PASSED_TESTS"
  print_message "$RED" "Failed Tests:  $FAILED_TESTS"
  print_message "$YELLOW" "Skipped Tests: $SKIPPED_TESTS"
  echo ""
  
  # Calculate pass rate
  if [ $TOTAL_TESTS -gt 0 ]; then
    PASS_RATE=$((PASSED_TESTS * 100 / TOTAL_TESTS))
    if [ $PASS_RATE -eq 100 ]; then
      print_message "$GREEN" "🎉 Pass Rate: ${PASS_RATE}% - ALL TESTS PASSED!"
    elif [ $PASS_RATE -ge 80 ]; then
      print_message "$YELLOW" "⚠️  Pass Rate: ${PASS_RATE}%"
    else
      print_message "$RED" "❌ Pass Rate: ${PASS_RATE}%"
    fi
  fi
  
  echo ""
  
  # Exit with failure if any tests failed
  if [ $FAILED_TESTS -gt 0 ]; then
    exit 1
  fi
}

# Main execution
main() {
  print_header "Kainuguru API - Integration Test Suite"
  
  print_message "$BLUE" "Test Configuration:"
  echo "  API URL:    $API_URL"
  echo "  Health URL: $HEALTH_URL"
  echo ""
  
  # Check prerequisites
  check_prerequisites
  
  # Optionally load fixtures
  if [ "$LOAD_FIXTURES" = "1" ]; then
    run_fixtures
  else
    print_message "$YELLOW" "ℹ️  Skipping fixture loading (set LOAD_FIXTURES=1 to enable)"
  fi
  
  # Run tests in logical order
  print_header "Executing Test Suite"
  
  # Phase 1: Basic CRUD Tests
  print_message "$BLUE" "Phase 1: Basic CRUD Operations"
  run_test "test_shopping_list.sh" "Shopping List CRUD"
  run_test "test_shopping_items.sh" "Shopping List Items CRUD"
  run_test "test_create.sh" "Create Shopping List"
  run_test "test_delete_item.sh" "Delete Shopping List Item"
  
  # Phase 2: Core Feature Tests
  print_message "$BLUE" "Phase 2: Core Features"
  run_test "test_product_master.sh" "Product Master Operations"
  run_test "test_search_verification.sh" "Search Functionality"
  
  # Phase 3: Enrichment Pipeline Tests
  print_message "$BLUE" "Phase 3: Enrichment Pipeline"
  run_test "test_enrichment.sh" "Product Enrichment"
  run_test "test_enrichment_cycle.sh" "Full Enrichment Cycle"
  
  # Phase 4: Wizard Integration Tests (NEW)
  print_message "$BLUE" "Phase 4: Wizard Integration (MVP)"
  run_test "test_wizard_integration.sh" "Wizard Full Integration"
  
  # Phase 5: Infrastructure Tests
  print_message "$BLUE" "Phase 5: Infrastructure"
  run_test "test-docker-pdf.sh" "Docker PDF Processing"
  
  # Print final summary
  print_summary
}

# Run main function
main
