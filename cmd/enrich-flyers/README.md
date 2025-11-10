# Flyer Enrichment Command

AI-powered flyer enrichment service that extracts product information from flyer images using OpenAI's Vision API.

## Overview

This command processes flyer pages that have been scraped and stored, extracting product information including:
- Product names (Lithuanian)
- Prices (current and original)
- Discounts
- Units and quantities
- Brands
- Categories
- Bounding boxes and positions

## Usage

### Basic Usage

```bash
# Process all active flyers for today
./bin/enrich-flyers

# Process specific store
./bin/enrich-flyers --store=iki

# Process for specific date
./bin/enrich-flyers --date=2025-11-20

# Dry run (preview what would be processed)
./bin/enrich-flyers --dry-run

# Debug mode with verbose logging
./bin/enrich-flyers --debug
```

### Advanced Options

```bash
# Limit number of pages
./bin/enrich-flyers --max-pages=50

# Adjust batch size
./bin/enrich-flyers --batch-size=5

# Force reprocess completed pages
./bin/enrich-flyers --force-reprocess

# Combine options
./bin/enrich-flyers --store=maxima --date=2025-11-15 --max-pages=20 --debug
```

## Command Line Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--store` | string | "" | Process specific store (iki/maxima/rimi) |
| `--date` | string | today | Override date (YYYY-MM-DD format) |
| `--force-reprocess` | bool | false | Reprocess completed pages |
| `--max-pages` | int | 0 | Maximum pages to process (0=all) |
| `--batch-size` | int | 10 | Pages per batch |
| `--dry-run` | bool | false | Preview what would be processed |
| `--debug` | bool | false | Enable debug logging |
| `--config` | string | "" | Path to custom config file |

## How It Works

### Processing Flow

1. **Flyer Selection**: Gets all active flyers for the specified date
2. **Page Filtering**: Identifies pages that need processing (pending/failed)
3. **AI Extraction**: Uses OpenAI Vision API to extract products from each page
4. **Quality Assessment**: Evaluates extraction quality and flags issues
5. **Product Creation**: Stores extracted products in the database
6. **Product Matching**: Links products to product masters (future)

### Quality Control

The system automatically:
- Flags pages with no products as "warning"
- Marks low product count pages for manual review
- Validates prices and data formats
- Tracks extraction confidence scores
- Limits retry attempts to 3 per page

### Database Updates

Products are stored with:
- Normalized names for search
- Full-text search vectors
- Valid date ranges from flyer
- Extraction metadata (confidence, method)
- Bounding box coordinates
- Processing timestamps

## Configuration

### Environment Variables

Required:
- `OPENAI_API_KEY` - OpenAI API key for Vision API

Optional:
- `ENV` - Environment (development/production)
- `DATABASE_*` - Database connection settings
- `OPENAI_*` - OpenAI configuration

### Config File

See `config/development.yaml` and `config/production.yaml` for full configuration options.

## Examples

### Process Today's IKI Flyers

```bash
./bin/enrich-flyers --store=iki
```

Output:
```
{"level":"info","time":1699459200,"message":"Starting Flyer Enrichment Service"}
{"level":"info","message":"Database connection established"}
{"level":"info","message":"Starting flyer processing"}
{"level":"info","count":1,"message":"Found eligible flyers"}
{"level":"info","flyer_id":123,"store":"IKI","message":"Processing flyer"}
{"level":"info","flyer_id":123,"pages_processed":20,"products_extracted":156,"message":"Flyer processing completed"}
```

### Dry Run for Specific Date

```bash
./bin/enrich-flyers --date=2025-11-20 --dry-run
```

Output:
```
{"level":"info","message":"Dry run mode - listing flyers that would be processed:"}
{"level":"info","id":123,"store":"IKI","valid_from":"2025-11-18","valid_to":"2025-11-24","message":"Would process flyer"}
{"level":"info","id":124,"store":"Maxima","valid_from":"2025-11-19","valid_to":"2025-11-25","message":"Would process flyer"}
```

### Limited Processing with Debug

```bash
./bin/enrich-flyers --max-pages=10 --debug
```

## Error Handling

### Transient Errors
- API rate limits
- Network timeouts
- Temporary service issues

**Action**: Automatic retry with exponential backoff (max 3 attempts)

### Permanent Errors
- Invalid image URLs
- Corrupted images
- Missing images

**Action**: Mark as failed, require manual intervention

### Quality Issues
- Empty pages (0 products)
- Low product count (< 5)
- Low confidence (< 0.5)

**Action**: Mark as "warning", flag for manual review

## Monitoring

### Metrics Tracked

- Pages processed
- Products extracted
- Average confidence scores
- Processing duration
- Failure rates
- API token usage

### Log Levels

- `INFO`: Normal operation, progress updates
- `WARN`: Quality issues, retryable errors
- `ERROR`: Processing failures, system issues
- `DEBUG`: Detailed extraction data, API calls

## Scheduling

### Cron Examples

```cron
# Run daily at 6 AM for current day's flyers
0 6 * * * /path/to/enrich-flyers --date=$(date +\%Y-\%m-\%d)

# Retry failed pages every 4 hours
0 */4 * * * /path/to/enrich-flyers --force-reprocess

# Process specific store daily
0 7 * * * /path/to/enrich-flyers --store=iki
```

## Troubleshooting

### No Flyers Found

**Cause**: No active flyers for the specified date  
**Solution**: Check that flyers have been scraped and are within valid date range

### OpenAI API Error

**Cause**: Invalid API key or rate limit exceeded  
**Solution**: Verify `OPENAI_API_KEY` environment variable, check API quota

### Low Product Count

**Cause**: Poor image quality, complex layout, non-standard format  
**Solution**: Images will be flagged for manual review

### Database Connection Failed

**Cause**: Database not running or incorrect credentials  
**Solution**: Check database connection settings in config

## Performance

### Processing Speed
- Average: 10-15 pages per minute
- Per page: 3-5 seconds (including API calls)
- Batch size affects memory usage and parallelization

### Cost Estimates
- Per page: ~$0.05-0.10 (2000 tokens @ $0.03/1K)
- Per flyer (20 pages): ~$1.00-2.00
- Daily (5 flyers): ~$5.00-10.00
- Monthly: ~$150-300

### Resource Usage
- Memory: ~50-100 MB base + 5MB per concurrent page
- CPU: Minimal (I/O bound)
- Network: ~1-2 MB per page (image upload)

## Development

### Build

```bash
go build -o bin/enrich-flyers cmd/enrich-flyers/*.go
```

### Test

```bash
# Unit tests
go test ./internal/services/...

# Integration tests
go test -tags=integration ./cmd/enrich-flyers/...
```

### Adding New Features

1. Update `enrichment_service.go` for core logic
2. Update `enricher.go` for command orchestration
3. Add tests
4. Update this README

## Related Documentation

- [Flyer Enrichment Plan](../../FLYER_ENRICHMENT_PLAN.md)
- [AI Prompts](../../FLYER_AI_PROMPTS.md)
- [Developer Guidelines](../../DEVELOPER_GUIDELINES.md)

## Support

For issues or questions:
1. Check logs with `--debug` flag
2. Review flyer page status in database
3. Check OpenAI API status
4. Contact development team
