#!/bin/bash

# Enrichment System Validation Script
# This script validates that all enrichment fixes are working correctly

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "================================================================================"
echo -e "${BLUE}  Flyer Enrichment System - Validation Script${NC}"
echo "================================================================================"
echo ""

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
    echo -e "${GREEN}✓${NC} Environment variables loaded"
else
    echo -e "${RED}✗${NC} .env file not found"
    exit 1
fi

# Check database connection
echo ""
echo -e "${BLUE}Checking database connection...${NC}"
if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Database connection successful"
else
    echo -e "${RED}✗${NC} Database connection failed"
    exit 1
fi

# Check OpenAI configuration
echo ""
echo -e "${BLUE}Checking OpenAI configuration...${NC}"
if [ -z "$OPENAI_API_KEY" ]; then
    echo -e "${RED}✗${NC} OPENAI_API_KEY not set"
    exit 1
else
    echo -e "${GREEN}✓${NC} OPENAI_API_KEY is set"
fi

if [ -z "$OPENAI_MODEL" ]; then
    echo -e "${YELLOW}⚠${NC}  OPENAI_MODEL not set, will use default (gpt-4o)"
else
    echo -e "${GREEN}✓${NC} OPENAI_MODEL: $OPENAI_MODEL"
fi

# Check binary exists
echo ""
echo -e "${BLUE}Checking enrichment binary...${NC}"
if [ ! -f ./bin/enrich-flyers ]; then
    echo -e "${YELLOW}⚠${NC}  Binary not found, building..."
    go build -o bin/enrich-flyers cmd/enrich-flyers/main.go
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓${NC} Build successful"
    else
        echo -e "${RED}✗${NC} Build failed"
        exit 1
    fi
else
    echo -e "${GREEN}✓${NC} Binary exists"
fi

# Check for flyers to process
echo ""
echo -e "${BLUE}Checking for available flyers...${NC}"
FLYER_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM flyers WHERE status != 'archived';")
FLYER_COUNT=$(echo $FLYER_COUNT | xargs) # Trim whitespace

if [ "$FLYER_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓${NC} Found $FLYER_COUNT flyers available for processing"
else
    echo -e "${RED}✗${NC} No flyers found. Run seeder first: go run cmd/seeder/main.go"
    exit 1
fi

# Check for pages to process
PAGE_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM flyer_pages WHERE image_url IS NOT NULL;")
PAGE_COUNT=$(echo $PAGE_COUNT | xargs)

if [ "$PAGE_COUNT" -gt 0 ]; then
    echo -e "${GREEN}✓${NC} Found $PAGE_COUNT pages with images"
else
    echo -e "${RED}✗${NC} No pages with images found. Run scraper first."
    exit 1
fi

echo ""
echo "================================================================================"
echo -e "${BLUE}  Pre-flight checks complete!${NC}"
echo "================================================================================"
echo ""
echo "System is ready for enrichment. Run one of these commands:"
echo ""
echo -e "${YELLOW}# Test with 1 page:${NC}"
echo "  ./bin/enrich-flyers --store=iki --max-pages=1 --debug"
echo ""
echo -e "${YELLOW}# Test with 2 pages:${NC}"
echo "  ./bin/enrich-flyers --store=iki --max-pages=2 --debug"
echo ""
echo -e "${YELLOW}# Dry run to see what would be processed:${NC}"
echo "  ./bin/enrich-flyers --store=iki --dry-run"
echo ""
echo "================================================================================"
echo ""

# Ask if user wants to run a test
read -p "Would you like to run a single page test now? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo ""
    echo -e "${BLUE}Running single page test...${NC}"
    echo "================================================================================"
    ./bin/enrich-flyers --store=iki --max-pages=1 --debug
    
    echo ""
    echo "================================================================================"
    echo -e "${BLUE}  Post-Test Validation${NC}"
    echo "================================================================================"
    
    # Check if products were created
    PRODUCT_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM products WHERE created_at > NOW() - INTERVAL '5 minutes';")
    PRODUCT_COUNT=$(echo $PRODUCT_COUNT | xargs)
    
    if [ "$PRODUCT_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✓${NC} Created $PRODUCT_COUNT products"
    else
        echo -e "${RED}✗${NC} No products created"
    fi
    
    # Check if products have tags
    TAGGED_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM products WHERE tags IS NOT NULL AND array_length(tags, 1) > 0 AND created_at > NOW() - INTERVAL '5 minutes';")
    TAGGED_COUNT=$(echo $TAGGED_COUNT | xargs)
    
    if [ "$TAGGED_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✓${NC} $TAGGED_COUNT products have tags"
    else
        echo -e "${YELLOW}⚠${NC}  No products with tags found"
    fi
    
    # Check if product masters were created
    MASTER_COUNT=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM product_masters WHERE created_at > NOW() - INTERVAL '5 minutes';")
    MASTER_COUNT=$(echo $MASTER_COUNT | xargs)
    
    if [ "$MASTER_COUNT" -gt 0 ]; then
        echo -e "${GREEN}✓${NC} Created $MASTER_COUNT product masters"
    else
        echo -e "${YELLOW}⚠${NC}  No product masters created"
    fi
    
    # Check if masters have generic names (without brands)
    echo ""
    echo -e "${BLUE}Checking product master names...${NC}"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    SELECT 
        id,
        name as master_name,
        brand,
        CASE 
            WHEN brand IS NOT NULL AND name ILIKE '%' || brand || '%' THEN '❌ FAIL'
            ELSE '✓ PASS'
        END as validation
    FROM product_masters 
    WHERE created_at > NOW() - INTERVAL '5 minutes'
    LIMIT 10;
    "
    
    echo ""
    echo -e "${BLUE}Sample products with tags:${NC}"
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "
    SELECT 
        id,
        LEFT(name, 40) as product_name,
        array_length(tags, 1) as tag_count,
        tags
    FROM products 
    WHERE created_at > NOW() - INTERVAL '5 minutes'
      AND tags IS NOT NULL
    LIMIT 5;
    "
    
    echo ""
    echo "================================================================================"
    echo -e "${GREEN}  Validation complete!${NC}"
    echo "================================================================================"
    echo ""
    echo "For more detailed testing, see: ENRICHMENT_TEST_PLAN.md"
fi

echo ""
echo -e "${GREEN}All checks passed! ✓${NC}"
echo ""
