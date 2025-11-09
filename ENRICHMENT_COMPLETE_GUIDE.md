# Complete Flyer Enrichment Implementation Guide

## Overview

The flyer enrichment system extracts product information from grocery store flyer images using AI (OpenAI Vision API). It creates structured product data with prices, discounts, tags, and establishes cross-store product relationships through the Product Master system.

## ✅ Implementation Status: COMPLETE

All features have been implemented and tested:
- ✅ AI product extraction from images
- ✅ Special discount detection (1+1, 3 už 2, etc.)
- ✅ Automatic tag generation
- ✅ Product master matching with generic names
- ✅ Cross-store product relationships
- ✅ Search functionality
- ✅ Configurable OpenAI/OpenRouter support
- ✅ GraphQL API exposure

## Architecture

### Package Structure
```
cmd/enrich-flyers/
  main.go                    # CLI entry point

internal/services/
  enrichment/
    orchestrator.go          # Coordinates processing
    service.go               # Page-level logic
    utils.go                 # Helpers
  
  ai/
    extractor.go             # AI extraction
    prompt_builder.go        # Prompt construction
    validator.go             # Result validation
  
  product_master_service.go  # Product normalization
  
pkg/openai/
  client.go                  # OpenAI API client
```

## Configuration

### Required Environment Variables

```bash
# Database (must be running)
DB_HOST=localhost
DB_PORT=5439
DB_USER=kainuguru
DB_PASSWORD=kainuguru_password
DB_NAME=kainuguru_db

# OpenAI API
OPENAI_API_KEY=sk-...                      # Required
OPENAI_BASE_URL=https://api.openai.com/v1 # Optional
OPENAI_MODEL=gpt-4o                        # Optional
OPENAI_MAX_TOKENS=4000                     # Optional
OPENAI_TEMPERATURE=0.1                     # Optional

# Storage (where flyer images are stored)
STORAGE_BASE_PATH=../kainuguru-public
```

### OpenAI Provider Options

#### Option 1: Standard OpenAI (Recommended)
```bash
OPENAI_BASE_URL=https://api.openai.com/v1
OPENAI_API_KEY=sk-proj-...
OPENAI_MODEL=gpt-4o
```

#### Option 2: OpenRouter
```bash
OPENAI_BASE_URL=https://openrouter.ai/api/v1
OPENAI_API_KEY=sk-or-v1-...
OPENAI_MODEL=anthropic/claude-3-opus
# or
OPENAI_MODEL=openai/gpt-4o
```

**Important**: 
- OpenRouter URL must include `/api` in the path
- Model names for OpenRouter should use format: `provider/model`
- Do NOT use `openrouter/` prefix (incorrect)

## Usage

### Build Command
```bash
make build-enrich
```

### Command Line Options
```bash
# Dry run - preview what would be processed
./bin/enrich-flyers --dry-run

# Process single page for testing
./bin/enrich-flyers --max-pages=1

# Process specific store
./bin/enrich-flyers --store=iki --max-pages=10

# Enable debug logging
./bin/enrich-flyers --store=iki --max-pages=1 --debug

# Force reprocess already completed pages
./bin/enrich-flyers --force-reprocess

# Custom batch size
./bin/enrich-flyers --store=iki --batch-size=5 --max-pages=20

# Process specific date
./bin/enrich-flyers --date=2025-11-08
```

### Available Flags
- `--store` - Filter by store code (iki/maxima/rimi)
- `--date` - Override target date (YYYY-MM-DD format)
- `--max-pages` - Maximum pages to process (0 = all pages)
- `--batch-size` - Pages per batch (default: 10)
- `--dry-run` - Preview mode, no actual processing
- `--force-reprocess` - Reprocess completed pages
- `--debug` - Enable detailed debug logging
- `--config` - Path to custom config file

## Workflow

### 1. Scrape Flyers
```bash
# Fetch latest flyers from stores
go run cmd/scraper/main.go
```

This creates:
- Flyer records
- Flyer pages with image URLs
- Initial status: `pending`

### 2. Enrich with AI
```bash
# Process all pending pages
./bin/enrich-flyers

# Or test with single page
./bin/enrich-flyers --store=iki --max-pages=1 --debug
```

This:
- Loads flyer page images
- Converts to base64 for AI
- Extracts products with AI
- Creates product records
- Generates tags automatically
- Matches to product masters

### 3. Query via GraphQL
```graphql
query GetProducts {
  products(storeCode: "iki", limit: 10) {
    id
    name
    brand
    category
    price {
      current
      original
      discountPercent
      specialDiscount  # e.g., "1+1", "3 už 2 €"
    }
    tags
    productMaster {
      id
      name  # Generic name without brand
    }
  }
}
```

## Features

### AI Product Extraction

Extracts from flyer images:
- **Product name** (Lithuanian)
- **Current price** (numeric)
- **Original price** (if discounted)
- **Discount percentage**
- **Special discounts** (1+1, 3 už 2, etc.)
- **Unit size** (e.g., "1L", "500g")
- **Unit type** (e.g., "kg", "vnt")
- **Brand name**
- **Category** (if visible)
- **Bounding box** (x, y, width, height)
- **Confidence score** (0-1)

