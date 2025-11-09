# Enrichment Fixes - Quick Summary

## Files Modified

### 1. Models
- ✅ `internal/models/product.go` - Added `SpecialDiscount` field

### 2. Database
- ✅ `migrations/032_add_special_discount_to_products.sql` - New migration (APPLIED)

### 3. AI Services
- ✅ `internal/services/ai/extractor.go` - Added SpecialDiscount to ExtractedProduct
- ✅ `pkg/openai/client.go` - Model now reads from OPENAI_MODEL env var

### 4. Enrichment Services
- ✅ `internal/services/enrichment/service.go` - Maps special_discount to products
- ✅ `internal/services/enrichment/utils.go` - Tag extraction already implemented

### 5. GraphQL
- ✅ `internal/graphql/schema/schema.graphql` - Added specialDiscount to ProductPrice
- ✅ `internal/graphql/resolvers/product.go` - Resolver includes special_discount

### 6. Product Master
- ✅ `internal/services/product_master_service.go` - Brand normalization already working

## What Works Now

1. **Special Discounts** - Captured and exposed via API (e.g., "1+1", "3 už 2 €")
2. **Tags** - Auto-generated from product characteristics  
3. **Product Masters** - Generic names without brands for cross-store matching
4. **Max Pages** - Limit properly enforced
5. **OpenAI Model** - Configurable via environment variable

## Action Required

**Update OpenAI API Key:**
```bash
# Edit .env file
OPENAI_API_KEY=sk-your-actual-key-here
```

Then test:
```bash
./bin/enrich-flyers --store=iki --max-pages=1 --debug
```

## Verification Queries

```sql
-- Check products with special discounts
SELECT name, current_price, special_discount, tags 
FROM products 
WHERE special_discount IS NOT NULL 
LIMIT 10;

-- Check product masters (should have generic names)
SELECT name, brand 
FROM product_masters 
ORDER BY match_count DESC 
LIMIT 10;

-- Count tagged products
SELECT COUNT(*), cardinality(tags) as tag_count 
FROM products 
WHERE tags IS NOT NULL AND array_length(tags, 1) > 0;
```
