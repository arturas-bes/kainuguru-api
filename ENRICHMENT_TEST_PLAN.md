# Flyer Enrichment System - Test Plan

## Pre-Test Setup

### 1. Verify Environment Configuration
```bash
# Check all required environment variables are set
cat .env | grep -E "(OPENAI|DB_)" 

# Ensure model is correct
grep OPENAI_MODEL .env
# Should show: OPENAI_MODEL=gpt-4o
```

### 2. Verify Database is Running
```bash
# Check database connectivity
make db-status

# Or manually
PGPASSWORD=$DB_PASSWORD psql -h localhost -p 5439 -U kainuguru_user -d kainuguru_db -c "SELECT COUNT(*) FROM stores;"
```

### 3. Verify Flyer Data Exists
```bash
# Check for flyers and pages in database
./bin/enrich-flyers --dry-run

# Should show available flyers to process
```

## Test Suite

### Test 1: Single Page Enrichment ✅
**Objective**: Verify basic enrichment works for one page

```bash
# Run enrichment for 1 page
./bin/enrich-flyers --store=iki --max-pages=1 --debug

# Expected outcome:
# - 1 page processed
# - Products extracted with AI
# - Products saved to database
# - Tags populated
# - Product masters created or linked
# - No errors
```

**Validation SQL**:
```sql
-- Check latest products
SELECT 
    id, 
    name, 
    brand,
    current_price,
    tags,
    product_master_id
FROM products 
ORDER BY created_at DESC 
LIMIT 5;

-- Check product masters created
SELECT 
    id,
    name,
    brand,
    tags,
    match_count
FROM product_masters 
ORDER BY created_at DESC 
LIMIT 5;

-- Verify tags are populated
SELECT COUNT(*) as products_with_tags 
FROM products 
WHERE tags IS NOT NULL AND array_length(tags, 1) > 0;

-- Verify masters have generic names (without brands in name)
SELECT 
    id,
    name,
    brand,
    CASE 
        WHEN name ILIKE '%' || brand || '%' THEN 'FAIL: Brand in name'
        ELSE 'PASS: Generic name'
    END as validation
FROM product_masters 
WHERE brand IS NOT NULL
ORDER BY created_at DESC 
LIMIT 10;
```

### Test 2: Max Pages Limit Enforcement ✅
**Objective**: Verify --max-pages flag stops after specified pages

```bash
# Process exactly 2 pages
./bin/enrich-flyers --store=iki --max-pages=2 --debug

# Check logs for:
# - "Processing pages flyer_id=X page_count=N"
# - "Reached maximum pages limit" message
# - Total pages_processed should be exactly 2
```

**Expected Log Output**:
```
Processing pages flyer_id=16 page_count=57
Processing page page_id=X page_number=1
Processing page page_id=Y page_number=2
Reached maximum pages limit after flyer processing
Processing summary flyers_processed=1 pages_processed=2
```

### Test 3: Tags Extraction ✅
**Objective**: Verify tags are correctly extracted from product data

**Check for various tag types**:
```sql
-- Products with category tags
SELECT name, tags 
FROM products 
WHERE 'mėsa ir žuvis' = ANY(tags) 
LIMIT 5;

-- Products with discount tags
SELECT name, tags 
FROM products 
WHERE 'nuolaida' = ANY(tags) OR 'akcija' = ANY(tags)
LIMIT 5;

-- Products with characteristic tags
SELECT name, tags 
FROM products 
WHERE 'ekologiškas' = ANY(tags) OR 'šviežias' = ANY(tags)
LIMIT 5;

-- Tag distribution
SELECT 
    unnest(tags) as tag,
    COUNT(*) as count
FROM products
WHERE tags IS NOT NULL
GROUP BY tag
ORDER BY count DESC
LIMIT 20;
```

### Test 4: Product Master Normalization ✅
**Objective**: Verify product masters have brand-free generic names

**Test Cases**:

1. **"Saulėgrąžų aliejus NATURA 1L"** should create master:
   - name: "Saulėgrąžų aliejus"
   - brand: "NATURA"

2. **"Glaistytas varškės sūrelis MAGIJA"** should create master:
   - name: "Glaistytas varškės sūrelis"
   - brand: "MAGIJA"

3. **"SOSTINĖS batonas"** should create master:
   - name: "Batonas"
   - brand: "SOSTINĖS"