### Automatic Tag Generation

Tags are extracted from:
- **Category** → "pieno-gaminiai", "mesa", "daržovės"
- **Brand** → brand name as tag
- **Discounts** → "nuolaida", "akcija"
- **Units** → "svoris", "tūris"
- **Characteristics** → "ekologiškas", "šviežias", "šaldytas", "lengvas", "naujiena"

Example product tags:
```json
["pieno-gaminiai", "iki", "nuolaida", "šviežias"]
```

### Product Master System

Product masters normalize products across stores:

**Before (Products)**:
- "Saulėgrąžų aliejus NATURA" @ IKI
- "Saulėgrąžų aliejus ELITA" @ Maxima

**After (Product Master)**:
- Name: "Saulėgrąžų aliejus" (generic, no brand)
- Links to both products
- Enables price comparison across stores

**Brand Removal Examples**:
- "IKI varškė" → "Varškė"
- "SOSTINĖS batonas" → "Batonas"
- "Glaistytas varškės sūrelis MAGIJA" → "Glaistytas varškės sūrelis"

### Matching Logic

When a product is enriched:

1. **Search for existing masters**
   - Compare normalized names
   - Check category similarity
   - Calculate confidence score

2. **Auto-link if high confidence** (≥ 0.85)
   - Same category
   - Very similar name
   - Automatic association

3. **Flag for review if medium** (0.65-0.85)
   - Similar but uncertain
   - Requires manual verification

4. **Create new master if no match** (< 0.65)
   - Unique product
   - Generic name generated
   - Available for future matching

### Quality Control

**Validation Checks**:
- Product name is not empty
- Price is a valid positive number
- Discount percent is 0-100 if present
- Unit size format is valid
- Bounding box coordinates are valid

**Retry Logic**:
- Max 3 attempts per page
- Exponential backoff
- Rate limit handling
- Context cancellation support

**Error Tracking**:
- Failed pages logged
- Attempt count tracked
- Status: pending → processing → completed/failed

## Database Schema

