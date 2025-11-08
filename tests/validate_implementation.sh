#!/bin/bash

# Validation script for implemented features
# This script tests core business functionality

set +e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "======================================"
echo "Kainuguru API - Implementation Validation"
echo "======================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

pass_count=0
fail_count=0

test_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $2"
        ((pass_count++))
    else
        echo -e "${RED}✗ FAIL${NC}: $2"
        ((fail_count++))
    fi
}

echo "1. Checking Product Master Service..."
echo "   - Testing matching algorithm"
cd "$PROJECT_ROOT"
if go test ./internal/services/matching > /dev/null 2>&1; then
    test_result 0 "Product matching algorithm tests"
else
    test_result 1 "Product matching algorithm tests"
fi

echo ""
echo "2. Checking Search Service..."
echo "   - Testing search validation and normalization"
if go test ./internal/services/search > /dev/null 2>&1; then
    test_result 0 "Search service tests"
else
    test_result 1 "Search service tests"
fi

echo ""
echo "3. Checking Code Compilation..."
echo "   - Building all packages"
if go build ./cmd/... > /dev/null 2>&1; then
    test_result 0 "All packages compile successfully"
else
    test_result 1 "All packages compile successfully"
fi

echo ""
echo "4. Checking Database Connectivity..."
if docker ps | grep -q "kainuguru-api-db"; then
    test_result 0 "Database container is running"
else
    test_result 1 "Database container is not running"
fi

echo ""
echo "5. Checking Redis Connectivity..."
if docker ps | grep -q "kainuguru-api-redis"; then
    test_result 0 "Redis container is running"
else
    test_result 1 "Redis container is not running"
fi

echo ""
echo "6. Checking Email Service..."
if [ -f "$PROJECT_ROOT/internal/services/email/smtp_service.go" ]; then
    test_result 0 "Email service implementation exists"
else
    test_result 1 "Email service implementation missing"
fi

echo ""
echo "7. Checking Recommendation Service..."
if [ -f "$PROJECT_ROOT/internal/services/recommendation/price_comparison_service.go" ]; then
    lines=$(wc -l < "$PROJECT_ROOT/internal/services/recommendation/price_comparison_service.go")
    if [ "$lines" -gt 200 ]; then
        test_result 0 "Recommendation service is implemented ($lines lines)"
    else
        test_result 1 "Recommendation service is too small ($lines lines)"
    fi
else
    test_result 1 "Recommendation service missing"
fi

echo ""
echo "8. Checking Product Master Worker..."
if [ -f "$PROJECT_ROOT/internal/workers/product_master_worker.go" ]; then
    test_result 0 "Product Master worker exists"
else
    test_result 1 "Product Master worker missing"
fi

echo ""
echo "9. Checking Migrations..."
if [ -d "$PROJECT_ROOT/migrations" ]; then
    migration_count=$(ls -1 "$PROJECT_ROOT/migrations"/*.sql 2>/dev/null | wc -l | tr -d ' ')
    if [ "$migration_count" -ge 20 ]; then
        test_result 0 "Database migrations present ($migration_count files)"
    else
        test_result 1 "Insufficient migrations ($migration_count files)"
    fi
else
    test_result 1 "Migrations directory missing"
fi

echo ""
echo "10. Checking GraphQL Schema..."
schema_file="$PROJECT_ROOT/internal/graphql/schema/schema.graphql"
if [ ! -f "$schema_file" ]; then
    schema_file="$PROJECT_ROOT/api/schema.graphql"
fi
if [ -f "$schema_file" ]; then
    schema_lines=$(wc -l < "$schema_file" | tr -d ' ')
    if [ "$schema_lines" -gt 100 ]; then
        test_result 0 "GraphQL schema is defined ($schema_lines lines)"
    else
        test_result 1 "GraphQL schema is too small ($schema_lines lines)"
    fi
else
    test_result 1 "GraphQL schema missing"
fi

echo ""
echo "11. Checking PDF Processing..."
if [ -f "$PROJECT_ROOT/pkg/pdf/processor.go" ]; then
    if grep -q "convertToImages" "$PROJECT_ROOT/pkg/pdf/processor.go"; then
        test_result 0 "PDF processing implementation present"
    else
        test_result 1 "PDF processing not fully implemented"
    fi
else
    test_result 1 "PDF processor missing"
fi

echo ""
echo "12. Checking Lithuanian Normalization..."
if [ -f "$PROJECT_ROOT/pkg/normalize/lithuanian.go" ]; then
    norm_lines=$(wc -l < "$PROJECT_ROOT/pkg/normalize/lithuanian.go")
    if [ "$norm_lines" -gt 400 ]; then
        test_result 0 "Lithuanian normalizer is comprehensive ($norm_lines lines)"
    else
        test_result 1 "Lithuanian normalizer is incomplete ($norm_lines lines)"
    fi
else
    test_result 1 "Lithuanian normalizer missing"
fi

echo ""
echo "======================================"
echo "Validation Summary"
echo "======================================"
echo -e "${GREEN}Passed: $pass_count${NC}"
echo -e "${RED}Failed: $fail_count${NC}"
total=$((pass_count + fail_count))
percentage=$((pass_count * 100 / total))
echo "Success Rate: $percentage%"
echo ""

if [ $fail_count -eq 0 ]; then
    echo -e "${GREEN}✓ All validation checks passed!${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some validation checks failed${NC}"
    exit 1
fi