4. **"IKI varškė 9%"** should create master:
   - name: "Varškė"
   - brand: "IKI"

**Validation**:
```sql
-- Check that no masters have brands in their names
SELECT 
    pm.id,
    pm.name as master_name,
    pm.brand,
    p.name as original_product_name
FROM product_masters pm
JOIN products p ON p.product_master_id = pm.id
WHERE pm.brand IS NOT NULL
  AND pm.name ILIKE '%' || pm.brand || '%'
ORDER BY pm.created_at DESC;

-- Should return 0 rows

-- Verify masters are properly normalized
SELECT 
    name,
    brand,
    match_count,
    tags
FROM product_masters
WHERE brand IS NOT NULL
ORDER BY created_at DESC
LIMIT 20;
```

### Test 5: Product Matching ✅
**Objective**: Verify products correctly link to existing masters

```bash
# Run enrichment twice with force-reprocess
./bin/enrich-flyers --store=iki --max-pages=1 --force-reprocess

# Check matching worked
```

**Validation**:
```sql
-- Check match counts on masters
SELECT 
    name,
    brand,
    match_count,
    confidence_score
FROM product_masters
WHERE match_count > 1
ORDER BY match_count DESC
LIMIT 10;

-- Check products linked to same master
SELECT 
    pm.name as master_name,
    pm.brand as master_brand,
    COUNT(p.id) as linked_products,
    array_agg(p.name ORDER BY p.id LIMIT 3) as sample_products
FROM product_masters pm
JOIN products p ON p.product_master_id = pm.id
GROUP BY pm.id, pm.name, pm.brand
HAVING COUNT(p.id) > 1
ORDER BY COUNT(p.id) DESC
LIMIT 10;
```

### Test 6: Error Handling ✅
**Objective**: Verify system handles errors gracefully

**Test scenarios**:

1. **Invalid API Key**:
```bash
# Temporarily set invalid key
OPENAI_API_KEY=invalid ./bin/enrich-flyers --store=iki --max-pages=1
# Should show clear error message
```

2. **Rate Limiting**:
```bash
# Process many pages quickly
./bin/enrich-flyers --store=iki --max-pages=20 --debug
# Should handle rate limits with retries
```

3. **Missing Image**:
```sql
-- Create a test page with invalid image URL
INSERT INTO flyer_pages (flyer_id, page_number, image_url, extraction_status)
VALUES (16, 999, '/invalid/path.jpg', 'pending');

-- Run enrichment
./bin/enrich-flyers --store=iki --debug
# Should mark page as failed with clear error
```

### Test 7: Performance ✅
**Objective**: Measure and validate performance metrics

```bash
# Process 5 pages and measure time
time ./bin/enrich-flyers --store=iki --max-pages=5 --debug

# Expected metrics:
# - Time per page: 2-5 seconds
# - Products per page: 5-20
# - Token usage per page: 1000-3000 tokens
# - Success rate: > 90%
```

**Performance Queries**:
```sql
-- Average extraction time
SELECT 
    AVG(
        EXTRACT(EPOCH FROM updated_at - created_at)
    ) as avg_processing_seconds
FROM flyer_pages
WHERE extraction_status = 'completed'
  AND updated_at > NOW() - INTERVAL '1 hour';

-- Products extracted per page
SELECT 
    fp.id as page_id,
    COUNT(p.id) as products_extracted
FROM flyer_pages fp
LEFT JOIN products p ON p.flyer_page_id = fp.id
WHERE fp.extraction_status = 'completed'
GROUP BY fp.id
ORDER BY fp.id DESC
LIMIT 10;

-- Extraction success rate
SELECT 
    extraction_status,
    COUNT(*) as count,
    ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER (), 2) as percentage
FROM flyer_pages
GROUP BY extraction_status
ORDER BY count DESC;
```

### Test 8: Data Quality ✅
**Objective**: Verify extracted data meets quality standards

```sql
-- Check for products with invalid prices
SELECT id, name, current_price
FROM products
WHERE current_price <= 0 OR current_price > 1000
LIMIT 10;

-- Check for products with missing required fields
SELECT 
    COUNT(*) as total,
    COUNT(name) as has_name,
    COUNT(current_price) as has_price,
    COUNT(category) as has_category,
    COUNT(brand) as has_brand
FROM products
WHERE created_at > NOW() - INTERVAL '1 hour';

-- Check extraction confidence scores
SELECT 
    CASE 
        WHEN extraction_confidence >= 0.8 THEN 'High (>= 0.8)'
        WHEN extraction_confidence >= 0.6 THEN 'Medium (0.6-0.8)'
        ELSE 'Low (< 0.6)'
    END as confidence_level,
    COUNT(*) as count
FROM products
WHERE extraction_method = 'ai_vision'
GROUP BY confidence_level
ORDER BY MIN(extraction_confidence) DESC;
```