### Products Table
```sql
CREATE TABLE products (
  id SERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  normalized_name VARCHAR(255),
  brand VARCHAR(100),
  category VARCHAR(100),
  subcategory VARCHAR(100),
  current_price NUMERIC(10,2) NOT NULL,
  original_price NUMERIC(10,2),
  discount_percent INTEGER,
  special_discount TEXT,           -- NEW: "1+1", "3 už 2"
  unit_size VARCHAR(50),
  unit_type VARCHAR(20),
  tags TEXT[],
  product_master_id BIGINT REFERENCES product_masters(id),
  flyer_page_id INTEGER REFERENCES flyer_pages(id),
  store_id INTEGER REFERENCES stores(id),
  search_vector TSVECTOR,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

### Product Masters Table
```sql
CREATE TABLE product_masters (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,      -- Generic name (no brand)
  normalized_name VARCHAR(255),
  brand VARCHAR(100),              -- Stored separately
  category VARCHAR(100),
  subcategory VARCHAR(100),
  tags TEXT[],
  standard_unit VARCHAR(50),
  match_count INTEGER DEFAULT 0,
  confidence_score FLOAT DEFAULT 0,
  last_seen_date TIMESTAMP,
  status VARCHAR(20) DEFAULT 'active',
  merged_into_id BIGINT,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

### Product Master Matches Table
```sql
CREATE TABLE product_master_matches (
  id BIGSERIAL PRIMARY KEY,
  product_id BIGINT NOT NULL REFERENCES products(id),
  product_master_id BIGINT NOT NULL REFERENCES product_masters(id),
  confidence FLOAT NOT NULL,
  match_type VARCHAR(50),
  review_status VARCHAR(20) DEFAULT 'pending',
  created_at TIMESTAMP DEFAULT NOW()
);
```

## GraphQL API

### Query Products
```graphql
query SearchProducts($query: String!, $storeCode: String, $limit: Int) {
  searchProducts(
    query: $query
    storeCode: $storeCode
    limit: $limit
    preferFuzzy: true
  ) {
    totalCount
    results {
      product {
        id
        name
        brand
        category
        tags
        price {
          current
          original
          discountPercent
          specialDiscount
        }
        productMaster {
          id
          name
          alternativeProducts {
            id
            name
            store {
              code
              name
            }
            price {
              current
            }
          }
        }
      }
      score
    }
  }
}
```

### Query Product Masters
```graphql
query GetProductMasters($category: String, $limit: Int) {
  productMasters(category: $category, limit: $limit) {
    id
    name
    category
    matchCount
    products {
      id
      name
      brand
      store {
        code
        name
      }
      price {
        current
      }
    }
  }
}
```

## Testing

### Unit Test
```bash
# Test individual page
./bin/enrich-flyers --store=iki --max-pages=1 --debug
```

### Verify Database
```bash
# Check products created
docker exec kainuguru-api-db-1 psql -U kainuguru -d kainuguru_db -c \
  "SELECT id, name, brand, special_discount FROM products ORDER BY created_at DESC LIMIT 10;"

# Check product masters
docker exec kainuguru-api-db-1 psql -U kainuguru -d kainuguru_db -c \
  "SELECT id, name, brand, match_count FROM product_masters ORDER BY created_at DESC LIMIT 10;"

# Check tags
docker exec kainuguru-api-db-1 psql -U kainuguru -d kainuguru_db -c \
  "SELECT name, tags FROM products WHERE tags IS NOT NULL LIMIT 5;"

# Test search
docker exec kainuguru-api-db-1 psql -U kainuguru -d kainuguru_db -c \
  "SELECT product_id, name, combined_similarity FROM fuzzy_search_products('varške', 0.3, 10, 0);"
```

### GraphQL Test
```bash
# Start API server
make dev

# Query via GraphQL playground
http://localhost:8080/graphql
```

## Troubleshooting

### Issue: Invalid OpenAI API Key
```
Error: API error: Invalid API key
```
**Solution**: Update `OPENAI_API_KEY` in `.env` with valid key

### Issue: Model Not Found
```
Error: The model `gpt-4-vision-preview` has been deprecated
```
**Solution**: Update `OPENAI_MODEL=gpt-4o` in `.env`

### Issue: Invalid Image URL
```
Error: Invalid image URL: Expected a base64-encoded data URL
```
**Solution**: Image conversion is automatic. Check `STORAGE_BASE_PATH` is correct.

### Issue: Rate Limited
```
Error: rate limited after 1 attempts
```
**Solution**: 
- Wait and retry
- Reduce `--batch-size`
- Increase retry delays in code

### Issue: Search Not Finding Products
```
Query: "varške"
Result: No products found
```
**Solution**:
- Search IS working correctly
- Verify products have `search_vector` populated
- Check if products exist with that name
- Use fuzzy search with lower threshold

### Issue: OpenRouter 404 Error
```
Error: unexpected status code 404
```
**Solution**: 
- Check `OPENAI_BASE_URL=https://openrouter.ai/api/v1` (must include `/api`)
- Verify model name format: `anthropic/claude-3-opus` (not `openrouter/...`)
- Check API key is valid for OpenRouter

### Issue: Product Masters Still Have Brands
```
Name: "Saulėgrąžų aliejus NATURA"
Expected: "Saulėgrąžų aliejus"
```
**Solution**: Run database update:
```sql
UPDATE product_masters SET name = TRIM(
  regexp_replace(name, 'NATURA|MAGIJA|SOSTINĖS|IKI|CLEVER|...', '', 'g')
)
WHERE name ~ '(NATURA|MAGIJA|...)';
```

## Performance

### Benchmarks (Approximate)
- **Single page processing**: 5-10 seconds
- **Products per page**: 5-15 products
- **AI API cost**: ~$0.01-0.05 per page
- **Batch processing**: 10 pages in ~1-2 minutes

### Optimization Tips
1. Use `--batch-size` to control concurrency
2. Use `--max-pages` to limit processing
3. Process during off-peak hours
4. Monitor API rate limits
5. Cache product masters in memory

## Monitoring

### Metrics to Track
- Pages processed per hour
- Products extracted per page
- Average confidence scores
- API error rates
- Processing duration
- Cost per page

### Logs
```bash
# View processing logs
tail -f logs/enrichment.log

# Check for errors
grep "ERR" logs/enrichment.log

# View statistics
grep "Processing summary" logs/enrichment.log
```

## Production Deployment

### Prerequisites
1. PostgreSQL database running
2. Valid OpenAI/OpenRouter API key
3. Flyer images accessible
4. Environment variables configured

### Deployment Steps
```bash
# 1. Build production binary
make build-enrich

# 2. Set environment
export ENV=production

# 3. Run enrichment
./bin/enrich-flyers --store=iki

# 4. Monitor logs
tail -f logs/enrichment.log
```

### Cron Job (Optional)
```bash
# Run enrichment daily at 2 AM
0 2 * * * cd /app/kainuguru-api && ./bin/enrich-flyers >> /var/log/enrichment.log 2>&1
```

## Support & Documentation

- **Implementation Plan**: `FLYER_ENRICHMENT_PLAN.md`
- **AI Prompts**: `FLYER_AI_PROMPTS.md`
- **Previous Fixes**: `ENRICHMENT_FIXES_COMPLETE.md`
- **Final Fixes**: `ENRICHMENT_FINAL_FIXES.md`
- **This Guide**: `ENRICHMENT_COMPLETE_GUIDE.md`

## Conclusion

The flyer enrichment system is **fully implemented and production-ready**. All major features are working:
- ✅ AI extraction
- ✅ Product normalization
- ✅ Cross-store matching
- ✅ Tag generation
- ✅ Special discounts
- ✅ Search functionality
- ✅ GraphQL API

The system is ready for production use with proper configuration and monitoring.
