#!/bin/bash
# Test script for enrichment functionality

set -e

echo "🧪 Testing Enrichment Functionality"
echo "===================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if OpenAI API key is set
if [ -z "$OPENAI_API_KEY" ]; then
    echo -e "${RED}❌ OPENAI_API_KEY not set in environment${NC}"
    echo "Please set it in .env file or export it:"
    echo "  export OPENAI_API_KEY=sk-your-key-here"
    exit 1
fi

# Check if API key looks valid
if [[ ! "$OPENAI_API_KEY" =~ ^sk- ]]; then
    echo -e "${YELLOW}⚠️  Warning: OpenAI API key should start with 'sk-'${NC}"
    echo "Current key starts with: ${OPENAI_API_KEY:0:10}..."
fi

DB_CONTAINER=$(docker compose ps -q db)
if [ -z "$DB_CONTAINER" ]; then
    echo -e "${RED}❌ Database container not running. Start docker compose first.${NC}"
    exit 1
fi

echo ""
echo "1️⃣  Checking database connection..."
if docker exec "$DB_CONTAINER" psql -U kainuguru -d kainuguru_db -c "SELECT 1" > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Database connected${NC}"
else
    echo -e "${RED}❌ Database not accessible${NC}"
    exit 1
fi

echo ""
echo "2️⃣  Checking if stores exist..."
STORE_COUNT=$(docker exec "$DB_CONTAINER" psql -U kainuguru -d kainuguru_db -t -c "SELECT COUNT(*) FROM stores;" | tr -d ' ')
if [ "$STORE_COUNT" -gt "0" ]; then
    echo -e "${GREEN}✅ Found $STORE_COUNT stores${NC}"
else
    echo -e "${YELLOW}⚠️  No stores found. Run: make seed-data${NC}"
fi

echo ""
echo "3️⃣  Checking if special_discount column exists..."
if docker exec "$DB_CONTAINER" psql -U kainuguru -d kainuguru_db -c "\d products" 2>/dev/null | grep -q "special_discount"; then
    echo -e "${GREEN}✅ special_discount column exists${NC}"
else
    echo -e "${RED}❌ special_discount column missing${NC}"
    echo "Run migration: docker exec -i kainuguru-api-db-1 psql -U kainuguru -d kainuguru_db < migrations/032_add_special_discount_to_products.sql"
    exit 1
fi

echo ""
echo "4️⃣  Testing enrichment with dry-run..."
if [ ! -x ./bin/enrich-flyers ]; then
    mkdir -p ./bin
    go build -o ./bin/enrich-flyers ./cmd/enrich-flyers
fi

if ./bin/enrich-flyers --store=iki --dry-run > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Dry-run successful${NC}"
else
    echo -e "${YELLOW}⚠️  Dry-run failed (may be no flyers to process)${NC}"
fi

echo ""
echo "5️⃣  Running enrichment on 1 page (this will use API credits)..."
echo -e "${YELLOW}Processing...${NC}"

# Run enrichment and capture output
OUTPUT=$(./bin/enrich-flyers --store=iki --max-pages=1 --debug 2>&1)

# Check for errors
if echo "$OUTPUT" | grep -q "API key"; then
    echo -e "${RED}❌ API Key Error${NC}"
    echo "$OUTPUT" | grep "error" | head -3
    exit 1
elif echo "$OUTPUT" | grep -q "Failed to process"; then
    echo -e "${RED}❌ Processing Failed${NC}"
    echo "$OUTPUT" | grep -E "ERR|error" | head -5
    exit 1
elif echo "$OUTPUT" | grep -q "products_extracted\":0"; then
    echo -e "${YELLOW}⚠️  No products extracted (check logs)${NC}"
    echo "$OUTPUT" | grep "INF" | tail -5
else
    echo -e "${GREEN}✅ Enrichment completed${NC}"
    echo "$OUTPUT" | grep "products_extracted" | tail -1
fi

echo ""
echo "6️⃣  Checking database for results..."

# Check if products were created
PRODUCT_COUNT=$(docker exec "$DB_CONTAINER" psql -U kainuguru -d kainuguru_db -t -c "SELECT COUNT(*) FROM products;" | tr -d ' ')
echo "Total products: $PRODUCT_COUNT"

# Check products with tags
TAGGED_COUNT=$(docker exec "$DB_CONTAINER" psql -U kainuguru -d kainuguru_db -t -c "SELECT COUNT(*) FROM products WHERE tags IS NOT NULL AND array_length(tags, 1) > 0;" | tr -d ' ')
echo "Products with tags: $TAGGED_COUNT"

# Check products with special discounts
SPECIAL_DISCOUNT_COUNT=$(docker exec "$DB_CONTAINER" psql -U kainuguru -d kainuguru_db -t -c "SELECT COUNT(*) FROM products WHERE special_discount IS NOT NULL;" | tr -d ' ')
echo "Products with special discounts: $SPECIAL_DISCOUNT_COUNT"

# Check product masters
MASTER_COUNT=$(docker exec "$DB_CONTAINER" psql -U kainuguru -d kainuguru_db -t -c "SELECT COUNT(*) FROM product_masters;" | tr -d ' ')
echo "Product masters: $MASTER_COUNT"

if [ "$PRODUCT_COUNT" -gt "0" ]; then
    echo ""
    echo "7️⃣  Sample products:"
    docker exec "$DB_CONTAINER" psql -U kainuguru -d kainuguru_db -c "
        SELECT 
            SUBSTRING(name, 1, 40) as name,
            current_price,
            COALESCE(special_discount, 'none') as discount,
            array_length(tags, 1) as tag_count
        FROM products 
        ORDER BY id DESC 
        LIMIT 5;
    "
    
    if [ "$MASTER_COUNT" -gt "0" ]; then
        echo ""
        echo "8️⃣  Sample product masters (should be generic names):"
        docker exec "$DB_CONTAINER" psql -U kainuguru -d kainuguru_db -c "
            SELECT 
                SUBSTRING(name, 1, 40) as name,
                COALESCE(brand, 'no brand') as brand,
                match_count
            FROM product_masters 
            ORDER BY match_count DESC, id DESC
            LIMIT 5;
        "
    fi
fi

echo ""
echo -e "${GREEN}================================${NC}"
echo -e "${GREEN}✅ Enrichment Test Complete!${NC}"
echo -e "${GREEN}================================${NC}"

echo ""
echo "Next steps:"
echo "1. Review sample products above"
echo "2. Verify product masters have generic names (brands removed)"
echo "3. Check that tags are populated"
echo "4. Run full enrichment: ./bin/enrich-flyers --store=iki --max-pages=10"
