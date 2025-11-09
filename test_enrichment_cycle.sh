#!/bin/bash

# Enrichment Testing Cycle Script
# Usage: ./test_enrichment_cycle.sh [model_name]
# Example: ./test_enrichment_cycle.sh "openai/gpt-4o"

set -e

MODEL=${1:-"openrouter/polaris-alpha"}
MAX_PAGES=${2:-3}

echo "🧪 Starting Enrichment Test Cycle"
echo "================================="
echo "Model: $MODEL"
echo "Max Pages: $MAX_PAGES"
echo ""

# Step 1: Reset products
echo "📝 Step 1: Resetting products..."
go run scripts/reset_products.go
echo "✅ Products reset"
echo ""

# Step 2: Reset page statuses
echo "📝 Step 2: Resetting page extraction status..."
docker exec -i kainuguru-api-db-1 psql -U kainuguru -d kainuguru_db -c \
  "UPDATE flyer_pages 
   SET extraction_status = 'pending', extraction_attempts = 0 
   WHERE flyer_id IN (
     SELECT id FROM flyers 
     WHERE store_id = (SELECT id FROM stores WHERE code = 'iki')
     ORDER BY valid_from DESC LIMIT 1
   ) AND page_number <= $MAX_PAGES;"
echo "✅ Page statuses reset"
echo ""

# Step 3: Rebuild command
echo "📝 Step 3: Building enricher..."
go build -o ./bin/enrich-flyers ./cmd/enrich-flyers
echo "✅ Build complete"
echo ""

# Step 4: Set model and run enrichment
echo "📝 Step 4: Running enrichment..."
echo "Using model: $MODEL"
export OPENAI_MODEL="$MODEL"
./bin/enrich-flyers --store=iki --max-pages=$MAX_PAGES --debug 2>&1 | tee enrichment_test_output.log
echo "✅ Enrichment complete"
echo ""

# Step 5: Validate results
echo "📝 Step 5: Validating results..."
echo ""

echo "Product Count:"
docker exec -i kainuguru-api-db-1 psql -U kainuguru -d kainuguru_db -t -c \
  "SELECT COUNT(*) FROM products WHERE extraction_method = 'ai_vision';" | xargs

echo ""
echo "Average Confidence:"
docker exec -i kainuguru-api-db-1 psql -U kainuguru -d kainuguru_db -t -c \
  "SELECT ROUND(AVG(extraction_confidence)::numeric, 2) 
   FROM products WHERE extraction_method = 'ai_vision';" | xargs

echo ""
echo "Products with Special Discounts:"
docker exec -i kainuguru-api-db-1 psql -U kainuguru -d kainuguru_db -t -c \
  "SELECT COUNT(*) FROM products 
   WHERE extraction_method = 'ai_vision' AND special_discount IS NOT NULL;" | xargs

echo ""
echo "Products by Category:"
docker exec -i kainuguru-api-db-1 psql -U kainuguru -d kainuguru_db -c \
  "SELECT category, COUNT(*) as count 
   FROM products 
   WHERE extraction_method = 'ai_vision' 
   GROUP BY category 
   ORDER BY count DESC;"

echo ""
echo "Sample Products:"
docker exec -i kainuguru-api-db-1 psql -U kainuguru -d kainuguru_db -c \
  "SELECT 
     LEFT(name, 50) as name,
     current_price,
     special_discount,
     ROUND(extraction_confidence::numeric, 2) as confidence
   FROM products 
   WHERE extraction_method = 'ai_vision'
   ORDER BY created_at DESC 
   LIMIT 10;"

echo ""
echo "✅ Test cycle complete!"
echo ""
echo "Expected: ~13 products from $MAX_PAGES pages"
echo "Check enrichment_test_output.log for detailed logs"