## Post-Test Validation

### 1. Review Logs
```bash
# Check for errors in recent run
grep -E "(ERR|FATAL)" logs/enrichment.log

# Check for warnings
grep "WRN" logs/enrichment.log

# Check processing summary
grep "Processing summary" logs/enrichment.log
```

### 2. Database Integrity
```sql
-- Verify no orphaned products
SELECT COUNT(*) 
FROM products p
LEFT JOIN flyers f ON p.flyer_id = f.id
WHERE f.id IS NULL;

-- Verify no orphaned product masters
SELECT COUNT(*)
FROM product_masters pm
WHERE pm.match_count = 0
  AND pm.status = 'active';

-- Check for duplicate masters
SELECT 
    normalized_name,
    brand,
    COUNT(*) as duplicates
FROM product_masters
WHERE status = 'active'
GROUP BY normalized_name, brand
HAVING COUNT(*) > 1;
```

### 3. Cost Analysis
```bash
# Estimate OpenAI costs
# Assuming $0.01 per 1K tokens for gpt-4o
# Average 2000 tokens per page = $0.02 per page
```

```sql
-- Calculate theoretical cost based on pages processed
SELECT 
    COUNT(*) as pages_processed,
    COUNT(*) * 2000 as estimated_tokens,
    ROUND(COUNT(*) * 2000 * 0.01 / 1000, 2) as estimated_cost_usd
FROM flyer_pages
WHERE extraction_status = 'completed'
  AND updated_at > NOW() - INTERVAL '1 day';
```

## Success Criteria

### Must Pass:
- [ ] Single page enrichment works without errors
- [ ] Max pages limit is enforced correctly
- [ ] Tags are populated for all products
- [ ] Product masters have generic names (no brands)
- [ ] Brand is stored in separate column
- [ ] Products successfully link to masters
- [ ] No database integrity issues

### Should Pass:
- [ ] Processing time < 5 seconds per page
- [ ] Success rate > 90%
- [ ] Confidence scores > 0.7 average
- [ ] No rate limiting errors (or handled gracefully)

### Nice to Have:
- [ ] Processing time < 3 seconds per page
- [ ] Success rate > 95%
- [ ] Confidence scores > 0.8 average
- [ ] Zero manual review required

## Troubleshooting Guide

### Issue: "rate limited after N attempts"
**Solution**: 
- Increase `OPENAI_MAX_RETRIES` in .env
- Add delay between pages with `--batch-size` smaller number
- Use exponential backoff (already implemented)

### Issue: "failed to load image"
**Solution**:
- Verify image paths in database match actual files
- Check `kainuguru-public` directory is accessible
- Ensure images were downloaded by scraper

### Issue: "No products extracted"
**Solution**:
- Check page image quality
- Verify OpenAI model has vision capabilities
- Review AI prompt in `prompt_builder.go`
- Check page doesn't just have headers/footers

### Issue: "Failed to create product master"
**Solution**:
- Check for unique constraint violations
- Verify normalized_name generation
- Check database constraints

## Cleanup After Tests

```sql
-- Optional: Remove test data
DELETE FROM products WHERE created_at > NOW() - INTERVAL '1 hour';
DELETE FROM product_masters WHERE created_at > NOW() - INTERVAL '1 hour';
UPDATE flyer_pages SET extraction_status = 'pending', extraction_attempts = 0 
WHERE extraction_status IN ('completed', 'failed');
```

## Next Steps After Tests Pass

1. **Production Deployment**:
   - Review and merge code changes
   - Update production environment variables
   - Deploy to production
   - Monitor first production runs

2. **Monitoring**:
   - Set up alerting for failed enrichments
   - Monitor OpenAI costs
   - Track success rates
   - Review flagged products

3. **Optimization**:
   - Implement caching for common products
   - Optimize matching algorithms
   - Add batch processing for multiple pages
   - Implement parallel processing

4. **Documentation**:
   - Update API documentation
   - Create operator runbook
   - Document common issues and solutions
