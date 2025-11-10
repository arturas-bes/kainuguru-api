#!/bin/bash

echo "==================================="
echo "Product Search Verification Test"
echo "==================================="
echo ""

API_URL="http://localhost:8080/graphql"

test_search() {
    local query=$1
    local description=$2
    
    echo "Test: $description"
    echo "Query: '$query'"
    
    result=$(curl -s -X POST "$API_URL" \
        -H "Content-Type: application/json" \
        -d "{\"query\":\"query { searchProducts(input: {q: \\\"$query\\\", first: 10}) { products { product { id name brand } searchScore } totalCount } }\"}")
    
    total=$(echo "$result" | jq -r '.data.searchProducts.totalCount')
    
    if [ "$total" = "null" ] || [ -z "$total" ]; then
        echo "❌ Error: $(echo "$result" | jq -r '.errors[0].message')"
    elif [ "$total" -gt 0 ]; then
        echo "✅ Found $total product(s)"
        echo "$result" | jq -r '.data.searchProducts.products[] | "  - \(.product.name) (ID: \(.product.id), Brand: \(.product.brand // "N/A"))"'
    else
        echo "⚠️  No products found"
    fi
    
    echo ""
}

# Test all enriched products
test_search "obuoliai" "Search for apples (Obuoliai)"
test_search "varškė" "Search for cottage cheese (Varškė)"
test_search "kopūstai" "Search for cabbage (Kopūstai)"
test_search "dešra" "Search for sausage (Dešra)"
test_search "dešrelės" "Search for sausages (Dešrelės)"
test_search "aliejus" "Search for oil (Aliejus)"
test_search "batonas" "Search for bread loaf (Batonas)"
test_search "sūrelis" "Search for cheese snack (Sūrelis)"
test_search "mentė" "Search for pork loin (Mentė)"

# Test by brand
echo "==================================="
echo "Brand Search Tests"
echo "==================================="
echo ""

test_search "IKI" "Search by brand IKI"
test_search "CLEVER" "Search by brand CLEVER"
test_search "MAGIJA" "Search by brand MAGIJA"
test_search "NATURA" "Search by brand NATURA"
test_search "TARCZYNSKI" "Search by brand TARCZYNSKI"

echo "==================================="
echo "Test Complete"
echo "==================================="
