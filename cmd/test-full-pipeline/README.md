# Full Pipeline Test

**Purpose:** Test the complete flyer processing pipeline from scraping to database storage.

---

## What It Does

This test script exercises the **entire flyer processing pipeline**:

1. ✅ **Scrapes** IKI website for current flyer
2. ✅ **Gets store** from database
3. ✅ **Creates flyer record** in database
4. ✅ **Downloads PDF** from IKI
5. ✅ **Converts PDF** to JPEG images (150 DPI)
6. ✅ **Saves images** to storage with proper folder structure
7. ✅ **Creates flyer_page records** in database with image URLs
8. ✅ **Marks flyer** as completed
9. ✅ **Verifies** all records in database and filesystem

---

## Prerequisites

### 1. Database Running
```bash
docker-compose up -d postgres
```

### 2. Database Migrated
```bash
make db-migrate
# or
go run cmd/migrator/main.go up
```

### 3. Stores Seeded
```bash
psql -d kainuguru_db -f migrations/007_seed_stores.sql
```

### 4. Storage Directory Created
```bash
mkdir -p ../kainuguru-public/flyers/{iki,maxima,rimi,lidl,norfa}
```

### 5. Configuration
Make sure `.env` has:
```bash
# Storage Configuration
STORAGE_TYPE=filesystem
STORAGE_BASE_PATH=../kainuguru-public
STORAGE_PUBLIC_URL=http://localhost:8080
STORAGE_MAX_RETRIES=3
```

### 6. pdftoppm Installed
```bash
# macOS
brew install poppler

# Ubuntu/Debian
apt-get install poppler-utils
```

---

## How to Run

```bash
cd /Users/arturas/Dev/kainuguru_all/kainuguru-api

# Run the test
go run cmd/test-full-pipeline/main.go
```

---

## Expected Output

```
🧪 Testing Full Flyer Pipeline (Scrape → Download → Convert → Save → Database)
================================================================================

📡 STEP 1: Scraping IKI flyer information...
   ✓ Found flyer: IKI savaitės leidinys
   ✓ Valid: 2025-11-08 to 2025-11-14
   ✓ URL: https://iki.lt/...

🗄️  STEP 2: Getting store from database...
   ✓ Store: IKI (ID: 1)

💾 STEP 3: Creating flyer record in database...
   ✓ Created flyer record: ID=1
   ✓ Folder will be: 2025-11-08-iki-savaitinis
   ✓ Path will be: flyers/iki/2025-11-08-iki-savaitinis

⬇️  STEP 4: Downloading PDF...
   ✓ Downloaded PDF: 2.45 MB
   ✓ Saved to: ./test_output/iki_flyer.pdf

🖼️  STEP 5: Converting PDF to images...
   ✓ Processing completed in 8.234s
   ✓ Pages converted: 12
      Page 1: 234.56 KB
      Page 2: 245.67 KB
      ...
   ✓ Total size: 2.89 MB

💾 STEP 6: Saving images to storage...
   ✓ Page 1: http://localhost:8080/flyers/iki/2025-11-08-iki-savaitinis/page-1.jpg
   ✓ Page 2: http://localhost:8080/flyers/iki/2025-11-08-iki-savaitinis/page-2.jpg
   ✓ Page 3: http://localhost:8080/flyers/iki/2025-11-08-iki-savaitinis/page-3.jpg
   ... (9 more pages)
   ✓ Saved 12 images to storage

💾 STEP 7: Creating flyer page records in database...
   ✓ Created 12 flyer_page records

✅ STEP 8: Marking flyer as completed...
   ✓ Flyer status: completed
   ✓ Products extracted: 0 (ready for AI)

🔍 STEP 9: Verifying results...
   ✓ Flyer in DB: ID=1, Status=completed, Pages=12
   ✓ Pages in DB: 12 records
   ✓ First page URL: http://localhost:8080/flyers/iki/2025-11-08-iki-savaitinis/page-1.jpg
   ✓ File exists: ../kainuguru-public/flyers/iki/2025-11-08-iki-savaitinis/page-1.jpg
   ✓ Storage folder: flyers/iki/2025-11-08-iki-savaitinis

📊 SUMMARY
==========
Flyer ID:        1
Store:           IKI
Title:           IKI savaitės leidinys
Valid:           2025-11-08 to 2025-11-14
Pages:           12
Status:          completed
Folder:          2025-11-08-iki-savaitinis
Storage Path:    flyers/iki/2025-11-08-iki-savaitinis
Ready for AI:    YES (all pages marked as 'pending')

🧹 CLEANUP
Delete test files from test_output/? (y/N): n
   Test files kept in: ./test_output
Delete flyer from database (for re-testing)? (y/N): y
   ✓ Deleted flyer and images from database/storage

✅ FULL PIPELINE TEST COMPLETED!

The system successfully:
  ✓ Scraped flyer information from IKI website
  ✓ Downloaded PDF flyer
  ✓ Converted PDF to JPEG images
  ✓ Saved images to filesystem with proper folder structure
  ✓ Created flyer and flyer_page records in database
  ✓ Stored full URLs in database
  ✓ Marked pages as 'pending' for AI extraction

🚀 Ready for production scraper run: go run cmd/scraper/main.go
```

