# Enrichment Quick Reference Guide

**Last Updated:** 2025-11-09

---

## 🚀 Quick Start

### Build
```bash
go build -o bin/enrich-flyers cmd/enrich-flyers/main.go
```

### Run
```bash
# Preview (no changes)
./bin/enrich-flyers --dry-run

# Process one page
./bin/enrich-flyers --max-pages=1

# Process specific store
./bin/enrich-flyers --store=iki --max-pages=10

# Debug mode
./bin/enrich-flyers --debug
```

---

## 📋 Command Options

```bash
--store=<code>          # iki, maxima, rimi
--date=YYYY-MM-DD       # Override date (default: today)
--max-pages=<n>         # Limit pages (0=all)
--batch-size=<n>        # Pages per batch (default: 10)
--force-reprocess       # Reprocess completed pages
--dry-run               # Preview only, no changes
--debug                 # Verbose logging
```

---

## 🔧 Environment Variables

### Required
```bash
DB_HOST=localhost
DB_PORT=5439
DB_USER=kainuguru
DB_PASSWORD=***
DB_NAME=kainuguru_db

OPENAI_API_KEY=sk-proj-***  # Or sk-or-v1-*** for OpenRouter
```

### Optional (with defaults)
```bash
OPENAI_MODEL=gpt-4o                          # Default: gpt-4o
OPENAI_BASE_URL=https://api.openai.com/v1   # Default: OpenAI
OPENAI_MAX_TOKENS=4000                       # Default: 4000
OPENAI_TEMPERATURE=0.1                       # Default: 0.1
OPENAI_TIMEOUT=120s                          # Default: 120s
OPENAI_MAX_RETRIES=3                         # Default: 3

FLYER_BASE_URL=http://localhost:8080         # For image URLs
```

---

## 🤖 AI Provider Options

### OpenAI (Recommended)
```bash
OPENAI_API_KEY=sk-proj-[YOUR_KEY]
OPENAI_MODEL=gpt-4o
OPENAI_BASE_URL=https://api.openai.com/v1
```

### OpenRouter - Google Gemini
```bash
OPENAI_API_KEY=sk-or-v1-[YOUR_KEY]
OPENAI_MODEL=google/gemini-2.5-flash-lite
OPENAI_BASE_URL=https://openrouter.ai/api/v1
```

### OpenRouter - Claude
```bash
OPENAI_API_KEY=sk-or-v1-[YOUR_KEY]
OPENAI_MODEL=anthropic/claude-3-opus
OPENAI_BASE_URL=https://openrouter.ai/api/v1
```

---

## 📊 Common Queries

### Check Enrichment Status
```sql
-- Overall stats
SELECT 
  COUNT(*) as total_products,
  COUNT(special_discount) as with_special_discount,
  AVG(extraction_confidence) as avg_confidence
FROM products;

-- Recent enrichments
SELECT 
  f.store_id,
  COUNT(p.id) as products,
  AVG(p.extraction_confidence) as confidence
FROM flyer_pages fp
JOIN products p ON p.flyer_page_id = fp.id
JOIN flyers f ON f.id = fp.flyer_id
WHERE p.created_at > NOW() - INTERVAL '1 day'
GROUP BY f.store_id;
```

### Check Page Status
```sql
SELECT 
  extraction_status,
  COUNT(*) as count
FROM flyer_pages
GROUP BY extraction_status;
```

### Products with Special Discounts
```sql
SELECT 
  name,
  current_price,
  special_discount
FROM products
WHERE special_discount IS NOT NULL
LIMIT 10;
```

---

## 🐛 Troubleshooting

### Issue: No Products Extracted
**Check:**
```bash
# View logs
./bin/enrich-flyers --debug

# Check page status
SELECT id, extraction_status, extraction_error 
FROM flyer_pages 
WHERE extraction_status != 'completed';
```

**Solutions:**
1. Switch to OpenAI GPT-4o
2. Check API key validity
3. Verify image files exist
4. Check API rate limits

---

### Issue: Rate Limiting
**Error:** `rate limited after 1 attempts`

**Solutions:**
1. Reduce `OPENAI_MAX_RETRIES`
2. Increase batch delay
3. Upgrade API tier
4. Switch provider

---

### Issue: Image Not Found
**Error:** `failed to read image from .../page-X.jpg`

**Check:**
```bash
# Verify image exists
ls -la ../kainuguru-public/flyers/iki/*/page-*.jpg

# Check database
SELECT image_url FROM flyer_pages WHERE id=X;
```

**Solutions:**
1. Run scraper first
2. Check `STORAGE_BASE_PATH`
3. Verify file permissions

---

## 📈 Performance Tips

### Optimize for Speed
```bash
# Smaller batches
--batch-size=5

# Limit pages
--max-pages=20

# Reduce retries
OPENAI_MAX_RETRIES=1
```

### Optimize for Quality
```bash
# Larger batches
--batch-size=10

# More retries
OPENAI_MAX_RETRIES=3

# Better model
OPENAI_MODEL=gpt-4o
```

### Optimize for Cost
```bash
# Use cheaper model
OPENAI_MODEL=google/gemini-2.5-flash-lite

# Reduce tokens
OPENAI_MAX_TOKENS=2000

# Process selectively
--store=iki --max-pages=50
```

---

## 🎯 Common Workflows

### Daily Enrichment
```bash
#!/bin/bash
# enrich-daily.sh

./bin/enrich-flyers --store=iki --max-pages=100
./bin/enrich-flyers --store=maxima --max-pages=100
./bin/enrich-flyers --store=rimi --max-pages=100
```

### Re-process Failed Pages
```bash
./bin/enrich-flyers --force-reprocess
```

### Test New AI Model
```bash
# Backup current settings
cp .env .env.backup

# Test new model
OPENAI_MODEL=gpt-4o ./bin/enrich-flyers --max-pages=5 --debug

# Compare results
psql -d kainuguru_db -c "SELECT * FROM products WHERE created_at > NOW() - INTERVAL '5 minutes';"
```

---

## 📦 Maintenance

### Weekly Tasks
- Check extraction success rate
- Review products with low confidence
- Monitor API costs
- Cleanup old flyers

### Monthly Tasks
- Optimize prompts
- A/B test AI models
- Review special discount extraction
- Update product masters

---

## 🆘 Quick Help

**Build fails:**
```bash
go mod download
go mod tidy
```

**Database connection fails:**
```bash
# Check Docker
docker ps | grep postgres

# Check credentials
cat .env | grep DB_
```

**API key invalid:**
```bash
# Test key
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"
```

---

## 📚 Full Documentation

- `SYSTEM_VALIDATION_COMPLETE.md` - Complete validation report
- `ENRICHMENT_FINAL_STATUS.md` - Detailed system status
- `ENRICHMENT_COMPREHENSIVE_FIXES.md` - All fixes explained
- `DEVELOPER_GUIDELINES.md` - Development standards

---

**Quick Reference Version:** 1.0  
**Last Updated:** 2025-11-09
