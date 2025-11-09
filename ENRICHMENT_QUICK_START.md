# Flyer Enrichment - Quick Start Guide

## Prerequisites

1. **Database Running:**
   ```bash
   docker-compose up -d db
   ```

2. **Environment Variables Set:**
   ```bash
   # Copy and edit .env
   cp .env.dist .env
   
   # Required variables:
   OPENAI_API_KEY=your_key_here
   OPENAI_BASE_URL=https://openrouter.ai/api/v1  # or OpenAI URL
   OPENAI_MODEL=openai/gpt-4o
   ```

3. **Build the Command:**
   ```bash
   make build-enrich
   ```

## Quick Commands

### Test with One Page
```bash
./bin/enrich-flyers --store=iki --max-pages=1 --debug
```

### Process Multiple Pages
```bash
./bin/enrich-flyers --store=iki --max-pages=10
```

### Dry Run (No Changes)
```bash
./bin/enrich-flyers --store=iki --max-pages=5 --dry-run
```

### All Stores
```bash
./bin/enrich-flyers --max-pages=50 --batch-size=10
```

## Verify Results

### Check Products
```bash
docker exec -i kainuguru-api-db-1 psql -U kainuguru -d kainuguru_db -c \
  "SELECT id, name, current_price, special_discount FROM products ORDER BY id DESC LIMIT 10;"
```

### Check Tags
```bash
docker exec -i kainuguru-api-db-1 psql -U kainuguru -d kainuguru_db -c \
  "SELECT id, name, tags FROM products WHERE tags IS NOT NULL LIMIT 10;"
```

### Check Product Masters
```bash
docker exec -i kainuguru-api-db-1 psql -U kainuguru -d kainuguru_db -c \
  "SELECT id, normalized_name, original_name FROM product_masters ORDER BY id DESC LIMIT 10;"
```

## GraphQL Query

```graphql
query {
  products(storeCode: "iki", limit: 10) {
    name
    brand
    tags
    price {
      current
      original
      specialDiscount
    }
    productMaster {
      normalizedName
    }
  }
}
```

## Command Options

```
--store string         Store code (iki, maxima, rimi)
--date string          Process flyers valid on date (YYYY-MM-DD)
--max-pages int        Maximum pages to process (default: unlimited)
--batch-size int       Pages per batch (default: 10)
--dry-run             Show what would be processed
--force-reprocess     Reprocess completed pages
--debug               Enable debug logging
```

## What Gets Extracted

✅ **Product Information:**
- Name (with Lithuanian characters)
- Current price
- Original price
- Brand
- Category
- Unit/quantity

✅ **Special Features:**
- **Special Discounts:** "1+1", "SUPER KAINA", "TIK", etc.
- **Tags:** Auto-generated from category, brand, characteristics
- **Product Masters:** Normalized names without brands

✅ **Quality Data:**
- Confidence scores
- Bounding boxes
- Page positions

## Common Issues

### "Invalid API Key"
- Check `OPENAI_API_KEY` in `.env`
- Ensure key is valid for OpenRouter or OpenAI

### "Rate Limited"
- Reduce `--batch-size`
- Add delays between runs
- Check your API tier limits

### "No Pages to Process"
- Check if flyers exist: `SELECT * FROM flyers;`
- Run scraper first: `go run cmd/scraper/main.go`
- Use `--force-reprocess` to reprocess

### "Database Connection Failed"
- Verify database is running: `docker ps`
- Check DB credentials in `.env`

## Performance Tips

1. **Batch Size:** Start with 5-10 pages per batch
2. **Max Pages:** Limit to 20-50 for testing
3. **Debug Mode:** Only use for troubleshooting (verbose)
4. **Store Filter:** Process one store at a time to manage costs

## Cost Estimation

**Approximate costs with OpenRouter (openai/gpt-4o):**
- Per page: ~$0.01-0.03
- 100 pages: ~$1-3
- Full flyer (50 pages): ~$0.50-1.50

**Factors affecting cost:**
- Image size
- Products per page
- Model used
- Retry attempts

## Monitoring

### Watch Progress
```bash
# In one terminal
./bin/enrich-flyers --store=iki --max-pages=10 --debug

# In another terminal
watch -n 2 'docker exec -i kainuguru-api-db-1 psql -U kainuguru -d kainuguru_db -c "SELECT COUNT(*) FROM products;"'
```

### Check Logs
```bash
# JSON logs
./bin/enrich-flyers --store=iki --max-pages=1 2>&1 | jq

# Human-readable
./bin/enrich-flyers --store=iki --max-pages=1 --debug
```

## Success Indicators

After running enrichment, you should see:

✅ **In Logs:**
```
✅ Found eligible flyers count=X
✅ Processing pages flyer_id=X page_count=Y
✅ Flyer processing completed products_extracted=Z
✅ Enrichment completed successfully
```

✅ **In Database:**
```sql
-- Products with special discounts
SELECT COUNT(*) FROM products WHERE special_discount IS NOT NULL;

-- Products with tags
SELECT COUNT(*) FROM products WHERE tags IS NOT NULL;

-- Product masters created
SELECT COUNT(*) FROM product_masters;
```

✅ **In GraphQL:**
Query returns products with complete data including special_discount and tags.

## Next Steps

1. **Run Initial Enrichment:**
   ```bash
   ./bin/enrich-flyers --store=iki --max-pages=5
   ```

2. **Verify Results in Database**

3. **Test GraphQL API**

4. **Scale Up Processing:**
   ```bash
   ./bin/enrich-flyers --max-pages=100 --batch-size=10
   ```

5. **Set Up Cron Job (Optional):**
   ```bash
   # Daily at 2 AM
   0 2 * * * /path/to/bin/enrich-flyers --max-pages=50
   ```

## Support

For issues or questions, check:
- `ENRICHMENT_COMPLETE_STATUS.md` - Full feature documentation
- `ENRICHMENT_VALIDATED.md` - Test results and validation
- `ENRICHMENT_SPECIAL_DISCOUNT_FIX.md` - Special discount details
- Code comments in `internal/services/enrichment/`

---

**Ready to enrich your flyers!** 🚀