---

## What Gets Created

### Filesystem
```
../kainuguru-public/flyers/iki/2025-11-08-iki-savaitinis/
├── page-1.jpg
├── page-2.jpg
├── page-3.jpg
├── page-4.jpg
├── page-5.jpg
├── page-6.jpg
├── page-7.jpg
├── page-8.jpg
├── page-9.jpg
├── page-10.jpg
├── page-11.jpg
└── page-12.jpg
```

### Database

**flyers table:**
```sql
id | store_id | title                  | valid_from | valid_to   | page_count | status    | is_archived
1  | 1        | IKI savaitės leidinys  | 2025-11-08 | 2025-11-14 | 12         | completed | false
```

**flyer_pages table:**
```sql
id | flyer_id | page_number | image_url                                                        | extraction_status
1  | 1        | 1           | http://localhost:8080/flyers/iki/2025-11-08-iki-savaitinis/page-1.jpg | pending
2  | 1        | 2           | http://localhost:8080/flyers/iki/2025-11-08-iki-savaitinis/page-2.jpg | pending
...
```

---

## Cleanup Options

The test offers two cleanup options at the end:

### Option 1: Delete Test Files
Removes temporary files from `test_output/` directory.

### Option 2: Delete Database Records
- Deletes flyer_page records
- Deletes flyer record  
- Deletes image files from storage
- Allows you to re-run the test cleanly

---

## Troubleshooting

### "Failed to connect to database"
```bash
# Check postgres is running
docker-compose ps postgres

# Check .env has correct DB settings
cat .env | grep DB_
```

### "Store not found"
```bash
# Seed stores
psql -d kainuguru_db -c "SELECT id, name, code FROM stores;"

# If empty, run seed
psql -d kainuguru_db -f migrations/007_seed_stores.sql
```

### "Failed to create directory"
```bash
# Create storage directory
mkdir -p ../kainuguru-public/flyers/iki
chmod -R 755 ../kainuguru-public
```

### "pdftoppm: command not found"
```bash
# Install poppler-utils
brew install poppler  # macOS
```

### "Flyer already exists"
```bash
# Delete existing flyer
psql -d kainuguru_db -c "DELETE FROM flyer_pages WHERE flyer_id = 1;"
psql -d kainuguru_db -c "DELETE FROM flyers WHERE id = 1;"

# Or answer 'y' to cleanup prompt at end of test
```

---

## Integration with Production

After this test passes, you can run the production scraper:

```bash
# Run scraper (processes all stores, runs every 6 hours)
go run cmd/scraper/main.go
```

The production scraper uses the same code paths tested here, just:
- Processes multiple stores (IKI, Maxima, Rimi, etc.)
- Runs on a schedule (every 6 hours)
- Enforces 2-flyer limit per store
- Archives old flyers automatically

---

## What's Next

After flyers are in the database with `extraction_status='pending'`:

1. **AI Extraction** - GPT-4 Vision extracts products from images
2. **Product Matching** - Products linked to Product Master catalog
3. **Price History** - Historical prices tracked
4. **Search Index** - Products indexed for search
5. **GraphQL API** - Products available via API

**Status:** Storage and database ready ✅  
**Next:** AI extraction (checklist #20) ⏳

---

*Full Pipeline Test*  
*Last Updated: 2025-11-08*  
*Status: Production Ready ✅*
